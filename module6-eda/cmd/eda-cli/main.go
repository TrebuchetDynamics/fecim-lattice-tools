// cmd/eda-cli/main.go
// CLI tool for FeCIM array design generation and export
//
// Supports three operation modes:
//   - storage: High-density non-volatile storage (NAND replacement)
//   - memory:  High-speed zero-refresh memory (DRAM replacement)
//   - compute: Analog compute-in-memory for AI inference
//
// For storage and memory modes, no input file is required.
// For compute mode, weights are optional - omit -input for unprogrammed arrays.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/compiler"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/export"
)

// WeightsFile represents a JSON file containing neural network weights
// Only used in compute mode when pre-programming with trained weights
type WeightsFile struct {
	Name    string      `json:"name"`
	Rows    int         `json:"rows"`
	Cols    int         `json:"cols"`
	Weights [][]float64 `json:"weights"`
}

func main() {
	// Operation mode
	mode := flag.String("mode", "compute", "Operation mode: storage, memory, or compute")

	// Input (optional for compute mode, ignored for others)
	inputFile := flag.String("input", "", "Input weights JSON file (optional, compute mode only)")

	// Output
	outputDir := flag.String("output", ".", "Output directory")
	designName := flag.String("name", "fecim_array", "Design name for output files")

	// Array parameters
	rows := flag.Int("rows", 128, "Array rows")
	cols := flag.Int("cols", 128, "Array cols")
	levels := flag.Int("levels", 30, "Conductance levels (2-30)")

	// Technology selection
	tech := flag.String("tech", "SKY130", "Technology: SKY130, GF180MCU, IHP_SG13G2")
	arch := flag.String("arch", "passive", "Architecture: passive or 1T1R")

	// Electrical parameters
	vdd := flag.Float64("vdd", 1.8, "Supply voltage (V)")
	gmin := flag.Float64("gmin", 1.0, "Min conductance (μS)")
	gmax := flag.Float64("gmax", 100.0, "Max conductance (μS)")

	// Export options
	exportJSON := flag.Bool("json", true, "Export JSON mapping")
	exportCSV := flag.Bool("csv", true, "Export CSV cell assignments")
	exportSPICE := flag.Bool("spice", true, "Export SPICE netlist")
	exportVerilog := flag.Bool("verilog", true, "Export Verilog netlist")
	exportDEF := flag.Bool("def", true, "Export DEF placement")

	flag.Parse()

	// Parse operation mode
	var opMode compiler.OperationMode
	switch strings.ToLower(*mode) {
	case "storage":
		opMode = compiler.ModeStorage
	case "memory":
		opMode = compiler.ModeMemory
	case "compute":
		opMode = compiler.ModeCompute
	default:
		fmt.Printf("Error: unknown mode '%s'. Use: storage, memory, or compute\n", *mode)
		os.Exit(1)
	}

	fmt.Printf("FeCIM Array Generator - %s Mode\n", strings.Title(*mode))
	fmt.Printf("========================================\n\n")

	// Create configuration
	config := compiler.NewArrayConfig(opMode, *rows, *cols)
	config.Name = *designName
	config.Technology = *tech
	config.Levels = *levels
	config.GMin = *gmin
	config.GMax = *gmax
	config.Peripherals.VDD = *vdd

	// Handle architecture
	if strings.ToLower(*arch) == "1t1r" {
		config.With1T1R()
	}

	// Load weights for compute mode (if provided)
	if opMode == compiler.ModeCompute && *inputFile != "" {
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Printf("Error reading weights file: %v\n", err)
			os.Exit(1)
		}

		var wf WeightsFile
		if err := json.Unmarshal(data, &wf); err != nil {
			fmt.Printf("Error parsing weights JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Loaded weights: %s (%dx%d = %d weights)\n",
			wf.Name, len(wf.Weights), len(wf.Weights[0]),
			len(wf.Weights)*len(wf.Weights[0]))

		config.ComputeConfig.InitialWeights = wf.Weights
	}

	// Print configuration
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Mode:         %s\n", config.Mode)
	fmt.Printf("  Array Size:   %d × %d (%d cells)\n", config.ArrayRows, config.ArrayCols, config.ArrayRows*config.ArrayCols)
	fmt.Printf("  Technology:   %s\n", config.Technology)
	fmt.Printf("  Architecture: %s\n", config.Architecture)
	fmt.Printf("  Levels:       %d (%.2f bits/cell)\n", config.Levels, float64(config.Levels)/6.0)
	fmt.Printf("  Conductance:  %.1f - %.1f μS\n", config.GMin, config.GMax)
	if opMode == compiler.ModeCompute && config.ComputeConfig.InitialWeights != nil {
		fmt.Printf("  Weights:      %dx%d loaded\n",
			len(config.ComputeConfig.InitialWeights),
			len(config.ComputeConfig.InitialWeights[0]))
	} else if opMode == compiler.ModeCompute {
		fmt.Printf("  Weights:      None (unprogrammed array)\n")
	}
	fmt.Println()

	// Generate design
	design, err := compiler.GenerateDesign(config)
	if err != nil {
		fmt.Printf("Design generation error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Design Statistics:\n")
	fmt.Printf("  Total Cells:  %d\n", design.Stats.TotalCells)
	fmt.Printf("  Active Cells: %d\n", design.Stats.ActiveCells)
	fmt.Printf("  Area:         %.4f mm²\n", design.Stats.AreaMM2)
	fmt.Printf("  Est. Power:   %.2f mW\n", design.Stats.PowerMW)

	if opMode == compiler.ModeCompute {
		fmt.Printf("  Throughput:   %.2f GOPS\n", design.Stats.ThroughputGOPS)
		if config.ComputeConfig.InitialWeights != nil {
			fmt.Printf("  Weight Range: [%.4f, %.4f]\n", design.Stats.WeightMin, design.Stats.WeightMax)
			fmt.Printf("  Quant PSNR:   %.2f dB\n", design.Stats.QuantPSNR)
		}
	}
	fmt.Println()

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Export files
	fmt.Printf("Exporting files to %s/\n", *outputDir)

	if *exportJSON {
		path := filepath.Join(*outputDir, *designName+"_design.json")
		if err := export.ExportJSON(design, path); err != nil {
			fmt.Printf("  JSON export error: %v\n", err)
		} else {
			fmt.Printf("  ✓ %s\n", path)
		}
	}

	if *exportCSV {
		path := filepath.Join(*outputDir, *designName+"_cells.csv")
		if err := export.ExportCSV(design, path); err != nil {
			fmt.Printf("  CSV export error: %v\n", err)
		} else {
			fmt.Printf("  ✓ %s\n", path)
		}
	}

	if *exportSPICE {
		path := filepath.Join(*outputDir, *designName+".sp")
		if err := export.ExportSPICE(design, path, *vdd); err != nil {
			fmt.Printf("  SPICE export error: %v\n", err)
		} else {
			fmt.Printf("  ✓ %s\n", path)
		}
	}

	if *exportVerilog {
		path := filepath.Join(*outputDir, *designName+".v")
		if err := export.ExportVerilog(design, path); err != nil {
			fmt.Printf("  Verilog export error: %v\n", err)
		} else {
			fmt.Printf("  ✓ %s\n", path)
		}
	}

	if *exportDEF {
		path := filepath.Join(*outputDir, *designName+".def")
		if err := export.ExportDEF(design, path); err != nil {
			fmt.Printf("  DEF export error: %v\n", err)
		} else {
			fmt.Printf("  ✓ %s\n", path)
		}
	}

	fmt.Println("\nDone!")
}
