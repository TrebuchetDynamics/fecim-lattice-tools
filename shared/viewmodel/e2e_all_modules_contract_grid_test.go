package viewmodel_test

import (
	"math"
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

func TestAllModulesE2ESnapshotContractGrid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules E2E contract grid: initializes all modules and runs PUND/FORC")
	}
	cases := []struct {
		name           string
		module         viewmodel.ModulePort
		prepare        []viewmodel.Action
		wantSections   []string
		wantActions    []string
		wantPlotIDs    []string
		contentMarkers []string
	}{
		{name: "hysteresis", module: hystvm.New(), prepare: []viewmodel.Action{{ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": hystvm.WaveformTriangle}}, {ID: hystvm.EventRunPUND}, {ID: hystvm.EventRunFORC}, {ID: hystvm.EventRunLevelCalibration}}, wantSections: []string{"level_calibration_summary", "diagnostic_pund", "diagnostic_forc", "edu_pe_loop", "research_citations"}, wantActions: []string{hystvm.EventRunPUND, hystvm.EventRunFORC, hystvm.EventExportCSV}, wantPlotIDs: []string{"pe_loop", "retention"}, contentMarkers: []string{"Preisach", "PUND", "FORC"}},
		{name: "crossbar", module: crossbarvm.New(8, 6), prepare: []viewmodel.Action{{ID: "toggle_ir"}, {ID: "run_mvm"}}, wantSections: []string{"ir_drop", "sneak", "mvm", "edu_ohm", "design_array"}, wantActions: []string{"resize", "run_mvm", "toggle_ir"}, wantPlotIDs: []string{"conductance_matrix", "mvm_result"}, contentMarkers: []string{"Ohm", "Kirchhoff", "conductance"}},
		{name: "mnist", module: mnistvm.New(), prepare: []viewmodel.Action{{ID: "sweep_levels", Payload: map[string]string{"levels": "30"}}, {ID: "run_inference"}}, wantSections: []string{"pipeline", "nonideality", "edu_pipeline", "design_tradeoff", "research_benchmark"}, wantActions: []string{"run_inference", "sweep_levels"}, wantPlotIDs: []string{"accuracy_sweep"}, contentMarkers: []string{"quantize", "MNIST", "educational"}},
		{name: "circuits", module: circuitsvm.New(), prepare: []viewmodel.Action{{ID: circuitsvm.ActionRunRead}, {ID: circuitsvm.ActionRunWrite}, {ID: circuitsvm.ActionRunCompute}}, wantSections: []string{"read_path", "write_path", "operation_log", "reference_specs", "reference_timing"}, wantActions: []string{circuitsvm.ActionRunRead, circuitsvm.ActionRunWrite, circuitsvm.ActionRunCompute}, wantPlotIDs: []string{"timing_waveform_active", "ispp_convergence"}, contentMarkers: []string{"ADC", "DAC", "TIA"}},
		{name: "comparison", module: comparisonvm.New(), wantSections: []string{"fecim-cim", "traditional-cpu-dram", "gpu-accelerator"}, contentMarkers: []string{"estimated", "CPU", "GPU"}},
		{name: "eda", module: edavm.New(), prepare: []viewmodel.Action{{ID: "generate_all"}}, wantSections: []string{"design_stats", "spice_content", "verilog_content", "def_content", "lef_content", "research_validation"}, wantActions: []string{"generate_all", "generate_spice"}, contentMarkers: []string{"SPICE", "Verilog", "Validation"}},
		{name: "docs", module: docsvm.New(), prepare: []viewmodel.Action{{ID: "search", Payload: map[string]string{"query": "validation trust"}}, {ID: "start_curriculum"}}, wantSections: []string{"curriculum", "citations", "glossary", "design_guide", "honesty", "trust", "search_results"}, wantActions: []string{"search", "start_curriculum"}, contentMarkers: []string{"trust", "validation", "curriculum"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for cycle := 0; cycle < 2; cycle++ {
				tc.module.Start()
				tc.module.Stop()
			}
			for _, action := range tc.prepare {
				if err := tc.module.ApplyAction(action); err != nil {
					t.Fatalf("prepare %s: %v", action.ID, err)
				}
			}
			snap := tc.module.Snapshot()
			if desc := tc.module.Descriptor(); desc.ID != snap.Descriptor.ID || desc.Title != snap.Descriptor.Title || desc.Status != snap.Descriptor.Status {
				t.Fatalf("Descriptor() and Snapshot descriptor diverged: direct=%+v snapshot=%+v", desc, snap.Descriptor)
			}
			assertGenericModuleSnapshotE2E(t, snap)
			assertUniqueSnapshotIDsE2E(t, tc.name, snap)
			for _, id := range tc.wantSections {
				if body := sectionBodyAllE2E(snap, id); body == "" {
					t.Fatalf("missing section %s", id)
				}
			}
			for _, id := range tc.wantActions {
				if !hasActionIDAllE2E(snap, id) {
					t.Fatalf("missing action %s in %+v", id, snap.Actions)
				}
			}
			for _, id := range tc.wantPlotIDs {
				plot := plotByIDAllE2E(snap, id)
				if plot == nil {
					t.Fatalf("missing plot %s in %+v", id, snap.Plots)
				}
				assertPlotFiniteAllE2E(t, *plot)
			}
			joinedSections := strings.ToLower(joinSectionBodiesAllE2E(snap))
			for _, marker := range tc.contentMarkers {
				if !strings.Contains(joinedSections, strings.ToLower(marker)) {
					t.Fatalf("sections missing marker %q in %q", marker, joinedSections)
				}
			}
		})
	}
}

func assertUniqueSnapshotIDsE2E(t *testing.T, name string, snap viewmodel.ModuleSnapshot) {
	t.Helper()
	metricIDs := map[string]bool{}
	for _, m := range snap.Metrics {
		if m.ID == "" || m.Label == "" || m.Value == "" {
			t.Fatalf("%s invalid metric: %+v", name, m)
		}
		if metricIDs[m.ID] {
			t.Fatalf("%s duplicate metric id %s", name, m.ID)
		}
		metricIDs[m.ID] = true
	}
	sectionIDs := map[string]bool{}
	for _, sec := range snap.Sections {
		if sec.ID == "" || sec.Title == "" || sec.Body == "" {
			t.Fatalf("%s invalid section: %+v", name, sec)
		}
		if sectionIDs[sec.ID] {
			t.Fatalf("%s duplicate section id %s", name, sec.ID)
		}
		sectionIDs[sec.ID] = true
	}
	actionIDs := map[string]bool{}
	for _, action := range snap.Actions {
		if action.ID == "" || action.Label == "" || action.Kind == "" {
			t.Fatalf("%s invalid action: %+v", name, action)
		}
		if actionIDs[action.ID] {
			t.Fatalf("%s duplicate action id %s", name, action.ID)
		}
		actionIDs[action.ID] = true
	}
}

func hasActionIDAllE2E(s viewmodel.ModuleSnapshot, id string) bool {
	for _, action := range s.Actions {
		if action.ID == id {
			return true
		}
	}
	return false
}

func plotByIDAllE2E(s viewmodel.ModuleSnapshot, id string) *viewmodel.PlotData {
	for i := range s.Plots {
		if s.Plots[i].ID == id {
			return &s.Plots[i]
		}
	}
	return nil
}

func assertPlotFiniteAllE2E(t *testing.T, p viewmodel.PlotData) {
	t.Helper()
	if p.ID == "" || p.Title == "" || len(p.Series) == 0 {
		t.Fatalf("invalid plot header: %+v", p)
	}
	for _, series := range p.Series {
		if series.Name == "" || len(series.Points) == 0 {
			t.Fatalf("invalid plot series in %s: %+v", p.ID, series)
		}
		for _, pt := range series.Points {
			if math.IsNaN(pt.X) || math.IsNaN(pt.Y) || math.IsNaN(pt.V) || math.IsInf(pt.X, 0) || math.IsInf(pt.Y, 0) || math.IsInf(pt.V, 0) {
				t.Fatalf("non-finite point in %s/%s: %+v", p.ID, series.Name, pt)
			}
		}
	}
}

func joinSectionBodiesAllE2E(s viewmodel.ModuleSnapshot) string {
	var b strings.Builder
	for _, sec := range s.Sections {
		b.WriteString(sec.Title)
		b.WriteByte('\n')
		b.WriteString(sec.Body)
		b.WriteByte('\n')
	}
	return b.String()
}
