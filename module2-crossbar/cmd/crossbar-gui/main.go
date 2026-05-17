package crossbarcmd

import "fecim-lattice-tools/internal/gogpucommand"

func RunGUI(args []string) error {
	return gogpucommand.Error(
		"module2-crossbar/cmd/crossbar-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module crossbar",
	)
}
