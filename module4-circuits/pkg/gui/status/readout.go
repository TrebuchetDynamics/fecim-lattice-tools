//go:build legacy_fyne

// Package status contains pure status-line formatting helpers for the module 4 GUI.
package status

import "fmt"

// FormatReadStatusLine formats the compact READ operation status text.
func FormatReadStatusLine(row, col, level int, currentUA, tiaVoltageV float64, adcCode int) string {
	return fmt.Sprintf("READ [%d,%d]: State=%d | I=%+.2f µA -> TIA=%+.2f V -> ADC=%d | ~76ns, ~46fJ",
		row, col, level, currentUA, tiaVoltageV, adcCode)
}
