package constants

import "testing"

func TestConstantsConsistency(t *testing.T) {
	// kT at 300 K in J
	kt300 := BoltzmannConstantJPerK * 300.0
	if kt300 <= 0 {
		t.Errorf("kT at 300K = %e, expected positive", kt300)
	}

	// Ratio k in J/K to k in eV/K should equal e
	ratio := BoltzmannConstantJPerK / BoltzmannConstanteVPerK
	const eps = 0.01 // 1% tolerance for floating rounding
	if ratio < ElectronChargeC*(1-eps) || ratio > ElectronChargeC*(1+eps) {
		t.Errorf("k_J/K / k_eV/K = %e, expected %e (ratio should equal e)", ratio, ElectronChargeC)
	}

	// All constants positive
	for name, v := range map[string]float64{
		"BoltzmannConstantJPerK":  BoltzmannConstantJPerK,
		"BoltzmannConstanteVPerK": BoltzmannConstanteVPerK,
		"ElectronChargeC":         ElectronChargeC,
		"VacuumPermittivityFPerM": VacuumPermittivityFPerM,
	} {
		if v <= 0 {
			t.Errorf("%s = %e, expected positive", name, v)
		}
	}
}

func TestConstantsKnownValues(t *testing.T) {
	tests := []struct {
		name string
		got  float64
		want float64
		eps  float64
	}{
		{"BoltzmannConstantJPerK", BoltzmannConstantJPerK, 1.380649e-23, 1e-30},
		{"BoltzmannConstanteVPerK", BoltzmannConstanteVPerK, 8.617333262145e-05, 1e-10},
		{"ElectronChargeC", ElectronChargeC, 1.602176634e-19, 1e-25},
		{"VacuumPermittivityFPerM", VacuumPermittivityFPerM, 8.8541878128e-12, 1e-18},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := tt.got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.eps {
				t.Errorf("%s = %.20e, want %.20e (diff %e > %e)", tt.name, tt.got, tt.want, diff, tt.eps)
			}
		})
	}
}
