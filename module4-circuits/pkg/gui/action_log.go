//go:build legacy_fyne

// Package gui provides logging helpers for module 4 UI actions and inputs.
package gui

import (
	"fecim-lattice-tools/module4-circuits/pkg/gui/status"
	"fecim-lattice-tools/module4-circuits/pkg/gui/uilog"
	"fecim-lattice-tools/shared/logging"
)

func getCircuitsLog() *logging.Logger {
	return uilog.Logger()
}

func logAction(format string, args ...interface{}) {
	uilog.Action(format, args...)
}

func logInput(format string, args ...interface{}) {
	uilog.Input(format, args...)
}

func opModeLabel(mode OpMode) string {
	return status.OpModeLabel(int(mode))
}

func dacModeLabel(mode DACMode) string {
	return status.DACModeLabel(int(mode))
}

func dacRangeLabel(mode DACRangeMode) string {
	return status.DACRangeLabel(int(mode))
}
