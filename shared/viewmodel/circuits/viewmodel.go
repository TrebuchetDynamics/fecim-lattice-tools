package circuits

import "fecim-lattice-tools/shared/viewmodel"

type Module struct{ state CircuitsState }

func New() *Module {
	return &Module{state: CircuitsState{
		ADCResolution: 5, DACResolution: 5, TIAGain: 1e4,
		ChargePumpStages: 4, SupplyVoltage: 1.8, ISPPEnabled: true,
	}}
}
func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleCircuits, Title: "FeCIM Peripheral Circuits Visualizer",
		Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
		Status: viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "run_read":
		return nil
	case "run_write":
		return nil
	case "toggle_ispp":
		m.state.ISPPEnabled = !m.state.ISPPEnabled
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
