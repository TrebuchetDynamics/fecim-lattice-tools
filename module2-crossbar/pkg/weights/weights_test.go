// pkg/weights/weights_test.go
// Tests for weight management and serialization utilities
package weights

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

type boundedReadRequestReader struct {
	reader  io.Reader
	maxRead int
	largest int
}

func (reader *boundedReadRequestReader) Read(p []byte) (int, error) {
	if len(p) > reader.maxRead {
		return 0, errors.New("read request exceeds scratch bound")
	}
	if len(p) > reader.largest {
		reader.largest = len(p)
	}
	return reader.reader.Read(p)
}

func TestReadFloat64SliceUsesBoundedScratchReads(t *testing.T) {
	want := make([]float64, 8193)
	for i := range want {
		want[i] = float64(i) - 4096.25
	}
	var encoded bytes.Buffer
	if err := binary.Write(&encoded, binary.LittleEndian, want); err != nil {
		t.Fatal(err)
	}
	reader := &boundedReadRequestReader{reader: bytes.NewReader(encoded.Bytes()), maxRead: 32 << 10}
	got := make([]float64, len(want))
	if err := readFloat64Slice(reader, got); err != nil {
		t.Fatalf("readFloat64Slice error=%v", err)
	}
	if reader.largest == 0 || reader.largest > reader.maxRead {
		t.Fatalf("largest read request=%d want 1..%d", reader.largest, reader.maxRead)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("readFloat64Slice values differ from encoded values")
	}
}

func writeModelBinaryFixture(t *testing.T, writePayload func(io.Writer)) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "model.bin")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	writePayload(gz)
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeModelHeader(t *testing.T, w io.Writer, metadata ModelMetadata, layers uint32) {
	t.Helper()
	meta, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []uint32{0x4D4F444C, 1, uint32(len(meta))} {
		if err := binary.Write(w, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := w.Write(meta); err != nil {
		t.Fatal(err)
	}
	if err := binary.Write(w, binary.LittleEndian, layers); err != nil {
		t.Fatal(err)
	}
}

func writeLayerPrefix(t *testing.T, w io.Writer, name, layerType string, shape []uint32) {
	t.Helper()
	for _, value := range []string{name, layerType} {
		if err := binary.Write(w, binary.LittleEndian, uint32(len(value))); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(shape))); err != nil {
		t.Fatal(err)
	}
	for _, dim := range shape {
		if err := binary.Write(w, binary.LittleEndian, dim); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSaveBinaryRejectsRaggedWeightsWithoutTouchingDestination(t *testing.T) {
	model := NewModel("ragged", "linear")
	model.AddLayer("fc1", "linear", [][]float64{{1, 2}, {3}}, nil)
	path := filepath.Join(t.TempDir(), "ragged.bin")
	if err := os.WriteFile(path, []byte("preserve-me"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := model.SaveBinary(path)
	if err == nil || !strings.Contains(err.Error(), "rectangular") {
		t.Fatalf("SaveBinary error=%v want rectangularity error", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "preserve-me" {
		t.Fatalf("destination changed on validation failure: %q", got)
	}
}

func TestMarshalBoundedModelMetadataRejectsOversizeBeforeMarshal(t *testing.T) {
	metadata := ModelMetadata{
		Name:   "oversized",
		Custom: map[string]string{"payload": strings.Repeat("x", int(maxModelMetadataBytes))},
	}
	called := false
	_, err := marshalBoundedModelMetadata(metadata, func(any) ([]byte, error) {
		called = true
		return nil, errors.New("marshal must not be called")
	})
	if err == nil || !strings.Contains(err.Error(), "metadata length") {
		t.Fatalf("marshalBoundedModelMetadata error=%v want metadata length error", err)
	}
	if called {
		t.Fatal("marshal function was called for oversized metadata")
	}
}

func TestModelMetadataEncodedSizeMatchesJSONMarshalWithoutAllocations(t *testing.T) {
	invalidUTF8 := string([]byte{'x', 0xff, 'y'})
	minInt := -int(^uint(0)>>1) - 1
	tests := []struct {
		name     string
		metadata ModelMetadata
	}{
		{name: "minimal", metadata: ModelMetadata{}},
		{name: "escaping-and-optionals", metadata: ModelMetadata{
			Name:         "quote\" slash\\ controls\x00\b\f\n\r\t",
			Version:      "<html>&",
			Architecture: "unicode\u2028\u2029é/" + invalidUTF8,
			NumLayers:    minInt,
			TotalParams:  int(^uint(0) >> 1),
			Quantized:    true,
			QuantBits:    -8,
			Custom: map[string]string{
				"key\"<":    "value&\u2028",
				invalidUTF8: "control\x01",
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.metadata)
			if err != nil {
				t.Fatal(err)
			}
			size, err := modelMetadataEncodedSize(test.metadata)
			if err != nil {
				t.Fatal(err)
			}
			if size != uint64(len(encoded)) {
				t.Fatalf("modelMetadataEncodedSize=%d json.Marshal length=%d\njson=%s", size, len(encoded), encoded)
			}
			if allocations := testing.AllocsPerRun(100, func() {
				if _, err := modelMetadataEncodedSize(test.metadata); err != nil {
					panic(err)
				}
			}); allocations != 0 {
				t.Fatalf("modelMetadataEncodedSize allocations=%g want 0", allocations)
			}
		})
	}
}

func TestSaveBinaryRejectsOversizedMetadataWithoutTouchingDestination(t *testing.T) {
	model := NewModel("oversized", "linear")
	model.Metadata.Custom["payload"] = strings.Repeat("x", int(maxModelMetadataBytes))
	path := filepath.Join(t.TempDir(), "model.bin")
	if err := os.WriteFile(path, []byte("preserve-me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := model.SaveBinary(path); err == nil || !strings.Contains(err.Error(), "metadata length") {
		t.Fatalf("SaveBinary error=%v want metadata length error", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "preserve-me" {
		t.Fatalf("destination changed on validation failure: %q", got)
	}
}

func TestSaveBinaryRejectsCumulativeRetainedAllocationWithoutTouchingDestination(t *testing.T) {
	model := NewModel("retained-budget", "linear")
	for i := 0; i < 3; i++ {
		rows := make([][]float64, 16)
		for row := range rows {
			rows[row] = make([]float64, 0)
		}
		model.AddLayer("z", "linear", rows, nil)
	}
	path := filepath.Join(t.TempDir(), "model.bin")
	if err := os.WriteFile(path, []byte("preserve-me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := model.saveBinaryWithRetainedLimit(path, 2800); err == nil || !strings.Contains(err.Error(), "retained model allocation") {
		t.Fatalf("saveBinaryWithRetainedLimit error=%v want retained allocation error", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "preserve-me" {
		t.Fatalf("destination changed on retained allocation validation failure: %q", got)
	}
}

func TestSaveBinaryRejectsInvalidModelWithoutTouchingDestination(t *testing.T) {
	tests := []struct {
		name  string
		label string
		model func() *Model
	}{
		{name: "layer-count", label: "layer count", model: func() *Model {
			model := NewModel("layers", "linear")
			model.Layers = make([]SerializedLayer, int(maxModelLayers)+1)
			return model
		}},
		{name: "name-length", label: "layer name length", model: func() *Model {
			model := NewModel("name", "linear")
			model.Layers = []SerializedLayer{{Name: strings.Repeat("x", int(maxModelStringBytes)+1)}}
			return model
		}},
		{name: "type-length", label: "layer type length", model: func() *Model {
			model := NewModel("type", "linear")
			model.Layers = []SerializedLayer{{Type: strings.Repeat("x", int(maxModelStringBytes)+1)}}
			return model
		}},
		{name: "shape-count", label: "shape dimension count", model: func() *Model {
			model := NewModel("shape", "linear")
			model.Layers = []SerializedLayer{{Shape: make([]int, int(maxModelShapeDims)+1)}}
			return model
		}},
		{name: "shape-dimension", label: "shape dimension", model: func() *Model {
			model := NewModel("shape", "linear")
			model.Layers = []SerializedLayer{{Shape: []int{-1}}}
			return model
		}},
		{name: "bias-length", label: "bias length", model: func() *Model {
			model := NewModel("bias", "linear")
			model.Layers = []SerializedLayer{{Biases: make([]float64, int(maxModelBiases)+1)}}
			return model
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "model.bin")
			if err := os.WriteFile(path, []byte("preserve-me"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := test.model().SaveBinary(path); err == nil || !strings.Contains(err.Error(), test.label) {
				t.Fatalf("SaveBinary error=%v want %s error", err, test.label)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != "preserve-me" {
				t.Fatalf("destination changed on validation failure: %q", got)
			}
		})
	}
}

func TestLoadModelBinaryRejectsOversizedMetadataLength(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.bin")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	_ = binary.Write(gz, binary.LittleEndian, uint32(0x4D4F444C))
	_ = binary.Write(gz, binary.LittleEndian, uint32(1))
	_ = binary.Write(gz, binary.LittleEndian, uint32((1<<20)+1))
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "metadata length") {
		t.Fatalf("LoadModelBinary error=%v want metadata length error", err)
	}
}

func TestLoadModelBinaryRejectsOversizedLayerCount(t *testing.T) {
	path := writeModelBinaryFixture(t, func(w io.Writer) {
		writeModelHeader(t, w, ModelMetadata{Name: "layers"}, maxModelLayers+1)
	})
	if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "layer count") {
		t.Fatalf("LoadModelBinary error=%v want layer count error", err)
	}
}

func TestLoadModelBinaryRejectsOversizedNameAndTypeLengths(t *testing.T) {
	tests := []struct {
		name  string
		label string
		write func(io.Writer)
	}{
		{name: "name", label: "layer name length", write: func(w io.Writer) {
			_ = binary.Write(w, binary.LittleEndian, maxModelStringBytes+1)
		}},
		{name: "type", label: "layer type length", write: func(w io.Writer) {
			_ = binary.Write(w, binary.LittleEndian, uint32(0))
			_ = binary.Write(w, binary.LittleEndian, maxModelStringBytes+1)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeModelBinaryFixture(t, func(w io.Writer) {
				writeModelHeader(t, w, ModelMetadata{Name: test.name}, 1)
				test.write(w)
			})
			if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), test.label) {
				t.Fatalf("LoadModelBinary error=%v want %s error", err, test.label)
			}
		})
	}
}

func TestLoadModelBinaryRejectsInvalidShape(t *testing.T) {
	tests := []struct {
		name  string
		label string
		write func(io.Writer)
	}{
		{name: "dimension-count", label: "shape dimension count", write: func(w io.Writer) {
			writeLayerPrefix(t, w, "x", "linear", nil)
		}},
		{name: "dimension", label: "shape dimension", write: func(w io.Writer) {
			writeLayerPrefix(t, w, "x", "linear", []uint32{maxModelMatrixElements + 1})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeModelBinaryFixture(t, func(w io.Writer) {
				writeModelHeader(t, w, ModelMetadata{Name: test.name}, 1)
				if test.name == "dimension-count" {
					_ = binary.Write(w, binary.LittleEndian, uint32(1))
					_, _ = io.WriteString(w, "x")
					_ = binary.Write(w, binary.LittleEndian, uint32(6))
					_, _ = io.WriteString(w, "linear")
					_ = binary.Write(w, binary.LittleEndian, maxModelShapeDims+1)
					return
				}
				test.write(w)
			})
			if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), test.label) {
				t.Fatalf("LoadModelBinary error=%v want %s error", err, test.label)
			}
		})
	}
}

func TestRetainedAllocationBudgetRejectsArithmeticOverflow(t *testing.T) {
	const maxUint64 = ^uint64(0)
	tests := []struct {
		name   string
		budget retainedAllocationBudget
		count  uint64
		size   uint64
	}{
		{name: "multiply", budget: retainedAllocationBudget{limit: maxUint64}, count: maxUint64, size: 2},
		{name: "add", budget: retainedAllocationBudget{used: maxUint64 - 1, limit: maxUint64}, count: 1, size: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			used := test.budget.used
			if err := test.budget.reserve("test", test.count, test.size); err == nil || !strings.Contains(err.Error(), "overflow") {
				t.Fatalf("reserve error=%v want arithmetic overflow", err)
			}
			if test.budget.used != used {
				t.Fatalf("failed reserve changed used bytes from %d to %d", used, test.budget.used)
			}
		})
	}
}

func TestRetainedAllocationBudgetReportsProductionLimit(t *testing.T) {
	budget := retainedAllocationBudget{used: maxRetainedModelBytes, limit: maxRetainedModelBytes}
	if err := budget.reserve("test", 1, 1); err == nil || !strings.Contains(err.Error(), "retained model allocation 536870913 exceeds 512 MiB limit") {
		t.Fatalf("reserve error=%v want production retained allocation limit", err)
	}
}

func TestValidateMatrixDimensionsRejectsOverflow(t *testing.T) {
	if _, err := validateMatrixDimensions(2, ^uint64(0)); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("validateMatrixDimensions error=%v want overflow error", err)
	}
}

func TestLoadModelBinaryRejectsOversizedMatrixRowCountBeforeAllocation(t *testing.T) {
	path := writeModelBinaryFixture(t, func(w io.Writer) {
		writeModelHeader(t, w, ModelMetadata{Name: "row-bound"}, 1)
		writeLayerPrefix(t, w, "x", "linear", nil)
		_ = binary.Write(w, binary.LittleEndian, uint32((1<<20)+1))
		_ = binary.Write(w, binary.LittleEndian, uint32(1))
	})
	if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "matrix row count") {
		t.Fatalf("LoadModelBinary error=%v want matrix row count error", err)
	}
}

func TestLoadModelBinaryRejectsInvalidMatrixDimensions(t *testing.T) {
	tests := []struct {
		name string
		rows uint32
		cols uint32
	}{
		{name: "element-limit", rows: maxModelMatrixRows, cols: 17},
		{name: "overflow-safe-product", rows: maxModelMatrixRows, cols: ^uint32(0)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeModelBinaryFixture(t, func(w io.Writer) {
				writeModelHeader(t, w, ModelMetadata{Name: test.name}, 1)
				writeLayerPrefix(t, w, "x", "linear", nil)
				_ = binary.Write(w, binary.LittleEndian, test.rows)
				_ = binary.Write(w, binary.LittleEndian, test.cols)
			})
			if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "matrix dimensions") {
				t.Fatalf("LoadModelBinary error=%v want matrix dimensions error", err)
			}
		})
	}
}

func TestLoadModelBinaryRejectsOversizedBiasLength(t *testing.T) {
	path := writeModelBinaryFixture(t, func(w io.Writer) {
		writeModelHeader(t, w, ModelMetadata{Name: "biases"}, 1)
		writeLayerPrefix(t, w, "x", "linear", nil)
		_ = binary.Write(w, binary.LittleEndian, uint32(0))
		_ = binary.Write(w, binary.LittleEndian, maxModelBiases+1)
	})
	if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "bias length") {
		t.Fatalf("LoadModelBinary error=%v want bias length error", err)
	}
}

func TestLoadModelBinaryRejectsTrailingDecodedData(t *testing.T) {
	path := writeModelBinaryFixture(t, func(w io.Writer) {
		writeModelHeader(t, w, ModelMetadata{Name: "trailing"}, 0)
		_, _ = w.Write([]byte{0x7f})
	})
	if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "trailing decoded data") {
		t.Fatalf("LoadModelBinary error=%v want trailing decoded data error", err)
	}
}

func TestLoadModelBinaryRejectsCumulativeRetainedAllocation(t *testing.T) {
	fixture := func(layers uint32) string {
		return writeModelBinaryFixture(t, func(w io.Writer) {
			writeModelHeader(t, w, ModelMetadata{Name: "retained-budget"}, layers)
			for i := uint32(0); i < layers; i++ {
				writeLayerPrefix(t, w, "z", "linear", []uint32{16, 0})
				_ = binary.Write(w, binary.LittleEndian, uint32(16))
				_ = binary.Write(w, binary.LittleEndian, uint32(0))
				_ = binary.Write(w, binary.LittleEndian, uint32(0))
			}
		})
	}
	if _, err := loadModelBinaryWithLimits(fixture(1), maxDecodedModelBytes, 2800); err != nil {
		t.Fatalf("single compact layer exceeded retained budget: %v", err)
	}
	if _, err := loadModelBinaryWithLimits(fixture(3), maxDecodedModelBytes, 2800); err == nil || !strings.Contains(err.Error(), "retained model allocation") {
		t.Fatalf("loadModelBinaryWithLimits error=%v want cumulative retained allocation error", err)
	}
}

func TestLoadModelBinaryRejectsDecodedPayloadOverLimit(t *testing.T) {
	var decoded bytes.Buffer
	writeModelHeader(t, &decoded, ModelMetadata{Name: "bounded"}, 0)
	decodedLimit := int64(decoded.Len())
	decoded.WriteByte(0)
	path := writeModelBinaryFixture(t, func(w io.Writer) {
		_, _ = w.Write(decoded.Bytes())
	})
	if _, err := loadModelBinaryWithLimit(path, decodedLimit); err == nil || !strings.Contains(err.Error(), "decoded model exceeds 512 MiB limit") {
		t.Fatalf("loadModelBinaryWithLimit error=%v want decoded payload limit error", err)
	}
}

func TestLoadModelBinaryRejectsGzipTruncationAndChecksumErrors(t *testing.T) {
	validPath := writeModelBinaryFixture(t, func(w io.Writer) {
		writeModelHeader(t, w, ModelMetadata{Name: "gzip"}, 0)
	})
	valid, err := os.ReadFile(validPath)
	if err != nil {
		t.Fatal(err)
	}
	fixtures := map[string][]byte{
		"truncated": valid[:len(valid)-1],
		"checksum":  append(append([]byte(nil), valid[:len(valid)-8]...), append([]byte{valid[len(valid)-8] ^ 0xff}, valid[len(valid)-7:]...)...),
	}
	for name, data := range fixtures {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), name+".bin")
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadModelBinary(path); err == nil {
				t.Fatal("LoadModelBinary returned nil error")
			}
		})
	}
}

func TestLoadModelBinaryReadsFrozenVersionOneFixtures(t *testing.T) {
	// These fixed gzip streams were generated from the pre-hardening version-1
	// layout and deliberately do not call the current serializer.
	const (
		frozenV1EmptyModel = "H4sIAAAAAAACAyWOsQ6DMAxE01/xDBVrM3ds1U9AVjAlUuJQx0ECxFf0hxvU4Ya7ezrd4/56Xowx36odGCOBhVHSRtxSnHWFBhaS7BPXorveqkdxk1dyWuSkOTHVlEvsA66VBds1oEkx9DMKxn/wKcjqNxrAjhgyNeBK1hTB7pBTEXduzULthDIQe37DcdRX5gfzuYc1owAAAA=="
		frozenV1ZeroRows   = "H4sIAAAAAAACAzVNQQrCQAzcKnjxFzm3otdevHhUfEIJNcWF3WxNs9VW/LUPcG11YEhmMkOOh/MpM8b0iU9g9AQlNBJG4mIkCYWEewc59CSdDZyOu802aZT6apVqjTI17GPacuDoK4dDikO5y0GDoqtaFPSzcYvIake6QNmg6+j1/b5OJN/qUFhuo66SdJYJZWFmLH8zM3+89x+suDg8uwAAAA=="
		frozenV1ZeroCols   = "H4sIAAAAAAACA22NMQ7CMAxFAwMDx/CcIsrYhYURxBEqq3UhUuoUJ62giLNxFY6CWzEyPP3/bX/5eDifFsaYQXkCY0tQQCNhJM5GkpBVwUewMJBEF1iX+WarGaW6ukRV6mVuuPvsLHDflh4feg5FbiGFhL7sULDVwc7CrUdObqQaigZ9pNf0fa3UdCEmwUQrTd4xoSzVTZg/asxn/9P3F5y2o1LGAAAA"
	)

	tests := []struct {
		name    string
		fixture string
		want    *Model
	}{
		{name: "empty-model", fixture: frozenV1EmptyModel, want: &Model{
			Metadata: ModelMetadata{Name: "frozen-empty", Version: "0.9", Architecture: "none", Custom: map[string]string{"source": "pre-hardening"}},
			Layers:   []SerializedLayer{},
		}},
		{name: "zero-rows", fixture: frozenV1ZeroRows, want: &Model{
			Metadata: ModelMetadata{Name: "frozen-zero-rows", Version: "1.0", Architecture: "fixture", NumLayers: 1, TotalParams: 1},
			Layers: []SerializedLayer{{
				Name: "empty-input", Type: "linear", Shape: []int{0, 3}, Biases: []float64{1.25}, Extra: map[string]interface{}{},
			}},
		}},
		{name: "nonzero-rows-zero-columns", fixture: frozenV1ZeroCols, want: &Model{
			Metadata: ModelMetadata{Name: "frozen-zero-cols", Version: "1.0", Architecture: "fixture", NumLayers: 1, TotalParams: 2},
			Layers: []SerializedLayer{{
				Name: "degenerate", Type: "linear", Shape: []int{2, 0}, Weights: [][]float64{make([]float64, 0), make([]float64, 0)}, Biases: []float64{0.5, -0.5}, Extra: map[string]interface{}{},
			}},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := base64.StdEncoding.DecodeString(test.fixture)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "frozen-v1.bin")
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := LoadModelBinary(path)
			if err != nil {
				t.Fatalf("LoadModelBinary error=%v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("frozen version-1 model mismatch\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}

func TestModelSaveLoadBinaryPreservesZeroColumnRows(t *testing.T) {
	model := NewModel("zero-columns", "linear")
	model.Metadata.Custom["source"] = "roundtrip"
	model.AddLayer("degenerate", "linear", [][]float64{make([]float64, 0), make([]float64, 0)}, []float64{0.5, -0.5})
	path := filepath.Join(t.TempDir(), "zero-columns.bin")
	if err := model.SaveBinary(path); err != nil {
		t.Fatalf("SaveBinary error=%v", err)
	}
	loaded, err := LoadModelBinary(path)
	if err != nil {
		t.Fatalf("LoadModelBinary error=%v", err)
	}
	if !reflect.DeepEqual(loaded.Metadata, model.Metadata) {
		t.Fatalf("metadata mismatch\ngot:  %#v\nwant: %#v", loaded.Metadata, model.Metadata)
	}
	if len(loaded.Layers) != 1 {
		t.Fatalf("loaded layer count=%d want 1", len(loaded.Layers))
	}
	layer := loaded.Layers[0]
	if layer.Name != "degenerate" || layer.Type != "linear" || !reflect.DeepEqual(layer.Shape, []int{2, 0}) || !reflect.DeepEqual(layer.Biases, []float64{0.5, -0.5}) {
		t.Fatalf("loaded zero-column layer metadata mismatch: %#v", layer)
	}
	if len(layer.Weights) != 2 || layer.Weights[0] == nil || layer.Weights[1] == nil || len(layer.Weights[0]) != 0 || len(layer.Weights[1]) != 0 {
		t.Fatalf("loaded weights=%#v want two non-nil empty rows", layer.Weights)
	}
}

func TestModelSaveLoadBinaryPreservesIndependentRowAppendSemantics(t *testing.T) {
	model := NewModel("independent-rows", "linear")
	model.AddLayer("matrix", "linear", [][]float64{{1, 2}, {3, 4}}, nil)
	path := filepath.Join(t.TempDir(), "independent-rows.bin")
	if err := model.SaveBinary(path); err != nil {
		t.Fatalf("SaveBinary error=%v", err)
	}
	loaded, err := LoadModelBinary(path)
	if err != nil {
		t.Fatalf("LoadModelBinary error=%v", err)
	}
	loaded.Layers[0].Weights[0] = append(loaded.Layers[0].Weights[0], 9)
	if !reflect.DeepEqual(loaded.Layers[0].Weights[1], []float64{3, 4}) {
		t.Fatalf("appending row 0 changed row 1: %v", loaded.Layers[0].Weights[1])
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

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	var magic, version uint32
	if err := binary.Read(gz, binary.LittleEndian, &magic); err != nil {
		t.Fatal(err)
	}
	if err := binary.Read(gz, binary.LittleEndian, &version); err != nil {
		t.Fatal(err)
	}
	if magic != 0x4D4F444C || version != 1 {
		t.Fatalf("binary header magic=%#x version=%d want MODL version 1", magic, version)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
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
