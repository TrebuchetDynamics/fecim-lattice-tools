// cmd/lattice-gen/main.go
// CLI tool for generating FeCIM lattice Verilog and DEF files

package main

import (
	"flag"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/shared/logging"
)

func main() {
	// Initialize logger
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".fecim", "logs", "module6-eda-lattice-gen.log")
	if err := logging.Init("module6-eda-lattice-gen", logPath); err != nil {
		// Fallback to standard error if logger init fails
		os.Stderr.WriteString("Failed to initialize logging: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logging.CloseGlobal()

	// Enable logging by default
	logging.SetVerbosity(logging.VerbosityInfo)

	rows := flag.Int("rows", 4, "Number of rows")
	cols := flag.Int("cols", 4, "Number of columns")
	outputDir := flag.String("output", "output/lattices", "Output directory")
	flag.Parse()

	// Create output directory if needed
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logging.GlobalError("Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	// Generate files
	verilogPath, err := export.WriteLatticeVerilog(*rows, *cols, *outputDir)
	if err != nil {
		logging.GlobalError("Error writing Verilog: %v\n", err)
		os.Exit(1)
	}
	logging.Printf("Generated: %s\n", verilogPath)

	defPath, err := export.WriteLatticeDEF(*rows, *cols, *outputDir)
	if err != nil {
		logging.GlobalError("Error writing DEF: %v\n", err)
		os.Exit(1)
	}
	logging.Printf("Generated: %s\n", defPath)

	logging.Printf("\nLattice %dx%d generated successfully (%d cells)\n", *rows, *cols, (*rows)*(*cols))
}
