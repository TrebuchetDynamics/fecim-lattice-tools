package gui

import (
	"reflect"
	"testing"
)

func TestM5Data04_SensitivityRankingSorted(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	missing := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	delete(missing.Inputs, "gpu_tco_usd_per_month")
	if SensitivityRanking(missing, "tco_reduction_pct") != nil {
		t.Fatal("missing denominator admitted")
	}
	r := SensitivityRanking(run, "tco_reduction_pct")
	for i := 0; i < 20; i++ {
		if !reflect.DeepEqual(r, SensitivityRanking(run, "tco_reduction_pct")) {
			t.Fatal("sensitivity ties are nondeterministic")
		}
	}
	if len(r) == 0 {
		t.Fatal("expected non-empty ranking")
	}
	if r[0].ImpactPct < r[len(r)-1].ImpactPct {
		t.Fatal("expected sorted descending by impact")
	}
}
