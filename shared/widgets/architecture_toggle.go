//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/selection"
)

// ArchitectureToggleStyle controls how the toggle buttons are labeled.
type ArchitectureToggleStyle = selection.ArchitectureToggleStyle

const (
	ArchitectureToggleStylePlain  = selection.ArchitectureToggleStylePlain
	ArchitectureToggleStyleBullet = selection.ArchitectureToggleStyleBullet
)

// ArchitectureToggleOptions configures the architecture toggle buttons.
type ArchitectureToggleOptions = selection.ArchitectureToggleOptions

// ArchitectureToggle provides a shared PASSIVE/1T1R/2T1R button group.
type ArchitectureToggle = selection.ArchitectureToggle

// NewArchitectureToggle creates a new architecture toggle button group.
func NewArchitectureToggle(opts ArchitectureToggleOptions) *ArchitectureToggle {
	return selection.NewArchitectureToggle(opts)
}

// Compile-time reference retains legacy widget import expectations.
var _ *widget.Button
