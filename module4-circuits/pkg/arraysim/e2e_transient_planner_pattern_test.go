package arraysim

import (
	"math"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestModule4ArraySimE2ETransientPlannerPatternWorkflow(t *testing.T) {
	cfg := ArrayConfig{
		Rows: 2, Cols: 3, ReadVoltageV: 0.2, CouplingMode: CouplingTierA,
		Sense:    SenseChain{TIA: TIAConfig{Rf: 2e4, Vref: 0, Vmin: 0, Vmax: 1.2}, ADC: ADCConfig{Bits: 6, Vmin: 0, Vmax: 1.2}},
		Material: sharedphysics.DefaultHZO(),
	}
	waveforms := map[string][]PulseStep{
		"write_ramp_read_tail": {{Voltage: 2.8, DurationNs: 20, RiseTimeNs: 5}, {Voltage: 0.2, DurationNs: 10, RiseTimeNs: 0}},
		"subcoercive_read":     {{Voltage: 0.15, DurationNs: 12, RiseTimeNs: 2}},
		"ignored_zero_step":    {{Voltage: 3.0, DurationNs: 0, RiseTimeNs: 0}, {Voltage: 1.5, DurationNs: 5, RiseTimeNs: 1}},
	}
	for name, waveform := range waveforms {
		t.Run(name, func(t *testing.T) {
			results := TransientSolve(cfg, waveform, 0.5)
			if len(results) != cfg.Rows*cfg.Cols {
				t.Fatalf("TransientSolve result count=%d, want %d", len(results), cfg.Rows*cfg.Cols)
			}
			for i, res := range results {
				if len(res.TimeNs) != len(res.Polarization) || len(res.TimeNs) != len(res.Current) || len(res.TimeNs) == 0 {
					t.Fatalf("cell %d trace shape invalid: %+v", i, res)
				}
				if math.IsNaN(res.FinalP) || math.IsInf(res.FinalP, 0) || math.IsNaN(res.Energy_fJ) || math.IsInf(res.Energy_fJ, 0) {
					t.Fatalf("cell %d transient scalars invalid: %+v", i, res)
				}
				for j := 1; j < len(res.TimeNs); j++ {
					if res.TimeNs[j] <= res.TimeNs[j-1] {
						t.Fatalf("cell %d time not monotonic at %d: %v", i, j, res.TimeNs)
					}
				}
				ch := CharacterizeTransientResult(cfg, res)
				if ch.WriteTimeNs < 0 || ch.ReadTimeNs < 0 || math.IsNaN(ch.WriteEnergy_fJ) || math.IsNaN(ch.ReadEnergy_fJ) {
					t.Fatalf("cell %d characterization invalid: %+v", i, ch)
				}
			}
		})
	}
	quiescent := TransientSolve(cfg, nil, 0)
	if len(quiescent) != cfg.Rows*cfg.Cols {
		t.Fatalf("quiescent result count=%d", len(quiescent))
	}
	for _, res := range quiescent {
		if len(res.TimeNs) != 0 || res.FinalP >= 0 {
			t.Fatalf("quiescent transient should preserve negative initial polarization: %+v", res)
		}
	}

	plannerCases := []MixedPrecisionPlannerInput{
		{AccuracyTarget: 0.90, EnergyBudgetPJ: 40, LatencyBudgetNS: 400},
		{AccuracyTarget: 0.95, EnergyBudgetPJ: 80, LatencyBudgetNS: 600},
		{AccuracyTarget: 0.98, EnergyBudgetPJ: 120, LatencyBudgetNS: 900},
	}
	for _, in := range plannerCases {
		plan, err := PlanMixedPrecisionConfig(in)
		if err != nil {
			t.Fatalf("PlanMixedPrecisionConfig(%+v): %v", in, err)
		}
		if plan.ExpectedAccuracy < in.AccuracyTarget || plan.ExpectedEnergyPJ > in.EnergyBudgetPJ || plan.ExpectedLatencyNS > in.LatencyBudgetNS || plan.Levels <= 0 || plan.ADCBits <= 0 || plan.TileRows <= 0 || plan.TileCols <= 0 {
			t.Fatalf("plan violates constraints: input=%+v plan=%+v", in, plan)
		}
	}
	for _, bad := range []MixedPrecisionPlannerInput{{AccuracyTarget: 0}, {AccuracyTarget: 1.2, EnergyBudgetPJ: 1, LatencyBudgetNS: 1}, {AccuracyTarget: 0.9, EnergyBudgetPJ: 0, LatencyBudgetNS: 1}, {AccuracyTarget: 0.9, EnergyBudgetPJ: 1, LatencyBudgetNS: 0}, {AccuracyTarget: 0.995, EnergyBudgetPJ: 1, LatencyBudgetNS: 1}} {
		if _, err := PlanMixedPrecisionConfig(bad); err == nil {
			t.Fatalf("PlanMixedPrecisionConfig(%+v) expected error", bad)
		}
	}
}

func TestModule4ArraySimE2EPatternSelectorAndEnduranceWorkflow(t *testing.T) {
	patterns := map[string][][]int{
		"checker":       GenerateCheckerboard(4, 5, 30),
		"ones":          GenerateAllOnes(4, 5, 30),
		"zeros":         GenerateAllZeros(4, 5),
		"walking_one":   GenerateWalkingOnes(4, 5, 7, 30),
		"walking_zero":  GenerateWalkingZeros(4, 5, 7, 30),
		"diagonal":      GenerateDiagonal(4, 5, 30),
		"row_stripe":    GenerateRowStripe(4, 5, 30),
		"random_seeded": GenerateRandom(4, 5, 30, 123),
	}
	for name, pattern := range patterns {
		t.Run(name, func(t *testing.T) {
			assertLevelPatternE2E(t, pattern, 4, 5, 29)
		})
	}
	if got := GenerateRandom(4, 5, 30, 123); !sameIntMatrixE2E(got, patterns["random_seeded"]) {
		t.Fatalf("GenerateRandom with same seed should be deterministic: %+v vs %+v", got, patterns["random_seeded"])
	}
	if p := GenerateWalkingOnes(2, 2, 99, 30); countNonZeroE2E(p) != 0 {
		t.Fatalf("out-of-range walking one should be all zero: %+v", p)
	}
	if p := GenerateAllOnes(-1, 3, 30); len(p) != 0 {
		t.Fatalf("negative rows should yield empty pattern: %+v", p)
	}

	params := SolveParams{
		WLVoltages:  []float64{0.2, 0.2, 0.2},
		BLVoltages:  []float64{0, 0, 0},
		Conductance: [][]float64{{40e-6, 50e-6, 60e-6}, {70e-6, 80e-6, 90e-6}, {100e-6, 110e-6, 120e-6}},
		Geometry:    DefaultCellGeometry(),
		Wire:        WireParams{RWordLine: 0.4, RBitLine: 2.5},
	}
	base, err := NewTierBSolver().Solve(params)
	if err != nil {
		t.Fatalf("TierB base solve: %v", err)
	}
	params.SelectorMode = SelectorRead
	params.ReadMask = [][]bool{{true, false, true}, {false, true, false}, {true, false, true}}
	masked, err := NewTierBSolver().Solve(params)
	if err != nil {
		t.Fatalf("TierB masked solve: %v", err)
	}
	params.Selector = SelectorDeviceParams{Enabled: true, OnConductance: math.Inf(1), OffConductance: 1e-12}
	softMasked, err := NewTierBSolver().Solve(params)
	if err != nil {
		t.Fatalf("TierB soft selector solve: %v", err)
	}
	baseAbs := absCurrentSumE2E(base)
	maskedAbs := absCurrentSumE2E(masked)
	softAbs := absCurrentSumE2E(softMasked)
	if baseAbs <= 0 || maskedAbs <= 0 || maskedAbs >= baseAbs || softAbs <= 0 || softAbs >= baseAbs {
		t.Fatalf("selector masks should reduce current: base=%g masked=%g soft=%g", baseAbs, maskedAbs, softAbs)
	}

	cycles := []float64{-1, 0, 1e6, 1e8, 1e9, 5e9}
	points := SimulateEnduranceAccuracy(cycles, EnduranceAccuracyConfig{BaselineAccuracy: 0.97, EnduranceLimit: 1e9, DriftAtLimit: 0.2, Sensitivity: 0.5})
	if len(points) != len(cycles) {
		t.Fatalf("endurance point count=%d", len(points))
	}
	prevAcc := math.Inf(1)
	prevDrift := -1.0
	for _, point := range points {
		if point.Cycles < 0 || point.ConductanceDrift < 0 || point.Accuracy < 0 || point.Accuracy > 1 || math.IsNaN(point.Accuracy) {
			t.Fatalf("endurance point invalid: %+v", point)
		}
		if point.Accuracy > prevAcc+1e-12 || point.ConductanceDrift < prevDrift-1e-12 {
			t.Fatalf("endurance should monotonically degrade: prevAcc=%g prevDrift=%g point=%+v", prevAcc, prevDrift, point)
		}
		prevAcc = point.Accuracy
		prevDrift = point.ConductanceDrift
	}
	defaults := SimulateEnduranceAccuracy([]float64{0, 1e9}, EnduranceAccuracyConfig{})
	if len(defaults) != 2 || defaults[0].Accuracy <= defaults[1].Accuracy || defaults[0].Accuracy <= 0 {
		t.Fatalf("default endurance config invalid: %+v", defaults)
	}
}

func assertLevelPatternE2E(t *testing.T, pattern [][]int, rows, cols, high int) {
	t.Helper()
	if len(pattern) != rows {
		t.Fatalf("rows=%d, want %d", len(pattern), rows)
	}
	for r := range pattern {
		if len(pattern[r]) != cols {
			t.Fatalf("row %d cols=%d, want %d", r, len(pattern[r]), cols)
		}
		for c, v := range pattern[r] {
			if v < 0 || v > high {
				t.Fatalf("pattern[%d][%d]=%d outside [0,%d]", r, c, v, high)
			}
		}
	}
}

func sameIntMatrixE2E(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for r := range a {
		if len(a[r]) != len(b[r]) {
			return false
		}
		for c := range a[r] {
			if a[r][c] != b[r][c] {
				return false
			}
		}
	}
	return true
}

func countNonZeroE2E(matrix [][]int) int {
	count := 0
	for r := range matrix {
		for _, v := range matrix[r] {
			if v != 0 {
				count++
			}
		}
	}
	return count
}

func absCurrentSumE2E(result SolveResult) float64 {
	sum := 0.0
	for r := range result.CellCurrents {
		for _, current := range result.CellCurrents[r] {
			sum += math.Abs(current)
		}
	}
	return sum
}
