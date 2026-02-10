package gui

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
)

// ShowPhysicsEquationsDialog displays the equation modal.
//
// This is exported so automated visual tests and screenshot crawlers can
// reliably open the modal without fragile widget-tree clicking.
func (a *App) ShowPhysicsEquationsDialog() {
	a.showPhysicsEquationsDialog()
}

func (a *App) showPhysicsEquationsDialog() {
	if a.mainWindow == nil {
		return
	}

	wasPaused := a.paused
	if !wasPaused {
		a.paused = true
		if a.pauseBtn != nil {
			a.pauseBtn.SetText("Resume")
		}
	}

	content := widgets.NewPhysicsEquationsWidget(a.mainWindow)
	framed := container.NewPadded(content)

	// Responsive sizing: keep it large on desktop but never exceed the window.
	canvasSize := a.mainWindow.Canvas().Size()
	width := float32(1000)
	height := float32(620)
	if canvasSize.Width > 0 {
		width = float32(math.Min(float64(width), float64(canvasSize.Width*0.95)))
		width = float32(math.Max(float64(width), 640))
	}
	if canvasSize.Height > 0 {
		height = float32(math.Min(float64(height), float64(canvasSize.Height*0.88)))
		height = float32(math.Max(float64(height), 420))
	}

	var d dialog.Dialog
	closeBtn := widget.NewButton("Close", func() {
		if d != nil {
			d.Hide()
		}
		if !wasPaused && a.paused {
			a.paused = false
			if a.pauseBtn != nil {
				a.pauseBtn.SetText("Pause")
			}
		}
	})

	footer := container.NewHBox(layout.NewSpacer(), closeBtn)
	body := container.NewBorder(nil, footer, nil, nil, framed)

	d = dialog.NewCustom("Physics Equations", "", body, a.mainWindow)
	d.Resize(fyne.NewSize(width, height))
	d.Show()
}
