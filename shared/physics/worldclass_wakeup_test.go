package physics

import (
	"math"
	"testing"
)

func TestWakeUpPolarization_IncreasesThenFatigues(t *testing.T) {
	cfg := WakeUpModelConfig{
		PrInitial_Cm2:      0.2,
		WakeUpGainFraction: 0.3,
		WakeUpTauCycles:    100,
		FatigueOnsetCycles: 1000,
		FatigueTauCycles:   500,
	}

	p0, err := WakeUpPolarization(0, cfg)
	if err != nil {
		t.Fatalf("WakeUpPolarization error: %v", err)
	}
	p200, _ := WakeUpPolarization(200, cfg)
	p1000, _ := WakeUpPolarization(1000, cfg)
	p3000, _ := WakeUpPolarization(3000, cfg)

	if !(p200 > p0) {
		t.Fatalf("expected wake-up increase: p200=%g p0=%g", p200, p0)
	}
	if !(p1000 >= p200) {
		t.Fatalf("expected continued wake-up before fatigue onset: p1000=%g p200=%g", p1000, p200)
	}
	if !(p3000 < p1000) {
		t.Fatalf("expected fatigue reduction after onset: p3000=%g p1000=%g", p3000, p1000)
	}
}

func TestDefaultWakeUpModelConfig(t *testing.T) {
	pr := 0.24
	cfg := DefaultWakeUpModelConfig(pr)
	if cfg.PrInitial_Cm2 != pr {
		t.Fatalf("expected PrInitial=%g, got %g", pr, cfg.PrInitial_Cm2)
	}
	if cfg.WakeUpTauCycles <= 0 || cfg.FatigueTauCycles <= 0 {
		t.Fatalf("time constants must be positive: wakeTau=%g fatigueTau=%g",
			cfg.WakeUpTauCycles, cfg.FatigueTauCycles)
	}

	// Verify the default config produces valid non-monotonic behavior.
	p0, err := WakeUpPolarization(0, cfg)
	if err != nil {
		t.Fatalf("error at N=0: %v", err)
	}
	pMid, _ := WakeUpPolarization(cfg.WakeUpTauCycles*5, cfg)
	pLate, _ := WakeUpPolarization(cfg.FatigueOnsetCycles+cfg.FatigueTauCycles*3, cfg)

	if pMid <= p0 {
		t.Fatalf("wake-up should increase Pr: p0=%g pMid=%g", p0, pMid)
	}
	if pLate >= pMid {
		t.Fatalf("fatigue should decrease Pr: pMid=%g pLate=%g", pMid, pLate)
	}
}

// --- NonMonotonicEndurance tests ---

func TestNonMonotonicEndurance_ThreePhaseShape(t *testing.T) {
	cfg := DefaultNonMonotonicEnduranceConfig(0.24)

	// Phase 1: Wake-up. Pr should rise from near-zero at N=0 toward the peak.
	pr1, err := NonMonotonicEndurance(1, cfg)
	if err != nil {
		t.Fatalf("error at N=1: %v", err)
	}
	pr100, _ := NonMonotonicEndurance(100, cfg)
	pr1k, _ := NonMonotonicEndurance(1e3, cfg)
	pr10k, _ := NonMonotonicEndurance(1e4, cfg)

	if pr100 <= pr1 {
		t.Fatalf("wake-up phase: Pr(100)=%g should exceed Pr(1)=%g", pr100, pr1)
	}
	if pr1k <= pr100 {
		t.Fatalf("wake-up phase: Pr(1k)=%g should exceed Pr(100)=%g", pr1k, pr100)
	}

	// Phase 2: Plateau region (10^4 to ~10^6). Pr should be near peak.
	pr100k, _ := NonMonotonicEndurance(1e5, cfg)
	if pr100k < 0.9*pr10k {
		t.Fatalf("plateau: Pr(100k)=%g should be near Pr(10k)=%g", pr100k, pr10k)
	}

	// Phase 3: Fatigue. At very high cycle counts, Pr should decline.
	prFatigue, _ := NonMonotonicEndurance(1e10, cfg)
	if prFatigue >= pr10k {
		t.Fatalf("fatigue phase: Pr(1e10)=%g should be < Pr(10k)=%g", prFatigue, pr10k)
	}
}

func TestNonMonotonicEndurance_PeakExceedsEndpoints(t *testing.T) {
	cfg := DefaultNonMonotonicEnduranceConfig(0.24)

	// Find approximate peak by scanning.
	var peakPr float64
	var peakN float64
	for logN := 0.0; logN <= 12.0; logN += 0.1 {
		n := math.Pow(10, logN)
		pr, _ := NonMonotonicEndurance(n, cfg)
		if pr > peakPr {
			peakPr = pr
			peakN = n
		}
	}

	// Peak must be > Pr at N=1 (virgin) and > Pr at N=10^10 (fatigued).
	prVirgin, _ := NonMonotonicEndurance(1, cfg)
	prFatigued, _ := NonMonotonicEndurance(1e10, cfg)

	if peakPr <= prVirgin {
		t.Fatalf("peak Pr=%g at N=%g should exceed virgin Pr=%g", peakPr, peakN, prVirgin)
	}
	if peakPr <= prFatigued {
		t.Fatalf("peak Pr=%g at N=%g should exceed fatigued Pr=%g", peakPr, peakN, prFatigued)
	}
	// Peak should be close to Pr0 (the configured peak value).
	if peakPr > cfg.Pr0*1.01 {
		t.Fatalf("peak Pr=%g exceeds configured Pr0=%g by more than 1%%", peakPr, cfg.Pr0)
	}

	t.Logf("Non-monotonic peak: Pr=%.4f at N=%.0f (virgin=%.4f, fatigued=%.4f)",
		peakPr, peakN, prVirgin, prFatigued)
}

func TestNonMonotonicEndurance_VirginStateIsReduced(t *testing.T) {
	cfg := DefaultNonMonotonicEnduranceConfig(0.24)

	// At N=0, wakeup term = [1-exp(0)]^alpha = 0, so Pr(0) = 0.
	pr0, _ := NonMonotonicEndurance(0, cfg)
	if pr0 != 0 {
		t.Fatalf("expected Pr(0)=0 (no wakeup yet), got %g", pr0)
	}

	// At very small N, Pr should be much less than peak.
	pr10, _ := NonMonotonicEndurance(10, cfg)
	if pr10 >= cfg.Pr0*0.5 {
		t.Fatalf("at N=10, Pr=%g should be well below peak %g", pr10, cfg.Pr0)
	}
}

func TestNonMonotonicEndurance_InvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  NonMonotonicEnduranceConfig
	}{
		{"negative N", NonMonotonicEnduranceConfig{Pr0: 0.2, NWakeup: 1e3, AlphaW: 0.3, NFatigue: 1e8, BetaF: 0.5, GammaF: 0.2}},
		{"zero Pr0", NonMonotonicEnduranceConfig{Pr0: 0, NWakeup: 1e3, AlphaW: 0.3, NFatigue: 1e8, BetaF: 0.5, GammaF: 0.2}},
		{"zero NWakeup", NonMonotonicEnduranceConfig{Pr0: 0.2, NWakeup: 0, AlphaW: 0.3, NFatigue: 1e8, BetaF: 0.5, GammaF: 0.2}},
		{"zero NFatigue", NonMonotonicEnduranceConfig{Pr0: 0.2, NWakeup: 1e3, AlphaW: 0.3, NFatigue: 0, BetaF: 0.5, GammaF: 0.2}},
		{"zero AlphaW", NonMonotonicEnduranceConfig{Pr0: 0.2, NWakeup: 1e3, AlphaW: 0, NFatigue: 1e8, BetaF: 0.5, GammaF: 0.2}},
		{"BetaF=1", NonMonotonicEnduranceConfig{Pr0: 0.2, NWakeup: 1e3, AlphaW: 0.3, NFatigue: 1e8, BetaF: 1.0, GammaF: 0.2}},
		{"zero GammaF", NonMonotonicEnduranceConfig{Pr0: 0.2, NWakeup: 1e3, AlphaW: 0.3, NFatigue: 1e8, BetaF: 0.5, GammaF: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := 100.0
			if tt.name == "negative N" {
				n = -1
			}
			_, err := NonMonotonicEndurance(n, tt.cfg)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}

func TestNonMonotonicEndurance_FatigueFloorAtZero(t *testing.T) {
	cfg := NonMonotonicEnduranceConfig{
		Pr0:      0.2,
		NWakeup:  100,
		AlphaW:   0.3,
		NFatigue: 1e4,
		BetaF:    0.5,
		GammaF:   0.5, // Aggressive fatigue for fast rolloff
	}

	// At extremely high cycles, fatigue term would go negative without clamping.
	prExtreme, err := NonMonotonicEndurance(1e12, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prExtreme < 0 {
		t.Fatalf("Pr must never be negative, got %g at N=1e12", prExtreme)
	}
}

func TestNonMonotonicEndurance_MonotonicDeclineAfterPeak(t *testing.T) {
	cfg := DefaultNonMonotonicEnduranceConfig(0.24)

	// After the wake-up plateau (say N > 10^5), Pr should decline monotonically.
	var prevPr float64
	prevPr, _ = NonMonotonicEndurance(1e5, cfg)

	for logN := 5.5; logN <= 11.0; logN += 0.5 {
		n := math.Pow(10, logN)
		pr, _ := NonMonotonicEndurance(n, cfg)
		if pr > prevPr*1.001 { // Allow tiny numerical tolerance
			t.Fatalf("non-monotonic decline: Pr(%.0e)=%g > Pr(prev)=%g", n, pr, prevPr)
		}
		prevPr = pr
	}
}
