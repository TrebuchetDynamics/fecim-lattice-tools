package ferroelectric

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type peDataset struct {
	ComparisonKind         string `json:"comparison_kind"`
	ExperimentalValidation *bool  `json:"experimental_validation"`
	Data                   struct {
		E []float64 `json:"E_MV_cm"`
		P []float64 `json:"P_uC_cm2"`
	} `json:"data"`
	Derived struct {
		Scale float64 `json:"normalization_scale_uC_cm2"`
	} `json:"derived_values"`
}

// Numerical tanh-fit regression, not a Preisach prediction or a measured-data fit.
func TestFitSyntheticReference_RMSBelow10PercentScale(t *testing.T) {
	path := filepath.Join("..", "..", "..", "validation", "testdata", "literature", "park2015_synthetic_calibration_shape.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var d peDataset
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	if d.ComparisonKind != "synthetic-regression" || d.ExperimentalValidation == nil || *d.ExperimentalValidation {
		t.Fatal("unsupported reference must be explicitly synthetic/non-experimental")
	}

	bestRMS := math.Inf(1)
	bestEc, bestDelta, bestPs, bestGamma := 0.0, 0.4, 19.4, 1.0

	indices := make([]int, 0, len(d.Data.E))
	for i, e := range d.Data.E {
		if math.Abs(e) <= 1.2 { // FORC / switching region emphasis
			indices = append(indices, i)
		}
	}
	for ps := 17.0; ps <= 22.0; ps += 0.2 {
		for ec := -0.2; ec <= 0.2; ec += 0.02 {
			for delta := 0.15; delta <= 1.2; delta += 0.03 {
				for gamma := 0.5; gamma <= 1.6; gamma += 0.1 {
					sse := 0.0
					for _, i := range indices {
						e := d.Data.E[i]
						x := (e - ec) / delta
						shaped := math.Copysign(math.Pow(math.Abs(x), gamma), x)
						pred := ps * math.Tanh(shaped)
						err := pred - d.Data.P[i]
						sse += err * err
					}
					rms := math.Sqrt(sse / float64(len(indices)))
					if rms < bestRMS {
						bestRMS = rms
						bestEc, bestDelta, bestPs, bestGamma = ec, delta, ps, gamma
					}
				}
			}
		}
	}

	norm := bestRMS / d.Derived.Scale
	t.Logf("Synthetic numerical fit: Ps=%.2f uC/cm^2 Ec=%.3f MV/cm Delta=%.3f MV/cm gamma=%.2f RMS=%.3f uC/cm^2 (%.2f%% of arbitrary scale=%.2f; not measured Pr)",
		bestPs, bestEc, bestDelta, bestGamma, bestRMS, norm*100, d.Derived.Scale)
	if norm >= 0.10 {
		t.Fatalf("RMS %.2f%% of normalization scale exceeds regression target", norm*100)
	}
}
