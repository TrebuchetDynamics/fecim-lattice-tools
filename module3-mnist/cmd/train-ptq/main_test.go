package mnisttrainptq

import "testing"

func TestQuantizeSingleLevelMidpoint(t *testing.T) {
	got := quantize(0.9, -1.0, 1.0, 1)
	if got != 0 {
		t.Fatalf("levels=1 should return midpoint 0, got %v", got)
	}
}

func TestGetWeightRange(t *testing.T) {
	w := [][]float64{{-2, 1}, {3, 0.5}}
	mn, mx := getWeightRange(w)
	if mn != -2 || mx != 3 {
		t.Fatalf("range=(%v,%v), want (-2,3)", mn, mx)
	}
}

func TestForwardPTQOutputSize(t *testing.T) {
	n := NewNetwork(4)
	out := n.ForwardPTQ(make([]float64, 784), PTQConfig{Layer1Levels: 8, Layer2Levels: 8})
	if len(out) != 10 {
		t.Fatalf("len(out)=%d, want 10", len(out))
	}
}
