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
	StateResetting  // Internal reset due to overshoot
)

func (s WriteState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateApply:
		return "APPLY"
	case StateWait:
		return "WAIT"
	case StateVerify:
		return "VERIFY"
	case StateHold:
		return "HOLD"
	case StateSuccess:
		return "SUCCESS"
	case StateFailed:
		return "FAILED"
	case StateForceReset:
		return "FORCE_RESET"
	case StateResetting:
		return "RESETTING"
	default:
		return "UNKNOWN"
	}
}

// WriteController manages the ISPP (Incremental Step Pulse Programming) loop
type WriteController struct {
	// Configuration
	NumLevels       int
	EcField         float64
	MaxField        float64
	MaxRetries      int // Max ISPP pulses before giving up
	ForceResetLimit int // Max retries before forcing a full reset
	Attempts        int
	SuccessCount    int
	FailureCount    int
	PulseDuration   float64 // Configuration field for simulation sync

	// Added for search acceleration (Slope estimation)
	previousLevel int
	previousField float64

	// Dependencies
	CalibManager *algo.CalibrationManager

	// Current Target
	TargetLevel    int
	FromSaturation bool

	// Dynamic State
	State        WriteState
	PhaseTimer   float64
	CurrentField float64 // The target field for the current pulse
	PulseCount   int
	TotalPulses  int // Accumulated pulses across retries
	RetryCount   int // Number of times we've restarted the ISPP loop

	InitialLevel    int
	InitialLevelSet bool

	// Overshoot tracking (for autonomous recalibration)
	OvershootCount int // Overshoots in current target cycle
	OvershootTotal int // Overshoots across runtime
	resetDirection int // Sticky reset direction after overshoot (-1 or +1)

	// Binary Search State (for ISPP)
	VMin float64 // Lower bound of safe voltage (won't overshoot)
	VMax float64 // Upper bound of voltage search space

	// Outputs
	LastVerifyLevel int
	LastError       int
}

func NewWriteController(numLevels int, ec, emax float64, calib *algo.CalibrationManager) *WriteController {
	return &WriteController{
		NumLevels:       numLevels,
		EcField:         ec,
		MaxField:        emax,
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
	wc.OvershootCount = 0
	wc.CurrentField = 0
	wc.LastVerifyLevel = 0
	wc.LastError = 0
	wc.previousLevel = 0
	wc.previousField = 0
	wc.resetDirection = 0

	// Initialize Binary Search Bounds
	wc.VMin = 0
	wc.VMax = wc.MaxField

	wc.InitialLevel = 0 // Will be captured in Update
	wc.InitialLevelSet = false
	// CurrentField will be set when calculateNextField is called during first Update

	log.Printf("ISPP START: target=%d fromSaturation=%v Ec=%.3g MaxField=%.3g bounds=[%.3f, %.3f]×Ec",
		wc.TargetLevel, wc.FromSaturation, wc.EcField, wc.MaxField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
}

// Reset clears the controller state for a completely new operation
func (wc *WriteController) ResetState() {
	wc.State = StateIdle
	wc.TotalPulses = 0
	wc.RetryCount = 0
	wc.Attempts = 0
	wc.SuccessCount = 0
	wc.FailureCount = 0
	wc.OvershootCount = 0
	wc.OvershootTotal = 0
}

// Update advances the controller state logic.
func (wc *WriteController) Update(dt float64, currentField float64, currentLevel int) (targetField float64, done bool) {
	wc.PhaseTimer += dt

	// Capture initial level on first update
	if !wc.InitialLevelSet {
		wc.InitialLevel = currentLevel
		wc.InitialLevelSet = true
		wc.LastVerifyLevel = currentLevel
		wc.previousLevel = currentLevel
	}

	// Pulse duration constant (could be configurable)
	pulseDur := wc.PulseDuration

	switch wc.State {
	case StateApply:
		// If we're already at target, skip pulses entirely.
		if wc.PulseCount == 0 && currentLevel == wc.TargetLevel {
			wc.LastVerifyLevel = currentLevel
			wc.LastError = 0
			wc.State = StateSuccess
			wc.SuccessCount++
			log.Printf("ISPP EARLY: already at target level %d, no pulse needed", currentLevel)
			return 0, true
		}

		// Calculate field for first pulse if not done yet
		if wc.PulseCount == 0 && wc.CurrentField == 0 {
			wc.calculateNextField(currentLevel)
		}

		// Target is the pulse field
		targetField = wc.CurrentField
		if wc.PhaseTimer <= dt {
			log.Printf("ISPP APPLY: pulse=%d currentLevel=%d targetLevel=%d pulseDir=%d E=%.3f×Ec bounds=[%.3f, %.3f]×Ec",
				wc.PulseCount+1, currentLevel, wc.TargetLevel, pulseDirection(wc.CurrentField),
				wc.CurrentField/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
		}

		// If we reached the target field (approx), switch to WAIT
		if wc.PhaseTimer > pulseDur*0.4 && math.Abs(currentField-wc.CurrentField) < 0.01*wc.MaxField {
			wc.State = StateWait
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateWait:
		targetField = wc.CurrentField
		if wc.PhaseTimer <= dt {
			log.Printf("ISPP WAIT: holding E=%.3f×Ec for verify (pulse=%d)", wc.CurrentField/wc.EcField, wc.PulseCount+1)
		}
		if wc.PhaseTimer > pulseDur*0.3 {
			wc.State = StateVerify // Go to Verify
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateResetting:
		// Determine reset polarity based on direction
		// If we were going UP and overshot, we are stuck High. Reset Low (-Max).
		// If we were going DOWN and overshot, we are stuck Low. Reset High (+Max).
		resetDir := wc.resetDirection
		if resetDir == 0 {
			resetDir = wc.directionToTarget(currentLevel)
			if resetDir == 0 {
				// Fall back to last pulse direction if we're exactly on target.
				if pulseDirection(wc.CurrentField) >= 0 {
					resetDir = -1
				} else {
					resetDir = 1
				}
			}
		}
		if resetDir < 0 {
			targetField = -wc.MaxField * 1.5 // Deep Negative
		} else {
			targetField = wc.MaxField * 1.5 // Deep Positive
		}
		if wc.PhaseTimer <= dt {
			log.Printf("ISPP RESET: resetDir=%d targetField=%.3f×Ec (pulse=%d)", resetDir, targetField/wc.EcField, wc.PulseCount+1)
		}
		if wc.PhaseTimer > pulseDur*0.5 && math.Abs(currentField-targetField) >= 0.01*wc.MaxField {
			log.Printf("ISPP RESETTING: currentField=%.3f×Ec targetField=%.3f×Ec phase=%.3f",
				currentField/wc.EcField, targetField/wc.EcField, wc.PhaseTimer)
		}

		// Wait for reset pulse to actually REACH the target (critical for ramp speed)
		if wc.PhaseTimer > pulseDur*0.8 && math.Abs(currentField-targetField) < 0.01*wc.MaxField {
			wc.State = StateApply
			wc.PhaseTimer = 0

			// BINARY SEARCH BOUNDS UPDATE
			// The pulse that caused overshoot was too strong
			// Set VMax to that failing voltage, VMin stays at 0 (or previous VMin if it exists)
			// Next try: VPulse = (VMin + VMax) / 2
			failedVoltage := math.Abs(wc.CurrentField)
			wc.VMax = failedVoltage
			wc.VMin = 0 // Reset to safe baseline
			wc.CurrentField = (wc.VMin + wc.VMax) / 2.0

			// Apply sign based on next target direction (not reset direction).
			nextDir := wc.directionToTarget(currentLevel)
			if nextDir == 0 {
				nextDir = -resetDir
			}
			if nextDir < 0 {
				wc.CurrentField = -wc.CurrentField
			}

			// CRITICAL: Reset InitialLevel to current level after reset
			// This ensures direction logic is consistent between calculateNextField and overshoot detection
			wc.InitialLevel = currentLevel

			wc.PulseCount++
			wc.resetDirection = 0
			log.Printf("ISPP RESET DONE. Binary search: VMax=%.3f×Ec (failed), VMin=%.3f×Ec, trying E=%.3f×Ec",
				wc.VMax/wc.EcField, wc.VMin/wc.EcField, wc.CurrentField/wc.EcField)
		}
		return targetField, false

	case StateVerify:
		// Target is 0V for verification
		targetField = 0.0

		// Wait for field to settle to 0
		if wc.PhaseTimer > pulseDur*0.3 && math.Abs(currentField) < 0.01*wc.MaxField {
			// VERIFY LOGIC
			prevLevel := wc.LastVerifyLevel
			if prevLevel == 0 {
				prevLevel = currentLevel
			}
			log.Printf("ISPP READ: currentLevel=%d targetLevel=%d prevLevel=%d currentField=%.3f×Ec",
				currentLevel, wc.TargetLevel, prevLevel, currentField/wc.EcField)
			wc.LastError = currentLevel - wc.TargetLevel

			// STRICT CONVERGENCE: Only accept exact match
			if wc.LastError == 0 {
				log.Printf("ISPP VERIFY RESULT: hit target level %d", currentLevel)
				wc.LastVerifyLevel = currentLevel
				wc.State = StateSuccess
				wc.SuccessCount++
				// Update calibration (simple average learning)
				// Note: In real FeCIM, we'd be more careful about updating calib from "stubborn" writes
				// as they might represent edges of the distribution.
				return 0, true
			}

			// Check for OVERSHOOT
			// If we passed the target, we are on the wrong hysteresis branch.
			// Standard servoing (reducing voltage) won't work due to remanence.
			// Must RESET.
			pulseDir := pulseDirection(wc.CurrentField)
			overshoot := false
			if pulseDir > 0 {
				overshoot = prevLevel <= wc.TargetLevel && currentLevel > wc.TargetLevel
			} else if pulseDir < 0 {
				overshoot = prevLevel >= wc.TargetLevel && currentLevel < wc.TargetLevel
			}

			// DEBUG: Always log the overshoot check
			log.Printf("ISPP VERIFY: currentLevel=%d, targetLevel=%d, prevLevel=%d, pulseDir=%d, overshoot=%v",
				currentLevel, wc.TargetLevel, prevLevel, pulseDir, overshoot)
			wc.LastVerifyLevel = currentLevel
			log.Printf("ISPP VERIFY RESULT: error=%d bounds=[%.3f, %.3f]×Ec",
				wc.LastError, wc.VMin/wc.EcField, wc.VMax/wc.EcField)

			if overshoot {
				wc.OvershootCount++
				wc.OvershootTotal++
				if pulseDir != 0 {
					wc.resetDirection = -pulseDir
				} else {
					wc.resetDirection = -wc.directionToTarget(currentLevel)
				}
				log.Printf("ISPP OVERSHOOT detected! Resetting state... (count=%d)", wc.OvershootCount)
				wc.State = StateResetting
				wc.PhaseTimer = 0
				return 0, false
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
			wc.calculateNextField(currentLevel)
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

// calculateNextField determines the field for the next pulse using binary search
func (wc *WriteController) calculateNextField(currentLevel int) {
	targetLevel := wc.TargetLevel
	direction := wc.directionToTarget(currentLevel)

	// Determine which saturation state we're writing from
	atNegativeSat := currentLevel <= 3                  // Near Level 1 (-Pr)
	atPositiveSat := currentLevel >= (wc.NumLevels - 2) // Near Level 30 (+Pr)

	// Initial Pulse: Use calibration as starting point
	if wc.PulseCount == 0 {
		wrdTargetIdx := targetLevel - 1
		var initialGuess float64

		// Use calibration if available and we're at saturation
		if wc.FromSaturation && wc.CalibManager != nil && wrdTargetIdx >= 0 && wrdTargetIdx < len(wc.CalibManager.CalibrationUp) {
			if atNegativeSat {
				// Starting from -Pr, use CalibrationUp
				initialGuess = math.Abs(wc.CalibManager.CalibrationUp[wrdTargetIdx])
				log.Printf("ISPP INIT: target=%d, calib=%.3f×Ec (from -Pr), bounds=[%.3f, %.3f]×Ec",
					targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
			} else if atPositiveSat {
				// Starting from +Pr, use CalibrationDown (already negative)
				initialGuess = math.Abs(wc.CalibManager.CalibrationDown[wrdTargetIdx])
				log.Printf("ISPP INIT: target=%d, calib=%.3f×Ec (from +Pr), bounds=[%.3f, %.3f]×Ec",
					targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
			} else {
				// Not at saturation - take a single Ec step toward target
				initialGuess = math.Abs(wc.EcField)
				log.Printf("ISPP INIT: target=%d, Ec-step=%.3f×Ec (not at sat), bounds=[%.3f, %.3f]×Ec",
					targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
			}
		} else {
			// No calibration or not from saturation - take a single Ec step toward target
			initialGuess = math.Abs(wc.EcField)
			log.Printf("ISPP INIT: target=%d, Ec-step=%.3f×Ec (no calib), bounds=[%.3f, %.3f]×Ec",
				targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
		}

		if direction == 0 {
			wc.CurrentField = 0
			log.Printf("ISPP INIT: already at target level %d, no pulse scheduled", currentLevel)
			return
		}

		if direction < 0 {
			initialGuess = -initialGuess
		}

		wc.CurrentField = initialGuess
		return
	}

	// SUBSEQUENT PULSES: Binary Search Logic
	// error = currentLevel - targetLevel
	// error > 0: We're ABOVE target (overshoot if we were going up, OR undershoot if going down)
	// error < 0: We're BELOW target (undershoot if we were going up, OR overshoot if going down)

	error := currentLevel - targetLevel

	if error == 0 {
		// Perfect hit! (shouldn't reach here as StateVerify handles this)
		log.Printf("ISPP: Perfect hit at level %d", currentLevel)
		return
	}

	// Determine if we undershot or made progress
	// Case 1: Undershoot - field too weak, need stronger field
	//   - Going UP (target > initial): currentLevel < targetLevel → error < 0
	//   - Going DOWN (target < initial): currentLevel > targetLevel → error > 0
	// Case 2: We're making progress in the right direction

	// Update binary search bounds
	absCurrentField := math.Abs(wc.CurrentField)

	if direction > 0 {
		// We're BELOW target, need stronger positive field
		wc.VMin = absCurrentField
		log.Printf("ISPP BINARY: Undershoot (level=%d < target=%d), VMin=%.3f×Ec → VMin=%.3f×Ec",
			currentLevel, targetLevel, wc.VMin/wc.EcField, absCurrentField/wc.EcField)
	} else if direction < 0 {
		// We're ABOVE target, need stronger negative field
		wc.VMin = absCurrentField
		log.Printf("ISPP BINARY: Undershoot (level=%d > target=%d), VMin=%.3f×Ec → VMin=%.3f×Ec",
			currentLevel, targetLevel, wc.VMin/wc.EcField, absCurrentField/wc.EcField)
	} else {
		// Already at target; nothing to do.
		wc.CurrentField = 0
		log.Printf("ISPP BINARY: already at target level %d, no pulse scheduled", currentLevel)
		return
	}

	// Calculate next voltage as midpoint of bounds
	if wc.VMax < wc.VMin {
		wc.VMax = wc.VMin
	}
	nextVoltage := (wc.VMin + wc.VMax) / 2.0

	// Apply sign based on direction
	if direction < 0 {
		nextVoltage = -nextVoltage
	}

	wc.CurrentField = nextVoltage
	log.Printf("ISPP BINARY: Next pulse E=%.3f×Ec, bounds=[%.3f, %.3f]×E c",
		wc.CurrentField/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
}

func (wc *WriteController) directionToTarget(currentLevel int) int {
	if currentLevel < wc.TargetLevel {
		return 1
	}
	if currentLevel > wc.TargetLevel {
		return -1
	}
	return 0
}

func pulseDirection(field float64) int {
	if field > 0 {
		return 1
	}
	if field < 0 {
		return -1
	}
	return 0
}
