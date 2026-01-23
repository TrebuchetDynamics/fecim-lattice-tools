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

	// Check all cells have valid conductance
	for _, cell := range mapping.Cells {
		if cell.Conductance < config.GMin || cell.Conductance > config.GMax {
			t.Errorf("Cell conductance %f outside range [%f, %f]",
				cell.Conductance, config.GMin, config.GMax)
		}
	}

	t.Logf("Stats: %+v", mapping.Stats)
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

func TestCompile_EmptyMatrix(t *testing.T) {
	weights := [][]float64{}

	config := DefaultConfig()
	_, err := Compile(weights, config)
	if err == nil {
		t.Error("Expected error for empty matrix")
	}
}

func TestCompile_8x8Sample(t *testing.T) {
	// Sample from plan file
	weights := [][]float64{
		{0.1, -0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8},
		{-0.1, 0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8},
		{0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75, -0.85},
		{-0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75, 0.85},
		{0.05, -0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75},
		{-0.05, 0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75},
		{0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8, -0.9},
		{-0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8, 0.9},
	}

	config := DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8

	mapping, err := Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Check expected stats
	if mapping.Stats.UsedCells != 64 {
		t.Errorf("Expected 64 used cells, got %d", mapping.Stats.UsedCells)
	}

	if mapping.Stats.Utilization != 1.0 {
		t.Errorf("Expected 100%% utilization, got %f", mapping.Stats.Utilization)
	}

	t.Logf("8x8 sample compiled successfully:")
	t.Logf("  UsedCells: %d", mapping.Stats.UsedCells)
	t.Logf("  UniqueLevels: %d", mapping.Stats.UniqueLevels)
	t.Logf("  WeightRange: [%.2f, %.2f]", mapping.Stats.WeightMin, mapping.Stats.WeightMax)
	t.Logf("  QuantPSNR: %.2f dB", mapping.Stats.QuantPSNR)
}

// ============================================================================
// Tests for new GenerateDesign API with three operation modes
// ============================================================================

func TestGenerateDesign_StorageMode(t *testing.T) {
	config := NewStorageConfig(64, 64)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Check mode is correct
	if design.Config.Mode != ModeStorage {
		t.Errorf("Expected ModeStorage, got %v", design.Config.Mode)
	}

	// Check total cells
	expectedCells := 64 * 64
	if design.Stats.TotalCells != expectedCells {
		t.Errorf("Expected %d total cells, got %d", expectedCells, design.Stats.TotalCells)
	}

	// Check all cells are at middle level (unprogrammed state)
	midLevel := config.Levels / 2
	for _, cell := range design.Cells {
		if cell.Level != midLevel {
			t.Errorf("Storage cell at [%d,%d] should be at mid level %d, got %d",
				cell.Row, cell.Col, midLevel, cell.Level)
			break
		}
	}

	t.Logf("Storage mode design: %d cells, %.4f mm², %.2f mW",
		design.Stats.TotalCells, design.Stats.AreaMM2, design.Stats.PowerMW)
}

func TestGenerateDesign_MemoryMode(t *testing.T) {
	config := NewMemoryConfig(32, 32)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Check mode is correct
	if design.Config.Mode != ModeMemory {
		t.Errorf("Expected ModeMemory, got %v", design.Config.Mode)
	}

	// Memory mode starts at reset state (level 0)
	for _, cell := range design.Cells {
		if cell.Level != 0 {
			t.Errorf("Memory cell at [%d,%d] should be at reset level 0, got %d",
				cell.Row, cell.Col, cell.Level)
			break
		}
	}

	t.Logf("Memory mode design: %d cells, %.4f mm², %.2f mW",
		design.Stats.TotalCells, design.Stats.AreaMM2, design.Stats.PowerMW)
}

func TestGenerateDesign_ComputeMode_NoWeights(t *testing.T) {
	config := NewComputeConfig(16, 16)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Check mode is correct
	if design.Config.Mode != ModeCompute {
		t.Errorf("Expected ModeCompute, got %v", design.Config.Mode)
	}

	// Without weights, cells should be at middle level (zero weight equivalent)
	midLevel := config.Levels / 2
	for _, cell := range design.Cells {
		if cell.Level != midLevel {
			t.Errorf("Unprogrammed compute cell should be at mid level %d, got %d",
				midLevel, cell.Level)
			break
		}
	}

	// Throughput should be estimated
	if design.Stats.ThroughputGOPS <= 0 {
		t.Error("Compute mode should have estimated throughput")
	}

	t.Logf("Compute mode (no weights): %d cells, %.2f GOPS",
		design.Stats.TotalCells, design.Stats.ThroughputGOPS)
}

func TestGenerateDesign_ComputeMode_WithWeights(t *testing.T) {
	config := NewComputeConfig(8, 8)

	// Provide initial weights
	weights := [][]float64{
		{0.1, -0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8},
		{-0.1, 0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8},
		{0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75, -0.85},
		{-0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75, 0.85},
		{0.05, -0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75},
		{-0.05, 0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75},
		{0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8, -0.9},
		{-0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8, 0.9},
	}
	config.ComputeConfig.InitialWeights = weights

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Check weights were compiled
	if design.Stats.QuantPSNR == 0 {
		t.Error("Compute mode with weights should have PSNR > 0")
	}

	// Check cells have InitialWeight populated
	hasWeights := false
	for _, cell := range design.Cells {
		if cell.InitialWeight != 0 {
			hasWeights = true
			break
		}
	}
	if !hasWeights {
		t.Error("Cells should have InitialWeight populated when compiled with weights")
	}

	t.Logf("Compute mode (with weights): PSNR=%.2f dB, weight range=[%.2f, %.2f]",
		design.Stats.QuantPSNR, design.Stats.WeightMin, design.Stats.WeightMax)
}

func TestGenerateDesign_With1T1R(t *testing.T) {
	config := NewStorageConfig(16, 16).With1T1R()

	if config.Architecture != Arch1T1R {
		t.Errorf("Expected 1T1R architecture, got %s", config.Architecture)
	}

	// 1T1R should have larger cell pitch
	if config.CellPitch <= 0.46 {
		t.Error("1T1R should have larger cell pitch than passive")
	}

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	t.Logf("1T1R design: %d cells, cell pitch=%.2f μm",
		design.Stats.TotalCells, config.CellPitch)
}

func TestGenerateDesign_NilConfig(t *testing.T) {
	_, err := GenerateDesign(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestGenerateDesign_InvalidDimensions(t *testing.T) {
	config := NewStorageConfig(0, 10)
	_, err := GenerateDesign(config)
	if err == nil {
		t.Error("Expected error for zero rows")
	}

	config = NewStorageConfig(10, 0)
	_, err = GenerateDesign(config)
	if err == nil {
		t.Error("Expected error for zero cols")
	}
}
