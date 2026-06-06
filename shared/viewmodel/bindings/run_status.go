package bindings

import (
	"fmt"
	"math"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

// State describes the lifecycle of a long-running UI operation.
type State string

const (
	StateIdle      State = "idle"
	StateRunning   State = "running"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

// String returns the stable binding value for s.
func (s State) String() string { return string(s) }

// RunStatus exposes operation lifecycle state through Fyne data bindings so
// status bars, toolbar actions, and progress views can share one source of
// truth without manual widget fan-out.
type RunStatus struct {
	Operation  binding.String
	Phase      binding.String
	Detail     binding.String
	State      binding.String
	StatusLine binding.String
	Progress   binding.Float
	Running    binding.Bool
	Terminal   binding.Bool
}

// NewRunStatus creates an idle status model for operation.
func NewRunStatus(operation string) *RunStatus {
	r := &RunStatus{
		Operation:  binding.NewString(),
		Phase:      binding.NewString(),
		Detail:     binding.NewString(),
		State:      binding.NewString(),
		StatusLine: binding.NewString(),
		Progress:   binding.NewFloat(),
		Running:    binding.NewBool(),
		Terminal:   binding.NewBool(),
	}
	r.set(operation, "", "", StateIdle, 0)
	return r
}

// Start marks the operation running and sets the initial phase/detail.
func (r *RunStatus) Start(phase, detail string) {
	r.set(mustGetString(r.Operation), phase, detail, StateRunning, 0)
}

// Update changes the running phase/detail/progress. Progress is clamped to
// [0,1] for direct use with Fyne progress bars.
func (r *RunStatus) Update(phase, detail string, progress float64) {
	r.set(mustGetString(r.Operation), phase, detail, StateRunning, progress)
}

// Complete marks the operation completed and pins progress to 100%.
func (r *RunStatus) Complete(detail string) {
	r.set(mustGetString(r.Operation), "Complete", detail, StateCompleted, 1)
}

// Fail marks the operation failed.
func (r *RunStatus) Fail(detail string) {
	r.set(mustGetString(r.Operation), "Failed", detail, StateFailed, mustGetFloat(r.Progress))
}

// Cancel marks the operation cancelled.
func (r *RunStatus) Cancel(detail string) {
	r.set(mustGetString(r.Operation), "Cancelled", detail, StateCancelled, mustGetFloat(r.Progress))
}

func (r *RunStatus) set(operation, phase, detail string, state State, progress float64) {
	progress = clamp01(progress)
	_ = r.Operation.Set(operation)
	_ = r.Phase.Set(phase)
	_ = r.Detail.Set(detail)
	_ = r.State.Set(state.String())
	_ = r.Progress.Set(progress)
	_ = r.Running.Set(state == StateRunning)
	_ = r.Terminal.Set(state == StateCompleted || state == StateFailed || state == StateCancelled)
	_ = r.StatusLine.Set(formatStatusLine(operation, phase, detail, state))
}

func formatStatusLine(operation, phase, detail string, state State) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(phase) != "" {
		parts = append(parts, phase)
	} else {
		parts = append(parts, state.String())
	}
	if strings.TrimSpace(detail) != "" {
		parts = append(parts, detail)
	}
	return fmt.Sprintf("%s: %s", operation, strings.Join(parts, " — "))
}

func clamp01(value float64) float64 {
	if math.IsNaN(value) || value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func mustGetString(item binding.String) string {
	value, _ := item.Get()
	return value
}

func mustGetFloat(item binding.Float) float64 {
	value, _ := item.Get()
	return value
}
