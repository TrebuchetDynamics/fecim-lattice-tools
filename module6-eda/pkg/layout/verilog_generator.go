// pkg/layout/verilog_generator.go
package layout

import (
	"fmt"
	"strings"
)

func GenerateVerilog(moduleName string, rows, cols int, cellName string) string {
	var sb strings.Builder

	// Module definition
	sb.WriteString(fmt.Sprintf("module %s (\n", moduleName))

	// Ports
	sb.WriteString(fmt.Sprintf("    input [%d:0] wl,\n", rows-1))
	sb.WriteString(fmt.Sprintf("    output [%d:0] bl\n", cols-1))
	sb.WriteString(");\n\n")

	// Cell instantiations
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			instName := fmt.Sprintf("cell_%d_%d", r, c)

			// Connect WL to row, BL to column
			// Assuming cell has ports .WL and .BL
			sb.WriteString(fmt.Sprintf("    %s %s (\n", cellName, instName))
			sb.WriteString(fmt.Sprintf("        .WL(wl[%d]),\n", r))
			sb.WriteString(fmt.Sprintf("        .BL(bl[%d])\n", c))
			// Comma management for last item?
			// Standard Verilog instantiations end with ); so no comma issue between ports usually
			sb.WriteString("    );\n")
		}
	}

	sb.WriteString("endmodule\n")
	return sb.String()
}
