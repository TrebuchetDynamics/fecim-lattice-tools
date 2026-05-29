// Package physics provides shared physics utilities for FeCIM simulations.
// This includes unit formatting, conductance calculations, and physical constants.
package physics

import physicsunits "fecim-lattice-tools/shared/physics/units"

// Electric field unit conversions.
//
// Internally, simulations store electric field in V/m (SI).
// UI/logs often display in MV/cm, common in ferroelectric literature.
//
// 1 MV/cm = 10^6 V/cm = 10^6 V per 10^-2 m = 10^8 V/m.
const VPerMPerMVPerCm = physicsunits.VPerMPerMVPerCm

// VPerMToMVPerCm converts electric field from V/m to MV/cm.
func VPerMToMVPerCm(vPerM float64) float64 { return physicsunits.VPerMToMVPerCm(vPerM) }

// MVPerCmToVPerM converts electric field from MV/cm to V/m.
func MVPerCmToVPerM(mvPerCm float64) float64 { return physicsunits.MVPerCmToVPerM(mvPerCm) }

// FormatEnergy formats energy in Joules with appropriate SI prefix.
func FormatEnergy(joules float64) string { return physicsunits.FormatEnergy(joules) }

// FormatEnergyMJ formats energy given in millijoules with appropriate SI prefix.
func FormatEnergyMJ(mj float64) string { return physicsunits.FormatEnergyMJ(mj) }

// FormatEnergyUJ formats energy given in microjoules with appropriate SI prefix.
func FormatEnergyUJ(uj float64) string { return physicsunits.FormatEnergyUJ(uj) }

// FormatConductance formats conductance in Siemens with appropriate SI prefix.
func FormatConductance(siemens float64) string { return physicsunits.FormatConductance(siemens) }

// FormatCurrent formats current in Amperes with appropriate SI prefix.
func FormatCurrent(amperes float64) string { return physicsunits.FormatCurrent(amperes) }

// FormatVoltage formats voltage in Volts with appropriate SI prefix.
func FormatVoltage(volts float64) string { return physicsunits.FormatVoltage(volts) }

// FormatTime formats time in seconds with appropriate SI prefix.
func FormatTime(seconds float64) string { return physicsunits.FormatTime(seconds) }

// FormatFrequency formats frequency in Hertz with appropriate SI prefix.
func FormatFrequency(hz float64) string { return physicsunits.FormatFrequency(hz) }

// FormatResistance formats resistance in Ohms with appropriate SI prefix.
func FormatResistance(ohms float64) string { return physicsunits.FormatResistance(ohms) }

// FormatCapacitance formats capacitance in Farads with appropriate SI prefix.
func FormatCapacitance(farads float64) string { return physicsunits.FormatCapacitance(farads) }

// FormatPower formats power in Watts with appropriate SI prefix.
func FormatPower(watts float64) string { return physicsunits.FormatPower(watts) }

// FormatCharge formats electric charge in Coulombs with appropriate SI prefix.
func FormatCharge(coulombs float64) string { return physicsunits.FormatCharge(coulombs) }

// FormatPolarization formats polarization in C/m² as µC/cm².
func FormatPolarization(cm2 float64) string { return physicsunits.FormatPolarization(cm2) }

// FormatElectricField formats electric field in V/m as MV/cm or kV/cm.
func FormatElectricField(vm float64) string { return physicsunits.FormatElectricField(vm) }
