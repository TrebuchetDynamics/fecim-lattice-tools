package calibration

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

type targetResult struct {
	Name        string         `json:"name"`
	Source      string         `json:"source_citation"`
	ModelParams map[string]any `json:"model_params"`
	FitError    float64        `json:"fit_error"`
	ErrorMetric string         `json:"error_metric"`
	Tolerance   float64        `json:"tolerance"`
	TargetMet   bool           `json:"target_met"`
}

type calibrationReport struct {
	Title        string         `json:"title"`
	Results      []targetResult `json:"results"`
	TargetsMet   int            `json:"targets_met"`
	TargetsTotal int            `json:"targets_total"`
}

func TestCalibrationPack_GenerateReport(t *testing.T) {
	tg := DefaultTargets()
	results := make([]targetResult, 0, 3)

	freqRes := evalFrequencySweep(tg.FrequencySweep)
	switchRes := evalSwitchingTime(tg.SwitchingTime)
	marginRes := evalReadMargin(tg.ReadMargin)
	results = append(results, freqRes, switchRes, marginRes)

	met := 0
	for _, r := range results {
		if r.TargetMet {
			met++
		}
	}

	report := calibrationReport{
		Title:        "Calibration pack report (literature-pinned targets)",
		Results:      results,
		TargetsMet:   met,
		TargetsTotal: len(results),
	}

	blob, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	blob = append(blob, '\n')
	if err := os.WriteFile(filepath.Join(".", "calibration_report.json"), blob, 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}

	if report.TargetsTotal == 0 {
		t.Fatal("no targets evaluated")
	}
	t.Logf("calibration targets met: %d/%d", report.TargetsMet, report.TargetsTotal)
}

func evalFrequencySweep(targets []FrequencySweepTarget) targetResult {
	base := sharedphysics.HysteresisMetrics{FrequencyHz: 1e3, Pr_Cm2: 21e-2, Ec_Vm: 1.02e8, LoopArea_Jm3: 2.35e8}
	cfg := sharedphysics.FrequencyDispersionConfig{
		ReferenceHz:        1e3,
		EcLogSlope:         0.030,
		PrLogSlope:         -0.035,
		LoopAreaLogSlope:   -0.045,
		MinMultiplierClamp: 0.4,
	}
	var sum float64
	for _, tg := range targets {
		m, _ := sharedphysics.ApplyFrequencyDispersion(base, tg.FrequencyHz, cfg)
		prModel := m.Pr_Cm2 * 100.0
		ecModel := m.Ec_Vm / 1e8
		dPr := prModel - tg.Pr_uCcm2
		dEc := ecModel - tg.Ec_MVcm
		sum += dPr*dPr + dEc*dEc
	}
	rmse := math.Sqrt(sum / float64(2*len(targets)))
	return targetResult{
		Name:        "P-E loop shape frequency sweep (1 Hz, 1 kHz, 1 MHz)",
		Source:      targets[0].Citation,
		ModelParams: map[string]any{"model": "log-frequency dispersion", "config": cfg},
		FitError:    rmse,
		ErrorMetric: "RMSE over [Pr(uC/cm^2), Ec(MV/cm)]",
		Tolerance:   2.5,
		TargetMet:   rmse <= 2.5,
	}
}

func evalSwitchingTime(targets []SwitchingTimeTarget) targetResult {
	const (
		t0Ns = 0.9
		ea   = 6.4
	)
	// t_ns = t0 * exp(ea / E_MVcm)
	var ape float64
	for _, tg := range targets {
		pred := t0Ns * math.Exp(ea/tg.PulseMVcm)
		ape += math.Abs(pred-tg.TimeNs) / tg.TimeNs * 100
	}
	mape := ape / float64(len(targets))
	return targetResult{
		Name:        "Switching time vs pulse amplitude",
		Source:      targets[0].Citation,
		ModelParams: map[string]any{"model": "Merz-like", "t0_ns": t0Ns, "Ea_over_E_units": ea},
		FitError:    mape,
		ErrorMetric: "MAPE (%)",
		Tolerance:   15.0,
		TargetMet:   mape <= 15.0,
	}
}

func evalReadMargin(targets []ReadMarginTarget) targetResult {
	const (
		baseMv = 180.0
		k      = 0.35
	)
	var sum float64
	for _, tg := range targets {
		ratio := float64(tg.ArraySize) / 64.0
		pred := baseMv / (1.0 + k*math.Log2(ratio))
		d := pred - tg.MarginMv
		sum += d * d
	}
	rmse := math.Sqrt(sum / float64(len(targets)))
	return targetResult{
		Name:        "Read margin trend vs array size",
		Source:      targets[0].Citation,
		ModelParams: map[string]any{"model": "log-size IR/sneak trend", "base_mV": baseMv, "k": k},
		FitError:    rmse,
		ErrorMetric: "RMSE (mV)",
		Tolerance:   18.0,
		TargetMet:   rmse <= 18.0,
	}
}
