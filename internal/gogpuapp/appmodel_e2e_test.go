package gogpuapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	edaconfig "fecim-lattice-tools/module6-eda/pkg/config"
	edaexport "fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
	"fecim-lattice-tools/shared/viewmodel/design"
	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
)

func TestFullAppE2ENavigateAndDispatchRepresentativeActions(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)

	cases := []struct {
		module   viewmodel.ModuleID
		action   viewmodel.Action
		readOnly bool
	}{
		{module: viewmodel.ModuleHysteresis, action: viewmodel.Action{ID: "toggle_simulation", Kind: viewmodel.ActionToggle}},
		{module: viewmodel.ModuleCrossbar, action: viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand}},
		{module: viewmodel.ModuleMNIST, action: viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "16"}}},
		{module: viewmodel.ModuleCircuits, action: viewmodel.Action{ID: "run_read", Kind: viewmodel.ActionCommand}},
		{module: viewmodel.ModuleComparison, action: viewmodel.Action{ID: "unsupported_e2e_probe", Kind: viewmodel.ActionCommand}, readOnly: true},
		{module: viewmodel.ModuleEDA, action: viewmodel.Action{ID: "generate_spice", Kind: viewmodel.ActionCommand}},
		{module: viewmodel.ModuleDocs, action: viewmodel.Action{ID: "start_curriculum", Kind: viewmodel.ActionCommand}},
	}

	for _, tc := range cases {
		t.Run(string(tc.module), func(t *testing.T) {
			if !model.SelectModule(tc.module) {
				t.Fatalf("SelectModule(%s) returned false", tc.module)
			}
			before := model.ActivePort().Snapshot()
			if before.Descriptor.ID != tc.module {
				t.Fatalf("active snapshot descriptor = %s, want %s", before.Descriptor.ID, tc.module)
			}
			if !tc.readOnly && !snapshotHasAction(before, tc.action.ID) {
				t.Fatalf("module %s snapshot does not expose representative action %q", tc.module, tc.action.ID)
			}

			err := model.DispatchAction(tc.action)
			if tc.readOnly {
				if !errors.Is(err, viewmodel.ErrUnsupportedAction) {
					t.Fatalf("read-only module dispatch error = %v, want ErrUnsupportedAction", err)
				}
			} else if err != nil {
				t.Fatalf("DispatchAction(%s/%s) returned error: %v", tc.module, tc.action.ID, err)
			}

			after := model.ActivePort().Snapshot()
			if after.Descriptor.ID != tc.module {
				t.Fatalf("dispatch changed active module to %s, want %s", after.Descriptor.ID, tc.module)
			}
			if len(after.Sections) == 0 {
				t.Fatalf("module %s has no sections after dispatch", tc.module)
			}
		})
	}
}

func TestFullAppE2ENavigationPreservesPerModuleState(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)

	if !model.SelectModule(viewmodel.ModuleMNIST) {
		t.Fatal("SelectModule(mnist) returned false")
	}
	if err := model.DispatchAction(viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "16"}}); err != nil {
		t.Fatalf("DispatchAction(mnist/sweep_levels): %v", err)
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "levels"); got != "16 levels" {
		t.Fatalf("mnist levels metric after dispatch = %q, want 16 levels", got)
	}

	for _, id := range []viewmodel.ModuleID{
		viewmodel.ModuleHysteresis,
		viewmodel.ModuleCrossbar,
		viewmodel.ModuleCircuits,
		viewmodel.ModuleComparison,
		viewmodel.ModuleEDA,
		viewmodel.ModuleDocs,
	} {
		if !model.SelectModule(id) {
			t.Fatalf("SelectModule(%s) returned false", id)
		}
		if got := model.ActivePort().Snapshot().Descriptor.ID; got != id {
			t.Fatalf("active descriptor = %s, want %s", got, id)
		}
	}

	if !model.SelectModule(viewmodel.ModuleMNIST) {
		t.Fatal("SelectModule(mnist) returned false on return")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "levels"); got != "16 levels" {
		t.Fatalf("mnist levels metric after round-trip navigation = %q, want preserved 16 levels", got)
	}
}

func TestFullAppE2EAllModulesPublishTrustBoundariesAndContent(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)

	for _, descriptor := range viewmodel.KnownDescriptors() {
		t.Run(string(descriptor.ID), func(t *testing.T) {
			if !model.SelectModule(descriptor.ID) {
				t.Fatalf("SelectModule(%s) returned false", descriptor.ID)
			}
			snapshot := model.ActivePort().Snapshot()
			if snapshot.Descriptor.ID != descriptor.ID {
				t.Fatalf("snapshot descriptor = %s, want %s", snapshot.Descriptor.ID, descriptor.ID)
			}
			if snapshot.Descriptor.BoundaryNotice == "" {
				t.Fatalf("module %s omitted boundary notice", descriptor.ID)
			}
			if len(snapshot.Metrics) == 0 {
				t.Fatalf("module %s omitted user-visible metrics", descriptor.ID)
			}
			if len(snapshot.Sections) == 0 {
				t.Fatalf("module %s omitted user-visible sections", descriptor.ID)
			}
		})
	}
}

func TestFullAppE2EStartsAndStopsAllRegisteredModules(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)
	wantDescriptors := viewmodel.KnownDescriptors()
	if len(model.Ports) != len(wantDescriptors) {
		t.Fatalf("ports = %d, want %d known descriptors", len(model.Ports), len(wantDescriptors))
	}

	model.StartAllModules()
	for i, port := range model.Ports {
		if got, want := port.Snapshot().Descriptor.ID, wantDescriptors[i].ID; got != want {
			t.Fatalf("after StartAllModules port[%d] descriptor = %s, want %s", i, got, want)
		}
	}

	model.StopAllModules()
	for i, port := range model.Ports {
		if got, want := port.Snapshot().Descriptor.ID, wantDescriptors[i].ID; got != want {
			t.Fatalf("after StopAllModules port[%d] descriptor = %s, want %s", i, got, want)
		}
	}
}

func TestFullAppE2EConfiguresDesignPipelineAndExportsArtifacts(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)
	outDir := t.TempDir()
	dispatch := func(module viewmodel.ModuleID, action viewmodel.Action) {
		t.Helper()
		if !model.SelectModule(module) {
			t.Fatalf("SelectModule(%s) returned false", module)
		}
		if err := model.DispatchAction(action); err != nil {
			t.Fatalf("DispatchAction(%s/%s): %v", module, action.ID, err)
		}
	}

	calibrationPath := filepath.Join(outDir, "level-calibration.json")
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": hysteresisvm.WaveformTriangle}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetLevelCalibrationLevelCount, Kind: viewmodel.ActionCommand, Payload: map[string]string{"level_count": "16"}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetLevelCalibrationTargetRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"target_range": "0.70"}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetLevelCalibrationTemperature, Kind: viewmodel.ActionCommand, Payload: map[string]string{"temperature_k": "350"}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventRunLevelCalibration, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventExportLevelCalibration, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": calibrationPath}})

	hysteresisSnapshot := model.ActivePort().Snapshot()
	if got := appModelE2EMetricValue(hysteresisSnapshot, "level_calibration_inputs"); got != "16 levels, target 70%, 350 K" {
		t.Fatalf("level calibration inputs = %q, want configured workflow values", got)
	}
	assertFileContains(t, calibrationPath, "\"artifact_type\": \"level_calibration\"", "SIMULATION OUTPUT")
	calibrationArtifact := readJSONArtifact[struct {
		ArtifactType   string `json:"artifact_type"`
		BoundaryNotice string `json:"boundary_notice"`
		Inputs         struct {
			LevelCount      int     `json:"level_count"`
			TargetRangeFrac float64 `json:"target_range_frac"`
			TemperatureK    float64 `json:"temperature_k"`
		} `json:"inputs"`
		Summary struct {
			AscendingEntries  int    `json:"ascending_entries"`
			DescendingEntries int    `json:"descending_entries"`
			Monotonicity      string `json:"monotonicity"`
		} `json:"summary"`
		Levels []struct {
			LevelIndex       int     `json:"level_index"`
			NormalizedTarget float64 `json:"normalized_target"`
		} `json:"levels"`
	}](t, calibrationPath)
	if calibrationArtifact.ArtifactType != "level_calibration" || !strings.Contains(calibrationArtifact.BoundaryNotice, "SIMULATION OUTPUT") {
		t.Fatalf("unexpected calibration artifact header: %+v", calibrationArtifact)
	}
	if calibrationArtifact.Inputs.LevelCount != 16 || calibrationArtifact.Inputs.TargetRangeFrac != 0.70 || calibrationArtifact.Inputs.TemperatureK != 350 {
		t.Fatalf("calibration inputs = %+v, want configured 16 levels / 70%% / 350 K", calibrationArtifact.Inputs)
	}
	if len(calibrationArtifact.Levels) != 16 || calibrationArtifact.Summary.AscendingEntries != 16 || calibrationArtifact.Summary.DescendingEntries != 16 {
		t.Fatalf("calibration level counts = levels %d summary %+v, want 16/16/16", len(calibrationArtifact.Levels), calibrationArtifact.Summary)
	}
	if calibrationArtifact.Levels[0].LevelIndex != 0 || calibrationArtifact.Levels[len(calibrationArtifact.Levels)-1].LevelIndex != 15 {
		t.Fatalf("calibration level endpoints = first %+v last %+v, want index 0..15", calibrationArtifact.Levels[0], calibrationArtifact.Levels[len(calibrationArtifact.Levels)-1])
	}

	dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "16", "cols": "32"}})
	dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand})
	crossbarSnapshot := model.ActivePort().Snapshot()
	if got := appModelE2EMetricValue(crossbarSnapshot, "rows"); got != "16" {
		t.Fatalf("crossbar rows = %q, want 16", got)
	}
	if got := appModelE2EMetricValue(crossbarSnapshot, "cols"); got != "32" {
		t.Fatalf("crossbar cols = %q, want 32", got)
	}
	if len(crossbarSnapshot.Plots) == 0 {
		t.Fatal("crossbar workflow did not publish plots after MVM")
	}

	operationLogPath := filepath.Join(outDir, "operation-log.json")
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionResizeArray, Kind: viewmodel.ActionCommand, Payload: map[string]string{"size": "16"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetArchitecture, Kind: viewmodel.ActionSelect, Payload: map[string]string{"architecture": circuitsvm.Architecture1T1R}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetDACBits, Kind: viewmodel.ActionCommand, Payload: map[string]string{"bits": "5"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetADCBits, Kind: viewmodel.ActionCommand, Payload: map[string]string{"bits": "6"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionRunCompute, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionExportOperationLog, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": operationLogPath}})
	assertFileContains(t, operationLogPath, "\"schema\": \"fecim.circuits.operation_log.v1\"", "COMPUTE on 16x16 1T1R")
	operationLog := readJSONArtifact[struct {
		Schema              string `json:"schema"`
		OperationMode       string `json:"operation_mode"`
		Architecture        string `json:"architecture"`
		Rows                int    `json:"rows"`
		Cols                int    `json:"cols"`
		ExportedEntries     int    `json:"exported_entries"`
		LastOperationStatus string `json:"last_operation_status"`
		ComputeRun          struct {
			Schema        string `json:"schema"`
			ArraySize     string `json:"array_size"`
			ExportedCells int    `json:"exported_cells"`
			RowResults    []struct {
				Row       int  `json:"row"`
				Active    bool `json:"active"`
				Saturated bool `json:"saturated"`
			} `json:"row_results"`
		} `json:"compute_run"`
	}](t, operationLogPath)
	if operationLog.Schema != "fecim.circuits.operation_log.v1" || operationLog.ComputeRun.Schema != "fecim.circuits.compute_run.v1" {
		t.Fatalf("unexpected operation log schemas: %+v", operationLog)
	}
	if operationLog.OperationMode != circuitsvm.OperationCompute || operationLog.Architecture != circuitsvm.Architecture1T1R || operationLog.Rows != 16 || operationLog.Cols != 16 {
		t.Fatalf("operation log state = mode %q arch %q %dx%d, want compute 1T1R 16x16", operationLog.OperationMode, operationLog.Architecture, operationLog.Rows, operationLog.Cols)
	}
	if operationLog.ExportedEntries < 5 || operationLog.ComputeRun.ArraySize != "16x16" || operationLog.ComputeRun.ExportedCells != 256 || len(operationLog.ComputeRun.RowResults) != 16 {
		t.Fatalf("operation log compute payload incomplete: entries=%d compute=%+v", operationLog.ExportedEntries, operationLog.ComputeRun)
	}

	dispatch(viewmodel.ModuleMNIST, viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "64"}})
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "levels"); got != "64 levels" {
		t.Fatalf("mnist levels metric = %q, want 64 levels", got)
	}
	dispatch(viewmodel.ModuleDocs, viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand, Payload: map[string]string{"query": "IR drop validation"}})
	dispatch(viewmodel.ModuleDocs, viewmodel.Action{ID: "start_curriculum", Kind: viewmodel.ActionCommand})
	if got := appModelE2ESectionBody(model.ActivePort().Snapshot(), "search_results"); !strings.Contains(got, "IR drop validation") {
		t.Fatalf("docs search section = %q, want query echoed", got)
	}

	if !model.SelectModule(viewmodel.ModuleHysteresis) {
		t.Fatal("SelectModule(hysteresis) returned false on persistence check")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "level_calibration_inputs"); got != "16 levels, target 70%, 350 K" {
		t.Fatalf("hysteresis calibration lost after multi-module workflow: %q", got)
	}
	if !model.SelectModule(viewmodel.ModuleMNIST) {
		t.Fatal("SelectModule(mnist) returned false on persistence check")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "levels"); got != "64 levels" {
		t.Fatalf("mnist sweep lost after multi-module workflow: %q", got)
	}

	composition := design.Composition{
		Hysteresis: appModelE2EPortByID(t, model, viewmodel.ModuleHysteresis),
		Crossbar:   appModelE2EPortByID(t, model, viewmodel.ModuleCrossbar),
		Circuits:   appModelE2EPortByID(t, model, viewmodel.ModuleCircuits),
		EDA:        appModelE2EPortByID(t, model, viewmodel.ModuleEDA),
	}
	designSnapshot := composition.Snapshot()
	if designSnapshot.ArrayRows != 16 || designSnapshot.ArrayCols != 32 {
		t.Fatalf("design array = %dx%d, want 16x32", designSnapshot.ArrayRows, designSnapshot.ArrayCols)
	}
	if designSnapshot.ADCResolution != 6 || designSnapshot.DACResolution != 5 {
		t.Fatalf("design converter resolution = ADC %d DAC %d, want ADC 6 DAC 5", designSnapshot.ADCResolution, designSnapshot.DACResolution)
	}
	for _, want := range []string{"fecim_crossbar_8x8", "16×32", "6-bit ADC/5-bit DAC", "sky130"} {
		if !strings.Contains(designSnapshot.Summary, want) {
			t.Fatalf("design summary %q missing %q", designSnapshot.Summary, want)
		}
	}
	if err := composition.ExportDesign(); err != nil {
		t.Fatalf("ExportDesign: %v", err)
	}
}

func TestFullAppE2EWideModuleActionAndExportMatrix(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)
	outDir := t.TempDir()
	dispatch := func(module viewmodel.ModuleID, action viewmodel.Action) viewmodel.ModuleSnapshot {
		t.Helper()
		if !model.SelectModule(module) {
			t.Fatalf("SelectModule(%s) returned false", module)
		}
		if err := model.DispatchAction(action); err != nil {
			t.Fatalf("DispatchAction(%s/%s): %v", module, action.ID, err)
		}
		snapshot := model.ActivePort().Snapshot()
		if snapshot.Descriptor.ID != module {
			t.Fatalf("after %s/%s active module = %s", module, action.ID, snapshot.Descriptor.ID)
		}
		if snapshot.Descriptor.BoundaryNotice == "" {
			t.Fatalf("after %s/%s boundary notice empty", module, action.ID)
		}
		return snapshot
	}

	hysteresisCSV := filepath.Join(outDir, "hysteresis-loop.csv")
	pundCSV := filepath.Join(outDir, "pund.csv")
	forcSweep := filepath.Join(outDir, "forc-sweep.csv")
	forcMatrix := filepath.Join(outDir, "forc-matrix.csv")
	forcMeta := filepath.Join(outDir, "forc-meta.json")
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetFieldRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"min": "-2400", "max": "2400"}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": hysteresisvm.WaveformSquare}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventExportCSV, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": hysteresisCSV}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventRunPUND, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventExportPUNDCSV, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": pundCSV}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventRunFORC, Kind: viewmodel.ActionCommand, Payload: map[string]string{"reversals": "7"}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventExportFORCSweep, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": forcSweep, "reversals": "7"}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventExportFORCMatrix, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": forcMatrix, "reversals": "7"}})
	hysteresisSnapshot := dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventExportFORCMeta, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": forcMeta, "reversals": "7"}})
	if got := appModelE2EMetricValue(hysteresisSnapshot, "waveform"); got != hysteresisvm.WaveformSquare {
		t.Fatalf("hysteresis waveform = %q, want square", got)
	}
	assertFileContains(t, hysteresisCSV, "Index,E_field_kV_cm,Polarization_uC_cm2")
	assertFileContains(t, pundCSV, "metric,value,unit", "Qsw_positive")
	assertFileContains(t, forcSweep, "reversal_field_vm,applied_field_vm,polarization_cm2")
	assertFileContains(t, forcMatrix, "Ea_Vm,Eb_Vm,density")
	forcMetadata := readJSONArtifact[struct {
		Material string `json:"material"`
		Waveform string `json:"waveform"`
		Curves   int    `json:"curves"`
		Boundary string `json:"boundary"`
	}](t, forcMeta)
	if forcMetadata.Waveform != hysteresisvm.WaveformSquare || forcMetadata.Curves != 7 || !strings.Contains(forcMetadata.Boundary, "SIMULATION OUTPUT") {
		t.Fatalf("FORC metadata = %+v, want square waveform, 7 curves, simulation boundary", forcMetadata)
	}

	mnistSnapshot := dispatch(viewmodel.ModuleMNIST, viewmodel.Action{ID: "run_inference", Kind: viewmodel.ActionCommand})
	if got := appModelE2EMetricValue(mnistSnapshot, "accuracy"); got == "" {
		t.Fatal("mnist inference snapshot omitted accuracy metric")
	}
	for _, levels := range []string{"8", "16", "32", "64"} {
		mnistSnapshot = dispatch(viewmodel.ModuleMNIST, viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": levels}})
		if got := appModelE2EMetricValue(mnistSnapshot, "levels"); got != levels+" levels" {
			t.Fatalf("mnist levels after sweep %s = %q", levels, got)
		}
	}
	if len(mnistSnapshot.Plots) == 0 {
		t.Fatal("mnist sweep omitted accuracy plot")
	}

	specPath := filepath.Join(outDir, "reference-specs.json")
	timingPath := filepath.Join(outDir, "reference-timing.json")
	svgPath := filepath.Join(outDir, "reference-timing.svg")
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetOperationMode, Kind: viewmodel.ActionSelect, Payload: map[string]string{"mode": circuitsvm.OperationCompute}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetTimingOperation, Kind: viewmodel.ActionSelect, Payload: map[string]string{"operation": "COMPUTE"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionPlayReferenceTiming, Kind: viewmodel.ActionCommand, Payload: map[string]string{"interval_ms": "250"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionStepReferenceTiming, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionPauseReferenceTiming, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionExportReferenceSpecs, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": specPath}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionExportReferenceTiming, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": timingPath}})
	circuitsSnapshot := dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionExportReferenceTimingSVG, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": svgPath}})
	if got := appModelE2EMetricValue(circuitsSnapshot, "timing_operation"); got != "COMPUTE" {
		t.Fatalf("circuits timing operation = %q, want COMPUTE", got)
	}
	assertFileContains(t, specPath, "\"schema\": \"fecim.circuits.reference_specs.v1\"", "educational")
	assertFileContains(t, timingPath, "\"schema\": \"fecim.circuits.reference_timing.v1\"", "COMPUTE")
	assertFileContains(t, svgPath, `<title>COMPUTE Timing Waveform</title>`, "educational reference timing waveform")

	for _, query := range []string{"ISPP", "FORC density", "OpenLane export"} {
		docsSnapshot := dispatch(viewmodel.ModuleDocs, viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand, Payload: map[string]string{"query": query}})
		if got := appModelE2EMetricValue(docsSnapshot, "search_query"); got != query {
			t.Fatalf("docs search query metric = %q, want %q", got, query)
		}
		if body := appModelE2ESectionBody(docsSnapshot, "search_results"); !strings.Contains(body, query) {
			t.Fatalf("docs search results for %q = %q", query, body)
		}
	}

	if !model.SelectModule(viewmodel.ModuleComparison) {
		t.Fatal("SelectModule(comparison) returned false")
	}
	if err := model.DispatchAction(viewmodel.Action{ID: "unsupported_wide_e2e", Kind: viewmodel.ActionCommand}); !errors.Is(err, viewmodel.ErrUnsupportedAction) {
		t.Fatalf("comparison unsupported action error = %v, want ErrUnsupportedAction", err)
	}
	if !model.SelectModule(viewmodel.ModuleHysteresis) {
		t.Fatal("SelectModule(hysteresis) returned false on final state check")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "waveform"); got != hysteresisvm.WaveformSquare {
		t.Fatalf("hysteresis state was not preserved after wide workflow: waveform=%q", got)
	}
}

func TestFullAppE2EWideInvalidActionMatrixPreservesState(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)

	if !model.SelectModule(viewmodel.ModuleHysteresis) {
		t.Fatal("SelectModule(hysteresis) returned false")
	}
	if err := model.DispatchAction(viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": hysteresisvm.WaveformTriangle}}); err != nil {
		t.Fatalf("seed hysteresis waveform: %v", err)
	}
	if !model.SelectModule(viewmodel.ModuleCrossbar) {
		t.Fatal("SelectModule(crossbar) returned false")
	}
	if err := model.DispatchAction(viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "12", "cols": "10"}}); err != nil {
		t.Fatalf("seed crossbar resize: %v", err)
	}
	if !model.SelectModule(viewmodel.ModuleMNIST) {
		t.Fatal("SelectModule(mnist) returned false")
	}
	if err := model.DispatchAction(viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "32"}}); err != nil {
		t.Fatalf("seed mnist levels: %v", err)
	}

	cases := []struct {
		name              string
		module            viewmodel.ModuleID
		action            viewmodel.Action
		wantUnsupported   bool
		wantErrSubstrings []string
		preservedMetricID string
		preservedValue    string
	}{
		{
			name:              "hysteresis rejects unknown waveform",
			module:            viewmodel.ModuleHysteresis,
			action:            viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": "sawtooth"}},
			wantErrSubstrings: []string{"unsupported waveform"},
			preservedMetricID: "waveform",
			preservedValue:    hysteresisvm.WaveformTriangle,
		},
		{
			name:              "crossbar rejects invalid resize",
			module:            viewmodel.ModuleCrossbar,
			action:            viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "0", "cols": "256"}},
			wantErrSubstrings: []string{"invalid dimensions"},
			preservedMetricID: "rows",
			preservedValue:    "12",
		},
		{
			name:              "mnist rejects nonnumeric level sweep",
			module:            viewmodel.ModuleMNIST,
			action:            viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "many"}},
			wantErrSubstrings: []string{"levels"},
			preservedMetricID: "levels",
			preservedValue:    "32 levels",
		},
		{
			name:              "circuits rejects path traversal export",
			module:            viewmodel.ModuleCircuits,
			action:            viewmodel.Action{ID: circuitsvm.ActionExportReferenceSpecs, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": "../escape.json"}},
			wantErrSubstrings: []string{"path traversal"},
			preservedMetricID: "reference_spec_export_path",
			preservedValue:    "none",
		},
		{
			name:              "docs rejects search without query",
			module:            viewmodel.ModuleDocs,
			action:            viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand},
			wantUnsupported:   true,
			preservedMetricID: "active_page",
			preservedValue:    "overview",
		},
		{
			name:              "eda rejects unknown export command",
			module:            viewmodel.ModuleEDA,
			action:            viewmodel.Action{ID: "export_magic", Kind: viewmodel.ActionCommand},
			wantUnsupported:   true,
			preservedMetricID: "process",
			preservedValue:    "sky130",
		},
		{
			name:              "comparison remains read-only",
			module:            viewmodel.ModuleComparison,
			action:            viewmodel.Action{ID: "mutate_architecture", Kind: viewmodel.ActionCommand},
			wantUnsupported:   true,
			preservedMetricID: "count",
			preservedValue:    "3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !model.SelectModule(tc.module) {
				t.Fatalf("SelectModule(%s) returned false", tc.module)
			}
			before := model.ActivePort().Snapshot()
			err := model.DispatchAction(tc.action)
			if err == nil {
				t.Fatalf("DispatchAction(%s/%s) returned nil, want rejection", tc.module, tc.action.ID)
			}
			if tc.wantUnsupported && !errors.Is(err, viewmodel.ErrUnsupportedAction) {
				t.Fatalf("DispatchAction(%s/%s) error = %v, want ErrUnsupportedAction", tc.module, tc.action.ID, err)
			}
			for _, want := range tc.wantErrSubstrings {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("DispatchAction(%s/%s) error = %v, want substring %q", tc.module, tc.action.ID, err, want)
				}
			}
			after := model.ActivePort().Snapshot()
			if after.Descriptor.ID != before.Descriptor.ID || after.Descriptor.ID != tc.module {
				t.Fatalf("active module changed from %s to %s after rejected action", before.Descriptor.ID, after.Descriptor.ID)
			}
			if got := appModelE2EMetricValue(after, tc.preservedMetricID); got != tc.preservedValue {
				t.Fatalf("metric %s after rejected action = %q, want preserved %q", tc.preservedMetricID, got, tc.preservedValue)
			}
			if after.Descriptor.BoundaryNotice == "" {
				t.Fatalf("module %s lost boundary notice after rejected action", tc.module)
			}
		})
	}
}

func TestFullAppE2EWideActionStormPreservesSnapshotContracts(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)
	dispatch := func(module viewmodel.ModuleID, action viewmodel.Action) {
		t.Helper()
		if !model.SelectModule(module) {
			t.Fatalf("SelectModule(%s) returned false", module)
		}
		if err := model.DispatchAction(action); err != nil {
			t.Fatalf("DispatchAction(%s/%s): %v", module, action.ID, err)
		}
	}

	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": hysteresisvm.WaveformManual}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetFieldRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"min": "-1800", "max": "2600"}})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventRunPUND, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventRunFORC, Kind: viewmodel.ActionCommand, Payload: map[string]string{"reversals": "5"}})

	dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "6", "cols": "10"}})
	dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "toggle_ir", Kind: viewmodel.ActionToggle})
	dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "toggle_ir", Kind: viewmodel.ActionToggle})
	dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand})

	for _, levels := range []string{"2", "8", "64"} {
		dispatch(viewmodel.ModuleMNIST, viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": levels}})
	}
	dispatch(viewmodel.ModuleMNIST, viewmodel.Action{ID: "run_inference", Kind: viewmodel.ActionCommand})

	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionResizeArray, Kind: viewmodel.ActionCommand, Payload: map[string]string{"size": "4"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetArchitecture, Kind: viewmodel.ActionSelect, Payload: map[string]string{"architecture": circuitsvm.Architecture2T1R}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSelectCell, Kind: viewmodel.ActionCommand, Payload: map[string]string{"row": "3", "col": "2"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetWriteTarget, Kind: viewmodel.ActionCommand, Payload: map[string]string{"level": "21"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetDACBits, Kind: viewmodel.ActionCommand, Payload: map[string]string{"bits": "6"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetADCBits, Kind: viewmodel.ActionCommand, Payload: map[string]string{"bits": "6"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionRunRead, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionRunWrite, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionRunCompute, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetTimingOperation, Kind: viewmodel.ActionSelect, Payload: map[string]string{"operation": "WRITE"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionAnimateReferenceTiming, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionPlayReferenceTiming, Kind: viewmodel.ActionCommand, Payload: map[string]string{"interval_ms": "100"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionStepReferenceTiming, Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionResetReferenceTiming, Kind: viewmodel.ActionCommand})

	dispatch(viewmodel.ModuleEDA, viewmodel.Action{ID: "generate_spice", Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleEDA, viewmodel.Action{ID: "generate_all", Kind: viewmodel.ActionCommand})
	dispatch(viewmodel.ModuleDocs, viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand, Payload: map[string]string{"query": "write verify EDA"}})
	dispatch(viewmodel.ModuleDocs, viewmodel.Action{ID: "start_curriculum", Kind: viewmodel.ActionCommand})

	if !model.SelectModule(viewmodel.ModuleComparison) {
		t.Fatal("SelectModule(comparison) returned false")
	}
	if err := model.DispatchAction(viewmodel.Action{ID: "action_storm_mutation", Kind: viewmodel.ActionCommand}); !errors.Is(err, viewmodel.ErrUnsupportedAction) {
		t.Fatalf("comparison mutation error = %v, want ErrUnsupportedAction", err)
	}

	for _, descriptor := range viewmodel.KnownDescriptors() {
		if !model.SelectModule(descriptor.ID) {
			t.Fatalf("SelectModule(%s) returned false during contract audit", descriptor.ID)
		}
		assertSnapshotContractAfterActionStorm(t, model.ActivePort().Snapshot())
	}

	if !model.SelectModule(viewmodel.ModuleHysteresis) {
		t.Fatal("SelectModule(hysteresis) returned false on storm persistence check")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "waveform"); got != hysteresisvm.WaveformManual {
		t.Fatalf("hysteresis waveform after action storm = %q, want manual", got)
	}
	if !model.SelectModule(viewmodel.ModuleCrossbar) {
		t.Fatal("SelectModule(crossbar) returned false on storm persistence check")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "rows"); got != "6" {
		t.Fatalf("crossbar rows after action storm = %q, want 6", got)
	}
	if !model.SelectModule(viewmodel.ModuleCircuits) {
		t.Fatal("SelectModule(circuits) returned false on storm persistence check")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "selected_cell"); got != "[3,2]" {
		t.Fatalf("circuits selected cell after action storm = %q, want [3,2]", got)
	}
}

func TestFullAppE2EWideNavigationPermutationStateIsolation(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleDocs)
	dispatch := func(module viewmodel.ModuleID, action viewmodel.Action) {
		t.Helper()
		if !model.SelectModule(module) {
			t.Fatalf("SelectModule(%s) returned false", module)
		}
		if err := model.DispatchAction(action); err != nil {
			t.Fatalf("DispatchAction(%s/%s): %v", module, action.ID, err)
		}
	}

	dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": hysteresisvm.WaveformSquare}})
	dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "20", "cols": "12"}})
	dispatch(viewmodel.ModuleMNIST, viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "16"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSelectCell, Kind: viewmodel.ActionCommand, Payload: map[string]string{"row": "5", "col": "4"}})
	dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSetOperationMode, Kind: viewmodel.ActionSelect, Payload: map[string]string{"mode": circuitsvm.OperationWrite}})
	dispatch(viewmodel.ModuleDocs, viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand, Payload: map[string]string{"query": "navigation isolation"}})
	dispatch(viewmodel.ModuleEDA, viewmodel.Action{ID: "generate_all", Kind: viewmodel.ActionCommand})

	model.StartAllModules()
	model.StopAllModules()

	path := []viewmodel.ModuleID{
		viewmodel.ModuleHysteresis,
		viewmodel.ModuleCrossbar,
		viewmodel.ModuleMNIST,
		viewmodel.ModuleCircuits,
		viewmodel.ModuleComparison,
		viewmodel.ModuleEDA,
		viewmodel.ModuleDocs,
		viewmodel.ModuleEDA,
		viewmodel.ModuleComparison,
		viewmodel.ModuleCircuits,
		viewmodel.ModuleMNIST,
		viewmodel.ModuleCrossbar,
		viewmodel.ModuleHysteresis,
		viewmodel.ModuleDocs,
	}
	for step, id := range path {
		if !model.SelectModule(id) {
			t.Fatalf("navigation step %d SelectModule(%s) returned false", step, id)
		}
		snapshot := model.ActivePort().Snapshot()
		if snapshot.Descriptor.ID != id {
			t.Fatalf("navigation step %d active descriptor = %s, want %s", step, snapshot.Descriptor.ID, id)
		}
		assertSnapshotContractAfterActionStorm(t, snapshot)
	}

	activeBeforeMissing := model.ActivePort().Snapshot().Descriptor.ID
	if model.SelectModule(viewmodel.ModuleID("missing-module")) {
		t.Fatal("SelectModule(missing-module) returned true")
	}
	if got := model.ActivePort().Snapshot().Descriptor.ID; got != activeBeforeMissing {
		t.Fatalf("missing module selection changed active module from %s to %s", activeBeforeMissing, got)
	}

	checks := []struct {
		module viewmodel.ModuleID
		metric string
		want   string
	}{
		{viewmodel.ModuleHysteresis, "waveform", hysteresisvm.WaveformSquare},
		{viewmodel.ModuleCrossbar, "rows", "20"},
		{viewmodel.ModuleCrossbar, "cols", "12"},
		{viewmodel.ModuleMNIST, "levels", "16 levels"},
		{viewmodel.ModuleCircuits, "selected_cell", "[5,4]"},
		{viewmodel.ModuleCircuits, "mode", "WRITE"},
		{viewmodel.ModuleDocs, "search_query", "navigation isolation"},
		{viewmodel.ModuleEDA, "process", "sky130"},
		{viewmodel.ModuleComparison, "count", "3"},
	}
	for _, check := range checks {
		if !model.SelectModule(check.module) {
			t.Fatalf("SelectModule(%s) returned false for final state check", check.module)
		}
		if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), check.metric); got != check.want {
			t.Fatalf("%s metric %s after navigation permutation = %q, want %q", check.module, check.metric, got, check.want)
		}
	}
}

func TestFullAppE2EWideIndependentSessionsDoNotShareState(t *testing.T) {
	alpha := NewAppModel(viewmodel.ModuleHysteresis)
	beta := NewAppModel(viewmodel.ModuleDocs)
	dispatch := func(name string, model *AppModel, module viewmodel.ModuleID, action viewmodel.Action) {
		t.Helper()
		if !model.SelectModule(module) {
			t.Fatalf("%s SelectModule(%s) returned false", name, module)
		}
		if err := model.DispatchAction(action); err != nil {
			t.Fatalf("%s DispatchAction(%s/%s): %v", name, module, action.ID, err)
		}
	}
	metric := func(name string, model *AppModel, module viewmodel.ModuleID, id string) string {
		t.Helper()
		if !model.SelectModule(module) {
			t.Fatalf("%s SelectModule(%s) returned false", name, module)
		}
		return appModelE2EMetricValue(model.ActivePort().Snapshot(), id)
	}

	dispatch("alpha", &alpha, viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": hysteresisvm.WaveformTriangle}})
	dispatch("alpha", &alpha, viewmodel.ModuleCrossbar, viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "9", "cols": "11"}})
	dispatch("alpha", &alpha, viewmodel.ModuleMNIST, viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "8"}})
	dispatch("alpha", &alpha, viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSelectCell, Kind: viewmodel.ActionCommand, Payload: map[string]string{"row": "2", "col": "1"}})
	dispatch("alpha", &alpha, viewmodel.ModuleDocs, viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand, Payload: map[string]string{"query": "alpha session"}})

	dispatch("beta", &beta, viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": hysteresisvm.WaveformSquare}})
	dispatch("beta", &beta, viewmodel.ModuleCrossbar, viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": "13", "cols": "7"}})
	dispatch("beta", &beta, viewmodel.ModuleMNIST, viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "64"}})
	dispatch("beta", &beta, viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSelectCell, Kind: viewmodel.ActionCommand, Payload: map[string]string{"row": "6", "col": "5"}})
	dispatch("beta", &beta, viewmodel.ModuleDocs, viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand, Payload: map[string]string{"query": "beta session"}})

	alpha.StartAllModules()
	beta.StartAllModules()
	alpha.StopAllModules()
	beta.StopAllModules()

	checks := []struct {
		name   string
		model  *AppModel
		module viewmodel.ModuleID
		metric string
		want   string
	}{
		{"alpha", &alpha, viewmodel.ModuleHysteresis, "waveform", hysteresisvm.WaveformTriangle},
		{"beta", &beta, viewmodel.ModuleHysteresis, "waveform", hysteresisvm.WaveformSquare},
		{"alpha", &alpha, viewmodel.ModuleCrossbar, "rows", "9"},
		{"alpha", &alpha, viewmodel.ModuleCrossbar, "cols", "11"},
		{"beta", &beta, viewmodel.ModuleCrossbar, "rows", "13"},
		{"beta", &beta, viewmodel.ModuleCrossbar, "cols", "7"},
		{"alpha", &alpha, viewmodel.ModuleMNIST, "levels", "8 levels"},
		{"beta", &beta, viewmodel.ModuleMNIST, "levels", "64 levels"},
		{"alpha", &alpha, viewmodel.ModuleCircuits, "selected_cell", "[2,1]"},
		{"beta", &beta, viewmodel.ModuleCircuits, "selected_cell", "[6,5]"},
		{"alpha", &alpha, viewmodel.ModuleDocs, "search_query", "alpha session"},
		{"beta", &beta, viewmodel.ModuleDocs, "search_query", "beta session"},
	}
	for _, check := range checks {
		if got := metric(check.name, check.model, check.module, check.metric); got != check.want {
			t.Fatalf("%s %s metric %s = %q, want %q", check.name, check.module, check.metric, got, check.want)
		}
	}

	for _, model := range []*AppModel{&alpha, &beta} {
		for _, descriptor := range viewmodel.KnownDescriptors() {
			if !model.SelectModule(descriptor.ID) {
				t.Fatalf("SelectModule(%s) returned false", descriptor.ID)
			}
			assertSnapshotContractAfterActionStorm(t, model.ActivePort().Snapshot())
		}
	}
}

func TestFullAppE2EWideParallelIndependentWorkflowArtifacts(t *testing.T) {
	cases := []struct {
		name     string
		waveform string
		rows     string
		cols     string
		levels   string
		cellRow  string
		cellCol  string
		query    string
	}{
		{name: "small-triangle", waveform: hysteresisvm.WaveformTriangle, rows: "5", cols: "7", levels: "8", cellRow: "1", cellCol: "2", query: "parallel small workflow"},
		{name: "medium-square", waveform: hysteresisvm.WaveformSquare, rows: "10", cols: "6", levels: "16", cellRow: "3", cellCol: "4", query: "parallel medium workflow"},
		{name: "wide-manual", waveform: hysteresisvm.WaveformManual, rows: "12", cols: "9", levels: "64", cellRow: "5", cellCol: "6", query: "parallel wide workflow"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			model := NewAppModel(viewmodel.ModuleHysteresis)
			outDir := t.TempDir()
			dispatch := func(module viewmodel.ModuleID, action viewmodel.Action) {
				t.Helper()
				if !model.SelectModule(module) {
					t.Fatalf("SelectModule(%s) returned false", module)
				}
				if err := model.DispatchAction(action); err != nil {
					t.Fatalf("DispatchAction(%s/%s): %v", module, action.ID, err)
				}
			}

			hysteresisCSV := filepath.Join(outDir, "loop.csv")
			operationLog := filepath.Join(outDir, "operation-log.json")
			dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": tc.waveform}})
			dispatch(viewmodel.ModuleHysteresis, viewmodel.Action{ID: hysteresisvm.EventExportCSV, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": hysteresisCSV}})
			dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "resize", Kind: viewmodel.ActionCommand, Payload: map[string]string{"rows": tc.rows, "cols": tc.cols}})
			dispatch(viewmodel.ModuleCrossbar, viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand})
			dispatch(viewmodel.ModuleMNIST, viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": tc.levels}})
			dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionSelectCell, Kind: viewmodel.ActionCommand, Payload: map[string]string{"row": tc.cellRow, "col": tc.cellCol}})
			dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionRunCompute, Kind: viewmodel.ActionCommand})
			dispatch(viewmodel.ModuleCircuits, viewmodel.Action{ID: circuitsvm.ActionExportOperationLog, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": operationLog}})
			dispatch(viewmodel.ModuleDocs, viewmodel.Action{ID: "search", Kind: viewmodel.ActionCommand, Payload: map[string]string{"query": tc.query}})

			assertFileContains(t, hysteresisCSV, "Index,E_field_kV_cm,Polarization_uC_cm2")
			assertFileContains(t, operationLog, "\"schema\": \"fecim.circuits.operation_log.v1\"", "COMPUTE")
			operationArtifact := readJSONArtifact[struct {
				OperationMode string `json:"operation_mode"`
				ComputeRun    struct {
					ExportedCells int `json:"exported_cells"`
				} `json:"compute_run"`
			}](t, operationLog)
			if operationArtifact.OperationMode != circuitsvm.OperationCompute || operationArtifact.ComputeRun.ExportedCells == 0 {
				t.Fatalf("operation artifact = %+v, want compute artifact with exported cells", operationArtifact)
			}

			checks := []struct {
				module viewmodel.ModuleID
				metric string
				want   string
			}{
				{viewmodel.ModuleHysteresis, "waveform", tc.waveform},
				{viewmodel.ModuleCrossbar, "rows", tc.rows},
				{viewmodel.ModuleCrossbar, "cols", tc.cols},
				{viewmodel.ModuleMNIST, "levels", tc.levels + " levels"},
				{viewmodel.ModuleCircuits, "selected_cell", "[" + tc.cellRow + "," + tc.cellCol + "]"},
				{viewmodel.ModuleDocs, "search_query", tc.query},
			}
			for _, check := range checks {
				if !model.SelectModule(check.module) {
					t.Fatalf("SelectModule(%s) returned false", check.module)
				}
				if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), check.metric); got != check.want {
					t.Fatalf("%s metric %s = %q, want %q", check.module, check.metric, got, check.want)
				}
				assertSnapshotContractAfterActionStorm(t, model.ActivePort().Snapshot())
			}
		})
	}
}

func TestFullAppE2EModule6ExportSurfaceMatchesCompilerArtifacts(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleEDA)
	if !model.SelectModule(viewmodel.ModuleEDA) {
		t.Fatal("SelectModule(eda) returned false")
	}
	for _, action := range []viewmodel.Action{
		{ID: "generate_spice", Kind: viewmodel.ActionCommand},
		{ID: "generate_all", Kind: viewmodel.ActionCommand},
	} {
		if err := model.DispatchAction(action); err != nil {
			t.Fatalf("DispatchAction(eda/%s): %v", action.ID, err)
		}
	}

	snapshot := model.ActivePort().Snapshot()
	assertSnapshotContractAfterActionStorm(t, snapshot)
	if snapshot.Descriptor.ID != viewmodel.ModuleEDA {
		t.Fatalf("active descriptor = %s, want EDA", snapshot.Descriptor.ID)
	}
	if !strings.Contains(snapshot.Descriptor.BoundaryNotice, "EDUCATIONAL EDA") || !strings.Contains(snapshot.Descriptor.BoundaryNotice, "Not a production chip design tool") {
		t.Fatalf("EDA boundary notice = %q, want educational/non-production warning", snapshot.Descriptor.BoundaryNotice)
	}

	cfg := compiler.NewComputeConfig(8, 8)
	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign(8x8 compute): %v", err)
	}
	wantMetrics := map[string]string{
		"design":  "fecim_crossbar_8x8",
		"process": "sky130",
		"array":   "8×8",
		"cells":   fmt.Sprintf("%d", design.Stats.TotalCells),
		"area":    fmt.Sprintf("%.3f mm²", design.Stats.AreaMM2),
		"power":   fmt.Sprintf("%.1f mW", design.Stats.PowerMW),
		"spice":   "ready",
		"verilog": "ready",
		"liberty": "ready",
		"def":     "ready",
		"lef":     "ready",
	}
	for id, want := range wantMetrics {
		if got := appModelE2EMetricValue(snapshot, id); got != want {
			t.Fatalf("EDA metric %s = %q, want %q", id, got, want)
		}
	}
	for _, actionID := range []string{"generate_spice", "generate_all"} {
		if !snapshotHasAction(snapshot, actionID) {
			t.Fatalf("EDA snapshot missing action %s", actionID)
		}
	}

	spice := edaexport.GenerateSPICE(design, 1.8)
	verilog := edaexport.GenerateVerilogWithDefaults(design)
	def := edaexport.GenerateDEFWithDefaults(design)
	lef := edaexport.GenerateLEF(edaconfig.DefaultCellConfig())
	liberty := edaexport.GenerateLiberty(edaconfig.DefaultCellConfig())
	openlaneConfig := edaexport.GenerateOpenLaneConfig(edaconfig.ArrayConfig{Rows: 8, Cols: 8, Mode: "compute", Architecture: "passive", Technology: "sky130", CellWidth: 0.46, CellHeight: 2.72})

	sections := map[string]string{
		"spice_content":   appModelE2ESectionBody(snapshot, "spice_content"),
		"verilog_content": appModelE2ESectionBody(snapshot, "verilog_content"),
		"def_content":     appModelE2ESectionBody(snapshot, "def_content"),
		"lef_content":     appModelE2ESectionBody(snapshot, "lef_content"),
	}
	for id, body := range sections {
		if body == "" {
			t.Fatalf("EDA section %s is empty", id)
		}
		if got := len(strings.Split(strings.TrimSpace(body), "\n")); got > 16 {
			t.Fatalf("EDA section %s has %d lines, want truncated preview", id, got)
		}
	}
	for _, check := range []struct {
		name    string
		section string
		full    string
		markers []string
	}{
		{name: "SPICE", section: sections["spice_content"], full: spice, markers: []string{"FeCIM", ".subckt"}},
		{name: "Verilog", section: sections["verilog_content"], full: verilog, markers: []string{"module", "input"}},
		{name: "DEF", section: sections["def_content"], full: def, markers: []string{"VERSION", "DESIGN"}},
		{name: "LEF", section: sections["lef_content"], full: lef, markers: []string{"MACRO", "PIN"}},
	} {
		if strings.TrimSpace(check.full) == "" {
			t.Fatalf("generated %s artifact is empty", check.name)
		}
		firstLine := strings.Split(strings.TrimSpace(check.section), "\n")[0]
		if !strings.Contains(check.full, firstLine) {
			t.Fatalf("EDA %s preview first line %q not found in generated artifact", check.name, firstLine)
		}
		for _, marker := range check.markers {
			if !strings.Contains(check.full, marker) {
				t.Fatalf("generated %s artifact missing marker %q", check.name, marker)
			}
		}
	}
	for _, marker := range []string{"library", "fecim_bitcell", "cell_rise", "cell_fall"} {
		if !strings.Contains(liberty, marker) {
			t.Fatalf("generated Liberty missing marker %q", marker)
		}
	}
	var openlane map[string]any
	if err := json.Unmarshal([]byte(openlaneConfig), &openlane); err != nil {
		t.Fatalf("OpenLane config JSON invalid: %v\n%s", err, openlaneConfig)
	}
	for _, key := range []string{"DESIGN_NAME", "VERILOG_FILES", "VERILOG_FILES_BLACKBOX", "FP_DEF_TEMPLATE", "EXTRA_LEFS", "EXTRA_LIBS", "DIE_AREA"} {
		if _, ok := openlane[key]; !ok {
			t.Fatalf("OpenLane config missing %s: %s", key, openlaneConfig)
		}
	}
	for _, sectionID := range []string{"edu_spice", "edu_flow", "research_validation", "design_workflow"} {
		body := appModelE2ESectionBody(snapshot, sectionID)
		if body == "" {
			t.Fatalf("EDA educational/research section %s is empty", sectionID)
		}
		if sectionID != "design_workflow" && !strings.Contains(body, "SPICE") && !strings.Contains(body, "OpenLane") && !strings.Contains(body, "validated") {
			t.Fatalf("EDA section %s body lacks expected technical context: %q", sectionID, body)
		}
	}
}

func snapshotHasAction(snapshot viewmodel.ModuleSnapshot, id string) bool {
	for _, action := range snapshot.Actions {
		if action.ID == id {
			return true
		}
	}
	return false
}

func appModelE2EMetricValue(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range snapshot.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}

func appModelE2ESectionBody(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, section := range snapshot.Sections {
		if section.ID == id {
			return section.Body
		}
	}
	return ""
}

func appModelE2EPortByID(t *testing.T, model AppModel, id viewmodel.ModuleID) viewmodel.ModulePort {
	t.Helper()
	for _, port := range model.Ports {
		if port.Descriptor().ID == id {
			return port
		}
	}
	t.Fatalf("port %s not found", id)
	return nil
}

func assertFileContains(t *testing.T, path string, substrings ...string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(b)
	for _, substring := range substrings {
		if !strings.Contains(text, substring) {
			t.Fatalf("%s missing %q in:\n%s", path, substring, text)
		}
	}
}

func readJSONArtifact[T any](t *testing.T, path string) T {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var artifact T
	if err := json.Unmarshal(b, &artifact); err != nil {
		t.Fatalf("decode %s: %v\n%s", path, err, b)
	}
	return artifact
}

func assertSnapshotContractAfterActionStorm(t *testing.T, snapshot viewmodel.ModuleSnapshot) {
	t.Helper()
	if snapshot.Descriptor.ID == "" || snapshot.Descriptor.Title == "" || snapshot.Descriptor.Description == "" || snapshot.Descriptor.Status == "" {
		t.Fatalf("snapshot has incomplete descriptor: %+v", snapshot.Descriptor)
	}
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Fatalf("module %s omitted boundary notice after action storm", snapshot.Descriptor.ID)
	}
	if len(snapshot.Metrics) == 0 {
		t.Fatalf("module %s omitted metrics after action storm", snapshot.Descriptor.ID)
	}
	if len(snapshot.Sections) == 0 {
		t.Fatalf("module %s omitted sections after action storm", snapshot.Descriptor.ID)
	}
	metricIDs := map[string]bool{}
	for _, metric := range snapshot.Metrics {
		if metric.ID == "" || metric.Label == "" {
			t.Fatalf("module %s has incomplete metric: %+v", snapshot.Descriptor.ID, metric)
		}
		if metricIDs[metric.ID] {
			t.Fatalf("module %s has duplicate metric ID %q", snapshot.Descriptor.ID, metric.ID)
		}
		metricIDs[metric.ID] = true
	}
	sectionIDs := map[string]bool{}
	for _, section := range snapshot.Sections {
		if section.ID == "" || section.Title == "" || section.Body == "" || section.Category == "" {
			t.Fatalf("module %s has incomplete section: %+v", snapshot.Descriptor.ID, section)
		}
		if sectionIDs[section.ID] {
			t.Fatalf("module %s has duplicate section ID %q", snapshot.Descriptor.ID, section.ID)
		}
		sectionIDs[section.ID] = true
	}
	actionIDs := map[string]bool{}
	for _, action := range snapshot.Actions {
		if action.ID == "" || action.Label == "" {
			t.Fatalf("module %s has incomplete action: %+v", snapshot.Descriptor.ID, action)
		}
		if action.Kind != viewmodel.ActionCommand && action.Kind != viewmodel.ActionToggle && action.Kind != viewmodel.ActionSelect {
			t.Fatalf("module %s action %s has unknown kind %q", snapshot.Descriptor.ID, action.ID, action.Kind)
		}
		if actionIDs[action.ID] {
			t.Fatalf("module %s has duplicate action ID %q", snapshot.Descriptor.ID, action.ID)
		}
		actionIDs[action.ID] = true
	}
	for _, plot := range snapshot.Plots {
		if plot.ID == "" || plot.Title == "" {
			t.Fatalf("module %s has incomplete plot: %+v", snapshot.Descriptor.ID, plot)
		}
		if len(plot.Series) == 0 {
			t.Fatalf("module %s plot %s has no series", snapshot.Descriptor.ID, plot.ID)
		}
		for _, series := range plot.Series {
			if series.Name == "" || len(series.Points) == 0 {
				t.Fatalf("module %s plot %s has incomplete series: %+v", snapshot.Descriptor.ID, plot.ID, series)
			}
			for _, p := range series.Points {
				for _, value := range []float64{p.X, p.Y, p.V} {
					if math.IsNaN(value) || math.IsInf(value, 0) {
						t.Fatalf("module %s plot %s has non-finite point %+v", snapshot.Descriptor.ID, plot.ID, p)
					}
				}
			}
		}
	}
}
