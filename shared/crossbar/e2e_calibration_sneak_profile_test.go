package crossbar

import (
	"math"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/crossbar/levels"
)

func TestModule2CrossbarE2EDriftCalibrationSneakAnalyzerAndLevelWorkflow(t *testing.T) {
	data := []RetentionDatum{
		{TimeS: 3600, TemperatureK: 358, Retention: 0.995},
		{TimeS: 24 * 3600, TemperatureK: 358, Retention: 0.992},
		{TimeS: 7 * 24 * 3600, TemperatureK: 385, Retention: 0.980},
		{TimeS: 30 * 24 * 3600, TemperatureK: 400, Retention: 0.940},
	}
	cal := CalibrateDriftToRetention(data)
	if cal.Coeff <= 0 || cal.Exponent <= 0 || cal.ActivationE <= 0 || math.IsInf(cal.RMSE, 0) || math.IsNaN(cal.RMSE) {
		t.Fatalf("drift calibration invalid: %+v", cal)
	}
	for _, d := range data {
		pred := retentionFromParams(d.TimeS, d.TemperatureK, cal.Coeff, cal.Exponent, cal.ActivationE)
		if pred < 0 || pred > 1 || math.IsNaN(pred) {
			t.Fatalf("retention prediction invalid for %+v: %g", d, pred)
		}
	}
	if retentionFromParams(0, 300, cal.Coeff, cal.Exponent, cal.ActivationE) != 1 {
		t.Fatalf("zero-time retention should be 1")
	}

	sp := NewSneakPathAnalyzer(4, 4)
	for r := 0; r < sp.Rows; r++ {
		for c := 0; c < sp.Cols; c++ {
			sp.SetConductance(r, c, float64(1+r+c)*10e-6)
		}
	}
	sp.SetConductance(-1, 0, 1) // side-effect safe boundary
	sp.AnalyzeTarget(1, 1, 0.2)
	stats := sp.GetStats(0.2)
	if stats.TargetCurrent <= 0 || stats.TotalSneakCurrent <= 0 || stats.SneakRatio <= 0 || stats.NumSneakPaths != 9 || stats.WorstSneakPath <= 0 || math.IsNaN(stats.SignalToNoiseRatio) {
		t.Fatalf("sneak stats invalid: %+v", stats)
	}
	mitigated := sp.AnalyzeWithMitigation(1, 1, 0.2, SneakMitigation{UseSelector: true, SelectorOnOff: 100, UseHalfSelect: true, HalfSelectVoltage: 0.1})
	if mitigated.TotalSneakCurrent >= stats.TotalSneakCurrent || mitigated.SneakRatio >= stats.SneakRatio {
		t.Fatalf("mitigation did not reduce sneak: base=%+v mitigated=%+v", stats, mitigated)
	}
	zeroV := sp.AnalyzeWithMitigation(1, 1, 0.2, SneakMitigation{UseHalfSelect: true, HalfSelectVoltage: 0.3})
	if zeroV.TotalSneakCurrent != 0 || zeroV.SignalToNoiseRatio != 100 {
		t.Fatalf("over-half-select should eliminate sneak current: %+v", zeroV)
	}

	for _, value := range []float64{-1, 0, 0.017, 0.5, 0.983, 1, 2} {
		q := levels.QuantizeToDefaultLevels(value)
		if q < 0 || q > 1 || math.IsNaN(q) {
			t.Fatalf("QuantizeToDefaultLevels(%g)=%g", value, q)
		}
		level := levels.DefaultLevelFor(q)
		if level < 0 || level >= levels.DefaultQuantizationLevels {
			t.Fatalf("DefaultLevelFor(%g)=%d outside [0,%d)", q, level, levels.DefaultQuantizationLevels)
		}
	}
}

func TestModule2CrossbarE2ETemperatureProfileMVMMatrix(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 4, Cols: 4, NoiseLevel: 0.02, ADCBits: 8, DACBits: 8, ProcessVariation: &ProcessVariationConfig{DeviceSigma: 0.01}})
	if err != nil {
		t.Fatalf("NewArray error = %v", err)
	}
	defer arr.Destroy()
	if err := arr.ProgramWeightMatrix([][]float64{{0.1, 0.3, 0.5, 0.7}, {0.9, 0.2, 0.4, 0.6}, {0.8, 1, 0.2, 0.4}, {0.25, 0.5, 0.75, 1}}); err != nil {
		t.Fatalf("ProgramWeightMatrix error = %v", err)
	}
	input := []float64{0.2, 0.4, 0.6, 0.8}
	profiles := []*TemperatureProfile{nil, {Enable: false}, DefaultTemperatureProfile(), {Enable: true, ApplyConductanceWindow: true}, {Enable: true, ApplyVariationNoise: true}, {Enable: true, ApplyDrift: true}}
	for idx, profile := range profiles {
		t.Run(strings.ReplaceAll(profileNameE2E(profile, idx), " ", "_"), func(t *testing.T) {
			for _, temp := range []float64{77, 300, 400} {
				opts := DefaultMVMOptions()
				opts.Temperature = temp
				opts.TemperatureProfile = profile
				opts.EnableIRDrop = true
				opts.EnableSneakPaths = true
				opts.EnableVariation = true
				opts.EnableDrift = true
				res, err := arr.MVMWithNonIdealities(input, opts)
				if err != nil {
					t.Fatalf("MVMWithNonIdealities temp=%g profile=%+v error=%v", temp, profile, err)
				}
				assertE2EMVMResultContract(t, res, arr.Rows(), arr.Rows()*len(input))
				if res.IRDropAnalysis == nil || res.SneakPathAnalysis == nil || res.SneakTrace == nil {
					t.Fatalf("missing analyses at temp=%g profile=%+v", temp, profile)
				}
			}
		})
	}
}

func TestModule2CrossbarE2EAnalyzerBoundaryContracts(t *testing.T) {
	sp := NewSneakPathAnalyzer(2, 2)
	sp.SetConductance(0, 0, 0)
	sp.AnalyzeTarget(0, 0, 0.2)
	stats := sp.GetStats(0.2)
	if !math.IsInf(stats.SneakRatio, 1) || stats.TargetCurrent != 0 {
		t.Fatalf("zero target sneak stats invalid: %+v", stats)
	}

	arr, err := NewArray(&Config{Rows: 3, Cols: 3, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	defer arr.Destroy()
	if err := arr.ProgramWeightMatrix([][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}); err != nil {
		t.Fatalf("ProgramWeightMatrix zeros: %v", err)
	}
	mh := arr.AnalyzeSneakPathsMultiHop(1, 1, 1, 2)
	if mh.FiveHopSneak != 0 || mh.TotalSneakMultiHop != mh.TotalSneak {
		t.Fatalf("zero-conductance multihop invalid: %+v", mh)
	}
}

func profileNameE2E(profile *TemperatureProfile, idx int) string {
	if profile == nil {
		return "legacy-nil"
	}
	return strings.Join([]string{boolNameE2E(profile.Enable), boolNameE2E(profile.ApplyConductanceWindow), boolNameE2E(profile.ApplyVariationNoise), boolNameE2E(profile.ApplyDrift), string(rune('0' + idx))}, "-")
}

func boolNameE2E(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
