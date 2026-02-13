package arraysim

import "testing"

func TestReadMargin_ReflectsCoupledSolverBehavior(t *testing.T) {
	cfg := ArrayConfig{Rows: 8, Cols: 8, CouplingMode: CouplingTierA}
	resTierA := readMarginAnalysisDeterministic(cfg, 4)
	resIdeal := readMarginAnalysisDeterministic(ArrayConfig{Rows: 8, Cols: 8, CouplingMode: CouplingIdeal}, 4)

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

func TestReadMargin_NoisyMarginsAreBelowDeterministic(t *testing.T) {
	cfg := ArrayConfig{Rows: 6, Cols: 6, CouplingMode: CouplingTierA}
	deterministic := readMarginAnalysisDeterministic(cfg, 8)
	noisy := ReadMarginAnalysis(cfg, 8)

	if noisy.MinMarginV > deterministic.MinMarginV {
		t.Fatalf("expected noisy min margin <= deterministic: noisy=%.6g deterministic=%.6g", noisy.MinMarginV, deterministic.MinMarginV)
	}
	if len(noisy.MarginPerLevel) != len(deterministic.MarginPerLevel) {
		t.Fatalf("mismatch margin vector length noisy=%d deterministic=%d", len(noisy.MarginPerLevel), len(deterministic.MarginPerLevel))
	}
	for i := range noisy.MarginPerLevel {
		if noisy.MarginPerLevel[i] > deterministic.MarginPerLevel[i] {
			t.Fatalf("expected noisy margin[%d] <= deterministic margin[%d]: noisy=%.6g deterministic=%.6g", i, i, noisy.MarginPerLevel[i], deterministic.MarginPerLevel[i])
		}
	}
}
