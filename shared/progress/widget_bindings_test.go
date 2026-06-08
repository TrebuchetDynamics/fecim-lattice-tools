package progress

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestProgressWidgetUsesProgressBindingsForCoreDisplay(t *testing.T) {
	fyneTest.NewTempApp(t)

	p := NewProgress("Export", 10)
	w := NewProgressWidget(p)
	bindings := w.Bindings()
	if bindings == nil {
		t.Fatal("ProgressWidget should expose the ProgressBindings it uses for display")
	}

	_ = bindings.Operation.Set("Bound export")
	_ = bindings.Phase.Set("Writing artifacts")
	_ = bindings.Detail.Set("Verilog 2/4")
	_ = bindings.Fraction.Set(0.75)

	waitForWidgetText(t, func() string { return w.titleLabel.Text }, "Bound export")
	waitForWidgetText(t, func() string { return w.phaseLabel.Text }, "Writing artifacts")
	waitForWidgetText(t, func() string { return w.detailLabel.Text }, "Verilog 2/4")
	waitForProgressValue(t, w.progressBar, 0.75)
}

func waitForWidgetText(t *testing.T, current func() string, want string) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		fyne.DoAndWait(func() {})
		if got := current(); got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("widget text = %q, want %q", current(), want)
}

func waitForProgressValue(t *testing.T, bar *widget.ProgressBar, want float64) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		fyne.DoAndWait(func() {})
		if got := bar.Value; got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("progress value = %v, want %v", bar.Value, want)
}
