package viewmodel_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"
	docsvm "fecim-lattice-tools/shared/viewmodel/docs"
	edavm "fecim-lattice-tools/shared/viewmodel/eda"
	hystvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
	mnistvm "fecim-lattice-tools/shared/viewmodel/mnist"
)

func TestAllModulesE2ESnapshotTimelineStateTransitions(t *testing.T) {
	tmp := t.TempDir()
	cases := []struct {
		name     string
		module   viewmodel.ModulePort
		phases   []timelinePhaseE2E
		validate func(t *testing.T, snaps []viewmodel.ModuleSnapshot)
	}{
		{
			name:   "hysteresis",
			module: hystvm.New(),
			phases: []timelinePhaseE2E{
				{label: "baseline"},
				{label: "configured", actions: []viewmodel.Action{{ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": hystvm.WaveformSquare}}, {ID: hystvm.EventSetFieldRange, Payload: map[string]string{"min": "-1600", "max": "1600"}}}},
				{label: "diagnostics", actions: []viewmodel.Action{{ID: hystvm.EventRunPUND}, {ID: hystvm.EventRunFORC}, {ID: hystvm.EventRunLevelCalibration}}},
				{label: "exports", actions: []viewmodel.Action{{ID: hystvm.EventExportCSV, Payload: map[string]string{"path": filepath.Join(tmp, "timeline-hysteresis-loop.csv")}}, {ID: hystvm.EventExportPUNDCSV, Payload: map[string]string{"path": filepath.Join(tmp, "timeline-hysteresis-pund.csv")}}}},
			},
			validate: validateHysteresisTimelineE2E,
		},
		{
			name:   "crossbar",
			module: crossbarvm.New(4, 4),
			phases: []timelinePhaseE2E{
				{label: "baseline"},
				{label: "resized", actions: []viewmodel.Action{{ID: "resize", Payload: map[string]string{"rows": "14", "cols": "6"}}}},
				{label: "nonideal", actions: []viewmodel.Action{{ID: "toggle_ir"}}},
				{label: "mvm", actions: []viewmodel.Action{{ID: "run_mvm"}}},
			},
			validate: validateCrossbarTimelineE2E,
		},
		{
			name:   "mnist",
			module: mnistvm.New(),
			phases: []timelinePhaseE2E{
				{label: "baseline"},
				{label: "low_levels", actions: []viewmodel.Action{{ID: "sweep_levels", Payload: map[string]string{"levels": "8"}}}},
				{label: "nominal_levels", actions: []viewmodel.Action{{ID: "sweep_levels", Payload: map[string]string{"levels": "30"}}}},
				{label: "inference", actions: []viewmodel.Action{{ID: "run_inference"}}},
			},
			validate: validateMNISTTimelineE2E,
		},
		{
			name:   "circuits",
			module: circuitsvm.New(),
			phases: []timelinePhaseE2E{
				{label: "baseline"},
				{label: "configured", actions: []viewmodel.Action{{ID: circuitsvm.ActionResizeArray, Payload: map[string]string{"rows": "32", "cols": "16"}}, {ID: circuitsvm.ActionSetArchitecture, Payload: map[string]string{"architecture": circuitsvm.Architecture2T1R}}, {ID: circuitsvm.ActionSetADCBits, Payload: map[string]string{"bits": "8"}}, {ID: circuitsvm.ActionSetDACBits, Payload: map[string]string{"bits": "7"}}}},
				{label: "run_paths", actions: []viewmodel.Action{{ID: circuitsvm.ActionRunRead}, {ID: circuitsvm.ActionRunWrite}, {ID: circuitsvm.ActionRunCompute}}},
				{label: "exports", actions: []viewmodel.Action{{ID: circuitsvm.ActionExportOperationLog, Payload: map[string]string{"path": filepath.Join(tmp, "timeline-circuits-log.json")}}, {ID: circuitsvm.ActionExportReferenceTimingSVG, Payload: map[string]string{"path": filepath.Join(tmp, "timeline-circuits-timing.svg")}}}},
			},
			validate: validateCircuitsTimelineE2E,
		},
		{
			name:     "comparison",
			module:   comparisonvm.New(),
			phases:   []timelinePhaseE2E{{label: "baseline"}, {label: "read_again"}},
			validate: validateComparisonTimelineE2E,
		},
		{
			name:   "eda",
			module: edavm.New(),
			phases: []timelinePhaseE2E{
				{label: "baseline"},
				{label: "spice", actions: []viewmodel.Action{{ID: "generate_spice"}}},
				{label: "all", actions: []viewmodel.Action{{ID: "generate_all"}}},
			},
			validate: validateEDATimelineE2E,
		},
		{
			name:   "docs",
			module: docsvm.New(),
			phases: []timelinePhaseE2E{
				{label: "baseline"},
				{label: "search", actions: []viewmodel.Action{{ID: "search", Payload: map[string]string{"query": "timeline trust"}}}},
				{label: "curriculum", actions: []viewmodel.Action{{ID: "start_curriculum"}}},
			},
			validate: validateDocsTimelineE2E,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.module.Start()
			defer tc.module.Stop()
			snaps := make([]viewmodel.ModuleSnapshot, 0, len(tc.phases))
			for _, phase := range tc.phases {
				for _, action := range phase.actions {
					if err := tc.module.ApplyAction(action); err != nil {
						t.Fatalf("phase %s action %s: %v", phase.label, action.ID, err)
					}
				}
				snap := tc.module.Snapshot()
				assertGenericModuleSnapshotE2E(t, snap)
				snaps = append(snaps, snap)
			}
			tc.validate(t, snaps)
		})
	}
}

type timelinePhaseE2E struct {
	label   string
	actions []viewmodel.Action
}

func validateHysteresisTimelineE2E(t *testing.T, snaps []viewmodel.ModuleSnapshot) {
	t.Helper()
	if metricValueAllE2E(snaps[1], "waveform") != hystvm.WaveformSquare || !strings.Contains(metricValueAllE2E(snaps[1], "field_min"), "-1600") {
		t.Fatalf("hysteresis configured metrics invalid: %+v", snaps[1].Metrics)
	}
	if metricValueAllE2E(snaps[2], "pund_switching_ratio") == "" || metricValueAllE2E(snaps[2], "forc_curves") == "" {
		t.Fatalf("hysteresis diagnostics missing: %+v", snaps[2].Metrics)
	}
	assertMetricContainsE2E(t, snaps[3], "csv_export", "wrote")
	assertMetricContainsE2E(t, snaps[3], "pund_export", "wrote")
	assertExistingFileMetricE2E(t, snaps[3], "csv_export_path")
	assertExistingFileMetricE2E(t, snaps[3], "pund_export_path")
}

func validateCrossbarTimelineE2E(t *testing.T, snaps []viewmodel.ModuleSnapshot) {
	t.Helper()
	if metricValueAllE2E(snaps[0], "rows") == metricValueAllE2E(snaps[1], "rows") || metricValueAllE2E(snaps[1], "rows") != "14" || metricValueAllE2E(snaps[1], "cols") != "6" {
		t.Fatalf("crossbar resize transition invalid: before=%+v after=%+v", snaps[0].Metrics, snaps[1].Metrics)
	}
	if sectionBodyAllE2E(snaps[2], "ir_drop") == "" || sectionBodyAllE2E(snaps[2], "sneak") == "" {
		t.Fatalf("crossbar nonideal sections missing after toggle: %+v", snaps[2].Sections)
	}
	if plotByIDAllE2E(snaps[3], "mvm_result") == nil {
		t.Fatal("crossbar MVM plot missing")
	}
}

func validateMNISTTimelineE2E(t *testing.T, snaps []viewmodel.ModuleSnapshot) {
	t.Helper()
	if metricValueAllE2E(snaps[1], "levels") != "8 levels" || metricValueAllE2E(snaps[2], "levels") != "30 levels" {
		t.Fatalf("mnist level transitions invalid: %q -> %q", metricValueAllE2E(snaps[1], "levels"), metricValueAllE2E(snaps[2], "levels"))
	}
	if metricValueAllE2E(snaps[3], "correct") == "" || sectionBodyAllE2E(snaps[3], "research_benchmark") == "" {
		t.Fatalf("mnist inference surface missing: %+v", snaps[3])
	}
}

func validateCircuitsTimelineE2E(t *testing.T, snaps []viewmodel.ModuleSnapshot) {
	t.Helper()
	assertMetricContainsE2E(t, snaps[1], "array", "32x16")
	assertMetricContainsE2E(t, snaps[1], "architecture", circuitsvm.Architecture2T1R)
	assertMetricContainsE2E(t, snaps[1], "adc", "8-bit")
	assertMetricContainsE2E(t, snaps[1], "dac", "7-bit")
	assertMetricContainsE2E(t, snaps[2], "operation_log_count", "total")
	assertMetricContainsE2E(t, snaps[3], "operation_log_export", "wrote")
	assertMetricContainsE2E(t, snaps[3], "reference_timing_svg_export", "wrote")
	assertExistingFileMetricE2E(t, snaps[3], "operation_log_export_path")
	assertExistingFileMetricE2E(t, snaps[3], "reference_timing_svg_export_path")
}

func validateComparisonTimelineE2E(t *testing.T, snaps []viewmodel.ModuleSnapshot) {
	t.Helper()
	if metricValueAllE2E(snaps[0], "count") != "3" || allModuleCoreSignatureE2E(snaps[0], []string{"count"}) != allModuleCoreSignatureE2E(snaps[1], []string{"count"}) {
		t.Fatalf("comparison read-only timeline changed unexpectedly: before=%+v after=%+v", snaps[0], snaps[1])
	}
}

func validateEDATimelineE2E(t *testing.T, snaps []viewmodel.ModuleSnapshot) {
	t.Helper()
	if sectionBodyAllE2E(snaps[1], "spice_content") == "" {
		t.Fatal("EDA SPICE content missing after generate_spice")
	}
	for _, id := range []string{"spice_content", "verilog_content", "def_content", "lef_content"} {
		if sectionBodyAllE2E(snaps[2], id) == "" {
			t.Fatalf("EDA all-format section %s missing", id)
		}
	}
}

func validateDocsTimelineE2E(t *testing.T, snaps []viewmodel.ModuleSnapshot) {
	t.Helper()
	if metricValueAllE2E(snaps[1], "search_query") != "timeline trust" || sectionBodyAllE2E(snaps[1], "search_results") == "" {
		t.Fatalf("docs search state missing: %+v", snaps[1])
	}
	if metricValueAllE2E(snaps[2], "active_page") != "curriculum" || sectionBodyAllE2E(snaps[2], "curriculum") == "" {
		t.Fatalf("docs curriculum state missing: %+v", snaps[2])
	}
}

func assertMetricContainsE2E(t *testing.T, snap viewmodel.ModuleSnapshot, metricID, want string) {
	t.Helper()
	if got := metricValueAllE2E(snap, metricID); !strings.Contains(got, want) {
		t.Fatalf("metric %s=%q want contains %q", metricID, got, want)
	}
}

func assertExistingFileMetricE2E(t *testing.T, snap viewmodel.ModuleSnapshot, metricID string) {
	t.Helper()
	path := metricValueAllE2E(snap, metricID)
	if path == "" {
		t.Fatalf("metric %s missing path", metricID)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s from metric %s: %v", path, metricID, err)
	}
	if info.Size() == 0 {
		t.Fatalf("file %s from metric %s is empty", path, metricID)
	}
}
