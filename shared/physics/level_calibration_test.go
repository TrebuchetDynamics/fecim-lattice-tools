package physics

import "testing"

func TestEstimateUniformLevelFieldsLabelsHeuristicAndProducesMonotonicRange(t *testing.T) {
	result, err := EstimateUniformLevelFields(LevelCalibrationInput{
		Material:        DefaultHZO(),
		LevelCount:      16,
		TargetRangeFrac: 0.70,
		TemperatureK:    300,
	})
	if err != nil {
		t.Fatalf("EstimateUniformLevelFields: %v", err)
	}
	if result.Method != LevelCalibrationMethodUniformFieldSpacing {
		t.Fatalf("Method = %q, want %q", result.Method, LevelCalibrationMethodUniformFieldSpacing)
	}
	if len(result.AscendingFields) != 16 || len(result.DescendingFields) != 16 {
		t.Fatalf("entry counts = %d/%d, want 16/16", len(result.AscendingFields), len(result.DescendingFields))
	}
	if !result.AscendingMonotonic || !result.DescendingMonotonic {
		t.Fatalf("monotonicity = %v/%v, want true/true", result.AscendingMonotonic, result.DescendingMonotonic)
	}
	if result.AscendingMinField != result.DescendingMinField || result.AscendingMaxField != result.DescendingMaxField {
		t.Fatal("uniform field-spacing estimate should report the same range for both directions")
	}
	if !(result.AscendingMinField < 0 && result.AscendingMaxField > 0) {
		t.Fatalf("field range = %.3e..%.3e, want negative-to-positive span", result.AscendingMinField, result.AscendingMaxField)
	}
}

func TestCalibrateLevelsLKCompatibilityAliasIsHonest(t *testing.T) {
	result, err := CalibrateLevelsLK(LevelCalibrationInput{
		Material:        DefaultHZO(),
		LevelCount:      4,
		TargetRangeFrac: 0.70,
		TemperatureK:    300,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Method != LevelCalibrationMethodUniformFieldSpacing {
		t.Fatalf("compatibility method label = %q, want %q", result.Method, LevelCalibrationMethodUniformFieldSpacing)
	}
}

func TestEstimateUniformLevelFieldsRejectsInvalidInputs(t *testing.T) {
	base := LevelCalibrationInput{Material: DefaultHZO(), LevelCount: 30, TargetRangeFrac: 0.90, TemperatureK: 300}
	for _, tc := range []struct {
		name  string
		input LevelCalibrationInput
	}{
		{name: "nil material", input: LevelCalibrationInput{LevelCount: 30, TargetRangeFrac: 0.90, TemperatureK: 300}},
		{name: "level count too low", input: LevelCalibrationInput{Material: base.Material, LevelCount: 1, TargetRangeFrac: base.TargetRangeFrac, TemperatureK: base.TemperatureK}},
		{name: "target range too high", input: LevelCalibrationInput{Material: base.Material, LevelCount: base.LevelCount, TargetRangeFrac: 1.20, TemperatureK: base.TemperatureK}},
		{name: "temperature too low", input: LevelCalibrationInput{Material: base.Material, LevelCount: base.LevelCount, TargetRangeFrac: base.TargetRangeFrac, TemperatureK: 100}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := EstimateUniformLevelFields(tc.input); err == nil {
				t.Fatal("EstimateUniformLevelFields accepted invalid input")
			}
		})
	}
}
