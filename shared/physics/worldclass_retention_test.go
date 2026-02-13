package physics

import (
	"math"
	"testing"
)

func TestGenerateLogTimeSweep(t *testing.T) {
	times, err := GenerateLogTimeSweep(1e-6, 1, 7)
	if err != nil {
		t.Fatalf("GenerateLogTimeSweep error: %v", err)
	}
	if times[0] != 1e-6 {
		t.Fatalf("first time = %g, want 1e-6", times[0])
	}
	if math.Abs(times[len(times)-1]-1) > 1e-15 {
		t.Fatalf("last time = %g, want 1", times[len(times)-1])
	}
	for i := 1; i < len(times); i++ {
		if !(times[i] > times[i-1]) {
			t.Fatalf("not increasing at i=%d", i)
		}
	}
}

func TestSimulateRetentionExponential(t *testing.T) {
	times := []float64{0, 1, 2}
	pts, err := SimulateRetentionExponential(0.3, 0.1, 1, times)
	if err != nil {
		t.Fatalf("SimulateRetentionExponential error: %v", err)
	}
	if math.Abs(pts[0].Polarization_Cm-0.3) > 1e-15 {
		t.Fatalf("P(0) = %g, want 0.3", pts[0].Polarization_Cm)
	}
	want1 := 0.1 + 0.2*math.Exp(-1)
	if math.Abs(pts[1].Polarization_Cm-want1) > 1e-15 {
		t.Fatalf("P(1) = %g, want %g", pts[1].Polarization_Cm, want1)
	}
}
