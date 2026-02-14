package arraysim

import (
	"math"
	"strconv"
	"testing"
)

func TestPatternWalkingOnes_KCLNoNaNAndBounds(t *testing.T) {
	for _, n := range []int{4, 8} {
		t.Run("size_"+strconv.Itoa(n), func(t *testing.T) {
			for pos := 0; pos < n*n; pos++ {
				pattern := GenerateWalkingOnes(n, n, pos, 30)
				params := solveParamsFromPattern(pattern, 1e-6, 1e-4)
				res, err := NewTierBSolver().SolveDC(params)
				if err != nil {
					t.Fatalf("SolveDC failed at pos=%d: %v", pos, err)
				}
				assertFiniteDCResult(t, res)
				assertKCLInternalNodes(t, params, res, 1e-6)
				assertCurrentBounds(t, res, n, 1e-4, 0.6)
			}
		})
	}
}

func TestPatternWalkingZeros_KCLNoNaNAndBounds(t *testing.T) {
	for _, n := range []int{4, 8} {
		t.Run("size_"+strconv.Itoa(n), func(t *testing.T) {
			for pos := 0; pos < n*n; pos++ {
				pattern := GenerateWalkingZeros(n, n, pos, 30)
				params := solveParamsFromPattern(pattern, 1e-6, 1e-4)
				res, err := NewTierBSolver().SolveDC(params)
				if err != nil {
					t.Fatalf("SolveDC failed at pos=%d: %v", pos, err)
				}
				assertFiniteDCResult(t, res)
				assertKCLInternalNodes(t, params, res, 1e-6)
				assertCurrentBounds(t, res, n, 1e-4, 0.6)
			}
		})
	}
}

func solveParamsFromPattern(levels [][]int, gmin, gmax float64) SolveParams {
	rows := len(levels)
	cols := len(levels[0])
	g := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		g[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			g[r][c] = levelToConductance(levels[r][c], 29, gmin, gmax)
		}
	}
	wl := make([]float64, rows)
	bl := make([]float64, cols)
	for i := range wl {
		wl[i] = 0.6
	}
	return SolveParams{
		WLVoltages:  wl,
		BLVoltages:  bl,
		Conductance: g,
		Wire:        WireParams{RWordLine: 5.0, RBitLine: 7.5},
		Boundary:    BoundaryParams{WLDriveResistance: 2.0, BLDriveResistance: 2.0},
	}
}

func levelToConductance(level, maxLevel int, gmin, gmax float64) float64 {
	if maxLevel <= 0 {
		return gmin
	}
	if level < 0 {
		level = 0
	}
	if level > maxLevel {
		level = maxLevel
	}
	return gmin + (gmax-gmin)*float64(level)/float64(maxLevel)
}

func assertFiniteDCResult(t *testing.T, res DCResult) {
	t.Helper()
	check2D := func(name string, m [][]float64) {
		for r := range m {
			for c := range m[r] {
				v := m[r][c]
				if math.IsNaN(v) || math.IsInf(v, 0) {
					t.Fatalf("%s has non-finite value at [%d,%d]: %v", name, r, c, v)
				}
			}
		}
	}
	check1D := func(name string, a []float64) {
		for i, v := range a {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("%s has non-finite value at [%d]: %v", name, i, v)
			}
		}
	}
	check2D("WLNodes", res.WLNodes)
	check2D("BLNodes", res.BLNodes)
	check2D("CellVoltages", res.CellVoltages)
	check2D("CellCurrents", res.CellCurrents)
	check1D("RowCurrents", res.RowCurrents)
	check1D("ColCurrents", res.ColCurrents)
}

func assertCurrentBounds(t *testing.T, res DCResult, cols int, gmax, vread float64) {
	t.Helper()
	imaxCell := gmax * vread
	imaxRow := float64(cols) * imaxCell
	for r := range res.CellCurrents {
		for c := range res.CellCurrents[r] {
			if math.Abs(res.CellCurrents[r][c]) > imaxCell*1.05 {
				t.Fatalf("cell current out of bound at [%d,%d]: got=%g A limit=%g A", r, c, res.CellCurrents[r][c], imaxCell)
			}
		}
	}
	for r := range res.RowCurrents {
		if math.Abs(res.RowCurrents[r]) > imaxRow*1.05 {
			t.Fatalf("row current out of bound at [%d]: got=%g A limit=%g A", r, res.RowCurrents[r], imaxRow)
		}
	}
}
