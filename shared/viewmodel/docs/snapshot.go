package docs

import "fecim-lattice-tools/shared/viewmodel"

func buildSnapshot(state DocsState) viewmodel.ModuleSnapshot {
	sections := []viewmodel.Section{
		{ID: "curriculum", Title: "Learning Curriculum", Body: "FeCIM 101 → Hysteresis → Crossbar Arrays → CIM Inference → Design Export. Guided walkthrough with interactive modules. Each module includes education, research, and design layers for progressive depth.", Category: "education"},
		{ID: "citations", Title: "Citation Browser", Body: "Filter by module, paper, or author. Track verified vs. educational claims via the honesty audit dashboard. 230+ indexed references across 23 topics. All claims are cited or marked educational.", Category: "research"},
		{ID: "glossary", Title: "Interactive Glossary", Body: "Click any term to see definition, equation, and citation. Physics terms linked to live simulation modules. Key terms: Ec (coercive field), Pr (remanent polarization), ISPP, MVM, IR drop, sneak path.", Category: "education"},
		{ID: "design_guide", Title: "Design Guide", Body: "Step-by-step accelerator design workflow. Cross-module integration: Material selection (M1) → Array configuration (M2) → Circuit specification (M4) → EDA export (M6). Design snapshot captures full system state.", Category: "design"},
		{ID: "honesty", Title: "Honesty Audit", Body: "Verified claims (peer-reviewed): HZO parameters from Materlik 2015, Park 2015. Educational defaults: 30-level quantization, energy models. Not verified: accuracy/efficiency claims without published measurement evidence. Full audit at docs/4-research/honesty-audit.md.", Category: "research"},
		{ID: "trust", Title: "Trust Boundaries", Body: "Each output is labeled: validated (golden data), literature-backed (cited), educational (simulation default), planned (not yet built), or not validated. See docs/TRUST.md for the full trust matrix.", Category: "research"},
	}
	actions := []viewmodel.Action{
		{ID: "search", Label: "Search Docs", Kind: viewmodel.ActionCommand},
		{ID: "start_curriculum", Label: "Start Curriculum", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleDocs, Title: "Documentation",
			Description:    "Curriculum, validation references, trust boundaries, and research notes.",
			Status:         viewmodel.StatusFunctional,
			BoundaryNotice: "This documentation module does not produce simulation output. All references are cited and categorized by validation status.",
		},
		Metrics: []viewmodel.Metric{
			{ID: "modules", Label: "Modules", Value: "7"},
			{ID: "papers", Label: "References", Value: "230+"},
		},
		Sections: sections, Actions: actions,
	}
}
