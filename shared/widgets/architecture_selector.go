//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/selection"
)

// Architecture constants for crossbar array types.
const (
	Architecture1T1R = selection.Architecture1T1R
	Architecture0T1R = selection.Architecture0T1R
	Architecture2T1R = selection.Architecture2T1R
)

// ArchitectureSelector is a shared widget for selecting crossbar architecture.
type ArchitectureSelector = selection.ArchitectureSelector

// NewArchitectureSelector creates a new architecture selector widget.
func NewArchitectureSelector(onChanged func(architecture string)) *ArchitectureSelector {
	return selection.NewArchitectureSelector(onChanged)
}

// ArchitectureInfo returns educational content about the selected architecture.
func ArchitectureInfo(arch string) (title, content string) { return selection.ArchitectureInfo(arch) }

// Compile-time reference retains legacy widget import expectations.
var _ *widget.Select
