package main

import "fecim-lattice-tools/internal/gogpucommand"

func main() {
	gogpucommand.Exit(
		"cmd/demo-frames",
		"CGO_ENABLED=0 go run ./cmd/fecim-screenshotter -only docs",
	)
}
