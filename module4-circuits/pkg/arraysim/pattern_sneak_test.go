package arraysim

import "testing"

func TestPatternWorstCaseSneak_CurrentRatioIdentifiableAndBounded(t *testing.T) {
	const n = 8
	targetRow := 3
	targetCol := 2

	levels := GenerateAllOnes(n, n, 30)
	for r := 0; r < n; r++ {
		levels[r][targetCol] = 0
	}

	params := solveParamsFromPattern(levels, 1e-6, 1e-4)
	for r := range params.WLVoltages {
		params.WLVoltages[r] = 0
	}
	params.WLVoltages[targetRow] = 0.6

	res, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC failed: %v", err)
	}

	signal := absf(res.ColCurrents[targetCol])
	sneak := 0.0
	for c := 0; c < n; c++ {
		if c == targetCol {
			continue
		}
		sneak += absf(res.ColCurrents[c])
	}
	if signal <= 0 {
		t.Fatalf("target signal is zero; cannot evaluate sneak ratio")
	}
	ratio := sneak / signal

	if sneak <= 1e-8 {
		t.Fatalf("sneak current not identifiable: sneak=%g A", sneak)
	}
	if ratio >= 1e4 {
		t.Fatalf("sneak ratio unbounded: ratio=%g sneak=%g A signal=%g A", ratio, sneak, signal)
	}
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
