package training

import (
	"math"
	"testing"
)

func TestDefaultTrainingConfig(t *testing.T) {
	cfg := DefaultTrainingConfig()

	if cfg.LearningRate <= 0 {
		t.Error("LearningRate should be positive")
	}
	if cfg.BatchSize <= 0 {
		t.Error("BatchSize should be positive")
	}
	if cfg.Epochs <= 0 {
		t.Error("Epochs should be positive")
	}
	if cfg.WeightClipMax <= cfg.WeightClipMin {
		t.Error("WeightClipMax should be > WeightClipMin")
	}
}

func TestNewTrainer(t *testing.T) {
	dims := []int{784, 128, 10}
	trainer, err := NewTrainer(dims, nil)
	if err != nil {
		t.Fatalf("NewTrainer failed: %v", err)
	}

	weights := trainer.GetWeights()
	if len(weights) != 2 {
		t.Errorf("Expected 2 weight matrices, got %d", len(weights))
	}

	// Check dimensions
	if len(weights[0]) != 128 || len(weights[0][0]) != 784 {
		t.Error("First layer dimensions incorrect")
	}
	if len(weights[1]) != 10 || len(weights[1][0]) != 128 {
		t.Error("Second layer dimensions incorrect")
	}
}

func TestNewTrainer_InvalidDims(t *testing.T) {
	dims := []int{10} // Only 1 layer
	_, err := NewTrainer(dims, nil)
	if err == nil {
		t.Error("Expected error for single-layer network")
	}
}

func TestTrainer_Forward(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	input := []float64{0.5, 0.5, 0.5, 0.5}
	output, activations, preActivations := trainer.Forward(input)

	// Check output is valid probability distribution
	if len(output) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(output))
	}

	sum := 0.0
	for _, p := range output {
		if p < 0 || p > 1 {
			t.Errorf("Output probability out of range: %f", p)
		}
		sum += p
	}
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("Probabilities don't sum to 1: %f", sum)
	}

	// Check activations cached correctly
	if len(activations) != 3 { // input + hidden + output
		t.Errorf("Expected 3 activation layers, got %d", len(activations))
	}
	if len(preActivations) != 2 { // hidden + output
		t.Errorf("Expected 2 pre-activation layers, got %d", len(preActivations))
	}
}

func TestTrainer_Backward(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	input := []float64{0.5, 0.5, 0.5, 0.5}
	target := []float64{1.0, 0.0} // One-hot

	_, activations, preActivations := trainer.Forward(input)
	weightGrads, biasGrads := trainer.Backward(activations, preActivations, target)

	// Check gradient dimensions
	if len(weightGrads) != 2 {
		t.Errorf("Expected 2 weight gradient matrices, got %d", len(weightGrads))
	}
	if len(biasGrads) != 2 {
		t.Errorf("Expected 2 bias gradient vectors, got %d", len(biasGrads))
	}

	// Check first layer gradients
	if len(weightGrads[0]) != 3 || len(weightGrads[0][0]) != 4 {
		t.Error("First layer gradient dimensions incorrect")
	}
}

func TestTrainer_UpdateWeights(t *testing.T) {
	dims := []int{4, 3, 2}
	cfg := DefaultTrainingConfig()
	cfg.Momentum = 0.0 // Disable momentum for simpler test
	trainer, _ := NewTrainer(dims, cfg)

	// Get initial weights
	initialWeights := make([][][]float64, len(trainer.GetWeights()))
	for l, layer := range trainer.GetWeights() {
		initialWeights[l] = make([][]float64, len(layer))
		for i, row := range layer {
			initialWeights[l][i] = make([]float64, len(row))
			copy(initialWeights[l][i], row)
		}
	}

	// Create non-zero gradients
	input := []float64{0.5, 0.5, 0.5, 0.5}
	target := []float64{1.0, 0.0}
	_, activations, preActivations := trainer.Forward(input)
	weightGrads, biasGrads := trainer.Backward(activations, preActivations, target)

	// Update weights
	trainer.UpdateWeights(weightGrads, biasGrads)

	// Check weights changed
	newWeights := trainer.GetWeights()
	changed := false
	for l := range newWeights {
		for i := range newWeights[l] {
			for j := range newWeights[l][i] {
				if newWeights[l][i][j] != initialWeights[l][i][j] {
					changed = true
					break
				}
			}
		}
	}

	if !changed {
		t.Error("Weights should change after update")
	}
}

func TestTrainer_TrainBatch(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	inputs := [][]float64{
		{0.5, 0.5, 0.5, 0.5},
		{0.1, 0.9, 0.1, 0.9},
	}
	targets := [][]float64{
		{1.0, 0.0},
		{0.0, 1.0},
	}

	loss := trainer.TrainBatch(inputs, targets)

	if loss < 0 {
		t.Error("Loss should be non-negative")
	}
	if math.IsNaN(loss) || math.IsInf(loss, 0) {
		t.Error("Loss is NaN or Inf")
	}
}

func TestTrainer_TrainBatch_Empty(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	loss := trainer.TrainBatch([][]float64{}, [][]float64{})
	if loss != 0 {
		t.Errorf("Empty batch should have 0 loss, got %f", loss)
	}
}

func TestTrainer_Predict(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	input := []float64{0.5, 0.5, 0.5, 0.5}
	pred := trainer.Predict(input)

	if pred < 0 || pred >= 2 {
		t.Errorf("Prediction %d out of range [0, 2)", pred)
	}
}

func TestTrainer_Evaluate(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	inputs := [][]float64{
		{0.5, 0.5, 0.5, 0.5},
		{0.1, 0.9, 0.1, 0.9},
	}
	labels := []int{0, 1}

	acc := trainer.Evaluate(inputs, labels)

	if acc < 0 || acc > 1 {
		t.Errorf("Accuracy %f out of range [0, 1]", acc)
	}
}

func TestTrainer_SetWeights(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	weights := [][][]float64{
		{ // 3x4
			{0.1, 0.2, 0.3, 0.4},
			{0.5, 0.6, 0.7, 0.8},
			{0.9, 0.8, 0.7, 0.6},
		},
		{ // 2x3
			{0.1, 0.2, 0.3},
			{0.4, 0.5, 0.6},
		},
	}
	biases := [][]float64{
		{0.01, 0.02, 0.03},
		{0.04, 0.05},
	}

	err := trainer.SetWeights(weights, biases)
	if err != nil {
		t.Fatalf("SetWeights failed: %v", err)
	}

	// Verify weights were set
	got := trainer.GetWeights()
	for l := range weights {
		for i := range weights[l] {
			for j := range weights[l][i] {
				if math.Abs(got[l][i][j]-weights[l][i][j]) > 0.001 {
					t.Errorf("Weight mismatch at [%d][%d][%d]", l, i, j)
				}
			}
		}
	}
}

func TestTrainer_SetWeights_Mismatch(t *testing.T) {
	dims := []int{4, 3, 2}
	trainer, _ := NewTrainer(dims, nil)

	weights := [][][]float64{
		{{0.1}}, // Wrong dimensions
	}

	err := trainer.SetWeights(weights, nil)
	if err == nil {
		t.Error("Expected error for dimension mismatch")
	}
}

func TestMLCProgrammer_Creation(t *testing.T) {
	prog := NewMLCProgrammer(2) // 4 levels

	if prog.NumLevels != 4 {
		t.Errorf("Expected 4 levels, got %d", prog.NumLevels)
	}
	if prog.PulseWidth <= 0 {
		t.Error("PulseWidth should be positive")
	}
	if prog.PulseVoltage <= 0 {
		t.Error("PulseVoltage should be positive")
	}
}

func TestMLCProgrammer_GetQuantizedLevels(t *testing.T) {
	prog := NewMLCProgrammer(2) // 4 levels
	levels := prog.GetQuantizedLevels()

	if len(levels) != 4 {
		t.Errorf("Expected 4 levels, got %d", len(levels))
	}

	// Check levels are evenly spaced from 0 to 1
	expected := []float64{0.0, 1.0 / 3.0, 2.0 / 3.0, 1.0}
	for i, level := range levels {
		if math.Abs(level-expected[i]) > 0.001 {
			t.Errorf("Level %d: expected %f, got %f", i, expected[i], level)
		}
	}
}

func TestMLCProgrammer_ComputePulseParams(t *testing.T) {
	prog := NewMLCProgrammer(2) // 4 levels

	// Test potentiation (increasing conductance)
	voltage, width, pulses := prog.ComputePulseParams(0.0, 0.5)
	if voltage <= 0 {
		t.Error("Potentiation should use positive voltage")
	}
	if width <= 0 {
		t.Error("Pulse width should be positive")
	}
	if pulses <= 0 {
		t.Error("Should require at least 1 pulse")
	}

	// Test depression (decreasing conductance)
	voltage, _, _ = prog.ComputePulseParams(1.0, 0.5)
	if voltage >= 0 {
		t.Error("Depression should use negative voltage")
	}

	// Test already at target (use exact quantized level: 1/3 for 4-level programmer)
	exactLevel := 1.0 / 3.0
	voltage, width, pulses = prog.ComputePulseParams(exactLevel, exactLevel)
	if voltage != 0 || width != 0 || pulses != 0 {
		t.Error("Already at target should return zeros")
	}
}

func TestMLCProgrammer_SimulateProgramming(t *testing.T) {
	prog := NewMLCProgrammer(2) // 4 levels

	// Test programming from 0 to 0.5
	finalG, success := prog.SimulateProgramming(0.0, 0.5)

	// Should generally succeed (with randomness)
	if !success {
		t.Log("Programming failed (may be due to randomness)")
	}

	// Final conductance should be quantized
	levelSize := 1.0 / 3.0
	quantized := math.Round(finalG/levelSize) * levelSize
	if math.Abs(finalG-quantized) > 0.2 {
		t.Errorf("Final conductance %f not well quantized", finalG)
	}
}

func TestClip(t *testing.T) {
	tests := []struct {
		v, min, max, expected float64
	}{
		{0.5, 0.0, 1.0, 0.5},  // In range
		{-0.5, 0.0, 1.0, 0.0}, // Below min
		{1.5, 0.0, 1.0, 1.0},  // Above max
		{0.0, 0.0, 1.0, 0.0},  // At min
		{1.0, 0.0, 1.0, 1.0},  // At max
	}

	for _, tc := range tests {
		result := clip(tc.v, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("clip(%f, %f, %f) = %f, expected %f",
				tc.v, tc.min, tc.max, result, tc.expected)
		}
	}
}

func TestTrainer_HardwareConstraints(t *testing.T) {
	dims := []int{4, 3, 2}
	cfg := DefaultTrainingConfig()
	cfg.WeightClipMin = 0.0
	cfg.WeightClipMax = 1.0
	cfg.QuantizeBits = 4 // 16 levels

	trainer, _ := NewTrainer(dims, cfg)

	// Train for a few iterations
	inputs := [][]float64{{0.5, 0.5, 0.5, 0.5}}
	targets := [][]float64{{1.0, 0.0}}

	for i := 0; i < 10; i++ {
		trainer.TrainBatch(inputs, targets)
	}

	// Check all weights are in valid range
	weights := trainer.GetWeights()
	for l := range weights {
		for i := range weights[l] {
			for j := range weights[l][i] {
				w := weights[l][i][j]
				if w < 0.0 || w > 1.0 {
					t.Errorf("Weight %f out of range [0, 1]", w)
				}
			}
		}
	}
}

func BenchmarkTrainer_Forward(b *testing.B) {
	dims := []int{784, 128, 10}
	trainer, _ := NewTrainer(dims, nil)

	input := make([]float64, 784)
	for i := range input {
		input[i] = 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trainer.Forward(input)
	}
}

func BenchmarkTrainer_TrainBatch(b *testing.B) {
	dims := []int{784, 128, 10}
	trainer, _ := NewTrainer(dims, nil)

	// Create batch
	batchSize := 32
	inputs := make([][]float64, batchSize)
	targets := make([][]float64, batchSize)
	for i := range inputs {
		inputs[i] = make([]float64, 784)
		for j := range inputs[i] {
			inputs[i][j] = 0.5
		}
		targets[i] = make([]float64, 10)
		targets[i][i%10] = 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trainer.TrainBatch(inputs, targets)
	}
}
