package peripherals

import "math"

// VoltageRegulator models a basic LDO-style regulator used by peripheral rails.
type VoltageRegulator struct {
	NominalVoltage   float64 // Regulated output target (V)
	DropoutVoltage   float64 // Required Vin-Vout headroom (V)
	OutputResistance float64 // Small-signal output resistance (ohm)
	QuiescentCurrent float64 // Internal bias current (A)
	PSRRdB           float64 // Power supply rejection (dB)
}

// DefaultVoltageRegulator returns a basic 1.2V peripheral rail regulator.
func DefaultVoltageRegulator() *VoltageRegulator {
	return &VoltageRegulator{
		NominalVoltage:   1.2,
		DropoutVoltage:   0.15,
		OutputResistance: 0.5,
		QuiescentCurrent: 12e-6,
		PSRRdB:           45,
	}
}

// Regulate estimates output voltage under finite headroom and load current.
func (r *VoltageRegulator) Regulate(vin, loadCurrent float64) float64 {
	if r == nil {
		return 0
	}
	maxVout := vin - r.DropoutVoltage
	if maxVout < 0 {
		maxVout = 0
	}
	ideal := r.NominalVoltage
	if ideal > maxVout {
		ideal = maxVout
	}
	vout := ideal - loadCurrent*r.OutputResistance
	if vout < 0 {
		vout = 0
	}
	return vout
}

// SupplyNoiseTransfer returns output-referred ripple for a given supply ripple input.
func (r *VoltageRegulator) SupplyNoiseTransfer(vinRipple float64) float64 {
	if r == nil {
		return vinRipple
	}
	attn := math.Pow(10, -r.PSRRdB/20)
	return vinRipple * attn
}
