package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectsBareWidgetMutationInsideGoroutine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.go")
	src := `package gui
import "fyne.io/fyne/v2/widget"
func bad(label *widget.Label) {
	go func() {
		label.SetText("done")
	}()
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected one violation, got %d: %#v", len(violations), violations)
	}
	if !strings.Contains(violations[0].Message, "label.SetText") {
		t.Fatalf("expected SetText violation, got %#v", violations[0])
	}
}

func TestDetectsBareWidgetMutationInsideNamedGoroutineTarget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.go")
	src := `package gui
import "fyne.io/fyne/v2/widget"
type App struct {
	label *widget.Label
}
func (a *App) Start() {
	go a.updateLoop()
}
func (a *App) updateLoop() {
	a.label.SetText("done")
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected one violation from the go a.updateLoop() target, got %d: %#v", len(violations), violations)
	}
	if !strings.Contains(violations[0].Message, "a.label.SetText") {
		t.Fatalf("expected SetText violation, got %#v", violations[0])
	}
}

// TestNamedGoroutineTargetResolvesReceiverType guards against a precision
// bug in the named-goroutine-target resolution: this codebase has several
// unrelated types with same-named methods (e.g. App.run and
// HysteresisDataLogger.run). Matching goroutine targets by method name
// alone, without considering the receiver's static type, would falsely
// attribute a synchronously-called method's body to an unrelated type's
// actual `go x.method()` call, flagging safe main-thread-only code as a
// thread-safety violation.
func TestNamedGoroutineTargetResolvesReceiverType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "typed.go")
	src := `package gui
import "fyne.io/fyne/v2/widget"

type App struct {
	label *widget.Label
}

func (a *App) Start() {
	// Called synchronously, not via go - must not be treated as a
	// goroutine target just because Logger has a same-named method.
	a.sameName()
}

func (a *App) sameName() {
	a.label.SetText("safe: called directly on the main thread")
}

type Logger struct {
	label *widget.Label
}

func NewLogger() *Logger {
	l := &Logger{}
	go l.sameName()
	return l
}

func (l *Logger) sameName() {
	l.label.SetText("unsafe: actually spawned via goroutine")
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected exactly one violation (Logger.sameName only), got %d: %#v", len(violations), violations)
	}
	if !strings.Contains(violations[0].Message, "l.label.SetText") {
		t.Fatalf("expected violation attributed to Logger's l.label.SetText, got %#v", violations[0])
	}
}

func TestAllowsWidgetMutationInsideFyneDo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "good.go")
	src := `package gui
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)
func good(label *widget.Label) {
	go func() {
		fyne.Do(func() { label.SetText("done") })
	}()
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %#v", violations)
	}
}

func TestAllowsWidgetMutationInsideSharedSafeDo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "safe_do.go")
	src := `package gui
import (
	"fyne.io/fyne/v2/widget"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)
func good(label *widget.Label) {
	go func() {
		sharedwidgets.SafeDo(func() {
			label.Enable()
			label.SetText("done")
		})
	}()
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %#v", violations)
	}
}

func TestAllowsKnownSelfProtectingWidgetWrapperFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wrappers.go")
	src := `package gui
type OperationLog struct{}
func (o *OperationLog) Add(string) {}
type KeyStat struct{}
func (k *KeyStat) SetValue(string) {}
type App struct {
	operationLog *OperationLog
	keyStat *KeyStat
}
func (a *App) Start() { go a.run() }
func (a *App) run() {
	a.operationLog.Add("safe wrapper")
	a.keyStat.SetValue("42")
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %#v", violations)
	}
}
