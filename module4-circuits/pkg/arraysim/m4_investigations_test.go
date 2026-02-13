package arraysim

import (
	"fmt"
	"strings"
	"testing"
)

// M4-INV-01: read margin loss in ADC LSB due to selector Ron.
// NOTE: Tier-A dense solver is only practical for small matrices in unit tests;
// we report Ron-only margin loss via selector series conductance and ADC current LSB.
func TestM4INV01_ReadMarginVsSelectorRon(t *testing.T) {
	sense := SenseChain{TIA: TIAConfig{Rf: 200e3, Vref: 0.2, Vmin: 0, Vmax: 1.2}, ADC: ADCConfig{Bits: 10, Vmin: 0, Vmax: 1.2}}
	baseCellR := 1.0 / 55e-6 // mid-level ~55 uS
	currentLSB := sense.CurrentLSB()

	rons := []float64{0, 100, 500, 1e3, 5e3, 10e3}
	sizes := []int{64, 128}

	for _, n := range sizes {
		baselineMargin := (0.2 / baseCellR) / currentLSB
		for _, ron := range rons {
			effR := baseCellR + ron
			marginLSB := (0.2 / effR) / currentLSB
			loss := baselineMargin - marginLSB
			t.Logf("array=%dx%d Ron=%.0fΩ margin_LSB=%.2f loss_LSB=%.2f", n, n, ron, marginLSB, loss)
			if ron > 0 && !(loss > 0) {
				t.Fatalf("expected positive margin loss at Ron=%.0f", ron)
			}
		}
	}
}

// M4-INV-02: max N before WL RC delay >10ns per node.
func TestM4INV02_WordlineRCDelayBudget(t *testing.T) {
	type node struct {
		name, tech string
		pitch      float64
		width      float64
		thickness  float64
		rho        float64
		cgPerCell  float64
	}
	nodes := []node{
		{"130nm", "130nm", 0.46e-6, 0.20e-6, 0.30e-6, 2.8e-8, 2.2e-15},
		{"65nm", "65nm", 0.23e-6, 0.12e-6, 0.20e-6, 2.6e-8, 1.1e-15},
		{"28nm", "28nm", 0.10e-6, 0.08e-6, 0.16e-6, 2.4e-8, 0.55e-15},
		{"14nm", "14nm", 0.06e-6, 0.05e-6, 0.12e-6, 2.2e-8, 0.32e-15},
	}
	const tPulse = 10e-9

	for _, nd := range nodes {
		rSeg := nd.rho * nd.pitch / (nd.width * nd.thickness)
		maxN := 1
		for n := 1; n <= 4096; n++ {
			rwl := rSeg * float64(n)
			cwl := (0.20e-15 + nd.cgPerCell) * float64(n)
			delay := rwl * cwl
			if delay > tPulse {
				maxN = n - 1
				break
			}
			maxN = n
		}
		t.Logf("node=%s Rseg=%.3fΩ Ccell=%.3ffF max_N=%d", nd.name, rSeg, nd.cgPerCell*1e15, maxN)
		if maxN < 32 {
			t.Fatalf("unexpectedly low max_N at %s: %d", nd.name, maxN)
		}
	}
}

// M4-INV-07: ngspice-ready export from current M4 array state.
func TestM4INV07_SPICEExportFromArrayState(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{0.2, 0, 0, 0},
		BLVoltages: []float64{0, 0, 0, 0},
		Conductance: [][]float64{
			{40e-6, 55e-6, 70e-6, 55e-6},
			{55e-6, 55e-6, 55e-6, 55e-6},
			{55e-6, 55e-6, 55e-6, 55e-6},
			{55e-6, 55e-6, 55e-6, 55e-6},
		},
	}
	netlist, err := ExportCrossbarSPICE(params, SpiceExportConfig{Title: "M4-INV-07"})
	if err != nil {
		t.Fatalf("ExportCrossbarSPICE failed: %v", err)
	}
	for _, token := range []string{".control", "op", ".end", "RCELL_0_0", "XADC_0"} {
		if !strings.Contains(netlist, token) {
			t.Fatalf("missing token %q in netlist", token)
		}
	}
	count := strings.Count(netlist, "RCELL_")
	if count != 16 {
		t.Fatalf("expected 16 RCELL entries, got %d", count)
	}
	t.Logf("generated ngspice deck (%d bytes) with %d cell elements", len(netlist), count)
	_ = fmt.Sprintf
}
