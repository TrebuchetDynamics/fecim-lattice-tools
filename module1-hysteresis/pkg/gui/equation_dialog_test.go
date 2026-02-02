package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestShowFrankesteinEquationDialog_NoWindow(t *testing.T) {
	a := &App{}
	a.showFrankesteinEquationDialog()
}

func TestShowFrankesteinEquationDialog_AddsOverlay(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))
	a := &App{mainWindow: win}

	before := len(win.Canvas().Overlays().List())
	a.showFrankesteinEquationDialog()
	after := len(win.Canvas().Overlays().List())

	if after != before+1 {
		t.Fatalf("expected overlays to increase by 1, got before=%d after=%d", before, after)
	}
}
