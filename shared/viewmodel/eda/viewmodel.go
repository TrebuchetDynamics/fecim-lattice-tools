package eda

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/shared/viewmodel"
	"fecim-lattice-tools/shared/viewmodel/design"
	"fecim-lattice-tools/workbench/moduleapi"
	webeda "fecim-lattice-tools/workbench/modules/eda"
)

type Module struct{ state EDAState }

func New() *Module {
	return &Module{state: newDesignGenerationWorkflow(8, 8).buildState()}
}

func truncateContent(s string, maxLines int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n... (%d more lines)", len(lines)-maxLines)
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleEDA,
		Title:       "FeCIM EDA Design Suite",
		Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
		Status:      viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "generate_all", "generate_spice":
		// Legacy compatibility regenerates the actual in-memory example; this is
		// not a filesystem export or an installed-tool validation claim.
		m.state = newDesignGenerationWorkflow(m.state.ArrayRows, m.state.ArrayCols).buildState()
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}

// WebPort self-supplies the native adapter without a shared factory/route edit.
func (m *Module) WebPort() moduleapi.Port { return webeda.New() }

func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) DesignState() design.ModuleDesignState {
	return design.ModuleDesignState{ProcessNode: m.state.ProcessNode, DesignName: m.state.DesignName}
}
