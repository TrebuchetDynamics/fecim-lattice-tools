package controller

import (
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// Remanent sweep diagnostic for polydomain Landau-Khalatnikov.
//
// Goal: with ensemble enabled, the plant (physics engine) should produce MANY distinct
// remanent levels after verify-at-E=0, as pulse amplitude increases.
//
// This test does NOT require hitting all 30 levels yet; it just measures whether the
// remanent staircase exists (a prerequisite for ISPP convergence).
func TestLandauKEnsemble_RemanentStaircase_Superlattice(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()

	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	// Enable NLS in ensemble mode to allow partial switching (probabilistic nucleation) with
	// deterministic per-domain RNG. This is a simple proxy for domain-to-domain threshold spread.
	solver.UseNLS = true
	solver.EnableEnsemble(256, mat, 0)

	// Pulse timing: short pulses to avoid fully switching every domain.
	dt := 2e-9
	pulseSteps := 4
	relaxSteps := 20

	seen := map[int]bool{}
	// Sweep magnitude from 0 to MaxField (2.5*Ec), positive branch.
	for k := 0; k <= 60; k++ {
		mag := (2.5 * float64(k) / 60.0) * mat.Ec
		// Start from negative saturation each trial.
		solver.SetState(-mat.Ps)
		// Apply pulse
		for i := 0; i < pulseSteps; i++ {
			solver.Step(mag, dt)
		}
		// Verify/relax at E=0
		for i := 0; i < relaxSteps; i++ {
			solver.Step(0, dt)
		}
		lvl := levelFromP(solver.GetState(), mat.Ps, 30)
		seen[lvl] = true
	}

	t.Logf("distinct remanent levels observed (ensemble): %d", len(seen))
	// Bare minimum: must be more than binary.
	if len(seen) < 6 {
		t.Fatalf("expected multi-level remanent staircase; got only %d distinct levels", len(seen))
	}
}
