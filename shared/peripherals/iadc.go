// Package peripherals provides shared peripheral circuit models for FeCIM simulations.
package peripherals

// ADCModel is the interface all ADC implementations must satisfy.
//
// Inspired by CrossSim's IADC hierarchy (SAR, ramp, pipeline, cyclic, quantizer).
// Current implementation: *ADC (flash ADC, optionally with SAR noise model).
// Future implementations (see Phase 9 of shared-refactor-plan.md):
//   - SarADC      — successive-approximation; best accuracy/power tradeoff
//   - PipelineADC — high throughput for column-parallel readout
//   - RampADC     — simplest, lowest area, worst speed
//   - CyclicADC   — recycling architecture for medium resolution
type ADCModel interface {
	// Convert quantizes an analog voltage to a digital code in [0, Levels()-1].
	Convert(voltage float64) int

	// Resolution returns the voltage step size (LSB) in volts.
	Resolution() float64

	// Levels returns the total number of output codes (= 2^bits).
	Levels() int

	// EnergyPerConversion returns the estimated energy consumed per conversion (J).
	EnergyPerConversion() float64

	// AreaEstimate returns the estimated silicon area (µm²).
	AreaEstimate() float64

	// LatencyNS returns the conversion latency in nanoseconds.
	LatencyNS() float64

	// ENOB returns the effective number of bits including noise and nonlinearity.
	ENOB() float64

	// TypeString returns a human-readable architecture name (e.g. "Flash", "SAR").
	TypeString() string
}

// Compile-time check: *ADC must satisfy ADCModel.
var _ ADCModel = (*ADC)(nil)
