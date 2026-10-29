// Package presentation derives display and speech text without changing the
// canonical scientific records used for sorting, analysis and export.
package presentation

import (
	"math"
	"strconv"
	"strings"

	"fecim-lattice-tools/workbench/experiment"
)

// Value is presentation only. Consumers must retain the metric's identity,
// model, assumptions and evidence alongside it. Availability is "available" or
// "unavailable"; Reason explains unavailable values.
type Value struct {
	Display      string
	Speakable    string
	Availability string
	Reason       string
}

// FormatMetric uses Go's shortest round-trippable decimal representation without
// rounding or unit conversion. A nonempty unavailableReason takes precedence over
// the metric, including its numeric value. Missing metrics must be declared with
// that argument: a finite zero in a supported unit is a real available value.
// Speakable is deterministic text, not a claim of screen-reader qualification.
func FormatMetric(metric experiment.Metric, unavailableReason string) Value {
	if unavailableReason != "" {
		return unavailable(unavailableReason)
	}
	if math.IsNaN(metric.Value) || math.IsInf(metric.Value, 0) {
		return unavailable("Metric value is not finite")
	}

	displayUnit := metric.Unit
	var spokenUnit string
	switch metric.Unit {
	case "um2":
		displayUnit, spokenUnit = "µm²", "square micrometers"
	case "pJ":
		spokenUnit = "picojoules"
	case "ns":
		spokenUnit = "nanoseconds"
	case "dB":
		spokenUnit = "decibels"
	case "MV/cm":
		spokenUnit = "megavolts per centimeter"
	case "µC/cm²":
		spokenUnit = "microcoulombs per square centimeter"
	case "A":
		spokenUnit = "amperes"
	case "s":
		spokenUnit = "seconds"
	case "C":
		spokenUnit = "coulombs"
	case "C/V²":
		spokenUnit = "coulombs per square volt"
	case "m²":
		spokenUnit = "square meters"
	case "m":
		spokenUnit = "meters"
	case "J":
		spokenUnit = "joules"
	case "V":
		spokenUnit = "volts"
	case "S":
		spokenUnit = "siemens"
	case "Ω·m":
		spokenUnit = "ohm meters"
	case "V·m/C":
		spokenUnit = "volt meters per coulomb"
	case "J·m⁵/C⁴":
		spokenUnit = "joule meters to the fifth power per coulombs to the fourth power"
	case "J·m⁹/C⁶":
		spokenUnit = "joule meters to the ninth power per coulombs to the sixth power"
	case "V/(m·K)":
		spokenUnit = "volts per meter per kelvin"
	case "C/(m²·K)":
		spokenUnit = "coulombs per square meter per kelvin"
	case "m⁴/C²":
		spokenUnit = "meters to the fourth power per square coulomb"
	case "cycles":
		spokenUnit = "cycles"
	case "Hz":
		spokenUnit = "hertz"
	case "K":
		spokenUnit = "kelvin"
	case "Pa":
		spokenUnit = "pascals"
	case "ratio":
		spokenUnit = "ratio"
	case "Ω":
		spokenUnit = "ohms"
	case "%":
		spokenUnit = "percent"
	case "bits":
		spokenUnit = "bits"
	case "bits/cell":
		spokenUnit = "bits per cell"
	case "cells":
		spokenUnit = "cells"
	case "levels":
		spokenUnit = "levels"
	case "samples":
		spokenUnit = "samples"
	case "count":
		spokenUnit = "count"
	case "operations":
		spokenUnit = "operations"
	case "MACs":
		spokenUnit = "multiply accumulate operations"
	case "mV":
		spokenUnit = "millivolts"
	case "µA":
		spokenUnit = "microamperes"
	case "mA":
		spokenUnit = "milliamperes"
	case "pA":
		spokenUnit = "picoamperes"
	case "µS":
		spokenUnit = "microsiemens"
	case "nS":
		spokenUnit = "nanosiemens"
	case "kΩ":
		spokenUnit = "kiloohms"
	case "F":
		spokenUnit = "farads"
	case "pF":
		spokenUnit = "picofarads"
	case "fF":
		spokenUnit = "femtofarads"
	case "W":
		spokenUnit = "watts"
	case "mW":
		spokenUnit = "milliwatts"
	case "µW":
		spokenUnit = "microwatts"
	case "nW":
		spokenUnit = "nanowatts"
	case "fJ":
		spokenUnit = "femtojoules"
	case "nJ":
		spokenUnit = "nanojoules"
	case "ms":
		spokenUnit = "milliseconds"
	case "µs":
		spokenUnit = "microseconds"
	case "ps":
		spokenUnit = "picoseconds"
	case "nm":
		spokenUnit = "nanometers"
	case "µm":
		spokenUnit = "micrometers"
	case "TOPS":
		spokenUnit = "trillion operations per second"
	case "GOPS":
		spokenUnit = "billion operations per second"
	case "TOPS/W":
		spokenUnit = "trillion operations per second per watt"
	case "GOPS/W":
		spokenUnit = "billion operations per second per watt"
	case "GB/s":
		spokenUnit = "gigabytes per second"
	case "kWh":
		spokenUnit = "kilowatt hours"
	case "mm2":
		displayUnit, spokenUnit = "mm²", "square millimeters"
	case "":
		return unavailable("Metric unit is missing")
	default:
		return unavailable("Unsupported metric unit " + strconv.Quote(metric.Unit))
	}

	number := strconv.FormatFloat(metric.Value, 'g', -1, 64)
	mantissa, exponent, scientific := strings.Cut(number, "e")
	spoken := speakMantissa(mantissa)
	if scientific {
		sign := "plus"
		if exponent[0] == '-' {
			sign = "minus"
		}
		// FormatFloat always supplies a signed exponent. Drop padding, not scale.
		spoken += " times 10 to the " + sign + " " + strings.TrimLeft(exponent[1:], "0")
	}
	return Value{
		Display:      number + " " + displayUnit,
		Speakable:    spoken + " " + spokenUnit,
		Availability: "available",
	}
}

// FormatElectricField converts SI V/m for presentation only. The caller keeps
// the original SI value; an explicit unavailable reason takes precedence.
func FormatElectricField(si float64, unavailableReason string) Value {
	return FormatMetric(experiment.Metric{Value: si * 1e-8, Unit: "MV/cm"}, unavailableReason)
}

// FormatPolarization converts SI C/m² to microcoulombs per square centimeter.
func FormatPolarization(si float64, unavailableReason string) Value {
	return FormatMetric(experiment.Metric{Value: si * 100, Unit: "µC/cm²"}, unavailableReason)
}

func unavailable(reason string) Value {
	return Value{
		Display:      "Not available",
		Speakable:    "Not available: " + reason,
		Availability: "unavailable",
		Reason:       reason,
	}
}

func speakMantissa(number string) string {
	spoken := ""
	if strings.HasPrefix(number, "-") {
		spoken = "minus "
		number = strings.TrimPrefix(number, "-")
	}
	integer, fraction, decimal := strings.Cut(number, ".")
	spoken += integer
	if decimal {
		// Separate fractional digits so precision, including zeros, stays explicit.
		spoken += " point " + strings.Join(strings.Split(fraction, ""), " ")
	}
	return spoken
}
