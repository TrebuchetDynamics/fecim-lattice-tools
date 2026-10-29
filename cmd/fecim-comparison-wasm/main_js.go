//go:build js && wasm

package main

import comparison "fecim-lattice-tools/workbench/modules/comparison"

func main() { comparison.ServeJS() }
