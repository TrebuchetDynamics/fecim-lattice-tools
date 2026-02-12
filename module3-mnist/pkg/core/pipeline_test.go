package core

import (
	"math"
	"sort"
	"testing"
)

func TestCIMInferencePipelineEndToEnd_Digit0(t *testing.T) {
	net := NewDualModeNetwork(784, 30, 10)
	net.SetNoiseLevel(0.0) // deterministic for regression
	net.SetNumLevels(30)
	net.SetADCBits(8)
	net.SetDACBits(8)

	programDigitZeroWeights(t, net)
	input := syntheticDigitZeroPattern()

	result := net.Infer(input)
	if result == nil {
		t.Fatal("Infer returned nil")
	}

	if len(result.CIMProbabilities) != 10 {
		t.Fatalf("expected 10 CIM probabilities, got %d", len(result.CIMProbabilities))
	}

	if result.CIMPrediction == 0 {
		t.Logf("CIM argmax is digit 0 with confidence %.4f", result.CIMConfidence)
	} else {
		top3 := topKIndices(result.CIMProbabilities, 3)
		inTop3 := false
		for _, idx := range top3 {
			if idx == 0 {
				inTop3 = true
				break
			}
		}
		if !inTop3 {
			t.Fatalf("digit 0 not in CIM top-3: pred=%d top3=%v probs=%v", result.CIMPrediction, top3, result.CIMProbabilities)
		}
		t.Logf("CIM argmax=%d but digit 0 is in top-3 %v", result.CIMPrediction, top3)
	}

	if result.EnergyUsed <= 0 {
		t.Fatalf("expected positive energy, got %.9f µJ", result.EnergyUsed)
	}
	if result.EnergyUsed > 1.0 {
		t.Fatalf("energy out of reasonable range: %.9f µJ (>1.0 µJ)", result.EnergyUsed)
	}

	est := EstimateInferenceEnergyJ(net.Config, net.InputSize, net.HiddenSize, net.OutputSize)
	expectedEnergyMicroJ := est.TotalJ * 1e6
	relErr := math.Abs(result.EnergyUsed-expectedEnergyMicroJ) / expectedEnergyMicroJ
	if relErr > 0.05 {
		t.Fatalf("energy mismatch too large: got %.9f µJ expected %.9f µJ relErr=%.3f", result.EnergyUsed, expectedEnergyMicroJ, relErr)
	}

	t.Logf("pipeline ok: CIM pred=%d conf=%.4f energy=%.9f µJ", result.CIMPrediction, result.CIMConfidence, result.EnergyUsed)
}

func programDigitZeroWeights(t *testing.T, net *DualModeNetwork) {
	t.Helper()

	centerStart, centerEnd := 8, 20

	for h := 0; h < net.HiddenSize; h++ {
		for r := 0; r < 28; r++ {
			for c := 0; c < 28; c++ {
				idx := r*28 + c
				if r >= centerStart && r < centerEnd && c >= centerStart && c < centerEnd {
					net.FPWeights1[h][idx] = 0.15
				} else {
					net.FPWeights1[h][idx] = -0.08
				}
			}
		}
		net.FPBias1[h] = 0.0
	}

	for out := 0; out < net.OutputSize; out++ {
		for h := 0; h < net.HiddenSize; h++ {
			if out == 0 {
				net.FPWeights2[out][h] = 0.35
			} else {
				net.FPWeights2[out][h] = -0.10
			}
		}
		if out == 0 {
			net.FPBias2[out] = 0.25
		} else {
			net.FPBias2[out] = -0.10
		}
	}

	// Program into CIM crossbar (quantized conductance representation)
	net.RequantizeWeights()

	if net.QuantWeights1[0][0] == 0 && net.QuantWeights1[0][14*28+14] == 0 {
		t.Fatal("crossbar programming appears to have failed: representative quantized weights are zero")
	}
}

func syntheticDigitZeroPattern() []float64 {
	in := make([]float64, 28*28)
	for r := 0; r < 28; r++ {
		for c := 0; c < 28; c++ {
			idx := r*28 + c
			if r >= 8 && r < 20 && c >= 8 && c < 20 {
				in[idx] = 0.95 // center pixels high
			} else {
				in[idx] = 0.05 // edges/background low
			}
		}
	}
	return in
}

func topKIndices(values []float64, k int) []int {
	idx := make([]int, len(values))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool {
		return values[idx[i]] > values[idx[j]]
	})
	if k > len(idx) {
		k = len(idx)
	}
	return idx[:k]
}
