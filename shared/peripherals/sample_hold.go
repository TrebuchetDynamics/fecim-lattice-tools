package peripherals

import "math"

// SampleAndHold models a basic switched-capacitor sample-and-hold front-end.
type SampleAndHold struct {
	HoldCapacitance   float64 // Hold capacitor (F)
	SwitchResistance  float64 // Sampling switch on-resistance (ohm)
	LeakageResistance float64 // Effective leakage resistance during hold (ohm)
	AcquisitionTimeNS float64 // Typical acquisition time (ns)
}

// DefaultSampleAndHold returns a conservative read-path S/H configuration.
func DefaultSampleAndHold() *SampleAndHold {
	return &SampleAndHold{
		HoldCapacitance:   1e-12, // 1 pF
		SwitchResistance:  1e3,   // 1 kΩ
		LeakageResistance: 5e9,   // 5 GΩ
		AcquisitionTimeNS: 20.0,  // 20 ns
	}
}

// SettledFraction returns the acquisition settling fraction after t seconds.
func (s *SampleAndHold) SettledFraction(tSeconds float64) float64 {
	if s == nil || s.HoldCapacitance <= 0 || s.SwitchResistance <= 0 || tSeconds <= 0 {
		return 0
	}
	tau := s.SwitchResistance * s.HoldCapacitance
	return 1 - math.Exp(-tSeconds/tau)
}

// HoldDroop returns held-voltage decay ratio V(t)/V0 for hold duration t seconds.
func (s *SampleAndHold) HoldDroop(tSeconds float64) float64 {
	if s == nil || s.HoldCapacitance <= 0 || s.LeakageResistance <= 0 || tSeconds <= 0 {
		return 1
	}
	tau := s.LeakageResistance * s.HoldCapacitance
	return math.Exp(-tSeconds / tau)
}
