package main

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/internal/launcher"
)

func CreateLauncherContent(onDemoSelected func(int)) fyne.CanvasObject {
	return launcher.CreateLauncherContent(onDemoSelected)
}
