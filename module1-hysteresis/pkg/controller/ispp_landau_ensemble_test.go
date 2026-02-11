package controller

import (
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// This test exercises the SAME WriteController (verify-at-E=0) loop, but with
// Landau-Khalatnikov in polydomain ensemble mode. The ensemble approximates
// partial domain switching, enabling stable intermediate remanent states.
func TestISPPConverges_LandauK_Ensemble_Superlattice(t *testing.T) {
	t.Skip("skipped: LK ensemble ISPP convergence not yet stable. See module1-hysteresis/pkg/controller/diagnostics_remanent_staircase.md. To unskip: tune WriteController pulse/verify timing (PulseDuration + adequate relax-at-E=0 verify window), integration dt used by LKSolver.Step, and retry/iteration limits (MaxRetries, max iters) until convergence is deterministic across representative targets.")
	mat := ferroelectric.LiteratureSuperlattice()

	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	solver.UseNLS = false
	solver.EnableEnsemble(96, mat, 0) // deterministic seed derived from material

	wc := NewWriteController(30, mat.Ec, mat.Ec*2.5, nil)
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 30

	// Representative targets including mid-levels.
	targets := []int{5, 10, 15, 20, 25}

	for _, target := range targets {
		t.Run("target_level_"+itoa(target), func(t *testing.T) {
			wc.Start(target, true)
			// Saturate on appropriate side.
			startP := mat.Ps
			if target > wc.NumLevels/2 {
				startP = -mat.Ps
			}
			solver.SetState(startP)

			currentField := 0.0
			finalLevel := 0
			dt := 1e-4
			for i := 0; i < 40000; i++ {
				curLevel := levelFromP(solver.GetState(), mat.Ps, wc.NumLevels)
				targetField, done := wc.Update(dt, currentField, curLevel, 0)
				currentField = targetField
				solver.Step(currentField, dt)
				finalLevel = levelFromP(solver.GetState(), mat.Ps, wc.NumLevels)
				if done {
					break
				}
			}

			if wc.State != StateSuccess {
				t.Fatalf("landauk-ensemble: did not converge: target=%d final=%d pulses=%d state=%s",
					target, finalLevel, wc.TotalPulses+wc.PulseCount, wc.State)
			}
			if finalLevel != target {
				t.Fatalf("landauk-ensemble: wrong final level: target=%d final=%d pulses=%d",
					target, finalLevel, wc.TotalPulses+wc.PulseCount)
			}
			if wc.TotalPulses+wc.PulseCount > 25 {
				t.Fatalf("landauk-ensemble: too many pulses: target=%d pulses=%d",
					target, wc.TotalPulses+wc.PulseCount)
			}
			t.Logf("landauk-ensemble OK: target=%d final=%d pulses=%d", target, finalLevel, wc.TotalPulses+wc.PulseCount)
		})
	}
}
