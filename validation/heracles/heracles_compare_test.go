package heracles

import (
	"math"
	"path/filepath"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestHeraclesComparator_GenerateReport(t *testing.T) {
	ref := Reference10nmHfO2_300K()

	solver := sharedphysics.NewLKSolver()
	mat := sharedphysics.MaterlikHfO2()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = ref.TemperatureK
	solver.UseNLS = false
	solver.EnableNoise = false
	solver.SetState(-math.Abs(mat.Pr))

	ascModel := make([]PEPoint, 0, len(ref.Ascending))
	for _, pt := range ref.Ascending {
		p := settleAtField(solver, pt.E_MVcm*1e8, 1800, 2e-12)
		ascModel = append(ascModel, PEPoint{E_MVcm: pt.E_MVcm, P_uCcm: p * 100.0})
	}
	descModel := make([]PEPoint, 0, len(ref.Descending))
	for _, pt := range ref.Descending {
		p := settleAtField(solver, pt.E_MVcm*1e8, 1800, 2e-12)
		descModel = append(descModel, PEPoint{E_MVcm: pt.E_MVcm, P_uCcm: p * 100.0})
	}

	rmse := rmseUCcm2(ref.Ascending, ascModel, ref.Descending, descModel)
	prRef, ecRef := estimatePrEc(ref.Ascending, ref.Descending)
	prModel, ecModel := estimatePrEc(ascModel, descModel)
	areaRef := loopAreaJm3(ref.Ascending, ref.Descending)
	areaModel := loopAreaJm3(ascModel, descModel)

	report := CompareReport{
		Title:     "Heracles comparator harness (digitized reference vs LK)",
		Reference: ref.SourceCitation,
		Dataset:   "10 nm HfO2, 300 K",
		Parameters: map[string]any{
			"material":        mat.Name,
			"temperature_K":   solver.Temperature,
			"use_nls":         solver.UseNLS,
			"beta":            solver.Beta,
			"gamma":           solver.Gamma,
			"rho":             solver.Rho,
			"k_dep":           solver.K_dep,
			"thickness_m":     solver.Thickness,
			"steps_per_point": 1800,
			"dt_s":            2e-12,
		},
		Ascending:  CurvePair{Reference: ref.Ascending, Model: ascModel},
		Descending: CurvePair{Reference: ref.Descending, Model: descModel},
		Metrics: CompareMetrics{
			RMSE_uCcm2:          rmse,
			EcRef_MVcm:          ecRef,
			EcModel_MVcm:        ecModel,
			EcMismatchPct:       mismatchPct(ecModel, ecRef),
			PrRef_uCcm2:         prRef,
			PrModel_uCcm2:       prModel,
			PrMismatchPct:       mismatchPct(prModel, prRef),
			LoopAreaRef_Jm3:     areaRef,
			LoopAreaModel_Jm3:   areaModel,
			LoopAreaMismatchPct: mismatchPct(areaModel, areaRef),
		},
	}

	reportPath := filepath.Join(".", "heracles_compare_report.json")
	if err := WriteCompareReport(reportPath, report); err != nil {
		t.Fatalf("write report: %v", err)
	}

	if report.Metrics.RMSE_uCcm2 <= 0 {
		t.Fatalf("unexpected RMSE %.3f", report.Metrics.RMSE_uCcm2)
	}
	t.Logf("Heracles compare: RMSE=%.2f uC/cm^2, Ec mismatch=%.1f%%, Pr mismatch=%.1f%%, area mismatch=%.1f%%",
		report.Metrics.RMSE_uCcm2,
		report.Metrics.EcMismatchPct,
		report.Metrics.PrMismatchPct,
		report.Metrics.LoopAreaMismatchPct,
	)
}

func settleAtField(s *sharedphysics.LKSolver, E float64, steps int, dt float64) float64 {
	for i := 0; i < steps; i++ {
		s.Step(E, dt)
	}
	return s.GetState()
}

func rmseUCcm2(aRef, aModel, dRef, dModel []PEPoint) float64 {
	var sum float64
	var n int
	accum := func(ref, model []PEPoint) {
		m := len(ref)
		if len(model) < m {
			m = len(model)
		}
		for i := 0; i < m; i++ {
			d := model[i].P_uCcm - ref[i].P_uCcm
			sum += d * d
			n++
		}
	}
	accum(aRef, aModel)
	accum(dRef, dModel)
	if n == 0 {
		return 0
	}
	return math.Sqrt(sum / float64(n))
}

func mismatchPct(model, ref float64) float64 {
	if ref == 0 {
		return 0
	}
	return math.Abs(model-ref) / math.Abs(ref) * 100.0
}
