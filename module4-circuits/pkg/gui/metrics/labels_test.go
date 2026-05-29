//go:build legacy_fyne

package metrics

import (
	"strings"
	"testing"
)

func TestReadModeMetricLabels(t *testing.T) {
	labels := ReadModeLabels()
	want := []string{
		"I_cell (µA)",
		"V_TIA (V)",
		"ADC Code (0–2^N-1)",
		"Noise RMS (µA)",
		"SNR (dB)",
		"I_LSB (µA/code)",
	}
	if len(labels) != len(want) {
		t.Fatalf("labels length = %d, want %d", len(labels), len(want))
	}
	for i := range want {
		if labels[i] != want[i] {
			t.Fatalf("labels[%d] = %q, want %q", i, labels[i], want[i])
		}
	}
}

func TestMetricFormatters(t *testing.T) {
	checks := map[string]string{
		FormatVTIAMV(0.01234):                     "+0.01 V",
		FormatICellUA(-1.236):                     "-1.24 µA",
		FormatADCCode(127):                        "127",
		FormatConductanceUS(14.44):                "14.4 µS",
		FormatLevel(17):                           "17",
		FormatOverlayBottomValue("Icell", 2.5e-6): "I: +2.50 µA",
		FormatOverlayBottomValue("Vcell", -0.25):  "V: -0.25 V",
	}
	for got, want := range checks {
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestSneakPathMetricsSummary(t *testing.T) {
	currents := [][]float64{
		{0, 1e-7, 2e-7},
		{3e-7, 9e-7, 4e-7},
		{5e-7, 6e-7, 7e-7},
	}
	metrics := ComputeSneakPath(currents, 1, 1)
	if metrics.AffectedCells != 7 {
		t.Fatalf("AffectedCells = %d, want 7", metrics.AffectedCells)
	}
	if len(metrics.TopAffectedCells) != 3 {
		t.Fatalf("TopAffectedCells len = %d, want 3", len(metrics.TopAffectedCells))
	}
	if metrics.TopAffectedCells[0].Row != 2 || metrics.TopAffectedCells[0].Col != 2 {
		t.Fatalf("top cell = [%d,%d], want [2,2]", metrics.TopAffectedCells[0].Row, metrics.TopAffectedCells[0].Col)
	}
	summary := FormatSneakPathSummary(metrics)
	if !strings.Contains(summary, "0T1R: sneak=") || !strings.Contains(summary, "top=[2,2]") {
		t.Fatalf("summary missing expected detail: %q", summary)
	}
}
