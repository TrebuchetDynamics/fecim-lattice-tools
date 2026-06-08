package crossbar

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestModule2CrossbarE2ETilingMemTorchExportWorkflow(t *testing.T) {
	weights := [][]float64{
		{0.0, 0.2, 0.4, 0.6, 0.8},
		{1.0, 0.1, 0.3, 0.5, 0.7},
		{0.9, 0.0, 0.25, 0.5, 0.75},
		{0.11, 0.22, 0.33, 0.44, 0.55},
		{0.66, 0.77, 0.88, 0.99, 0.12},
	}
	cfgs := []TilingConfig{
		{MaxRows: 2, MaxCols: 3, Overlap: 0, Padding: "zero"},
		{MaxRows: 3, MaxCols: 2, Overlap: 0, Padding: "replicate"},
	}
	input := []float64{1, 0.5, 0.25, 0.75, 0.1}
	wantDense := denseMVMNoNormalizeE2E(weights, input)
	for _, cfg := range cfgs {
		t.Run(cfg.Padding, func(t *testing.T) {
			tiled, err := TileWeightMatrix(weights, cfg)
			if err != nil {
				t.Fatalf("TileWeightMatrix error = %v", err)
			}
			if tiled.OrigShape != [2]int{5, 5} || tiled.TotalTiles() != tiled.TileGrid[0]*tiled.TileGrid[1] || tiled.Efficiency() != 1 {
				t.Fatalf("tiling metadata invalid: shape=%v grid=%v total=%d eff=%g", tiled.OrigShape, tiled.TileGrid, tiled.TotalTiles(), tiled.Efficiency())
			}
			for i, tile := range tiled.Tiles {
				tr, tc := tiled.TileIndex(i)
				if tr < 0 || tc < 0 || tr >= tiled.TileGrid[0] || tc >= tiled.TileGrid[1] {
					t.Fatalf("TileIndex(%d) = %d,%d outside grid %v", i, tr, tc, tiled.TileGrid)
				}
				if tile.RowOffset < 0 || tile.ColOffset < 0 || tile.Rows <= 0 || tile.Cols <= 0 || len(tile.Weights) != tile.Rows {
					t.Fatalf("tile %d invalid: %+v", i, tile)
				}
			}
			got, err := tiled.TiledMVM(input, func(w [][]float64, inp []float64) ([]float64, error) {
				return denseMVMNoNormalizeE2E(w, inp), nil
			})
			if err != nil {
				t.Fatalf("TiledMVM error = %v", err)
			}
			assertFloatVectorCloseE2E(t, got, wantDense, 1e-12)

			params := ExportToMemTorch(&Config{Rows: 5, Cols: 5, NoiseLevel: 0.03, ADCBits: 8, DACBits: 8})
			if params.ROn <= 0 || params.ROff <= params.ROn || params.PNoiseStd != 0.03 || params.ROnSigma <= 0 || params.ROffSigma <= 0 {
				t.Fatalf("MemTorch params invalid: %+v", params)
			}
			imported := ImportFromMemTorch(params)
			if math.Abs(imported.NoiseLevel-0.03) > 1e-12 {
				t.Fatalf("imported noise = %.12f, want 0.03", imported.NoiseLevel)
			}
			data, err := ExportWeightsAsMemTorchJSON(weights, params)
			if err != nil {
				t.Fatalf("ExportWeightsAsMemTorchJSON error = %v", err)
			}
			var exported MemTorchWeightExport
			if err := json.Unmarshal(data, &exported); err != nil {
				t.Fatalf("decode MemTorch export: %v\n%s", err, data)
			}
			if exported.Rows != 5 || exported.Cols != 5 || len(exported.Weights) != 5 || exported.DeviceParams.ROn != params.ROn || exported.Metadata.Tool == "" || exported.Metadata.Disclaimer == "" {
				t.Fatalf("MemTorch export contract invalid: %+v", exported)
			}
		})
	}
}

func TestModule2CrossbarE2ETilingMemTorchInvalidMatrix(t *testing.T) {
	invalidConfigs := []TilingConfig{{MaxRows: 0, MaxCols: 2}, {MaxRows: 2, MaxCols: 0}, {MaxRows: 2, MaxCols: 2, Overlap: -1}, {MaxRows: 2, MaxCols: 2, Overlap: 2}, {MaxRows: 2, MaxCols: 2, Padding: "mystery"}}
	for _, cfg := range invalidConfigs {
		if _, err := TileWeightMatrix([][]float64{{1}}, cfg); err == nil {
			t.Fatalf("TileWeightMatrix invalid config %+v returned nil error", cfg)
		}
	}
	if _, err := TileWeightMatrix(nil, TilingConfig{MaxRows: 2, MaxCols: 2}); err == nil {
		t.Fatal("TileWeightMatrix empty matrix returned nil error")
	}
	if _, err := TileWeightMatrix([][]float64{{1, 2}, {3}}, TilingConfig{MaxRows: 2, MaxCols: 2}); err == nil {
		t.Fatal("TileWeightMatrix jagged matrix returned nil error")
	}
	tiled, err := TileWeightMatrix([][]float64{{1, 2}, {3, 4}}, TilingConfig{MaxRows: 1, MaxCols: 1})
	if err != nil {
		t.Fatalf("valid TileWeightMatrix error = %v", err)
	}
	if _, err := tiled.ReconstructOutput(nil); err == nil {
		t.Fatal("ReconstructOutput wrong tile count returned nil error")
	}
	if _, err := tiled.ReconstructOutput([][]float64{{1}, {2}, {3}, {4, 5}}); err == nil {
		t.Fatal("ReconstructOutput wrong tile output length returned nil error")
	}
	if _, err := tiled.TiledMVM([]float64{1}, func(w [][]float64, inp []float64) ([]float64, error) { return nil, nil }); err == nil {
		t.Fatal("TiledMVM wrong input length returned nil error")
	}

	params := ImportFromMemTorch(MemTorchDeviceParams{ROn: 0, ROff: 1})
	if params.Rows != 0 || params.Cols != 0 || params.NoiseLevel != 0 {
		t.Fatalf("invalid MemTorch import should return zero config, got %+v", params)
	}
	invalidWeights := [][][]float64{{}, {{0.1}, {0.2, 0.3}}, {{math.NaN()}}, {{-0.1}}, {{1.1}}}
	for _, matrix := range invalidWeights {
		_, err := ExportWeightsAsMemTorchJSON(matrix, MemTorchDeviceParams{ROn: 1, ROff: 2})
		if err == nil {
			t.Fatalf("ExportWeightsAsMemTorchJSON(%v) returned nil error", matrix)
		}
		if !strings.Contains(err.Error(), "weight") && !strings.Contains(err.Error(), "row") {
			t.Fatalf("unexpected MemTorch error for %v: %v", matrix, err)
		}
	}
}

func denseMVMNoNormalizeE2E(weights [][]float64, input []float64) []float64 {
	out := make([]float64, len(weights))
	for r := range weights {
		for c := range input {
			out[r] += weights[r][c] * input[c]
		}
	}
	return out
}

func assertFloatVectorCloseE2E(t *testing.T, got, want []float64, tol float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("vector length = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > tol {
			t.Fatalf("vector[%d] = %.12f, want %.12f", i, got[i], want[i])
		}
	}
}
