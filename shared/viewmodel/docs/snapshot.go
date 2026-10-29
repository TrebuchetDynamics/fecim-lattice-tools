package docs

import (
	"fmt"
	"strconv"
	"strings"

	"fecim-lattice-tools/shared/viewmodel"
	core "fecim-lattice-tools/workbench/modules/docs"
)

func buildSnapshot(state DocsState, corpus *core.Corpus, loadErr error) viewmodel.ModuleSnapshot {
	s := viewmodel.ModuleSnapshot{Descriptor: viewmodel.ModuleDescriptor{ID: viewmodel.ModuleDocs, Title: "Documentation", Description: "Curated learning, glossary, citations, trust and validation evidence.", Status: viewmodel.StatusFunctional, BoundaryNotice: "This documentation module does not produce simulation output. Educational text and imported reports are not measured device validation."}, Actions: []viewmodel.Action{{ID: "search", Label: "Search Docs", Kind: viewmodel.ActionCommand}, {ID: "start_curriculum", Label: "Start Curriculum", Kind: viewmodel.ActionCommand}}}
	active := state.ActivePage
	if active == "" {
		active = "overview"
	}
	count, refs := 0, 0
	if corpus != nil {
		count = len(corpus.Documents)
		for _, d := range corpus.Documents {
			if d.Category == "Citation" {
				refs += externalCount(d.Nodes)
			}
		}
	}
	s.Metrics = []viewmodel.Metric{{ID: "modules", Label: "Modules", Value: strconv.Itoa(len(viewmodel.KnownDescriptors()))}, {ID: "papers", Label: "Curated citation links", Value: strconv.Itoa(refs)}, {ID: "documents", Label: "Indexed documents", Value: strconv.Itoa(count)}, {ID: "active_page", Label: "Active Page", Value: active}}
	if loadErr != nil {
		s.Sections = []viewmodel.Section{{ID: "error", Title: "Documentation unavailable", Body: loadErr.Error()}}
		return s
	}
	for _, item := range []struct{ id, doc, title string }{{"curriculum", "curriculum", "Learning Curriculum"}, {"citations", "citations", "Citation Browser"}, {"glossary", "glossary", "Interactive Glossary"}, {"design_guide", "crossbar", "Design Guide"}, {"honesty", "honesty", "Honesty Audit"}, {"trust", "trust", "Trust Boundaries"}} {
		d, err := corpus.Document(item.doc)
		if err != nil {
			continue
		}
		s.Sections = append(s.Sections, viewmodel.Section{ID: item.id, Title: item.title, Body: nodeText(d.Nodes), Category: "education"})
	}
	if state.SearchQuery != "" {
		p, err := corpus.Search(state.SearchQuery, "", 1)
		body := fmt.Sprintf("Search query %q: %d matches in %d indexed documents.", state.SearchQuery, p.Total, count)
		if err != nil {
			body = err.Error()
		} else {
			for _, hit := range p.Results {
				body += "\n" + hit.Title + ": " + hit.Snippet
			}
		}
		s.Sections = append([]viewmodel.Section{{ID: "search_results", Title: "Search Results", Body: body, Category: "education"}}, s.Sections...)
		s.Metrics = append(s.Metrics, viewmodel.Metric{ID: "search_query", Label: "Search Query", Value: state.SearchQuery})
	}
	return s
}
func nodeText(nodes []core.Node) string {
	var b strings.Builder
	for _, n := range nodes {
		b.WriteString(n.Text)
		b.WriteString(nodeText(n.Children))
		if n.Kind == "paragraph" || n.Kind == "heading" {
			b.WriteString("\n")
		}
	}
	return b.String()
}
func externalCount(nodes []core.Node) int {
	n := 0
	for _, v := range nodes {
		if v.Kind == "link" && strings.HasPrefix(v.URL, "https://") {
			n++
		}
		n += externalCount(v.Children)
	}
	return n
}
