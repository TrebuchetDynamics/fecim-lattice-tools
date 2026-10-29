//go:build ignore

package main

import (
	"encoding/json"
	comparison "fecim-lattice-tools/workbench/modules/comparison"
	"os"
)

func main() {
	json.NewEncoder(os.Stdout).Encode(struct {
		Identity   any `json:"identity"`
		Descriptor any `json:"descriptor"`
	}{comparison.Identity(), comparison.New().Descriptor()})
}
