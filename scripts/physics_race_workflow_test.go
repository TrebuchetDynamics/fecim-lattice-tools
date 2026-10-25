package scripts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNightlyPhysicsRaceHasDedicatedLongRunningJob(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "nightly.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		On struct {
			Schedule []map[string]string `yaml:"schedule"`
			Dispatch yaml.Node           `yaml:"workflow_dispatch"`
		} `yaml:"on"`
		Jobs map[string]struct {
			Timeout  int       `yaml:"timeout-minutes"`
			Continue yaml.Node `yaml:"continue-on-error"`
			Steps    []struct {
				Timeout  int       `yaml:"timeout-minutes"`
				Run      string    `yaml:"run"`
				If       string    `yaml:"if"`
				Uses     string    `yaml:"uses"`
				Continue yaml.Node `yaml:"continue-on-error"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(body, &workflow); err != nil {
		t.Fatal(err)
	}
	if len(workflow.On.Schedule) == 0 || workflow.On.Dispatch.Kind == 0 {
		t.Fatal("physics lane needs scheduled and manual triggers")
	}
	job, ok := workflow.Jobs["physics-race"]
	if !ok {
		t.Fatal("missing dedicated physics-race job; broad runs must not share the short local/PR timeout")
	}
	if job.Timeout < 90 || job.Continue.Kind != 0 {
		t.Fatal("physics job needs >=90 minutes and must not mask failure")
	}
	found, upload := false, false
	for _, step := range job.Steps {
		if strings.Contains(step.Run, "go test -race") {
			found = true
			if step.Timeout < 75 || step.Timeout >= job.Timeout {
				t.Errorf("physics race step timeout=%d, want >=75 minutes and room within job timeout=%d for log upload", step.Timeout, job.Timeout)
			}
			for _, token := range []string{"-count=1", "-timeout=60m", "./shared/physics/...", "./module1-hysteresis/...", "./shared/crossbar/..."} {
				if !strings.Contains(step.Run, token) {
					t.Errorf("physics race command missing %s", token)
				}
			}
			if strings.Contains(step.Run, "-short") || step.If != "" || step.Continue.Kind != 0 {
				t.Error("full physics race must not be short, conditional, or non-gating")
			}
			if strings.Contains(step.Run, "tee") && !strings.Contains(step.Run, "set -euo pipefail") {
				t.Error("race log pipeline must preserve test exit status")
			}
		}
		if strings.HasPrefix(step.Uses, "actions/upload-artifact@") && step.If == "always()" {
			upload = true
		}
	}
	if !found || !upload {
		t.Fatal("physics race command and always-uploaded logs are required")
	}
}
