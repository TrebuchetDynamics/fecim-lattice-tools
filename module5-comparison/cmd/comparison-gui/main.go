package comparisongui

import "fecim-lattice-tools/internal/legacycommand"

func Run(args []string) error {
	return legacycommand.Error(
		"module5-comparison/cmd/comparison-gui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module comparison",
		"go run ./module5-comparison/cmd/comparison-gui-fyne",
	)
}
