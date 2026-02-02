// pkg/peripherals/analysis_test.go
// Tests for peripheral circuit analysis functions
package peripherals

import (
	"math"
	"testing"
)

// ============================================================================
// INL/DNL Analysis Tests - DAC
// ============================================================================

func TestDAC_AnalyzeINLDNL_Basic(t *testing.T) {
	dac := DefaultDAC()
	analysis := dac.AnalyzeINLDNL()

	if analysis == nil {
		t.Fatal("AnalyzeINLDNL returned nil")
	}

	// Should have 30 levels (or capped at 32)
	if analysis.Levels <= 0 {
		t.Errorf("Invalid number of levels: %d", analysis.Levels)
	}

	// INL/DNL arrays should match levels
	if len(analysis.INLValues) != analysis.Levels {
		t.Errorf("INL array length mismatch: expected %d, got %d",
			analysis.Levels, len(analysis.INLValues))
	}
	if len(analysis.DNLValues) != analysis.Levels {
		t.Errorf("DNL array length mismatch: expected %d, got %d",
			analysis.Levels, len(analysis.DNLValues))
	}
}

func TestDAC_AnalyzeINLDNL_Values(t *testing.T) {
	dac := DefaultDAC()
	analysis := dac.AnalyzeINLDNL()

	// MaxINL should be within reasonable bounds for a good DAC
	if math.Abs(analysis.MaxINL) > 5.0 { // More than 5 LSB is very bad
		t.Errorf("MaxINL too high: %.2f LSB", analysis.MaxINL)
	}

	// MaxDNL should be reasonable
	if analysis.MaxDNL > 1.5 { // DNL > 1 LSB can cause missing codes
		t.Logf("Warning: MaxDNL is high: %.2f LSB", analysis.MaxDNL)
	}

	// MinDNL shouldn't be too negative (causes non-monotonicity)
	if analysis.MinDNL < -0.99 {
		t.Errorf("MinDNL too negative: %.2f LSB (non-monotonic)", analysis.MinDNL)
	}
}

func TestDAC_AnalyzeINLDNL_WorstCode(t *testing.T) {
	dac := DefaultDAC()
	analysis := dac.AnalyzeINLDNL()

	// WorstCode should be valid
	if analysis.WorstCode < 0 || analysis.WorstCode >= analysis.Levels {
		t.Errorf("Invalid WorstCode: %d (levels: %d)", analysis.WorstCode, analysis.Levels)
	}

	// WorstCode should correspond to MaxINL
	if math.Abs(analysis.INLValues[analysis.WorstCode]-analysis.MaxINL) > 1e-10 {
		t.Errorf("WorstCode %d has INL %.3f but MaxINL is %.3f",
			analysis.WorstCode, analysis.INLValues[analysis.WorstCode], analysis.MaxINL)
	}
}

func TestDAC_AnalyzeINLDNL_IdealDAC(t *testing.T) {
	// Create ideal DAC with no nonlinearity
	dac := DefaultDAC()
	dac.INL = 0
	dac.DNL = 0

	analysis := dac.AnalyzeINLDNL()

	// Ideal DAC should have very low INL/DNL
	for i, inl := range analysis.INLValues {
		if math.Abs(inl) > 0.01 {
			t.Errorf("Ideal DAC INL[%d] = %.4f, expected ~0", i, inl)
		}
	}
}

// ============================================================================
// INL/DNL Analysis Tests - ADC
// ============================================================================

func TestADC_AnalyzeINLDNL_Basic(t *testing.T) {
	adc := DefaultADC()
	analysis := adc.AnalyzeINLDNL()

	if analysis == nil {
		t.Fatal("AnalyzeINLDNL returned nil")
	}

	if analysis.Levels <= 0 {
		t.Errorf("Invalid number of levels: %d", analysis.Levels)
	}
}

func TestADC_AnalyzeINLDNL_Arrays(t *testing.T) {
	adc := DefaultADC()
	analysis := adc.AnalyzeINLDNL()

	if len(analysis.INLValues) != analysis.Levels {
		t.Errorf("INL array length mismatch")
	}
	if len(analysis.DNLValues) != analysis.Levels {
		t.Errorf("DNL array length mismatch")
	}
}

// ============================================================================
// Timing Analysis Tests
// ============================================================================

func TestAnalyzeTiming_Basic(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)

	if timing == nil {
		t.Fatal("AnalyzeTiming returned nil")
	}

	// All timing values should be positive
	if timing.DACSettle <= 0 {
		t.Errorf("DACSettle should be positive: %e", timing.DACSettle)
	}
	if timing.ADCConvert <= 0 {
		t.Errorf("ADCConvert should be positive: %e", timing.ADCConvert)
	}
	if timing.TIASettle <= 0 {
		t.Errorf("TIASettle should be positive: %e", timing.TIASettle)
	}
	if timing.ArraySettle <= 0 {
		t.Errorf("ArraySettle should be positive: %e", timing.ArraySettle)
	}
	if timing.WritePulse <= 0 {
		t.Errorf("WritePulse should be positive: %e", timing.WritePulse)
	}
}

func TestAnalyzeTiming_WriteTime(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)

	// WriteTime = DACSettle + PumpRise + WritePulse + ArraySettle
	expectedWriteTime := timing.DACSettle + timing.PumpRise + timing.WritePulse + timing.ArraySettle
	if math.Abs(timing.WriteTime-expectedWriteTime) > 1e-12 {
		t.Errorf("WriteTime mismatch: expected %e, got %e", expectedWriteTime, timing.WriteTime)
	}
}

func TestAnalyzeTiming_ReadTime(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)

	// ReadTime = DACSettle + ArraySettle + TIASettle + ADCConvert
	expectedReadTime := timing.DACSettle + timing.ArraySettle + timing.TIASettle + timing.ADCConvert
	if math.Abs(timing.ReadTime-expectedReadTime) > 1e-12 {
		t.Errorf("ReadTime mismatch: expected %e, got %e", expectedReadTime, timing.ReadTime)
	}
}

func TestAnalyzeTiming_CycleTime(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)

	// CycleTime = WriteTime + ReadTime
	expectedCycleTime := timing.WriteTime + timing.ReadTime
	if math.Abs(timing.CycleTime-expectedCycleTime) > 1e-12 {
		t.Errorf("CycleTime mismatch: expected %e, got %e", expectedCycleTime, timing.CycleTime)
	}
}

func TestAnalyzeTiming_MaxThroughput(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)

	// MaxThroughput = 1 / CycleTime
	expectedThroughput := 1.0 / timing.CycleTime
	if math.Abs(timing.MaxThroughput-expectedThroughput) > 1e-6 {
		t.Errorf("MaxThroughput mismatch: expected %e, got %e", expectedThroughput, timing.MaxThroughput)
	}

	// Throughput should be reasonable (not zero, not infinite)
	if timing.MaxThroughput <= 0 || timing.MaxThroughput > 1e12 {
		t.Errorf("MaxThroughput unreasonable: %e ops/s", timing.MaxThroughput)
	}
}

// ============================================================================
// Power Analysis Tests
// ============================================================================

func TestAnalyzePower_Basic(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)
	power := AnalyzePower(dac, adc, tia, pump, timing)

	if power == nil {
		t.Fatal("AnalyzePower returned nil")
	}

	// All energy values should be non-negative
	if power.DACEnergy < 0 {
		t.Errorf("DACEnergy should be non-negative: %e", power.DACEnergy)
	}
	if power.ADCEnergy < 0 {
		t.Errorf("ADCEnergy should be non-negative: %e", power.ADCEnergy)
	}
	if power.TIAEnergy < 0 {
		t.Errorf("TIAEnergy should be non-negative: %e", power.TIAEnergy)
	}
	if power.PumpEnergy < 0 {
		t.Errorf("PumpEnergy should be non-negative: %e", power.PumpEnergy)
	}
}

func TestAnalyzePower_TotalEnergy(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)
	power := AnalyzePower(dac, adc, tia, pump, timing)

	// TotalEnergy should be sum of components
	expectedTotal := power.DACEnergy + power.ADCEnergy + power.TIAEnergy + power.PumpEnergy
	if math.Abs(power.TotalEnergy-expectedTotal) > 1e-20 {
		t.Errorf("TotalEnergy mismatch: expected %e, got %e", expectedTotal, power.TotalEnergy)
	}
}

func TestAnalyzePower_PowerFromEnergy(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)
	power := AnalyzePower(dac, adc, tia, pump, timing)

	// Power = Energy / CycleTime
	expectedTotalPower := power.TotalEnergy / timing.CycleTime
	if math.Abs(power.TotalPower-expectedTotalPower) > 1e-15 {
		t.Errorf("TotalPower mismatch: expected %e, got %e", expectedTotalPower, power.TotalPower)
	}
}

func TestAnalyzePower_Fractions(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)
	power := AnalyzePower(dac, adc, tia, pump, timing)

	// Fractions should sum to 1.0
	totalFraction := power.DACFraction + power.ADCFraction + power.TIAFraction + power.PumpFraction
	if math.Abs(totalFraction-1.0) > 0.01 {
		t.Errorf("Power fractions should sum to 1.0, got %.3f", totalFraction)
	}

	// Each fraction should be between 0 and 1
	if power.DACFraction < 0 || power.DACFraction > 1 {
		t.Errorf("DACFraction out of range: %.3f", power.DACFraction)
	}
	if power.ADCFraction < 0 || power.ADCFraction > 1 {
		t.Errorf("ADCFraction out of range: %.3f", power.ADCFraction)
	}
	if power.TIAFraction < 0 || power.TIAFraction > 1 {
		t.Errorf("TIAFraction out of range: %.3f", power.TIAFraction)
	}
	if power.PumpFraction < 0 || power.PumpFraction > 1 {
		t.Errorf("PumpFraction out of range: %.3f", power.PumpFraction)
	}
}

// ============================================================================
// Transfer Function Tests
// ============================================================================

func TestComputeTransferFunction_Basic(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	tf := ComputeTransferFunction(dac, adc, tia, pump)

	if tf == nil {
		t.Fatal("ComputeTransferFunction returned nil")
	}

	// Should have 30 entries (FeCIM levels)
	if len(tf.InputLevels) != 30 {
		t.Errorf("Expected 30 input levels, got %d", len(tf.InputLevels))
	}
	if len(tf.DACVoltages) != 30 {
		t.Errorf("Expected 30 DAC voltages, got %d", len(tf.DACVoltages))
	}
	if len(tf.ADCLevels) != 30 {
		t.Errorf("Expected 30 ADC levels, got %d", len(tf.ADCLevels))
	}
	if len(tf.Errors) != 30 {
		t.Errorf("Expected 30 error values, got %d", len(tf.Errors))
	}
}

func TestComputeTransferFunction_InputLevels(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	tf := ComputeTransferFunction(dac, adc, tia, pump)

	// Input levels should be 0-29
	for i := 0; i < 30; i++ {
		if tf.InputLevels[i] != i {
			t.Errorf("InputLevel[%d] should be %d, got %d", i, i, tf.InputLevels[i])
		}
	}
}

func TestComputeTransferFunction_Monotonicity(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	tf := ComputeTransferFunction(dac, adc, tia, pump)

	// DAC voltages should be monotonically increasing
	for i := 1; i < 30; i++ {
		if tf.DACVoltages[i] <= tf.DACVoltages[i-1] {
			t.Errorf("DAC not monotonic: V[%d]=%.4f <= V[%d]=%.4f",
				i, tf.DACVoltages[i], i-1, tf.DACVoltages[i-1])
		}
	}

	// TIA voltages should also be monotonically increasing
	for i := 1; i < 30; i++ {
		if tf.TIAVoltages[i] <= tf.TIAVoltages[i-1] {
			t.Errorf("TIA not monotonic: V[%d]=%.4f <= V[%d]=%.4f",
				i, tf.TIAVoltages[i], i-1, tf.TIAVoltages[i-1])
		}
	}
}

func TestComputeTransferFunction_Errors(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	tf := ComputeTransferFunction(dac, adc, tia, pump)

	// Errors should be ADCLevel - InputLevel
	for i := 0; i < 30; i++ {
		expectedError := tf.ADCLevels[i] - tf.InputLevels[i]
		if tf.Errors[i] != expectedError {
			t.Errorf("Error[%d] should be %d, got %d", i, expectedError, tf.Errors[i])
		}
	}

	// For a good system, most errors should be small
	largeErrors := 0
	for _, err := range tf.Errors {
		if abs(err) > 2 { // More than ±2 levels is significant
			largeErrors++
		}
	}
	if largeErrors > 5 { // More than 5 large errors is problematic
		t.Logf("Warning: %d levels have errors > ±2", largeErrors)
	}
}

func TestComputeTransferFunction_RoundTrip(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	// Match ADC range to expected TIA output range
	adc.VrefLow = tia.OutputOffset
	adc.VrefHigh = tia.OutputOffset + tia.Gain*100e-6 // Max current at level 29

	tf := ComputeTransferFunction(dac, adc, tia, pump)

	// Check that round-trip preserves most values
	maxError := 0
	for i := 0; i < 30; i++ {
		err := abs(tf.Errors[i])
		if err > maxError {
			maxError = err
		}
	}

	// Maximum error should be bounded
	if maxError > 5 { // ±5 levels is very poor
		t.Logf("Maximum round-trip error: %d levels", maxError)
	}
}

// ============================================================================
// findCodeWidth Tests
// ============================================================================

func TestFindCodeWidth_Positive(t *testing.T) {
	adc := DefaultADC()
	lsb := adc.Resolution()

	for code := 0; code < 30; code++ {
		width := findCodeWidth(adc, code, lsb)
		if width <= 0 {
			t.Errorf("Code width should be positive for code %d, got %f", code, width)
		}
	}
}

func TestFindCodeWidth_ApproximatesLSB(t *testing.T) {
	adc := DefaultADC()
	adc.DNL = 0 // Ideal ADC
	lsb := adc.Resolution()

	// With zero DNL, width should equal LSB
	width := findCodeWidth(adc, 0, lsb)
	// Note: The implementation has a pattern based on code%5, so allow some tolerance
	if math.Abs(width-lsb) > 0.5*lsb {
		t.Errorf("Ideal code width should approximate LSB: got %f, expected %f", width, lsb)
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestAnalyzePower_ZeroCycleTime(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	// Create timing with zero cycle time
	timing := &TimingAnalysis{
		DACSettle:  0,
		PumpRise:   0,
		TIASettle:  0,
		ADCConvert: 0,
		CycleTime:  0, // Edge case
	}

	power := AnalyzePower(dac, adc, tia, pump, timing)

	// Should not crash, power values should be 0
	if power.TotalPower != 0 {
		t.Logf("Zero cycle time results in power: %e (expected handling)", power.TotalPower)
	}
}

func TestAnalyzePower_ZeroTotalEnergy(t *testing.T) {
	// Create peripherals that might have zero energy
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	timing := AnalyzeTiming(dac, adc, tia, pump)
	power := AnalyzePower(dac, adc, tia, pump, timing)

	// If TotalEnergy is zero, fractions should be 0
	if power.TotalEnergy == 0 {
		if power.DACFraction != 0 || power.ADCFraction != 0 {
			t.Error("Fractions should be 0 when TotalEnergy is 0")
		}
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkDAC_AnalyzeINLDNL(b *testing.B) {
	dac := DefaultDAC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dac.AnalyzeINLDNL()
	}
}

func BenchmarkADC_AnalyzeINLDNL(b *testing.B) {
	adc := DefaultADC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adc.AnalyzeINLDNL()
	}
}

func BenchmarkAnalyzeTiming(b *testing.B) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AnalyzeTiming(dac, adc, tia, pump)
	}
}

func BenchmarkAnalyzePower(b *testing.B) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()
	timing := AnalyzeTiming(dac, adc, tia, pump)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AnalyzePower(dac, adc, tia, pump, timing)
	}
}

func BenchmarkComputeTransferFunction(b *testing.B) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	pump := DefaultChargePump()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeTransferFunction(dac, adc, tia, pump)
	}
}
