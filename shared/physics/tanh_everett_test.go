package physics

import (
	"math"
	"testing"
)

func TestTanhEverettRejectsInvalidInputs(t *testing.T) {
	valid := TanhEverett{Ps: 0.3, Ec: 1e8, Delta: 2e7}
	cases := []struct {
		name  string
		ev    *TanhEverett
		alpha float64
		beta  float64
	}{
		{name: "nil receiver", ev: nil, alpha: valid.Ec, beta: -valid.Ec},
		{name: "nan alpha", ev: &valid, alpha: math.NaN(), beta: -valid.Ec},
		{name: "positive infinite alpha", ev: &valid, alpha: math.Inf(1), beta: -valid.Ec},
		{name: "nan beta", ev: &valid, alpha: valid.Ec, beta: math.NaN()},
		{name: "negative infinite beta", ev: &valid, alpha: valid.Ec, beta: math.Inf(-1)},
		{name: "zero ps", ev: &TanhEverett{Ps: 0, Ec: valid.Ec, Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "negative ps", ev: &TanhEverett{Ps: -0.3, Ec: valid.Ec, Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "nan ps", ev: &TanhEverett{Ps: math.NaN(), Ec: valid.Ec, Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "infinite ps", ev: &TanhEverett{Ps: math.Inf(1), Ec: valid.Ec, Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "zero ec", ev: &TanhEverett{Ps: valid.Ps, Ec: 0, Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "negative ec", ev: &TanhEverett{Ps: valid.Ps, Ec: -valid.Ec, Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "nan ec", ev: &TanhEverett{Ps: valid.Ps, Ec: math.NaN(), Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "infinite ec", ev: &TanhEverett{Ps: valid.Ps, Ec: math.Inf(1), Delta: valid.Delta}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "nan delta", ev: &TanhEverett{Ps: valid.Ps, Ec: valid.Ec, Delta: math.NaN()}, alpha: valid.Ec, beta: -valid.Ec},
		{name: "infinite delta", ev: &TanhEverett{Ps: valid.Ps, Ec: valid.Ec, Delta: math.Inf(1)}, alpha: valid.Ec, beta: -valid.Ec},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("expected invalid Everett input to be rejected without panic, got panic: %v", r)
				}
			}()

			got := tc.ev.Calculate(tc.alpha, tc.beta)
			if got != 0 {
				t.Fatalf("expected invalid Everett input to return 0 C/m², got %.6g C/m²", got)
			}
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("expected invalid Everett input to return finite polarization, got %.6g C/m²", got)
			}
		})
	}
}
