package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

func TestFocus54_ConductanceModelsAnd30LevelQuantization(t *testing.T) {
	cfg := &Config{Rows: 1, Cols: 1, ADCBits: 8, DACBits: 8, ConductanceModel: ConductanceLinear}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// 30-level quantization should expose exactly 30 distinct levels.
	seen := map[float64]bool{}
	for i := 0; i <= 1000; i++ {
		q := QuantizeToLevels(float64(i) / 1000.0)
		seen[q] = true
	}
	if len(seen) != DefaultQuantizationLevels {
		t.Fatalf("quantization levels = %d, want %d", len(seen), DefaultQuantizationLevels)
	}

	// Linear and exponential endpoints should map to physical GMin/GMax.
	arr.SetConductanceModel(ConductanceLinear)
	if got := arr.GetPhysicalConductance(0.0); math.Abs(got-GMin) > 1e-15 {
		t.Fatalf("linear g(0) = %e, want %e", got, GMin)
	}
	if got := arr.GetPhysicalConductance(1.0); math.Abs(got-GMax) > 1e-15 {
		t.Fatalf("linear g(1) = %e, want %e", got, GMax)
	}

	arr.SetConductanceModel(ConductanceExponential)
	if got := arr.GetPhysicalConductance(0.0); math.Abs(got-GMin) > 1e-15 {
		t.Fatalf("exp g(0) = %e, want %e", got, GMin)
	}
	if got := arr.GetPhysicalConductance(1.0); math.Abs(got-GMax) > 1e-15 {
		t.Fatalf("exp g(1) = %e, want %e", got, GMax)
	}

	// Lookup should return exact table values at each 30-level point.
	table := make([]float64, DefaultQuantizationLevels)
	for i := range table {
		table[i] = GMin + (GMax-GMin)*float64(i)/float64(DefaultQuantizationLevels-1)
	}
	if err := arr.SetConductanceTable(table); err != nil {
		t.Fatalf("SetConductanceTable failed: %v", err)
	}
	arr.SetConductanceModel(ConductanceLookup)
	for level := 0; level < DefaultQuantizationLevels; level++ {
		gNorm := float64(level) / float64(DefaultQuantizationLevels-1)
		if got := arr.GetPhysicalConductance(gNorm); math.Abs(got-table[level]) > 1e-15 {
			t.Fatalf("lookup level %d = %e, want %e", level, got, table[level])
		}
	}
}

func TestFocus55_MVMVMMOhmsLawAndNormalization(t *testing.T) {
	cfg := &Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8, NoiseLevel: 0}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Program deterministic matrix:
	// [1.0 0.5]
	// [0.0 1.0]
	_ = arr.ProgramWeight(0, 0, 1.0)
	_ = arr.ProgramWeight(0, 1, 0.5)
	_ = arr.ProgramWeight(1, 0, 0.0)
	_ = arr.ProgramWeight(1, 1, 1.0)

	input := []float64{1.0, 0.5}
	mvm, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	v0 := arr.quantizeDAC(input[0])
	v1 := arr.quantizeDAC(input[1])
	g := arr.GetConductanceMatrix()
	expected0 := arr.quantizeADC((g[0][0]*v0 + g[0][1]*v1) / 2.0)
	expected1 := arr.quantizeADC((g[1][0]*v0 + g[1][1]*v1) / 2.0)
	if math.Abs(mvm[0]-expected0) > 1e-12 || math.Abs(mvm[1]-expected1) > 1e-12 {
		t.Fatalf("MVM = %v, want [%v %v]", mvm, expected0, expected1)
	}

	vmmInput := []float64{1.0, 0.5}
	vmm, err := arr.VMM(vmmInput)
	if err != nil {
		t.Fatalf("VMM failed: %v", err)
	}
	vi0 := arr.quantizeDAC(vmmInput[0])
	vi1 := arr.quantizeDAC(vmmInput[1])
	expectedCol0 := arr.quantizeADC((g[0][0]*vi0 + g[1][0]*vi1) / 2.0)
	expectedCol1 := arr.quantizeADC((g[0][1]*vi0 + g[1][1]*vi1) / 2.0)
	if math.Abs(vmm[0]-expectedCol0) > 1e-12 || math.Abs(vmm[1]-expectedCol1) > 1e-12 {
		t.Fatalf("VMM = %v, want [%v %v]", vmm, expectedCol0, expectedCol1)
	}
}

func TestFocus56_IRDropIterativeSolverConsistency(t *testing.T) {
	cfg := &Config{Rows: 4, Cols: 4, ADCBits: 8, DACBits: 8}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			_ = arr.ProgramWeight(r, c, 1.0)
		}
	}
	input := []float64{1, 1, 1, 1}
	params := DefaultWireParams()

	a1 := arr.AnalyzeIRDrop(input, params)
	a2 := arr.AnalyzeIRDropIterative(input, params, &IRDropSolverConfig{MaxIterations: 200, Tolerance: 1e-9, Damping: 0.5})

	if math.Abs(a1.MaxIRDrop-a2.MaxIRDrop) > 5e-3 {
		t.Fatalf("max IR drop mismatch: direct=%g iterative=%g", a1.MaxIRDrop, a2.MaxIRDrop)
	}
	if a2.MaxIRDrop <= 0 {
		t.Fatalf("expected non-zero IR drop")
	}

	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			v := a2.EffectiveVoltage[r][c]
			if v < 0 || v > 1 {
				t.Fatalf("effective voltage out of range at (%d,%d): %g", r, c, v)
			}
		}
	}
}

func TestFocus57_SneakPathThreeCellAndSNRMath(t *testing.T) {
	sp := NewSneakPathAnalyzer(2, 2)
	voltage := 1.0
	sp.AnalyzeTarget(0, 0, voltage)

	if len(sp.SneakPaths) != 1 {
		t.Fatalf("paths = %d, want 1", len(sp.SneakPaths))
	}
	if sp.SneakPaths[0].PathLength != 3 {
		t.Fatalf("path length = %d, want 3", sp.SneakPaths[0].PathLength)
	}

	stats := sp.GetStats(voltage)
	expectedSNR := 20 * math.Log10(stats.TargetCurrent/stats.TotalSneakCurrent)
	if math.Abs(stats.SignalToNoiseRatio-expectedSNR) > 1e-12 {
		t.Fatalf("SNR = %.12f dB, want %.12f dB", stats.SignalToNoiseRatio, expectedSNR)
	}
}

func TestFocus58_DriftTemperatureArrheniusAndVariation(t *testing.T) {
	rand.Seed(42)
	sim300 := NewDriftSimulator(4, 4, 30)
	sim300.Temperature = 300
	sim300.SimulateTimeStep(1e5)
	drift300 := sim300.GetStats().AvgDrift

	rand.Seed(42)
	sim400 := NewDriftSimulator(4, 4, 30)
	sim400.Temperature = 400
	sim400.SimulateTimeStep(1e5)
	drift400 := sim400.GetStats().AvgDrift

	if math.Abs(drift400) <= math.Abs(drift300) {
		t.Fatalf("expected |drift| at 400K (%.3e) > 300K (%.3e)", drift400, drift300)
	}
}

func TestFocus59_EnduranceAndHalfSelectDisturb(t *testing.T) {
	cfg := &Config{Rows: 5, Cols: 5, ADCBits: 8, DACBits: 8, ConductanceModel: ConductanceLinear,
		Endurance:  &EnduranceConfig{Enabled: true, FatigueThreshold: 10, FailureThreshold: 100},
		HalfSelect: &HalfSelectConfig{Enabled: true, DisturbThreshold: 0.3, DisturbRate: 0.01},
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	_ = arr.ProgramWeight(2, 2, 1.0)
	baseline := arr.GetPhysicalConductanceForCell(2, 2)
	for i := 0; i < 150; i++ {
		_ = arr.ProgramWeight(2, 2, 1.0)
	}
	aged := arr.GetPhysicalConductanceForCell(2, 2)
	if aged >= baseline {
		t.Fatalf("expected fatigue-degraded conductance: aged=%e baseline=%e", aged, baseline)
	}

	if err := arr.ProgramWeightWithDisturb(2, 2, 0.6, true); err != nil {
		t.Fatalf("ProgramWeightWithDisturb failed: %v", err)
	}
	// Same row (4 cells) + same col (4 cells) should be half-selected.
	halfSelected := int64(0)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i == 2 && j == 2 {
				continue
			}
			if arr.cells[i][j].HalfSelectCount > 0 {
				halfSelected++
			}
		}
	}
	if halfSelected != 8 {
		t.Fatalf("half-selected cells = %d, want 8", halfSelected)
	}
}

func TestFocus60_MVMWithNonIdealitiesPipelineOrdering(t *testing.T) {
	cfg := &Config{Rows: 8, Cols: 8, ADCBits: 8, DACBits: 8}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			_ = arr.ProgramWeight(i, j, 0.5)
		}
	}
	degradation, err := arr.ComputeAccuracyDegradation([]float64{1, 1, 1, 1, 1, 1, 1, 1}, 100)
	if err != nil {
		t.Fatalf("ComputeAccuracyDegradation failed: %v", err)
	}
	got := []string{}
	for _, s := range degradation.Degradations {
		got = append(got, s.Source)
	}
	want := []string{"ADC/DAC Quantization", "IR Drop", "Device Variation", "Sneak Paths"}
	if len(got) != len(want) {
		t.Fatalf("pipeline step count = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pipeline[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
