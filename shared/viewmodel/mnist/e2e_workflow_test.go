package mnist

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule3MNISTViewModelE2EWideQuantizationWorkflow(t *testing.T) {
	m := New()
	levels := []int{2, 4, 8, 16, 30, 64, 128}
	for _, level := range levels {
		if err := m.ApplyAction(viewmodel.Action{ID: "sweep_levels", Payload: map[string]string{"levels": itoaE2E(level)}}); err != nil {
			t.Fatalf("sweep_levels %d: %v", level, err)
		}
		s := m.Snapshot()
		assertMNISTSnapshotContractE2E(t, s)
		if got := metricValueE2E(s, "levels"); got != itoaE2E(level)+" levels" {
			t.Fatalf("levels metric=%q", got)
		}
		if !strings.Contains(sectionBodyE2E(s, "pipeline"), itoaE2E(level)+"-level") || !strings.Contains(sectionBodyE2E(s, "design_tradeoff"), itoaE2E(level)) {
			t.Fatalf("sections did not reflect level %d", level)
		}
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "run_inference"}); err != nil {
		t.Fatalf("run_inference: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "unknown"}); err == nil {
		t.Fatal("unknown action should fail")
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "sweep_levels", Payload: map[string]string{"levels": "bad"}}); err == nil {
		t.Fatal("bad level payload should fail")
	}
}

func assertMNISTSnapshotContractE2E(t *testing.T, s viewmodel.ModuleSnapshot) {
	t.Helper()
	if s.Descriptor.ID != viewmodel.ModuleMNIST || s.Descriptor.BoundaryNotice == "" {
		t.Fatalf("descriptor invalid: %+v", s.Descriptor)
	}
	for _, id := range []string{"accuracy", "levels", "correct"} {
		if metricValueE2E(s, id) == "" {
			t.Fatalf("missing metric %s", id)
		}
	}
	for _, id := range []string{"pipeline", "nonideality", "edu_pipeline", "research_benchmark", "design_tradeoff"} {
		if sectionBodyE2E(s, id) == "" {
			t.Fatalf("missing section %s", id)
		}
	}
	if len(s.Plots) != 1 || s.Plots[0].ID != "accuracy_sweep" || len(s.Plots[0].Series) != 1 || len(s.Plots[0].Series[0].Points) == 0 {
		t.Fatalf("accuracy sweep plot invalid: %+v", s.Plots)
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

func itoaE2E(v int) string {
	if v == 0 {
		return "0"
	}
	digits := []byte{}
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	return string(digits)
}
