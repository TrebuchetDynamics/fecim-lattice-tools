//go:build legacy_fyne

// Package metrics contains pure formatting and readout helpers for the module 4 GUI.
package metrics

import "fmt"

// FormatVTIAMV formats a transimpedance amplifier voltage readout.
func FormatVTIAMV(voltageV float64) string {
	return fmt.Sprintf("%+.2f V", voltageV)
}

// FormatICellUA formats a cell-current readout in microamps.
func FormatICellUA(currentUA float64) string {
	return fmt.Sprintf("%+.2f µA", currentUA)
}

// FormatADCCode formats an ADC code readout.
func FormatADCCode(code int) string {
	return fmt.Sprintf("%d", code)
}

// FormatConductanceUS formats a conductance readout in microsiemens.
func FormatConductanceUS(conductanceUS float64) string {
	return fmt.Sprintf("%.1f µS", conductanceUS)
}

// FormatLevel formats a quantized conductance level readout.
func FormatLevel(level int) string {
	return fmt.Sprintf("%d", level)
}

// FormatOverlayBottomValue formats the secondary overlay annotation.
func FormatOverlayBottomValue(mode string, value float64) string {
	if mode == "Icell" {
		// value in A -> µA
		return fmt.Sprintf("I: %+.2f µA", value*1e6)
	}
	return fmt.Sprintf("V: %+.2f V", value)
}

// ReadModeLabels returns the canonical READ-mode metric labels shown in UI/docs.
func ReadModeLabels() []string {
	return []string{
		"I_cell (µA)",
		"V_TIA (V)",
		"ADC Code (0–2^N-1)",
		"Noise RMS (µA)",
		"SNR (dB)",
		"I_LSB (µA/code)",
	}
}
