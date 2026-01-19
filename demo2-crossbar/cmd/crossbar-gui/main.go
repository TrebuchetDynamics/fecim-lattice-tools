// Demo 2 GUI: Crossbar Array Visualization with Fyne
//
// This provides an interactive GUI for visualizing matrix-vector multiplication
// operations on a simulated ferroelectric crossbar array.
//
// Features:
// - Interactive heatmap visualization of conductance states
// - IR drop analysis with heatmap overlay
// - Sneak path current analysis
// - Real-time MVM operations
// - 30 discrete IronLattice levels
//
// Run with: go run ./cmd/crossbar-gui
package main

import (
	"ironlattice-vis/demo2-crossbar/pkg/gui"
)

func main() {
	app := gui.NewCrossbarApp()
	app.Run()
}
