package simulation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func TestNewMultiCellArrayRejectsNonPhysicalThickness(t *testing.T) {
	tests := []struct {
		name      string
		thickness float64
	}{
		{name: "zero thickness", thickness: 0},
		{name: "negative thickness", thickness: -1e-9},
		{name: "nan thickness", thickness: math.NaN()},
		{name: "positive infinite thickness", thickness: math.Inf(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			material := *ferroelectric.DefaultHZO()
			material.Thickness = tt.thickness

			_, err := NewMultiCellArray(1, 1, &material)
			if err == nil {
				t.Fatalf("expected error for thickness %.3e m", tt.thickness)
			}
		})
	}
}

func TestNewMultiCellArrayRejectsNonPhysicalCoreMaterial(t *testing.T) {
	material := *ferroelectric.DefaultHZO()
	material.Ps = 0

	_, err := NewMultiCellArray(1, 1, &material)
	if err == nil {
		t.Fatal("expected error for nonphysical core material")
	}
}

func TestNewMultiCellArrayRejectsUnrepresentableDimensions(t *testing.T) {
	cases := []struct {
		name string
		rows int
		cols int
	}{
		{name: "huge rows", rows: math.MaxInt, cols: 1},
		{name: "huge cols", rows: 1, cols: math.MaxInt},
		{name: "overflowing product", rows: math.MaxInt, cols: math.MaxInt},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("expected dimensions %dx%d to be rejected without panic, got panic: %v", tc.rows, tc.cols, r)
				}
			}()

			array, err := NewMultiCellArray(tc.rows, tc.cols, ferroelectric.DefaultHZO())
			if err == nil {
				t.Fatalf("expected dimensions %dx%d to be rejected, got array %#v", tc.rows, tc.cols, array)
			}
		})
	}
}

func TestMultiCellSnapshotsMaterialAtConstruction(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	originalVoltage := material.CoerciveVoltage()
	originalThickness := material.Thickness
	array, err := NewMultiCellArray(1, 1, material)
	if err != nil {
		t.Fatal(err)
	}

	material.Thickness = 0
	material.Ps = 0
	material.Pr = 0
	material.Ec = 0

	state, err := array.StepCell(0, 0, originalVoltage)
	if err != nil {
		t.Fatal(err)
	}
	wantField := originalVoltage / originalThickness
	if math.Abs(state.ElectricField-wantField) > math.Abs(wantField)*1e-12 {
		t.Fatalf("array used mutated material thickness: got E %.12e V/m want %.12e V/m", state.ElectricField, wantField)
	}
	assertCellStateFinite(t, state)
}

func TestMultiCellIndependentStates(t *testing.T) {
	m, err := NewMultiCellArray(2, 2, ferroelectric.DefaultHZO())
	if err != nil {
		t.Fatal(err)
	}
	vc := ferroelectric.DefaultHZO().CoerciveVoltage() * 2
	_, _ = m.StepCell(0, 0, vc)
	_, _ = m.StepCell(1, 1, -vc)
	a, _ := m.GetCellState(0, 0)
	b, _ := m.GetCellState(1, 1)
	if a.NormPol <= b.NormPol {
		t.Fatalf("expected independent states P00=%.4f P11=%.4f", a.NormPol, b.NormPol)
	}
}

func TestMultiCellStepWithVoltageMapAndSelector(t *testing.T) {
	m, err := NewMultiCellArray(3, 3, ferroelectric.DefaultHZO())
	if err != nil {
		t.Fatal(err)
	}
	if err := m.StepWithVoltageMap([][]float64{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}, {0.7, 0.8, 0.9}}); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if math.Abs(s[2][2].Voltage-0.9) > 1e-12 {
		t.Fatalf("unexpected voltage %.6f", s[2][2].Voltage)
	}
	if err := m.StepWithSelector([]CellCoord{{0, 1}, {2, 0}}, 0.33); err != nil {
		t.Fatal(err)
	}
	a, _ := m.GetCellState(0, 1)
	b, _ := m.GetCellState(2, 0)
	if math.Abs(a.Voltage-0.33) > 1e-12 || math.Abs(b.Voltage-0.33) > 1e-12 {
		t.Fatal("selector update failed")
	}
}

func TestMultiCellRejectsInvalidVoltagesWithoutMutation(t *testing.T) {
	tests := []struct {
		name string
		step func(*MultiCellArray) error
	}{
		{name: "single_cell_nan", step: func(m *MultiCellArray) error {
			_, err := m.StepCell(0, 0, math.NaN())
			return err
		}},
		{name: "voltage_map_positive_inf", step: func(m *MultiCellArray) error {
			return m.StepWithVoltageMap([][]float64{{0.1, math.Inf(1)}, {0.3, 0.4}})
		}},
		{name: "selector_negative_inf", step: func(m *MultiCellArray) error {
			return m.StepWithSelector([]CellCoord{{0, 0}, {1, 1}}, math.Inf(-1))
		}},
		{name: "single_cell_field_overflow", step: func(m *MultiCellArray) error {
			_, err := m.StepCell(0, 0, math.MaxFloat64)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMultiCellArray(2, 2, ferroelectric.DefaultHZO())
			if err != nil {
				t.Fatal(err)
			}
			if err := m.StepWithVoltageMap([][]float64{{0.11, 0.12}, {0.21, 0.22}}); err != nil {
				t.Fatal(err)
			}
			before := m.Snapshot()

			err = tt.step(m)
			if err == nil {
				t.Fatal("expected invalid voltage to be rejected")
			}

			after := m.Snapshot()
			assertSnapshotsEqual(t, before, after)
			assertSnapshotFinite(t, after)
		})
	}
}

func assertCellStateFinite(t *testing.T, state CellState) {
	t.Helper()
	values := map[string]float64{
		"Voltage":       state.Voltage,
		"ElectricField": state.ElectricField,
		"Polarization":  state.Polarization,
		"NormPol":       state.NormPol,
	}
	for name, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("expected finite %s, got %.3g", name, value)
		}
	}
}

func assertSnapshotsEqual(t *testing.T, want, got [][]CellState) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("snapshot row count changed: got %d want %d", len(got), len(want))
	}
	for r := range want {
		if len(got[r]) != len(want[r]) {
			t.Fatalf("snapshot col count changed at row %d: got %d want %d", r, len(got[r]), len(want[r]))
		}
		for c := range want[r] {
			if got[r][c] != want[r][c] {
				t.Fatalf("cell (%d,%d) mutated after rejected voltage: got %+v want %+v", r, c, got[r][c], want[r][c])
			}
		}
	}
}

func assertSnapshotFinite(t *testing.T, snapshot [][]CellState) {
	t.Helper()
	for r := range snapshot {
		for c, state := range snapshot[r] {
			values := map[string]float64{
				"Voltage":       state.Voltage,
				"ElectricField": state.ElectricField,
				"Polarization":  state.Polarization,
				"NormPol":       state.NormPol,
			}
			for name, value := range values {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					t.Fatalf("cell (%d,%d) has non-finite %s after rejected voltage: %.3g", r, c, name, value)
				}
			}
		}
	}
}
