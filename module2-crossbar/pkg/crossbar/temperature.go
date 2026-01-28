// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import "math"

// TemperatureEffects models temperature-dependent physics effects.
// FeFET devices show enhanced properties at cryogenic temperatures
// and degraded performance at elevated temperatures.
type TemperatureEffects struct {
	AmbientK float64 // Operating temperature in Kelvin
}

// NewTemperatureEffects creates a temperature effects model.
func NewTemperatureEffects(tempK float64) *TemperatureEffects {
	if tempK < 0 {
		tempK = 300 // Default to room temperature
	}
	return &TemperatureEffects{AmbientK: tempK}
}

// TemperaturePresets provides common operating temperatures.
const (
	TempCryogenic  = 77.0  // Liquid nitrogen temperature
	TempColdSpace  = 4.0   // Deep space / liquid helium
	TempRoom       = 300.0 // Room temperature (27°C)
	TempAutomotive = 400.0 // Automotive Grade 0 (125°C)
	TempIndustrial = 358.0 // Industrial grade (85°C)
)

// AdjustedWireResistance applies temperature coefficient of resistance (TCR) to wire resistance.
// Uses copper TCR = 0.00393 /K (3.93% per Kelvin change from 300K reference).
// Higher temperature → higher resistance → more IR drop.
func (t *TemperatureEffects) AdjustedWireResistance(R0 float64) float64 {
	const copperTCR = 0.00393 // Copper temperature coefficient
	return R0 * (1.0 + copperTCR*(t.AmbientK-300.0))
}

// AdjustedConductanceRange scales Gmin/Gmax conductance window with temperature.
// Physics basis:
//   - Cryogenic (<100K): Enhanced ferroelectric polarization → wider window
//   - High temp (>300K): Thermal noise reduces effective window
//
// Returns (adjustedGmin, adjustedGmax).
func (t *TemperatureEffects) AdjustedConductanceRange(gMin, gMax float64) (float64, float64) {
	if t.AmbientK < 100 {
		// Cryogenic enhancement
		// At 4K, Pr can reach 75 µC/cm² vs 15-34 µC/cm² at RT (Adv. Elec. Mat. 2024)
		// Model as window expansion factor
		enhancementFactor := 1.0 + 0.5*(100-t.AmbientK)/100
		// Gmin decreases (better OFF state), Gmax increases (better ON state)
		return gMin / enhancementFactor, gMax * enhancementFactor
	}

	if t.AmbientK > 300 {
		// High temperature degradation
		// Thermal fluctuations reduce effective polarization
		// Model as window narrowing (conservative estimate)
		degradationFactor := 1.0 - 0.1*(t.AmbientK-300)/100
		if degradationFactor < 0.5 {
			degradationFactor = 0.5 // Cap at 50% degradation
		}
		// Window narrows symmetrically toward center
		gMid := (gMin + gMax) / 2
		gRange := (gMax - gMin) * degradationFactor
		return gMid - gRange/2, gMid + gRange/2
	}

	// Room temperature - no adjustment
	return gMin, gMax
}

// AdjustedDriftRate scales drift rate with temperature using Arrhenius model.
// Drift processes are thermally activated: rate ∝ exp(-Ea/kT)
// Higher temperature → faster drift.
//
// Reference: Thermal activation energy Ea ≈ 0.5 eV for typical ferroelectric switching.
func (t *TemperatureEffects) AdjustedDriftRate(driftCoeff float64) float64 {
	const kB = 1.38e-23 // Boltzmann constant (J/K)
	const Ea = 0.5      // Activation energy (eV)
	const eV = 1.6e-19  // Electron-volt to Joules

	// Reference rate at 300K
	refRate := math.Exp(-Ea * eV / (kB * 300))
	// Rate at operating temperature
	newRate := math.Exp(-Ea * eV / (kB * t.AmbientK))

	// Scale drift coefficient by rate ratio
	return driftCoeff * (newRate / refRate)
}

// AdjustedRetention estimates retention time scaling with temperature.
// Returns a factor to multiply the nominal retention time.
// Lower temperature → exponentially better retention.
func (t *TemperatureEffects) AdjustedRetention() float64 {
	const kB = 1.38e-23
	const Ea = 0.5
	const eV = 1.6e-19

	// Higher activation energy ratio → longer retention
	refRate := math.Exp(-Ea * eV / (kB * 300))
	newRate := math.Exp(-Ea * eV / (kB * t.AmbientK))

	// Retention scales inversely with rate
	return refRate / newRate
}

// AdjustedNoise scales thermal noise with temperature.
// Thermal noise voltage Vn ∝ sqrt(kT), so noise power ∝ T.
// Returns a noise multiplier relative to 300K.
func (t *TemperatureEffects) AdjustedNoise() float64 {
	// RMS voltage noise scales as sqrt(T)
	return math.Sqrt(t.AmbientK / 300.0)
}

// AdjustedSwitchingEnergy estimates switching energy scaling with temperature.
// Ferroelectric switching energy scales roughly linearly with coercive field,
// which can vary with temperature.
func (t *TemperatureEffects) AdjustedSwitchingEnergy(baseEnergy float64) float64 {
	// At cryogenic temperatures, coercive field can be slightly lower
	// At high temperatures, enhanced ionic motion can reduce Ec
	// This is a simplified model - actual behavior is material-dependent

	if t.AmbientK < 100 {
		// Slight reduction at cryo
		return baseEnergy * (0.9 + 0.1*t.AmbientK/100)
	}
	if t.AmbientK > 300 {
		// Slight reduction at high temp (but reliability concerns)
		factor := 1.0 - 0.05*(t.AmbientK-300)/100
		if factor < 0.8 {
			factor = 0.8
		}
		return baseEnergy * factor
	}
	return baseEnergy
}

// GetTemperatureLabel returns a human-readable label for the temperature.
func (t *TemperatureEffects) GetTemperatureLabel() string {
	switch {
	case t.AmbientK < 10:
		return "Deep Cryogenic"
	case t.AmbientK < 100:
		return "Cryogenic"
	case t.AmbientK < 273:
		return "Cold"
	case t.AmbientK < 323:
		return "Room Temperature"
	case t.AmbientK < 373:
		return "Industrial"
	case t.AmbientK < 423:
		return "Automotive"
	default:
		return "Extreme Heat"
	}
}

// TemperatureEffectsForMVM returns temperature effects configured for MVM simulation.
// Adjusts wire parameters and provides conductance scaling.
type TemperatureAdjustedParams struct {
	WireResistanceFactor float64 // Multiply nominal wire R by this
	GminAdjusted         float64 // Adjusted minimum conductance
	GmaxAdjusted         float64 // Adjusted maximum conductance
	DriftRateFactor      float64 // Multiply nominal drift rate by this
	NoiseFactor          float64 // Multiply nominal noise by this
	RetentionFactor      float64 // Multiply nominal retention by this
}

// GetAdjustedParams returns all temperature-adjusted parameters.
func (t *TemperatureEffects) GetAdjustedParams() *TemperatureAdjustedParams {
	adjGmin, adjGmax := t.AdjustedConductanceRange(GMin, GMax)

	return &TemperatureAdjustedParams{
		WireResistanceFactor: t.AdjustedWireResistance(1.0), // Factor for 1Ω base
		GminAdjusted:         adjGmin,
		GmaxAdjusted:         adjGmax,
		DriftRateFactor:      t.AdjustedDriftRate(1.0),      // Factor for base rate
		NoiseFactor:          t.AdjustedNoise(),
		RetentionFactor:      t.AdjustedRetention(),
	}
}
