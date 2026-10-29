package neural

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
)

const MaxMNISTModelBytes = 32 << 20
const MaxMNISTScalars = 2_000_000

// MNISTModel is an admitted immutable model. Decoding never mutates a network.
// Raw and scale/offset formats retain LoadWeights' reconstruction semantics.
type MNISTModel struct {
	weights        WeightsFile
	SHA256         string
	Bytes, Scalars int
}

func (m *MNISTModel) Hidden() int          { return len(m.weights.Layer1Weights) }
func (m *MNISTModel) HasCalibration() bool { return len(m.weights.SingleLayerWeights) == 10 }
func (m *MNISTModel) TrainingLevels() (int, int, int) {
	return m.weights.QuantLevels, m.weights.Layer1QuantLevels, m.weights.Layer2QuantLevels
}

// DecodeMNISTModel bounds bytes, rows, columns and numeric tokens before any
// constructor or dimension-derived network allocation. Unknown/duplicate fields,
// ragged arrays, non-finite reconstruction and trailing JSON are rejected.
func DecodeMNISTModel(r io.Reader) (*MNISTModel, error) {
	b, e := io.ReadAll(io.LimitReader(r, MaxMNISTModelBytes+1))
	if e != nil {
		return nil, e
	}
	if len(b) > MaxMNISTModelBytes {
		return nil, errors.New("model exceeds 32 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	m := &MNISTModel{Bytes: len(b)}
	w := &m.weights
	token, e := d.Token()
	if e != nil || token != json.Delim('{') {
		return nil, errors.New("model must be an object")
	}
	seen := map[string]bool{}
	number := func() (float64, error) {
		t, e := d.Token()
		if e != nil {
			return 0, e
		}
		n, ok := t.(json.Number)
		if !ok {
			return 0, errors.New("expected finite number")
		}
		v, e := n.Float64()
		if e != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, errors.New("expected finite number")
		}
		return v, nil
	}
	vector := func(max int) ([]float64, error) {
		t, e := d.Token()
		if e == nil && t == nil {
			return nil, nil
		} // absent optional bias vector
		if e != nil || t != json.Delim('[') {
			return nil, errors.New("expected numeric vector")
		}
		v := []float64{}
		for d.More() {
			if len(v) >= max || m.Scalars >= MaxMNISTScalars {
				return nil, errors.New("model scalar/shape limit")
			}
			x, e := number()
			if e != nil {
				return nil, e
			}
			v = append(v, x)
			m.Scalars++
		}
		_, e = d.Token()
		return v, e
	}
	matrix := func(rows, cols int) ([][]float64, error) {
		t, e := d.Token()
		if e != nil || t != json.Delim('[') {
			return nil, errors.New("expected matrix")
		}
		v := [][]float64{}
		for d.More() {
			if len(v) >= rows {
				return nil, errors.New("model row limit")
			}
			row, e := vector(cols)
			if e != nil {
				return nil, e
			}
			v = append(v, row)
		}
		_, e = d.Token()
		return v, e
	}
	for d.More() {
		t, e := d.Token()
		if e != nil {
			return nil, e
		}
		k, ok := t.(string)
		if !ok || seen[k] {
			return nil, errors.New("invalid or duplicate model key")
		}
		seen[k] = true
		switch k {
		case "layer1_weights":
			w.Layer1Weights, e = matrix(1024, 784)
		case "layer2_weights":
			w.Layer2Weights, e = matrix(10, 1024)
		case "single_layer_weights":
			w.SingleLayerWeights, e = matrix(10, 784)
		case "biases1":
			w.Biases1, e = vector(1024)
		case "biases2":
			w.Biases2, e = vector(10)
		case "single_layer_bias":
			w.SingleLayerBias, e = vector(10)
		case "l1_scale":
			w.L1Scale, e = number()
		case "l2_scale":
			w.L2Scale, e = number()
		case "l1_offset":
			w.L1Offset, e = number()
		case "l2_offset":
			w.L2Offset, e = number()
		case "quant_levels", "layer1_quant_levels", "layer2_quant_levels":
			var v float64
			v, e = number()
			if e == nil && (v != math.Trunc(v) || v < 2 || v > MaxDemoLevels) {
				e = errors.New("invalid declared quantization levels")
			}
			if k == "quant_levels" {
				w.QuantLevels = int(v)
			} else if k == "layer1_quant_levels" {
				w.Layer1QuantLevels = int(v)
			} else {
				w.Layer2QuantLevels = int(v)
			}
		default:
			e = fmt.Errorf("unknown model field %s", k)
		}
		if e != nil {
			return nil, e
		}
	}
	if _, e = d.Token(); e != nil {
		return nil, e
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, errors.New("trailing model content")
	}
	h := m.Hidden()
	if h < 1 || h > 1024 {
		return nil, errors.New("hidden size must be 1..1024")
	}
	check := func(a [][]float64, rows, cols int, scale, offset float64) error {
		if len(a) != rows {
			return errors.New("model row mismatch")
		}
		for _, row := range a {
			if len(row) != cols {
				return errors.New("ragged or mismatched model columns")
			}
			for _, v := range row {
				if scale > 0 {
					v = v*scale + offset
				}
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return errors.New("non-finite reconstructed weight")
				}
			}
		}
		return nil
	}
	if w.L1Scale < 0 || w.L2Scale < 0 {
		return nil, errors.New("negative weight scale")
	}
	if e = check(w.Layer1Weights, h, 784, w.L1Scale, w.L1Offset); e != nil {
		return nil, e
	}
	if e = check(w.Layer2Weights, 10, h, w.L2Scale, w.L2Offset); e != nil {
		return nil, e
	}
	if len(w.SingleLayerWeights) > 0 {
		if e = check(w.SingleLayerWeights, 10, 784, 0, 0); e != nil {
			return nil, e
		}
	}
	for _, v := range []struct {
		a []float64
		n int
	}{{w.Biases1, h}, {w.Biases2, 10}, {w.SingleLayerBias, 10}} {
		if len(v.a) != 0 && len(v.a) != v.n {
			return nil, errors.New("bias length mismatch")
		}
	}
	if len(w.SingleLayerBias) > 0 && !m.HasCalibration() {
		return nil, errors.New("calibration bias without weights")
	}
	if (w.Layer1QuantLevels == 0) != (w.Layer2QuantLevels == 0) {
		return nil, errors.New("incomplete per-layer metadata")
	}
	sum := sha256.Sum256(b)
	m.SHA256 = hex.EncodeToString(sum[:])
	return m, nil
}

// NewNetwork creates a CPU-only isolated job network with an explicit noise seed.
// Missing calibration weights remain unavailable, never randomly synthesized.
func (m *MNISTModel) NewNetwork(seed uint64) *DualModeNetwork {
	w := m.weights
	n := NewDualModeNetwork(784, m.Hidden(), 10)
	n.rng = NewRandomSource(seed)
	reconstruct := func(src, dst [][]float64, scale, offset float64) {
		for i, row := range src {
			for j, v := range row {
				if scale > 0 {
					v = v*scale + offset
				}
				dst[i][j] = v
			}
		}
	}
	reconstruct(w.Layer1Weights, n.FPWeights1, w.L1Scale, w.L1Offset)
	reconstruct(w.Layer2Weights, n.FPWeights2, w.L2Scale, w.L2Offset)
	copy(n.FPBias1, w.Biases1)
	copy(n.FPBias2, w.Biases2)
	if m.HasCalibration() {
		reconstruct(w.SingleLayerWeights, n.SingleLayerWeights, 0, 0)
		copy(n.SingleLayerBias, w.SingleLayerBias)
	}
	if w.Layer1QuantLevels > 0 {
		n.Config.Layer1Levels = w.Layer1QuantLevels
		n.Config.Layer2Levels = w.Layer2QuantLevels
	} else if w.QuantLevels > 0 {
		n.Config.Layer1Levels = w.QuantLevels
		n.Config.Layer2Levels = w.QuantLevels
	}
	n.RequantizeWeights()
	return n
}
