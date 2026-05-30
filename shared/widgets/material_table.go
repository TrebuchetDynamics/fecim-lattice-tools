//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/shared/widgets/materials"
)

// MaterialTable displays all properties of a single material in a categorized tabbed view.
type MaterialTable = materials.MaterialTable

// NewMaterialTable creates a new material table widget.
func NewMaterialTable(material *physics.Material) *MaterialTable {
	return materials.NewMaterialTable(material)
}

// MaterialDetailPanel shows detailed information about a material.
type MaterialDetailPanel = materials.MaterialDetailPanel

// NewMaterialDetailPanel creates a new detailed material panel.
func NewMaterialDetailPanel(material *physics.Material) *MaterialDetailPanel {
	return materials.NewMaterialDetailPanel(material)
}

// Compile-time reference retains legacy Fyne import expectations.
var _ fyne.CanvasObject
