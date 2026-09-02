package fecim

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
)

func fixtureDesign() project.Design {
	return project.Design{
		SchemaVersion: 1,
		Device:        project.Device{Material: "HZO", ConductanceLevels: 30, GMinS: 1e-6, GMaxS: 30e-6},
		Array:         project.Array{Rows: 32, Cols: 32, ReadVoltageV: 0.2},
		Circuit:       project.Circuit{ADCBits: 6, DACBits: 4, TIAGainOhm: 10_000, TechNode: "65nm"},
	}
}

func TestEvaluateHZOProducesTraceableMetrics(t *testing.T) {
	got, err := Evaluate(context.Background(), fixtureDesign(), 17)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != experiment.StatusSuccess {
		t.Fatalf("result=%+v", got)
	}
	names := make([]string, 0, len(got.Metrics))
	for _, metric := range got.Metrics {
		names = append(names, metric.Name+":"+metric.Unit)
		if metric.Model == "" || len(metric.Assumptions) == 0 || metric.Evidence == "" {
			t.Fatalf("metric lacks provenance: %+v", metric)
		}
		if math.IsNaN(metric.Value) || math.IsInf(metric.Value, 0) {
			t.Fatalf("metric is non-finite: %+v", metric)
		}
	}
	want := []string{"area_um2:um2", "energy_pj:pJ", "latency_ns:ns", "tia_snr_db:dB"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("got=%v want=%v", names, want)
	}
	if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "simulation") {
		t.Fatalf("warnings=%v", got.Warnings)
	}
	golden := map[string]float64{
		"area_um2":   3844.3264,
		"energy_pj":  51.466103190918965,
		"latency_ns": 27.494033983191414,
		"tia_snr_db": 75.686679471952004,
	}
	for _, metric := range got.Metrics {
		want := golden[metric.Name]
		if math.Abs(metric.Value-want) > 1e-9*math.Max(1, math.Abs(want)) {
			t.Fatalf("%s=%g want %g", metric.Name, metric.Value, want)
		}
	}
}

func TestEvaluateRejectsUnsupportedMaterialAsDesignFailure(t *testing.T) {
	design := fixtureDesign()
	design.Device.Material = "PZT"
	got, err := Evaluate(context.Background(), design, 17)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != experiment.StatusFailed || got.Failure == nil || got.Failure.Kind != "unsupported-material" {
		t.Fatalf("result=%+v", got)
	}
}

func TestEvaluateHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Evaluate(ctx, fixtureDesign(), 17)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want context.Canceled", err)
	}
}
