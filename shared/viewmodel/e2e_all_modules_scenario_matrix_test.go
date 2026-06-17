package viewmodel_test

import (
	"strconv"
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

func TestAllModulesE2EScenarioMatrixDescriptorCompositionAndTrust(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules E2E scenario matrix: runs full descriptor composition across module configurations")
	}
	type scenario struct {
		name         string
		rows         string
		cols         string
		adcBits      string
		dacBits      string
		arch         string
		coupling     string
		isppEngine   string
		waveform     string
		mnistLevels  string
		docQuery     string
		wantArray    string
		wantSummary  []string
		wantMaterial string
	}
	scenarios := []scenario{
		{name: "compact_education", rows: "8", cols: "8", adcBits: "6", dacBits: "6", arch: circuitsvm.ArchitecturePassive, coupling: circuitsvm.CouplingIdeal, isppEngine: circuitsvm.ISPPEngineLevel, waveform: hystvm.WaveformSine, mnistLevels: "16", docQuery: "education simulation", wantArray: "8x8", wantSummary: []string{"8×8", "6-bit ADC/6-bit DAC"}, wantMaterial: "HZO"},
		{name: "balanced_crossbar", rows: "16", cols: "32", adcBits: "8", dacBits: "7", arch: circuitsvm.Architecture1T1R, coupling: circuitsvm.CouplingTierA, isppEngine: circuitsvm.ISPPEngineLevel, waveform: hystvm.WaveformTriangle, mnistLevels: "30", docQuery: "crossbar validation", wantArray: "16x32", wantSummary: []string{"16×32", "8-bit ADC/7-bit DAC"}, wantMaterial: "HZO"},
		{name: "research_export", rows: "32", cols: "16", adcBits: "8", dacBits: "8", arch: circuitsvm.Architecture2T1R, coupling: circuitsvm.CouplingTierB, isppEngine: circuitsvm.ISPPEngineLK, waveform: hystvm.WaveformSquare, mnistLevels: "64", docQuery: "EDA artifact trust", wantArray: "32x16", wantSummary: []string{"32×16", "8-bit ADC/8-bit DAC"}, wantMaterial: "HZO"},
	}

	known := map[viewmodel.ModuleID]viewmodel.ModuleDescriptor{}
	for _, desc := range viewmodel.KnownDescriptors() {
		known[desc.ID] = desc
	}
	if len(known) != 7 {
		t.Fatalf("KnownDescriptors count=%d", len(known))
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			h := hystvm.New()
			x := crossbarvm.New(4, 4)
			mn := mnistvm.New()
			c := circuitsvm.New()
			cmp := comparisonvm.New()
			eda := edavm.New()
			docs := docsvm.New()

			modules := []viewmodel.ModulePort{h, x, mn, c, cmp, eda, docs}
			for _, m := range modules {
				m.Start()
				defer m.Stop()
			}

			applyScenarioActionE2E(t, h, viewmodel.Action{ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": sc.waveform}})
			applyScenarioActionE2E(t, h, viewmodel.Action{ID: hystvm.EventRunPUND})
			applyScenarioActionE2E(t, h, viewmodel.Action{ID: hystvm.EventRunLevelCalibration})

			applyScenarioActionE2E(t, x, viewmodel.Action{ID: "resize", Payload: map[string]string{"rows": sc.rows, "cols": sc.cols}})
			applyScenarioActionE2E(t, x, viewmodel.Action{ID: "toggle_ir"})
			applyScenarioActionE2E(t, x, viewmodel.Action{ID: "run_mvm"})

			applyScenarioActionE2E(t, mn, viewmodel.Action{ID: "sweep_levels", Payload: map[string]string{"levels": sc.mnistLevels}})
			applyScenarioActionE2E(t, mn, viewmodel.Action{ID: "run_inference"})

			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionResizeArray, Payload: map[string]string{"rows": sc.rows, "cols": sc.cols}})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionSetADCBits, Payload: map[string]string{"bits": sc.adcBits}})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionSetDACBits, Payload: map[string]string{"bits": sc.dacBits}})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionSetArchitecture, Payload: map[string]string{"architecture": sc.arch}})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionSetCouplingTier, Payload: map[string]string{"tier": sc.coupling}})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionSetISPPEngine, Payload: map[string]string{"engine": sc.isppEngine}})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionRunRead})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionRunWrite})
			applyScenarioActionE2E(t, c, viewmodel.Action{ID: circuitsvm.ActionRunCompute})

			applyScenarioActionE2E(t, eda, viewmodel.Action{ID: "generate_all"})
			applyScenarioActionE2E(t, docs, viewmodel.Action{ID: "search", Payload: map[string]string{"query": sc.docQuery}})

			seen := map[viewmodel.ModuleID]bool{}
			for _, m := range modules {
				snap := m.Snapshot()
				assertGenericModuleSnapshotE2E(t, snap)
				seen[snap.Descriptor.ID] = true
				knownDesc, ok := known[snap.Descriptor.ID]
				if !ok {
					t.Fatalf("unknown descriptor id %s", snap.Descriptor.ID)
				}
				if knownDesc.Title != snap.Descriptor.Title || knownDesc.Status != snap.Descriptor.Status {
					t.Fatalf("descriptor drift for %s: known=%+v snap=%+v", snap.Descriptor.ID, knownDesc, snap.Descriptor)
				}
				if strings.TrimSpace(snap.Descriptor.BoundaryNotice) == "" {
					t.Fatalf("%s missing boundary notice", snap.Descriptor.ID)
				}
			}
			for id := range known {
				if !seen[id] {
					t.Fatalf("scenario did not exercise module %s", id)
				}
			}

			if got := metricValueAllE2E(x.Snapshot(), "rows") + "x" + metricValueAllE2E(x.Snapshot(), "cols"); got != sc.wantArray {
				t.Fatalf("crossbar array=%s want %s", got, sc.wantArray)
			}
			if got := metricValueAllE2E(c.Snapshot(), "array"); got != sc.wantArray {
				t.Fatalf("circuits array=%s want %s", got, sc.wantArray)
			}
			if got := metricValueAllE2E(c.Snapshot(), "adc"); !strings.Contains(got, sc.adcBits+"-bit") {
				t.Fatalf("adc metric=%q", got)
			}
			if got := metricValueAllE2E(c.Snapshot(), "dac"); !strings.Contains(got, sc.dacBits+"-bit") {
				t.Fatalf("dac metric=%q", got)
			}
			if got := metricValueAllE2E(mn.Snapshot(), "levels"); got != sc.mnistLevels+" levels" {
				t.Fatalf("mnist levels=%q", got)
			}
			if got := metricValueAllE2E(docs.Snapshot(), "search_query"); got != sc.docQuery {
				t.Fatalf("docs search_query=%q", got)
			}

			composition := design.Composition{Hysteresis: h, Crossbar: x, Circuits: c, EDA: eda}
			designSnap := composition.Snapshot()
			if !strings.Contains(designSnap.Material, sc.wantMaterial) {
				t.Fatalf("material=%q", designSnap.Material)
			}
			rows, _ := strconv.Atoi(sc.rows)
			cols, _ := strconv.Atoi(sc.cols)
			adcBits, _ := strconv.Atoi(sc.adcBits)
			dacBits, _ := strconv.Atoi(sc.dacBits)
			if designSnap.ArrayRows != rows || designSnap.ArrayCols != cols || designSnap.ADCResolution != adcBits || designSnap.DACResolution != dacBits {
				t.Fatalf("composition mismatch: %+v", designSnap)
			}
			for _, marker := range sc.wantSummary {
				if !strings.Contains(designSnap.Summary, marker) {
					t.Fatalf("summary %q missing %q", designSnap.Summary, marker)
				}
			}
			if err := composition.ExportDesign(); err != nil {
				t.Fatalf("composition export: %v", err)
			}
		})
	}
}

func applyScenarioActionE2E(t *testing.T, m viewmodel.ModulePort, action viewmodel.Action) {
	t.Helper()
	if err := m.ApplyAction(action); err != nil {
		t.Fatalf("%s action %s: %v", m.Descriptor().ID, action.ID, err)
	}
}
