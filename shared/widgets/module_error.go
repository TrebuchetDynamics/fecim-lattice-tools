//go:build legacy_fyne

package widgets

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/widgets/display"
)

// NewModuleErrorContent returns a consistent fallback view for embeddable
// modules that fail during initialization.
func NewModuleErrorContent(moduleName string, err error) fyne.CanvasObject {
	return display.NewModuleErrorContent(moduleName, err)
}
