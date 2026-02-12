package undo

import (
	"reflect"
	"testing"
)

type undoSystemState struct {
	// Module 2 / crossbar-style state
	ProgrammedCellRow int
	ProgrammedCellCol int
	ProgrammedValue   float64

	// Module 1 / materials-style state
	Material string

	// Module 3 / array configuration state
	ArrayRows int
	ArrayCols int
}

func TestUndoComprehensiveAcrossModules(t *testing.T) {
	m := NewManager(10)

	state := undoSystemState{
		ProgrammedCellRow: 1,
		ProgrammedCellCol: 2,
		ProgrammedValue:   0.25,
		Material:          "HZO",
		ArrayRows:         64,
		ArrayCols:         64,
	}
	initial := state

	// (1) Perform actions representative of multiple modules:
	// - program cell
	// - change material
	// - resize array
	beforeProgram := state
	programCmd := NewFuncCommand(
		"Program cell (module2-crossbar)",
		func() {
			state.ProgrammedCellRow = 3
			state.ProgrammedCellCol = 5
			state.ProgrammedValue = 0.91
		},
		func() {
			state = beforeProgram
		},
	)
	m.Execute(programCmd)
	afterProgram := state

	materialCmd := NewStringCommand(
		"Change material (module1-hysteresis)",
		state.Material,
		"AlScN",
		func(v string) {
			state.Material = v
		},
	)
	m.Execute(materialCmd)
	afterMaterial := state

	beforeResize := state
	resizeCmd := NewFuncCommand(
		"Resize array (module3-mnist)",
		func() {
			state.ArrayRows = 128
			state.ArrayCols = 32
		},
		func() {
			state = beforeResize
		},
	)
	m.Execute(resizeCmd)
	afterResize := state

	if m.UndoCount() != 3 {
		t.Fatalf("expected 3 undo commands after three actions, got %d", m.UndoCount())
	}
	if !reflect.DeepEqual(state, undoSystemState{
		ProgrammedCellRow: 3,
		ProgrammedCellCol: 5,
		ProgrammedValue:   0.91,
		Material:          "AlScN",
		ArrayRows:         128,
		ArrayCols:         32,
	}) {
		t.Fatalf("unexpected post-action state: %+v", state)
	}

	// (2) Undo and verify exact state reversion for each step.
	if !m.Undo() {
		t.Fatal("expected undo (resize) to succeed")
	}
	if !reflect.DeepEqual(state, afterMaterial) {
		t.Fatalf("resize undo mismatch\nwant: %+v\n got: %+v", afterMaterial, state)
	}

	if !m.Undo() {
		t.Fatal("expected undo (material) to succeed")
	}
	if !reflect.DeepEqual(state, afterProgram) {
		t.Fatalf("material undo mismatch\nwant: %+v\n got: %+v", afterProgram, state)
	}

	if !m.Undo() {
		t.Fatal("expected undo (program) to succeed")
	}
	if !reflect.DeepEqual(state, initial) {
		t.Fatalf("program undo mismatch\nwant: %+v\n got: %+v", initial, state)
	}

	// (3) Redo and verify exact post-action restoration.
	if !m.Redo() {
		t.Fatal("expected redo (program) to succeed")
	}
	if !reflect.DeepEqual(state, afterProgram) {
		t.Fatalf("program redo mismatch\nwant: %+v\n got: %+v", afterProgram, state)
	}

	if !m.Redo() {
		t.Fatal("expected redo (material) to succeed")
	}
	if !reflect.DeepEqual(state, afterMaterial) {
		t.Fatalf("material redo mismatch\nwant: %+v\n got: %+v", afterMaterial, state)
	}

	if !m.Redo() {
		t.Fatal("expected redo (resize) to succeed")
	}
	if !reflect.DeepEqual(state, afterResize) {
		t.Fatalf("resize redo mismatch\nwant: %+v\n got: %+v", afterResize, state)
	}
}

func TestUndoComprehensiveDepthLimitAndRedoClear(t *testing.T) {
	// (4) Verify undo stack depth limit is respected.
	m := NewManager(3)
	value := 0

	for i := 1; i <= 5; i++ {
		oldValue := value
		newValue := i
		m.Execute(NewIntCommand("depth-limited-step", oldValue, newValue, func(v int) {
			value = v
		}))
	}

	if got := m.UndoCount(); got != 3 {
		t.Fatalf("undo depth limit not respected: expected 3, got %d", got)
	}

	// Undo only 3 retained history entries: 5->4->3->2.
	for i := 0; i < 3; i++ {
		if !m.Undo() {
			t.Fatalf("expected undo %d to succeed", i+1)
		}
	}
	if value != 2 {
		t.Fatalf("expected value 2 after undoing retained history, got %d", value)
	}
	if m.Undo() {
		t.Fatal("expected no more undo operations after retained history exhausted")
	}

	// (5) Verify undo after a new action clears redo stack.
	m2 := NewManager(10)
	value2 := 0

	m2.Execute(NewIntCommand("step1", 0, 10, func(v int) { value2 = v }))
	m2.Execute(NewIntCommand("step2", 10, 20, func(v int) { value2 = v }))
	if !m2.Undo() {
		t.Fatal("expected undo to succeed")
	}
	if got := m2.RedoCount(); got != 1 {
		t.Fatalf("expected redo stack size 1 after undo, got %d", got)
	}

	m2.Execute(NewIntCommand("step3", 10, 30, func(v int) { value2 = v }))
	if got := m2.RedoCount(); got != 0 {
		t.Fatalf("expected redo stack cleared after new action, got %d", got)
	}
	if m2.Redo() {
		t.Fatal("expected redo to fail after new action clears redo timeline")
	}
	if value2 != 30 {
		t.Fatalf("expected value 30 after new action, got %d", value2)
	}
}
