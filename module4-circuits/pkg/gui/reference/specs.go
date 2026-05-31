//go:build legacy_fyne

package reference

import "fecim-lattice-tools/module4-circuits/pkg/gui/reference/metrics"

// SpecSummary contains derived values shown in the reference specification summary.
type SpecSummary = metrics.SpecSummary

// ParseArraySize parses a selected array size and falls back to 32 when invalid.
func ParseArraySize(selected string) int {
	return metrics.ParseArraySize(selected)
}

// NewSpecSummary returns the derived summary values for an array size.
func NewSpecSummary(size int, latencyNS float64) SpecSummary {
	return metrics.NewSpecSummary(size, latencyNS)
}
