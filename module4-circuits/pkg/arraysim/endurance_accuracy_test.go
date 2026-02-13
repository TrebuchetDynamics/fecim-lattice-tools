package arraysim

import "testing"

func TestSimulateEnduranceAccuracy_MonotonicDegradation(t *testing.T) {
	pts := SimulateEnduranceAccuracy([]float64{0, 1e6, 1e8, 1e9}, EnduranceAccuracyConfig{
		BaselineAccuracy: 0.98,
		EnduranceLimit:   1e9,
		DriftAtLimit:     0.20,
		Sensitivity:      0.50,
	})
	if len(pts) != 4 {
		t.Fatalf("len=%d want 4", len(pts))
	}
	for i := 1; i < len(pts); i++ {
		if pts[i].ConductanceDrift < pts[i-1].ConductanceDrift {
			t.Fatalf("drift not monotonic at %d", i)
		}
		if pts[i].Accuracy > pts[i-1].Accuracy {
			t.Fatalf("accuracy not monotonic at %d", i)
		}
	}
}

func TestSimulateEnduranceAccuracy_ClampsNegativeCycles(t *testing.T) {
	pts := SimulateEnduranceAccuracy([]float64{-1, 0}, EnduranceAccuracyConfig{})
	if pts[0].Cycles != 0 {
		t.Fatalf("negative cycles should clamp to 0, got %.2f", pts[0].Cycles)
	}
}
