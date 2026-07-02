//go:build cgo

package main

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fyneTest "fyne.io/fyne/v2/test"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
	"fecim-lattice-tools/shared/testutil"
	sharedtheme "fecim-lattice-tools/shared/theme"
)

type lifecycleModule interface {
	BuildContent(fyne.App, fyne.Window) fyne.CanvasObject
	Start()
	Stop()
}

func TestEmbeddedModulesStartStopDoNotLeakWorkers(t *testing.T) {
	t.Setenv("FECIM_DISABLE_STARTUP_CALIBRATION", "1")

	crossbar, err := demo2gui.NewEmbeddedCrossbarApp()
	if err != nil {
		t.Fatalf("create crossbar module: %v", err)
	}

	modules := []struct {
		name     string
		module   lifecycleModule
		maxExtra int
	}{
		{name: "hysteresis", module: demo1gui.NewEmbeddedApp(), maxExtra: 2},
		{name: "crossbar", module: crossbar, maxExtra: 2},
		{name: "mnist", module: demo3gui.NewEmbeddedDualModeApp(), maxExtra: 2},
		{name: "circuits", module: demo4gui.NewEmbeddedCircuitsApp(), maxExtra: 2},
		{name: "comparison", module: demo5gui.NewEmbeddedComparisonApp(), maxExtra: 2},
		{name: "eda", module: demo6gui.NewEmbeddedEDAApp(), maxExtra: 2},
		{name: "docs", module: demo7gui.NewEmbeddedDocsApp(), maxExtra: 2},
	}

	for _, tc := range modules {
		t.Run(tc.name, func(t *testing.T) {
			app := fyneTest.NewTempApp(t)
			app.Settings().SetTheme(&sharedtheme.FeCIMTheme{})
			window := app.NewWindow("lifecycle-" + tc.name)
			defer window.Close()

			content := tc.module.BuildContent(app, window)
			if content == nil {
				t.Fatal("BuildContent returned nil")
			}
			window.SetContent(content)
			window.Resize(fyne.NewSize(1400, 900))
			window.Show()

			testutil.AssertNoGoroutineLeak(t, tc.maxExtra, func() {
				tc.module.Start()
				time.Sleep(80 * time.Millisecond)
				tc.module.Stop()
			})
		})
	}
}
