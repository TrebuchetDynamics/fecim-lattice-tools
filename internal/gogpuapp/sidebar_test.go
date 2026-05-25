//go:build !cgo

package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
	uiwidget "github.com/gogpu/ui/widget"
)

func TestSidebarBuildsForAllModules(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	w := buildSidebar(descriptors, 0)
	if w == nil {
		t.Fatal("buildSidebar returned nil")
	}
}

func TestSidebarActiveIndex(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	w := buildSidebar(descriptors, 2)
	if w == nil {
		t.Fatal("buildSidebar with activeIndex=2 returned nil")
	}
}

func TestSidebarMaterialBuildsForAllModules(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	theme := material3.New(uiwidget.Hex(0x2F5D50))
	w := buildSidebarMaterial(descriptors, 0, theme)
	if w == nil {
		t.Fatal("buildSidebarMaterial returned nil")
	}
}

func TestSidebarMaterialSwitchButtonsInvokeSelection(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	theme := material3.New(uiwidget.Hex(0x2F5D50))
	var selected viewmodel.ModuleID

	w := buildSidebarMaterialWithSelect(descriptors, 0, theme, func(id viewmodel.ModuleID) {
		selected = id
	})
	buttons := collectSidebarButtons(w)
	if len(buttons) != len(descriptors) {
		t.Fatalf("sidebar button count = %d, want %d", len(buttons), len(descriptors))
	}

	want := descriptors[2].ID
	clickButton(buttons[2])
	if selected != want {
		t.Fatalf("selected module = %q, want %q", selected, want)
	}
}

func collectSidebarButtons(w uiwidget.Widget) []*button.Widget {
	var buttons []*button.Widget
	if b, ok := w.(*button.Widget); ok {
		buttons = append(buttons, b)
	}
	for _, child := range w.Children() {
		buttons = append(buttons, collectSidebarButtons(child)...)
	}
	return buttons
}

func clickButton(btn *button.Widget) {
	btn.SetBounds(geometry.NewRect(0, 0, 220, 40))
	ctx := uiwidget.NewContext()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, geometry.Pt(40, 20), geometry.Pt(40, 20), event.ModNone)
	btn.Event(ctx, press)
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, geometry.Pt(40, 20), geometry.Pt(40, 20), event.ModNone)
	btn.Event(ctx, release)
}
