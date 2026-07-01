//go:build cgo

package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// TestFlattenLayoutNodes_DescendsIntoScrollAccordionAndSplit guards against
// the accessibility smoke test silently skipping the (very common) Scroll,
// Accordion, and Split wrapper widgets used throughout every module's GUI.
// Without descending into them, TestEmbeddedModulesLayoutAccessibilitySmoke
// never inspects whatever content they contain. The accordion item must be
// Open: a closed item's Detail is never laid out by Fyne (see
// TestFlattenLayoutNodes_SkipsClosedAccordionItems), so walking it would
// only manufacture false positives.
func TestFlattenLayoutNodes_DescendsIntoScrollAccordionAndSplit(t *testing.T) {
	scrollLabel := widget.NewLabel("inside scroll")
	scroll := container.NewScroll(scrollLabel)

	accordionLabel := widget.NewLabel("inside accordion")
	accordionItem := widget.NewAccordionItem("title", accordionLabel)
	accordionItem.Open = true
	accordion := widget.NewAccordion(accordionItem)

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

// TestFlattenLayoutNodes_SkipsClosedAccordionItems guards against a false
// positive: Fyne's accordion layout never resizes a closed item's Detail
// (widget/accordion.go only touches ai.Detail when ai.Open), so its content
// legitimately stays at zero size. Walking into it anyway makes the
// accessibility smoke test flag perfectly normal collapsed UI as broken.
func TestFlattenLayoutNodes_SkipsClosedAccordionItems(t *testing.T) {
	closedLabel := widget.NewLabel("inside closed accordion")
	accordionItem := widget.NewAccordionItem("title", closedLabel)
	accordionItem.Open = false
	accordion := widget.NewAccordion(accordionItem)

	nodes := flattenLayoutNodes(accordion)

	for _, n := range nodes {
		if label, ok := n.obj.(*widget.Label); ok && label.Text == "inside closed accordion" {
			t.Fatal("flattenLayoutNodes descended into a closed accordion item's Detail")
		}
	}
}

// TestWidgetMissingLayoutSize guards the accessibility smoke test's zero-size
// check against false positives on widgets that legitimately want zero
// space right now (e.g. an empty breadcrumb trail before any document is
// selected, whose own MinSize() is computed as zero) versus widgets that
// want space (non-zero MinSize) but were never given any due to a real
// layout bug.
func TestWidgetMissingLayoutSize(t *testing.T) {
	wantsSpace := widget.NewLabel("hello")
	wantsSpace.Resize(fyne.NewSize(0, 0))
	if !widgetMissingLayoutSize(wantsSpace) {
		t.Error("widget with non-zero MinSize but zero actual size should be reported as missing layout size")
	}

	sized := widget.NewLabel("hello")
	sized.Resize(sized.MinSize())
	if widgetMissingLayoutSize(sized) {
		t.Error("widget sized to its own MinSize should not be reported as missing layout size")
	}

	// An accordion with no items has a legitimately zero MinSize (mirrors
	// the empty breadcrumb trail / collapsed accordion case): nothing to
	// lay out, so zero actual size is correct, not a bug.
	empty := widget.NewAccordion()
	empty.Resize(fyne.NewSize(0, 0))
	if widgetMissingLayoutSize(empty) {
		t.Error("widget with zero MinSize and zero actual size should not be reported as missing layout size")
	}
}
