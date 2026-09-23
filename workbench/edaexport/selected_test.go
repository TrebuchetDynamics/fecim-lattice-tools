package edaexport

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
)

func testRun(t *testing.T) (project.Bundle, string) {
	t.Helper()
	bundle := project.Bundle{
		Root: t.TempDir(),
		Project: project.Project{
			SchemaVersion: 1,
			ID:            "eda-export-test",
			Hypothesis:    "Export preserves the selected simulated design.",
			ModelVersion:  "test-v1",
			Objectives:    []project.Objective{{Metric: "energy_pj", Direction: project.Minimize}, {Metric: "latency_ns", Direction: project.Minimize}},
		},
		Design: project.Design{
			SchemaVersion: 1,
			Device:        project.Device{Material: "HZO", ConductanceLevels: 30, GMinS: 1e-6, GMaxS: 30e-6},
			Array:         project.Array{Rows: 4, Cols: 4, ReadVoltageV: 0.2},
			Circuit:       project.Circuit{ADCBits: 6, DACBits: 4, TIAGainOhm: 10000, TechNode: "65nm"},
		},
		Sweep: project.Sweep{Seed: 17, MaxPoints: 1, Parameters: []project.Parameter{{Path: "circuit.adc_bits", Values: []float64{6}}}},
	}
	summary, err := experiment.Run(context.Background(), bundle, experiment.RunOptions{
		EvaluatorVersion: "test-v1",
		Workers:          1,
		Now:              func() time.Time { return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC) },
		Evaluator: func(context.Context, project.Design, int64) (experiment.Result, error) {
			return experiment.Result{Status: experiment.StatusSuccess, Metrics: []experiment.Metric{
				{Name: "energy_pj", Value: 1, Unit: "pJ", Model: "fixture", Assumptions: []string{"fixture"}, Evidence: experiment.EvidenceDerived},
				{Name: "latency_ns", Value: 1, Unit: "ns", Model: "fixture", Assumptions: []string{"fixture"}, Evidence: experiment.EvidenceDerived},
			}}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return bundle, summary.Runs[0].Manifest.RunID
}

func TestExportSelectedRunUsesVerifiedRunAndWritesAllFormats(t *testing.T) {
	bundle, runID := testRun(t)
	dir, err := ExportSelectedRun(bundle.Root, runID)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"design.spice", "design.v", "design.def", "cell.lef", "cell.lib", "manifest.json"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.Size() == 0 {
			t.Fatalf("artifact %s size=%v err=%v", name, info, err)
		}
	}
	var manifest struct {
		RunID string            `json:"run_id"`
		Files map[string]string `json:"files_sha256"`
	}
	blob, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(blob, &manifest); err != nil || manifest.RunID != runID || len(manifest.Files) != 5 {
		t.Fatalf("manifest=%+v err=%v", manifest, err)
	}
	if _, err := ExportSelectedRun(bundle.Root, runID); err == nil {
		t.Fatal("repeat export unexpectedly overwrote existing artifacts")
	}
}

func TestExportSelectedRunRejectsTamperedRun(t *testing.T) {
	bundle, runID := testRun(t)
	path := filepath.Join(bundle.Root, "runs", runID, "result.json")
	if err := os.WriteFile(path, []byte(`{"status":"success"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ExportSelectedRun(bundle.Root, runID); err == nil {
		t.Fatal("tampered run unexpectedly exported")
	}
	if _, err := os.Stat(filepath.Join(bundle.Root, "exports")); !os.IsNotExist(err) {
		t.Fatalf("failed export created output directory: %v", err)
	}
}
