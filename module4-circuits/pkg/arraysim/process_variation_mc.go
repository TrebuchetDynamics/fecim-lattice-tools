package arraysim

import (
	"math"
	"math/rand"
)

// ProcessVariationConfig defines Monte Carlo process variation settings.
type ProcessVariationConfig struct {
	NominalEc          float64
	NominalPr          float64
	VariationFraction  float64
	Samples            int
	Seed               int64
	MinReadMarginRatio float64
}

// ProcessVariationResult summarizes Monte Carlo variation and yield.
type ProcessVariationResult struct {
	MeanEc      float64
	MeanPr      float64
	StdEc       float64
	StdPr       float64
	Yield       float64
	PassSamples int
}

// RunProcessVariationMC sweeps Ec/Pr using Gaussian ±variation and reports yield.
func RunProcessVariationMC(cfg ProcessVariationConfig) ProcessVariationResult {
	if cfg.Samples < 1 {
		cfg.Samples = 1
	}
	if cfg.VariationFraction < 0 {
		cfg.VariationFraction = 0
	}
	if cfg.MinReadMarginRatio <= 0 {
		cfg.MinReadMarginRatio = 0.85
	}

	rng := rand.New(rand.NewSource(cfg.Seed))
	sigmaEc := cfg.NominalEc * cfg.VariationFraction / 3.0
	sigmaPr := cfg.NominalPr * cfg.VariationFraction / 3.0

	var sumEc, sumPr, sumEc2, sumPr2 float64
	pass := 0
	for i := 0; i < cfg.Samples; i++ {
		ec := clampPositive(cfg.NominalEc + rng.NormFloat64()*sigmaEc)
		pr := clampPositive(cfg.NominalPr + rng.NormFloat64()*sigmaPr)

		sumEc += ec
		sumPr += pr
		sumEc2 += ec * ec
		sumPr2 += pr * pr

		margin := (pr / cfg.NominalPr) / (ec / cfg.NominalEc)
		if margin >= cfg.MinReadMarginRatio {
			pass++
		}
	}

	n := float64(cfg.Samples)
	meanEc := sumEc / n
	meanPr := sumPr / n
	stdEc := math.Sqrt(math.Max(0, sumEc2/n-meanEc*meanEc))
	stdPr := math.Sqrt(math.Max(0, sumPr2/n-meanPr*meanPr))

	return ProcessVariationResult{
		MeanEc:      meanEc,
		MeanPr:      meanPr,
		StdEc:       stdEc,
		StdPr:       stdPr,
		Yield:       float64(pass) / n,
		PassSamples: pass,
	}
}

func clampPositive(v float64) float64 {
	if v < 1e-12 {
		return 1e-12
	}
	return v
}
