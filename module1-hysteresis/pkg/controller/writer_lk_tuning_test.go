package controller

import (
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

type lkRunStats struct {
	overshoots int
	pulses     int
	finalLevel int
	success    bool
}

func runLKWrite(t *testing.T, targetLevel int, enableOptimizations bool) lkRunStats {
	t.Helper()
	mat := ferroelectric.LiteratureSuperlattice()
	numLevels := 30

	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	solver.UseNLS = false

	wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.EnableLKMidOptimizations = enableOptimizations
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 30
	wc.Start(targetLevel, true)

	startP := mat.Ps
	if targetLevel > numLevels/2 {
		startP = -mat.Ps
	}
	solver.SetState(startP)

	currentField := 0.0
	finalLevel := levelFromP(solver.GetState(), mat.Ps, numLevels)
	const maxIters = 60000
	const dt = 1e-4
	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(solver.GetState(), mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		solver.Step(currentField, dt)
		finalLevel = levelFromP(solver.GetState(), mat.Ps, numLevels)
		if done {
			break
		}
	}

	return lkRunStats{
		overshoots: wc.OvershootTotal + wc.OvershootCount,
		pulses:     wc.TotalPulses + wc.PulseCount,
		finalLevel: finalLevel,
		success:    wc.State == StateSuccess,
	}
}

func TestWriteController_LKMidOptimizationsReduceOvershoot(t *testing.T) {
	target := 15 // MID target where LK overshoot behavior is most problematic
	baseline := runLKWrite(t, target, false)
	tuned := runLKWrite(t, target, true)

	if tuned.overshoots > baseline.overshoots {
		t.Fatalf("overshoots increased with LK tuning: baseline=%d tuned=%d", baseline.overshoots, tuned.overshoots)
	}

	// Keep quality at least as good (same level or better proximity to target).
	baseErr := absInt(baseline.finalLevel - target)
	tunedErr := absInt(tuned.finalLevel - target)
	if tunedErr > baseErr {
		t.Fatalf("final level error regressed: baseline=%d tuned=%d (target=%d)", baseErr, tunedErr, target)
	}

	t.Logf("LK MID tuning: overshoots %d -> %d, pulses %d -> %d, level %d -> %d",
		baseline.overshoots, tuned.overshoots,
		baseline.pulses, tuned.pulses,
		baseline.finalLevel, tuned.finalLevel)
}

func TestWriteController_WaitSettleScaleIsLongerNearMID(t *testing.T) {
	wcMid := NewWriteController(30, 1.0, 2.5, nil)
	wcMid.EnableLKMidOptimizations = true
	wcMid.TargetLevel = 15
	midScale := wcMid.waitSettleScale()

	wcEdge := NewWriteController(30, 1.0, 2.5, nil)
	wcEdge.EnableLKMidOptimizations = true
	wcEdge.TargetLevel = 1
	edgeScale := wcEdge.waitSettleScale()

	if midScale <= edgeScale {
		t.Fatalf("expected longer settle near MID: midScale=%.3f edgeScale=%.3f", midScale, edgeScale)
	}
	if edgeScale < 1.0 {
		t.Fatalf("edge settle scale should not shrink below 1.0, got %.3f", edgeScale)
	}
}
