package algo

import (
	"math"
	"reflect"
	"testing"
)

func TestModule1AlgoE2ERuntimeCalibrationManagerWideUpdateMatrix(t *testing.T) {
	const ec = 1.1e8
	cm := NewCalibrationManager(12)

	upEvents := []struct {
		target int
		errors []int
	}{
		{target: 3, errors: []int{-3, -1, 2, -1}},
		{target: 6, errors: []int{-2, -1, 1, -1}},
		{target: 9, errors: []int{-4, -2, 2, 1}},
	}
	for _, event := range upEvents {
		for _, err := range event.errors {
			cm.UpdateCalibrationUp(event.target, err, ec)
		}
		if cm.LastErrorUp[event.target] != event.errors[len(event.errors)-1] {
			t.Fatalf("LastErrorUp[%d] = %d, want %d", event.target, cm.LastErrorUp[event.target], event.errors[len(event.errors)-1])
		}
		if cm.CalibrationUp[event.target] < 0.3*ec || cm.CalibrationUp[event.target] > 2.5*ec {
			t.Fatalf("CalibrationUp[%d] out of soft bounds: %.6e", event.target, cm.CalibrationUp[event.target])
		}
		if cm.RelaxCompUp[event.target] < -0.05 || cm.RelaxCompUp[event.target] > 0.25 {
			t.Fatalf("RelaxCompUp[%d] out of bounds: %.6f", event.target, cm.RelaxCompUp[event.target])
		}
	}

	downEvents := []struct {
		target int
		errors []int
	}{
		{target: 2, errors: []int{3, 1, -2, 1}},
		{target: 5, errors: []int{2, 1, -1, 1}},
		{target: 8, errors: []int{4, 2, -2, -1}},
	}
	for _, event := range downEvents {
		for _, err := range event.errors {
			cm.UpdateCalibrationDown(event.target, err, ec)
		}
		if cm.LastErrorDown[event.target] != event.errors[len(event.errors)-1] {
			t.Fatalf("LastErrorDown[%d] = %d, want %d", event.target, cm.LastErrorDown[event.target], event.errors[len(event.errors)-1])
		}
		if cm.CalibrationDown[event.target] < -2.5*ec || cm.CalibrationDown[event.target] > -0.3*ec {
			t.Fatalf("CalibrationDown[%d] out of soft bounds: %.6e", event.target, cm.CalibrationDown[event.target])
		}
		if cm.RelaxCompDown[event.target] < -0.05 || cm.RelaxCompDown[event.target] > 0.25 {
			t.Fatalf("RelaxCompDown[%d] out of bounds: %.6f", event.target, cm.RelaxCompDown[event.target])
		}
	}

	assertFiniteCalibrationManagerE2E(t, cm)
	for _, idx := range []int{3, 6, 9} {
		cm.EnforceMonotonicityUp(idx, ec)
	}
	for _, idx := range []int{2, 5, 8} {
		cm.EnforceMonotonicityDown(idx, ec)
	}
	assertLocalUpMonotonicE2E(t, cm, ec, 3, 6, 9)
	assertLocalDownMonotonicE2E(t, cm, ec, 2, 5, 8)
}

func TestModule1AlgoE2ECalibrationInvalidInputsAreSideEffectSafe(t *testing.T) {
	cm := NewCalibrationManager(5)
	cm.UpdateCalibrationUp(2, -2, 1e8)
	cm.UpdateCalibrationDown(2, 2, 1e8)
	before := cloneCalibrationManagerE2E(cm)

	cm.UpdateCalibrationUp(-1, -10, 1e8)
	cm.UpdateCalibrationUp(5, -10, 1e8)
	cm.UpdateCalibrationDown(-1, 10, 1e8)
	cm.UpdateCalibrationDown(5, 10, 1e8)
	cm.EnforceMonotonicityUp(0, 1e8)
	cm.EnforceMonotonicityUp(5, 1e8)
	cm.EnforceMonotonicityDown(0, 1e8)
	cm.EnforceMonotonicityDown(4, 1e8)

	if !reflect.DeepEqual(before, cloneCalibrationManagerE2E(cm)) {
		t.Fatalf("invalid calibration operations mutated manager\nbefore=%+v\nafter=%+v", before, cloneCalibrationManagerE2E(cm))
	}
}

type calibrationManagerSnapshotE2E struct {
	Up, Down                         []float64
	UpLow, UpHigh, DownLow, DownHigh []float64
	LastUp, LastDown                 []int
	RelaxUp, RelaxDown               []float64
}

func cloneCalibrationManagerE2E(cm *CalibrationManager) calibrationManagerSnapshotE2E {
	return calibrationManagerSnapshotE2E{
		Up:        append([]float64(nil), cm.CalibrationUp...),
		Down:      append([]float64(nil), cm.CalibrationDown...),
		UpLow:     append([]float64(nil), cm.CalibUpLow...),
		UpHigh:    append([]float64(nil), cm.CalibUpHigh...),
		DownLow:   append([]float64(nil), cm.CalibDownLow...),
		DownHigh:  append([]float64(nil), cm.CalibDownHigh...),
		LastUp:    append([]int(nil), cm.LastErrorUp...),
		LastDown:  append([]int(nil), cm.LastErrorDown...),
		RelaxUp:   append([]float64(nil), cm.RelaxCompUp...),
		RelaxDown: append([]float64(nil), cm.RelaxCompDown...),
	}
}

func assertFiniteCalibrationManagerE2E(t *testing.T, cm *CalibrationManager) {
	t.Helper()
	all := [][]float64{cm.CalibrationUp, cm.CalibrationDown, cm.CalibUpLow, cm.CalibUpHigh, cm.CalibDownLow, cm.CalibDownHigh, cm.RelaxCompUp, cm.RelaxCompDown}
	for groupIdx, group := range all {
		for i, value := range group {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("calibration group %d index %d invalid: %g", groupIdx, i, value)
			}
		}
	}
}

func assertLocalUpMonotonicE2E(t *testing.T, cm *CalibrationManager, ec float64, indices ...int) {
	t.Helper()
	minStep := ec * 0.02
	for _, idx := range indices {
		if idx > 0 && cm.CalibrationUp[idx] < cm.CalibrationUp[idx-1]+minStep-1e-6 {
			t.Fatalf("up monotonicity failed at %d: %.6e vs prev %.6e", idx, cm.CalibrationUp[idx], cm.CalibrationUp[idx-1])
		}
		if idx < len(cm.CalibrationUp)-1 && cm.CalibrationUp[idx+1] < cm.CalibrationUp[idx]+minStep-1e-6 {
			t.Fatalf("up neighbor monotonicity failed at %d: next %.6e current %.6e", idx, cm.CalibrationUp[idx+1], cm.CalibrationUp[idx])
		}
	}
}

func assertLocalDownMonotonicE2E(t *testing.T, cm *CalibrationManager, ec float64, indices ...int) {
	t.Helper()
	minStep := ec * 0.02
	for _, idx := range indices {
		if idx > 0 && cm.CalibrationDown[idx] < cm.CalibrationDown[idx-1]+minStep-1e-6 {
			t.Fatalf("down monotonicity failed at %d: %.6e vs prev %.6e", idx, cm.CalibrationDown[idx], cm.CalibrationDown[idx-1])
		}
		if idx < len(cm.CalibrationDown)-1 && cm.CalibrationDown[idx+1] < cm.CalibrationDown[idx]+minStep-1e-6 {
			t.Fatalf("down neighbor monotonicity failed at %d: next %.6e current %.6e", idx, cm.CalibrationDown[idx+1], cm.CalibrationDown[idx])
		}
	}
}
