package crossbar

import (
	"math"
	"testing"
)

// TestMVMWithUncertainty_BasicOutput verifies that MVMWithUncertainty returns
// the same output slice as the standard MVM and populates uncertainty fields.
func TestMVMWithUncertainty_BasicOutput(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.02,
		ADCBits:    8,
		DACBits:    8,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Program a known weight pattern: identity-like diagonal.
	for i := 0; i < 4; i++ {
		if err := arr.ProgramWeight(i, i, 1.0); err != nil {
			t.Fatal(err)
		}
	}

	input := []float64{0.5, 0.5, 0.5, 0.5}

	// Run both MVM and MVMWithUncertainty.
	mvmOut, err := arr.MVM(input)
	if err != nil {
		t.Fatal(err)
	}

	result, err := arr.MVMWithUncertainty(input)
	if err != nil {
		t.Fatal(err)
	}

	// Output lengths must match.
	if len(result.Output) != len(mvmOut) {
		t.Fatalf("output length mismatch: got %d, want %d", len(result.Output), len(mvmOut))
	}
	if len(result.Uncertainty) != len(mvmOut) {
		t.Fatalf("uncertainty length mismatch: got %d, want %d", len(result.Uncertainty), len(mvmOut))
	}

	// Output values must match the standard MVM (same quantization path).
	for i, v := range result.Output {
		if v != mvmOut[i] {
			t.Errorf("output[%d] = %v, want %v", i, v, mvmOut[i])
		}
	}
}

// TestMVMWithUncertainty_ZeroNoise verifies that with zero noise level,
// device-variation uncertainty is zero and only DAC quantization noise remains.
func TestMVMWithUncertainty_ZeroNoise(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:       2,
		Cols:       3,
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Program all cells to 0.5 conductance.
	for i := 0; i < 2; i++ {
		for j := 0; j < 3; j++ {
			if err := arr.ProgramWeight(i, j, 0.5); err != nil {
				t.Fatal(err)
			}
		}
	}

	input := []float64{0.8, 0.4, 0.6}
	result, err := arr.MVMWithUncertainty(input)
	if err != nil {
		t.Fatal(err)
	}

	// With zero noise, uncertainty should come only from DAC quantization.
	// DAC sigma = LSB / sqrt(12), LSB = 1/255 for 8-bit DAC.
	// Each uncertainty value should be small but nonzero (from DAC noise).
	for i, u := range result.Uncertainty {
		if u < 0 {
			t.Errorf("uncertainty[%d] = %v, must be non-negative", i, u)
		}
		// With 8-bit DAC and zero device noise, uncertainty should be small.
		if u > 0.01 {
			t.Errorf("uncertainty[%d] = %v, unexpectedly large for zero device noise", i, u)
		}
	}
}

// TestMVMWithUncertainty_WithProcessVariation verifies that process variation
// config is used for conductance sigma instead of NoiseLevel.
func TestMVMWithUncertainty_WithProcessVariation(t *testing.T) {
	// Array with process variation but zero NoiseLevel.
	arr, err := NewArray(&Config{
		Rows:       2,
		Cols:       2,
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
		ProcessVariation: &ProcessVariationConfig{
			DeviceSigma: 0.10, // 10% device variation
			GradientX:   0,
			GradientY:   0,
			EdgeEffect:  0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Program nonzero conductances.
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if err := arr.ProgramWeight(i, j, 0.8); err != nil {
				t.Fatal(err)
			}
		}
	}

	input := []float64{0.5, 0.5}
	result, err := arr.MVMWithUncertainty(input)
	if err != nil {
		t.Fatal(err)
	}

	// With 10% device sigma and nonzero conductance, uncertainty must be
	// meaningfully larger than DAC-only noise.
	for i, u := range result.Uncertainty {
		if u < 0.005 {
			t.Errorf("uncertainty[%d] = %v, expected larger uncertainty with 10%% device sigma", i, u)
		}
	}
}

// TestMVMWithUncertainty_Saturation verifies that the saturated count
// correctly identifies outputs at ADC rails.
func TestMVMWithUncertainty_Saturation(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:       3,
		Cols:       2,
		NoiseLevel: 0,
		ADCBits:    4, // Low resolution to make saturation more likely
		DACBits:    4,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Row 0: all zeros -> output = 0 (saturated at min)
	// Row 1: moderate conductance -> should be mid-range
	// Row 2: all ones -> output = 1 (saturated at max)
	if err := arr.ProgramWeight(1, 0, 0.3); err != nil {
		t.Fatal(err)
	}
	if err := arr.ProgramWeight(1, 1, 0.3); err != nil {
		t.Fatal(err)
	}
	if err := arr.ProgramWeight(2, 0, 1.0); err != nil {
		t.Fatal(err)
	}
	if err := arr.ProgramWeight(2, 1, 1.0); err != nil {
		t.Fatal(err)
	}

	input := []float64{1.0, 1.0}
	result, err := arr.MVMWithUncertainty(input)
	if err != nil {
		t.Fatal(err)
	}

	// At minimum, rows 0 (all-zero conductance) should register as saturated.
	// Row 2 with all-ones may also saturate at ADC max.
	if result.Saturated < 1 {
		t.Errorf("expected at least 1 saturated output, got %d (outputs: %v)",
			result.Saturated, result.Output)
	}
}

// TestMVMWithUncertainty_ErrorCases checks that invalid inputs propagate errors.
func TestMVMWithUncertainty_ErrorCases(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:       2,
		Cols:       2,
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Empty input.
	_, err = arr.MVMWithUncertainty([]float64{})
	if err == nil {
		t.Error("expected error for empty input")
	}

	// Oversized input.
	_, err = arr.MVMWithUncertainty([]float64{1, 2, 3})
	if err == nil {
		t.Error("expected error for oversized input")
	}

	// NaN input.
	_, err = arr.MVMWithUncertainty([]float64{math.NaN(), 0.5})
	if err == nil {
		t.Error("expected error for NaN input")
	}
}

// TestMVMWithUncertainty_UncertaintyFormula validates the error propagation
// formula against a manual calculation for a simple 1x1 array.
func TestMVMWithUncertainty_UncertaintyFormula(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows:       1,
		Cols:       1,
		NoiseLevel: 0.05, // 5% device sigma
		ADCBits:    10,
		DACBits:    10,
	})
	if err != nil {
		t.Fatal(err)
	}

	gRaw := 0.6
	if err := arr.ProgramWeight(0, 0, gRaw); err != nil {
		t.Fatal(err)
	}

	// ProgramWeight quantizes to 30 discrete levels, so the stored
	// conductance is QuantizeToLevels(0.6) = round(0.6*29)/29 = 17/29.
	gVal := QuantizeToLevels(gRaw)

	inputVal := 0.7
	input := []float64{inputVal}
	result, err := arr.MVMWithUncertainty(input)
	if err != nil {
		t.Fatal(err)
	}

	// Manual calculation using the quantized conductance:
	// dacLevels = 2^10 = 1024
	// LSB = 1/1023
	// dacSigma = LSB / sqrt(12)
	// dacIn = round(0.7 * 1023) / 1023 (quantized input)
	// maxCurrent = 1 (cols=1)
	// deviceSigma = 0.05
	//
	// term1 = gVal * dacSigma / maxCurrent
	// term2 = dacIn * (gVal * deviceSigma) / maxCurrent
	// sigma_out = sqrt(term1^2 + term2^2)

	dacLevels := 1 << 10
	lsb := 1.0 / float64(dacLevels-1)
	dacSigma := lsb / math.Sqrt(12.0)
	dacIn := math.Round(inputVal*float64(dacLevels-1)) / float64(dacLevels-1)
	maxCurrent := 1.0

	term1 := gVal * dacSigma / maxCurrent
	term2 := dacIn * (gVal * 0.05) / maxCurrent
	expected := math.Sqrt(term1*term1 + term2*term2)

	got := result.Uncertainty[0]
	relErr := math.Abs(got-expected) / expected
	if relErr > 1e-10 {
		t.Errorf("uncertainty[0] = %.12e, want %.12e (relErr = %.2e)", got, expected, relErr)
	}
}
