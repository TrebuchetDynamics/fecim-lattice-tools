package arraysim

import (
	"math"
	"testing"
)

func TestReferenceSolveDense_KCLResidual_2x2(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{1.0, 0.5},
		BLVoltages: []float64{0.0, 0.2},
		Conductance: [][]float64{
			{1.0, 2.0},
			{0.5, 1.5},
		},
		Wire: WireParams{RWordLine: 1.0, RBitLine: 2.0},
	}

	res, err := referenceSolveDense(params)
	if err != nil {
		t.Fatalf("referenceSolveDense: %v", err)
	}

	maxResidual := kclMaxResidual(params, res)
	if maxResidual > 1e-9 {
		t.Fatalf("KCL residual too large: %g", maxResidual)
	}

	// Row/col currents must equal per-cell sums.
	checkRowColCurrentSums(t, params, res)
}

func TestReferenceSolveDense_KCLResidual_4x4_WithInactiveRow(t *testing.T) {
	g := [][]float64{
		{1, 1, 1, 1},
		{1, 2, 3, 4},
		{0.5, 0.5, 0.5, 0.5},
		{2, 2, 2, 2},
	}
	params := SolveParams{
		WLVoltages:  []float64{1, 1, 0, 0.25},
		BLVoltages:  []float64{0, 0.1, 0.2, 0.3},
		Conductance: g,
		ActiveRows:  []bool{true, false, true, true},
		Wire:        WireParams{RWordLine: 0.8, RBitLine: 0.9},
	}

	res, err := referenceSolveDense(params)
	if err != nil {
		t.Fatalf("referenceSolveDense: %v", err)
	}

	maxResidual := kclMaxResidual(params, res)
	if maxResidual > 1e-9 {
		t.Fatalf("KCL residual too large: %g", maxResidual)
	}

	// Inactive row should contribute zero current.
	if math.Abs(res.RowCurrents[1]) > 1e-12 {
		t.Fatalf("inactive row current not ~0: %g", res.RowCurrents[1])
	}
	for c := 0; c < 4; c++ {
		if math.Abs(res.CellCurrents[1][c]) > 1e-12 {
			t.Fatalf("inactive row cell current not ~0 at (1,%d): %g", c, res.CellCurrents[1][c])
		}
	}

	checkRowColCurrentSums(t, params, res)
}

func checkRowColCurrentSums(t *testing.T, params SolveParams, res DCResult) {
	t.Helper()
	rows := len(params.Conductance)
	cols := len(params.BLVoltages)
	if cols == 0 {
		for _, row := range params.Conductance {
			if len(row) > cols {
				cols = len(row)
			}
		}
	}

	rowSums := make([]float64, rows)
	colSums := make([]float64, cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			i := res.CellCurrents[r][c]
			rowSums[r] += i
			colSums[c] += i
		}
	}
	for r := 0; r < rows; r++ {
		if math.Abs(rowSums[r]-res.RowCurrents[r]) > 1e-12 {
			t.Fatalf("row current mismatch r=%d: sum=%g res=%g", r, rowSums[r], res.RowCurrents[r])
		}
	}
	for c := 0; c < cols; c++ {
		if math.Abs(colSums[c]-res.ColCurrents[c]) > 1e-12 {
			t.Fatalf("col current mismatch c=%d: sum=%g res=%g", c, colSums[c], res.ColCurrents[c])
		}
	}
}

func kclMaxResidual(params SolveParams, res DCResult) float64 {
	rows := len(params.Conductance)
	cols := len(params.BLVoltages)
	if cols == 0 {
		for _, row := range params.Conductance {
			if len(row) > cols {
				cols = len(row)
			}
		}
	}

	geom := params.Geometry.WithDefaults()
	wire := params.Wire.WithDefaults(geom)
	gWL := 1.0 / wire.RWordLine
	gBL := 1.0 / wire.RBitLine

	rowActive := func(r int) bool {
		if params.ActiveRows == nil {
			return true
		}
		if r < 0 || r >= len(params.ActiveRows) {
			return false
		}
		return params.ActiveRows[r]
	}

	maxAbs := 0.0
	abs := func(v float64) float64 {
		if v < 0 {
			return -v
		}
		return v
	}

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			vw := res.WLNodes[r][c]
			vb := res.BLNodes[r][c]

			// WL KCL.
			residWL := 0.0
			if c > 0 {
				residWL += gWL * (vw - res.WLNodes[r][c-1])
			}
			if c < cols-1 {
				residWL += gWL * (vw - res.WLNodes[r][c+1])
			}
			if c == 0 {
				wlV := 0.0
				if r < len(params.WLVoltages) {
					wlV = params.WLVoltages[r]
				}
				residWL += gWL * (vw - wlV)
			}
			gcell := 0.0
			if r < len(params.Conductance) && c < len(params.Conductance[r]) {
				gcell = params.Conductance[r][c]
			}
			if !rowActive(r) {
				gcell = 0
			}
			residWL += gcell * (vw - vb)

			// BL KCL.
			residBL := 0.0
			if r > 0 {
				residBL += gBL * (vb - res.BLNodes[r-1][c])
			}
			if r < rows-1 {
				residBL += gBL * (vb - res.BLNodes[r+1][c])
			}
			if r == 0 {
				blV := 0.0
				if c < len(params.BLVoltages) {
					blV = params.BLVoltages[c]
				}
				residBL += gBL * (vb - blV)
			}
			residBL += gcell * (vb - vw)

			if abs(residWL) > maxAbs {
				maxAbs = abs(residWL)
			}
			if abs(residBL) > maxAbs {
				maxAbs = abs(residBL)
			}
		}
	}

	return maxAbs
}
