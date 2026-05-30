//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/widgets/help"
)

// GlossaryEntry represents a single glossary term.
type GlossaryEntry = help.GlossaryEntry

// TermsData contains all technical terms used across modules.
var TermsData = help.TermsData

// GlossaryWidget displays searchable glossary with expandable terms.
type GlossaryWidget = help.GlossaryWidget

// NewGlossaryWidget creates a new glossary widget.
func NewGlossaryWidget() *GlossaryWidget { return help.NewGlossaryWidget() }

// ShowGlossary displays a popup dialog with term definition.
func ShowGlossary(term string, parent fyne.Window) { help.ShowGlossary(term, parent) }

// ShowFullGlossary displays the complete searchable glossary.
func ShowFullGlossary(parent fyne.Window) { help.ShowFullGlossary(parent) }

// ReferenceEntry represents a literature reference.
type ReferenceEntry = help.ReferenceEntry

// ReferencesData contains project references.
var ReferencesData = help.ReferencesData

// ReferencesWidget displays project references.
type ReferencesWidget = help.ReferencesWidget

// NewReferencesWidget creates a references widget.
func NewReferencesWidget() *ReferencesWidget { return help.NewReferencesWidget() }

// ShowReferences displays references in a dialog.
func ShowReferences(parent fyne.Window) { help.ShowReferences(parent) }

// CreateHelpMenuItems creates glossary/reference help menu items.
func CreateHelpMenuItems(parent fyne.Window) []*fyne.MenuItem {
	return help.CreateHelpMenuItems(parent)
}

// QuickTermLookup returns the definition for a term, case-insensitive.
func QuickTermLookup(term string) string { return help.QuickTermLookup(term) }

// GetTermsByCategory returns glossary terms in a category.
func GetTermsByCategory(category string) []GlossaryEntry { return help.GetTermsByCategory(category) }

// GetCategories returns sorted glossary categories.
func GetCategories() []string { return help.GetCategories() }

// CreateGlossaryButton creates a button that opens the glossary.
func CreateGlossaryButton(parent fyne.Window) *widget.Button {
	return help.CreateGlossaryButton(parent)
}

// CreateReferencesButton creates a button that opens references.
func CreateReferencesButton(parent fyne.Window) *widget.Button {
	return help.CreateReferencesButton(parent)
}
