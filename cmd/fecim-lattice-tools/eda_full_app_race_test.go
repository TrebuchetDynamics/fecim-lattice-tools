//go:build cgo

package main

import (
	"testing"
	"time"

	sharedwidgets "fecim-lattice-tools/shared/widgets"

	"fyne.io/fyne/v2"
	fyneTest "fyne.io/fyne/v2/test"
)

// TestAllModulesInitialLayoutIsStable exercises initial module layout with the
// headless driver. Unlike the desktop dispatcher, test.Driver runs fyne.Do inline;
// construction, layout, and cache resets must join the same UI lock as SafeDo.
func TestAllModulesInitialLayoutIsStable(t *testing.T) {
	t.Setenv("FECIM_DISABLE_STARTUP_CALIBRATION", "1")

	var app fyne.App
	sharedwidgets.WithUILock(func() { app = fyneTest.NewApp() })
	// NewTempApp's implicit cleanup resets global theme/font caches without the
	// shared lock. Own that reset explicitly, including after late SafeDo calls.
	t.Cleanup(func() {
		sharedwidgets.WithUILock(func() { fyneTest.NewApp() })
	})

	for _, module := range newLifecycleModulesForTest(t) {
		t.Run(module.name, func(t *testing.T) {
			var window fyne.Window
			sharedwidgets.WithUILock(func() {
				window = app.NewWindow("initial-layout-" + module.name)
				content := module.module.BuildContent(app, window)
				if content == nil {
					t.Fatal("BuildContent returned nil")
				}
				window.Resize(fyne.NewSize(1024, 768))
				window.SetContent(content)
				window.Show()
				fyne.DoAndWait(func() { window.Canvas().Refresh(content) })
			})
			defer sharedwidgets.WithUILock(window.Close)
			// Leave the lock available to real asynchronous status updates.
			time.Sleep(100 * time.Millisecond)
		})
	}
}
