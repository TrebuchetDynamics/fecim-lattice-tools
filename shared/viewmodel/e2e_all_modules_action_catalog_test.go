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

func TestAllModulesE2EAdvertisedActionCatalogExecutableOrClassified(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules action catalog E2E: initializes all modules and executes actions")
	}
	modules := []viewmodel.ModulePort{
		hystvm.New(),
		crossbarvm.New(8, 8),
		mnistvm.New(),
		circuitsvm.New(),
		comparisonvm.New(),
		edavm.New(),
		docsvm.New(),
	}
	tmp := t.TempDir()

	for _, module := range modules {
		module := module
		t.Run(string(module.Descriptor().ID), func(t *testing.T) {
			module.Start()
			defer module.Stop()
			primeActionCatalogE2E(t, module)
			initial := module.Snapshot()
			seenActions := map[string]bool{}
			for _, advertised := range initial.Actions {
				if advertised.ID == "" || advertised.Label == "" || advertised.Kind == "" {
					t.Fatalf("invalid advertised action: %+v", advertised)
				}
				if seenActions[advertised.ID] {
					t.Fatalf("duplicate advertised action %s", advertised.ID)
				}
				seenActions[advertised.ID] = true
				payload, classified, reason := actionCatalogPayloadE2E(module.Descriptor().ID, advertised.ID, advertised.Payload, tmp)
				if classified {
					t.Logf("classified non-executed action %s/%s: %s", module.Descriptor().ID, advertised.ID, reason)
					continue
				}
				if err := module.ApplyAction(viewmodel.Action{ID: advertised.ID, Kind: advertised.Kind, Payload: payload}); err != nil {
					t.Fatalf("advertised action %s/%s failed with payload %+v: %v", module.Descriptor().ID, advertised.ID, payload, err)
				}
				snap := module.Snapshot()
				assertGenericModuleSnapshotE2E(t, snap)
			}
			if module.Descriptor().ID != viewmodel.ModuleComparison && len(seenActions) == 0 {
				t.Fatalf("%s advertised no actions", module.Descriptor().ID)
			}
			assertCatalogArtifactsE2E(t, module.Descriptor().ID, tmp)
		})
	}
}

func primeActionCatalogE2E(t *testing.T, module viewmodel.ModulePort) {
	t.Helper()
	var actions []viewmodel.Action
	switch module.Descriptor().ID {
	case viewmodel.ModuleHysteresis:
		actions = []viewmodel.Action{{ID: hystvm.EventRunPUND}, {ID: hystvm.EventRunFORC}, {ID: hystvm.EventRunLevelCalibration}}
	case viewmodel.ModuleCrossbar:
		actions = []viewmodel.Action{{ID: "run_mvm"}}
	case viewmodel.ModuleCircuits:
		actions = []viewmodel.Action{{ID: circuitsvm.ActionRunRead}, {ID: circuitsvm.ActionRunWrite}, {ID: circuitsvm.ActionRunCompute}}
	case viewmodel.ModuleEDA:
		actions = []viewmodel.Action{{ID: "generate_all"}}
	case viewmodel.ModuleDocs:
		actions = []viewmodel.Action{{ID: "search", Payload: map[string]string{"query": "action catalog"}}}
	}
	for _, action := range actions {
		if err := module.ApplyAction(action); err != nil {
			t.Fatalf("prime %s/%s: %v", module.Descriptor().ID, action.ID, err)
		}
	}
}

func actionCatalogPayloadE2E(moduleID viewmodel.ModuleID, actionID string, advertised map[string]string, tmp string) (map[string]string, bool, string) {
	if advertised != nil {
		copyPayload := map[string]string{}
		for k, v := range advertised {
			copyPayload[k] = v
		}
		if actionID == circuitsvm.ActionPlayReferenceTiming {
			return copyPayload, true, "playback starts timer-like UI behavior; covered by timing-specific tests"
		}
		return copyPayload, false, ""
	}
	safePath := func(name string) string {
		return filepath.Join(tmp, strings.ReplaceAll(string(moduleID)+"-"+name, "_", "-"))
	}
	switch actionID {
	case hystvm.EventSelectMaterial:
		return map[string]string{"material": "HZO (Si-doped, Park 2015 midpoint)"}, false, ""
	case hystvm.EventSetFieldRange:
		return map[string]string{"min": "-1400", "max": "1400"}, false, ""
	case hystvm.EventSetWaveform:
		return map[string]string{"waveform": hystvm.WaveformTriangle}, false, ""
	case hystvm.EventSetLevelCalibrationLevelCount:
		return map[string]string{"level_count": "30"}, false, ""
	case hystvm.EventSetLevelCalibrationTargetRange:
		return map[string]string{"target_range": "0.90"}, false, ""
	case hystvm.EventSetLevelCalibrationTemperature:
		return map[string]string{"temperature_k": "300"}, false, ""
	case hystvm.EventExportCSV:
		return map[string]string{"path": safePath("loop.csv")}, false, ""
	case hystvm.EventExportLevelCalibration:
		return map[string]string{"path": safePath("levels.json")}, false, ""
	case hystvm.EventExportPUNDCSV:
		return map[string]string{"path": safePath("pund.csv")}, false, ""
	case hystvm.EventExportFORCSweep:
		return map[string]string{"path": safePath("forc-sweep.csv")}, false, ""
	case hystvm.EventExportFORCMatrix:
		return map[string]string{"path": safePath("forc-matrix.csv")}, false, ""
	case hystvm.EventExportFORCMeta:
		return map[string]string{"path": safePath("forc-meta.json")}, false, ""
	case "resize":
		return map[string]string{"rows": "8", "cols": "8"}, false, ""
	case "sweep_levels":
		return map[string]string{"levels": "30"}, false, ""
	case "search":
		return map[string]string{"query": "catalog trust"}, false, ""
	case circuitsvm.ActionExportOperationLog:
		return map[string]string{"path": safePath("operation-log.json")}, false, ""
	case circuitsvm.ActionExportReferenceSpecs:
		return map[string]string{"path": safePath("reference-specs.json")}, false, ""
	case circuitsvm.ActionExportReferenceTiming:
		return map[string]string{"path": safePath("reference-timing.json")}, false, ""
	case circuitsvm.ActionExportReferenceTimingSVG:
		return map[string]string{"path": safePath("reference-timing.svg")}, false, ""
	case circuitsvm.ActionResizeArray:
		return map[string]string{"rows": "8", "cols": "8"}, false, ""
	case circuitsvm.ActionSetOperationMode:
		return map[string]string{"mode": circuitsvm.OperationRead}, false, ""
	case circuitsvm.ActionSetArchitecture:
		return map[string]string{"architecture": circuitsvm.Architecture1T1R}, false, ""
	case circuitsvm.ActionSelectCell:
		return map[string]string{"row": "0", "col": "0"}, false, ""
	case circuitsvm.ActionSetWriteTarget:
		return map[string]string{"level": "12"}, false, ""
	case circuitsvm.ActionSetDACBits:
		return map[string]string{"bits": "6"}, false, ""
	case circuitsvm.ActionSetADCBits:
		return map[string]string{"bits": "6"}, false, ""
	case circuitsvm.ActionSetTIAGain:
		return map[string]string{"gain_ohm": "100000"}, false, ""
	case circuitsvm.ActionSetCouplingTier:
		return map[string]string{"tier": circuitsvm.CouplingTierA}, false, ""
	case circuitsvm.ActionSetISPPEngine:
		return map[string]string{"engine": circuitsvm.ISPPEngineLevel}, false, ""
	case circuitsvm.ActionSetTimingOperation:
		return map[string]string{"operation": "READ"}, false, ""
	case circuitsvm.ActionSetLoggerVerbosity:
		return map[string]string{"verbosity": "normal"}, false, ""
	case circuitsvm.ActionPlayReferenceTiming:
		return nil, true, "playback starts timer-like UI behavior; covered by timing-specific tests"
	}
	return nil, false, ""
}

func assertCatalogArtifactsE2E(t *testing.T, moduleID viewmodel.ModuleID, tmp string) {
	t.Helper()
	patterns := map[viewmodel.ModuleID][]string{
		viewmodel.ModuleHysteresis: {"hysteresis-loop.csv", "hysteresis-levels.json", "hysteresis-pund.csv", "hysteresis-forc-sweep.csv", "hysteresis-forc-matrix.csv", "hysteresis-forc-meta.json"},
		viewmodel.ModuleCircuits:   {"circuits-operation-log.json", "circuits-reference-specs.json", "circuits-reference-timing.json", "circuits-reference-timing.svg"},
	}
	for _, name := range patterns[moduleID] {
		path := filepath.Join(tmp, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("artifact %s is empty", path)
		}
	}
}
