package mnist

import (
	"errors"
	"fmt"

	"fecim-lattice-tools/shared/neural"
	"fecim-lattice-tools/shared/viewmodel"
	"fecim-lattice-tools/workbench/moduleapi"
	webmnist "fecim-lattice-tools/workbench/modules/mnist"
)

type Module struct{ state MNISTState }

func New() *Module                                       { return &Module{state: MNISTState{NumLevels: 30}} }
func (m *Module) WebPort() moduleapi.Port                { return webmnist.New() }
func (m *Module) Descriptor() viewmodel.ModuleDescriptor { return buildSnapshot(m.state).Descriptor }
func (m *Module) Snapshot() viewmodel.ModuleSnapshot     { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "run_inference":
		return errors.New("MNIST inference unavailable in the legacy summary: select real model/data/input in the local web module")
	case "sweep_levels":
		levels, e := viewmodel.PayloadInt(action.Payload, "levels")
		if e != nil {
			return fmt.Errorf("mnist: %w", e)
		}
		if levels < 2 || levels > neural.MaxDemoLevels {
			return errors.New("mnist: levels outside simulation bounds")
		}
		m.state.NumLevels = levels
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start() {}
func (m *Module) Stop()  {}
