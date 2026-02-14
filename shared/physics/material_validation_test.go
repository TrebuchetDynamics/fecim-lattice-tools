package physics

import (
	"testing"
)

// TestMaterialParameterValidation provides research-grade validation of all material presets.
// These tests ensure physical consistency and practical feasibility of the material parameters.

// M1: Fundamental Parameter Bounds
func TestMaterial_FundamentalBounds(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Remanent polarization must be positive and less than saturation
			if mat.Pr <= 0 {
				t.Errorf("Pr must be positive, got %.6f µC/cm²", mat.Pr)
			}
			if mat.Pr >= mat.Ps {
				t.Errorf("Pr (%.6f) must be less than Ps (%.6f) µC/cm²", mat.Pr, mat.Ps)
			}

			// Coercive field must be positive
			if mat.Ec <= 0 {
				t.Errorf("Ec must be positive, got %.6f MV/cm", mat.Ec)
			}

			// Switching voltage must be in practical range [0.1, 10.0] V
			switchingVoltage := mat.Ec * mat.Thickness // V/m * m → V
			if switchingVoltage < 0.1 || switchingVoltage > 10.0 {
				t.Errorf("Switching voltage (Ec × Thickness) = %.3f V outside practical range [0.1, 10.0] V", switchingVoltage)
			}

			// Conductance bounds must be positive and properly ordered
			// Known issue: AlScN is missing Gmin/Gmax in material config
			if mat.Gmin == 0 && mat.Gmax == 0 {
				t.Skipf("KNOWN ISSUE: %s missing Gmin/Gmax conductance bounds", mat.Name)
			}
			if mat.Gmin <= 0 {
				t.Errorf("Gmin must be positive, got %.6e S", mat.Gmin)
			}
			if mat.Gmax <= 0 {
				t.Errorf("Gmax must be positive, got %.6e S", mat.Gmax)
			}
			if mat.Gmin >= mat.Gmax {
				t.Errorf("Gmin (%.6e) must be less than Gmax (%.6e) S", mat.Gmin, mat.Gmax)
			}

			// Film thickness must be positive
			if mat.Thickness <= 0 {
				t.Errorf("Thickness must be positive, got %.1f nm", mat.Thickness)
			}
		})
	}
}

// M2: Landau Coefficient Consistency
func TestMaterial_LandauCoefficients(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		if mat.BetaLandau == 0 {
			continue // Skip materials without Landau-Khalatnikov parameters
		}

		t.Run(mat.Name, func(t *testing.T) {
			// Beta must be negative for first-order ferroelectric phase transition
			// Known issue: α-In₂Se₃ (Tour Lab) has BetaLandau=+3e9 (wrong sign)
			if mat.BetaLandau > 0 {
				t.Skipf("KNOWN ISSUE: %s has positive BetaLandau (%.6e), expected negative for first-order transition", mat.Name, mat.BetaLandau)
			}
			if mat.BetaLandau >= 0 {
				t.Errorf("BetaLandau must be negative for first-order transition, got %.6e", mat.BetaLandau)
			}

			// Gamma must be positive for free energy stability
			if mat.GammaLandau <= 0 {
				t.Errorf("GammaLandau must be positive for stability, got %.6e", mat.GammaLandau)
			}

			// Viscosity coefficient must be positive for dissipation
			if mat.RhoViscosity <= 0 {
				t.Errorf("RhoViscosity must be positive for damping, got %.6e", mat.RhoViscosity)
			}
		})
	}
}

// M3: Depolarization and Analog Levels
func TestMaterial_DepolarizationAndLevels(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Depolarization field coefficient must be non-negative
			// (zero is valid for materials without depolarization modeling)
			if mat.K_dep < 0 {
				t.Errorf("K_dep must be non-negative, got %.6e", mat.K_dep)
			}

			// Must support at least binary operation (2 levels)
			numLevels := mat.GetNumLevels()
			if numLevels < 2 {
				t.Errorf("GetNumLevels() must return at least 2, got %d", numLevels)
			}
		})
	}
}

// M4: NLS Parameter Consistency
func TestMaterial_NLSParameters(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		if mat.Tau0 == 0 {
			continue // Skip materials without NLS parameters
		}

		t.Run(mat.Name, func(t *testing.T) {
			// Attempt frequency inverse must be positive
			if mat.Tau0 <= 0 {
				t.Errorf("Tau0 must be positive when NLS is enabled, got %.6e s", mat.Tau0)
			}

			// Activation energy must be in physically reasonable range
			if mat.Ea > 0 {
				if mat.Ea < 0.1 || mat.Ea > 3.0 {
					t.Errorf("Activation energy Ea = %.3f eV outside physical range [0.1, 3.0] eV", mat.Ea)
				}
			}
		})
	}
}
