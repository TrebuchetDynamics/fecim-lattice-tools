package arraysim

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestModule4ArraySimE2EWidePeripheralArrayWorkflow(t *testing.T) {
	cfg := ArrayConfig{
		Rows: 4, Cols: 4, ReadVoltageV: 0.2, CouplingMode: CouplingTierA,
		Wire:  WireParams{RWordLine: 0.45, RBitLine: 2.8},
		Sense: SenseChain{TIA: TIAConfig{Rf: 2e4, Vref: 0, Vmin: 0, Vmax: 1.2}, ADC: ADCConfig{Bits: 6, Vmin: 0, Vmax: 1.2}},
	}
	targets := [][]int{{0, 4, 8, 12}, {16, 20, 24, 29}, {3, 9, 15, 21}, {27, 18, 6, 1}}
	for _, order := range []string{"row-major", "col-major", "checkerboard"} {
		t.Run("program-"+order, func(t *testing.T) {
			result, err := ProgramArray(cfg, targets, ProgramOpts{Order: order, MaxPulses: 18, VerifyAfter: true, AccumDisturb: true})
			if err != nil {
				t.Fatalf("ProgramArray(%s): %v", order, err)
			}
			if len(result.Cells) != cfg.Rows || result.TotalPulses <= 0 || result.ProgramTimeNs <= 0 || result.TotalEnergy_fJ < 0 || result.WorstError > 2 || result.MaxDisturb <= 0 {
				t.Fatalf("program result invalid for %s: %+v", order, result)
			}
			for r := range result.Cells {
				for c, cell := range result.Cells[r] {
					if cell.Row != r || cell.Col != c || cell.TargetLevel != targets[r][c] || cell.FinalLevel < 0 || cell.FinalLevel > 29 || cell.PulsesUsed < 0 {
						t.Fatalf("cell[%d][%d] invalid: %+v", r, c, cell)
					}
				}
			}
			payload, err := json.Marshal(result)
			if err != nil || !strings.Contains(string(payload), "TotalEnergy_fJ") {
				t.Fatalf("program result JSON invalid: %v %s", err, payload)
			}
		})
	}

	conductance := [][]float64{{20e-6, 35e-6, 50e-6, 65e-6}, {80e-6, 30e-6, 45e-6, 60e-6}, {75e-6, 90e-6, 25e-6, 40e-6}, {55e-6, 70e-6, 85e-6, 100e-6}}
	wl := []float64{0.2, 0.1, 0.0, 0.2}
	bl := []float64{0, 0, 0, 0}
	for _, mode := range []CouplingMode{CouplingIdeal, CouplingTierA, CouplingTierB} {
		t.Run("solve-"+mode.String(), func(t *testing.T) {
			cfg.CouplingMode = mode
			res, ok := solveRead(cfg, SolveParams{WLVoltages: wl, BLVoltages: bl, Conductance: conductance, Geometry: DefaultCellGeometry(), Wire: cfg.Wire})
			if !ok || len(res.CellVoltages) != cfg.Rows || len(res.RowCurrents) != cfg.Rows || len(res.ColCurrents) != cfg.Cols {
				t.Fatalf("solveRead(%s) invalid: ok=%v result=%+v", mode.String(), ok, res)
			}
			for _, row := range res.CellCurrents {
				for _, current := range row {
					if math.IsNaN(current) || math.IsInf(current, 0) {
						t.Fatalf("solveRead(%s) current invalid: %g", mode.String(), current)
					}
				}
			}
		})
	}

	margin := ReadMarginAnalysis(cfg, 8)
	if margin.ArraySize != cfg.Rows || margin.Levels != 8 || len(margin.MarginPerLevel) != 7 || margin.MinMarginV < 0 || math.IsNaN(margin.MinMarginV) || margin.CouplingMode == "" {
		t.Fatalf("read margin invalid: %+v", margin)
	}
	invalidMargin := ReadMarginAnalysis(cfg, 1)
	if invalidMargin.Levels != 1 || len(invalidMargin.MarginPerLevel) != 0 || invalidMargin.Reliable {
		t.Fatalf("invalid-level margin contract changed: %+v", invalidMargin)
	}

	sense := cfg.Sense
	currents := []float64{-1e-3, 0, 10e-6, 40e-6, 1e-3}
	converted := sense.ConvertCurrents(currents)
	if len(converted) != len(currents) {
		t.Fatalf("ConvertCurrents len=%d", len(converted))
	}
	for i, got := range converted {
		if got.Code < 0 || got.Code >= 64 || got.Vout < sense.TIA.Vmin || got.Vout > sense.TIA.Vmax {
			t.Fatalf("converted[%d] invalid: %+v", i, got)
		}
	}
	if lo, hi := sense.CurrentRange(); lo >= hi || sense.CurrentLSB() <= 0 {
		t.Fatalf("sense range invalid: lo=%g hi=%g lsb=%g", lo, hi, sense.CurrentLSB())
	}
}

func TestModule4ArraySimE2EDesignBenchmarkScheduleAndSpiceWorkflow(t *testing.T) {
	points := BuildDesignSpacePoints([]int{0, 8, 16}, []int{0, 4, 6}, []string{"FeFET", "RRAM", "PCM", "unknown"})
	if len(points) != 16 {
		t.Fatalf("design sweep produced %d points, want 16", len(points))
	}
	front := ParetoFront(points)
	if len(front) == 0 || len(front) > len(points) {
		t.Fatalf("pareto front invalid: len=%d", len(front))
	}
	var csv bytes.Buffer
	if err := ExportParetoCSV(front, &csv); err != nil {
		t.Fatalf("ExportParetoCSV: %v", err)
	}
	if !strings.Contains(csv.String(), "array_size,adc_bits,device,latency_ns,energy_pj,accuracy") || !strings.Contains(csv.String(), "FeFET") {
		t.Fatalf("pareto CSV missing expected fields: %s", csv.String())
	}

	bench := RunBatchMNISTBenchmark([]BenchmarkCase{{Name: "z", ArraySize: 16, ADCBits: 6, Device: "FeFET"}, {Name: "a", ArraySize: 8, ADCBits: 4, Device: "RRAM"}})
	if len(bench) != 2 || bench[0].Name != "a" || bench[1].Name != "z" || bench[0].Accuracy <= 0 || bench[0].Throughput <= 0 {
		t.Fatalf("batch benchmark invalid/sort changed: %+v", bench)
	}

	var rowMajorDisturb, checkerDisturb float64
	for _, mode := range []ProgramOrderMode{ProgramOrderRowMajor, ProgramOrderCheckerboard, ProgramOrderSerpentine, ProgramOrderAdaptive} {
		sched, err := GenerateProgramSchedule(5, 4, mode, 0.2)
		if err != nil {
			t.Fatalf("GenerateProgramSchedule(%s): %v", mode, err)
		}
		if sched.Rows != 5 || sched.Cols != 4 || len(sched.Order) != 20 || sched.CumulativeDisturb < 0 {
			t.Fatalf("schedule invalid for %s: %+v", mode, sched)
		}
		if mode == ProgramOrderRowMajor {
			rowMajorDisturb = sched.CumulativeDisturb
		}
		if mode == ProgramOrderCheckerboard {
			checkerDisturb = sched.CumulativeDisturb
		}
	}
	if checkerDisturb == rowMajorDisturb || checkerDisturb < 0 || rowMajorDisturb < 0 {
		t.Fatalf("schedule disturb scores should be finite and order-sensitive: row=%g checker=%g", rowMajorDisturb, checkerDisturb)
	}
	for _, bad := range []struct {
		rows, cols int
		mode       ProgramOrderMode
	}{{0, 4, ProgramOrderRowMajor}, {4, 0, ProgramOrderRowMajor}, {4, 4, "bad"}} {
		if _, err := GenerateProgramSchedule(bad.rows, bad.cols, bad.mode, 0); err == nil {
			t.Fatalf("GenerateProgramSchedule(%+v) expected error", bad)
		}
	}

	cfg := ArrayConfig{Rows: 2, Cols: 3, ReadVoltageV: 0.2, CouplingMode: CouplingTierA}
	params := SolveParams{WLVoltages: []float64{0.2, 0}, BLVoltages: []float64{0, 0, 0}, Conductance: [][]float64{{20e-6, 40e-6, 60e-6}, {30e-6, 50e-6, 70e-6}}, Geometry: DefaultCellGeometry()}
	deck, err := ExportCrossbarSPICE(params, SpiceExportConfig{Title: "Module4 E2E deck"})
	if err != nil {
		t.Fatalf("ExportCrossbarSPICE: %v", err)
	}
	for _, marker := range []string{"Module4 E2E deck", ".param RWL", "XDAC_WL_0", "XADC_2", "RCELL_1_2", ".end"} {
		if !strings.Contains(deck, marker) {
			t.Fatalf("SPICE deck missing %q:\n%s", marker, deck)
		}
	}
	if _, err := ExportCrossbarSPICE(SolveParams{}, SpiceExportConfig{}); err == nil {
		t.Fatal("empty SPICE export should fail")
	}
	if _, err := ProgramArray(cfg, [][]int{{1, 2}}, ProgramOpts{}); err == nil {
		t.Fatal("ProgramArray jagged/mismatched targets should fail")
	}
}

func TestModule4ArraySimE2EWriteVerifyVariationBoundaryWorkflow(t *testing.T) {
	model := LinearProgramModel{Gain: 20e-6, Gmin: 0, Gmax: 90e-6}
	ok := RunWriteVerifyLoop(10e-6, model, WriteVerifyConfig{TargetG: 30e-6, Tolerance: 2e-6, StartVoltage: 1.0, StepVoltage: 0.2, MaxIters: 20})
	if !ok.Converged || ok.Iterations <= 0 || math.Abs(ok.FinalG-30e-6) > 2e-6 {
		t.Fatalf("write-verify should converge: %+v", ok)
	}
	fail := RunWriteVerifyLoop(10e-6, model, WriteVerifyConfig{TargetG: 200e-6, Tolerance: 1e-9, StartVoltage: 1.0, StepVoltage: 0.1, MaxIters: 5})
	if fail.Converged || fail.Iterations != 5 || fail.FinalG > model.Gmax {
		t.Fatalf("write-verify failure boundary invalid: %+v", fail)
	}

	mc := RunProcessVariationMC(ProcessVariationConfig{NominalEc: 1.0e8, NominalPr: 0.3, VariationFraction: 0.15, Samples: 64, Seed: 7, MinReadMarginRatio: 0.8})
	if mc.MeanEc <= 0 || mc.MeanPr <= 0 || mc.StdEc <= 0 || mc.StdPr <= 0 || mc.Yield <= 0 || mc.Yield > 1 || mc.PassSamples <= 0 || mc.PassSamples > 64 {
		t.Fatalf("process variation invalid: %+v", mc)
	}
	bounded := RunProcessVariationMC(ProcessVariationConfig{NominalEc: 1.0e8, NominalPr: 0.3, VariationFraction: -1, Samples: 0, Seed: 1, MinReadMarginRatio: 0})
	if bounded.StdEc != 0 || bounded.StdPr != 0 || bounded.PassSamples != 1 || bounded.Yield != 1 {
		t.Fatalf("process variation defaults/bounds invalid: %+v", bounded)
	}
}
