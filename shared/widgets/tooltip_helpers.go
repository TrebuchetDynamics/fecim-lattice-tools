//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/help"
)

// WithInfoButton wraps content with an info button that shows tooltip on click.
func WithInfoButton(content fyne.CanvasObject, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.WithInfoButton(content, tc, window)
}

// WithInfoButtonLeft places the info button on the left side.
func WithInfoButtonLeft(content fyne.CanvasObject, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.WithInfoButtonLeft(content, tc, window)
}

// LabelWithTooltip creates a label with an adjacent info button.
func LabelWithTooltip(text string, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.LabelWithTooltip(text, tc, window)
}

// SliderWithTooltip creates a complete slider row with label and info button.
func SliderWithTooltip(labelText string, slider *widget.Slider, valueLabel *widget.Label, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.SliderWithTooltip(labelText, slider, valueLabel, tc, window)
}

// SelectWithTooltip creates a select widget with info button.
func SelectWithTooltip(label string, options []string, onChanged func(string), tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.SelectWithTooltip(label, options, onChanged, tc, window)
}

// ButtonWithTooltip creates a button with an info button next to it.
func ButtonWithTooltip(label string, onTapped func(), tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.ButtonWithTooltip(label, onTapped, tc, window)
}

// EntryWithTooltip creates an entry widget with info button.
func EntryWithTooltip(placeholder string, tc TooltipContent, window fyne.Window) (*widget.Entry, fyne.CanvasObject) {
	return help.EntryWithTooltip(placeholder, tc, window)
}

// CheckWithTooltip creates a checkbox with info button.
func CheckWithTooltip(label string, onChanged func(bool), tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.CheckWithTooltip(label, onChanged, tc, window)
}

// ShowQuickTooltip shows a brief tooltip near the cursor.
func ShowQuickTooltip(text string, pos fyne.Position, canvas fyne.Canvas) *widget.PopUp {
	return help.ShowQuickTooltip(text, pos, canvas)
}

// CreateTooltipCard creates a card-style container for a section with an overall tooltip.
func CreateTooltipCard(title string, tc TooltipContent, window fyne.Window, content ...fyne.CanvasObject) *widget.Card {
	return help.CreateTooltipCard(title, tc, window, content...)
}

// EducationalDialog shows a detailed explanation dialog with formatted content.
func EducationalDialog(title string, tc TooltipContent, window fyne.Window) {
	help.EducationalDialog(title, tc, window)
}

// SectionWithTooltip creates a collapsible section with a tooltip for the header.
type SectionWithTooltip = help.SectionWithTooltip

// NewSectionWithTooltip creates a new section with tooltip support.
func NewSectionWithTooltip(title string, tc TooltipContent, window fyne.Window, content fyne.CanvasObject) *SectionWithTooltip {
	return help.NewSectionWithTooltip(title, tc, window, content)
}

var ModuleOverviewTooltips = help.ModuleOverviewTooltips

// ShowModuleOverview displays a module overview dialog.
func ShowModuleOverview(module string, window fyne.Window) { help.ShowModuleOverview(module, window) }

// HelpButton creates a module help button.
func HelpButton(module string, window fyne.Window) *widget.Button {
	return help.HelpButton(module, window)
}

// QuickReferenceCard creates a quick reference card.
func QuickReferenceCard(tc TooltipContent) *widget.Card { return help.QuickReferenceCard(tc) }
