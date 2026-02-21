package widgets

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestAdaptiveLayout_InitialMobileBreakpointSwitchesOnFirstLayout(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := app.NewWindow("adaptive-initial-mobile")
	defer w.Close()

	zones := []fyne.CanvasObject{
		widget.NewLabel("zone-1"),
		widget.NewLabel("zone-2"),
		widget.NewLabel("zone-3"),
	}
	adaptive := NewAdaptiveLayout(zones, []string{"One", "Two", "Three"})
	adaptive.SetDesktopLayout(func(z []fyne.CanvasObject) fyne.CanvasObject {
		return container.NewHBox(z...)
	})

	root := adaptive.Content()
	fyne.DoAndWait(func() {
		w.SetContent(container.NewMax(root))
		w.Resize(fyne.NewSize(390, 844))
		w.Show()
	})

	deadline := time.Now().Add(500 * time.Millisecond)
	for !adaptive.IsMobile() && time.Now().Before(deadline) {
		fyne.DoAndWait(func() {
			w.Canvas().Refresh(root)
		})
		time.Sleep(10 * time.Millisecond)
	}

	if !adaptive.IsMobile() {
		t.Fatalf("expected initial 390px layout to switch to mobile tabs")
	}
}

