package crossbar

import (
	"math"
	"reflect"
	"testing"

	core "fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/viewmodel"
)

// A dot-product adapter, a miswired IR toggle, or a swallowed shape error
// breaks this workflow at the legacy public action seam.
func TestCrossbarWebWorkflow(t *testing.T) {
	m := New(2, 3)
	m.state.Conductances = [][]float64{{0.2, 0.5, 0.8}, {0.9, 0.4, 0.1}}
	m.state.InputVector = []float64{0.5, 1, 0.25}
	if err := m.ApplyAction(viewmodel.Action{ID: "toggle_ir"}); err != nil {
		t.Fatal(err)
	}
	if m.state.SneakPaths {
		t.Fatal("IR toggle changed sneak paths")
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "run_mvm"}); err != nil {
		t.Fatal(err)
	}
	a, err := core.NewArray(&core.Config{Rows: 2, Cols: 3, ADCBits: 16, DACBits: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Destroy()
	if err = a.ProgramWeightMatrix(m.state.Conductances); err != nil {
		t.Fatal(err)
	}
	want, err := a.MVMWithNonIdealities(m.state.InputVector, &core.MVMOptions{EnableIRDrop: true, Temperature: 300, Architecture: "0T1R"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.state.OutputVector, want.ActualOutput) || m.state.IRDrop != want.IRDropAnalysis.MaxIRDrop {
		t.Fatalf("adapter did not execute real IR core: %+v want %+v", m.state, want)
	}
	before := append([]float64(nil), m.state.OutputVector...)
	for _, invalid := range [][]float64{{1}, {1, math.NaN(), 0}} {
		m.state.InputVector = invalid
		if err = m.ApplyAction(viewmodel.Action{ID: "run_mvm"}); err == nil {
			t.Fatal("invalid vector accepted")
		}
		if !reflect.DeepEqual(before, m.state.OutputVector) {
			t.Fatal("failed run replaced output")
		}
	}
}
