//go:build !cgo

package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildCircuitsView(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildCircuitsView(snapshot, theme)
	if w == nil {
		t.Fatal("buildCircuitsView returned nil")
	}
}

func TestCircuitsBoundaryNotice(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Error("Expected boundary notice in circuits descriptor")
	}
}

func TestCircuitsMetrics(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Metrics) == 0 {
		t.Error("Expected metrics in circuits snapshot")
	}
	foundADC := false
	for _, m := range snapshot.Metrics {
		if m.ID == "adc" {
			foundADC = true
		}
	}
	if !foundADC {
		t.Error("Expected ADC metric in circuits snapshot")
	}
}

func TestCircuitsViewActionButtonsDispatchActions(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	var actions []viewmodel.Action

	w := buildCircuitsViewWithActions(snapshot, theme, func(action viewmodel.Action) {
		actions = append(actions, action)
	})
	buttons := collectSidebarButtons(w)
	if len(buttons) < 12 {
		t.Fatalf("circuits button count = %d, want at least 12 command buttons", len(buttons))
	}

	clickButton(buttons[0])
	clickButton(buttons[1])
	clickButton(buttons[2])
	clickButton(buttons[3])
	clickButton(buttons[4])
	clickButton(buttons[5])
	clickButton(buttons[6])
	clickButton(buttons[7])
	clickButton(buttons[8])
	clickButton(buttons[9])
	clickButton(buttons[10])
	clickButton(buttons[11])

	wantIDs := []string{
		circuitsvm.ActionRunRead,
		circuitsvm.ActionRunWrite,
		circuitsvm.ActionRunCompute,
		circuitsvm.ActionExportOperationLog,
		circuitsvm.ActionExportReferenceSpecs,
		circuitsvm.ActionExportReferenceTiming,
		circuitsvm.ActionExportReferenceTimingSVG,
		circuitsvm.ActionAnimateReferenceTiming,
		circuitsvm.ActionPlayReferenceTiming,
		circuitsvm.ActionPauseReferenceTiming,
		circuitsvm.ActionStepReferenceTiming,
		circuitsvm.ActionResetReferenceTiming,
	}
	if len(actions) != len(wantIDs) {
		t.Fatalf("dispatched action count = %d, want %d", len(actions), len(wantIDs))
	}
	for i, want := range wantIDs {
		if actions[i].ID != want {
			t.Fatalf("action[%d].ID = %q, want %q", i, actions[i].ID, want)
		}
	}
}

func TestCircuitsViewSelectorButtonsDispatchPayloads(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	var actions []viewmodel.Action

	w := buildCircuitsViewWithActions(snapshot, theme, func(action viewmodel.Action) {
		actions = append(actions, action)
	})
	buttons := collectSidebarButtons(w)
	if len(buttons) < 26 {
		t.Fatalf("circuits button count = %d, want selector buttons", len(buttons))
	}

	clickButton(buttons[13])
	clickButton(buttons[16])
	clickButton(buttons[26])

	if got := actions[0]; got.ID != circuitsvm.ActionSetOperationMode || got.Payload["mode"] != circuitsvm.OperationWrite {
		t.Fatalf("first selector action = %#v, want write mode action", got)
	}
	if got := actions[1]; got.ID != circuitsvm.ActionSetTimingOperation || got.Payload["operation"] != "WRITE" {
		t.Fatalf("second selector action = %#v, want write timing operation action", got)
	}
	if got := actions[2]; got.ID != circuitsvm.ActionResizeArray || got.Payload["rows"] != "32" || got.Payload["cols"] != "32" {
		t.Fatalf("third selector action = %#v, want 32x32 resize action", got)
	}
}

func TestReferenceTimingPanelStateFollowsSnapshot(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetTimingOperation,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"operation": "COMPUTE"},
	}); err != nil {
		t.Fatalf("set compute timing operation: %v", err)
	}
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionPlayReferenceTiming, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("play compute timing: %v", err)
	}
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionStepReferenceTiming, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("step compute timing: %v", err)
	}

	state := referenceTimingPanelStateFromSnapshot(vm.Snapshot())
	if !state.available {
		t.Fatal("reference timing panel state should be available")
	}
	if state.title != "COMPUTE Reference Timing Panel" {
		t.Fatalf("panel title = %q, want COMPUTE Reference Timing Panel", state.title)
	}
	if state.summary != "COMPUTE panel / 6 signals / 5 markers / 3 phases / 76 ns" {
		t.Fatalf("panel summary = %q, want compute panel summary", state.summary)
	}
	if state.playbackStep != "2/6" {
		t.Fatalf("playbackStep = %q, want 2/6", state.playbackStep)
	}
	if len(state.signalRows) != 6 || state.signalRows[5] != "OUTPUT_VALID · 5 points" {
		t.Fatalf("signalRows = %+v, want OUTPUT_VALID row", state.signalRows)
	}
}

func TestBuildCircuitsViewIncludesReferenceTimingPanel(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))

	panel := buildReferenceTimingPanel(snapshot, theme)
	if panel == nil {
		t.Fatal("buildReferenceTimingPanel returned nil")
	}
	w := buildCircuitsView(snapshot, theme)
	if w == nil {
		t.Fatal("buildCircuitsView returned nil")
	}
}

func TestCircuitsViewCheckboxDispatchesToggle(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	var actions []viewmodel.Action

	w := buildCircuitsViewWithActions(snapshot, theme, func(action viewmodel.Action) {
		actions = append(actions, action)
	})
	checkboxes := collectCircuitCheckboxes(w)
	if len(checkboxes) != 1 {
		t.Fatalf("circuits checkbox count = %d, want 1", len(checkboxes))
	}

	clickCheckbox(checkboxes[0])

	if len(actions) != 1 {
		t.Fatalf("dispatched action count = %d, want 1", len(actions))
	}
	if got := actions[0].ID; got != circuitsvm.ActionToggleISPP {
		t.Fatalf("dispatched action = %q, want %q", got, circuitsvm.ActionToggleISPP)
	}
}

func collectCircuitCheckboxes(w widget.Widget) []*checkbox.Widget {
	var checkboxes []*checkbox.Widget
	if cb, ok := w.(*checkbox.Widget); ok {
		checkboxes = append(checkboxes, cb)
	}
	for _, child := range w.Children() {
		checkboxes = append(checkboxes, collectCircuitCheckboxes(child)...)
	}
	return checkboxes
}

func clickCheckbox(cb *checkbox.Widget) {
	cb.SetBounds(geometry.NewRect(0, 0, 220, 40))
	ctx := widget.NewContext()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, press)
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, release)
}
