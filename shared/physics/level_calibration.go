package physics

import (
	"fmt"
	"math"
)

const LevelCalibrationMethodLKQuasiStatic = "LK quasi-static simulation"

// LevelCalibrationInput describes the UI-neutral inputs for simulated level calibration.
type LevelCalibrationInput struct {
	Material        *HZOMaterial
	LevelCount      int
	TargetRangeFrac float64
	TemperatureK    float64
}

// LevelCalibrationResult summarizes simulated field mappings for discrete levels.
type LevelCalibrationResult struct {
	MaterialName    string
	LevelCount      int
	TargetRangeFrac float64
	TemperatureK    float64
	Method          string

	AscendingFields  []float64
	DescendingFields []float64

	AscendingMinField  float64
	AscendingMaxField  float64
	DescendingMinField float64
	DescendingMaxField float64

	AscendingMonotonic  bool
	DescendingMonotonic bool
}

// CalibrateLevelsLK computes a deterministic, UI-neutral simulated level mapping
// using the same LK-oriented material path as the default hysteresis viewmodel.
func CalibrateLevelsLK(input LevelCalibrationInput) (LevelCalibrationResult, error) {
	if input.Material == nil {
		return LevelCalibrationResult{}, fmt.Errorf("level calibration: material is required")
	}
	if input.LevelCount < 2 || input.LevelCount > 64 {
		return LevelCalibrationResult{}, fmt.Errorf("level calibration: level count %d outside [2,64]", input.LevelCount)
	}
	if !isFiniteLevelCalibrationValue(input.TargetRangeFrac) || input.TargetRangeFrac < 0.5 || input.TargetRangeFrac > 1.0 {
		return LevelCalibrationResult{}, fmt.Errorf("level calibration: target range %.3f outside [0.5,1.0]", input.TargetRangeFrac)
	}
	if !isFiniteLevelCalibrationValue(input.TemperatureK) || input.TemperatureK < 200 || input.TemperatureK > 700 {
		return LevelCalibrationResult{}, fmt.Errorf("level calibration: temperature %.1f K outside [200,700]", input.TemperatureK)
	}

	effectiveEc := levelCalibrationEffectiveEc(input.Material, input.TemperatureK)
	if !isFiniteLevelCalibrationValue(effectiveEc) || effectiveEc <= 0 {
		return LevelCalibrationResult{}, fmt.Errorf("level calibration: non-positive effective coercive field at %.1f K", input.TemperatureK)
	}
	effectiveP := levelCalibrationEffectivePolarization(input.Material, input.TemperatureK)
	if !isFiniteLevelCalibrationValue(effectiveP) || effectiveP <= 0 {
		return LevelCalibrationResult{}, fmt.Errorf("level calibration: non-positive effective polarization at %.1f K", input.TemperatureK)
	}

	// Exercise the LK material path so invalid material/runtime combinations fail
	// before a summary is reported as fresh.
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(input.Material)
	solver.Temperature = input.TemperatureK
	solver.UpdateParams()
	if !isFiniteLevelCalibrationValue(solver.Alpha) {
		return LevelCalibrationResult{}, fmt.Errorf("level calibration: invalid LK runtime parameters at %.1f K", input.TemperatureK)
	}

	ascending := make([]float64, input.LevelCount)
	descending := make([]float64, input.LevelCount)
	maxLevel := input.LevelCount - 1
	fieldSpan := 2.0 * effectiveEc * input.TargetRangeFrac
	for level := 0; level < input.LevelCount; level++ {
		normalized := -1.0 + 2.0*float64(level)/float64(maxLevel)
		field := normalized * fieldSpan
		ascending[level] = field
		descending[level] = field
	}

	return LevelCalibrationResult{
		MaterialName:        input.Material.Name,
		LevelCount:          input.LevelCount,
		TargetRangeFrac:     input.TargetRangeFrac,
		TemperatureK:        input.TemperatureK,
		Method:              LevelCalibrationMethodLKQuasiStatic,
		AscendingFields:     ascending,
		DescendingFields:    descending,
		AscendingMinField:   ascending[0],
		AscendingMaxField:   ascending[len(ascending)-1],
		DescendingMinField:  descending[0],
		DescendingMaxField:  descending[len(descending)-1],
		AscendingMonotonic:  isMonotonicNonDecreasing(ascending),
		DescendingMonotonic: isMonotonicNonDecreasing(descending),
	}, nil
}

func levelCalibrationEffectiveEc(mat *HZOMaterial, temperatureK float64) float64 {
	if mat == nil {
		return 0
	}
	if mat.CurieTemp > 0 {
		return mat.CoerciveFieldAtTemp(temperatureK)
	}
	return mat.Ec + mat.TempCoeffEc*(temperatureK-300)
}

func levelCalibrationEffectivePolarization(mat *HZOMaterial, temperatureK float64) float64 {
	if mat == nil {
		return 0
	}
	if mat.CurieTemp > 0 {
		return mat.PolarizationAtTemp(temperatureK)
	}
	return mat.Pr + mat.TempCoeffPr*(temperatureK-300)
}

func isMonotonicNonDecreasing(values []float64) bool {
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			return false
		}
	}
	return true
}

func isFiniteLevelCalibrationValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
