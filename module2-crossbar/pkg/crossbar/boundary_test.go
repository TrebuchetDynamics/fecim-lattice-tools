package crossbar

import (
	"fecim-lattice-tools/shared/physics"
	"fmt"
	"math"
	"testing"
)

// TestSingleCellArray verifies 1x1 array edge case
func TestSingleCellArray(t *testing.T) {
	cfg := &Config{
		Rows:       1,
		Cols:       1,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create 1x1 array: %v", err)
	}
	defer arr.Destroy()

	arr.ProgramWeight(0, 0, 0.5)

	t.Run("MVM works", func(t *testing.T) {
		output, err := arr.MVM([]float64{1.0})
		if err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
		if len(output) != 1 {
			t.Errorf("Output length = %d, want 1", len(output))
		}
		if math.IsNaN(output[0]) || math.IsInf(output[0], 0) {
			t.Errorf("Output is NaN or Inf")
		}
	})

	t.Run("IR drop analysis works", func(t *testing.T) {
		analysis := arr.AnalyzeIRDrop([]float64{1.0}, nil)
		// 1x1 should have minimal IR drop
		t.Logf("1x1 IR drop: %.4f%%", analysis.MaxIRDrop*100)
	})

	t.Run("Sneak path analysis works", func(t *testing.T) {
		analysis := arr.AnalyzeSneakPaths(0, 0)
		// 1x1 should have zero sneak - using TotalSneak field
		if analysis.TotalSneak != 0 {
			t.Logf("1x1 sneak current (expected minimal): %e", analysis.TotalSneak)
		}
	})
}

// TestExtremeAspectRatioArrays verifies non-square arrays
func TestExtremeAspectRatioArrays(t *testing.T) {
	testCases := []struct {
		rows, cols int
	}{
		{1, 64},  // 1 row, 64 cols
		{64, 1},  // 64 rows, 1 col
		{2, 128}, // Wide array
		{128, 2}, // Tall array
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%dx%d", tc.rows, tc.cols), func(t *testing.T) {
			cfg := &Config{
				Rows:       tc.rows,
				Cols:       tc.cols,
				NoiseLevel: 0.0,
				ADCBits:    8,
				DACBits:    8,
			}

			arr, err := NewArray(cfg)
			if err != nil {
				t.Fatalf("Failed to create array: %v", err)
			}
			defer arr.Destroy()

			// Program weights
			for i := 0; i < tc.rows; i++ {
				for j := 0; j < tc.cols; j++ {
					arr.ProgramWeight(i, j, 0.5)
				}
			}

			// MVM with correct input size
			input := make([]float64, tc.cols)
			for j := range input {
				input[j] = 0.5
			}

			output, err := arr.MVM(input)
			if err != nil {
				t.Fatalf("MVM failed: %v", err)
			}

			if len(output) != tc.rows {
				t.Errorf("Output length = %d, want %d", len(output), tc.rows)
			}

			// Check no NaN/Inf
			for i, v := range output {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					t.Errorf("Output[%d] is NaN or Inf", i)
				}
			}
		})
	}
}

// TestZeroConductanceCell verifies handling of minimum conductance
func TestZeroConductanceCell(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)
	defer arr.Destroy()

	// Program one cell to minimum (level 0)
	arr.ProgramWeight(0, 0, 0.0)

	// Other cells at mid-range
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i != 0 || j != 0 {
				arr.ProgramWeight(i, j, 0.5)
			}
		}
	}

	t.Run("MVM handles minimum G", func(t *testing.T) {
		output, err := arr.MVM([]float64{1.0, 1.0, 1.0, 1.0})
		if err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
		// Should have no NaN/Inf
		for i, v := range output {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Errorf("Output[%d] is NaN or Inf", i)
			}
		}
	})
}

// TestMaximumConductanceSaturation verifies values > 1.0 saturate
func TestMaximumConductanceSaturation(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)
	defer arr.Destroy()

	// Try to program above 1.0 (should saturate to level 29)
	arr.ProgramWeight(0, 0, 1.5)

	matrix := arr.GetConductanceMatrix()
	if matrix[0][0] > 1.0 {
		t.Errorf("Weight %f not saturated to <= 1.0", matrix[0][0])
	}

	// Also test negative values should clamp to 0
	arr.ProgramWeight(1, 1, -0.5)
	matrix = arr.GetConductanceMatrix()
	if matrix[1][1] < 0.0 {
		t.Errorf("Weight %f not clamped to >= 0.0", matrix[1][1])
	}
}

// TestQuantizationBoundaryValues verifies edge quantization behavior
func TestQuantizationBoundaryValues(t *testing.T) {
	t.Run("Exactly 30 unique levels", func(t *testing.T) {
		seen := make(map[float64]bool)
		for i := 0; i <= 1000; i++ {
			input := float64(i) / 1000.0
			quantized := physics.QuantizeTo30Levels(input)
			seen[quantized] = true
		}
		if len(seen) != physics.DefaultLevels {
			t.Errorf("Found %d unique levels, want %d", len(seen), physics.DefaultLevels)
		}
	})

	t.Run("Boundary at 0", func(t *testing.T) {
		q := physics.QuantizeTo30Levels(0.0)
		if q != 0.0 {
			t.Errorf("QuantizeTo30Levels(0.0) = %f, want 0.0", q)
		}
	})

	t.Run("Boundary at 1", func(t *testing.T) {
		q := physics.QuantizeTo30Levels(1.0)
		if q != 1.0 {
			t.Errorf("QuantizeTo30Levels(1.0) = %f, want 1.0", q)
		}
	})

	t.Run("Midpoint quantization", func(t *testing.T) {
		// 0.5 should quantize to level 14 or 15 (14.5 rounded)
		q := physics.QuantizeTo30Levels(0.5)
		level := GetLevel(q)
		if level < 14 || level > 15 {
			t.Errorf("QuantizeTo30Levels(0.5) -> level %d, expected 14 or 15", level)
		}
	})
}

// TestArrayDestroyCleanup verifies Destroy() doesn't panic
func TestArrayDestroyCleanup(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)

	// Should not panic
	arr.Destroy()

	// Double destroy should also not panic
	arr.Destroy()
}

// TestLargeArrayPerformance verifies large arrays work (smoke test)
func TestLargeArrayPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large array test in short mode")
	}

	cfg := &Config{
		Rows:       256,
		Cols:       256,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create 256x256 array: %v", err)
	}
	defer arr.Destroy()

	// Quick program and MVM
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	input := make([]float64, 256)
	for j := range input {
		input[j] = 0.5
	}

	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	if len(output) != 256 {
		t.Errorf("Output length = %d, want 256", len(output))
	}
}
