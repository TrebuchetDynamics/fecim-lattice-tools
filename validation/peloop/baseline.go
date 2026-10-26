package peloop

import (
	"fmt"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

// BaselineSettings records the sampling protocol and unit conversions.
// Temperature/stress are constructor defaults, not experimental conditions;
// conditioning prepares numerical history, not measured wake-up dynamics.
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

// BaselineResult retains simulated physical samples, the full unchanged material
// snapshot (in native field units), and settings. It is neither a fit nor a claim
// that the model reproduces the reference measurement frequency or wake-up.
type BaselineResult struct {
	Samples  []Sample
	Material physics.HZOMaterial
	Settings BaselineSettings
}

// SimulateBaseline drives a fresh DefaultHZO/NewPreisachModel using only fields
// and input order. After Reset, it applies the minimum field and two min->max->min
// conditioning cycles. Each half-cycle uses 512 equal steps (excluding its start,
// including its endpoint). It then samples Update at each reference field without
// resetting between branches. Measured P is validated but never drives the model.
func SimulateBaseline(reference []Sample) (BaselineResult, error) {
	if err := validatePhysicalSamples(reference); err != nil {
		return BaselineResult{}, err
	}
	minField, maxField := reference[0].FieldMVcm, reference[0].FieldMVcm
	for i, sample := range reference {
		if !finite(sample.FieldMVcm * 1e8) {
			return BaselineResult{}, fmt.Errorf("sample %d: V/m conversion overflows", i)
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
		return BaselineResult{}, fmt.Errorf("baseline: conditioning span overflows")
	}
	material := physics.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)
	result := BaselineResult{
		Samples: make([]Sample, len(reference)), Material: *material,
		Settings: BaselineSettings{
			Preset: "shared/physics.DefaultHZO", Model: "module1-hysteresis/pkg/ferroelectric.NewPreisachModel",
			TemperatureK: model.Temperature, StressGPa: model.Stress, Reset: true,
			ConditioningCycles: 2, StepsPerHalfCycle: 512, FieldMinMVcm: minField, FieldMaxMVcm: maxField,
			ConditioningPath: "min -> max -> min", SamplingOrder: "input order",
			UpdateMode: "quasi-static Update; not frequency- or wake-up-matched",
			FieldToVm:  1e8, PolarizationToUCcm2: 1e2,
		},
	}
	model.Reset()
	model.Update(lo)
	for cycle := 0; cycle < result.Settings.ConditioningCycles; cycle++ {
		for step := 1; step <= result.Settings.StepsPerHalfCycle; step++ {
			model.Update(lo + (hi-lo)*(float64(step)/float64(result.Settings.StepsPerHalfCycle)))
		}
		for step := 1; step <= result.Settings.StepsPerHalfCycle; step++ {
			model.Update(hi - (hi-lo)*(float64(step)/float64(result.Settings.StepsPerHalfCycle)))
		}
	}
	for i, sample := range reference {
		sample.PolarizationUCcm2 = model.Update(sample.FieldMVcm*result.Settings.FieldToVm) * result.Settings.PolarizationToUCcm2
		if !finite(sample.PolarizationUCcm2) {
			return BaselineResult{}, fmt.Errorf("sample %d: nonfinite baseline polarization", i)
		}
		result.Samples[i] = sample
	}
	return result, nil
}
