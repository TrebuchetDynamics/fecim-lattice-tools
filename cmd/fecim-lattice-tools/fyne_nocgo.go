//go:build !cgo

package main

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

var (
	screenshotOutputDir = "screenshots"
	recordingOutputDir  = "recordings"
)

func runFyneApp(module viewmodel.ModuleID) error {
	if module == "" {
		return fmt.Errorf("Fyne desktop UI requires CGO; rebuild with CGO_ENABLED=1 or use a headless subcommand")
	}
	return fmt.Errorf("Fyne desktop UI for module %s requires CGO; rebuild with CGO_ENABLED=1 or use a headless subcommand", module)
}
