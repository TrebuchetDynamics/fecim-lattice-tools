//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/interaction"
)

// SafeUpdateLabel updates a label's text safely from any goroutine.
func SafeUpdateLabel(label *widget.Label, text string) { interaction.SafeUpdateLabel(label, text) }

// SafeUpdateProgress updates a progress bar's value safely from any goroutine.
func SafeUpdateProgress(progress *widget.ProgressBar, value float64) {
	interaction.SafeUpdateProgress(progress, value)
}

// SafeUpdateProgressInfinite sets a progress bar to infinite mode safely.
func SafeUpdateProgressInfinite(progress *widget.ProgressBarInfinite, start bool) {
	interaction.SafeUpdateProgressInfinite(progress, start)
}

// SafeRefresh refreshes a canvas object safely from any goroutine.
func SafeRefresh(obj fyne.CanvasObject) { interaction.SafeRefresh(obj) }

// SafeShow shows a canvas object safely from any goroutine.
func SafeShow(obj fyne.CanvasObject) { interaction.SafeShow(obj) }

// SafeHide hides a canvas object safely from any goroutine.
func SafeHide(obj fyne.CanvasObject) { interaction.SafeHide(obj) }

// SafeShowHide shows or hides a canvas object safely from any goroutine.
func SafeShowHide(obj fyne.CanvasObject, show bool) { interaction.SafeShowHide(obj, show) }

// SafeEnable enables a disableable widget safely from any goroutine.
func SafeEnable(w fyne.Disableable) { interaction.SafeEnable(w) }

// SafeDisable disables a disableable widget safely from any goroutine.
func SafeDisable(w fyne.Disableable) { interaction.SafeDisable(w) }

// SafeEnableDisable enables or disables a widget safely from any goroutine.
func SafeEnableDisable(w fyne.Disableable, enable bool) { interaction.SafeEnableDisable(w, enable) }

// SafeSetEntry sets an entry's text safely from any goroutine.
func SafeSetEntry(entry *widget.Entry, text string) { interaction.SafeSetEntry(entry, text) }

// SafeSetCheck sets a check widget's checked state safely from any goroutine.
func SafeSetCheck(check *widget.Check, checked bool) { interaction.SafeSetCheck(check, checked) }

// SafeSetSlider sets a slider's value safely from any goroutine.
func SafeSetSlider(slider *widget.Slider, value float64) { interaction.SafeSetSlider(slider, value) }

// SafeSetSelect sets a select widget's selected value safely from any goroutine.
func SafeSetSelect(sel *widget.Select, selected string) { interaction.SafeSetSelect(sel, selected) }

// SafeSetSelectIndex sets a select widget's selected index safely from any goroutine.
func SafeSetSelectIndex(sel *widget.Select, index int) { interaction.SafeSetSelectIndex(sel, index) }
