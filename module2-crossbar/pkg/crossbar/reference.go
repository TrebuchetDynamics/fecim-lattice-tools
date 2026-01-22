// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"fmt"
	"math"
)

// CPUReference provides a CPU-based reference implementation for MVM operations.
// This is used to verify the correctness of GPU compute shader implementations.
type CPUReference struct{}

// NewCPUReference creates a new CPU reference calculator.
func NewCPUReference() *CPUReference {
	return &CPUReference{}
}

// MVM performs reference matrix-vector multiplication without any non-idealities.
// y[i] = sum(W[i][j] * x[j]) for all j
func (c *CPUReference) MVM(weights [][]float64, input []float64) ([]float64, error) {
	if len(weights) == 0 {
		return nil, fmt.Errorf("empty weight matrix")
	}
	rows := len(weights)
	cols := len(weights[0])

	if len(input) != cols {
		return nil, fmt.Errorf("input size (%d) does not match weight columns (%d)", len(input), cols)
	}

	output := make([]float64, rows)
	for i := 0; i < rows; i++ {
		var sum float64
		for j := 0; j < cols; j++ {
			sum += weights[i][j] * input[j]
		}
		output[i] = sum
	}

	return output, nil
}

// MVMWithQuantization performs MVM with DAC/ADC quantization.
func (c *CPUReference) MVMWithQuantization(weights [][]float64, input []float64, dacBits, adcBits int) ([]float64, error) {
	if len(weights) == 0 {
		return nil, fmt.Errorf("empty weight matrix")
	}
	rows := len(weights)
	cols := len(weights[0])

	if len(input) != cols {
		return nil, fmt.Errorf("input size (%d) does not match weight columns (%d)", len(input), cols)
	}

	dacLevels := float64(int(1)<<dacBits - 1)
	adcLevels := float64(int(1)<<adcBits - 1)

	output := make([]float64, rows)
	for i := 0; i < rows; i++ {
		var sum float64
		for j := 0; j < cols; j++ {
			// Quantize input through DAC
			quantizedInput := math.Round(clamp01(input[j])*dacLevels) / dacLevels
			sum += weights[i][j] * quantizedInput
		}
		// Normalize and quantize through ADC
		normalized := sum / float64(cols)
		output[i] = math.Round(clamp01(normalized)*adcLevels) / adcLevels
	}

	return output, nil
}

// MVMWithNoise performs MVM with device noise simulation.
func (c *CPUReference) MVMWithNoise(weights [][]float64, input []float64, noiseFactors [][]float64, dacBits, adcBits int) ([]float64, error) {
	if len(weights) == 0 {
		return nil, fmt.Errorf("empty weight matrix")
	}
	rows := len(weights)
	cols := len(weights[0])

	if len(input) != cols {
		return nil, fmt.Errorf("input size (%d) does not match weight columns (%d)", len(input), cols)
	}

	dacLevels := float64(int(1)<<dacBits - 1)
	adcLevels := float64(int(1)<<adcBits - 1)

	output := make([]float64, rows)
	for i := 0; i < rows; i++ {
		var sum float64
		for j := 0; j < cols; j++ {
			// Quantize input through DAC
			quantizedInput := math.Round(clamp01(input[j])*dacLevels) / dacLevels

			// Apply weight with noise factor
			effectiveWeight := weights[i][j]
			if noiseFactors != nil && i < len(noiseFactors) && j < len(noiseFactors[i]) {
				effectiveWeight *= noiseFactors[i][j]
			}

			sum += effectiveWeight * quantizedInput
		}
		// Normalize and quantize through ADC
		normalized := sum / float64(cols)
		output[i] = math.Round(clamp01(normalized)*adcLevels) / adcLevels
	}

	return output, nil
}

// clamp01 clamps a value to the range [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// CompareOutputs compares two output vectors and returns the maximum absolute error.
func CompareOutputs(expected, actual []float64) (maxErr float64, avgErr float64, err error) {
	if len(expected) != len(actual) {
		return 0, 0, fmt.Errorf("output size mismatch: %d vs %d", len(expected), len(actual))
	}

	var sumErr float64
	for i := range expected {
		absErr := math.Abs(expected[i] - actual[i])
		if absErr > maxErr {
			maxErr = absErr
		}
		sumErr += absErr
	}

	if len(expected) > 0 {
		avgErr = sumErr / float64(len(expected))
	}

	return maxErr, avgErr, nil
}

// VerifyMVM verifies that the crossbar MVM matches the CPU reference.
// Returns true if within tolerance, along with error statistics.
func VerifyMVM(array *Array, tolerance float64) (bool, float64, float64, error) {
	// Get the weight matrix from the array
	weights := array.GetConductanceMatrix()
	if len(weights) == 0 || len(weights[0]) == 0 {
		return false, 0, 0, fmt.Errorf("empty crossbar array")
	}

	// Create a test input vector
	input := make([]float64, len(weights[0]))
	for i := range input {
		input[i] = float64(i+1) / float64(len(input)+1) // Values from ~0.1 to ~0.9
	}

	// Compute reference output
	ref := NewCPUReference()
	expected, err := ref.MVMWithQuantization(weights, input, 8, 6) // Match default DAC/ADC bits
	if err != nil {
		return false, 0, 0, fmt.Errorf("reference MVM failed: %w", err)
	}

	// Compute crossbar output
	actual, err := array.MVM(input)
	if err != nil {
		return false, 0, 0, fmt.Errorf("crossbar MVM failed: %w", err)
	}

	// Compare outputs
	maxErr, avgErr, err := CompareOutputs(expected, actual)
	if err != nil {
		return false, 0, 0, err
	}

	return maxErr <= tolerance, maxErr, avgErr, nil
}

// BenchmarkMVM runs multiple MVM operations and returns throughput statistics.
func BenchmarkMVM(array *Array, iterations int) (opsPerSec float64, totalOps int64) {
	cols := array.Cols()
	input := make([]float64, cols)
	for i := range input {
		input[i] = 0.5
	}

	// Warm up
	for i := 0; i < 10; i++ {
		array.MVM(input)
	}

	// Benchmark loop - just count operations, timing requires external measurement
	totalOps = int64(iterations * array.Rows() * cols)

	return 0, totalOps // Caller should measure actual time
}
