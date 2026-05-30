//go:build legacy_fyne

package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"fecim-lattice-tools/shared/widgets/interaction"
)

// LazyTabItem defers heavy content construction until the tab is first selected.
type LazyTabItem = interaction.LazyTabItem

// NewLazyTabItem creates a lazy tab item.
func NewLazyTabItem(title string, builder func() fyne.CanvasObject) *LazyTabItem {
	return interaction.NewLazyTabItem(title, builder)
}

// ApplyToTabs wires lazy initialization to tab selection changes.
func ApplyToTabs(tabs *container.AppTabs, lazyItems map[string]*LazyTabItem) {
	interaction.ApplyToTabs(tabs, lazyItems)
}
