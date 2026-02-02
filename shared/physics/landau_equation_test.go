package physics

import (
	"math"
	"testing"
)

func TestLKSolver_dPdT_Equation(t *testing.T) {
	s := &LKSolver{
		Alpha: 1.2,
		Beta:  -0.5,
		Gamma: 0.25,
		K_dep: 2.0,
	}

	P := 0.3
	E := 5.0
	noise := -0.2
	rhoEff := 1.5

	Edep := s.K_dep * P
	Deff := E - Edep
	dGdP := (2 * s.Alpha * P) + (4 * s.Beta * math.Pow(P, 3)) + (6 * s.Gamma * math.Pow(P, 5))
	expected := (Deff + noise - dGdP) / rhoEff

	got := s.dPdT(0, P, E, noise, rhoEff)
	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf("dPdT mismatch: got %.12f, expected %.12f", got, expected)
	}
}

func TestLKSolver_effectiveRho(t *testing.T) {
	s := &LKSolver{
		Rho:                   0.1,
		UseEffectiveViscosity: true,
		SeriesResistance:      50,
		Area:                  2,
		Thickness:             5,
	}

	expected := 0.1 + (50*2)/5
	got := s.effectiveRho()
	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf("effectiveRho mismatch: got %.12f, expected %.12f", got, expected)
	}
}

func TestLKSolver_effectiveRho_Disabled(t *testing.T) {
	s := &LKSolver{
		Rho:                   0.25,
		UseEffectiveViscosity: false,
		SeriesResistance:      50,
		Area:                  2,
		Thickness:             5,
	}

	got := s.effectiveRho()
	if math.Abs(got-s.Rho) > 1e-12 {
		t.Fatalf("expected rho without series contribution: got %.12f, expected %.12f", got, s.Rho)
	}
}
