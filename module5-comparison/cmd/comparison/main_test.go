package comparisoncli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module5-comparison/pkg/comparison"
)

func TestBuildComparisonResult_PopulatesArchitectures(t *testing.T) {
	w := comparison.MNISTWorkload()
	comp := comparison.CompareArchitectures(w, 1, 10000)
	adv := comparison.CalculateAdvantages(comp)

	res := buildComparisonResult(comp, adv, "mnist", 10000)
	if res.Workload != "mnist" {
		t.Fatalf("workload=%q want mnist", res.Workload)
	}
	if len(res.Architectures) == 0 {
		t.Fatal("expected architectures in JSON result")
	}
	for _, a := range res.Architectures {
		if a.Name == "" {
			t.Fatal("architecture name should not be empty")
		}
	}
}

func TestRunComparisonCLIE2EWideJSONWorkloadMatrix(t *testing.T) {
	workloads := []string{"mnist", "resnet", "bert", "gpt2", "llm", "unknown"}
	for _, workload := range workloads {
		t.Run(workload, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), workload+".json")
			var stdout, stderr bytes.Buffer
			if err := runComparison([]string{"--json", "--all", "--workload", workload, "--throughput", "1234", "--output", out}, &stdout, &stderr); err != nil {
				t.Fatalf("runComparison json %s: %v stderr=%s", workload, err, stderr.String())
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("output-file JSON should not write buffers stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			raw, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("read JSON output: %v", err)
			}
			var result ComparisonJSONResult
			if err := json.Unmarshal(raw, &result); err != nil {
				t.Fatalf("invalid JSON: %v data=%s", err, raw)
			}
			if result.Workload != workload || result.Throughput != 1234 || len(result.Architectures) != 3 || result.Advantages.VsCPU_EnergyReduction <= 0 {
				t.Fatalf("JSON result invalid: %+v", result)
			}
		})
	}
}

func TestRunComparisonCLIE2ETextConfigHelpAndInvalidMatrix(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "comparison.json")
	if err := os.WriteFile(cfg, []byte(`{"workload":"bert","throughput":4321}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	var stdout, stderr bytes.Buffer
	global := captureComparisonStdoutE2E(t, func() {
		if err := runComparison([]string{"--config", cfg, "--specs", "--inference", "--datacenter", "--advantages", "--no-color"}, &stdout, &stderr); err != nil {
			t.Fatalf("text config run: %v stderr=%s", err, stderr.String())
		}
	})
	for _, marker := range []string{"MODEL INPUTS ONLY", "ARCHITECTURE SPECIFICATIONS", "BERT-Base", "Data Center", "FeCIM Advantages"} {
		if !strings.Contains(global, marker) {
			t.Fatalf("text output missing %q:\n%s", marker, global)
		}
	}
	stdout.Reset()
	stderr.Reset()
	if err := runComparison([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(stdout.String(), "FeCIM Architecture Comparison CLI") || !strings.Contains(stdout.String(), "--json") {
		t.Fatalf("help stdout missing usage: %q", stdout.String())
	}
	badCfg := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(badCfg, []byte(`{"workload":`), 0644); err != nil {
		t.Fatalf("write bad cfg: %v", err)
	}
	if err := runComparison([]string{"--config", badCfg, "--json", "--output", filepath.Join(tmp, "badout.json")}, &stdout, &stderr); err == nil {
		t.Fatal("malformed config should fail")
	}
	if err := runComparison([]string{"--json", "--output", filepath.Join(tmp, "missing", "out.json")}, &stdout, &stderr); err == nil {
		t.Fatal("missing output dir should fail")
	}
}

func captureComparisonStdoutE2E(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	_ = r.Close()
	return buf.String()
}

func TestRunComparisonReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runComparison([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runComparison error = nil, want invalid flag error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want invalid flag context", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want error prefix", text)
	}
	if !strings.Contains(text, "FeCIM Architecture Comparison CLI") {
		t.Fatalf("stderr = %q, want usage", text)
	}
}
