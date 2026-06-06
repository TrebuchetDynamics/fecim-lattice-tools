package render

import (
	"image/color"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestVisualizationPaletteFromThemeProvidesSemanticPlotColors(t *testing.T) {
	fyneTest.NewTempApp(t)

	palette := VisualizationPaletteFromTheme(theme.DarkTheme(), theme.VariantDark)

	assertOpaque(t, "Background", palette.Background)
	assertOpaque(t, "Text", palette.Text)
	assertOpaque(t, "Axis", palette.Axis)
	assertVisible(t, "Grid", palette.Grid)
	assertOpaque(t, "TracePrimary", palette.TracePrimary)
	assertOpaque(t, "TraceSecondary", palette.TraceSecondary)
	assertOpaque(t, "Positive", palette.Positive)
	assertOpaque(t, "Negative", palette.Negative)
	assertOpaque(t, "Success", palette.Success)
	assertOpaque(t, "Warning", palette.Warning)
	assertOpaque(t, "Error", palette.Error)

	if rgba(palette.Background) == rgba(palette.Axis) {
		t.Fatal("axis color should contrast with background")
	}
	if rgba(palette.Positive) == rgba(palette.Negative) {
		t.Fatal("positive and negative traces should be visually distinct")
	}
}

func TestVisualizationPaletteHeatmapClampsAndInterpolates(t *testing.T) {
	fyneTest.NewTempApp(t)

	palette := VisualizationPaletteFromTheme(theme.LightTheme(), theme.VariantLight)

	low := palette.HeatmapColor(-1)
	zero := palette.HeatmapColor(0)
	mid := palette.HeatmapColor(0.5)
	one := palette.HeatmapColor(1)
	high := palette.HeatmapColor(2)

	if rgba(low) != rgba(zero) {
		t.Fatalf("HeatmapColor should clamp low values: low=%#v zero=%#v", low, zero)
	}
	if rgba(high) != rgba(one) {
		t.Fatalf("HeatmapColor should clamp high values: high=%#v one=%#v", high, one)
	}
	if rgba(mid) == rgba(zero) || rgba(mid) == rgba(one) {
		t.Fatalf("mid heatmap color should be interpolated, got mid=%#v zero=%#v one=%#v", mid, zero, one)
	}
}

func assertOpaque(t *testing.T, name string, c color.Color) {
	t.Helper()
	_, _, _, a := c.RGBA()
	if a != 0xffff {
		t.Fatalf("%s alpha = %#x, want opaque", name, a)
	}
}

func assertVisible(t *testing.T, name string, c color.Color) {
	t.Helper()
	_, _, _, a := c.RGBA()
	if a == 0 {
		t.Fatalf("%s alpha = 0, want visible", name)
	}
}

func rgba(c color.Color) [4]uint32 {
	r, g, b, a := c.RGBA()
	return [4]uint32{r, g, b, a}
}
