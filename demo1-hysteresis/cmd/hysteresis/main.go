// Command hysteresis provides an interactive visualization of ferroelectric
// hysteresis in HfO2-ZrO2 superlattice materials.
//
// This is Demo 1 of the IronLattice Visualizer project.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"ironlattice-vis/demo1-hysteresis/pkg/ferroelectric"
	"ironlattice-vis/demo1-hysteresis/pkg/render"
	"ironlattice-vis/demo1-hysteresis/pkg/simulation"
)

func main() {
	// Command line flags
	optimized := flag.Bool("optimized", false, "Use optimized superlattice parameters")
	freq := flag.Float64("freq", 1e6, "Waveform frequency in Hz")
	headless := flag.Bool("headless", false, "Run in headless mode (no graphics)")
	flag.Parse()

	fmt.Println("===========================================")
	fmt.Println("  IronLattice Hysteresis Visualizer")
	fmt.Println("  Demo 1: Ferroelectric P-E Curve")
	fmt.Println("===========================================")
	fmt.Println()

	// Select material parameters
	var material *ferroelectric.HZOMaterial
	if *optimized {
		material = ferroelectric.OptimizedHZO()
		fmt.Println("Using: Optimized HfO2/ZrO2 Superlattice")
	} else {
		material = ferroelectric.DefaultHZO()
		fmt.Println("Using: Default HZO Parameters")
	}

	// Print material parameters
	printMaterialInfo(material)

	// Create simulation engine
	engine := simulation.NewEngine(material)
	engine.SetFrequency(*freq)

	if *headless {
		// Run headless simulation and print results
		runHeadless(engine)
	} else {
		// Run with graphics
		runGraphical(engine)
	}
}

func printMaterialInfo(m *ferroelectric.HZOMaterial) {
	fmt.Println("\nMaterial Parameters:")
	fmt.Printf("  Remanent Polarization (Pr): %.1f μC/cm²\n", m.Pr*1e4)
	fmt.Printf("  Saturation Polarization (Ps): %.1f μC/cm²\n", m.Ps*1e4)
	fmt.Printf("  Coercive Field (Ec): %.2f MV/cm\n", m.Ec/1e8)
	fmt.Printf("  Coercive Voltage (Vc): %.2f V\n", m.CoerciveVoltage())
	fmt.Printf("  Film Thickness: %.0f nm\n", m.Thickness*1e9)
	fmt.Printf("  Relative Permittivity: %.0f\n", m.Epsilon)
	fmt.Println()
}

func runHeadless(engine *simulation.Engine) {
	fmt.Println("Running headless simulation...")
	fmt.Println()

	// Generate hysteresis loop data
	E, P := engine.GetHysteresisData()

	fmt.Println("Hysteresis Loop Data (E, P):")
	fmt.Println("-----------------------------")

	// Print a subset of points
	step := len(E) / 20
	for i := 0; i < len(E); i += step {
		fmt.Printf("  E: %+8.2e V/m, P: %+8.4f (normalized)\n", E[i], P[i])
	}

	fmt.Println()
	fmt.Println("Discrete States (30-level analog):")
	fmt.Println("-----------------------------------")

	model := ferroelectric.NewPreisachModel(ferroelectric.DefaultHZO())
	states := model.DiscreteStates(30)
	for i, s := range states {
		fmt.Printf("  State %2d: P = %+.4f C/m²\n", i, s)
	}

	fmt.Println()
	fmt.Println("Simulation complete.")
}

func runGraphical(engine *simulation.Engine) {
	fmt.Println("Starting Vulkan-based graphical interface...")
	fmt.Println("Press ESC or close window to exit.")
	fmt.Println()

	// Create Vulkan renderer
	config := render.DefaultConfig()
	renderer := render.NewVulkanRenderer(config)

	// Create hysteresis plot
	material := ferroelectric.DefaultHZO()
	Emax := material.Ec * 1.5
	Pmax := material.Ps * 1.2
	plot := render.NewHysteresisPlot(Emax, Pmax)
	renderer.SetHysteresisPlot(plot)

	// Set up update callback
	frameCount := 0
	engine.Start()
	renderer.SetUpdateCallback(func() {
		// Step simulation
		engine.Step()
		state := engine.State()

		// Update renderer with new polarization
		renderer.UpdatePolarization(state.NormPol)

		// Add point to plot
		plot.AddPoint(state.ElectricField, state.Polarization)

		frameCount++
	})

	// Initialize Vulkan
	if err := renderer.Initialize(); err != nil {
		log.Printf("Failed to initialize Vulkan renderer: %v", err)
		fmt.Println()
		fmt.Println("Vulkan initialization failed. Running in headless mode instead.")
		fmt.Println()
		runHeadless(engine)
		os.Exit(0)
	}
	defer renderer.Cleanup()

	// Run render loop
	if err := renderer.Run(); err != nil {
		log.Fatalf("Renderer error: %v", err)
	}

	fmt.Printf("\nSimulation completed. Rendered %d frames.\n", frameCount)
}
