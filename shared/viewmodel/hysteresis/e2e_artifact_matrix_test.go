package hysteresis

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule1ViewModelE2EWideArtifactMatrix(t *testing.T) {
	tests := []struct {
		name        string
		material    string
		waveform    string
		fieldMin    string
		fieldMax    string
		levels      string
		targetRange string
		temperature string
		reversals   string
	}{
		{name: "default-triangle-16-levels", material: "HZO", waveform: WaveformTriangle, fieldMin: "-2200", fieldMax: "2200", levels: "16", targetRange: "0.70", temperature: "325", reversals: "5"},
		{name: "fecim-square-24-levels", material: "FeCIM HZO", waveform: WaveformSquare, fieldMin: "-1800", fieldMax: "1900", levels: "24", targetRange: "0.80", temperature: "350", reversals: "7"},
		{name: "cryogenic-manual-12-levels", material: "Cryogenic", waveform: WaveformManual, fieldMin: "-3200", fieldMax: "3100", levels: "12", targetRange: "0.60", temperature: "250", reversals: "4"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			module := New()
			outDir := t.TempDir()
			paths := map[string]string{
				"loop":        filepath.Join(outDir, "loop.csv"),
				"pund":        filepath.Join(outDir, "pund.csv"),
				"forc_sweep":  filepath.Join(outDir, "forc-sweep.csv"),
				"forc_matrix": filepath.Join(outDir, "forc-matrix.csv"),
				"forc_meta":   filepath.Join(outDir, "forc-meta.json"),
				"levels":      filepath.Join(outDir, "levels.json"),
			}

			materialName := module1E2EMaterialNameContaining(t, module, tc.material)
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSelectMaterial, Kind: viewmodel.ActionSelect, Payload: map[string]string{"material": materialName}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSetFieldRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"min": tc.fieldMin, "max": tc.fieldMax}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": tc.waveform}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSetLevelCalibrationLevelCount, Kind: viewmodel.ActionCommand, Payload: map[string]string{"level_count": tc.levels}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSetLevelCalibrationTargetRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"target_range": tc.targetRange}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSetLevelCalibrationTemperature, Kind: viewmodel.ActionCommand, Payload: map[string]string{"temperature_k": tc.temperature}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventExportLevelCalibration, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": paths["levels"]}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventExportCSV, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": paths["loop"]}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventRunPUND, Kind: viewmodel.ActionCommand})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventExportPUNDCSV, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": paths["pund"]}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventRunFORC, Kind: viewmodel.ActionCommand, Payload: map[string]string{"reversals": tc.reversals}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventExportFORCSweep, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": paths["forc_sweep"], "reversals": tc.reversals}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventExportFORCMatrix, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": paths["forc_matrix"], "reversals": tc.reversals}})
			dispatchModule1E2E(t, module, viewmodel.Action{ID: EventExportFORCMeta, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": paths["forc_meta"], "reversals": tc.reversals}})

			snapshot := module.Snapshot()
			assertModule1E2EMetric(t, snapshot, "material", materialName)
			assertModule1E2EMetric(t, snapshot, "waveform", tc.waveform)
			assertModule1E2EMetric(t, snapshot, "level_calibration_state", string(LevelCalibrationFresh))
			assertModule1E2EHasSection(t, snapshot, "level_calibration_summary")
			assertModule1E2EHasSection(t, snapshot, "diagnostic_pund")
			assertModule1E2EHasSection(t, snapshot, "diagnostic_forc")
			for _, id := range []string{"pe_loop", "pund_current_waveforms", "forc_density_heatmap"} {
				if !module1E2EHasPlot(snapshot, id) {
					t.Fatalf("snapshot missing plot %s", id)
				}
			}

			assertModule1E2EFileContains(t, paths["loop"], "Index,E_field_kV_cm,Polarization_uC_cm2")
			assertModule1E2EFileContains(t, paths["pund"], "metric,value,unit", "Qsw_positive", "switching_ratio")
			assertModule1E2EFileContains(t, paths["forc_sweep"], "reversal_field_vm,applied_field_vm,polarization_cm2")
			assertModule1E2EFileContains(t, paths["forc_matrix"], "Ea_Vm,Eb_Vm,density")

			levels := readModule1E2EJSON[struct {
				ArtifactType   string `json:"artifact_type"`
				BoundaryNotice string `json:"boundary_notice"`
				Material       string `json:"material"`
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
					LevelIndex int `json:"level_index"`
				} `json:"levels"`
			}](t, paths["levels"])
			if levels.ArtifactType != "level_calibration" || !strings.Contains(levels.BoundaryNotice, "SIMULATION OUTPUT") || levels.Material != materialName {
				t.Fatalf("level artifact header = %+v", levels)
			}
			if len(levels.Levels) != levels.Inputs.LevelCount || levels.Summary.AscendingEntries != levels.Inputs.LevelCount || levels.Summary.DescendingEntries != levels.Inputs.LevelCount {
				t.Fatalf("level artifact counts inconsistent: %+v", levels)
			}
			if levels.Summary.Monotonicity == "" || levels.Levels[0].LevelIndex != 0 || levels.Levels[len(levels.Levels)-1].LevelIndex != levels.Inputs.LevelCount-1 {
				t.Fatalf("level artifact endpoints/monotonicity invalid: %+v", levels)
			}

			meta := readModule1E2EJSON[struct {
				Material string `json:"material"`
				Waveform string `json:"waveform"`
				Curves   int    `json:"curves"`
				Boundary string `json:"boundary"`
			}](t, paths["forc_meta"])
			if meta.Material != materialName || meta.Waveform != tc.waveform || meta.Curves == 0 || !strings.Contains(meta.Boundary, "SIMULATION OUTPUT") {
				t.Fatalf("FORC metadata = %+v, want material/waveform/curves/boundary", meta)
			}
		})
	}
}

func TestModule1ViewModelE2EInvalidActionMatrixPreservesState(t *testing.T) {
	module := New()
	dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": WaveformTriangle}})
	dispatchModule1E2E(t, module, viewmodel.Action{ID: EventSetLevelCalibrationLevelCount, Kind: viewmodel.ActionCommand, Payload: map[string]string{"level_count": "16"}})
	before := module.Snapshot()

	invalid := []viewmodel.Action{
		{ID: EventSelectMaterial, Kind: viewmodel.ActionSelect, Payload: map[string]string{"material": "not-a-material"}},
		{ID: EventSetWaveform, Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": "sawtooth"}},
		{ID: EventSetFieldRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"min": "bad-number"}},
		{ID: EventSetLevelCalibrationLevelCount, Kind: viewmodel.ActionCommand, Payload: map[string]string{"level_count": "1"}},
		{ID: EventSetLevelCalibrationTargetRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"target_range": "1.5"}},
		{ID: EventSetLevelCalibrationTemperature, Kind: viewmodel.ActionCommand, Payload: map[string]string{"temperature_k": "900"}},
		{ID: EventExportLevelCalibration, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": filepath.Join(t.TempDir(), "not-ready.json")}},
		{ID: "unknown", Kind: viewmodel.ActionCommand},
	}
	for _, action := range invalid {
		if err := module.ApplyAction(action); err == nil {
			t.Fatalf("ApplyAction(%s) error = nil, want rejection", action.ID)
		}
	}
	after := module.Snapshot()
	for _, id := range []string{"material", "waveform", "level_calibration_inputs"} {
		if got, want := module1E2EMetric(after, id), module1E2EMetric(before, id); got != want {
			t.Fatalf("metric %s changed after invalid actions: got %q want %q", id, got, want)
		}
	}
	if !errors.Is(module.ApplyAction(viewmodel.Action{ID: "unknown", Kind: viewmodel.ActionCommand}), viewmodel.ErrUnsupportedAction) {
		t.Fatalf("unknown action should wrap/return ErrUnsupportedAction")
	}
}

func module1E2EMaterialNameContaining(t *testing.T, module *Module, needle string) string {
	t.Helper()
	for _, section := range module.Snapshot().Sections {
		if strings.HasPrefix(section.ID, "material_") && strings.Contains(section.Title, needle) {
			return section.Title
		}
	}
	t.Fatalf("no material containing %q found", needle)
	return ""
}

func dispatchModule1E2E(t *testing.T, module *Module, action viewmodel.Action) {
	t.Helper()
	if err := module.ApplyAction(action); err != nil {
		t.Fatalf("ApplyAction(%s): %v", action.ID, err)
	}
}

func assertModule1E2EMetric(t *testing.T, snapshot viewmodel.ModuleSnapshot, id, want string) {
	t.Helper()
	if got := module1E2EMetric(snapshot, id); got != want {
		t.Fatalf("metric %s = %q, want %q", id, got, want)
	}
}

func module1E2EMetric(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range snapshot.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}

func assertModule1E2EHasSection(t *testing.T, snapshot viewmodel.ModuleSnapshot, id string) {
	t.Helper()
	for _, section := range snapshot.Sections {
		if section.ID == id && strings.TrimSpace(section.Body) != "" {
			return
		}
	}
	t.Fatalf("snapshot missing non-empty section %s", id)
}

func module1E2EHasPlot(snapshot viewmodel.ModuleSnapshot, id string) bool {
	for _, plot := range snapshot.Plots {
		if plot.ID == id && len(plot.Series) > 0 && len(plot.Series[0].Points) > 0 {
			return true
		}
	}
	return false
}

func assertModule1E2EFileContains(t *testing.T, path string, needles ...string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(data)
	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			t.Fatalf("%s missing %q\n%s", path, needle, text)
		}
	}
}

func readModule1E2EJSON[T any](t *testing.T, path string) T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode %s: %v\n%s", path, err, string(data))
	}
	return value
}
