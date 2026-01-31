package controller

import (
	"log"
	"math"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
)

// WriteState represents the sub-state of the writing process
type WriteState int

const (
	StateIdle WriteState = iota
	StateApply
	StateWait
	StateVerify
	StateHold
	StateSuccess
	StateFailed     // Needs external intervention (e.g. Reset)
	StateForceReset // Explicitly requests a full system reset
)

// WriteController manages the ISPP (Incremental Step Pulse Programming) loop
type WriteController struct {
	// Configuration
	NumLevels       int
	Ec              float64
	Emax            float64
	MaxRetries      int // Max ISPP pulses before giving up
	ForceResetLimit int // Max retries before forcing a full reset
	Attempts        int
	SuccessCount    int
	FailureCount    int
	PulseDuration   float64 // Configuration field for simulation sync

	// Added for search acceleration (Slope estimation)
	previousLevel   int
	previousVoltage float64

	// Dependencies
	CalibManager *algo.CalibrationManager

	// Current Target
	TargetLevel    int
	FromSaturation bool

	// Dynamic State
	State          WriteState
	PhaseTimer     float64
	CurrentVoltage float64 // The target voltage for the current pulse
	PulseCount     int
	TotalPulses    int // Accumulated pulses across retries
	RetryCount     int // Number of times we've restarted the ISPP loop

	// Servo State
	PreviousDiff int     // Difference from target in previous step
	StepModifier float64 // Adaptive multiplier for step size

	// Outputs
	LastVerifyLevel int
	LastError       int
}

func NewWriteController(numLevels int, ec, emax float64, calib *algo.CalibrationManager) *WriteController {
	return &WriteController{
		NumLevels:       numLevels,
		Ec:              ec,
		Emax:            emax,
		MaxRetries:      50,   // Stubborn: Try hard to converge
		ForceResetLimit: 100,  // Effectively disabled, only for total failure
		PulseDuration:   0.15, // Default safe value
		CalibManager:    calib,
		State:           StateIdle,
	}
}

// Start begins a new write operation to the target level
func (wc *WriteController) Start(targetLevel int, fromSaturation bool) {
	wc.TargetLevel = targetLevel
	wc.FromSaturation = fromSaturation
	wc.State = StateApply
	wc.PhaseTimer = 0
	wc.PulseCount = 0
	// Reset Servo State
	wc.PreviousDiff = 0
	wc.StepModifier = 1.0

	// Reset slope estimation state
	wc.previousLevel = -1
	wc.previousVoltage = 0

	wc.calculateNextVoltage(0) // 0 for current level, but will be refined
}

// Reset clears the controller state for a completely new operation
func (wc *WriteController) ResetState() {
	wc.State = StateIdle
	wc.TotalPulses = 0
	wc.RetryCount = 0
	wc.Attempts = 0
	wc.SuccessCount = 0
	wc.FailureCount = 0
}

// Update advances the controller state logic.
func (wc *WriteController) Update(dt float64, currentField float64, currentLevel int) (targetField float64, done bool) {
	wc.PhaseTimer += dt

	// Pulse duration constant (could be configurable)
	pulseDur := wc.PulseDuration

	switch wc.State {
	case StateApply:
		// Target is the pulse voltage
		targetField = wc.CurrentVoltage

		// If we reached the target voltage (approx), switch to WAIT
		if wc.PhaseTimer > pulseDur*0.4 && math.Abs(currentField-wc.CurrentVoltage) < 0.01*wc.Emax {
			wc.State = StateWait
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateWait:
		targetField = wc.CurrentVoltage
		if wc.PhaseTimer > pulseDur*0.3 {
			wc.State = StateVerify // Go to Verify
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateVerify:
		// Target is 0V for verification
		targetField = 0.0

		// Wait for field to settle to 0
		if wc.PhaseTimer > pulseDur*0.3 && math.Abs(currentField) < 0.01*wc.Emax {
			// VERIFY LOGIC
			wc.LastVerifyLevel = currentLevel
			wc.LastError = currentLevel - wc.TargetLevel

			// STRICT CONVERGENCE: Only accept exact match
			if wc.LastError == 0 {
				wc.State = StateSuccess
				wc.SuccessCount++
				// Update calibration (simple average learning)
				// Note: In real FeCIM, we'd be more careful about updating calib from "stubborn" writes
				// as they might represent edges of the distribution.
				return 0, true
			}

			// Not converged. Check retries.
			if wc.PulseCount >= wc.MaxRetries {
				// We tried hard. If we still failed, we might need a BIG reset.
				// But we are "Stubborn", so we only give up if we hit the limit.
				wc.RetryCount++
				wc.FailureCount++

				// Don't resets immediately. Just count it as a "failed cycle" but maybe keep trying?
				// For now, adhere to the "Give Up" logic if MaxRetries hit, but MaxRetries is high (50).
				wc.State = StateFailed
				return 0, true
			}

			// Continue ISPP (Next Pulse)
			wc.PulseCount++
			wc.calculateNextVoltage(currentLevel)
			wc.State = StateApply
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateSuccess, StateFailed, StateForceReset:
		return 0, true

	default:
		return 0, false
	}
}

// calculateNextVoltage determines the voltage for the next pulse
func (wc *WriteController) calculateNextVoltage(currentLevel int) {
	targetLevel := wc.TargetLevel
	goingUp := targetLevel > currentLevel
	levelDiff := int(math.Abs(float64(targetLevel - currentLevel)))
	wrdTargetIdx := targetLevel - 1

	// Initial Pulse Logic
	if wc.PulseCount == 0 {
		var targetVoltage float64

		// Check if we start from saturation (normal path) or mid-state (overshoot recovery)
		if wc.FromSaturation && wrdTargetIdx >= 0 && wrdTargetIdx < len(wc.CalibManager.CalibrationUp) {
			// NORMAL PATH: Starting from saturation, use calibration directly
			if goingUp {
				targetVoltage = wc.CalibManager.CalibrationUp[wrdTargetIdx]
			} else {
				targetVoltage = wc.CalibManager.CalibrationDown[wrdTargetIdx]
			}
			log.Printf("ISPP INIT (SAT): target=%d, calibV=%.3f×Ec", targetLevel, targetVoltage/wc.Ec)
		} else {
			// OVERSHOOT RECOVERY or FIRST ATTEMPT from mid-state
			// Use estimation logic
			// Proportional Step
			// Increased gain from 0.005 to 0.015 to speed up convergence when far from target
			propStep := float64(levelDiff) * wc.Emax * 0.015
			estV := propStep

			if goingUp {
				targetVoltage = estV
			} else {
				targetVoltage = -estV
			}
			log.Printf("ISPP INIT (MID): target=%d, current=%d, estV=%.3f×Ec", targetLevel, currentLevel, targetVoltage/wc.Ec)
		}

		wc.CurrentVoltage = targetVoltage
		return
	}

	// SUBSEQUENT PULSES: Stubborn Servo Logic
	// We are adjusting 'wc.CurrentVoltage' based on the error.

	// 1. Calculate Error
	diff := currentLevel - targetLevel // +ve means READ > TARGET (Overshoot) -> Needs Negative Nudge
	levelError := diff                 // Renamed for clarity in new logic

	// 2. Servo Logic: Detect Oscillation
	// If sign of diff flipped compared to previous, we overshot the target in the servo loop.
	// Dampen the step size.
	// If sign matches, we are approaching or stuck. Maintain or accelerate.

	signChanged := false
	if (diff > 0 && wc.PreviousDiff < 0) || (diff < 0 && wc.PreviousDiff > 0) {
		signChanged = true
	}

	// SIGN CHANGED - Dampen or Binary Search logic
	if signChanged {
		wc.StepModifier *= 0.5 // Standard dampening
		if wc.StepModifier < 0.1 {
			wc.StepModifier = 0.1
		}
		// Reset tracking on flip
		wc.previousVoltage = 0
		wc.previousLevel = -1
	} else {
		// NO SIGN CHANGE - We haven't crossed target yet.
		// If we made NO progress (level read same), kick it harder
		if diff == wc.PreviousDiff {
			// AGGRESSIVE KICK when stuck in sub-threshold or plateau
			// If field is weak (<0.8*Ec), jump immediately to switching region
			EcEst := wc.Emax * 0.4
			if math.Abs(wc.CurrentVoltage) < 0.8*EcEst && math.Abs(float64(levelError)) > 2 {
				if levelError > 0 {
					wc.CurrentVoltage = -0.9 * EcEst // Kick negative
				} else {
					wc.CurrentVoltage = 0.9 * EcEst // Kick positive
				}
				log.Printf("ISPP KICK: V=%.3f", wc.CurrentVoltage)
			} else {
				wc.StepModifier *= 1.5 // Standard stuck recovery
			}
		} else {
			// WE MADE PROGRESS - but are we moving fast enough?
			// Use slope estimation if we have two data points on the same branch
			if wc.previousVoltage != 0 && currentLevel != wc.previousLevel {
				slope := (wc.CurrentVoltage - wc.previousVoltage) / float64(currentLevel-wc.previousLevel)
				if (levelError > 0 && slope < 0) || (levelError < 0 && slope > 0) {
					// Predict voltage needed to close remaining level error
					estDeltaV := slope * float64(-levelError)
					// Use a blend of current and estimate to avoid overshoot
					wc.CurrentVoltage += estDeltaV * 0.8
					log.Printf("ISPP SLOPE ESTIMATE: newV=%.3f", wc.CurrentVoltage)
					// Reset tracking to avoid double-stepping
					wc.previousVoltage = 0
					wc.previousLevel = -1
					return
				}
			}
			wc.StepModifier = 1.0
		}
	}

	// Update tracking for next slope calculation
	wc.previousLevel = currentLevel
	wc.previousVoltage = wc.CurrentVoltage

	// 3. Calculate Nudge
	voltagePerLevel := (2.0 * wc.Ec) / float64(wc.NumLevels-1)
	baseStep := voltagePerLevel * 0.15 // Finer base nudge unit (reduced from 0.5)

	// Scale nudge by error, but with diminishing returns for large errors
	errorScale := math.Min(float64(math.Abs(float64(diff))), 3.0) // Cap at 3 levels
	nudge := baseStep * errorScale * wc.StepModifier

	// Direction
	if diff > 0 {
		// Read > Target => Go DOWN => Subtract voltage (or make more negative)
		wc.CurrentVoltage -= nudge
	} else {
		// Read < Target => Go UP => Add voltage
		wc.CurrentVoltage += nudge
	}

	// 4. Update State
	wc.PreviousDiff = diff

	// Clamp
	if wc.CurrentVoltage > 2.0*wc.Ec {
		wc.CurrentVoltage = 2.0 * wc.Ec
	} else if wc.CurrentVoltage < -2.0*wc.Ec {
		wc.CurrentVoltage = -2.0 * wc.Ec
	}

	log.Printf("ISPP SERVO: trg=%d, read=%d, err=%+d, mod=%.2f, newV=%.3f×Ec",
		targetLevel, currentLevel, diff, wc.StepModifier, wc.CurrentVoltage/wc.Ec)
}
