package physics

// CellSwitchingEnergy returns E_switch = C_fe * V² for one polarization reversal.
func CellSwitchingEnergy(cFe, voltage float64) float64 {
	if cFe <= 0 {
		return 0
	}
	return cFe * voltage * voltage
}

// CellDynamicPower returns P_dyn = C_eff * V² * f for a cell at given frequency.
func CellDynamicPower(cEff, voltage, frequency float64) float64 {
	if cEff <= 0 || frequency <= 0 {
		return 0
	}
	return cEff * voltage * voltage * frequency
}

// CellLeakagePower returns P_leak = V * I_off for selector leakage.
func CellLeakagePower(voltage, iOff float64) float64 {
	if iOff <= 0 {
		return 0
	}
	return voltage * iOff
}

// ArrayPowerParams defines array-level parameters for dynamic and static power estimation.
type ArrayPowerParams struct {
	Rows           int
	Cols           int
	ActiveFraction float64 // Fraction of cells switching per cycle [0,1]

	CellCapacitance float64 // F
	WriteVoltage    float64 // V (used for dynamic/switching)
	ReadVoltage     float64 // V (used for leakage)
	Frequency       float64 // Hz
	SelectorIoff    float64 // A
	PeripheralPower float64 // W (DAC + TIA + ADC + control)

	// Optional area term for power density reporting.
	// If zero or negative, PowerDensity is reported as 0.
	ArrayAreaMM2 float64 // mm²
}

// ArrayPowerResult captures aggregate array power metrics.
type ArrayPowerResult struct {
	DynamicPower    float64 // W
	LeakagePower    float64 // W
	PeripheralPower float64 // W
	TotalPower      float64 // W

	EnergyPerOp float64 // J (array-level energy per operation/cycle)
	PowerDensity float64 // W/mm²
}

// ArrayPower computes total array power for an array:
// P_total = N_active * P_dyn_cell + N_total * P_leak_cell + P_peripheral.
func ArrayPower(params ArrayPowerParams) ArrayPowerResult {
	rows := params.Rows
	if rows < 0 {
		rows = 0
	}
	cols := params.Cols
	if cols < 0 {
		cols = 0
	}
	nTotal := float64(rows * cols)

	activeFraction := params.ActiveFraction
	if activeFraction < 0 {
		activeFraction = 0
	}
	if activeFraction > 1 {
		activeFraction = 1
	}
	nActive := nTotal * activeFraction

	pDynCell := CellDynamicPower(params.CellCapacitance, params.WriteVoltage, params.Frequency)
	pLeakCell := CellLeakagePower(params.ReadVoltage, params.SelectorIoff)

	dynamic := nActive * pDynCell
	leakage := nTotal * pLeakCell
	peripheral := params.PeripheralPower
	if peripheral < 0 {
		peripheral = 0
	}
	total := dynamic + leakage + peripheral

	energyPerOp := 0.0
	if params.Frequency > 0 {
		energyPerOp = total / params.Frequency
	}

	powerDensity := 0.0
	if params.ArrayAreaMM2 > 0 {
		powerDensity = total / params.ArrayAreaMM2
	}

	return ArrayPowerResult{
		DynamicPower:    dynamic,
		LeakagePower:    leakage,
		PeripheralPower: peripheral,
		TotalPower:      total,
		EnergyPerOp:     energyPerOp,
		PowerDensity:    powerDensity,
	}
}
