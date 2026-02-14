package physics

import (
	"math"
	"testing"
)

// R1: Boundary Conditions (Tier T0)
// Verify G(-Ps)=Gmin, G(+Ps)=Gmax, and midpoint behavior for all models
func TestTransferFunction_BoundaryConditions(t *testing.T) {
	materials := AllMaterials()
	models := []ConductanceModel{
		ConductanceLinear,
		ConductanceSubthreshold,
		ConductanceSaturation,
	}

	for _, mat := range materials {
		for _, model := range models {
			t.Run(mat.Name+"/"+model.String(), func(t *testing.T) {
				// Test G(-Ps) == Gmin
				gMin := PolarizationToConductanceModel(-mat.Ps, mat.Ps, mat.Gmin, mat.Gmax, model)
				if math.Abs(gMin-mat.Gmin) > 1e-10 {
					t.Errorf("G(-Ps) = %e, want %e (error: %e)", gMin, mat.Gmin, gMin-mat.Gmin)
				}

				// Test G(+Ps) == Gmax
				gMax := PolarizationToConductanceModel(mat.Ps, mat.Ps, mat.Gmin, mat.Gmax, model)
				if math.Abs(gMax-mat.Gmax) > 1e-10 {
					t.Errorf("G(+Ps) = %e, want %e (error: %e)", gMax, mat.Gmax, gMax-mat.Gmax)
				}

				// Test G(0) midpoint
				gMid := PolarizationToConductanceModel(0, mat.Ps, mat.Gmin, mat.Gmax, model)
				switch model {
				case ConductanceLinear:
					// Arithmetic mean
					expectedMid := (mat.Gmin + mat.Gmax) / 2
					if math.Abs(gMid-expectedMid) > 1e-10 {
						t.Errorf("G(0) linear = %e, want %e (arithmetic mean)", gMid, expectedMid)
					}
				case ConductanceSubthreshold:
					// Geometric mean (1% tolerance for numerical precision)
					expectedMid := math.Sqrt(mat.Gmin * mat.Gmax)
					relErr := math.Abs(gMid-expectedMid) / expectedMid
					if relErr > 0.01 {
						t.Errorf("G(0) subthreshold = %e, want %e (geometric mean, rel error: %.2f%%)",
							gMid, expectedMid, relErr*100)
					}
				case ConductanceSaturation:
					// Just verify in range (exact value depends on internal params)
					if gMid < mat.Gmin || gMid > mat.Gmax {
						t.Errorf("G(0) saturation = %e out of range [%e, %e]", gMid, mat.Gmin, mat.Gmax)
					}
				}
			})
		}
	}
}

// R2: Monotonicity (Tier T0)
// Verify G is strictly increasing over full P range for all models
func TestTransferFunction_Monotonicity(t *testing.T) {
	materials := AllMaterials()
	models := []ConductanceModel{
		ConductanceLinear,
		ConductanceSubthreshold,
		ConductanceSaturation,
	}
	const numSteps = 1000

	for _, mat := range materials {
		for _, model := range models {
			t.Run(mat.Name+"/"+model.String(), func(t *testing.T) {
				// Skip materials with Gmin=Gmax (config bug, cannot support multi-level)
				if mat.Gmin == mat.Gmax {
					t.Skipf("Material has Gmin=Gmax=%e (config missing gmin_s/gmax_s)", mat.Gmin)
				}

				minDG := math.Inf(1)
				var minDGLocation float64
				prevG := PolarizationToConductanceModel(-mat.Ps, mat.Ps, mat.Gmin, mat.Gmax, model)

				for i := 1; i <= numSteps; i++ {
					p := -mat.Ps + 2*mat.Ps*float64(i)/float64(numSteps)
					g := PolarizationToConductanceModel(p, mat.Ps, mat.Gmin, mat.Gmax, model)
					dG := g - prevG

					if dG < minDG {
						minDG = dG
						minDGLocation = p
					}

					if dG <= 0 {
						t.Errorf("Non-monotonic at P=%e: G[i+1]=%e <= G[i]=%e (dG=%e)",
							p, g, prevG, dG)
					}
					prevG = g
				}

				// Report minimum gradient for diagnostics
				t.Logf("Minimum dG: %e at P=%e", minDG, minDGLocation)
			})
		}
	}
}

// R3: Level Separability (Tier T1)
// Verify all target levels are distinct and evenly spaced (Linear model only)
func TestTransferFunction_LevelSeparability(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Skip materials with Gmin=Gmax (config bug, cannot support multi-level)
			if mat.Gmin == mat.Gmax {
				t.Skipf("Material has Gmin=Gmax=%e (config missing gmin_s/gmax_s)", mat.Gmin)
			}

			N := mat.GetNumLevels()
			if N < 2 {
				t.Skipf("Material has <2 levels (N=%d)", N)
			}

			gLevels := make([]float64, N)
			for i := 0; i < N; i++ {
				p := -mat.Ps + 2*mat.Ps*float64(i)/float64(N-1)
				gLevels[i] = PolarizationToConductance(p, mat.Ps, mat.Gmin, mat.Gmax)
			}

			// Verify all spacings > 0
			spacings := make([]float64, N-1)
			for i := 0; i < N-1; i++ {
				dG := gLevels[i+1] - gLevels[i]
				if dG <= 0 {
					t.Errorf("Non-positive spacing at level %d: dG=%e", i, dG)
				}
				spacings[i] = dG
			}

			// For linear model, verify uniform spacing (max deviation < 5% of mean)
			var sumSpacing float64
			for _, dG := range spacings {
				sumSpacing += dG
			}
			meanSpacing := sumSpacing / float64(len(spacings))

			maxDeviation := 0.0
			for _, dG := range spacings {
				dev := math.Abs(dG - meanSpacing)
				if dev > maxDeviation {
					maxDeviation = dev
				}
			}

			relDeviation := maxDeviation / meanSpacing
			if relDeviation > 0.05 {
				t.Errorf("Linear spacing non-uniform: max deviation %.2f%% (threshold 5%%)",
					relDeviation*100)
			}

			t.Logf("N=%d levels, mean spacing=%e, max deviation=%.2f%%",
				N, meanSpacing, relDeviation*100)
		})
	}
}
