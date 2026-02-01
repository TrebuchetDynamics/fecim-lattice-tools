package physics

import (
	"math"
)

// AdaptiveISPP implements Incremental Step Pulse Programming
// with Binary Search for high-speed FeCIM writing.
type AdaptiveISPP struct {
	Solver *LKSolver
	Material *HZOMaterial
	
	// Control Parameters
	MaxVoltage    float64
	MinVoltage    float64
	TargetTolerance float64 // Acceptable error in P/Ps
	MaxIterations int
	PulseWidth    float64 // Duration of write pulse (s)
}

func NewAdaptiveISPP(solver *LKSolver, mat *HZOMaterial) *AdaptiveISPP {
	return &AdaptiveISPP{
		Solver: solver,
		Material: mat,
		MaxVoltage: 3.0, // 3V max for 10nm HZO
		MinVoltage: -3.0,
		TargetTolerance: 0.01, // 1% error
		MaxIterations: 10,
		PulseWidth: mat.Tau, // Use characteristic switching time (e.g. 1ns)
	}
}

// PredictState estimates the required voltage to reach target Polarization
// using a simplified inverse model (Linear or Tanh inverse).
func (c *AdaptiveISPP) PredictState(targetP float64) float64 {
	// Simple Tanh Inverse: P = Ps * Tanh((V - Vc)/Delta)
	// V = Vc + Delta * Atanh(P/Ps)
	
	Ps := c.Material.Ps
	Vc := c.Material.Ec * c.Material.Thickness
	Delta := 0.5 // Fitting parameter
	
	// Clamp target ratio to avoid infinity
	ratio := targetP / Ps
	if ratio > 0.95 { ratio = 0.95 }
	if ratio < -0.95 { ratio = -0.95 }
	
	V_est := Vc + Delta * math.Atanh(ratio)
	return V_est
}

// BinarySearchWrite performs the write-verify-correct loop.
func (c *AdaptiveISPP) BinarySearchWrite(targetP float64) (float64, int, bool) {
	// 1. Prediction (Initial Guess)
	V_next := c.PredictState(targetP)
	
	// Range for Binary Search
	V_min := c.MinVoltage
	V_max := c.MaxVoltage
	
	success := false
	iter := 0
	
	for iter < c.MaxIterations {
		iter++
		
		// 2. Pulse (Apply Voltage)
		// Convert Voltage to Field: E = V / thickness
		E_field := V_next / c.Material.Thickness
		
		// Run Physics (L-K Solver)
		// We step for the duration of PulseWidth
		c.Solver.Step(E_field, c.PulseWidth)
		
		// 3. Verify (Read State)
		currentP := c.Solver.GetState()
		
		// Check error
		errorP := currentP - targetP
		
		if math.Abs(errorP) <= c.TargetTolerance * c.Material.Ps {
			success = true
			break
		}
		
		// 4. Correction (Adaptive Binary Search)
		if errorP < 0 {
			// Too low (Need higher V)
			V_min = V_next
			V_next = (V_min + V_max) / 2
		} else {
			// Too high (Overshoot) - CRITICAL
			// We cannot just lower V next time, we are already stuck at high P.
			// We must RESET/ERASE slightly to drop P, then try lower V.
			// Ideally, apply negative pulse.
			
			// Simple reset pulse strategy:
			c.Solver.Step(-c.Material.Ec * c.Material.Thickness / c.Material.Thickness, c.PulseWidth)
			
			V_max = V_next
			V_next = (V_min + V_max) / 2
		}
	}
	
	return c.Solver.GetState(), iter, success
}
