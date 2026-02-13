package physics

import "fmt"

// PulseSample is a current transient sample acquired during one pulse.
type PulseSample struct {
	TimeS    float64 // Seconds
	CurrentA float64 // Amperes
}

// PUNDResult captures integrated charge per pulse and separated switching charge.
type PUNDResult struct {
	QP_C float64 // Program pulse integrated charge
	QU_C float64 // Up pulse integrated charge (non-switching baseline after P)
	QN_C float64 // Negative pulse integrated charge
	QD_C float64 // Down pulse integrated charge (non-switching baseline after N)

	SwitchingPositive_C float64 // QP-QU
	SwitchingNegative_C float64 // QN-QD
}

// IntegrateCurrent integrates current over time with the trapezoidal rule.
func IntegrateCurrent(samples []PulseSample) (float64, error) {
	if len(samples) < 2 {
		return 0, fmt.Errorf("need at least 2 samples, got %d", len(samples))
	}
	q := 0.0
	for i := 1; i < len(samples); i++ {
		dt := samples[i].TimeS - samples[i-1].TimeS
		if dt <= 0 {
			return 0, fmt.Errorf("non-monotonic time at index %d", i)
		}
		q += 0.5 * (samples[i-1].CurrentA + samples[i].CurrentA) * dt
	}
	return q, nil
}

// AnalyzePUND calculates pulse charges and switching components from P/U/N/D traces.
func AnalyzePUND(programP, upU, negativeN, downD []PulseSample) (PUNDResult, error) {
	qP, err := IntegrateCurrent(programP)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate P pulse: %w", err)
	}
	qU, err := IntegrateCurrent(upU)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate U pulse: %w", err)
	}
	qN, err := IntegrateCurrent(negativeN)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate N pulse: %w", err)
	}
	qD, err := IntegrateCurrent(downD)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate D pulse: %w", err)
	}

	return PUNDResult{
		QP_C:                qP,
		QU_C:                qU,
		QN_C:                qN,
		QD_C:                qD,
		SwitchingPositive_C: qP - qU,
		SwitchingNegative_C: qN - qD,
	}, nil
}
