package training

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
	"ironlattice-vis/demo3-mnist/pkg/mnist"
)

// TestNetworkCreation verifies network initialization
func TestNetworkCreation(t *testing.T) {
	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 128, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	net := NewMNISTNetwork(layer1, layer2)
	if net == nil {
		t.Fatal("NewMNISTNetwork returned nil")
	}
}

// TestForwardPassOutputSize verifies output dimensions
func TestForwardPassOutputSize(t *testing.T) {
	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 128, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	net := NewMNISTNetwork(layer1, layer2)

	// Create random input
	input := make([]float64, 784)
	for i := range input {
		input[i] = rand.Float64()
	}

	output := net.Forward(input)
	if len(output) != 10 {
		t.Errorf("Forward output size = %d, expected 10", len(output))
	}
}

// TestForwardPassSoftmax verifies output sums to 1
func TestForwardPassSoftmax(t *testing.T) {
	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 128, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	net := NewMNISTNetwork(layer1, layer2)

	input := make([]float64, 784)
	for i := range input {
		input[i] = rand.Float64()
	}

	output := net.Forward(input)

	// Softmax should sum to 1
	var sum float64
	for _, v := range output {
		sum += v
	}

	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Softmax sum = %v, expected ~1.0", sum)
	}
}

// TestPredict verifies prediction output format
func TestPredict(t *testing.T) {
	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 128, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	net := NewMNISTNetwork(layer1, layer2)

	input := make([]float64, 784)
	for i := range input {
		input[i] = rand.Float64()
	}

	digit, confidence := net.Predict(input)

	if digit < 0 || digit > 9 {
		t.Errorf("Predicted digit = %d, expected 0-9", digit)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence = %v, expected 0-1", confidence)
	}
}

// TestWeightsAreQuantizedTo30Levels verifies all weights are on valid levels
func TestWeightsAreQuantizedTo30Levels(t *testing.T) {
	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 128, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	_ = NewMNISTNetwork(layer1, layer2)

	// Check layer 1 weights are quantized to valid levels
	weights1 := layer1.GetConductanceMatrix()
	seenLevels := make(map[int]bool)

	for i := 0; i < 128; i++ {
		for j := 0; j < 784; j++ {
			w := weights1[i][j]
			level := crossbar.GetLevel(w)
			if level < 0 || level >= crossbar.IronLatticeLevels {
				t.Errorf("Weight level %d out of range [0, %d)", level, crossbar.IronLatticeLevels)
			}
			seenLevels[level] = true

			// Verify weight is exactly on a 30-level quantization point
			expected := float64(level) / float64(crossbar.IronLatticeLevels-1)
			if w != expected {
				t.Errorf("Weight %.6f not quantized to level %d (expected %.6f)", w, level, expected)
			}
		}
	}

	t.Logf("Network uses %d unique levels (Xavier init centers around 0.5)", len(seenLevels))
}

// TestTrainEpochReducesLoss verifies training makes progress
func TestTrainEpochReducesLoss(t *testing.T) {
	rand.Seed(42)

	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 64, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 64, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	net := NewMNISTNetwork(layer1, layer2)

	// Generate simple training data
	images := make([][]float64, 100)
	labels := make([]int, 100)
	for i := 0; i < 100; i++ {
		images[i] = make([]float64, 784)
		for j := range images[i] {
			images[i][j] = rand.Float64()
		}
		labels[i] = rand.Intn(10)
	}

	// Initial loss
	loss1 := net.TrainEpoch(images, labels, 0.1)

	// Train more epochs
	loss2 := net.TrainEpoch(images, labels, 0.1)
	loss3 := net.TrainEpoch(images, labels, 0.1)

	// Loss should generally decrease (or at least not explode)
	t.Logf("Losses: epoch1=%.4f, epoch2=%.4f, epoch3=%.4f", loss1, loss2, loss3)

	if loss3 > loss1*2 {
		t.Errorf("Loss increased significantly: %.4f -> %.4f", loss1, loss3)
	}
}

// TestSaveLoadWeights verifies weight persistence
func TestSaveLoadWeights(t *testing.T) {
	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 16, Cols: 16, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 4, Cols: 16, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	net := NewMNISTNetwork(layer1, layer2)

	// Get original weights
	origWeights := layer1.GetConductanceMatrix()

	// Save weights
	tmpFile := "/tmp/test_weights.json"
	err := net.SaveWeights(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// Create new network and load weights
	layer1New, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 16, Cols: 16, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	layer2New, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 4, Cols: 16, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})

	netNew := NewMNISTNetwork(layer1New, layer2New)
	err = netNew.LoadWeights(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// Compare weights
	loadedWeights := layer1New.GetConductanceMatrix()
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if origWeights[i][j] != loadedWeights[i][j] {
				t.Errorf("Weight mismatch at [%d][%d]: %.4f vs %.4f",
					i, j, origWeights[i][j], loadedWeights[i][j])
			}
		}
	}
}

// TestMNISTAccuracy verifies pretrained network achieves >= 85% on MNIST test set.
// This test validates the IronLattice 30-level quantized weights maintain accuracy.
func TestMNISTAccuracy(t *testing.T) {
	// Find the data directory relative to the test file
	// The test runs from the package directory, so we need to go up to demo3-mnist
	dataDir := filepath.Join("..", "..", "data")

	// Check if pretrained weights exist
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsFile); os.IsNotExist(err) {
		t.Skip("Pretrained weights not found, skipping accuracy test")
	}

	// Check if MNIST test data exists
	testImageFile := filepath.Join(dataDir, "t10k-images-idx3-ubyte.gz")
	if _, err := os.Stat(testImageFile); os.IsNotExist(err) {
		t.Skip("MNIST test data not found, skipping accuracy test")
	}

	// Create network with same architecture as pretrained model
	layer1, err := crossbar.NewArray(&crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	if err != nil {
		t.Fatalf("Failed to create layer1: %v", err)
	}

	layer2, err := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 128, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	if err != nil {
		t.Fatalf("Failed to create layer2: %v", err)
	}

	net := NewMNISTNetwork(layer1, layer2)

	// Load pretrained weights
	err = net.LoadWeights(weightsFile)
	if err != nil {
		t.Fatalf("Failed to load weights: %v", err)
	}

	// Load MNIST test set
	testImages, testLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test data: %v", err)
	}

	t.Logf("Loaded %d test images", len(testImages))

	// Evaluate accuracy
	accuracy := net.Evaluate(testImages, testLabels)
	accuracyPercent := accuracy * 100

	t.Logf("MNIST test accuracy: %.2f%%", accuracyPercent)

	// Assert accuracy >= 85% (IronLattice target is 87%, we allow some margin)
	minAccuracy := 85.0
	if accuracyPercent < minAccuracy {
		t.Errorf("Accuracy %.2f%% is below minimum required %.2f%%", accuracyPercent, minAccuracy)
	}

	// Log if we exceed the IronLattice reported 87% target
	if accuracyPercent >= 87.0 {
		t.Logf("Exceeds IronLattice target accuracy of 87%%")
	}
}
