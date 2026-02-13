package arraysim

import "testing"

func TestReadMargin_ReflectsCoupledSolverBehavior(t *testing.T) {
	resTierA := ReadMarginAnalysis(ArrayConfig{Rows: 8, Cols: 8, CouplingMode: CouplingTierA}, 4)
	resIdeal := ReadMarginAnalysis(ArrayConfig{Rows: 8, Cols: 8, CouplingMode: CouplingIdeal}, 4)

	if resTierA.CouplingMode != CouplingTierA.String() {
		t.Fatalf("unexpected coupling mode: got %q", resTierA.CouplingMode)
	}
	if resTierA.MinMarginV <= 0 {
		t.Fatalf("expected positive margin, got %.6g", resTierA.MinMarginV)
	}
	if resTierA.MinMarginV >= resIdeal.MinMarginV {
		t.Fatalf("expected Tier-A coupled margin < Ideal margin, got tier-a=%.6g ideal=%.6g", resTierA.MinMarginV, resIdeal.MinMarginV)
	}
}
