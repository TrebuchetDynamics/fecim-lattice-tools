package comparison

import (
	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

// Module is a read-only viewmodel for the FeCIM comparison module.
// Implements viewmodel.ModulePort. Architectures are captured at construction
// so Snapshot is deterministic across calls.
type Module struct {
	architectures []*pkg.Architecture
}

// New constructs a Module from the canonical architecture set.
func New() *Module {
	return &Module{architectures: pkg.Architectures()}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor { return descriptor() }
func (m *Module) Snapshot() viewmodel.ModuleSnapshot     { return buildSnapshot(m.architectures) }
func (m *Module) ApplyAction(viewmodel.Action) error     { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                                 {}
func (m *Module) Stop()                                  {}
