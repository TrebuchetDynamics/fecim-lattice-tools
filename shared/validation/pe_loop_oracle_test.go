package validation

import (
	"math"
	"testing"
)

// TestLDParams_dFdP_AtEquilibrium verifies dF/dP = 0 at the spontaneous polarisation.
// For E=0, the equilibrium is at P_s ≈ sqrt(-α/(2β)) (6th-order term neglected).
func TestLDParams_dFdP_AtEquilibrium(t *testing.T) {
	p := LDParamsForHZO()
	Ps := math.Sqrt(-p.Alpha / (2 * p.Beta))
	// dF/dP at P=Ps, E=0 should be small (near zero by construction)
	residual := math.Abs(p.dFdP(Ps, 0))
	// For cubic: 2α·Ps + 4β·Ps³ = 2Ps(α + 2β·Ps²) = 2Ps(α - α) = 0 exactly
	if residual > 1e-6*math.Abs(p.Alpha*Ps) {
		t.Errorf("dF/dP at Ps not near zero: residual=%.4e", residual)
	}
}

// TestLDParams_d2FdP2_Sign verifies curvature is positive at equilibrium (stable).
func TestLDParams_d2FdP2_Sign(t *testing.T) {
	p := LDParamsForHZO()
	Ps := math.Sqrt(-p.Alpha / (2 * p.Beta))
	curv := p.d2FdP2(Ps)
	if curv <= 0 {
		t.Errorf("d²F/dP² at Ps should be > 0 (stable), got %.4e", curv)
	}
}

// TestSolveEquilibrium_ZeroField verifies the solver finds ±Ps at E=0.
func TestSolveEquilibrium_ZeroField(t *testing.T) {
	p := LDParamsForHZO()
	Ps := math.Sqrt(-p.Alpha / (2 * p.Beta))

	Ppos, err := p.SolveEquilibrium(0, Ps*0.9)
	if err != nil {
		t.Fatalf("SolveEquilibrium(E=0, P0=+Ps): %v", err)
	}
	if math.Abs(Ppos-Ps) > 1e-6 {
		t.Errorf("positive equilibrium: got %.6f, want %.6f", Ppos, Ps)
	}

	Pneg, err := p.SolveEquilibrium(0, -Ps*0.9)
	if err != nil {
		t.Fatalf("SolveEquilibrium(E=0, P0=-Ps): %v", err)
	}
	if math.Abs(Pneg+Ps) > 1e-6 {
		t.Errorf("negative equilibrium: got %.6f, want %.6f", Pneg, -Ps)
	}
}

// TestPELoopOracle_Shape validates the hysteresis loop has the correct topology:
// - Correct number of points
// - Loop opens (P switches sign within the sweep)
// - Saturation in correct direction at ±Emax
func TestPELoopOracle_Shape(t *testing.T) {
	p := LDParamsForHZO()
	Emax := 3e6 // 3 MV/m (3× above Ec ≈ 1 MV/m for HZO params)
	nPts := 200

	res, err := PELoopOracle(p, Emax, nPts)
	if err != nil {
		t.Fatalf("PELoopOracle: %v", err)
	}
	if len(res.Points) != 2*nPts {
		t.Errorf("expected %d points, got %d", 2*nPts, len(res.Points))
	}

	// First point: E = +Emax, P should be positive (positive saturation)
	if res.Points[0].P <= 0 {
		t.Errorf("P at +Emax should be positive, got %.4e", res.Points[0].P)
	}

	// End of first half-cycle: E = -Emax, P should be negative (switched)
	midIdx := nPts - 1
	if res.Points[midIdx].P >= 0 {
		t.Errorf("P at -Emax (first half) should be negative, got %.4e", res.Points[midIdx].P)
	}

	// Last point: E = +Emax return branch, P should be positive again
	if res.Points[len(res.Points)-1].P <= 0 {
		t.Errorf("P at +Emax (return) should be positive, got %.4e", res.Points[len(res.Points)-1].P)
	}
}

// TestPELoopOracle_Pr validates that remnant polarisation is physically reasonable.
// For HZO: literature Pr ≈ 10–30 µC/cm² = 0.10–0.30 C/m².
func TestPELoopOracle_Pr(t *testing.T) {
	p := LDParamsForHZO()
	res, err := PELoopOracle(p, 3e6, 500)
	if err != nil {
		t.Fatalf("PELoopOracle: %v", err)
	}
	if res.Pr < 0.05 || res.Pr > 0.50 {
		t.Errorf("Pr out of expected range [0.05, 0.50] C/m²: got %.4e", res.Pr)
	}
}

// TestPELoopOracle_Ec validates that the coercive field is physically reasonable.
// For HZO: literature Ec ≈ 0.5–3 MV/m.
func TestPELoopOracle_Ec(t *testing.T) {
	p := LDParamsForHZO()
	res, err := PELoopOracle(p, 3e6, 500)
	if err != nil {
		t.Fatalf("PELoopOracle: %v", err)
	}
	if res.Ec < 0.1e6 || res.Ec > 5e6 {
		t.Errorf("Ec out of expected range [0.1, 5] MV/m: got %.4e V/m", res.Ec)
	}
}

// TestPELoopOracle_Symmetry validates that the loop is approximately symmetric:
// P(+Emax) is positive on both the forward and return branches at saturation.
func TestPELoopOracle_Symmetry(t *testing.T) {
	p := LDParamsForHZO()
	nPts := 300
	res, err := PELoopOracle(p, 3e6, nPts)
	if err != nil {
		t.Fatalf("PELoopOracle: %v", err)
	}

	// First point (+Emax) and last point (+Emax return branch) should have same P
	first := res.Points[0]
	last := res.Points[len(res.Points)-1]
	if math.Abs(first.E-last.E) > 1e3 {
		t.Errorf("first/last E mismatch: %.4e vs %.4e", first.E, last.E)
	}
	if math.Abs(first.P-last.P)/math.Abs(first.P) > 0.01 {
		t.Errorf("first/last P differ by >1%%: %.4e vs %.4e", first.P, last.P)
	}
}

// TestPELoopOracle_InvalidInputs checks error handling.
func TestPELoopOracle_InvalidInputs(t *testing.T) {
	p := LDParamsForHZO()

	if _, err := PELoopOracle(p, 5e6, 1); err == nil {
		t.Error("expected error for nPoints=1")
	}
	if _, err := PELoopOracle(p, -1e6, 100); err == nil {
		t.Error("expected error for Emax<=0")
	}
	if _, err := PELoopOracle(p, 0, 100); err == nil {
		t.Error("expected error for Emax=0")
	}
}

// TestEstimatePsat verifies analytic Psat estimate against Newton equilibrium.
func TestEstimatePsat(t *testing.T) {
	p := LDParamsForHZO()
	Psat := estimatePsat(p)
	// Psat should satisfy dF/dP ≈ 0 at E=0 (within 1%)
	residual := math.Abs(p.dFdP(Psat, 0)) / (math.Abs(p.Alpha) * Psat)
	if residual > 0.01 {
		t.Errorf("estimatePsat: relative dFdP residual %.4f > 1%%", residual)
	}
}
