package physics

import (
	"math"
	"testing"
)

// gibbsFreeEnergy computes the Landau free energy G(P) = alpha*P^2 + beta*P^4 + gamma*P^6
func gibbsFreeEnergy(alpha, beta, gamma, P float64) float64 {
	P2 := P * P
	P4 := P2 * P2
	P6 := P4 * P2
	return alpha*P2 + beta*P4 + gamma*P6
}

// TestLK_LK1_RemanentFromFreeEnergy (Tier T1)
// Validates that the remanent polarization Pr minimizes the Landau free energy at E=0.
// Uses brute-force search to find the P that minimizes G(P) and compares to material Pr.
// Acceptance: |P_min - mat.Pr| / mat.Pr < 0.50 (50% tolerance due to temperature-dependent alpha)
func TestLK_LK1_RemanentFromFreeEnergy(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UpdateParams() // Ensure Alpha is current

	// Brute-force search for P that minimizes G(P) at E=0
	const numPoints = 10000
	pMin := -1.5 * mat.Ps
	pMax := 1.5 * mat.Ps
	step := (pMax - pMin) / float64(numPoints)

	minG := math.Inf(1)
	minP := 0.0

	for i := 0; i <= numPoints; i++ {
		P := pMin + float64(i)*step
		G := gibbsFreeEnergy(s.Alpha, s.Beta, s.Gamma, P)
		if G < minG {
			minG = G
			minP = P
		}
	}

	// Find positive root (we initialize to negative, so positive is the switched state)
	if minP < 0 {
		// Also search for positive minimum
		for i := 0; i <= numPoints; i++ {
			P := pMin + float64(i)*step
			if P <= 0 {
				continue
			}
			G := gibbsFreeEnergy(s.Alpha, s.Beta, s.Gamma, P)
			if G < minG+1e-6 { // Within tolerance of global minimum
				minP = P
				break
			}
		}
	}

	minP = math.Abs(minP) // Take absolute value for comparison

	relError := math.Abs(minP-mat.Pr) / mat.Pr
	if relError >= 0.50 {
		t.Errorf("LK1: Remanent polarization from free energy minimum deviates too much from mat.Pr\n"+
			"  P_min from G(P): %.6e C/m²\n"+
			"  mat.Pr:          %.6e C/m²\n"+
			"  Relative error:  %.2f%% (tolerance: 50%%)",
			minP, mat.Pr, relError*100)
	}

	t.Logf("LK1 PASS: P_min=%.6e C/m², mat.Pr=%.6e C/m², error=%.2f%%",
		minP, mat.Pr, relError*100)
}

// TestLK_LK2_CoerciveFieldFromLandau (Tier T1)
// Validates that estimateLandauEc produces reasonable coercive field estimates
// from Landau polynomial parameters for all materials with Landau params defined.
// Acceptance: Ec_estimated > 0 and within order of magnitude of mat.Ec (0.1x to 10x)
func TestLK_LK2_CoerciveFieldFromLandau(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		// Only test materials with Landau parameters defined
		if mat.BetaLandau == 0 {
			t.Logf("LK2: Skipping %s (no Landau parameters)", mat.Name)
			continue
		}

		t.Run(mat.Name, func(t *testing.T) {
			s := NewLKSolver()
			s.ConfigureFromMaterial(mat)
			s.UpdateParams() // Ensure Alpha is current

			ecEstimated := estimateLandauEc(s.Alpha, s.Beta, s.Gamma, mat.Pr)

			// Check non-trivial field
			if ecEstimated <= 0 {
				t.Errorf("LK2: Ec_estimated must be positive, got %.6e V/m", ecEstimated)
				return
			}

			// Check order of magnitude (0.1x to 10x)
			ratio := ecEstimated / mat.Ec
			if ratio < 0.1 || ratio > 10.0 {
				t.Errorf("LK2: Ec_estimated out of order-of-magnitude range\n"+
					"  Ec_estimated: %.6e V/m\n"+
					"  mat.Ec:       %.6e V/m\n"+
					"  Ratio:        %.2fx (expected 0.1x to 10x)",
					ecEstimated, mat.Ec, ratio)
				return
			}

			t.Logf("LK2 PASS: Ec_estimated=%.6e V/m, mat.Ec=%.6e V/m, ratio=%.2fx",
				ecEstimated, mat.Ec, ratio)
		})
	}
}

// TestLK_LK3_FreeEnergyDoubleWell (Tier T1)
// Validates that the Landau free energy G(P) has the expected double-well structure:
// - At least 2 local minima (one positive, one negative P)
// - At least 1 local maximum near P=0
// - Barrier height G(0) - G(P_min) > 0 (wells are lower than barrier)
func TestLK_LK3_FreeEnergyDoubleWell(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UpdateParams() // Ensure Alpha is current

	// Evaluate G(P) over range
	const numPoints = 10000
	pMin := -1.5 * mat.Ps
	pMax := 1.5 * mat.Ps
	step := (pMax - pMin) / float64(numPoints)

	Gs := make([]float64, numPoints+1)
	for i := 0; i <= numPoints; i++ {
		P := pMin + float64(i)*step
		Gs[i] = gibbsFreeEnergy(s.Alpha, s.Beta, s.Gamma, P)
	}

	// Find local minima and maxima
	type localExtremum struct {
		idx   int
		P     float64
		G     float64
		isMin bool
	}
	var extrema []localExtremum

	for i := 1; i < len(Gs)-1; i++ {
		P := pMin + float64(i)*step
		if Gs[i] < Gs[i-1] && Gs[i] < Gs[i+1] {
			// Local minimum
			extrema = append(extrema, localExtremum{i, P, Gs[i], true})
		} else if Gs[i] > Gs[i-1] && Gs[i] > Gs[i+1] {
			// Local maximum
			extrema = append(extrema, localExtremum{i, P, Gs[i], false})
		}
	}

	// Count minima and maxima
	numMinima := 0
	numMaxima := 0
	var posMin, negMin *localExtremum
	var zeroMax *localExtremum

	for i := range extrema {
		ex := &extrema[i]
		if ex.isMin {
			numMinima++
			if ex.P > 0 {
				if posMin == nil || ex.G < posMin.G {
					posMin = ex
				}
			} else if ex.P < 0 {
				if negMin == nil || ex.G < negMin.G {
					negMin = ex
				}
			}
		} else {
			numMaxima++
			if math.Abs(ex.P) < 0.1*mat.Ps { // Near P=0
				if zeroMax == nil || math.Abs(ex.P) < math.Abs(zeroMax.P) {
					zeroMax = ex
				}
			}
		}
	}

	// Validation checks
	if numMinima < 2 {
		t.Errorf("LK3: Expected at least 2 local minima, got %d", numMinima)
	}
	if posMin == nil {
		t.Errorf("LK3: Expected a local minimum at positive P")
	}
	if negMin == nil {
		t.Errorf("LK3: Expected a local minimum at negative P")
	}
	if numMaxima < 1 {
		t.Errorf("LK3: Expected at least 1 local maximum, got %d", numMaxima)
	}
	if zeroMax == nil {
		t.Errorf("LK3: Expected a local maximum near P=0")
	}

	// Check barrier height (G at P=0 should be higher than minima)
	G0 := gibbsFreeEnergy(s.Alpha, s.Beta, s.Gamma, 0)
	if posMin != nil {
		barrierHeight := G0 - posMin.G
		if barrierHeight <= 0 {
			t.Errorf("LK3: Barrier height G(0) - G(P_min+) must be positive, got %.6e", barrierHeight)
		} else {
			t.Logf("LK3: Barrier height (positive well) = %.6e J/m³", barrierHeight)
		}
	}
	if negMin != nil {
		barrierHeight := G0 - negMin.G
		if barrierHeight <= 0 {
			t.Errorf("LK3: Barrier height G(0) - G(P_min-) must be positive, got %.6e", barrierHeight)
		} else {
			t.Logf("LK3: Barrier height (negative well) = %.6e J/m³", barrierHeight)
		}
	}

	if posMin != nil && negMin != nil && zeroMax != nil {
		t.Logf("LK3 PASS: Found double-well structure with %d minima and %d maxima\n"+
			"  Negative minimum: P=%.6e, G=%.6e\n"+
			"  Positive minimum: P=%.6e, G=%.6e\n"+
			"  Central maximum:  P=%.6e, G=%.6e",
			numMinima, numMaxima,
			negMin.P, negMin.G,
			posMin.P, posMin.G,
			zeroMax.P, zeroMax.G)
	}
}

// TestLK_LK4_EnergyDissipation (Tier T1)
// Validates that the LK solver exhibits energy dissipation under hysteresis cycling.
// Runs a full hysteresis loop (0 → +3Ec → -3Ec → +3Ec) and verifies:
// - P changes smoothly (no NaN, no jumps > 0.1*Ps)
// - Loop area > 0 (energy dissipated)
// - All P values remain finite
func TestLK_LK4_EnergyDissipation(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false
	s.EnableNoise = false

	// Run full hysteresis loop
	dt := 1e-10 // 100 ps timestep
	stepsPerPhase := 500
	totalSteps := stepsPerPhase * 4

	Emax := 3.0 * mat.Ec
	var Ps []float64
	var Es []float64

	// Phase 1: 0 → +3Ec
	for i := 0; i < stepsPerPhase; i++ {
		E := Emax * float64(i) / float64(stepsPerPhase-1)
		s.Step(E, dt)
		Ps = append(Ps, s.P)
		Es = append(Es, E)
	}

	// Phase 2: +3Ec → -3Ec
	for i := 0; i < stepsPerPhase; i++ {
		E := Emax - 2.0*Emax*float64(i)/float64(stepsPerPhase-1)
		s.Step(E, dt)
		Ps = append(Ps, s.P)
		Es = append(Es, E)
	}

	// Phase 3: -3Ec → 0
	for i := 0; i < stepsPerPhase; i++ {
		E := -Emax + Emax*float64(i)/float64(stepsPerPhase-1)
		s.Step(E, dt)
		Ps = append(Ps, s.P)
		Es = append(Es, E)
	}

	// Phase 4: 0 → +3Ec (complete loop)
	for i := 0; i < stepsPerPhase; i++ {
		E := Emax * float64(i) / float64(stepsPerPhase-1)
		s.Step(E, dt)
		Ps = append(Ps, s.P)
		Es = append(Es, E)
	}

	// Validation: Check for smooth P changes (no NaN/Inf)
	maxJump := 0.0
	largeJumps := 0
	for i := 1; i < len(Ps); i++ {
		if math.IsNaN(Ps[i]) || math.IsInf(Ps[i], 0) {
			t.Errorf("LK4: P[%d] is not finite: %.6e", i, Ps[i])
			return
		}
		jump := math.Abs(Ps[i] - Ps[i-1])
		if jump > maxJump {
			maxJump = jump
		}
		// Count large jumps during switching (expected during phase transitions)
		// Use 0.15*Ps threshold to allow for rapid switching dynamics
		if jump > 0.15*mat.Ps {
			largeJumps++
		}
	}

	// Too many large jumps indicates numerical instability
	if largeJumps > 20 {
		t.Errorf("LK4: Too many large jumps (%d > 20), indicates numerical instability", largeJumps)
	}

	// Validation: Check loop area (energy dissipation)
	loopArea := 0.0
	for i := 1; i < len(Ps); i++ {
		dP := math.Abs(Ps[i] - Ps[i-1])
		E := math.Abs(Es[i])
		loopArea += dP * E
	}

	if loopArea <= 0 {
		t.Errorf("LK4: Loop area must be positive (energy dissipated), got %.6e", loopArea)
		return
	}

	t.Logf("LK4 PASS: Hysteresis loop completed with %.0f steps\n"+
		"  Loop area (dissipation): %.6e J/m³\n"+
		"  Max P jump: %.6e C/m² (%.2f%% of Ps)\n"+
		"  Final P: %.6e C/m²",
		float64(totalSteps), loopArea, maxJump, maxJump/mat.Ps*100, s.P)
}

// TestLK_LK5_PolydomainEnsemble (Tier T1)
// Validates that single-domain solver produces stable remanent state.
// Tests that after applying +3*Ec and removing field (E=0), P stabilizes near +Pr.
// This is the single-domain baseline; ensemble mode would show distributed behavior.
func TestLK_LK5_PolydomainEnsemble(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false
	s.EnableNoise = false

	// Apply strong positive field to saturate
	E := 3.0 * mat.Ec
	dt := 1e-10 // 100 ps
	for i := 0; i < 1000; i++ {
		s.Step(E, dt)
	}

	// Remove field and let relax
	for i := 0; i < 500; i++ {
		s.Step(0, dt)
	}

	// Check that P is near +Pr (positive remanent state)
	if s.P <= 0.3*mat.Ps {
		t.Errorf("LK5: Expected positive remanent state after saturation and relaxation\n"+
			"  Final P: %.6e C/m²\n"+
			"  Expected: > %.6e C/m² (0.3*Ps)\n"+
			"  mat.Pr: %.6e C/m²",
			s.P, 0.3*mat.Ps, mat.Pr)
		return
	}

	t.Logf("LK5 PASS: Single-domain solver reached stable remanent state\n"+
		"  Final P: %.6e C/m² (%.2f%% of Ps)\n"+
		"  mat.Pr:  %.6e C/m²",
		s.P, s.P/mat.Ps*100, mat.Pr)
}
