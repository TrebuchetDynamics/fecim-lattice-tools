package edagui

import "fecim-lattice-tools/internal/legacycommand"

func Run(args []string) error {
	return legacycommand.Error(
		"module6-eda/cmd/eda-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module eda",
		"go run ./module6-eda/cmd/eda-gui-fyne",
	)
}
