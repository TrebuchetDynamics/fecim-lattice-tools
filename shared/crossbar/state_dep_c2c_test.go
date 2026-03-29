package crossbar

import (
	"math"
	"testing"
)

// TestStateDepC2CDisabledByDefault verifies that state-dependent C2C variation
// is opt-in and does not affect existing behavior when not enabled.
func TestStateDepC2CDisabledByDefault(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:    4,
		Cols:    4,
		ADCBits: 8,
		DACBits: 8,
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	// applyStateDepC2CNoise should return 1.0 when disabled
	for i := 0; i < 100; i++ {
		factor := arr.applyStateDepC2CNoise(0.5)
		if factor != 1.0 {
			t.Fatalf("expected factor 1.0 when disabled, got %v", factor)
		}
	}
}

// TestStateDepC2CNoiseReturnsUnityWhenZeroSigma verifies that even when
// enabled, zero sigma produces no noise.
func TestStateDepC2CNoiseReturnsUnityWhenZeroSigma(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:               4,
		Cols:               4,
		ADCBits:            8,
		DACBits:            8,
		NoiseLevel:         0.0,
		StateDepC2CEnabled: true,
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	for i := 0; i < 100; i++ {
		factor := arr.applyStateDepC2CNoise(0.1)
		if factor != 1.0 {
			t.Fatalf("expected factor 1.0 with zero sigma, got %v", factor)
		}
	}
}

// TestStateDepC2CHRSHigherVariationThanLRS is the core physics test.
// It verifies that low-conductance (HRS) cells exhibit higher cycle-to-cycle
// variation than high-conductance (LRS) cells, consistent with RRAM/FeFET
// literature (Jang et al., IEEE TED 2015; Fantini et al., IMW 2014).
func TestStateDepC2CHRSHigherVariationThanLRS(t *testing.T) {
	const (
		sigmaBase = 0.05
		k         = 1.5
		nSamples  = 10_000

		gHRS = 0.1 // Near high-resistance state (low conductance)
		gLRS = 0.9 // Near low-resistance state (high conductance)
	)

	arr, err := NewArray(&Config{
		Rows:               4,
		Cols:               4,
		ADCBits:            8,
		DACBits:            8,
		NoiseLevel:         sigmaBase,
		StateDepC2CEnabled: true,
		StateDepC2CScaling: k,
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	// Collect noise samples for HRS and LRS
	hrsFactors := make([]float64, nSamples)
	lrsFactors := make([]float64, nSamples)

	for i := 0; i < nSamples; i++ {
		hrsFactors[i] = arr.applyStateDepC2CNoise(gHRS)
		lrsFactors[i] = arr.applyStateDepC2CNoise(gLRS)
	}

	hrsStd := stddev(hrsFactors)
	lrsStd := stddev(lrsFactors)

	t.Logf("HRS (G_norm=%.1f) sigma = %.6f", gHRS, hrsStd)
	t.Logf("LRS (G_norm=%.1f) sigma = %.6f", gLRS, lrsStd)
	t.Logf("Ratio HRS/LRS sigma = %.2f", hrsStd/lrsStd)

	// Expected ratio: sigma_HRS / sigma_LRS = (1 + k*(1-gHRS)) / (1 + k*(1-gLRS))
	// = (1 + 1.5*0.9) / (1 + 1.5*0.1) = 2.35 / 1.15 ≈ 2.04
	expectedRatio := (1.0 + k*(1.0-gHRS)) / (1.0 + k*(1.0-gLRS))
	t.Logf("Expected ratio = %.2f", expectedRatio)

	// The measured ratio should be within 20% of the expected ratio
	// (Monte Carlo sampling variance requires some tolerance).
	measuredRatio := hrsStd / lrsStd
	if math.Abs(measuredRatio-expectedRatio)/expectedRatio > 0.20 {
		t.Errorf("HRS/LRS sigma ratio %.2f deviates >20%% from expected %.2f",
			measuredRatio, expectedRatio)
	}

	// HRS must have strictly higher variation than LRS
	if hrsStd <= lrsStd {
		t.Errorf("HRS sigma (%.6f) should be > LRS sigma (%.6f)", hrsStd, lrsStd)
	}
}

// TestStateDepC2CDefaultScaling verifies that when StateDepC2CScaling is zero
// (unset), the default of 1.5 is used.
func TestStateDepC2CDefaultScaling(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:               4,
		Cols:               4,
		ADCBits:            8,
		DACBits:            8,
		NoiseLevel:         0.05,
		StateDepC2CEnabled: true,
		// StateDepC2CScaling intentionally left at 0 (default)
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	if got := arr.stateDepC2CScaling(); got != defaultStateDepC2CScaling {
		t.Errorf("default scaling = %v, want %v", got, defaultStateDepC2CScaling)
	}
}

// TestStateDepC2CMVMOutputVariation verifies that MVM output for HRS-loaded
// rows has a higher coefficient of variation (CV = std/mean) than LRS-loaded rows.
//
// Note: absolute output std may be lower for HRS due to smaller mean current,
// but the relative variation (CV) must be higher -- this is the physical
// observable that state-dependent C2C aims to capture.
func TestStateDepC2CMVMOutputVariation(t *testing.T) {
	const (
		size     = 8
		nReads   = 500
		gHRS     = 0.1
		gLRS     = 0.9
		inputVal = 1.0
	)

	arr, err := NewArray(&Config{
		Rows:               size,
		Cols:               size,
		ADCBits:            16, // High ADC resolution to not mask noise
		DACBits:            16,
		NoiseLevel:         0.05,
		StateDepC2CEnabled: true,
		StateDepC2CScaling: 1.5,
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	// Program: row 0 = all HRS, row 1 = all LRS
	weights := make([][]float64, size)
	for i := range weights {
		weights[i] = make([]float64, size)
		for j := range weights[i] {
			if i == 0 {
				weights[i][j] = gHRS
			} else if i == 1 {
				weights[i][j] = gLRS
			} else {
				weights[i][j] = 0.5
			}
		}
	}
	if err := arr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix: %v", err)
	}

	input := make([]float64, size)
	for j := range input {
		input[j] = inputVal
	}

	// Collect MVM outputs for row 0 (HRS) and row 1 (LRS)
	hrsOutputs := make([]float64, nReads)
	lrsOutputs := make([]float64, nReads)

	for r := 0; r < nReads; r++ {
		out, err := arr.MVM(input)
		if err != nil {
			t.Fatalf("MVM read %d: %v", r, err)
		}
		hrsOutputs[r] = out[0]
		lrsOutputs[r] = out[1]
	}

	hrsStd := stddev(hrsOutputs)
	lrsStd := stddev(lrsOutputs)
	hrsMean := mean(hrsOutputs)
	lrsMean := mean(lrsOutputs)

	hrsCV := hrsStd / hrsMean
	lrsCV := lrsStd / lrsMean

	t.Logf("MVM HRS row: mean=%.6f std=%.8f CV=%.4f", hrsMean, hrsStd, hrsCV)
	t.Logf("MVM LRS row: mean=%.6f std=%.8f CV=%.4f", lrsMean, lrsStd, lrsCV)

	// HRS row should show higher relative variation (coefficient of variation)
	// due to state-dependent C2C noise
	if hrsCV <= lrsCV {
		t.Errorf("HRS row CV (%.4f) should be > LRS row CV (%.4f)", hrsCV, lrsCV)
	}
}

// TestStateDepC2CWithProcessVariation verifies that state-dependent C2C noise
// composes correctly with the existing process variation factor.
func TestStateDepC2CWithProcessVariation(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:       4,
		Cols:       4,
		ADCBits:    8,
		DACBits:    8,
		NoiseLevel: 0.05,
		ProcessVariation: &ProcessVariationConfig{
			DeviceSigma: 0.03,
		},
		StateDepC2CEnabled: true,
		StateDepC2CScaling: 1.5,
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	// Verify sigma uses ProcessVariation.DeviceSigma as base (not NoiseLevel)
	// by checking that HRS noise is scaled relative to DeviceSigma=0.03.
	const nSamples = 10_000
	factors := make([]float64, nSamples)
	for i := range factors {
		factors[i] = arr.applyStateDepC2CNoise(0.0) // HRS: maximum scaling
	}

	std := stddev(factors)
	// At G_norm=0: effectiveSigma = 0.03 * (1 + 1.5) = 0.075
	expectedSigma := 0.03 * (1.0 + 1.5)
	if math.Abs(std-expectedSigma)/expectedSigma > 0.20 {
		t.Errorf("HRS sigma = %.4f, expected ~%.4f (within 20%%)", std, expectedSigma)
	}
}

// stddev computes the sample standard deviation of a float64 slice.
func stddev(xs []float64) float64 {
	n := float64(len(xs))
	if n < 2 {
		return 0
	}
	var sum, sumSq float64
	for _, x := range xs {
		sum += x
		sumSq += x * x
	}
	mean := sum / n
	variance := (sumSq/n - mean*mean)
	if variance < 0 {
		variance = 0
	}
	return math.Sqrt(variance)
}
