package gui

import (
	"context"
	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

func (ds *DeviceState) numericalSettingsLocked() arraysim.DeviceSettings {
	return arraysim.DeviceSettings{Material: ds.material, Geometry: ds.cellGeometry, Wire: ds.wireParams, Coupling: ds.couplingMode, Passive: ds.isPassive, Read: ds.opMode == OpModeRead, Write: ds.dacRangeMode == DACRangeWrite, ActiveRows: ds.activeRows, WL: ds.wlVoltages, BL: ds.dacVoltages, TIA: ds.tia, ADC: ds.adc, DAC: ds.dac, DACNonlinearity: ds.enableDACNonlinearity, Temperature: ds.peripheralTemperature, Corner: ds.processCorner, ReadMax: ds.readRange.Max, SelectorEnabled: ds.selectorEnabled, SelectorRon: ds.selectorRon, SelectorLeakage: ds.selectorLeakageConductance}
}
func (ds *DeviceState) Compute(weights [][]int, quantLevels int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	s := ds.numericalSettingsLocked()
	result, err := arraysim.EvaluateDevice(context.Background(), s, weights, quantLevels)
	if err != nil { // Preserve the legacy GUI's ideal fallback. Web ports surface errors.
		s.Coupling = arraysim.CouplingIdeal
		result, _ = arraysim.EvaluateDevice(context.Background(), s, weights, quantLevels)
	}
	ds.coupledCellVoltages, ds.coupledCellCurrents = result.CellVoltages, result.CellCurrents
	if s.Coupling == arraysim.CouplingIdeal {
		ds.coupledCellVoltages, ds.coupledCellCurrents = nil, nil
	}
	ds.rowCurrents, ds.rowVoltages, ds.rowLevels, ds.saturated = result.RowCurrents, result.RowVoltages, result.RowLevels, result.Saturated
	ds.LogCompute(weights, quantLevels)
}
func (ds *DeviceState) programLevelFromCoupledVoltage(level int, voltage, width float64, levels int) int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return arraysim.ProgramDeviceLevel(ds.numericalSettingsLocked(), level, voltage, width, levels)
}

func (ds *DeviceState) convertSenseLocked(current, gain, vref float64) (float64, int, bool) {
	return arraysim.DeviceSense(ds.numericalSettingsLocked(), current, gain, vref)
}
