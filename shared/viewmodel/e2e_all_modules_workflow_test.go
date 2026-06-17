package viewmodel_test

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"
	"fecim-lattice-tools/shared/viewmodel/design"
	docsvm "fecim-lattice-tools/shared/viewmodel/docs"
	edavm "fecim-lattice-tools/shared/viewmodel/eda"
	hystvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
	mnistvm "fecim-lattice-tools/shared/viewmodel/mnist"
)

func TestAllModulesE2EWideViewModelWorkflowMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules E2E wide workflow matrix: exercises all module action sequences")
	}
	mods := []struct {
		name     string
		module   viewmodel.ModulePort
		actions  []viewmodel.Action
		metrics  []string
		sections []string
	}{
		{name: "hysteresis", module: hystvm.New(), actions: []viewmodel.Action{{ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": hystvm.WaveformTriangle}}, {ID: hystvm.EventSetFieldRange, Payload: map[string]string{"min": "-1200", "max": "1200"}}, {ID: hystvm.EventRunPUND}, {ID: hystvm.EventRunLevelCalibration}}, metrics: []string{"material", "field_min", "field_max", "waveform"}, sections: []string{"level_calibration_summary", "diagnostic_pund", "edu_pe_loop"}},
		{name: "crossbar", module: crossbarvm.New(4, 5), actions: []viewmodel.Action{{ID: "resize", Payload: map[string]string{"rows": "6", "cols": "4"}}, {ID: "run_mvm"}, {ID: "toggle_ir"}}, metrics: []string{"rows", "cols"}, sections: []string{"ir_drop", "sneak", "mvm"}},
		{name: "mnist", module: mnistvm.New(), actions: []viewmodel.Action{{ID: "sweep_levels", Payload: map[string]string{"levels": "64"}}, {ID: "run_inference"}}, metrics: []string{"accuracy", "levels", "correct"}, sections: []string{"pipeline", "design_tradeoff"}},
		{name: "circuits", module: circuitsvm.New(), actions: []viewmodel.Action{{ID: circuitsvm.ActionResizeArray, Payload: map[string]string{"rows": "16", "cols": "8"}}, {ID: circuitsvm.ActionSetArchitecture, Payload: map[string]string{"architecture": circuitsvm.Architecture1T1R}}, {ID: circuitsvm.ActionRunRead}, {ID: circuitsvm.ActionRunWrite}, {ID: circuitsvm.ActionRunCompute}}, metrics: []string{"array", "architecture", "adc", "dac", "tia"}, sections: []string{"read_path", "write_path", "operation_log"}},
		{name: "comparison", module: comparisonvm.New(), actions: nil, metrics: []string{"count"}, sections: []string{"fecim-cim"}},
		{name: "eda", module: edavm.New(), actions: []viewmodel.Action{{ID: "generate_spice"}, {ID: "generate_all"}}, metrics: []string{"process", "design"}, sections: []string{"design_stats", "design_workflow"}},
		{name: "docs", module: docsvm.New(), actions: []viewmodel.Action{{ID: "search", Payload: map[string]string{"query": "trust boundary"}}, {ID: "start_curriculum"}}, metrics: []string{"modules", "papers", "active_page"}, sections: []string{"curriculum", "trust", "search_results"}},
	}
	seen := map[viewmodel.ModuleID]bool{}
	for _, tc := range mods {
		t.Run(tc.name, func(t *testing.T) {
			tc.module.Start()
			defer tc.module.Stop()
			for _, action := range tc.actions {
				if err := tc.module.ApplyAction(action); err != nil {
					t.Fatalf("%s ApplyAction(%s): %v", tc.name, action.ID, err)
				}
			}
			s := tc.module.Snapshot()
			if s.Descriptor.ID == "" || s.Descriptor.Title == "" || s.Descriptor.Status != viewmodel.StatusFunctional {
				t.Fatalf("descriptor invalid: %+v", s.Descriptor)
			}
			seen[s.Descriptor.ID] = true
			for _, id := range tc.metrics {
				if got := metricValueAllE2E(s, id); got == "" {
					t.Fatalf("missing metric %s in %+v", id, s.Metrics)
				}
			}
			for _, id := range tc.sections {
				if got := sectionBodyAllE2E(s, id); got == "" {
					t.Fatalf("missing section %s", id)
				}
			}
			if err := tc.module.ApplyAction(viewmodel.Action{ID: "definitely_unsupported"}); err == nil {
				t.Fatalf("%s unsupported action should fail", tc.name)
			}
		})
	}
	for _, id := range []viewmodel.ModuleID{viewmodel.ModuleHysteresis, viewmodel.ModuleCrossbar, viewmodel.ModuleMNIST, viewmodel.ModuleCircuits, viewmodel.ModuleComparison, viewmodel.ModuleEDA, viewmodel.ModuleDocs} {
		if !seen[id] {
			t.Fatalf("module %s not exercised; seen=%v", id, seen)
		}
	}
}

func TestAllModulesE2EDesignCompositionPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules E2E design composition pipeline: full multi-module design flow")
	}
	h := hystvm.New()
	x := crossbarvm.New(12, 10)
	c := circuitsvm.New()
	e := edavm.New()
	if err := x.ApplyAction(viewmodel.Action{ID: "resize", Payload: map[string]string{"rows": "32", "cols": "16"}}); err != nil {
		t.Fatalf("resize crossbar: %v", err)
	}
	if err := c.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionSetADCBits, Payload: map[string]string{"bits": "7"}}); err != nil {
		t.Fatalf("set adc: %v", err)
	}
	if err := c.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionSetDACBits, Payload: map[string]string{"bits": "8"}}); err != nil {
		t.Fatalf("set dac: %v", err)
	}
	comp := design.Composition{Hysteresis: h, Crossbar: x, Circuits: c, EDA: e}
	snap := comp.Snapshot()
	if snap.Material == "" || snap.ArrayRows != 32 || snap.ArrayCols != 16 || snap.ADCResolution != 7 || snap.DACResolution != 8 || snap.ProcessNode == "" || snap.DesignName == "" {
		t.Fatalf("design snapshot incomplete: %+v", snap)
	}
	if !strings.Contains(snap.Summary, "32×16") || !strings.Contains(snap.Summary, "7-bit ADC/8-bit DAC") {
		t.Fatalf("design summary missing state: %q", snap.Summary)
	}
	if err := comp.ExportDesign(); err != nil {
		t.Fatalf("ExportDesign: %v", err)
	}
	if err := (&design.Composition{}).ExportDesign(); err == nil {
		t.Fatal("ExportDesign without EDA should fail")
	}
}

func metricValueAllE2E(s viewmodel.ModuleSnapshot, id string) string {
	for _, m := range s.Metrics {
		if m.ID == id {
			return m.Value
		}
	}
	return ""
}

func sectionBodyAllE2E(s viewmodel.ModuleSnapshot, id string) string {
	for _, sec := range s.Sections {
		if sec.ID == id {
			return sec.Body
		}
	}
	return ""
}
