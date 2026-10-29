package circuits

import (
	"fecim-lattice-tools/workbench/moduleapi"
	webcircuits "fecim-lattice-tools/workbench/modules/circuits"
)

// WebPort supplies the real material/device workflow, not the legacy synthetic
// compatibility snapshot. Hosted builds construct their own bounded CPU port.
func (m *Module) WebPort() moduleapi.Port { return webcircuits.New(128) }
