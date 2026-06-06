package comparison

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule5ComparisonViewModelE2ESnapshotContract(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if s.Descriptor.ID != viewmodel.ModuleComparison || s.Descriptor.BoundaryNotice == "" {
		t.Fatalf("descriptor invalid: %+v", s.Descriptor)
	}
	if metricValueE2E(s, "count") != "3" || len(s.Sections) != 3 {
		t.Fatalf("comparison count/sections invalid metrics=%+v sections=%d", s.Metrics, len(s.Sections))
	}
	wantSections := []string{"traditional-cpu-dram", "gpu-accelerator", "fecim-cim"}
	for _, id := range wantSections {
		body := sectionBodyE2E(s, id)
		if body == "" || !strings.Contains(body, "TDP") || !strings.Contains(body, "TOPS") || !strings.Contains(body, "Memory") {
			t.Fatalf("section %s invalid: %q", id, body)
		}
	}
	if !strings.Contains(sectionBodyE2E(s, "fecim-cim"), "estimated") || !strings.Contains(sectionBodyE2E(s, "fecim-cim"), "compute-in-memory") {
		t.Fatalf("FeCIM section missing estimated/CIM context: %q", sectionBodyE2E(s, "fecim-cim"))
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "anything"}); err == nil {
		t.Fatal("comparison viewmodel should reject actions")
	}
	m.Start()
	m.Stop()
	if got := m.Snapshot(); metricValueE2E(got, "count") != "3" || len(got.Sections) != len(s.Sections) {
		t.Fatalf("snapshot changed across lifecycle: %+v", got)
	}
}

func metricValueE2E(s viewmodel.ModuleSnapshot, id string) string {
	for _, m := range s.Metrics {
		if m.ID == id {
			return m.Value
		}
	}
	return ""
}

func sectionBodyE2E(s viewmodel.ModuleSnapshot, id string) string {
	for _, sec := range s.Sections {
		if sec.ID == id {
			return sec.Body
		}
	}
	return ""
}
