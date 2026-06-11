package isppconv

import "testing"

func TestEvaluateGuardDecisionUsesLimitedCorrectionBeforeAccepting(t *testing.T) {
	decision := EvaluateGuardDecision(GuardInput{
		LevelError:      0,
		GuardSign:       -3,
		GuardPulseCount: 1,
		MaxGuardPulses:  2,
	})

	if !decision.GuardActive {
		t.Fatal("expected guard correction to be active before pulse limit")
	}
	if decision.Accepted {
		t.Fatal("guard correction before pulse limit must not accept convergence")
	}
	if decision.EffectiveError != -1 {
		t.Fatalf("effective error = %d, want normalized guard sign -1", decision.EffectiveError)
	}
	if decision.NextGuardPulseCount != 2 {
		t.Fatalf("next guard pulse count = %d, want 2", decision.NextGuardPulseCount)
	}
	if decision.Reason == "" {
		t.Fatal("guard decision should include a receipt reason")
	}
}

func TestEvaluateGuardDecisionAcceptsAfterGuardLimit(t *testing.T) {
	decision := EvaluateGuardDecision(GuardInput{
		LevelError:      0,
		GuardSign:       1,
		GuardPulseCount: 2,
		MaxGuardPulses:  2,
	})

	if decision.GuardActive {
		t.Fatal("guard must not stay active after pulse limit")
	}
	if !decision.Accepted {
		t.Fatal("guard limit should accept exact target level as converged")
	}
	if decision.EffectiveError != 0 {
		t.Fatalf("effective error = %d, want exact target error 0", decision.EffectiveError)
	}
	if decision.NextGuardPulseCount != 3 {
		t.Fatalf("next guard pulse count = %d, want 3", decision.NextGuardPulseCount)
	}
}

func TestEvaluateGuardDecisionResetsCountWhenOffTarget(t *testing.T) {
	decision := EvaluateGuardDecision(GuardInput{
		LevelError:      2,
		GuardSign:       1,
		GuardPulseCount: 2,
		MaxGuardPulses:  2,
	})

	if decision.GuardActive || decision.Accepted {
		t.Fatalf("off-target guard decision should not activate or accept: %+v", decision)
	}
	if decision.EffectiveError != 2 {
		t.Fatalf("effective error = %d, want original level error 2", decision.EffectiveError)
	}
	if decision.NextGuardPulseCount != 0 {
		t.Fatalf("next guard pulse count = %d, want reset to 0", decision.NextGuardPulseCount)
	}
}

func TestRecoverCollapsedBoundsWidensDirectionallyWhenMoreFieldNeeded(t *testing.T) {
	bounds := Bounds{Min: 1.20, Max: 1.00, MinSet: true, MaxSet: true}

	recovered := RecoverCollapsedBounds(bounds, RecoveryInput{
		NeedMore:         true,
		CurrentMagnitude: 1.10,
		MaxMagnitude:     2.50,
		MinimumWidth:     0.04,
	})

	if !recovered.Changed {
		t.Fatal("expected collapsed bounds recovery to report a change")
	}
	if recovered.Bounds.Min < 1.10 {
		t.Fatalf("recovered minimum = %.3f, want to preserve at least current safe field 1.10", recovered.Bounds.Min)
	}
	if recovered.Bounds.Max <= recovered.Bounds.Min {
		t.Fatalf("recovered bounds still collapsed: min=%.3f max=%.3f", recovered.Bounds.Min, recovered.Bounds.Max)
	}
	if got := recovered.Bounds.Max - recovered.Bounds.Min; got < 0.04 {
		t.Fatalf("recovered width = %.3f, want at least 0.04", got)
	}
	if recovered.ResetToFullRange {
		t.Fatal("directional recovery must not reset to full range")
	}
}

func TestRecoverCollapsedBoundsResetsToFullRangeWhenDirectionUnknown(t *testing.T) {
	bounds := Bounds{Min: 1.20, Max: 1.00, MinSet: true, MaxSet: true}

	recovered := RecoverCollapsedBounds(bounds, RecoveryInput{
		CurrentMagnitude: 1.10,
		MaxMagnitude:     2.50,
		MinimumWidth:     0.04,
	})

	if !recovered.Changed {
		t.Fatal("expected collapsed bounds recovery to report a change")
	}
	if !recovered.ResetToFullRange {
		t.Fatal("unknown-direction recovery should explicitly report full-range reset")
	}
	if recovered.Bounds.Min != 0 || recovered.Bounds.Max != 2.50 {
		t.Fatalf("recovered bounds = [%.3f, %.3f], want [0, 2.50]", recovered.Bounds.Min, recovered.Bounds.Max)
	}
	if recovered.Bounds.MinSet || recovered.Bounds.MaxSet {
		t.Fatalf("full-range recovery should clear bound evidence flags: %+v", recovered.Bounds)
	}
}
