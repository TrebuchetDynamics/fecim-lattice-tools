//go:build legacy_fyne

// Package widgets provides accessibility helper functions.
package widgets

import (
	"fyne.io/fyne/v2"

	widgetsa11y "fecim-lattice-tools/shared/widgets/accessibility"
)

// AccessibleColors provides pre-computed WCAG AA compliant color pairs.
var AccessibleColors = widgetsa11y.AccessibleColors

// MinTextSize constants for accessibility compliance.
const (
	MinBodyTextSize    = widgetsa11y.MinBodyTextSize
	MinCaptionTextSize = widgetsa11y.MinCaptionTextSize
	MinHeaderTextSize  = widgetsa11y.MinHeaderTextSize
	MinLargeHeaderSize = widgetsa11y.MinLargeHeaderSize
)

// KeyboardHandler provides a standardized interface for keyboard accessibility.
type KeyboardHandler = widgetsa11y.KeyboardHandler

// StandardKeyBindings maps common keyboard shortcuts.
var StandardKeyBindings = widgetsa11y.StandardKeyBindings

// WrapWithFocus wraps a canvas object with a focus indicator.
func WrapWithFocus(content fyne.CanvasObject) *FocusIndicator {
	return widgetsa11y.WrapWithFocus(content)
}

// MakeKeyboardNavigable adds keyboard event handling to a window.
func MakeKeyboardNavigable(window fyne.Window, onHelp func()) {
	widgetsa11y.MakeKeyboardNavigable(window, onHelp)
}

// KeyboardDrawable provides keyboard drawing support for canvas widgets.
type KeyboardDrawable = widgetsa11y.KeyboardDrawable

// NewKeyboardDrawable creates a keyboard-accessible drawing handler.
func NewKeyboardDrawable(width, height int, onDraw func(x, y int)) *KeyboardDrawable {
	return widgetsa11y.NewKeyboardDrawable(width, height, onDraw)
}

// GridNavigator provides keyboard navigation for grid-based widgets.
type GridNavigator = widgetsa11y.GridNavigator

// NewGridNavigator creates a grid keyboard navigator.
func NewGridNavigator(rows, cols int, onSelect func(row, col int)) *GridNavigator {
	return widgetsa11y.NewGridNavigator(rows, cols, onSelect)
}

// EnsureMinTextSize returns the larger of the input size or minimum accessible size.
func EnsureMinTextSize(size float32) float32 { return widgetsa11y.EnsureMinTextSize(size) }

// EnsureMinCaptionSize returns the larger of the input size or minimum caption size.
func EnsureMinCaptionSize(size float32) float32 { return widgetsa11y.EnsureMinCaptionSize(size) }

// ScaleForAccessibility applies accessibility scaling to a base size.
func ScaleForAccessibility(baseSize, scale float32) float32 {
	return widgetsa11y.ScaleForAccessibility(baseSize, scale)
}
