//go:build cgo

package main

import (
	"context"
	"os"
	"os/exec"
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
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

type lifecycleModule interface {
	BuildContent(fyne.App, fyne.Window) fyne.CanvasObject
	Start()
	Stop()
}

// The headless driver executes Do inline. Make direct fyne.Do and SafeDo join
// the same re-entrant lock instead of concurrently mutating widget renderers.
// This adapter belongs only to the lifecycle fixture, not production dispatch.
type lifecycleSerialApp struct {
	fyne.App
	driver fyne.Driver
}

func (a lifecycleSerialApp) Driver() fyne.Driver { return a.driver }

type lifecycleSerialDriver struct{ fyne.Driver }

func (d lifecycleSerialDriver) DoFromGoroutine(fn func(), _ bool) {
	sharedwidgets.WithUILock(fn)
}

func TestEmbeddedModulesStartStopDoNotLeakWorkers(t *testing.T) {
	const childModuleEnv = "FECIM_TEST_LIFECYCLE_MODULE"
	selected := os.Getenv(childModuleEnv)
	if selected == "" {
		for _, name := range []string{"hysteresis", "crossbar", "mnist", "circuits", "comparison", "eda", "docs"} {
			t.Run(name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestEmbeddedModulesStartStopDoNotLeakWorkers$", "-test.timeout=40s")
				cmd.Env = append(os.Environ(), childModuleEnv+"="+name)
				output, err := cmd.CombinedOutput()
				if ctx.Err() != nil {
					t.Fatalf("lifecycle subprocess timed out: %v\n%s", ctx.Err(), output)
				}
				if err != nil {
					t.Fatalf("lifecycle subprocess failed: %v\n%s", err, output)
				}
			})
		}
		return
	}

	t.Setenv("FECIM_DISABLE_STARTUP_CALIBRATION", "1")
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	var module lifecycleModule
	var window fyne.Window
	sharedwidgets.WithUILock(func() {
		// No NewTempApp cleanup: resetting global caches can race workers allowed
		// by the existing leak tolerance. Each module owns one child process;
		// process exit contains late updates, but does not prove worker joining.
		base := fyneTest.NewApp()
		app := lifecycleSerialApp{App: base, driver: lifecycleSerialDriver{base.Driver()}}
		fyne.SetCurrentApp(app)
		app.Settings().SetTheme(&sharedtheme.FeCIMTheme{})
		switch selected {
		case "hysteresis":
			module = demo1gui.NewEmbeddedApp()
		case "crossbar":
			var err error
			module, err = demo2gui.NewEmbeddedCrossbarApp()
			if err != nil {
				t.Fatalf("create crossbar module: %v", err)
			}
		case "mnist":
			module = demo3gui.NewEmbeddedDualModeApp()
		case "circuits":
			module = demo4gui.NewEmbeddedCircuitsApp()
		case "comparison":
			module = demo5gui.NewEmbeddedComparisonApp()
		case "eda":
			module = demo6gui.NewEmbeddedEDAApp()
		case "docs":
			module = demo7gui.NewEmbeddedDocsApp()
		default:
			t.Fatalf("unknown lifecycle module %q", selected)
		}
		window = app.NewWindow("lifecycle-" + selected)
		content := module.BuildContent(app, window)
		if content == nil {
			t.Fatal("BuildContent returned nil")
		}
		window.SetContent(content)
		window.Resize(fyne.NewSize(1400, 900))
		window.Show()
	})
	defer sharedwidgets.WithUILock(window.Close)

	// Preserve the original two-worker allowance and real Start/Stop calls.
	// Stop and leak polling stay outside the UI lock: Stop may join workers
	// that themselves need to dispatch a final UI update.
	testutil.AssertNoGoroutineLeak(t, 2, func() {
		sharedwidgets.WithUILock(module.Start)
		time.Sleep(80 * time.Millisecond)
		module.Stop()
	})
}
