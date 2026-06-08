package crossbar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestModule2CrossbarE2EParallelIndependentArrayWorkflows(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		rows int
		cols int
	}{
		{name: "small-square", rows: 4, cols: 4},
		{name: "wide", rows: 3, cols: 6},
		{name: "tall", rows: 6, cols: 3},
	}
	var wg sync.WaitGroup
	for _, tc := range cases {
		tc := tc
		wg.Add(1)
		go func() {
			defer wg.Done()
			arr, err := NewArray(&Config{Rows: tc.rows, Cols: tc.cols, NoiseLevel: 0.01, ADCBits: 8, DACBits: 8, ProcessVariation: &ProcessVariationConfig{DeviceSigma: 0.005}})
			if err != nil {
				t.Errorf("%s NewArray: %v", tc.name, err)
				return
			}
			defer arr.Destroy()
			if err := arr.ProgramWeightMatrix(parallelE2EWeights(tc.rows, tc.cols)); err != nil {
				t.Errorf("%s ProgramWeightMatrix: %v", tc.name, err)
				return
			}
			input := deterministicE2EInput(tc.cols)
			for i := 0; i < 8; i++ {
				mvm, err := arr.MVM(input)
				if err != nil || len(mvm) != tc.rows {
					t.Errorf("%s MVM[%d] len/err = %d/%v", tc.name, i, len(mvm), err)
					return
				}
				unc, err := arr.MVMWithUncertainty(input)
				if err != nil || len(unc.Output) != tc.rows || len(unc.Uncertainty) != tc.rows {
					t.Errorf("%s uncertainty[%d] invalid: %+v err=%v", tc.name, i, unc, err)
					return
				}
				res, err := arr.MVMWithNonIdealities(input, DefaultMVMOptions())
				if err != nil || res.MACOperations != tc.rows*tc.cols {
					t.Errorf("%s nonideal[%d] invalid: %+v err=%v", tc.name, i, res, err)
					return
				}
			}
			reads, writes := arr.GetStats()
			if writes != int64(tc.rows*tc.cols) || reads < int64(tc.rows*8) {
				t.Errorf("%s stats invalid: reads=%d writes=%d", tc.name, reads, writes)
				return
			}
			outDir := t.TempDir()
			csvPath := filepath.Join(outDir, tc.name+".csv")
			jsonPath := filepath.Join(outDir, tc.name+".json")
			res, err := arr.MVMWithNonIdealities(input, DefaultMVMOptions())
			if err != nil {
				t.Errorf("%s final nonideal: %v", tc.name, err)
				return
			}
			if err := arr.ExportWeightsCSV(csvPath); err != nil {
				t.Errorf("%s ExportWeightsCSV: %v", tc.name, err)
				return
			}
			if err := arr.ExportAnalysisJSON(jsonPath, res); err != nil {
				t.Errorf("%s ExportAnalysisJSON: %v", tc.name, err)
				return
			}
			if _, err := os.Stat(csvPath); err != nil {
				t.Errorf("%s csv missing: %v", tc.name, err)
			}
			data, err := os.ReadFile(jsonPath)
			if err != nil {
				t.Errorf("%s read json: %v", tc.name, err)
				return
			}
			var report map[string]any
			if err := json.Unmarshal(data, &report); err != nil || report["array_size"] == nil {
				t.Errorf("%s invalid json report=%v err=%v data=%s", tc.name, report, err, data)
			}
		}()
	}
	wg.Wait()
}

func TestModule2CrossbarE2EConcurrentReadOnlyOperationsOnSingleArray(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 5, Cols: 5, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	defer arr.Destroy()
	if err := arr.ProgramWeightMatrix(parallelE2EWeights(5, 5)); err != nil {
		t.Fatalf("ProgramWeightMatrix: %v", err)
	}
	input := deterministicE2EInput(5)
	var wg sync.WaitGroup
	for worker := 0; worker < 12; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for iter := 0; iter < 20; iter++ {
				if _, err := arr.MVM(input); err != nil {
					t.Errorf("worker %d MVM[%d]: %v", worker, iter, err)
					return
				}
				if _, err := arr.VMM(input); err != nil {
					t.Errorf("worker %d VMM[%d]: %v", worker, iter, err)
					return
				}
				if _, err := arr.MVMWithUncertainty(input); err != nil {
					t.Errorf("worker %d uncertainty[%d]: %v", worker, iter, err)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
	reads, writes := arr.GetStats()
	if writes != 25 || reads < 12*20*10 { // MVM + VMM each add rows/cols; uncertainty wraps another MVM.
		t.Fatalf("concurrent stats invalid: reads=%d writes=%d", reads, writes)
	}
}

func parallelE2EWeights(rows, cols int) [][]float64 {
	weights := make([][]float64, rows)
	for r := range weights {
		weights[r] = make([]float64, cols)
		for c := range weights[r] {
			weights[r][c] = float64((r*7+c*5+3)%29) / 28.0
		}
	}
	return weights
}
