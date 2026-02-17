package training

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

func TestCrossEntropyLossGradient(t *testing.T) {
	lossFn := CrossEntropyLoss{}
	logits := []float64{2.0, 0.5, -1.0}
	loss, grad := lossFn.Forward(logits, 0)
	if loss <= 0 {
		t.Fatalf("loss should be > 0, got %v", loss)
	}
	sum := 0.0
	for _, g := range grad {
		sum += g
	}
	if math.Abs(sum) > 1e-9 {
		t.Fatalf("softmax-cross-entropy grad should sum to 0, got %e", sum)
	}
}

func TestTrainEpochWithAdamConfig(t *testing.T) {
	layer1, _ := crossbar.NewArray(&crossbar.Config{Rows: 8, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	layer2, _ := crossbar.NewArray(&crossbar.Config{Rows: 10, Cols: 8, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	net := NewMNISTNetwork(layer1, layer2)

	images := make([][]float64, 6)
	labels := make([]int, 6)
	for i := 0; i < 6; i++ {
		images[i] = make([]float64, 784)
		images[i][i] = 1
		labels[i] = i % 10
	}

	cfg := DefaultTrainingConfig()
	cfg.LearningRate = 0.001
	cfg.Optimizer = NewAdamOptimizer(cfg.LearningRate)
	loss := net.TrainEpochWithConfig(images, labels, cfg)
	if math.IsNaN(loss) || math.IsInf(loss, 0) {
		t.Fatalf("invalid loss from Adam epoch: %v", loss)
	}
}
