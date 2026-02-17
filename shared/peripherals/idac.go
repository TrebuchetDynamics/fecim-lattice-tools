// Package peripherals provides shared peripheral circuit models for FeCIM simulations.
package peripherals

// DACModel is the interface all DAC implementations must satisfy.
//
// Inspired by CrossSim's IDAC abstraction.
// Current implementation: *DAC (R-2R ladder / binary-weighted).
// Future implementations may include current-steering, sigma-delta, and
// charge-redistribution DACs for column-driver applications.
type DACModel interface {
	// Convert maps a digital code in [0, Levels()-1] to an analog voltage (V).
	Convert(level int) float64

	// Resolution returns the voltage step size (LSB) in volts.
	Resolution() float64

	// Levels returns the total number of input codes (= 2^bits).
	Levels() int

	// VoltageRange returns the minimum and maximum output voltages (V).
	VoltageRange() (min, max float64)

	// EnergyPerConversion returns the estimated energy consumed per conversion (J).
	EnergyPerConversion() float64
}

// Compile-time check: *DAC must satisfy DACModel.
var _ DACModel = (*DAC)(nil)
