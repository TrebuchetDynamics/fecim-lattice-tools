// cmd/eda-gui/main.go
package main

import (
	"fyne.io/fyne/v2/app"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui"
)

func main() {
	a := app.New()
	w := gui.CreateMainWindow(a)
	w.ShowAndRun()
}
