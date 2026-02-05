// Package arraysim provides approximate array coupling solvers for module4-circuits.
package arraysim

import "math"

// TIAConfig captures the minimal transimpedance amplifier model.
type TIAConfig struct {
	Rf   float64 // Feedback resistance (Ohm)
	Vref float64 // Output reference (V)
	Vmin float64 // Minimum output rail (V)
	Vmax float64 // Maximum output rail (V)
}

// ADCConfig captures the minimal ADC model for sense conversion.
type ADCConfig struct {
	Bits int     // Resolution in bits
	Vmin float64 // Minimum input voltage (V)
	Vmax float64 // Maximum input voltage (V)
}

// SenseChain chains a TIA and ADC model for current sensing.
type SenseChain struct {
	TIA TIAConfig
	ADC ADCConfig
}

// SenseResult captures the sense conversion outputs with saturation flags.
type SenseResult struct {
	Vout         float64
	Code         int
	TIASaturated bool
	ADCSaturated bool
}

// ConvertCurrent converts a single row current to TIA output and ADC code.
func (s SenseChain) ConvertCurrent(currentA float64) SenseResult {
	vout := s.TIA.Vref + currentA*s.TIA.Rf
	tiaSat := false
	if vout < s.TIA.Vmin {
		vout = s.TIA.Vmin
		tiaSat = true
	}
	if vout > s.TIA.Vmax {
		vout = s.TIA.Vmax
		tiaSat = true
	}

	adcBits := s.ADC.Bits
	if adcBits < 1 {
		adcBits = 1
	}
	levels := 1 << adcBits
	adcSat := false
	adcV := vout
	if adcV < s.ADC.Vmin {
		adcV = s.ADC.Vmin
		adcSat = true
	}
	if adcV > s.ADC.Vmax {
		adcV = s.ADC.Vmax
		adcSat = true
	}
	if tiaSat {
		adcSat = true
	}

	fraction := 0.0
	span := s.ADC.Vmax - s.ADC.Vmin
	if span > 0 {
		fraction = (adcV - s.ADC.Vmin) / span
	}
	code := int(math.Round(fraction * float64(levels-1)))
	if code < 0 {
		code = 0
	}
	if code >= levels {
		code = levels - 1
	}

	return SenseResult{
		Vout:         vout,
		Code:         code,
		TIASaturated: tiaSat,
		ADCSaturated: adcSat,
	}
}

// ConvertCurrents converts a slice of row currents.
func (s SenseChain) ConvertCurrents(currents []float64) []SenseResult {
	results := make([]SenseResult, len(currents))
	for i, current := range currents {
		results[i] = s.ConvertCurrent(current)
	}
	return results
}
