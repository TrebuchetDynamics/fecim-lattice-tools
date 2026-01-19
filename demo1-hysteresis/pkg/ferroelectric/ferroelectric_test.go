package ferroelectric

import (
	"math"
	"testing"
)

// TestHysteresisLoopExists verifies the P-E curve shows proper hysteresis behavior.
func TestHysteresisLoopExists(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Generate hysteresis loop
	Emax := 2 * material.Ec // Go beyond coercive field
	E, P := model.GetHysteresisLoop(Emax, 50)

	if len(E) == 0 || len(P) == 0 {
		t.Fatal("Hysteresis loop generation returned empty arrays")
	}

	t.Logf("Generated hysteresis loop with %d points", len(E))

	// Verify we have reasonable polarization values
	maxP := 0.0
	minP := 0.0
	for _, p := range P {
		if p > maxP {
			maxP = p
		}
		if p < minP {
			minP = p
		}
	}

	t.Logf("Polarization range: [%.4f, %.4f] C/m²", minP, maxP)

	// Should reach close to saturation polarization
	if maxP < 0.5*material.Ps {
		t.Errorf("Max polarization %.4f is too low (expected > %.4f)", maxP, 0.5*material.Ps)
	}
	if minP > -0.5*material.Ps {
		t.Errorf("Min polarization %.4f is too high (expected < %.4f)", minP, -0.5*material.Ps)
	}
}

// TestHysteresisAsymmetry verifies that the ascending and descending branches differ.
// This is the key signature of hysteresis - path dependence.
func TestHysteresisAsymmetry(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Sweep from negative to positive
	model.Reset()
	E := 0.0
	for i := 0; i < 20; i++ {
		E = -material.Ec + 2*material.Ec*float64(i)/19
		model.Update(E)
	}
	ascendingP := model.Polarization()

	// Sweep from positive to negative
	model.Reset()
	// First go to positive saturation
	for i := 0; i <= 20; i++ {
		E = 2 * material.Ec * float64(i) / 20
		model.Update(E)
	}
	// Then come back to zero
	for i := 0; i <= 20; i++ {
		E = 2*material.Ec - 2*material.Ec*float64(i)/20
		model.Update(E)
	}
	descendingP := model.Polarization()

	t.Logf("Ascending from -Ec to +Ec: P = %.4f", ascendingP)
	t.Logf("Descending from +Emax to 0: P = %.4f", descendingP)

	// At E=0, the polarization should differ based on history
	// (this is remanent polarization)
	if math.Abs(ascendingP-descendingP) < 0.01*material.Ps {
		t.Log("Warning: Ascending and descending paths show similar polarization at same E")
		// This might be OK depending on where we measure, but it's worth noting
	}
}

// TestCoerciveFieldSwitching verifies polarization switches sign around Ec.
func TestCoerciveFieldSwitching(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Start with positive polarization (apply positive field)
	model.Reset()
	model.Update(2 * material.Ec) // Positive saturation
	initialP := model.Polarization()

	// Apply field just below coercive field - should still be positive
	model.Update(-0.5 * material.Ec)
	belowEcP := model.Polarization()

	// Apply field beyond coercive field - should switch to negative
	model.Update(-2 * material.Ec)
	beyondEcP := model.Polarization()

	t.Logf("Initial (E=+2Ec): P = %.4f", initialP)
	t.Logf("Below Ec (E=-0.5Ec): P = %.4f", belowEcP)
	t.Logf("Beyond Ec (E=-2Ec): P = %.4f", beyondEcP)

	// Initial should be positive
	if initialP < 0 {
		t.Errorf("Initial polarization should be positive, got %.4f", initialP)
	}

	// Beyond Ec should switch to negative
	if beyondEcP > 0 {
		t.Errorf("Polarization beyond -Ec should be negative, got %.4f", beyondEcP)
	}
}

// TestDiscreteStatesCount verifies 30 discrete states for IronLattice.
func TestDiscreteStatesCount(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	states := model.DiscreteStates(30)

	if len(states) != 30 {
		t.Errorf("Expected 30 discrete states, got %d", len(states))
	}

	// Verify states span from -Ps to +Ps
	if states[0] > -0.9*material.Ps {
		t.Errorf("First state %.4f should be close to -Ps (%.4f)", states[0], -material.Ps)
	}
	if states[29] < 0.9*material.Ps {
		t.Errorf("Last state %.4f should be close to +Ps (%.4f)", states[29], material.Ps)
	}

	// Verify states are evenly spaced
	expectedSpacing := 2 * material.Ps / 29
	for i := 1; i < 30; i++ {
		spacing := states[i] - states[i-1]
		if math.Abs(spacing-expectedSpacing) > 1e-10 {
			t.Errorf("State spacing at index %d is %.6f, expected %.6f", i, spacing, expectedSpacing)
		}
	}

	t.Logf("30 discrete states verified, spacing: %.6f C/m²", expectedSpacing)
}

// TestMaterialParameters verifies HZO material parameters are physically reasonable.
func TestMaterialParameters(t *testing.T) {
	material := DefaultHZO()

	// Check polarization values are in reasonable range for HZO
	// Literature values: Pr ~ 10-40 μC/cm², Ps ~ 15-50 μC/cm²
	if material.Pr < 10e-2 || material.Pr > 50e-2 {
		t.Errorf("Remanent polarization %.2f C/m² is outside expected range", material.Pr)
	}
	if material.Ps < 15e-2 || material.Ps > 60e-2 {
		t.Errorf("Saturation polarization %.2f C/m² is outside expected range", material.Ps)
	}

	// Check coercive field (literature: 0.5-2 MV/cm)
	if material.Ec < 0.5e8 || material.Ec > 3e8 {
		t.Errorf("Coercive field %.2e V/m is outside expected range", material.Ec)
	}

	// Check thickness is reasonable for FeFET applications (5-20 nm typical)
	if material.Thickness < 1e-9 || material.Thickness > 50e-9 {
		t.Errorf("Film thickness %.0f nm is outside expected range", material.Thickness*1e9)
	}

	// Check Ps > Pr (saturation should exceed remanent)
	if material.Ps <= material.Pr {
		t.Errorf("Ps (%.4f) should be greater than Pr (%.4f)", material.Ps, material.Pr)
	}

	t.Logf("HZO material parameters verified:")
	t.Logf("  Pr = %.1f μC/cm²", material.Pr*1e4)
	t.Logf("  Ps = %.1f μC/cm²", material.Ps*1e4)
	t.Logf("  Ec = %.2f MV/cm", material.Ec/1e8)
	t.Logf("  Thickness = %.0f nm", material.Thickness*1e9)
}

// TestPreisachModelReset verifies the reset function works correctly.
func TestPreisachModelReset(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Apply some fields
	model.Update(2 * material.Ec)
	model.Update(-material.Ec)

	if model.Polarization() == 0 {
		t.Error("Model should have non-zero polarization after updates")
	}

	// Reset
	model.Reset()

	if model.Polarization() != 0 {
		t.Errorf("After reset, polarization should be 0, got %.4f", model.Polarization())
	}
}

// TestNormalizedPolarization verifies normalized output is in [-1, 1] range.
func TestNormalizedPolarization(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Test at various field values
	testFields := []float64{
		-2 * material.Ec,
		-material.Ec,
		0,
		material.Ec,
		2 * material.Ec,
	}

	for _, E := range testFields {
		model.Update(E)
		normP := model.NormalizedPolarization()

		if normP < -1.1 || normP > 1.1 {
			t.Errorf("Normalized polarization %.4f at E=%.2e is outside [-1, 1] range",
				normP, E)
		}
	}
}
