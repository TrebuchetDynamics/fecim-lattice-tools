package arraysim

import "math"

// EnduranceAccuracyPoint captures degradation state at a given cycle count.
type EnduranceAccuracyPoint struct {
	Cycles           float64
	ConductanceDrift float64
	Accuracy         float64
}

// EnduranceAccuracyConfig defines fatigue-to-accuracy mapping.
type EnduranceAccuracyConfig struct {
	BaselineAccuracy float64
	EnduranceLimit   float64
	DriftAtLimit     float64
	Sensitivity      float64
}

// SimulateEnduranceAccuracy maps cycles -> conductance drift -> accuracy drop.
func SimulateEnduranceAccuracy(cycles []float64, cfg EnduranceAccuracyConfig) []EnduranceAccuracyPoint {
	if cfg.BaselineAccuracy <= 0 {
		cfg.BaselineAccuracy = 0.98
	}
	if cfg.EnduranceLimit <= 0 {
		cfg.EnduranceLimit = 1e9
	}
	if cfg.DriftAtLimit <= 0 {
		cfg.DriftAtLimit = 0.22
	}
	if cfg.Sensitivity <= 0 {
		cfg.Sensitivity = 0.55
	}

	out := make([]EnduranceAccuracyPoint, 0, len(cycles))
	for _, c := range cycles {
		if c < 0 {
			c = 0
		}
		ratio := c / cfg.EnduranceLimit
		if ratio > 2.0 {
			ratio = 2.0
		}
		drift := cfg.DriftAtLimit * (1.0 - math.Exp(-3.0*ratio))
		acc := cfg.BaselineAccuracy - cfg.Sensitivity*drift
		if acc < 0 {
			acc = 0
		}
		out = append(out, EnduranceAccuracyPoint{Cycles: c, ConductanceDrift: drift, Accuracy: acc})
	}
	return out
}
