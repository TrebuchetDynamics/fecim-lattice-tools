package gui

import "testing"

func TestM5Data03_ConfidenceIsNotAStatisticalInterval(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	for key, v := range run.Outputs {
		if v.Uncertainty != 0 {
			t.Fatalf("%s claims unsupported interval", key)
		}
	}
	cfg := PresetScenarioConfig(ScenarioBaseline)
	cfg.GPUEnergyPJPerMAC = 0
	invalid := BuildScenarioRun(ScenarioBaseline, cfg)
	if invalid.Outputs["energy_reduction_pct"].UnavailableReason == "" {
		t.Fatal("missing denominator exposed as numeric output")
	}
	if SensitivityRanking(run, "co2_reduction_pct") != nil {
		t.Fatal("currency-based CO2 must be withheld")
	}
}
