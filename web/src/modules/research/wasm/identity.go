//go:build ignore

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/moduleapi/research"
	"os"
)

func main() { json.NewEncoder(os.Stdout).Encode(research.Identity()) }
