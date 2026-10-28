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
	tau := n.Tau0 * preisachExp(n.Ea/absE)
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
	decay := preisachExp(-dt / tau)
	if !isFiniteNLSValue(decay) {
		return currentP
	}
	relaxed := targetP + (currentP-targetP)*decay
	if !isFiniteNLSValue(relaxed) {
		return currentP
	}
	return relaxed
}

// preisachExp fixes the evaluation order of the existing exponential equation
// to Go's portable range-reduced implementation. math.Exp selects different
// native assembly polynomials even with FMA disabled; that changed programming
// history by one ULP relative to WASM. No precision rounding or equation change
// is applied at a target boundary. Explicit binary64 intermediates forbid FMA.
//
// Adapted from Go 1.26 math/exp.go (exp/expmulti), Copyright 2009 The Go Authors.
// Original constants/algorithm: Copyright (C) 2004 Sun Microsystems, Inc.
// Permission to use, copy, modify, and distribute this software is freely
// granted, provided that this notice is preserved.
//
// Go BSD license: Redistribution and use in source and binary forms, with or
// without modification, are permitted provided that the following conditions
// are met:
//   - Redistributions of source code must retain the above copyright notice,
//     this list of conditions and the following disclaimer.
//   - Redistributions in binary form must reproduce the above copyright notice,
//     this list of conditions and the following disclaimer in the documentation
//     and/or other materials provided with the distribution.
//   - Neither the name of Google LLC nor the names of its contributors may be
//     used to endorse or promote products derived from this software without
//     specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
func preisachExp(x float64) float64 {
	const (
		ln2Hi     = 6.93147180369123816490e-01
		ln2Lo     = 1.90821492927058770002e-10
		log2e     = 1.44269504088896338700e+00
		overflow  = 7.09782712893383973096e+02
		underflow = -7.45133219101941108420e+02
		nearZero  = 1.0 / (1 << 28)
		p1        = 1.66666666666666657415e-01
		p2        = -2.77777777770155933842e-03
		p3        = 6.61375632143793436117e-05
		p4        = -1.65339022054652515390e-06
		p5        = 4.13813679705723846039e-08
	)
	switch {
	case math.IsNaN(x):
		return x
	case x > overflow:
		return math.Inf(1)
	case x < underflow:
		return 0
	case -nearZero < x && x < nearZero:
		return 1 + x
	}
	var k int
	if x < 0 {
		k = int(float64(log2e*x) - 0.5)
	} else {
		k = int(float64(log2e*x) + 0.5)
	}
	hi := x - float64(float64(k)*ln2Hi)
	lo := float64(k) * ln2Lo
	r := hi - lo
	t := r * r
	polynomial := p4 + float64(t*p5)
	polynomial = p3 + float64(t*polynomial)
	polynomial = p2 + float64(t*polynomial)
	polynomial = p1 + float64(t*polynomial)
	c := r - float64(t*polynomial)
	y := 1 - ((lo - float64(r*c)/(2-c)) - hi)
	return math.Ldexp(y, k)
}

func (n *NLSKinetics) hasValidParameters() bool {
	return n != nil && n.Tau0 > 0 && isFiniteNLSValue(n.Tau0) && n.Ea >= 0 && isFiniteNLSValue(n.Ea)
}

func isFiniteNLSValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
