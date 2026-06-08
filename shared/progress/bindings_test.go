package progress

import (
	"strings"
	"testing"
	"time"

	fyneTest "fyne.io/fyne/v2/test"
)

func TestProgressBindingsTrackProgressState(t *testing.T) {
	fyneTest.NewTempApp(t)

	p := NewProgress("EDA export", 4)
	bindings := NewProgressBindings(p)

	assertStringBinding(t, bindings.Operation, "EDA export")
	assertStringBinding(t, bindings.State, StateIdle.String())
	assertFloatBinding(t, bindings.Fraction, 0)

	p.Start()
	p.UpdateWithStatus(2, "Writing SPICE", "2 of 4 artifacts")

	waitForStringBinding(t, bindings.Phase, "Writing SPICE")
	waitForStringBinding(t, bindings.Detail, "2 of 4 artifacts")
	assertFloatBinding(t, bindings.Fraction, 0.5)
	assertStringBinding(t, bindings.State, StateRunning.String())

	status := mustStringBinding(t, bindings.StatusLine)
	if !strings.Contains(status, "EDA export: Writing SPICE") {
		t.Fatalf("StatusLine = %q, want operation and phase", status)
	}

	p.Complete()
	waitForStringBinding(t, bindings.State, StateCompleted.String())
	assertFloatBinding(t, bindings.Fraction, 1)
}

func waitForStringBinding(t *testing.T, item stringBinding, want string) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := mustStringBinding(t, item); got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("binding value = %q, want %q", mustStringBinding(t, item), want)
}

func assertStringBinding(t *testing.T, item stringBinding, want string) {
	t.Helper()
	if got := mustStringBinding(t, item); got != want {
		t.Fatalf("binding value = %q, want %q", got, want)
	}
}

func mustStringBinding(t *testing.T, item stringBinding) string {
	t.Helper()
	got, err := item.Get()
	if err != nil {
		t.Fatalf("get string binding: %v", err)
	}
	return got
}

func assertFloatBinding(t *testing.T, item floatBinding, want float64) {
	t.Helper()
	got, err := item.Get()
	if err != nil {
		t.Fatalf("get float binding: %v", err)
	}
	if got != want {
		t.Fatalf("binding value = %v, want %v", got, want)
	}
}

type stringBinding interface {
	Get() (string, error)
}

type floatBinding interface {
	Get() (float64, error)
}
