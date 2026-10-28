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

// TestHysteresisValues checks SI conversion independently of the session/GUI.
func TestHysteresisValues(t *testing.T) {
	for _, tt := range []struct {
		name                  string
		format                func(float64, string) presentation.Value
		si, factor            float64
		unit, speech, missing string
	}{
		{"negative E", presentation.FormatElectricField, -1.25e8, 1e-8, "MV/cm", "minus 1 point 2 5 megavolts per centimeter", ""},
		{"zero E", presentation.FormatElectricField, 0, 1e-8, "MV/cm", "0 megavolts per centimeter", ""},
		{"signed zero P", presentation.FormatPolarization, math.Copysign(0, -1), 100, "µC/cm²", "minus 0 microcoulombs per square centimeter", ""},
		{"negative P", presentation.FormatPolarization, -0.25, 100, "µC/cm²", "minus 25 microcoulombs per square centimeter", ""},
		{"exponent E", presentation.FormatElectricField, 1e-12, 1e-8, "MV/cm", "1 times 10 to the minus 20 megavolts per centimeter", ""},
		{"exponent P", presentation.FormatPolarization, 1e20, 100, "µC/cm²", "1 times 10 to the plus 22 microcoulombs per square centimeter", ""},
		{"missing", presentation.FormatPolarization, 0, 100, "µC/cm²", "", "No sample was produced"},
		{"nonfinite", presentation.FormatElectricField, math.NaN(), 1e-8, "MV/cm", "", ""},
		{"overflow", presentation.FormatPolarization, math.MaxFloat64, 100, "µC/cm²", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bits := math.Float64bits(tt.si)
			got := tt.format(tt.si, tt.missing)
			if math.Float64bits(tt.si) != bits {
				t.Fatal("canonical SI changed")
			}
			if tt.missing != "" || math.IsNaN(tt.si) || math.IsInf(tt.si*tt.factor, 0) {
				if got.Availability != "unavailable" || got.Reason == "" || !strings.Contains(got.Speakable, got.Reason) {
					t.Fatalf("missing reason: %+v", got)
				}
				if tt.missing != "" && got.Reason != tt.missing {
					t.Fatal("reason lost")
				}
				return
			}
			want := strconv.FormatFloat(tt.si*tt.factor, 'g', -1, 64) + " " + tt.unit
			if got.Display != want || got.Speakable != tt.speech || got.Availability != "available" {
				t.Fatalf("got %+v want %q / %q", got, want, tt.speech)
			}
		})
	}
	if got := presentation.FormatMetric(experiment.Metric{Value: 1, Unit: "V/m"}, ""); got.Reason != `Unsupported metric unit "V/m"` {
		t.Fatal("unsupported-unit compatibility changed")
	}
}

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
