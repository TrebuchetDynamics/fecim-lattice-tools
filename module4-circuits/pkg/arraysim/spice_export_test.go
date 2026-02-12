package arraysim

import (
	"strings"
	"testing"
)

func TestExportCrossbarSPICE_IncludesCellsWiresAndPeripherals(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{0.8, 0.2},
		BLVoltages: []float64{0.1, 0.0},
		Conductance: [][]float64{
			{2e-6, 5e-6},
			{1e-6, 9e-6},
		},
	}
	deck, err := ExportCrossbarSPICE(params, SpiceExportConfig{Title: "test deck"})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	checks := []string{
		"* test deck",
		"RWL_0_0",
		"RBL_0_0",
		"RCELL_1_1",
		"XSH_0",
		"XTIA_1",
		"XADC_1",
		"XREG",
		".control",
	}
	for _, c := range checks {
		if !strings.Contains(deck, c) {
			t.Fatalf("expected %q in deck", c)
		}
	}
}
