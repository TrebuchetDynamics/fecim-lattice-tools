// Package mnist provides CPU-first local MNIST jobs over shared/neural.
package mnist

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	"fecim-lattice-tools/shared/neural"
)

const JobTimeout = 15 * time.Minute
const MaxAllocationBytes int64 = 256 << 20
const MaxMACs int64 = 100_000_000_000
const Warning = "Educational CPU simulation; not a device measurement. Accuracy requires actual labelled data; no accuracy threshold is claimed."

type Config struct {
	Levels       int     `json:"levels"`
	Layer1       int     `json:"layer1"`
	Layer2       int     `json:"layer2"`
	ADC          int     `json:"adc_bits"`
	DAC          int     `json:"dac_bits"`
	Noise        float64 `json:"noise"`
	Seed         uint64  `json:"seed"`
	Quantization string  `json:"quantization"`
	Architecture string  `json:"architecture"`
	IRDrop       bool    `json:"ir_drop"`
	Sneak        bool    `json:"sneak"`
}

func Defaults() Config {
	return Config{Levels: 30, Layer1: 30, Layer2: 30, ADC: 8, DAC: 8, Noise: 0.01, Seed: 42, Quantization: "uniform-ptq", Architecture: "784-H-10"}
}
func (c Config) validate(m *neural.MNISTModel) error {
	for _, v := range []int{c.Levels, c.Layer1, c.Layer2} {
		if v < 2 || v > neural.MaxDemoLevels {
			return errors.New("levels outside 2..MaxDemoLevels")
		}
	}
	if !finite(c.Noise) || c.Noise < 0 || c.Noise > 0.2 || c.ADC < 3 || c.ADC > 16 || c.DAC < 3 || c.DAC > 16 {
		return errors.New("noise must be 0..0.20 and ADC/DAC bits 3..16")
	}
	if c.Quantization != "uniform-ptq" && c.Quantization != "per-layer-ptq" && c.Quantization != "model-qat" {
		return errors.New("unknown quantization method")
	}
	if c.Quantization == "model-qat" {
		if m == nil {
			return errors.New("QAT requires an explicitly selected model")
		}
		q, a, b := m.TrainingLevels()
		if q == 0 && (a == 0 || b == 0) {
			return errors.New("model declares no QAT/PTQ levels; select PTQ explicitly")
		}
	}
	if c.Architecture != "784-H-10" && c.Architecture != "784-10" {
		return errors.New("unknown architecture")
	}
	if c.Architecture == "784-10" && (m == nil || !m.HasCalibration()) {
		return errors.New("calibration unavailable: selected model has no direct weights")
	}
	return nil
}
func (c Config) configure(n *neural.DualModeNetwork, m *neural.MNISTModel) {
	n.SetNumLevels(c.Levels)
	n.SetPerLayerQuant(false)
	if c.Quantization == "per-layer-ptq" {
		n.SetPerLayerLevels(c.Layer1, c.Layer2)
	}
	if c.Quantization == "model-qat" {
		q, a, b := m.TrainingLevels()
		if a > 0 && b > 0 {
			n.SetPerLayerLevels(a, b)
		} else {
			n.SetNumLevels(q)
		}
	}
	n.SetNoiseLevel(c.Noise)
	n.SetADCBits(c.ADC)
	n.SetDACBits(c.DAC)
	n.SetSingleLayer(c.Architecture == "784-10")
	n.Config.IRDrop = c.IRDrop
	n.Config.EnableSneak = c.Sneak
}
func configure(c Config, input map[string]string) (Config, error) {
	for k, v := range input {
		var e error
		switch k {
		case "levels":
			c.Levels, e = strconv.Atoi(v)
		case "layer1":
			c.Layer1, e = strconv.Atoi(v)
		case "layer2":
			c.Layer2, e = strconv.Atoi(v)
		case "adc_bits":
			c.ADC, e = strconv.Atoi(v)
		case "dac_bits":
			c.DAC, e = strconv.Atoi(v)
		case "noise":
			c.Noise, e = strconv.ParseFloat(v, 64)
		case "seed":
			c.Seed, e = strconv.ParseUint(v, 10, 64)
		case "quantization":
			c.Quantization = v
		case "architecture":
			c.Architecture = v
		case "ir_drop":
			c.IRDrop, e = strconv.ParseBool(v)
		case "sneak":
			c.Sneak, e = strconv.ParseBool(v)
		case "preset":
			switch v {
			case "ideal":
				c.Noise = 0
				c.ADC = 16
				c.DAC = 16
				c.IRDrop = false
				c.Sneak = false
			case "hardware":
				c.Noise = .01
				c.ADC = 8
				c.DAC = 8
				c.IRDrop = true
				c.Sneak = true
			case "noisy":
				c.Noise = .1
				c.ADC = 8
				c.DAC = 8
				c.IRDrop = true
				c.Sneak = true
			default:
				e = errors.New("unknown preset")
			}
		default:
			e = errors.New("unknown MNIST setting")
		}
		if e != nil {
			return c, e
		}
	}
	return c, nil
}

// Finite stored weights can still overflow the existing quantizer/statistics.
// Check the prospective inspection network before committing any session state.
func prepareInspection(m *neural.MNISTModel, c Config) (*neural.DualModeNetwork, error) {
	n := m.NewNetwork(c.Seed)
	c.configure(n, m)
	for _, matrix := range [][][]float64{n.QuantWeights1, n.QuantWeights2, n.QuantSingleLayerWeights, {n.QuantBias1, n.QuantBias2, n.QuantSingleLayerBias}} {
		for _, row := range matrix {
			for _, v := range row {
				if !finite(v) {
					return nil, errors.New("model/configuration causes non-finite quantization")
				}
			}
		}
	}
	a, b := n.GetQuantizationStats()
	for _, s := range []neural.QuantizationStats{a, b} {
		if !finite(s.MSE) || !finite(s.MaxError) {
			return nil, errors.New("model/configuration causes non-finite inspection statistics")
		}
	}
	return n, nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func digest(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// AdmitWork is a pre-constructor budget, not a latency or measured RSS claim.
func AdmitWork(hidden, images, cases int, sweep bool) error {
	if hidden < 1 || hidden > 1024 || images < 1 || images > 10000 || cases < 1 || cases > 64 || (!sweep && cases != 1) || (sweep && images > 500) {
		return errors.New("MNIST shape/corpus/sweep admission limit")
	}
	macs := 2 * int64(images) * int64(cases) * int64(794*hidden)
	// Include resident and isolated FP/quantized models, raw token/IDX buffers,
	// transactional dataset replacement, decoded corpus and inference scratch.
	allocation := int64(2*32<<20) + 2*16<<20 + int64(2*images*784*8) + int64(6*(794*hidden+7840+hidden+20)*8)
	if macs > MaxMACs || allocation > MaxAllocationBytes {
		return errors.New("MNIST job exceeds MAC/allocation budget")
	}
	return nil
}

type Case struct {
	Value          float64              `json:"value"`
	Config         Config               `json:"config"`
	ResolvedConfig neural.NetworkConfig `json:"resolved_config"`
	Total          int                  `json:"total"`
	FPCorrect      int                  `json:"fp_correct"`
	CIMCorrect     int                  `json:"cim_correct"`
	FPConfusion    [10][10]int          `json:"fp_confusion"`
	CIMConfusion   [10][10]int          `json:"cim_confusion"`
}
type Result struct {
	Version        string                  `json:"version"`
	ModelSHA       string                  `json:"model_sha256"`
	DatasetSHA     string                  `json:"dataset_sha256,omitempty"`
	InputSHA       string                  `json:"input_sha256,omitempty"`
	RawInputSHA    string                  `json:"raw_input_sha256,omitempty"`
	Source         string                  `json:"source"`
	Preprocessing  string                  `json:"preprocessing"`
	Config         Config                  `json:"config"`
	ResolvedConfig neural.NetworkConfig    `json:"resolved_config"`
	TrainingLevels [3]int                  `json:"training_levels"`
	Start          int                     `json:"start"`
	Count          int                     `json:"count"`
	CorpusCount    int                     `json:"corpus_count"`
	Sweep          string                  `json:"sweep,omitempty"`
	Inference      *neural.InferenceResult `json:"inference,omitempty"`
	Cases          []Case                  `json:"cases,omitempty"`
}
