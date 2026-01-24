// pkg/export/openlane_config.go
package export

import (
	"encoding/json"
	"fmt"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
)

// GenerateOpenLaneConfig generates an OpenLane config.json for the FeCIM crossbar design
// This configures OpenLane to use pre-placed DEF and custom cell libraries
func GenerateOpenLaneConfig(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	
	// Calculate die area with margins (in nanometers, converted to microns in DEF units)
	dieWidth := cfg.Cols*int(cfg.CellWidth*1000) + 2000   // Add 2μm margin
	dieHeight := cfg.Rows*int(cfg.CellHeight*1000) + 2000
	
	config := map[string]interface{}{
		"DESIGN_NAME": designName,
		"VERILOG_FILES": fmt.Sprintf("dir::output/%s.v", designName),
		"CLOCK_PORT": "",
		"CLOCK_PERIOD": 10.0,
		
		// PDK configuration
		"PDK": "sky130A",
		"STD_CELL_LIBRARY": "sky130_fd_sc_hd",
		
		// Custom cell integration
		"EXTRA_LEFS": "dir::cells/fecim_bitcell/fecim_bitcell.lef",
		"EXTRA_LIBS": "dir::cells/fecim_bitcell/fecim_bitcell.lib",
		"VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bitcell/fecim_bitcell.v",
		
		// Floorplan configuration
		"FP_SIZING": "absolute",
		"DIE_AREA": fmt.Sprintf("0 0 %.3f %.3f", float64(dieWidth)/1000.0, float64(dieHeight)/1000.0),
		"DESIGN_IS_CORE": 0,
		
		// Pre-placed placement strategy
		"PLACEMENT_CURRENT_DEF": fmt.Sprintf("dir::output/%s.def", designName),
		"PL_SKIP_INITIAL_PLACEMENT": 1,
		"PL_TARGET_DENSITY": 0.6,
		
		// Skip CTS for this macro
		"RUN_CTS": 0,
		
		// Synthesis configuration
		"SYNTH_ELABORATE_ONLY": 1,
	}
	
	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}
