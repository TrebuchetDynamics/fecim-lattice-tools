//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/help"
)

// TooltipContent holds structured tooltip information.
type TooltipContent = help.TooltipContent

var HysteresisTooltips = help.HysteresisTooltips
var CrossbarTooltips = help.CrossbarTooltips
var MNISTTooltips = help.MNISTTooltips
var CircuitsTooltips = help.CircuitsTooltips
var ComparisonTooltips = help.ComparisonTooltips
var EDATooltips = help.EDATooltips

// HoverTooltip is a widget wrapper that shows tooltip content on hover.
type HoverTooltip = help.HoverTooltip

// NewHoverTooltip creates a new hover tooltip wrapper.
func NewHoverTooltip(content fyne.CanvasObject, tooltip string, window fyne.Window) *HoverTooltip {
	return help.NewHoverTooltip(content, tooltip, window)
}

// NewHoverTooltipFromContent creates a hover tooltip from structured content.
func NewHoverTooltipFromContent(content fyne.CanvasObject, tc TooltipContent, window fyne.Window) *HoverTooltip {
	return help.NewHoverTooltipFromContent(content, tc, window)
}

// InfoButton creates a small info button for tooltip dialogs.
func InfoButton(tc TooltipContent, window fyne.Window) *widget.Button {
	return help.InfoButton(tc, window)
}

// ShowTooltipDialog shows a structured tooltip dialog.
func ShowTooltipDialog(tc TooltipContent, window fyne.Window) { help.ShowTooltipDialog(tc, window) }

// AddLabelTooltip wraps a label with tooltip help.
func AddLabelTooltip(label *widget.Label, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.AddLabelTooltip(label, tc, window)
}

// AddSliderTooltip wraps a slider with tooltip help.
func AddSliderTooltip(slider *widget.Slider, label *widget.Label, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return help.AddSliderTooltip(slider, label, tc, window)
}
