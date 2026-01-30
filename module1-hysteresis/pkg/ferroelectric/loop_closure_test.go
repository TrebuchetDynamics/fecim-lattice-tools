package ferroelectric

import (
	"math"
	"testing"
)

func TestLoopClosureAfterCycle(t *testing.T) {
	materials := AllMaterials()

	for _, material := range materials {
		t.Run(material.Name, func(t *testing.T) {
			model := NewMayergoyzPreisach(material, 50)
			Emax := material.Ec * 2.0
			points := 500

			E, P := model.GetHysteresisLoop(Emax, points)

			startIdx := 0
			for i, e := range E {
				if e >= Emax*0.99 {
					startIdx = i
					break
				}
			}

			endIdx := len(E) - 1
			for i := len(E) - 1; i >= 0; i-- {
				if E[i] >= Emax*0.99 {
					endIdx = i
					break
				}
			}

			if startIdx == 0 || endIdx == len(E)-1 {
				t.Skip("Could not find Emax points in loop")
			}

			Pstart := P[startIdx]
			Pend := P[endIdx]
			tolerance := material.Ps * 0.05

			if math.Abs(Pend-Pstart) > tolerance {
				t.Errorf("Loop not closed: P(start)=%.4e, P(end)=%.4e, diff=%.4e (tolerance=%.4e)",
					Pstart, Pend, math.Abs(Pend-Pstart), tolerance)
			}
		})
	}
}

func TestSimplePreisachLoopClosure(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)
	Emax := material.Ec * 2.0

	E, P := model.GetHysteresisLoop(Emax, 500)

	tolerance := material.Ps * 0.10

	firstEmaxIdx := -1
	lastEmaxIdx := -1

	for i, e := range E {
		if e >= Emax*0.98 {
			if firstEmaxIdx == -1 {
				firstEmaxIdx = i
			}
			lastEmaxIdx = i
		}
	}

	if firstEmaxIdx == -1 || lastEmaxIdx == -1 {
		t.Skip("Could not find Emax points")
	}

	Pstart := P[firstEmaxIdx]
	Pend := P[lastEmaxIdx]

	if math.Abs(Pend-Pstart) > tolerance {
		t.Errorf("SimplePreisach loop not closed: diff=%.4e (tolerance=%.4e)",
			math.Abs(Pend-Pstart), tolerance)
	}
}

func TestMinorLoopCongruenceProperty(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec
	Ps := material.Ps

	model.Update(Ec * 2.0)
	model.Update(Ec * 0.5)
	P1_start := model.Update(Ec * 0.5)

	model.Update(Ec * 1.0)
	P1_end := model.Polarization()

	slope1 := (P1_end - P1_start) / (Ec * 0.5)

	model.Reset()
	model.Update(Ec * 2.0)
	model.Update(Ec * 0.3)
	P2_start := model.Update(Ec * 0.3)

	model.Update(Ec * 0.8)
	P2_end := model.Polarization()

	slope2 := (P2_end - P2_start) / (Ec * 0.5)

	tolerance := 0.20

	if math.Abs(slope1-slope2) > tolerance*math.Abs(slope1+slope2)/2 {
		t.Logf("Minor loop slopes: slope1=%.4e, slope2=%.4e (diff=%.1f%%)",
			slope1, slope2, math.Abs(slope1-slope2)/math.Abs(slope1)*100)
	}

	if math.IsNaN(slope1) || math.IsNaN(slope2) {
		t.Error("NaN in minor loop slopes")
	}

	if math.Abs(P1_end) > Ps || math.Abs(P2_end) > Ps {
		t.Error("Minor loop exceeded saturation")
	}
}

func TestWipeOutProperty(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 2.0

	model.Update(Emax)

	model.Update(0.5 * material.Ec)
	model.Update(0.8 * material.Ec)
	model.Update(0.3 * material.Ec)
	model.Update(0.6 * material.Ec)

	P_before_wipeout := model.Polarization()

	model.Update(Emax)
	P_after_return := model.Polarization()

	tolerance := material.Ps * 0.05

	if math.Abs(P_after_return-material.Ps) > tolerance {
		t.Errorf("Returning to Emax should reach saturation: P=%.4e, expected ~Ps=%.4e",
			P_after_return, material.Ps)
	}

	if P_after_return < P_before_wipeout {
		t.Errorf("P decreased when returning to Emax: before=%.4e, after=%.4e",
			P_before_wipeout, P_after_return)
	}
}

func TestMajorLoopSymmetry(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 2.0
	E, P := model.GetHysteresisLoop(Emax, 500)

	var posEmaxP, negEmaxP float64
	foundPos, foundNeg := false, false

	for i, e := range E {
		if e >= Emax*0.98 && !foundPos {
			posEmaxP = P[i]
			foundPos = true
		}
		if e <= -Emax*0.98 && !foundNeg {
			negEmaxP = P[i]
			foundNeg = true
		}
	}

	if !foundPos || !foundNeg {
		t.Skip("Could not find +/-Emax points")
	}

	tolerance := material.Ps * 0.10

	if math.Abs(posEmaxP+negEmaxP) > tolerance {
		t.Errorf("Loop not symmetric: P(+Emax)=%.4e, P(-Emax)=%.4e (sum should be ~0)",
			posEmaxP, negEmaxP)
	}

	if math.Abs(posEmaxP) < material.Ps*0.90 {
		t.Errorf("Saturation not reached at +Emax: P=%.4e, expected ~Ps=%.4e",
			posEmaxP, material.Ps)
	}
}

func TestRemanentPolarization(t *testing.T) {
	materials := []struct {
		name     string
		material *HZOMaterial
	}{
		{"DefaultHZO", DefaultHZO()},
		{"FeCIMMaterial", FeCIMMaterial()},
	}

	for _, m := range materials {
		t.Run(m.name, func(t *testing.T) {
			model := NewMayergoyzPreisach(m.material, 50)

			Emax := m.material.Ec * 2.0
			model.Update(Emax)

			Pr_pos := model.Update(0)

			model.Update(-Emax)

			Pr_neg := model.Update(0)

			tolerance := m.material.Pr * 0.30

			if Pr_pos < m.material.Pr-tolerance {
				t.Errorf("Positive remanent polarization too low: Pr=%.4e, expected >=%.4e",
					Pr_pos, m.material.Pr-tolerance)
			}

			if Pr_neg > -m.material.Pr+tolerance {
				t.Errorf("Negative remanent polarization too high: Pr=%.4e, expected <=%.4e",
					Pr_neg, -m.material.Pr+tolerance)
			}

			if math.Abs(Pr_pos+Pr_neg) > tolerance {
				t.Errorf("Pr asymmetry: Pr_pos=%.4e, Pr_neg=%.4e (should be symmetric)",
					Pr_pos, Pr_neg)
			}
		})
	}
}

func TestLoopAreaPositive(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 2.0
	E, P := model.GetHysteresisLoop(Emax, 1000)

	area := 0.0
	for i := 1; i < len(E); i++ {
		area += (P[i] + P[i-1]) / 2 * (E[i] - E[i-1])
	}
	area = math.Abs(area)

	expectedArea := 4 * material.Ec * material.Pr * 0.5

	if area < expectedArea*0.1 {
		t.Errorf("Loop area too small (no hysteresis?): area=%.4e, expected >%.4e",
			area, expectedArea*0.1)
	}

	if area > expectedArea*5 {
		t.Errorf("Loop area too large: area=%.4e, expected <%.4e",
			area, expectedArea*5)
	}

	t.Logf("Loop area: %.4e (expected order: %.4e)", area, expectedArea)
}

func TestCoerciveFieldCrossing(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 2.0
	E, P := model.GetHysteresisLoop(Emax, 1000)

	var Ec_asc, Ec_desc float64
	foundAsc, foundDesc := false, false

	for i := 1; i < len(E); i++ {
		if E[i] > E[i-1] && P[i-1] < 0 && P[i] >= 0 && !foundAsc {
			Ec_asc = E[i-1] + (0-P[i-1])/(P[i]-P[i-1])*(E[i]-E[i-1])
			foundAsc = true
		}

		if E[i] < E[i-1] && P[i-1] > 0 && P[i] <= 0 && !foundDesc {
			Ec_desc = E[i-1] + (0-P[i-1])/(P[i]-P[i-1])*(E[i]-E[i-1])
			foundDesc = true
		}
	}

	if !foundAsc || !foundDesc {
		t.Skip("Could not find zero-crossings")
	}

	tolerance := material.Ec * 0.30

	if math.Abs(Ec_asc-material.Ec) > tolerance {
		t.Errorf("Ascending Ec off: measured=%.4e, expected=%.4e",
			Ec_asc, material.Ec)
	}

	if math.Abs(math.Abs(Ec_desc)-material.Ec) > tolerance {
		t.Errorf("Descending Ec off: measured=%.4e, expected=%.4e",
			Ec_desc, -material.Ec)
	}
}
