package main

import "fecim-lattice-tools/internal/gogpucommand"

func main() {
	gogpucommand.Exit(
		"cmd/write-proof",
		"CGO_ENABLED=0 go run ./cmd/fecim-screenshotter -only circuits",
	)
}
