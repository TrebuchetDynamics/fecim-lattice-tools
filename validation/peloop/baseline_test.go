package peloop_test

import (
	"math"
	"reflect"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/validation/peloop"
)

func TestBaselineUsesUnchangedPreset(t *testing.T) {
	preset := physics.DefaultHZO()
	before := *preset
	ref := []peloop.Sample{{"ascending", 0, -2, -12}, {"ascending", 1, 0, 8}, {"ascending", 2, 3, 23}, {"descending", 0, 3, 24}, {"descending", 1, 0, -7}, {"descending", 2, -2, -11}}
	original := append([]peloop.Sample(nil), ref...)
	got, err := peloop.SimulateBaseline(ref)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := peloop.SimulateBaseline(ref)
	if err != nil || !reflect.DeepEqual(repeated, got) {
		t.Fatalf("nondeterministic baseline: %v", err)
	}
	if !reflect.DeepEqual(ref, original) {
		t.Fatal("simulation changed reference samples")
	}
	for i := range ref {
		ref[i].PolarizationUCcm2 = float64(i) * 1234
	}
	changedP, err := peloop.SimulateBaseline(ref)
	if err != nil || !reflect.DeepEqual(changedP, got) {
		t.Fatalf("measured P affected unfitted baseline: %v", err)
	}
	if !reflect.DeepEqual(got.Material, before) || !reflect.DeepEqual(*preset, before) || !reflect.DeepEqual(*physics.DefaultHZO(), before) {
		t.Fatal("source preset or material snapshot changed")
	}
	wantSettings := peloop.BaselineSettings{
		Preset: "shared/physics.DefaultHZO", Model: "module1-hysteresis/pkg/ferroelectric.NewPreisachModel",
		TemperatureK: 300, StressGPa: 1, Reset: true,
		ConditioningCycles: 2, StepsPerHalfCycle: 512, FieldMinMVcm: -2, FieldMaxMVcm: 3,
		ConditioningPath: "min -> max -> min", SamplingOrder: "input order",
		UpdateMode: "quasi-static Update; not frequency- or wake-up-matched",
		FieldToVm:  1e8, PolarizationToUCcm2: 1e2,
	}
	if got.Settings != wantSettings || len(got.Samples) != len(original) {
		t.Fatalf("unrecorded protocol: %+v", got.Settings)
	}
	// Independent drive through the public unchanged model: 2048 updates form
	// two triangular cycles, each 1024 steps, over the asymmetric field span.
	model := ferroelectric.NewPreisachModel(&before)
	model.Reset()
	model.Update(-2e8)
	for step := 1; step <= 2048; step++ {
		phase := step % 1024
		if phase > 512 {
			phase = 1024 - phase
		}
		model.Update(-2e8 + 5e8*float64(phase)/512)
	}
	for i, sample := range original {
		wantP := model.Update(sample.FieldMVcm*1e8) * 100
		sim := got.Samples[i]
		if sim.Branch != sample.Branch || sim.PointIndex != sample.PointIndex || sim.FieldMVcm != sample.FieldMVcm || math.IsNaN(sim.PolarizationUCcm2) || math.Abs(sim.PolarizationUCcm2-wantP) > 1e-12 {
			t.Fatalf("point %d: %+v, want P=%g", i, sim, wantP)
		}
	}
	for _, bad := range [][]peloop.Sample{nil, {{"ascending", 0, math.NaN(), 1}}, {{"ascending", 0, 1, math.Inf(1)}}, {{"ascending", 0, math.MaxFloat64, 1}}, {{"ascending", 0, -1e300, 1}, {"ascending", 1, 1e300, 2}}} {
		if _, err := peloop.SimulateBaseline(bad); err == nil {
			t.Errorf("accepted invalid baseline input: %v", bad)
		}
	}
}
