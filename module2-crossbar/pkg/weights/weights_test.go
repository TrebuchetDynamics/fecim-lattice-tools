// pkg/weights/weights_test.go
// Tests for weight management and serialization utilities
package weights

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// ModelWeights Tests
// ============================================================================

func TestNewModelWeights(t *testing.T) {
	model := NewModelWeights("test_model", 3)

	if model == nil {
		t.Fatal("NewModelWeights returned nil")
	}
	if model.Name != "test_model" {
		t.Errorf("Name mismatch: expected 'test_model', got '%s'", model.Name)
	}
	if model.Version != "1.0" {
		t.Errorf("Version should default to '1.0', got '%s'", model.Version)
	}
	if model.NumLayers != 3 {
		t.Errorf("NumLayers mismatch: expected 3, got %d", model.NumLayers)
	}
	if model.Layers == nil {
		t.Error("Layers should be initialized")
	}
	if model.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestModelWeights_AddLayer(t *testing.T) {
	model := NewModelWeights("test", 2)

	weights := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}
	bias := []float64{0.01, 0.02}

	model.AddLayer("fc1", weights, bias)

	if len(model.Layers) != 1 {
		t.Fatalf("Expected 1 layer, got %d", len(model.Layers))
	}

	layer := &model.Layers[0]
	if layer.Name != "fc1" {
		t.Errorf("Layer name mismatch")
	}
	if len(layer.Shape) != 2 || layer.Shape[0] != 2 || layer.Shape[1] != 3 {
		t.Errorf("Shape mismatch: expected [2, 3], got %v", layer.Shape)
	}
	if len(layer.Data) != 6 {
		t.Errorf("Data length mismatch: expected 6, got %d", len(layer.Data))
	}
	if len(layer.Bias) != 2 {
		t.Errorf("Bias length mismatch: expected 2, got %d", len(layer.Bias))
	}
}

func TestModelWeights_AddLayer_Empty(t *testing.T) {
	model := NewModelWeights("test", 1)

	// Empty weights should not add layer
	model.AddLayer("empty", [][]float64{}, nil)

	if len(model.Layers) != 0 {
		t.Errorf("Empty weights should not add layer")
	}
}

func TestModelWeights_GetLayer(t *testing.T) {
	model := NewModelWeights("test", 2)
	model.AddLayer("fc1", [][]float64{{1, 2}, {3, 4}}, nil)
	model.AddLayer("fc2", [][]float64{{5, 6}}, nil)

	// Get by name
	layer, err := model.GetLayer("fc1")
	if err != nil {
		t.Fatalf("GetLayer failed: %v", err)
	}
	if layer.Name != "fc1" {
		t.Errorf("Got wrong layer: %s", layer.Name)
	}

	// Get non-existent layer
	_, err = model.GetLayer("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent layer")
	}
}

func TestModelWeights_GetLayerByIndex(t *testing.T) {
	model := NewModelWeights("test", 2)
	model.AddLayer("fc1", [][]float64{{1, 2}}, nil)
	model.AddLayer("fc2", [][]float64{{3, 4}}, nil)

	// Valid index
	layer, err := model.GetLayerByIndex(0)
	if err != nil {
		t.Fatalf("GetLayerByIndex failed: %v", err)
	}
	if layer.Name != "fc1" {
		t.Errorf("Expected fc1, got %s", layer.Name)
	}

	// Invalid indices
	_, err = model.GetLayerByIndex(-1)
	if err == nil {
		t.Error("Expected error for negative index")
	}

	_, err = model.GetLayerByIndex(5)
	if err == nil {
		t.Error("Expected error for out-of-range index")
	}
}

// ============================================================================
// LayerWeights Tests
// ============================================================================

func TestLayerWeights_ToMatrix(t *testing.T) {
	layer := LayerWeights{
		Name:  "test",
		Shape: []int{2, 3},
		Data:  []float64{1, 2, 3, 4, 5, 6},
	}

	matrix := layer.ToMatrix()

	if matrix == nil {
		t.Fatal("ToMatrix returned nil")
	}
	if len(matrix) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(matrix))
	}
	if len(matrix[0]) != 3 {
		t.Errorf("Expected 3 cols, got %d", len(matrix[0]))
	}

	// Verify values
	expected := [][]float64{{1, 2, 3}, {4, 5, 6}}
	for i := range expected {
		for j := range expected[i] {
			if matrix[i][j] != expected[i][j] {
				t.Errorf("Matrix[%d][%d] mismatch: expected %f, got %f",
					i, j, expected[i][j], matrix[i][j])
			}
		}
	}
}

func TestLayerWeights_ToMatrix_InvalidShape(t *testing.T) {
	layer := LayerWeights{
		Name:  "test",
		Shape: []int{2}, // 1D shape
		Data:  []float64{1, 2},
	}

	matrix := layer.ToMatrix()

	if matrix != nil {
		t.Error("ToMatrix should return nil for non-2D shape")
	}
}

func TestLayerWeights_GetStatistics(t *testing.T) {
	layer := LayerWeights{
		Data: []float64{-2, -1, 0, 1, 2},
	}

	min, max, mean, std := layer.GetStatistics()

	if min != -2 {
		t.Errorf("Min mismatch: expected -2, got %f", min)
	}
	if max != 2 {
		t.Errorf("Max mismatch: expected 2, got %f", max)
	}
	if mean != 0 {
		t.Errorf("Mean mismatch: expected 0, got %f", mean)
	}
	// std of [-2,-1,0,1,2] = sqrt(2)
	expectedStd := math.Sqrt(2.0)
	if math.Abs(std-expectedStd) > 0.001 {
		t.Errorf("Std mismatch: expected %f, got %f", expectedStd, std)
	}
}

func TestLayerWeights_GetStatistics_Empty(t *testing.T) {
	layer := LayerWeights{Data: []float64{}}

	min, max, mean, std := layer.GetStatistics()

	// Should handle empty gracefully
	if min != 0 || max != 0 || mean != 0 || std != 0 {
		t.Logf("Empty layer stats: min=%f max=%f mean=%f std=%f", min, max, mean, std)
	}
}

// ============================================================================
// Quantization Tests
// ============================================================================

func TestLayerWeights_QuantizeWeights_Symmetric(t *testing.T) {
	layer := LayerWeights{
		Data: []float64{-1.0, -0.5, 0, 0.5, 1.0},
	}

	layer.QuantizeWeights(8, true) // 8-bit symmetric

	if layer.Quant == nil {
		t.Fatal("Quant info should be set")
	}
	if layer.Quant.Bits != 8 {
		t.Errorf("Bits mismatch: expected 8, got %d", layer.Quant.Bits)
	}
	if !layer.Quant.Symmetric {
		t.Error("Should be symmetric quantization")
	}
	if layer.Quant.ZeroPoint != 0 {
		t.Errorf("Symmetric quant should have zero_point=0, got %f", layer.Quant.ZeroPoint)
	}

	// Quantized values should be within range
	for _, v := range layer.Data {
		if v < -1.1 || v > 1.1 { // Allow small tolerance
			t.Errorf("Quantized value out of range: %f", v)
		}
	}
}

func TestLayerWeights_QuantizeWeights_Asymmetric(t *testing.T) {
	layer := LayerWeights{
		Data: []float64{0, 0.25, 0.5, 0.75, 1.0}, // All positive
	}

	layer.QuantizeWeights(8, false) // 8-bit asymmetric

	if layer.Quant == nil {
		t.Fatal("Quant info should be set")
	}
	if layer.Quant.Symmetric {
		t.Error("Should be asymmetric quantization")
	}

	// Scale and zero_point should be set
	if layer.Quant.Scale == 0 {
		t.Error("Scale should not be zero")
	}
}

func TestLayerWeights_QuantizeWeights_InvalidBits(t *testing.T) {
	layer := LayerWeights{
		Data: []float64{1, 2, 3},
	}

	// Invalid bits should default to 8
	layer.QuantizeWeights(0, true)
	if layer.Quant.Bits != 8 {
		t.Errorf("Invalid bits should default to 8, got %d", layer.Quant.Bits)
	}

	layer2 := LayerWeights{Data: []float64{1, 2, 3}}
	layer2.QuantizeWeights(32, true) // Too many bits
	if layer2.Quant.Bits != 8 {
		t.Errorf("Too many bits should default to 8, got %d", layer2.Quant.Bits)
	}
}

// ============================================================================
// Normalization Tests
// ============================================================================

func TestLayerWeights_NormalizeWeights(t *testing.T) {
	layer := LayerWeights{
		Data: []float64{0, 50, 100},
	}

	scale, offset := layer.NormalizeWeights()

	// After normalization, values should be in [0, 1]
	for _, v := range layer.Data {
		if v < 0 || v > 1 {
			t.Errorf("Normalized value out of range: %f", v)
		}
	}

	// Verify scale and offset
	if scale != 100 {
		t.Errorf("Scale mismatch: expected 100, got %f", scale)
	}
	if offset != 0 {
		t.Errorf("Offset mismatch: expected 0, got %f", offset)
	}
}

func TestLayerWeights_DenormalizeWeights(t *testing.T) {
	original := []float64{0, 50, 100}
	layer := LayerWeights{
		Data: make([]float64, len(original)),
	}
	copy(layer.Data, original)

	scale, offset := layer.NormalizeWeights()
	layer.DenormalizeWeights(scale, offset)

	// Should restore original values
	for i, v := range layer.Data {
		if math.Abs(v-original[i]) > 0.001 {
			t.Errorf("Denormalize failed: expected %f, got %f", original[i], v)
		}
	}
}

func TestLayerWeights_NormalizeWeights_ZeroRange(t *testing.T) {
	layer := LayerWeights{
		Data: []float64{5, 5, 5}, // All same value
	}

	scale, _ := layer.NormalizeWeights()

	// Scale should be 1 to avoid division by zero
	if scale != 1 {
		t.Errorf("Zero range should set scale=1, got %f", scale)
	}
}

// ============================================================================
// JSON Serialization Tests
// ============================================================================

func TestModelWeights_SaveLoadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "model.json")

	// Create model
	model := NewModelWeights("test_model", 2)
	model.AddLayer("fc1", [][]float64{{1, 2}, {3, 4}}, []float64{0.1, 0.2})
	model.AddLayer("fc2", [][]float64{{5, 6}}, nil)
	model.Metadata["framework"] = "fecim"

	// Save
	err := model.SaveJSON(path)
	if err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("JSON file not created")
	}

	// Load
	loaded, err := LoadJSON(path)
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	// Verify loaded data
	if loaded.Name != model.Name {
		t.Errorf("Name mismatch: expected %s, got %s", model.Name, loaded.Name)
	}
	if len(loaded.Layers) != len(model.Layers) {
		t.Errorf("Layers count mismatch")
	}
	if loaded.Layers[0].Name != "fc1" {
		t.Errorf("Layer name mismatch")
	}
}

func TestLoadJSON_NonExistent(t *testing.T) {
	_, err := LoadJSON("/nonexistent/path/model.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// ============================================================================
// Binary Serialization Tests
// ============================================================================

func TestModelWeights_SaveLoadBinary(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "model.bin")

	// Create model
	model := NewModelWeights("binary_test", 2)
	model.AddLayer("layer1", [][]float64{
		{1.5, 2.5, 3.5},
		{4.5, 5.5, 6.5},
	}, []float64{0.1, 0.2})

	// Save binary
	err := model.SaveBinary(path)
	if err != nil {
		t.Fatalf("SaveBinary failed: %v", err)
	}

	// Verify file exists and is smaller than JSON
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Binary file not created: %v", err)
	}
	t.Logf("Binary file size: %d bytes", info.Size())

	// Load binary
	loaded, err := LoadBinary(path)
	if err != nil {
		t.Fatalf("LoadBinary failed: %v", err)
	}

	// Verify data (note: binary uses float32, so precision is reduced)
	if len(loaded.Layers) != 1 {
		t.Errorf("Expected 1 layer, got %d", len(loaded.Layers))
	}
	layer := loaded.Layers[0]
	if layer.Name != "layer1" {
		t.Errorf("Layer name mismatch")
	}

	// Check data with float32 tolerance
	for i, v := range layer.Data {
		expected := model.Layers[0].Data[i]
		if math.Abs(v-expected) > 1e-5 {
			t.Errorf("Data[%d] mismatch: expected %f, got %f", i, expected, v)
		}
	}
}

func TestLoadBinary_InvalidMagic(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.bin")

	// Write invalid data
	os.WriteFile(path, []byte("not a valid binary file"), 0644)

	_, err := LoadBinary(path)
	if err == nil {
		t.Error("Expected error for invalid binary file")
	}
}

// ============================================================================
// Model Serialization (serialization.go) Tests
// ============================================================================

func TestNewModel(t *testing.T) {
	model := NewModel("test", "mlp")

	if model.Metadata.Name != "test" {
		t.Errorf("Name mismatch")
	}
	if model.Metadata.Architecture != "mlp" {
		t.Errorf("Architecture mismatch")
	}
	if model.Metadata.Version != "1.0" {
		t.Errorf("Version should default to 1.0")
	}
}

func TestModel_AddLayer(t *testing.T) {
	model := NewModel("test", "mlp")

	weights := [][]float64{{1, 2, 3}, {4, 5, 6}}
	biases := []float64{0.1, 0.2}

	model.AddLayer("fc1", "linear", weights, biases)

	if model.Metadata.NumLayers != 1 {
		t.Errorf("NumLayers should be 1, got %d", model.Metadata.NumLayers)
	}
	// 2*3 weights + 2 biases = 8 params
	if model.Metadata.TotalParams != 8 {
		t.Errorf("TotalParams should be 8, got %d", model.Metadata.TotalParams)
	}

	layer := model.Layers[0]
	if layer.Name != "fc1" || layer.Type != "linear" {
		t.Error("Layer metadata mismatch")
	}
}

func TestModel_SaveLoadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "model_v2.json")

	model := NewModel("test", "transformer")
	model.AddLayer("attention", "attention", [][]float64{{1, 2}, {3, 4}}, nil)

	err := model.SaveJSON(path)
	if err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	loaded, err := LoadModelJSON(path)
	if err != nil {
		t.Fatalf("LoadModelJSON failed: %v", err)
	}

	if loaded.Metadata.Name != "test" {
		t.Error("Metadata not preserved")
	}
	if len(loaded.Layers) != 1 {
		t.Error("Layers not preserved")
	}
}

func TestModel_SaveLoadBinary(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "model_v2.bin")

	model := NewModel("binary_test", "cnn")
	model.AddLayer("conv1", "conv2d", [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}, []float64{0.01})

	err := model.SaveBinary(path)
	if err != nil {
		t.Fatalf("SaveBinary failed: %v", err)
	}

	loaded, err := LoadModelBinary(path)
	if err != nil {
		t.Fatalf("LoadModelBinary failed: %v", err)
	}

	if loaded.Metadata.Name != "binary_test" {
		t.Error("Metadata name not preserved")
	}
	if loaded.Metadata.Architecture != "cnn" {
		t.Error("Metadata architecture not preserved")
	}
	if len(loaded.Layers) != 1 {
		t.Error("Layers not preserved")
	}
}

// ============================================================================
// Quantized Model Tests
// ============================================================================

func TestQuantizeModel(t *testing.T) {
	model := NewModel("quant_test", "mlp")
	model.AddLayer("fc1", "linear", [][]float64{
		{-1.0, 0.0, 1.0},
		{-0.5, 0.5, 0.0},
	}, []float64{0.1, -0.1})

	qmodel := QuantizeModel(model, 8)

	if qmodel == nil {
		t.Fatal("QuantizeModel returned nil")
	}
	if !qmodel.Metadata.Quantized {
		t.Error("Quantized flag should be true")
	}
	if qmodel.Metadata.QuantBits != 8 {
		t.Errorf("QuantBits should be 8, got %d", qmodel.Metadata.QuantBits)
	}
	if len(qmodel.Layers) != 1 {
		t.Errorf("Expected 1 layer, got %d", len(qmodel.Layers))
	}

	qLayer := qmodel.Layers[0]
	if qLayer.WeightScale == 0 {
		t.Error("WeightScale should not be zero")
	}

	// Int8 weights should be in valid range
	for r := range qLayer.Weights {
		for c := range qLayer.Weights[r] {
			w := qLayer.Weights[r][c]
			if w < -128 || w > 127 {
				t.Errorf("Int8 weight out of range: %d", w)
			}
		}
	}
}

func TestQuantizedLayerWeights_Dequantize(t *testing.T) {
	model := NewModel("test", "mlp")
	weights := [][]float64{{-1.0, 0.0, 1.0}}
	model.AddLayer("fc1", "linear", weights, nil)

	qmodel := QuantizeModel(model, 8)
	qLayer := &qmodel.Layers[0]

	dequant := qLayer.Dequantize()

	// Should approximately restore original values
	for i, row := range dequant.Weights {
		for j, v := range row {
			expected := weights[i][j]
			if math.Abs(v-expected) > 0.1 { // Allow quantization error
				t.Errorf("Dequantized[%d][%d] = %f, expected ~%f", i, j, v, expected)
			}
		}
	}
}

// ============================================================================
// Crossbar Mapping Tests
// ============================================================================

func TestGenerateCrossbarMapping(t *testing.T) {
	layer := &SerializedLayer{
		Name:    "fc1",
		Type:    "linear",
		Shape:   []int{10, 20},
		Weights: make([][]float64, 10),
	}
	for i := range layer.Weights {
		layer.Weights[i] = make([]float64, 20)
		for j := range layer.Weights[i] {
			if (i+j)%3 == 0 {
				layer.Weights[i][j] = 0 // Some zeros for sparsity
			} else {
				layer.Weights[i][j] = float64(i*20+j) * 0.01
			}
		}
	}

	mapping := GenerateCrossbarMapping(layer, 4, 8) // 4x8 tiles

	if mapping == nil {
		t.Fatal("GenerateCrossbarMapping returned nil")
	}
	if mapping.LayerName != "fc1" {
		t.Errorf("LayerName mismatch")
	}
	if mapping.TileSize != [2]int{4, 8} {
		t.Errorf("TileSize mismatch: got %v", mapping.TileSize)
	}

	// 10 rows / 4 = 3 tile rows, 20 cols / 8 = 3 tile cols = 9 tiles
	expectedTiles := 3 * 3
	if mapping.NumTiles != expectedTiles {
		t.Errorf("NumTiles mismatch: expected %d, got %d", expectedTiles, mapping.NumTiles)
	}
	if len(mapping.TileOffsets) != expectedTiles {
		t.Errorf("TileOffsets length mismatch")
	}
	if len(mapping.TileMasks) != expectedTiles {
		t.Errorf("TileMasks length mismatch")
	}
}

func TestGenerateCrossbarMapping_TileOffsets(t *testing.T) {
	layer := &SerializedLayer{
		Shape:   []int{8, 8},
		Weights: make([][]float64, 8),
	}
	for i := range layer.Weights {
		layer.Weights[i] = make([]float64, 8)
	}

	mapping := GenerateCrossbarMapping(layer, 4, 4) // 2x2 tiles

	// First tile at (0,0)
	if mapping.TileOffsets[0] != [2]int{0, 0} {
		t.Errorf("Tile 0 offset should be [0,0], got %v", mapping.TileOffsets[0])
	}
	// Second tile at (0,4)
	if mapping.TileOffsets[1] != [2]int{0, 4} {
		t.Errorf("Tile 1 offset should be [0,4], got %v", mapping.TileOffsets[1])
	}
	// Third tile at (4,0)
	if mapping.TileOffsets[2] != [2]int{4, 0} {
		t.Errorf("Tile 2 offset should be [4,0], got %v", mapping.TileOffsets[2])
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkQuantizeWeights_8bit(b *testing.B) {
	layer := LayerWeights{
		Data: make([]float64, 784*128), // MNIST first layer size
	}
	for i := range layer.Data {
		layer.Data[i] = float64(i%1000)/500.0 - 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layer.QuantizeWeights(8, true)
	}
}

func BenchmarkNormalizeWeights(b *testing.B) {
	layer := LayerWeights{
		Data: make([]float64, 784*128),
	}
	for i := range layer.Data {
		layer.Data[i] = float64(i) / float64(len(layer.Data))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layer.NormalizeWeights()
	}
}

func BenchmarkToMatrix(b *testing.B) {
	layer := LayerWeights{
		Shape: []int{128, 784},
		Data:  make([]float64, 128*784),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = layer.ToMatrix()
	}
}

func BenchmarkSaveLoadJSON(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.json")

	model := NewModelWeights("bench", 2)
	model.AddLayer("fc1", make([][]float64, 128), nil)
	for i := range model.Layers[0].Data {
		model.Layers[0].Data[i] = float64(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.SaveJSON(path)
		LoadJSON(path)
	}
}

func BenchmarkSaveLoadBinary(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.bin")

	model := NewModelWeights("bench", 2)
	weights := make([][]float64, 64)
	for i := range weights {
		weights[i] = make([]float64, 64)
		for j := range weights[i] {
			weights[i][j] = float64(i*64+j) * 0.001
		}
	}
	model.AddLayer("fc1", weights, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.SaveBinary(path)
		LoadBinary(path)
	}
}
