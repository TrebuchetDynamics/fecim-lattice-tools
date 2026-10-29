package physics

import "math"

// TanhEverett implements EverettFunction using the product form of the
// Preisach Everett integral for sech^2-distributed hysterons.
//
// The Everett function F(alpha,beta) is the integral of the Preisach density
// mu(alpha,beta) over the triangle {(a,b) : beta <= b <= a <= alpha} in the
// Preisach half-plane. For the factorizable sech^2 density:
//
//	mu(a,b) = (Ps / 4*Delta^2) * sech^2((a - Ec)/Delta) * sech^2((b + Ec)/Delta)
//
// the exact integral yields the product form:
//
//	F(alpha,beta) = [1 + tanh((alpha - Ec)/Delta)] * [1 - tanh((beta + Ec)/Delta)] * Ps/4
//
// This product form is guaranteed non-negative for all (alpha,beta), unlike the
// older difference form [tanh(.) - tanh(.)] which can go negative for minor
// loops within the coercive gap (|alpha - beta| < 2*Ec). See MEMORY.md entry
// "Preisach Everett Zero-Clamp Teleportation" for the bug this form fixed.
//
// Units: Ps, Ec in SI (C/m^2, V/m). Delta controls distribution width (V/m).
type TanhEverett struct {
	Ps    float64 // Irreversible saturation polarization (C/m^2)
	Ec    float64 // Coercive field (V/m)
	Delta float64 // Distribution width (V/m); controls loop squareness
}

// Calculate returns the Everett integral F(alpha, beta) for the half-plane
// region defined by switching thresholds alpha (ascending) and beta (descending).
// The result is in C/m^2 and is guaranteed non-negative by the product form.
//
// If Delta == 0 the distribution collapses to a delta function; the Everett
// integral is Ps when both alpha > Ec and beta < -Ec, and 0 otherwise.
func (t *TanhEverett) Calculate(alpha, beta float64) float64 {
	if !t.hasValidBinding() || !isFiniteTanhEverettValue(alpha) || !isFiniteTanhEverettValue(beta) {
		return 0
	}
	if t.Delta <= 0 {
		// Degenerate case: infinitely sharp distribution (square loop).
		// Delta <= 0 prevents NaN from 0/0 and sign-flipped tanh arguments.
		if alpha > t.Ec && beta < -t.Ec {
			return t.Ps
		}
		return 0
	}

	ascCDF := 1.0 + PreisachTanh((alpha-t.Ec)/t.Delta)
	descSurv := 1.0 - PreisachTanh((beta+t.Ec)/t.Delta)

	val := ascCDF * descSurv * t.Ps * 0.25
	if !isFiniteTanhEverettValue(val) || val < 0 {
		return 0
	}
	if val > t.Ps {
		return t.Ps
	}
	return val
}

// PreisachTanh evaluates the existing tanh equation with a fixed exponential
// backend, shared by Everett and the adapter's reversible contribution. The
// small-argument math.Tanh path is the same rational polynomial in native Go
// and WASM; its large-argument path otherwise calls architecture-specific Exp.
func PreisachTanh(x float64) float64 {
	z := math.Abs(x)
	if z < 0.625 || math.IsNaN(x) {
		return math.Tanh(x)
	}
	if z > 0.5*8.8029691931113054295988e+01 {
		return math.Copysign(1, x)
	}
	return math.Copysign(1-2/(preisachExp(2*z)+1), x)
}

func (t *TanhEverett) hasValidBinding() bool {
	return t != nil &&
		t.Ps > 0 && isFiniteTanhEverettValue(t.Ps) &&
		t.Ec > 0 && isFiniteTanhEverettValue(t.Ec) &&
		!math.IsNaN(t.Delta) && !math.IsInf(t.Delta, 0)
}

func isFiniteTanhEverettValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
