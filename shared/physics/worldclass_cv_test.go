package physics

import (
	"math"
	"testing"
)

func TestExtractButterflyCV_LinearPEGivesConstantC(t *testing.T) {
	// P = kE, so dP/dE = k and C = (A/t)k
	k := 8e-10
	area := 1e-12
	thickness := 10e-9
	pe := []PEPoint{
		{Field_Vm: -2e8, Polarization_Cm: -2e8 * k},
		{Field_Vm: -1e8, Polarization_Cm: -1e8 * k},
		{Field_Vm: 0, Polarization_Cm: 0},
		{Field_Vm: 1e8, Polarization_Cm: 1e8 * k},
		{Field_Vm: 2e8, Polarization_Cm: 2e8 * k},
	}
	cv, err := ExtractButterflyCV(pe, area, thickness)
	if err != nil {
		t.Fatalf("ExtractButterflyCV error: %v", err)
	}
	want := (area / thickness) * k
	for i, pt := range cv {
		if math.Abs(pt.Capacitance_F-want) > 1e-18 {
			t.Fatalf("cv[%d].C = %g, want %g", i, pt.Capacitance_F, want)
		}
	}
}
