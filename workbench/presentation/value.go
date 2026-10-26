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
	case "µC/cm²":
		spokenUnit = "microcoulombs per square centimeter"
	case "A":
		spokenUnit = "amperes"
	case "Ω":
		spokenUnit = "ohms"
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
