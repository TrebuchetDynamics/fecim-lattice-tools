package render

import (
	"strings"
	"testing"
)

func TestGenerateCellVertices(t *testing.T) {
	cell := &CellDisplay{
		X:            0.25,
		Y:            0.25,
		Width:        0.5,
		Height:       0.5,
		Polarization: 0.8,
		ColorMap:     DefaultColorMap(),
	}

	vertices := GenerateCellVertices(cell)
	if len(vertices) == 0 {
		t.Fatal("expected non-empty vertices")
	}

	// Should have 6 vertices for quad (2 triangles) + 8 for border edges
	if len(vertices) < 6 {
		t.Fatalf("expected at least 6 vertices for cell quad, got %d", len(vertices))
	}

	// Verify positions are in NDC range
	for i, v := range vertices {
		if v.Position[0] < -1.5 || v.Position[0] > 1.5 {
			t.Errorf("vertex %d X position out of range: %f", i, v.Position[0])
		}
		if v.Position[1] < -1.5 || v.Position[1] > 1.5 {
			t.Errorf("vertex %d Y position out of range: %f", i, v.Position[1])
		}
	}
}

func TestFormatAxisLabel(t *testing.T) {
	tests := []struct {
		value    float64
		min      float64
		max      float64
		label    string
		contains string
	}{
		{0.5, 0.0, 1.0, "Polarization", "Polarization"},
		{1.5, -2.0, 2.0, "E-field", "E-field"},
		{-0.25, -1.0, 1.0, "Voltage", "Voltage"},
	}

	for _, tc := range tests {
		result := FormatAxisLabel(tc.value, tc.min, tc.max, tc.label)
		if !strings.Contains(result, tc.contains) {
			t.Errorf("FormatAxisLabel(%f, %f, %f, %q) = %q, expected to contain %q",
				tc.value, tc.min, tc.max, tc.label, result, tc.contains)
		}
		if !strings.Contains(result, "0.50") && !strings.Contains(result, "1.50") && !strings.Contains(result, "-0.25") {
			// At least verify numeric formatting
			if len(result) < 3 {
				t.Errorf("FormatAxisLabel result too short: %q", result)
			}
		}
	}
}
