// Package peripherals provides shared peripheral circuit configurations for FeCIM systems.
package peripherals

import "fecim-lattice-tools/shared/physics"

// Standard peripheral configuration constants for the demo 30-level baseline.
const (
	// DefaultBits is the standard resolution for ADC/DAC (5 bits = 32 levels, we use 30).
	DefaultBits = 5

	// DefaultLevels is the number of discrete levels (2^5 = 32, demo baseline uses 30).
	DefaultLevels = 32

	// FeCIMLevels is the baseline number of analog states used in this demo.
	// Re-exported from shared/physics for backward compatibility.
	FeCIMLevels = physics.DefaultLevels
)

// DAC reference voltage constants
const (
	// DACVrefHigh is the high reference voltage for write operations (+1.5V).
	// This default is a simulation baseline chosen to bracket common FeFET write
	// windows (roughly ±(1.5-3)x Ec in practical programming flows).
	DACVrefHigh = 1.5

	// DACVrefLow is the low reference voltage for write operations (-1.5V).
	// Symmetric with DACVrefHigh for bipolar set/reset experiments.
	DACVrefLow = -1.5

	// DACSettleTime is the typical DAC settling time in nanoseconds.
	DACSettleTime = 10.0
)

// ADC reference voltage constants
const (
	// ADCVrefHigh is the high reference voltage for read operations (1.0V).
	// Chosen to match the default TIA output clamp (0-1V) for direct sensing
	// without extra level shifting.
	ADCVrefHigh = 1.0

	// ADCVrefLow is the low reference voltage for read operations (0.0V).
	ADCVrefLow = 0.0

	// ADCConversionTime is the typical SAR ADC conversion time in nanoseconds.
	ADCConversionTime = 50.0
)

// Nonlinearity specifications (in LSB)
const (
	// DefaultINL is the typical integral nonlinearity.
	// Placeholder value for medium-resolution SAR-style converters; use silicon
	// characterization for tapeout decisions.
	DefaultINL = 0.5

	// DefaultDNL is the typical differential nonlinearity.
	// Placeholder value paired with DefaultINL for educational what-if studies.
	DefaultDNL = 0.25
)

// NOTE: Full DAC, ADC, TIA, ChargePump structs and their methods are now in:
//   - dac.go (DAC struct with DefaultDAC())
//   - adc.go (ADC struct with DefaultADC(), ADCType enum)
//   - tia.go (TIA struct with DefaultTIA())
//   - chargepump.go (ChargePump struct with DefaultChargePump())
//   - analysis.go (INL/DNL analysis, timing, power breakdown)
//
// The simplified Config structs that were here have been removed to avoid
// duplication. Use the full structs from the individual files instead.
