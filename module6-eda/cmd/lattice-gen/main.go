// cmd/lattice-gen/main.go
// CLI tool for generating FeCIM lattice Verilog and DEF files

package main

import (
	"flag"
	"fmt"
	"os"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/export"
)

func main() {
	rows := flag.Int("rows", 4, "Number of rows")
	cols := flag.Int("cols", 4, "Number of columns")
	outputDir := flag.String("output", ".", "Output directory")
	flag.Parse()

	// Create output directory if needed
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	// Generate files
	verilogPath, err := export.WriteLatticeVerilog(*rows, *cols, *outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Verilog: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s\n", verilogPath)

	defPath, err := export.WriteLatticeDEF(*rows, *cols, *outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing DEF: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s\n", defPath)

	fmt.Printf("\nLattice %dx%d generated successfully (%d cells)\n", *rows, *cols, (*rows)*(*cols))
}
