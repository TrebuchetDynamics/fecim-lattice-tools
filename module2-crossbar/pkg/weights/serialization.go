// Package weights provides weight management utilities for crossbar neural networks.
package weights

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"unicode/utf8"
)

const (
	maxDecodedModelBytes    int64  = 512 << 20
	maxRetainedModelBytes   uint64 = 512 << 20
	maxModelMetadataBytes   uint32 = 1 << 20
	modelRetainedBytes      uint64 = 256
	modelMetadataMultiplier uint64 = 8
	modelLayerRetainedBytes uint64 = 192
	modelStringMultiplier   uint64 = 2
	modelIntBytes           uint64 = 8
	modelRowDescriptorBytes uint64 = 24
	modelFloat64Bytes       uint64 = 8
	maxModelLayers          uint32 = 4096
	maxModelStringBytes     uint32 = 64 << 10
	maxModelShapeDims       uint32 = 8
	maxModelMatrixRows      uint32 = 1 << 20
	maxModelMatrixElements  uint32 = 16 << 20
	maxModelBiases          uint32 = 1 << 20
	float64ReadScratchBytes        = 32 << 10
)

func readFloat64Slice(r io.Reader, values []float64) error {
	var scratch [float64ReadScratchBytes]byte
	const valuesPerChunk = float64ReadScratchBytes / 8
	for offset := 0; offset < len(values); {
		count := len(values) - offset
		if count > valuesPerChunk {
			count = valuesPerChunk
		}
		chunk := scratch[:count*8]
		if _, err := io.ReadFull(r, chunk); err != nil {
			return err
		}
		for index := 0; index < count; index++ {
			bits := binary.LittleEndian.Uint64(chunk[index*8 : index*8+8])
			values[offset+index] = math.Float64frombits(bits)
		}
		offset += count
	}
	return nil
}

func checkedLength(label string, value, limit uint32) error {
	if value > limit {
		return fmt.Errorf("%s %d exceeds limit %d", label, value, limit)
	}
	return nil
}

type retainedAllocationBudget struct {
	used  uint64
	limit uint64
}

func (budget *retainedAllocationBudget) reserve(label string, count, size uint64) error {
	const maxUint64 = ^uint64(0)
	if count != 0 && size > maxUint64/count {
		return fmt.Errorf("retained model allocation overflow while charging %s", label)
	}
	bytes := count * size
	if budget.used > maxUint64-bytes {
		return fmt.Errorf("retained model allocation overflow while charging %s", label)
	}
	total := budget.used + bytes
	if total > budget.limit {
		if budget.limit == maxRetainedModelBytes {
			return fmt.Errorf("retained model allocation %d exceeds 512 MiB limit while charging %s", total, label)
		}
		return fmt.Errorf("retained model allocation %d exceeds configured %d-byte limit while charging %s", total, budget.limit, label)
	}
	budget.used = total
	return nil
}

func (budget *retainedAllocationBudget) reserveString(label string, length uint64) error {
	return budget.reserve(label, length, modelStringMultiplier)
}

func (budget *retainedAllocationBudget) reserveShape(dimensions uint64) error {
	return budget.reserve("shape", dimensions, modelIntBytes)
}

func (budget *retainedAllocationBudget) reserveMatrix(rows, elements uint64) error {
	if err := budget.reserve("matrix row descriptors", rows, modelRowDescriptorBytes); err != nil {
		return err
	}
	return budget.reserve("matrix elements", elements, modelFloat64Bytes)
}

func (budget *retainedAllocationBudget) reserveBiases(count uint64) error {
	return budget.reserve("biases", count, modelFloat64Bytes)
}

func validateMatrixDimensions(rows, cols uint64) (uint64, error) {
	if rows > uint64(maxModelMatrixRows) {
		return 0, fmt.Errorf("matrix row count %d exceeds limit %d", rows, maxModelMatrixRows)
	}
	elements := rows * cols
	if rows != 0 && elements/rows != cols {
		return 0, fmt.Errorf("matrix dimensions %dx%d overflow", rows, cols)
	}
	if elements > uint64(maxModelMatrixElements) {
		return 0, fmt.Errorf("matrix dimensions %dx%d exceed limit", rows, cols)
	}
	return elements, nil
}

// ModelMetadata stores model information for serialization.
type ModelMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	NumLayers    int               `json:"num_layers"`
	TotalParams  int               `json:"total_params"`
	Quantized    bool              `json:"quantized"`
	QuantBits    int               `json:"quant_bits,omitempty"`
	Custom       map[string]string `json:"custom,omitempty"`
}

func addEncodedSize(total *uint64, amount uint64) error {
	const maxUint64 = ^uint64(0)
	if *total > maxUint64-amount {
		return fmt.Errorf("metadata encoded size overflow")
	}
	*total += amount
	return nil
}

func jsonQuotedStringSize(value string) (uint64, error) {
	size := uint64(2) // surrounding quotes
	for index := 0; index < len(value); {
		byteValue := value[index]
		if byteValue < utf8.RuneSelf {
			var encoded uint64 = 1
			switch byteValue {
			case '\\', '"', '\b', '\f', '\n', '\r', '\t':
				encoded = 2
			default:
				if byteValue < 0x20 || byteValue == '<' || byteValue == '>' || byteValue == '&' {
					encoded = 6
				}
			}
			if err := addEncodedSize(&size, encoded); err != nil {
				return 0, err
			}
			index++
			continue
		}

		runeValue, width := utf8.DecodeRuneInString(value[index:])
		encoded := uint64(width)
		if runeValue == utf8.RuneError && width == 1 {
			encoded = 6 // encoding/json emits invalid UTF-8 as \ufffd.
		} else if runeValue == '\u2028' || runeValue == '\u2029' {
			encoded = 6
		}
		if err := addEncodedSize(&size, encoded); err != nil {
			return 0, err
		}
		index += width
	}
	return size, nil
}

func jsonIntSize(value int) uint64 {
	var magnitude uint64
	size := uint64(0)
	if value < 0 {
		size = 1
		magnitude = uint64(-(value + 1)) + 1
	} else {
		magnitude = uint64(value)
	}
	size++
	for magnitude >= 10 {
		magnitude /= 10
		size++
	}
	return size
}

func modelMetadataEncodedSize(metadata ModelMetadata) (uint64, error) {
	size := uint64(1) // opening object brace
	addLiteral := func(literal string) error {
		return addEncodedSize(&size, uint64(len(literal)))
	}
	addString := func(value string) error {
		quoted, err := jsonQuotedStringSize(value)
		if err != nil {
			return err
		}
		return addEncodedSize(&size, quoted)
	}

	if err := addLiteral(`"name":`); err != nil {
		return 0, err
	}
	if err := addString(metadata.Name); err != nil {
		return 0, err
	}
	if err := addLiteral(`,"version":`); err != nil {
		return 0, err
	}
	if err := addString(metadata.Version); err != nil {
		return 0, err
	}
	if err := addLiteral(`,"architecture":`); err != nil {
		return 0, err
	}
	if err := addString(metadata.Architecture); err != nil {
		return 0, err
	}
	if err := addLiteral(`,"num_layers":`); err != nil {
		return 0, err
	}
	if err := addEncodedSize(&size, jsonIntSize(metadata.NumLayers)); err != nil {
		return 0, err
	}
	if err := addLiteral(`,"total_params":`); err != nil {
		return 0, err
	}
	if err := addEncodedSize(&size, jsonIntSize(metadata.TotalParams)); err != nil {
		return 0, err
	}
	if err := addLiteral(`,"quantized":`); err != nil {
		return 0, err
	}
	if metadata.Quantized {
		if err := addLiteral("true"); err != nil {
			return 0, err
		}
	} else if err := addLiteral("false"); err != nil {
		return 0, err
	}
	if metadata.QuantBits != 0 {
		if err := addLiteral(`,"quant_bits":`); err != nil {
			return 0, err
		}
		if err := addEncodedSize(&size, jsonIntSize(metadata.QuantBits)); err != nil {
			return 0, err
		}
	}
	if len(metadata.Custom) > 0 {
		if err := addLiteral(`,"custom":{`); err != nil {
			return 0, err
		}
		first := true
		for key, value := range metadata.Custom {
			if !first {
				if err := addLiteral(","); err != nil {
					return 0, err
				}
			}
			first = false
			if err := addString(key); err != nil {
				return 0, err
			}
			if err := addLiteral(":"); err != nil {
				return 0, err
			}
			if err := addString(value); err != nil {
				return 0, err
			}
		}
		if err := addLiteral("}"); err != nil {
			return 0, err
		}
	}
	if err := addLiteral("}"); err != nil {
		return 0, err
	}
	return size, nil
}

func marshalBoundedModelMetadata(metadata ModelMetadata, marshalFn func(any) ([]byte, error)) ([]byte, error) {
	encodedSize, err := modelMetadataEncodedSize(metadata)
	if err != nil {
		return nil, err
	}
	if encodedSize > uint64(maxModelMetadataBytes) {
		return nil, fmt.Errorf("metadata length %d exceeds limit", encodedSize)
	}
	encoded, err := marshalFn(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}
	if len(encoded) > int(maxModelMetadataBytes) {
		return nil, fmt.Errorf("metadata length %d exceeds limit", len(encoded))
	}
	return encoded, nil
}

// SerializedLayer stores weights for a single layer in model serialization.
type SerializedLayer struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"` // "linear", "conv2d", "attention", etc.
	Shape   []int                  `json:"shape"`
	Weights [][]float64            `json:"weights,omitempty"`
	Biases  []float64              `json:"biases,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// Model represents a complete neural network model for serialization.
type Model struct {
	Metadata ModelMetadata     `json:"metadata"`
	Layers   []SerializedLayer `json:"layers"`
}

// NewModel creates a new model for serialization.
func NewModel(name, architecture string) *Model {
	return &Model{
		Metadata: ModelMetadata{
			Name:         name,
			Version:      "1.0",
			Architecture: architecture,
			Custom:       make(map[string]string),
		},
		Layers: make([]SerializedLayer, 0),
	}
}

// AddLayer adds a layer to the model.
func (m *Model) AddLayer(name, layerType string, weights [][]float64, biases []float64) {
	shape := make([]int, 2)
	if len(weights) > 0 {
		shape[0] = len(weights)
		if len(weights[0]) > 0 {
			shape[1] = len(weights[0])
		}
	}

	// Count parameters
	params := shape[0] * shape[1]
	if biases != nil {
		params += len(biases)
	}
	m.Metadata.TotalParams += params
	m.Metadata.NumLayers++

	m.Layers = append(m.Layers, SerializedLayer{
		Name:    name,
		Type:    layerType,
		Shape:   shape,
		Weights: weights,
		Biases:  biases,
		Extra:   make(map[string]interface{}),
	})
}

// SaveJSON saves model to JSON file.
func (m *Model) SaveJSON(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(m)
}

// LoadModelJSON loads model from JSON file.
func LoadModelJSON(path string) (*Model, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var model Model
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &model, nil
}

func validateSerializedLayer(layer *SerializedLayer) error {
	if len(layer.Name) > int(maxModelStringBytes) {
		return fmt.Errorf("layer name length %d exceeds limit", len(layer.Name))
	}
	if len(layer.Type) > int(maxModelStringBytes) {
		return fmt.Errorf("layer type length %d exceeds limit", len(layer.Type))
	}
	if len(layer.Shape) > int(maxModelShapeDims) {
		return fmt.Errorf("shape dimension count %d exceeds limit", len(layer.Shape))
	}
	for index, dim := range layer.Shape {
		if dim < 0 || uint64(dim) > uint64(maxModelMatrixElements) {
			return fmt.Errorf("shape dimension %d value %d is invalid", index, dim)
		}
	}

	rows := len(layer.Weights)
	cols := 0
	if rows > 0 {
		cols = len(layer.Weights[0])
	}
	if _, err := validateMatrixDimensions(uint64(rows), uint64(cols)); err != nil {
		return err
	}
	for row := range layer.Weights {
		if len(layer.Weights[row]) != cols {
			return fmt.Errorf("weights must be rectangular: row %d has %d columns, want %d", row, len(layer.Weights[row]), cols)
		}
	}
	if len(layer.Biases) > int(maxModelBiases) {
		return fmt.Errorf("bias length %d exceeds limit", len(layer.Biases))
	}
	return nil
}

func serializedLayerSize(layer *SerializedLayer) uint64 {
	rows := len(layer.Weights)
	cols := 0
	if rows > 0 {
		cols = len(layer.Weights[0])
	}
	size := uint64(4+len(layer.Name)) + uint64(4+len(layer.Type))
	size += 4 + uint64(len(layer.Shape))*4
	size += 4
	if rows > 0 {
		size += 4 + uint64(rows)*uint64(cols)*8
	}
	return size + 4 + uint64(len(layer.Biases))*8
}

func validateModelForBinary(model *Model) ([]byte, error) {
	return validateModelForBinaryWithRetainedLimit(model, maxRetainedModelBytes)
}

func validateModelForBinaryWithRetainedLimit(model *Model, retainedLimit uint64) ([]byte, error) {
	if len(model.Layers) > int(maxModelLayers) {
		return nil, fmt.Errorf("layer count %d exceeds limit", len(model.Layers))
	}
	metadata, err := marshalBoundedModelMetadata(model.Metadata, json.Marshal)
	if err != nil {
		return nil, err
	}

	budget := &retainedAllocationBudget{limit: retainedLimit}
	if err := budget.reserve("metadata", uint64(len(metadata)), modelMetadataMultiplier); err != nil {
		return nil, err
	}
	if err := budget.reserve("model state", 1, modelRetainedBytes); err != nil {
		return nil, err
	}
	if err := budget.reserve("layer state", uint64(len(model.Layers)), modelLayerRetainedBytes); err != nil {
		return nil, err
	}

	decodedSize := uint64(4 + 4 + 4 + len(metadata) + 4)
	for index := range model.Layers {
		layer := &model.Layers[index]
		if err := validateSerializedLayer(layer); err != nil {
			return nil, fmt.Errorf("layer %d: %w", index, err)
		}
		rows := len(layer.Weights)
		cols := 0
		if rows > 0 {
			cols = len(layer.Weights[0])
		}
		elements, err := validateMatrixDimensions(uint64(rows), uint64(cols))
		if err != nil {
			return nil, fmt.Errorf("layer %d: %w", index, err)
		}
		if err := budget.reserveString("layer name", uint64(len(layer.Name))); err != nil {
			return nil, fmt.Errorf("layer %d: %w", index, err)
		}
		if err := budget.reserveString("layer type", uint64(len(layer.Type))); err != nil {
			return nil, fmt.Errorf("layer %d: %w", index, err)
		}
		if err := budget.reserveShape(uint64(len(layer.Shape))); err != nil {
			return nil, fmt.Errorf("layer %d: %w", index, err)
		}
		if err := budget.reserveMatrix(uint64(rows), elements); err != nil {
			return nil, fmt.Errorf("layer %d: %w", index, err)
		}
		if err := budget.reserveBiases(uint64(len(layer.Biases))); err != nil {
			return nil, fmt.Errorf("layer %d: %w", index, err)
		}
		decodedSize += serializedLayerSize(layer)
		if decodedSize > uint64(maxDecodedModelBytes) {
			return nil, fmt.Errorf("decoded model exceeds 512 MiB limit")
		}
	}
	return metadata, nil
}

// SaveBinary saves model in compact binary format.
// Format: header + metadata + layers
func (m *Model) SaveBinary(path string) error {
	return m.saveBinaryWithRetainedLimit(path, maxRetainedModelBytes)
}

func (m *Model) saveBinaryWithRetainedLimit(path string, retainedLimit uint64) error {
	metaBytes, err := validateModelForBinaryWithRetainedLimit(m, retainedLimit)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	// Use gzip compression. Validation above is complete before the destination is opened.
	gzWriter := gzip.NewWriter(file)
	failWrite := func(writeErr error) error {
		_ = gzWriter.Close()
		_ = file.Close()
		return writeErr
	}

	// Write magic number and version.
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(0x4D4F444C)); err != nil { // "MODL"
		return failWrite(err)
	}
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(1)); err != nil { // version 1
		return failWrite(err)
	}

	// Write the metadata bytes prepared during preflight validation.
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(len(metaBytes))); err != nil {
		return failWrite(err)
	}
	if _, err := gzWriter.Write(metaBytes); err != nil {
		return failWrite(err)
	}

	// Write number of layers.
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(len(m.Layers))); err != nil {
		return failWrite(err)
	}

	// Write each layer.
	for index := range m.Layers {
		layer := &m.Layers[index]
		if err := writeSerializedLayerBinary(gzWriter, layer); err != nil {
			return failWrite(fmt.Errorf("failed to write layer %s: %w", layer.Name, err))
		}
	}

	if err := gzWriter.Close(); err != nil {
		_ = file.Close()
		return fmt.Errorf("finalize gzip stream: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close model file: %w", err)
	}
	return nil
}

func writeSerializedLayerBinary(w io.Writer, layer *SerializedLayer) error {
	if err := validateSerializedLayer(layer); err != nil {
		return err
	}

	// Write layer name
	nameBytes := []byte(layer.Name)
	if err := binary.Write(w, binary.LittleEndian, uint32(len(nameBytes))); err != nil {
		return err
	}
	if _, err := w.Write(nameBytes); err != nil {
		return err
	}

	// Write layer type
	typeBytes := []byte(layer.Type)
	if err := binary.Write(w, binary.LittleEndian, uint32(len(typeBytes))); err != nil {
		return err
	}
	if _, err := w.Write(typeBytes); err != nil {
		return err
	}

	// Write shape
	if err := binary.Write(w, binary.LittleEndian, uint32(len(layer.Shape))); err != nil {
		return err
	}
	for _, dim := range layer.Shape {
		if err := binary.Write(w, binary.LittleEndian, uint32(dim)); err != nil {
			return err
		}
	}

	// Write weights
	rows := len(layer.Weights)
	if err := binary.Write(w, binary.LittleEndian, uint32(rows)); err != nil {
		return err
	}
	if rows > 0 {
		cols := len(layer.Weights[0])
		if err := binary.Write(w, binary.LittleEndian, uint32(cols)); err != nil {
			return err
		}
		for i := range layer.Weights {
			for j := range layer.Weights[i] {
				if err := binary.Write(w, binary.LittleEndian, layer.Weights[i][j]); err != nil {
					return err
				}
			}
		}
	}

	// Write biases
	biasLen := len(layer.Biases)
	if err := binary.Write(w, binary.LittleEndian, uint32(biasLen)); err != nil {
		return err
	}
	for _, b := range layer.Biases {
		if err := binary.Write(w, binary.LittleEndian, b); err != nil {
			return err
		}
	}

	return nil
}

// LoadModelBinary loads model from binary format.
func LoadModelBinary(path string) (*Model, error) {
	return loadModelBinaryWithLimits(path, maxDecodedModelBytes, maxRetainedModelBytes)
}

func loadModelBinaryWithLimit(path string, decodedLimit int64) (*Model, error) {
	return loadModelBinaryWithLimits(path, decodedLimit, maxRetainedModelBytes)
}

func loadModelBinaryWithLimits(path string, decodedLimit int64, retainedLimit uint64) (*Model, error) {
	budget := &retainedAllocationBudget{limit: retainedLimit}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()
	limited := &io.LimitedReader{R: gzReader, N: decodedLimit + 1}

	// Read magic number
	var magic uint32
	if err := binary.Read(limited, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if magic != 0x4D4F444C {
		return nil, fmt.Errorf("invalid magic number: expected MODL")
	}

	// Read version
	var version uint32
	if err := binary.Read(limited, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	if version != 1 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	// Read metadata
	var metaLen uint32
	if err := binary.Read(limited, binary.LittleEndian, &metaLen); err != nil {
		return nil, err
	}
	if err := checkedLength("metadata length", metaLen, maxModelMetadataBytes); err != nil {
		return nil, err
	}
	if err := budget.reserve("metadata", uint64(metaLen), modelMetadataMultiplier); err != nil {
		return nil, err
	}
	if err := budget.reserve("model state", 1, modelRetainedBytes); err != nil {
		return nil, err
	}
	metaBytes := make([]byte, int(metaLen))
	if _, err := io.ReadFull(limited, metaBytes); err != nil {
		return nil, err
	}

	model := &Model{}
	if err := json.Unmarshal(metaBytes, &model.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Read number of layers
	var numLayers uint32
	if err := binary.Read(limited, binary.LittleEndian, &numLayers); err != nil {
		return nil, err
	}
	if err := checkedLength("layer count", numLayers, maxModelLayers); err != nil {
		return nil, err
	}
	if err := budget.reserve("layer state", uint64(numLayers), modelLayerRetainedBytes); err != nil {
		return nil, err
	}

	// Read each layer
	model.Layers = make([]SerializedLayer, int(numLayers))
	for i := uint32(0); i < numLayers; i++ {
		layer, err := readSerializedLayerBinary(limited, budget)
		if err != nil {
			return nil, fmt.Errorf("failed to read layer %d: %w", i, err)
		}
		model.Layers[int(i)] = *layer
	}

	var trailing [1]byte
	n, readErr := limited.Read(trailing[:])
	if n > 0 {
		if limited.N <= 0 {
			return nil, fmt.Errorf("decoded model exceeds 512 MiB limit")
		}
		if readErr != nil && readErr != io.EOF {
			return nil, fmt.Errorf("validate gzip stream: %w", readErr)
		}
		if readErr == nil {
			_, drainErr := io.Copy(io.Discard, limited)
			if limited.N <= 0 {
				return nil, fmt.Errorf("decoded model exceeds 512 MiB limit")
			}
			if drainErr != nil {
				return nil, fmt.Errorf("validate gzip stream: %w", drainErr)
			}
		}
		return nil, fmt.Errorf("trailing decoded data")
	}
	if limited.N <= 0 {
		return nil, fmt.Errorf("decoded model exceeds 512 MiB limit")
	}
	if readErr == nil {
		return nil, fmt.Errorf("validate gzip stream: reader made no progress")
	}
	if readErr != io.EOF {
		return nil, fmt.Errorf("validate gzip stream: %w", readErr)
	}

	return model, nil
}

func readSerializedLayerBinary(r io.Reader, budget *retainedAllocationBudget) (*SerializedLayer, error) {
	layer := &SerializedLayer{
		Extra: make(map[string]interface{}),
	}

	// Read layer name
	var nameLen uint32
	if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
		return nil, err
	}
	if err := checkedLength("layer name length", nameLen, maxModelStringBytes); err != nil {
		return nil, err
	}
	if err := budget.reserveString("layer name", uint64(nameLen)); err != nil {
		return nil, err
	}
	nameBytes := make([]byte, int(nameLen))
	if _, err := io.ReadFull(r, nameBytes); err != nil {
		return nil, err
	}
	layer.Name = string(nameBytes)

	// Read layer type
	var typeLen uint32
	if err := binary.Read(r, binary.LittleEndian, &typeLen); err != nil {
		return nil, err
	}
	if err := checkedLength("layer type length", typeLen, maxModelStringBytes); err != nil {
		return nil, err
	}
	if err := budget.reserveString("layer type", uint64(typeLen)); err != nil {
		return nil, err
	}
	typeBytes := make([]byte, int(typeLen))
	if _, err := io.ReadFull(r, typeBytes); err != nil {
		return nil, err
	}
	layer.Type = string(typeBytes)

	// Read shape
	var shapeDims uint32
	if err := binary.Read(r, binary.LittleEndian, &shapeDims); err != nil {
		return nil, err
	}
	if err := checkedLength("shape dimension count", shapeDims, maxModelShapeDims); err != nil {
		return nil, err
	}
	if err := budget.reserveShape(uint64(shapeDims)); err != nil {
		return nil, err
	}
	layer.Shape = make([]int, int(shapeDims))
	for i := uint32(0); i < shapeDims; i++ {
		var dim uint32
		if err := binary.Read(r, binary.LittleEndian, &dim); err != nil {
			return nil, err
		}
		if dim > maxModelMatrixElements {
			return nil, fmt.Errorf("shape dimension %d exceeds limit %d", dim, maxModelMatrixElements)
		}
		layer.Shape[int(i)] = int(dim)
	}

	// Read weights
	var rows uint32
	if err := binary.Read(r, binary.LittleEndian, &rows); err != nil {
		return nil, err
	}
	if rows > 0 {
		var cols uint32
		if err := binary.Read(r, binary.LittleEndian, &cols); err != nil {
			return nil, err
		}
		elements, err := validateMatrixDimensions(uint64(rows), uint64(cols))
		if err != nil {
			return nil, err
		}
		if err := budget.reserveMatrix(uint64(rows), elements); err != nil {
			return nil, err
		}
		backing := make([]float64, int(elements))
		if err := readFloat64Slice(r, backing); err != nil {
			return nil, err
		}
		layer.Weights = make([][]float64, int(rows))
		for row := uint32(0); row < rows; row++ {
			start := int(uint64(row) * uint64(cols))
			end := start + int(cols)
			layer.Weights[int(row)] = backing[start:end:end]
		}
	}

	// Read biases
	var biasLen uint32
	if err := binary.Read(r, binary.LittleEndian, &biasLen); err != nil {
		return nil, err
	}
	if err := checkedLength("bias length", biasLen, maxModelBiases); err != nil {
		return nil, err
	}
	if err := budget.reserveBiases(uint64(biasLen)); err != nil {
		return nil, err
	}
	layer.Biases = make([]float64, int(biasLen))
	for i := uint32(0); i < biasLen; i++ {
		if err := binary.Read(r, binary.LittleEndian, &layer.Biases[int(i)]); err != nil {
			return nil, err
		}
	}

	return layer, nil
}

// QuantizedModel represents a quantized model for efficient storage.
type QuantizedModel struct {
	Metadata     ModelMetadata
	Layers       []QuantizedLayerWeights
	ScaleFactors []float64 // Per-layer scale factors
}

// QuantizedLayerWeights stores quantized weights.
type QuantizedLayerWeights struct {
	Name        string
	Type        string
	Shape       []int
	Weights     [][]int8 // Quantized to int8
	Biases      []int16  // Quantized to int16
	WeightScale float64
	BiasScale   float64
}

// QuantizeModel quantizes a model to int8 weights.
func QuantizeModel(model *Model, bits int) *QuantizedModel {
	qmodel := &QuantizedModel{
		Metadata: model.Metadata,
		Layers:   make([]QuantizedLayerWeights, len(model.Layers)),
	}
	qmodel.Metadata.Quantized = true
	qmodel.Metadata.QuantBits = bits

	maxVal := float64(int(1)<<(bits-1) - 1)

	for i, layer := range model.Layers {
		qLayer := QuantizedLayerWeights{
			Name:  layer.Name,
			Type:  layer.Type,
			Shape: layer.Shape,
		}

		// Find max absolute weight
		weightMax := 0.0
		for _, row := range layer.Weights {
			for _, w := range row {
				if abs := math.Abs(w); abs > weightMax {
					weightMax = abs
				}
			}
		}

		// Quantize weights
		if weightMax > 0 {
			qLayer.WeightScale = weightMax / maxVal
		} else {
			qLayer.WeightScale = 1.0
		}

		qLayer.Weights = make([][]int8, len(layer.Weights))
		for r := range layer.Weights {
			qLayer.Weights[r] = make([]int8, len(layer.Weights[r]))
			for c := range layer.Weights[r] {
				qLayer.Weights[r][c] = int8(layer.Weights[r][c] / qLayer.WeightScale)
			}
		}

		// Quantize biases (use 16-bit for better precision)
		if len(layer.Biases) > 0 {
			biasMax := 0.0
			for _, b := range layer.Biases {
				if abs := math.Abs(b); abs > biasMax {
					biasMax = abs
				}
			}

			biasMaxVal := float64(int(1)<<15 - 1)
			if biasMax > 0 {
				qLayer.BiasScale = biasMax / biasMaxVal
			} else {
				qLayer.BiasScale = 1.0
			}

			qLayer.Biases = make([]int16, len(layer.Biases))
			for j, b := range layer.Biases {
				qLayer.Biases[j] = int16(b / qLayer.BiasScale)
			}
		}

		qmodel.Layers[i] = qLayer
	}

	return qmodel
}

// DequantizeLayer converts quantized layer back to float64.
func (qw *QuantizedLayerWeights) Dequantize() SerializedLayer {
	layer := SerializedLayer{
		Name:    qw.Name,
		Type:    qw.Type,
		Shape:   qw.Shape,
		Weights: make([][]float64, len(qw.Weights)),
		Biases:  make([]float64, len(qw.Biases)),
		Extra:   make(map[string]interface{}),
	}

	for i := range qw.Weights {
		layer.Weights[i] = make([]float64, len(qw.Weights[i]))
		for j := range qw.Weights[i] {
			layer.Weights[i][j] = float64(qw.Weights[i][j]) * qw.WeightScale
		}
	}

	for i := range qw.Biases {
		layer.Biases[i] = float64(qw.Biases[i]) * qw.BiasScale
	}

	return layer
}

// CrossbarMapping stores information about crossbar tile mapping.
type CrossbarMapping struct {
	LayerName   string
	TileSize    [2]int // [rows, cols]
	NumTiles    int
	TileOffsets [][2]int   // Starting offset for each tile
	TileMasks   [][][]bool // Skip masks for sparse weights
}

// GenerateCrossbarMapping creates mapping for deploying weights to crossbar tiles.
func GenerateCrossbarMapping(layer *SerializedLayer, tileRows, tileCols int) *CrossbarMapping {
	mapping := &CrossbarMapping{
		LayerName: layer.Name,
		TileSize:  [2]int{tileRows, tileCols},
	}

	rows := layer.Shape[0]
	cols := layer.Shape[1]

	numTileRows := (rows + tileRows - 1) / tileRows
	numTileCols := (cols + tileCols - 1) / tileCols
	mapping.NumTiles = numTileRows * numTileCols

	mapping.TileOffsets = make([][2]int, mapping.NumTiles)
	mapping.TileMasks = make([][][]bool, mapping.NumTiles)

	tileIdx := 0
	for tr := 0; tr < numTileRows; tr++ {
		for tc := 0; tc < numTileCols; tc++ {
			mapping.TileOffsets[tileIdx] = [2]int{tr * tileRows, tc * tileCols}

			// Generate skip mask
			mask := make([][]bool, tileRows)
			for i := range mask {
				mask[i] = make([]bool, tileCols)
				srcRow := tr*tileRows + i
				for j := range mask[i] {
					srcCol := tc*tileCols + j
					if srcRow < rows && srcCol < cols {
						mask[i][j] = layer.Weights[srcRow][srcCol] == 0
					} else {
						mask[i][j] = true // Padding
					}
				}
			}
			mapping.TileMasks[tileIdx] = mask
			tileIdx++
		}
	}

	return mapping
}
