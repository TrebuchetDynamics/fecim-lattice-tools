package physics

import (
	"fmt"
	"math"
)

// WakeUpModelConfig captures wake-up then fatigue behavior of remanent polarization.
type WakeUpModelConfig struct {
	PrInitial_Cm2      float64
	WakeUpGainFraction float64 // Fractional gain above initial Pr at full wake-up
	WakeUpTauCycles    float64 // Characteristic wake-up cycle count
	FatigueOnsetCycles float64 // Fatigue starts after this cycle count
	FatigueTauCycles   float64 // Characteristic fatigue decay after onset
}

// WakeUpPolarization returns Pr at the given cycle count.
func WakeUpPolarization(cycles float64, cfg WakeUpModelConfig) (float64, error) {
	if cycles < 0 {
		return 0, fmt.Errorf("cycles must be non-negative, got %g", cycles)
	}
	if cfg.PrInitial_Cm2 <= 0 || cfg.WakeUpTauCycles <= 0 || cfg.FatigueTauCycles <= 0 {
		return 0, fmt.Errorf("invalid config: PrInitial=%g wakeTau=%g fatigueTau=%g", cfg.PrInitial_Cm2, cfg.WakeUpTauCycles, cfg.FatigueTauCycles)
	}
	wake := 1.0 + cfg.WakeUpGainFraction*(1.0-math.Exp(-cycles/cfg.WakeUpTauCycles))
	fatigue := 1.0
	if cycles > cfg.FatigueOnsetCycles {
		fatigue = math.Exp(-(cycles-cfg.FatigueOnsetCycles) / cfg.FatigueTauCycles)
	}
	return cfg.PrInitial_Cm2 * wake * fatigue, nil
}
