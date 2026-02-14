package gui

import "testing"

func TestPassiveBehavior_0T1R_HalfSelectResidue_RowAndColumn(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Requirement: validate behavior on an 8x8 passive array.
	ca.resizeArray(8, 8)
	ca.deviceState.SetPassiveMode(true)
	ca.deviceState.SetOperationMode(OpModeWrite)

	targetRow, targetCol := 2, 3
	ca.deviceState.SetSelectedCell(targetRow, targetCol)

	writeV := ca.deviceState.GetWriteRange().Max
	if writeV <= 0 {
		writeV = 1.8
	}
	ca.deviceState.ApplyHalfSelectWrite(targetRow, targetCol, writeV)
	_ = ca.applyHalfSelectDisturb(targetRow, targetCol)

	// Verify stress accumulated on half-selected cells via WriteDisturbEngine.
	if ca.writeDisturbEngine == nil {
		t.Fatal("expected writeDisturbEngine to be initialized after disturb")
	}

	// Same row (r=2, c!=3) should show non-zero stress.
	for c := 0; c < 8; c++ {
		if c == targetCol {
			continue
		}
		if ca.writeDisturbEngine.GetCellStress(targetRow, c) == 0 {
			t.Fatalf("expected non-zero stress at same-row cell (%d,%d)", targetRow, c)
		}
	}

	// Same column (c=3, r!=2) should show non-zero stress.
	for r := 0; r < 8; r++ {
		if r == targetRow {
			continue
		}
		if ca.writeDisturbEngine.GetCellStress(r, targetCol) == 0 {
			t.Fatalf("expected non-zero stress at same-col cell (%d,%d)", r, targetCol)
		}
	}

	// Unselected cell (different row AND column) should have zero stress.
	if got := ca.writeDisturbEngine.GetCellStress(0, 0); got != 0 {
		t.Fatalf("expected zero stress at unselected cell (0,0), got %.6f", got)
	}
}
