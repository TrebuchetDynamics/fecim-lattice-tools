package peripherals

import (
	"math"
	"testing"
)

// TestDACConversion verifies DAC converts levels to correct voltages.
func TestDACConversion(t *testing.T) {
	dac := DefaultDAC()

	// Test endpoints
	minV := dac.Convert(0)
	maxV := dac.Convert(dac.Levels() - 1)

	if math.Abs(minV-dac.VrefLow) > 0.001 {
		t.Errorf("Level 0 should be VrefLow (%.3f), got %.3f", dac.VrefLow, minV)
	}
	if math.Abs(maxV-dac.VrefHigh) > 0.001 {
		t.Errorf("Max level should be VrefHigh (%.3f), got %.3f", dac.VrefHigh, maxV)
	}

	// Test monotonicity
	prevV := dac.Convert(0)
	for level := 1; level < dac.Levels(); level++ {
		v := dac.Convert(level)
		if v <= prevV {
			t.Errorf("DAC not monotonic: level %d (%.4fV) <= level %d (%.4fV)",
				level, v, level-1, prevV)
		}
		prevV = v
	}
}

// TestDACLevels verifies 30 FeCIM levels.
func TestDACLevels(t *testing.T) {
	dac := DefaultDAC()

	if dac.Levels() < 30 {
		t.Errorf("DAC should support at least 30 levels, has %d", dac.Levels())
	}
}

// TestADCConversion verifies ADC converts voltages to correct levels.
func TestADCConversion(t *testing.T) {
	adc := DefaultADC()

	// Test endpoints
	minLevel := adc.Convert(adc.VrefLow)
	maxLevel := adc.Convert(adc.VrefHigh)

	if minLevel != 0 {
		t.Errorf("VrefLow should convert to 0, got %d", minLevel)
	}
	if maxLevel != adc.Levels()-1 {
		t.Errorf("VrefHigh should convert to %d, got %d", adc.Levels()-1, maxLevel)
	}

	// Test clamping
	belowMin := adc.Convert(adc.VrefLow - 1.0)
	aboveMax := adc.Convert(adc.VrefHigh + 1.0)

	if belowMin != 0 {
		t.Errorf("Below VrefLow should clamp to 0, got %d", belowMin)
	}
	if aboveMax != adc.Levels()-1 {
		t.Errorf("Above VrefHigh should clamp to %d, got %d", adc.Levels()-1, aboveMax)
	}
}

// TestADCENOB verifies ENOB calculation.
func TestADCENOB(t *testing.T) {
	adc := DefaultADC()

	enob := adc.ENOB()
	if enob <= 0 || enob > float64(adc.Bits) {
		t.Errorf("ENOB %.2f should be between 0 and %d bits", enob, adc.Bits)
	}

	// With ideal (0 INL/DNL), ENOB should equal bits
	idealADC := DefaultADC()
	idealADC.INL = 0
	idealADC.DNL = 0
	idealENOB := idealADC.ENOB()

	if math.Abs(idealENOB-float64(idealADC.Bits)) > 0.01 {
		t.Errorf("Ideal ADC ENOB should be %d, got %.2f", idealADC.Bits, idealENOB)
	}
}

// TestTIAConversion verifies current-to-voltage conversion.
func TestTIAConversion(t *testing.T) {
	tia := DefaultTIA()

	// Test zero current (should equal offset)
	v0 := tia.Convert(0)
	if math.Abs(v0-tia.OutputOffset) > 0.001 {
		t.Errorf("Zero current should give offset (%.3fV), got %.3fV", tia.OutputOffset, v0)
	}

	// Test linearity
	i1 := 10e-6
	i2 := 20e-6
	v1 := tia.Convert(i1)
	v2 := tia.Convert(i2)

	expectedRatio := 2.0
	actualRatio := (v2 - tia.OutputOffset) / (v1 - tia.OutputOffset)

	if math.Abs(actualRatio-expectedRatio) > 0.01 {
		t.Errorf("TIA should be linear: 2x current should give 2x voltage (got %.2fx)", actualRatio)
	}
}

// TestTIAClamping verifies output clamping.
func TestTIAClamping(t *testing.T) {
	tia := DefaultTIA()

	// High current should clamp
	vMax := tia.Convert(1.0) // 1 A would saturate any TIA
	if vMax > tia.MaxOutputVoltage {
		t.Errorf("TIA should clamp at %.2fV, got %.2fV", tia.MaxOutputVoltage, vMax)
	}
}

// TestChargePumpBoost verifies voltage boost.
func TestChargePumpBoost(t *testing.T) {
	pump := DefaultChargePump()

	idealV := pump.IdealOutputVoltage()
	actualV := pump.ActualOutputVoltage()

	// Ideal should be (N+1) * Vin
	expectedIdeal := float64(pump.Stages+1) * pump.InputVoltage
	if math.Abs(idealV-expectedIdeal) > 0.01 {
		t.Errorf("Ideal output should be %.2fV, got %.2fV", expectedIdeal, idealV)
	}

	// Actual should be less than ideal
	if actualV >= idealV {
		t.Errorf("Actual output (%.2fV) should be less than ideal (%.2fV)", actualV, idealV)
	}

	// Should still boost
	if actualV <= pump.InputVoltage {
		t.Errorf("Output (%.2fV) should exceed input (%.2fV)", actualV, pump.InputVoltage)
	}
}

// TestChargePumpEfficiency verifies energy calculations.
func TestChargePumpEfficiency(t *testing.T) {
	pump := DefaultChargePump()

	pIn := pump.PowerInput()
	pOut := pump.PowerOutput()
	pLoss := pump.PowerLoss()

	// Power balance
	if math.Abs(pIn-pOut-pLoss) > 1e-15 {
		t.Errorf("Power balance: Pin (%.2e) != Pout (%.2e) + Ploss (%.2e)",
			pIn, pOut, pLoss)
	}

	// Efficiency check
	calculatedEff := pOut / pIn
	if math.Abs(calculatedEff-pump.Efficiency) > 0.01 {
		t.Errorf("Efficiency mismatch: calculated %.2f, specified %.2f",
			calculatedEff, pump.Efficiency)
	}
}

// TestDACToADCRoundTrip verifies end-to-end conversion.
func TestDACToADCRoundTrip(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()

	// Set ADC range to match DAC output
	adc.VrefLow = dac.VrefLow
	adc.VrefHigh = dac.VrefHigh

	// Test round-trip for several levels
	testLevels := []int{0, 7, 15, 22, 29}
	for _, level := range testLevels {
		if level >= dac.Levels() || level >= adc.Levels() {
			continue
		}

		voltage := dac.Convert(level)
		recoveredLevel := adc.Convert(voltage)

		// Allow ±1 level due to quantization
		if abs(recoveredLevel-level) > 1 {
			t.Errorf("Round-trip for level %d: got %d (Δ=%d)",
				level, recoveredLevel, abs(recoveredLevel-level))
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
