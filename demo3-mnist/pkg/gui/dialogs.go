// Package gui provides Fyne-based GUI components for MNIST visualization.
// dialogs.go provides educational info dialogs.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowWhy30LevelsDialog displays information about the 30 analog levels.
func ShowWhy30LevelsDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Why 30 Analog Levels?", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Physics Justification", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• HZO Ferroelectric: ~30 stable polarization states"),
		widget.NewLabel("• Domain Wall Pinning: Natural quantization from crystal defects"),
		widget.NewLabel("• ADC Resolution: 6-bit (64 levels) → 30 reliably distinguishable"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Competitive Advantage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Technology     | Levels | Notes"),
		widget.NewLabel("---------------|--------|------------------"),
		widget.NewLabel("Flash (NAND)   | 2-4    | TLC/QLC"),
		widget.NewLabel("ReRAM          | 4-16   | Limited by variability"),
		widget.NewLabel("FeCIM (HZO)    | 30     | 5x better than ReRAM"),
		widget.NewLabel("Ideal (FP32)   | 2^32   | Baseline"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Impact on MNIST Accuracy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• 2 levels (binary): ~50% (worse than random!)"),
		widget.NewLabel("• 8 levels: ~75%"),
		widget.NewLabel("• 30 levels: ~87% (FeCIM hardware)"),
		widget.NewLabel("• Float32: ~98% (theoretical)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Why Not 64 Levels (6-bit ADC)?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Only 30 are reliably distinguishable due to:"),
		widget.NewLabel("1. Device-to-device variation (~2.75%)"),
		widget.NewLabel("2. Cycle-to-cycle variation (~1.5%)"),
		widget.NewLabel("3. Read noise (~0.5% σ/μ)"),
		widget.NewLabel(""),
		widget.NewLabel("With 3σ separation requirement, 30 levels is the practical limit."),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("Why 30 Levels?", "Close", scroll, window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}

// ShowHardwareRealityDialog displays the hardware reality check information.
func ShowHardwareRealityDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Hardware Reality Check", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Why 87% and Not 98%?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
		widget.NewLabel("Simulation (this demo): Can achieve 95-98% under ideal conditions."),
		widget.NewLabel("FeCIM Hardware (Dr. Tour): 87% measured, 88% theoretical max."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("The Gap Explained", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Non-Ideality          | Sim | HW  | Impact"),
		widget.NewLabel("----------------------|-----|-----|--------"),
		widget.NewLabel("Weight quantization   | Yes | Yes | -1%"),
		widget.NewLabel("Read noise            | Yes | Yes | -2%"),
		widget.NewLabel("IR drop               | No  | Yes | -3%"),
		widget.NewLabel("Sneak paths           | No  | Yes | -2%"),
		widget.NewLabel("ADC non-linearity     | No  | Yes | -1%"),
		widget.NewLabel("Retention drift       | No  | Yes | -1%"),
		widget.NewLabel("Cycle-to-cycle var.   | No  | Yes | -2%"),
		widget.NewLabel(""),
		widget.NewLabel("Total: ~12% gap between ideal (98%) and hardware (87%)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("How to Match Hardware in Simulation", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Set noise level to ~0.08 in the GUI."),
		widget.NewLabel("This empirically matches the 87% target."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Energy Efficiency", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("FeCIM Energy: ~50 fJ/MAC (Jerry et al. IEDM 2017)"),
		widget.NewLabel("GPU Energy:   ~500 pJ/MAC (V100 with DRAM)"),
		widget.NewLabel(""),
		widget.NewLabel("For MNIST (101,632 MACs):"),
		widget.NewLabel("• FeCIM: 5.08 μJ"),
		widget.NewLabel("• GPU:   50.8 mJ"),
		widget.NewLabel("• Ratio: 10,000x more efficient!"),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("Hardware Reality Check", "Close", scroll, window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}

// ShowFailureModesDialog displays information about failure modes.
func ShowFailureModesDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Failure Modes Explained", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("1. Quantization Cliff (< 4 levels)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Preset: 'Quant Cliff' button"),
		widget.NewLabel("Settings: Levels=2, Noise=0.01, ADC=8"),
		widget.NewLabel("Result: Accuracy ~50% (worse than random!)"),
		widget.NewLabel(""),
		widget.NewLabel("Why: Binary weights {-1, +1} cannot represent"),
		widget.NewLabel("the 128-dimensional weight space."),
		widget.NewLabel("Network loses ability to distinguish classes."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("2. Noise Wall (> 0.10 noise)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Preset: 'Noisy' button"),
		widget.NewLabel("Settings: Levels=30, Noise=0.15, ADC=6"),
		widget.NewLabel("Result: Accuracy ~70%, confidence drops to 40-60%"),
		widget.NewLabel(""),
		widget.NewLabel("Why: Gaussian noise in MVM corrupts output currents."),
		widget.NewLabel("ADC reads wrong values. '8' misclassified as '3'."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("3. ADC Quantization Artifacts (< 4-bit)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Preset: 'Broken ADC' button"),
		widget.NewLabel("Settings: Levels=30, Noise=0.01, ADC=3"),
		widget.NewLabel("Result: Accuracy ~65%, staircase artifacts"),
		widget.NewLabel(""),
		widget.NewLabel("Why: 3-bit ADC = only 8 output levels."),
		widget.NewLabel("Hidden layer activations coarsely quantized."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("4. Confidence Collapse (Extreme)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Settings: Levels=2, Noise=0.20, ADC=3"),
		widget.NewLabel("Result: All probabilities → ~10% (uniform)"),
		widget.NewLabel(""),
		widget.NewLabel("Why: Combination of insufficient precision,"),
		widget.NewLabel("high noise, and coarse ADC."),
		widget.NewLabel("Network cannot extract meaningful features."),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("Failure Modes", "Close", scroll, window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}

// ShowAboutDialog displays information about the demo.
func ShowAboutDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabelWithStyle("MNIST FeCIM Demo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("87% Hardware Target"),
		widget.NewSeparator(),

		widget.NewLabel("\"We're at 87% validation here... theoretical is 88%.\""),
		widget.NewLabel("— Dr. external research group, external research institution (Nov 2024)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("This Demo Answers:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("1. What are 30 analog levels? (Physics + advantage)"),
		widget.NewLabel("2. Why does FeCIM achieve 87%? (Reality vs simulation)"),
		widget.NewLabel("3. What happens when hardware fails? (Failure modes)"),
		widget.NewLabel("4. Why does this matter? (10,000x energy savings)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Architecture", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Input: 784 neurons (28x28 pixels)"),
		widget.NewLabel("Hidden: 128 neurons (ReLU activation)"),
		widget.NewLabel("Output: 10 neurons (Softmax, digits 0-9)"),
		widget.NewLabel("Total MACs: 101,632 per inference"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("References", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• Dr. external research group, external research institution (2024)"),
		widget.NewLabel("• Jerry et al., IEDM 2017 (FeFET Synapse)"),
		widget.NewLabel("• MNIST Dataset - Yann LeCun"),
		widget.NewSeparator(),

		widget.NewLabel("GitHub: XelHaku/multilayer-ferroelectric-cim-visualizer"),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(450, 400))

	d := dialog.NewCustom("About", "Close", scroll, window)
	d.Resize(fyne.NewSize(500, 500))
	d.Show()
}
