//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/shared/widgets/materialdisplay"
)

// Property categories organized by role in the Frankenstein L-K equation.
const (
	CategoryCore        = materialdisplay.CategoryCore
	CategoryGeometry    = materialdisplay.CategoryGeometry
	CategoryLandau      = materialdisplay.CategoryLandau
	CategoryAlpha       = materialdisplay.CategoryAlpha
	CategoryDepol       = materialdisplay.CategoryDepol
	CategoryCircuit     = materialdisplay.CategoryCircuit
	CategoryNLS         = materialdisplay.CategoryNLS
	CategoryConductance = materialdisplay.CategoryConductance
)

// ModelUsage indicates which physics models use a parameter.
type ModelUsage = materialdisplay.ModelUsage

// FormattedProperty holds a material property with display formatting.
type FormattedProperty = materialdisplay.FormattedProperty

// Model usage markers retained for legacy package tests and internal callers.
var (
	lkModel       = ModelUsage{LandauKh: true}
	preisachModel = ModelUsage{Preisach: true}
	bothModels    = ModelUsage{LandauKh: true, Preisach: true}

	// CategoryOrder defines the display order matching the L-K equation structure.
	CategoryOrder = materialdisplay.CategoryOrder
)

// FormatPolarization converts C/m² to µC/cm² display string.
func FormatPolarization(cM2 float64) string { return materialdisplay.FormatPolarization(cM2) }

// FormatField converts V/m to MV/cm display string.
func FormatField(vM float64) string { return materialdisplay.FormatField(vM) }

// FormatThickness converts m to nm display string.
func FormatThickness(m float64) string { return materialdisplay.FormatThickness(m) }

// FormatArea converts m² to nm² display string.
func FormatArea(m2 float64) string { return materialdisplay.FormatArea(m2) }

// FormatTime converts seconds to appropriate time unit.
func FormatTime(s float64) string { return materialdisplay.FormatTime(s) }

// FormatEndurance formats cycle count with superscript notation.
func FormatEndurance(cycles float64) string { return materialdisplay.FormatEndurance(cycles) }

// FormatTemperature converts K to display string with Celsius.
func FormatTemperature(k float64) string { return materialdisplay.FormatTemperature(k) }

// FormatEnergy formats energy in eV.
func FormatEnergy(ev float64) string { return materialdisplay.FormatEnergy(ev) }

// FormatConductanceRatio formats a ratio for display.
func FormatConductanceRatio(ratio float64) string {
	return materialdisplay.FormatConductanceRatio(ratio)
}

// FormatVoltage formats voltage in V.
func FormatVoltage(v float64) string { return materialdisplay.FormatVoltage(v) }

// FormatDimensionless formats a dimensionless value.
func FormatDimensionless(v float64) string { return materialdisplay.FormatDimensionless(v) }

// FormatPercent formats a fraction as percentage.
func FormatPercent(v float64) string { return materialdisplay.FormatPercent(v) }

// GetMaterialProperties extracts properties relevant to the Frankenstein L-K equation.
func GetMaterialProperties(mat *physics.Material) []FormattedProperty {
	return materialdisplay.GetMaterialProperties(mat)
}

// GetPropertiesByCategory filters properties by category.
func GetPropertiesByCategory(props []FormattedProperty, category string) []FormattedProperty {
	return materialdisplay.GetPropertiesByCategory(props, category)
}

// HasCategory returns true if any property has the given category.
func HasCategory(props []FormattedProperty, category string) bool {
	return materialdisplay.HasCategory(props, category)
}

// TruncateString truncates a string to maxLen with ellipsis.
func TruncateString(s string, maxLen int) string { return materialdisplay.TruncateString(s, maxLen) }

// WrapText wraps text to roughly maxWidth characters per line.
func WrapText(s string, maxWidth int) string { return materialdisplay.WrapText(s, maxWidth) }
