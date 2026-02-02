package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestNewFrankesteinEquationWidget(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))

	obj := NewFrankesteinEquationWidget(win)
	if obj == nil {
		t.Fatal("NewFrankesteinEquationWidget should return non-nil")
	}

	container, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatalf("expected *fyne.Container, got %T", obj)
	}

	if len(container.Objects) != 6 {
		t.Errorf("expected 6 top-level objects, got %d", len(container.Objects))
	}
}

func TestTermChipTooltipLifecycle(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))

	chip := NewTermChip(win, "\\rho_{eff}", "Effective viscosity tooltip")
	event := &desktop.MouseEvent{
		PointEvent: fyne.PointEvent{
			AbsolutePosition: fyne.NewPos(10, 10),
		},
	}

	chip.MouseIn(event)
	if chip.tooltipPopup == nil {
		t.Fatal("expected tooltip popup to be created on hover")
	}

	chip.MouseOut()
	if chip.tooltipPopup != nil {
		t.Fatal("expected tooltip popup to be cleared on mouse out")
	}
}

func TestTermChipNoTooltip(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))

	chip := NewTermChip(win, "\\alpha", "")
	event := &desktop.MouseEvent{
		PointEvent: fyne.PointEvent{
			AbsolutePosition: fyne.NewPos(5, 5),
		},
	}

	chip.MouseIn(event)
	if chip.tooltipPopup != nil {
		t.Fatal("expected no tooltip popup when tooltip text is empty")
	}
}
