package main

import "fecim-lattice-tools/internal/legacycommand"

func main() {
	legacycommand.Exit(
		"cmd/demo-frames",
		"CGO_ENABLED=0 go run ./cmd/fecim-screenshotter -only docs",
		"go run ./cmd/demo-frames-fyne",
	)
}
