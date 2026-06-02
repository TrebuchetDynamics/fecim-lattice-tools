package crossbar

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule2ViewModelE2EWideResizeMVMContractMatrix(t *testing.T) {
	sizes := []struct{ rows, cols int }{{2, 3}, {4, 4}, {6, 5}, {8, 2}}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", size.rows, size.cols), func(t *testing.T) {
			m := New(1, 1)
			if err := m.ApplyAction(viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": strconv.Itoa(size.rows), "cols": strconv.Itoa(size.cols)}}); err != nil {
				t.Fatalf("resize error = %v", err)
			}
			snap := m.Snapshot()
			assertCrossbarE2ESnapshotContract(t, snap, size.rows, size.cols)
			conductancePlot := crossbarE2EPlotByID(t, snap, "conductance_matrix")
			if got, want := len(conductancePlot.Series[0].Points), size.rows*size.cols; got != want {
				t.Fatalf("conductance plot points = %d, want %d", got, want)
			}
			initialOutput := crossbarE2EPlotByID(t, snap, "mvm_result").Series[0].Points
			for _, point := range initialOutput {
				if point.Y != 0 {
					t.Fatalf("initial output point = %+v, want zero vector", point)
				}
			}

			if err := m.ApplyAction(viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand}); err != nil {
				t.Fatalf("run_mvm error = %v", err)
			}
			after := m.Snapshot()
			assertCrossbarE2ESnapshotContract(t, after, size.rows, size.cols)
			outputPlot := crossbarE2EPlotByID(t, after, "mvm_result")
			if len(outputPlot.Series[0].Points) != size.rows {
				t.Fatalf("MVM output points = %d, want rows %d", len(outputPlot.Series[0].Points), size.rows)
			}
			positive := false
			for _, point := range outputPlot.Series[0].Points {
				if point.Y > 0 {
					positive = true
				}
			}
			if !positive {
				t.Fatalf("run_mvm produced no positive output points: %+v", outputPlot.Series[0].Points)
			}

			designState := m.DesignState()
			if designState.ArrayRows != size.rows || designState.ArrayCols != size.cols {
				t.Fatalf("DesignState = %dx%d, want %dx%d", designState.ArrayRows, designState.ArrayCols, size.rows, size.cols)
			}
		})
	}
}

func TestModule2ViewModelE2EInvalidActionsPreserveSnapshot(t *testing.T) {
	m := New(4, 4)
	if err := m.ApplyAction(viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("initial run_mvm error = %v", err)
	}
	baseline := m.Snapshot()
	invalids := []viewmodel.Action{
		{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "0", "cols": "4"}},
		{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "4", "cols": "129"}},
		{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "wide", "cols": "4"}},
		{ID: "unsupported", Kind: viewmodel.ActionCommand},
	}
	for _, action := range invalids {
		err := m.ApplyAction(action)
		if err == nil {
			t.Fatalf("invalid action %+v returned nil error", action)
		}
		after := m.Snapshot()
		if crossbarE2EMetric(after, "rows") != crossbarE2EMetric(baseline, "rows") || crossbarE2EMetric(after, "cols") != crossbarE2EMetric(baseline, "cols") {
			t.Fatalf("invalid action %+v changed dimensions: before %sx%s after %sx%s", action, crossbarE2EMetric(baseline, "rows"), crossbarE2EMetric(baseline, "cols"), crossbarE2EMetric(after, "rows"), crossbarE2EMetric(after, "cols"))
		}
		if len(crossbarE2EPlotByID(t, after, "mvm_result").Series[0].Points) != len(crossbarE2EPlotByID(t, baseline, "mvm_result").Series[0].Points) {
			t.Fatalf("invalid action %+v changed output vector length", action)
		}
	}
}

func TestModule2ViewModelE2EToggleAndLifecycleWorkflow(t *testing.T) {
	m := New(3, 3)
	m.Start()
	defer m.Stop()
	before := m.Snapshot()
	if crossbarE2EMetric(before, "ir_drop") != "0.0%" {
		t.Fatalf("initial ir_drop metric = %q", crossbarE2EMetric(before, "ir_drop"))
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "toggle_ir", Kind: viewmodel.ActionToggle}); err != nil {
		t.Fatalf("toggle_ir error = %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run_mvm error = %v", err)
	}
	after := m.Snapshot()
	assertCrossbarE2ESnapshotContract(t, after, 3, 3)
	for _, marker := range []string{"SIMULATION OUTPUT", "30 levels", "Ohm", "Kirchhoff"} {
		joined := after.Descriptor.BoundaryNotice + "\n" + crossbarE2EAllSectionText(after)
		if !strings.Contains(joined, marker) {
			t.Fatalf("snapshot documentation missing %q", marker)
		}
	}
}

func assertCrossbarE2ESnapshotContract(t *testing.T, snap viewmodel.ModuleSnapshot, rows, cols int) {
	t.Helper()
	if snap.Descriptor.ID != viewmodel.ModuleCrossbar || snap.Descriptor.Status != viewmodel.StatusFunctional {
		t.Fatalf("descriptor = %+v", snap.Descriptor)
	}
	if crossbarE2EMetric(snap, "rows") != strconv.Itoa(rows) || crossbarE2EMetric(snap, "cols") != strconv.Itoa(cols) {
		t.Fatalf("metrics rows/cols = %q/%q, want %d/%d", crossbarE2EMetric(snap, "rows"), crossbarE2EMetric(snap, "cols"), rows, cols)
	}
	for _, id := range []string{"ir_drop", "sneak", "mvm", "edu_ohm", "edu_irdrop", "research_drift", "design_array"} {
		if !crossbarE2EHasSection(snap, id) {
			t.Fatalf("snapshot missing section %q", id)
		}
	}
	for _, id := range []string{"resize", "run_mvm", "toggle_ir"} {
		if !crossbarE2EHasAction(snap, id) {
			t.Fatalf("snapshot missing action %q", id)
		}
	}
	_ = crossbarE2EPlotByID(t, snap, "conductance_matrix")
	_ = crossbarE2EPlotByID(t, snap, "mvm_result")
}

func crossbarE2EMetric(snap viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range snap.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}

func crossbarE2EHasSection(snap viewmodel.ModuleSnapshot, id string) bool {
	for _, section := range snap.Sections {
		if section.ID == id {
			return true
		}
	}
	return false
}

func crossbarE2EHasAction(snap viewmodel.ModuleSnapshot, id string) bool {
	for _, action := range snap.Actions {
		if action.ID == id {
			return true
		}
	}
	return false
}

func crossbarE2EPlotByID(t *testing.T, snap viewmodel.ModuleSnapshot, id string) viewmodel.PlotData {
	t.Helper()
	for _, plot := range snap.Plots {
		if plot.ID == id {
			if len(plot.Series) == 0 {
				t.Fatalf("plot %q has no series", id)
			}
			return plot
		}
	}
	t.Fatalf("snapshot missing plot %q", id)
	return viewmodel.PlotData{}
}

func crossbarE2EAllSectionText(snap viewmodel.ModuleSnapshot) string {
	var b strings.Builder
	for _, section := range snap.Sections {
		b.WriteString(section.Title)
		b.WriteString("\n")
		b.WriteString(section.Body)
		b.WriteString("\n")
	}
	return b.String()
}
