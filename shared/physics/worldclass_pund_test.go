package physics

import (
	"math"
	"testing"
)

func makeFlatPulse(durationS, currentA float64, n int) []PulseSample {
	out := make([]PulseSample, n)
	dt := durationS / float64(n-1)
	for i := range out {
		out[i] = PulseSample{TimeS: float64(i) * dt, CurrentA: currentA}
	}
	return out
}

func TestAnalyzePUND_SeparatesSwitchingAndNonSwitching(t *testing.T) {
	// Q = I*t. Use 1us pulses.
	pulseDur := 1e-6
	p := makeFlatPulse(pulseDur, 8e-6, 11) // 8 pC
	u := makeFlatPulse(pulseDur, 3e-6, 11) // 3 pC (non-switching)
	n := makeFlatPulse(pulseDur, -7e-6, 11)
	d := makeFlatPulse(pulseDur, -2e-6, 11)

	res, err := AnalyzePUND(p, u, n, d)
	if err != nil {
		t.Fatalf("AnalyzePUND error: %v", err)
	}

	if math.Abs(res.SwitchingPositive_C-5e-12) > 1e-18 {
		t.Fatalf("SwitchingPositive_C = %g, want %g", res.SwitchingPositive_C, 5e-12)
	}
	if math.Abs(res.SwitchingNegative_C-(-5e-12)) > 1e-18 {
		t.Fatalf("SwitchingNegative_C = %g, want %g", res.SwitchingNegative_C, -5e-12)
	}
}

func TestIntegrateCurrent_RejectsNonMonotonicTime(t *testing.T) {
	_, err := IntegrateCurrent([]PulseSample{{TimeS: 0, CurrentA: 1}, {TimeS: 0, CurrentA: 1}})
	if err == nil {
		t.Fatal("expected error for non-monotonic time")
	}
}
