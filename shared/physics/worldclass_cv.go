package physics

import "fmt"

// PEPoint is one sample from a hysteresis sweep.
type PEPoint struct {
	Field_Vm        float64 // Electric field (V/m)
	Polarization_Cm float64 // Polarization (C/m^2)
}

// CVPoint contains bias voltage and extracted capacitance dQ/dV.
type CVPoint struct {
	Voltage_V     float64
	Capacitance_F float64
}

// ExtractButterflyCV computes C(V)=dQ/dV=(A/t)*dP/dE from P-E samples.
func ExtractButterflyCV(pe []PEPoint, areaM2, thicknessM float64) ([]CVPoint, error) {
	if len(pe) < 3 {
		return nil, fmt.Errorf("need at least 3 P-E points, got %d", len(pe))
	}
	if areaM2 <= 0 || thicknessM <= 0 {
		return nil, fmt.Errorf("area and thickness must be positive")
	}

	scale := areaM2 / thicknessM
	out := make([]CVPoint, 0, len(pe)-2)
	for i := 1; i < len(pe)-1; i++ {
		dE := pe[i+1].Field_Vm - pe[i-1].Field_Vm
		if dE == 0 {
			return nil, fmt.Errorf("zero field spacing around index %d", i)
		}
		dPdE := (pe[i+1].Polarization_Cm - pe[i-1].Polarization_Cm) / dE
		voltage := pe[i].Field_Vm * thicknessM
		out = append(out, CVPoint{
			Voltage_V:     voltage,
			Capacitance_F: scale * dPdE,
		})
	}
	return out, nil
}
