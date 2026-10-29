//go:build ignore

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/modules/docs"
	"os"
)

func main() {
	json.NewEncoder(os.Stdout).Encode(map[string]any{"identity": docs.Identity(), "descriptor": docs.Descriptor()})
}
