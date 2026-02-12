package arraysim

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/shared/peripherals"
)

// SpiceExportConfig controls behavioral SPICE export details.
type SpiceExportConfig struct {
	Title string
	DAC   *peripherals.DAC
	ADC   *peripherals.ADC
	TIA   *peripherals.TIA
	SH    *peripherals.SampleAndHold
	VReg  *peripherals.VoltageRegulator
}

// ExportCrossbarSPICE exports a behavioral SPICE deck for the array + peripherals.
func ExportCrossbarSPICE(params SolveParams, cfg SpiceExportConfig) (string, error) {
	rows := len(params.Conductance)
	if rows == 0 {
		return "", fmt.Errorf("arraysim: empty conductance matrix")
	}
	cols := len(params.Conductance[0])
	if cols == 0 {
		return "", fmt.Errorf("arraysim: empty conductance row")
	}
	for r := 1; r < rows; r++ {
		if len(params.Conductance[r]) != cols {
			return "", fmt.Errorf("arraysim: jagged conductance matrix")
		}
	}

	geom := params.Geometry.WithDefaults()
	wire := params.Wire.WithDefaults(geom)
	boundary := params.Boundary.WithDefaults(wire)

	title := cfg.Title
	if strings.TrimSpace(title) == "" {
		title = "FeCIM crossbar behavioral export"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "* %s\n", title)
	fmt.Fprintf(&b, ".param RWL=%.9g RBL=%.9g\n", wire.RWordLine, wire.RBitLine)
	fmt.Fprintf(&b, ".param RWLDRV=%.9g RBLDRV=%.9g\n\n", boundary.WLDriveResistance, boundary.BLDriveResistance)

	b.WriteString(peripherals.BuildBehavioralSpiceSubcircuits(cfg.DAC, cfg.ADC, cfg.TIA, cfg.SH, cfg.VReg))
	b.WriteString("\n")
	b.WriteString("* Regulated supply for peripherals\n")
	b.WriteString("VDD_RAW vdd_raw 0 1.8\n")
	b.WriteString("XREG vdd_raw vdd_periph 0 VREG_BASIC\n\n")

	for r := 0; r < rows; r++ {
		vwl := 0.0
		if r < len(params.WLVoltages) {
			vwl = params.WLVoltages[r]
		}
		fmt.Fprintf(&b, "VWL_SRC_%d wl_src_%d 0 %.9g\n", r, r, vwl)
		fmt.Fprintf(&b, "XDAC_WL_%d wl_src_%d wl_drv_%d 0 DAC5\n", r, r, r)
		fmt.Fprintf(&b, "RWL_DRV_%d wl_drv_%d wl_%d_0 {RWLDRV}\n", r, r, r)
	}
	b.WriteString("\n")

	for c := 0; c < cols; c++ {
		vbl := 0.0
		if c < len(params.BLVoltages) {
			vbl = params.BLVoltages[c]
		}
		fmt.Fprintf(&b, "VBL_SRC_%d bl_src_%d 0 %.9g\n", c, c, vbl)
		fmt.Fprintf(&b, "RBL_DRV_%d bl_src_%d bl_0_%d {RBLDRV}\n", c, c, c)
	}
	b.WriteString("\n")

	b.WriteString("* WL wire resistances\n")
	for r := 0; r < rows; r++ {
		for c := 0; c < cols-1; c++ {
			fmt.Fprintf(&b, "RWL_%d_%d wl_%d_%d wl_%d_%d {RWL}\n", r, c, r, c, r, c+1)
		}
	}
	b.WriteString("\n* BL wire resistances\n")
	for c := 0; c < cols; c++ {
		for r := 0; r < rows-1; r++ {
			fmt.Fprintf(&b, "RBL_%d_%d bl_%d_%d bl_%d_%d {RBL}\n", r, c, r, c, r+1, c)
		}
	}

	b.WriteString("\n* Memory cell conductances\n")
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			g := params.Conductance[r][c]
			res := 1e15
			if g > 0 {
				res = 1.0 / g
			}
			fmt.Fprintf(&b, "RCELL_%d_%d wl_%d_%d bl_%d_%d %.9g\n", r, c, r, c, r, c, res)
		}
	}

	b.WriteString("\n* Readout peripherals per BL\n")
	for c := 0; c < cols; c++ {
		fmt.Fprintf(&b, "XSH_%d bl_%d_%d bl_sh_%d sh_clk 0 SAMPLE_HOLD\n", c, rows-1, c, c)
		fmt.Fprintf(&b, "XTIA_%d bl_sh_%d vout_%d 0 TIA_BASIC\n", c, c, c)
		fmt.Fprintf(&b, "XADC_%d vout_%d code_%d 0 ADC5\n", c, c, c)
	}

	b.WriteString("\n.control\n")
	b.WriteString("op\n")
	b.WriteString("print all\n")
	b.WriteString(".endc\n\n.end\n")

	return b.String(), nil
}
