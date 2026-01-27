// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device state for the simulation view.
package gui

import (
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module4-circuits/pkg/peripherals"
)

// OperationMode represents the current operation mode (legacy, kept for compatibility)
type OperationMode int

const (
	ModeWrite OperationMode = iota
	ModeRead
	ModeCompute
)

// WLMode represents word line selection mode
type WLMode int

const (
	WLSingle WLMode = iota // One row selected (for program/read single cell)
	WLAll                  // All rows active (for MVM compute)
	WLCustom               // User-defined pattern
)

// DACMode represents how DAC voltages were set
type DACMode int

const (
	DACManual DACMode = iota // User entered each voltage
	DACReadPreset            // All columns at readVoltage (0-1V range)
	DACWritePreset           // Selected column at write voltage (Vmin-Vmax range)
	DACInputVector           // From digital input vector (0-255 -> 0-1V)
	DACRandom                // Random voltages
)

// DACRangeMode represents the DAC output range mode
type DACRangeMode int

const (
	DACRangeRead  DACRangeMode = iota // 0 to 1V (read/compute safe zone)
	DACRangeWrite                     // MinWriteV to MaxWriteV (write zone)
)

// VoltageRange holds the min/max voltages for a given operation
type VoltageRange struct {
	Min       float64 // Minimum voltage
	Max       float64 // Maximum voltage
	StepSize  float64 // Voltage step between states
	NumLevels int     // Number of discrete levels
}

// Default voltage thresholds (used when no calibration data available)
const (
	DefaultVoltageReadMax   = 1.0  // Max safe read voltage
	DefaultVoltageWriteMin  = 1.2  // Min write voltage (above Vc)
	DefaultVoltageWriteMax  = 1.5  // Max write voltage
	DefaultVoltageComputeMax = 1.0 // Max voltage for compute (MVM)
)

// DeviceState holds the unified simulation state
type DeviceState struct {
	// Dimensions
	rows int
	cols int

	// WL configuration
	wlMode     WLMode
	activeRows []bool // true = WL HIGH for that row

	// DAC inputs (per column)
	dacVoltages  []float64
	dacMode      DACMode
	dacRangeMode DACRangeMode // Current DAC range (read vs write)

	// Voltage ranges (derived from material calibration)
	readRange  VoltageRange // 0 to ~1V for read/compute
	writeRange VoltageRange // Vc to Vmax for write operations

	// Computed outputs (per row)
	rowCurrents []float64 // TIA input currents (uA)
	rowVoltages []float64 // TIA output voltages (V)
	rowLevels   []int     // ADC output levels

	// Saturation flags
	saturated []bool

	// Selected cell (for single-cell operations)
	selectedRow int
	selectedCol int

	// Material physics model (from hysteresis calibration)
	material *ferroelectric.HZOMaterial

	// Peripherals reference
	tia *peripherals.TIA
	adc *peripherals.ADC
}

// NewDeviceState creates a new device state with specified dimensions
func NewDeviceState(rows, cols int, tia *peripherals.TIA, adc *peripherals.ADC) *DeviceState {
	ds := &DeviceState{
		rows:         rows,
		cols:         cols,
		wlMode:       WLSingle,
		activeRows:   make([]bool, rows),
		dacVoltages:  make([]float64, cols),
		dacMode:      DACReadPreset,
		dacRangeMode: DACRangeRead,
		rowCurrents:  make([]float64, rows),
		rowVoltages:  make([]float64, rows),
		rowLevels:    make([]int, rows),
		saturated:    make([]bool, rows),
		selectedRow:  0,
		selectedCol:  0,
		material:     ferroelectric.FeCIMMaterial(), // Default to FeCIM material
		tia:          tia,
		adc:          adc,
	}

	// Calculate voltage ranges from material properties
	ds.updateVoltageRanges()

	// Initialize with read preset (uses read range)
	ds.SetDACRangeMode(DACRangeRead)
	ds.SetDACPreset(DACReadPreset)

	// Default: single row 0 active
	ds.activeRows[0] = true

	return ds
}

// updateVoltageRanges calculates voltage ranges from material properties
// Read range: 0 to safe_read_voltage (below Vc to avoid disturbing states)
// Write range: Vc (coercive voltage) to ~1.5*Vc (programming window)
func (ds *DeviceState) updateVoltageRanges() {
	if ds.material == nil {
		// Use defaults if no material
		ds.readRange = VoltageRange{Min: 0, Max: DefaultVoltageReadMax, NumLevels: 30}
		ds.writeRange = VoltageRange{Min: DefaultVoltageWriteMin, Max: DefaultVoltageWriteMax, NumLevels: 30}
		return
	}

	// Coercive voltage = Ec * thickness
	Vc := ds.material.CoerciveVoltage()

	// Read range: 0 to ~0.8*Vc (safe zone, won't disturb polarization)
	// For typical HZO: Vc ~ 1.2V, so safe read up to ~1V
	safeReadMax := 0.8 * Vc
	if safeReadMax > 1.0 {
		safeReadMax = 1.0 // Cap at 1V for practical DAC range
	}
	if safeReadMax < 0.3 {
		safeReadMax = 0.3 // Minimum useful read voltage
	}

	ds.readRange = VoltageRange{
		Min:       0,
		Max:       safeReadMax,
		NumLevels: 30,
	}

	// Write range: Vc to ~1.3*Vc (programming window)
	// Need to exceed Vc to switch polarization
	writeMin := Vc
	writeMax := 1.3 * Vc
	if writeMax > 2.0 {
		writeMax = 2.0 // Practical limit for most devices
	}

	ds.writeRange = VoltageRange{
		Min:       writeMin,
		Max:       writeMax,
		StepSize:  (writeMax - writeMin) / 29.0, // 30 levels
		NumLevels: 30,
	}
}

// SetMaterial changes the ferroelectric material used for conductance calculation
func (ds *DeviceState) SetMaterial(mat *ferroelectric.HZOMaterial) {
	ds.material = mat
	ds.updateVoltageRanges() // Recalculate voltage ranges for new material
}

// GetMaterial returns the current material
func (ds *DeviceState) GetMaterial() *ferroelectric.HZOMaterial {
	return ds.material
}

// GetMaterialName returns the name of the current material
func (ds *DeviceState) GetMaterialName() string {
	if ds.material != nil {
		return ds.material.Name
	}
	return "Unknown"
}

// GetReadRange returns the voltage range for read/compute operations
func (ds *DeviceState) GetReadRange() VoltageRange {
	return ds.readRange
}

// GetWriteRange returns the voltage range for write operations
func (ds *DeviceState) GetWriteRange() VoltageRange {
	return ds.writeRange
}

// GetDACRangeMode returns the current DAC range mode
func (ds *DeviceState) GetDACRangeMode() DACRangeMode {
	return ds.dacRangeMode
}

// SetDACRangeMode sets the DAC range mode (read vs write)
func (ds *DeviceState) SetDACRangeMode(mode DACRangeMode) {
	ds.dacRangeMode = mode
}

// GetCurrentVoltageRange returns the voltage range for the current mode
func (ds *DeviceState) GetCurrentVoltageRange() VoltageRange {
	if ds.dacRangeMode == DACRangeWrite {
		return ds.writeRange
	}
	return ds.readRange
}

// SetWLSingle activates only the specified row
func (ds *DeviceState) SetWLSingle(row int) {
	ds.wlMode = WLSingle
	ds.selectedRow = row
	for i := range ds.activeRows {
		ds.activeRows[i] = (i == row)
	}
}

// SetWLAll activates all rows for MVM
func (ds *DeviceState) SetWLAll() {
	ds.wlMode = WLAll
	for i := range ds.activeRows {
		ds.activeRows[i] = true
	}
}

// SetWLCustom sets a custom WL pattern
func (ds *DeviceState) SetWLCustom(pattern []bool) {
	ds.wlMode = WLCustom
	copy(ds.activeRows, pattern)
}

// SetDACVoltage sets voltage for a single column
func (ds *DeviceState) SetDACVoltage(col int, voltage float64) {
	if col >= 0 && col < ds.cols {
		ds.dacVoltages[col] = voltage
		ds.dacMode = DACManual
	}
}

// SetDACPreset applies a preset pattern using material-derived voltage ranges
func (ds *DeviceState) SetDACPreset(preset DACMode, params ...float64) {
	ds.dacMode = preset

	switch preset {
	case DACReadPreset:
		// Use read range from material calibration
		ds.dacRangeMode = DACRangeRead
		voltage := ds.readRange.Max * 0.5 // Default to 50% of safe read range
		if len(params) > 0 {
			voltage = params[0]
		}
		// Clamp to read range
		if voltage > ds.readRange.Max {
			voltage = ds.readRange.Max
		}
		for i := range ds.dacVoltages {
			ds.dacVoltages[i] = voltage
		}

	case DACWritePreset:
		// Use write range from material calibration
		ds.dacRangeMode = DACRangeWrite
		// Default to middle of write range for selected column
		writeVoltage := (ds.writeRange.Min + ds.writeRange.Max) / 2
		if len(params) > 0 {
			writeVoltage = params[0]
		}
		// Clamp to write range
		if writeVoltage < ds.writeRange.Min {
			writeVoltage = ds.writeRange.Min
		}
		if writeVoltage > ds.writeRange.Max {
			writeVoltage = ds.writeRange.Max
		}
		for i := range ds.dacVoltages {
			if i == ds.selectedCol {
				ds.dacVoltages[i] = writeVoltage
			} else {
				ds.dacVoltages[i] = 0
			}
		}

	case DACInputVector:
		// Convert input vector (0-255) to voltage using read range
		// Maps 0-255 to readRange.Min-readRange.Max
		ds.dacRangeMode = DACRangeRead
		for i := range ds.dacVoltages {
			if i < len(params) {
				normalized := params[i] / 255.0
				ds.dacVoltages[i] = ds.readRange.Min + normalized*(ds.readRange.Max-ds.readRange.Min)
			}
		}

	case DACRandom:
		// Random voltages in read range (compute-safe)
		// Note: actual random generation done by caller
		ds.dacRangeMode = DACRangeRead
	}
}

// SetDACVoltageForState sets the write voltage for a target state (0 to numLevels-1)
// Maps the state to the appropriate voltage in the write range
func (ds *DeviceState) SetDACVoltageForState(col int, targetState int) {
	if col < 0 || col >= ds.cols {
		return
	}

	// Clamp target state
	if targetState < 0 {
		targetState = 0
	}
	if targetState >= ds.writeRange.NumLevels {
		targetState = ds.writeRange.NumLevels - 1
	}

	// Linear interpolation within write range
	normalized := float64(targetState) / float64(ds.writeRange.NumLevels-1)
	voltage := ds.writeRange.Min + normalized*(ds.writeRange.Max-ds.writeRange.Min)

	ds.dacVoltages[col] = voltage
	ds.dacRangeMode = DACRangeWrite
	ds.dacMode = DACManual
}

// SetAllDACVoltages sets all DAC columns to the same voltage
func (ds *DeviceState) SetAllDACVoltages(voltage float64) {
	ds.dacMode = DACManual
	for i := range ds.dacVoltages {
		ds.dacVoltages[i] = voltage
	}
}

// SetSelectedCell sets the currently selected cell
func (ds *DeviceState) SetSelectedCell(row, col int) {
	ds.selectedRow = row
	ds.selectedCol = col
	if ds.wlMode == WLSingle {
		ds.SetWLSingle(row)
	}
}

// Compute runs the device simulation given the weight matrix
func (ds *DeviceState) Compute(weights [][]int, quantLevels int) {
	for r := 0; r < ds.rows; r++ {
		if !ds.activeRows[r] {
			ds.rowCurrents[r] = 0
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		// Sum currents from all active columns
		totalCurrent := 0.0
		for c := 0; c < ds.cols; c++ {
			voltage := ds.dacVoltages[c]
			if voltage < 0.01 {
				continue
			}

			// Get cell conductance from weight using material physics model
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}

			// Use material's DiscreteLevel for physics-accurate conductance
			// DiscreteLevel returns conductance in Siemens (S)
			var conductanceS float64
			if ds.material != nil {
				conductanceS = ds.material.DiscreteLevel(level, quantLevels)
			} else {
				// Fallback: linear mapping 1-100 µS
				conductanceS = (1.0 + float64(level)/float64(quantLevels-1)*99.0) * 1e-6
			}

			// Convert to µS for current calculation
			conductanceUS := conductanceS * 1e6
			current := conductanceUS * voltage // I = G * V (in µA since G is in µS)
			totalCurrent += current
		}

		ds.rowCurrents[r] = totalCurrent

		// TIA conversion: current (A) to voltage (V)
		if ds.tia != nil {
			ds.rowVoltages[r] = ds.tia.Convert(totalCurrent * 1e-6) // µA to A
		}

		// ADC conversion: voltage to level
		if ds.adc != nil {
			ds.rowLevels[r] = ds.adc.Convert(ds.rowVoltages[r])
		}

		// Check saturation (TIA saturates around 100 µA)
		ds.saturated[r] = totalCurrent > 100.0 || ds.rowLevels[r] >= 31
	}
}

// GetRowCurrent returns the computed current for a row
func (ds *DeviceState) GetRowCurrent(row int) float64 {
	if row >= 0 && row < ds.rows {
		return ds.rowCurrents[row]
	}
	return 0
}

// GetRowVoltage returns the TIA output voltage for a row
func (ds *DeviceState) GetRowVoltage(row int) float64 {
	if row >= 0 && row < ds.rows {
		return ds.rowVoltages[row]
	}
	return 0
}

// GetRowLevel returns the ADC output level for a row
func (ds *DeviceState) GetRowLevel(row int) int {
	if row >= 0 && row < ds.rows {
		return ds.rowLevels[row]
	}
	return 0
}

// IsSaturated returns whether a row's output is saturated
func (ds *DeviceState) IsSaturated(row int) bool {
	if row >= 0 && row < ds.rows {
		return ds.saturated[row]
	}
	return false
}

// IsRowActive returns whether a row's WL is active
func (ds *DeviceState) IsRowActive(row int) bool {
	if row >= 0 && row < ds.rows {
		return ds.activeRows[row]
	}
	return false
}

// GetDACVoltage returns the DAC voltage for a column
func (ds *DeviceState) GetDACVoltage(col int) float64 {
	if col >= 0 && col < ds.cols {
		return ds.dacVoltages[col]
	}
	return 0
}

// GetWLMode returns the current WL selection mode
func (ds *DeviceState) GetWLMode() WLMode {
	return ds.wlMode
}

// GetDACMode returns the current DAC preset mode
func (ds *DeviceState) GetDACMode() DACMode {
	return ds.dacMode
}

// GetSelectedRow returns the selected row index
func (ds *DeviceState) GetSelectedRow() int {
	return ds.selectedRow
}

// GetSelectedCol returns the selected column index
func (ds *DeviceState) GetSelectedCol() int {
	return ds.selectedCol
}

// ClassifyOperation determines what operation the current configuration represents
func (ds *DeviceState) ClassifyOperation() string {
	// Check if any column has write voltage (above write range minimum)
	hasWriteVoltage := false
	hasReadVoltage := false
	for _, v := range ds.dacVoltages {
		if v >= ds.writeRange.Min {
			hasWriteVoltage = true
		}
		if v > 0.01 && v <= ds.readRange.Max {
			hasReadVoltage = true
		}
	}

	activeRowCount := 0
	for _, active := range ds.activeRows {
		if active {
			activeRowCount++
		}
	}

	// Classify based on WL mode and voltage levels
	switch {
	case activeRowCount == 1 && hasWriteVoltage:
		return "WRITE"
	case activeRowCount == 1 && hasReadVoltage:
		return "READ"
	case activeRowCount > 1 && !hasWriteVoltage:
		return "COMPUTE (MVM)"
	case activeRowCount > 1 && hasWriteVoltage:
		return "BULK WRITE (CAUTION)"
	default:
		return "IDLE"
	}
}

// Resize updates the device state dimensions
func (ds *DeviceState) Resize(rows, cols int) {
	if rows != ds.rows {
		ds.rows = rows
		ds.activeRows = make([]bool, rows)
		ds.rowCurrents = make([]float64, rows)
		ds.rowVoltages = make([]float64, rows)
		ds.rowLevels = make([]int, rows)
		ds.saturated = make([]bool, rows)
		// Reset to single row 0
		if rows > 0 {
			ds.activeRows[0] = true
		}
	}

	if cols != ds.cols {
		ds.cols = cols
		ds.dacVoltages = make([]float64, cols)
		// Reset to read preset (use material-derived safe read voltage)
		readVoltage := ds.readRange.Max * 0.5 // 50% of max safe read voltage
		for i := range ds.dacVoltages {
			ds.dacVoltages[i] = readVoltage
		}
	}

	// Ensure selected cell is within bounds
	if ds.selectedRow >= ds.rows {
		ds.selectedRow = 0
	}
	if ds.selectedCol >= ds.cols {
		ds.selectedCol = 0
	}
}
