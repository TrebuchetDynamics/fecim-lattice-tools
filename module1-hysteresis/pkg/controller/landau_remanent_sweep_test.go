package controller

import (
	"hash/fnv"
	"math"
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

	runSweep := func() (levels []int, distinct int, hash uint64, decreases int, worstDrop int, maxRelaxDeltaFrac float64, levelChangeAtEndOfRelax int) {
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

		ps := math.Abs(mat.Ps)
		if ps == 0 {
			ps = 1
		}

		seen := map[int]bool{}
		levels = make([]int, 0, 61)
		maxRelaxDeltaFrac = 0

		h := fnv.New64a()
		// Sweep magnitude from 0 to MaxField (2.5*Ec), positive branch.
		for k := 0; k <= 60; k++ {
			mag := (2.5 * float64(k) / 60.0) * mat.Ec
			// Start from negative saturation each trial.
			solver.SetState(-mat.Ps)
			// Apply pulse
			for i := 0; i < pulseSteps; i++ {
				solver.Step(mag, dt)
			}

			// Verify/relax at E=0; check stability via the final-step delta at E=0.
			prevRelaxP := solver.GetState()
			prevRelaxLevel := levelFromP(prevRelaxP, mat.Ps, 30)
			for i := 0; i < relaxSteps; i++ {
				solver.Step(0, dt)
				if i == relaxSteps-2 { // value immediately before the final relax step
					prevRelaxP = solver.GetState()
					prevRelaxLevel = levelFromP(prevRelaxP, mat.Ps, 30)
				}
			}
			pEnd := solver.GetState()
			lvl := levelFromP(pEnd, mat.Ps, 30)
			if lvl != prevRelaxLevel {
				levelChangeAtEndOfRelax++
			}
			deltaFrac := math.Abs(pEnd-prevRelaxP) / ps
			if deltaFrac > maxRelaxDeltaFrac {
				maxRelaxDeltaFrac = deltaFrac
			}

			levels = append(levels, lvl)
			seen[lvl] = true
			_, _ = h.Write([]byte{byte(lvl)})
		}

		// Monotonicity metrics (allow small quantization jitter; see diagnostics_remanent_staircase.md).
		for i := 1; i < len(levels); i++ {
			if levels[i] < levels[i-1] {
				decreases++
				drop := levels[i-1] - levels[i]
				if drop > worstDrop {
					worstDrop = drop
				}
			}
		}

		return levels, len(seen), h.Sum64(), decreases, worstDrop, maxRelaxDeltaFrac, levelChangeAtEndOfRelax
	}

	levels1, distinct1, hash1, dec1, worstDrop1, maxDeltaFrac1, relaxLevelChanges1 := runSweep()
	levels2, _, hash2, _, _, _, _ := runSweep()

	t.Logf("distinct remanent levels observed (ensemble): %d (hash=%016x)", distinct1, hash1)
	t.Logf("monotonicity: decreases=%d worstDrop=%d", dec1, worstDrop1)
	t.Logf("remanent stability at E=0: maxDeltaFrac(last-step)=%0.3e levelChanges=%d", maxDeltaFrac1, relaxLevelChanges1)

	// Determinism: with fixed seed and noise disabled, the level sequence must be identical run-to-run.
	if hash1 != hash2 || len(levels1) != len(levels2) {
		t.Fatalf("expected deterministic sweep hash/length; got hash1=%016x hash2=%016x len1=%d len2=%d", hash1, hash2, len(levels1), len(levels2))
	}
	for i := range levels1 {
		if levels1[i] != levels2[i] {
			t.Fatalf("expected deterministic sweep levels; mismatch at i=%d: run1=%d run2=%d (hash1=%016x hash2=%016x)", i, levels1[i], levels2[i], hash1, hash2)
		}
	}

	// Acceptance metrics are documented in diagnostics_remanent_staircase.md.
	// Bare minimum: must be more than binary.
	if distinct1 < 6 {
		t.Fatalf("expected multi-level remanent staircase; got only %d distinct levels", distinct1)
	}
	// Monotonic trend tolerance: allow small quantization jitter (<=2 single-level drops).
	if dec1 > 2 || worstDrop1 > 1 {
		t.Fatalf("remanent staircase not monotonic enough: decreases=%d worstDrop=%d", dec1, worstDrop1)
	}
	// Remanent stability after relax-at-E=0: final relax step should not change the quantized level,
	// and the final-step polarization drift should be small.
	if relaxLevelChanges1 != 0 {
		t.Fatalf("remanent not stable at E=0: level changed during final relax step (count=%d)", relaxLevelChanges1)
	}
	if maxDeltaFrac1 > 1e-3 {
		t.Fatalf("remanent not stable at E=0: maxDeltaFrac(last-step)=%0.3e exceeds 1e-3", maxDeltaFrac1)
	}
}
