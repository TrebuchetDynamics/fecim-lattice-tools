package export

import (
	"encoding/json"
	"fmt"
	"os"
)

// ScenarioAction is one replayable action in an experiment trace.
type ScenarioAction struct {
	Name   string             `json:"name"`
	Params map[string]float64 `json:"params,omitempty"`
}

// ScenarioState stores the full deterministic replay payload.
type ScenarioState struct {
	Config  map[string]float64 `json:"config"`
	Seeds   map[string]int64   `json:"seeds"`
	Actions []ScenarioAction   `json:"actions"`
	Results []float64          `json:"results"`
}

// SaveScenarioState persists the full scenario state to JSON.
func SaveScenarioState(path string, state ScenarioState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal scenario state: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write scenario state: %w", err)
	}
	return nil
}

// LoadScenarioState restores scenario state from JSON.
func LoadScenarioState(path string) (ScenarioState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ScenarioState{}, fmt.Errorf("read scenario state: %w", err)
	}
	var s ScenarioState
	if err := json.Unmarshal(data, &s); err != nil {
		return ScenarioState{}, fmt.Errorf("unmarshal scenario state: %w", err)
	}
	return s, nil
}

// ReplayScenarioDeterministic recomputes results from actions using a deterministic
// step function and verifies exact match with recorded results.
func ReplayScenarioDeterministic(state ScenarioState, stepFn func(cfg map[string]float64, seeds map[string]int64, action ScenarioAction) float64) ([]float64, error) {
	if stepFn == nil {
		return nil, fmt.Errorf("step function is required")
	}
	replayed := make([]float64, len(state.Actions))
	for i, action := range state.Actions {
		replayed[i] = stepFn(state.Config, state.Seeds, action)
		if i < len(state.Results) && replayed[i] != state.Results[i] {
			return replayed, fmt.Errorf("non-deterministic replay at step %d: got %.9g want %.9g", i, replayed[i], state.Results[i])
		}
	}
	if len(state.Results) != len(replayed) {
		return replayed, fmt.Errorf("result length mismatch: got %d replayed vs %d recorded", len(replayed), len(state.Results))
	}
	return replayed, nil
}
