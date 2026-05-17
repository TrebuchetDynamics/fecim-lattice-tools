package main

import "fecim-lattice-tools/internal/legacycommand"

func main() {
	legacycommand.Exit(
		"cmd/write-proof",
		"CGO_ENABLED=0 go run ./cmd/fecim-screenshotter -only circuits",
		"go run ./cmd/write-proof-fyne",
	)
}
