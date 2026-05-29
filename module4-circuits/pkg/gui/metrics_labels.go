//go:build legacy_fyne

package gui

import guimetrics "fecim-lattice-tools/module4-circuits/pkg/gui/metrics"

func formatMetricVTIAMV(voltageV float64) string {
	return guimetrics.FormatVTIAMV(voltageV)
}

func formatMetricICellUA(currentUA float64) string {
	return guimetrics.FormatICellUA(currentUA)
}

func formatMetricADCCode(code int) string {
	return guimetrics.FormatADCCode(code)
}

func formatMetricConductanceUS(conductanceUS float64) string {
	return guimetrics.FormatConductanceUS(conductanceUS)
}

func formatMetricLevel(level int) string {
	return guimetrics.FormatLevel(level)
}

func formatOverlayBottomValue(mode string, value float64) string {
	return guimetrics.FormatOverlayBottomValue(mode, value)
}
