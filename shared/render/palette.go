package render

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// VisualizationPalette is the shared semantic color contract for FeCIM plots,
// heatmaps, axes, and scientific annotations.
type VisualizationPalette struct {
	Background     color.Color
	Surface        color.Color
	Text           color.Color
	TextMuted      color.Color
	Axis           color.Color
	Grid           color.Color
	TracePrimary   color.Color
	TraceSecondary color.Color
	Positive       color.Color
	Negative       color.Color
	Success        color.Color
	Warning        color.Color
	Error          color.Color
	HeatmapStops   []color.Color
}

// VisualizationPaletteFromTheme derives a plot/heatmap palette from a Fyne
// theme and variant without consulting global app state. Passing nil uses the
// default Fyne theme.
func VisualizationPaletteFromTheme(th fyne.Theme, variant fyne.ThemeVariant) VisualizationPalette {
	if th == nil {
		th = theme.DefaultTheme()
	}

	primary := th.Color(theme.ColorNamePrimary, variant)
	foreground := th.Color(theme.ColorNameForeground, variant)
	background := th.Color(theme.ColorNameBackground, variant)
	separator := th.Color(theme.ColorNameSeparator, variant)
	warning := th.Color(theme.ColorNameWarning, variant)
	success := th.Color(theme.ColorNameSuccess, variant)
	errorColor := th.Color(theme.ColorNameError, variant)

	return VisualizationPalette{
		Background:     background,
		Surface:        th.Color(theme.ColorNameMenuBackground, variant),
		Text:           foreground,
		TextMuted:      withPaletteAlpha(foreground, 180),
		Axis:           foreground,
		Grid:           withPaletteAlpha(separator, 120),
		TracePrimary:   primary,
		TraceSecondary: th.Color(theme.ColorNameHyperlink, variant),
		Positive:       errorColor,
		Negative:       primary,
		Success:        success,
		Warning:        warning,
		Error:          errorColor,
		HeatmapStops: []color.Color{
			color.NRGBA{R: 68, G: 1, B: 84, A: 255},
			color.NRGBA{R: 59, G: 82, B: 139, A: 255},
			color.NRGBA{R: 33, G: 145, B: 140, A: 255},
			color.NRGBA{R: 94, G: 201, B: 98, A: 255},
			color.NRGBA{R: 253, G: 231, B: 37, A: 255},
		},
	}
}

// HeatmapColor returns an interpolated color from HeatmapStops for value in
// [0,1]. Out-of-range values are clamped.
func (p VisualizationPalette) HeatmapColor(value float64) color.Color {
	stops := p.HeatmapStops
	if len(stops) == 0 {
		stops = VisualizationPaletteFromTheme(nil, theme.VariantDark).HeatmapStops
	}
	if len(stops) == 1 {
		return stops[0]
	}

	value = math.Max(0, math.Min(1, value))
	position := value * float64(len(stops)-1)
	idx := int(math.Floor(position))
	if idx >= len(stops)-1 {
		return stops[len(stops)-1]
	}
	return interpolateColor(stops[idx], stops[idx+1], position-float64(idx))
}

func interpolateColor(a, b color.Color, t float64) color.Color {
	ar, ag, ab, aa := rgba8(a)
	br, bg, bb, ba := rgba8(b)
	return color.NRGBA{
		R: interpolateChannel(ar, br, t),
		G: interpolateChannel(ag, bg, t),
		B: interpolateChannel(ab, bb, t),
		A: interpolateChannel(aa, ba, t),
	}
}

func interpolateChannel(a, b uint8, t float64) uint8 {
	return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t))
}

func withPaletteAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := rgba8(c)
	return color.NRGBA{R: r, G: g, B: b, A: alpha}
}

func rgba8(c color.Color) (uint8, uint8, uint8, uint8) {
	r, g, b, a := c.RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)
}
