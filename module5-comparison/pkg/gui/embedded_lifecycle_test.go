package gui

import (
	"testing"
	"time"

	fyneTest "fyne.io/fyne/v2/test"

	"fecim-lattice-tools/shared/testutil"
)

func TestEmbeddedComparisonStartStopDoesNotLeakAnimationLoop(t *testing.T) {
	app := fyneTest.NewApp()
	defer app.Quit()
	window := app.NewWindow("m5-lifecycle")
	defer window.Close()

	embedded := NewEmbeddedComparisonApp()
	content := embedded.BuildContent(app, window)
	window.SetContent(content)

	testutil.AssertNoGoroutineLeak(t, 1, func() {
		embedded.Start()
		time.Sleep(80 * time.Millisecond)
		embedded.Stop()
	})
	if embedded.running {
		t.Fatal("embedded comparison app still marked running after Stop")
	}
}
