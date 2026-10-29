//go:build js && wasm

package main

import "fecim-lattice-tools/workbench/modules/crossbar/wasm"

func main() { wasm.Run() }
