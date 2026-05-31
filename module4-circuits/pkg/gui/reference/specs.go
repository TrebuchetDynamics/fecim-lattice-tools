//go:build legacy_fyne

package reference

import "fecim-lattice-tools/module4-circuits/pkg/gui/reference/metrics"

// ComponentRow contains one row in the reference component summary table.
type ComponentRow = metrics.ComponentRow

// SpecSummary contains derived values shown in the reference specification summary.
type SpecSummary = metrics.SpecSummary

// ParseArraySize parses a selected array size and falls back to 32 when invalid.
func ParseArraySize(selected string) int {
	return metrics.ParseArraySize(selected)
}

// ComponentRows returns the canonical reference component summary rows for an array size.
func ComponentRows(size int) []ComponentRow {
	return metrics.ComponentRows(size)
}

// NewSpecSummary returns the derived summary values for an array size.
func NewSpecSummary(size int, latencyNS float64) SpecSummary {
	return metrics.NewSpecSummary(size, latencyNS)
}
