package gui

import "testing"

func TestApplyHalfSelectDisturb_AccumulatesResidueDuringWrite(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca.deviceState == nil {
		t.Fatal("expected device state")
	}
	ca.deviceState.SetPassiveMode(true)

	targetRow, targetCol := 1, 1
	writeV := ca.deviceState.GetWriteRange().Max
	if writeV <= 0 {
		writeV = 1.8
	}
	ca.deviceState.ApplyHalfSelectWrite(targetRow, targetCol, writeV)

	changes := ca.applyHalfSelectDisturb(targetRow, targetCol)
	if changes < 0 {
		t.Fatalf("unexpected negative change count: %d", changes)
	}

	// Half-selected neighbors should accumulate signed residue.
	nonZero := 0
	for r := range ca.halfSelectResidue {
		for c := range ca.halfSelectResidue[r] {
			if r == targetRow && c == targetCol {
				continue
			}
			if ca.halfSelectResidue[r][c] != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatalf("expected non-zero half-select residue after write disturb pulse")
	}
}

func TestApplyHalfSelectDisturb_8x8_WriteAt2x3_DisturbsFullRowAndColumn(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca.deviceState == nil {
		t.Fatal("expected device state")
	}
	ca.resizeArray(8, 8)
	ca.deviceState.SetPassiveMode(true)

	targetRow, targetCol := 2, 3
	writeV := ca.deviceState.GetWriteRange().Max
	if writeV <= 0 {
		writeV = 1.8
	}
	ca.deviceState.ApplyHalfSelectWrite(targetRow, targetCol, writeV)
	ca.applyHalfSelectDisturb(targetRow, targetCol)

	for c := 0; c < 8; c++ {
		if c == targetCol {
			continue
		}
		if ca.halfSelectResidue[targetRow][c] == 0 {
			t.Fatalf("expected non-zero halfSelectResidue on target row cell (%d,%d)", targetRow, c)
		}
	}
	for r := 0; r < 8; r++ {
		if r == targetRow {
			continue
		}
		if ca.halfSelectResidue[r][targetCol] == 0 {
			t.Fatalf("expected non-zero halfSelectResidue on target col cell (%d,%d)", r, targetCol)
		}
	}
}
