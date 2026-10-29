//go:build ignore

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/modules/crossbar"
	"os"
)

func main() {
	p := crossbar.New(32)
	defer p.Close()
	json.NewEncoder(os.Stdout).Encode(map[string]any{"identity": crossbar.Identity(), "descriptor": p.Descriptor()})
}
