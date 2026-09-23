//go:build cgo

package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestParseSweepValues(t *testing.T) {
	got, err := parseSweepValues(" 8, 16,30 ")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != 8 || got[1] != 16 || got[2] != 30 {
		t.Fatalf("parsed values = %v", got)
	}
	for _, raw := range []string{"", "1,,2", "1.5,2", "x,2"} {
		if _, err := parseSweepValues(raw); err == nil {
			t.Errorf("parseSweepValues(%q) accepted malformed input", raw)
		}
	}
}

func TestResearchWorkbenchContentRenders(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Research workbench test")
	content := CreateResearchWorkbenchContent(window)
	if content == nil || content.MinSize().Width <= 0 || content.MinSize().Height <= 0 {
		t.Fatalf("invalid workbench content: %v", content)
	}
	window.SetContent(content)
	window.Show()
	labels := make(map[string]bool)
	collectWorkbenchButtons(content, labels)
	for _, required := range []string{"Create starter project", "Configure starter sweep…", "Validate project", "Simulate sweep", "Analyze · generate reports", "Validate · run integrity/provenance", "Export selected run · SPICE/Verilog/DEF/LEF/Liberty", "Cancel simulation"} {
		if !labels[required] {
			t.Errorf("workbench is missing action %q", required)
		}
	}
}

func collectWorkbenchButtons(object fyne.CanvasObject, labels map[string]bool) {
	if button, ok := object.(*widget.Button); ok {
		labels[button.Text] = true
	}
	if children, ok := object.(*fyne.Container); ok {
		for _, child := range children.Objects {
			collectWorkbenchButtons(child, labels)
		}
	}
}
