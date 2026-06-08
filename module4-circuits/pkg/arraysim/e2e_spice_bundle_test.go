package arraysim

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/peripherals"
)

func TestModule4ArraySimE2ESPICEPeripheralBundleWideMatrix(t *testing.T) {
	cases := []struct {
		name string
		rows int
		cols int
		cfg  SpiceExportConfig
	}{
		{name: "defaults", rows: 2, cols: 2, cfg: SpiceExportConfig{}},
		{name: "custom_peripherals", rows: 3, cols: 4, cfg: SpiceExportConfig{
			Title: "Module4 custom peripheral bundle",
			DAC:   peripherals.ThermometerDAC(5),
			ADC:   peripherals.RampADC(6, 4),
			TIA:   &peripherals.TIA{Gain: 25e3, Bandwidth: 80e6, InputNoiseRMS: 2e-12, OutputOffset: 0.02, MaxInputCurrent: 80e-6, MaxOutputVoltage: 1.1},
			SH:    &peripherals.SampleAndHold{HoldCapacitance: 2e-12, SwitchResistance: 750, LeakageResistance: 8e9, AcquisitionTimeNS: 15},
			VReg:  &peripherals.VoltageRegulator{NominalVoltage: 1.1, DropoutVoltage: 0.12, OutputResistance: 0.8, QuiescentCurrent: 15e-6, PSRRdB: 50},
		}},
		{name: "terminated_boundary", rows: 4, cols: 3, cfg: SpiceExportConfig{Title: "Terminated array"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := spiceBundleParamsE2E(tc.rows, tc.cols)
			if tc.name == "terminated_boundary" {
				params.Boundary = BoundaryParams{WLDriveResistance: 2, BLDriveResistance: 3, WLTerminationResistance: 40, BLTerminationResistance: 70, WLTerminationVoltage: 0.01}
			}
			deck, err := ExportCrossbarSPICE(params, tc.cfg)
			if err != nil {
				t.Fatalf("ExportCrossbarSPICE: %v", err)
			}
			wantTitle := tc.cfg.Title
			if strings.TrimSpace(wantTitle) == "" {
				wantTitle = "FeCIM crossbar behavioral export"
			}
			for _, marker := range []string{wantTitle, "Peripheral behavioral subcircuits", ".subckt DAC5", ".subckt SAMPLE_HOLD", ".subckt TIA_BASIC", ".subckt ADC5", ".subckt VREG_BASIC", "VDD_RAW", ".control", ".end"} {
				if !strings.Contains(deck, marker) {
					t.Fatalf("SPICE deck missing %q:\n%s", marker, deck)
				}
			}
			assertSPICEElementCountE2E(t, deck, `(?m)^RCELL_`, tc.rows*tc.cols)
			assertSPICEElementCountE2E(t, deck, `(?m)^VWL_SRC_`, tc.rows)
			assertSPICEElementCountE2E(t, deck, `(?m)^VBL_SRC_`, tc.cols)
			assertSPICEElementCountE2E(t, deck, `(?m)^XADC_`, tc.cols)
			assertSPICEElementCountE2E(t, deck, `(?m)^XSH_`, tc.cols)
			assertSPICEElementCountE2E(t, deck, `(?m)^RWL_[0-9]`, tc.rows*(tc.cols-1))
			assertSPICEElementCountE2E(t, deck, `(?m)^RBL_[0-9]`, tc.cols*(tc.rows-1))
			for r := 0; r < tc.rows; r++ {
				for c := 0; c < tc.cols; c++ {
					g := params.Conductance[r][c]
					wantR := 1.0 / g
					marker := fmt.Sprintf("RCELL_%d_%d wl_%d_%d bl_%d_%d %.9g", r, c, r, c, r, c, wantR)
					if !strings.Contains(deck, marker) {
						t.Fatalf("SPICE deck missing resistor marker %q", marker)
					}
				}
			}
		})
	}
}

func TestModule4ArraySimE2ESPICERejectsMalformedMatrices(t *testing.T) {
	bad := []SolveParams{
		{},
		{Conductance: [][]float64{{}}},
		{Conductance: [][]float64{{1e-6, 2e-6}, {3e-6}}},
	}
	for i, params := range bad {
		if deck, err := ExportCrossbarSPICE(params, SpiceExportConfig{}); err == nil || deck != "" {
			t.Fatalf("bad SPICE case %d should fail without deck, deck=%q err=%v", i, deck, err)
		}
	}
}

func spiceBundleParamsE2E(rows, cols int) SolveParams {
	conductance := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		conductance[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			conductance[r][c] = 20e-6 + float64(r*cols+c+1)*5e-6
		}
	}
	wl := make([]float64, rows)
	for r := range wl {
		wl[r] = 0.2 - 0.01*float64(r)
	}
	bl := make([]float64, cols)
	return SolveParams{WLVoltages: wl, BLVoltages: bl, Conductance: conductance, Geometry: DefaultCellGeometry(), Wire: WireParams{RWordLine: 0.7, RBitLine: 2.9}, Boundary: BoundaryParams{WLDriveResistance: 1.5, BLDriveResistance: 2.5}}
}

func assertSPICEElementCountE2E(t *testing.T, deck, expr string, want int) {
	t.Helper()
	re := regexp.MustCompile(expr)
	if got := len(re.FindAllString(deck, -1)); got != want {
		t.Fatalf("SPICE count %s = %d, want %d\n%s", expr, got, want, deck)
	}
}
