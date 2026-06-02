package crossbar

import (
	"math"
	"strings"
	"testing"
)

func TestModule2CrossbarE2EVariationTemperatureDeviceErrorWorkflow(t *testing.T) {
	crossSim := DefaultCrossSimHZO()
	pv := ImportCrossSimVariation(crossSim)
	if pv.DeviceSigma <= 0 || pv.GradientX <= 0 || pv.GradientY <= 0 || pv.EdgeEffect <= 0 || pv.EdgeEffect > 0.20 {
		t.Fatalf("imported CrossSim variation invalid: %+v", pv)
	}
	exported := ExportToCrossSimFormat(pv)
	if exported.ProgramNoiseSigma <= 0 || exported.ReadNoiseSigma <= 0 || exported.D2DSigmaHRS <= exported.D2DSigmaLRS || !strings.Contains(exported.Disclaimer, "SIMULATION") {
		t.Fatalf("exported CrossSim variation invalid: %+v", exported)
	}
	if zero := ExportToCrossSimFormat(nil); zero.ProgramNoiseSigma != 0 || zero.Disclaimer != "" {
		t.Fatalf("nil CrossSim export should be zero value: %+v", zero)
	}

	hot := NewHotspotTemperatureProfile(5, 5, 300, 50)
	center := hot.TemperatureAt(2, 2)
	corner := hot.TemperatureAt(0, 0)
	if center <= corner || corner != 300 || hot.TemperatureAt(-1, 0) != 300 {
		t.Fatalf("hotspot profile invalid: center=%g corner=%g", center, corner)
	}
	gradX := NewGradientTemperatureProfile(3, 4, 290, 350, "x")
	gradY := NewGradientTemperatureProfile(3, 4, 290, 350, "y")
	if gradX.TemperatureAt(0, 0) != 290 || gradX.TemperatureAt(0, 3) != 350 || gradY.TemperatureAt(2, 0) != 350 || gradY.TemperatureAt(99, 99) != 290 {
		t.Fatalf("gradient profiles invalid: x=%v y=%v", gradX.MapK, gradY.MapK)
	}
	fallback := NewHotspotTemperatureProfile(0, 0, -1, -5)
	if fallback.Rows != 1 || fallback.Cols != 1 || fallback.AmbientK != 300 || fallback.TemperatureAt(0, 0) != 300 {
		t.Fatalf("hotspot fallback invalid: %+v", fallback)
	}

	target := [][]float64{{0.05, 0.25, 0.5}, {0.75, 1, 0.1}, {0.2, 0.4, 0.8}}
	models := []ErrorModel{ErrorModelNone, ErrorModelNormalIndependent, ErrorModelNormalProportional, ErrorModelNormalInverseProportional, ErrorModelUniformIndependent, ErrorModelUniformProportional, ErrorModel(99)}
	for _, model := range models {
		t.Run(model.String(), func(t *testing.T) {
			engine := NewDeviceErrorEngine(&ProgrammingErrorConfig{Enable: true, Model: model, Sigma: 0.05, Symmetric: false, Seed: 12}, &ReadNoiseConfig{Enable: true, Model: model, Sigma: 0.02, Persistent: true, Seed: 34})
			programmed := engine.ApplyProgrammingErrorToMatrix(target)
			read1 := engine.ApplyReadNoiseToMatrix(programmed)
			read2 := engine.ApplyReadNoiseToMatrix(programmed)
			assertDeviceMatrix01E2E(t, programmed)
			assertDeviceMatrix01E2E(t, read1)
			if !sameE2EMatrix(read1, read2) {
				t.Fatalf("persistent read noise not stable for model %s", model.String())
			}
			engine.ClearPersistentNoise()
			_ = engine.ApplyReadNoiseToMatrix(programmed)
			stats := ComputeErrorStatistics(target, programmed)
			if stats == nil || math.IsNaN(stats.RMSE) || stats.MaxAbsError < 0 || stats.MinAbsError < 0 || stats.PercentOutliers < 0 {
				t.Fatalf("error statistics invalid for %s: %+v", model.String(), stats)
			}
		})
	}
	if ComputeErrorStatistics(nil, nil) != nil {
		t.Fatal("ComputeErrorStatistics empty inputs should return nil")
	}
	loss := SimulateAccuracyDegradation(0.03, 0.01, 64)
	progBudget, readBudget := RecommendErrorBudget(0.95, 64)
	if loss <= 0 || loss > 1 || progBudget <= 0 || readBudget <= 0 || math.Abs(progBudget-readBudget) > 1e-12 {
		t.Fatalf("accuracy/error-budget helpers invalid: loss=%g budgets=%g/%g", loss, progBudget, readBudget)
	}
	cfg := DefaultDeviceNonIdealityConfig()
	cfg.EnableTypicalErrors()
	if !cfg.ProgrammingError.Enable || !cfg.ReadNoise.Enable {
		t.Fatalf("typical errors not enabled: %+v", cfg)
	}
	cfg.EnableWorstCaseErrors()
	if cfg.ProgrammingError.Sigma < 0.10 || cfg.ReadNoise.Sigma < 0.03 {
		t.Fatalf("worst-case errors invalid: %+v", cfg)
	}
}

func TestModule2CrossbarE2EMultiHopSneakSamplingWorkflow(t *testing.T) {
	small, err := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray small: %v", err)
	}
	defer small.Destroy()
	if err := small.ProgramWeightMatrix([][]float64{{1, 0.8, 0.6, 0.4}, {0.7, 0.5, 0.3, 0.9}, {0.2, 1, 0.4, 0.6}, {0.8, 0.1, 0.9, 0.5}}); err != nil {
		t.Fatalf("ProgramWeightMatrix small: %v", err)
	}
	exact := small.AnalyzeSneakPathsMultiHop(1, 2, 1, 2)
	if exact == nil || exact.SneakPathAnalysis == nil || exact.IsSampled || exact.SampleCount == 0 || exact.FiveHopSneak <= 0 || exact.TotalSneakMultiHop < exact.TotalSneak {
		t.Fatalf("exact multihop invalid: %+v", exact)
	}
	isolated := small.AnalyzeSneakPathsMultiHop(1, 2, 0.01, 2)
	if isolated.TotalSneakMultiHop >= exact.TotalSneakMultiHop {
		t.Fatalf("isolation did not reduce sneak: exact=%g isolated=%g", exact.TotalSneakMultiHop, isolated.TotalSneakMultiHop)
	}

	large, err := NewArray(&Config{Rows: 33, Cols: 33, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray large: %v", err)
	}
	defer large.Destroy()
	for r := 0; r < large.Rows(); r++ {
		for c := 0; c < large.Cols(); c++ {
			if err := large.ProgramWeight(r, c, float64((r+c)%29+1)/29); err != nil {
				t.Fatalf("ProgramWeight large[%d][%d]: %v", r, c, err)
			}
		}
	}
	sampled := large.AnalyzeSneakPathsMultiHop(16, 16, 1, 2)
	if sampled == nil || !sampled.IsSampled || sampled.SampleCount != 1000 || sampled.FiveHopSneak <= 0 || sampled.TotalSneakMultiHop <= sampled.TotalSneak {
		t.Fatalf("sampled multihop invalid: %+v", sampled)
	}

	tiny, err := NewArray(&Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray tiny: %v", err)
	}
	defer tiny.Destroy()
	if err := tiny.ProgramWeightMatrix([][]float64{{1, 1}, {1, 1}}); err != nil {
		t.Fatalf("ProgramWeightMatrix tiny: %v", err)
	}
	noFiveHop := tiny.AnalyzeSneakPathsMultiHop(0, 0, 1, 2)
	if noFiveHop.FiveHopSneak != 0 || noFiveHop.IsSampled || noFiveHop.SampleCount != 0 {
		t.Fatalf("tiny multihop should have no 5-cell paths: %+v", noFiveHop)
	}
}

func assertDeviceMatrix01E2E(t *testing.T, matrix [][]float64) {
	t.Helper()
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
	for r := range matrix {
		for c, v := range matrix[r] {
			if math.IsNaN(v) || v < 0 || v > 1 {
				t.Fatalf("matrix[%d][%d] invalid: %g", r, c, v)
			}
		}
	}
}
