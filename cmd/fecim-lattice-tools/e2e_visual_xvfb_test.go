//go:build !ci
// +build !ci

package main

import (
	"image"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
)

// These tests intentionally use a real fyne app/driver (not test.NewApp()).
// They are meant to be run under Xvfb (e.g. xvfb-run -a go test ... -run VisualXvfb)
// to capture screenshots for modules that don't work with test.NewApp().

func TestVisualXvfbRegressionCrossbar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	module, err := demo2gui.NewEmbeddedCrossbarApp()
	if err != nil {
		t.Fatalf("Failed to create crossbar module: %v", err)
	}

	img := captureWithRealApp(t, module, fyne.NewSize(1200, 800), 1200*time.Millisecond)
	savePath := saveTestScreenshot(t, img, "crossbar_initial")
	t.Logf("Crossbar screenshot saved: %s", savePath)
	verifyImageNotEmpty(t, img, "crossbar")
}

func TestVisualXvfbRegressionMNIST(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	module := demo3gui.NewEmbeddedDualModeApp()

	img := captureWithRealApp(t, module, fyne.NewSize(1200, 800), 1500*time.Millisecond)
	savePath := saveTestScreenshot(t, img, "mnist_initial")
	t.Logf("MNIST screenshot saved: %s", savePath)
	verifyImageNotEmpty(t, img, "mnist")
}

func captureWithRealApp(t *testing.T, module moduleLifecycle, size fyne.Size, settle time.Duration) image.Image {
	t.Helper()

	a := app.New()

	imgCh := make(chan image.Image, 1)

	// All UI work must happen on the Fyne call thread. With the GLFW driver this
	// effectively means: do UI setup inside the lifecycle callbacks that run after
	// the app event loop has started.
	a.Lifecycle().SetOnStarted(func() {
		w := a.NewWindow("VisualXvfb")
		w.Resize(size)

		content := module.BuildContent(a, w)
		w.SetContent(container.NewMax(content))
		w.Show()

		module.Start()

		// Capture after settle, then stop/close/quit.
		time.AfterFunc(settle, func() {
			fyne.Do(func() {
				imgCh <- w.Canvas().Capture()
				module.Stop()
				w.Close()
				a.Quit()
			})
		})
	})

	// NOTE: GLFW requires Run() be called from the main goroutine.
	a.Run()

	select {
	case img := <-imgCh:
		return img
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for screenshot capture")
		return nil
	}
}
