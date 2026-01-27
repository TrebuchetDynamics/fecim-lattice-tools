// Package physics provides shared physics utilities for FeCIM simulations.
// This includes unit formatting, conductance calculations, and physical constants.
package physics

import "fmt"

// FormatEnergy formats energy in Joules with appropriate SI prefix.
// Automatically scales from fJ (femtojoules) to J (joules).
//
// Example:
//
//	FormatEnergy(1.5e-15) // "1.50 fJ"
//	FormatEnergy(2.3e-12) // "2.30 pJ"
//	FormatEnergy(4.5e-9)  // "4.50 nJ"
//	FormatEnergy(6.7e-6)  // "6.70 µJ"
//	FormatEnergy(8.9e-3)  // "8.90 mJ"
//	FormatEnergy(1.2)     // "1.20 J"
func FormatEnergy(joules float64) string {
	switch {
	case joules <= 0:
		return "0 J"
	case joules < 1e-12:
		return fmt.Sprintf("%.2f fJ", joules*1e15)
	case joules < 1e-9:
		return fmt.Sprintf("%.2f pJ", joules*1e12)
	case joules < 1e-6:
		return fmt.Sprintf("%.2f nJ", joules*1e9)
	case joules < 1e-3:
		return fmt.Sprintf("%.2f µJ", joules*1e6)
	case joules < 1:
		return fmt.Sprintf("%.2f mJ", joules*1e3)
	default:
		return fmt.Sprintf("%.2f J", joules)
	}
}

// FormatEnergyMJ formats energy given in millijoules with appropriate SI prefix.
// Convenience wrapper for data already in mJ.
func FormatEnergyMJ(mj float64) string {
	return FormatEnergy(mj * 1e-3)
}

// FormatEnergyUJ formats energy given in microjoules with appropriate SI prefix.
// Convenience wrapper for data already in µJ.
func FormatEnergyUJ(uj float64) string {
	return FormatEnergy(uj * 1e-6)
}

// FormatConductance formats conductance in Siemens with appropriate SI prefix.
// Automatically scales from nS (nanosiemens) to S (siemens).
//
// Example:
//
//	FormatConductance(1e-9)  // "1.00 nS"
//	FormatConductance(50e-6) // "50.00 µS"
//	FormatConductance(1e-3)  // "1.00 mS"
func FormatConductance(siemens float64) string {
	switch {
	case siemens <= 0:
		return "0 S"
	case siemens < 1e-6:
		return fmt.Sprintf("%.2f nS", siemens*1e9)
	case siemens < 1e-3:
		return fmt.Sprintf("%.2f µS", siemens*1e6)
	case siemens < 1:
		return fmt.Sprintf("%.2f mS", siemens*1e3)
	default:
		return fmt.Sprintf("%.2f S", siemens)
	}
}

// FormatCurrent formats current in Amperes with appropriate SI prefix.
// Automatically scales from pA (picoamperes) to A (amperes).
//
// Example:
//
//	FormatCurrent(1e-12) // "1.00 pA"
//	FormatCurrent(50e-9) // "50.00 nA"
//	FormatCurrent(1e-6)  // "1.00 µA"
//	FormatCurrent(1e-3)  // "1.00 mA"
func FormatCurrent(amperes float64) string {
	switch {
	case amperes <= 0:
		return "0 A"
	case amperes < 1e-9:
		return fmt.Sprintf("%.2f pA", amperes*1e12)
	case amperes < 1e-6:
		return fmt.Sprintf("%.2f nA", amperes*1e9)
	case amperes < 1e-3:
		return fmt.Sprintf("%.2f µA", amperes*1e6)
	case amperes < 1:
		return fmt.Sprintf("%.2f mA", amperes*1e3)
	default:
		return fmt.Sprintf("%.2f A", amperes)
	}
}

// FormatVoltage formats voltage in Volts with appropriate SI prefix.
// Automatically scales from µV (microvolts) to V (volts).
//
// Example:
//
//	FormatVoltage(1e-6) // "1.00 µV"
//	FormatVoltage(1e-3) // "1.00 mV"
//	FormatVoltage(1.5)  // "1.50 V"
func FormatVoltage(volts float64) string {
	switch {
	case volts <= 0:
		return "0 V"
	case volts < 1e-3:
		return fmt.Sprintf("%.2f µV", volts*1e6)
	case volts < 1:
		return fmt.Sprintf("%.2f mV", volts*1e3)
	default:
		return fmt.Sprintf("%.2f V", volts)
	}
}

// FormatTime formats time in seconds with appropriate SI prefix.
// Automatically scales from ps (picoseconds) to s (seconds).
//
// Example:
//
//	FormatTime(1e-12) // "1.00 ps"
//	FormatTime(1e-9)  // "1.00 ns"
//	FormatTime(1e-6)  // "1.00 µs"
//	FormatTime(1e-3)  // "1.00 ms"
func FormatTime(seconds float64) string {
	switch {
	case seconds <= 0:
		return "0 s"
	case seconds < 1e-9:
		return fmt.Sprintf("%.2f ps", seconds*1e12)
	case seconds < 1e-6:
		return fmt.Sprintf("%.2f ns", seconds*1e9)
	case seconds < 1e-3:
		return fmt.Sprintf("%.2f µs", seconds*1e6)
	case seconds < 1:
		return fmt.Sprintf("%.2f ms", seconds*1e3)
	default:
		return fmt.Sprintf("%.2f s", seconds)
	}
}

// FormatFrequency formats frequency in Hertz with appropriate SI prefix.
// Automatically scales from Hz to GHz.
//
// Example:
//
//	FormatFrequency(1e3) // "1.00 kHz"
//	FormatFrequency(1e6) // "1.00 MHz"
//	FormatFrequency(1e9) // "1.00 GHz"
func FormatFrequency(hz float64) string {
	switch {
	case hz <= 0:
		return "0 Hz"
	case hz < 1e3:
		return fmt.Sprintf("%.2f Hz", hz)
	case hz < 1e6:
		return fmt.Sprintf("%.2f kHz", hz/1e3)
	case hz < 1e9:
		return fmt.Sprintf("%.2f MHz", hz/1e6)
	default:
		return fmt.Sprintf("%.2f GHz", hz/1e9)
	}
}
