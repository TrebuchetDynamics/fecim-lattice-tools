package moduleapi

import (
	"math"
	"strconv"
	"strings"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/presentation"
	"fecim-lattice-tools/workbench/project"
)

type ConfigurationField struct {
	project.ConfigurationField
	Label        string             `json:"label"`
	Unit         string             `json:"unit"`
	Presentation presentation.Value `json:"presentation"`
}

// FieldFromConfiguration retains literal editor input. Unit comes from the Go
// schema/constraint/sweep target, not inferred from input text in JavaScript.
// Integer inputs are spoken exactly, without float64 coercion (notably seeds).
func FieldFromConfiguration(f project.ConfigurationField, unit string) ConfigurationField {
	label := map[string]string{"/name": "Study name", "/hypothesis": "Hypothesis", "/device/material": "Material", "/device/conductance_levels": "Conductance levels", "/device/g_min_s": "Minimum conductance", "/device/g_max_s": "Maximum conductance", "/array/rows": "Array rows", "/array/cols": "Array columns", "/array/read_voltage_v": "Read voltage", "/circuit/adc_bits": "ADC resolution", "/circuit/dac_bits": "DAC resolution", "/circuit/tia_gain_ohm": "TIA gain", "/circuit/tech_node": "Technology node", "/seed": "Random seed", "/max_points": "Maximum sweep points"}[f.Path]
	if label == "" {
		label = f.Document + " " + f.Path
	}
	value := presentation.Value{Display: f.Value, Speakable: f.Value, Availability: "available"}
	if f.Kind != "string" && unit != "" {
		number, err := strconv.ParseFloat(f.Value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			value = presentation.Value{Display: "Not available", Speakable: "Not available: literal numeric input needs validation", Availability: "unavailable", Reason: "Literal numeric input needs validation"}
		} else {
			value = presentation.FormatMetric(experiment.Metric{Value: number, Unit: unit}, "")
			if f.Kind == "int" || f.Kind == "int64" {
				suffix := strings.TrimPrefix(presentation.FormatMetric(experiment.Metric{Unit: unit}, "").Speakable, "0")
				value.Display = f.Value + " " + unit
				value.Speakable = f.Value + suffix
			}
		}
	}
	return ConfigurationField{ConfigurationField: f, Label: label, Unit: unit, Presentation: value}
}

// ConfigurationUnits selects canonical input units, including constraints and
// numeric sweep coordinates. Labels/reasons remain Go-owned literal text.
func ConfigurationUnits(f project.ConfigurationField, b project.Bundle) string {
	target := strings.TrimPrefix(f.Path, "/")
	if f.Document == "sweep.yaml" {
		parts := strings.Split(target, "/")
		if len(parts) >= 3 && parts[0] == "parameters" {
			i, e := strconv.Atoi(parts[1])
			if e == nil && i >= 0 && i < len(b.Sweep.Parameters) {
				if parts[2] == "range" && len(parts) == 4 && parts[3] == "count" {
					return "count"
				}
				target = strings.ReplaceAll(b.Sweep.Parameters[i].Path, ".", "/")
			}
		}
	}
	if f.Document == "project.yaml" {
		parts := strings.Split(target, "/")
		if len(parts) == 3 && parts[0] == "constraints" && parts[2] == "value" {
			i, e := strconv.Atoi(parts[1])
			if e == nil && i >= 0 && i < len(b.Project.Constraints) {
				return b.Project.Constraints[i].Unit
			}
		}
	}
	return map[string]string{"device/conductance_levels": "levels", "device/g_min_s": "S", "device/g_max_s": "S", "array/rows": "count", "array/cols": "count", "array/read_voltage_v": "V", "circuit/adc_bits": "bits", "circuit/dac_bits": "bits", "circuit/tia_gain_ohm": "Ω", "max_points": "count"}[target]
}
