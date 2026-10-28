package gui

import (
	"math"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
)

// Calibration EMA (Exponential Moving Average) tuning constants.
const (
	// relaxAdjustRate is the per-retry relaxation compensation adjustment (2% of Ec).
	// Controls how quickly the calibration adapts to systematic read-back bias.
	relaxAdjustRate = 0.02

	// emaRetainFrac is the retain fraction for the EMA filter (70% old value).
	// Higher values → smoother but slower convergence; lower → faster but noisier.
	emaRetainFrac = 0.7

	// emaUpdateFrac is the update fraction (1 - emaRetainFrac = 30% new value).
	emaUpdateFrac = 1.0 - emaRetainFrac

	// relaxCompMin/Max are the clamp bounds for relaxation compensation.
	// Prevents runaway drift if calibration oscillates.
	relaxCompMin = -0.05
	relaxCompMax = 0.25
)

// calibrateLevelsAtTemperature performs calibration at a specific temperature.
// Sets Preisach temperature before calibrating, stores result in cache.
// MUST be called with a.mu held.
func (a *App) calibrateLevelsAtTemperature(tempK float64) {
	if a.material == nil {
		return
	}
	// Allow fast cancellation when module view/window switches and Stop() is called.
	if !a.running.Load() {
		return
	}

	// Set engine temperature before calibrating
	if a.useLKSolver() {
		if a.lkSolver != nil {
			a.lkSolver.Temperature = tempK
		}
	} else if a.preisach != nil {
		a.preisach.SetTemperature(tempK)
	}
	a.calibrationTemp = tempK

	// Perform calibration
	a.calibrateLevels()

	if !a.running.Load() {
		return
	}

	// Store in cache
	tempKRounded := int(math.Round(tempK))
	if a.tempCalibrations == nil {
		a.tempCalibrations = make(map[int]*TempCalibration)
	}
	a.tempCalibrations[tempKRounded] = &TempCalibration{
		Temperature:     tempK,
		CalibrationUp:   append([]float64(nil), a.calibrationUp...),
		CalibrationDown: append([]float64(nil), a.calibrationDown...),
		CalibUpLow:      append([]float64(nil), a.calibUpLow...),
		CalibUpHigh:     append([]float64(nil), a.calibUpHigh...),
		CalibDownLow:    append([]float64(nil), a.calibDownLow...),
		CalibDownHigh:   append([]float64(nil), a.calibDownHigh...),
		LastErrorUp:     append([]int(nil), a.lastErrorUp...),
		LastErrorDown:   append([]int(nil), a.lastErrorDown...),
		RelaxCompUp:     append([]float64(nil), a.relaxCompUp...),
		RelaxCompDown:   append([]float64(nil), a.relaxCompDown...),
	}

	log.Printf("Level calibration complete for material: %s at %.0fK", a.material.Name, tempK)
}

// calibrateLevels performs a calibration sweep to map field→level relationship.
// This mimics how real ferroelectric memory controllers characterize each device
// and build lookup tables for programming. Called at startup and when material changes.
// MUST be called with a.mu held.
//
// FIXED: Uses binary search with fresh saturation for each level to avoid
// Preisach state corruption from the previous sweep-based approach.
func (a *App) calibrateLevels() {
	if a.material == nil || !a.running.Load() {
		return
	}
	if a.useLKSolver() {
		a.calibrateLevelsLK()
		return
	}
	if a.preisach == nil {
		return
	}
	a.calibrationTemp = a.preisach.Temperature
	a.runLevelCalibration(false, a.calibrationTemp)
}

// calibrateLevelsLK retains native single-domain controller calibration.
// MUST be called with a.mu held.
func (a *App) calibrateLevelsLK() {
	if a.material == nil || !a.running.Load() {
		return
	}
	a.runLevelCalibration(true, a.currentTemperature())
}

// runLevelCalibration drains the shared resumable sweep for the legacy GUI.
// Browser/workbench callers advance it explicitly, without the GUI timer/save.
func (a *App) runLevelCalibration(landau bool, temperature float64) {
	job, err := controller.NewLevelCalibration(*a.material, landau, a.numLevels, a.wrdRangeFrac, temperature, 1e9)
	if err != nil {
		log.Printf("Level calibration unavailable: %v", err)
		return
	}
	for !job.Done() {
		if !a.running.Load() {
			return
		}
		job.Step()
	}
	cm, _ := job.Result()
	a.calibrationUp, a.calibrationDown = cm.CalibrationUp, cm.CalibrationDown
	a.calibUpLow, a.calibUpHigh = cm.CalibUpLow, cm.CalibUpHigh
	a.calibDownLow, a.calibDownHigh = cm.CalibDownLow, cm.CalibDownHigh
	a.lastErrorUp, a.lastErrorDown = cm.LastErrorUp, cm.LastErrorDown
	a.relaxCompUp, a.relaxCompDown = cm.RelaxCompUp, cm.RelaxCompDown
	a.finalizeCalibration(job.EffectiveEc())
}

// finalizeCalibration enforces monotonicity, resets state, and syncs to CalibrationManager.
// MUST be called with a.mu held.
func (a *App) finalizeCalibration(Ec float64) {
	numLevels := a.numLevels
	if numLevels <= 0 {
		numLevels = 1
	}

	cm := &algo.CalibrationManager{NumLevels: numLevels, CalibrationUp: a.calibrationUp, CalibrationDown: a.calibrationDown}
	cm.RepairLevelTables(Ec)

	// Reset physics state to neutral after calibration
	if a.useLKSolver() {
		if a.lkSolver != nil {
			resetP := a.lkDefaultPolarization()
			a.lkSolver.SetState(resetP)
			a.lkSolver.Time = 0
			a.polarization = a.lkSolver.GetState()
		} else {
			a.polarization = a.lkDefaultPolarization()
		}
	} else if a.preisach != nil {
		a.preisach.Reset()
		a.polarization = 0
	}
	if !a.useLKSolver() && a.preisach == nil {
		a.polarization = 0
	}
	a.electricField = 0
	a.normalizedP = 0
	a.syncDiscreteLevelLocked()

	// Sync to CalibrationManager (Refactoring)
	if a.calibManager != nil {
		a.calibManager.CalibrationUp = append([]float64(nil), a.calibrationUp...)
		a.calibManager.CalibrationDown = append([]float64(nil), a.calibrationDown...)
		a.calibManager.CalibUpLow = append([]float64(nil), a.calibUpLow...)
		a.calibManager.CalibUpHigh = append([]float64(nil), a.calibUpHigh...)
		a.calibManager.CalibDownLow = append([]float64(nil), a.calibDownLow...)
		a.calibManager.CalibDownHigh = append([]float64(nil), a.calibDownHigh...)
		a.calibManager.LastErrorUp = append([]int(nil), a.lastErrorUp...)
		a.calibManager.LastErrorDown = append([]int(nil), a.lastErrorDown...)
		a.calibManager.RelaxCompUp = append([]float64(nil), a.relaxCompUp...)
		a.calibManager.RelaxCompDown = append([]float64(nil), a.relaxCompDown...)
	}

	a.calibrated = true
}

// enforceMonotonicityUp ensures calibrationUp[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
// MUST be called with a.mu held.
func (a *App) enforceMonotonicityUp(idx int, Ec float64) {
	if idx <= 0 || idx >= len(a.calibrationUp) {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is greater than previous level
	if a.calibrationUp[idx] <= a.calibrationUp[idx-1] {
		a.calibrationUp[idx] = a.calibrationUp[idx-1] + step
	}

	// Ensure value is less than next level (if not at end)
	if idx < len(a.calibrationUp)-1 && a.calibrationUp[idx] >= a.calibrationUp[idx+1] {
		// Push next value up to maintain monotonicity
		a.calibrationUp[idx+1] = a.calibrationUp[idx] + step
	}
}

// enforceMonotonicityDown ensures calibrationDown[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
// MUST be called with a.mu held.
func (a *App) enforceMonotonicityDown(idx int, Ec float64) {
	if idx <= 0 || idx >= len(a.calibrationDown)-1 {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is more negative than next level (lower index = more negative)
	if a.calibrationDown[idx] >= a.calibrationDown[idx+1] {
		a.calibrationDown[idx] = a.calibrationDown[idx+1] - step
	}

	// Ensure value is less negative than previous level
	if idx > 0 && a.calibrationDown[idx] <= a.calibrationDown[idx-1] {
		// Push previous value down to maintain monotonicity
		a.calibrationDown[idx-1] = a.calibrationDown[idx] - step
	}
}

// updateCalibrationUp updates ascending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
// MUST be called with a.mu held.
func (a *App) updateCalibrationUp(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(a.calibrationUp) {
		return
	}

	currentE := a.calibrationUp[targetIdx]
	lastErr := a.lastErrorUp[targetIdx]

	// Update bounds based on error direction
	if levelError > 0 {
		// Overshot (read level > target) → field too strong → update upper bound
		if currentE < a.calibUpHigh[targetIdx] {
			a.calibUpHigh[targetIdx] = currentE
		}
	} else {
		// Undershot (read level < target) → field too weak → update lower bound
		if currentE > a.calibUpLow[targetIdx] {
			a.calibUpLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if a.calibUpLow[targetIdx] < a.calibUpHigh[targetIdx] {
		newVal = (a.calibUpLow[targetIdx] + a.calibUpHigh[targetIdx]) / 2
	} else {
		// Bounds crossed - use direct error-proportional adjustment
		// This is more aggressive than midpoint to escape stuck states
		adjustment := float64(levelError) * Ec * 0.05 // 5% per level error
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Soft constraints: guide towards reasonable range but don't hard-clamp
	// This allows the binary search to explore beyond if physics requires it
	minE := Ec * 0.3 // Absolute minimum: 0.3×Ec
	maxE := Ec * 2.5 // Absolute maximum: 2.5×Ec

	if newVal < minE {
		newVal = minE
	} else if newVal > maxE {
		newVal = maxE
	}

	oldVal := a.calibrationUp[targetIdx]
	a.calibrationUp[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For ascending calibration: higher indices = higher fields
	//   calibrationUp[0] lowest, calibrationUp[29] highest
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal < oldVal {
		// Made field LOWER - cascade to lower indices (they must be even lower)
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := a.calibrationUp[i+1] - step // Must be lower than next
			if a.calibrationUp[i] > maxAllowed {
				a.calibrationUp[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal > oldVal {
		// Made field HIGHER - cascade to higher indices (they must be even higher)
		for i := targetIdx + 1; i < len(a.calibrationUp); i++ {
			minAllowed := a.calibrationUp[i-1] + step // Must be higher than previous
			if a.calibrationUp[i] < minAllowed {
				a.calibrationUp[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call enforceMonotonicityUp - cascade handles it and won't fight with newVal
	a.lastErrorUp[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(a.relaxCompUp) {
		oldRelaxComp := a.relaxCompUp[targetIdx]
		var targetRelax float64
		if levelError > 0 {
			// Overshot: read level > target, reduce compensation (we overshot too much)
			targetRelax = oldRelaxComp - relaxAdjustRate*float64(levelError)
		} else {
			// Undershot: read level < target, increase compensation (need more overshoot)
			targetRelax = oldRelaxComp - relaxAdjustRate*float64(levelError) // levelError is negative, so this adds
		}
		a.relaxCompUp[targetIdx] = emaRetainFrac*oldRelaxComp + emaUpdateFrac*targetRelax
		if a.relaxCompUp[targetIdx] < relaxCompMin {
			a.relaxCompUp[targetIdx] = relaxCompMin
		} else if a.relaxCompUp[targetIdx] > relaxCompMax {
			a.relaxCompUp[targetIdx] = relaxCompMax
		}
		if a.relaxCompUp[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_UP[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, a.relaxCompUp[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_UP[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		a.calibUpLow[targetIdx]/1e8, a.calibUpHigh[targetIdx]/1e8)

	// Refactoring: Sync to CalibrationManager
	if a.calibManager != nil {
		copy(a.calibManager.CalibrationUp, a.calibrationUp)
	}
}

// updateCalibrationDown updates descending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
// MUST be called with a.mu held.
func (a *App) updateCalibrationDown(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(a.calibrationDown) {
		return
	}

	currentE := a.calibrationDown[targetIdx]
	lastErr := a.lastErrorDown[targetIdx]

	// Update bounds based on error direction (descending uses negative fields)
	// calibDownLow = most negative limit (e.g., -2.5×Ec)
	// calibDownHigh = least negative limit (e.g., -0.3×Ec)
	// midpoint searches between them
	if levelError > 0 {
		// Positive error = didn't go down enough = field not negative enough
		// Current field is too weak → becomes new UPPER bound (least negative limit)
		// Next search will be between Low and currentE (more negative)
		if currentE < a.calibDownHigh[targetIdx] {
			a.calibDownHigh[targetIdx] = currentE
		}
	} else {
		// Negative error = went too far down = field too negative
		// Current field is too strong → becomes new LOWER bound (most negative limit)
		// Next search will be between currentE and High (less negative)
		if currentE > a.calibDownLow[targetIdx] {
			a.calibDownLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if a.calibDownLow[targetIdx] < a.calibDownHigh[targetIdx] {
		newVal = (a.calibDownLow[targetIdx] + a.calibDownHigh[targetIdx]) / 2
	} else {
		// Bounds crossed - use direct error-proportional adjustment
		// For descending, positive error means we need MORE negative field
		adjustment := float64(levelError) * Ec * 0.05 // 5% per level error
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Soft constraints: guide towards reasonable range but don't hard-clamp
	minE := -Ec * 2.5 // Absolute minimum (most negative): -2.5×Ec
	maxE := -Ec * 0.3 // Absolute maximum (least negative): -0.3×Ec

	if newVal > maxE {
		newVal = maxE
	} else if newVal < minE {
		newVal = minE
	}

	oldVal := a.calibrationDown[targetIdx]
	a.calibrationDown[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For descending calibration: lower indices = more negative (stronger field)
	//   calibrationDown[0] most negative, calibrationDown[29] least negative (or 0)
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal > oldVal {
		// Made field LESS negative (weaker) - cascade to higher indices
		// Higher indices must be even less negative to maintain monotonicity
		for i := targetIdx + 1; i < len(a.calibrationDown); i++ {
			minAllowed := a.calibrationDown[i-1] + step // Must be less negative than previous
			if a.calibrationDown[i] < minAllowed {
				a.calibrationDown[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal < oldVal {
		// Made field MORE negative (stronger) - cascade to lower indices
		// Lower indices must be even more negative to maintain monotonicity
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := a.calibrationDown[i+1] - step // Must be more negative than next
			if a.calibrationDown[i] > maxAllowed {
				a.calibrationDown[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call enforceMonotonicityDown - cascade handles it and won't fight with newVal
	a.lastErrorDown[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(a.relaxCompDown) {
		oldRelaxComp := a.relaxCompDown[targetIdx]
		var targetRelax float64
		if levelError < 0 {
			// Went too far down (error < 0), reduce compensation
			targetRelax = oldRelaxComp + relaxAdjustRate*float64(levelError) // levelError is negative, so this subtracts
		} else {
			// Didn't go down enough (error > 0), increase compensation
			targetRelax = oldRelaxComp + relaxAdjustRate*float64(levelError)
		}
		a.relaxCompDown[targetIdx] = emaRetainFrac*oldRelaxComp + emaUpdateFrac*targetRelax
		if a.relaxCompDown[targetIdx] < relaxCompMin {
			a.relaxCompDown[targetIdx] = relaxCompMin
		} else if a.relaxCompDown[targetIdx] > relaxCompMax {
			a.relaxCompDown[targetIdx] = relaxCompMax
		}
		if a.relaxCompDown[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_DOWN[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, a.relaxCompDown[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_DOWN[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		a.calibDownLow[targetIdx]/1e8, a.calibDownHigh[targetIdx]/1e8)

	// Refactoring: Sync to CalibrationManager
	if a.calibManager != nil {
		copy(a.calibManager.CalibrationDown, a.calibrationDown)
	}
}
