package mnistgui

import "fecim-lattice-tools/internal/gogpucommand"

func Run(args []string) error {
	return gogpucommand.Error(
		"module3-mnist/cmd/mnist-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module mnist",
	)
}
