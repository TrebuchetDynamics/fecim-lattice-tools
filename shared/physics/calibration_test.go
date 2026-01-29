package physics

import (
	"testing"
)

func TestNewCalibrator(t *testing.T) {
	Ec := 1e8 // 1 MV/cm in V/m
	c := NewCalibrator(30, Ec)

	if c.NumLevels != 30 {
		t.Errorf("NumLevels = %d, want 30", c.NumLevels)
	}
	if c.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", c.MaxRetries)
	}
	if len(c.Up) != 30 {
		t.Errorf("len(Up) = %d, want 30", len(c.Up))
	}
	if len(c.Down) != 30 {
		t.Errorf("len(Down) = %d, want 30", len(c.Down))
	}
}

func TestUpdateAscending(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(30, Ec)

	// Set initial value
	c.Up[15].Value = 1.2 * Ec
	c.Up[15].Bounds = CalibrationBounds{Low: 0.5 * Ec, High: 2.0 * Ec}

	// Simulate overshoot (error > 0) - field too strong
	newVal := c.UpdateAscending(15, +3)
	if newVal >= 1.2*Ec {
		t.Errorf("After overshoot, newVal = %.2f*Ec should be < 1.2*Ec", newVal/Ec)
	}
	if c.Up[15].Bounds.High >= 2.0*Ec {
		t.Errorf("Upper bound should have decreased from 2.0*Ec")
	}

	// Simulate undershoot (error < 0) - field too weak
	c.Up[15].Value = 0.8 * Ec
	c.Up[15].Bounds = CalibrationBounds{Low: 0.5 * Ec, High: 2.0 * Ec}
	newVal = c.UpdateAscending(15, -2)
	if newVal <= 0.8*Ec {
		t.Errorf("After undershoot, newVal = %.2f*Ec should be > 0.8*Ec", newVal/Ec)
	}
}

func TestUpdateDescending(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(30, Ec)

	// For level 15 (middle), the field constraints are roughly -1.2*Ec to -0.7*Ec
	// Set initial value within constraints
	c.Down[15].Value = -0.9 * Ec
	c.Down[15].Bounds = CalibrationBounds{Low: -1.5 * Ec, High: -0.5 * Ec}

	// Simulate overshoot UP (error > 0) - didn't go negative enough
	// The binary search should try a more negative value
	originalVal := c.Down[15].Value
	c.UpdateDescending(15, +3)
	// Bounds should be updated: lower bound moves up
	if c.Down[15].Bounds.Low <= -1.5*Ec {
		t.Errorf("After overshoot UP, lower bound should have increased")
	}

	// Simulate undershoot (error < 0) - went too negative
	c.Down[15].Value = -1.1 * Ec
	c.Down[15].Bounds = CalibrationBounds{Low: -1.5 * Ec, High: -0.5 * Ec}
	c.UpdateDescending(15, -2)
	// Bounds should be updated: upper bound moves down
	if c.Down[15].Bounds.High >= -0.5*Ec {
		t.Errorf("After undershoot, upper bound should have decreased (more negative)")
	}

	// Verify the value is within field constraints
	minE, maxE := c.FieldConstraintsDescending(15)
	if c.Down[15].Value < minE || c.Down[15].Value > maxE {
		t.Errorf("Value %.2f*Ec should be within [%.2f, %.2f]*Ec",
			c.Down[15].Value/Ec, minE/Ec, maxE/Ec)
	}

	_ = originalVal // suppress unused warning
}

func TestFieldConstraintsAscending(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(30, Ec)

	// Level 0 should have lowest field range
	minE0, maxE0 := c.FieldConstraintsAscending(0)
	// Level 29 should have highest field range
	minE29, maxE29 := c.FieldConstraintsAscending(29)

	if minE29 <= minE0 {
		t.Errorf("minE for level 29 (%.2f*Ec) should be > minE for level 0 (%.2f*Ec)",
			minE29/Ec, minE0/Ec)
	}
	if maxE29 <= maxE0 {
		t.Errorf("maxE for level 29 (%.2f*Ec) should be > maxE for level 0 (%.2f*Ec)",
			maxE29/Ec, maxE0/Ec)
	}
}

func TestFieldConstraintsDescending(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(30, Ec)

	// Level 0 should have most negative field range
	minE0, maxE0 := c.FieldConstraintsDescending(0)
	// Level 29 should have least negative field range
	minE29, maxE29 := c.FieldConstraintsDescending(29)

	if minE0 >= minE29 {
		t.Errorf("minE for level 0 (%.2f*Ec) should be < minE for level 29 (%.2f*Ec)",
			minE0/Ec, minE29/Ec)
	}
	if maxE0 >= maxE29 {
		t.Errorf("maxE for level 0 (%.2f*Ec) should be < maxE for level 29 (%.2f*Ec)",
			maxE0/Ec, maxE29/Ec)
	}
}

func TestEnforceMonotonicityAscending(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(30, Ec)

	// Create a non-monotonic sequence
	c.Up[10].Value = 1.0 * Ec
	c.Up[11].Value = 0.8 * Ec // Spike! Lower than previous
	c.Up[12].Value = 1.2 * Ec

	c.EnforceMonotonicityAscending(11)

	if c.Up[11].Value <= c.Up[10].Value {
		t.Errorf("After enforcement, Up[11] (%.2f*Ec) should be > Up[10] (%.2f*Ec)",
			c.Up[11].Value/Ec, c.Up[10].Value/Ec)
	}
}

func TestEnforceGlobalMonotonicity(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(10, Ec)

	// Create random values
	c.Up[0].Value = 0.5 * Ec
	c.Up[1].Value = 0.4 * Ec // Non-monotonic
	c.Up[2].Value = 0.9 * Ec
	c.Up[3].Value = 0.6 * Ec // Non-monotonic
	c.Up[4].Value = 1.0 * Ec

	c.EnforceGlobalMonotonicity()

	// Verify ascending is monotonic
	for i := 1; i < len(c.Up); i++ {
		if c.Up[i].Value <= c.Up[i-1].Value {
			t.Errorf("After global enforcement, Up[%d] (%.4f) should be > Up[%d] (%.4f)",
				i, c.Up[i].Value, i-1, c.Up[i-1].Value)
		}
	}
}

func TestCheckVerify(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(30, Ec)
	c.Tolerance = 1
	c.MaxRetries = 3

	// Test success case
	result := c.CheckVerify(15, 15, 0)
	if !result.Success {
		t.Error("Exact match should be success")
	}
	if result.ShouldRetry {
		t.Error("Success should not retry")
	}

	// Test within tolerance
	result = c.CheckVerify(15, 16, 0)
	if !result.Success {
		t.Error("Within tolerance (1 level) should be success")
	}

	// Test failure with retries remaining
	result = c.CheckVerify(15, 20, 0)
	if result.Success {
		t.Error("5-level error should be failure")
	}
	if !result.ShouldRetry {
		t.Error("Should retry when retries remaining")
	}

	// Test failure with max retries reached
	result = c.CheckVerify(15, 20, 3)
	if result.Success {
		t.Error("5-level error should still be failure")
	}
	if result.ShouldRetry {
		t.Error("Should not retry when max retries reached")
	}
}

func TestGetSetValues(t *testing.T) {
	Ec := 1e8
	c := NewCalibrator(5, Ec)

	// Set values
	upVals := []float64{0.5 * Ec, 0.7 * Ec, 0.9 * Ec, 1.1 * Ec, 1.3 * Ec}
	c.SetAscendingValues(upVals)

	// Get values and verify
	gotVals := c.GetAscendingValues()
	for i, v := range upVals {
		if gotVals[i] != v {
			t.Errorf("GetAscendingValues[%d] = %f, want %f", i, gotVals[i], v)
		}
	}
}
