package peloop_test

import (
	"math"
	"testing"

	"fecim-lattice-tools/validation/peloop"
)

// Retain the overflow-sensitive regression coverage from e60a3e4f alongside
// the stronger branch validation and asymmetric-drive tests in local main.
func TestCompareFiniteHugeResidualsRemainFiniteOrFail(t *testing.T) {
	reference, simulated := make([]peloop.Sample, 28), make([]peloop.Sample, 28)
	for i := range reference {
		reference[i] = peloop.Sample{"ascending", i, float64(i), 0}
		simulated[i] = peloop.Sample{"ascending", i, float64(i), math.MaxFloat64}
	}
	metrics, err := peloop.Compare(reference, simulated)
	if err == nil && (math.IsInf(metrics.Pooled.RMSEUCcm2, 0) || math.IsNaN(metrics.Pooled.RMSEUCcm2)) {
		t.Fatalf("successful comparison returned nonfinite accumulated errors: %+v", metrics.Pooled)
	}
}

func TestBaselineFiniteWideSpanRetainsNegativeRemanence(t *testing.T) {
	wide, err := peloop.SimulateBaseline([]peloop.Sample{{"ascending", 0, 0, 0}, {"ascending", 1, 1e298, 0}, {"descending", 0, -1e298, 0}})
	if err != nil {
		t.Fatal(err)
	}
	if !(wide.Samples[0].PolarizationUCcm2 < 0) {
		t.Fatalf("conditioning did not reach finite minimum: %+v", wide.Samples)
	}
}

func TestBaselineRetainsLegacyFivePointRegression(t *testing.T) {
	reference := []peloop.Sample{{"ascending", 0, -2, -9}, {"ascending", 1, -.5, 3}, {"ascending", 2, 2, 8}, {"descending", 0, 1, 7}, {"descending", 1, -1.5, -1}}
	got, err := peloop.SimulateBaseline(reference)
	if err != nil {
		t.Fatal(err)
	}
	// Existing numerical regression values from e60a3e4f, not measurements and
	// not newly fitted or regenerated for this merge.
	want := []float64{-25.465138481825001, -24.693257261526774, 25.044156017912311, 23.984560054976079, -15.375794436140053}
	if len(got.Samples) != len(want) {
		t.Fatalf("sample count = %d", len(got.Samples))
	}
	for i, sample := range got.Samples {
		if math.Abs(sample.PolarizationUCcm2-want[i]) > 1e-10 {
			t.Errorf("sample %d: P=%.17g, want %.17g", i, sample.PolarizationUCcm2, want[i])
		}
	}
}
