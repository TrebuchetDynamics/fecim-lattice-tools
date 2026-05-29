//go:build legacy_fyne

package gui

import guimetrics "fecim-lattice-tools/module4-circuits/pkg/gui/metrics"

type SneakCellImpact = guimetrics.SneakCellImpact

type SneakPathMetrics = guimetrics.SneakPathMetrics

func computeSneakPathMetrics(currents [][]float64, selectedRow, selectedCol int) SneakPathMetrics {
	return guimetrics.ComputeSneakPath(currents, selectedRow, selectedCol)
}

func formatSneakPathSummary(metrics SneakPathMetrics) string {
	return guimetrics.FormatSneakPathSummary(metrics)
}
