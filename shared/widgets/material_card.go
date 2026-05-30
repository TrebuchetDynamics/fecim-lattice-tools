//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/shared/widgets/materials"
)

// MaterialCard displays a compact summary of a material for list/grid display.
type MaterialCard = materials.MaterialCard

// NewMaterialCard creates a new material card widget.
func NewMaterialCard(materialID string, material *physics.Material, onTapped func(string)) *MaterialCard {
	return materials.NewMaterialCard(materialID, material, onTapped)
}

// Compile-time reference retains legacy Fyne import expectations.
var _ fyne.CanvasObject
