//go:build legacy_fyne

package materialdisplay

import "testing"

func TestFormatHelpers(t *testing.T) {
	checks := map[string]string{
		FormatPolarization(0.245):    "24.5 µC/cm²",
		FormatField(2.5e8):           "2.5 MV/cm",
		FormatThickness(10e-9):       "10 nm",
		FormatArea(1500e-18):         "1500 nm²",
		FormatTime(2.5e-6):           "2.5 µs",
		FormatEndurance(1e9):         "10^9 cycles",
		FormatTemperature(300):       "300 K (27°C)",
		FormatEnergy(0.67):           "0.67 eV",
		FormatConductanceRatio(1200): "1k:1",
		FormatVoltage(0.2):           "200 mV",
		FormatDimensionless(1.25):    "1.25",
		FormatPercent(0.125):         "12.5%",
	}
	for got, want := range checks {
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestTextHelpers(t *testing.T) {
	if got := TruncateString("abcdef", 5); got != "ab..." {
		t.Fatalf("TruncateString = %q", got)
	}
	wrapped := WrapText("alpha beta gamma", 10)
	if wrapped != "alpha beta\ngamma" {
		t.Fatalf("WrapText = %q", wrapped)
	}
}
