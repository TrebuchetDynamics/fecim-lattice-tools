//go:build legacy_fyne

// Package uilog contains module 4 UI logging helpers.
package uilog

import (
	"sync"

	"fecim-lattice-tools/shared/logging"
)

var circuitsLogOnce sync.Once
var circuitsLog *logging.Logger

// Logger returns the lazily initialized circuits logger.
func Logger() *logging.Logger {
	circuitsLogOnce.Do(func() {
		circuitsLog = logging.NewLogger("circuits")
	})
	return circuitsLog
}

// Action logs a debug-level user action.
func Action(format string, args ...interface{}) {
	if !logging.IsVerbose(logging.VerbosityDebug) {
		return
	}
	Logger().Debug("ACTION: "+format, args...)
}

// Input logs a debug-level user input event.
func Input(format string, args ...interface{}) {
	if !logging.IsVerbose(logging.VerbosityDebug) {
		return
	}
	Logger().Debug("INPUT: "+format, args...)
}
