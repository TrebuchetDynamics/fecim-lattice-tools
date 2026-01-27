// pkg/compiler/compiler_extended_test.go
// Extended tests for GenerateDesign, operation modes, and config builders
package compiler

import (
	"testing"
)

// ============================================================================
// GenerateDesign Tests - All Operation Modes
// ============================================================================

func TestGenerateDesign_StorageMode(t *testing.T) {
	config := NewArrayConfig(ModeStorage, 16, 16)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed for storage mode: %v", err)
	}

	// Verify mode
	if design.Config.Mode != ModeStorage {
		t.Errorf("Expected ModeStorage, got %v", design.Config.Mode)
	}

	// Verify dimensions
	expectedCells := 16 * 16
	if len(design.Cells) != expectedCells {
		t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
	}

	// Storage mode should have blank initialization
	if design.Stats.ActiveCells != 0 {
		t.Errorf("Storage mode blank array should have 0 active cells, got %d", design.Stats.ActiveCells)
	}

	// All cells should be at level 0
	for _, cell := range design.Cells {
		if cell.Level != 0 {
			t.Errorf("Cell (%d,%d) should be level 0, got %d", cell.Row, cell.Col, cell.Level)
		}
	}

	// Verify storage config is set
	if design.Config.StorageConfig == nil {
		t.Error("StorageConfig should be set for storage mode")
	}
	if design.Config.StorageConfig.RetentionYears != 10 {
		t.Errorf("Expected 10 year retention, got %f", design.Config.StorageConfig.RetentionYears)
	}
}

func TestGenerateDesign_MemoryMode(t *testing.T) {
	config := NewArrayConfig(ModeMemory, 32, 32)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed for memory mode: %v", err)
	}

	// Verify mode
	if design.Config.Mode != ModeMemory {
		t.Errorf("Expected ModeMemory, got %v", design.Config.Mode)
	}

	// Verify memory config
	if design.Config.MemoryConfig == nil {
		t.Error("MemoryConfig should be set for memory mode")
	}
	if design.Config.MemoryConfig.AccessTimeNs != 10.0 {
		t.Errorf("Expected 10ns access time, got %f", design.Config.MemoryConfig.AccessTimeNs)
	}
}

func TestGenerateDesign_ComputeModeWithoutWeights(t *testing.T) {
	config := NewArrayConfig(ModeCompute, 64, 64)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed for compute mode without weights: %v", err)
	}

	// Verify mode
	if design.Config.Mode != ModeCompute {
		t.Errorf("Expected ModeCompute, got %v", design.Config.Mode)
	}

	// Without weights, should generate blank array
	if design.Stats.ActiveCells != 0 {
		t.Errorf("Compute mode without weights should have 0 active cells, got %d", design.Stats.ActiveCells)
	}

	// Verify compute config
	if design.Config.ComputeConfig == nil {
		t.Error("ComputeConfig should be set for compute mode")
	}
}

func TestGenerateDesign_ComputeModeWithWeights(t *testing.T) {
	weights := [][]float64{
		{0.1, -0.2, 0.3, 0.4},
		{-0.5, 0.6, -0.7, 0.8},
		{0.9, -1.0, 0.1, -0.2},
	}

	config := NewArrayConfig(ModeCompute, 8, 8)
	config.WithWeights(weights)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed for compute mode with weights: %v", err)
	}

	// Verify active cells matches weight matrix size
	expectedActive := 3 * 4 // 3 rows x 4 cols
	if design.Stats.ActiveCells != expectedActive {
		t.Errorf("Expected %d active cells, got %d", expectedActive, design.Stats.ActiveCells)
	}

	// Verify PSNR is reasonable for 30 levels
	if design.Stats.QuantPSNR < 20 {
		t.Errorf("PSNR too low: %f dB (expected >= 20)", design.Stats.QuantPSNR)
	}

	t.Logf("PSNR: %.2f dB, MSE: %.6f", design.Stats.QuantPSNR, design.Stats.QuantMSE)
}

// ============================================================================
// GenerateBlank Tests
// ============================================================================

func TestGenerateBlank_AllModes(t *testing.T) {
	modes := []struct {
		mode OperationMode
		name string
	}{
		{ModeStorage, "Storage"},
		{ModeMemory, "Memory"},
		{ModeCompute, "Compute"},
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			config := NewArrayConfig(m.mode, 8, 8)
			design := GenerateBlank(config)

			if design == nil {
				t.Fatal("GenerateBlank returned nil")
			}

			// All cells should be at level 0
			for _, cell := range design.Cells {
				if cell.Level != 0 {
					t.Errorf("Cell (%d,%d) should be level 0, got %d", cell.Row, cell.Col, cell.Level)
				}
				// Conductance should be GMin
				if cell.Conductance != config.GMin {
					t.Errorf("Cell conductance should be %f, got %f", config.GMin, cell.Conductance)
				}
			}
		})
	}
}

func TestGenerateBlank_DimensionScaling(t *testing.T) {
	sizes := []struct{ rows, cols int }{
		{4, 4},
		{16, 16},
		{32, 64},
		{128, 128},
	}

	for _, size := range sizes {
		config := NewArrayConfig(ModeStorage, size.rows, size.cols)
		design := GenerateBlank(config)

		expectedCells := size.rows * size.cols
		if len(design.Cells) != expectedCells {
			t.Errorf("Size %dx%d: expected %d cells, got %d",
				size.rows, size.cols, expectedCells, len(design.Cells))
		}

		if design.Stats.TotalCells != expectedCells {
			t.Errorf("Size %dx%d: TotalCells stat should be %d, got %d",
				size.rows, size.cols, expectedCells, design.Stats.TotalCells)
		}
	}
}

// ============================================================================
// Config Builder Tests
// ============================================================================

func TestNewArrayConfig_Defaults(t *testing.T) {
	config := NewArrayConfig(ModeCompute, 64, 64)

	// Verify defaults
	if config.Technology != TechSKY130 {
		t.Errorf("Default technology should be SKY130, got %s", config.Technology)
	}
	if config.Architecture != ArchPassive {
		t.Errorf("Default architecture should be passive, got %s", config.Architecture)
	}
	if config.Levels != 30 {
		t.Errorf("Default levels should be 30, got %d", config.Levels)
	}
	if config.CellPitch != 0.46 {
		t.Errorf("Default cell pitch should be 0.46, got %f", config.CellPitch)
	}
}

func TestNewStorageConfig(t *testing.T) {
	config := NewStorageConfig(64, 64)

	if config.Mode != ModeStorage {
		t.Errorf("Expected ModeStorage, got %v", config.Mode)
	}
	if config.StorageConfig == nil {
		t.Error("StorageConfig should not be nil")
	}
}

func TestNewMemoryConfig(t *testing.T) {
	config := NewMemoryConfig(64, 64)

	if config.Mode != ModeMemory {
		t.Errorf("Expected ModeMemory, got %v", config.Mode)
	}
	if config.MemoryConfig == nil {
		t.Error("MemoryConfig should not be nil")
	}
}

func TestNewComputeConfig(t *testing.T) {
	config := NewComputeConfig(64, 64)

	if config.Mode != ModeCompute {
		t.Errorf("Expected ModeCompute, got %v", config.Mode)
	}
	if config.ComputeConfig == nil {
		t.Error("ComputeConfig should not be nil")
	}
}

func TestWith1T1R(t *testing.T) {
	config := NewComputeConfig(64, 64).With1T1R()

	if config.Architecture != Arch1T1R {
		t.Errorf("Architecture should be 1T1R, got %s", config.Architecture)
	}
	if config.CellPitch != 0.92 {
		t.Errorf("1T1R cell pitch should be 0.92, got %f", config.CellPitch)
	}
}

func TestWithWeights(t *testing.T) {
	weights := [][]float64{{1.0, 2.0}, {3.0, 4.0}}

	config := NewComputeConfig(64, 64).WithWeights(weights)

	if config.ComputeConfig.InitialWeights == nil {
		t.Error("InitialWeights should be set")
	}
	if len(config.ComputeConfig.InitialWeights) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(config.ComputeConfig.InitialWeights))
	}
}

func TestWithWeights_NonComputeMode(t *testing.T) {
	weights := [][]float64{{1.0, 2.0}}

	// WithWeights should be ignored for non-compute modes
	config := NewStorageConfig(64, 64).WithWeights(weights)

	if config.ComputeConfig != nil && config.ComputeConfig.InitialWeights != nil {
		t.Error("WithWeights should be ignored for storage mode")
	}
}

// ============================================================================
// OperationMode.String() Tests
// ============================================================================

func TestOperationMode_String(t *testing.T) {
	tests := []struct {
		mode     OperationMode
		expected string
	}{
		{ModeStorage, "Storage"},
		{ModeMemory, "Memory"},
		{ModeCompute, "Compute"},
		{OperationMode(99), "Unknown"},
	}

	for _, tc := range tests {
		result := tc.mode.String()
		if result != tc.expected {
			t.Errorf("Mode %d: expected %q, got %q", tc.mode, tc.expected, result)
		}
	}
}

// ============================================================================
// Edge Cases and Error Handling
// ============================================================================

func TestGenerateDesign_EmptyWeights(t *testing.T) {
	config := NewComputeConfig(8, 8)
	config.ComputeConfig.InitialWeights = [][]float64{}

	_, err := GenerateDesign(config)
	if err == nil {
		t.Error("Expected error for empty weight matrix")
	}
}

func TestGenerateDesign_OversizedWeights(t *testing.T) {
	// Weight matrix larger than array
	weights := make([][]float64, 100)
	for i := range weights {
		weights[i] = make([]float64, 100)
	}

	config := NewComputeConfig(8, 8)
	config.WithWeights(weights)

	_, err := GenerateDesign(config)
	if err == nil {
		t.Error("Expected error for oversized weights")
	}
}

func TestGenerateDesign_SingleCellWeights(t *testing.T) {
	// Edge case: 1x1 weight matrix
	weights := [][]float64{{0.5}}

	config := NewComputeConfig(4, 4)
	config.WithWeights(weights)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("Failed for 1x1 weights: %v", err)
	}

	if design.Stats.ActiveCells != 1 {
		t.Errorf("Expected 1 active cell, got %d", design.Stats.ActiveCells)
	}
}

func TestGenerateDesign_AllZeroWeights(t *testing.T) {
	weights := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
	}

	config := NewComputeConfig(8, 8)
	config.WithWeights(weights)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("Failed for all-zero weights: %v", err)
	}

	// Should still map cells, just all at level 15 (middle for zero)
	if design.Stats.ActiveCells != 6 {
		t.Errorf("Expected 6 active cells even for zero weights, got %d", design.Stats.ActiveCells)
	}
}

// ============================================================================
// Quantization Quality Tests
// ============================================================================

func TestQuantization_SymmetricBipolar(t *testing.T) {
	// Test that quantization is symmetric: q(-x) ≈ -q(x)
	weights := [][]float64{
		{1.0, -1.0},
		{0.5, -0.5},
		{0.1, -0.1},
	}

	config := NewComputeConfig(8, 8)
	config.WithWeights(weights)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Find cells by their position (cells are stored row-major for full array)
	// For 8x8 array: cell at (row, col) is at index row*8 + col
	arrayCols := 8

	// Check symmetry of levels for opposite weights in same row
	// Row 0: (0,0)=1.0 and (0,1)=-1.0 should have symmetric levels
	testPairs := []struct {
		row, col1, col2 int
	}{
		{0, 0, 1}, // 1.0 and -1.0
		{1, 0, 1}, // 0.5 and -0.5
		{2, 0, 1}, // 0.1 and -0.1
	}

	for _, pair := range testPairs {
		idx1 := pair.row*arrayCols + pair.col1
		idx2 := pair.row*arrayCols + pair.col2

		level1 := design.Cells[idx1].Level
		level2 := design.Cells[idx2].Level

		// Levels should sum to 29 (max level) for symmetric quantization
		sum := level1 + level2
		if sum != 29 {
			t.Errorf("Row %d: Levels %d + %d = %d (expected 29 for symmetric quantization)",
				pair.row, level1, level2, sum)
		}
	}
}

func TestQuantization_30Levels(t *testing.T) {
	// Verify that quantization uses all 30 levels when appropriate
	weights := make([][]float64, 1)
	weights[0] = make([]float64, 30)
	for i := 0; i < 30; i++ {
		weights[0][i] = float64(i) / 29.0 * 2.0 - 1.0 // Range from -1 to 1
	}

	config := NewComputeConfig(64, 64)
	config.Levels = 30
	config.WithWeights(weights)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Count unique levels
	levelSet := make(map[int]bool)
	for _, cell := range design.Cells {
		if cell.InitialWeight != 0 || cell.Row == 0 { // Only count mapped cells
			levelSet[cell.Level] = true
		}
	}

	// Should have close to 30 unique levels
	if len(levelSet) < 20 {
		t.Errorf("Expected ~30 unique levels, got %d", len(levelSet))
	}
	t.Logf("Unique levels used: %d", len(levelSet))
}

// ============================================================================
// Legacy API Compatibility Tests
// ============================================================================

func TestLegacyCompile_Compatibility(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}

	// Legacy API
	legacyConfig := DefaultConfig()
	legacyConfig.ArrayRows = 8
	legacyConfig.ArrayCols = 8

	legacyMapping, err := Compile(weights, legacyConfig)
	if err != nil {
		t.Fatalf("Legacy Compile failed: %v", err)
	}

	// New API
	newConfig := NewComputeConfig(8, 8)
	newConfig.WithWeights(weights)

	newDesign, err := GenerateDesign(newConfig)
	if err != nil {
		t.Fatalf("New GenerateDesign failed: %v", err)
	}

	// Results should match
	if legacyMapping.Stats.ActiveCells != newDesign.Stats.ActiveCells {
		t.Errorf("ActiveCells mismatch: legacy=%d, new=%d",
			legacyMapping.Stats.ActiveCells, newDesign.Stats.ActiveCells)
	}
}

func TestDefaultConfig_Returns1T1RWhenCalled(t *testing.T) {
	config := Config1T1R()

	if config.Architecture != Arch1T1R {
		t.Errorf("Config1T1R should return 1T1R architecture, got %s", config.Architecture)
	}
}

// ============================================================================
// Area Calculation Tests
// ============================================================================

func TestDesignStats_AreaCalculation(t *testing.T) {
	config := NewComputeConfig(64, 64)
	design := GenerateBlank(config)

	// Expected area = rows * cols * (CellPitch * RowHeight) * 1e-6 mm^2
	expectedArea := float64(64*64) * (0.46 * 2.72) * 1e-6

	// Allow small floating point tolerance
	if diff := design.Stats.AreaMM2 - expectedArea; diff > 1e-10 || diff < -1e-10 {
		t.Errorf("Area calculation: expected %e mm^2, got %e mm^2", expectedArea, design.Stats.AreaMM2)
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGenerateDesign_SmallCompute(b *testing.B) {
	weights := make([][]float64, 16)
	for i := range weights {
		weights[i] = make([]float64, 16)
		for j := range weights[i] {
			weights[i][j] = float64(i*16+j) / 256.0
		}
	}

	config := NewComputeConfig(32, 32)
	config.WithWeights(weights)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateDesign(config)
	}
}

func BenchmarkGenerateBlank_Large(b *testing.B) {
	config := NewStorageConfig(256, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateBlank(config)
	}
}

func BenchmarkGenerateDesign_LargeWeights(b *testing.B) {
	weights := make([][]float64, 128)
	for i := range weights {
		weights[i] = make([]float64, 128)
		for j := range weights[i] {
			weights[i][j] = float64(i-64) * float64(j-64) / 4096.0
		}
	}

	config := NewComputeConfig(128, 128)
	config.WithWeights(weights)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateDesign(config)
	}
}
