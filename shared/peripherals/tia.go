package peripherals

import (
	"math"
)

// TIA represents a Transimpedance Amplifier for current-to-voltage conversion.
// Used in crossbar read path to sense column currents.
type TIA struct {
	Gain             float64 // Transimpedance gain (Ohms)
	Bandwidth        float64 // -3dB bandwidth (Hz)
	InputNoiseRMS    float64 // Input-referred noise (A/sqrt(Hz))
	OutputOffset     float64 // Output offset voltage (V)
	MaxInputCurrent  float64 // Maximum input current (A)
	MaxOutputVoltage float64 // Maximum output voltage (V)
}

// DefaultTIA returns a TIA configured for crossbar sense operations.
//
// Defaults are heuristic and chosen so that 0-100 µA array current maps into
// the default ADC 0-1V range via Vout ~= Iin * 10kΩ.
func DefaultTIA() *TIA {
	tia := &TIA{
		Gain:             10e3,   // 10 kΩ transimpedance (heuristic baseline)
		Bandwidth:        100e6,  // 100 MHz bandwidth (heuristic baseline)
		InputNoiseRMS:    1e-12,  // 1 pA/sqrt(Hz) placeholder input-referred noise
		OutputOffset:     5e-3,   // 5 mV output offset
		MaxInputCurrent:  100e-6, // 100 µA max input
		MaxOutputVoltage: 1.0,    // 1V max output (aligned with ADCVrefHigh)
	}
	log.Calculation("DefaultTIA", map[string]interface{}{
		"gain":               tia.Gain,
		"bandwidth":          tia.Bandwidth,
		"input_noise_rms":    tia.InputNoiseRMS,
		"max_input_current":  tia.MaxInputCurrent,
		"max_output_voltage": tia.MaxOutputVoltage,
	}, tia)
	return tia
}

// Convert performs current-to-voltage conversion.
func (t *TIA) Convert(current float64) float64 {
	log.Input("TIA.Convert", map[string]interface{}{
		"current": current,
		"gain":    t.Gain,
		"offset":  t.OutputOffset,
	})

	// Vout = Iin * Gain + Offset
	output := current*t.Gain + t.OutputOffset

	// Clamp to output range
	if output < 0 {
		output = 0
	}
	if output > t.MaxOutputVoltage {
		output = t.MaxOutputVoltage
	}

	log.Calculation("TIA.Convert", map[string]interface{}{
		"current": current,
	}, output)

	return output
}

// ConvertWithNoise adds thermal noise to conversion.
func (t *TIA) ConvertWithNoise(current float64) float64 {
	idealOutput := t.Convert(current)

	// Calculate RMS noise voltage
	// Vnoise = Inoise * Gain * sqrt(BW)
	noiseVoltage := t.InputNoiseRMS * t.Gain * math.Sqrt(t.Bandwidth)

	// Deterministic RMS noise injection (non-random, for repeatable demos)
	noiseContribution := noiseVoltage

	noisy := idealOutput + noiseContribution
	if noisy < 0 {
		noisy = 0
	}
	if noisy > t.MaxOutputVoltage {
		noisy = t.MaxOutputVoltage
	}

	return noisy
}

// SNR returns the signal-to-noise ratio for a given input current.
func (t *TIA) SNR(current float64) float64 {
	log.Input("TIA.SNR", map[string]interface{}{
		"current":   current,
		"gain":      t.Gain,
		"bandwidth": t.Bandwidth,
	})

	signal := current * t.Gain
	noise := t.InputNoiseRMS * t.Gain * math.Sqrt(t.Bandwidth)

	if noise == 0 {
		return math.Inf(1)
	}

	snr := 20 * math.Log10(signal/noise) // dB

	log.Calculation("TIA.SNR", map[string]interface{}{
		"current": current,
		"signal":  signal,
		"noise":   noise,
	}, snr)

	return snr
}

// MinDetectableCurrent returns minimum detectable current (SNR=1).
func (t *TIA) MinDetectableCurrent() float64 {
	// I_min = Inoise * sqrt(BW)
	return t.InputNoiseRMS * math.Sqrt(t.Bandwidth)
}

// DynamicRange returns the dynamic range in dB.
func (t *TIA) DynamicRange() float64 {
	log.Input("TIA.DynamicRange", map[string]interface{}{
		"max_input_current": t.MaxInputCurrent,
	})

	minCurrent := t.MinDetectableCurrent()
	dr := 20 * math.Log10(t.MaxInputCurrent/minCurrent)

	log.Calculation("TIA.DynamicRange", map[string]interface{}{
		"min_current": minCurrent,
		"max_current": t.MaxInputCurrent,
	}, dr)

	return dr
}

// SettlingTime estimates the step response settling time.
func (t *TIA) SettlingTime() float64 {
	log.Input("TIA.SettlingTime", map[string]interface{}{
		"bandwidth": t.Bandwidth,
	})

	// Single-pole settling: t = ln(1/accuracy) / (2*pi*BW)
	// For 0.1% accuracy: t ≈ 7 / (2*pi*BW)
	accuracy := 0.001 // 0.1%
	settleTime := math.Log(1/accuracy) / (2 * math.Pi * t.Bandwidth)

	log.Calculation("TIA.SettlingTime", map[string]interface{}{
		"bandwidth": t.Bandwidth,
		"accuracy":  accuracy,
	}, settleTime)

	return settleTime
}

// PowerConsumption estimates TIA power based on bandwidth and gain.
func (t *TIA) PowerConsumption() float64 {
	// Typical: P ≈ 2 * kT * BW * Gain / η
	// Simplified estimate
	kT := 4.14e-21 // kT at 300K
	efficiency := 0.1
	return 2 * kT * t.Bandwidth * t.Gain / efficiency
}
