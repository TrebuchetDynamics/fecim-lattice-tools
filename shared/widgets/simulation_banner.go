// Package widgets provides shared UI components for FeCIM visualizers.
//
// simulation_banner.go provides a thin amber banner reminding users that
// all results are simulation-only and not validated against fabricated devices.
package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// NewSimulationBanner returns a thin horizontal bar with an amber background
// and white text reading "Simulation Only -- Not Validated Against Fabricated Devices".
func NewSimulationBanner() *fyne.Container {
	bg := canvas.NewRectangle(color.RGBA{210, 150, 30, 255})
	bg.SetMinSize(fyne.NewSize(0, 22))
	text := canvas.NewText("Simulation Only \u2014 Not Validated Against Fabricated Devices", color.White)
	text.TextSize = 11
	text.Alignment = fyne.TextAlignCenter
	return container.NewStack(bg, container.NewCenter(text))
}
