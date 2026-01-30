package ferroelectric

import (
	"math"
	"testing"
)

func TestMonotonicityEnforcement(t *testing.T) {
	cases := []struct {
		name     string
		input    []float64
		expected bool
	}{
		{"AlreadyMonotonic", []float64{0.1, 0.2, 0.3, 0.4, 0.5}, true},
		{"SingleSpike", []float64{0.1, 0.2, 0.15, 0.3, 0.4}, false},
		{"MultipleSpikes", []float64{0.1, 0.3, 0.2, 0.4, 0.35, 0.5}, false},
		{"AllSame", []float64{0.2, 0.2, 0.2, 0.2}, false},
		{"Descending", []float64{0.5, 0.4, 0.3, 0.2, 0.1}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isMonotonic := true
			for i := 1; i < len(tc.input); i++ {
				if tc.input[i] <= tc.input[i-1] {
					isMonotonic = false
					break
				}
			}
			if isMonotonic != tc.expected {
				t.Errorf("Monotonicity check failed: expected %v, got %v", tc.expected, isMonotonic)
			}
		})
	}
}

func assertMonotonic(t *testing.T, name string, values []float64) bool {
	t.Helper()
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			t.Errorf("%s not monotonic at [%d]: %.6e < [%d]: %.6e",
				name, i, values[i], i-1, values[i-1])
			return false
		}
	}
	return true
}

func assertMonotonicDescending(t *testing.T, name string, values []float64) bool {
	t.Helper()
	for i := 1; i < len(values); i++ {
		if values[i] > values[i-1] {
			t.Errorf("%s not monotonic descending at [%d]: %.6e > [%d]: %.6e",
				name, i, values[i], i-1, values[i-1])
			return false
		}
	}
	return true
}

func TestMayergoyzLevelMapping(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	levels := 30
	states := model.DiscreteStates(levels)

	if len(states) != levels {
		t.Errorf("Expected %d discrete states, got %d", levels, len(states))
	}

	pValues := make([]float64, len(states))
	for i, s := range states {
		pValues[i] = s.Polarization
	}

	assertMonotonic(t, "DiscreteStates.Polarization", pValues)

	for i := 1; i < len(pValues); i++ {
		gap := pValues[i] - pValues[i-1]
		expectedGap := 2 * material.Ps / float64(levels-1)
		tolerance := expectedGap * 0.50

		if math.Abs(gap-expectedGap) > tolerance {
			t.Errorf("Non-uniform level spacing at [%d]: gap=%.4e, expected=%.4e",
				i, gap, expectedGap)
		}
	}
}

func TestGridResolutionVsLevelCount(t *testing.T) {
	material := DefaultHZO()

	gridSizes := []int{20, 30, 40, 50, 60, 80, 100}

	for _, gridSize := range gridSizes {
		t.Run("grid"+string(rune('0'+gridSize/10))+string(rune('0'+gridSize%10)), func(t *testing.T) {
			model := NewMayergoyzPreisach(material, gridSize)

			Emax := material.Ec * 2.0
			_, P := model.GetHysteresisLoop(Emax, 500)

			uniqueP := make(map[float64]struct{})
			for _, p := range P {
				rounded := math.Round(p*1e6) / 1e6
				uniqueP[rounded] = struct{}{}
			}

			minRequired := 30

			if len(uniqueP) < minRequired {
				t.Logf("Grid size %d provides %d unique P values (recommend >=%d)",
					gridSize, len(uniqueP), minRequired)
			}
		})
	}
}

func TestPolarizationVsFieldMonotonicity(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec
	steps := 100

	var prevP float64 = -material.Ps

	for i := 0; i <= steps; i++ {
		E := -Ec + 2*Ec*float64(i)/float64(steps)
		P := model.Update(E)

		if E > Ec*0.2 && P < prevP-material.Ps*0.01 {
			t.Errorf("P decreased while E increased in saturation approach: E=%.4e, P=%.4e < prevP=%.4e",
				E, P, prevP)
		}

		prevP = P
	}
}

func TestDescendingBranchMonotonicity(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 2.0
	model.Update(Emax)

	steps := 100
	prevP := model.Polarization()

	for i := 1; i <= steps; i++ {
		E := Emax - 2*Emax*float64(i)/float64(steps)
		P := model.Update(E)

		if E < -material.Ec*0.2 && P > prevP+material.Ps*0.01 {
			t.Errorf("P increased while E decreased in saturation approach: E=%.4e, P=%.4e > prevP=%.4e",
				E, P, prevP)
		}

		prevP = P
	}
}

func TestSwitchingRegionBehavior(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec
	Ps := material.Ps

	model.Update(-Ec * 2.0)
	P_negsat := model.Polarization()

	var switchingStarted, switchingComplete bool
	var P_at_Ec float64

	for i := 0; i <= 100; i++ {
		E := -Ec*0.5 + Ec*2.0*float64(i)/100
		P := model.Update(E)

		if !switchingStarted && P > P_negsat+Ps*0.1 {
			switchingStarted = true
		}

		if math.Abs(E-Ec) < Ec*0.05 {
			P_at_Ec = P
		}

		if !switchingComplete && P > Ps*0.9 {
			switchingComplete = true
		}
	}

	if !switchingStarted {
		t.Error("Switching did not start in expected field range")
	}

	if math.Abs(P_at_Ec) > Ps*0.20 {
		t.Errorf("P at E=Ec should be near zero for symmetric material: P=%.4e", P_at_Ec)
	}

	if !switchingComplete {
		t.Error("Switching did not complete in expected field range")
	}
}

func TestCalibrationMonotonicityAfterBinarySearch(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	numLevels := 30
	calibUp := make([]float64, numLevels)
	calibDown := make([]float64, numLevels)

	Ec := material.Ec

	for level := 1; level < numLevels; level++ {
		targetP := -material.Ps + 2*material.Ps*float64(level)/float64(numLevels-1)

		model.Reset()
		model.Update(-Ec * 3)

		lowE := 0.0
		highE := Ec * 3

		for iter := 0; iter < 20; iter++ {
			midE := (lowE + highE) / 2
			model.Reset()
			model.Update(-Ec * 3)
			P := model.Update(midE)

			if P < targetP {
				lowE = midE
			} else {
				highE = midE
			}

			if highE-lowE < Ec*0.001 {
				break
			}
		}

		calibUp[level] = (lowE + highE) / 2
	}

	for level := numLevels - 2; level >= 0; level-- {
		targetP := -material.Ps + 2*material.Ps*float64(level)/float64(numLevels-1)

		model.Reset()
		model.Update(Ec * 3)

		lowE := -Ec * 3
		highE := 0.0

		for iter := 0; iter < 20; iter++ {
			midE := (lowE + highE) / 2
			model.Reset()
			model.Update(Ec * 3)
			P := model.Update(midE)

			if P > targetP {
				lowE = midE
			} else {
				highE = midE
			}

			if highE-lowE < Ec*0.001 {
				break
			}
		}

		calibDown[level] = (lowE + highE) / 2
	}

	assertMonotonic(t, "calibrationUp", calibUp[1:numLevels])
	assertMonotonicDescending(t, "calibrationDown", calibDown[:numLevels-1])
}

func BenchmarkMonotonicityCheck(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 1; j < len(data); j++ {
			if data[j] <= data[j-1] {
				break
			}
		}
	}
}
