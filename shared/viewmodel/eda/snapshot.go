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
		{ID: "spice_export", Title: "SPICE Netlist", Body: fmt.Sprintf("Netlist for %d×%d FeCIM crossbar with FeFET compact model.", state.ArrayRows, state.ArrayCols), Category: "research"},
		{ID: "verilog_export", Title: "Verilog Module", Body: "Behavioral model for digital control logic (WL decoder, BL multiplexer, read/write FSM).", Category: "research"},
		{ID: "liberty_export", Title: "Liberty Timing", Body: fmt.Sprintf("Timing and power for %s process at TT/FF/SS corners.", state.ProcessNode), Category: "research"},
		{ID: "physical_export", Title: "Physical Design (DEF/LEF)", Body: fmt.Sprintf("LEF macro for %d×%d array with placed cells and routed interconnect.", state.ArrayRows, state.ArrayCols), Category: "research"},
	}
	sections = append(sections, viewmodel.Section{
		ID: "edu_spice", Title: "What is SPICE?",
		Body: "SPICE (Simulation Program with Integrated Circuit Emphasis) describes circuits as netlists — text files listing components and how they connect. SPICE simulators solve Kirchhoff's laws numerically. Our export generates a netlist with FeFET compact models.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "edu_flow", Title: "Design Flow",
		Body: fmt.Sprintf("RTL → Synthesis → Floorplan → Place → Route → DRC/LVS → GDSII. Our %s flow targets OpenLane for open-source tapeout. SPICE for circuit simulation, Verilog for digital control, Liberty for timing closure.", state.ProcessNode),
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "research_validation", Title: "Validation Status",
		Body: "Generated SPICE netlists validated with ngspice roundtrip tests. Netlist syntax checked against Verilog-A compact model. Liberty corners verified (TT/FF/SS). All exports are educational baselines — not validated against silicon measurements.",
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID: "design_workflow", Title: "Design Composition",
		Body: "Design wizard: Start with material selection (Module 1) → Array configuration (Module 2) → Peripheral specs (Module 4) → Export (here). Use parameter sweep to generate netlist families across voltage corners. All formats exportable in one step.",
		Category: "design",
	})
	actions := []viewmodel.Action{
		{ID: "generate_all", Label: "Export All Formats", Kind: viewmodel.ActionCommand},
		{ID: "generate_spice", Label: "Generate SPICE", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleEDA, Title: "FeCIM EDA Design Suite",
			Description:    "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
			Status:         viewmodel.StatusFunctional,
			BoundaryNotice: "EDUCATIONAL EDA — Not a production chip design tool. Generated netlists and layouts are educational examples. Does not use proprietary foundry PDKs.",
		},
		Metrics: metrics, Sections: sections, Actions: actions,
	}
}
