// Package validation provides oracle implementations for validating physics simulations.
//
// This file contains a static Landau-Devonshire P-E loop oracle, inspired by
// ferro_scripts' get_pol_vs_e() and analyze_energy_chi() functions. It serves
// as a ground-truth reference for validating M1's dynamic Landau-Khalatnikov
// hysteresis curves (shared/physics/landau.go).
//
// Model: Landau-Devonshire polynomial free energy
//
//	F(P) = α·P² + β·P⁴ + γ·P⁶ − E·P
//
// Static equilibrium (no domain-wall kinetics, no thermal noise):
//
//	dF/dP = 2α·P + 4β·P³ + 6γ·P⁶ = E   (Newton's method)
//
// The quasi-static P-E loop is traced by sweeping E and following the stable
// branch. A branch-jump (hysteresis switching) occurs when the curvature
// d²F/dP² = 2α + 12β·P² + 30γ·P⁴ < 0 (spinodal instability).
package validation

import (
	"fmt"
	"math"
)

// LDParams holds Landau-Devonshire polynomial coefficients for one material.
// All fields use SI units (energy in J/m³, polarisation in C/m²).
type LDParams struct {
	// Alpha is the second-order Landau coefficient (J·m/C² < 0 for ferroelectric).
	Alpha float64

	// Beta is the fourth-order Landau coefficient (J·m⁵/C⁴ > 0 for stability).
	Beta float64

	// Gamma is the sixth-order Landau coefficient (J·m⁹/C⁶, may be zero).
	Gamma float64
}

// LDParamsForHZO returns approximate Landau-Devonshire coefficients for HfO₂-based
// ferroelectrics (orthorhombic phase). These are representative simulation defaults;
// calibrate against measured P-E loops for quantitative use.
//
// Coefficients chosen to reproduce Pr ≈ 20 µC/cm² (0.20 C/m²) and Ec ≈ 1 MV/m:
//
//	β = 3√3·Ec / (8·Psat³)  ≈ 8.12×10⁷ J·m⁵/C⁴
//	α = −2β·Psat²            ≈ −6.50×10⁶ J·m/C²
func LDParamsForHZO() LDParams {
	return LDParams{
		Alpha: -6.50e6, // J·m/C²  (negative: ferroelectric phase at room temperature)
		Beta:  8.12e7,  // J·m⁵/C⁴
		Gamma: 0.0,     // neglect sixth-order for simplicity
	}
}

// ── Equilibrium solver ────────────────────────────────────────────────────────

// dFdP computes dF/dP = 2α·P + 4β·P³ + 6γ·P⁵ − E.
func (p *LDParams) dFdP(P, E float64) float64 {
	return 2*p.Alpha*P + 4*p.Beta*P*P*P + 6*p.Gamma*P*P*P*P*P - E
}

// d2FdP2 computes d²F/dP² = 2α + 12β·P² + 30γ·P⁴. Negative → spinodal instability.
func (p *LDParams) d2FdP2(P float64) float64 {
	return 2*p.Alpha + 12*p.Beta*P*P + 30*p.Gamma*P*P*P*P
}

// SolveEquilibrium finds the polarisation P that satisfies dF/dP = E using Newton's
// method starting from P0. Returns an error if the solver fails to converge, or if
// the solution is on an unstable branch (curvature < 0).
func (p *LDParams) SolveEquilibrium(E, P0 float64) (float64, error) {
	const maxIter = 100
	const tol = 1e-12

	P := P0
	for i := 0; i < maxIter; i++ {
		f := p.dFdP(P, E)
		df := p.d2FdP2(P)
		if math.Abs(df) < 1e-30 {
			return 0, fmt.Errorf("pe_loop_oracle: Newton stalled at P=%.4e (zero curvature)", P)
		}
		dP := -f / df
		P += dP
		if math.Abs(dP) < tol {
			// Check stability
			if p.d2FdP2(P) < 0 {
				return 0, fmt.Errorf("pe_loop_oracle: solution P=%.4e is on unstable branch", P)
			}
			return P, nil
		}
	}
	return 0, fmt.Errorf("pe_loop_oracle: Newton did not converge from P0=%.4e at E=%.4e", P0, E)
}

// ── P-E loop oracle ───────────────────────────────────────────────────────────

// PELoopPoint is one point on a P-E hysteresis curve.
type PELoopPoint struct {
	E float64 // applied electric field (V/m)
	P float64 // equilibrium polarisation (C/m²)
}

// PELoopResult holds the full quasi-static hysteresis loop from the oracle.
type PELoopResult struct {
	// Points holds all points in sweep order: positive→negative→positive half-cycle.
	Points []PELoopPoint

	// Pr is the estimated remnant polarisation (C/m²), defined as |P(E=0)|.
	Pr float64

	// Ec is the estimated coercive field (V/m), defined as the field where P changes sign.
	Ec float64
}

// PELoopOracle generates a quasi-static Landau-Devonshire P-E hysteresis loop.
//
// It sweeps E from +Emax → -Emax → +Emax with nPoints per half-cycle, following
// the stable equilibrium branch. A branch-jump occurs automatically when the
// current solution becomes unstable — matching ferro_scripts' get_pol_vs_e().
//
// Parameters:
//   - p: Landau-Devonshire coefficients
//   - Emax: maximum applied electric field (V/m); should exceed coercive field
//   - nPoints: number of field steps per half-cycle (total = 2×nPoints)
func PELoopOracle(p LDParams, Emax float64, nPoints int) (*PELoopResult, error) {
	if nPoints < 2 {
		return nil, fmt.Errorf("pe_loop_oracle: nPoints must be ≥ 2, got %d", nPoints)
	}
	if Emax <= 0 {
		return nil, fmt.Errorf("pe_loop_oracle: Emax must be > 0, got %.4e", Emax)
	}

	// Saturation polarisation estimate: |P_sat| ≈ (-α/β)^0.5 for γ=0
	Psat := estimatePsat(p)

	// Start at positive saturation (+Emax, positive branch)
	P0, err := p.SolveEquilibrium(Emax, Psat)
	if err != nil {
		// Fallback: use analytic estimate
		P0 = Psat
	}

	points := make([]PELoopPoint, 0, 2*nPoints)

	// ── Half-cycle 1: +Emax → -Emax ──────────────────────────────────────────
	P := P0
	for i := 0; i < nPoints; i++ {
		E := Emax - 2*Emax*float64(i)/float64(nPoints-1)
		P, err = trySolveOrJump(&p, E, P, Psat)
		if err != nil {
			return nil, fmt.Errorf("pe_loop_oracle: half-cycle 1 at E=%.4e: %w", E, err)
		}
		points = append(points, PELoopPoint{E: E, P: P})
	}

	// ── Half-cycle 2: -Emax → +Emax ──────────────────────────────────────────
	P = -Psat // restart from negative saturation
	if Psol, err2 := p.SolveEquilibrium(-Emax, -Psat); err2 == nil {
		P = Psol
	}

	for i := nPoints - 1; i >= 0; i-- {
		E := Emax - 2*Emax*float64(i)/float64(nPoints-1)
		P, err = trySolveOrJump(&p, E, P, -Psat)
		if err != nil {
			return nil, fmt.Errorf("pe_loop_oracle: half-cycle 2 at E=%.4e: %w", E, err)
		}
		points = append(points, PELoopPoint{E: E, P: P})
	}

	result := &PELoopResult{Points: points}
	result.Pr = estimatePr(points)
	result.Ec = estimateEc(points)
	return result, nil
}

// trySolveOrJump attempts to solve for the new equilibrium P given E, starting
// from the previous P. If the solution would be on an unstable branch, it jumps
// to the opposite-sign saturation branch.
func trySolveOrJump(p *LDParams, E, prevP, fallbackSign float64) (float64, error) {
	Pnew, err := p.SolveEquilibrium(E, prevP)
	if err != nil {
		// Branch became unstable — jump to opposite branch
		Psign := math.Copysign(estimatePsat(*p), -prevP)
		Pnew, err = p.SolveEquilibrium(E, Psign)
		if err != nil {
			// Last resort: use sign of fallback
			Pnew, err = p.SolveEquilibrium(E, fallbackSign)
			if err != nil {
				return 0, err
			}
		}
	}
	return Pnew, nil
}

// estimatePsat returns an analytic estimate of |P| at saturation for α<0, γ=0:
// P_sat ≈ sqrt(-α / (2β)).
func estimatePsat(p LDParams) float64 {
	if p.Beta == 0 {
		return 0.3 // fallback
	}
	return math.Sqrt(-p.Alpha / (2 * p.Beta))
}

// estimatePr returns |P(E=0)| from the first half of the loop (positive branch through zero).
func estimatePr(points []PELoopPoint) float64 {
	// Find pair of consecutive points where E crosses zero from positive to negative
	for i := 1; i < len(points)/2; i++ {
		if points[i-1].E >= 0 && points[i].E <= 0 {
			// Linear interpolation
			dE := points[i].E - points[i-1].E
			if math.Abs(dE) < 1e-30 {
				return math.Abs(points[i].P)
			}
			t := -points[i-1].E / dE
			P := points[i-1].P + t*(points[i].P-points[i-1].P)
			return math.Abs(P)
		}
	}
	return 0
}

// estimateEc returns the coercive field magnitude: the |E| where P first changes sign
// on the negative-going branch.
func estimateEc(points []PELoopPoint) float64 {
	for i := 1; i < len(points)/2; i++ {
		if points[i-1].P > 0 && points[i].P <= 0 {
			dP := points[i].P - points[i-1].P
			if math.Abs(dP) < 1e-30 {
				return math.Abs(points[i].E)
			}
			t := -points[i-1].P / dP
			E := points[i-1].E + t*(points[i].E-points[i-1].E)
			return math.Abs(E)
		}
	}
	return 0
}
