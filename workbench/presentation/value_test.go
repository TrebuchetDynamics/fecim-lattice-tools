package presentation_test

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/presentation"
)

// These synthetic fixtures catch lost signs, scale, precision, unit vocabulary,
// fabricated missing values and mutation of the scientific record. They do not
// qualify actual Orca pronunciation.
func TestFormatMetricSpeech(t *testing.T) {
	tests := []struct {
		name, unit                       string
		number                           float64
		missing, display, speech, reason string
	}{
		{name: "area", unit: "um2", number: 12.5, display: "12.5 µm²", speech: "12 point 5 square micrometers"},
		{name: "zero energy", unit: "pJ", display: "0 pJ", speech: "0 picojoules"},
		{name: "latency precision", unit: "ns", number: 0.12345678901234568, display: "0.12345678901234568 ns", speech: "0 point 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 8 nanoseconds"},
		{name: "negative decibels", unit: "dB", number: -3.25, display: "-3.25 dB", speech: "minus 3 point 2 5 decibels"},
		{name: "polarization example", unit: "µC/cm²", number: 15, display: "15 µC/cm²", speech: "15 microcoulombs per square centimeter"},
		{name: "negative exponent example", unit: "A", number: 1e-12, display: "1e-12 A", speech: "1 times 10 to the minus 12 amperes"},
		{name: "ohm example shortest decimal", unit: "Ω", number: 2e3, display: "2000 Ω", speech: "2000 ohms"},
		{name: "positive exponent", unit: "Ω", number: 2e20, display: "2e+20 Ω", speech: "2 times 10 to the plus 20 ohms"},
		{name: "negative mantissa and padded exponent", unit: "A", number: -1.25e-9, display: "-1.25e-09 A", speech: "minus 1 point 2 5 times 10 to the minus 9 amperes"},
		{name: "negative zero", unit: "pJ", number: math.Copysign(0, -1), display: "-0 pJ", speech: "minus 0 picojoules"},
		{name: "smallest nonzero", unit: "A", number: math.SmallestNonzeroFloat64, display: "5e-324 A", speech: "5 times 10 to the minus 324 amperes"},
		{name: "largest finite", unit: "ns", number: math.MaxFloat64, display: "1.7976931348623157e+308 ns", speech: "1 point 7 9 7 6 9 3 1 3 4 8 6 2 3 1 5 7 times 10 to the plus 308 nanoseconds"},
		{name: "explicit missing", unit: "pJ", missing: "Metric was not produced", reason: "Metric was not produced"},
		{name: "missing reason wins", unit: "unknown", number: math.NaN(), missing: "Evaluation failed", reason: "Evaluation failed"},
		{name: "unknown unit", unit: "mystery", number: 5, reason: `Unsupported metric unit "mystery"`},
		{name: "empty unit", number: 5, reason: "Metric unit is missing"},
		{name: "NaN", unit: "ns", number: math.NaN(), reason: "Metric value is not finite"},
		{name: "positive infinity", unit: "ns", number: math.Inf(1), reason: "Metric value is not finite"},
		{name: "negative infinity", unit: "ns", number: math.Inf(-1), reason: "Metric value is not finite"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := experiment.Metric{
				Name: "synthetic", Value: tt.number, Unit: tt.unit,
				Model: "synthetic-model", Assumptions: []string{"synthetic assumption"},
				Evidence: experiment.EvidenceDefault,
			}
			before := metric
			before.Assumptions = append([]string(nil), metric.Assumptions...)
			want := presentation.Value{Display: tt.display, Speakable: tt.speech, Availability: "available"}
			if tt.reason != "" {
				want = presentation.Value{Display: "Not available", Speakable: "Not available: " + tt.reason, Availability: "unavailable", Reason: tt.reason}
			}
			got := presentation.FormatMetric(metric, tt.missing)
			if got != want {
				t.Errorf("FormatMetric() = %+v, want %+v", got, want)
			}
			// Compare bits separately: NaN != NaN, and ordinary equality hides -0.
			if math.Float64bits(metric.Value) != math.Float64bits(before.Value) {
				t.Error("FormatMetric mutated the numeric value")
			}
			metric.Value, before.Value = 0, 0
			if !reflect.DeepEqual(metric, before) {
				t.Errorf("FormatMetric mutated metric metadata: %+v, want %+v", metric, before)
			}
			if got.Availability == "available" {
				numeric, _, _ := strings.Cut(got.Display, " ")
				// Verify display keeps the exact float, including signed zero.
				parsed, err := strconv.ParseFloat(numeric, 64)
				if err != nil || math.Float64bits(parsed) != math.Float64bits(tt.number) {
					t.Errorf("display %q does not round-trip %v: %v", got.Display, tt.number, err)
				}
			}
		})
	}
}
