//go:build legacy_fyne

package widgets

import (
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/display"
)

// FormatUncertaintyAnnotation returns text in the required form:
// "value ± uncertainty (confidence%)".
func FormatUncertaintyAnnotation(value, uncertainty float64, confidencePct float64) string {
	return display.FormatUncertaintyAnnotation(value, uncertainty, confidencePct)
}

// NewUncertaintyOverlay creates a readout annotation label for UI overlays.
// Text format is fixed as: "value ± uncertainty (confidence%)".
func NewUncertaintyOverlay(value, uncertainty float64, confidencePct float64) *widget.Label {
	return display.NewUncertaintyOverlay(value, uncertainty, confidencePct)
}
