package circuits

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule4CircuitsViewModelE2EWideWorkflowArtifactMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E wide workflow matrix: full circuits module action sequence")
	}
	m := New()
	workflow := []viewmodel.Action{
		{ID: ActionResizeArray, Payload: map[string]string{"rows": "16", "cols": "8"}},
		{ID: ActionSetArchitecture, Payload: map[string]string{"architecture": Architecture2T1R}},
		{ID: ActionSetCouplingTier, Payload: map[string]string{"tier": CouplingTierB}},
		{ID: ActionSelectCell, Payload: map[string]string{"row": "15", "col": "7"}},
		{ID: ActionSetWriteTarget, Payload: map[string]string{"level": "23"}},
		{ID: ActionSetDACBits, Payload: map[string]string{"bits": "8"}},
		{ID: ActionSetADCBits, Payload: map[string]string{"bits": "7"}},
		{ID: ActionSetTIAGain, Payload: map[string]string{"gain_ohm": "75000"}},
		{ID: ActionSetISPPEngine, Payload: map[string]string{"engine": ISPPEngineLK}},
		{ID: ActionSetTimingOperation, Payload: map[string]string{"operation": "compute"}},
		{ID: ActionSetLoggerVerbosity, Payload: map[string]string{"verbosity": "debug"}},
		{ID: ActionRunRead},
		{ID: ActionRunWrite},
		{ID: ActionRunCompute},
		{ID: ActionAnimateReferenceTiming},
		{ID: ActionStepReferenceTiming},
		{ID: ActionPlayReferenceTiming, Payload: map[string]string{"interval_ms": "250"}},
		{ID: ActionPauseReferenceTiming},
	}
	for _, action := range workflow {
		if err := m.ApplyAction(action); err != nil {
			t.Fatalf("ApplyAction(%s): %v", action.ID, err)
		}
		assertCircuitsSnapshotContractE2E(t, m.Snapshot())
	}

	s := m.Snapshot()
	want := map[string]string{
		"array":                              "16x8",
		"mode":                               "COMPUTE",
		"architecture":                       Architecture2T1R,
		"selected_cell":                      "[15,7]",
		"write_target":                       "23/29",
		"dac":                                "8-bit R-2R",
		"adc":                                "7-bit SAR",
		"tia":                                "75 kΩ",
		"coupling":                           CouplingTierB,
		"ispp_engine":                        ISPPEngineLK,
		"timing_operation":                   "COMPUTE",
		"logger_verbosity":                   "debug",
		"reference_timing_playback_interval": "250 ms",
	}
	for id, expected := range want {
		if got := metricValue(s, id); got != expected {
			t.Fatalf("metric %s = %q, want %q", id, got, expected)
		}
	}
	if got := metricValue(s, "compute_run"); !strings.Contains(got, "16x8") {
		t.Fatalf("compute_run metric = %q, want 16x8", got)
	}
	if !hasSection(s, "compute_run_log") || !hasSection(s, "reference_timing_waveform_panel") || len(s.Plots) == 0 {
		t.Fatalf("snapshot missing expected sections/plots: %+v", s)
	}

	tmp := t.TempDir()
	exports := []struct {
		action string
		path   string
		key    string
		want   string
	}{
		{ActionExportOperationLog, filepath.Join(tmp, "operation-log.json"), "schema", "fecim.circuits.operation_log.v1"},
		{ActionExportReferenceSpecs, filepath.Join(tmp, "reference-specs.json"), "schema", "fecim.circuits.reference_specs.v1"},
		{ActionExportReferenceTiming, filepath.Join(tmp, "reference-timing.json"), "schema", "fecim.circuits.reference_timing.v1"},
	}
	for _, ex := range exports {
		if err := m.ApplyAction(viewmodel.Action{ID: ex.action, Payload: map[string]string{"path": ex.path}}); err != nil {
			t.Fatalf("%s: %v", ex.action, err)
		}
		raw, err := os.ReadFile(ex.path)
		if err != nil {
			t.Fatalf("read %s: %v", ex.path, err)
		}
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("invalid JSON export %s: %v data=%s", ex.path, err, raw)
		}
		if payload[ex.key] != ex.want {
			t.Fatalf("%s %s = %v, want %s", ex.path, ex.key, payload[ex.key], ex.want)
		}
	}
	svg := filepath.Join(tmp, "reference-timing.svg")
	if err := m.ApplyAction(viewmodel.Action{ID: ActionExportReferenceTimingSVG, Payload: map[string]string{"path": svg}}); err != nil {
		t.Fatalf("export SVG: %v", err)
	}
	rawSVG, err := os.ReadFile(svg)
	if err != nil {
		t.Fatalf("read SVG: %v", err)
	}
	if !strings.Contains(string(rawSVG), "<svg") || !strings.Contains(string(rawSVG), "COMPUTE") {
		t.Fatalf("SVG missing expected waveform markers: %s", rawSVG)
	}
}

func TestModule4CircuitsViewModelE2EInvalidActionsPreserveState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E invalid-action state preservation: full circuits module sequence")
	}
	m := New()
	valid := []viewmodel.Action{
		{ID: ActionResizeArray, Payload: map[string]string{"rows": "8", "cols": "4"}},
		{ID: ActionSetArchitecture, Payload: map[string]string{"architecture": Architecture1T1R}},
		{ID: ActionSetOperationMode, Payload: map[string]string{"mode": OperationWrite}},
		{ID: ActionSelectCell, Payload: map[string]string{"row": "7", "col": "3"}},
		{ID: ActionSetWriteTarget, Payload: map[string]string{"level": "11"}},
	}
	for _, action := range valid {
		if err := m.ApplyAction(action); err != nil {
			t.Fatalf("valid setup %s: %v", action.ID, err)
		}
	}
	before := m.Snapshot()
	beforeCore := map[string]string{"array": metricValue(before, "array"), "mode": metricValue(before, "mode"), "architecture": metricValue(before, "architecture"), "selected_cell": metricValue(before, "selected_cell"), "write_target": metricValue(before, "write_target")}
	invalid := []viewmodel.Action{
		{ID: ActionResizeArray, Payload: map[string]string{"rows": "3", "cols": "8"}},
		{ID: ActionSetArchitecture, Payload: map[string]string{"architecture": "4T4R"}},
		{ID: ActionSetOperationMode, Payload: map[string]string{"mode": "erase"}},
		{ID: ActionSelectCell, Payload: map[string]string{"row": "8", "col": "0"}},
		{ID: ActionSetWriteTarget, Payload: map[string]string{"level": "30"}},
		{ID: ActionSetDACBits, Payload: map[string]string{"bits": "3"}},
		{ID: ActionSetADCBits, Payload: map[string]string{"bits": "9"}},
		{ID: ActionSetTIAGain, Payload: map[string]string{"gain_ohm": "0"}},
		{ID: ActionSetCouplingTier, Payload: map[string]string{"tier": "Tier-Z"}},
		{ID: ActionSetISPPEngine, Payload: map[string]string{"engine": "magic"}},
		{ID: ActionSetTimingOperation, Payload: map[string]string{"operation": "sleep"}},
		{ID: ActionSetLoggerVerbosity, Payload: map[string]string{"verbosity": "verbose++"}},
		{ID: "unknown"},
	}
	for _, action := range invalid {
		if err := m.ApplyAction(action); err == nil {
			t.Fatalf("invalid action %s unexpectedly succeeded", action.ID)
		}
		after := m.Snapshot()
		for id, expected := range beforeCore {
			if got := metricValue(after, id); got != expected {
				t.Fatalf("after invalid %s metric %s = %q, want %q", action.ID, id, got, expected)
			}
		}
		assertCircuitsSnapshotContractE2E(t, after)
	}
}

func assertCircuitsSnapshotContractE2E(t *testing.T, s viewmodel.ModuleSnapshot) {
	t.Helper()
	if s.Descriptor.ID != viewmodel.ModuleCircuits || s.Descriptor.BoundaryNotice == "" {
		t.Fatalf("descriptor contract invalid: %+v", s.Descriptor)
	}
	for _, id := range []string{"array", "mode", "architecture", "selected_cell", "write_target", "adc", "dac", "tia", "reference_specs", "reference_timing", "operation_log", "compute_run_log"} {
		if id == "reference_specs" || id == "reference_timing" || id == "operation_log" || id == "compute_run_log" {
			if !hasSection(s, id) {
				t.Fatalf("missing section %s", id)
			}
			continue
		}
		if metricValue(s, id) == "" {
			t.Fatalf("missing metric %s", id)
		}
	}
	if len(s.Actions) < 10 {
		t.Fatalf("too few actions: %d", len(s.Actions))
	}
}
