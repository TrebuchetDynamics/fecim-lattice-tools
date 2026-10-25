package peloop

import (
	"fmt"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

// BaselineSettings records the complete sampling protocol. Temperature and stress
// are constructor defaults, not inferred experimental conditions. Conditioning is
// numerical history preparation, not a simulation of measured wake-up cycles.
type BaselineSettings struct {
	Preset              string
	Model               string
	TemperatureK        float64
	StressGPa           float64
	Reset               bool
	ConditioningCycles  int
	StepsPerHalfCycle   int
	FieldMinMVcm        float64
	FieldMaxMVcm        float64
	ConditioningPath    string
	SamplingOrder       string
	UpdateMode          string
	FieldToVm           float64
	PolarizationToUCcm2 float64
}

// BaselineResult retains all preset fields (in their native SI units), the fixed
// simulation settings and simulated physical samples. It is not a fitted model
// or a claim to reproduce measurement frequency or wake-up dynamics.
type BaselineResult struct {
	Samples  []Sample
	Material physics.HZOMaterial
	Settings BaselineSettings
}

// SimulateBaseline uses only reference fields and their input order to drive a
// fresh, unchanged DefaultHZO Preisach model. After Reset it applies the minimum
// field, then two min->max->min conditioning cycles with 512 equal intervals per
// half-cycle (excluding each start, including each endpoint). It then calls Update
// once per reference field in input order, without resetting between branches.
// Measured polarization never enters the model or its settings.
func SimulateBaseline(reference []Sample) (BaselineResult, error) {
	if err := validateSamples(reference); err != nil {
		return BaselineResult{}, err
	}
	minField, maxField := reference[0].FieldMVcm, reference[0].FieldMVcm
	for i, sample := range reference {
		if !finite(sample.FieldMVcm * 1e8) {
			return BaselineResult{}, fmt.Errorf("sample %d: field conversion overflows V/m", i)
		}
		if sample.FieldMVcm < minField {
			minField = sample.FieldMVcm
		}
		if sample.FieldMVcm > maxField {
			maxField = sample.FieldMVcm
		}
	}
	lo, hi := minField*1e8, maxField*1e8
	if !finite(hi - lo) {
		return BaselineResult{}, fmt.Errorf("conditioning field span overflows V/m")
	}
	material := physics.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)
	result := BaselineResult{
		Material: *material,
		Settings: BaselineSettings{
			Preset: "shared/physics.DefaultHZO", Model: "module1-hysteresis/pkg/ferroelectric.NewPreisachModel",
			TemperatureK: model.Temperature, StressGPa: model.Stress, Reset: true,
			ConditioningCycles: 2, StepsPerHalfCycle: 512, FieldMinMVcm: minField, FieldMaxMVcm: maxField,
			ConditioningPath: "min -> max -> min", SamplingOrder: "input order",
			UpdateMode: "quasi-static Update; not frequency- or wake-up-matched",
			FieldToVm:  1e8, PolarizationToUCcm2: 1e2,
		},
		Samples: make([]Sample, len(reference)),
	}
	model.Reset()
	model.Update(lo)
	for cycle := 0; cycle < result.Settings.ConditioningCycles; cycle++ {
		for step := 1; step <= result.Settings.StepsPerHalfCycle; step++ {
			model.Update(lo + (hi-lo)*(float64(step)/float64(result.Settings.StepsPerHalfCycle)))
		}
		for step := 1; step <= result.Settings.StepsPerHalfCycle; step++ {
			model.Update(hi + (lo-hi)*(float64(step)/float64(result.Settings.StepsPerHalfCycle)))
		}
	}
	for i, sample := range reference {
		sample.PolarizationUCcm2 = model.Update(sample.FieldMVcm*result.Settings.FieldToVm) * result.Settings.PolarizationToUCcm2
		if !finite(sample.PolarizationUCcm2) {
			return BaselineResult{}, fmt.Errorf("sample %d: nonfinite model polarization", i)
		}
		result.Samples[i] = sample
	}
	return result, nil
}
