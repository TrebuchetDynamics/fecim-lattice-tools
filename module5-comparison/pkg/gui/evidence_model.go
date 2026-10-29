package gui

import model "fecim-lattice-tools/module5-comparison/pkg/comparison"

type TaggedValue = model.TaggedValue
type ScenarioRun = model.ScenarioRun
type SensitivityImpact = model.SensitivityImpact
type ScenarioDiff = model.ScenarioDiff

func BuildScenarioRun(p ScenarioProfile, c ScenarioConfig) ScenarioRun {
	return model.BuildScenarioRun(p, c)
}
func SensitivityRanking(r ScenarioRun, output string) []SensitivityImpact {
	return model.SensitivityRanking(r, output)
}
func DiffScenarios(a, b ScenarioRun) ScenarioDiff          { return model.DiffScenarios(a, b) }
func PlainTextEvidence(title string, r ScenarioRun) string { return model.PlainTextEvidence(title, r) }
