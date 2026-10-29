package docs

import (
	"fecim-lattice-tools/shared/viewmodel"
	"fecim-lattice-tools/workbench/moduleapi"
	core "fecim-lattice-tools/workbench/modules/docs"
)

type Module struct {
	state   DocsState
	corpus  *core.Corpus
	loadErr error
}

func New() *Module { c, err := core.Public(); return &Module{corpus: c, loadErr: err} }

// WebPort is discovered by the shared native factory, not a parallel registry.
func (m *Module) WebPort() moduleapi.Port { return newNativePort() }
func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleDocs, Title: "Documentation",
		Description: "Curriculum, validation references, trust boundaries, and research notes.",
		Status:      viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot {
	return buildSnapshot(m.state, m.corpus, m.loadErr)
}
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "search":
		if q, ok := action.Payload["query"]; ok {
			if m.loadErr != nil {
				return m.loadErr
			}
			if _, err := m.corpus.Search(q, "", 1); err != nil {
				return err
			}
			m.state.SearchQuery = q
			return nil
		}
		return viewmodel.ErrUnsupportedAction
	case "start_curriculum":
		m.state.ActivePage = "curriculum"
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start() {}
func (m *Module) Stop()  {}
