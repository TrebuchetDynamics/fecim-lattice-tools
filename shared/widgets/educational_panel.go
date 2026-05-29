//go:build legacy_fyne

// Package widgets provides shared widget utilities for Fyne GUI development.
package widgets

import (
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/display"
)

// EducationalPanel is a reusable widget for displaying educational content.
type EducationalPanel = display.EducationalPanel

// EducationalPanelConfig holds configuration for creating an EducationalPanel.
type EducationalPanelConfig = display.EducationalPanelConfig

// EducationalSection represents a collapsible section for educational content.
type EducationalSection = display.EducationalSection

// NewEducationalPanel creates a new educational panel widget.
func NewEducationalPanel(config EducationalPanelConfig) *EducationalPanel {
	return display.NewEducationalPanel(config)
}

// CreateEducationalAccordion creates an accordion widget from educational sections.
func CreateEducationalAccordion(sections []EducationalSection) *widget.Accordion {
	return display.CreateEducationalAccordion(sections)
}
