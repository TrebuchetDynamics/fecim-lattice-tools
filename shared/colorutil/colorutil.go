// Package colorutil provides shared color manipulation utilities for FeCIM UI.
// These functions are used by both shared/theme and shared/themes to avoid
// duplicated luminance and alpha-blending logic.
package colorutil

import "image/color"

// WithAlpha returns a new RGBA color with the specified alpha channel (0-255).
// The RGB values are preserved from the input color.
func WithAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: alpha,
	}
}

// Luminance returns the relative luminance of a color using the sRGB luminance
// coefficients (0.299 R + 0.587 G + 0.114 B), normalized to [0, 1].
func Luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	r8 := float64(r>>8) / 255.0
	g8 := float64(g>>8) / 255.0
	b8 := float64(b>>8) / 255.0
	return 0.299*r8 + 0.587*g8 + 0.114*b8
}

// GetContrastColor returns either pure white (255,255,255) or pure black (0,0,0)
// based on the luminance of the background color, ensuring readable contrast.
// Returns white for dark backgrounds (luminance < 0.5) and black for light backgrounds.
func GetContrastColor(bgColor color.Color) color.Color {
	if Luminance(bgColor) < 0.5 {
		return color.RGBA{255, 255, 255, 255}
	}
	return color.RGBA{0, 0, 0, 255}
}
