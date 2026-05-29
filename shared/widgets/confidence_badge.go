//go:build legacy_fyne

package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/widgets/display"
)

// ConfidenceLevel identifies the confidence tier displayed in a badge.
type ConfidenceLevel = display.ConfidenceLevel

const (
	Measured    = display.Measured
	Calibrated  = display.Calibrated
	Estimated   = display.Estimated
	Placeholder = display.Placeholder
)

func confidenceColor(level ConfidenceLevel) color.Color {
	return display.ConfidenceColor(level)
}

// NewConfidenceBadge creates a compact color-dot + text confidence badge.
func NewConfidenceBadge(level ConfidenceLevel) *fyne.Container {
	return display.NewConfidenceBadge(level)
}
