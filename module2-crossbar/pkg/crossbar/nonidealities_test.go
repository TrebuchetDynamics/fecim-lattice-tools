package crossbar

import (
	"math/rand"
	"testing"
)

func TestDefaultWireParams(t *testing.T) {
	params := DefaultWireParams()

	if params.RwordLine <= 0 {
		t.Error("Word line resistance should be positive")
	}
	if params.RbitLine <= 0 {
		t.Error("Bit line resistance should be positive")
	}
	if params.Rcontact <= 0 {
		t.Error("Contact resistance should be positive")
	}
}

func TestAnalyzeIRDrop(t *testing.T) {
	cfg := &Config{
		Rows:       16,
		Cols:       16,
		NoiseLevel: 0.01,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	// Program random weights
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, rand.Float64())
		}
	}

	// Create input
	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}

	analysis := arr.AnalyzeIRDrop(input, nil)

	// Check dimensions
	if len(analysis.WordLineVoltages) != cfg.Rows {
		t.Errorf("Expected %d rows in WL voltages, got %d", cfg.Rows, len(analysis.WordLineVoltages))
	}

	if len(analysis.BitLineVoltages) != cfg.Rows {
		t.Errorf("Expected %d rows in BL voltages, got %d", cfg.Rows, len(analysis.BitLineVoltages))
	}

	// Check that IR drop is non-negative
	if analysis.MaxIRDrop < 0 {
		t.Error("Max IR drop should be non-negative")
	}

	if analysis.AvgIRDrop < 0 {
		t.Error("Avg IR drop should be non-negative")
	}

	// Check worst case cell is within bounds
	if analysis.WorstCaseCell[0] < 0 || analysis.WorstCaseCell[0] >= cfg.Rows {
		t.Error("Worst case cell row out of bounds")
	}
	if analysis.WorstCaseCell[1] < 0 || analysis.WorstCaseCell[1] >= cfg.Cols {
		t.Error("Worst case cell col out of bounds")
	}
}

func TestAnalyzeSneakPaths(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.01,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	// Program weights to have some conductance
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, 0.5+rand.Float64()*0.5)
		}
	}

	selectedRow := 4
	selectedCol := 4

	analysis := arr.AnalyzeSneakPaths(selectedRow, selectedCol)

	// Check dimensions
	if len(analysis.SneakCurrents) != cfg.Rows {
		t.Errorf("Expected %d rows in sneak map, got %d", cfg.Rows, len(analysis.SneakCurrents))
	}

	// Check that selected cell has no sneak current to itself
	if analysis.SneakCurrents[selectedRow][selectedCol] != 0 {
		t.Error("Selected cell should have zero sneak current")
	}

	// Check that same row/column cells have sneak currents
	hasRowSneak := false
	hasColSneak := false

	for j := 0; j < cfg.Cols; j++ {
		if j != selectedCol && analysis.SneakCurrents[selectedRow][j] > 0 {
			hasRowSneak = true
			break
		}
	}

	for i := 0; i < cfg.Rows; i++ {
		if i != selectedRow && analysis.SneakCurrents[i][selectedCol] > 0 {
			hasColSneak = true
			break
		}
	}

	if !hasRowSneak {
		t.Error("Expected sneak currents in same row")
	}
	if !hasColSneak {
		t.Error("Expected sneak currents in same column")
	}
}

func TestMVMWithIRDrop(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0, // No noise for deterministic test
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	// Program uniform weights
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	// Create input
	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = 0.5
	}

	// Get ideal output
	idealOutput, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// Get output with IR drop
	actualOutput, irAnalysis, err := arr.MVMWithIRDrop(input, nil)
	if err != nil {
		t.Fatalf("MVMWithIRDrop failed: %v", err)
	}

	// Outputs should have same length
	if len(idealOutput) != len(actualOutput) {
		t.Error("Output lengths should match")
	}

	// IR analysis should be populated
	if irAnalysis == nil {
		t.Error("IR analysis should not be nil")
	}

	// For small arrays, outputs should be similar
	for i := range idealOutput {
		diff := idealOutput[i] - actualOutput[i]
		if diff < -0.1 || diff > 0.1 {
			t.Errorf("Output difference too large at index %d: %f", i, diff)
		}
	}
}

func TestIRDropMapNormalization(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	// Program weights
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = 1.0
	}

	analysis := arr.AnalyzeIRDrop(input, nil)
	normalized := analysis.GetIRDropMap()

	// Check all values are in [0, 1]
	for i := range normalized {
		for j := range normalized[i] {
			if normalized[i][j] < 0 || normalized[i][j] > 1.01 {
				t.Errorf("Normalized IR drop out of range at [%d,%d]: %f", i, j, normalized[i][j])
			}
		}
	}
}

func TestSneakMapNormalization(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	// Program weights
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	analysis := arr.AnalyzeSneakPaths(2, 2)
	normalized := analysis.GetSneakMap()

	// Check all values are in [0, 1]
	for i := range normalized {
		for j := range normalized[i] {
			if normalized[i][j] < 0 || normalized[i][j] > 1.01 {
				t.Errorf("Normalized sneak map out of range at [%d,%d]: %f", i, j, normalized[i][j])
			}
		}
	}
}

func TestComputeError(t *testing.T) {
	ideal := []float64{1.0, 0.5, 0.25}
	actual := []float64{0.9, 0.5, 0.3}

	err := ComputeError(ideal, actual)

	if err < 0 {
		t.Error("Error should be non-negative")
	}

	// Same vectors should have zero error
	zeroErr := ComputeError(ideal, ideal)
	if zeroErr != 0 {
		t.Errorf("Same vectors should have zero error, got %f", zeroErr)
	}
}

func BenchmarkIRDropAnalysis(b *testing.B) {
	cfg := &Config{
		Rows:       64,
		Cols:       64,
		NoiseLevel: 0.01,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, rand.Float64())
		}
	}

	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}

	params := DefaultWireParams()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.AnalyzeIRDrop(input, params)
	}
}

func BenchmarkSneakPathAnalysis(b *testing.B) {
	cfg := &Config{
		Rows:       64,
		Cols:       64,
		NoiseLevel: 0.01,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, rand.Float64())
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.AnalyzeSneakPaths(32, 32)
	}
}
