package physics

import (
	"math"
	"testing"
)

// TestLandauEquilibrium_Materlik2015_Analytical validates the L-K solver against
// the analytical Landau equilibrium curve derived from the LGD coefficients
// published in Materlik et al., J. Appl. Phys. 117, 134109 (2015).
//
// The equilibrium condition (dP/dt = 0, no depolarization) gives:
//
//	E_eq(P) = dG/dP = 2αP + 4βP³ + 6γP⁵
//
// where α comes from Curie-Weiss: α(T) = (T − Tc) / (2ε₀C₀).
//
// NOTE: The raw LGD coefficients give a THERMODYNAMIC coercive field (~0.1 MV/cm)
// which is much lower than the experimental value (~1 MV/cm). This discrepancy
// is well-understood: the Landau model gives the intrinsic switching barrier,
// while real polycrystalline HfO₂ has higher Ec due to domain wall pinning and
// grain boundaries. The ConfigureFromMaterial() rescaling addresses this gap.
func TestLandauEquilibrium_Materlik2015_Analytical(t *testing.T) {
	// Raw Materlik 2015 LGD coefficients (Table I, orthorhombic Pca21 HfO₂).
	// doi:10.1063/1.4916229
	const (
		beta  = -6.72e8   // J·m⁵/C⁴
		gamma = 1.95e10   // J·m⁹/C⁶
		Tc    = 598.0     // K (Curie temperature)
		C0    = 5.3e5     // K (Curie constant)
		T     = 300.0     // K (operating temperature)
		eps0  = 8.854e-12 // F/m (vacuum permittivity)
	)

	// α from Curie-Weiss law.
	alpha := (T - Tc) / (2 * eps0 * C0)
	t.Logf("Materlik raw α = %.6e V·m/C (T=%g K, Tc=%g K, C₀=%g K)", alpha, T, Tc, C0)

	// Analytical E(P) = dG/dP for the Landau free-energy polynomial.
	landauE := func(P float64) float64 {
		P2 := P * P
		P3 := P2 * P
		P5 := P3 * P2
		return 2*alpha*P + 4*beta*P3 + 6*gamma*P5
	}

	// --- Analytical spontaneous polarization P₀ ---
	// At E=0: 2α + 4βP₀² + 6γP₀⁴ = 0, a quadratic in u = P₀².
	aQ := 6 * gamma
	bQ := 4 * beta
	cQ := 2 * alpha
	disc := bQ*bQ - 4*aQ*cQ
	if disc < 0 {
		t.Fatalf("No real solution for P₀: discriminant = %.3e < 0", disc)
	}
	u := (-bQ + math.Sqrt(disc)) / (2 * aQ)
	if u <= 0 {
		t.Fatalf("Non-physical P₀² = %.6e (need > 0)", u)
	}
	P0 := math.Sqrt(u)
	P0uCcm2 := P0 * 100 // C/m² → µC/cm²
	t.Logf("Analytical P₀ = %.2f µC/cm² (%.6f C/m²)", P0uCcm2, P0)

	// Verify E(P₀) ≈ 0 as a sanity check.
	eAtP0 := landauE(P0)
	if math.Abs(eAtP0) > 1e3 {
		t.Errorf("Equilibrium sanity: E(P₀) = %.3e V/m, expected ≈ 0", eAtP0)
	}

	// --- Analytical coercive field Ec ---
	// Ec comes from the S-curve turning points where dE/dP = 0:
	//   2α + 12βP² + 30γP⁴ = 0  →  quadratic in v = P².
	aEc := 30 * gamma
	bEc := 12 * beta
	cEc := 2 * alpha
	discEc := bEc*bEc - 4*aEc*cEc
	if discEc < 0 {
		t.Fatalf("No S-curve turning point: discriminant = %.3e < 0", discEc)
	}
	sqrtDisc := math.Sqrt(discEc)
	vMinus := (-bEc - sqrtDisc) / (2 * aEc)
	vPlus := (-bEc + sqrtDisc) / (2 * aEc)
	var Pc float64
	switch {
	case vMinus > 0:
		Pc = math.Sqrt(vMinus)
	case vPlus > 0:
		Pc = math.Sqrt(vPlus)
	default:
		t.Fatalf("No positive turning point: v− = %.6e, v+ = %.6e", vMinus, vPlus)
	}
	Ec := math.Abs(landauE(Pc))
	EcMVcm := Ec / 1e8
	t.Logf("Analytical Ec = %.4f MV/cm (%.3e V/m, at Pc = %.4f C/m²)", EcMVcm, Ec, Pc)

	// Validate P₀ is in the expected range for HfO₂.
	if P0uCcm2 < 10 || P0uCcm2 > 35 {
		t.Errorf("P₀ = %.1f µC/cm² outside expected range [10, 35]", P0uCcm2)
	}
	// The thermodynamic Ec from raw LGD is much lower than experimental (~0.05–0.2 MV/cm).
	// This is expected and well-documented in the literature.
	if EcMVcm < 0.01 || EcMVcm > 1.0 {
		t.Errorf("Thermodynamic Ec = %.4f MV/cm outside expected range [0.01, 1.0]", EcMVcm)
	}

	// --- L-K solver comparison ---
	// Configure with raw Materlik coefficients, no depolarization or circuit parasitics.
	// Use low viscosity for fast quasi-static convergence at the low thermodynamic Ec.
	s := &LKSolver{
		Beta:                  beta,
		Gamma:                 gamma,
		Alpha:                 alpha,
		UseMaterialAlpha:      true,
		Rho:                   0.005, // Low viscosity for fast equilibration
		K_dep:                 0,     // Pure Landau (no depolarization)
		UseEffectiveViscosity: false,
		SeriesResistance:      0,
		Thickness:             10e-9,
		Area:                  100e-12,
		Temperature:           T,
		CurieTemp:             Tc,
		CurieConst:            C0,
		EnableNoise:           false,
		UseNLS:                false,
		P:                     -P0,
		PMax:                  P0 * 1.5,
	}

	// Sweep to 5× thermodynamic Ec for a clear hysteresis loop.
	Emax := Ec * 5.0
	const (
		nPtsHalf      = 300
		stepsPerPoint = 10000
		dt            = 1e-12
	)

	solverE := make([]float64, 0, 2*nPtsHalf+1)
	solverP := make([]float64, 0, 2*nPtsHalf+1)

	// Ascending: −Emax → +Emax
	for i := 0; i <= nPtsHalf; i++ {
		E := -Emax + 2*Emax*float64(i)/float64(nPtsHalf)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		solverE = append(solverE, E)
		solverP = append(solverP, s.P)
	}
	// Descending: +Emax → −Emax
	for i := 1; i <= nPtsHalf; i++ {
		E := Emax - 2*Emax*float64(i)/float64(nPtsHalf)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		solverE = append(solverE, E)
		solverP = append(solverP, s.P)
	}

	// Diagnostic: log P range and key values.
	pMin, pMax := solverP[0], solverP[0]
	for _, p := range solverP {
		if p < pMin {
			pMin = p
		}
		if p > pMax {
			pMax = p
		}
	}
	t.Logf("Solver P range: [%.4f, %.4f] C/m² (P₀ = %.4f)", pMin, pMax, P0)

	// Extract Pr from solver (P at E≈0 crossings).
	var prVals []float64
	for i := 1; i < len(solverE); i++ {
		if solverE[i-1]*solverE[i] <= 0 && solverE[i-1] != solverE[i] {
			p0, ok := interpYAtX0(solverE[i-1], solverP[i-1], solverE[i], solverP[i])
			if ok {
				prVals = append(prVals, math.Abs(p0))
			}
		}
	}
	if len(prVals) == 0 {
		t.Fatalf("Failed to extract solver Pr from %d E-P points", len(solverE))
	}
	solverPr := 0.0
	for _, v := range prVals {
		solverPr += v
	}
	solverPr /= float64(len(prVals))

	// Extract Ec from solver (E at P≈0 crossings).
	var ecVals []float64
	for i := 1; i < len(solverP); i++ {
		if solverP[i-1]*solverP[i] <= 0 && solverP[i-1] != solverP[i] {
			ec, ok := interpXAtY0(solverP[i-1], solverE[i-1], solverP[i], solverE[i])
			if ok {
				ecVals = append(ecVals, math.Abs(ec))
			}
		}
	}

	// Fallback Ec extraction: find E where |P| is minimum on each branch.
	if len(ecVals) == 0 {
		t.Logf("No P=0 crossings found; using min-|P| fallback for Ec extraction")
		// Ascending branch: first half of points
		ascEnd := nPtsHalf + 1
		bestAbsP := math.Abs(solverP[0])
		bestE := solverE[0]
		for i := 1; i < ascEnd; i++ {
			if ap := math.Abs(solverP[i]); ap < bestAbsP {
				bestAbsP = ap
				bestE = solverE[i]
			}
		}
		ecVals = append(ecVals, math.Abs(bestE))
		// Descending branch
		bestAbsP = math.Abs(solverP[ascEnd])
		bestE = solverE[ascEnd]
		for i := ascEnd + 1; i < len(solverP); i++ {
			if ap := math.Abs(solverP[i]); ap < bestAbsP {
				bestAbsP = ap
				bestE = solverE[i]
			}
		}
		ecVals = append(ecVals, math.Abs(bestE))
	}

	solverEc := 0.0
	for _, v := range ecVals {
		solverEc += v
	}
	solverEc /= float64(len(ecVals))

	solverPruCcm2 := solverPr * 100
	solverEcMVcm := solverEc / 1e8
	t.Logf("Solver quasi-static: Pr = %.2f µC/cm², Ec = %.4f MV/cm", solverPruCcm2, solverEcMVcm)

	// Solver Pr should match analytical P₀ within 15%.
	prRelErr := math.Abs(solverPr-P0) / P0
	t.Logf("Pr relative error: %.1f%%", prRelErr*100)
	if prRelErr > 0.15 {
		t.Errorf("Solver Pr deviates >15%% from analytical: solver=%.2f, analytical=%.2f µC/cm²",
			solverPruCcm2, P0uCcm2)
	}

	// Solver Ec should match analytical within 50% (dynamic Ec > thermodynamic Ec
	// due to viscosity-induced lag, even in quasi-static sweep).
	ecRelErr := math.Abs(solverEc-Ec) / Ec
	t.Logf("Ec relative error: %.1f%%", ecRelErr*100)
	if ecRelErr > 0.50 {
		t.Errorf("Solver Ec deviates >50%% from analytical: solver=%.4f, analytical=%.4f MV/cm",
			solverEcMVcm, EcMVcm)
	}

	// --- Curve shape validation on saturated branches ---
	// For |E| > 2×Ec the system is on a stable branch and the solver's P
	// should satisfy E ≈ E_eq(P) to within a viscosity-induced lag tolerance.
	saturatedMatches := 0
	saturatedTotal := 0
	maxBranchErr := 0.0

	for i, E := range solverE {
		if math.Abs(E) > 2.0*Ec {
			Psol := solverP[i]
			Eeq := landauE(Psol)
			branchErr := math.Abs(E-Eeq) / Emax
			if branchErr > maxBranchErr {
				maxBranchErr = branchErr
			}
			saturatedTotal++
			if branchErr < 0.15 {
				saturatedMatches++
			}
		}
	}

	if saturatedTotal > 0 {
		matchRate := float64(saturatedMatches) / float64(saturatedTotal) * 100
		t.Logf("Saturated branch: %.0f%% match (%d/%d within 15%%, max err %.1f%%)",
			matchRate, saturatedMatches, saturatedTotal, maxBranchErr*100)
		if matchRate < 70 {
			t.Errorf("Saturated branch match rate too low: %.0f%% (need >70%%)", matchRate)
		}
	}
}

// TestLandauEquilibrium_Materlik2015_AnalyticalCurve generates the full analytical
// Landau S-curve for Materlik 2015 parameters and validates its properties.
// This serves as a pure-math reference independent of the solver.
func TestLandauEquilibrium_Materlik2015_AnalyticalCurve(t *testing.T) {
	const (
		beta  = -6.72e8
		gamma = 1.95e10
		Tc    = 598.0
		C0    = 5.3e5
		T     = 300.0
		eps0  = 8.854e-12
	)

	alpha := (T - Tc) / (2 * eps0 * C0)

	// P₀ from equilibrium.
	aQ := 6 * gamma
	bQ := 4 * beta
	cQ := 2 * alpha
	u := (-bQ + math.Sqrt(bQ*bQ-4*aQ*cQ)) / (2 * aQ)
	P0 := math.Sqrt(u)

	nPoints := 500
	Pmax := P0 * 1.1
	eVals := make([]float64, nPoints+1)
	pVals := make([]float64, nPoints+1)
	for i := 0; i <= nPoints; i++ {
		P := -Pmax + 2*Pmax*float64(i)/float64(nPoints)
		P2 := P * P
		P3 := P2 * P
		P5 := P3 * P2
		pVals[i] = P
		eVals[i] = 2*alpha*P + 4*beta*P3 + 6*gamma*P5
	}

	// S-curve properties:
	// 1. E(0) = 0 (origin symmetry)
	midIdx := nPoints / 2
	if math.Abs(eVals[midIdx]) > 1e3 {
		t.Errorf("E(0) = %.3e, expected 0", eVals[midIdx])
	}

	// 2. E(P₀) ≈ 0
	eAtP0 := 2*alpha*P0 + 4*beta*math.Pow(P0, 3) + 6*gamma*math.Pow(P0, 5)
	if math.Abs(eAtP0) > 1e3 {
		t.Errorf("E(P₀) = %.3e, expected ≈ 0", eAtP0)
	}

	// 3. Exactly two local extrema (the coercive field turning points).
	extremaCount := 0
	for i := 1; i < len(eVals)-1; i++ {
		if (eVals[i] > eVals[i-1] && eVals[i] > eVals[i+1]) ||
			(eVals[i] < eVals[i-1] && eVals[i] < eVals[i+1]) {
			extremaCount++
		}
	}
	if extremaCount != 2 {
		t.Errorf("Expected 2 S-curve extrema, found %d", extremaCount)
	}

	// 4. Antisymmetry: E(-P) ≈ -E(P) (odd function).
	for i := 0; i < nPoints/2; i++ {
		j := nPoints - i
		if j >= len(pVals) {
			continue
		}
		symErr := math.Abs(eVals[i] + eVals[j])
		if symErr > 1e3 {
			t.Errorf("Antisymmetry violation at P=%.4f: E(P)=%.3e, E(-P)=%.3e, |sum|=%.3e",
				pVals[j], eVals[j], eVals[i], symErr)
			break
		}
	}

	t.Logf("Analytical S-curve: %d points, P₀=%.2f µC/cm², 2 extrema, antisymmetric",
		len(eVals), P0*100)
}

// TestLandauEquilibrium_ConfiguredMaterial validates the L-K solver against
// the analytical equilibrium using the MaterlikHfO2 preset as configured
// by ConfigureFromMaterial (with Ec-rescaling and Pr-matching).
// This tests the full engine pipeline, not just the raw LGD math.
func TestLandauEquilibrium_ConfiguredMaterial(t *testing.T) {
	mat := MaterlikHfO2()

	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.EnableNoise = false
	s.UseNLS = false
	// Disable depolarization for clean Landau comparison.
	s.K_dep = 0

	// Analytical equilibrium with the solver's actual (rescaled) coefficients.
	landauE := func(P float64) float64 {
		P2 := P * P
		P3 := P2 * P
		P5 := P3 * P2
		return 2*s.Alpha*P + 4*s.Beta*P3 + 6*s.Gamma*P5
	}

	// Find P₀ from the configured coefficients.
	aQ := 6 * s.Gamma
	bQ := 4 * s.Beta
	cQ := 2 * s.Alpha
	disc := bQ*bQ - 4*aQ*cQ
	if disc < 0 {
		t.Fatalf("No equilibrium for configured coefficients")
	}
	u := (-bQ + math.Sqrt(disc)) / (2 * aQ)
	if u <= 0 {
		t.Skipf("Configured coefficients have no positive P₀² (u=%.3e)", u)
	}
	P0 := math.Sqrt(u)
	t.Logf("Configured material P₀ = %.2f µC/cm² (material Pr = %.2f µC/cm²)",
		P0*100, mat.Pr*100)

	// Find thermodynamic Ec from the configured coefficients.
	aEc := 30 * s.Gamma
	bEc := 12 * s.Beta
	cEc := 2 * s.Alpha
	discEc := bEc*bEc - 4*aEc*cEc
	if discEc < 0 {
		t.Skipf("No turning point for configured coefficients")
	}
	sqrtDisc := math.Sqrt(discEc)
	vMinus := (-bEc - sqrtDisc) / (2 * aEc)
	var Pc float64
	if vMinus > 0 {
		Pc = math.Sqrt(vMinus)
	} else {
		vPlus := (-bEc + sqrtDisc) / (2 * aEc)
		if vPlus <= 0 {
			t.Skipf("No positive turning point")
		}
		Pc = math.Sqrt(vPlus)
	}
	Ec := math.Abs(landauE(Pc))
	t.Logf("Configured Ec = %.3f MV/cm (material Ec = %.3f MV/cm)", Ec/1e8, mat.Ec/1e8)

	// Run solver quasi-statically.
	Emax := math.Max(Ec*5.0, mat.Ec*3.0)
	s.SetState(-math.Abs(mat.Pr))

	const (
		nPtsHalf      = 300
		stepsPerPoint = 5000
		dt            = 2e-12
	)

	var solverE, solverP []float64
	for i := 0; i <= nPtsHalf; i++ {
		E := -Emax + 2*Emax*float64(i)/float64(nPtsHalf)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		solverE = append(solverE, E)
		solverP = append(solverP, s.P)
	}
	for i := 1; i <= nPtsHalf; i++ {
		E := Emax - 2*Emax*float64(i)/float64(nPtsHalf)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		solverE = append(solverE, E)
		solverP = append(solverP, s.P)
	}

	// Extract Pr.
	var prVals []float64
	for i := 1; i < len(solverE); i++ {
		if solverE[i-1]*solverE[i] <= 0 && solverE[i-1] != solverE[i] {
			p, ok := interpYAtX0(solverE[i-1], solverP[i-1], solverE[i], solverP[i])
			if ok {
				prVals = append(prVals, math.Abs(p))
			}
		}
	}
	if len(prVals) == 0 {
		t.Fatalf("Failed to extract Pr")
	}
	solverPr := 0.0
	for _, v := range prVals {
		solverPr += v
	}
	solverPr /= float64(len(prVals))

	t.Logf("Solver Pr = %.2f µC/cm² (material Pr = %.2f µC/cm²)", solverPr*100, mat.Pr*100)

	// Pr should be within 20% of the material's advertised Pr.
	prRelErr := math.Abs(solverPr-math.Abs(mat.Pr)) / math.Abs(mat.Pr)
	if prRelErr > 0.20 {
		t.Errorf("Solver Pr deviates >20%% from material Pr: solver=%.2f, material=%.2f µC/cm²",
			solverPr*100, mat.Pr*100)
	}
}
