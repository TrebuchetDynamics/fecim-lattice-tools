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

// drawRoundedRect draws a filled rectangle with rounded corners.
// cornerRadius specifies the radius of the corner rounding.
func drawRoundedRect(img *image.RGBA, x, y, rectW, rectH, cornerRadius int, c color.Color) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	for py := y; py < y+rectH; py++ {
		for px := x; px < x+rectW; px++ {
			if px < 0 || px >= maxX || py < 0 || py >= maxY {
				continue
			}

			// Check if pixel is in corner region
			inCorner := false
			var cx, cy int

			// Top-left corner
			if px < x+cornerRadius && py < y+cornerRadius {
				cx, cy = x+cornerRadius, y+cornerRadius
				inCorner = true
			}
			// Top-right corner
			if px >= x+rectW-cornerRadius && py < y+cornerRadius {
				cx, cy = x+rectW-cornerRadius-1, y+cornerRadius
				inCorner = true
			}
			// Bottom-left corner
			if px < x+cornerRadius && py >= y+rectH-cornerRadius {
				cx, cy = x+cornerRadius, y+rectH-cornerRadius-1
				inCorner = true
			}
			// Bottom-right corner
			if px >= x+rectW-cornerRadius && py >= y+rectH-cornerRadius {
				cx, cy = x+rectW-cornerRadius-1, y+rectH-cornerRadius-1
				inCorner = true
			}

			if inCorner {
				dx := px - cx
				dy := py - cy
				if dx*dx+dy*dy > cornerRadius*cornerRadius {
					continue // Outside rounded corner
				}
			}

			img.Set(px, py, c)
		}
	}
}

// drawGradientRect draws a rectangle with vertical gradient from topColor to bottomColor.
func drawGradientRect(img *image.RGBA, x, y, rectW, rectH int, topColor, bottomColor color.RGBA) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	for py := y; py < y+rectH; py++ {
		if py < 0 || py >= maxY {
			continue
		}
		// Calculate interpolation factor
		t := float64(py-y) / float64(rectH)

		// Interpolate colors
		r := uint8(float64(topColor.R)*(1-t) + float64(bottomColor.R)*t)
		g := uint8(float64(topColor.G)*(1-t) + float64(bottomColor.G)*t)
		b := uint8(float64(topColor.B)*(1-t) + float64(bottomColor.B)*t)
		a := uint8(float64(topColor.A)*(1-t) + float64(bottomColor.A)*t)

		rowColor := color.RGBA{r, g, b, a}

		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < maxX {
				img.Set(px, py, rowColor)
			}
		}
	}
}

// drawGlowCircle draws a circle with a soft glow effect.
func drawGlowCircle(img *image.RGBA, cx, cy, radius int, centerColor, glowColor color.RGBA) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	glowRadius := radius + 3

	for dy := -glowRadius; dy <= glowRadius; dy++ {
		for dx := -glowRadius; dx <= glowRadius; dx++ {
			px, py := cx+dx, cy+dy
			if px < 0 || px >= maxX || py < 0 || py >= maxY {
				continue
			}

			dist := dx*dx + dy*dy

			if dist <= radius*radius {
				// Inner solid circle
				img.Set(px, py, centerColor)
			} else if dist <= glowRadius*glowRadius {
				// Glow region - fade out
				t := float64(dist-radius*radius) / float64(glowRadius*glowRadius-radius*radius)
				alpha := uint8(float64(glowColor.A) * (1 - t))
				if alpha > 10 {
					// Blend with existing pixel
					existing := img.RGBAAt(px, py)
					blendAlpha := float64(alpha) / 255.0
					r := uint8(float64(glowColor.R)*blendAlpha + float64(existing.R)*(1-blendAlpha))
					g := uint8(float64(glowColor.G)*blendAlpha + float64(existing.G)*(1-blendAlpha))
					b := uint8(float64(glowColor.B)*blendAlpha + float64(existing.B)*(1-blendAlpha))
					img.Set(px, py, color.RGBA{r, g, b, 255})
				}
			}
		}
	}
}

// levelToColor converts a FeCIM level (0-29) to a visually appealing color.
// Uses a blue->cyan->green->yellow->red gradient for better visual differentiation.
func levelToColor(level, maxLevel int) color.RGBA {
	if maxLevel <= 1 {
		return color.RGBA{100, 100, 200, 255}
	}

	t := float64(level) / float64(maxLevel-1) // 0.0 to 1.0

	var r, g, b uint8

	if t < 0.25 {
		// Blue to Cyan (0.0 - 0.25)
		s := t / 0.25
		r = uint8(30)
		g = uint8(80 + s*120)
		b = uint8(200)
	} else if t < 0.5 {
		// Cyan to Green (0.25 - 0.5)
		s := (t - 0.25) / 0.25
		r = uint8(30 + s*50)
		g = uint8(200)
		b = uint8(200 - s*150)
	} else if t < 0.75 {
		// Green to Yellow (0.5 - 0.75)
		s := (t - 0.5) / 0.25
		r = uint8(80 + s*175)
		g = uint8(200)
		b = uint8(50 - s*20)
	} else {
		// Yellow to Red (0.75 - 1.0)
		s := (t - 0.75) / 0.25
		r = uint8(255)
		g = uint8(200 - s*150)
		b = uint8(30)
	}

	return color.RGBA{r, g, b, 255}
}

// drawThickLine draws a line with specified thickness.
func drawThickLine(img *image.RGBA, x1, y1, x2, y2, thickness int, c color.Color) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	// Simple Bresenham-like line drawing with thickness
	dx := x2 - x1
	dy := y2 - y1

	steps := max(abs(dx), abs(dy))
	if steps == 0 {
		steps = 1
	}

	xInc := float64(dx) / float64(steps)
	yInc := float64(dy) / float64(steps)

	halfT := thickness / 2

	x := float64(x1)
	y := float64(y1)

	for i := 0; i <= steps; i++ {
		px := int(x)
		py := int(y)

		// Draw thick point
		for ty := -halfT; ty <= halfT; ty++ {
			for tx := -halfT; tx <= halfT; tx++ {
				ppx, ppy := px+tx, py+ty
				if ppx >= 0 && ppx < maxX && ppy >= 0 && ppy < maxY {
					img.Set(ppx, ppy, c)
				}
			}
		}

		x += xInc
		y += yInc
	}
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
