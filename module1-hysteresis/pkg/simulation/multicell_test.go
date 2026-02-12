package simulation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func TestNewMultiCellArray(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	arr, err := NewMultiCellArray(4, 5, material)
	if err != nil {
		t.Fatalf("NewMultiCellArray failed: %v", err)
	}
	rows, cols := arr.Size()
	if rows != 4 || cols != 5 {
		t.Fatalf("Size mismatch: got %dx%d", rows, cols)
	}
}

func TestMultiCellIndependentStates(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	arr, err := NewMultiCellArray(2, 2, material)
	if err != nil {
		t.Fatalf("NewMultiCellArray failed: %v", err)
	}

	vc := material.CoerciveVoltage() * 2
	if _, err := arr.StepCell(0, 0, vc); err != nil {
		t.Fatalf("StepCell(0,0) failed: %v", err)
	}
	if _, err := arr.StepCell(1, 1, -vc); err != nil {
		t.Fatalf("StepCell(1,1) failed: %v", err)
	}

	a, _ := arr.GetCellState(0, 0)
	b, _ := arr.GetCellState(1, 1)
	if !(a.NormPol > b.NormPol) {
		t.Fatalf("expected independent polarization states, got P00=%.4f P11=%.4f", a.NormPol, b.NormPol)
	}
}

func TestMultiCellStepWithVoltageMap(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	arr, err := NewMultiCellArray(2, 3, material)
	if err != nil {
		t.Fatalf("NewMultiCellArray failed: %v", err)
	}

	voltageMap := [][]float64{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}}
	if err := arr.StepWithVoltageMap(voltageMap); err != nil {
		t.Fatalf("StepWithVoltageMap failed: %v", err)
	}

	s := arr.Snapshot()
	if len(s) != 2 || len(s[0]) != 3 {
		t.Fatalf("snapshot size mismatch")
	}
	if math.Abs(s[1][2].Voltage-0.6) > 1e-12 {
		t.Fatalf("unexpected voltage in snapshot: %.6f", s[1][2].Voltage)
	}
}

func TestStepWithSelector(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	arr, err := NewMultiCellArray(3, 3, material)
	if err != nil {
		t.Fatalf("NewMultiCellArray failed: %v", err)
	}

	sel := []CellCoord{{Row: 0, Col: 1}, {Row: 2, Col: 2}}
	if err := arr.StepWithSelector(sel, 0.7); err != nil {
		t.Fatalf("StepWithSelector failed: %v", err)
	}

	a, _ := arr.GetCellState(0, 1)
	b, _ := arr.GetCellState(2, 2)
	c, _ := arr.GetCellState(1, 1)

	if math.Abs(a.Voltage-0.7) > 1e-12 || math.Abs(b.Voltage-0.7) > 1e-12 {
		t.Fatalf("selected cells did not update correctly")
	}
	if c.Voltage != 0 {
		t.Fatalf("unselected cell changed unexpectedly: %.6f", c.Voltage)
	}
}
