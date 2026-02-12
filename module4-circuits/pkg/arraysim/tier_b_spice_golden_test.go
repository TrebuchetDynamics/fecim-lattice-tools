package arraysim

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type spiceGoldenFile struct {
	Cases []spiceGoldenCase `json:"cases"`
}

type spiceGoldenCase struct {
	Name     string            `json:"name"`
	Params   spiceGoldenParams `json:"params"`
	Expected DCResult          `json:"expected"`
}

type spiceGoldenParams struct {
	WLVoltages   []float64            `json:"WLVoltages"`
	BLVoltages   []float64            `json:"BLVoltages"`
	Conductance  [][]float64          `json:"Conductance"`
	ActiveRows   []bool               `json:"ActiveRows"`
	Wire         WireParams           `json:"Wire"`
	Boundary     BoundaryParams       `json:"Boundary"`
	SelectorMode string               `json:"SelectorMode"`
	ReadMask     [][]bool             `json:"ReadMask"`
	WriteMask    [][]bool             `json:"WriteMask"`
	Selector     SelectorDeviceParams `json:"Selector"`
}

func (p spiceGoldenParams) toSolveParams() SolveParams {
	mode := SelectorBypass
	switch p.SelectorMode {
	case "read":
		mode = SelectorRead
	case "write":
		mode = SelectorWrite
	}
	return SolveParams{
		WLVoltages:   p.WLVoltages,
		BLVoltages:   p.BLVoltages,
		Conductance:  p.Conductance,
		ActiveRows:   p.ActiveRows,
		Wire:         p.Wire,
		Boundary:     p.Boundary,
		SelectorMode: mode,
		ReadMask:     p.ReadMask,
		WriteMask:    p.WriteMask,
		Selector:     p.Selector,
	}
}

func TestTierB_AgreesWithSpiceGoldenVectors(t *testing.T) {
	path := filepath.Join("testdata", "tierb_spice_golden_vectors.json")
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	var gf spiceGoldenFile
	if err := json.Unmarshal(blob, &gf); err != nil {
		t.Fatalf("parse golden file: %v", err)
	}
	if len(gf.Cases) == 0 {
		t.Fatal("no golden cases")
	}

	for _, tc := range gf.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := NewTierBSolver().SolveDC(tc.Params.toSolveParams())
			if err != nil {
				t.Fatalf("tier-b solve: %v", err)
			}

			const tol = 2e-11
			assertCloseGrid(t, "WLNodes", tc.Expected.WLNodes, got.WLNodes, tol)
			assertCloseGrid(t, "BLNodes", tc.Expected.BLNodes, got.BLNodes, tol)
			assertCloseGrid(t, "CellVoltages", tc.Expected.CellVoltages, got.CellVoltages, tol)
			assertCloseGrid(t, "CellCurrents", tc.Expected.CellCurrents, got.CellCurrents, tol)
			assertCloseVec(t, "RowCurrents", tc.Expected.RowCurrents, got.RowCurrents, tol)
			assertCloseVec(t, "ColCurrents", tc.Expected.ColCurrents, got.ColCurrents, tol)
		})
	}
}

func assertCloseGrid(t *testing.T, name string, want, got [][]float64, tol float64) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("%s row count mismatch: got=%d want=%d", name, len(got), len(want))
	}
	for r := range want {
		if len(want[r]) != len(got[r]) {
			t.Fatalf("%s col count mismatch row=%d got=%d want=%d", name, r, len(got[r]), len(want[r]))
		}
		for c := range want[r] {
			d := math.Abs(want[r][c] - got[r][c])
			if d > tol {
				t.Fatalf("%s[%d][%d] got=%.12g want=%.12g |delta|=%.3g > %.3g", name, r, c, got[r][c], want[r][c], d, tol)
			}
		}
	}
}

func assertCloseVec(t *testing.T, name string, want, got []float64, tol float64) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("%s len mismatch: got=%d want=%d", name, len(got), len(want))
	}
	for i := range want {
		d := math.Abs(want[i] - got[i])
		if d > tol {
			t.Fatalf("%s[%d] got=%.12g want=%.12g |delta|=%.3g > %.3g", name, i, got[i], want[i], d, tol)
		}
	}
}
