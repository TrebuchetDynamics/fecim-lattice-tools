package widgets

import (
	"image/color"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	sharedrender "fecim-lattice-tools/shared/render"
)

func TestColorLegendCanUseVisualizationPaletteHeatmap(t *testing.T) {
	fyneTest.NewTempApp(t)

	palette := sharedrender.VisualizationPaletteFromTheme(theme.DarkTheme(), theme.VariantDark)
	legend := NewColorLegendWithPalette(0, 30, "levels", false, palette)

	if legend == nil {
		t.Fatal("NewColorLegendWithPalette returned nil")
	}
	if legend.colormapName != "visualization-palette" {
		t.Fatalf("colormapName = %q, want visualization-palette", legend.colormapName)
	}

	got := legend.colorFunc(0.5)
	want := palette.HeatmapColor(0.5)
	if rgba32(got) != rgba32(want) {
		t.Fatalf("legend midpoint color = %#v, want palette midpoint %#v", got, want)
	}
}

func rgba32(c color.Color) [4]uint32 {
	r, g, b, a := c.RGBA()
	return [4]uint32{r, g, b, a}
}
