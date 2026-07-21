package report

import (
	"bytes"
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
			ID:          "hzo-study",
			Name:        "HZO study",
			Hypothesis:  "ADC resolution trades area for energy fidelity.",
			Objectives:  []project.Objective{{Metric: "energy_pj", Direction: project.Minimize}},
			Constraints: []project.Constraint{{Metric: "area_um2", Operator: "<=", Value: 5000, Unit: "um2"}},
		},
	}
	run := experiment.RunRecord{
		Manifest: experiment.RunManifest{RunID: strings.Repeat("a", 64), Status: experiment.StatusSuccess, StartedAt: time.Now(), CompletedAt: time.Now()},
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
	if err := WriteJSON(&first, analysis); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(&second, analysis); err != nil {
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
	for _, required := range []string{"run_id", "area_um2", "simulation-default", "pareto"} {
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
	for _, required := range []string{"literature-calibrated pre-silicon", bundle.Project.Hypothesis, "Pareto", "simulation-default", "simulation only"} {
		if !strings.Contains(markdown, required) {
			t.Fatalf("Markdown missing %q", required)
		}
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
