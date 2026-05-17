package circuitsgui

import "fecim-lattice-tools/internal/gogpucommand"

func Run(args []string) error {
	return gogpucommand.Error(
		"module4-circuits/cmd/circuits-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module circuits",
	)
}
