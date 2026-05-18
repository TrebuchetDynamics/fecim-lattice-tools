//go:build !cgo

package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
)

func TestCircuitsOverlayStateIncludesHalfSelectStress(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetOperationMode,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"mode": circuitsvm.OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.halfSelectState != circuitsvm.HalfSelectStateColumnWriteActive {
		t.Fatalf("halfSelectState = %q, want %q", state.halfSelectState, circuitsvm.HalfSelectStateColumnWriteActive)
	}
	if state.halfSelectCells != 7 {
		t.Fatalf("halfSelectCells = %d, want 7", state.halfSelectCells)
	}
	if state.stressBudget != "400 pulses/level" {
		t.Fatalf("stressBudget = %q, want 400 pulses/level", state.stressBudget)
	}
}
