// Package isppconv contains UI-neutral ISPP convergence policy helpers.
package isppconv

// Bounds describes the established non-negative search interval for pulse
// magnitude. Min/Max are magnitude values in the caller's native units
// (electric field for Module 1, voltage for LK adapters).
type Bounds struct {
	Min    float64
	Max    float64
	MinSet bool
	MaxSet bool
}

// RecoveryInput describes one collapsed-bounds recovery decision.
type RecoveryInput struct {
	NeedMore         bool
	NeedLess         bool
	CurrentMagnitude float64
	MaxMagnitude     float64
	MinimumWidth     float64
}

// RecoveryReceipt reports the recovered bounds and why they changed.
type RecoveryReceipt struct {
	Bounds           Bounds
	Changed          bool
	ResetToFullRange bool
	Reason           string
}

// GuardInput describes one guard-band decision at zero-field verify.
type GuardInput struct {
	LevelError      int
	GuardSign       int
	GuardPulseCount int
	MaxGuardPulses  int
}

// GuardDecision reports whether a guard pulse should alter the effective error
// or whether the exact target level should be accepted as converged.
type GuardDecision struct {
	EffectiveError      int
	NextGuardPulseCount int
	GuardActive         bool
	Accepted            bool
	Reason              string
}

// EvaluateGuardDecision applies the UI-neutral guard-band acceptance rule used
// by ISPP adapters. When the verified level is already the target but a guard
// sign indicates bin-edge risk, the policy allows a limited number of guard
// corrections before accepting convergence to avoid direction flipping.
func EvaluateGuardDecision(input GuardInput) GuardDecision {
	maxGuardPulses := input.MaxGuardPulses
	if maxGuardPulses <= 0 {
		maxGuardPulses = 2
	}

	decision := GuardDecision{
		EffectiveError:      input.LevelError,
		NextGuardPulseCount: input.GuardPulseCount,
	}

	guardSign := 0
	if input.GuardSign > 0 {
		guardSign = 1
	} else if input.GuardSign < 0 {
		guardSign = -1
	}

	if input.LevelError == 0 && guardSign != 0 {
		decision.NextGuardPulseCount++
		if decision.NextGuardPulseCount <= maxGuardPulses {
			decision.EffectiveError = guardSign
			decision.GuardActive = true
			decision.Reason = "guard correction within pulse limit"
			return decision
		}
		decision.Accepted = true
		decision.Reason = "guard pulse limit reached"
		return decision
	}

	if input.LevelError != 0 {
		decision.NextGuardPulseCount = 0
		decision.Reason = "off target resets guard pulse count"
		return decision
	}

	decision.Accepted = true
	decision.Reason = "exact target without guard"
	return decision
}

// RecoverCollapsedBounds widens a collapsed search interval using directional
// evidence when available. It preserves locality for the ISPP guard rule that
// collapsed bounds should be widened minimally, not reset to the full range.
func RecoverCollapsedBounds(bounds Bounds, input RecoveryInput) RecoveryReceipt {
	receipt := RecoveryReceipt{Bounds: bounds}
	if !bounds.MinSet || !bounds.MaxSet || bounds.Min < bounds.Max {
		return receipt
	}

	width := input.MinimumWidth
	if width <= 0 {
		width = 1
	}

	maxMagnitude := input.MaxMagnitude
	if maxMagnitude <= 0 {
		maxMagnitude = bounds.Max
	}

	switch {
	case input.NeedMore:
		min := bounds.Min
		if input.CurrentMagnitude > min {
			min = input.CurrentMagnitude
		}
		max := min + width
		if maxMagnitude > 0 && max > maxMagnitude {
			max = maxMagnitude
			min = max - width
			if min < 0 {
				min = 0
			}
		}
		receipt.Bounds = Bounds{Min: min, Max: max, MinSet: true, MaxSet: true}
		receipt.Changed = true
		receipt.Reason = "directional need-more recovery"
	case input.NeedLess:
		max := bounds.Max
		if input.CurrentMagnitude > 0 && input.CurrentMagnitude < max {
			max = input.CurrentMagnitude
		}
		min := max - width
		if min < 0 {
			min = 0
			max = width
			if maxMagnitude > 0 && max > maxMagnitude {
				max = maxMagnitude
			}
		}
		receipt.Bounds = Bounds{Min: min, Max: max, MinSet: true, MaxSet: true}
		receipt.Changed = true
		receipt.Reason = "directional need-less recovery"
	default:
		receipt.Bounds = Bounds{Min: 0, Max: maxMagnitude, MinSet: false, MaxSet: false}
		receipt.Changed = true
		receipt.ResetToFullRange = true
		receipt.Reason = "unknown direction full-range recovery"
	}
	return receipt
}
