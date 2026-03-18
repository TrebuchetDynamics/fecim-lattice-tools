package physics_test

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

// ---------------------------------------------------------------------------
// LK <-> Preisach P-E Loop Consistency Tests
//
// Research-grade validation covering three gaps identified in the testing audit:
//   1. Loop closure: start/end polarization must match after full cycles.
//   2. Pr measurement: LK solver Pr vs. material Pr (depolarization causes ~7%
//      undershoot; assert within 15%).
//   3. Cross-model consistency: LK vs Preisach Pr and Ec agreement.
//
// All tests disable NLS and noise for deterministic, quasi-static comparison.
// ---------------------------------------------------------------------------

// hzoMaterials returns the three HZO material presets under test.
func hzoMaterials() []struct {
	name string
	mat  *physics.HZOMaterial
} {
	return []struct {
		name string
		mat  *physics.HZOMaterial
	}{
		{"DefaultHZO", physics.DefaultHZO()},
		{"FeCIM", physics.FeCIMMaterial()},
		{"LiteratureSuperlattice", physics.LiteratureSuperlattice()},
	}
}

// lkGeneratePELoop runs the LK solver through nCycles full sinusoidal P-E
// cycles and returns the field and polarization traces. The solver is configured
// from mat with NLS and noise disabled. dt is the integration timestep, eMax is
// the peak field amplitude, and pointsPerCycle controls the resolution.
func lkGeneratePELoop(
	mat *physics.HZOMaterial,
	nCycles int,
	dt float64,
	eMax float64,
	pointsPerCycle int,
) (fields, pols []float64, solver *physics.LKSolver) {
	s := physics.NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false
	s.EnableNoise = false
	s.SetState(-math.Abs(mat.Pr))

	totalPoints := nCycles * pointsPerCycle
	// Number of LK sub-steps per field point. More sub-steps let the solver
	// relax closer to equilibrium at each field value.
	stepsPerPoint := 200

	fields = make([]float64, 0, totalPoints)
	pols = make([]float64, 0, totalPoints)

	for i := 0; i < totalPoints; i++ {
		// Sinusoidal field: starts at -Emax, sweeps through +Emax and back.
		phase := 2.0 * math.Pi * float64(i) / float64(pointsPerCycle)
		E := eMax * math.Sin(phase)

		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		fields = append(fields, E)
		pols = append(pols, s.GetState())
	}

	return fields, pols, s
}

// extractPrEc extracts remanent polarization (Pr) and coercive field (Ec) from
// a P-E trace by interpolating E=0 crossings (for Pr) and P=0 crossings (for
// Ec). Handles both sign-change crossings and exact-zero field points.
// Returns averages over all crossings found.
func extractPrEc(fields, pols []float64) (pr, ec float64, prOK, ecOK bool) {
	// Pr: find P at E=0 crossings.
	var prVals []float64
	for i := 0; i < len(fields); i++ {
		// Exact zero: record P directly.
		if fields[i] == 0 {
			prVals = append(prVals, math.Abs(pols[i]))
			continue
		}
		// Sign-change interpolation.
		if i > 0 && fields[i-1]*fields[i] < 0 {
			dx := fields[i] - fields[i-1]
			if dx != 0 {
				f := -fields[i-1] / dx
				if f >= 0 && f <= 1 {
					p0 := pols[i-1] + f*(pols[i]-pols[i-1])
					prVals = append(prVals, math.Abs(p0))
				}
			}
		}
	}
	if len(prVals) > 0 {
		for _, v := range prVals {
			pr += v
		}
		pr /= float64(len(prVals))
		prOK = true
	}

	// Ec: find E at P=0 crossings.
	var ecVals []float64
	for i := 0; i < len(pols); i++ {
		// Exact zero: record E directly.
		if pols[i] == 0 {
			ecVals = append(ecVals, math.Abs(fields[i]))
			continue
		}
		// Sign-change interpolation.
		if i > 0 && pols[i-1]*pols[i] < 0 {
			dy := pols[i] - pols[i-1]
			if dy != 0 {
				f := -pols[i-1] / dy
				if f >= 0 && f <= 1 {
					ec0 := fields[i-1] + f*(fields[i]-fields[i-1])
					ecVals = append(ecVals, math.Abs(ec0))
				}
			}
		}
	}
	if len(ecVals) > 0 {
		for _, v := range ecVals {
			ec += v
		}
		ec /= float64(len(ecVals))
		ecOK = true
	}

	return
}

// ---------------------------------------------------------------------------
// Test 1: Loop Closure
// ---------------------------------------------------------------------------

// TestLK_LoopClosure verifies that two full sinusoidal P-E cycles return the
// polarization to its starting value. The start and end P must match within 1%
// of Ps, confirming no net drift from numerical dissipation or asymmetry.
func TestLK_LoopClosure(t *testing.T) {
	for _, tc := range hzoMaterials() {
		t.Run(tc.name, func(t *testing.T) {
			mat := tc.mat
			eMax := 2.0 * mat.Ec
			dt := 1e-12
			pointsPerCycle := 400

			// Run a conditioning half-cycle first to establish the hysteresis
			// state, then measure closure over 2 full cycles.
			s := physics.NewLKSolver()
			s.ConfigureFromMaterial(mat)
			s.UseNLS = false
			s.EnableNoise = false
			s.SetState(-math.Abs(mat.Pr))

			stepsPerPoint := 200

			// Conditioning: run one full cycle to settle the loop.
			for i := 0; i < pointsPerCycle; i++ {
				phase := 2.0 * math.Pi * float64(i) / float64(pointsPerCycle)
				E := eMax * math.Sin(phase)
				for k := 0; k < stepsPerPoint; k++ {
					s.Step(E, dt)
				}
			}

			// Record P at the start of the measured cycles.
			pStart := s.GetState()

			// Run 2 full cycles.
			nMeasuredCycles := 2
			for i := 0; i < nMeasuredCycles*pointsPerCycle; i++ {
				phase := 2.0 * math.Pi * float64(i) / float64(pointsPerCycle)
				E := eMax * math.Sin(phase)
				for k := 0; k < stepsPerPoint; k++ {
					s.Step(E, dt)
				}
			}

			pEnd := s.GetState()

			diff := math.Abs(pEnd - pStart)
			threshold := 0.01 * math.Abs(mat.Ps) // 1% of Ps
			pctPs := 100.0 * diff / math.Abs(mat.Ps)

			t.Logf("Loop closure: P_start=%.6e C/m2, P_end=%.6e C/m2, |diff|=%.3e (%.2f%% of Ps)",
				pStart, pEnd, diff, pctPs)

			if diff > threshold {
				t.Errorf("Loop closure failed: |P_end - P_start| = %.3e > 1%% of Ps (%.3e)",
					diff, threshold)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test 2: Pr / Ec Measurement from LK Solver
// ---------------------------------------------------------------------------

// TestLK_PrEcMeasurement generates a full P-E loop with the LK solver for each
// HZO material and compares the measured Pr and Ec against the material's
// advertised values.
//
// Expected discrepancies:
//   - Pr: depolarization field (K_dep) pulls remanent state inward by ~7%.
//     Assert within 15%.
//   - Ec: dynamic effects, depolarization, and viscosity-dependent sweep-rate
//     sensitivity all shift the dynamic coercive field from the material's
//     advertised Ec. The depolarization field K_dep*P introduces a restoring
//     force that reduces the apparent switching field in dynamic LK simulations.
//     Materials with low viscosity show larger Ec shifts because the solver
//     switches faster. Assert within 50%.
func TestLK_PrEcMeasurement(t *testing.T) {
	for _, tc := range hzoMaterials() {
		t.Run(tc.name, func(t *testing.T) {
			mat := tc.mat

			// Use 3*Ec to ensure full saturation and clear switching for all
			// materials, including those with low viscosity.
			eMax := 3.0 * mat.Ec
			dt := 1e-12
			pointsPerCycle := 500

			// Generate 2 cycles; extract from second cycle for steady state.
			fields, pols, _ := lkGeneratePELoop(mat, 2, dt, eMax, pointsPerCycle)

			// Use only the second cycle for measurement.
			secondCycleStart := pointsPerCycle
			fSlice := fields[secondCycleStart:]
			pSlice := pols[secondCycleStart:]

			prMeasured, ecMeasured, prOK, ecOK := extractPrEc(fSlice, pSlice)

			if !prOK {
				t.Fatal("Failed to extract Pr from LK P-E loop (no E=0 crossings found)")
			}
			if !ecOK {
				t.Fatal("Failed to extract Ec from LK P-E loop (no P=0 crossings found)")
			}

			// Pr comparison.
			prRef := math.Abs(mat.Pr)
			prPct := 100.0 * (prRef - prMeasured) / prRef
			prRelErr := math.Abs(prMeasured-prRef) / prRef

			t.Logf("Pr: material=%.4f C/m2 (%.1f uC/cm2), LK measured=%.4f C/m2 (%.1f uC/cm2), "+
				"discrepancy=%.1f%% (%+.1f%% undershoot)",
				prRef, prRef*1e6/1e4, prMeasured, prMeasured*1e6/1e4, prRelErr*100, prPct)

			if prRelErr > 0.15 {
				t.Errorf("Pr discrepancy %.1f%% exceeds 15%% tolerance (material Pr=%.4f, LK Pr=%.4f)",
					prRelErr*100, prRef, prMeasured)
			}

			// Ec comparison.
			// Dynamic LK Ec depends on sweep rate and depolarization strength.
			// K_dep creates a restoring field E_dep = K_dep*P that opposes
			// switching, effectively reducing the field at which P=0 is reached.
			// Low-viscosity materials switch faster and show larger Ec shifts.
			// Tolerance is 50% to accommodate these physical effects.
			ecRef := math.Abs(mat.Ec)
			ecRelErr := math.Abs(ecMeasured-ecRef) / ecRef

			t.Logf("Ec: material=%.4e V/m (%.2f MV/cm), LK measured=%.4e V/m (%.2f MV/cm), "+
				"discrepancy=%.1f%%",
				ecRef, ecRef/1e8, ecMeasured, ecMeasured/1e8, ecRelErr*100)

			if ecRelErr > 0.50 {
				t.Errorf("Ec discrepancy %.1f%% exceeds 50%% tolerance (material Ec=%.4e, LK Ec=%.4e)",
					ecRelErr*100, ecRef, ecMeasured)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test 3: LK vs Preisach Consistency
// ---------------------------------------------------------------------------

// TestLK_PreisachConsistency generates quasi-static P-E loops with both the LK
// solver and Preisach model for DefaultHZO and compares their Pr and Ec.
//
// The two engines are independent implementations of ferroelectric switching:
//   - LK: time-domain ODE integration of the Landau-Khalatnikov equation
//   - Preisach: history-dependent classical Preisach relay model
//
// Both are calibrated against the same material parameters so their macroscopic
// observables (Pr, Ec) should agree within reasonable bounds.
//
// Tolerances:
//   - Pr agreement within 20% (different physical models, same material)
//   - Ec agreement within 30% (LK dynamic Ec depends on sweep rate; Preisach
//     uses direct field-to-polarization mapping)
func TestLK_PreisachConsistency(t *testing.T) {
	mat := physics.DefaultHZO()

	// --- Preisach P-E loop ---
	preisachModel := ferroelectric.NewPreisachModel(mat)
	eMaxPreisach := 2.0 * mat.Ec
	preisachFields, preisachPols := preisachModel.GetHysteresisLoop(eMaxPreisach, 250)

	prPreisach, ecPreisach, prOKp, ecOKp := extractPrEc(preisachFields, preisachPols)
	if !prOKp {
		t.Fatal("Failed to extract Pr from Preisach P-E loop")
	}
	if !ecOKp {
		t.Fatal("Failed to extract Ec from Preisach P-E loop")
	}

	// --- LK P-E loop (quasi-static regime) ---
	// Use dt=2e-12 with 400 sub-steps per point for a total dwell time of
	// 8e-10 s per field value. This is slow enough relative to the LK
	// switching time (~ns range for HZO) to approximate quasi-static behavior.
	eMaxLK := 3.0 * mat.Ec
	dtLK := 2e-12
	pointsPerCycle := 500

	// Build a dedicated LK loop with more sub-steps for quasi-static behavior.
	sLK := physics.NewLKSolver()
	sLK.ConfigureFromMaterial(mat)
	sLK.UseNLS = false
	sLK.EnableNoise = false
	sLK.SetState(-math.Abs(mat.Pr))

	lkStepsPerPoint := 400

	// Run 2 conditioning cycles.
	for cycle := 0; cycle < 2; cycle++ {
		for i := 0; i < pointsPerCycle; i++ {
			phase := 2.0 * math.Pi * float64(i) / float64(pointsPerCycle)
			E := eMaxLK * math.Sin(phase)
			for k := 0; k < lkStepsPerPoint; k++ {
				sLK.Step(E, dtLK)
			}
		}
	}

	// Collect the measurement cycle.
	lkFields2 := make([]float64, 0, pointsPerCycle)
	lkPols2 := make([]float64, 0, pointsPerCycle)
	for i := 0; i < pointsPerCycle; i++ {
		phase := 2.0 * math.Pi * float64(i) / float64(pointsPerCycle)
		E := eMaxLK * math.Sin(phase)
		for k := 0; k < lkStepsPerPoint; k++ {
			sLK.Step(E, dtLK)
		}
		lkFields2 = append(lkFields2, E)
		lkPols2 = append(lkPols2, sLK.GetState())
	}

	prLK, ecLK, prOKlk, ecOKlk := extractPrEc(lkFields2, lkPols2)
	if !prOKlk {
		t.Fatal("Failed to extract Pr from LK P-E loop")
	}
	if !ecOKlk {
		t.Fatal("Failed to extract Ec from LK P-E loop")
	}

	// --- Comparison ---
	prRelDiff := math.Abs(prPreisach-prLK) / math.Max(prPreisach, prLK)
	ecRelDiff := math.Abs(ecPreisach-ecLK) / math.Max(ecPreisach, ecLK)

	t.Logf("Preisach: Pr=%.4f C/m2 (%.1f uC/cm2), Ec=%.4e V/m (%.2f MV/cm)",
		prPreisach, prPreisach*1e6/1e4, ecPreisach, ecPreisach/1e8)
	t.Logf("LK:       Pr=%.4f C/m2 (%.1f uC/cm2), Ec=%.4e V/m (%.2f MV/cm)",
		prLK, prLK*1e6/1e4, ecLK, ecLK/1e8)
	t.Logf("Pr relative difference: %.1f%% (tolerance: 20%%)", prRelDiff*100)
	t.Logf("Ec relative difference: %.1f%% (tolerance: 30%%)", ecRelDiff*100)

	if prRelDiff > 0.20 {
		t.Errorf("LK vs Preisach Pr mismatch exceeds 20%%: Preisach=%.4f, LK=%.4f, relDiff=%.1f%%",
			prPreisach, prLK, prRelDiff*100)
	}

	if ecRelDiff > 0.30 {
		t.Errorf("LK vs Preisach Ec mismatch exceeds 30%%: Preisach=%.4e, LK=%.4e, relDiff=%.1f%%",
			ecPreisach, ecLK, ecRelDiff*100)
	}
}
