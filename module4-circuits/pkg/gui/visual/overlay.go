//go:build legacy_fyne

package visual

import "image"

// OverlayDrawableDims returns safe drawable dimensions constrained by backing weights.
func OverlayDrawableDims(rows, cols int, weights [][]int) (int, int) {
	drawRows := rows
	if len(weights) < drawRows {
		drawRows = len(weights)
	}
	drawCols := cols
	for r := 0; r < drawRows; r++ {
		if r >= len(weights) {
			break
		}
		if len(weights[r]) < drawCols {
			drawCols = len(weights[r])
		}
	}
	if drawRows < 0 {
		drawRows = 0
	}
	if drawCols < 0 {
		drawCols = 0
	}
	return drawRows, drawCols
}

// CellRectFor returns the image rectangle for one array cell.
func CellRectFor(row, col, offsetX, offsetY, cellSize int) image.Rectangle {
	x0 := offsetX + col*cellSize
	y0 := offsetY + row*cellSize
	return image.Rect(x0, y0, x0+cellSize, y0+cellSize)
}

// SelectedHighlightRectFor returns the selected-cell highlight rectangle.
func SelectedHighlightRectFor(row, col, offsetX, offsetY, cellSize int) image.Rectangle {
	cell := CellRectFor(row, col, offsetX, offsetY, cellSize)
	return cell.Inset(1)
}
