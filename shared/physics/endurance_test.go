package physics

import (
	"math"
	"testing"
)

func polarizationToConductanceLinear(p, pSat, gMin, gMax float64) float64 {
	if pSat <= 0 {
		return gMin
	}
	norm := (p + pSat) / (2 * pSat)
	if norm < 0 {
		norm = 0
	}
	if norm > 1 {
		norm = 1
	}
	return gMin + norm*(gMax-gMin)
}

func TestPreisachCell_Endurance10000Cycles_DriftWithin5Percent(t *testing.T) {
	const (
		satE      = 1.0
		programE  = 0.75
		readE     = 0.05
		maxCycles = 10000
	)

	mat := FeCIMMaterial()
	ps := NewPreisachStack(satE, simpleUniformEverett{sat: satE})
	pSat := ps.Everett.Calculate(satE, -satE)

	// Program once and establish baseline read conductance.
	ps.Update(programE)
	baseP := ps.Update(readE)
	baseG := polarizationToConductanceLinear(baseP, pSat, mat.Gmin, mat.Gmax)
	if baseG <= 0 {
		t.Fatalf("baseline conductance must be positive, got %e", baseG)
	}

	maxDrift := 0.0
	for i := 0; i < maxCycles; i++ {
		ps.Update(programE) // write/program pulse
		p := ps.Update(readE)

		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Fatalf("invalid polarization at cycle %d: %v", i+1, p)
		}

		g := polarizationToConductanceLinear(p, pSat, mat.Gmin, mat.Gmax)
		drift := math.Abs(g-baseG) / baseG
		if drift > maxDrift {
			maxDrift = drift
		}
	}

	if maxDrift > 0.05 {
		t.Fatalf("conductance drift exceeded 5%%: max drift %.3f%% (baseline=%e S)", maxDrift*100, baseG)
	}
}

func TestPreisachCell_Retention_10YearEquivalentWithinBounds(t *testing.T) {
	const (
		satE          = 1.0
		programE      = 0.80
		readE         = 0.05
		tenYearsSec   = 10 * 365 * 24 * 3600
		referenceTemp = 358.0 // 85C reference for model
	)

	mat := FeCIMMaterial()
	ps := NewPreisachStack(satE, simpleUniformEverett{sat: satE})
	pSat := ps.Everett.Calculate(satE, -satE)

	ps.Update(programE)
	p0 := ps.Update(readE)
	g0 := polarizationToConductanceLinear(p0, pSat, mat.Gmin, mat.Gmax)

	prAged := mat.RetentionAtTime(tenYearsSec, referenceTemp)
	retentionFactor := prAged / mat.Pr

	pFinal := p0 * retentionFactor
	gFinal := polarizationToConductanceLinear(pFinal, pSat, mat.Gmin, mat.Gmax)

	if math.IsNaN(gFinal) || math.IsInf(gFinal, 0) {
		t.Fatalf("invalid retained conductance after 10-year equivalent: %v", gFinal)
	}

	// Model guarantees <=10% Pr loss at retention limit; allow small numerical slack.
	if retentionFactor < 0.89 || retentionFactor > 1.0 {
		t.Fatalf("retention factor out of expected bounds: %.4f", retentionFactor)
	}
	if gFinal <= 0 || gFinal > mat.Gmax {
		t.Fatalf("final conductance out of physical bounds: %e S", gFinal)
	}
	if gFinal < 0.85*g0 {
		t.Fatalf("retained conductance too low: got %e S, want >= %e S", gFinal, 0.85*g0)
	}
}

func TestPreisachCell_Endurance_NoNaNInf_AndNoOverDegradationVsModel(t *testing.T) {
	const (
		satE      = 1.0
		programE  = 0.70
		readE     = 0.05
		maxCycles = 10000
	)

	mat := FeCIMMaterial()
	ps := NewPreisachStack(satE, simpleUniformEverett{sat: satE})
	pSat := ps.Everett.Calculate(satE, -satE)

	ps.Update(programE)
	baseP := ps.Update(readE)
	baseG := polarizationToConductanceLinear(baseP, pSat, mat.Gmin, mat.Gmax)

	checkpoints := map[int]struct{}{
		100:   {},
		1000:  {},
		5000:  {},
		10000: {},
	}

	prevRatio := 1.0
	for n := 1; n <= maxCycles; n++ {
		ps.Update(programE)
		p := ps.Update(readE)

		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Fatalf("invalid polarization at cycle %d: %v", n, p)
		}

		g := polarizationToConductanceLinear(p, pSat, mat.Gmin, mat.Gmax)
		if math.IsNaN(g) || math.IsInf(g, 0) {
			t.Fatalf("invalid conductance at cycle %d: %v", n, g)
		}

		if _, ok := checkpoints[n]; ok {
			observedRatio := g / baseG
			predictedRatio := mat.EnduranceAtCycles(float64(n)) / mat.Pr

			// Do not allow monotonic degradation beyond model prediction envelope.
			if observedRatio < predictedRatio-0.02 {
				t.Fatalf("cycle %d over-degraded: observed ratio %.4f < predicted %.4f", n, observedRatio, predictedRatio)
			}

			if observedRatio > prevRatio+0.05 {
				t.Fatalf("cycle %d shows non-physical jump in retention ratio: prev %.4f -> now %.4f", n, prevRatio, observedRatio)
			}
			prevRatio = observedRatio
		}
	}
}
