package weights

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestModule2WeightsE2ESerializationQuantizationMappingWorkflow(t *testing.T) {
	model := NewModel("module2-e2e", "crossbar-mlp")
	model.Metadata.Custom["source"] = "e2e"
	model.AddLayer("hidden", "linear", [][]float64{{0.0, 0.25, -0.5, 0.75, 1.0}, {0.1, 0.0, 0.3, -0.2, 0.5}, {0.9, -0.4, 0.0, 0.2, -0.1}}, []float64{0.01, -0.02, 0.03})
	model.AddLayer("output", "linear", [][]float64{{0.2, 0.0, -0.2}, {0.0, 0.4, 0.8}}, []float64{0.05, -0.05})
	if model.Metadata.NumLayers != 2 || model.Metadata.TotalParams != 26 {
		t.Fatalf("metadata after AddLayer = layers %d params %d", model.Metadata.NumLayers, model.Metadata.TotalParams)
	}

	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "model.json")
	binPath := filepath.Join(dir, "model.bin.gz")
	if err := model.SaveJSON(jsonPath); err != nil {
		t.Fatalf("SaveJSON error = %v", err)
	}
	if err := model.SaveBinary(binPath); err != nil {
		t.Fatalf("SaveBinary error = %v", err)
	}
	loadedJSON, err := LoadModelJSON(jsonPath)
	if err != nil {
		t.Fatalf("LoadModelJSON error = %v", err)
	}
	loadedBinary, err := LoadModelBinary(binPath)
	if err != nil {
		t.Fatalf("LoadModelBinary error = %v", err)
	}
	assertModelE2EEquivalent(t, model, loadedJSON)
	assertModelE2EEquivalent(t, model, loadedBinary)

	q := QuantizeModel(model, 8)
	if !q.Metadata.Quantized || q.Metadata.QuantBits != 8 || len(q.Layers) != len(model.Layers) {
		t.Fatalf("quantized metadata invalid: %+v layers=%d", q.Metadata, len(q.Layers))
	}
	for i := range q.Layers {
		deq := q.Layers[i].Dequantize()
		if deq.Name != model.Layers[i].Name || deq.Type != model.Layers[i].Type || len(deq.Weights) != len(model.Layers[i].Weights) {
			t.Fatalf("dequantized layer %d metadata/shape mismatch", i)
		}
		if q.Layers[i].WeightScale <= 0 || (len(q.Layers[i].Biases) > 0 && q.Layers[i].BiasScale <= 0) {
			t.Fatalf("quantized layer %d invalid scales: %+v", i, q.Layers[i])
		}
	}

	mapping := GenerateCrossbarMapping(&model.Layers[0], 2, 3)
	if mapping.LayerName != "hidden" || mapping.TileSize != [2]int{2, 3} || mapping.NumTiles != 4 || len(mapping.TileOffsets) != 4 || len(mapping.TileMasks) != 4 {
		t.Fatalf("mapping metadata invalid: %+v", mapping)
	}
	wantOffsets := [][2]int{{0, 0}, {0, 3}, {2, 0}, {2, 3}}
	for i, want := range wantOffsets {
		if mapping.TileOffsets[i] != want {
			t.Fatalf("offset[%d] = %v, want %v", i, mapping.TileOffsets[i], want)
		}
	}
	// In tile (0,0), source hidden[0][0] is zero and hidden[0][1] is non-zero.
	if !mapping.TileMasks[0][0][0] || mapping.TileMasks[0][0][1] {
		t.Fatalf("tile mask did not preserve sparse/non-sparse cells: %v", mapping.TileMasks[0])
	}
	// Last tile includes padding at row 3 and col 5.
	if !mapping.TileMasks[3][1][2] {
		t.Fatalf("last tile padding mask missing: %v", mapping.TileMasks[3])
	}
}

func TestModule2WeightsE2EInvalidSerializationInputs(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadModelJSON(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("LoadModelJSON missing file returned nil error")
	}
	badJSON := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badJSON, []byte(`{"metadata":`), 0644); err != nil {
		t.Fatalf("write bad json: %v", err)
	}
	if _, err := LoadModelJSON(badJSON); err == nil {
		t.Fatal("LoadModelJSON malformed file returned nil error")
	}
	badBinary := filepath.Join(dir, "bad.bin")
	if err := os.WriteFile(badBinary, []byte("not-gzip"), 0644); err != nil {
		t.Fatalf("write bad binary: %v", err)
	}
	if _, err := LoadModelBinary(badBinary); err == nil {
		t.Fatal("LoadModelBinary malformed file returned nil error")
	}
	model := NewModel("bad-output", "mlp")
	if err := model.SaveJSON(filepath.Join(dir, "missing-parent", "model.json")); err == nil {
		t.Fatal("SaveJSON missing parent returned nil error")
	}
	if err := model.SaveBinary(filepath.Join(dir, "missing-parent", "model.bin")); err == nil {
		t.Fatal("SaveBinary missing parent returned nil error")
	}
}

func assertModelE2EEquivalent(t *testing.T, want, got *Model) {
	t.Helper()
	if got.Metadata.Name != want.Metadata.Name || got.Metadata.Architecture != want.Metadata.Architecture || got.Metadata.NumLayers != want.Metadata.NumLayers || got.Metadata.TotalParams != want.Metadata.TotalParams {
		t.Fatalf("metadata mismatch\nwant=%+v\ngot=%+v", want.Metadata, got.Metadata)
	}
	if len(got.Layers) != len(want.Layers) {
		t.Fatalf("layer count = %d, want %d", len(got.Layers), len(want.Layers))
	}
	for l := range want.Layers {
		if got.Layers[l].Name != want.Layers[l].Name || got.Layers[l].Type != want.Layers[l].Type || len(got.Layers[l].Weights) != len(want.Layers[l].Weights) || len(got.Layers[l].Biases) != len(want.Layers[l].Biases) {
			t.Fatalf("layer %d metadata mismatch", l)
		}
		for r := range want.Layers[l].Weights {
			for c := range want.Layers[l].Weights[r] {
				if math.Abs(got.Layers[l].Weights[r][c]-want.Layers[l].Weights[r][c]) > 1e-12 {
					t.Fatalf("layer %d weight[%d][%d] = %g, want %g", l, r, c, got.Layers[l].Weights[r][c], want.Layers[l].Weights[r][c])
				}
			}
		}
		for i := range want.Layers[l].Biases {
			if math.Abs(got.Layers[l].Biases[i]-want.Layers[l].Biases[i]) > 1e-12 {
				t.Fatalf("layer %d bias[%d] = %g, want %g", l, i, got.Layers[l].Biases[i], want.Layers[l].Biases[i])
			}
		}
	}
}
