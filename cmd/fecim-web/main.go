package main

import "fecim-lattice-tools/internal/gogpucommand"

func main() {
	gogpucommand.Exit(
		"cmd/fecim-web",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module docs",
	)
}
