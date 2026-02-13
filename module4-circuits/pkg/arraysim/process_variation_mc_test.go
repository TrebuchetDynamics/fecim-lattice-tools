package arraysim

import "testing"

func TestRunProcessVariationMC_YieldAndStats(t *testing.T) {
	cfg := ProcessVariationConfig{
		NominalEc:         1.2,
		NominalPr:         25.0,
		VariationFraction: 0.10,
		Samples:           5000,
		Seed:              7,
	}
	got := RunProcessVariationMC(cfg)
	if got.Yield <= 0 || got.Yield > 1 {
		t.Fatalf("invalid yield: %.4f", got.Yield)
	}
	if got.StdEc <= 0 || got.StdPr <= 0 {
		t.Fatalf("expected positive std dev, got Ec=%.6f Pr=%.6f", got.StdEc, got.StdPr)
	}
	if got.PassSamples <= 0 {
		t.Fatalf("expected non-zero passing samples")
	}
}

func TestRunProcessVariationMC_TightMarginLowersYield(t *testing.T) {
	base := ProcessVariationConfig{NominalEc: 1.0, NominalPr: 20, VariationFraction: 0.10, Samples: 3000, Seed: 11, MinReadMarginRatio: 0.80}
	tight := base
	tight.MinReadMarginRatio = 0.98
	r1 := RunProcessVariationMC(base)
	r2 := RunProcessVariationMC(tight)
	if r2.Yield >= r1.Yield {
		t.Fatalf("expected tighter margin to reduce yield: loose=%.4f tight=%.4f", r1.Yield, r2.Yield)
	}
}
