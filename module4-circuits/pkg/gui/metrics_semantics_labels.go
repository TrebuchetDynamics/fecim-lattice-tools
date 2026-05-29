//go:build legacy_fyne

package gui

import guimetrics "fecim-lattice-tools/module4-circuits/pkg/gui/metrics"

// readModeMetricLabels returns the canonical READ-mode metric labels shown in UI/docs.
func readModeMetricLabels() []string {
	return guimetrics.ReadModeLabels()
}
