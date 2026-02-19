// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains write-mode physics orchestration for the unified view.
package gui

// write_physics.go — Write-Mode Physics Orchestration
//
// Passive 0T1R (DAC-only column drive):
//   - WLs are grounded at 0V by the TIA virtual ground (hardware constraint — cannot be driven)
//   - Selected BL is driven to −V_write by the DAC
//   - Effective ΔV across cell = WL − BL = 0 − (−V_write) = +V_write  (full switching)
//   - Consequence: ALL cells in the selected column switch; unselected columns see 0V (safe)
//   - Same-row cells see 0V (WL=0, unselected BL=0): no disturb, no overlay
//
// Active 1T1R/2T1R (transistor isolation):
//   - Selected WL (gate) is raised to enable the target transistor
//   - Selected BL is driven to +V_write; all other BLs = 0V
//   - Transistors isolate non-selected cells: only the target cell switches
//   - No column-disturb; safe single-cell write
//
// Entry points consumed by the ISPP goroutines in tab_unified_voltage.go:
//   applyWriteVoltages      — quantise voltage through DAC, route to passive or active path
//   applyWritePhaseVoltages — drive the 5-phase write-sequence voltage for each phase
//   applyColumnWrite        — (passive only) apply full column disturb to all non-target cells
//   applyHalfSelectDisturb  — always returns 0; V/2 disturb does not exist in DAC-only drive

import "math"

func (ca *CircuitsApp) applyWritePhaseVoltages(phaseInfo WriteSequenceState) {
	if ca.deviceState == nil {
		return
	}

	row := phaseInfo.TargetRow
	col := phaseInfo.TargetCol
	isPassive := ca.deviceState.IsPassiveMode()

	switch phaseInfo.Phase {
	case PhaseWrite:
		writeVoltage := phaseInfo.PhaseVoltage
		if isppStatus := ca.deviceState.GetISPPStatus(); isppStatus.Active {
			writeVoltage = isppStatus.Voltage
		}
		if isPassive {
			ca.deviceState.ApplyHalfSelectWrite(row, col, writeVoltage)
		} else {
			ca.deviceState.SetWLSingle(row)
			ca.deviceState.SetAllDACVoltages(0)
			ca.deviceState.SetDACVoltage(col, writeVoltage)
		}
		ca.deviceState.SetDACRangeMode(DACRangeWrite)
		neighborChanges := ca.applyHalfSelectDisturb(row, col)
		if neighborChanges > 0 {
			logAction("write_disturb rows=%d cols=%d changes=%d", ca.arrayRows, ca.arrayCols, neighborChanges)
		}

	case PhaseVerify:
		verifyVoltage := phaseInfo.PhaseVoltage
		if !isPassive {
			ca.deviceState.SetWLSingle(row)
		}
		ca.deviceState.SetAllDACVoltages(0)
		ca.deviceState.SetDACVoltage(col, verifyVoltage)
		ca.deviceState.SetDACRangeMode(DACRangeRead)

	default:
		ca.deviceState.ResetWriteVoltages()
	}

	ca.recomputeAndRefresh()
}

// applyHalfSelectDisturb accumulates V/2 stress on half-selected cells.
// In passive 0T1R with DAC-only column drive:
//   - Same-column cells see the full write voltage → handled by applyColumnWrite
//   - Same-row cells see 0V (unselected BLs grounded)  → no disturb
//
// The V/2 half-select stress model does not apply to this architecture; this
// function always returns 0.
func (ca *CircuitsApp) applyHalfSelectDisturb(targetRow, targetCol int) int {
	if ca.deviceState == nil || !ca.deviceState.IsPassiveMode() {
		return 0
	}
	return 0
}

// applyColumnWrite applies the write voltage physics to every cell in the target column
// except the selected cell (which is managed by the ISPP loop).
//
// In passive 0T mode with DAC-only column drive (all WLs grounded, selected BL at V_write)
// every cell in the column sees the same full write voltage. This function simulates that
// effect by running a single-pulse LK step for each non-selected cell in the column.
func (ca *CircuitsApp) applyColumnWrite(selectedRow, col int, writeVoltage float64) {
	if ca.deviceState == nil || !ca.deviceState.IsPassiveMode() {
		return
	}
	if math.Abs(writeVoltage) < 1e-12 {
		return
	}
	nRows := ca.arrayRows
	if nRows == 0 {
		return
	}

	// Read current levels outside ca.mu to avoid holding two locks while calling
	// programLevelFromCoupledVoltage (which acquires ds.mu internally).
	ca.mu.RLock()
	currentLevels := make([]int, nRows)
	for r := 0; r < nRows && r < len(ca.arrayWeights); r++ {
		if col < len(ca.arrayWeights[r]) {
			currentLevels[r] = ca.arrayWeights[r][col]
		}
	}
	ca.mu.RUnlock()

	// Each cell independently responds to the write voltage based on its own state.
	newLevels := make([]int, nRows)
	copy(newLevels, currentLevels)
	for r := 0; r < nRows; r++ {
		if r == selectedRow {
			continue // selected cell is driven by the ISPP loop
		}
		newLevels[r] = ca.deviceState.programLevelFromCoupledVoltage(
			currentLevels[r], writeVoltage,
			float64(PhaseWriteDurationNs)*1e-9, ca.quantLevels,
		)
	}

	ca.mu.Lock()
	for r := 0; r < nRows && r < len(ca.arrayWeights); r++ {
		if r == selectedRow {
			continue
		}
		if col < len(ca.arrayWeights[r]) {
			ca.arrayWeights[r][col] = newLevels[r]
		}
	}
	ca.mu.Unlock()
}

// applyWriteVoltages converts the target voltage through the DAC and applies
// the resulting voltages to the array (DAC-only column drive in passive 0T1R mode).
func (ca *CircuitsApp) applyWriteVoltages(row, col int, targetVoltage float64) (float64, int) {
	if ca.deviceState == nil {
		return targetVoltage, -1
	}
	applied, dacCode := ca.deviceState.DACWriteVoltage(targetVoltage)

	if ca.deviceState.IsPassiveMode() {
		ca.deviceState.ApplyHalfSelectWrite(row, col, applied)
	} else {
		// Non-passive: pass-through voltages should be 0 on unselected BLs.
		ca.deviceState.SetAllDACVoltages(0)
		ca.deviceState.SetDACVoltage(col, applied)
	}
	ca.deviceState.SetDACRangeMode(DACRangeWrite)

	return applied, dacCode
}
