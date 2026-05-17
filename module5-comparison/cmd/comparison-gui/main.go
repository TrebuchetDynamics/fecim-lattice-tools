package comparisongui

import "fecim-lattice-tools/internal/gogpucommand"

func Run(args []string) error {
	return gogpucommand.Error(
		"module5-comparison/cmd/comparison-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module comparison",
	)
}
