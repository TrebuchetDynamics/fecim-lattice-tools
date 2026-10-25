//go:build cgo

package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/study"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestParseSweepValues(t *testing.T) {
	got, err := parseSweepValues(" 8, 16,30 ")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != 8 || got[1] != 16 || got[2] != 30 {
		t.Fatalf("parsed values = %v", got)
	}
	for _, raw := range []string{"", "1,,2", "1.5,2", "x,2"} {
		if _, err := parseSweepValues(raw); err == nil {
			t.Errorf("parseSweepValues(%q) accepted malformed input", raw)
		}
	}
}

func TestResearchWorkbenchContentRenders(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Research workbench test")
	content := CreateResearchWorkbenchContent(window)
	if content == nil || content.MinSize().Width <= 0 || content.MinSize().Height <= 0 {
		t.Fatalf("invalid workbench content: %v", content)
	}
	window.SetContent(content)
	window.Show()
	labels := make(map[string]bool)
	collectWorkbenchButtons(content, labels)
	for _, required := range []string{"Create starter project", "Configure starter sweep…", "Validate project", "Simulate sweep", "Analyze · generate reports", "Validate · run integrity/provenance", "Export selected run · SPICE/Verilog/DEF/LEF/Liberty", "Cancel simulation"} {
		if !labels[required] {
			t.Errorf("workbench is missing action %q", required)
		}
	}
}

// Queue UI dispatches so the test applies terminal updates on its own thread.
// The real study operation remains blocked until the test releases it.
type workbenchLifecycleApp struct {
	fyne.App
	driver fyne.Driver
}

func (a workbenchLifecycleApp) Driver() fyne.Driver { return a.driver }

type workbenchLifecycleDriver struct {
	fyne.Driver
	updates chan func()
}

func (d workbenchLifecycleDriver) DoFromGoroutine(fn func(), _ bool) { d.updates <- fn }

func TestWorkbenchCancellationLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name, terminal string
		err            error
		results        study.OutcomeResultsState
	}{
		{"cancelled partial", "Simulation cancelled · partial", context.Canceled, study.ResultsPartial},
		{"failure despite cancelled context", "Simulation failed", errors.New("report unavailable"), study.ResultsComplete},
		{"late cancel after success", "Simulation complete", nil, study.ResultsComplete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := study.CreateStarterProject(root, study.DefaultStarterSweepOptions()); err != nil {
				t.Fatal(err)
			}
			app := test.NewApp()
			defer app.Quit()
			updates := make(chan func(), 1)
			fyne.SetCurrentApp(workbenchLifecycleApp{App: app, driver: workbenchLifecycleDriver{Driver: app.Driver(), updates: updates}})
			window := app.NewWindow("Cancellation lifecycle")
			started := make(chan context.Context, 1)
			release := make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			defer unblock()
			content := createResearchWorkbenchContent(window, func(ctx context.Context, gotRoot string, opts study.RunOptions) (study.RunOutcome, error) {
				if gotRoot != root || opts.Workers != 1 {
					return study.RunOutcome{}, errors.New("wrong study arguments")
				}
				started <- ctx
				<-release
				return study.RunOutcome{
					ResultsState: tc.results, AnalysisAvailable: true,
					Summary:  experiment.Summary{Runs: []experiment.RunRecord{{Result: experiment.Result{Status: experiment.StatusSuccess}, Reused: true}}},
					Analysis: experiment.Analysis{Counts: map[string]int{"failed": 0, "feasible": 1, "pareto": 1}},
				}, tc.err
			})
			window.SetContent(content)
			buttons := map[string]*widget.Button{}
			var path, output *widget.Entry
			var status *widget.Label
			var inspect func(fyne.CanvasObject)
			inspect = func(obj fyne.CanvasObject) {
				switch obj := obj.(type) {
				case *widget.Button:
					buttons[obj.Text] = obj
				case *widget.Entry:
					if obj.MultiLine {
						output = obj
					} else {
						path = obj
					}
				case *widget.Label:
					if strings.HasPrefix(obj.Text, "Local projects only") {
						status = obj
					}
				case *fyne.Container:
					for _, child := range obj.Objects {
						inspect(child)
					}
				}
			}
			inspect(content)
			path.SetText(root)
			simulate, cancel := buttons["Simulate sweep"], buttons["Cancel simulation"]
			test.Tap(simulate)
			var ctx context.Context
			select {
			case ctx = <-started:
			case <-time.After(10 * time.Second):
				t.Fatal("study did not start")
			}
			test.Tap(cancel)
			if ctx.Err() != context.Canceled || !strings.HasPrefix(status.Text, "Cancel requested") || !simulate.Disabled() || !cancel.Disabled() {
				t.Fatalf("cancellation must remain requested with work disabled: %q", status.Text)
			}
			// Even a queued callback must not start a second task while cancelling.
			simulate.OnTapped()
			buttons["Validate project"].OnTapped()
			if !strings.HasPrefix(status.Text, "Cancel requested") {
				t.Fatalf("busy guard allowed duplicate work: %q", status.Text)
			}
			select {
			case <-updates:
				t.Fatal("terminal UI update before study returned")
			default:
			}
			unblock()
			select {
			case update := <-updates:
				if !strings.HasPrefix(status.Text, "Cancel requested") || !simulate.Disabled() || !cancel.Disabled() {
					t.Fatalf("controls/status changed before terminal UI dispatch: %q", status.Text)
				}
				update()
			case <-time.After(10 * time.Second):
				t.Fatal("study did not publish terminal state")
			}
			if !strings.HasPrefix(status.Text, tc.terminal) || !strings.Contains(output.Text, "total=1 reused=1 failed=0 feasible=1 pareto=1") || simulate.Disabled() || !cancel.Disabled() {
				t.Fatalf("terminal state/counts/controls: status=%q details=%q", status.Text, output.Text)
			}
			if strings.Contains(output.Text, "Partial results") != (tc.results == study.ResultsPartial) || (tc.err != nil && !strings.Contains(output.Text, tc.err.Error())) {
				t.Fatalf("terminal details lost partial qualification or error: %q", output.Text)
			}
		})
	}
}

func collectWorkbenchButtons(object fyne.CanvasObject, labels map[string]bool) {
	if button, ok := object.(*widget.Button); ok {
		labels[button.Text] = true
	}
	if children, ok := object.(*fyne.Container); ok {
		for _, child := range children.Objects {
			collectWorkbenchButtons(child, labels)
		}
	}
}
