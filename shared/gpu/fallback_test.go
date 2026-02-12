package gpu

import (
	"math"
	"strings"
	"testing"
)

// forwardWithCPUFallback executes GPU when available, otherwise computes on CPU.
func forwardWithCPUFallback(net *GPUNetwork, input []float32, layers []LayerWeights) ([]float32, error) {
	if net != nil && net.IsAvailable() {
		return net.Forward(input, layers)
	}

	current := input
	for _, layer := range layers {
		relu := layer.Activation == ActivationReLU
		current = cpuDenseLayer(layer.Weights, layer.Bias, current, layer.Rows, layer.Cols, relu)
	}
	return current, nil
}

func TestGPUFallback_UnavailableFallsBackToCPUGracefully(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	layers := []LayerWeights{
		{
			Weights:    createTestWeights(4, 3),
			Bias:       createTestBias(4),
			Rows:       4,
			Cols:       3,
			Activation: ActivationReLU,
		},
		{
			Weights:    createTestWeights(2, 4),
			Bias:       createTestBias(2),
			Rows:       2,
			Cols:       4,
			Activation: ActivationNone,
		},
	}
	input := []float32{0.25, -0.5, 1.0}

	if net.IsAvailable() {
		t.Skip("GPU available; fallback-specific path requires unavailable GPU")
	}

	got, err := forwardWithCPUFallback(net, input, layers)
	if err != nil {
		t.Fatalf("fallback execution returned error: %v", err)
	}

	hidden := cpuDenseLayer(layers[0].Weights, layers[0].Bias, input, layers[0].Rows, layers[0].Cols, true)
	want := cpuDenseLayer(layers[1].Weights, layers[1].Bias, hidden, layers[1].Rows, layers[1].Cols, false)
	assertFloat32SliceEqual(t, want, got, 1e-6, "CPU fallback output")
}

func TestGPUFallback_CPUAndGPUResultsMatchWithinTolerance(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available; parity validation requires GPU")
	}

	rows, cols := 8, 16
	weights := createRandomWeights(rows, cols, 1234)
	bias := createRandomBias(rows, 1235)
	input := createRandomInput(cols, 1236)

	gpuOut, err := net.denseLayer.Forward(weights, bias, input, rows, cols, ActivationReLU)
	if err != nil {
		t.Fatalf("GPU forward failed: %v", err)
	}
	cpuOut := cpuDenseLayer(weights, bias, input, rows, cols, true)

	assertFloat32SliceEqual(t, cpuOut, gpuOut, 1e-4, "CPU vs GPU parity")
}

func TestGPUFallback_ErrorMessageWhenGPUInitFails(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if net.IsAvailable() {
		t.Skip("GPU available; cannot validate init-failure message")
	}

	layers := []LayerWeights{{
		Weights:    []float32{1, 2, 3, 4, 5, 6},
		Bias:       []float32{0.1, 0.2},
		Rows:       2,
		Cols:       3,
		Activation: ActivationNone,
	}}
	_, err := net.Forward([]float32{1, 2, 3}, layers)
	if err == nil {
		t.Fatal("expected error when GPU unavailable")
	}

	errMsg := strings.ToLower(err.Error())
	if !strings.Contains(errMsg, "gpu") || !strings.Contains(errMsg, "not available") {
		t.Fatalf("error message should be informative, got: %q", err.Error())
	}
}

func TestGPUFallback_NoPanicsOnGPUUnavailability(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if net.IsAvailable() {
		t.Skip("GPU available; panic-safety check targets unavailability path")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic on GPU unavailability path: %v", r)
		}
	}()

	layers := []LayerWeights{{
		Weights:    createTestWeights(2, 3),
		Bias:       createTestBias(2),
		Rows:       2,
		Cols:       3,
		Activation: ActivationNone,
	}}

	for i := 0; i < 20; i++ {
		input := []float32{float32(i), float32(i) * 0.5, -float32(i)}
		out, err := forwardWithCPUFallback(net, input, layers)
		if err != nil {
			t.Fatalf("iteration %d fallback failed: %v", i, err)
		}
		if len(out) != 2 {
			t.Fatalf("iteration %d unexpected output length: %d", i, len(out))
		}
		if math.IsNaN(float64(out[0])) || math.IsNaN(float64(out[1])) {
			t.Fatalf("iteration %d output contains NaN: %v", i, out)
		}
	}
}
