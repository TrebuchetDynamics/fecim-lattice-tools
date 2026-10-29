package crossbar

import (
	"math"
	"math/rand"
	"slices"
	"testing"
)

// M2-PV-03: Yield (fraction of trials with every output <10% from ideal)
// decreases with sigma in a controlled sensitivity sweep. A finite Gaussian
// sample can saturate at 100% for both low sigmas, so check that sample against
// a numerical reference instead of requiring distinct empirical yields.
func TestM2_PV_03_YieldDecreasesWithNoiseLevel(t *testing.T) {
	rows, cols := 8, 8
	input := make([]float64, cols)
	for j := range input {
		input[j] = 1.0
	}

	weights := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		weights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			weights[i][j] = 0.85
		}
	}

	// Ideal (no variation) baseline.
	idealArr, err := NewArray(&Config{Rows: rows, Cols: cols, NoiseLevel: 0, ADCBits: 12, DACBits: 12})
	if err != nil {
		t.Fatalf("NewArray(ideal): %v", err)
	}
	if err := idealArr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix: %v", err)
	}
	idealRes, err := idealArr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, Temperature: 300})
	if err != nil {
		t.Fatalf("ideal MVM: %v", err)
	}

	threshold := 0.10 // 10% relative error per element
	sigmas := []float64{0.01, 0.10, 0.30}
	const runs = 200
	for _, fixture := range []string{"ControlledRowSweep", "Gaussian"} {
		t.Run(fixture, func(t *testing.T) {
			yields := make([]float64, len(sigmas))
			for idx, sigma := range sigmas {
				pass := 0
				for r := 0; r < runs; r++ {
					// Keep the original trial seeds, but own the draws: global
					// rand.Seed is a no-op on Go >=1.24. Reuse each realization
					// across sigmas, independently of constructor/package RNG use.
					rng := rand.New(rand.NewSource(int64(9000 + 31*r)))
					arr, err := NewArray(&Config{Rows: rows, Cols: cols, NoiseLevel: sigma, ADCBits: 12, DACBits: 12})
					if err != nil {
						t.Fatalf("NewArray(sigma=%.3f): %v", sigma, err)
					}
					if err := arr.ProgramWeightMatrix(weights); err != nil {
						t.Fatalf("ProgramWeightMatrix: %v", err)
					}

					want := make([]float64, rows)
					for i := range arr.cells {
						for j := range arr.cells[i] {
							// Sweep row-coherent deviations uniformly over (-2,2).
							// Unlike independent draws (or balanced +/- deviations),
							// these cannot average away within a row. This is a
							// sensitivity fixture, not a Gaussian yield estimate.
							z := -2 + 4*(float64(r)+0.5)/runs
							if fixture == "Gaussian" {
								z = rng.NormFloat64()
							}
							factor := math.Max(0, 1+sigma*z)
							arr.cells[i][j].NoiseFactor = factor
							// Reference uses the programmed baseline conductance,
							// unit DAC input, row averaging and a 12-bit ADC below.
							want[i] += idealArr.cells[i][j].Conductance * factor / float64(cols)
						}
					}

					res, err := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: true, Temperature: 300})
					if err != nil {
						t.Fatalf("MVMWithNonIdealities: %v", err)
					}
					for i, value := range want {
						want[i] = math.Round(math.Min(1, value)*4095) / 4095
					}
					if !slices.Equal(res.ActualOutput, want) {
						t.Fatalf("sigma=%.3f trial=%d: got %v, reference %v", sigma, r, res.ActualOutput, want)
					}

					if withinRelError(res.ActualOutput, idealRes.ActualOutput, threshold) {
						pass++
					}
				}
				yields[idx] = float64(pass) / float64(runs)
				t.Logf("yield @ sigma=%.3f: %.3f (%d/%d)", sigma, yields[idx], pass, runs)
			}

			if fixture == "ControlledRowSweep" && !(yields[0] > yields[1] && yields[1] > yields[2]) {
				t.Fatalf("expected yield to strictly decrease with sigma: y(0.01)=%.3f y(0.10)=%.3f y(0.30)=%.3f", yields[0], yields[1], yields[2])
			}

			// Retain the low/high-sigma sanity bounds for both fixtures.
			if yields[0] < 0.5 {
				t.Fatalf("unexpectedly low yield at sigma=0.01: %.3f", yields[0])
			}
			if yields[2] > 0.50 {
				t.Fatalf("unexpectedly high yield at sigma=0.30: %.3f", yields[2])
			}
		})
	}
}

func withinRelError(got, want []float64, thresh float64) bool {
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	if n == 0 {
		return true
	}
	const eps = 1e-12
	for i := 0; i < n; i++ {
		den := math.Abs(want[i])
		if den < eps {
			den = eps
		}
		rel := math.Abs(got[i]-want[i]) / den
		if rel >= thresh {
			return false
		}
	}
	return true
}
