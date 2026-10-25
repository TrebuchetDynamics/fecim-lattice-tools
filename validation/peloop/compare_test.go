package peloop_test

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/validation/peloop"
)

func TestCompareMatchedBranches(t *testing.T) {
	ref := []peloop.Sample{{"ascending", 0, -1, -4}, {"ascending", 1, 1, 2}, {"descending", 0, 1, 8}}
	sim := []peloop.Sample{{"ascending", 0, -1, -1}, {"ascending", 1, 1, -2}, {"descending", 0, 1, 8}}
	before := append([]peloop.Sample(nil), ref...)
	got, err := peloop.Compare(ref, sim)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]struct {
		count                 int
		rmse, mae, normalizer float64
	}{
		"ascending":  {2, math.Sqrt(12.5), 3.5, 4},
		"descending": {1, 0, 0, 8},
		"pooled":     {3, math.Sqrt(25.0 / 3), 7.0 / 3, 8},
	} {
		m := got.PerBranch[name]
		if name == "pooled" {
			m = got.Pooled
		}
		if m.Count != want.count || math.Abs(m.RMSEUCcm2-want.rmse) > 1e-14 || math.Abs(m.MAEUCcm2-want.mae) > 1e-14 || m.ObservedMaxAbsReferencePUCcm2 != want.normalizer || m.Normalizer != "observed max(abs(reference P))" || m.NormalizedRMSE == nil || math.Abs(*m.NormalizedRMSE-want.rmse/want.normalizer) > 1e-14 {
			t.Errorf("%s metrics = %+v, want %+v", name, m, want)
		}
	}
	if !reflect.DeepEqual(ref, before) {
		t.Fatal("comparison mutated reference")
	}
	zero, err := peloop.Compare([]peloop.Sample{{"ascending", 0, 0, 0}}, []peloop.Sample{{"ascending", 0, 0, 2}})
	if err != nil || zero.Pooled.NormalizedRMSE != nil || zero.PerBranch["ascending"].NormalizedRMSE != nil || zero.Pooled.RMSEUCcm2 != 2 {
		t.Fatalf("zero observed normalization = %+v, %v", zero, err)
	}
	// Finite input must never produce successful nonfinite metrics, including
	// when observed P is zero and normalized RMSE is unavailable.
	hugeRef, hugeSim := make([]peloop.Sample, 28), make([]peloop.Sample, 28)
	for i := range hugeRef {
		hugeRef[i] = peloop.Sample{"ascending", i, float64(i), 0}
		hugeSim[i] = peloop.Sample{"ascending", i, float64(i), math.MaxFloat64}
	}
	if m, err := peloop.Compare(hugeRef, hugeSim); err == nil && (math.IsInf(m.Pooled.RMSEUCcm2, 0) || math.IsNaN(m.Pooled.RMSEUCcm2)) {
		t.Fatalf("successful comparison returned nonfinite accumulated errors: %+v", m.Pooled)
	}
	for _, tc := range []struct {
		name     string
		ref, sim []peloop.Sample
	}{
		{"empty", nil, nil},
		{"length", ref, sim[:2]},
		{"branch", ref[:1], []peloop.Sample{{"descending", 0, -1, -1}}},
		{"index", ref[:1], []peloop.Sample{{"ascending", 1, -1, -1}}},
		{"field", ref[:1], []peloop.Sample{{"ascending", 0, math.Nextafter(-1, 0), -1}}},
		{"nonfinite reference", []peloop.Sample{{"ascending", 0, -1, math.NaN()}}, sim[:1]},
		{"nonfinite simulation", ref[:1], []peloop.Sample{{"ascending", 0, -1, math.Inf(1)}}},
		{"reordered branches", ref, []peloop.Sample{sim[2], sim[0], sim[1]}},
		{"duplicate index", []peloop.Sample{{"ascending", 0, -1, 0}, {"ascending", 0, 1, 0}}, []peloop.Sample{{"ascending", 0, -1, 0}, {"ascending", 0, 1, 0}}},
		{"duplicate field", []peloop.Sample{{"ascending", 0, -1, 0}, {"ascending", 1, -1, 0}}, []peloop.Sample{{"ascending", 0, -1, 0}, {"ascending", 1, -1, 0}}},
		{"nonfinite field", []peloop.Sample{{"ascending", 0, math.Inf(1), 0}}, []peloop.Sample{{"ascending", 0, math.Inf(1), 0}}},
		{"unknown branch", []peloop.Sample{{"other", 0, 0, 0}}, []peloop.Sample{{"other", 0, 0, 0}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := peloop.Compare(tc.ref, tc.sim); err == nil {
				t.Fatal("accepted invalid or unmatched samples")
			}
		})
	}
}

func TestReadSamples(t *testing.T) {
	const header = "branch,point_index,field_MV_cm,polarization_uC_cm2\n"
	got, err := peloop.ReadSamples(strings.NewReader(header + "ascending,0,-1,-4\nascending,1,1,2\ndescending,0,1,8\n"))
	want := []peloop.Sample{{"ascending", 0, -1, -4}, {"ascending", 1, 1, 2}, {"descending", 0, 1, 8}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("samples = %v, %v", got, err)
	}
	for _, input := range []string{
		"", header, "branch,point_index,x_px,y_px\nascending,0,1,2\n",
		header + "ascending,0,NaN,1\n", header + "ascending,0,1,Inf\n",
		header + "ascending,0,1,no\n", header + "ascending,0,1\n",
		header + "ascending,-1,1,2\n", header + "ascending,1,1,2\n",
		header + "other,0,1,2\n", header + "ascending,0,1,2\nascending,0,2,3\n",
		header + "ascending,0,1,2\nascending,1,1,3\n",
		header + "ascending,0,0." + strings.Repeat("0", 2*1024*1024) + ",0\n",
	} {
		if _, err := peloop.ReadSamples(strings.NewReader(input)); err == nil {
			t.Errorf("accepted invalid CSV (length %d)", len(input))
		}
	}
	var bounded strings.Builder
	bounded.WriteString(header)
	for i := 0; i < 10000; i++ {
		fmt.Fprintf(&bounded, "ascending,%d,%d,0\n", i, i)
	}
	if samples, err := peloop.ReadSamples(strings.NewReader(bounded.String())); err != nil || len(samples) != 10000 {
		t.Fatalf("10000-row boundary: count %d, %v", len(samples), err)
	}
	bounded.WriteString("ascending,10000,10000,0\n")
	if _, err := peloop.ReadSamples(strings.NewReader(bounded.String())); err == nil {
		t.Fatal("accepted 10001 rows")
	}
}
