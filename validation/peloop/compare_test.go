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
	ref := []peloop.Sample{{"ascending", 0, -1, -4}, {"ascending", 1, 1, 8}, {"descending", 0, 1, 2}}
	sim := []peloop.Sample{{"ascending", 0, -1, -7}, {"ascending", 1, 1, 12}, {"descending", 0, 1, -4}}
	before := append([]peloop.Sample(nil), ref...)
	got, err := peloop.Compare(ref, sim)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name                  string
		got                   peloop.ErrorMetrics
		count                 int
		rmse, mae, normalizer float64
	}{
		{"ascending", got.PerBranch["ascending"], 2, math.Sqrt(12.5), 3.5, 8},
		{"descending", got.PerBranch["descending"], 1, 6, 6, 2},
		{"pooled", got.Pooled, 3, math.Sqrt(61.0 / 3), 13.0 / 3, 8},
	} {
		m := tc.got
		if m.Count != tc.count || math.Abs(m.RMSEUCcm2-tc.rmse) > 1e-12 || math.Abs(m.MAEUCcm2-tc.mae) > 1e-12 || m.ObservedMaxAbsReferencePUCcm2 != tc.normalizer || m.Normalizer != "observed max(abs(reference P))" || m.NormalizedRMSE == nil || math.Abs(*m.NormalizedRMSE-tc.rmse/tc.normalizer) > 1e-12 {
			t.Errorf("%s: %+v", tc.name, m)
		}
	}
	if len(got.PerBranch) != 2 || !reflect.DeepEqual(ref, before) {
		t.Fatal("comparison changed reference or branch inventory")
	}
	zero, err := peloop.Compare([]peloop.Sample{{"ascending", 0, 0, 0}}, []peloop.Sample{{"ascending", 0, 0, 3}})
	if err != nil || zero.Pooled.RMSEUCcm2 != 3 || zero.Pooled.NormalizedRMSE != nil || zero.PerBranch["ascending"].NormalizedRMSE != nil {
		t.Fatalf("zero observed normalization must be unavailable: %+v, %v", zero, err)
	}
	for name, bad := range map[string][]peloop.Sample{
		"empty":        nil,
		"count":        sim[:2],
		"branch":       {{"descending", 0, -1, -7}, sim[1], sim[2]},
		"branch order": {sim[2], sim[0], sim[1]},
		"index":        {{"ascending", 1, -1, -7}, sim[1], sim[2]},
		"field":        {{"ascending", 0, math.Nextafter(-1, 0), -7}, sim[1], sim[2]},
		"duplicate":    {sim[0], sim[0], sim[2]},
		"nan P":        {{"ascending", 0, -1, math.NaN()}, sim[1], sim[2]},
		"infinite E":   {{"ascending", 0, math.Inf(1), -7}, sim[1], sim[2]},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := peloop.Compare(ref, bad); err == nil {
				t.Fatal("accepted invalid simulation")
			}
			if _, err := peloop.Compare(bad, ref); err == nil {
				t.Fatal("accepted invalid reference")
			}
		})
	}
}

func TestReadSamples(t *testing.T) {
	const header = "branch,point_index,field_MV_cm,polarization_uC_cm2\n"
	want := []peloop.Sample{{"ascending", 0, -1, -4}, {"ascending", 1, 1, 8}, {"descending", 0, 1, 2}}
	got, err := peloop.ReadSamples(strings.NewReader(header + "ascending,0,-1,-4\nascending,1,1,8\ndescending,0,1,2\n"))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("physical samples: %v, %v", got, err)
	}
	for _, bad := range []string{
		"", header, "branch,point_index,x_px,y_px\nascending,0,1,2\n",
		header + "ascending,0,NaN,1\n", header + "ascending,0,1,+Inf\n",
		header + "ascending,no,1,2\n", header + "ascending,0,no,2\n", header + "ascending,0,1,no\n",
		header + "other,0,1,2\n", header + "ascending,-1,1,2\n",
		header + "ascending,0,1,2\nascending,2,2,3\n",
		header + "ascending,0,1,2\nascending,1,1,3\n",
		header + "descending,0,1,2\ndescending,1,2,3\n",
		header + "ascending,0,1\n", header + "\"unterminated",
		header + "ascending,0,0,0." + strings.Repeat("0", 2*1024*1024),
	} {
		if samples, err := peloop.ReadSamples(strings.NewReader(bad)); err == nil || samples != nil {
			t.Errorf("accepted malformed CSV %.100q: %v", bad, err)
		}
	}
	var rows strings.Builder
	rows.WriteString(header)
	for i := 0; i < 10000; i++ {
		fmt.Fprintf(&rows, "ascending,%d,%d,0\n", i, i)
	}
	if samples, err := peloop.ReadSamples(strings.NewReader(rows.String())); err != nil || len(samples) != 10000 {
		t.Fatalf("rejected 10000 valid rows: count=%d, %v", len(samples), err)
	}
	rows.WriteString("ascending,10000,10000,0\n")
	if _, err := peloop.ReadSamples(strings.NewReader(rows.String())); err == nil {
		t.Fatal("accepted 10001 valid rows")
	}
	if _, err := peloop.ReadSamples(nil); err == nil {
		t.Fatal("accepted nil reader")
	}
}
