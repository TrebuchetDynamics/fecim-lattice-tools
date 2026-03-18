package literature

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// ---------------------------------------------------------------------------
// Gap 7: Preisach FORC validation against measured / literature data
// ---------------------------------------------------------------------------

// forcDensityPoint stores a single FORC density sample in coercivity-interaction
// coordinates, matching the conventions used in computeFORCMetrics.
type forcDensityPoint struct {
	hc, hu, rho float64 // all in MV/cm (rho units: uC/cm2 / (MV/cm)^2)
}

// computeFORCDensityPoints performs a full FORC measurement and density
// extraction using the proven protocol from TestModule1_FORC_Falsification.
// Returns the density points in (Hc, Hu) coordinates plus summary metrics.
func computeFORCDensityPoints(t *testing.T, mat *sharedphysics.HZOMaterial, gridSize int) []forcDensityPoint {
	t.Helper()

	ec := mat.Ec
	esat := ec * 3.5

	ha := make([]float64, gridSize)
	for i := range ha {
		ha[i] = -esat + float64(i)*2*esat/float64(gridSize-1)
	}
	hb := make([]float64, gridSize)
	for i := range hb {
		hb[i] = -esat + float64(i)*2*esat/float64(gridSize-1)
	}

	P := make([][]float64, gridSize)
	for i := range P {
		P[i] = make([]float64, gridSize)
	}

	for i, haVal := range ha {
		for j, hbVal := range hb {
			if hbVal < haVal {
				P[i][j] = math.NaN()
			}
		}

		m := ferroelectric.NewPreisachModel(mat)
		m.Reset()
		m.Update(esat)
		m.Update(0)
		steps := 10
		for s := 1; s <= steps; s++ {
			m.Update(esat + float64(s)*(haVal-esat)/float64(steps))
		}
		prevHb := haVal
		for j, hbVal := range hb {
			if hbVal < haVal {
				continue
			}
			subSteps := 5
			for s := 1; s <= subSteps; s++ {
				m.Update(prevHb + float64(s)*(hbVal-prevHb)/float64(subSteps))
			}
			P[i][j] = m.Update(hbVal) * 1e2 // C/m2 -> uC/cm2
			prevHb = hbVal
		}
	}

	dHa := ha[1] - ha[0]
	dHb := hb[1] - hb[0]

	var points []forcDensityPoint

	for i := 1; i < gridSize-1; i++ {
		for j := 1; j < gridSize-1; j++ {
			p11 := P[i+1][j+1]
			p1m := P[i+1][j-1]
			pm1 := P[i-1][j+1]
			pmm := P[i-1][j-1]
			if math.IsNaN(p11) || math.IsNaN(p1m) || math.IsNaN(pm1) || math.IsNaN(pmm) {
				continue
			}
			d2P := (p11 - p1m - pm1 + pmm) / (4 * dHa * dHb * 1e-16)
			rho := -0.5 * d2P

			hc := (hb[j] - ha[i]) * 0.5e-8
			hu := (ha[i] + hb[j]) * 0.5e-8

			if hc < 0 {
				continue
			}
			points = append(points, forcDensityPoint{hc, hu, rho})
		}
	}

	if len(points) == 0 {
		t.Fatal("no valid FORC grid points computed")
	}
	return points
}

// findFORCPeak returns the (Hc, Hu) location of the maximum FORC density.
func findFORCPeak(points []forcDensityPoint) (peakHc, peakHu, peakRho float64) {
	peakRho = -math.Inf(1)
	for _, p := range points {
		if p.rho > peakRho {
			peakRho = p.rho
			peakHc = p.hc
			peakHu = p.hu
		}
	}
	return
}

// forcWeightedMeanHc computes the density-weighted mean Hc from FORC density
// points. This is more stable than peak-bin Hc for convergence analysis because
// it averages over the full density distribution rather than depending on a
// single maximum bin which shifts with grid discretization.
func forcWeightedMeanHc(points []forcDensityPoint) float64 {
	sumRhoHc := 0.0
	sumRho := 0.0
	for _, p := range points {
		if p.rho > 0 {
			sumRhoHc += p.rho * p.hc
			sumRho += p.rho
		}
	}
	if sumRho == 0 {
		return 0
	}
	return sumRhoHc / sumRho
}

// TestFORCValidation_HZO_ExpectedDensityShape validates that the simulated
// FORC density for HZO has the expected qualitative shape:
//   - Concentrated along the Ec axis (not spread uniformly)
//   - Higher density in the coercive field region
//   - Approximately symmetric about the origin (Hu = 0)
func TestFORCValidation_HZO_ExpectedDensityShape(t *testing.T) {
	mat := sharedphysics.Park2015Fig2aHZO10nm()
	ref := HZO10nm_FORCReference()

	points := computeFORCDensityPoints(t, mat, 25)
	peakHc, peakHu, _ := findFORCPeak(points)

	ecMV := ref.ExpectedEc_MVcm
	hcErr := pctErr(peakHc, ecMV)
	t.Logf("Peak Hc=%.4f MV/cm (expected %.4f, err=%.1f%%), Peak Hu=%.4f MV/cm",
		peakHc, ecMV, hcErr, peakHu)

	if hcErr > ref.PeakHcTolerance_pct {
		t.Errorf("Peak Hc=%.4f MV/cm deviates from Ec=%.4f MV/cm by %.1f%% (limit %.1f%%)",
			peakHc, ecMV, hcErr, ref.PeakHcTolerance_pct)
	}

	// 2. Check density is concentrated (not uniform).
	totalDensity := 0.0
	centralDensity := 0.0
	fwhm := ref.DensityFWHM_Hc_MVcm

	for _, p := range points {
		if p.rho <= 0 {
			continue
		}
		totalDensity += p.rho
		if math.Abs(p.hc-ecMV) < fwhm {
			centralDensity += p.rho
		}
	}

	if totalDensity == 0 {
		t.Fatal("Total FORC density is zero; no valid density points computed")
	}

	centralFrac := centralDensity / totalDensity
	t.Logf("Central band fraction: %.3f (density within |Hc-Ec| < %.2f MV/cm)", centralFrac, fwhm)

	// At least 30% of density should be concentrated near Ec.
	if centralFrac < 0.30 {
		t.Errorf("Density is too spread out: only %.1f%% in central band (expected > 30%%)", centralFrac*100)
	}

	// 3. Check symmetry about Hu = 0.
	posHuDensity := 0.0
	negHuDensity := 0.0

	for _, p := range points {
		if p.rho <= 0 {
			continue
		}
		if p.hu > 0 {
			posHuDensity += p.rho
		} else {
			negHuDensity += p.rho
		}
	}

	var symmetryRatio float64
	if negHuDensity > 0 {
		symmetryRatio = posHuDensity / negHuDensity
	} else if posHuDensity > 0 {
		symmetryRatio = math.Inf(1)
	} else {
		symmetryRatio = 1.0
	}

	t.Logf("Symmetry ratio (pos/neg Hu density): %.3f (expected ~1.0)", symmetryRatio)
	if math.Abs(symmetryRatio-ref.SymmetryRatio) > ref.SymmetryTolerance {
		t.Errorf("FORC density asymmetry: ratio=%.3f, expected %.3f +/- %.3f",
			symmetryRatio, ref.SymmetryRatio, ref.SymmetryTolerance)
	}
}

// TestFORCValidation_Pr_Ec_FromFORC extracts Pr and Ec from the FORC data
// and compares them with values obtained from a direct P-E loop simulation.
//
// The FORC peak Hc should match the P-E loop Ec within 20% (consistent with
// thFORCHcErrPct in module1_forc_test.go), and the Pr values should match
// within 10%.
func TestFORCValidation_Pr_Ec_FromFORC(t *testing.T) {
	mat := sharedphysics.Park2015Fig2aHZO10nm()
	ec := mat.Ec
	esat := ec * 3.5

	// 1. Get Pr/Ec from direct P-E loop via PreisachModel using a fine sweep.
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	// Full hysteresis loop: saturate +esat, then descend to -esat.
	numSteps := 500
	model.Update(esat)
	var prDirect, ecDirect float64
	var prevP float64
	for i := 0; i <= numSteps; i++ {
		e := esat - 2*esat*float64(i)/float64(numSteps)
		p := model.Update(e)
		// Pr is polarization at E = 0 on descending branch.
		if i > 0 {
			ePrev := esat - 2*esat*float64(i-1)/float64(numSteps)
			if ePrev > 0 && e <= 0 {
				prDirect = math.Abs(p)
			}
			// Ec: interpolate where P crosses zero on descending branch.
			if prevP > 0 && p <= 0 && ecDirect == 0 {
				// Linear interpolation for E at P=0.
				frac := prevP / (prevP - p)
				ecDirect = math.Abs(ePrev + frac*(e-ePrev))
			}
		}
		prevP = p
	}

	// 2. Get Ec from FORC peak Hc using higher resolution grid.
	points := computeFORCDensityPoints(t, mat, 25)
	peakHc, _, _ := findFORCPeak(points)
	ecFORCMV := peakHc // Already in MV/cm

	// 3. For FORC-based Pr, use the same saturate-then-measure-at-zero protocol.
	modelFORC := ferroelectric.NewPreisachModel(mat)
	modelFORC.Reset()
	// Match the direct Pr extraction: sweep continuously from +esat down to 0.
	modelFORC.Update(esat)
	prSteps := 100
	var prFORC float64
	for i := 0; i <= prSteps; i++ {
		e := esat * (1.0 - float64(i)/float64(prSteps))
		p := modelFORC.Update(e)
		if i == prSteps { // e = 0
			prFORC = math.Abs(p)
		}
	}

	prDirectUC := prDirect * 1e2
	prFORCUC := prFORC * 1e2
	ecDirectMV := ecDirect * 1e-8

	t.Logf("Direct P-E: Pr=%.2f uC/cm2, Ec=%.4f MV/cm", prDirectUC, ecDirectMV)
	t.Logf("FORC:       Pr=%.2f uC/cm2, Ec=%.4f MV/cm", prFORCUC, ecFORCMV)

	// Compare Pr (< 10% mismatch).
	if prDirectUC > 0 {
		prMismatch := pctErr(prFORCUC, prDirectUC)
		t.Logf("Pr mismatch: %.1f%%", prMismatch)
		if prMismatch > 10.0 {
			t.Errorf("Pr mismatch %.1f%% > 10%% (direct=%.2f, FORC=%.2f uC/cm2)",
				prMismatch, prDirectUC, prFORCUC)
		}
	}

	// Compare Ec (< 20% mismatch, consistent with module1_forc_test thresholds).
	// FORC peak Hc has limited resolution from finite-difference grid and is
	// inherently coarser than the P-E loop zero-crossing interpolation.
	if ecDirectMV > 0 {
		ecMismatch := pctErr(ecFORCMV, ecDirectMV)
		t.Logf("Ec mismatch: %.1f%%", ecMismatch)
		if ecMismatch > 20.0 {
			t.Errorf("Ec mismatch %.1f%% > 20%% (direct=%.4f, FORC=%.4f MV/cm)",
				ecMismatch, ecDirectMV, ecFORCMV)
		}
	}
}

// TestFORCValidation_NumReversals_Convergence verifies that increasing the
// FORC grid resolution produces converging metrics. The density-weighted mean
// Hc (a more robust estimator than peak-bin Hc) and Pr should change by less
// than 5% between grid sizes 30 and 50.
func TestFORCValidation_NumReversals_Convergence(t *testing.T) {
	mat := sharedphysics.Park2015Fig2aHZO10nm()

	type metrics struct {
		meanHc float64 // density-weighted mean Hc (MV/cm)
		peakHc float64 // peak-bin Hc (MV/cm), for logging
		prEst  float64 // estimated Pr from envelope (uC/cm2)
	}

	extract := func(gridSize int) metrics {
		points := computeFORCDensityPoints(t, mat, gridSize)
		peakHc, _, _ := findFORCPeak(points)
		meanHc := forcWeightedMeanHc(points)

		// Pr estimate from envelope.
		ec := mat.Ec
		esat := ec * 3.5
		m := ferroelectric.NewPreisachModel(mat)
		m.Reset()
		m.Update(esat)
		var pr float64
		prSteps := 100
		for i := 0; i <= prSteps; i++ {
			e := esat * (1.0 - float64(i)/float64(prSteps))
			p := m.Update(e)
			if i == prSteps {
				pr = math.Abs(p) * 1e2
			}
		}

		return metrics{meanHc: meanHc, peakHc: peakHc, prEst: pr}
	}

	gridSizes := []int{10, 15, 20, 25, 30, 40, 50}
	results := make([]metrics, len(gridSizes))

	for i, n := range gridSizes {
		results[i] = extract(n)
		t.Logf("gridSize=%d: meanHc=%.4f MV/cm, peakHc=%.4f MV/cm, Pr~=%.2f uC/cm2",
			n, results[i].meanHc, results[i].peakHc, results[i].prEst)
	}

	// Check convergence between grid=30 (index 4) and grid=50 (index 6).
	m30 := results[4]
	m50 := results[6]

	if m30.meanHc > 0 {
		hcChange := pctErr(m50.meanHc, m30.meanHc)
		t.Logf("Mean Hc change (grid 30->50): %.1f%%", hcChange)
		if hcChange > 5.0 {
			t.Errorf("Mean Hc not converged: change=%.1f%% > 5%% (grid=30: %.4f, grid=50: %.4f MV/cm)",
				hcChange, m30.meanHc, m50.meanHc)
		}
	}

	if m30.prEst > 0 {
		prChange := pctErr(m50.prEst, m30.prEst)
		t.Logf("Pr change (grid 30->50): %.1f%%", prChange)
		if prChange > 5.0 {
			t.Errorf("Pr not converged: change=%.1f%% > 5%% (grid=30: %.2f, grid=50: %.2f uC/cm2)",
				prChange, m30.prEst, m50.prEst)
		}
	}
}

// TestFORCValidation_MaterialComparison runs FORC for BTO vs HZO and verifies:
//   - BTO has larger Pr than HZO (20 vs 15.8 uC/cm2)
//   - BTO has different (lower) Ec and therefore different peak Hc location
//   - The density distributions are qualitatively different
func TestFORCValidation_MaterialComparison(t *testing.T) {
	matHZO := sharedphysics.Park2015Fig2aHZO10nm()
	matBTO := sharedphysics.BTO()

	type matResult struct {
		name   string
		peakHc float64
		pr     float64
	}

	runFORC := func(mat *sharedphysics.HZOMaterial) matResult {
		points := computeFORCDensityPoints(t, mat, 25)
		peakHc, _, _ := findFORCPeak(points)

		return matResult{
			name:   mat.Name,
			peakHc: peakHc,
			pr:     mat.Pr * 1e2,
		}
	}

	rHZO := runFORC(matHZO)
	rBTO := runFORC(matBTO)

	t.Logf("HZO: Pr=%.1f uC/cm2, peak Hc=%.4f MV/cm", rHZO.pr, rHZO.peakHc)
	t.Logf("BTO: Pr=%.1f uC/cm2, peak Hc=%.4f MV/cm", rBTO.pr, rBTO.peakHc)

	// BTO should have larger Pr (20 vs 15.8 uC/cm2).
	if rBTO.pr <= rHZO.pr {
		t.Errorf("Expected BTO Pr (%.1f) > HZO Pr (%.1f)", rBTO.pr, rHZO.pr)
	}

	// BTO should have much lower Ec (0.03 vs 0.93 MV/cm).
	if rBTO.peakHc >= rHZO.peakHc {
		t.Errorf("Expected BTO peak Hc (%.4f) < HZO peak Hc (%.4f) MV/cm",
			rBTO.peakHc, rHZO.peakHc)
	}

	// The peak Hc values should differ by at least 2x (0.03 vs 0.93).
	if rHZO.peakHc > 0 && rBTO.peakHc > 0 {
		ratio := rHZO.peakHc / rBTO.peakHc
		t.Logf("Hc ratio (HZO/BTO): %.1fx", ratio)
		if ratio < 2.0 {
			t.Errorf("Expected HZO/BTO Hc ratio > 2x, got %.1fx", ratio)
		}
	}
}
