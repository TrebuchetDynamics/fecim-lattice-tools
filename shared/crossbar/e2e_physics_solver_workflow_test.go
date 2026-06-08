package crossbar

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestModule2CrossbarE2EParasiticNonlinearThermalWorkflow(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 4, Cols: 4, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray error = %v", err)
	}
	defer arr.Destroy()
	weights := [][]float64{{0.1, 0.3, 0.5, 0.7}, {0.9, 0.2, 0.4, 0.6}, {0.8, 1.0, 0.0, 0.25}, {0.33, 0.66, 0.99, 0.12}}
	if err := arr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix error = %v", err)
	}
	input := []float64{0.2, 0.4, 0.6, 0.8}

	wire := DefaultWireParams()
	for _, node := range []float64{45, 28, 14, 7} {
		t.Run("node", func(t *testing.T) {
			scaled := wire.ScaleForTechNode(45, node)
			rc := arr.AnalyzeRCDelay(scaled, 0.2)
			if len(rc.WordLineDelay) != arr.Rows() || len(rc.BitLineDelay) != arr.Cols() || rc.TotalArrayDelay <= 0 || rc.MaxFrequency <= 0 || rc.ChargingEnergy <= 0 {
				t.Fatalf("RC delay invalid for node %.0f: %+v", node, rc)
			}
			if scaled.GetRCTimeConstant(arr.Cols(), true) <= 0 || scaled.GetPropagationDelay(arr.Cols(), true) <= 0 || scaled.GetSettlingTime(arr.Cols(), true, 0) <= 0 {
				t.Fatalf("wire helper invalid for node %.0f", node)
			}
		})
	}

	iter := arr.AnalyzeIRDropIterative(input, wire, &IRDropSolverConfig{MaxIterations: 30, Tolerance: 1e-8, Damping: 0.6})
	if len(iter.EffectiveVoltage) != arr.Rows() || len(iter.EffectiveVoltage[0]) != arr.Cols() || iter.MaxIRDrop < 0 || iter.AvgIRDrop < 0 || iter.WorstCaseCell[0] < 0 {
		t.Fatalf("iterative IR drop invalid: %+v", iter)
	}

	iv := &FeFETIVParams{SubthSlope: 1.2, TempK: 300}
	if iv.ThermalVoltage() <= 0 || iv.VSat() <= 0 || iv.Current(1e-6, 0) != 0 || iv.LinearityError(0) != 0 || iv.LinearityError(0.2) <= 0 {
		t.Fatalf("IV helper invalid: Vt=%g Vsat=%g", iv.ThermalVoltage(), iv.VSat())
	}
	nonlinear, err := arr.MVMNonLinear(input, iv)
	if err != nil {
		t.Fatalf("MVMNonLinear error = %v", err)
	}
	assertE2EVectorFinite01(t, "nonlinear", nonlinear, arr.Rows())
	if _, err := arr.MVMNonLinear(input[:2], iv); err == nil || !strings.Contains(err.Error(), "input length") {
		t.Fatalf("MVMNonLinear wrong input error = %v", err)
	}
	fecap, err := NewArray(&Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8, CellType: CellTypeFeCAP})
	if err != nil {
		t.Fatalf("NewArray FeCAP error = %v", err)
	}
	defer fecap.Destroy()
	if _, err := fecap.MVMNonLinear([]float64{1, 1}, nil); err == nil || !strings.Contains(err.Error(), "FeCAP") {
		t.Fatalf("MVMNonLinear FeCAP error = %v", err)
	}

	solver, err := NewParasiticSolver(4, 4, &SORConfig{MaxIterations: 200, Tolerance: 1e-9, OmegaInitial: 1.0, OmegaMin: 0.01, OmegaDecay: 0.9, AdaptiveOmega: true})
	if err != nil {
		t.Fatalf("NewParasiticSolver error = %v", err)
	}
	solver.SetConductances(arr.GetConductanceMatrix())
	for _, parasitic := range []struct{ row, col float64 }{{0, 0}, {0.001, 0.001}, {0.01, 0.02}} {
		solver.SetParasitics(parasitic.row, parasitic.col)
		res, err := solver.SolveMVM(input)
		if err != nil {
			t.Fatalf("SolveMVM parasitic %+v error = %v", parasitic, err)
		}
		if !res.Converged || res.Iterations <= 0 || len(res.OutputCurrents) != 4 || len(res.DeviceCurrents) != 4 || len(res.DeviceVoltages) != 4 || res.FinalOmega <= 0 {
			t.Fatalf("parasitic result invalid for %+v: %+v", parasitic, res)
		}
	}
	if _, err := solver.SolveMVM([]float64{1}); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("SolveMVM invalid input error = %v", err)
	}

	thermalCfg := ThermalConfig{AmbientTempK: 300, ThermalResistance: 25, ThermalCapacitance: 2e-6, SubstrateTempK: 300}
	tm := NewThermalModel(4, 4, thermalCfg)
	power := tm.PowerFromMVM(arr, input)
	if len(power) != 4 || len(power[0]) != 4 || tm.PowerFromMVM(nil, input) != nil {
		t.Fatalf("PowerFromMVM shape/nil invalid: %v", power)
	}
	steady := tm.ComputeSteadyState(power)
	if steady.PeakTempK < thermalCfg.AmbientTempK || steady.AvgTempK < thermalCfg.AmbientTempK || steady.PowerDensityWm2 <= 0 {
		t.Fatalf("steady thermal invalid: %+v", steady)
	}
	transient := tm.ComputeTransient(power, 1e-7, 6)
	if len(transient) != 6 || transient[len(transient)-1].PeakTempK < transient[0].PeakTempK {
		t.Fatalf("transient thermal invalid: %+v", transient)
	}
	if tm.TimeConstant() != thermalCfg.ThermalResistance*thermalCfg.ThermalCapacitance {
		t.Fatalf("thermal time constant changed")
	}
	if SteadyStateTemp(1e-3, thermalCfg) <= thermalCfg.AmbientTempK || MaxAllowedPower(thermalCfg.AmbientTempK-1, thermalCfg) != 0 || TransientTemp(1e-3, tm.TimeConstant(), thermalCfg) <= thermalCfg.AmbientTempK {
		t.Fatalf("thermal helper contract changed")
	}
}

func TestModule2CrossbarE2EParasiticSolverInvalidAndNonConvergence(t *testing.T) {
	if solver, err := NewParasiticSolver(0, 4, nil); !errors.Is(err, ErrInvalidConfiguration) || solver != nil {
		t.Fatalf("NewParasiticSolver invalid rows = solver %v err %v", solver, err)
	}
	solver, err := NewParasiticSolver(2, 2, &SORConfig{MaxIterations: 1, Tolerance: 1e-30, OmegaInitial: 1.9, OmegaMin: 0.01, OmegaDecay: 0.5, AdaptiveOmega: true})
	if err != nil {
		t.Fatalf("NewParasiticSolver constrained error = %v", err)
	}
	solver.SetConductances([][]float64{{1, 1}, {1, 1}})
	solver.SetParasitics(10, 10)
	res, err := solver.SolveMVM([]float64{1, 1})
	if err == nil {
		t.Fatalf("expected constrained solver to fail, got result %+v", res)
	}
	if !errors.Is(err, ErrConvergenceFailed) && !errors.Is(err, ErrMaxIterationsExceeded) {
		t.Fatalf("unexpected constrained solver error: %v", err)
	}

	acceptedIV := &FeFETIVParams{SubthSlope: -1, TempK: 300}
	arr, err := NewArray(&Config{Rows: 1, Cols: 1, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	defer arr.Destroy()
	ivOut, err := arr.MVMNonLinear([]float64{1}, acceptedIV)
	if err != nil || len(ivOut) != 1 || math.IsNaN(ivOut[0]) {
		t.Fatalf("MVMNonLinear accepted-IV boundary output=%v err=%v", ivOut, err)
	}

	zeroThermal := NewThermalModel(0, 0, DefaultThermalConfig())
	if zeroThermal.TimeConstant() <= 0 || len(zeroThermal.ComputeTransient(nil, -1, 0)) != 0 {
		t.Fatalf("thermal zero/boundary contract changed")
	}
	if math.IsNaN(TransientTemp(1e-3, 1, ThermalConfig{AmbientTempK: 300, ThermalResistance: 25, ThermalCapacitance: 0})) {
		t.Fatalf("TransientTemp zero capacitance returned NaN")
	}
}
