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

	content := widgets.NewFrankesteinEquationWidget(a.mainWindow)
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(640, 320))

	var dialog *widget.PopUp
	closeBtn := widget.NewButton("Close", func() {
		if dialog != nil {
			dialog.Hide()
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
