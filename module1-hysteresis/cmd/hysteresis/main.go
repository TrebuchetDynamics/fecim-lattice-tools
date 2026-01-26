// Command hysteresis provides an interactive visualization of ferroelectric
// hysteresis in HfO2-ZrO2 superlattice materials.
//
// This is Demo 1 of the FeCIM Visualizer project.
//
// Run modes:
//   - Default: Fyne GUI with real-time P-E curve animation (recommended)
//   - --tui: Terminal user interface (for SSH/remote)
//   - --headless: ASCII terminal output (static, no interactivity)
//   - --vulkan: Vulkan-based graphical interface (advanced)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui"
	"fecim-lattice-tools/module1-hysteresis/pkg/render"
	"fecim-lattice-tools/module1-hysteresis/pkg/simulation"
	"fecim-lattice-tools/module1-hysteresis/pkg/tui"
)

func main() {
	// Command line flags
	optimized := flag.Bool("optimized", false, "Use optimized superlattice parameters")
	freq := flag.Float64("freq", 1e6, "Waveform frequency in Hz")
	headless := flag.Bool("headless", false, "Run in headless mode (static ASCII output)")
	tuiMode := flag.Bool("tui", false, "Run terminal UI mode (for SSH/remote)")
	vulkan := flag.Bool("vulkan", false, "Run with Vulkan graphics (GPU accelerated)")
	flag.Parse()

	// Determine run mode based on flags
	if *headless {
		// Headless mode - static ASCII output
		fmt.Println("===========================================")
		fmt.Println("  FeCIM Hysteresis Visualizer")
		fmt.Println("  Demo 1: Ferroelectric P-E Curve")
		fmt.Println("===========================================")
		fmt.Println()

		var material *ferroelectric.HZOMaterial
		if *optimized {
			material = ferroelectric.OptimizedHZO()
			fmt.Println("Using: Optimized HfO2/ZrO2 Superlattice")
		} else {
			material = ferroelectric.DefaultHZO()
			fmt.Println("Using: Default HZO Parameters")
		}

		printMaterialInfo(material)
		engine := simulation.NewEngine(material)
		engine.SetFrequency(*freq)
		runHeadless(engine)
		return
	}

	if *tuiMode {
		// Terminal UI mode
		if err := tui.Run(); err != nil {
			log.Printf("TUI error: %v\n", err)
			fmt.Println("\nFalling back to headless mode...")

			var material *ferroelectric.HZOMaterial
			if *optimized {
				material = ferroelectric.OptimizedHZO()
			} else {
				material = ferroelectric.DefaultHZO()
			}
			engine := simulation.NewEngine(material)
			engine.SetFrequency(*freq)
			runHeadless(engine)
		}
		return
	}

	if *vulkan {
		// Vulkan graphical mode
		fmt.Println("===========================================")
		fmt.Println("  FeCIM Hysteresis Visualizer")
		fmt.Println("  Demo 1: Ferroelectric P-E Curve (Vulkan)")
		fmt.Println("===========================================")
		fmt.Println()

		var material *ferroelectric.HZOMaterial
		if *optimized {
			material = ferroelectric.OptimizedHZO()
			fmt.Println("Using: Optimized HfO2/ZrO2 Superlattice")
		} else {
			material = ferroelectric.DefaultHZO()
			fmt.Println("Using: Default HZO Parameters")
		}

		printMaterialInfo(material)
		engine := simulation.NewEngine(material)
		engine.SetFrequency(*freq)
		runGraphical(engine)
		return
	}

	// Default: Fyne GUI mode (recommended)
	if err := gui.Run(); err != nil {
		log.Printf("GUI error: %v\n", err)
		fmt.Println("\nFalling back to TUI mode...")

		if err := tui.Run(); err != nil {
			log.Printf("TUI error: %v\n", err)
			fmt.Println("\nFalling back to headless mode...")

			var material *ferroelectric.HZOMaterial
			if *optimized {
				material = ferroelectric.OptimizedHZO()
			} else {
				material = ferroelectric.DefaultHZO()
			}
			engine := simulation.NewEngine(material)
			engine.SetFrequency(*freq)
			runHeadless(engine)
		}
	}
}

func printMaterialInfo(m *ferroelectric.HZOMaterial) {
	fmt.Println("\nMaterial Parameters:")
	fmt.Printf("  Remanent Polarization (Pr): %.1f μC/cm²\n", m.Pr*100)
	fmt.Printf("  Saturation Polarization (Ps): %.1f μC/cm²\n", m.Ps*100)
	fmt.Printf("  Coercive Field (Ec): %.2f MV/cm\n", m.Ec/1e8)
	fmt.Printf("  Coercive Voltage (Vc): %.2f V\n", m.CoerciveVoltage())
	fmt.Printf("  Film Thickness: %.0f nm\n", m.Thickness*1e9)
	fmt.Printf("  Relative Permittivity: %.0f\n", m.Epsilon)
	fmt.Println()
}

func runHeadless(engine *simulation.Engine) {
	fmt.Println("Running enhanced terminal visualization...")
	fmt.Println()

	// Get material
	material := ferroelectric.DefaultHZO()

	// Create advanced Preisach model
	model := ferroelectric.NewMayergoyzPreisach(material, 40)

	// Create renderer
	renderer := ferroelectric.NewPERenderer()

	// Generate and render P-E loop
	Emax := material.Ec * 2
	E, P := model.GetHysteresisLoop(Emax, 100)
	fmt.Println(renderer.RenderPELoop(E, P, material))

	// Render domain states
	alphas, betas, states := model.GetPreisachPlane()
	fmt.Println(renderer.RenderDomainStates(alphas, betas, states))

	// Render discrete states
	discreteStates := model.DiscreteStates(30)
	fmt.Println(renderer.RenderDiscreteStates(discreteStates))

	// Render switching dynamics
	times, pols, switched := model.SimulateDomainSwitching(Emax, 10*material.Tau, 50)
	fmt.Println(renderer.RenderSwitchingDynamics(times, pols, switched, material))

	// Render temperature dependence
	fmt.Println(renderer.RenderTemperatureDependence(material))

	// Render material comparison
	fmt.Println(renderer.RenderMaterialComparison())

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                     SIMULATION SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Material: %s\n", material.Name)
	fmt.Printf("  Remanent Polarization: %.1f µC/cm²\n", material.Pr*100)
	fmt.Printf("  Coercive Field: %.2f MV/cm\n", material.Ec/1e8)
	fmt.Printf("  Switching Time: %.2f ns\n", material.Tau*1e9)
	fmt.Printf("  Endurance: %.0e cycles\n", material.EnduranceCycles)
	fmt.Printf("  30 Discrete States: %.1f bits/cell\n", 4.91)
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("  \"It's got 30 discrete states. So it's not 0-1-0-1.\"")
	fmt.Println("  - Dr. external research group")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()
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
