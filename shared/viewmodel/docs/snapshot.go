package docs

import "fecim-lattice-tools/shared/viewmodel"

func buildSnapshot(state DocsState) viewmodel.ModuleSnapshot {
	sections := []viewmodel.Section{
		{ID: "curriculum", Title: "Learning Curriculum", Body: "FeCIM 101 → Hysteresis → Crossbar Arrays → CIM Inference → Design Export. Guided walkthrough with interactive modules."},
		{ID: "citations", Title: "Citation Browser", Body: "Filter by module, paper, or author. Track verified vs. educational claims via the honesty audit dashboard."},
		{ID: "glossary", Title: "Interactive Glossary", Body: "Click any term to see definition, equation, and citation. Physics terms linked to live simulation modules."},
		{ID: "design_guide", Title: "Design Guide", Body: "Step-by-step accelerator design workflow. Cross-module integration: Material → Array → Circuits → Export."},
	}
	actions := []viewmodel.Action{
		{ID: "search", Label: "Search Docs", Kind: viewmodel.ActionCommand},
		{ID: "start_curriculum", Label: "Start Curriculum", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleDocs, Title: "Documentation",
			Description: "Curriculum, validation references, trust boundaries, and research notes.",
			Status: viewmodel.StatusFunctional,
		},
		Metrics: []viewmodel.Metric{
			{ID: "modules", Label: "Modules", Value: "7"},
			{ID: "papers", Label: "References", Value: "230+"},
		},
		Sections: sections, Actions: actions,
	}
}
