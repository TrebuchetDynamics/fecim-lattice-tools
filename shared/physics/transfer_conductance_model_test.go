package physics

import (
	"math"
	"testing"
)

func TestConductance_LinearModel(t *testing.T) {
	P, Ps := 0.12, 0.30
	gmin, gmax := 1e-6, 100e-6
	got := PolarizationToConductanceModel(P, Ps, gmin, gmax, ConductanceLinear)
	want := gmin + (gmax-gmin)*(P/Ps+1)/2
	if math.Abs(got-want) > 1e-18 {
		t.Fatalf("linear model mismatch: got=%e want=%e", got, want)
	}
}

func TestConductance_SubthresholdModel(t *testing.T) {
	Ps := 0.3
	gmin, gmax := 1e-6, 100e-6
	lowP := -0.2 * Ps
	highP := 0.2 * Ps
	gLow := PolarizationToConductanceModel(lowP, Ps, gmin, gmax, ConductanceSubthreshold)
	gMid := PolarizationToConductanceModel(0, Ps, gmin, gmax, ConductanceSubthreshold)
	gHigh := PolarizationToConductanceModel(highP, Ps, gmin, gmax, ConductanceSubthreshold)
	if !(gLow < gMid && gMid < gHigh) {
		t.Fatalf("subthreshold monotonicity broken: low=%e mid=%e high=%e", gLow, gMid, gHigh)
	}
	linearLow := PolarizationToConductanceModel(lowP, Ps, gmin, gmax, ConductanceLinear)
	if !(gLow < linearLow) {
		t.Fatalf("expected low-P compression for subthreshold: sub=%e linear=%e", gLow, linearLow)
	}
}

func TestConductance_SubthresholdWindow(t *testing.T) {
	Ps := 0.3
	gmin, gmax := 1e-6, 100e-6
	gLo := PolarizationToConductanceModel(-Ps, Ps, gmin, gmax, ConductanceSubthreshold)
	gHi := PolarizationToConductanceModel(Ps, Ps, gmin, gmax, ConductanceSubthreshold)
	ratio := gHi / gLo
	want := gmax / gmin
	if math.Abs(ratio-want)/want > 1e-9 {
		t.Fatalf("window not preserved: got ratio=%g want=%g", ratio, want)
	}
}
