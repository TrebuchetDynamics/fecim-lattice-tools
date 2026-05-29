// Package levels contains conductance-level quantization helpers for crossbar arrays.
package levels

import "fecim-lattice-tools/shared/physics"

// DefaultQuantizationLevels is the standard number of discrete analog states.
// The 30-level default is a simulation baseline, not a validated device claim.
const DefaultQuantizationLevels = physics.DefaultLevels

// QuantizeToDefaultLevels quantizes a normalized conductance value to the
// default discrete level grid and clamps out-of-range inputs.
func QuantizeToDefaultLevels(value float64) float64 {
	return physics.QuantizeTo30Levels(value)
}

// DefaultLevelFor returns the default discrete level index for a normalized
// conductance value. It preserves the legacy shared/physics level calculation;
// callers should quantize or clamp inputs first when they need bounded levels.
func DefaultLevelFor(conductance float64) int {
	return physics.GetLevelFor30(conductance)
}
