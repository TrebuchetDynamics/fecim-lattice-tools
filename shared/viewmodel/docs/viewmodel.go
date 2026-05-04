package docs

import "fecim-lattice-tools/shared/viewmodel"

type Module struct{ state DocsState }

func New() *Module { return &Module{} }
func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleDocs, Title: "Documentation",
		Description: "Curriculum, validation references, trust boundaries, and research notes.",
		Status: viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
