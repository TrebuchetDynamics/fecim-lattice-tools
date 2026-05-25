package physics

import (
	"math"
)

// NLSKinetics implements Nucleation Limited Switching dynamics (Merz's Law).
// This introduces time-dependence to the polarization switching process.
//
// Theoretical Basis:
// Switching time tau follows an exponential activation law:
// tau(E) = tau0 * exp(Ea / |E|)
//
// where:
//   - tau0: Intrinsic switching time limit (at infinite field)
//   - Ea: Activation field (characteristic energy barrier)
//   - E: Applied electric field
//
// The polarization relaxation towards the static target (P_eq) follows:
// dP/dt = -(P - P_eq) / tau(E)
//
// For a constant field step over duration dt:
// P_new = P_eq + (P_old - P_eq) * exp(-dt / tau(E))
type NLSKinetics struct {
	Tau0 float64 // Intrinsic time constant (s)
	Ea   float64 // Activation field (V/m)
}

// NewNLSKinetics creates a new NLS model with default parameters typical for HZO.
func NewNLSKinetics() *NLSKinetics {
	return &NLSKinetics{
		Tau0: 1e-10, // Intrinsic attempt time for HZO (~100 ps); Guo et al. APL 112, 262903 (2018)
		Ea:   8e8,   // Activation field (~8 MV/cm), typically higher than Ec
	}
}

// CalculateTau computes the characteristic switching time for a given field E.
func (n *NLSKinetics) CalculateTau(E float64) float64 {
	if !isFiniteNLSValue(E) || !n.hasValidParameters() {
		return math.Inf(1)
	}

	absE := math.Abs(E)
	if absE < 1e-9 {
		// Zero field: effective infinite retention (tau -> infinity)
		// Return a very large number to prevent numerical issues
		return 1e15 // ~30 million years
	}
	tau := n.Tau0 * math.Exp(n.Ea/absE)
	if !isFiniteNLSValue(tau) || tau <= 0 {
		return math.Inf(1)
	}
	return tau
}

// Relax updates polarization P towards equilibrium P_target over time dt at field E.
func (n *NLSKinetics) Relax(currentP, targetP, E, dt float64) float64 {
	if !isFiniteNLSValue(currentP) {
		return 0
	}
	if dt <= 0 || !isFiniteNLSValue(dt) || !isFiniteNLSValue(targetP) || !isFiniteNLSValue(E) || !n.hasValidParameters() {
		return currentP
	}

	tau := n.CalculateTau(E)
	if !isFiniteNLSValue(tau) {
		return currentP
	}

	// If tau is extremely small relative to dt, we reach equilibrium instantly.
	if tau < dt*1e-6 {
		return targetP
	}

	// P(t) decay
	decay := math.Exp(-dt / tau)
	if !isFiniteNLSValue(decay) {
		return currentP
	}
	relaxed := targetP + (currentP-targetP)*decay
	if !isFiniteNLSValue(relaxed) {
		return currentP
	}
	return relaxed
}

func (n *NLSKinetics) hasValidParameters() bool {
	return n != nil && n.Tau0 > 0 && isFiniteNLSValue(n.Tau0) && n.Ea >= 0 && isFiniteNLSValue(n.Ea)
}

func isFiniteNLSValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
