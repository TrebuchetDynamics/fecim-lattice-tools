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

// ============================================================================
// Additional Tests for Complete Coverage
// ============================================================================

func TestOperationModeString(t *testing.T) {
	tests := []struct {
		mode OperationMode
		want string
	}{
		{ModeStorage, "Storage"},
		{ModeMemory, "Memory"},
		{ModeCompute, "Compute"},
		{OperationMode(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("OperationMode(%d).String() = %v, want %v", tt.mode, got, tt.want)
		}
	}
}

func TestCellPitchScaling(t *testing.T) {
	// Default passive config
	passiveConfig := NewStorageConfig(16, 16)
	if passiveConfig.CellPitch != 0.46 {
		t.Errorf("Passive cell pitch = %f, expected 0.46", passiveConfig.CellPitch)
	}
	if passiveConfig.Architecture != ArchPassive {
		t.Errorf("Default architecture should be passive, got %s", passiveConfig.Architecture)
	}

	// Switch to 1T1R
	config1T1R := NewStorageConfig(16, 16).With1T1R()
	if config1T1R.CellPitch != 0.92 {
		t.Errorf("1T1R cell pitch = %f, expected 0.92", config1T1R.CellPitch)
	}
	if config1T1R.Architecture != Arch1T1R {
		t.Errorf("Architecture should be 1T1R after With1T1R(), got %s", config1T1R.Architecture)
	}
}

func TestStorageModeRetentionConfig(t *testing.T) {
	config := NewStorageConfig(256, 256)

	if config.StorageConfig == nil {
		t.Fatal("StorageConfig should not be nil for storage mode")
	}

	if config.StorageConfig.RetentionYears != 10 {
		t.Errorf("Expected 10-year retention, got %f", config.StorageConfig.RetentionYears)
	}

	if config.StorageConfig.EnduranceCycles != 1000000 {
		t.Errorf("Expected 1M endurance cycles, got %d", config.StorageConfig.EnduranceCycles)
	}

	// Verify no compute config for storage mode
	if config.ComputeConfig != nil {
		t.Error("Storage mode should not have ComputeConfig")
	}
}

func TestMemoryModeAccessTimeConfig(t *testing.T) {
	config := NewMemoryConfig(128, 128)

	if config.MemoryConfig == nil {
		t.Fatal("MemoryConfig should not be nil for memory mode")
	}

	if config.MemoryConfig.AccessTimeNs != 10.0 {
		t.Errorf("Expected 10ns access time, got %f", config.MemoryConfig.AccessTimeNs)
	}

	if config.MemoryConfig.BandwidthGBps != 10.0 {
		t.Errorf("Expected 10 GB/s bandwidth, got %f", config.MemoryConfig.BandwidthGBps)
	}
}

func TestArrayDesignStatistics(t *testing.T) {
	weights := make([][]float64, 16)
	for i := range weights {
		weights[i] = make([]float64, 16)
		for j := range weights[i] {
			weights[i][j] = float64(i*16+j) / 256.0
		}
	}

	config := NewComputeConfig(16, 16)
	config.ComputeConfig.InitialWeights = weights

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Verify cell count
	if design.Stats.TotalCells != 256 {
		t.Errorf("Expected 256 total cells, got %d", design.Stats.TotalCells)
	}

	if design.Stats.ActiveCells != 256 {
		t.Errorf("Expected 256 active cells, got %d", design.Stats.ActiveCells)
	}

	// Verify area is calculated
	if design.Stats.AreaMM2 <= 0 {
		t.Error("Area should be calculated and > 0")
	}

	// Verify power is estimated
	if design.Stats.PowerMW <= 0 {
		t.Error("Power should be estimated and > 0")
	}

	// For compute mode with weights, quantization stats should exist
	if design.Stats.QuantMSE == 0 && design.Stats.QuantPSNR == 0 {
		t.Error("QuantMSE/QuantPSNR should be calculated for compute mode with weights")
	}
}

func TestWithWeightsNonComputeMode(t *testing.T) {
	// WithWeights should be ignored for non-compute modes
	weights := [][]float64{{0.1, 0.2}, {0.3, 0.4}}

	storageConfig := NewStorageConfig(8, 8).WithWeights(weights)

	// Should not panic and config should remain valid
	design, err := GenerateDesign(storageConfig)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Storage mode should not have weights in cells
	for _, cell := range design.Cells {
		if cell.InitialWeight != 0 {
			t.Error("Storage mode cells should not have InitialWeight")
			break
		}
	}
}

func TestLargeArrayGeneration(t *testing.T) {
	// Test large array (256x256) to ensure scalability
	config := NewStorageConfig(256, 256)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("Large array generation failed: %v", err)
	}

	expectedCells := 256 * 256
	if design.Stats.TotalCells != expectedCells {
		t.Errorf("Expected %d cells, got %d", expectedCells, design.Stats.TotalCells)
	}

	if len(design.Cells) != expectedCells {
		t.Errorf("Expected %d cell assignments, got %d", expectedCells, len(design.Cells))
	}

	t.Logf("256x256 array: %d cells, %.4f mm², %.2f mW",
		design.Stats.TotalCells, design.Stats.AreaMM2, design.Stats.PowerMW)
}

func TestTechnologySelection(t *testing.T) {
	config := NewStorageConfig(16, 16)

	// Default should be SKY130
	if config.Technology != TechSKY130 {
		t.Errorf("Default technology should be SKY130, got %s", config.Technology)
	}

	// Test technology constants
	techs := []string{TechSKY130, TechGF180, TechIHP}
	for _, tech := range techs {
		if tech == "" {
			t.Error("Technology constant should not be empty")
		}
	}
}

func TestPeripheralConfig(t *testing.T) {
	config := NewComputeConfig(16, 16)

	// Check default peripheral configuration
	if config.Peripherals.DACBits != 8 {
		t.Errorf("Expected 8-bit DAC, got %d", config.Peripherals.DACBits)
	}

	if config.Peripherals.ADCBits != 8 {
		t.Errorf("Expected 8-bit ADC, got %d", config.Peripherals.ADCBits)
	}

	if config.Peripherals.VDD != 1.8 {
		t.Errorf("Expected 1.8V VDD, got %f", config.Peripherals.VDD)
	}

	if config.Peripherals.ClockFreq != 100.0 {
		t.Errorf("Expected 100 MHz clock, got %f", config.Peripherals.ClockFreq)
	}
}

func TestCellConductanceRange(t *testing.T) {
	config := NewComputeConfig(4, 4)
	weights := [][]float64{
		{-1.0, 0.0, 0.5, 1.0},
		{0.25, -0.25, 0.75, -0.75},
		{0.1, -0.1, 0.9, -0.9},
		{0.0, 0.0, 0.0, 0.0},
	}
	config.ComputeConfig.InitialWeights = weights

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// All cells should have valid conductance in range
	for _, cell := range design.Cells {
		if cell.Conductance < config.GMin || cell.Conductance > config.GMax {
			t.Errorf("Cell [%d,%d] conductance %.2f outside range [%.2f, %.2f]",
				cell.Row, cell.Col, cell.Conductance, config.GMin, config.GMax)
		}
		if cell.Level < 0 || cell.Level >= config.Levels {
			t.Errorf("Cell [%d,%d] level %d outside range [0, %d)",
				cell.Row, cell.Col, cell.Level, config.Levels)
		}
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGenerateDesign_Small(b *testing.B) {
	config := NewStorageConfig(16, 16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateDesign(config)
	}
}

func BenchmarkGenerateDesign_Medium(b *testing.B) {
	config := NewStorageConfig(64, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateDesign(config)
	}
}

func BenchmarkGenerateDesign_Large(b *testing.B) {
	config := NewStorageConfig(256, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateDesign(config)
	}
}

func BenchmarkCompileWeights(b *testing.B) {
	weights := make([][]float64, 64)
	for i := range weights {
		weights[i] = make([]float64, 64)
		for j := range weights[i] {
			weights[i][j] = float64(i*64+j) / 4096.0
		}
	}
	config := DefaultConfig()
	config.ArrayRows = 64
	config.ArrayCols = 64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compile(weights, config)
	}
}
