package physics

import (
	"testing"
)

func TestHZO10nmTemperatureCalibration_DataIntegrity(t *testing.T) {
	cal := HZO10nmTemperatureCalibration()

	if cal.Material == "" {
		t.Fatal("Material name must not be empty")
	}
	if len(cal.Points) != 4 {
		t.Fatalf("expected 4 calibration points, got %d", len(cal.Points))
	}

	expectedTemps := []float64{200, 300, 400, 500}
	for i, pt := range cal.Points {
		if pt.TemperatureK != expectedTemps[i] {
			t.Errorf("point %d: expected T=%.0fK, got T=%.0fK", i, expectedTemps[i], pt.TemperatureK)
		}
		if pt.Pr_Cm2 <= 0 {
			t.Errorf("point %d (T=%.0fK): Pr must be positive, got %.6e", i, pt.TemperatureK, pt.Pr_Cm2)
		}
		if pt.Ec_Vm <= 0 {
			t.Errorf("point %d (T=%.0fK): Ec must be positive, got %.6e", i, pt.TemperatureK, pt.Ec_Vm)
		}
		if pt.LoopArea_Jm3 <= 0 {
			t.Errorf("point %d (T=%.0fK): LoopArea must be positive, got %.6e", i, pt.TemperatureK, pt.LoopArea_Jm3)
		}
		if pt.Source == "" {
			t.Errorf("point %d (T=%.0fK): Source must not be empty", i, pt.TemperatureK)
		}
	}

	// Validate that Pr and Ec are in physically reasonable ranges for HZO.
	// At 300K: Pr should be around 0.10-0.30 C/m², Ec around 0.5e8-2.0e8 V/m.
	for _, pt := range cal.Points {
		if pt.Pr_Cm2 > 0.50 {
			t.Errorf("T=%.0fK: Pr=%.4f C/m² exceeds physically reasonable range for HZO", pt.TemperatureK, pt.Pr_Cm2)
		}
		if pt.Ec_Vm > 3.0e8 {
			t.Errorf("T=%.0fK: Ec=%.3e V/m exceeds physically reasonable range for HZO", pt.TemperatureK, pt.Ec_Vm)
		}
	}
}

func TestValidateTemperatureResponse_TrendCorrect(t *testing.T) {
	cal := HZO10nmTemperatureCalibration()
	solver := NewLKSolver()

	results := ValidateTemperatureResponse(solver, cal)

	if len(results) != len(cal.Points) {
		t.Fatalf("expected %d results, got %d", len(cal.Points), len(results))
	}

	// Verify Curie-Weiss trend: Pr decreases with increasing T.
	for i := 1; i < len(results); i++ {
		if results[i].PrRef >= results[i-1].PrRef {
			t.Errorf("reference Pr not decreasing: T=%.0fK Pr=%.6e >= T=%.0fK Pr=%.6e",
				results[i].TemperatureK, results[i].PrRef,
				results[i-1].TemperatureK, results[i-1].PrRef)
		}
	}

	// Verify trend in model output: Pr(model) should also decrease with T.
	for i := 1; i < len(results); i++ {
		if results[i].PrModel > results[i-1].PrModel+1e-9 {
			t.Errorf("model Pr not decreasing with T: T=%.0fK Pr=%.6e > T=%.0fK Pr=%.6e",
				results[i].TemperatureK, results[i].PrModel,
				results[i-1].TemperatureK, results[i-1].PrModel)
		}
	}

	// Verify Curie-Weiss trend: Ec decreases with increasing T.
	for i := 1; i < len(results); i++ {
		if results[i].EcRef >= results[i-1].EcRef {
			t.Errorf("reference Ec not decreasing: T=%.0fK Ec=%.6e >= T=%.0fK Ec=%.6e",
				results[i].TemperatureK, results[i].EcRef,
				results[i-1].TemperatureK, results[i-1].EcRef)
		}
	}

	// Model Ec should also decrease with T (with tolerance for LK numerical noise).
	for i := 1; i < len(results); i++ {
		if results[i].EcModel > results[i-1].EcModel*(1.05) {
			t.Errorf("model Ec not decreasing with T: T=%.0fK Ec=%.6e > T=%.0fK Ec=%.6e",
				results[i].TemperatureK, results[i].EcModel,
				results[i-1].TemperatureK, results[i-1].EcModel)
		}
	}
}

func TestValidateTemperatureResponse_ReasonableMismatch(t *testing.T) {
	cal := HZO10nmTemperatureCalibration()
	solver := NewLKSolver()

	results := ValidateTemperatureResponse(solver, cal)

	// Mismatch threshold: < 30% for our simplified LK vs FerroX phase-field.
	// The LK single-domain model omits domain nucleation/growth dynamics,
	// polycrystalline grain effects, and electrode dead-layer coupling,
	// so deviations up to ~30% are expected and acceptable.
	const maxMismatchPct = 30.0

	for _, r := range results {
		if r.PrRef > 0 && r.PrModel > 0 {
			if r.PrMismatchPct > maxMismatchPct {
				t.Errorf("T=%.0fK: Pr mismatch %.1f%% exceeds %.0f%% threshold (model=%.6e ref=%.6e)",
					r.TemperatureK, r.PrMismatchPct, maxMismatchPct, r.PrModel, r.PrRef)
			}
		}
		if r.EcRef > 0 && r.EcModel > 0 {
			if r.EcMismatchPct > maxMismatchPct {
				t.Errorf("T=%.0fK: Ec mismatch %.1f%% exceeds %.0f%% threshold (model=%.6e ref=%.6e)",
					r.TemperatureK, r.EcMismatchPct, maxMismatchPct, r.EcModel, r.EcRef)
			}
		}
	}

	// Sanity check: at least some extracted values are non-zero.
	nonzero := 0
	for _, r := range results {
		if r.PrModel > 0 && r.EcModel > 0 {
			nonzero++
		}
	}
	if nonzero == 0 {
		t.Fatal("all model Pr/Ec values are zero; hysteresis extraction likely failed")
	}
}

func TestHZO10nmTemperatureCalibration_MonotonicDecay(t *testing.T) {
	cal := HZO10nmTemperatureCalibration()

	// Verify that reference Pr and Ec values decrease monotonically with temperature
	// (Curie-Weiss decay toward Tc). This is a fundamental physics constraint.
	for i := 1; i < len(cal.Points); i++ {
		if cal.Points[i].Pr_Cm2 >= cal.Points[i-1].Pr_Cm2 {
			t.Errorf("reference Pr not decreasing: T=%.0fK Pr=%.6e >= T=%.0fK Pr=%.6e",
				cal.Points[i].TemperatureK, cal.Points[i].Pr_Cm2,
				cal.Points[i-1].TemperatureK, cal.Points[i-1].Pr_Cm2)
		}
		if cal.Points[i].Ec_Vm >= cal.Points[i-1].Ec_Vm {
			t.Errorf("reference Ec not decreasing: T=%.0fK Ec=%.6e >= T=%.0fK Ec=%.6e",
				cal.Points[i].TemperatureK, cal.Points[i].Ec_Vm,
				cal.Points[i-1].TemperatureK, cal.Points[i-1].Ec_Vm)
		}
		if cal.Points[i].LoopArea_Jm3 >= cal.Points[i-1].LoopArea_Jm3 {
			t.Errorf("reference LoopArea not decreasing: T=%.0fK W=%.6e >= T=%.0fK W=%.6e",
				cal.Points[i].TemperatureK, cal.Points[i].LoopArea_Jm3,
				cal.Points[i-1].TemperatureK, cal.Points[i-1].LoopArea_Jm3)
		}
	}

	// Verify that all Pr values are below room-temperature material Pr (physically bounded).
	mat := DefaultHZO()
	for _, pt := range cal.Points {
		if pt.Pr_Cm2 > mat.Pr*1.1 {
			t.Errorf("T=%.0fK: reference Pr=%.6e exceeds material Pr=%.6e (should be bounded)",
				pt.TemperatureK, pt.Pr_Cm2, mat.Pr)
		}
	}
}
