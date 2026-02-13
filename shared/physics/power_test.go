package physics

import (
	"math"
	"testing"
)

func TestCellDynamicPower_KnownValue(t *testing.T) {
	// 2 fF, 1.8 V, 100 MHz => 6.48e-7 W
	got := CellDynamicPower(2e-15, 1.8, 1e8)
	want := 6.48e-7
	if math.Abs(got-want) > 1e-15 {
		t.Fatalf("CellDynamicPower mismatch: got %.12e W want %.12e W", got, want)
	}
}

func TestArrayPower_IncludesAllComponents(t *testing.T) {
	params := ArrayPowerParams{
		Rows:            4,
		Cols:            4,
		ActiveFraction:  0.5,
		CellCapacitance: 2e-15,
		WriteVoltage:    1.8,
		ReadVoltage:     0.2,
		Frequency:       1e8,
		SelectorIoff:    10e-12,
		SelectorIShort:  2e-6,
		OverlapFactor:   0.1,
		PeripheralPower: 1e-6,
	}
	got := ArrayPower(params)

	pDynCell := 2e-15 * 1.8 * 1.8 * 1e8
	pLeakCell := 0.2 * 10e-12
	pShortCell := 1.8 * 2e-6 * 0.1
	wantDyn := 8.0 * pDynCell
	wantLeak := 16.0 * pLeakCell
	wantShort := 8.0 * pShortCell
	wantTotal := wantDyn + wantLeak + wantShort + 1e-6

	if math.Abs(got.DynamicPower-wantDyn) > 1e-15 {
		t.Fatalf("dynamic mismatch: got %.12e want %.12e", got.DynamicPower, wantDyn)
	}
	if math.Abs(got.LeakagePower-wantLeak) > 1e-18 {
		t.Fatalf("leakage mismatch: got %.12e want %.12e", got.LeakagePower, wantLeak)
	}
	if math.Abs(got.ShortCircuitPower-wantShort) > 1e-12 {
		t.Fatalf("short-circuit mismatch: got %.12e want %.12e", got.ShortCircuitPower, wantShort)
	}
	if math.Abs(got.TotalPower-wantTotal) > 1e-12 {
		t.Fatalf("total mismatch: got %.12e want %.12e", got.TotalPower, wantTotal)
	}
}
