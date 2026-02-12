package simulation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func TestMultiCellIndependentStates(t *testing.T) {
	m, err := NewMultiCellArray(2, 2, ferroelectric.DefaultHZO())
	if err != nil { t.Fatal(err) }
	vc := ferroelectric.DefaultHZO().CoerciveVoltage() * 2
	_, _ = m.StepCell(0, 0, vc)
	_, _ = m.StepCell(1, 1, -vc)
	a, _ := m.GetCellState(0, 0)
	b, _ := m.GetCellState(1, 1)
	if a.NormPol <= b.NormPol { t.Fatalf("expected independent states P00=%.4f P11=%.4f", a.NormPol, b.NormPol) }
}

func TestMultiCellStepWithVoltageMapAndSelector(t *testing.T) {
	m, err := NewMultiCellArray(3, 3, ferroelectric.DefaultHZO())
	if err != nil { t.Fatal(err) }
	if err := m.StepWithVoltageMap([][]float64{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}, {0.7, 0.8, 0.9}}); err != nil { t.Fatal(err) }
	s := m.Snapshot()
	if math.Abs(s[2][2].Voltage-0.9) > 1e-12 { t.Fatalf("unexpected voltage %.6f", s[2][2].Voltage) }
	if err := m.StepWithSelector([]CellCoord{{0, 1}, {2, 0}}, 0.33); err != nil { t.Fatal(err) }
	a, _ := m.GetCellState(0, 1)
	b, _ := m.GetCellState(2, 0)
	if math.Abs(a.Voltage-0.33) > 1e-12 || math.Abs(b.Voltage-0.33) > 1e-12 { t.Fatal("selector update failed") }
}
