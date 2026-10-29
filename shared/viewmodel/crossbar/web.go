package crossbar

import (
	"fecim-lattice-tools/workbench/moduleapi"
	web "fecim-lattice-tools/workbench/modules/crossbar"
)

// WebPort self-supplies through the existing native factory capability check.
// The legacy shell retains its own state and does not start background work.
func (m *Module) WebPort() moduleapi.Port { return web.New(128) }
