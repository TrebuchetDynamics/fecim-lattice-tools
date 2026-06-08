package arraysim

import (
	"math"
	"testing"
)

func TestModule4ArraySimE2EDCSolverDenseTierBOracleMatrix(t *testing.T) {
	base := SolveParams{
		WLVoltages:  []float64{0.3, 0.15, 0.05},
		BLVoltages:  []float64{0, 0.02, 0.01, 0},
		Conductance: [][]float64{{20e-6, 35e-6, 50e-6, 65e-6}, {80e-6, 25e-6, 45e-6, 70e-6}, {30e-6, 55e-6, 75e-6, 95e-6}},
		Geometry:    DefaultCellGeometry(),
		Wire:        WireParams{RWordLine: 0.8, RBitLine: 3.2},
		Boundary: BoundaryParams{
			WLDriveResistance:       1.1,
			BLDriveResistance:       3.5,
			WLTerminationResistance: 40,
			BLTerminationResistance: 70,
			WLTerminationVoltage:    0.01,
			BLTerminationVoltage:    0,
		},
	}
	cases := []struct {
		name   string
		mutate func(*SolveParams)
	}{
		{name: "base", mutate: func(*SolveParams) {}},
		{name: "active_rows", mutate: func(p *SolveParams) { p.ActiveRows = []bool{true, false, true} }},
		{name: "read_mask", mutate: func(p *SolveParams) {
			p.SelectorMode = SelectorRead
			p.ReadMask = [][]bool{{true, false, true, true}, {false, false, false, false}, {true, true, false, true}}
		}},
		{name: "write_mask", mutate: func(p *SolveParams) {
			p.SelectorMode = SelectorWrite
			p.WriteMask = [][]bool{{true, true, false, true}, {true, false, true, false}, {false, true, true, true}}
		}},
		{name: "selector_ron", mutate: func(p *SolveParams) { p.SelectorEnabled = true; p.SelectorRon = 1500 }},
		{name: "selector_device", mutate: func(p *SolveParams) {
			p.SelectorMode = SelectorRead
			p.ReadMask = [][]bool{{true, false, true, false}, {true, true, false, true}, {false, true, true, true}}
			p.Selector = SelectorDeviceParams{Enabled: true, OnConductance: 1.0 / 2e3, OffConductance: 1e-12}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := cloneSolveParamsE2E(base)
			tc.mutate(&params)
			dense, err := referenceSolveDense(params)
			if err != nil {
				t.Fatalf("referenceSolveDense: %v", err)
			}
			tierB, err := NewTierBSolver().SolveDC(params)
			if err != nil {
				t.Fatalf("TierB SolveDC: %v", err)
			}
			assertDCShapeE2E(t, dense, 3, 4)
			assertDCShapeE2E(t, tierB, 3, 4)
			assertSolveCloseE2E(t, dense.SolveResult, tierB.SolveResult, 5e-8, 5e-11)
			assertNodeCloseE2E(t, dense.WLNodes, tierB.WLNodes, 5e-8)
			assertNodeCloseE2E(t, dense.BLNodes, tierB.BLNodes, 5e-8)
			assertKCLPowerContractE2E(t, tierB.SolveResult)
		})
	}
}

func TestModule4ArraySimE2EDCSolverFailureAndDegenerateContracts(t *testing.T) {
	if got, err := referenceSolveDense(SolveParams{}); err != nil || len(got.CellVoltages) != 0 {
		t.Fatalf("empty dense solve contract changed: got=%+v err=%v", got, err)
	}
	if got, err := NewTierBSolver().SolveDC(SolveParams{}); err != nil || len(got.CellVoltages) != 0 {
		t.Fatalf("empty TierB solve contract changed: got=%+v err=%v", got, err)
	}
	if got, err := referenceSolveDense(SolveParams{Conductance: [][]float64{{1e-6}}, Wire: WireParams{RWordLine: -1, RBitLine: -1}}); err != nil || len(got.CellVoltages) != 1 {
		t.Fatalf("dense solve should default negative wire parameters: got=%+v err=%v", got, err)
	}
	if got, err := NewTierBSolver().SolveDC(SolveParams{Conductance: [][]float64{{1e-6}}, Wire: WireParams{RWordLine: -1, RBitLine: -1}}); err != nil || len(got.CellVoltages) != 1 {
		t.Fatalf("TierB should default negative wire parameters: got=%+v err=%v", got, err)
	}
	if _, err := gaussianElimSolve([][]float64{{1, 2}}, []float64{1}); err == nil {
		t.Fatal("gaussianElimSolve should reject non-square matrix")
	}
	if _, err := gaussianElimSolve([][]float64{{0, 0}, {0, 0}}, []float64{1, 1}); err == nil {
		t.Fatal("gaussianElimSolve should reject singular matrix")
	}
	if _, err := pcgSolve(func(x, y []float64) {}, []float64{0}, []float64{1}, 1, 1e-8, 1e-12); err == nil {
		t.Fatal("pcgSolve should reject non-positive diagonal")
	}
	params := SolveParams{
		WLVoltages:  []float64{0.2},
		BLVoltages:  []float64{0},
		Conductance: [][]float64{{50e-6}},
		Geometry:    DefaultCellGeometry(),
		Wire:        WireParams{RWordLine: 0.5, RBitLine: 2.0},
	}
	if _, err := (&TierBSolver{MaxIter: 1, RelativeTolerance: 1e-30, AbsoluteTolerance: 1e-30}).SolveDC(params); err == nil {
		t.Fatal("TierB should expose non-convergence under impossible tolerance/iteration budget")
	}
}

func cloneSolveParamsE2E(in SolveParams) SolveParams {
	out := in
	out.WLVoltages = append([]float64(nil), in.WLVoltages...)
	out.BLVoltages = append([]float64(nil), in.BLVoltages...)
	out.ActiveRows = append([]bool(nil), in.ActiveRows...)
	out.Conductance = cloneFloatMatrixE2E(in.Conductance)
	out.ReadMask = cloneBoolMatrixE2E(in.ReadMask)
	out.WriteMask = cloneBoolMatrixE2E(in.WriteMask)
	return out
}

func cloneFloatMatrixE2E(in [][]float64) [][]float64 {
	if in == nil {
		return nil
	}
	out := make([][]float64, len(in))
	for i := range in {
		out[i] = append([]float64(nil), in[i]...)
	}
	return out
}

func cloneBoolMatrixE2E(in [][]bool) [][]bool {
	if in == nil {
		return nil
	}
	out := make([][]bool, len(in))
	for i := range in {
		out[i] = append([]bool(nil), in[i]...)
	}
	return out
}

func assertDCShapeE2E(t *testing.T, dc DCResult, rows, cols int) {
	t.Helper()
	if len(dc.CellVoltages) != rows || len(dc.CellCurrents) != rows || len(dc.WLNodes) != rows || len(dc.BLNodes) != rows || len(dc.RowCurrents) != rows || len(dc.ColCurrents) != cols {
		t.Fatalf("DC shape invalid: %+v", dc)
	}
	for r := 0; r < rows; r++ {
		if len(dc.CellVoltages[r]) != cols || len(dc.CellCurrents[r]) != cols || len(dc.WLNodes[r]) != cols || len(dc.BLNodes[r]) != cols {
			t.Fatalf("DC row %d shape invalid: %+v", r, dc)
		}
	}
}

func assertSolveCloseE2E(t *testing.T, a, b SolveResult, voltTol, currentTol float64) {
	t.Helper()
	for r := range a.CellVoltages {
		for c := range a.CellVoltages[r] {
			if math.Abs(a.CellVoltages[r][c]-b.CellVoltages[r][c]) > voltTol {
				t.Fatalf("cell voltage[%d][%d] mismatch: %g vs %g", r, c, a.CellVoltages[r][c], b.CellVoltages[r][c])
			}
			if math.Abs(a.CellCurrents[r][c]-b.CellCurrents[r][c]) > currentTol {
				t.Fatalf("cell current[%d][%d] mismatch: %g vs %g", r, c, a.CellCurrents[r][c], b.CellCurrents[r][c])
			}
		}
	}
}

func assertNodeCloseE2E(t *testing.T, a, b [][]float64, tol float64) {
	t.Helper()
	for r := range a {
		for c := range a[r] {
			if math.Abs(a[r][c]-b[r][c]) > tol {
				t.Fatalf("node[%d][%d] mismatch: %g vs %g", r, c, a[r][c], b[r][c])
			}
		}
	}
}

func assertKCLPowerContractE2E(t *testing.T, result SolveResult) {
	t.Helper()
	rowSum := 0.0
	colSum := 0.0
	for _, v := range result.RowCurrents {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("row current invalid: %g", v)
		}
		rowSum += v
	}
	for _, v := range result.ColCurrents {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("col current invalid: %g", v)
		}
		colSum += v
	}
	if math.Abs(rowSum-colSum) > 1e-10 {
		t.Fatalf("row/col current mismatch: row=%g col=%g", rowSum, colSum)
	}
}
