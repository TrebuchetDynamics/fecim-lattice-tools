package docs

import (
	"fecim-lattice-tools/shared/viewmodel"
	"strings"
	"testing"
)

// An unmatched query must not manufacture the old cross-module matches.
func TestDocsWebWorkflow(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: "search", Payload: map[string]string{"query": "zzzznotincorpus"}}); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if body := sectionBodyDocsE2E(s, "search_results"); !strings.Contains(body, "0 matches") || strings.Contains(body, "matched curriculum") {
		t.Fatalf("unmatched search must report 0 matches, got %q", body)
	}
}
