package arraysim

import "testing"

func TestTierB_BoundaryTerminationPullsFarEnd(t *testing.T) {
	paramsOpen := SolveParams{
		WLVoltages:  []float64{1.0},
		BLVoltages:  []float64{0.0},
		Conductance: [][]float64{{0}, {0}},
		Wire:        WireParams{RWordLine: 1.0, RBitLine: 1.0},
	}
	openRes, err := NewTierBSolver().SolveDC(paramsOpen)
	if err != nil {
		t.Fatalf("open solve: %v", err)
	}

	paramsTerm := paramsOpen
	paramsTerm.Boundary = BoundaryParams{
		WLTerminationResistance: 2.0,
		WLTerminationVoltage:    0.0,
	}
	termRes, err := NewTierBSolver().SolveDC(paramsTerm)
	if err != nil {
		t.Fatalf("terminated solve: %v", err)
	}

	if !(termRes.WLNodes[0][0] < openRes.WLNodes[0][0]) {
		t.Fatalf("expected termination to reduce driven WL node: open=%.6f terminated=%.6f", openRes.WLNodes[0][0], termRes.WLNodes[0][0])
	}
}

func TestTierB_SelectorOffLeakageProducesSmallCurrent(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{1.0},
		BLVoltages: []float64{0.0},
		Conductance: [][]float64{{
			10e-6,
		}},
		Wire:         WireParams{RWordLine: 0.5, RBitLine: 0.5},
		SelectorMode: SelectorRead,
		ReadMask:     [][]bool{{false}},
		Selector: SelectorDeviceParams{
			Enabled:        true,
			OnConductance:  50e-6,
			OffConductance: 2e-9,
		},
	}
	res, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("solve: %v", err)
	}
	if !(res.CellCurrents[0][0] > 0) {
		t.Fatalf("expected non-zero leakage current, got %g", res.CellCurrents[0][0])
	}
	if !(res.CellCurrents[0][0] < 5e-9) {
		t.Fatalf("expected leakage-scale current, got %g", res.CellCurrents[0][0])
	}
}
