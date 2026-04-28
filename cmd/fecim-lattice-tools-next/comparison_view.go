//go:build !cgo

package main

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
)

// buildComparisonView renders a comparison ModuleSnapshot into a gogpu/ui
// widget tree. Pure: same input → same widget tree, no side effects.
func buildComparisonView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	descriptor := snapshot.Descriptor

	children := []widget.Widget{
		primitives.Text(descriptor.Title).FontSize(20).Bold(),
		primitives.Text(descriptor.Description).FontSize(13),
		primitives.Text(string(descriptor.ID) + " | " + descriptor.Status).FontSize(11),
	}

	for _, m := range snapshot.Metrics {
		children = append(children, primitives.Text(m.Label+": "+m.Value).FontSize(12))
	}

	for _, section := range snapshot.Sections {
		children = append(children, comparisonCard(section, theme))
	}

	return primitives.Box(children...).
		Padding(20).
		Gap(12).
		Background(theme.Colors.Surface)
}

func comparisonCard(section viewmodel.Section, theme *material3.Theme) widget.Widget {
	return primitives.Box(
		primitives.Text(section.Title).FontSize(16).Bold(),
		primitives.Text(section.Body).FontSize(13),
	).
		Padding(14).
		Gap(6).
		Background(theme.Colors.SurfaceContainer)
}
