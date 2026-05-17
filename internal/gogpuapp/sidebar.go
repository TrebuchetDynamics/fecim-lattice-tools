//go:build !cgo

package gogpuapp

import (
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// buildSidebar returns a sidebar widget listing all module descriptors.
func buildSidebar(descriptors []viewmodel.ModuleDescriptor, activeIndex int) widget.Widget {
	items := []widget.Widget{
		primitives.Text("FeCIM Modules").FontSize(14).Bold(),
	}

	for i, d := range descriptors {
		highlight := ""
		if i == activeIndex {
			highlight = " →"
		}
		items = append(items, primitives.Box(
			primitives.Text(highlight+" "+d.Title).FontSize(13),
			primitives.Text(string(d.ID)+" · "+d.Status).FontSize(10),
		).
			Padding(8).
			Gap(2),
		)
	}

	return primitives.Box(items...).
		Padding(16).
		Gap(8)
}

// buildSidebarMaterial returns a themed sidebar widget with Material 3 colors.
func buildSidebarMaterial(descriptors []viewmodel.ModuleDescriptor, activeIndex int, theme *material3.Theme) widget.Widget {
	return buildSidebarMaterialWithSelect(descriptors, activeIndex, theme, nil)
}

func buildSidebarMaterialWithSelect(descriptors []viewmodel.ModuleDescriptor, activeIndex int, theme *material3.Theme, onSelect func(viewmodel.ModuleID)) widget.Widget {
	items := []widget.Widget{
		primitives.Text("FeCIM Modules").FontSize(14).Bold(),
	}

	for i, d := range descriptors {
		descriptor := d
		variant := button.Tonal
		statusColor := theme.Colors.OnSurfaceVariant
		if d.Status == viewmodel.StatusFunctional {
			statusColor = theme.Colors.Primary
		}
		if i == activeIndex {
			variant = button.Filled
		}

		moduleButton := button.New(
			button.Text(d.Title),
			button.VariantOpt(variant),
			button.SizeOpt(button.Small),
			button.PainterOpt(material3.ButtonPainter{Theme: theme}),
			button.A11yHint("Switch to "+d.Title),
			button.OnClick(func() {
				if onSelect != nil {
					onSelect(descriptor.ID)
				}
			}),
		).MinWidth(180)

		items = append(items, primitives.Box(
			moduleButton,
			primitives.Text(string(d.ID)+" · "+d.Status).FontSize(10).Color(statusColor),
		).
			Padding(8).
			Gap(2).
			Background(theme.Colors.Surface),
		)
	}

	return primitives.Box(items...).
		Padding(16).
		Gap(8).
		Background(theme.Colors.SurfaceContainer)
}
