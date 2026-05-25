package main

import (
	"fmt"
	"io"
	"os"

	"fecim-lattice-tools/internal/gogpuscreenshot"
)

func main() {
	os.Exit(runScreenshotter(os.Args[1:], os.Stderr))
}

func runScreenshotter(args []string, stderr io.Writer) int {
	if err := gogpuscreenshot.Run(args); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}
