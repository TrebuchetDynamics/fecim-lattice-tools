package gui

import (
	"math"
	"testing"
)

func TestComputeSignedDifference(t *testing.T) {
	// Create test widget with test data
	widget := NewBeforeAfterToggle(3, 3)

	// Set up ideal data (uniform)
	ideal := [][]float64{
		{0.5, 0.5, 0.5},
		{0.5, 0.5, 0.5},
		{0.5, 0.5, 0.5},
	}

	// Set up actual data (with variations)
	actual := [][]float64{
		{0.3, 0.5, 0.7}, // Less, same, more than ideal
		{0.4, 0.5, 0.6},
		{0.2, 0.5, 0.8},
	}

	widget.SetData(ideal, actual)

	// Compute difference
	diff := widget.computeSignedDifference()

	// Verify dimensions
	if len(diff) != 3 || len(diff[0]) != 3 {
		t.Fatalf("Expected 3x3 difference matrix, got %dx%d", len(diff), len(diff[0]))
	}

	// Find max absolute difference for normalization check
	maxAbsDiff := 0.0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			absDiff := math.Abs(actual[i][j] - ideal[i][j])
			if absDiff > maxAbsDiff {
				maxAbsDiff = absDiff
			}
		}
	}

	// Verify normalization: all values should be in [-1, 1]
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if diff[i][j] < -1.0 || diff[i][j] > 1.0 {
				t.Errorf("Difference at [%d][%d] = %.3f is out of range [-1, 1]", i, j, diff[i][j])
			}
		}
	}

	// Verify sign correctness
	// diff[0][0]: actual=0.3, ideal=0.5, should be negative (actual < ideal)
	if diff[0][0] >= 0 {
		t.Errorf("Expected negative difference at [0][0], got %.3f", diff[0][0])
	}

	// diff[0][1]: actual=0.5, ideal=0.5, should be zero
	if math.Abs(diff[0][1]) > 1e-9 {
		t.Errorf("Expected zero difference at [0][1], got %.3f", diff[0][1])
	}

	// diff[0][2]: actual=0.7, ideal=0.5, should be positive (actual > ideal)
	if diff[0][2] <= 0 {
		t.Errorf("Expected positive difference at [0][2], got %.3f", diff[0][2])
	}

	// Verify max difference is normalized to ±1
	expectedMax := (actual[2][2] - ideal[2][2]) / maxAbsDiff // 0.3 / 0.3 = 1.0
	if math.Abs(diff[2][2]-expectedMax) > 1e-9 {
		t.Errorf("Expected max difference at [2][2] = %.3f, got %.3f", expectedMax, diff[2][2])
	}

	expectedMin := (actual[2][0] - ideal[2][0]) / maxAbsDiff // -0.3 / 0.3 = -1.0
	if math.Abs(diff[2][0]-expectedMin) > 1e-9 {
		t.Errorf("Expected min difference at [2][0] = %.3f, got %.3f", expectedMin, diff[2][0])
	}
}

func TestComputeStats(t *testing.T) {
	widget := NewBeforeAfterToggle(2, 2)

	// Simple test case
	ideal := [][]float64{
		{0.0, 1.0},
		{0.5, 0.5},
	}

	actual := [][]float64{
		{0.1, 0.9}, // Differences: 0.1, -0.1
		{0.6, 0.4}, // Differences: 0.1, -0.1
	}

	widget.SetData(ideal, actual)

	rmse, mae, maxDiff := widget.computeStats()

	// Expected values
	// All differences: 0.1, -0.1, 0.1, -0.1
	// MAE = (0.1 + 0.1 + 0.1 + 0.1) / 4 = 0.1
	// RMSE = sqrt((0.01 + 0.01 + 0.01 + 0.01) / 4) = sqrt(0.01) = 0.1
	// Max = 0.1

	if math.Abs(mae-0.1) > 1e-9 {
		t.Errorf("Expected MAE = 0.1, got %.3f", mae)
	}

	if math.Abs(rmse-0.1) > 1e-9 {
		t.Errorf("Expected RMSE = 0.1, got %.3f", rmse)
	}

	if math.Abs(maxDiff-0.1) > 1e-9 {
		t.Errorf("Expected Max Diff = 0.1, got %.3f", maxDiff)
	}
}

func TestModeChanges(t *testing.T) {
	widget := NewBeforeAfterToggle(2, 2)

	ideal := [][]float64{{0.2, 0.3}, {0.4, 0.5}}
	actual := [][]float64{{0.3, 0.4}, {0.5, 0.6}}

	widget.SetData(ideal, actual)

	// Test split mode
	widget.SetMode("split")
	if widget.mode != "split" {
		t.Errorf("Expected mode 'split', got '%s'", widget.mode)
	}

	// Test before mode
	widget.SetMode("before")
	if widget.mode != "before" {
		t.Errorf("Expected mode 'before', got '%s'", widget.mode)
	}

	// Test after mode
	widget.SetMode("after")
	if widget.mode != "after" {
		t.Errorf("Expected mode 'after', got '%s'", widget.mode)
	}

	// Test diff mode
	widget.SetMode("diff")
	if widget.mode != "diff" {
		t.Errorf("Expected mode 'diff', got '%s'", widget.mode)
	}

	// Verify colormap changes for diff mode
	if widget.leftHeatmap.colormap != "diverging" {
		t.Errorf("Expected diverging colormap in diff mode, got '%s'", widget.leftHeatmap.colormap)
	}
}
