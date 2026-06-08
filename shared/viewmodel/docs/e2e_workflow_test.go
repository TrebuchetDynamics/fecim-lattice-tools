package docs

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule7DocsViewModelE2ESearchCurriculumTrustWorkflow(t *testing.T) {
	m := New()
	queries := []string{"hysteresis", "crossbar IR drop", "EDA trust boundary", "honesty audit"}
	for _, query := range queries {
		if err := m.ApplyAction(viewmodel.Action{ID: "search", Payload: map[string]string{"query": query}}); err != nil {
			t.Fatalf("search %q: %v", query, err)
		}
		s := m.Snapshot()
		assertDocsSnapshotE2E(t, s)
		if got := metricValueDocsE2E(s, "search_query"); got != query {
			t.Fatalf("search_query=%q want %q", got, query)
		}
		if body := sectionBodyDocsE2E(s, "search_results"); !strings.Contains(body, query) || !strings.Contains(body, "cross-module") {
			t.Fatalf("search results body invalid: %q", body)
		}
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "start_curriculum"}); err != nil {
		t.Fatalf("start_curriculum: %v", err)
	}
	if got := metricValueDocsE2E(m.Snapshot(), "active_page"); got != "curriculum" {
		t.Fatalf("active_page=%q", got)
	}
	for _, action := range []viewmodel.Action{{ID: "search"}, {ID: "unknown"}} {
		if err := m.ApplyAction(action); err == nil {
			t.Fatalf("action %s should fail", action.ID)
		}
	}
}

func assertDocsSnapshotE2E(t *testing.T, s viewmodel.ModuleSnapshot) {
	t.Helper()
	if s.Descriptor.ID != viewmodel.ModuleDocs || s.Descriptor.BoundaryNotice == "" {
		t.Fatalf("descriptor invalid: %+v", s.Descriptor)
	}
	for _, id := range []string{"modules", "papers", "active_page"} {
		if metricValueDocsE2E(s, id) == "" {
			t.Fatalf("missing metric %s", id)
		}
	}
	for _, id := range []string{"curriculum", "citations", "glossary", "design_guide", "honesty", "trust"} {
		if sectionBodyDocsE2E(s, id) == "" {
			t.Fatalf("missing section %s", id)
		}
	}
	if len(s.Actions) != 2 {
		t.Fatalf("actions len=%d", len(s.Actions))
	}
}

func metricValueDocsE2E(s viewmodel.ModuleSnapshot, id string) string {
	for _, m := range s.Metrics {
		if m.ID == id {
			return m.Value
		}
	}
	return ""
}

func sectionBodyDocsE2E(s viewmodel.ModuleSnapshot, id string) string {
	for _, sec := range s.Sections {
		if sec.ID == id {
			return sec.Body
		}
	}
	return ""
}
