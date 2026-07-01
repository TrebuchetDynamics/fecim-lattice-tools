package tabs

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// TestShowRelayout_GivesPreviouslyHiddenWidgetANonZeroSize guards against a
// real layout bug: Fyne's box layouts (layout/boxlayout.go) skip Hidden
// children entirely when computing positions/sizes, so a widget hidden
// before its container's first layout pass never receives a size. Calling
// widget.Show() alone does not retrigger that layout pass — only the
// parent Container's Refresh() (which re-runs Layout()) does. Without
// showRelayout, a button shown asynchronously after the initial layout
// (e.g. the "Pull OpenLane Image" button once Docker detection completes)
// stays visible but stuck at zero size.
func TestShowRelayout_GivesPreviouslyHiddenWidgetANonZeroSize(t *testing.T) {
	btn := widget.NewButton("Pull OpenLane Image", func() {})
	btn.Hide() // mirrors pullImageBtn.Hide() called before the row is ever laid out

	row := container.NewHBox(widget.NewLabel("Docker:"), btn)
	row.Resize(row.MinSize()) // initial layout pass, btn still hidden

	if got := btn.Size(); got.Width > 0 || got.Height > 0 {
		t.Fatalf("setup invariant broken: hidden button already has size %v", got)
	}

	showRelayout(btn, row)

	if got := btn.Size(); got.Width <= 0 || got.Height <= 0 {
		t.Fatalf("button shown after initial layout still has zero size %v; showRelayout should force the parent row to relayout", got)
	}
}

// TestHideRelayout_ShrinksRowAfterHidingAWidget is the mirror case: hiding a
// previously-visible widget should make the parent row's MinSize shrink
// immediately, not wait for some unrelated future relayout.
func TestHideRelayout_ShrinksRowAfterHidingAWidget(t *testing.T) {
	btn := widget.NewButton("Pull OpenLane Image", func() {})

	row := container.NewHBox(widget.NewLabel("Docker:"), btn)
	row.Resize(row.MinSize())

	withBtn := row.MinSize()

	hideRelayout(btn, row)

	withoutBtn := row.MinSize()
	if !(withoutBtn.Width < withBtn.Width) {
		t.Fatalf("expected row MinSize width to shrink after hideRelayout, got withBtn=%v withoutBtn=%v", withBtn, withoutBtn)
	}
}

// TestShowRelayout_NilParentDoesNotPanic covers the construction-time race
// where the relayout helpers may be invoked before the enclosing row
// variable has been assigned.
func TestShowRelayout_NilParentDoesNotPanic(t *testing.T) {
	btn := widget.NewButton("x", func() {})
	btn.Hide()
	var row *fyne.Container
	showRelayout(btn, row)
	hideRelayout(btn, row)
}
