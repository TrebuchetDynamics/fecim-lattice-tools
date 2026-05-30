//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fecim-lattice-tools/shared/validation"
	"fecim-lattice-tools/shared/widgets/validationui"
)

// ToolStatusText formats a tool status label.
func ToolStatusText(status validation.ToolStatus, name string, mode ToolStatusLabelMode) string {
	return validationui.ToolStatusText(status, name, mode)
}

// ToolBusyText formats the temporary "busy" status label.
func ToolBusyText(name string, mode ToolStatusLabelMode) string {
	return validationui.ToolBusyText(name, mode)
}

// FormatToolValidationResult formats a validation result into a short status message.
func FormatToolValidationResult(result *validation.ValidationResult, style ToolValidationMessageStyle) string {
	return validationui.FormatToolValidationResult(result, style)
}
