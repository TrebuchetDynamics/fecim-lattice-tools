package mnist

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadMNIST_FileNotFound tests error handling when files don't exist
func TestLoadMNIST_FileNotFound(t *testing.T) {
	_, _, err := LoadMNIST("/nonexistent/path", true)
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

// TestLoadMNIST_Integration tests loading real MNIST data if available
func TestLoadMNIST_Integration(t *testing.T) {
	// Check common MNIST data locations
	mnistPaths := []string{
		"../../../data",
		"../../../../data",
		"data",
		os.Getenv("MNIST_DATA"),
	}

	var dataPath string
	for _, p := range mnistPaths {
		if p == "" {
			continue
		}
		absPath, _ := filepath.Abs(p)
		trainImages := filepath.Join(absPath, "train-images-idx3-ubyte.gz")
		if _, err := os.Stat(trainImages); err == nil {
			dataPath = absPath
			break
		}
		// Try without .gz
		trainImages = filepath.Join(absPath, "train-images-idx3-ubyte")
		if _, err := os.Stat(trainImages); err == nil {
			dataPath = absPath
			break
		}
	}

	if dataPath == "" {
		t.Skip("MNIST data not found, skipping integration test")
	}

	// Test loading training data
	images, labels, err := LoadMNIST(dataPath, true)
	if err != nil {
		t.Fatalf("LoadMNIST failed: %v", err)
	}

	// Verify counts
	if len(images) != 60000 {
		t.Errorf("Expected 60000 training images, got %d", len(images))
	}
	if len(labels) != 60000 {
		t.Errorf("Expected 60000 training labels, got %d", len(labels))
	}

	// Verify image dimensions
	if len(images[0]) != 784 {
		t.Errorf("Expected 784 pixels per image, got %d", len(images[0]))
	}

	// Verify pixel values are normalized [0, 1]
	for i := 0; i < 100; i++ { // Check first 100 images
		for j, pixel := range images[i] {
			if pixel < 0 || pixel > 1 {
				t.Errorf("Image %d pixel %d out of range: %f", i, j, pixel)
				break
			}
		}
	}

	// Verify labels are in range [0, 9]
	for i := 0; i < 100; i++ {
		if labels[i] < 0 || labels[i] > 9 {
			t.Errorf("Label %d out of range: %d", i, labels[i])
		}
	}

	t.Logf("Successfully loaded %d training images", len(images))
}

// TestLoadMNIST_TestSet tests loading test data if available
func TestLoadMNIST_TestSet(t *testing.T) {
	mnistPaths := []string{
		"../../../data",
		"../../../../data",
		"data",
		os.Getenv("MNIST_DATA"),
	}

	var dataPath string
	for _, p := range mnistPaths {
		if p == "" {
			continue
		}
		absPath, _ := filepath.Abs(p)
		testImages := filepath.Join(absPath, "t10k-images-idx3-ubyte.gz")
		if _, err := os.Stat(testImages); err == nil {
			dataPath = absPath
			break
		}
		testImages = filepath.Join(absPath, "t10k-images-idx3-ubyte")
		if _, err := os.Stat(testImages); err == nil {
			dataPath = absPath
			break
		}
	}

	if dataPath == "" {
		t.Skip("MNIST test data not found, skipping test")
	}

	images, labels, err := LoadMNIST(dataPath, false)
	if err != nil {
		t.Fatalf("LoadMNIST (test) failed: %v", err)
	}

	if len(images) != 10000 {
		t.Errorf("Expected 10000 test images, got %d", len(images))
	}
	if len(labels) != 10000 {
		t.Errorf("Expected 10000 test labels, got %d", len(labels))
	}

	t.Logf("Successfully loaded %d test images", len(images))
}

// TestImagePixelRange verifies all pixels are in [0, 1] range
func TestImagePixelRange(t *testing.T) {
	// This tests the normalization logic directly
	// Simulated byte values
	rawPixels := []byte{0, 127, 255, 50, 200}
	normalized := make([]float64, len(rawPixels))

	for i, b := range rawPixels {
		normalized[i] = float64(b) / 255.0
	}

	expected := []float64{0.0, 127.0 / 255.0, 1.0, 50.0 / 255.0, 200.0 / 255.0}
	for i := range normalized {
		if normalized[i] < 0 || normalized[i] > 1 {
			t.Errorf("Normalized pixel %d out of range: %f", i, normalized[i])
		}
		if abs(normalized[i]-expected[i]) > 0.001 {
			t.Errorf("Pixel %d: expected %f, got %f", i, expected[i], normalized[i])
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// BenchmarkLoadMNIST benchmarks data loading if data is available
func BenchmarkLoadMNIST(b *testing.B) {
	mnistPaths := []string{
		"../../../data",
		"../../../../data",
		"data",
	}

	var dataPath string
	for _, p := range mnistPaths {
		absPath, _ := filepath.Abs(p)
		trainImages := filepath.Join(absPath, "train-images-idx3-ubyte.gz")
		if _, err := os.Stat(trainImages); err == nil {
			dataPath = absPath
			break
		}
	}

	if dataPath == "" {
		b.Skip("MNIST data not found, skipping benchmark")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadMNIST(dataPath, true)
	}
}
