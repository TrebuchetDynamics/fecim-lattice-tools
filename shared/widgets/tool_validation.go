//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/validationui"
)

// ToolStatusLabelMode controls how the tool status labels are rendered.
type ToolStatusLabelMode = validationui.ToolStatusLabelMode

const (
	ToolStatusSymbolOnly     = validationui.ToolStatusSymbolOnly
	ToolStatusSymbolWithName = validationui.ToolStatusSymbolWithName
)

// ToolValidationMessageStyle controls result text formatting.
type ToolValidationMessageStyle = validationui.ToolValidationMessageStyle

const (
	ToolMessageUnicode     = validationui.ToolMessageUnicode
	ToolMessageASCII       = validationui.ToolMessageASCII
	ToolMessageUnicodeSkip = validationui.ToolMessageUnicodeSkip
)

// ToolValidationOptions configures the shared tool validation UI.
type ToolValidationOptions = validationui.ToolValidationOptions

// ToolValidationWidgets bundles the status labels and button for tool validation.
type ToolValidationWidgets = validationui.ToolValidationWidgets

// NewToolValidationWidgets creates status labels and a validation button.
func NewToolValidationWidgets(opts ToolValidationOptions) *ToolValidationWidgets {
	return validationui.NewToolValidationWidgets(opts)
}

// Compile-time references retain legacy imports expected by old docs/examples.
var _ fyne.Window
var _ *widget.Button
