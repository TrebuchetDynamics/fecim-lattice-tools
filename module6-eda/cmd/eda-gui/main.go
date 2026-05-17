package edagui

import "fecim-lattice-tools/internal/gogpucommand"

func Run(args []string) error {
	return gogpucommand.Error(
		"module6-eda/cmd/eda-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module eda",
	)
}
