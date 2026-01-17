// Command phasefield runs a GPU-accelerated TDGL phase-field simulation
// of ferroelectric domain evolution.
//
// This is Demo 3 of the IronLattice Visualizer project.
package main

import (
	"flag"
	"fmt"
	"os"

	"ironlattice-vis/demo3-phasefield/pkg/physics"
)

func main() {
	// Command line flags
	gridSize := flag.Int("grid", 32, "Grid size (NxNxN)")
	steps := flag.Int("steps", 1000, "Number of simulation steps")
	temp := flag.Float64("temp", 300, "Temperature in Kelvin")
	field := flag.Float64("field", 0, "External electric field (V/m)")
	headless := flag.Bool("headless", false, "Run in headless mode (no graphics)")
	printInterval := flag.Int("print", 100, "Print interval for stats")
	flag.Parse()

	fmt.Println("===========================================")
	fmt.Println("  IronLattice Phase-Field Simulator")
	fmt.Println("  Demo 3: GPU-Accelerated TDGL Solver")
	fmt.Println("===========================================")
	fmt.Println()

	// Create material
	material := physics.DefaultHZO()

	// Print material info
	printMaterialInfo(material, *temp)

	// Create computational grid
	n := *gridSize
	dx := material.LatticeCellSize * 2 // Grid spacing
	grid := physics.NewGrid3D(n, n, n, dx)

	fmt.Printf("Grid: %dx%dx%d (%.1f nm)³\n", n, n, n, dx*float64(n)*1e9)
	fmt.Printf("Cell size: %.2f nm\n", dx*1e9)
	fmt.Println()

	// Create TDGL solver
	solver := physics.NewTDGLSolver(grid, material)
	solver.SetTemperature(*temp)
	solver.SetExternalField(*field)
	solver.AutoSetTimeStep()

	fmt.Printf("Time step: %.3e s\n", solver.StabilityLimit())
	fmt.Printf("Stability limit: %.3e s\n", solver.StabilityLimit())
	fmt.Println()

	// Initialize domain pattern
	Ps := material.SpontaneousPolarization(*temp)
	fmt.Printf("Spontaneous polarization Ps: %.4f C/m²\n", Ps)
	solver.InitializeDomainPattern(Ps)
	fmt.Println()

	if *headless {
		runHeadless(solver, *steps, *printInterval)
	} else {
		runGraphical(solver, *steps)
	}
}

func printMaterialInfo(m *physics.HZOMaterial, T float64) {
	fmt.Println("HZO Material Parameters:")
	fmt.Printf("  α = %.2e Vm/C (at T=%.0f K)\n", m.AlphaTemperature(T), T)
	fmt.Printf("  β = %.2e Vm⁵/C³\n", m.Beta)
	fmt.Printf("  γ = %.2e Vm⁹/C⁵\n", m.Gamma)
	fmt.Printf("  κ = %.2e Vm³/C\n", m.Kappa)
	fmt.Printf("  L = %.2e m³/VsC\n", m.L)
	fmt.Printf("  Tc = %.0f K\n", m.Tc)
	fmt.Println()
}

func runHeadless(solver *physics.TDGLSolver, steps, printInterval int) {
	fmt.Println("Running headless simulation...")
	fmt.Println()

	// Initial state
	fmt.Printf("Initial: %s\n", solver.Stats())
	fmt.Println()

	// Run simulation
	for i := 1; i <= steps; i++ {
		solver.Step()

		if i%printInterval == 0 {
			fmt.Printf("%s\n", solver.Stats())
		}
	}

	fmt.Println()
	fmt.Printf("Final: %s\n", solver.Stats())

	// Print domain statistics
	posF, negF := solver.ComputeDomainFraction()
	fmt.Println()
	fmt.Println("Domain Statistics:")
	fmt.Printf("  Positive domains (P>0): %.1f%%\n", posF*100)
	fmt.Printf("  Negative domains (P<0): %.1f%%\n", negF*100)
	fmt.Printf("  Average polarization: %.4e C/m²\n", solver.ComputeAveragePolarization())
	fmt.Printf("  Total free energy: %.4e J\n", solver.ComputeFreeEnergy())

	fmt.Println()
	fmt.Println("Simulation complete.")
}

func runGraphical(solver *physics.TDGLSolver, steps int) {
	fmt.Println("Graphical mode not yet implemented.")
	fmt.Println()
	fmt.Println("To run the simulation, use --headless flag:")
	fmt.Println("  ./phasefield --headless --grid 32 --steps 1000")
	fmt.Println()
	fmt.Println("Coming soon:")
	fmt.Println("  - Real-time 3D domain visualization")
	fmt.Println("  - Interactive temperature/field control")
	fmt.Println("  - Volume rendering with Vulkan")
	fmt.Println()

	// For now, run headless
	runHeadless(solver, steps, 100)

	os.Exit(0)
}
