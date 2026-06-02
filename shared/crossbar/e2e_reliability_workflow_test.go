package crossbar

import (
	"math"
	"testing"
)

func TestModule2CrossbarE2EReliabilityDriftTemperatureUncertaintyWorkflow(t *testing.T) {
	sim := NewDriftSimulatorWithModel(4, 5, 30, DriftModelLiterature)
	sim.DriftNoiseSigma = 0
	sim.Temperature = TempIndustrial
	for r := 0; r < sim.Rows; r++ {
		for c := 0; c < sim.Cols; c++ {
			sim.SetConductanceLevel(r, c, (r*sim.Cols+c)%sim.Levels)
		}
	}
	info := sim.GetDriftModelInfo()
	if info.ModelName == "" || info.Coefficient != FeFETDriftCoefficients.Literature || !info.IsAssumed {
		t.Fatalf("literature drift info invalid: %+v", info)
	}
	levelsBefore := make([][]int, sim.Rows)
	for r := range levelsBefore {
		levelsBefore[r] = make([]int, sim.Cols)
		for c := range levelsBefore[r] {
			levelsBefore[r][c] = sim.GetCurrentLevel(r, c)
		}
	}
	for _, dt := range []float64{1, 60, 3600, 24 * 3600} {
		sim.SimulateTimeStep(dt)
		sim.RecordSnapshot()
	}
	if len(sim.DriftHistory) != 4 {
		t.Fatalf("drift history length = %d, want 4", len(sim.DriftHistory))
	}
	stats := sim.GetStats()
	if stats.ElapsedTime <= 0 || stats.RetentionPrediction <= 0 || stats.RetentionPrediction > 100 || stats.TechnologyComparison.FeFETAdvantage <= 0 {
		t.Fatalf("drift stats invalid: %+v", stats)
	}
	for _, snap := range sim.DriftHistory {
		if snap.Time <= 0 || snap.WorstCellRow < 0 || snap.WorstCellCol < 0 || len(snap.Conductances) != sim.Cols || math.IsNaN(snap.AvgDrift) || math.IsNaN(snap.MaxDrift) {
			t.Fatalf("invalid drift snapshot: %+v", snap)
		}
	}
	sim.RefreshCell(0, 0)
	if sim.GetCurrentLevel(-1, 0) != 0 || sim.GetCurrentLevel(99, 99) != 0 {
		t.Fatalf("invalid current-level boundary changed")
	}
	sim.RefreshAll()
	sim.Reset()
	if sim.Time != 0 || len(sim.DriftHistory) != 0 {
		t.Fatalf("drift reset left time/history: time=%g history=%d", sim.Time, len(sim.DriftHistory))
	}
	for r := range levelsBefore {
		for c := range levelsBefore[r] {
			if sim.GetCurrentLevel(r, c) != levelsBefore[r][c] {
				t.Fatalf("reset changed level[%d][%d] = %d, want %d", r, c, sim.GetCurrentLevel(r, c), levelsBefore[r][c])
			}
		}
	}

	tech := CompareTechnologies(3, 3, 3600)
	for _, name := range []string{"FeCIM (FeFET)", "RRAM", "PCM", "Flash"} {
		if _, ok := tech[name]; !ok {
			t.Fatalf("CompareTechnologies missing %q: %#v", name, tech)
		}
	}

	for _, temp := range []float64{TempColdSpace, TempCryogenic, TempRoom, TempIndustrial, TempAutomotive, 500} {
		t.Run(NewTemperatureEffects(temp).GetTemperatureLabel(), func(t *testing.T) {
			te := NewTemperatureEffects(temp)
			params := te.GetAdjustedParams()
			if math.IsNaN(params.WireResistanceFactor) || params.GminAdjusted <= 0 || params.GmaxAdjusted <= params.GminAdjusted || math.IsNaN(params.DriftRateFactor) || params.NoiseFactor <= 0 || math.IsNaN(params.RetentionFactor) || params.RetentionFactor <= 0 {
				t.Fatalf("temperature params invalid for %.1fK: %+v", temp, params)
			}
			if te.AdjustedSwitchingEnergy(10) <= 0 || te.GetTemperatureLabel() == "" {
				t.Fatalf("temperature helper invalid for %.1fK", temp)
			}
		})
	}

	arr, err := NewArray(&Config{Rows: 3, Cols: 3, NoiseLevel: 0.04, ADCBits: 6, DACBits: 6, ProcessVariation: &ProcessVariationConfig{DeviceSigma: 0.03}})
	if err != nil {
		t.Fatalf("NewArray uncertainty: %v", err)
	}
	defer arr.Destroy()
	if err := arr.ProgramWeightMatrix([][]float64{{1, 1, 1}, {0.5, 0.25, 0.75}, {0, 0.5, 1}}); err != nil {
		t.Fatalf("ProgramWeightMatrix uncertainty: %v", err)
	}
	unc, err := arr.MVMWithUncertainty([]float64{1, 1, 1})
	if err != nil {
		t.Fatalf("MVMWithUncertainty error = %v", err)
	}
	if len(unc.Output) != 3 || len(unc.Uncertainty) != 3 || unc.Saturated < 0 || unc.Saturated > 3 {
		t.Fatalf("uncertainty contract invalid: %+v", unc)
	}
	for i, sigma := range unc.Uncertainty {
		if sigma < 0 || math.IsNaN(sigma) {
			t.Fatalf("uncertainty[%d] invalid: %g", i, sigma)
		}
	}
	if _, err := arr.MVMWithUncertainty([]float64{}); err == nil {
		t.Fatal("MVMWithUncertainty empty input returned nil error")
	}
}

func TestModule2CrossbarE2EWriteDisturbArchitectureWorkflow(t *testing.T) {
	cfg := &WriteDisturbConfig{Enable: true, HalfSelectRatio: 0.5, StressAccumulationRate: 0.4, StressThreshold: 1.0, Architecture1T1R: false, Architecture1T1RReduction: 0.1}
	engine := NewWriteDisturbEngine(4, 4, cfg)
	engine.RecordWrite(1, 1)
	engine.RecordBatchWrite([][2]int{{1, 2}, {3, 1}})
	engine.RecordWrite(1, 1)
	stats := engine.GetStressStats()
	if stats.TotalWriteOps != 4 || stats.TotalHalfSelects != 24 || stats.CellsAtRisk == 0 {
		t.Fatalf("write disturb stats after records invalid: %+v", stats)
	}
	stress := engine.GetStressMatrix()
	if len(stress) != 4 || len(stress[0]) != 4 || engine.GetCellStress(1, 0) <= 0 || engine.GetCellStress(-1, 0) != 0 || engine.GetCellStress(9, 9) != 0 {
		t.Fatalf("stress matrix/bounds invalid: stress=%v", stress)
	}
	conductances := [][]float64{
		{GMin, GMin, GMax, GMax},
		{GMax, GMax, GMin, GMin},
		{GMin, GMax, GMin, GMax},
		{GMax, GMin, GMax, GMin},
	}
	disturbed := engine.ApplyDisturbEffects(conductances, 30)
	if disturbed == 0 {
		t.Fatalf("expected at least one disturbed cell, stats=%+v stress=%v", engine.GetStressStats(), engine.GetStressMatrix())
	}
	stats = engine.GetStressStats()
	if stats.DisturbedCells != disturbed || stats.CellsAtRisk < 0 {
		t.Fatalf("disturb stats after apply invalid: disturbed=%d stats=%+v", disturbed, stats)
	}

	passiveRate, activeRate := CompareArchitectures(10_000)
	if passiveRate <= activeRate || passiveRate <= 0 || activeRate <= 0 {
		t.Fatalf("architecture comparison invalid: passive=%g active=%g", passiveRate, activeRate)
	}
	if EstimateDisturbRate(100, nil) != 0 {
		t.Fatalf("default disabled disturb rate should be zero")
	}
	if HalfSelectVoltage(3, "V/2") != 1.5 || HalfSelectVoltage(3, "V/3") != 1 || HalfSelectVoltage(3, "floating") != 1.2000000000000002 {
		t.Fatalf("half-select voltage helpers changed")
	}
	if !IsDisturbCritical(1.0, 1.1, 0.2) || IsDisturbCritical(0.5, 1.1, 0.2) {
		t.Fatalf("disturb critical helper changed")
	}

	engine.Reset()
	if reset := engine.GetStressStats(); reset.TotalWriteOps != 0 || reset.TotalHalfSelects != 0 || reset.MaxStress != 0 {
		t.Fatalf("reset stats invalid: %+v", reset)
	}
	engine.Resize(2, 3)
	engine.RecordWrite(0, 1)
	resized := engine.GetStressMatrix()
	if len(resized) != 2 || len(resized[0]) != 3 || engine.GetStressStats().TotalHalfSelects != 3 {
		t.Fatalf("resize/record contract invalid: stress=%v stats=%+v", resized, engine.GetStressStats())
	}

	disabled := NewWriteDisturbEngine(2, 2, &WriteDisturbConfig{Enable: false})
	disabled.RecordWrite(0, 0)
	if disabled.GetStressStats().TotalWriteOps != 0 || disabled.ApplyDisturbEffects(conductances, 30) != 0 {
		t.Fatalf("disabled engine should be inert: %+v", disabled.GetStressStats())
	}
}
