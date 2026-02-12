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
