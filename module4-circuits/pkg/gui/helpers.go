// Package gui provides UI components for the circuits module.
package gui

import (
	"image"
	"image/color"
	"time"
)

// drawRect fills a rectangular region in an RGBA image with the specified color.
// It performs boundary checks to ensure all pixels are within image bounds.
func drawRect(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	for py := y; py < y+rectH; py++ {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.Set(px, py, c)
			}
		}
	}
}

// drawRectBorder draws only the border (outline) of a rectangular region.
// It performs boundary checks to ensure all pixels are within image bounds.
func drawRectBorder(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	// Top edge
	if y >= 0 && y < maxY {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < maxX {
				img.Set(px, y, c)
			}
		}
	}
	// Bottom edge
	bottomY := y + rectH - 1
	if bottomY >= 0 && bottomY < maxY {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < maxX {
				img.Set(px, bottomY, c)
			}
		}
	}
	// Left edge
	if x >= 0 && x < maxX {
		for py := y; py < y+rectH; py++ {
			if py >= 0 && py < maxY {
				img.Set(x, py, c)
			}
		}
	}
	// Right edge
	rightX := x + rectW - 1
	if rightX >= 0 && rightX < maxX {
		for py := y; py < y+rectH; py++ {
			if py >= 0 && py < maxY {
				img.Set(rightX, py, c)
			}
		}
	}
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sleep pauses execution for the specified number of milliseconds.
// Used for animation timing.
func (ca *CircuitsApp) sleep(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}
