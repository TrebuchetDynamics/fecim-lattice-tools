// Package arraysim shares the native GUI numerical device path with browser ports.
// These calculations were extracted without changing native arithmetic.
package arraysim

import (
	"context"
	"fecim-lattice-tools/shared/peripherals"
	sharedphysics "fecim-lattice-tools/shared/physics"
	"math"
)

// DeviceSettings contains explicit SI inputs; no filesystem, GUI or logging authority.
type DeviceSettings struct {
	Material                     *sharedphysics.HZOMaterial
	Geometry                     CellGeometry
	Wire                         WireParams
	Coupling                     CouplingMode
	Passive, Read, Write         bool
	ActiveRows                   []bool
	WL, BL                       []float64
	TIA                          *peripherals.TIA
	ADC                          *peripherals.ADC
	DAC                          *peripherals.DAC
	DACNonlinearity              bool
	Temperature                  float64
	Corner                       peripherals.ProcessCorner
	ReadMax                      float64
	SelectorEnabled              bool
	SelectorRon, SelectorLeakage float64
	ReadMask, WriteMask          [][]bool
	MaxIterations                int
}

// DeviceResult retains signed cell currents; coupled sense rows retain native magnitude polarity.
type DeviceResult struct {
	CellVoltages, CellCurrents [][]float64
	RowCurrents, RowVoltages   []float64 // native row currents in microamperes
	RowLevels                  []int
	Saturated                  []bool
}
type deviceCalculation struct {
	ctx                                      context.Context
	err                                      error
	rows, cols                               int
	material                                 *sharedphysics.HZOMaterial
	cellGeometry                             CellGeometry
	wireParams                               WireParams
	couplingMode                             CouplingMode
	arrayEngine                              Engine
	isPassive, read, write                   bool
	activeRows                               []bool
	wlVoltages, dacVoltages                  []float64
	tia                                      *peripherals.TIA
	adc                                      *peripherals.ADC
	dac                                      *peripherals.DAC
	enableDACNonlinearity                    bool
	peripheralTemperature                    float64
	processCorner                            peripherals.ProcessCorner
	readRange                                struct{ Max float64 }
	selectorEnabled                          bool
	selectorRon, selectorLeakageConductance  float64
	coupledCellVoltages, coupledCellCurrents [][]float64
	rowCurrents, rowVoltages                 []float64
	rowLevels                                []int
	saturated                                []bool
	readMask, writeMask                      [][]bool
}

func calculation(ctx context.Context, s DeviceSettings, rows, cols int) *deviceCalculation {
	d := &deviceCalculation{ctx: ctx, rows: rows, cols: cols, material: s.Material, cellGeometry: s.Geometry, wireParams: s.Wire, couplingMode: s.Coupling, isPassive: s.Passive, read: s.Read, write: s.Write, activeRows: s.ActiveRows, wlVoltages: s.WL, dacVoltages: s.BL, tia: s.TIA, adc: s.ADC, dac: s.DAC, enableDACNonlinearity: s.DACNonlinearity, peripheralTemperature: s.Temperature, processCorner: s.Corner, selectorEnabled: s.SelectorEnabled, selectorRon: s.SelectorRon, selectorLeakageConductance: s.SelectorLeakage, readMask: s.ReadMask, writeMask: s.WriteMask, rowCurrents: make([]float64, rows), rowVoltages: make([]float64, rows), rowLevels: make([]int, rows), saturated: make([]bool, rows)}
	if d.material == nil {
		d.material = sharedphysics.FeCIMMaterial()
	}
	d.readRange.Max = s.ReadMax
	switch s.Coupling {
	case CouplingTierA:
		d.arrayEngine = NewTierASolver()
	case CouplingTierB:
		b := NewTierBSolver()
		if s.MaxIterations > 0 {
			b.MaxIter = s.MaxIterations
		}
		d.arrayEngine = b
	}
	return d
}

func (ds *deviceCalculation) maskSelected(r, c int) bool {
	mask := ds.readMask
	if ds.write {
		mask = ds.writeMask
	}
	return mask == nil || maskAt(mask, r, c)
}

// EvaluateDevice is the shared native read/compute path. Errors never silently downgrade coupling.
func EvaluateDevice(ctx context.Context, s DeviceSettings, weights [][]int, levels int) (DeviceResult, error) {
	if err := ctx.Err(); err != nil {
		return DeviceResult{}, err
	}
	d := calculation(ctx, s, len(weights), len(s.BL))
	if s.Coupling == CouplingIdeal {
		d.computeIdealLocked(weights, levels)
		d.coupledCellVoltages = make([][]float64, d.rows)
		d.coupledCellCurrents = make([][]float64, d.rows)
		for r := 0; r < d.rows; r++ {
			d.coupledCellVoltages[r] = make([]float64, d.cols)
			d.coupledCellCurrents[r] = make([]float64, d.cols)
			for c := 0; c < d.cols; c++ {
				v := d.effectiveCellVoltageLocked(r, c)
				d.coupledCellVoltages[r][c] = v
				if d.activeRows[r] && d.maskSelected(r, c) {
					d.coupledCellCurrents[r][c] = d.levelToConductance(weights[r][c], levels) * v
				}
			}
		}
	} else if !d.computeWithArraysimLocked(weights, levels) {
		return DeviceResult{}, d.err
	}
	if err := ctx.Err(); err != nil {
		return DeviceResult{}, err
	}
	return DeviceResult{d.coupledCellVoltages, d.coupledCellCurrents, d.rowCurrents, d.rowVoltages, d.rowLevels, d.saturated}, nil
}
func DeviceConductance(s DeviceSettings, level, levels int) float64 {
	return calculation(context.Background(), s, 0, 0).levelToConductance(level, levels)
}
func DeviceConductanceBounds(s DeviceSettings) (float64, float64) {
	return calculation(context.Background(), s, 0, 0).conductanceBounds()
}
func DeviceConductanceLevel(s DeviceSettings, g float64, levels int) int {
	return calculation(context.Background(), s, 0, 0).conductanceToLevel(g, levels)
}
func ProgramDeviceLevel(s DeviceSettings, level int, voltage, width float64, levels int) int {
	return calculation(context.Background(), s, 0, 0).programLevelFromCoupledVoltage(level, voltage, width, levels)
}

func (ds *deviceCalculation) conductanceBounds() (float64, float64) {
	mat := ds.material
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	gmin := mat.Gmin
	gmax := mat.Gmax
	if gmin == 0 && gmax == 0 {
		gmin = 1e-6   // Match material.DiscreteLevel fallback
		gmax = 100e-6 // Match material.DiscreteLevel fallback
	}
	scale := ds.cellGeometry.Film.ConductanceScale(sharedphysics.GeometryFromMaterial(mat))
	return gmin * scale, gmax * scale
}

// levelToConductance maps a discrete level to conductance using the active material.
func (ds *deviceCalculation) levelToConductance(level, levels int) float64 {
	mat := ds.material
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	levels = ds.resolveConductanceLevels(levels)
	baseG := mat.DiscreteLevel(level, levels)
	referenceGeom := sharedphysics.GeometryFromMaterial(mat)
	return baseG * ds.cellGeometry.Film.ConductanceScale(referenceGeom)
}

// resolveConductanceLevels picks quantization levels used for level->conductance mapping.
// READ defaults to material-native levels so material selection propagates into READ current.
func (ds *deviceCalculation) resolveConductanceLevels(quantLevels int) int {
	if ds.read && ds.material != nil {
		if materialLevels := ds.material.GetNumLevels(); materialLevels > 0 {
			return materialLevels
		}
	}
	if quantLevels > 0 {
		return quantLevels
	}
	if ds.material != nil {
		if materialLevels := ds.material.GetNumLevels(); materialLevels > 0 {
			return materialLevels
		}
	}
	return 30
}

// conductanceToLevel maps conductance to a discrete level using material bounds.
func (ds *deviceCalculation) conductanceToLevel(gPhys float64, levels int) int {
	if levels <= 0 {
		levels = ds.material.GetNumLevels()
	}
	if levels <= 0 {
		levels = sharedphysics.DefaultLevels
	}
	gmin, gmax := ds.conductanceBounds()
	if gmax <= gmin {
		return 0
	}
	gNorm := (gPhys - gmin) / (gmax - gmin)
	return sharedphysics.GetLevel(gNorm, levels)
}

func (ds *deviceCalculation) computeIdealLocked(weights [][]int, quantLevels int) {
	// During programming, the sense chain (TIA/ADC) is disconnected.
	// Don't compute row currents using write-level WL/BL voltages.
	if ds.write {
		for r := 0; r < ds.rows; r++ {
			ds.rowCurrents[r] = 0
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
		}
		return
	}

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
			if !ds.maskSelected(r, c) {
				continue
			}
			voltage := ds.effectiveCellVoltageLocked(r, c)

			// Apply DAC nonlinearity if enabled (read/compute path only).
			if ds.enableDACNonlinearity && ds.dac != nil && !ds.write {
				voltage = ds.applyDACNonlinearityLocked(voltage)
			} else if ds.isPassive {
				// For passive mode, use WL/BL effective voltage even without DAC nonlinearity.
				voltage = ds.wlVoltages[r] - ds.dacVoltages[c]
			}

			if math.Abs(voltage) < 1e-9 {
				continue
			}

			// Get cell conductance from weight using material physics model
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}

			// Use shared level->conductance mapping so ideal and coupled paths
			// respect the same material + geometry scaling dependencies.
			conductanceS := ds.levelToConductance(level, quantLevels)

			// Convert to uS for current calculation.
			conductanceUS := conductanceS * 1e6
			current := conductanceUS * voltage // I = G * V (signed uA since G is in uS)
			totalCurrent += current
		}

		ds.rowCurrents[r] = totalCurrent

		if ds.tia == nil {
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		currentA := totalCurrent * 1e-6 // uA to A

		// Count active columns (columns with |effective voltage| > 0.01V)
		activeColCount := 0
		for c := 0; c < ds.cols; c++ {
			v := ds.effectiveCellVoltageLocked(r, c)
			if math.Abs(v) > 1e-9 {
				activeColCount++
			}
		}

		effectiveGain := ds.tia.Gain
		if activeColCount > 1 {
			// MVM mode: scale gain based on number of active columns to reduce saturation.
			effectiveGain = ds.tia.Gain / float64(activeColCount)
		}

		vref := 0.0
		if activeColCount <= 1 {
			// Single-cell read path keeps the TIA output offset.
			vref = ds.tia.OutputOffset
		}

		ds.rowVoltages[r], ds.rowLevels[r], ds.saturated[r] = ds.convertSenseLocked(currentA, effectiveGain, vref)
	}
}

func (ds *deviceCalculation) convertSenseLocked(rowCurrentA, gainOhm, vrefV float64) (float64, int, bool) {
	if ds.tia == nil {
		return 0, 0, false
	}

	if ds.adc == nil {
		vout := vrefV + rowCurrentA*gainOhm
		if vout < 0 {
			vout = 0
		}
		if vout > ds.tia.MaxOutputVoltage {
			vout = ds.tia.MaxOutputVoltage
		}
		return vout, 0, false
	}

	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   gainOhm,
			Vref: vrefV,
			Vmin: 0,
			Vmax: ds.tia.MaxOutputVoltage,
		},
		ADC: ADCConfig{
			Bits: ds.adc.Bits,
			Vmin: ds.adc.VrefLow,
			Vmax: ds.adc.VrefHigh,
		},
	}
	res := sense.ConvertCurrent(rowCurrentA)
	adcMaxLevel := (1 << ds.adc.Bits) - 1
	saturated := res.TIASaturated || res.ADCSaturated || res.Code >= adcMaxLevel
	return res.Vout, res.Code, saturated
}

func (ds *deviceCalculation) computeWithArraysimLocked(weights [][]int, quantLevels int) bool {
	if ds.arrayEngine == nil {
		return false
	}

	rows := ds.rows
	cols := ds.cols
	if rows == 0 || cols == 0 {
		return true
	}

	conductance := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		conductance[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			conductance[r][c] = ds.levelToConductance(level, quantLevels)
		}
	}

	blApplied := make([]float64, cols)
	for c := 0; c < cols; c++ {
		blApplied[c] = ds.applyDACNonlinearityLocked(ds.dacVoltages[c])
	}

	wlApplied := make([]float64, rows)
	if ds.isPassive {
		copy(wlApplied, ds.wlVoltages)
	}

	blSolve := blApplied
	if !ds.isPassive {
		blSolve = make([]float64, cols)
		for c := 0; c < cols; c++ {
			blSolve[c] = -blApplied[c]
		}
	}

	params := SolveParams{
		WLVoltages:      wlApplied,
		BLVoltages:      blSolve,
		Conductance:     conductance,
		ActiveRows:      ds.activeRows,
		SelectorEnabled: ds.selectorEnabled,
		SelectorRon:     ds.selectorRon,
		Selector: SelectorDeviceParams{
			Enabled:        ds.selectorLeakageConductance > 0,
			OnConductance:  math.Inf(1),
			OffConductance: ds.selectorLeakageConductance,
		},
		ReadMask: ds.readMask, WriteMask: ds.writeMask,
		Geometry: ds.cellGeometry,
		Wire:     ds.wireParams,
	}
	if ds.write {
		params.SelectorMode = SelectorWrite
	} else {
		params.SelectorMode = SelectorRead
	}
	result, err := SolveContext(ds.ctx, ds.arrayEngine, params)
	if err != nil {
		ds.err = err
		return false
	}

	ds.coupledCellVoltages = result.CellVoltages
	ds.coupledCellCurrents = result.CellCurrents

	for r := 0; r < rows; r++ {
		if !ds.activeRows[r] {
			ds.rowCurrents[r] = 0
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		rowCurrentA := 0.0
		if r < len(result.RowCurrents) {
			rowCurrentA = result.RowCurrents[r]
		} else if r < len(result.CellCurrents) {
			for c := 0; c < len(result.CellCurrents[r]); c++ {
				rowCurrentA += result.CellCurrents[r][c]
			}
		}
		// arraysim uses solver-current sign conventions; sense chain expects read-magnitude polarity.
		rowCurrentA = math.Abs(rowCurrentA)

		ds.rowCurrents[r] = rowCurrentA * 1e6

		if ds.tia == nil {
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		activeColCount := 0
		expectedMaxCurrentA := 0.0
		for c := 0; c < cols; c++ {
			vcell := blApplied[c]
			if ds.isPassive {
				vcell = wlApplied[r] - blApplied[c]
			}
			if math.Abs(vcell) > 1e-9 {
				activeColCount++
			}
			expectedMaxCurrentA += conductance[r][c] * math.Abs(vcell)
		}

		effectiveGain := ds.tia.Gain
		if activeColCount > 1 {
			effectiveGain = effectiveGain / float64(activeColCount)
		}
		if expectedMaxCurrentA > 0 {
			targetV := 0.9 * ds.tia.MaxOutputVoltage
			if expectedMaxCurrentA*effectiveGain > targetV {
				effectiveGain = targetV / expectedMaxCurrentA
			}
		}

		vref := 0.0
		if activeColCount <= 1 {
			vref = ds.tia.OutputOffset
		}

		ds.rowVoltages[r], ds.rowLevels[r], ds.saturated[r] = ds.convertSenseLocked(rowCurrentA, effectiveGain, vref)
	}

	return true
}

// programLevelFromCoupledVoltage advances one cell's programmed level using the
// *actual* cell voltage seen after DAC + array coupling (IR drop / half-select).
// This is used by the level-engine ISPP write path to avoid ideal-voltage updates.
func (ds *deviceCalculation) programLevelFromCoupledVoltage(currentLevel int, effectiveV float64, pulseWidth float64, quantLevels int) int {

	levels := ds.resolveConductanceLevels(quantLevels)
	if levels < 2 {
		levels = 2
	}
	if currentLevel < 0 {
		currentLevel = 0
	}
	if currentLevel >= levels {
		currentLevel = levels - 1
	}
	if math.Abs(effectiveV) < 1e-12 {
		return currentLevel
	}

	mat := ds.material
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	if mat.Ps == 0 {
		return currentLevel
	}

	if pulseWidth <= 0 {
		pulseWidth = mat.Tau
	}
	if pulseWidth <= 0 {
		pulseWidth = float64(200) * 1e-9
	}

	gmin, gmax := ds.conductanceBounds()
	currentG := ds.levelToConductance(currentLevel, levels)
	currentP := sharedphysics.ConductanceToPolarizationWithParams(currentG, mat.Ps, gmin, gmax, sharedphysics.ParseConductanceModel(mat.ConductanceModel), mat.KvT, mat.VGSReadV, mat.VT0V)

	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false
	if !solver.UseMaterialAlpha {
		solver.UpdateParams()
	}
	solver.SetState(currentP)

	geom := ds.cellGeometry.WithDefaults()
	eField := geom.Film.ElectricField(effectiveV)
	solver.Step(eField, pulseWidth)

	newP := solver.GetState()
	newG := sharedphysics.PolarizationToConductanceWithParams(newP, mat.Ps, gmin, gmax, sharedphysics.ParseConductanceModel(mat.ConductanceModel), mat.KvT, mat.VGSReadV, mat.VT0V)
	newLevel := ds.conductanceToLevel(newG, levels)
	if newLevel < 0 {
		newLevel = 0
	}
	if newLevel >= levels {
		newLevel = levels - 1
	}
	return newLevel
}

func (ds *deviceCalculation) effectiveCellVoltageLocked(row, col int) float64 {
	if row < 0 || row >= ds.rows || col < 0 || col >= ds.cols {
		return 0
	}
	if ds.couplingMode != CouplingIdeal && ds.coupledCellVoltages != nil {
		if row < len(ds.coupledCellVoltages) && col < len(ds.coupledCellVoltages[row]) {
			v := ds.coupledCellVoltages[row][col]
			if math.Abs(v) < 1e-12 {
				return 0
			}
			return v
		}
	}
	bl := ds.dacVoltages[col]
	if ds.isPassive {
		wl := ds.wlVoltages[row]
		v := wl - bl
		if math.Abs(v) < 1e-12 {
			return 0
		}
		return v
	}
	if math.Abs(bl) < 1e-12 {
		return 0
	}
	return bl
}

// applyDACNonlinearityLocked applies DAC nonlinearity for read/compute path voltages.
// Caller must hold ds.mu.
func (ds *deviceCalculation) applyDACNonlinearityLocked(voltage float64) float64 {
	if !ds.enableDACNonlinearity || ds.dac == nil || ds.write {
		return voltage
	}
	voltageMag := math.Abs(voltage)
	normalized := 0.0
	if ds.readRange.Max > 0 {
		normalized = voltageMag / ds.readRange.Max
	}
	if normalized > 1.0 {
		normalized = 1.0
	}
	if normalized < 0 {
		normalized = 0
	}
	level := int(normalized * float64(ds.dac.Levels()-1))
	applied := ds.dac.ConvertWithCondition(level, ds.peripheralTemperature, ds.processCorner)
	return math.Copysign(applied, voltage)
}
