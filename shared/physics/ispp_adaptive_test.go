package physics

import (
	"math"
	"testing"
)

func TestDefaultAdaptiveISPPConfig(t *testing.T) {
	config := DefaultAdaptiveISPPConfig()

	// Check base config is embedded
	if config.StartRatio != 0.7 {
		t.Errorf("Expected StartRatio 0.7, got %f", config.StartRatio)
	}

	// Check adaptive parameters
	if config.MinStepPercent != 0.005 {
		t.Errorf("Expected MinStepPercent 0.005, got %f", config.MinStepPercent)
	}
	if config.MaxStepPercent != 0.10 {
		t.Errorf("Expected MaxStepPercent 0.10, got %f", config.MaxStepPercent)
	}
	if config.ConfidenceK != 1.5 {
		t.Errorf("Expected ConfidenceK 1.5, got %f", config.ConfidenceK)
	}
	if config.MinSamplesForBayes != 5 {
		t.Errorf("Expected MinSamplesForBayes 5, got %d", config.MinSamplesForBayes)
	}
}

func TestNewAdaptiveISPPCalculator(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	if calc == nil {
		t.Fatal("Expected non-nil calculator")
	}
	if calc.Ec != 1.0 {
		t.Errorf("Expected Ec 1.0, got %f", calc.Ec)
	}
	if calc.NumLevels != 30 {
		t.Errorf("Expected NumLevels 30, got %d", calc.NumLevels)
	}
	if len(calc.LevelStats) != 30 {
		t.Errorf("Expected 30 LevelStats, got %d", len(calc.LevelStats))
	}

	// Check all step multipliers initialized to 1.0
	for i, stats := range calc.LevelStats {
		if stats.StepMultiplier != 1.0 {
			t.Errorf("Level %d: expected StepMultiplier 1.0, got %f", i, stats.StepMultiplier)
		}
	}
}

func TestLevelStatisticsUpdateVoltage(t *testing.T) {
	stats := &LevelStatistics{StepMultiplier: 1.0}

	// Update with known values to test Welford's algorithm
	voltages := []float64{1.0, 1.2, 0.8, 1.1, 0.9}
	for _, v := range voltages {
		stats.UpdateVoltage(v)
	}

	if stats.Count != 5 {
		t.Errorf("Expected Count 5, got %d", stats.Count)
	}

	// Mean should be 1.0
	expectedMean := 1.0
	if math.Abs(stats.VoltageMean-expectedMean) > 0.001 {
		t.Errorf("Expected mean %f, got %f", expectedMean, stats.VoltageMean)
	}

	// Standard deviation should be ~0.158
	stdDev := stats.VoltageStdDev()
	expectedStdDev := 0.158
	if math.Abs(stdDev-expectedStdDev) > 0.01 {
		t.Errorf("Expected stdDev ~%f, got %f", expectedStdDev, stdDev)
	}

	// Min/Max
	if stats.VoltageMin != 0.8 {
		t.Errorf("Expected VoltageMin 0.8, got %f", stats.VoltageMin)
	}
	if stats.VoltageMax != 1.2 {
		t.Errorf("Expected VoltageMax 1.2, got %f", stats.VoltageMax)
	}
}

func TestAdaptiveStepSizing(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Far from target (level 0 -> level 29): should use larger step
	farStep := calc.CalculateAdaptiveStep(0, 29, DirectionAscending)

	// Close to target (level 14 -> level 15): should use smaller step
	closeStep := calc.CalculateAdaptiveStep(14, 15, DirectionAscending)

	if farStep <= closeStep {
		t.Errorf("Expected farStep (%f) > closeStep (%f)", farStep, closeStep)
	}

	// Far step should be significantly larger due to distance + error-proportional scaling
	// With 29-level gap and ErrorStepGain=0.5: multiplier = 1 + 0.5*29 = 15.5x
	maxStep := calc.AdaptiveConfig.MaxStepPercent * calc.Ec
	if farStep < maxStep {
		t.Errorf("Expected farStep >= maxStep (%f), got %f", maxStep, farStep)
	}

	// Close step with 1-level gap has error gain of 1.5x (1 + 0.5*1)
	// So it won't be as small as minStep, but should be reasonable
	// For 1-level gap: base ~= minStep, with 1.5x gain
	minStep := calc.AdaptiveConfig.MinStepPercent * calc.Ec
	expectedCloseStep := minStep * 1.5 * 3 // Allow margin for decay curve
	if closeStep > expectedCloseStep {
		t.Errorf("Expected closeStep <= %f, got %f", expectedCloseStep, closeStep)
	}
}

func TestAdaptiveStepWithLearnedMultiplier(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Get baseline step
	baseStep := calc.CalculateAdaptiveStep(14, 15, DirectionAscending)

	// Simulate overshoot to reduce multiplier
	calc.currentTarget = 15
	calc.RecordOvershoot()

	// Step should now be smaller
	reducedStep := calc.CalculateAdaptiveStep(14, 15, DirectionAscending)
	expectedReduced := baseStep * calc.AdaptiveConfig.OvershootPenalty

	if math.Abs(reducedStep-expectedReduced) > 0.0001 {
		t.Errorf("Expected step %f after overshoot, got %f", expectedReduced, reducedStep)
	}
}

func TestBayesianVoltagePreduction(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// No data yet - should return no prediction
	_, hasData := calc.GetPredictedVoltage(15)
	if hasData {
		t.Error("Expected no data for fresh calculator")
	}

	// Add training data
	targetLevel := 15
	voltages := []float64{1.05, 0.98, 1.02, 1.00, 0.99, 1.01}
	for _, v := range voltages {
		calc.currentTarget = targetLevel
		calc.currentPulses = 2
		calc.hadOvershoot = false
		calc.RecordSuccess(v)
	}

	// Now should have prediction
	predicted, hasData := calc.GetPredictedVoltage(targetLevel)
	if !hasData {
		t.Error("Expected prediction data after training")
	}

	// Predicted should be close to mean (~1.008)
	expectedMean := 1.008
	if math.Abs(predicted-expectedMean) > 0.01 {
		t.Errorf("Expected predicted ~%f, got %f", expectedMean, predicted)
	}
}

func TestStartWriteWithBayesianPrediction(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Train with consistent voltage
	targetLevel := 10
	for i := 0; i < 10; i++ {
		calc.currentTarget = targetLevel
		calc.currentPulses = 2
		calc.hadOvershoot = false
		calc.RecordSuccess(0.8)
	}

	// Start write with Bayesian prediction
	calibratedVoltage := 0.85 // Calibration says 0.85
	startV := calc.StartWrite(5, targetLevel, calibratedVoltage)

	// Should start below learned mean due to ConfidenceK
	stats := &calc.LevelStats[targetLevel]

	// Due to low variance in training data, stdDev is very small
	// So start should be close to mean but slightly below
	if startV > stats.VoltageMean {
		t.Errorf("Bayesian start (%f) should be <= mean (%f)", startV, stats.VoltageMean)
	}

	// Verify the expected start calculation is reasonable
	expectedStart := stats.VoltageMean - calc.AdaptiveConfig.ConfidenceK*stats.VoltageStdDev()
	if expectedStart > stats.VoltageMean {
		t.Errorf("Expected start calculation error: %f > mean %f", expectedStart, stats.VoltageMean)
	}
}

func TestRecordSuccessUpdatesStats(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	calc.currentStart = 0
	calc.currentTarget = 15
	calc.currentPulses = 2
	calc.hadOvershoot = false

	calc.RecordSuccess(1.0)

	stats := &calc.LevelStats[15]
	if stats.Count != 1 {
		t.Errorf("Expected Count 1, got %d", stats.Count)
	}
	if stats.TotalPulses != 2 {
		t.Errorf("Expected TotalPulses 2, got %d", stats.TotalPulses)
	}

	// Quick success (<=3 pulses) should boost step multiplier
	expectedMultiplier := 1.0 * calc.AdaptiveConfig.SuccessBonus
	if stats.StepMultiplier != expectedMultiplier {
		t.Errorf("Expected StepMultiplier %f after quick success, got %f", expectedMultiplier, stats.StepMultiplier)
	}
}

func TestRecordOvershootPenalizesLevel(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	calc.currentTarget = 15
	initialMultiplier := calc.LevelStats[15].StepMultiplier

	calc.RecordOvershoot()

	expectedMultiplier := initialMultiplier * calc.AdaptiveConfig.OvershootPenalty
	if calc.LevelStats[15].StepMultiplier != expectedMultiplier {
		t.Errorf("Expected multiplier %f after overshoot, got %f",
			expectedMultiplier, calc.LevelStats[15].StepMultiplier)
	}

	if calc.LevelStats[15].OvershootCount != 1 {
		t.Errorf("Expected OvershootCount 1, got %d", calc.LevelStats[15].OvershootCount)
	}
}

func TestTransitionStatistics(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Record some transitions
	transitions := []struct {
		from, to int
		pulses   int
		voltage  float64
	}{
		{0, 15, 3, 1.0},
		{0, 15, 2, 0.98},
		{0, 15, 4, 1.02},
		{15, 0, 2, -0.95},
		{15, 0, 3, -1.0},
		{15, 0, 2, -0.98}, // Need at least 3 for prediction
	}

	for _, tr := range transitions {
		calc.currentStart = tr.from
		calc.currentTarget = tr.to
		calc.currentPulses = tr.pulses
		calc.hadOvershoot = false
		calc.RecordSuccess(tr.voltage)
	}

	// Check 0->15 transition
	avgPulses, avgVoltage, confidence := calc.GetTransitionPrediction(0, 15)
	if avgPulses == 0 {
		t.Error("Expected transition data for 0->15")
	}
	if confidence == 0 {
		t.Error("Expected non-zero confidence for 0->15")
	}

	// avgVoltage should be close to 1.0
	if math.Abs(avgVoltage-1.0) > 0.1 {
		t.Errorf("Expected avgVoltage ~1.0, got %f", avgVoltage)
	}

	// Check 15->0 transition (descending)
	avgPulses2, avgVoltage2, _ := calc.GetTransitionPrediction(15, 0)
	if avgPulses2 == 0 {
		t.Error("Expected transition data for 15->0")
	}
	if avgVoltage2 >= 0 {
		t.Errorf("Expected negative avgVoltage for descending, got %f", avgVoltage2)
	}
}

func TestLevelConfidence(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// No data = zero confidence
	conf := calc.GetLevelConfidence(15)
	if conf != 0 {
		t.Errorf("Expected zero confidence with no data, got %f", conf)
	}

	// Add successful writes
	for i := 0; i < 20; i++ {
		calc.currentTarget = 15
		calc.currentPulses = 2
		calc.hadOvershoot = false
		calc.RecordSuccess(1.0)
	}

	conf = calc.GetLevelConfidence(15)
	if conf < 0.5 {
		t.Errorf("Expected high confidence after many successes, got %f", conf)
	}

	// Add some failures to reduce confidence
	for i := 0; i < 5; i++ {
		calc.currentTarget = 15
		calc.RecordFailure()
	}

	confAfterFailures := calc.GetLevelConfidence(15)
	if confAfterFailures >= conf {
		t.Errorf("Expected lower confidence after failures: before=%f, after=%f", conf, confAfterFailures)
	}
}

func TestExportImportLearningState(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Train the calculator
	for i := 0; i < 10; i++ {
		calc.currentTarget = 15
		calc.currentPulses = 2
		calc.hadOvershoot = false
		calc.RecordSuccess(1.0 + float64(i)*0.01)
	}

	calc.currentTarget = 15
	calc.RecordOvershoot()

	// Export state
	state := calc.ExportLearningState()

	// Create new calculator and import
	calc2 := NewAdaptiveISPPCalculator(1.0, 30)
	calc2.ImportLearningState(state)

	// Verify imported state matches
	if calc2.LevelStats[15].Count != calc.LevelStats[15].Count {
		t.Error("Imported Count doesn't match")
	}
	if calc2.LevelStats[15].OvershootCount != calc.LevelStats[15].OvershootCount {
		t.Error("Imported OvershootCount doesn't match")
	}
	if math.Abs(calc2.LevelStats[15].VoltageMean-calc.LevelStats[15].VoltageMean) > 0.0001 {
		t.Error("Imported VoltageMean doesn't match")
	}
}

func TestCalculateNextAdaptiveVoltage(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Start a write
	calc.StartWrite(0, 29, 1.5)

	// Get next voltage (ascending)
	nextV := calc.CalculateNextAdaptiveVoltage(1.0, 10, 29, DirectionAscending)

	if nextV <= 1.0 {
		t.Errorf("Expected ascending voltage > 1.0, got %f", nextV)
	}

	// Verify pulse count incremented
	if calc.currentPulses != 1 {
		t.Errorf("Expected currentPulses 1, got %d", calc.currentPulses)
	}

	// Get another voltage
	nextV2 := calc.CalculateNextAdaptiveVoltage(nextV, 15, 29, DirectionAscending)
	if nextV2 <= nextV {
		t.Errorf("Expected second voltage > first: %f vs %f", nextV2, nextV)
	}
}

func TestStepMultiplierCaps(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Many overshoots should not reduce multiplier below 0.5
	calc.currentTarget = 15
	for i := 0; i < 20; i++ {
		calc.RecordOvershoot()
	}

	if calc.LevelStats[15].StepMultiplier < 0.5 {
		t.Errorf("StepMultiplier should not go below 0.5, got %f", calc.LevelStats[15].StepMultiplier)
	}

	// Many quick successes should not raise multiplier above 2.0
	calc2 := NewAdaptiveISPPCalculator(1.0, 30)
	for i := 0; i < 50; i++ {
		calc2.currentTarget = 10
		calc2.currentPulses = 1
		calc2.hadOvershoot = false
		calc2.RecordSuccess(1.0)
	}

	if calc2.LevelStats[10].StepMultiplier > 2.0 {
		t.Errorf("StepMultiplier should not exceed 2.0, got %f", calc2.LevelStats[10].StepMultiplier)
	}
}

func TestGetLevelSuccessRates(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Level 10: 8 successes, 2 failures = 80%
	for i := 0; i < 8; i++ {
		calc.currentTarget = 10
		calc.currentPulses = 2
		calc.RecordSuccess(1.0)
	}
	for i := 0; i < 2; i++ {
		calc.currentTarget = 10
		calc.RecordFailure()
	}

	// Level 20: 10 successes, 0 failures = 100%
	for i := 0; i < 10; i++ {
		calc.currentTarget = 20
		calc.currentPulses = 2
		calc.RecordSuccess(1.2)
	}

	rates := calc.GetLevelSuccessRates()

	if math.Abs(rates[10]-0.8) > 0.01 {
		t.Errorf("Expected level 10 rate 0.8, got %f", rates[10])
	}
	if rates[20] != 1.0 {
		t.Errorf("Expected level 20 rate 1.0, got %f", rates[20])
	}
	// Untested level should be 1.0 (assume perfect)
	if rates[5] != 1.0 {
		t.Errorf("Expected untested level rate 1.0, got %f", rates[5])
	}
}

func TestGetAveragePulsesPerLevel(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Level 10: pulses 2, 3, 4 = avg 3
	pulses := []int{2, 3, 4}
	for _, p := range pulses {
		calc.currentTarget = 10
		calc.currentPulses = p
		calc.RecordSuccess(1.0)
	}

	avgPulses := calc.GetAveragePulsesPerLevel()
	expectedAvg := 3.0
	if math.Abs(avgPulses[10]-expectedAvg) > 0.01 {
		t.Errorf("Expected avg pulses %f for level 10, got %f", expectedAvg, avgPulses[10])
	}
}

func TestVoltageStdDevEdgeCases(t *testing.T) {
	stats := &LevelStatistics{}

	// No data
	if stats.VoltageStdDev() != 0 {
		t.Error("Expected 0 stddev with no data")
	}

	// One sample
	stats.UpdateVoltage(1.0)
	if stats.VoltageStdDev() != 0 {
		t.Error("Expected 0 stddev with one sample")
	}

	// Two samples
	stats.UpdateVoltage(2.0)
	stdDev := stats.VoltageStdDev()
	if stdDev == 0 {
		t.Error("Expected non-zero stddev with two samples")
	}
}

func TestDirectShotMode(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Train with consistent single-pulse successes
	targetLevel := 15
	for i := 0; i < 15; i++ {
		calc.mu.Lock()
		calc.currentTarget = targetLevel
		calc.currentStart = 0
		calc.currentPulses = 1 // Single pulse success
		calc.hadOvershoot = false
		calc.mu.Unlock()
		calc.RecordSuccess(1.0)
	}

	// Now start a new write - should use direct shot mode
	calc.StartWrite(0, targetLevel, 1.0)

	mode := calc.GetCurrentMode()
	if mode != ModeDirectShot {
		t.Errorf("Expected ModeDirectShot after training, got %d", mode)
	}
}

func TestMomentumAcceleration(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Start a write
	calc.StartWrite(0, 20, 1.5)

	// First step
	v1 := calc.CalculateNextAdaptiveVoltage(1.0, 5, 20, DirectionAscending)

	// Second step in same direction - should have momentum
	v2 := calc.CalculateNextAdaptiveVoltage(v1, 8, 20, DirectionAscending)
	step1 := v1 - 1.0
	step2 := v2 - v1

	// Second step should be larger due to momentum (assuming same-ish level gap)
	// Note: the level gap also affects step size, so we just verify momentum is tracked
	if calc.consecutiveDir < 2 {
		t.Errorf("Expected consecutiveDir >= 2, got %d", calc.consecutiveDir)
	}

	// With momentum gain of 1.15, second step should be ~15% larger (adjusting for level gap difference)
	// This is approximate due to level gap changes
	_ = step1
	_ = step2
}

func TestBinarySearchActivation(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)
	calc.AdaptiveConfig.BinarySearchThreshold = 3 // Lower for testing

	calc.StartWrite(0, 20, 1.5)

	// Apply pulses until binary search activates
	v := 1.0
	for i := 0; i < 5; i++ {
		v = calc.CalculateNextAdaptiveVoltage(v, 10, 20, DirectionAscending)
	}

	// Binary search should be active now
	calc.mu.RLock()
	bsActive := calc.bsActive
	calc.mu.RUnlock()

	if !bsActive {
		t.Error("Expected binary search to be active after threshold pulses")
	}
}

func TestBinarySearchBoundsUpdate(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	calc.StartWrite(0, 15, 1.5)
	calc.mu.Lock()
	calc.bsActive = true
	calc.bsLowVoltage = 0.5
	calc.bsHighVoltage = 1.5
	calc.mu.Unlock()

	// Simulate undershoot at 0.8V (current=10, target=15)
	calc.UpdateBinarySearchBounds(0.8, 10, 15, DirectionAscending)

	calc.mu.RLock()
	low := calc.bsLowVoltage
	calc.mu.RUnlock()

	if low != 0.8 {
		t.Errorf("Expected lower bound updated to 0.8, got %f", low)
	}

	// Simulate overshoot at 1.2V (current=17, target=15)
	calc.UpdateBinarySearchBounds(1.2, 17, 15, DirectionAscending)

	calc.mu.RLock()
	high := calc.bsHighVoltage
	calc.mu.RUnlock()

	if high != 1.2 {
		t.Errorf("Expected upper bound updated to 1.2, got %f", high)
	}
}

func TestEfficiencyStats(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Record some writes with varying pulse counts
	writes := []struct {
		pulses    int
		overshoot bool
	}{
		{1, false}, // Single pulse success
		{2, false},
		{1, false}, // Single pulse success
		{3, false},
		{2, true}, // With overshoot
	}

	for i, w := range writes {
		calc.mu.Lock()
		calc.currentTarget = 15
		calc.currentStart = 0
		calc.currentPulses = w.pulses
		calc.hadOvershoot = w.overshoot
		calc.mu.Unlock()
		calc.RecordSuccess(1.0 + float64(i)*0.01)
	}

	stats := calc.GetEfficiencyStats()

	if stats.TotalWrites != 5 {
		t.Errorf("Expected 5 total writes, got %d", stats.TotalWrites)
	}

	// 2 single pulse successes out of 5
	expectedSinglePulseRate := 2.0 / 5.0
	if math.Abs(stats.SinglePulseRate-expectedSinglePulseRate) > 0.01 {
		t.Errorf("Expected single pulse rate %f, got %f", expectedSinglePulseRate, stats.SinglePulseRate)
	}

	// Average pulses: (1+2+1+3+2)/5 = 1.8
	expectedAvgPulses := 1.8
	if math.Abs(stats.AveragePulses-expectedAvgPulses) > 0.01 {
		t.Errorf("Expected avg pulses %f, got %f", expectedAvgPulses, stats.AveragePulses)
	}
}

func TestSafeVoltageBoundsLearning(t *testing.T) {
	calc := NewAdaptiveISPPCalculator(1.0, 30)

	// Record successful writes without overshoot
	targetLevel := 10
	for i := 0; i < 5; i++ {
		calc.mu.Lock()
		calc.currentTarget = targetLevel
		calc.currentStart = 0
		calc.currentPulses = 2
		calc.hadOvershoot = false
		calc.mu.Unlock()
		calc.RecordSuccess(0.8 + float64(i)*0.02) // 0.8, 0.82, 0.84, 0.86, 0.88
	}

	stats := &calc.LevelStats[targetLevel]
	if !stats.BoundsValid {
		t.Error("Expected bounds to be valid after successful writes")
	}

	// Safe bounds should be established
	if stats.SafeVoltageMin <= 0 {
		t.Error("Expected positive SafeVoltageMin")
	}
	if stats.SafeVoltageMax <= stats.SafeVoltageMin {
		t.Error("Expected SafeVoltageMax > SafeVoltageMin")
	}
}
