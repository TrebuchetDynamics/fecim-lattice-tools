// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the FeCIM theme definition for consistent branding.
package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Theme colors
var (
	colorBackground   = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
	colorPrimary      = color.RGBA{0, 212, 255, 255}   // Cyan
	colorAccent       = color.RGBA{255, 165, 0, 255}   // Orange for highlights
	colorSuccess      = color.RGBA{0, 200, 100, 255}   // Green for success
	colorWarning      = color.RGBA{255, 200, 0, 255}   // Yellow for warnings
	colorDanger       = color.RGBA{255, 80, 80, 255}   // Red for danger
	colorCPU          = color.RGBA{200, 100, 100, 255} // CPU color
	colorGPU          = color.RGBA{100, 200, 100, 255} // GPU color
	colorFeFET        = color.RGBA{100, 150, 255, 255} // FeFET color
	colorWriteZone    = color.RGBA{200, 50, 50, 200}   // Write zone (danger)
	colorReadZone     = color.RGBA{50, 150, 50, 200}   // Read zone (safe)
	colorThreshold    = color.RGBA{255, 200, 0, 200}   // Threshold line
	colorDAC          = color.RGBA{150, 100, 200, 255} // Purple for DAC
	colorADC          = color.RGBA{100, 200, 150, 255} // Green for ADC
	colorTIA          = color.RGBA{200, 150, 100, 255} // Orange for TIA
	colorArrayCell    = color.RGBA{100, 150, 200, 255} // Blue for array cells
	colorSelectedCell = color.RGBA{255, 200, 50, 255}  // Yellow for selected
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
