//go:build legacy_fyne

package display

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ConfidenceLevel identifies the confidence tier displayed in a badge.
type ConfidenceLevel string

const (
	Measured    ConfidenceLevel = "Measured"
	Calibrated  ConfidenceLevel = "Calibrated"
	Estimated   ConfidenceLevel = "Estimated"
	Placeholder ConfidenceLevel = "Placeholder"
)

// ConfidenceColor returns the color used for a confidence tier.
func ConfidenceColor(level ConfidenceLevel) color.Color {
	switch level {
	case Measured:
		return color.RGBA{60, 180, 75, 255} // green
	case Calibrated:
		return color.RGBA{70, 130, 255, 255} // blue
	case Estimated:
		return color.RGBA{240, 190, 60, 255} // yellow
	default:
		return color.RGBA{220, 80, 80, 255} // red
	}
}

// NewConfidenceBadge creates a compact color-dot + text confidence badge.
func NewConfidenceBadge(level ConfidenceLevel) *fyne.Container {
	if level == "" {
		level = Placeholder
	}
	dot := canvas.NewCircle(ConfidenceColor(level))
	dotWrap := container.NewGridWrap(fyne.NewSize(10, 10), dot)
	label := widget.NewLabel(string(level))
	label.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewHBox(dotWrap, label)
}
