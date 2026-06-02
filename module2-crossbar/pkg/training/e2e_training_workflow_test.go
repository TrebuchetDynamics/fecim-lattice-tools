package training

import (
	"math"
	"math/rand"
	"testing"
)

func TestModule2TrainingE2EWideHardwareAwareTrainingWorkflow(t *testing.T) {
	rand.Seed(42)
	cfg := &TrainingConfig{LearningRate: 0.05, WeightDecay: 1e-4, Momentum: 0.2, BatchSize: 2, Epochs: 2, WeightClipMin: 0.05, WeightClipMax: 0.95, UpdateNoise: 0, QuantizeBits: 3, AsymmetryRatio: 0.8}
	trainer, err := NewTrainer([]int{4, 5, 3}, cfg)
	if err != nil {
		t.Fatalf("NewTrainer error = %v", err)
	}
	inputs := [][]float64{{1, 0, 0.5, 0.25}, {0, 1, 0.25, 0.5}, {0.5, 0.5, 1, 0}}
	targets := [][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	labels := []int{0, 1, 2}

	beforeWeights := cloneWeightsE2E(trainer.GetWeights())
	beforeBiases := cloneBiasesE2E(trainer.GetBiases())
	output, activations, preActivations := trainer.Forward(inputs[0])
	assertTrainingProbabilityE2E(t, output, 3)
	if len(activations) != 3 || len(preActivations) != 2 {
		t.Fatalf("activation cache sizes = %d/%d, want 3/2", len(activations), len(preActivations))
	}
	wgrads, bgrads := trainer.Backward(activations, preActivations, targets[0])
	if len(wgrads) != 2 || len(bgrads) != 2 {
		t.Fatalf("gradient layer counts = %d/%d", len(wgrads), len(bgrads))
	}

	loss1 := trainer.TrainBatch(inputs, targets)
	loss2 := trainer.TrainBatch(inputs, targets)
	if !(loss1 > 0) || !(loss2 > 0) || math.IsNaN(loss1) || math.IsNaN(loss2) {
		t.Fatalf("losses invalid: %g %g", loss1, loss2)
	}
	afterWeights := trainer.GetWeights()
	if sameWeightsE2E(beforeWeights, afterWeights) {
		t.Fatalf("training did not change weights")
	}
	assertWeightsWithinQuantizedBoundsE2E(t, afterWeights, 0, 1, cfg.QuantizeBits)
	if sameBiasesE2E(beforeBiases, trainer.GetBiases()) {
		t.Fatalf("training did not change biases")
	}
	pred := trainer.Predict(inputs[0])
	if pred < 0 || pred >= 3 {
		t.Fatalf("Predict returned %d, want class range [0,3)", pred)
	}
	acc := trainer.Evaluate(inputs, labels)
	if math.IsNaN(acc) || acc < 0 || acc > 1 {
		t.Fatalf("Evaluate accuracy invalid: %g", acc)
	}

	weights := cloneWeightsE2E(afterWeights)
	biases := cloneBiasesE2E(trainer.GetBiases())
	clone, err := NewTrainer([]int{4, 5, 3}, cfg)
	if err != nil {
		t.Fatalf("clone NewTrainer error = %v", err)
	}
	if err := clone.SetWeights(weights, biases); err != nil {
		t.Fatalf("SetWeights cloned weights error = %v", err)
	}
	cloneOut, _, _ := clone.Forward(inputs[1])
	assertTrainingProbabilityE2E(t, cloneOut, 3)
}

func TestModule2TrainingE2EMLCProgrammingMatrix(t *testing.T) {
	rand.Seed(9)
	for _, bits := range []int{2, 3, 4} {
		t.Run("bits", func(t *testing.T) {
			programmer := NewMLCProgrammer(bits)
			levels := programmer.GetQuantizedLevels()
			if len(levels) != 1<<bits || levels[0] != 0 || levels[len(levels)-1] != 1 {
				t.Fatalf("levels for bits=%d invalid: %v", bits, levels)
			}
			for i := 1; i < len(levels); i++ {
				if levels[i] <= levels[i-1] {
					t.Fatalf("levels not increasing: %v", levels)
				}
			}
			for _, target := range []float64{0, 0.2, 0.5, 0.8, 1} {
				voltage, width, pulses := programmer.ComputePulseParams(0.5, target)
				if target != 0.5 && width <= 0 && pulses != 0 {
					t.Fatalf("invalid pulse params target %.2f: V=%g width=%g pulses=%d", target, voltage, width, pulses)
				}
				finalG, _ := programmer.SimulateProgramming(0.5, target)
				if math.IsNaN(finalG) || finalG < 0 || finalG > 1 {
					t.Fatalf("final conductance for bits=%d target=%.2f invalid: %g", bits, target, finalG)
				}
			}
		})
	}
}

func TestModule2TrainingE2EInvalidInputsAndSetWeightsIsolation(t *testing.T) {
	if _, err := NewTrainer([]int{3}, nil); err == nil {
		t.Fatal("NewTrainer with one dim returned nil error")
	}
	trainer, err := NewTrainer([]int{2, 3, 2}, DefaultTrainingConfig())
	if err != nil {
		t.Fatalf("NewTrainer valid error = %v", err)
	}
	before := cloneWeightsE2E(trainer.GetWeights())
	if loss := trainer.TrainBatch(nil, nil); loss != 0 {
		t.Fatalf("TrainBatch empty loss = %g, want 0", loss)
	}
	if err := trainer.SetWeights([][][]float64{{{0.1}}}, nil); err == nil {
		t.Fatal("SetWeights wrong layer count returned nil error")
	}
	bad := cloneWeightsE2E(trainer.GetWeights())
	bad[0] = append(bad[0], []float64{0.1, 0.2})
	if err := trainer.SetWeights(bad, nil); err == nil {
		t.Fatal("SetWeights wrong output shape returned nil error")
	}
	if !sameWeightsE2E(before, trainer.GetWeights()) {
		t.Fatal("invalid training operations mutated weights")
	}
}

func cloneWeightsE2E(in [][][]float64) [][][]float64 {
	out := make([][][]float64, len(in))
	for l := range in {
		out[l] = make([][]float64, len(in[l]))
		for r := range in[l] {
			out[l][r] = append([]float64(nil), in[l][r]...)
		}
	}
	return out
}

func cloneBiasesE2E(in [][]float64) [][]float64 {
	out := make([][]float64, len(in))
	for i := range in {
		out[i] = append([]float64(nil), in[i]...)
	}
	return out
}

func sameWeightsE2E(a, b [][][]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for l := range a {
		if len(a[l]) != len(b[l]) {
			return false
		}
		for r := range a[l] {
			if len(a[l][r]) != len(b[l][r]) {
				return false
			}
			for c := range a[l][r] {
				if a[l][r][c] != b[l][r][c] {
					return false
				}
			}
		}
	}
	return true
}

func sameBiasesE2E(a, b [][]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func assertTrainingProbabilityE2E(t *testing.T, out []float64, wantLen int) {
	t.Helper()
	if len(out) != wantLen {
		t.Fatalf("output len = %d, want %d", len(out), wantLen)
	}
	sum := 0.0
	for i, v := range out {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			t.Fatalf("output[%d] invalid: %g", i, v)
		}
		sum += v
	}
	if math.Abs(sum-1) > 1e-9 {
		t.Fatalf("output sum = %.12f, want 1", sum)
	}
}

func assertWeightsWithinQuantizedBoundsE2E(t *testing.T, weights [][][]float64, min, max float64, bits int) {
	t.Helper()
	levels := float64(int(1) << uint(bits))
	for l := range weights {
		for r := range weights[l] {
			for c, w := range weights[l][r] {
				if math.IsNaN(w) || w < min || w > max {
					t.Fatalf("weight[%d][%d][%d]=%g outside [%g,%g]", l, r, c, w, min, max)
				}
				q := math.Round(w*levels) / levels
				if math.Abs(w-q) > 1e-12 {
					t.Fatalf("weight[%d][%d][%d]=%g not quantized to %d bits", l, r, c, w, bits)
				}
			}
		}
	}
}
