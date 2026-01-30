package physics

import (
	"math"
	"testing"
)

func TestDefaultISPPConfig(t *testing.T) {
	cfg := DefaultISPPConfig()

	if cfg.StartRatio != 0.7 {
		t.Errorf("StartRatio = %v, want 0.7", cfg.StartRatio)
	}
	if cfg.StepPercent != 0.05 {
		t.Errorf("StepPercent = %v, want 0.05", cfg.StepPercent)
	}
	if cfg.MaxPulses != 10 {
		t.Errorf("MaxPulses = %v, want 10", cfg.MaxPulses)
	}
	if cfg.SafetyCap != 2.2 {
		t.Errorf("SafetyCap = %v, want 2.2", cfg.SafetyCap)
	}
	if cfg.Tolerance != 0 {
		t.Errorf("Tolerance = %v, want 0", cfg.Tolerance)
	}
}

func TestNewISPPCalculator(t *testing.T) {
	ec := 1.0  // 1V coercive voltage
	levels := 30

	calc := NewISPPCalculator(ec, levels)

	if calc.Ec != ec {
		t.Errorf("Ec = %v, want %v", calc.Ec, ec)
	}
	if calc.NumLevels != levels {
		t.Errorf("NumLevels = %v, want %v", calc.NumLevels, levels)
	}
	if calc.Config.StartRatio != 0.7 {
		t.Errorf("Config.StartRatio = %v, want 0.7", calc.Config.StartRatio)
	}
}

func TestNewISPPCalculatorWithConfig(t *testing.T) {
	ec := 1.5
	levels := 64
	cfg := ISPPConfig{
		StartRatio:  0.8,
		StepPercent: 0.03,
		MaxPulses:   15,
		SafetyCap:   2.5,
		Tolerance:   1,
	}

	calc := NewISPPCalculatorWithConfig(ec, levels, cfg)

	if calc.Ec != ec {
		t.Errorf("Ec = %v, want %v", calc.Ec, ec)
	}
	if calc.Config.StartRatio != 0.8 {
		t.Errorf("Config.StartRatio = %v, want 0.8", calc.Config.StartRatio)
	}
	if calc.Config.Tolerance != 1 {
		t.Errorf("Config.Tolerance = %v, want 1", calc.Config.Tolerance)
	}
}

func TestGetDirection(t *testing.T) {
	tests := []struct {
		name     string
		current  int
		target   int
		expected HysteresisDirection
	}{
		{"same level", 15, 15, DirectionUnknown},
		{"ascending small", 10, 11, DirectionAscending},
		{"ascending large", 0, 29, DirectionAscending},
		{"descending small", 15, 14, DirectionDescending},
		{"descending large", 29, 0, DirectionDescending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDirection(tt.current, tt.target)
			if result != tt.expected {
				t.Errorf("GetDirection(%d, %d) = %v, want %v",
					tt.current, tt.target, result, tt.expected)
			}
		})
	}
}

func TestHysteresisDirectionString(t *testing.T) {
	tests := []struct {
		dir      HysteresisDirection
		expected string
	}{
		{DirectionUnknown, "Unknown"},
		{DirectionAscending, "Ascending"},
		{DirectionDescending, "Descending"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.dir.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestISPPResultString(t *testing.T) {
	tests := []struct {
		result   ISPPResult
		expected string
	}{
		{ISPPContinue, "Continue"},
		{ISPPSuccess, "Success"},
		{ISPPOvershoot, "Overshoot"},
		{ISPPMaxPulses, "MaxPulses"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.result.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCalculateStartVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		calibrated float64
		expected   float64
	}{
		{1.0, 0.7},
		{2.0, 1.4},
		{0.5, 0.35},
	}

	for _, tt := range tests {
		result := calc.CalculateStartVoltage(tt.calibrated)
		if math.Abs(result-tt.expected) > 1e-9 {
			t.Errorf("CalculateStartVoltage(%v) = %v, want %v",
				tt.calibrated, result, tt.expected)
		}
	}
}

func TestCalculateVoltageStep(t *testing.T) {
	tests := []struct {
		name     string
		ec       float64
		expected float64
	}{
		{"1V Ec", 1.0, 0.05},
		{"2V Ec", 2.0, 0.10},
		{"0.5V Ec", 0.5, 0.025},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewISPPCalculator(tt.ec, 30)
			result := calc.CalculateVoltageStep()
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("CalculateVoltageStep() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateNextVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30) // Ec = 1V, step = 0.05V

	tests := []struct {
		name      string
		current   float64
		direction HysteresisDirection
		expected  float64
	}{
		{"ascending from 0.7V", 0.7, DirectionAscending, 0.75},
		{"ascending from 2.0V", 2.0, DirectionAscending, 2.05},
		{"descending from -0.7V", -0.7, DirectionDescending, -0.75},
		{"descending from -2.0V", -2.0, DirectionDescending, -2.05},
		{"unknown direction", 0.5, DirectionUnknown, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateNextVoltage(tt.current, tt.direction)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("CalculateNextVoltage(%v, %v) = %v, want %v",
					tt.current, tt.direction, result, tt.expected)
			}
		})
	}
}

func TestClampVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30) // SafetyCap = 2.2, so max = 2.2V

	tests := []struct {
		name      string
		voltage   float64
		direction HysteresisDirection
		expected  float64
	}{
		// Ascending: clamp to [0, 2.2]
		{"ascending in range", 1.5, DirectionAscending, 1.5},
		{"ascending too high", 3.0, DirectionAscending, 2.2},
		{"ascending negative", -0.5, DirectionAscending, 0},

		// Descending: clamp to [-2.2, 0]
		{"descending in range", -1.5, DirectionDescending, -1.5},
		{"descending too low", -3.0, DirectionDescending, -2.2},
		{"descending positive", 0.5, DirectionDescending, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.ClampVoltage(tt.voltage, tt.direction)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("ClampVoltage(%v, %v) = %v, want %v",
					tt.voltage, tt.direction, result, tt.expected)
			}
		})
	}
}

func TestIsOvershoot(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		name      string
		current   int
		target    int
		direction HysteresisDirection
		expected  bool
	}{
		// Ascending: overshoot if current > target
		{"ascending undershoot", 14, 15, DirectionAscending, false},
		{"ascending exact", 15, 15, DirectionAscending, false},
		{"ascending overshoot", 16, 15, DirectionAscending, true},

		// Descending: overshoot if current < target
		{"descending undershoot", 16, 15, DirectionDescending, false},
		{"descending exact", 15, 15, DirectionDescending, false},
		{"descending overshoot", 14, 15, DirectionDescending, true},

		// Unknown: never overshoot
		{"unknown direction", 20, 15, DirectionUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.IsOvershoot(tt.current, tt.target, tt.direction)
			if result != tt.expected {
				t.Errorf("IsOvershoot(%d, %d, %v) = %v, want %v",
					tt.current, tt.target, tt.direction, result, tt.expected)
			}
		})
	}
}

func TestCheckResult(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		name       string
		current    int
		target     int
		direction  HysteresisDirection
		pulseCount int
		expected   ISPPResult
	}{
		// Success cases (exact match, tolerance = 0)
		{"success exact", 15, 15, DirectionAscending, 1, ISPPSuccess},
		{"success descending", 10, 10, DirectionDescending, 3, ISPPSuccess},

		// Overshoot cases
		{"overshoot ascending", 16, 15, DirectionAscending, 2, ISPPOvershoot},
		{"overshoot descending", 14, 15, DirectionDescending, 2, ISPPOvershoot},

		// Max pulses
		{"max pulses ascending", 14, 15, DirectionAscending, 10, ISPPMaxPulses},
		{"max pulses descending", 16, 15, DirectionDescending, 10, ISPPMaxPulses},

		// Continue cases (undershoot, not at max)
		{"continue ascending", 14, 15, DirectionAscending, 5, ISPPContinue},
		{"continue descending", 16, 15, DirectionDescending, 5, ISPPContinue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CheckResult(tt.current, tt.target, tt.direction, tt.pulseCount)
			if result != tt.expected {
				t.Errorf("CheckResult(%d, %d, %v, %d) = %v, want %v",
					tt.current, tt.target, tt.direction, tt.pulseCount, result, tt.expected)
			}
		})
	}
}

func TestCheckResultWithTolerance(t *testing.T) {
	cfg := DefaultISPPConfig()
	cfg.Tolerance = 1
	calc := NewISPPCalculatorWithConfig(1.0, 30, cfg)

	tests := []struct {
		name     string
		current  int
		target   int
		expected ISPPResult
	}{
		{"exact match", 15, 15, ISPPSuccess},
		{"within +1 tolerance", 16, 15, ISPPSuccess},
		{"within -1 tolerance", 14, 15, ISPPSuccess},
		{"beyond +1 tolerance", 17, 15, ISPPOvershoot}, // Still overshoot
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CheckResult(tt.current, tt.target, DirectionAscending, 1)
			if result != tt.expected {
				t.Errorf("CheckResult(%d, %d, ...) = %v, want %v",
					tt.current, tt.target, result, tt.expected)
			}
		})
	}
}

func TestGetResetVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30) // max = 2.2V

	tests := []struct {
		name      string
		direction HysteresisDirection
		expected  float64
	}{
		// Ascending overshoot: reset to negative saturation
		{"ascending overshoot", DirectionAscending, -2.2},
		// Descending overshoot: reset to positive saturation
		{"descending overshoot", DirectionDescending, 2.2},
		// Unknown: no reset
		{"unknown", DirectionUnknown, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.GetResetVoltage(tt.direction)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("GetResetVoltage(%v) = %v, want %v",
					tt.direction, result, tt.expected)
			}
		})
	}
}

func TestGetSaturationVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30) // max = 2.2V

	tests := []struct {
		name      string
		direction HysteresisDirection
		expected  float64
	}{
		{"ascending", DirectionAscending, 2.2},
		{"descending", DirectionDescending, -2.2},
		{"unknown", DirectionUnknown, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.GetSaturationVoltage(tt.direction)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("GetSaturationVoltage(%v) = %v, want %v",
					tt.direction, result, tt.expected)
			}
		})
	}
}

func TestLevelError(t *testing.T) {
	tests := []struct {
		current  int
		target   int
		expected int
	}{
		{15, 15, 0},
		{16, 15, 1},  // current higher
		{14, 15, -1}, // current lower
		{20, 10, 10},
		{5, 25, -20},
	}

	for _, tt := range tests {
		result := LevelError(tt.current, tt.target)
		if result != tt.expected {
			t.Errorf("LevelError(%d, %d) = %d, want %d",
				tt.current, tt.target, result, tt.expected)
		}
	}
}

func TestEstimatePulsesNeeded(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	// Same level should need 1 pulse (verify only)
	pulses := calc.EstimatePulsesNeeded(15, 15, 1.0)
	if pulses != 1 {
		t.Errorf("EstimatePulsesNeeded(same level) = %d, want 1", pulses)
	}

	// Different levels need > 1 pulse
	pulses = calc.EstimatePulsesNeeded(10, 20, 1.5)
	if pulses < 1 {
		t.Errorf("EstimatePulsesNeeded(different levels) = %d, want >= 1", pulses)
	}

	// Should not exceed MaxPulses
	if pulses > calc.Config.MaxPulses {
		t.Errorf("EstimatePulsesNeeded() = %d, exceeds MaxPulses %d",
			pulses, calc.Config.MaxPulses)
	}
}

func TestISPPWorkflow(t *testing.T) {
	// Simulate a complete ISPP workflow
	calc := NewISPPCalculator(1.0, 30)

	currentLevel := 10
	targetLevel := 15
	calibratedVoltage := 1.5

	// Get direction
	direction := GetDirection(currentLevel, targetLevel)
	if direction != DirectionAscending {
		t.Errorf("Expected ascending direction, got %v", direction)
	}

	// Calculate start voltage
	startV := calc.CalculateStartVoltage(calibratedVoltage)
	expectedStart := 1.5 * 0.7 // 1.05V
	if math.Abs(startV-expectedStart) > 1e-9 {
		t.Errorf("Start voltage = %v, want %v", startV, expectedStart)
	}

	// Simulate undershoot (current = 13, target = 15)
	currentLevel = 13
	result := calc.CheckResult(currentLevel, targetLevel, direction, 1)
	if result != ISPPContinue {
		t.Errorf("After undershoot, expected Continue, got %v", result)
	}

	// Calculate next voltage
	nextV := calc.CalculateNextVoltage(startV, direction)
	expectedNext := startV + 0.05 // 1.10V
	if math.Abs(nextV-expectedNext) > 1e-9 {
		t.Errorf("Next voltage = %v, want %v", nextV, expectedNext)
	}

	// Simulate success (current = 15, target = 15)
	currentLevel = 15
	result = calc.CheckResult(currentLevel, targetLevel, direction, 2)
	if result != ISPPSuccess {
		t.Errorf("After reaching target, expected Success, got %v", result)
	}

	// Simulate overshoot (current = 16, target = 15)
	currentLevel = 16
	result = calc.CheckResult(currentLevel, targetLevel, direction, 3)
	if result != ISPPOvershoot {
		t.Errorf("After overshoot, expected Overshoot, got %v", result)
	}

	// Get reset voltage
	resetV := calc.GetResetVoltage(direction)
	if resetV >= 0 {
		t.Errorf("Reset voltage for ascending overshoot should be negative, got %v", resetV)
	}
}

func TestISPPDescendingWorkflow(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	currentLevel := 20
	targetLevel := 12
	calibratedVoltage := -1.2

	// Get direction
	direction := GetDirection(currentLevel, targetLevel)
	if direction != DirectionDescending {
		t.Errorf("Expected descending direction, got %v", direction)
	}

	// Start voltage should be negative
	startV := calc.CalculateStartVoltage(calibratedVoltage)
	if startV > 0 {
		t.Errorf("Start voltage for descending should be negative, got %v", startV)
	}

	// Undershoot (didn't go down enough): current = 16, target = 12
	currentLevel = 16
	result := calc.CheckResult(currentLevel, targetLevel, direction, 1)
	if result != ISPPContinue {
		t.Errorf("Undershoot descending: expected Continue, got %v", result)
	}

	// Overshoot (went too far down): current = 10, target = 12
	currentLevel = 10
	result = calc.CheckResult(currentLevel, targetLevel, direction, 2)
	if result != ISPPOvershoot {
		t.Errorf("Overshoot descending: expected Overshoot, got %v", result)
	}

	// Reset should be positive
	resetV := calc.GetResetVoltage(direction)
	if resetV <= 0 {
		t.Errorf("Reset voltage for descending overshoot should be positive, got %v", resetV)
	}
}
