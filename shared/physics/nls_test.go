package physics

import (
	"math"
	"testing"
)

// The same activation/decay equation must not select a CPU-specific Exp
// polynomial. References are rounded from 100-digit exp at binary64 inputs.
func TestNLSKineticsReproducibleExponentials(t *testing.T) {
	n := NLSKinetics{Tau0: 1, Ea: 1.74}
	if got := n.CalculateTau(1); got != 5.697343422671991 {
		t.Fatalf("activation: %.17g", got)
	}
	n.Ea = 0
	if got := n.Relax(1, 0, 1, 1.74); got != 0.17552040061699686 {
		t.Fatalf("relaxation: %.17g", got)
	}
}

func TestNLSKineticsRejectsNonFiniteInputs(t *testing.T) {
	const (
		currentP = 0.10
		targetP  = 0.20
		field    = 1e8
		dt       = 1e-9
	)

	cases := []struct {
		name  string
		setup func(*NLSKinetics) (float64, float64, float64, float64)
	}{
		{name: "nan field", setup: func(*NLSKinetics) (float64, float64, float64, float64) { return currentP, targetP, math.NaN(), dt }},
		{name: "positive infinite field", setup: func(*NLSKinetics) (float64, float64, float64, float64) { return currentP, targetP, math.Inf(1), dt }},
		{name: "nan dt", setup: func(*NLSKinetics) (float64, float64, float64, float64) { return currentP, targetP, field, math.NaN() }},
		{name: "positive infinite dt", setup: func(*NLSKinetics) (float64, float64, float64, float64) { return currentP, targetP, field, math.Inf(1) }},
		{name: "nan tau0", setup: func(n *NLSKinetics) (float64, float64, float64, float64) {
			n.Tau0 = math.NaN()
			return currentP, targetP, field, dt
		}},
		{name: "positive infinite tau0", setup: func(n *NLSKinetics) (float64, float64, float64, float64) {
			n.Tau0 = math.Inf(1)
			return currentP, targetP, field, dt
		}},
		{name: "zero tau0", setup: func(n *NLSKinetics) (float64, float64, float64, float64) {
			n.Tau0 = 0
			return currentP, targetP, field, dt
		}},
		{name: "nan activation field", setup: func(n *NLSKinetics) (float64, float64, float64, float64) {
			n.Ea = math.NaN()
			return currentP, targetP, field, dt
		}},
		{name: "positive infinite activation field", setup: func(n *NLSKinetics) (float64, float64, float64, float64) {
			n.Ea = math.Inf(1)
			return currentP, targetP, field, dt
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nls := NewNLSKinetics()
			current, target, eField, step := tc.setup(nls)

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("expected invalid NLS input to be rejected without panic, got panic: %v", r)
				}
			}()

			got := nls.Relax(current, target, eField, step)
			if got != current {
				t.Fatalf("expected invalid NLS input to preserve current polarization %.6g C/m², got %.6g C/m²", current, got)
			}
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("expected invalid NLS input to return finite polarization, got %.6g C/m²", got)
			}

			tau := nls.CalculateTau(eField)
			if math.IsNaN(tau) || tau <= 0 {
				t.Fatalf("expected invalid NLS input to yield a positive finite-or-infinite retention tau, got %.6g s", tau)
			}
		})
	}
}
