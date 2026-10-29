package gui

import model "fecim-lattice-tools/module5-comparison/pkg/comparison"

// Compatibility aliases keep the native scenario protocol on the shared Go model.
type ScenarioProfile = model.ScenarioProfile
type ScenarioConfig = model.ScenarioConfig

const (
	ScenarioConservative = model.ScenarioConservative
	ScenarioBaseline     = model.ScenarioBaseline
	ScenarioOptimistic   = model.ScenarioOptimistic
)

func PresetScenarioConfig(p ScenarioProfile) ScenarioConfig { return model.PresetScenarioConfig(p) }
func ScenarioProfileFromString(s string) ScenarioProfile    { return model.ScenarioProfileFromString(s) }
