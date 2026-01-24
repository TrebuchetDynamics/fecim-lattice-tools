// Package gui provides Fyne-based GUI components for crossbar visualization.
// callbacks.go contains event handlers for heatmap cell interactions.
package gui

import (
	"fmt"
	"math"

	"multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
)

// syncSelection updates the app-level selection and syncs it to all heatmaps.
func (ca *CrossbarApp) syncSelection(row, col int) {
	ca.selectedRow = row
	ca.selectedCol = col

	// Sync to all heatmaps
	if ca.conductanceHeatmap != nil {
		ca.conductanceHeatmap.SetSelection(row, col)
	}
	if ca.irDropHeatmap != nil {
		ca.irDropHeatmap.SetSelection(row, col)
	}
	if ca.sneakPathHeatmap != nil {
		ca.sneakPathHeatmap.SetSelection(row, col)
	}
	if ca.beforeAfterToggle != nil && ca.beforeAfterToggle.leftHeatmap != nil {
		ca.beforeAfterToggle.leftHeatmap.SetSelection(row, col)
		ca.beforeAfterToggle.rightHeatmap.SetSelection(row, col)
	}
}

// onCellTapped handles clicks on heatmap cells.
func (ca *CrossbarApp) onCellTapped(row, col int) {
	ca.modeIndicator.SetMode(DemoModeRead)

	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	matrix := ca.array.GetConductanceMatrix()
	value := matrix[row][col]
	level := crossbar.GetLevel(value)

	ca.levelIndicator.SetLevel(level)

	// Generate comprehensive tooltip
	tooltip := ConductanceTooltip(row, col, value, ca.array)

	// Display in stats label (formatted for readability)
	ca.statsLabel.SetText(tooltip)

	ca.updateStatus(fmt.Sprintf("READ | Cell [%d,%d] = Level %d/30 (%.2f µS)",
		row, col, level, value*99+1))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// onCellHover handles mouse hover over heatmap cells.
func (ca *CrossbarApp) onCellHover(row, col int, value float64) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells to see detailed physics data")
		return
	}
	level := crossbar.GetLevel(value)
	conductanceUS := value*99 + 1
	resistance := 1.0 / (conductanceUS * 1e-6) / 1000.0 // kΩ

	ca.hoverInfoLabel.SetText(fmt.Sprintf(
		"[%d,%d] │ L%d/29 (%.1f%%) │ G=%.2f µS │ R=%.1f kΩ │ %s %.6f",
		row, col, level, float64(level)/29.0*100, conductanceUS, resistance,
		"Norm:", value))
}

// onIRDropCellTapped handles clicks on IR Drop heatmap.
func (ca *CrossbarApp) onIRDropCellTapped(row, col int) {
	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	// Protected read of lastIRDropAnalysis
	ca.stateMu.RLock()
	analysis := ca.lastIRDropAnalysis
	ca.stateMu.RUnlock()

	// Generate comprehensive IR drop tooltip
	tooltip := IRDropTooltip(row, col, analysis, ca.array)
	ca.statsLabel.SetText(tooltip)

	// Update status with key info
	if analysis != nil && row < len(analysis.EffectiveVoltage) &&
		col < len(analysis.EffectiveVoltage[0]) {
		effectiveV := analysis.EffectiveVoltage[row][col]
		dropPercent := (1.0 - effectiveV) * 100
		ca.updateStatus(fmt.Sprintf("IR DROP | Cell [%d,%d]: %.3f V (%.1f%% drop)",
			row, col, effectiveV, dropPercent))
	}
}

// onIRDropCellHover handles hover on IR Drop heatmap.
func (ca *CrossbarApp) onIRDropCellHover(row, col int, value float64) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells for IR drop details")
		return
	}

	conductance := ca.array.GetConductanceMatrix()[row][col]
	level := crossbar.GetLevel(conductance)
	conductanceUS := conductance*99 + 1

	// Protected read of lastIRDropAnalysis
	ca.stateMu.RLock()
	analysis := ca.lastIRDropAnalysis
	ca.stateMu.RUnlock()

	// Get detailed voltage info if available
	if analysis != nil && row < len(analysis.EffectiveVoltage) &&
		col < len(analysis.EffectiveVoltage[0]) {
		effectiveV := analysis.EffectiveVoltage[row][col]
		wlV := analysis.WordLineVoltages[row][col]
		blV := analysis.BitLineVoltages[row][col]
		dropPercent := (1.0 - effectiveV) * 100

		// Calculate distance from drivers (WL driver on left, BL sense amp at top)
		wlDist := col // Distance from left WL driver
		blDist := row // Distance from top BL sense amp

		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] │ Veff=%.3fV (%.1f%% drop) │ WL=%.3fV BL=%.3fV │ G=%.1fµS L%d │ Dist=[WL:%d,BL:%d]",
			row, col, effectiveV, dropPercent, wlV, blV, conductanceUS, level, wlDist, blDist))
	} else {
		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] │ G=%.1f µS │ L%d/29 │ Run MVM for IR drop analysis",
			row, col, conductanceUS, level))
	}
}

// onSneakCellTapped handles clicks on Sneak Path heatmap.
func (ca *CrossbarApp) onSneakCellTapped(row, col int) {
	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	// Get selected target cell for sneak analysis (typically center)
	sneakTargetRow := ca.config.Rows / 2
	sneakTargetCol := ca.config.Cols / 2

	// Protected read of lastSneakAnalysis
	ca.stateMu.RLock()
	analysis := ca.lastSneakAnalysis
	ca.stateMu.RUnlock()

	// Generate comprehensive sneak path tooltip
	tooltip := SneakPathTooltip(row, col, analysis, sneakTargetRow, sneakTargetCol, ca.array)
	ca.statsLabel.SetText(tooltip)

	// Update status with key info
	if analysis != nil && row < len(analysis.SneakCurrents) &&
		col < len(analysis.SneakCurrents[0]) {
		sneakCurrent := analysis.SneakCurrents[row][col]
		sneakRatio := 0.0
		if analysis.TotalSignal > 0 {
			sneakRatio = sneakCurrent / analysis.TotalSignal * 100
		}
		ca.updateStatus(fmt.Sprintf("SNEAK | Cell [%d,%d]: %.6f µA (%.2f%% of signal)",
			row, col, sneakCurrent*1e6, sneakRatio))
	}
}

// onSneakCellHover handles hover on Sneak Path heatmap.
func (ca *CrossbarApp) onSneakCellHover(row, col int, value float64) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells for sneak path details")
		return
	}

	conductance := ca.array.GetConductanceMatrix()[row][col]
	level := crossbar.GetLevel(conductance)
	conductanceUS := conductance*99 + 1

	// Get selected cell (center)
	selectedRow := ca.config.Rows / 2
	selectedCol := ca.config.Cols / 2

	// Protected read of lastSneakAnalysis
	ca.stateMu.RLock()
	analysis := ca.lastSneakAnalysis
	ca.stateMu.RUnlock()

	// Get detailed sneak info if available
	if analysis != nil && row < len(analysis.SneakCurrents) &&
		col < len(analysis.SneakCurrents[0]) {
		sneakCurrent := analysis.SneakCurrents[row][col]
		sneakRatio := 0.0
		if analysis.TotalSignal > 0 {
			sneakRatio = sneakCurrent / analysis.TotalSignal * 100
		}

		// Determine path type
		pathType := "DIAG"
		if row == selectedRow && col == selectedCol {
			pathType = "TGT"
		} else if row == selectedRow {
			pathType = "ROW"
		} else if col == selectedCol {
			pathType = "COL"
		}

		// SNR in dB
		snrDB := -100.0
		if sneakCurrent > 0 && analysis.TotalSignal > 0 {
			snrDB = 20 * math.Log10(analysis.TotalSignal / sneakCurrent)
		}

		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] │ %s sneak │ I=%.3fµA (%.2f%%) │ SNR=%.1fdB │ G=%.1fµS L%d",
			row, col, pathType, sneakCurrent*1e6, sneakRatio, snrDB, conductanceUS, level))
	} else {
		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] │ G=%.1f µS │ L%d/29 │ Run MVM for sneak analysis",
			row, col, conductanceUS, level))
	}
}
