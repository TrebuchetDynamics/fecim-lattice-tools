//go:build legacy_fyne

// Package visual contains pure color and drawing-adjacent helpers for module 4 visualizations.
package visual

import (
	"image/color"
	"math"
)

// VCellOverlayColor maps a signed cell voltage to a cool-neutral-warm color ramp.
func VCellOverlayColor(voltage, maxAbs float64) color.RGBA {
	if maxAbs <= 0 {
		maxAbs = 1.0
	}
	n := voltage / maxAbs
	if n > 1 {
		n = 1
	}
	if n < -1 {
		n = -1
	}

	cool := color.RGBA{70, 130, 255, 255}     // blue: negative voltage
	neutral := color.RGBA{255, 255, 255, 255} // white: zero voltage
	warm := color.RGBA{255, 80, 80, 255}      // red: positive voltage

	if n < 0 {
		return LerpRGBA(neutral, cool, -n)
	}
	return LerpRGBA(neutral, warm, n)
}

// LerpRGBA linearly interpolates between two opaque RGBA colors.
func LerpRGBA(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	mix := func(x, y uint8) uint8 {
		return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t))
	}
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}

// BlendRGBA alpha-blends overlay onto base and returns an opaque color.
func BlendRGBA(base, overlay color.RGBA, alpha float64) color.RGBA {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	mix := func(a, b uint8) uint8 {
		return uint8(math.Round((1-alpha)*float64(a) + alpha*float64(b)))
	}
	return color.RGBA{mix(base.R, overlay.R), mix(base.G, overlay.G), mix(base.B, overlay.B), 255}
}
