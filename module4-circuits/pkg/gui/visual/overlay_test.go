//go:build legacy_fyne

package visual

import "testing"

func TestOverlayDrawableDims_ConstrainedByWeights(t *testing.T) {
	weights := [][]int{{1, 2, 3}, {4, 5}}
	rows, cols := OverlayDrawableDims(4, 4, weights)
	if rows != 2 || cols != 2 {
		t.Fatalf("expected drawable dims 2x2, got %dx%d", rows, cols)
	}
}

func TestOverlayDrawableDims_ClampsNegative(t *testing.T) {
	rows, cols := OverlayDrawableDims(-1, -2, nil)
	if rows != 0 || cols != 0 {
		t.Fatalf("expected clamped dims 0x0, got %dx%d", rows, cols)
	}
}

func TestSelectedHighlightRectFor_InsideCell(t *testing.T) {
	cell := CellRectFor(3, 5, 10, 20, 8)
	highlight := SelectedHighlightRectFor(3, 5, 10, 20, 8)
	if !highlight.In(cell) {
		t.Fatalf("highlight %v should be inside cell %v", highlight, cell)
	}
	if highlight.Dx() != cell.Dx()-2 || highlight.Dy() != cell.Dy()-2 {
		t.Fatalf("highlight should inset by 1px, cell=%v highlight=%v", cell, highlight)
	}
}
