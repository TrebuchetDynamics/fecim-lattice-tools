package report

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
)

func reportFixture() (project.Bundle, experiment.Analysis) {
	bundle := project.Bundle{
		Project: project.Project{
			ID:           "hzo-study",
			Name:         "HZO study",
			Hypothesis:   "ADC resolution trades area for energy fidelity.",
			ModelVersion: "study-model-v1",
			Objectives:   []project.Objective{{Metric: "energy_pj", Direction: project.Minimize}},
			Constraints:  []project.Constraint{{Metric: "area_um2", Operator: "<=", Value: 5000, Unit: "um2"}},
			Citations:    []string{"park2015_advmat_hzo"},
		},
		Inputs: []project.ResolvedInput{{Path: "inputs/loop.csv", SHA256: strings.Repeat("b", 64), Citation: "park2015_advmat_hzo", Evidence: "literature-backed"}},
	}
	run := experiment.RunRecord{
		Manifest: experiment.RunManifest{RunID: strings.Repeat("a", 64), ModelVersion: "study-model-v1", EvaluatorVersion: "evaluator-v1", RepositoryRevision: "abc123", Inputs: []project.ResolvedInput{{Path: "inputs/loop.csv", SHA256: strings.Repeat("b", 64), Citation: "park2015_advmat_hzo", Evidence: "literature-backed"}}, Status: experiment.StatusSuccess, StartedAt: time.Now(), CompletedAt: time.Now()},
		Design:   project.Design{SchemaVersion: 1, Device: project.Device{Material: "HZO"}},
		Result: experiment.Result{Status: experiment.StatusSuccess, Metrics: []experiment.Metric{
			{Name: "area_um2", Value: 100, Unit: "um2", Model: "area", Assumptions: []string{"default"}, Evidence: experiment.EvidenceDefault},
			{Name: "energy_pj", Value: 10, Unit: "pJ", Model: "energy", Assumptions: []string{"derived"}, Evidence: experiment.EvidenceDerived},
		}, Warnings: []string{"simulation only"}},
		Reused: true,
	}
	analysis := experiment.Analysis{
		Runs:   []experiment.AnalyzedRun{{Run: run, Feasible: true, Pareto: true, Constraints: []experiment.ConstraintOutcome{{Constraint: bundle.Project.Constraints[0], Passed: true, Actual: 100}}}},
		Counts: map[string]int{"success": 1, "failed": 0, "feasible": 1, "infeasible": 0, "unusable": 0, "pareto": 1},
	}
	return bundle, analysis
}

func TestReportsAreDeterministicAndExcludeRuntimeMetadata(t *testing.T) {
	bundle, analysis := reportFixture()
	var first, second bytes.Buffer
	if err := WriteJSON(&first, bundle, analysis); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(&second, bundle, analysis); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() {
		t.Fatal("JSON report changed between writes")
	}
	jsonText := first.String()
	for _, forbidden := range []string{"started_at", "completed_at", "reused"} {
		if strings.Contains(jsonText, forbidden) {
			t.Fatalf("JSON contains runtime field %q", forbidden)
		}
	}
	for _, required := range []string{"run_id", "area_um2", "simulation-default", "pareto", "model_version", "study-model-v1", "evaluator_version", "evaluator-v1", "repository_revision", "abc123", "inputs", "inputs/loop.csv", strings.Repeat("b", 64), "citations", "park2015_advmat_hzo"} {
		if !strings.Contains(jsonText, required) {
			t.Fatalf("JSON missing %q", required)
		}
	}

	first.Reset()
	if err := WriteCSV(&first, analysis); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(first.String(), "run_id,status,feasible,pareto,unusable_reason,area_um2,energy_pj,warnings\n") {
		t.Fatalf("CSV header=%q", strings.SplitN(first.String(), "\n", 2)[0])
	}

	first.Reset()
	if err := WriteMarkdown(&first, bundle, analysis); err != nil {
		t.Fatal(err)
	}
	markdown := first.String()
	for _, required := range []string{"Simplified simulation-only pre-silicon estimates", bundle.Project.Hypothesis, "Pareto", "simulation-default", "simulation only", "model: energy", "assumptions: derived", "study-model-v1", "evaluator-v1", "abc123", "inputs/loop.csv", strings.Repeat("b", 64), "park2015_advmat_hzo"} {
		if !strings.Contains(markdown, required) {
			t.Fatalf("Markdown missing %q", required)
		}
	}
}

func TestWriteMarkdownStopsAtFirstWriterError(t *testing.T) {
	bundle, analysis := reportFixture()
	writer := &failingWriter{failAt: 2}
	if err := WriteMarkdown(writer, bundle, analysis); err == nil {
		t.Fatal("WriteMarkdown ignored writer error")
	}
	if writer.calls != 2 {
		t.Fatalf("WriteMarkdown made %d writes after writer failed on write 2", writer.calls)
	}
}

type failingWriter struct {
	calls  int
	failAt int
}

func (w *failingWriter) Write(data []byte) (int, error) {
	w.calls++
	if w.calls >= w.failAt {
		return 0, errors.New("writer failed")
	}
	return len(data), nil
}

// This catches path-based authority, unpinned report directories, premature
// replacement, and cleanup that deletes a different writer's scratch file.
func TestGenerateRootConfinement(t *testing.T) {
	for _, scenario := range []string{"bytes", "outside reports link", "outside output link", "retarget", "sync failure", "retarget failure", "substituted scratch", "substituted scratch failure"} {
		t.Run(scenario, func(t *testing.T) {
			bundle, analysis := reportFixture()
			dir, outside := t.TempDir(), t.TempDir()
			bundle.Root = outside // Descriptive metadata must not grant authority.
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			put := func(name, body string) {
				t.Helper()
				if err := os.WriteFile(name, []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			read := func(name string) string {
				t.Helper()
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				return string(data)
			}
			names := []string{"results.json", "results.csv", "report.md"}
			for _, name := range names {
				put(filepath.Join(outside, name), "outside marker")
			}
			reports := filepath.Join(dir, "reports")
			if scenario == "outside reports link" {
				if err := os.Symlink(outside, reports); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(reports, 0o755); err != nil {
					t.Fatal(err)
				}
				for _, name := range names {
					put(filepath.Join(reports, name), "previous report")
				}
				put(filepath.Join(reports, ".tmp-unrelated"), "keep")
			}
			if scenario == "bytes" {
				// The handle remains authoritative after its original path is replaced.
				moved := filepath.Join(t.TempDir(), "pinned")
				if err := os.Rename(dir, moved); err != nil {
					t.Fatal(err)
				}
				reports = filepath.Join(moved, "reports")
				if err := os.Symlink(outside, dir); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "outside output link" {
				if err := root.Remove("reports/results.json"); err != nil {
					t.Fatal(err)
				}
				if err := root.Symlink(filepath.Join(outside, "results.json"), "reports/results.json"); err != nil {
					t.Fatal(err)
				}
			}

			fault := errors.New("precommit sync failure")
			originalSync := syncReportFile
			t.Cleanup(func() { syncReportFile = originalSync })
			substituted := ""
			first := true
			syncReportFile = func(f *os.File) error {
				if !first {
					return f.Sync()
				}
				first = false
				if scenario == "retarget" || scenario == "retarget failure" {
					if err := root.Rename("reports", "pinned-reports"); err != nil {
						t.Fatal(err)
					}
					reports = filepath.Join(dir, "pinned-reports")
					if err := root.Symlink(outside, "reports"); err != nil {
						t.Fatal(err)
					}
				}
				if strings.HasPrefix(scenario, "substituted scratch") {
					substituted = filepath.Base(f.Name())
					if err := os.Remove(filepath.Join(reports, substituted)); err != nil {
						t.Fatal(err)
					}
					put(filepath.Join(reports, substituted), "other writer")
				}
				if strings.Contains(scenario, "failure") {
					return fault
				}
				return f.Sync()
			}
			err = GenerateRoot(root, bundle, analysis)
			if scenario == "outside reports link" {
				if err == nil {
					t.Fatal("accepted escaping reports link")
				}
			} else if strings.Contains(scenario, "failure") {
				if !errors.Is(err, fault) {
					t.Fatalf("error=%v, want injected precommit failure", err)
				}
			} else if scenario == "substituted scratch" {
				if err == nil {
					t.Fatal("committed another writer's substituted scratch")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if _, err := root.Stat("."); err != nil {
				t.Fatalf("caller root closed: %v", err)
			}
			for _, name := range names {
				if got := read(filepath.Join(outside, name)); got != "outside marker" {
					t.Fatalf("outside %s changed: %s", name, got)
				}
			}
			entries, err := os.ReadDir(outside)
			if err != nil || len(entries) != len(names) {
				t.Fatalf("outside namespace changed: %v %v", entries, err)
			}
			if scenario == "outside reports link" {
				return
			}
			if strings.Contains(scenario, "failure") || scenario == "substituted scratch" {
				for _, name := range names {
					if got := read(filepath.Join(reports, name)); got != "previous report" {
						t.Fatalf("precommit failure replaced %s: %s", name, got)
					}
				}
			} else {
				legacy := filepath.Join(t.TempDir(), "new", "study")
				if err := Generate(legacy, bundle, analysis); err != nil {
					t.Fatal(err)
				}
				for i, write := range []func(io.Writer) error{
					func(w io.Writer) error { return WriteJSON(w, bundle, analysis) },
					func(w io.Writer) error { return WriteCSV(w, analysis) },
					func(w io.Writer) error { return WriteMarkdown(w, bundle, analysis) },
				} {
					var want bytes.Buffer
					if err := write(&want); err != nil {
						t.Fatal(err)
					}
					for _, path := range []string{filepath.Join(reports, names[i]), filepath.Join(legacy, "reports", names[i])} {
						if got := read(path); got != want.String() {
							t.Fatalf("%s differs from serializer bytes", path)
						}
					}
				}
			}
			entries, err = os.ReadDir(reports)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".tmp-") && entry.Name() != ".tmp-unrelated" && entry.Name() != substituted {
					t.Fatalf("owned scratch leaked: %s", entry.Name())
				}
			}
			if got := read(filepath.Join(reports, ".tmp-unrelated")); got != "keep" {
				t.Fatal("unrelated scratch changed")
			}
			if substituted != "" && read(filepath.Join(reports, substituted)) != "other writer" {
				t.Fatal("substituted scratch removed")
			}
		})
	}
}

func TestGenerateWritesThreeReports(t *testing.T) {
	bundle, analysis := reportFixture()
	bundle.Root = t.TempDir()
	if err := Generate(bundle.Root, bundle, analysis); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"results.json", "results.csv", "report.md"} {
		path := filepath.Join(bundle.Root, "reports", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			t.Fatalf("%s is empty", path)
		}
	}
	entries, err := os.ReadDir(filepath.Join(bundle.Root, "reports"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Fatalf("temporary report remained: %s", entry.Name())
		}
	}
}
