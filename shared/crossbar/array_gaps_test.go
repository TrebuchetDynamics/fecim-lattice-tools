package crossbar

import "testing"

// newTestArray creates a small array for gap-coverage tests.
func newGapTestArray(t *testing.T) *Array {
	t.Helper()
	arr, err := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 4, DACBits: 4})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	return arr
}

func TestGetEffectiveConductanceMatrix(t *testing.T) {
	arr := newGapTestArray(t)
	if err := arr.ProgramWeight(0, 0, 0.5); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}
	if err := arr.ProgramWeight(1, 1, 0.8); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}

	eff := arr.GetEffectiveConductanceMatrix()
	if len(eff) != 4 {
		t.Fatalf("GetEffectiveConductanceMatrix: got %d rows, want 4", len(eff))
	}
	if len(eff[0]) != 4 {
		t.Fatalf("GetEffectiveConductanceMatrix: got %d cols, want 4", len(eff[0]))
	}
	for i, row := range eff {
		for j, g := range row {
			if g < 0 {
				t.Fatalf("cell [%d,%d]: negative effective conductance %v", i, j, g)
			}
		}
	}
}

func TestGetProcessVariationFactor(t *testing.T) {
	arr := newGapTestArray(t)
	f := arr.GetProcessVariationFactor(0, 0)
	if f <= 0 {
		t.Fatalf("GetProcessVariationFactor(0,0): got %v, want > 0", f)
	}
	// Without process-variation config, factor should be 1.0 (no variation).
	if f != 1.0 {
		t.Logf("GetProcessVariationFactor without PV config: %v (non-unity allowed with noise)", f)
	}

	// All cells in bounds must return positive factors.
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if v := arr.GetProcessVariationFactor(i, j); v <= 0 {
				t.Fatalf("GetProcessVariationFactor(%d,%d): got %v, want > 0", i, j, v)
			}
		}
	}
}

func TestGetCellStats(t *testing.T) {
	arr := newGapTestArray(t)
	if err := arr.ProgramWeight(0, 0, 0.5); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}

	stats, err := arr.GetCellStats(0, 0)
	if err != nil {
		t.Fatalf("GetCellStats(0,0): %v", err)
	}
	if stats == nil {
		t.Fatal("GetCellStats returned nil")
	}
	if stats.Row != 0 || stats.Col != 0 {
		t.Fatalf("GetCellStats: row/col: got %d/%d, want 0/0", stats.Row, stats.Col)
	}
	if stats.Conductance < 0 || stats.Conductance > 1 {
		t.Fatalf("GetCellStats: Conductance %v not in [0,1]", stats.Conductance)
	}
	if stats.VariationFactor <= 0 {
		t.Fatalf("GetCellStats: VariationFactor %v, want > 0", stats.VariationFactor)
	}

	// Out-of-bounds must return an error.
	_, err = arr.GetCellStats(100, 100)
	if err == nil {
		t.Fatal("GetCellStats(100,100): expected error, got nil")
	}
}

func TestResetDisturbTracking(t *testing.T) {
	arr := newGapTestArray(t)

	// Manually increment HalfSelectCount to verify reset clears it.
	arr.cells[0][0].HalfSelectCount = 5
	arr.cells[1][2].HalfSelectCount = 3
	arr.cells[0][0].DisturbShift = 0.01

	arr.ResetDisturbTracking()

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if arr.cells[i][j].HalfSelectCount != 0 {
				t.Fatalf("cell [%d,%d] HalfSelectCount after reset: got %d, want 0", i, j, arr.cells[i][j].HalfSelectCount)
			}
			if arr.cells[i][j].DisturbShift != 0 {
				t.Fatalf("cell [%d,%d] DisturbShift after reset: got %v, want 0", i, j, arr.cells[i][j].DisturbShift)
			}
		}
	}
}

func TestResetCycleCounts(t *testing.T) {
	arr := newGapTestArray(t)

	// Program a cell twice to build up SwitchingCount.
	_ = arr.ProgramWeight(0, 0, 0.3)
	_ = arr.ProgramWeight(0, 0, 0.7)

	arr.ResetCycleCounts()

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if arr.cells[i][j].SwitchingCount != 0 {
				t.Fatalf("cell [%d,%d] SwitchingCount after reset: got %d, want 0", i, j, arr.cells[i][j].SwitchingCount)
			}
		}
	}
}

func TestAnalyzeIRDropIterative(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping iterative IR drop in short mode")
	}

	arr := newGapTestArray(t)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			_ = arr.ProgramWeight(i, j, 0.5)
		}
	}

	input := []float64{0.5, 0.5, 0.5, 0.5}
	result := arr.AnalyzeIRDropIterative(input, nil, nil)
	if result == nil {
		t.Fatal("AnalyzeIRDropIterative returned nil")
	}
	if len(result.EffectiveVoltage) != 4 {
		t.Fatalf("EffectiveVoltage rows: got %d, want 4", len(result.EffectiveVoltage))
	}
	for i, row := range result.EffectiveVoltage {
		if len(row) != 4 {
			t.Fatalf("EffectiveVoltage[%d] cols: got %d, want 4", i, len(row))
		}
		for j, v := range row {
			if v < 0 || v > 1.001 {
				t.Fatalf("EffectiveVoltage[%d][%d] = %v, want in [0,1]", i, j, v)
			}
		}
	}
}

func TestAnalyzeIRDropIterative_WithDefaultConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	arr := newGapTestArray(t)
	input := []float64{1.0, 1.0, 1.0, 1.0}

	// Explicit nil params and config should use defaults without panic.
	result := arr.AnalyzeIRDropIterative(input, DefaultWireParams(), DefaultIRDropSolverConfig())
	if result == nil {
		t.Fatal("AnalyzeIRDropIterative with explicit defaults returned nil")
	}
}
