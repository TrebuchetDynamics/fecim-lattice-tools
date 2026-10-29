//go:build ignore

// Build-time metadata only. Hosted EDA has no scientific WASM engine.
package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/modules/eda"
	"os"
)

func main() {
	json.NewEncoder(os.Stdout).Encode(map[string]any{"identity": hysteresis.BridgeIdentity(), "descriptor": eda.Descriptor()})
}
