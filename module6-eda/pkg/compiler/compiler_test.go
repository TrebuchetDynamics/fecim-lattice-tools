// pkg/compiler/compiler_test.go
package compiler

import (
	"testing"
)

func TestCompile_Basic(t *testing.T) {
	weights := [][]float64{
		{0.1, -0.2, 0.3},
		{-0.4, 0.5, -0.6},
	}

	config := DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8

	mapping, err := Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Check basic stats
	if mapping.Stats.UsedCells != 6 {
		t.Errorf("Expected 6 used cells, got %d", mapping.Stats.UsedCells)
	}

	if mapping.Stats.Utilization < 0 || mapping.Stats.Utilization > 1 {
		t.Errorf("Invalid utilization: %f", mapping.Stats.Utilization)
	}

	t.Logf("Stats: %+v", mapping.Stats)
}

func TestCompile_BlankArray(t *testing.T) {
	config := DefaultConfig()
	config.ArrayRows = 16
	config.ArrayCols = 16
	config.Levels = 30 // Ensure 30 levels

	// Compile with nil weights -> Should produce blank array
	mapping, err := Compile(nil, config)
	if err != nil {
		t.Fatalf("Blank generation failed: %v", err)
	}

	// Verify dimensions
	expectedCells := 16 * 16
	if len(mapping.Cells) != expectedCells {
		t.Errorf("Expected %d cells, got %d", expectedCells, len(mapping.Cells))
	}

	// Verify initialization
	for i, cell := range mapping.Cells {
		if cell.QuantLevel != 0 {
			t.Errorf("Cell %d initialized to level %d, expected 0", i, cell.QuantLevel)
		}
		if cell.Conductance != config.GMin {
			t.Errorf("Cell %d initialized to conductance %f, expected %f", i, cell.Conductance, config.GMin)
		}
	}

	t.Logf("Blank Array Generated: %d cells", len(mapping.Cells))
}

func TestCompile_Quantization(t *testing.T) {
	// Test that quantization preserves information
	weights := [][]float64{{0.0, 0.5, 1.0, -0.5, -1.0}}

	config := DefaultConfig()
	config.Levels = 30
	config.ArrayRows = 8
	config.ArrayCols = 8

	mapping, err := Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// PSNR should be reasonable for 30 levels
	if mapping.Stats.QuantPSNR < 20 {
		t.Errorf("PSNR too low: %f dB", mapping.Stats.QuantPSNR)
	}

	t.Logf("PSNR: %.2f dB, MSE: %.6f", mapping.Stats.QuantPSNR, mapping.Stats.QuantMSE)
}

func TestCompile_SizeValidation(t *testing.T) {
	weights := [][]float64{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, // 10 cols
	}

	config := DefaultConfig()
	config.ArrayCols = 5 // Too small

	_, err := Compile(weights, config)
	if err == nil {
		t.Error("Expected error for oversized weights")
	}
}
