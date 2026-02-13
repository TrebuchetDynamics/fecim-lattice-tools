package physics

import (
	"fmt"
	"math"
)

// HysteresisMetrics summarizes key loop observables.
type HysteresisMetrics struct {
	FrequencyHz  float64
	Pr_Cm2       float64
	Ec_Vm        float64
	LoopArea_Jm3 float64
}

// FrequencyDispersionConfig controls log-frequency scaling around a reference frequency.
type FrequencyDispersionConfig struct {
	ReferenceHz        float64
	EcLogSlope         float64 // Ec multiplier slope vs ln(f/f0)
	PrLogSlope         float64 // Pr multiplier slope vs ln(f/f0)
	LoopAreaLogSlope   float64 // Area multiplier slope vs ln(f/f0)
	MinMultiplierClamp float64 // Lower clamp to keep metrics physical
}

// ApplyFrequencyDispersion maps baseline loop metrics to a target frequency.
func ApplyFrequencyDispersion(base HysteresisMetrics, targetHz float64, cfg FrequencyDispersionConfig) (HysteresisMetrics, error) {
	if targetHz <= 0 || cfg.ReferenceHz <= 0 {
		return HysteresisMetrics{}, fmt.Errorf("target and reference frequencies must be positive")
	}
	if cfg.MinMultiplierClamp <= 0 {
		return HysteresisMetrics{}, fmt.Errorf("MinMultiplierClamp must be positive")
	}
	logRatio := math.Log(targetHz / cfg.ReferenceHz)
	clamp := func(v float64) float64 {
		if v < cfg.MinMultiplierClamp {
			return cfg.MinMultiplierClamp
		}
		return v
	}

	ecMult := clamp(1 + cfg.EcLogSlope*logRatio)
	prMult := clamp(1 + cfg.PrLogSlope*logRatio)
	areaMult := clamp(1 + cfg.LoopAreaLogSlope*logRatio)

	return HysteresisMetrics{
		FrequencyHz:  targetHz,
		Pr_Cm2:       base.Pr_Cm2 * prMult,
		Ec_Vm:        base.Ec_Vm * ecMult,
		LoopArea_Jm3: base.LoopArea_Jm3 * areaMult,
	}, nil
}
