package circuitsgui

import "fecim-lattice-tools/internal/legacycommand"

func Run(args []string) error {
	return legacycommand.Error(
		"module4-circuits/cmd/circuits-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module circuits",
		"go run ./module4-circuits/cmd/circuits-gui-fyne",
	)
}
