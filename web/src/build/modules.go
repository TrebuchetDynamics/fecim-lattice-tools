//go:build ignore

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/moduleapi"
	"os"
)

func main() { json.NewEncoder(os.Stdout).Encode(moduleapi.BrowserCatalogue()) }
