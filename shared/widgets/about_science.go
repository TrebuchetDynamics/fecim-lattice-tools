//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/display"
)

// ScienceSection represents a topic in the About the Science dialog.
type ScienceSection = display.ScienceSection

func aboutScienceData() []ScienceSection {
	return display.AboutScienceData()
}

// ShowAboutScience displays the unified "About the Science" dialog.
func ShowAboutScience(parent fyne.Window) {
	display.ShowAboutScience(parent)
}

// CreateAboutScienceButton creates a standardized button for the About Science dialog.
func CreateAboutScienceButton(parent fyne.Window) *widget.Button {
	return display.CreateAboutScienceButton(parent)
}
