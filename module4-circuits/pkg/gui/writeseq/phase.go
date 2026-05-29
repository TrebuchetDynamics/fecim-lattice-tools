//go:build legacy_fyne

// Package writeseq contains pure helpers for the module 4 write sequence state machine.
package writeseq

import "fmt"

// Phase values mirror gui.WritePhase without importing the stateful gui package.
const (
	PhaseIdle = iota
	PhaseReset
	PhaseHold1
	PhaseWrite
	PhaseHold2
	PhaseVerify
)

// Phase timing constants in nanoseconds for display, not real-time.
const (
	PhaseResetDurationNs  = 100
	PhaseHold1DurationNs  = 50
	PhaseWriteDurationNs  = 200
	PhaseHold2DurationNs  = 50
	PhaseVerifyDurationNs = 80
)

// CellKey generates a map key for a cell coordinate.
func CellKey(row, col int) string {
	return fmt.Sprintf("%d,%d", row, col)
}

// PhaseName returns a human-readable name for a write phase.
func PhaseName(phase int) string {
	switch phase {
	case PhaseIdle:
		return "IDLE"
	case PhaseReset:
		return "RESET"
	case PhaseHold1:
		return "HOLD"
	case PhaseWrite:
		return "WRITE"
	case PhaseHold2:
		return "HOLD"
	case PhaseVerify:
		return "VERIFY"
	default:
		return "UNKNOWN"
	}
}

// PhaseDuration returns the duration in nanoseconds for a phase.
func PhaseDuration(phase int) int {
	switch phase {
	case PhaseReset:
		return PhaseResetDurationNs
	case PhaseHold1:
		return PhaseHold1DurationNs
	case PhaseWrite:
		return PhaseWriteDurationNs
	case PhaseHold2:
		return PhaseHold2DurationNs
	case PhaseVerify:
		return PhaseVerifyDurationNs
	default:
		return 0
	}
}
