//go:build legacy_fyne

package status

import (
	"fmt"

	"fecim-lattice-tools/module4-circuits/pkg/gui/metrics"
)

// OverlayCellLabels contains the dual-line overlay labels for selected cell diagnostics.
type OverlayCellLabels struct {
	TopLabel    string
	BottomLabel string
}

// FormatOverlayCellInfo composes dual-line overlay labels for selected cell diagnostics.
func FormatOverlayCellInfo(level int, value float64, mode string) OverlayCellLabels {
	top := fmt.Sprintf("L: %d", level)
	bottom := metrics.FormatOverlayBottomValue(mode, value)
	return OverlayCellLabels{TopLabel: top, BottomLabel: bottom}
}
