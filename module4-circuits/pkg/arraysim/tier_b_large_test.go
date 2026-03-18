package arraysim

import (
	"math"
	"math/rand"
	"testing"
)

// TestTierBSolver_128x128_Convergence verifies that the PCG solver converges
// for a 128x128 array with random conductances and voltages.
func TestTierBSolver_128x128_Convergence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large array test in short mode")
	}

	const (
		rows = 128
		cols = 128
	)

	rng := rand.New(rand.NewSource(42))

	// Random conductances in the 1-100 uS range.
	cond := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		cond[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			cond[r][c] = (1.0 + 99.0*rng.Float64()) * 1e-6
		}
	}

	// Random WL voltages in [0, 1V]; BL voltages all 0V.
	wlV := make([]float64, rows)
	for r := 0; r < rows; r++ {
		wlV[r] = rng.Float64()
	}
	blV := make([]float64, cols) // all zeros

	params := SolveParams{
		WLVoltages:  wlV,
		BLVoltages:  blV,
		Conductance: cond,
	}

	res, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC 128x128: %v", err)
	}

	// Verify output dimensions.
	assertGridDims(t, "CellVoltages", res.CellVoltages, rows, cols)
	assertGridDims(t, "CellCurrents", res.CellCurrents, rows, cols)
	assertGridDims(t, "WLNodes", res.WLNodes, rows, cols)
	assertGridDims(t, "BLNodes", res.BLNodes, rows, cols)

	if len(res.RowCurrents) != rows {
		t.Fatalf("RowCurrents len=%d want %d", len(res.RowCurrents), rows)
	}
	if len(res.ColCurrents) != cols {
		t.Fatalf("ColCurrents len=%d want %d", len(res.ColCurrents), cols)
	}

	t.Logf("128x128 solve succeeded; sample WLNode[0][0]=%.6g V, BLNode[127][127]=%.6g V",
		res.WLNodes[0][0], res.BLNodes[rows-1][cols-1])
}

// TestTierBSolver_128x128_KCLValidation verifies Kirchhoff's Current Law
// at 10 randomly selected internal nodes of a 128x128 solution.
func TestTierBSolver_128x128_KCLValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large array test in short mode")
	}

	const (
		rows     = 128
		cols     = 128
		kclTol   = 1e-4
		numSpots = 10
	)

	rng := rand.New(rand.NewSource(42))

	cond := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		cond[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			cond[r][c] = (1.0 + 99.0*rng.Float64()) * 1e-6
		}
	}

	wlV := make([]float64, rows)
	for r := 0; r < rows; r++ {
		wlV[r] = rng.Float64()
	}
	blV := make([]float64, cols)

	params := SolveParams{
		WLVoltages:  wlV,
		BLVoltages:  blV,
		Conductance: cond,
	}

	res, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC 128x128: %v", err)
	}

	wire := params.Wire.WithDefaults(params.Geometry.WithDefaults())
	gWL := 1.0 / wire.RWordLine
	gBL := 1.0 / wire.RBitLine

	// Use a separate RNG for node selection to keep conductance generation
	// deterministic regardless of how many nodes we sample.
	spotRng := rand.New(rand.NewSource(99))
	for i := 0; i < numSpots; i++ {
		// Pick internal WL nodes (r any, c in [1, cols-2]).
		r := spotRng.Intn(rows)
		c := 1 + spotRng.Intn(cols-2)

		vw := res.WLNodes[r][c]
		residWL := gWL*(vw-res.WLNodes[r][c-1]) +
			gWL*(vw-res.WLNodes[r][c+1]) +
			effectiveCellConductance(params, r, c)*(vw-res.BLNodes[r][c])

		if math.Abs(residWL) > kclTol {
			t.Fatalf("KCL WL node[%d,%d] residual=%g A (tol=%g)", r, c, residWL, kclTol)
		}

		// Pick internal BL nodes (r in [1, rows-2], c any).
		r2 := 1 + spotRng.Intn(rows-2)
		c2 := spotRng.Intn(cols)

		vb := res.BLNodes[r2][c2]
		residBL := gBL*(vb-res.BLNodes[r2-1][c2]) +
			gBL*(vb-res.BLNodes[r2+1][c2]) +
			effectiveCellConductance(params, r2, c2)*(vb-res.WLNodes[r2][c2])

		if math.Abs(residBL) > kclTol {
			t.Fatalf("KCL BL node[%d,%d] residual=%g A (tol=%g)", r2, c2, residBL, kclTol)
		}
	}

	t.Logf("KCL validated at %d WL + %d BL internal nodes (tol=%g)", numSpots, numSpots, kclTol)
}

// TestTierBSolver_128x128_SubblockVsStandalone verifies that a 4x4 sub-block
// extracted from the 128x128 solution differs from a standalone 4x4 solve
// (because of coupling from surrounding cells), but both are individually
// valid per KCL.
func TestTierBSolver_128x128_SubblockVsStandalone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large array test in short mode")
	}

	const (
		rows    = 128
		cols    = 128
		subSize = 4
		subR0   = 60 // top-left corner of the sub-block
		subC0   = 60
	)

	rng := rand.New(rand.NewSource(42))

	cond := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		cond[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			cond[r][c] = (1.0 + 99.0*rng.Float64()) * 1e-6
		}
	}

	wlV := make([]float64, rows)
	for r := 0; r < rows; r++ {
		wlV[r] = rng.Float64()
	}
	blV := make([]float64, cols)

	fullParams := SolveParams{
		WLVoltages:  wlV,
		BLVoltages:  blV,
		Conductance: cond,
	}

	fullRes, err := NewTierBSolver().SolveDC(fullParams)
	if err != nil {
		t.Fatalf("SolveDC 128x128: %v", err)
	}

	// Extract the 4x4 sub-block conductances and corresponding voltages.
	subCond := make([][]float64, subSize)
	for r := 0; r < subSize; r++ {
		subCond[r] = make([]float64, subSize)
		copy(subCond[r], cond[subR0+r][subC0:subC0+subSize])
	}

	subWLV := make([]float64, subSize)
	copy(subWLV, wlV[subR0:subR0+subSize])

	subBLV := make([]float64, subSize)
	copy(subBLV, blV[subC0:subC0+subSize])

	subParams := SolveParams{
		WLVoltages:  subWLV,
		BLVoltages:  subBLV,
		Conductance: subCond,
	}

	subRes, err := referenceSolveDense(subParams)
	if err != nil {
		t.Fatalf("referenceSolveDense 4x4: %v", err)
	}

	// The sub-block from the full solution should NOT match the standalone
	// 4x4 solution due to coupling from the surrounding 128x128 network.
	anyDiff := false
	for r := 0; r < subSize; r++ {
		for c := 0; c < subSize; c++ {
			fullCellV := fullRes.CellVoltages[subR0+r][subC0+c]
			subCellV := subRes.CellVoltages[r][c]
			if math.Abs(fullCellV-subCellV) > 1e-6 {
				anyDiff = true
				break
			}
		}
		if anyDiff {
			break
		}
	}
	if !anyDiff {
		t.Fatal("expected sub-block cell voltages to differ from standalone 4x4 due to coupling, but they matched")
	}

	// Both solutions should satisfy KCL independently.
	fullKCL := kclMaxResidual(fullParams, fullRes)
	if fullKCL > 1e-4 {
		t.Fatalf("full 128x128 KCL residual too large: %g", fullKCL)
	}

	subKCL := kclMaxResidual(subParams, subRes)
	if subKCL > 1e-9 {
		t.Fatalf("standalone 4x4 KCL residual too large: %g", subKCL)
	}

	t.Logf("sub-block coupling verified: full KCL=%g, standalone KCL=%g", fullKCL, subKCL)
}

// assertGridDims checks that a 2D slice has the expected dimensions.
func assertGridDims(t *testing.T, name string, grid [][]float64, wantRows, wantCols int) {
	t.Helper()
	if len(grid) != wantRows {
		t.Fatalf("%s rows=%d want %d", name, len(grid), wantRows)
	}
	for r, row := range grid {
		if len(row) != wantCols {
			t.Fatalf("%s row %d cols=%d want %d", name, r, len(row), wantCols)
		}
	}
}
