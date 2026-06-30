//go:build cgo

package main

import (
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// TestFlattenLayoutNodes_DescendsIntoScrollAccordionAndSplit guards against
// the accessibility smoke test silently skipping the (very common) Scroll,
// Accordion, and Split wrapper widgets used throughout every module's GUI.
// Without descending into them, TestEmbeddedModulesLayoutAccessibilitySmoke
// never inspects whatever content they contain.
func TestFlattenLayoutNodes_DescendsIntoScrollAccordionAndSplit(t *testing.T) {
	scrollLabel := widget.NewLabel("inside scroll")
	scroll := container.NewScroll(scrollLabel)

	accordionLabel := widget.NewLabel("inside accordion")
	accordion := widget.NewAccordion(widget.NewAccordionItem("title", accordionLabel))

	splitLeading := widget.NewLabel("inside split leading")
	splitTrailing := widget.NewLabel("inside split trailing")
	split := container.NewHSplit(splitLeading, splitTrailing)

	root := container.NewVBox(scroll, accordion, split)

	nodes := flattenLayoutNodes(root)

	found := map[string]bool{}
	for _, n := range nodes {
		if label, ok := n.obj.(*widget.Label); ok {
			found[label.Text] = true
		}
	}

	for _, want := range []string{
		"inside scroll",
		"inside accordion",
		"inside split leading",
		"inside split trailing",
	} {
		if !found[want] {
			t.Errorf("flattenLayoutNodes did not descend to find label %q", want)
		}
	}
}
