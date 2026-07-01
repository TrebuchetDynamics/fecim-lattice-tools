package tabs

import "fyne.io/fyne/v2"

// showRelayout shows obj and forces parent to re-run its layout. Fyne's box
// layouts skip Hidden children when computing positions and sizes, so a
// widget hidden before its container's first layout pass never gets a
// size; toggling Show() alone does not retrigger that layout pass, only
// Container.Refresh() does. Without this, a widget shown asynchronously
// after the initial layout stays visible but stuck at zero size.
func showRelayout(obj fyne.CanvasObject, parent *fyne.Container) {
	obj.Show()
	if parent != nil {
		parent.Refresh()
	}
}

// hideRelayout is the mirror of showRelayout for Hide().
func hideRelayout(obj fyne.CanvasObject, parent *fyne.Container) {
	obj.Hide()
	if parent != nil {
		parent.Refresh()
	}
}
