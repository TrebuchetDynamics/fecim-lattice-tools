package eda

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct{ state EDAState }

func New() *Module {
	return &Module{state: EDAState{
		DesignName:  "fecim_crossbar_8x8",
		ProcessNode: "sky130",
		ArrayRows:   8, ArrayCols: 8,
		ExportFormats: []string{"spice", "verilog", "liberty", "def", "lef"},
	}}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleEDA, Title: "FeCIM EDA Design Suite",
		Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
		Status: viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "generate_all":
		return nil
	case "generate_spice":
		return nil
	case "set_design_name":
		if name, ok := action.Payload["name"]; ok {
			m.state.DesignName = name
			return nil
		}
		return fmt.Errorf("eda: design name required")
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
