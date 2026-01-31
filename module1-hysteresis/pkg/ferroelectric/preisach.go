// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"math"

	"fecim-lattice-tools/shared/logging"
)

// Package-level logger
var log *logging.Logger

func init() {
	log = logging.NewLogger("preisach")
}

// PreisachModel implements the Preisach hysteresis model for ferroelectrics.
// Based on Bartic et al. (2001) "Preisach model for the simulation of
// ferroelectric capacitors" and Bo Jiang's hyperbolic tangent method.
type PreisachModel struct {
	material *HZOMaterial

	// Distribution parameters
	EcMean  float64 // Mean coercive field
	EcSigma float64 // Coercive field distribution width
	EuMean  float64 // Mean interaction field
	EuSigma float64 // Interaction field distribution width

	// History tracking (LIFO stack for turning points)
	turningPointsE []float64 // E-field at reversal
	turningPointsP []float64 // Polarization at reversal
	lastE         float64
	increasing    bool

	// Current state
	polarization float64
}

// NewPreisachModel creates a new Preisach model with the given material.
func NewPreisachModel(material *HZOMaterial) *PreisachModel {
	log.Input("NewPreisachModel", map[string]interface{}{
		"material_name": material.Name,
		"Ec":            material.Ec,
		"Ps":            material.Ps,
		"Pr":            material.Pr,
	})

	p := &PreisachModel{
		material:      material,
		EcMean:        material.Ec,
		EcSigma:       material.Ec * 0.25, // 25% distribution width
		EuMean:        0,
		EuSigma:        material.Ec * 0.4,
		turningPointsE: make([]float64, 0, 100),
		turningPointsP: make([]float64, 0, 100),
		polarization:   0,
	}

	log.Calculation("NewPreisachModel", map[string]interface{}{
		"EcMean":  p.EcMean,
		"EcSigma": p.EcSigma,
		"EuMean":  p.EuMean,
		"EuSigma": p.EuSigma,
	}, p)

	return p
}

// Reset clears the history and sets polarization to zero.
func (p *PreisachModel) Reset() {
	p.turningPointsE = p.turningPointsE[:0]
	p.turningPointsP = p.turningPointsP[:0]
	p.polarization = 0
	p.lastE = 0
	p.increasing = true // Start assuming ascending direction to avoid first-point discontinuity
}

// Update applies a new electric field and returns the resulting polarization.
// The field E should be in V/m.
func (p *PreisachModel) Update(E float64) float64 {
	log.Input("Update", map[string]interface{}{
		"E_field": E,
	})

	// Determine direction
	increasing := E > p.lastE

	// Check for turning point (direction change)
	if len(p.turningPointsE) > 0 && increasing != p.increasing {
		p.addTurningPoint(p.lastE, p.polarization)
	} else if len(p.turningPointsE) == 0 && p.polarization == 0 && E != 0 {
		// First point from zero - implicitly track start
		// p.addTurningPoint(0, 0)
	}

	// Apply Wipe-out property strictly BEFORE calculation
	// This ensures we are always on the correct branch
	p.applyWipeOut(E, increasing)

	// Calculate polarization using hyperbolic tangent model with history
	// This captures the S-shaped switching characteristic and proper minor loops
	p.polarization = p.calculatePolarization(E, increasing)

	// Update state
	p.lastE = E
	p.increasing = increasing

	log.Calculation("Update", map[string]interface{}{
		"E_field":        E,
		"turning_points": len(p.turningPointsE),
		"increasing":     p.increasing,
	}, p.polarization)

	return p.polarization
}

// calculatePolarization computes P(E) using the Preisach distribution.
func (p *PreisachModel) calculatePolarization(E float64, increasing bool) float64 {
	// Hyperbolic tangent switching function (Bo Jiang method)
	Ps := p.material.Ps
	delta := p.EcSigma * 2 // Transition width

	// Calculate effective coercive field based on history
	EcEff := p.effectiveCoerciveField()

	// Base switching function
	majorP := func(e float64, inc bool) float64 {
		if inc {
			return Ps * math.Tanh((e-EcEff)/delta)
		}
		return Ps * math.Tanh((e+EcEff)/delta)
	}

	// Get Major Loop value at current E
	P_major := majorP(E, increasing)

	// If no history, return Major Loop value
	if len(p.turningPointsE) == 0 {
		return P_major
	}

	// INTERPOLATION for Minor Loops (Bartic et al.)
	// We scale the Major Loop shape to fit between the last turning point and saturation
	// P(E) = P_start + S * (P_major(E) - P_major(E_start))
	// where S scales the major loop slope to connect (E_start, P_start) to (Infinity, Ps)

	lastIdx := len(p.turningPointsE) - 1
	E_start := p.turningPointsE[lastIdx]
	P_start := p.turningPointsP[lastIdx]
	P_major_start := majorP(E_start, p.increasing)

	// Calculate scaling factor S
	// We want P -> Target as E -> Infinity
	// If increasing: Target is +Ps. If decreasing: Target is -Ps.
	var TargetP float64
	var TargetMajor float64
	if p.increasing {
		TargetP = Ps
		TargetMajor = Ps
	} else {
		TargetP = -Ps
		TargetMajor = -Ps
	}

	// Avoid division by zero
	denom := TargetMajor - P_major_start
	if math.Abs(denom) < 1e-9 {
		return P_start // Already at saturation
	}

	S := (TargetP - P_start) / denom

	// Interpolated P
	P := P_start + S*(P_major-P_major_start)

	// Safety clamp
	if P > Ps {
		P = Ps
	} else if P < -Ps {
		P = -Ps
	}

	return P
}



// effectiveCoerciveField returns Ec modified by the Preisach distribution.
func (p *PreisachModel) effectiveCoerciveField() float64 {
	// In a full Preisach model, this would integrate over the distribution
	// For simplicity, we use the mean with small random variation
	return p.EcMean
}

// addTurningPoint records a reversal in the field sweep direction.
func (p *PreisachModel) addTurningPoint(E, P float64) {
	p.turningPointsE = append(p.turningPointsE, E)
	p.turningPointsP = append(p.turningPointsP, P)
}

// applyWipeOut implements the Preisach "Wipe-out" property.
// If the field excursion goes beyond a previous turning point, that memory is erased.
func (p *PreisachModel) applyWipeOut(E float64, increasing bool) {
	if len(p.turningPointsE) == 0 {
		return
	}

	// Loop to handle multiple wipes (e.g. large spike)
	for len(p.turningPointsE) > 0 {
		lastIdx := len(p.turningPointsE) - 1


		shouldWipe := false
		if increasing {
			// Going UP: wipe if we exceed previous Max (which must be a turning point where we started going down)
			// Wait, the stack is [..., Min, Max, Min, Max...]
			// The last point was where we turned to come HERE.
			// If we are increasing, the last point was a MINIMUM (start of this branch).
			// We don't wipe the start of our own branch!
			// We check the point BEFORE that (a Maximum).
			// BUT, Mayergoyz stack logic:
			// If increasing, we are checking if E > Previous Max.
			// The stack top is the Minimum we just turned from.
			// Any Maxima inside the minor loop < E are wiped?
			//
			// Actually, simpler logic:
			// If increasing, and E > E_stack_top, does it mean anything?
			// If E > E_stack_top, and E_stack_top was a Maximum, we wiped it.
			// Use the alternating nature of stack.
			// If stack has 1 element: [Min]. We are going up. If E < Min, impossible (we are up).
			//
			// Correct Logic (Mayergoyz):
			// The stack contains pairs (M, m).
			// If we are increasing, we check against M (previous max).
			// The last element in `turningPoints` is the point we turned FROM.
			// If we are increasing, we turned from a Minimum (stack top).
			// We want to know if we exceed the Maximum prior to that.
			// Stack: [..., Max_prev, Min_last].
			// If E > Max_prev, then Min_last and Max_prev are wiped.
			
			if lastIdx >= 1 {
				prevMax := p.turningPointsE[lastIdx-1]
				if E >= prevMax {
					shouldWipe = true
				}
			}
		} else {
			// Going DOWN. Last point was a Maximum.
			// Check if E < Min_prev (point before last).
			if lastIdx >= 1 {
				prevMin := p.turningPointsE[lastIdx-1]
				if E <= prevMin {
					shouldWipe = true
				}
			}
		}

		if shouldWipe {
			// Pop TWO points (the Min and the Max defining the minor loop)
			// effectively returning to the outer loop
			p.turningPointsE = p.turningPointsE[:lastIdx-1]
			p.turningPointsP = p.turningPointsP[:lastIdx-1]
		} else {
			break
		}
	}
}

// Polarization returns the current polarization state.
func (p *PreisachModel) Polarization() float64 {
	return p.polarization
}

// NormalizedPolarization returns polarization as fraction of Ps (-1 to +1).
func (p *PreisachModel) NormalizedPolarization() float64 {
	return p.polarization / p.material.Ps
}

// GetHysteresisLoop generates a full P-E hysteresis loop.
// Returns slices of E and P values for plotting.
func (p *PreisachModel) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64) {
	log.Input("GetHysteresisLoop", map[string]interface{}{
		"Emax":   Emax,
		"points": points,
	})

	p.Reset()

	E := make([]float64, 0, points*4)
	P := make([]float64, 0, points*4)

	// First, establish initial saturation state at -Emax (not recorded)
	// This ensures we start from a well-defined state on the major loop
	p.Update(-Emax)

	// Sweep from -Emax to +Emax (ascending branch)
	for i := 0; i <= points*2; i++ {
		e := -Emax + 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		P = append(P, pol)
	}

	// Sweep from +Emax back to -Emax (descending branch)
	for i := 1; i <= points*2; i++ {
		e := Emax - 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		P = append(P, pol)
	}

	log.Output("GetHysteresisLoop", map[string]interface{}{
		"E_points": len(E),
		"P_points": len(P),
		"P_max":    maxFloat64(P),
		"P_min":    minFloat64(P),
	})

	return E, P
}

// DiscreteStates returns polarization values for N discrete analog states.
// This demonstrates the 30-state capability of FeCIM.
func (p *PreisachModel) DiscreteStates(N int) []float64 {
	log.Input("DiscreteStates", map[string]interface{}{
		"N_states": N,
		"Ps":       p.material.Ps,
	})

	states := make([]float64, N)
	Ps := p.material.Ps

	for i := 0; i < N; i++ {
		// Linear spacing from -Ps to +Ps
		states[i] = -Ps + 2*Ps*float64(i)/float64(N-1)
	}

	log.Output("DiscreteStates", map[string]interface{}{
		"N_states":   N,
		"state_min":  states[0],
		"state_max":  states[N-1],
		"state_step": (states[N-1] - states[0]) / float64(N-1),
	})

	return states
}

// Helper functions for logging
func maxFloat64(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	max := slice[0]
	for _, v := range slice {
		if v > max {
			max = v
		}
	}
	return max
}

func minFloat64(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	min := slice[0]
	for _, v := range slice {
		if v < min {
			min = v
		}
	}
	return min
}
