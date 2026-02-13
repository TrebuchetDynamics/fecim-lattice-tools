package export

import (
	"path/filepath"
	"reflect"
	"testing"
)

func deterministicStep(cfg map[string]float64, seeds map[string]int64, action ScenarioAction) float64 {
	base := cfg["gain"] + float64(seeds["main"]%100)
	sum := base
	for _, v := range action.Params {
		sum += v
	}
	return sum
}

func TestScenarioReplay_SaveLoadReplayDeterministic(t *testing.T) {
	state := ScenarioState{
		Config: map[string]float64{"gain": 1.5},
		Seeds:  map[string]int64{"main": 42},
		Actions: []ScenarioAction{
			{Name: "step1", Params: map[string]float64{"x": 0.5}},
			{Name: "step2", Params: map[string]float64{"x": 2.0}},
		},
	}
	for _, a := range state.Actions {
		state.Results = append(state.Results, deterministicStep(state.Config, state.Seeds, a))
	}

	path := filepath.Join(t.TempDir(), "scenario.json")
	if err := SaveScenarioState(path, state); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := LoadScenarioState(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if !reflect.DeepEqual(state, loaded) {
		t.Fatalf("loaded state mismatch\nwant=%+v\ngot=%+v", state, loaded)
	}

	replayed, err := ReplayScenarioDeterministic(loaded, deterministicStep)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	if !reflect.DeepEqual(replayed, loaded.Results) {
		t.Fatalf("replayed results mismatch\nwant=%v\ngot=%v", loaded.Results, replayed)
	}
}
