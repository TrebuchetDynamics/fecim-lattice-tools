package eda

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state EDAState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "design", Label: "Design", Value: state.DesignName},
		{ID: "process", Label: "Process Node", Value: state.ProcessNode},
		{ID: "spice", Label: "SPICE", Value: "ready"},
		{ID: "verilog", Label: "Verilog", Value: "ready"},
		{ID: "liberty", Label: "Liberty", Value: "ready"},
		{ID: "def", Label: "DEF", Value: "ready"},
		{ID: "lef", Label: "LEF", Value: "ready"},
	}
	sections := []viewmodel.Section{
		{ID: "spice_export", Title: "SPICE Netlist", Body: fmt.Sprintf("Netlist for %d×%d FeCIM crossbar with FeFET compact model.", state.ArrayRows, state.ArrayCols)},
		{ID: "verilog_export", Title: "Verilog Module", Body: "Behavioral model for digital control logic (WL decoder, BL multiplexer, read/write FSM)."},
		{ID: "liberty_export", Title: "Liberty Timing", Body: fmt.Sprintf("Timing and power for %s process at TT/FF/SS corners.", state.ProcessNode)},
		{ID: "physical_export", Title: "Physical Design (DEF/LEF)", Body: fmt.Sprintf("LEF macro for %d×%d array with placed cells and routed interconnect.", state.ArrayRows, state.ArrayCols)},
	}
	actions := []viewmodel.Action{
		{ID: "generate_all", Label: "Export All Formats", Kind: viewmodel.ActionCommand},
		{ID: "generate_spice", Label: "Generate SPICE", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleEDA, Title: "FeCIM EDA Design Suite",
			Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
			Status: viewmodel.StatusFunctional,
		},
		Metrics: metrics, Sections: sections, Actions: actions,
	}
}
