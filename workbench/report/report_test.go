package report

import (
	"bytes"
	"errors"
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
