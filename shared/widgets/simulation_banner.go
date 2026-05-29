//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/widgets/display"
)

// NewSimulationBanner returns a thin horizontal bar with an amber background
// and white text reading "Simulation Only -- Not Validated Against Fabricated Devices".
func NewSimulationBanner() *fyne.Container {
	return display.NewSimulationBanner()
}
