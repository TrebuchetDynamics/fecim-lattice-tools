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

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	// Same row (r=2, c!=3) should show non-zero residue.
	for c := 0; c < 8; c++ {
		if c == targetCol {
			continue
		}
		if ca.halfSelectResidue[targetRow][c] == 0 {
			t.Fatalf("expected non-zero halfSelectResidue at same-row cell (%d,%d)", targetRow, c)
		}
	}

	// Same column (c=3, r!=2) should show non-zero residue.
	for r := 0; r < 8; r++ {
		if r == targetRow {
			continue
		}
		if ca.halfSelectResidue[r][targetCol] == 0 {
			t.Fatalf("expected non-zero halfSelectResidue at same-col cell (%d,%d)", r, targetCol)
		}
	}

	// Unselected cell should stay zero/minimal; model uses exactly 0 for non-row/non-col.
	if got := ca.halfSelectResidue[0][0]; got != 0 {
		t.Fatalf("expected zero residue at unselected cell (0,0), got %.6f", got)
	}
}
