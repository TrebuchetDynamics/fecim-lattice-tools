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

// TestDACConvertWithNonlinearity verifies DAC nonlinearity application.
func TestDACConvertWithNonlinearity(t *testing.T) {
	dac := DefaultDAC()
	dac.INL = 0.5
	dac.DNL = 0.5

	level := 15
	idealV := dac.Convert(level)
	noisyV := dac.ConvertWithNonlinearity(level)

	if math.Abs(noisyV-idealV) < 1e-9 {
		t.Errorf("DAC nonlinearity should change output, but got same as ideal: %.4fV", idealV)
	}

	lsb := dac.Resolution()
	if math.Abs(noisyV-idealV) > 2.0*lsb {
		t.Errorf("DAC nonlinearity deviation too large: ideal %.4fV, noisy %.4fV (Δ=%.2f LSB)",
			idealV, noisyV, math.Abs(noisyV-idealV)/lsb)
	}
}

// TestADCConvertWithNonlinearity verifies ADC nonlinearity effects.
func TestADCConvertWithNonlinearity(t *testing.T) {
	adc := DefaultADC()
	adc.INL = 1.0
	adc.DNL = 1.0

	v := (adc.VrefHigh + adc.VrefLow) / 2.0
	idealL := adc.Convert(v)
	noisyL := adc.ConvertWithNonlinearity(v)

	if abs(noisyL-idealL) > 2 {
		t.Errorf("ADC nonlinearity caused too large deviation: ideal %d, noisy %d", idealL, noisyL)
	}
}

// TestADCTheoreticalSNR verifies SNR calculation.
func TestADCTheoreticalSNR(t *testing.T) {
	adc := DefaultADC()
	adc.Bits = 5
	snr := adc.TheoreticalSNR()
	expected := 6.02*5.0 + 1.76
	if math.Abs(snr-expected) > 0.01 {
		t.Errorf("Expected TheoreticalSNR %.2f, got %.2f", expected, snr)
	}
}

// TestADCEffectiveSNR verifies ENOB-based SNR.
func TestADCEffectiveSNR(t *testing.T) {
	adc := DefaultADC()
	adc.INL = 0.5
	adc.DNL = 0.5

	tSNR := adc.TheoreticalSNR()
	eSNR := adc.EffectiveSNR()

	if eSNR >= tSNR {
		t.Errorf("Effective SNR (%.2f) should be less than Theoretical SNR (%.2f) when nonlinearity exists", eSNR, tSNR)
	}

	adc.INL = 0
	adc.DNL = 0
	if math.Abs(adc.EffectiveSNR()-adc.TheoreticalSNR()) > 0.01 {
		t.Errorf("Ideal ADC should have EffectiveSNR == TheoreticalSNR")
	}
}

// TestTIAConvertWithNoise verifies TIA noise injection.
func TestTIAConvertWithNoise(t *testing.T) {
	tia := DefaultTIA()

	current := 10e-6
	vIdeal := tia.Convert(current)

	const iterations = 10
	for i := 0; i < iterations; i++ {
		v := tia.ConvertWithNoise(current)
		if math.Abs(v-vIdeal) > 0.1 {
			t.Errorf("TIA mean voltage (%.4fV) deviated too far from ideal (%.4fV)", v, vIdeal)
		}
	}
}

// TestTIASNR verifies TIA signal-to-noise ratio.
func TestTIASNR(t *testing.T) {
	tia := DefaultTIA()

	snrLow := tia.SNR(1e-7)
	snrHigh := tia.SNR(1e-5)

	if snrHigh <= snrLow {
		t.Errorf("Higher current should have better SNR: SNR(10uA)=%.2fdB, SNR(0.1uA)=%.2fdB", snrHigh, snrLow)
	}
}

// TestTIAMinDetectableCurrent verifies sensitivity.
func TestTIAMinDetectableCurrent(t *testing.T) {
	tia := DefaultTIA()
	minI := tia.MinDetectableCurrent()
	if minI <= 0 {
		t.Errorf("Min detectable current should be positive, got %.2e", minI)
	}
}

// TestTIADynamicRange verifies dynamic range.
func TestTIADynamicRange(t *testing.T) {
	tia := DefaultTIA()
	dr := tia.DynamicRange()
	if dr < 50 || dr > 120 {
		t.Errorf("Reasonable Dynamic Range expected, got %.2fdB", dr)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
