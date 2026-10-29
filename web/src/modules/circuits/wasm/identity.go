//go:build ignore

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/modules/circuits"
	"os"
)

func main() {
	p := circuits.New(8)
	defer p.Close()
	json.NewEncoder(os.Stdout).Encode(map[string]any{"identity": circuits.Identity(), "descriptor": p.Descriptor()})
}
