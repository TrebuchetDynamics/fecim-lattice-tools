package physics

import "testing"

func TestApplyFrequencyDispersion(t *testing.T) {
	base := HysteresisMetrics{FrequencyHz: 1e3, Pr_Cm2: 0.25, Ec_Vm: 1e6, LoopArea_Jm3: 5e5}
	cfg := FrequencyDispersionConfig{
		ReferenceHz:        1e3,
		EcLogSlope:         0.04,
		PrLogSlope:         -0.03,
		LoopAreaLogSlope:   0.02,
		MinMultiplierClamp: 0.5,
	}
	low, err := ApplyFrequencyDispersion(base, 1e2, cfg)
	if err != nil {
		t.Fatalf("low freq error: %v", err)
	}
	high, err := ApplyFrequencyDispersion(base, 1e5, cfg)
	if err != nil {
		t.Fatalf("high freq error: %v", err)
	}

	if !(high.Ec_Vm > base.Ec_Vm && low.Ec_Vm < base.Ec_Vm) {
		t.Fatalf("Ec should increase with frequency; low=%g base=%g high=%g", low.Ec_Vm, base.Ec_Vm, high.Ec_Vm)
	}
	if !(high.Pr_Cm2 < base.Pr_Cm2 && low.Pr_Cm2 > base.Pr_Cm2) {
		t.Fatalf("Pr should decrease with frequency; low=%g base=%g high=%g", low.Pr_Cm2, base.Pr_Cm2, high.Pr_Cm2)
	}
}
