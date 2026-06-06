package network

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

func TestModule2NetworkE2EWideForwardLoadWeightsMatrix(t *testing.T) {
	rand.Seed(7)
	configs := []Config{
		{InputSize: 4, HiddenSize: 5, OutputSize: 3, NumLayers: 3},
		{InputSize: 6, HiddenSize: 4, OutputSize: 2, NumLayers: 4},
		{InputSize: 3, HiddenSize: 3, OutputSize: 4, NumLayers: 2},
	}
	for _, cfg := range configs {
		t.Run(fmt.Sprintf("%dx%dx%d_layers%d", cfg.InputSize, cfg.HiddenSize, cfg.OutputSize, cfg.NumLayers), func(t *testing.T) {
			base, err := crossbar.NewArray(&crossbar.Config{Rows: cfg.HiddenSize, Cols: cfg.InputSize, ADCBits: 8, DACBits: 8, NoiseLevel: 0})
			if err != nil {
				t.Fatalf("base NewArray: %v", err)
			}
			defer base.Destroy()
			net, err := NewNetwork(&cfg, base)
			if err != nil {
				t.Fatalf("NewNetwork(%+v) error = %v", cfg, err)
			}
			if net.GetLayerCount() != cfg.NumLayers-1 {
				t.Fatalf("layer count = %d, want %d", net.GetLayerCount(), cfg.NumLayers-1)
			}
			for layer := 0; layer < net.GetLayerCount(); layer++ {
				in, out, err := net.GetLayerDimensions(layer)
				if err != nil {
					t.Fatalf("GetLayerDimensions(%d): %v", layer, err)
				}
				wantIn, wantOut := expectedLayerDimsE2E(cfg, layer)
				if in != wantIn || out != wantOut {
					t.Fatalf("layer %d dims = %dx%d, want %dx%d", layer, in, out, wantIn, wantOut)
				}
			}

			weights, biases := deterministicNetworkWeightsE2E(cfg)
			if err := net.LoadWeights(weights, biases); err != nil {
				t.Fatalf("LoadWeights() error = %v", err)
			}
			input := deterministicNetworkInputE2E(cfg.InputSize)
			out1, err := net.Forward(input)
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}
			assertProbabilityVectorE2E(t, out1, cfg.OutputSize)
			opsAfterFirst := net.GetOpsCount()
			wantOps := int64(0)
			for layer := 0; layer < net.GetLayerCount(); layer++ {
				in, out, _ := net.GetLayerDimensions(layer)
				wantOps += int64(in * out)
			}
			if opsAfterFirst != wantOps {
				t.Fatalf("ops after first forward = %d, want %d", opsAfterFirst, wantOps)
			}
			out2, err := net.Forward(input)
			if err != nil {
				t.Fatalf("second Forward() error = %v", err)
			}
			assertProbabilityVectorE2E(t, out2, cfg.OutputSize)
			if net.GetOpsCount() != 2*wantOps {
				t.Fatalf("ops after second forward = %d, want %d", net.GetOpsCount(), 2*wantOps)
			}
		})
	}
}

func TestModule2NetworkE2EInvalidLoadAndForwardBoundaries(t *testing.T) {
	cfg := Config{InputSize: 4, HiddenSize: 3, OutputSize: 2, NumLayers: 3}
	base, err := crossbar.NewArray(&crossbar.Config{Rows: 3, Cols: 4, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("base NewArray: %v", err)
	}
	defer base.Destroy()
	net, err := NewNetwork(&cfg, base)
	if err != nil {
		t.Fatalf("NewNetwork: %v", err)
	}
	if _, _, err := net.GetLayerDimensions(-1); err == nil {
		t.Fatal("GetLayerDimensions(-1) returned nil error")
	}
	if _, _, err := net.GetLayerDimensions(net.GetLayerCount()); err == nil {
		t.Fatal("GetLayerDimensions(out of range) returned nil error")
	}
	if err := net.LoadWeights([][][]float64{{{0.1}}}, nil); err == nil {
		t.Fatal("LoadWeights wrong layer count returned nil error")
	}
	weights, biases := deterministicNetworkWeightsE2E(cfg)
	weights[0] = append(weights[0], []float64{0.1, 0.2, 0.3, 0.4})
	if err := net.LoadWeights(weights, biases); err == nil {
		t.Fatal("LoadWeights wrong layer shape returned nil error")
	}
	if _, err := net.Forward([]float64{1, 2, 3, 4, 5}); err == nil {
		t.Fatal("Forward oversized input returned nil error")
	}
	if _, err := NewNetwork(&Config{InputSize: 2, HiddenSize: 2, OutputSize: 2, NumLayers: 1}, base); err == nil {
		t.Fatal("NewNetwork with one layer returned nil error")
	}
}

func expectedLayerDimsE2E(cfg Config, layer int) (int, int) {
	sizes := make([]int, cfg.NumLayers)
	sizes[0] = cfg.InputSize
	sizes[cfg.NumLayers-1] = cfg.OutputSize
	for i := 1; i < cfg.NumLayers-1; i++ {
		sizes[i] = cfg.HiddenSize
	}
	return sizes[layer], sizes[layer+1]
}

func deterministicNetworkWeightsE2E(cfg Config) ([][][]float64, [][]float64) {
	layers := cfg.NumLayers - 1
	weights := make([][][]float64, layers)
	biases := make([][]float64, layers)
	for l := 0; l < layers; l++ {
		in, out := expectedLayerDimsE2E(cfg, l)
		weights[l] = make([][]float64, out)
		biases[l] = make([]float64, out)
		for r := 0; r < out; r++ {
			weights[l][r] = make([]float64, in)
			biases[l][r] = float64((l+r)%3) * 0.01
			for c := 0; c < in; c++ {
				weights[l][r][c] = float64(((l+1)*(r+2)*(c+3))%9) / 10.0
			}
		}
	}
	return weights, biases
}

func deterministicNetworkInputE2E(n int) []float64 {
	input := make([]float64, n)
	for i := range input {
		input[i] = float64((i*2+1)%5) / 4.0
	}
	return input
}

func assertProbabilityVectorE2E(t *testing.T, vector []float64, wantLen int) {
	t.Helper()
	if len(vector) != wantLen {
		t.Fatalf("probability length = %d, want %d", len(vector), wantLen)
	}
	sum := 0.0
	for i, value := range vector {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			t.Fatalf("probability[%d] invalid: %g", i, value)
		}
		sum += value
	}
	if math.Abs(sum-1) > 1e-9 {
		t.Fatalf("probability sum = %.12f, want 1", sum)
	}
}
