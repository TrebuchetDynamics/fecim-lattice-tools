// Package physics provides shared physics utilities for FeCIM simulations.
// This file implements adaptive, self-learning ISPP (Incremental Step Pulse Programming).
//
// Improvements over basic ISPP:
//   1. Adaptive Step Sizing: Large steps when far, fine steps when close
//   2. Bayesian Voltage Prediction: Learn optimal start voltage per level
//   3. Online Calibration Refinement: Update calibration from runtime successes
//   4. Transition-Aware Programming: Track level-to-level transition statistics
//   5. Direct Shot Mode: Skip iteration when high confidence (single pulse)
//   6. Binary Search Fallback: Use binary search when linear stepping stalls
//   7. Error-Proportional Stepping: Step size proportional to level error
//   8. Momentum Acceleration: Consecutive same-direction steps increase in size
//
// The algorithm continuously improves as it observes more write operations,
// reducing average pulses-to-converge and overshoot rate over time.
package physics

import (
	"math"
	"sync"
)

// ISPPMode represents the current operating mode of the adaptive ISPP.
type ISPPMode int

const (
	// ModeLinear uses standard incremental stepping
	ModeLinear ISPPMode = iota
	// ModeDirectShot attempts single-pulse write when confidence is high
	ModeDirectShot
	// ModeBinarySearch uses binary search between known bounds
	ModeBinarySearch
	// ModeErrorProportional scales steps by level error magnitude
	ModeErrorProportional
)

// LevelStatistics tracks per-level voltage and convergence statistics.
type LevelStatistics struct {
	// Voltage tracking (Welford's online algorithm for mean/variance)
	Count        int     // Number of successful writes to this level
	VoltageMean  float64 // Running mean of successful voltages
	VoltageM2    float64 // Sum of squared differences (for variance)
	VoltageMin   float64 // Minimum successful voltage observed
	VoltageMax   float64 // Maximum successful voltage observed

	// Convergence tracking
	TotalPulses      int // Total pulses across all writes
	OvershootCount   int // Number of times we overshot this level
	FailCount        int // Number of failed attempts
	SinglePulseCount int // Number of times we succeeded with 1 pulse (direct shot)

	// Adaptive step multiplier (learned)
	// >1.0 means this level converges slowly, needs bigger steps
	// <1.0 means this level is sensitive, needs smaller steps
	StepMultiplier float64

	// Binary search bounds (learned safe operating range)
	SafeVoltageMin float64 // Minimum voltage that reached this level without undershoot
	SafeVoltageMax float64 // Maximum voltage that reached this level without overshoot
	BoundsValid    bool    // True if bounds have been established
}

// VoltageStdDev returns the standard deviation of successful voltages.
func (ls *LevelStatistics) VoltageStdDev() float64 {
	if ls.Count < 2 {
		return 0
	}
	return math.Sqrt(ls.VoltageM2 / float64(ls.Count-1))
}

// UpdateVoltage updates voltage statistics using Welford's online algorithm.
func (ls *LevelStatistics) UpdateVoltage(voltage float64) {
	ls.Count++
	delta := voltage - ls.VoltageMean
	ls.VoltageMean += delta / float64(ls.Count)
	delta2 := voltage - ls.VoltageMean
	ls.VoltageM2 += delta * delta2

	// Track min/max
	if ls.Count == 1 || voltage < ls.VoltageMin {
		ls.VoltageMin = voltage
	}
	if ls.Count == 1 || voltage > ls.VoltageMax {
		ls.VoltageMax = voltage
	}
}

// TransitionKey represents a level-to-level transition.
type TransitionKey struct {
	From int
	To   int
}

// TransitionStats tracks statistics for a specific state transition.
type TransitionStats struct {
	Count          int     // Number of times this transition was performed
	AvgPulses      float64 // Average pulses needed
	AvgVoltage     float64 // Average final voltage used
	OvershootRate  float64 // Fraction that overshot
}

// AdaptiveISPPConfig extends ISPPConfig with learning parameters.
type AdaptiveISPPConfig struct {
	ISPPConfig // Embed base config

	// Adaptive step sizing
	MinStepPercent   float64 // Minimum step (fine control near target), default 0.5%
	MaxStepPercent   float64 // Maximum step (far from target), default 10%
	StepDecayFactor  float64 // How fast steps shrink as we approach target, default 0.7

	// Bayesian prediction
	ConfidenceK        float64 // Std dev multiplier for conservative start, default 1.5
	MinSamplesForBayes int     // Minimum samples before using Bayesian prediction, default 5
	LearningRate       float64 // EMA weight for calibration updates, default 0.1

	// Overshoot adaptation
	OvershootPenalty float64 // Reduce step multiplier after overshoot, default 0.8
	SuccessBonus     float64 // Increase step multiplier after quick success, default 1.05

	// Direct shot mode (single pulse when confident)
	DirectShotConfidence float64 // Minimum confidence for direct shot attempt, default 0.85
	DirectShotMinSamples int     // Minimum samples for direct shot, default 10

	// Error-proportional stepping
	ErrorStepGain float64 // Multiplier: step = base * (1 + gain * |error|), default 0.5

	// Momentum (acceleration on consecutive same-direction steps)
	MomentumGain    float64 // Step increase per consecutive step, default 1.15
	MomentumMaxGain float64 // Maximum momentum multiplier, default 2.0

	// Binary search parameters
	BinarySearchThreshold int // Switch to binary search after this many pulses, default 5
}

// DefaultAdaptiveISPPConfig returns sensible defaults for adaptive ISPP.
func DefaultAdaptiveISPPConfig() AdaptiveISPPConfig {
	return AdaptiveISPPConfig{
		ISPPConfig: ISPPConfig{
			StartRatio:  0.7,  // Fallback when no statistics available
			StepPercent: 0.02, // Base step (overridden by adaptive logic)
			MaxPulses:   20,
			SafetyCap:   2.2,
			Tolerance:   0,
		},
		MinStepPercent:        0.005, // 0.5% of Ec minimum
		MaxStepPercent:        0.10,  // 10% of Ec maximum
		StepDecayFactor:       0.7,   // Steps shrink by 30% as we get closer
		ConfidenceK:           1.5,   // Start 1.5 sigma below mean
		MinSamplesForBayes:    5,     // Need 5+ samples for Bayesian
		LearningRate:          0.1,   // 10% weight to new observations
		OvershootPenalty:      0.8,   // 20% step reduction after overshoot
		SuccessBonus:          1.05,  // 5% step increase after quick convergence
		DirectShotConfidence:  0.85,  // 85% confidence for direct shot
		DirectShotMinSamples:  10,    // Need 10+ samples for direct shot
		ErrorStepGain:         0.5,   // 50% bonus step per level error
		MomentumGain:          1.15,  // 15% step increase per consecutive step
		MomentumMaxGain:       2.0,   // Cap momentum at 2x
		BinarySearchThreshold: 5,     // Switch to binary search after 5 pulses
	}
}

// AdaptiveISPPCalculator wraps ISPPCalculator with learning capabilities.
type AdaptiveISPPCalculator struct {
	mu sync.RWMutex

	// Base calculator
	*ISPPCalculator

	// Adaptive configuration
	AdaptiveConfig AdaptiveISPPConfig

	// Per-level learning (indexed by target level)
	LevelStats []LevelStatistics

	// Transition learning (from-level -> to-level)
	TransitionStats map[TransitionKey]*TransitionStats

	// Current write state
	currentTarget   int
	currentStart    int
	currentPulses   int
	currentVoltages []float64 // Voltages applied this write
	hadOvershoot    bool
	currentMode     ISPPMode // Active algorithm mode
	consecutiveDir  int      // Consecutive steps in same direction (for momentum)
	lastDirection   HysteresisDirection
	lastLevelError  int // Last observed level error (for binary search)

	// Binary search state
	bsLowVoltage  float64 // Known undershoot voltage
	bsHighVoltage float64 // Known overshoot voltage
	bsActive      bool    // Binary search is active
}

// NewAdaptiveISPPCalculator creates an adaptive calculator.
func NewAdaptiveISPPCalculator(ec float64, numLevels int) *AdaptiveISPPCalculator {
	config := DefaultAdaptiveISPPConfig()
	base := NewISPPCalculatorWithConfig(ec, numLevels, config.ISPPConfig)

	calc := &AdaptiveISPPCalculator{
		ISPPCalculator:  base,
		AdaptiveConfig:  config,
		LevelStats:      make([]LevelStatistics, numLevels),
		TransitionStats: make(map[TransitionKey]*TransitionStats),
	}

	// Initialize step multipliers to 1.0
	for i := range calc.LevelStats {
		calc.LevelStats[i].StepMultiplier = 1.0
	}

	return calc
}

// NewAdaptiveISPPCalculatorWithConfig creates an adaptive calculator with custom config.
func NewAdaptiveISPPCalculatorWithConfig(ec float64, numLevels int, config AdaptiveISPPConfig) *AdaptiveISPPCalculator {
	base := NewISPPCalculatorWithConfig(ec, numLevels, config.ISPPConfig)

	calc := &AdaptiveISPPCalculator{
		ISPPCalculator:  base,
		AdaptiveConfig:  config,
		LevelStats:      make([]LevelStatistics, numLevels),
		TransitionStats: make(map[TransitionKey]*TransitionStats),
	}

	for i := range calc.LevelStats {
		calc.LevelStats[i].StepMultiplier = 1.0
	}

	return calc
}

// StartWrite begins a new write operation, returning the predicted start voltage.
// This uses Bayesian prediction if sufficient statistics are available.
// It also selects the optimal mode (DirectShot, Linear, etc.) based on confidence.
func (c *AdaptiveISPPCalculator) StartWrite(currentLevel, targetLevel int, calibratedVoltage float64) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset write state
	c.currentStart = currentLevel
	c.currentTarget = targetLevel
	c.currentPulses = 0
	c.currentVoltages = nil
	c.hadOvershoot = false
	c.consecutiveDir = 0
	c.lastDirection = DirectionUnknown
	c.lastLevelError = 0
	c.bsActive = false
	c.currentMode = ModeLinear // Default

	if targetLevel < 0 || targetLevel >= len(c.LevelStats) {
		return c.CalculateStartVoltage(calibratedVoltage)
	}

	stats := &c.LevelStats[targetLevel]
	direction := GetDirection(currentLevel, targetLevel)

	// Check if we can use Direct Shot mode (single pulse, high confidence)
	if c.canUseDirectShot(stats, direction) {
		c.currentMode = ModeDirectShot
		directVoltage := stats.VoltageMean
		c.currentVoltages = append(c.currentVoltages, directVoltage)
		return directVoltage
	}

	// Use Bayesian prediction if we have enough samples
	if stats.Count >= c.AdaptiveConfig.MinSamplesForBayes {
		stdDev := stats.VoltageStdDev()
		if stdDev > 0 {
			// Conservative start: mean - k*sigma
			// This reduces overshoot risk while still being informed
			k := c.AdaptiveConfig.ConfidenceK

			var startVoltage float64
			if direction == DirectionAscending {
				// Ascending: start below predicted mean
				startVoltage = stats.VoltageMean - k*stdDev
				// Don't go below reasonable minimum
				if startVoltage < calibratedVoltage*0.5 {
					startVoltage = calibratedVoltage * 0.5
				}
			} else {
				// Descending: start above predicted mean (less negative)
				startVoltage = stats.VoltageMean + k*stdDev
				// Don't go above reasonable maximum (closer to 0)
				if startVoltage > calibratedVoltage*0.5 {
					startVoltage = calibratedVoltage * 0.5
				}
			}

			c.currentVoltages = append(c.currentVoltages, startVoltage)
			return startVoltage
		}
	}

	// Fallback to default conservative start
	startV := c.CalculateStartVoltage(calibratedVoltage)
	c.currentVoltages = append(c.currentVoltages, startV)
	return startV
}

// canUseDirectShot checks if conditions are met for single-pulse direct write.
func (c *AdaptiveISPPCalculator) canUseDirectShot(stats *LevelStatistics, direction HysteresisDirection) bool {
	// Need enough samples
	if stats.Count < c.AdaptiveConfig.DirectShotMinSamples {
		return false
	}

	// Need high success rate with single pulse
	singlePulseRate := float64(stats.SinglePulseCount) / float64(stats.Count)
	if singlePulseRate < 0.5 { // At least 50% single-pulse success historically
		return false
	}

	// Need low variance (tight voltage distribution)
	stdDev := stats.VoltageStdDev()
	if stdDev == 0 {
		return true // Perfect consistency
	}

	// Coefficient of variation should be low
	cv := stdDev / math.Abs(stats.VoltageMean)
	if cv > 0.05 { // More than 5% variation is too risky
		return false
	}

	// Need low overshoot rate
	totalAttempts := stats.Count + stats.FailCount
	if totalAttempts > 0 {
		overshootRate := float64(stats.OvershootCount) / float64(totalAttempts)
		if overshootRate > 0.1 { // More than 10% overshoot is too risky
			return false
		}
	}

	return true
}

// GetCurrentMode returns the active ISPP mode.
func (c *AdaptiveISPPCalculator) GetCurrentMode() ISPPMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMode
}

// CalculateAdaptiveStep returns the voltage step based on distance from target.
// Uses larger steps when far, smaller steps when close.
// Incorporates error-proportional scaling and momentum.
func (c *AdaptiveISPPCalculator) CalculateAdaptiveStep(currentLevel, targetLevel int, direction HysteresisDirection) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.calculateAdaptiveStepInternal(currentLevel, targetLevel, direction)
}

// calculateAdaptiveStepInternal is the lock-free internal implementation.
func (c *AdaptiveISPPCalculator) calculateAdaptiveStepInternal(currentLevel, targetLevel int, direction HysteresisDirection) float64 {
	// Base step calculation
	levelGap := absInt(targetLevel - currentLevel)
	maxGap := c.NumLevels - 1
	if maxGap < 1 {
		maxGap = 1
	}

	// Distance ratio: 1.0 when far, 0.0 when at target
	distanceRatio := float64(levelGap) / float64(maxGap)

	// Interpolate between min and max step based on distance
	minStep := c.AdaptiveConfig.MinStepPercent * c.Ec
	maxStep := c.AdaptiveConfig.MaxStepPercent * c.Ec

	// Exponential decay as we approach target
	// step = minStep + (maxStep - minStep) * distanceRatio^decay
	decayedRatio := math.Pow(distanceRatio, c.AdaptiveConfig.StepDecayFactor)
	step := minStep + (maxStep-minStep)*decayedRatio

	// Error-proportional scaling: bigger steps for bigger errors
	// This makes convergence faster when far from target
	errorGain := 1.0 + c.AdaptiveConfig.ErrorStepGain*float64(levelGap)
	step *= errorGain

	// Apply momentum: consecutive steps in same direction get progressively larger
	if c.consecutiveDir > 0 && direction == c.lastDirection {
		momentum := math.Pow(c.AdaptiveConfig.MomentumGain, float64(c.consecutiveDir))
		if momentum > c.AdaptiveConfig.MomentumMaxGain {
			momentum = c.AdaptiveConfig.MomentumMaxGain
		}
		step *= momentum
	}

	// Apply per-level learned multiplier
	if targetLevel >= 0 && targetLevel < len(c.LevelStats) {
		step *= c.LevelStats[targetLevel].StepMultiplier
	}

	// Clamp to bounds
	if step < minStep {
		step = minStep
	}
	if step > maxStep {
		step = maxStep
	}

	return step
}

// CalculateNextAdaptiveVoltage computes the next pulse voltage using adaptive stepping.
// Automatically switches to binary search mode when appropriate.
func (c *AdaptiveISPPCalculator) CalculateNextAdaptiveVoltage(currentVoltage float64, currentLevel, targetLevel int, direction HysteresisDirection) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Track momentum
	if direction == c.lastDirection {
		c.consecutiveDir++
	} else {
		c.consecutiveDir = 1
	}
	c.lastDirection = direction

	// Check if we should switch to binary search
	if c.currentPulses >= c.AdaptiveConfig.BinarySearchThreshold && !c.bsActive {
		c.activateBinarySearch(currentVoltage, currentLevel, targetLevel, direction)
	}

	var nextVoltage float64

	if c.bsActive && c.bsLowVoltage != 0 && c.bsHighVoltage != 0 {
		// Binary search mode: use midpoint
		nextVoltage = (c.bsLowVoltage + c.bsHighVoltage) / 2
		c.currentMode = ModeBinarySearch
	} else {
		// Standard adaptive stepping
		step := c.calculateAdaptiveStepInternal(currentLevel, targetLevel, direction)

		switch direction {
		case DirectionAscending:
			nextVoltage = currentVoltage + step
		case DirectionDescending:
			nextVoltage = currentVoltage - step
		default:
			nextVoltage = currentVoltage
		}
		c.currentMode = ModeErrorProportional
	}

	c.currentPulses++
	c.currentVoltages = append(c.currentVoltages, nextVoltage)

	return c.ClampVoltage(nextVoltage, direction)
}

// activateBinarySearch initializes binary search based on current state.
func (c *AdaptiveISPPCalculator) activateBinarySearch(currentVoltage float64, currentLevel, targetLevel int, direction HysteresisDirection) {
	c.bsActive = true

	// Use observed voltages to establish bounds
	levelError := currentLevel - targetLevel

	if direction == DirectionAscending {
		if levelError < 0 {
			// Undershot: current voltage is a lower bound
			c.bsLowVoltage = currentVoltage
			// Upper bound from safety cap or learned max
			if c.currentTarget >= 0 && c.currentTarget < len(c.LevelStats) {
				stats := &c.LevelStats[c.currentTarget]
				if stats.BoundsValid && stats.SafeVoltageMax > c.bsLowVoltage {
					c.bsHighVoltage = stats.SafeVoltageMax
				} else {
					c.bsHighVoltage = c.Ec * c.Config.SafetyCap
				}
			} else {
				c.bsHighVoltage = c.Ec * c.Config.SafetyCap
			}
		} else if levelError > 0 {
			// Overshot: current voltage is upper bound
			c.bsHighVoltage = currentVoltage
			c.bsLowVoltage = 0
		}
	} else { // Descending
		if levelError > 0 {
			// Undershot (didn't go down enough): current voltage is upper bound
			c.bsHighVoltage = currentVoltage
			if c.currentTarget >= 0 && c.currentTarget < len(c.LevelStats) {
				stats := &c.LevelStats[c.currentTarget]
				if stats.BoundsValid && stats.SafeVoltageMin < c.bsHighVoltage {
					c.bsLowVoltage = stats.SafeVoltageMin
				} else {
					c.bsLowVoltage = -c.Ec * c.Config.SafetyCap
				}
			} else {
				c.bsLowVoltage = -c.Ec * c.Config.SafetyCap
			}
		} else if levelError < 0 {
			// Overshot: current voltage is lower bound
			c.bsLowVoltage = currentVoltage
			c.bsHighVoltage = 0
		}
	}
}

// UpdateBinarySearchBounds should be called after each verify to refine bounds.
func (c *AdaptiveISPPCalculator) UpdateBinarySearchBounds(currentVoltage float64, currentLevel, targetLevel int, direction HysteresisDirection) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.bsActive {
		return
	}

	levelError := currentLevel - targetLevel
	c.lastLevelError = levelError

	if direction == DirectionAscending {
		if levelError < 0 {
			// Undershot: raise lower bound
			if currentVoltage > c.bsLowVoltage {
				c.bsLowVoltage = currentVoltage
			}
		} else if levelError > 0 {
			// Overshot: lower upper bound
			if currentVoltage < c.bsHighVoltage {
				c.bsHighVoltage = currentVoltage
			}
		}
	} else { // Descending
		if levelError > 0 {
			// Undershot: lower upper bound
			if currentVoltage < c.bsHighVoltage {
				c.bsHighVoltage = currentVoltage
			}
		} else if levelError < 0 {
			// Overshot: raise lower bound
			if currentVoltage > c.bsLowVoltage {
				c.bsLowVoltage = currentVoltage
			}
		}
	}
}

// RecordOvershoot records that an overshoot occurred.
func (c *AdaptiveISPPCalculator) RecordOvershoot() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.hadOvershoot = true

	if c.currentTarget >= 0 && c.currentTarget < len(c.LevelStats) {
		c.LevelStats[c.currentTarget].OvershootCount++

		// Penalize step multiplier - this level needs finer control
		c.LevelStats[c.currentTarget].StepMultiplier *= c.AdaptiveConfig.OvershootPenalty
		if c.LevelStats[c.currentTarget].StepMultiplier < 0.5 {
			c.LevelStats[c.currentTarget].StepMultiplier = 0.5 // Don't go below 50%
		}
	}
}

// RecordSuccess records a successful write, updating statistics.
func (c *AdaptiveISPPCalculator) RecordSuccess(finalVoltage float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentTarget < 0 || c.currentTarget >= len(c.LevelStats) {
		return
	}

	stats := &c.LevelStats[c.currentTarget]

	// Update voltage statistics
	stats.UpdateVoltage(finalVoltage)
	stats.TotalPulses += c.currentPulses

	// Track single-pulse successes for direct shot eligibility
	if c.currentPulses == 1 {
		stats.SinglePulseCount++
	}

	// Update safe voltage bounds (for binary search optimization)
	if !c.hadOvershoot {
		direction := GetDirection(c.currentStart, c.currentTarget)
		if direction == DirectionAscending {
			// This voltage worked without overshoot
			if !stats.BoundsValid || finalVoltage > stats.SafeVoltageMin {
				stats.SafeVoltageMin = finalVoltage * 0.95 // Slightly below actual
			}
			if !stats.BoundsValid || finalVoltage < stats.SafeVoltageMax {
				stats.SafeVoltageMax = finalVoltage * 1.05 // Slightly above actual
			}
		} else {
			// Descending: voltages are negative
			if !stats.BoundsValid || finalVoltage < stats.SafeVoltageMin {
				stats.SafeVoltageMin = finalVoltage * 1.05 // More negative
			}
			if !stats.BoundsValid || finalVoltage > stats.SafeVoltageMax {
				stats.SafeVoltageMax = finalVoltage * 0.95 // Less negative
			}
		}
		stats.BoundsValid = true
	}

	// Reward quick convergence with step multiplier bonus
	// "Quick" = 3 or fewer pulses without overshoot
	if c.currentPulses <= 3 && !c.hadOvershoot {
		stats.StepMultiplier *= c.AdaptiveConfig.SuccessBonus
		if stats.StepMultiplier > 2.0 {
			stats.StepMultiplier = 2.0 // Cap at 200%
		}
	}

	// Update transition statistics
	key := TransitionKey{From: c.currentStart, To: c.currentTarget}
	ts := c.TransitionStats[key]
	if ts == nil {
		ts = &TransitionStats{}
		c.TransitionStats[key] = ts
	}

	// Update with exponential moving average
	if ts.Count == 0 {
		ts.AvgPulses = float64(c.currentPulses)
		ts.AvgVoltage = finalVoltage
		ts.OvershootRate = boolToFloat(c.hadOvershoot)
	} else {
		lr := c.AdaptiveConfig.LearningRate
		ts.AvgPulses = ts.AvgPulses*(1-lr) + float64(c.currentPulses)*lr
		ts.AvgVoltage = ts.AvgVoltage*(1-lr) + finalVoltage*lr
		ts.OvershootRate = ts.OvershootRate*(1-lr) + boolToFloat(c.hadOvershoot)*lr
	}
	ts.Count++
}

// GetEfficiencyStats returns efficiency metrics for the adaptive ISPP.
func (c *AdaptiveISPPCalculator) GetEfficiencyStats() EfficiencyStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var stats EfficiencyStats

	totalWrites := 0
	totalPulses := 0
	totalSinglePulse := 0
	totalOvershoots := 0

	for _, ls := range c.LevelStats {
		totalWrites += ls.Count
		totalPulses += ls.TotalPulses
		totalSinglePulse += ls.SinglePulseCount
		totalOvershoots += ls.OvershootCount
	}

	if totalWrites > 0 {
		stats.AveragePulses = float64(totalPulses) / float64(totalWrites)
		stats.SinglePulseRate = float64(totalSinglePulse) / float64(totalWrites)
		stats.OvershootRate = float64(totalOvershoots) / float64(totalWrites)
	}

	stats.TotalWrites = totalWrites
	stats.DirectShotAttempts = totalSinglePulse

	return stats
}

// EfficiencyStats holds efficiency metrics.
type EfficiencyStats struct {
	TotalWrites       int
	AveragePulses     float64
	SinglePulseRate   float64 // Fraction of writes that succeeded in 1 pulse
	OvershootRate     float64
	DirectShotAttempts int
}

// RecordFailure records a failed write attempt.
func (c *AdaptiveISPPCalculator) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentTarget >= 0 && c.currentTarget < len(c.LevelStats) {
		c.LevelStats[c.currentTarget].FailCount++
	}
}

// GetLevelConfidence returns a confidence score (0-1) for a target level.
// Higher confidence = more reliable predictions.
func (c *AdaptiveISPPCalculator) GetLevelConfidence(targetLevel int) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if targetLevel < 0 || targetLevel >= len(c.LevelStats) {
		return 0
	}

	stats := &c.LevelStats[targetLevel]
	if stats.Count == 0 {
		return 0
	}

	// Confidence based on:
	// 1. Sample count (more samples = more confident)
	// 2. Success rate (fewer failures = more confident)
	// 3. Low overshoot rate
	// 4. Low voltage variance

	sampleConfidence := 1.0 - math.Exp(-float64(stats.Count)/10.0) // Asymptote to 1.0

	totalAttempts := stats.Count + stats.FailCount
	successRate := float64(stats.Count) / float64(totalAttempts)

	overshootRate := 0.0
	if totalAttempts > 0 {
		overshootRate = float64(stats.OvershootCount) / float64(totalAttempts)
	}
	overshootConfidence := 1.0 - overshootRate

	// Combine factors
	confidence := sampleConfidence * successRate * overshootConfidence
	return confidence
}

// GetPredictedVoltage returns the Bayesian predicted voltage for a level.
// Returns (voltage, hasData) where hasData is false if insufficient statistics.
func (c *AdaptiveISPPCalculator) GetPredictedVoltage(targetLevel int) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if targetLevel < 0 || targetLevel >= len(c.LevelStats) {
		return 0, false
	}

	stats := &c.LevelStats[targetLevel]
	if stats.Count < c.AdaptiveConfig.MinSamplesForBayes {
		return 0, false
	}

	return stats.VoltageMean, true
}

// GetTransitionPrediction returns predicted performance for a transition.
// Returns (avgPulses, avgVoltage, confidence) or zeros if no data.
func (c *AdaptiveISPPCalculator) GetTransitionPrediction(fromLevel, toLevel int) (float64, float64, float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := TransitionKey{From: fromLevel, To: toLevel}
	ts := c.TransitionStats[key]
	if ts == nil || ts.Count < 3 {
		return 0, 0, 0
	}

	// Confidence based on sample count
	confidence := 1.0 - math.Exp(-float64(ts.Count)/5.0)
	return ts.AvgPulses, ts.AvgVoltage, confidence
}

// ExportLearningState exports the learned statistics for persistence.
func (c *AdaptiveISPPCalculator) ExportLearningState() AdaptiveLearningState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state := AdaptiveLearningState{
		LevelStats:      make([]LevelStatistics, len(c.LevelStats)),
		TransitionStats: make(map[TransitionKey]*TransitionStats),
	}

	copy(state.LevelStats, c.LevelStats)
	for k, v := range c.TransitionStats {
		copyV := *v
		state.TransitionStats[k] = &copyV
	}

	return state
}

// ImportLearningState imports previously saved learning state.
func (c *AdaptiveISPPCalculator) ImportLearningState(state AdaptiveLearningState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(state.LevelStats) == len(c.LevelStats) {
		copy(c.LevelStats, state.LevelStats)
	}

	for k, v := range state.TransitionStats {
		copyV := *v
		c.TransitionStats[k] = &copyV
	}
}

// AdaptiveLearningState holds exportable learning data.
type AdaptiveLearningState struct {
	LevelStats      []LevelStatistics
	TransitionStats map[TransitionKey]*TransitionStats
}

// GetStepMultipliers returns all per-level step multipliers.
func (c *AdaptiveISPPCalculator) GetStepMultipliers() []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	multipliers := make([]float64, len(c.LevelStats))
	for i, stats := range c.LevelStats {
		multipliers[i] = stats.StepMultiplier
	}
	return multipliers
}

// GetLevelSuccessRates returns success rate per level.
func (c *AdaptiveISPPCalculator) GetLevelSuccessRates() []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rates := make([]float64, len(c.LevelStats))
	for i, stats := range c.LevelStats {
		total := stats.Count + stats.FailCount
		if total > 0 {
			rates[i] = float64(stats.Count) / float64(total)
		} else {
			rates[i] = 1.0 // No data = assume perfect
		}
	}
	return rates
}

// GetAveragePulsesPerLevel returns average pulses to converge per level.
func (c *AdaptiveISPPCalculator) GetAveragePulsesPerLevel() []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	avgPulses := make([]float64, len(c.LevelStats))
	for i, stats := range c.LevelStats {
		if stats.Count > 0 {
			avgPulses[i] = float64(stats.TotalPulses) / float64(stats.Count)
		}
	}
	return avgPulses
}

// Utility functions

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
