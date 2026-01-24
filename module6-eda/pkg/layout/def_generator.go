// pkg/layout/def_generator.go
package layout

import (
	"fmt"
	"strings"
)

// GenerateDEF creates a placement file for the array
func GenerateDEF(appName string, rows, cols int, cellName string, pitchX, pitchY int) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("VERSION 5.8 ;\n"))
	sb.WriteString(fmt.Sprintf("DESIGN %s ;\n", appName))
	sb.WriteString(fmt.Sprintf("UNITS DISTANCE MICRONS 1000 ;\n"))
	
	// Die Area calculation (pitch is in microns * 1000 for DEF units? usually microns, and UNITS sets the scale)
	// OpenLane usually expects DEF units to be congruent with the LEF. 
	// Standard DEF: UNITS DISTANCE MICRONS 1000; means 1000 units = 1 micron.
	// So pitchX and pitchY should be in database units (e.g. 460 for 0.46um)
	
	width := cols * pitchX
	height := rows * pitchY
	sb.WriteString(fmt.Sprintf("DIEAREA ( 0 0 ) ( %d %d ) ;\n", width, height))
	sb.WriteString("\n")

	// Components (The FeFET Cells)
	numCells := rows * cols
	sb.WriteString(fmt.Sprintf("COMPONENTS %d ;\n", numCells))

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			// Name: cell_row_col
			instName := fmt.Sprintf("cell_%d_%d", r, c)
			
			// Position: x = col * pitch, y = row * pitch
			posX := c * pitchX
			posY := r * pitchY
			
			// FIXED means regular placement/routing tools won't move it
			// N = North orientation (default)
			sb.WriteString(fmt.Sprintf("- %s %s + FIXED ( %d %d ) N ;\n",
				instName, cellName, posX, posY))
		}
	}
	sb.WriteString("END COMPONENTS\n")
	sb.WriteString("\n")

    // Nets would be defined here if we were doing detailed routing
	// But since we rely on OpenROAD to route based on logical connectivity,
	// we technically only need the COMPONENTS placement for the 'PL_SKIP_INITIAL_PLACEMENT' strategy
	// coupled with the Verilog netlist which defines the connectivity.
	// OpenROAD will then route the nets defined in Verilog between the pins defined in the LEF
	// at the locations defined in this DEF.
    
	sb.WriteString("END DESIGN\n")
	return sb.String()
}
