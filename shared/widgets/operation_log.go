//go:build legacy_fyne

// Package widgets provides shared widget utilities for Fyne GUI development.
package widgets

import "fecim-lattice-tools/shared/widgets/display"

// OperationLog is a reusable widget for displaying timestamped operation history.
type OperationLog = display.OperationLog

// OperationLogConfig holds configuration for creating an OperationLog.
type OperationLogConfig = display.OperationLogConfig

// FormattedLogEntry creates a formatted log entry with a result type.
type FormattedLogEntry = display.FormattedLogEntry

// NewOperationLog creates a new operation log widget.
func NewOperationLog(config OperationLogConfig) *OperationLog {
	return display.NewOperationLog(config)
}
