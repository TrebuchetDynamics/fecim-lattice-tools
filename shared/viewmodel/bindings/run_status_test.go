package bindings

import (
	"strings"
	"testing"
	"time"

	fyneTest "fyne.io/fyne/v2/test"
)

func TestRunStatusBindingsTrackLifecycle(t *testing.T) {
	fyneTest.NewTempApp(t)

	status := NewRunStatus("EDA export")

	assertStringBinding(t, status.Operation, "EDA export")
	assertStringBinding(t, status.State, StateIdle.String())
	assertBoolBinding(t, status.Running, false)
	assertBoolBinding(t, status.Terminal, false)
	assertFloatBinding(t, status.Progress, 0)

	status.Start("Generating", "preparing netlists")
	waitForStringBinding(t, status.Phase, "Generating")
	assertStringBinding(t, status.Detail, "preparing netlists")
	assertStringBinding(t, status.State, StateRunning.String())
	assertBoolBinding(t, status.Running, true)
	assertBoolBinding(t, status.Terminal, false)

	status.Update("Writing", "2 of 4 artifacts", 0.5)
	waitForStringBinding(t, status.Phase, "Writing")
	assertStringBinding(t, status.Detail, "2 of 4 artifacts")
	assertFloatBinding(t, status.Progress, 0.5)
	if got := mustStringBinding(t, status.StatusLine); !strings.Contains(got, "EDA export: Writing — 2 of 4 artifacts") {
		t.Fatalf("StatusLine = %q, want operation, phase, and detail", got)
	}

	status.Complete("4 artifacts exported")
	waitForStringBinding(t, status.State, StateCompleted.String())
	assertStringBinding(t, status.Detail, "4 artifacts exported")
	assertBoolBinding(t, status.Running, false)
	assertBoolBinding(t, status.Terminal, true)
	assertFloatBinding(t, status.Progress, 1)
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

func assertBoolBinding(t *testing.T, item boolBinding, want bool) {
	t.Helper()
	got, err := item.Get()
	if err != nil {
		t.Fatalf("get bool binding: %v", err)
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

type boolBinding interface {
	Get() (bool, error)
}
