//go:build !cgo

package main

import (
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildEDAView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(14),
	}

	formatBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		if m.ID == "spice" || m.ID == "verilog" || m.ID == "liberty" || m.ID == "def" || m.ID == "lef" {
			formatBoxes = append(formatBoxes, primitives.Box(
				primitives.Text(m.Label).FontSize(12).Bold(),
				primitives.Text(m.Value).FontSize(11).Color(theme.Colors.Primary),
			).Padding(8).Gap(2).Background(theme.Colors.SurfaceContainer))
		} else {
			infoBoxes := []widget.Widget{}
			infoBoxes = append(infoBoxes, primitives.Box(
				primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
				primitives.Text(m.Value).FontSize(14).Bold(),
			).Padding(10).Gap(2).Background(theme.Colors.SurfaceContainer))
			children = append(children, primitives.Box(infoBoxes...).Gap(8))
		}
	}
	children = append(children, primitives.Box(formatBoxes...).Gap(6))

	for _, section := range snapshot.Sections {
		children = append(children, primitives.Box(
			primitives.Text(section.Title).FontSize(15).Bold(),
			primitives.Text(section.Body).FontSize(12),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer))
	}

	return primitives.Box(children...).Padding(24).Gap(14).Background(theme.Colors.Surface)
}
