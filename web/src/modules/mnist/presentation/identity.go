//go:build ignore

// Build-time capability metadata only; no hosted MNIST inference engine.
package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/modules/mnist"
	"os"
)

func main() {
	json.NewEncoder(os.Stdout).Encode(map[string]any{"identity": hysteresis.BridgeIdentity(), "descriptor": mnist.Descriptor()})
}
