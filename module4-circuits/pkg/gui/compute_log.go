//go:build legacy_fyne

// Package gui provides compute logging for debugging MVM operations.
package gui

import (
	"fmt"
	"time"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/module4-circuits/pkg/gui/computelog"
)

type ComputeLogEntry = computelog.Entry
type ComputeRowResult = computelog.RowResult
type CellMVM = computelog.CellMVM

type computeLogAdapter struct {
	*computelog.Log
}

var globalComputeLog = &computeLogAdapter{Log: computelog.New()}

// EnableComputeLog enables or disables compute logging.
func EnableComputeLog(enabled bool) {
	globalComputeLog.Enable(enabled)
}

// ComputeLogEnabled reports whether compute logging is currently enabled.
func ComputeLogEnabled() bool {
	return globalComputeLog.Enabled()
}

// SetComputeLogPath sets the path for the compute log file.
// Returns an error if the path is empty or contains traversal sequences.
func SetComputeLogPath(path string) error {
	return globalComputeLog.SetPath(path)
}

// ClearComputeLog clears all logged entries.
func ClearComputeLog() {
	globalComputeLog.Clear()
}

// GetComputeLogEntries returns a copy of all logged entries.
func GetComputeLogEntries() []ComputeLogEntry {
	return globalComputeLog.Entries()
}

// LogCompute logs a compute operation (called from DeviceState.Compute).
func (ds *DeviceState) LogCompute(weights [][]int, quantLevels int) {
	if !globalComputeLog.Enabled() {
		return
	}

	entry := ComputeLogEntry{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05.000"),
		ArraySize:   fmt.Sprintf("%dx%d", ds.rows, ds.cols),
		QuantLevels: quantLevels,
	}

	// Material name
	if ds.material != nil {
		entry.Material = ds.material.Name
	} else {
		entry.Material = "default"
	}

	// Input vector (DAC voltages)
	entry.InputVector = make([]float64, ds.cols)
	copy(entry.InputVector, ds.dacVoltages)

	// Weight matrix
	entry.Weights = make([][]int, len(weights))
	for r := range weights {
		entry.Weights[r] = make([]int, len(weights[r]))
		copy(entry.Weights[r], weights[r])
	}

	// Conductance matrix
	entry.Conductances = make([][]float64, ds.rows)
	for r := 0; r < ds.rows; r++ {
		entry.Conductances[r] = make([]float64, ds.cols)
		for c := 0; c < ds.cols; c++ {
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			var conductanceS float64
			if ds.material != nil {
				conductanceS = ds.material.DiscreteLevel(level, quantLevels)
			} else {
				conductanceS = (1.0 + float64(level)/float64(quantLevels-1)*99.0) * 1e-6
			}
			entry.Conductances[r][c] = conductanceS * 1e6 // Convert to µS
		}
	}

	// Row results with cell details
	entry.RowResults = make([]ComputeRowResult, ds.rows)
	for r := 0; r < ds.rows; r++ {
		result := ComputeRowResult{
			Row:        r,
			Active:     ds.activeRows[r],
			CurrentUA:  ds.rowCurrents[r],
			TIAVoltage: ds.rowVoltages[r],
			ADCLevel:   ds.rowLevels[r],
			Saturated:  ds.saturated[r],
			CellDetail: make([]CellMVM, ds.cols),
		}

		// Per-cell breakdown
		for c := 0; c < ds.cols; c++ {
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			conductanceUS := entry.Conductances[r][c]
			voltage := ds.dacVoltages[c]
			currentUA := conductanceUS * voltage
			if ds.couplingMode == arraysim.CouplingTierA && ds.coupledCellVoltages != nil {
				if r < len(ds.coupledCellVoltages) && c < len(ds.coupledCellVoltages[r]) {
					voltage = ds.coupledCellVoltages[r][c]
				}
			}
			if ds.couplingMode == arraysim.CouplingTierA && ds.coupledCellCurrents != nil {
				if r < len(ds.coupledCellCurrents) && c < len(ds.coupledCellCurrents[r]) {
					currentUA = ds.coupledCellCurrents[r][c] * 1e6
				}
			}

			result.CellDetail[c] = CellMVM{
				Col:           c,
				Weight:        level,
				ConductanceUS: conductanceUS,
				VoltageV:      voltage,
				CurrentUA:     currentUA,
			}
		}
		entry.RowResults[r] = result
	}

	globalComputeLog.Append(entry, 100)
}

// SaveComputeLog saves all logged entries to the JSON file.
func SaveComputeLog() error {
	return globalComputeLog.Save()
}

// SaveComputeLogTo saves all logged entries to a specific file.
func SaveComputeLogTo(path string) error {
	return globalComputeLog.SaveTo(path)
}
