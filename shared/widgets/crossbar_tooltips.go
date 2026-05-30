//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/widgets/help"
)

// ConductanceTooltip generates tooltip for conductance cell with progressive disclosure.
func ConductanceTooltip(row, col int, G float64, array *crossbar.Array) string {
	return help.ConductanceTooltip(row, col, G, array)
}

// IRDropTooltip generates tooltip for IR drop analysis with progressive disclosure.
func IRDropTooltip(row, col int, irAnalysis *crossbar.IRDropAnalysis, array *crossbar.Array) string {
	return help.IRDropTooltip(row, col, irAnalysis, array)
}

// IRDropTooltipWithArch generates tooltip for IR drop analysis including architecture info.
func IRDropTooltipWithArch(row, col int, irAnalysis *crossbar.IRDropAnalysis, array *crossbar.Array, arch string) string {
	return help.IRDropTooltipWithArch(row, col, irAnalysis, array, arch)
}

// SneakPathTooltip generates tooltip for sneak path analysis with progressive disclosure.
func SneakPathTooltip(row, col int, sneakAnalysis *crossbar.SneakPathAnalysis, selectedRow, selectedCol int, array *crossbar.Array) string {
	return help.SneakPathTooltip(row, col, sneakAnalysis, selectedRow, selectedCol, array)
}

// SneakPathTooltipWithArch generates tooltip for sneak path analysis including architecture info.
func SneakPathTooltipWithArch(row, col int, sneakAnalysis *crossbar.SneakPathAnalysis, selectedRow, selectedCol int, array *crossbar.Array, arch string) string {
	return help.SneakPathTooltipWithArch(row, col, sneakAnalysis, selectedRow, selectedCol, array, arch)
}

// MVMResultTooltip generates tooltip for MVM output row.
func MVMResultTooltip(row int, mvmResult *crossbar.MVMResult) string {
	return help.MVMResultTooltip(row, mvmResult)
}

// ComprehensiveTooltip generates comprehensive tooltip combining all analyses.
func ComprehensiveTooltip(row, col int, array *crossbar.Array, irAnalysis *crossbar.IRDropAnalysis, sneakAnalysis *crossbar.SneakPathAnalysis, mvmResult *crossbar.MVMResult) string {
	return help.ComprehensiveTooltip(row, col, array, irAnalysis, sneakAnalysis, mvmResult)
}
