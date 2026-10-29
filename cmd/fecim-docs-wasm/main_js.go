//go:build js && wasm

package main

import "fecim-lattice-tools/workbench/modules/docs"

func main() { docs.InstallWASM() }
