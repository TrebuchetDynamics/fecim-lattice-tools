package main

import "fecim-lattice-tools/internal/legacycommand"

func main() {
	legacycommand.Exit(
		"cmd/fecim-web",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module docs",
		"go run ./cmd/fecim-web-fyne",
	)
}
