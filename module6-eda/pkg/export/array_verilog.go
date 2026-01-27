// Structural Verilog netlist generator for FeCIM crossbar arrays
//
// References:
// [1] IEEE Std 1364-2005 - Verilog HDL Structural Modeling
//
// This generates a STRUCTURAL netlist (instantiation list) of FeCIM bitcells.
// The bitcell itself uses a placeholder behavioral model (see cell_verilog.go).
package export

import (
	"fmt"
	"strings"
	"time"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GenerateArrayVerilog generates a structural Verilog netlist for a FeCIM crossbar array
// This instantiates the FeCIM bitcells in a grid pattern with WL/BL connections
// Format: Verilog HDL Structural [Ref 1]
// Supports both passive and 1T1R architectures:
//   - passive: WL[], BL[] ports (sneak path susceptible)
//   - 1t1r: WL[], BL[], SL[] ports (sneak path mitigated via select transistor)
func GenerateArrayVerilog(cfg config.ArrayConfig) string {
	var sb strings.Builder

	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	is1T1R := cfg.Architecture == "1t1r"

	// Determine cell name based on architecture
	cellName := "fecim_bitcell"
	if is1T1R {
		cellName = "fecim_1t1r_bitcell"
	}

	// Header with metadata
	sb.WriteString(fmt.Sprintf(`// FeCIM Crossbar Array - Auto-generated
// Date: %s
// Rows: %d, Cols: %d
// Mode: %s
// Architecture: %s
// NOTE: Cell is placeholder. Real behavior requires FeFET model.
`,
		time.Now().Format("2006-01-02"),
		cfg.Rows, cfg.Cols, cfg.Mode, cfg.Architecture))

	if is1T1R {
		sb.WriteString("// 1T1R: SL (Source Lines) connect to transistor source for sneak path mitigation\n")
	}
	sb.WriteString("\n")

	// Module declaration with ports
	sb.WriteString(fmt.Sprintf("module %s (\n", designName))
	sb.WriteString(fmt.Sprintf("    input  wire [%d:0] WL,    // Word Lines (row select)\n", cfg.Rows-1))
	sb.WriteString(fmt.Sprintf("    output wire [%d:0] BL,    // Bit Lines (column data)\n", cfg.Cols-1))
	if is1T1R {
		sb.WriteString(fmt.Sprintf("    input  wire [%d:0] SL,    // Source Lines (1T1R: transistor source, one per column)\n", cfg.Cols-1))
	}
	sb.WriteString("    inout  wire VPWR,         // Power\n")
	sb.WriteString("    inout  wire VGND          // Ground\n")
	sb.WriteString(");\n\n")

	sb.WriteString("// Cell instantiations\n")

	// Generate cell instances in row-major order
	for row := 0; row < cfg.Rows; row++ {
		for col := 0; col < cfg.Cols; col++ {
			sb.WriteString(fmt.Sprintf("%s cell_%d_%d (\n", cellName, row, col))
			sb.WriteString(fmt.Sprintf("    .WL(WL[%d]),\n", row))
			sb.WriteString(fmt.Sprintf("    .BL(BL[%d]),\n", col))
			if is1T1R {
				sb.WriteString(fmt.Sprintf("    .SL(SL[%d]),\n", col))
			}
			sb.WriteString("    .VPWR(VPWR),\n")
			sb.WriteString("    .VGND(VGND)\n")
			sb.WriteString(");\n\n")
		}
	}

	sb.WriteString("endmodule\n")
	return sb.String()
}
