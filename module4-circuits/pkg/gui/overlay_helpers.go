//go:build legacy_fyne

package gui

import (
	"image"

	"fecim-lattice-tools/module4-circuits/pkg/gui/status"
	"fecim-lattice-tools/module4-circuits/pkg/gui/visual"
)

type overlayCellLabels = status.OverlayCellLabels

// overlayDrawableDims returns safe drawable dimensions constrained by backing weights.
func (ca *CircuitsApp) overlayDrawableDims(rows, cols int, weights [][]int) (int, int) {
	return visual.OverlayDrawableDims(rows, cols, weights)
}

// overlayCellInfo composes dual-line overlay labels for selected cell diagnostics.
func (ca *CircuitsApp) overlayCellInfo(level, quantLevels int, value float64, mode string) overlayCellLabels {
	_ = quantLevels // kept for compatibility with call sites/tests
	return status.FormatOverlayCellInfo(level, value, mode)
}

func cellRectFor(row, col, offsetX, offsetY, cellSize int) image.Rectangle {
	return visual.CellRectFor(row, col, offsetX, offsetY, cellSize)
}

func selectedHighlightRectFor(row, col, offsetX, offsetY, cellSize int) image.Rectangle {
	return visual.SelectedHighlightRectFor(row, col, offsetX, offsetY, cellSize)
}
