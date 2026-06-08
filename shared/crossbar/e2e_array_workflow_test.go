package crossbar

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModule2CrossbarE2EWideArrayNonIdealityExportWorkflow(t *testing.T) {
	arrays := []struct {
		name string
		cfg  *Config
	}{
		{name: "ideal-rectangular-fefet", cfg: &Config{Rows: 3, Cols: 4, NoiseLevel: 0, ADCBits: 10, DACBits: 10, ConductanceModel: ConductanceLinear}},
		{name: "variation-gradient-endurance", cfg: &Config{Rows: 4, Cols: 3, NoiseLevel: 0, ADCBits: 8, DACBits: 8, ConductanceModel: ConductanceExponential, ProcessVariation: &ProcessVariationConfig{DeviceSigma: 0, GradientX: 0.002, GradientY: -0.001, EdgeEffect: 0.03}, Endurance: &EnduranceConfig{Enabled: true, FatigueThreshold: 2, FailureThreshold: 20}, HalfSelect: &HalfSelectConfig{Enabled: true, DisturbThreshold: 0.3, DisturbRate: 0.002}}},
	}

	for _, tc := range arrays {
		t.Run(tc.name, func(t *testing.T) {
			arr, err := NewArray(tc.cfg)
			if err != nil {
				t.Fatalf("NewArray() error = %v", err)
			}
			defer arr.Destroy()

			weights := deterministicE2EWeights(tc.cfg.Rows, tc.cfg.Cols)
			if err := arr.ProgramWeightMatrix(weights); err != nil {
				t.Fatalf("ProgramWeightMatrix() error = %v", err)
			}
			reads, writes := arr.GetStats()
			if reads != 0 || writes != int64(tc.cfg.Rows*tc.cfg.Cols) {
				t.Fatalf("stats after program = reads %d writes %d", reads, writes)
			}

			matrix := arr.GetConductanceMatrix()
			assertE2EConductanceMatrixQuantized(t, matrix, weights)
			effective := arr.GetEffectiveConductanceMatrix()
			assertE2EMatrixShapeAndFinite(t, effective, tc.cfg.Rows, tc.cfg.Cols)

			input := deterministicE2EInput(tc.cfg.Cols)
			mvm, err := arr.MVM(input)
			if err != nil {
				t.Fatalf("MVM() error = %v", err)
			}
			vmm, err := arr.VMM(deterministicE2EInput(tc.cfg.Rows))
			if err != nil {
				t.Fatalf("VMM() error = %v", err)
			}
			assertE2EVectorFinite01(t, "mvm", mvm, tc.cfg.Rows)
			assertE2EVectorFinite01(t, "vmm", vmm, tc.cfg.Cols)
			reads, writes = arr.GetStats()
			if reads != int64(tc.cfg.Rows+tc.cfg.Cols) || writes != int64(tc.cfg.Rows*tc.cfg.Cols) {
				t.Fatalf("stats after MVM/VMM = reads %d writes %d", reads, writes)
			}

			opts := DefaultMVMOptions()
			opts.EnableDrift = true
			opts.EnableVariation = true
			opts.Temperature = 330
			for _, arch := range []string{"0T1R", "1T1R", "2T1R"} {
				t.Run("arch-"+arch, func(t *testing.T) {
					opts.Architecture = arch
					result, err := arr.MVMWithNonIdealities(input, opts)
					if err != nil {
						t.Fatalf("MVMWithNonIdealities(%s) error = %v", arch, err)
					}
					assertE2EMVMResultContract(t, result, tc.cfg.Rows, tc.cfg.Rows*len(input))
					if result.IRDropAnalysis == nil || result.SneakPathAnalysis == nil || result.SneakTrace == nil {
						t.Fatalf("non-ideality analyses missing for %s: IR=%v sneak=%v trace=%v", arch, result.IRDropAnalysis != nil, result.SneakPathAnalysis != nil, result.SneakTrace != nil)
					}
					degradation, err := arr.ComputeAccuracyDegradationWithOptions(input, 92.5, opts)
					if err != nil {
						t.Fatalf("ComputeAccuracyDegradationWithOptions(%s) error = %v", arch, err)
					}
					if degradation.BaselineAccuracy != 92.5 || degradation.FinalAccuracy < 0 || len(degradation.Degradations) == 0 {
						t.Fatalf("invalid degradation report for %s: %+v", arch, degradation)
					}
				})
			}

			outDir := t.TempDir()
			weightsPath := filepath.Join(outDir, tc.name+"-weights.csv")
			analysisPath := filepath.Join(outDir, tc.name+"-analysis.json")
			result, err := arr.MVMWithNonIdealities(input, DefaultMVMOptions())
			if err != nil {
				t.Fatalf("MVMWithNonIdealities(default) error = %v", err)
			}
			if err := arr.ExportWeightsCSV(weightsPath); err != nil {
				t.Fatalf("ExportWeightsCSV() error = %v", err)
			}
			if err := arr.ExportAnalysisJSON(analysisPath, result); err != nil {
				t.Fatalf("ExportAnalysisJSON() error = %v", err)
			}
			assertE2EWeightsCSV(t, weightsPath, tc.cfg.Rows*tc.cfg.Cols)
			assertE2EAnalysisJSON(t, analysisPath)
		})
	}
}

func TestModule2CrossbarE2EFeCAPChargeWorkflowAndInvalidIsolation(t *testing.T) {
	cfg := &Config{Rows: 3, Cols: 3, ADCBits: 8, DACBits: 8, CellType: CellTypeFeCAP, CMin: 1e-15, CMax: 9e-15, PulseDuration: 2e-9, CapacitanceModel: CapModelLinear}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray(FeCAP) error = %v", err)
	}
	defer arr.Destroy()

	caps := [][]float64{{0, 0.25, 0.5}, {0.75, 1, 0.33}, {0.66, 0.1, 0.9}}
	if err := arr.ProgramCapacitanceMatrix(caps); err != nil {
		t.Fatalf("ProgramCapacitanceMatrix() error = %v", err)
	}
	capMatrix := arr.GetCapacitanceMatrix()
	assertE2EMatrixShapeAndFinite(t, capMatrix, 3, 3)
	charge, err := arr.MVMCharge([]float64{0.2, 0.5, 1.0})
	if err != nil {
		t.Fatalf("MVMCharge() error = %v", err)
	}
	quantized, err := arr.MVMChargeQuantized([]float64{0.2, 0.5, 1.0}, 5e-15)
	if err != nil {
		t.Fatalf("MVMChargeQuantized() error = %v", err)
	}
	assertE2EVectorFiniteNonNegative(t, "charge", charge, 3)
	assertE2EVectorFiniteNonNegative(t, "quantized", quantized, 3)
	if energy := arr.MVMChargeEnergy([]float64{0.2, 0.5, 1.0}); !(energy > 0) || math.IsNaN(energy) {
		t.Fatalf("MVMChargeEnergy() = %g, want positive finite", energy)
	}
	if _, err := arr.MVMWithNonIdealities([]float64{0.2, 0.5, 1.0}, DefaultMVMOptions()); err == nil || !strings.Contains(err.Error(), "CellTypeFeFET") {
		t.Fatalf("MVMWithNonIdealities on FeCAP error = %v, want FeFET guidance", err)
	}

	before := arr.GetCapacitanceMatrix()
	invalids := []func() error{
		func() error { return arr.ProgramCapacitance(-1, 0, 0.5) },
		func() error { return arr.ProgramCapacitance(0, 9, 0.5) },
		func() error { _, err := arr.MVMCharge([]float64{}); return err },
		func() error { _, err := arr.MVMCharge([]float64{1, 2, 3, 4}); return err },
	}
	for i, op := range invalids {
		if err := op(); err == nil {
			t.Fatalf("invalid FeCAP operation %d returned nil error", i)
		}
	}
	after := arr.GetCapacitanceMatrix()
	if !sameE2EMatrix(before, after) {
		t.Fatalf("invalid FeCAP operations mutated capacitance matrix: before=%v after=%v", before, after)
	}
}

func TestModule2CrossbarE2EInvalidConfigurationAndOperationMatrix(t *testing.T) {
	invalidConfigs := []*Config{{Rows: 0, Cols: 4, ADCBits: 8, DACBits: 8}, {Rows: 4, Cols: 0, ADCBits: 8, DACBits: 8}, {Rows: 4, Cols: 4, ADCBits: 0, DACBits: 8}, {Rows: 4, Cols: 4, ADCBits: 8, DACBits: 0}, {Rows: 4097, Cols: 1, ADCBits: 8, DACBits: 8}}
	for i, cfg := range invalidConfigs {
		if arr, err := NewArray(cfg); err == nil || arr != nil {
			t.Fatalf("invalid config %d produced arr=%v err=%v", i, arr, err)
		}
	}

	arr, err := NewArray(&Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray(valid) error = %v", err)
	}
	defer arr.Destroy()
	if err := arr.ProgramWeightMatrix([][]float64{{0.2, 0.4}, {0.6, 0.8}}); err != nil {
		t.Fatalf("initial ProgramWeightMatrix error = %v", err)
	}
	before := arr.GetConductanceMatrix()
	invalidOps := []func() error{
		func() error { return arr.ProgramWeight(-1, 0, 0.2) },
		func() error { return arr.ProgramWeight(0, 9, 0.2) },
		func() error { return arr.ProgramWeightMatrix(nil) },
		func() error { return arr.ProgramWeightMatrix([][]float64{{1, 2, 3}}) },
		func() error { _, err := arr.MVM(nil); return err },
		func() error { _, err := arr.MVM([]float64{1, math.NaN()}); return err },
		func() error { _, err := arr.MVM([]float64{1, 2, 3}); return err },
		func() error { _, err := arr.VMM([]float64{1, 2, 3}); return err },
	}
	for i, op := range invalidOps {
		if err := op(); err == nil {
			t.Fatalf("invalid operation %d returned nil error", i)
		}
	}
	after := arr.GetConductanceMatrix()
	if !sameE2EMatrix(before, after) {
		t.Fatalf("invalid operations mutated conductance matrix: before=%v after=%v", before, after)
	}
}

func deterministicE2EWeights(rows, cols int) [][]float64 {
	weights := make([][]float64, rows)
	for i := range weights {
		weights[i] = make([]float64, cols)
		for j := range weights[i] {
			weights[i][j] = float64((i+1)*(j+2)%11) / 10.0
		}
	}
	return weights
}

func deterministicE2EInput(n int) []float64 {
	input := make([]float64, n)
	for i := range input {
		input[i] = float64((i*3+2)%7) / 6.0
	}
	return input
}

func assertE2EConductanceMatrixQuantized(t *testing.T, matrix, source [][]float64) {
	t.Helper()
	assertE2EMatrixShapeAndFinite(t, matrix, len(source), len(source[0]))
	for i := range source {
		for j := range source[i] {
			want := QuantizeToLevels(source[i][j])
			if math.Abs(matrix[i][j]-want) > 1e-12 {
				t.Fatalf("matrix[%d][%d] = %.12f, want quantized %.12f", i, j, matrix[i][j], want)
			}
		}
	}
}

func assertE2EMatrixShapeAndFinite(t *testing.T, matrix [][]float64, rows, cols int) {
	t.Helper()
	if len(matrix) != rows {
		t.Fatalf("matrix rows = %d, want %d", len(matrix), rows)
	}
	for i, row := range matrix {
		if len(row) != cols {
			t.Fatalf("matrix row %d cols = %d, want %d", i, len(row), cols)
		}
		for j, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("matrix[%d][%d] invalid: %g", i, j, value)
			}
		}
	}
}

func assertE2EVectorFinite01(t *testing.T, name string, vector []float64, wantLen int) {
	t.Helper()
	assertE2EVectorFiniteNonNegative(t, name, vector, wantLen)
	for i, value := range vector {
		if value > 1 {
			t.Fatalf("%s[%d] = %g, want <= 1", name, i, value)
		}
	}
}

func assertE2EVectorFiniteNonNegative(t *testing.T, name string, vector []float64, wantLen int) {
	t.Helper()
	if len(vector) != wantLen {
		t.Fatalf("%s length = %d, want %d", name, len(vector), wantLen)
	}
	for i, value := range vector {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			t.Fatalf("%s[%d] invalid: %g", name, i, value)
		}
	}
}

func assertE2EMVMResultContract(t *testing.T, result *MVMResult, rows, macs int) {
	t.Helper()
	if result == nil {
		t.Fatal("nil MVMResult")
	}
	assertE2EVectorFinite01(t, "ideal", result.IdealOutput, rows)
	assertE2EVectorFinite01(t, "actual", result.ActualOutput, rows)
	if result.MACOperations != macs || result.TotalEnergy <= 0 || result.Latency <= 0 || result.Throughput <= 0 || result.EnergyEfficiency <= 0 {
		t.Fatalf("invalid MVM metrics: MAC=%d energy=%g latency=%g throughput=%g efficiency=%g", result.MACOperations, result.TotalEnergy, result.Latency, result.Throughput, result.EnergyEfficiency)
	}
	for name, value := range map[string]float64{"rmse": result.RMSE, "maxError": result.MaxError, "meanError": result.MeanError, "accuracyLoss": result.AccuracyLoss} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			t.Fatalf("%s invalid: %g", name, value)
		}
	}
}

func assertE2EWeightsCSV(t *testing.T, path string, wantDataRows int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open weights CSV: %v", err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read weights CSV: %v", err)
	}
	if len(rows) != wantDataRows+1 {
		t.Fatalf("weights CSV rows = %d, want %d", len(rows), wantDataRows+1)
	}
	if strings.Join(rows[0], ",") != "row,col,level,conductance,conductance_uS" {
		t.Fatalf("unexpected CSV header: %v", rows[0])
	}
}

func assertE2EAnalysisJSON(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read analysis JSON: %v", err)
	}
	var report map[string]any
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode analysis JSON: %v", err)
	}
	for _, key := range []string{"array_size", "total_energy_pj", "rmse", "energy_efficiency"} {
		if _, ok := report[key]; !ok {
			t.Fatalf("analysis JSON missing key %q: %s", key, data)
		}
	}
}

func sameE2EMatrix(a, b [][]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
