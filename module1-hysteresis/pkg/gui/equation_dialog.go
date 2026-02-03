package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
)

func (a *App) showFrankesteinEquationDialog() {
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

	content := widgets.NewFrankesteinEquationWidget(a.mainWindow)
	scroll := container.NewScroll(content)
	scroll.Direction = container.ScrollBoth
	canvasSize := a.mainWindow.Canvas().Size()
	width := canvasSize.Width * 0.9
	height := canvasSize.Height * 0.6
	if width <= 0 {
		width = 640
	}
	if height <= 0 {
		height = 360
	}
	scroll.SetMinSize(fyne.NewSize(width, height))

	var dialog *widget.PopUp
	closeBtn := widget.NewButton("Close", func() {
		if dialog != nil {
			dialog.Hide()
		}
		if !wasPaused && a.paused {
			a.paused = false
			if a.pauseBtn != nil {
				a.pauseBtn.SetText("Pause")
			}
		}
	})

	dialog = widget.NewModalPopUp(
		container.NewVBox(
			container.NewPadded(scroll),
			closeBtn,
		),
		a.mainWindow.Canvas(),
	)

	dialog.Show()
}
