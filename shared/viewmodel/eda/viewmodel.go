package eda

import "fecim-lattice-tools/shared/viewmodel"

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
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
