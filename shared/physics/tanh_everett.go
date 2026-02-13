package physics

import "math"

// TanhEverett implements EverettFunction with a factorized tanh formulation.
// It is suitable for PreisachStack-based ferroelectric hysteresis simulations.
type TanhEverett struct {
	Ps    float64
	Ec    float64
	Delta float64 // Distribution width
}

// Calculate returns the Everett integral contribution for (alpha, beta).
func (t *TanhEverett) Calculate(alpha, beta float64) float64 {
	ascCDF := 1.0 + math.Tanh((alpha-t.Ec)/t.Delta)
	descSurv := 1.0 - math.Tanh((beta+t.Ec)/t.Delta)

	val := ascCDF * descSurv * t.Ps * 0.25
	if val > t.Ps {
		return t.Ps
	}
	return val
}
