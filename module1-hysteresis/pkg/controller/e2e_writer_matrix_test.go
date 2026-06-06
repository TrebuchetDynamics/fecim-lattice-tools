package controller

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
)

func TestModule1ControllerE2EWideISPPConvergenceMatrix(t *testing.T) {
	tests := []struct {
		name           string
		numLevels      int
		target         int
		initial        int
		fromSaturation bool
		stepMode       string
		guardSign      int
	}{
		{name: "low-to-mid-log", numLevels: 30, target: 15, initial: 2, stepMode: "logarithmic"},
		{name: "high-to-low-log", numLevels: 30, target: 4, initial: 28, fromSaturation: true, stepMode: "logarithmic"},
		{name: "mid-to-high-linear", numLevels: 32, target: 29, initial: 16, stepMode: "linear"},
		{name: "guarded-target-accept", numLevels: 16, target: 8, initial: 8, stepMode: "logarithmic", guardSign: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ec := 1.2e8
			wc := NewWriteController(tc.numLevels, ec, 2.5*ec, algo.NewCalibrationManager(tc.numLevels))
			wc.StepMode = tc.stepMode
			wc.MaxRetries = 4
			wc.ForceResetLimit = 2
			wc.Start(tc.target, tc.fromSaturation)

			result := runModule1ControllerE2EWrite(wc, tc.initial, tc.target, tc.guardSign)
			if !result.done || wc.State != StateSuccess {
				t.Fatalf("write did not converge: result=%+v state=%s lastLevel=%d lastErr=%d pulses=%d", result, wc.State, wc.LastVerifyLevel, wc.LastError, wc.TotalPulses)
			}
			if wc.TargetLevel != tc.target || wc.InitialLevel != tc.initial || !wc.InitialLevelSet {
				t.Fatalf("controller target/initial tracking = target %d initial %d set=%v, want %d/%d", wc.TargetLevel, wc.InitialLevel, wc.InitialLevelSet, tc.target, tc.initial)
			}
			if wc.BestAbsError != 0 || wc.BestVerifyLevel != tc.target || wc.LastVerifyLevel != tc.target || wc.LastError != 0 {
				t.Fatalf("verify tracking = best L%d err %d last L%d err %d, want target %d", wc.BestVerifyLevel, wc.BestAbsError, wc.LastVerifyLevel, wc.LastError, tc.target)
			}
			if wc.SuccessCount != 1 || wc.FailureCount != 0 || wc.CumulativeFluence < 0 || math.IsNaN(wc.CumulativeFluence) {
				t.Fatalf("counters/fluence invalid: success=%d failure=%d fluence=%g", wc.SuccessCount, wc.FailureCount, wc.CumulativeFluence)
			}
			if tc.initial != tc.target && wc.CumulativeFluence == 0 {
				t.Fatalf("non-trivial write used zero programming fluence")
			}
			wc.ResetState()
			if wc.State != StateIdle || wc.TotalPulses != 0 || wc.RetryCount != 0 || wc.SuccessCount != 0 || wc.FailureCount != 0 {
				t.Fatalf("ResetState left counters/state dirty: state=%s total=%d retry=%d success=%d failure=%d", wc.State, wc.TotalPulses, wc.RetryCount, wc.SuccessCount, wc.FailureCount)
			}
		})
	}
}

func TestModule1ControllerE2EOvershootRecoveryAndFailureBoundaries(t *testing.T) {
	ec := 1e8
	wc := NewWriteController(30, ec, 2*ec, algo.NewCalibrationManager(30))
	wc.OvershootLimit = 2
	wc.Start(10, false)
	level := 2
	field := 0.0
	seenReset := false
	for i := 0; i < 80; i++ {
		if wc.State == StateVerify {
			field = 0
			level = 25 // persistent overshoot beyond target
		} else if wc.CurrentField != 0 {
			field = wc.CurrentField
		}
		target, done := wc.Update(0.1, field, level, 0)
		if wc.State == StateResetting {
			seenReset = true
		}
		field = target
		if done {
			break
		}
	}
	if !seenReset && wc.OvershootTotal == 0 {
		t.Fatalf("persistent overshoot did not exercise reset/overshoot accounting: state=%s overshootTotal=%d", wc.State, wc.OvershootTotal)
	}
	if wc.MaxOvershootDelta <= 0 || wc.LastVerifyLevel <= wc.TargetLevel {
		t.Fatalf("overshoot tracking missing: maxDelta=%d lastLevel=%d target=%d", wc.MaxOvershootDelta, wc.LastVerifyLevel, wc.TargetLevel)
	}

	wc.ResetState()
	wc.MaxRetries = 1
	wc.ForceResetLimit = 1
	wc.Start(29, false)
	if wc.ResetDirection() != 0 {
		t.Fatalf("new operation reset direction = %d, want 0", wc.ResetDirection())
	}
}

func TestModule1ControllerE2EEarlySuccessUsesNoPulse(t *testing.T) {
	wc := NewWriteController(30, 1e8, 2e8, algo.NewCalibrationManager(30))
	wc.Start(12, false)
	targetField, done := wc.Update(0.01, 0, 12, 0)
	if !done || targetField != 0 || wc.State != StateSuccess || wc.TotalPulses != 0 || wc.PulseCount != 0 {
		t.Fatalf("early success = targetField %.3e done=%v state=%s total=%d pulse=%d", targetField, done, wc.State, wc.TotalPulses, wc.PulseCount)
	}
}

type module1ControllerE2EResult struct {
	done      bool
	steps     int
	lastField float64
	lastLevel int
}

func runModule1ControllerE2EWrite(wc *WriteController, initialLevel, targetLevel, guardSign int) module1ControllerE2EResult {
	level := initialLevel
	field := 0.0
	for step := 0; step < 160; step++ {
		if wc.State == StateVerify {
			field = 0
			level = targetLevel
		} else if wc.CurrentField != 0 {
			field = wc.CurrentField
		}
		targetField, done := wc.Update(0.1, field, level, guardSign)
		if done {
			return module1ControllerE2EResult{done: true, steps: step + 1, lastField: targetField, lastLevel: level}
		}
		field = targetField
	}
	return module1ControllerE2EResult{done: false, steps: 160, lastField: field, lastLevel: level}
}
