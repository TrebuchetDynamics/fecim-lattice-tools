package physics

import "math"

// AdaptiveISPP implements a simple write-verify loop for reaching a target
// polarization using single-pulse evaluation and voltage bracketing.
//
// Each candidate voltage is evaluated from the same pre-pulse baseline
// polarization (the solver state at entry). This avoids trying to do a binary
// search while applying pulses cumulatively (which breaks monotonicity).
//
// NOTE: This is a pedagogical controller; for richer ISPP diagnostics, use
// WriteController in ispp_write.go.
type AdaptiveISPP struct {
	Solver   *LKSolver
	Material *HZOMaterial

	MaxVoltage      float64
	MinVoltage      float64
	TargetTolerance float64 // fraction of Ps
	MaxIterations   int
	PulseWidth      float64 // seconds
}

func NewAdaptiveISPP(solver *LKSolver, mat *HZOMaterial) *AdaptiveISPP {
	pw := 0.0
	if mat != nil {
		pw = mat.Tau
	}
	if pw <= 0 {
		pw = 10e-9
	}
	return &AdaptiveISPP{
		Solver:          solver,
		Material:        mat,
		MaxVoltage:      3.0,
		MinVoltage:      -3.0,
		TargetTolerance: 0.01,
		MaxIterations:   12,
		PulseWidth:      pw,
	}
}

// PredictState estimates an initial voltage guess using a monotonic atanh model:
//
//	P ≈ Ps*tanh(E/Ec)  →  E ≈ Ec*atanh(P/Ps)  →  V ≈ E*d
func (c *AdaptiveISPP) PredictState(targetP float64) float64 {
	if c == nil || c.Material == nil {
		return 0
	}
	Ps := c.Material.Ps
	if Ps == 0 || c.Material.Ec == 0 || c.Material.Thickness == 0 {
		return 0
	}
	ratio := targetP / Ps
	if ratio > 0.95 {
		ratio = 0.95
	}
	if ratio < -0.95 {
		ratio = -0.95
	}
	eEst := c.Material.Ec * math.Atanh(ratio)
	return eEst * c.Material.Thickness
}

func (c *AdaptiveISPP) BinarySearchWrite(targetP float64) (float64, int, bool) {
	if c == nil || c.Solver == nil || c.Material == nil {
		return 0, 0, false
	}
	if c.Material.Ps == 0 || c.Material.Thickness == 0 || c.PulseWidth <= 0 {
		return c.Solver.GetState(), 0, false
	}

	baseP := c.Solver.GetState()

	vMin := c.MinVoltage
	vMax := c.MaxVoltage
	if vMin > vMax {
		vMin, vMax = vMax, vMin
	}
	// Pick a voltage polarity consistent with the direction we need to move.
	if targetP >= baseP {
		if vMin < 0 {
			vMin = 0
		}
	} else {
		if vMax > 0 {
			vMax = 0
		}
	}

	vNext := c.PredictState(targetP)
	if vNext < vMin {
		vNext = vMin
	}
	if vNext > vMax {
		vNext = vMax
	}

	tolP := c.TargetTolerance * math.Abs(c.Material.Ps)
	if tolP <= 0 {
		tolP = 1e-6
	}

	bestP := baseP
	bestErr := math.Abs(bestP - targetP)

	for iter := 1; iter <= c.MaxIterations; iter++ {
		c.Solver.SetState(baseP)
		eField := vNext / c.Material.Thickness
		c.Solver.Step(eField, c.PulseWidth)
		p := c.Solver.GetState()
		err := p - targetP
		absErr := math.Abs(err)

		if absErr < bestErr {
			bestErr = absErr
			bestP = p
		}
		if absErr <= tolP {
			return p, iter, true
		}

		if err < 0 {
			vMin = vNext
		} else {
			vMax = vNext
		}
		vNext = 0.5 * (vMin + vMax)
	}

	c.Solver.SetState(bestP)
	return bestP, c.MaxIterations, bestErr <= tolP
}

// -----------------------------------------------------------------------------
// Level-based ISPP compatibility engine (shared single source of truth)
// -----------------------------------------------------------------------------

// Legacy compatibility for other modules (e.g. Module 4 Circuits)

// HysteresisDirection represents the direction of the field sweep.
type HysteresisDirection int

const (
	DirectionUnknown HysteresisDirection = iota
	DirectionAscending
	DirectionDescending
)

func (d HysteresisDirection) String() string {
	switch d {
	case DirectionAscending:
		return "Ascending"
	case DirectionDescending:
		return "Descending"
	default:
		return "Unknown"
	}
}

// GetDirection determines the sweep direction based on current and target levels.
func GetDirection(currentLevel, targetLevel int) HysteresisDirection {
	if currentLevel < targetLevel {
		return DirectionAscending
	} else if currentLevel > targetLevel {
		return DirectionDescending
	}
	return DirectionUnknown
}

// ISPPConfig holds configuration for the ISPP algorithm.
//
// NOTE: StartRatio/StepPercent/SafetyCap/MaxPulses are heuristic defaults for
// simulation stability. They represent standard ISPP write-verify behavior
// (increment + verify + clamp), but are not tied to one foundry-qualified
// FeFET PDK.
type ISPPConfig struct {
	StartRatio  float64 // Ratio of calibrated voltage to start at (default 0.7; heuristic)
	StepPercent float64 // Step size as percentage of Ec (default 0.01; heuristic)
	MaxPulses   int     // Max pulses before stopping (default 40; heuristic)
	SafetyCap   float64 // Maximum voltage multiplier relative to Ec (default 2.2; heuristic)
	Tolerance   int     // Allowed level error (default 0 for exact match)
}

// DefaultISPPConfig returns default configuration values.
func DefaultISPPConfig() ISPPConfig {
	return ISPPConfig{
		StartRatio:  0.7,
		StepPercent: 0.01,
		MaxPulses:   40,
		SafetyCap:   2.2,
		Tolerance:   0,
	}
}

// ISPPCalculator handles the logic for Incremental Step Pulse Programming.
type ISPPCalculator struct {
	Ec        float64 // Coercive Field (V)
	NumLevels int     // Total number of levels
	Config    ISPPConfig
}

// NewISPPCalculator creates a new calculator with default config.
func NewISPPCalculator(ec float64, numLevels int) *ISPPCalculator {
	return &ISPPCalculator{
		Ec:        ec,
		NumLevels: numLevels,
		Config:    DefaultISPPConfig(),
	}
}

// NewISPPCalculatorWithConfig creates a new calculator with custom config.
func NewISPPCalculatorWithConfig(ec float64, numLevels int, config ISPPConfig) *ISPPCalculator {
	return &ISPPCalculator{
		Ec:        ec,
		NumLevels: numLevels,
		Config:    config,
	}
}

// CalculateStartVoltage returns the initial programming voltage based on calibration.
func (c *ISPPCalculator) CalculateStartVoltage(calibratedVoltage float64) float64 {
	return calibratedVoltage * c.Config.StartRatio
}

// CalculateVoltageStep returns the voltage increment.
func (c *ISPPCalculator) CalculateVoltageStep() float64 {
	return c.Ec * c.Config.StepPercent
}

// CalculateNextVoltage computes the next voltage in the sequence.
func (c *ISPPCalculator) CalculateNextVoltage(currentVoltage float64, direction HysteresisDirection) float64 {
	step := c.CalculateVoltageStep()
	if direction == DirectionDescending {
		return currentVoltage - step // More negative
	}
	return currentVoltage + step // More positive
}

// ClampVoltage ensures the voltage stays within safe limits.
func (c *ISPPCalculator) ClampVoltage(voltage float64, direction HysteresisDirection) float64 {
	limit := c.Ec * c.Config.SafetyCap
	if direction == DirectionAscending {
		if voltage > limit {
			return limit
		}
		if voltage < 0 {
			return 0
		}
	} else {
		if voltage < -limit {
			return -limit
		}
		if voltage > 0 {
			return 0
		}
	}
	return voltage
}

// ISPPResult represents the outcome of a verify step.
type ISPPResult int

const (
	ISPPContinue ISPPResult = iota
	ISPPSuccess
	ISPPOvershoot
	ISPPMaxPulses
)

func (r ISPPResult) String() string {
	switch r {
	case ISPPSuccess:
		return "Success"
	case ISPPOvershoot:
		return "Overshoot"
	case ISPPMaxPulses:
		return "MaxPulses"
	default:
		return "Continue"
	}
}

// CheckResult evaluates the verify result against the target.
func (c *ISPPCalculator) CheckResult(currentLevel, targetLevel int, direction HysteresisDirection, pulseCount int) ISPPResult {
	if pulseCount >= c.Config.MaxPulses {
		return ISPPMaxPulses
	}

	diff := currentLevel - targetLevel
	// Tolerance check
	if int(math.Abs(float64(diff))) <= c.Config.Tolerance {
		return ISPPSuccess
	}

	if direction == DirectionAscending {
		if diff > 0 { // current > target
			return ISPPOvershoot
		}
	} else {
		if diff < 0 { // current < target (more negative than target)
			return ISPPOvershoot
		}
	}

	return ISPPContinue
}

// GetResetVoltage returns the voltage to reset to saturation (for overshoot recovery).
func (c *ISPPCalculator) GetResetVoltage(direction HysteresisDirection) float64 {
	limit := c.Ec * c.Config.SafetyCap
	if direction == DirectionAscending {
		return -limit // Reset to negative saturation
	} else if direction == DirectionDescending {
		return limit // Reset to positive saturation
	}
	return 0
}

// GetSaturationVoltage returns max voltage for the direction.
func (c *ISPPCalculator) GetSaturationVoltage(direction HysteresisDirection) float64 {
	limit := c.Ec * c.Config.SafetyCap
	if direction == DirectionAscending {
		return limit
	} else if direction == DirectionDescending {
		return -limit
	}
	return 0
}

// EstimatePulsesNeeded estimates how many pulses might be needed (heuristic).
func (c *ISPPCalculator) EstimatePulsesNeeded(currentLevel, targetLevel int, voltage float64) int {
	if currentLevel == targetLevel {
		return 1
	}
	// Simple heuristic
	return int(math.Abs(float64(targetLevel-currentLevel))) + 1
}

// LevelError calculates signed error.
func LevelError(current, target int) int {
	return current - target
}

// IsOvershoot check.
func (c *ISPPCalculator) IsOvershoot(current, target int, direction HysteresisDirection) bool {
	diff := current - target
	if direction == DirectionAscending {
		return diff > 0
	}
	return diff < 0
}
