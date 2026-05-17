package mnistgui

import "fecim-lattice-tools/internal/legacycommand"

func Run(args []string) error {
	return legacycommand.Error(
		"module3-mnist/cmd/mnist-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module mnist",
		"go run ./module3-mnist/cmd/mnist-gui-fyne",
	)
}
