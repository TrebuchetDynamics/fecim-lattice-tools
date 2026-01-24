// pkg/export/layout_wrapper.go
package export

import (
	"os"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/compiler"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/layout"
)

// ExportDEF generates and writes the DEF placement file
func ExportDEF(design *compiler.ArrayDesign, path string) error {
	rows := design.Config.ArrayRows
	cols := design.Config.ArrayCols
	
	// Default to SKY130 dimensions (units: 1000 = 1um)
	// Cell pitch 0.46um, Height 2.72um standard cell
	pitchX := 460 
	pitchY := 2720 
	
	if design.Config.Architecture == compiler.Arch1T1R {
		pitchX = 920 // Larger pitch for 1T1R
	}

	// Adjust for different technologies if needed
	if design.Config.Technology == compiler.TechGF180 {
		pitchX = 560
		pitchY = 3200
	}

	content := layout.GenerateDEF(design.Config.Name, rows, cols, "fecim_bitcell", pitchX, pitchY)
	return os.WriteFile(path, []byte(content), 0644)
}

// ExportVerilog generates and writes the structural Verilog netlist
func ExportVerilog(design *compiler.ArrayDesign, path string) error {
	content := layout.GenerateVerilog(design.Config.Name, design.Config.ArrayRows, design.Config.ArrayCols, "fecim_bitcell")
	return os.WriteFile(path, []byte(content), 0644)
}
