// Demo 2 GUI: Crossbar Array Visualization with Fyne
//
// This provides an interactive GUI for visualizing matrix-vector multiplication
// operations on a simulated ferroelectric crossbar array.
//
// Features:
// - Interactive heatmap visualization of conductance states
// - IR drop analysis with heatmap overlay
// - Sneak path current analysis
// - Real-time MVM operations with full physics
// - 30 discrete FeCIM levels (4.9 bits/cell, conference claim baseline)
//
// Standard Mode:
//
//	go run ./cmd/fecim-lattice-tools crossbar
//
// Enhanced Mode (all features):
//
//	go run ./cmd/fecim-lattice-tools crossbar -enhanced
//
// Terminal Inference (CLI):
//
//	go run ./cmd/fecim-lattice-tools crossbar inference [options]
//
// Enhanced features include:
// - Color legends for all heatmaps
// - Live metrics panel (accuracy, energy, performance)
// - Before/after comparison view
// - Accuracy waterfall chart
// - Energy comparison badges
// - Enhanced MVM with integrated non-idealities
// - Data export (CSV, JSON)
package crossbarcmd

import (
	"flag"
	"fmt"
	"os"

	"fecim-lattice-tools/module2-crossbar/pkg/gui"
)

func RunGUI(args []string) error {
	fs := flag.NewFlagSet("crossbar-gui", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	enhanced := fs.Bool("enhanced", false, "Enable enhanced UI with all features")
	help := fs.Bool("help", false, "Show help")
	helpShort := fs.Bool("h", false, "Show help (shorthand)")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Crossbar Array Visualization")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools crossbar [options]")
		fmt.Fprintln(out, "  fecim-lattice-tools crossbar gui [options]")
		fmt.Fprintln(out, "  fecim-lattice-tools crossbar inference [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, "  -enhanced    Enable enhanced layout")
		fmt.Fprintln(out, "  -help        Show this help message")
		fmt.Fprintln(out, "  -h           Show this help message (shorthand)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Inference:")
		fmt.Fprintln(out, "  Use: fecim-lattice-tools crossbar inference -help")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Implemented GUI capabilities:")
		fmt.Fprintln(out, "  • Conductance heatmap with hover/click inspection")
		fmt.Fprintln(out, "  • IR-drop analysis tab")
		fmt.Fprintln(out, "  • Sneak-path analysis tab")
		fmt.Fprintln(out, "  • Accuracy/energy metrics and comparison widgets")
		fmt.Fprintln(out, "  • Data export (weights CSV and analysis JSON)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Device model defaults:")
		fmt.Fprintln(out, "  • 30 discrete FeCIM levels (demo baseline)")
		fmt.Fprintln(out, "  • DAC/ADC quantization with configurable bit depth")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(fs.Output(), "Error:", err)
		fs.Usage()
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *help || *helpShort {
		fs.Usage()
		return nil
	}

	app, err := gui.NewCrossbarApp()
	if err != nil {
		return fmt.Errorf("failed to initialize crossbar app: %w", err)
	}

	if *enhanced {
		fmt.Println("Starting FeCIM Crossbar Visualizer (Enhanced Mode)")
		fmt.Println("→ All features enabled")
		app.RunEnhanced()
	} else {
		fmt.Println("Starting FeCIM Crossbar Visualizer (Standard Mode)")
		fmt.Println("→ Run with -enhanced for all features")
		app.Run()
	}

	return nil
}
