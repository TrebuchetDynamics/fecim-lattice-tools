package study

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/experiment"
)

func TestRunAndGenerateReportUseStructuredStudyOutcome(t *testing.T) {
	root := writeProjectFixture(t)
	outcome, err := Run(context.Background(), root, RunOptions{Workers: 2, RepositoryRevision: "test-revision"})
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Summary.Runs) != 2 || len(outcome.Analysis.Runs) != 2 {
		t.Fatalf("runs summary=%d analysis=%d, want 2/2", len(outcome.Summary.Runs), len(outcome.Analysis.Runs))
	}
	if outcome.Bundle.Project.ModelVersion != "study-model-v1" {
		t.Fatalf("model version = %q", outcome.Bundle.Project.ModelVersion)
	}
	if _, err := os.Stat(filepath.Join(root, "reports", "report.md")); err != nil {
		t.Fatalf("run did not generate report: %v", err)
	}

	analysis, err := GenerateReport(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Runs) != 2 {
		t.Fatalf("generated report analysis runs=%d, want 2", len(analysis.Runs))
	}
}

func TestRunForwardsProgress(t *testing.T) {
	for _, state := range []struct {
		name string
		want experiment.ProgressSnapshot
		err  error
	}{
		{"success", experiment.ProgressSnapshot{Total: 2, Completed: 2}, nil},
		{"reused", experiment.ProgressSnapshot{Total: 2, Completed: 2, Reused: 2}, nil},
		{"model failure", experiment.ProgressSnapshot{Total: 2, Completed: 2, Failed: 2}, nil},
		{"cancelled", experiment.ProgressSnapshot{Total: 2, Cancelled: 2}, context.Canceled},
		{"invalid cache", experiment.ProgressSnapshot{Total: 2}, experiment.ErrUnverifiableRun},
	} {
		for _, observer := range []string{"buffered", "nil", "full", "unread"} {
			t.Run(state.name+"/"+observer, func(t *testing.T) {
				root := writeProjectFixture(t)
				if state.name == "model failure" {
					path := filepath.Join(root, "design.yaml")
					data, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte(strings.Replace(string(data), "material: HZO", "material: BTO", 1)), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if state.name == "reused" || state.name == "invalid cache" {
					outcome, err := Run(context.Background(), root, RunOptions{Workers: 2})
					if err != nil {
						t.Fatal(err)
					}
					if state.name == "invalid cache" {
						path := filepath.Join(root, "runs", outcome.Summary.Runs[0].Manifest.RunID, "manifest.json")
						if err := os.Remove(path); err != nil {
							t.Fatal(err)
						}
					}
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if state.name == "cancelled" {
					cancel()
				}
				var progress chan experiment.ProgressSnapshot
				sentinel := experiment.ProgressSnapshot{Total: 99}
				switch observer {
				case "buffered":
					progress = make(chan experiment.ProgressSnapshot, 8)
				case "full":
					progress = make(chan experiment.ProgressSnapshot, 1)
					progress <- sentinel
				case "unread":
					progress = make(chan experiment.ProgressSnapshot)
				}
				outcome, err := Run(ctx, root, RunOptions{Workers: 2, Progress: progress})
				if !errors.Is(err, state.err) {
					t.Fatalf("error = %v, want %v", err, state.err)
				}
				if err == nil && (len(outcome.Summary.Runs) != 2 || len(outcome.Analysis.Runs) != 2) {
					t.Fatalf("outcome runs summary=%d analysis=%d, want 2/2", len(outcome.Summary.Runs), len(outcome.Analysis.Runs))
				}
				if progress != nil {
					close(progress) // The caller still owns the channel after Run.
				}
				if observer == "buffered" {
					var last experiment.ProgressSnapshot
					for snapshot := range progress {
						last = snapshot
					}
					if last != state.want {
						t.Fatalf("last progress = %+v, want %+v", last, state.want)
					}
				} else if observer == "full" {
					if got := <-progress; got != sentinel {
						t.Fatalf("full observer changed: %+v", got)
					}
				}
			})
		}
	}
}

func TestCreateStarterProjectRejectsInvalidSweepBeforeFilesystemChanges(t *testing.T) {
	for _, options := range []StarterSweepOptions{
		{ConductanceLevels: []int{8, 8}, Rows: []int{32}, ADCBits: []int{4}},
		{ConductanceLevels: []int{1}, Rows: []int{32}, ADCBits: []int{4}},
		{ConductanceLevels: []int{8}, Rows: []int{0}, ADCBits: []int{4}},
		{ConductanceLevels: []int{8}, Rows: []int{32}, ADCBits: []int{17}},
	} {
		root := filepath.Join(t.TempDir(), "invalid")
		if err := CreateStarterProject(root, options); err == nil {
			t.Fatal("invalid starter sweep was accepted")
		}
		if _, err := os.Stat(root); !os.IsNotExist(err) {
			t.Fatalf("invalid sweep changed target directory: %v", err)
		}
	}
}

func TestCreateStarterProjectUsesRequestedSweepAndPreservesExistingData(t *testing.T) {
	root := filepath.Join(t.TempDir(), "starter")
	options := StarterSweepOptions{ConductanceLevels: []int{8, 16}, Rows: []int{16, 32}, ADCBits: []int{4, 6}}
	if err := CreateStarterProject(root, options); err != nil {
		t.Fatal(err)
	}
	bundle, points, err := ValidateProject(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if points != 8 || len(bundle.Sweep.Parameters) != 3 {
		t.Fatalf("sweep points=%d parameters=%d, want 8 and 3", points, len(bundle.Sweep.Parameters))
	}
	if bundle.Sweep.Parameters[0].Values[0] != 8 || bundle.Sweep.Parameters[1].Values[0] != 16 || bundle.Sweep.Parameters[2].Values[0] != 4 {
		t.Fatalf("starter sweep values = %+v", bundle.Sweep.Parameters)
	}

	marker := filepath.Join(root, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CreateStarterProject(root, options); err == nil {
		t.Fatal("initialized over a non-empty project")
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "keep" {
		t.Fatalf("existing file changed: got=%q err=%v", got, err)
	}
}

func writeProjectFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"project.yaml": `schema_version: 1
id: hzo-study
name: HZO Study
hypothesis: ADC resolution changes energy and area.
model_version: study-model-v1
objectives:
  - metric: energy_pj
    direction: minimize
`,
		"design.yaml": `schema_version: 1
device:
  material: HZO
  conductance_levels: 30
  g_min_s: 0.000001
  g_max_s: 0.00003
array:
  rows: 32
  cols: 32
  read_voltage_v: 0.2
circuit:
  adc_bits: 6
  dac_bits: 4
  tia_gain_ohm: 10000
  tech_node: 65nm
`,
		"sweep.yaml": `schema_version: 1
seed: 17
max_points: 8
parameters:
  - path: circuit.adc_bits
    values: [4, 6]
`,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
