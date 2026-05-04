//go:build !cgo

package main

import (
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildCircuitsView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(14),
	}

	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		status := theme.Colors.Primary
		if m.ID == "ispp" && m.Value == "false" {
			status = theme.Colors.OnSurfaceVariant
		}
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(14).Bold().Color(status),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer))
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))

	for _, section := range snapshot.Sections {
		children = append(children, primitives.Box(
			primitives.Text(section.Title).FontSize(15).Bold(),
			primitives.Text(section.Body).FontSize(12),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer))
	}

	return primitives.Box(children...).Padding(24).Gap(14).Background(theme.Colors.Surface)
}
