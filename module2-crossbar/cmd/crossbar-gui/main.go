package crossbarcmd

import "fecim-lattice-tools/internal/legacycommand"

func RunGUI(args []string) error {
	return legacycommand.Error(
		"module2-crossbar/cmd/crossbar-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module crossbar",
		"go run ./module2-crossbar/cmd/crossbar-gui-fyne",
	)
}
