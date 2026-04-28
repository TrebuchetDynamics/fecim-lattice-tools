package comparison

import (
	"testing"

	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestNew_ReturnsModuleWithCanonicalArchitectures(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	snap := m.Snapshot()
	want := len(pkg.Architectures())
	if len(snap.Sections) != want {
		t.Errorf("New().Snapshot() Sections = %d, want %d (canonical architecture count)", len(snap.Sections), want)
	}
}

func TestModule_DescriptorMatchesSnapshot(t *testing.T) {
	m := New()
	desc := m.Descriptor()
	snap := m.Snapshot()
	if desc != snap.Descriptor {
		t.Errorf("Descriptor() and Snapshot().Descriptor disagree\n  desc: %+v\n  snap: %+v", desc, snap.Descriptor)
	}
	if desc.ID != viewmodel.ModuleComparison {
		t.Errorf("Descriptor().ID = %q, want %q", desc.ID, viewmodel.ModuleComparison)
	}
	if desc.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor().Status = %q, want %q (no longer placeholder)", desc.Status, viewmodel.StatusFunctional)
	}
}

func TestModule_ApplyAction_ReturnsErrUnsupported(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{ID: "anything"})
	if err != viewmodel.ErrUnsupportedAction {
		t.Errorf("ApplyAction error = %v, want viewmodel.ErrUnsupportedAction", err)
	}
}

func TestModule_StartStop_AreNoOpsAndIdempotent(t *testing.T) {
	m := New()
	m.Start()
	m.Start()
	m.Stop()
	m.Stop()
	// reaching here without panic is the assertion
}

func TestModule_SatisfiesModulePortInterface(t *testing.T) {
	var _ viewmodel.ModulePort = New()
}
