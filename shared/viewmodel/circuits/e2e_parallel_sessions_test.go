package circuits

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule4CircuitsViewModelE2EParallelIndependentSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E parallel sessions: full circuits viewmodel workflow per session")
	}
	cases := []struct {
		name         string
		rows         string
		cols         string
		architecture string
		mode         string
		cellRow      string
		cellCol      string
		target       string
		coupling     string
		engine       string
	}{
		{name: "passive-read", rows: "8", cols: "8", architecture: ArchitecturePassive, mode: OperationRead, cellRow: "1", cellCol: "2", target: "9", coupling: CouplingTierA, engine: ISPPEngineLevel},
		{name: "1t1r-write", rows: "16", cols: "4", architecture: Architecture1T1R, mode: OperationWrite, cellRow: "15", cellCol: "3", target: "21", coupling: CouplingTierB, engine: ISPPEngineLK},
		{name: "2t1r-compute", rows: "4", cols: "16", architecture: Architecture2T1R, mode: OperationCompute, cellRow: "3", cellCol: "15", target: "5", coupling: CouplingIdeal, engine: ISPPEngineLevel},
	}
	var wg sync.WaitGroup
	for _, tc := range cases {
		tc := tc
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := New()
			workflow := []viewmodel.Action{
				{ID: ActionResizeArray, Payload: map[string]string{"rows": tc.rows, "cols": tc.cols}},
				{ID: ActionSetArchitecture, Payload: map[string]string{"architecture": tc.architecture}},
				{ID: ActionSetCouplingTier, Payload: map[string]string{"tier": tc.coupling}},
				{ID: ActionSetISPPEngine, Payload: map[string]string{"engine": tc.engine}},
				{ID: ActionSetOperationMode, Payload: map[string]string{"mode": tc.mode}},
				{ID: ActionSelectCell, Payload: map[string]string{"row": tc.cellRow, "col": tc.cellCol}},
				{ID: ActionSetWriteTarget, Payload: map[string]string{"level": tc.target}},
				{ID: ActionRunRead},
				{ID: ActionRunWrite},
				{ID: ActionRunCompute},
				{ID: ActionSetTimingOperation, Payload: map[string]string{"operation": tc.mode}},
				{ID: ActionAnimateReferenceTiming},
				{ID: ActionStepReferenceTiming},
			}
			for _, action := range workflow {
				if err := m.ApplyAction(action); err != nil {
					t.Errorf("%s ApplyAction(%s): %v", tc.name, action.ID, err)
					return
				}
				assertCircuitsSnapshotContractE2E(t, m.Snapshot())
			}
			s := m.Snapshot()
			if got := metricValue(s, "array"); got != tc.rows+"x"+tc.cols {
				t.Errorf("%s array = %q", tc.name, got)
				return
			}
			if got := metricValue(s, "architecture"); got != tc.architecture {
				t.Errorf("%s architecture = %q", tc.name, got)
				return
			}
			if got := metricValue(s, "selected_cell"); got != "["+tc.cellRow+","+tc.cellCol+"]" {
				t.Errorf("%s selected_cell = %q", tc.name, got)
				return
			}
			if got := metricValue(s, "compute_run"); !strings.Contains(got, tc.rows+"x"+tc.cols) {
				t.Errorf("%s compute_run = %q", tc.name, got)
				return
			}

			dir := t.TempDir()
			exports := map[string]string{
				ActionExportOperationLog:       filepath.Join(dir, tc.name+"-operation.json"),
				ActionExportReferenceSpecs:     filepath.Join(dir, tc.name+"-specs.json"),
				ActionExportReferenceTiming:    filepath.Join(dir, tc.name+"-timing.json"),
				ActionExportReferenceTimingSVG: filepath.Join(dir, tc.name+"-timing.svg"),
			}
			for actionID, path := range exports {
				if err := m.ApplyAction(viewmodel.Action{ID: actionID, Payload: map[string]string{"path": path}}); err != nil {
					t.Errorf("%s export %s: %v", tc.name, actionID, err)
					return
				}
				raw, err := os.ReadFile(path)
				if err != nil || len(raw) == 0 {
					t.Errorf("%s export %s read/len: %d %v", tc.name, actionID, len(raw), err)
					return
				}
				if strings.HasSuffix(path, ".json") {
					var payload map[string]any
					if err := json.Unmarshal(raw, &payload); err != nil || payload["schema"] == nil {
						t.Errorf("%s export %s invalid JSON: %v %s", tc.name, actionID, err, raw)
						return
					}
				} else if !strings.Contains(string(raw), "<svg") {
					t.Errorf("%s SVG export missing svg marker: %s", tc.name, raw)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestModule4CircuitsViewModelE2EExportTraversalFailuresAreSideEffectSafe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E export failure traversal: full circuits module workflow")
	}
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: ActionRunCompute}); err != nil {
		t.Fatalf("run compute: %v", err)
	}
	before := m.Snapshot()
	failures := []struct {
		action string
		path   string
		metric string
	}{
		{ActionExportOperationLog, "../operation.json", "operation_log_export_path"},
		{ActionExportReferenceSpecs, "../specs.json", "reference_spec_export_path"},
		{ActionExportReferenceTiming, "../timing.json", "reference_timing_export_path"},
		{ActionExportReferenceTimingSVG, "../timing.svg", "reference_timing_svg_export_path"},
	}
	for _, f := range failures {
		if err := m.ApplyAction(viewmodel.Action{ID: f.action, Payload: map[string]string{"path": f.path}}); err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Fatalf("%s traversal should fail with context, err=%v", f.action, err)
		}
		after := m.Snapshot()
		if metricValue(after, "array") != metricValue(before, "array") || metricValue(after, "mode") != metricValue(before, "mode") || metricValue(after, f.metric) != metricValue(before, f.metric) {
			t.Fatalf("%s traversal changed snapshot unexpectedly: metric %s=%q before=%q", f.action, f.metric, metricValue(after, f.metric), metricValue(before, f.metric))
		}
	}
}
