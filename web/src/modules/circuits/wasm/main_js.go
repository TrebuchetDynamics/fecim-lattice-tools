//go:build js && wasm

package main

import "fecim-lattice-tools/workbench/modules/circuits/wasm"

func main() { wasm.Run() }
