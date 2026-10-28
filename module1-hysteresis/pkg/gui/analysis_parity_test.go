package gui

import (
	"math"
	"reflect"
	"testing"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/workbench/hysteresis"
)

// Native GUI constructions are the oracle, not the headless TanhEverett FORC.
func TestAnalysisGUIParity(t *testing.T) {
	cfg := hysteresis.DefaultConfig()
	cfg.Model = hysteresis.Landau
	cfg.Temperature = 450
	session, err := hysteresis.NewSession(cfg)
	if err != nil {
		t.Fatal(err)
	}
	input := session.Snapshot()
	for _, kind := range []hysteresis.AnalysisKind{hysteresis.AnalysisPUND, hysteresis.AnalysisFORC} {
		req := hysteresis.AnalysisRequest{Input: input, Kind: kind}
		if kind == hysteresis.AnalysisFORC {
			req.FORC = hysteresis.FORCParameters{EmaxVm: 1, Reversals: 21, Resolution: 31}
		}
		j, err := hysteresis.NewAnalysis(req)
		if err != nil {
			t.Fatal(err)
		}
		for {
			p, err := j.Advance(hysteresis.AnalysisBudget{Steps: 1})
			if err != nil {
				t.Fatal(err)
			}
			if p.State == "complete" {
				break
			}
		}
		r, err := j.Result()
		if err != nil {
			t.Fatal(err)
		}
		if kind == hysteresis.AnalysisPUND {
			// The actual numerical helper called by PUNDPanel.runPUND.
			charges, traces, err := runPUNDDiagnostic()
			if err != nil || r.PUND.Charges != charges || !reflect.DeepEqual(r.PUND.Traces, traces) {
				t.Fatal("GUI PUND differs", err)
			}
		} else {
			want, err := runFORCWorkflow(input.Material.Parameters.Ps, 1, 21, 31)
			if err != nil || !reflect.DeepEqual(r.FORC.Sweep, want.Sweep) || !reflect.DeepEqual(r.FORC.Matrix.EaAxis, want.Matrix.EaAxis) || !reflect.DeepEqual(r.FORC.Matrix.EbAxis, want.Matrix.EbAxis) || !reflect.DeepEqual(r.FORC.Matrix.Density, want.Matrix.Density) || r.FORC.Matrix.Rows != want.Matrix.Rows || r.FORC.Matrix.Cols != want.Matrix.Cols {
				t.Fatal("GUI FORC differs", err)
			}
		}
	}
	for _, ab := range [][2]float64{{1, -1}, {0.5, -0.2}, {4, -4}, {-1, 1}, {1, 1}} {
		alpha, beta := ab[0], ab[1]
		want := 0.0
		if alpha > beta {
			x := math.Max(0, math.Min(1, (alpha-beta)/2))
			center := (alpha + beta) / 2
			want = 0.25 * x * math.Exp(-(center*center)/(2*0.2*0.2))
		}
		if got := (guiFORCEverett{satE: 1, ps: 0.25}).Calculate(alpha, beta); got != want {
			t.Fatalf("GUI Everett characterization %v: %g != %g", ab, got, want)
		}
		if got := (physics.GUIFORCEverett{SaturationE: 1, Ps: 0.25}).Calculate(alpha, beta); got != want {
			t.Fatalf("extracted Everett %v: %g != %g", ab, got, want)
		}
	}
}
