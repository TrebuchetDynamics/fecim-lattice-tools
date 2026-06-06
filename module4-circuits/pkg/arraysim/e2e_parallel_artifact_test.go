package arraysim

import (
	"encoding/json"
	"math"
	"strings"
	"sync"
	"testing"
)

func TestModule4ArraySimE2EParallelPatternProgramSolveExportWorkflows(t *testing.T) {
	cases := []struct {
		name    string
		rows    int
		cols    int
		pattern func(int, int, int) [][]int
		mode    CouplingMode
		order   string
	}{
		{name: "checker-tier-a", rows: 4, cols: 4, pattern: GenerateCheckerboard, mode: CouplingTierA, order: "checkerboard"},
		{name: "diagonal-tier-b", rows: 5, cols: 3, pattern: GenerateDiagonal, mode: CouplingTierB, order: "row-major"},
		{name: "stripe-ideal", rows: 3, cols: 6, pattern: GenerateRowStripe, mode: CouplingIdeal, order: "col-major"},
		{name: "random-tier-b", rows: 4, cols: 5, pattern: func(r, c, q int) [][]int { return GenerateRandom(r, c, q, 77) }, mode: CouplingTierB, order: "checkerboard"},
	}

	var wg sync.WaitGroup
	for _, tc := range cases {
		tc := tc
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := ArrayConfig{
				Rows: tc.rows, Cols: tc.cols, ReadVoltageV: 0.2, CouplingMode: tc.mode,
				Wire:  WireParams{RWordLine: 0.6 + 0.1*float64(tc.rows), RBitLine: 2.0 + 0.2*float64(tc.cols)},
				Sense: SenseChain{TIA: TIAConfig{Rf: 2e4, Vref: 0, Vmin: 0, Vmax: 1.1}, ADC: ADCConfig{Bits: 6, Vmin: 0, Vmax: 1.1}},
			}
			targets := tc.pattern(tc.rows, tc.cols, 30)
			programmed, err := ProgramArray(cfg, targets, ProgramOpts{Order: tc.order, MaxPulses: 20, VerifyAfter: true, AccumDisturb: true})
			if err != nil {
				t.Errorf("%s ProgramArray: %v", tc.name, err)
				return
			}
			if programmed.TotalPulses <= 0 || len(programmed.Cells) != tc.rows || programmed.ProgramTimeNs <= 0 {
				t.Errorf("%s program summary invalid: %+v", tc.name, programmed)
				return
			}
			encoded, err := json.Marshal(programmed)
			if err != nil || !strings.Contains(string(encoded), "TotalPulses") {
				t.Errorf("%s program JSON invalid: %v %s", tc.name, err, encoded)
				return
			}

			conductance := conductanceFromPatternE2E(targets)
			params := SolveParams{WLVoltages: wlVectorE2E(tc.rows, 0.2), BLVoltages: make([]float64, tc.cols), Conductance: conductance, Geometry: DefaultCellGeometry(), Wire: cfg.Wire}
			if tc.mode == CouplingTierB {
				params.SelectorMode = SelectorRead
				params.ReadMask = checkerMaskE2E(tc.rows, tc.cols)
			}
			res, ok := solveRead(cfg, params)
			if !ok || len(res.CellCurrents) != tc.rows || len(res.ColCurrents) != tc.cols {
				t.Errorf("%s solve invalid ok=%v result=%+v", tc.name, ok, res)
				return
			}
			if math.IsNaN(absCurrentSumE2E(res)) || absCurrentSumE2E(res) < 0 {
				t.Errorf("%s current sum invalid: %+v", tc.name, res)
				return
			}
			margin := ReadMarginAnalysis(cfg, 6)
			if len(margin.MarginPerLevel) != 5 || margin.MinMarginV < 0 || margin.CouplingMode == "" {
				t.Errorf("%s read margin invalid: %+v", tc.name, margin)
				return
			}
			deck, err := ExportCrossbarSPICE(params, SpiceExportConfig{Title: "Module4 parallel " + tc.name})
			if err != nil || !strings.Contains(deck, "Module4 parallel "+tc.name) || strings.Count(deck, "RCELL_") != tc.rows*tc.cols {
				t.Errorf("%s SPICE invalid err=%v deck=%s", tc.name, err, deck)
				return
			}
			transient := TransientSolve(cfg, []PulseStep{{Voltage: 2.0, DurationNs: 4, RiseTimeNs: 1}, {Voltage: 0.2, DurationNs: 2}}, 0.5)
			if len(transient) != tc.rows*tc.cols || len(transient[0].TimeNs) == 0 {
				t.Errorf("%s transient invalid: %d %+v", tc.name, len(transient), transient[:minIntE2E(len(transient), 1)])
			}
		}()
	}
	wg.Wait()
}

func TestModule4ArraySimE2EConcurrentTierBSolvesAreDeterministic(t *testing.T) {
	params := SolveParams{
		WLVoltages:  []float64{0.25, 0.2, 0.15, 0.1},
		BLVoltages:  []float64{0, 0.01, 0, 0.02},
		Conductance: conductanceFromPatternE2E(GenerateDiagonal(4, 4, 30)),
		Geometry:    DefaultCellGeometry(),
		Wire:        WireParams{RWordLine: 0.9, RBitLine: 2.7},
		Boundary:    BoundaryParams{WLDriveResistance: 1.2, BLDriveResistance: 2.1},
	}
	oracle, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("oracle SolveDC: %v", err)
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 12; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for iter := 0; iter < 25; iter++ {
				got, err := NewTierBSolver().SolveDC(cloneSolveParamsE2E(params))
				if err != nil {
					t.Errorf("worker %d iter %d SolveDC: %v", worker, iter, err)
					return
				}
				if !solveResultEqualE2E(oracle.SolveResult, got.SolveResult, 1e-12) {
					t.Errorf("worker %d iter %d result drift", worker, iter)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
}

func conductanceFromPatternE2E(pattern [][]int) [][]float64 {
	out := make([][]float64, len(pattern))
	for r := range pattern {
		out[r] = make([]float64, len(pattern[r]))
		for c, level := range pattern[r] {
			out[r][c] = 10e-6 + float64(level)*3e-6
		}
	}
	return out
}

func wlVectorE2E(rows int, start float64) []float64 {
	out := make([]float64, rows)
	for r := range out {
		out[r] = start - 0.01*float64(r)
	}
	return out
}

func checkerMaskE2E(rows, cols int) [][]bool {
	mask := make([][]bool, rows)
	for r := range mask {
		mask[r] = make([]bool, cols)
		for c := range mask[r] {
			mask[r][c] = (r+c)%2 == 0
		}
	}
	return mask
}

func solveResultEqualE2E(a, b SolveResult, tol float64) bool {
	if len(a.CellCurrents) != len(b.CellCurrents) {
		return false
	}
	for r := range a.CellCurrents {
		if len(a.CellCurrents[r]) != len(b.CellCurrents[r]) {
			return false
		}
		for c := range a.CellCurrents[r] {
			if math.Abs(a.CellCurrents[r][c]-b.CellCurrents[r][c]) > tol || math.Abs(a.CellVoltages[r][c]-b.CellVoltages[r][c]) > tol {
				return false
			}
		}
	}
	return true
}

func minIntE2E(a, b int) int {
	if a < b {
		return a
	}
	return b
}
