package physics

import "testing"

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
