package arraysim

import (
	"math"
	"testing"
)

const testEpsilon = 1e-6

func TestTierA_PassiveHalfSelectPattern(t *testing.T) {
	solver := NewTierASolver()

	conductance := [][]float64{
		{10e-6, 10e-6},
		{10e-6, 10e-6},
	}
	params := SolveParams{
		WLVoltages:  []float64{0.5, 0.0},
		BLVoltages:  []float64{-0.5, 0.0},
		Conductance: conductance,
		ActiveRows:  []bool{true, true},
		Geometry:    CellGeometry{},
		Wire: WireParams{
			RWordLine: 1000,
			RBitLine:  1000,
		},
	}

	result, err := solver.Solve(params)
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if len(result.CellVoltages) != 2 || len(result.CellVoltages[0]) != 2 {
		t.Fatalf("unexpected cell voltage size: %#v", result.CellVoltages)
	}

	want := [][]float64{
		{0.9925, 0.4875},
		{0.4875, 0.0},
	}

	for r := range want {
		for c := range want[r] {
			got := result.CellVoltages[r][c]
			if math.Abs(got-want[r][c]) > testEpsilon {
				t.Fatalf("cell (%d,%d) voltage: got %.6f, want %.6f", r, c, got, want[r][c])
			}
		}
	}

	if result.CellVoltages[0][0] <= result.CellVoltages[0][1] {
		t.Fatalf("target cell should have highest voltage, got %.6f <= %.6f", result.CellVoltages[0][0], result.CellVoltages[0][1])
	}
	if result.CellVoltages[1][1] != 0 {
		t.Fatalf("diagonal cell should be 0V, got %.6f", result.CellVoltages[1][1])
	}
}

func TestSenseChain_ConvertCurrent(t *testing.T) {
	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   10e3,
			Vref: 0.1,
			Vmin: 0.0,
			Vmax: 1.0,
		},
		ADC: ADCConfig{
			Bits: 4,
			Vmin: 0.0,
			Vmax: 1.0,
		},
	}

	tests := []struct {
		name       string
		currentA   float64
		wantVout   float64
		wantCode   int
		wantTIASat bool
		wantADCSat bool
	}{
		{
			name:       "in-range",
			currentA:   40e-6,
			wantVout:   0.5,
			wantCode:   8,
			wantTIASat: false,
			wantADCSat: false,
		},
		{
			name:       "high-saturation",
			currentA:   200e-6,
			wantVout:   1.0,
			wantCode:   15,
			wantTIASat: true,
			wantADCSat: true,
		},
		{
			name:       "low-saturation",
			currentA:   -20e-6,
			wantVout:   0.0,
			wantCode:   0,
			wantTIASat: true,
			wantADCSat: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sense.ConvertCurrent(tc.currentA)
			if math.Abs(result.Vout-tc.wantVout) > testEpsilon {
				t.Fatalf("Vout: got %.6f, want %.6f", result.Vout, tc.wantVout)
			}
			if result.Code != tc.wantCode {
				t.Fatalf("Code: got %d, want %d", result.Code, tc.wantCode)
			}
			if result.TIASaturated != tc.wantTIASat {
				t.Fatalf("TIASaturated: got %v, want %v", result.TIASaturated, tc.wantTIASat)
			}
			if result.ADCSaturated != tc.wantADCSat {
				t.Fatalf("ADCSaturated: got %v, want %v", result.ADCSaturated, tc.wantADCSat)
			}
		})
	}
}
