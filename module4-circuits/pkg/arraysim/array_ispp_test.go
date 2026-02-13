package arraysim

import (
	"encoding/json"
	"math"
	"testing"
)

func testArrayCfg(rows, cols int) ArrayConfig {
	cfg := withAnalysisDefaults(ArrayConfig{Rows: rows, Cols: cols, CouplingMode: CouplingTierA})
	cfg.Wire.RWordLine = 0.02
	cfg.Wire.RBitLine = 0.02
	cfg.Boundary.WLDriveResistance = 0.02
	cfg.Boundary.BLDriveResistance = 0.02
	return cfg
}

func TestArrayISPP_8x8_Checkerboard(t *testing.T) {
	rows, cols := 8, 8
	target := make([][]int, rows)
	for r := 0; r < rows; r++ {
		target[r] = make([]int, cols)
		for c := 0; c < cols; c++ {
			if (r+c)%2 == 0 {
				target[r][c] = 5
			} else {
				target[r][c] = 25
			}
		}
	}
	res, err := ProgramArray(testArrayCfg(rows, cols), target, ProgramOpts{Order: "checkerboard", MaxPulses: 30, VerifyAfter: true, AccumDisturb: true})
	if err != nil {
		t.Fatalf("ProgramArray error: %v", err)
	}
	assertAllWithin(t, res, 1)
	logJSONSummary(t, res)
}

func TestArrayISPP_16x16_Gradient(t *testing.T) {
	rows, cols := 16, 16
	target := make([][]int, rows)
	for r := 0; r < rows; r++ {
		target[r] = make([]int, cols)
		for c := 0; c < cols; c++ {
			v := r + c
			if v > 29 {
				v = 29
			}
			target[r][c] = v
		}
	}
	res, err := ProgramArray(testArrayCfg(rows, cols), target, ProgramOpts{Order: "row-major", MaxPulses: 30, VerifyAfter: true, AccumDisturb: true})
	if err != nil {
		t.Fatalf("ProgramArray error: %v", err)
	}
	assertAllWithin(t, res, 1)
	logJSONSummary(t, res)
}

func TestArrayISPP_DisturbDoesNotFlip(t *testing.T) {
	rows, cols := 8, 8
	target := make([][]int, rows)
	for r := 0; r < rows; r++ {
		target[r] = make([]int, cols)
		for c := 0; c < cols; c++ {
			target[r][c] = 12
		}
	}
	res, err := ProgramArray(testArrayCfg(rows, cols), target, ProgramOpts{Order: "col-major", MaxPulses: 25, VerifyAfter: true, AccumDisturb: true})
	if err != nil {
		t.Fatalf("ProgramArray error: %v", err)
	}
	for r := range res.Cells {
		for c := range res.Cells[r] {
			if math.Abs(float64(res.Cells[r][c].LevelError)) > 1 {
				t.Fatalf("disturb flip exceeded tolerance at [%d,%d]: err=%d recv=%.4f", r, c, res.Cells[r][c].LevelError, res.Cells[r][c].DisturbRecv)
			}
		}
	}
	logJSONSummary(t, res)
}

func TestArrayISPP_ScalingNotQuartic(t *testing.T) {
	mkTarget := func(n int) [][]int {
		t := make([][]int, n)
		for r := 0; r < n; r++ {
			t[r] = make([]int, n)
			for c := 0; c < n; c++ {
				t[r][c] = (r + c) % 30
			}
		}
		return t
	}

	res8, err := ProgramArray(testArrayCfg(8, 8), mkTarget(8), ProgramOpts{Order: "row-major", MaxPulses: 20, VerifyAfter: false, AccumDisturb: true})
	if err != nil {
		t.Fatalf("ProgramArray 8x8 error: %v", err)
	}
	res16, err := ProgramArray(testArrayCfg(16, 16), mkTarget(16), ProgramOpts{Order: "row-major", MaxPulses: 20, VerifyAfter: false, AccumDisturb: true})
	if err != nil {
		t.Fatalf("ProgramArray 16x16 error: %v", err)
	}

	ratio := res16.ProgramTimeNs / math.Max(res8.ProgramTimeNs, 1)
	if ratio >= 8.0 {
		t.Fatalf("scaling looks quartic or worse: ratio=%.2f (8x8=%.1fns,16x16=%.1fns)", ratio, res8.ProgramTimeNs, res16.ProgramTimeNs)
	}
	t.Logf("scaling ratio 16x16/8x8 = %.2f", ratio)
	logJSONSummary(t, res8)
	logJSONSummary(t, res16)
}

func assertAllWithin(t *testing.T, res *ProgramResult, tol int) {
	t.Helper()
	if res == nil {
		t.Fatal("nil ProgramResult")
	}
	for r := range res.Cells {
		for c := range res.Cells[r] {
			if absInt(res.Cells[r][c].LevelError) > tol {
				t.Fatalf("cell [%d,%d] error too large: got %d tol %d", r, c, res.Cells[r][c].LevelError, tol)
			}
		}
	}
}

func logJSONSummary(t *testing.T, res *ProgramResult) {
	t.Helper()
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Logf("json marshal error: %v", err)
		return
	}
	t.Log(string(b))
}
