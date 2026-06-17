package viewmodel_test

import (
	"fmt"
	"strings"
	"sync"
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

func TestAllModulesE2EInvalidActionMatrixPreservesCoreContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules E2E resilience: runs valid+invalid action sequences across all modules")
	}
	cases := []struct {
		name    string
		module  viewmodel.ModulePort
		valid   []viewmodel.Action
		invalid []viewmodel.Action
		keep    []string
	}{
		{name: "hysteresis", module: hystvm.New(), valid: []viewmodel.Action{{ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": hystvm.WaveformTriangle}}}, invalid: []viewmodel.Action{{ID: hystvm.EventSetFieldRange, Payload: map[string]string{"min": "bad"}}, {ID: hystvm.EventSelectMaterial, Payload: map[string]string{"material": "not-real"}}, {ID: "unknown"}}, keep: []string{"material", "waveform"}},
		{name: "crossbar", module: crossbarvm.New(5, 4), valid: []viewmodel.Action{{ID: "run_mvm"}}, invalid: []viewmodel.Action{{ID: "resize", Payload: map[string]string{"rows": "0", "cols": "4"}}, {ID: "resize", Payload: map[string]string{"rows": "129", "cols": "4"}}, {ID: "unknown"}}, keep: []string{"rows", "cols"}},
		{name: "mnist", module: mnistvm.New(), valid: []viewmodel.Action{{ID: "sweep_levels", Payload: map[string]string{"levels": "30"}}}, invalid: []viewmodel.Action{{ID: "sweep_levels", Payload: map[string]string{"levels": "NaN"}}, {ID: "unknown"}}, keep: []string{"levels", "accuracy"}},
		{name: "circuits", module: circuitsvm.New(), valid: []viewmodel.Action{{ID: circuitsvm.ActionResizeArray, Payload: map[string]string{"rows": "8", "cols": "8"}}, {ID: circuitsvm.ActionRunCompute}}, invalid: []viewmodel.Action{{ID: circuitsvm.ActionResizeArray, Payload: map[string]string{"rows": "3", "cols": "8"}}, {ID: circuitsvm.ActionSetADCBits, Payload: map[string]string{"bits": "99"}}, {ID: circuitsvm.ActionSetArchitecture, Payload: map[string]string{"architecture": "invalid"}}, {ID: "unknown"}}, keep: []string{"array", "architecture", "adc"}},
		{name: "comparison", module: comparisonvm.New(), invalid: []viewmodel.Action{{ID: "any"}, {ID: "generate_all"}}, keep: []string{"count"}},
		{name: "eda", module: edavm.New(), valid: []viewmodel.Action{{ID: "generate_all"}}, invalid: []viewmodel.Action{{ID: "unknown"}, {ID: "resize"}}, keep: []string{"design", "process", "array"}},
		{name: "docs", module: docsvm.New(), valid: []viewmodel.Action{{ID: "search", Payload: map[string]string{"query": "validation"}}}, invalid: []viewmodel.Action{{ID: "search"}, {ID: "unknown"}}, keep: []string{"modules", "papers", "active_page", "search_query"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, action := range tc.valid {
				if err := tc.module.ApplyAction(action); err != nil {
					t.Fatalf("valid setup %s: %v", action.ID, err)
				}
			}
			before := tc.module.Snapshot()
			beforeSig := allModuleCoreSignatureE2E(before, tc.keep)
			for _, action := range tc.invalid {
				if err := tc.module.ApplyAction(action); err == nil {
					t.Fatalf("invalid action %s unexpectedly succeeded", action.ID)
				}
				after := tc.module.Snapshot()
				if got := allModuleCoreSignatureE2E(after, tc.keep); got != beforeSig {
					t.Fatalf("%s invalid %s changed core signature\nbefore=%s\nafter =%s", tc.name, action.ID, beforeSig, got)
				}
				assertGenericModuleSnapshotE2E(t, after)
			}
		})
	}
}

func TestAllModulesE2EParallelLifecycleAndActionStorm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules E2E parallel lifecycle: concurrent goroutines storm all module actions")
	}
	cases := []struct {
		name    string
		factory func() viewmodel.ModulePort
		actions []viewmodel.Action
	}{
		{name: "hysteresis", factory: func() viewmodel.ModulePort { return hystvm.New() }, actions: []viewmodel.Action{{ID: hystvm.EventToggleSimulation}, {ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": hystvm.WaveformSine}}, {ID: hystvm.EventRunPUND}}},
		{name: "crossbar", factory: func() viewmodel.ModulePort { return crossbarvm.New(4, 4) }, actions: []viewmodel.Action{{ID: "run_mvm"}, {ID: "toggle_ir"}, {ID: "resize", Payload: map[string]string{"rows": "8", "cols": "8"}}}},
		{name: "mnist", factory: func() viewmodel.ModulePort { return mnistvm.New() }, actions: []viewmodel.Action{{ID: "run_inference"}, {ID: "sweep_levels", Payload: map[string]string{"levels": "16"}}, {ID: "sweep_levels", Payload: map[string]string{"levels": "64"}}}},
		{name: "circuits", factory: func() viewmodel.ModulePort { return circuitsvm.New() }, actions: []viewmodel.Action{{ID: circuitsvm.ActionRunRead}, {ID: circuitsvm.ActionRunWrite}, {ID: circuitsvm.ActionRunCompute}, {ID: circuitsvm.ActionToggleISPP}}},
		{name: "comparison", factory: func() viewmodel.ModulePort { return comparisonvm.New() }},
		{name: "eda", factory: func() viewmodel.ModulePort { return edavm.New() }, actions: []viewmodel.Action{{ID: "generate_spice"}, {ID: "generate_all"}}},
		{name: "docs", factory: func() viewmodel.ModulePort { return docsvm.New() }, actions: []viewmodel.Action{{ID: "search", Payload: map[string]string{"query": "curriculum"}}, {ID: "start_curriculum"}}},
	}
	var wg sync.WaitGroup
	for _, tc := range cases {
		tc := tc
		for worker := 0; worker < 3; worker++ {
			worker := worker
			wg.Add(1)
			go func() {
				defer wg.Done()
				m := tc.factory()
				m.Start()
				defer m.Stop()
				for i := 0; i < 5; i++ {
					for _, action := range tc.actions {
						if err := m.ApplyAction(action); err != nil {
							t.Errorf("%s worker %d action %s: %v", tc.name, worker, action.ID, err)
							return
						}
						assertGenericModuleSnapshotE2E(t, m.Snapshot())
					}
				}
			}()
		}
	}
	wg.Wait()
}

func allModuleCoreSignatureE2E(s viewmodel.ModuleSnapshot, metricIDs []string) string {
	parts := []string{string(s.Descriptor.ID), s.Descriptor.Title, fmt.Sprintf("sections=%d", len(s.Sections)), fmt.Sprintf("plots=%d", len(s.Plots))}
	for _, id := range metricIDs {
		parts = append(parts, id+"="+metricValueAllE2E(s, id))
	}
	return strings.Join(parts, "|")
}

func assertGenericModuleSnapshotE2E(t *testing.T, s viewmodel.ModuleSnapshot) {
	t.Helper()
	if s.Descriptor.ID == "" || s.Descriptor.Title == "" || s.Descriptor.Status != viewmodel.StatusFunctional {
		t.Fatalf("invalid descriptor: %+v", s.Descriptor)
	}
	if len(s.Metrics) == 0 && len(s.Sections) == 0 {
		t.Fatalf("snapshot has no observable content: %+v", s)
	}
}
