package peloop_test

import (
	"math"
	"reflect"
	"testing"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/validation/peloop"
)

func TestBaselineUsesUnchangedPreset(t *testing.T) {
	ref := []peloop.Sample{{"ascending", 0, -2, -9}, {"ascending", 1, -.5, 3}, {"ascending", 2, 2, 8}, {"descending", 0, 1, 7}, {"descending", 1, -1.5, -1}}
	before := append([]peloop.Sample(nil), ref...)
	preset := *physics.DefaultHZO()
	got, err := peloop.SimulateBaseline(ref)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := peloop.SimulateBaseline(ref)
	if err != nil || !reflect.DeepEqual(got, repeated) {
		t.Fatalf("non-deterministic baseline: %v", err)
	}
	changedP := append([]peloop.Sample(nil), ref...)
	for i := range changedP {
		changedP[i].PolarizationUCcm2 = float64(i) * 1e6
	}
	withoutMeasuredP, err := peloop.SimulateBaseline(changedP)
	if err != nil || !reflect.DeepEqual(got, withoutMeasuredP) {
		t.Fatalf("baseline depends on measured P: %v", err)
	}
	if !reflect.DeepEqual(ref, before) || !reflect.DeepEqual(preset, *physics.DefaultHZO()) || !reflect.DeepEqual(got.Material, preset) {
		t.Fatal("baseline changed reference or material preset, or omitted snapshot fields")
	}
	wantSettings := peloop.BaselineSettings{
		Preset: "shared/physics.DefaultHZO", Model: "module1-hysteresis/pkg/ferroelectric.NewPreisachModel",
		TemperatureK: 300, StressGPa: 1, Reset: true,
		ConditioningCycles: 2, StepsPerHalfCycle: 512, FieldMinMVcm: -2, FieldMaxMVcm: 2,
		ConditioningPath: "min -> max -> min", SamplingOrder: "input order",
		UpdateMode: "quasi-static Update; not frequency- or wake-up-matched",
		FieldToVm:  1e8, PolarizationToUCcm2: 1e2,
	}
	if got.Settings != wantSettings {
		t.Fatalf("settings = %+v, want %+v", got.Settings, wantSettings)
	}
	// Fixed trajectory evaluated with the unchanged public Preisach engine:
	// Reset, min, 2*(512 ascending + 512 descending updates), then these fields.
	wantP := []float64{-25.465138481825001, -24.693257261526774, 25.044156017912311, 23.984560054976079, -15.375794436140053}
	if len(got.Samples) != len(ref) {
		t.Fatalf("sample count = %d", len(got.Samples))
	}
	for i, sample := range got.Samples {
		if sample.Branch != ref[i].Branch || sample.PointIndex != ref[i].PointIndex || sample.FieldMVcm != ref[i].FieldMVcm || math.Abs(sample.PolarizationUCcm2-wantP[i]) > 1e-10 {
			t.Errorf("sample %d = %+v, want unchanged field/identity and P %.17g", i, sample, wantP[i])
		}
	}
	// A finite but very wide span must still finish conditioning at its minimum.
	// Sampling zero first exposes the remanent sign if intermediate fields overflow
	// and are silently ignored by the unchanged model's nonfinite-input guard.
	wide, err := peloop.SimulateBaseline([]peloop.Sample{{"ascending", 0, 0, 0}, {"ascending", 1, 1e298, 0}, {"descending", 0, -1e298, 0}})
	if err != nil || !(wide.Samples[0].PolarizationUCcm2 < 0) {
		t.Fatalf("conditioning did not reach finite minimum: %+v, %v", wide.Samples, err)
	}
	for _, invalid := range [][]peloop.Sample{nil, {{"ascending", 0, math.NaN(), 1}}, {{"ascending", 0, math.MaxFloat64, 1}}} {
		if _, err := peloop.SimulateBaseline(invalid); err == nil {
			t.Fatal("accepted invalid baseline fields")
		}
	}
}
