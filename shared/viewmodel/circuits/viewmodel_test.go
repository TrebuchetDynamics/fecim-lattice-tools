package circuits

import (
	"testing"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New()
	if m == nil { t.Fatal("New() returned nil") }
}
func TestDescriptorHasCorrectID(t *testing.T) {
	if New().Descriptor().ID != viewmodel.ModuleCircuits {
		t.Error("wrong ID")
	}
}
func TestSnapshotContainsCircuitBlocks(t *testing.T) {
	s := New().Snapshot()
	ids := map[string]bool{}
	for _, m := range s.Metrics { ids[m.ID] = true }
	for _, id := range []string{"adc", "dac", "tia", "charge_pump", "ispp"} {
		if !ids[id] { t.Errorf("missing: %s", id) }
	}
}
