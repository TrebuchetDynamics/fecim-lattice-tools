// Package gui provides Fyne-based GUI components for crossbar visualization.
package gui

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/shared/logging"
	sharedtheme "fecim-lattice-tools/shared/theme"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Package-level logger using shared logging infrastructure
var debug *logging.Logger

func init() {
	debug = logging.NewLogger("crossbar-app")
}

// CrossbarApp is the main application for the crossbar demo.
type CrossbarApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Core components
	array  *crossbar.Array
	config *crossbar.Config

	// GUI components
	conductanceHeatmap *CrossbarHeatmap
	irDropHeatmap      *CrossbarHeatmap
	sneakPathHeatmap   *CrossbarHeatmap

	// Color legends for each heatmap
	condLegend  *sharedwidgets.ColorLegend
	irLegend    *sharedwidgets.ColorLegend
	sneakLegend *sharedwidgets.ColorLegend

	controlPanel   *ControlPanel
	statsPanel     *StatsPanel
	levelIndicator *LevelIndicator

	// Enhanced widgets
	metricsPanel      *MetricsPanel
	comparisonBadge   *ComparisonBadge
	accuracyWaterfall *AccuracyWaterfall
	beforeAfterToggle *BeforeAfterToggle

	// Live Slide components
	modeIndicator    *ModeIndicatorBox
	educationalPanel *EducationalPanel
	operationLog     *OperationLog
	ioDisplay        *InputOutputDisplay
	keyStat          *KeyStatBox

	// Simple left panel labels (replacing custom widgets)
	eduTitleLabel   *widget.Label
	eduContentLabel *widget.Label
	keyStatLabel    *widget.Label
	keyStatValue    *widget.Label

	// Simple right panel widgets (replacing custom widgets)
	runMVMButton    *widget.Button
	resetButton     *widget.Button
	arraySizeLabel  *widget.Label
	arraySizeSlider *widget.Slider
	noiseLabel      *widget.Label
	noiseSlider     *widget.Slider
	adcBitsLabel    *widget.Label
	adcBitsSlider   *widget.Slider
	colormapSelect  *widget.Select
	statsLabel      *widget.Label

	// Track colormap per tab
	condColormap  string
	irColormap    string
	sneakColormap string

	// Architecture selector (clarify 0T1R vs 1T1R)
	// Recommendation: clarify sneak path behavior depends on architecture
	ca.archSelect = widget.NewSelect([]string{"1T1R (Transistor)", "0T1R (Passive)"}, func(s string) {
		ca.stateMu.Lock()
		ca.architecture = s
		ca.stateMu.Unlock()
		// Update educational content based on architecture
		if s == "1T1R (Transistor)" {
			ca.setEducationalContent("1T1R Architecture",
				"1T1R = One Transistor per FeFET\n\n"+
					"How it works:\n"+
					"Transistor acts as controlled\n"+
					"switch, isolating unselected cells.\n\n"+
					"Advantages:\n"+
					"✓ Zero sneak paths\n"+
					"✓ Linear I-V characteristics\n"+
					"✓ Industry standard (SRAM-like)\n\n"+
					"Tradeoffs:\n"+
					"✗ 50% area overhead\n"+
					"✗ More complex fabrication\n\n"+
					"Best for: High-precision inference\n"+
					"(vision, language models)")
		} else {
			ca.setEducationalContent("0T1R Architecture",
				"0T1R = Passive Crossbar (no transistor)\n\n"+
					"How it works:\n"+
					"Direct connection between wires.\n"+
					"FeFET is the only device.\n\n"+
					"Advantages:\n"+
					"✓ Highest density (4F² per cell)\n"+
					"✓ Simpler fabrication\n"+
					"✓ Lower cost\n\n"+
					"Tradeoffs:\n"+
					"✗ Sneak paths (2-15% SNR loss)\n"+
					"✗ Requires selector device OR\n"+
					"    self-rectifying FeFET\n\n"+
					"FeFET advantage: Natural\n"+
					"rectification in HfO₂-ZrO₂!")
		}
	})
	ca.archSelect.SetSelected("1T1R (Transistor)")

	ca.statsLabel = widget.NewLabel("Analysis Results\n\nNo data yet.\nClick Run MVM to start.")
	ca.statsLabel.Wrapping = fyne.TextWrapOff
	ca.statsLabel.TextStyle = fyne.TextStyle{Monospace: true} // Fixed-width prevents resize

	// Create status labels
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	ca.infoLabel = widget.NewLabel(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))

	// Hover info label - shows cell info on mouse hover
	ca.hoverInfoLabel = widget.NewLabel("Hover over cells to see values")
	ca.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}
	ca.hoverInfoLabel.Wrapping = fyne.TextWrapOff
	ca.hoverInfoLabel.Truncation = fyne.TextTruncateEllipsis

	// Create tabbed heatmap view - use Max to fill available space
	ca.tabs = container.NewAppTabs(
		container.NewTabItem("Conductance", container.NewMax(ca.conductanceHeatmap)),
		container.NewTabItem("IR Drop", container.NewMax(ca.irDropHeatmap)),
		container.NewTabItem("Sneak Paths", container.NewMax(ca.sneakPathHeatmap)),
		container.NewTabItem("Input/Output", container.NewMax(ca.mvmVis)),
	)

	// Update educational panel based on selected tab
	ca.tabs.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "Conductance":
			ca.setEducationalContent("Conductance Matrix",
				"Each cell = one FeFET\n\n"+
					"Conductance G (1-100 µS)\n"+
					"stored as 30 discrete levels\n"+
					"(~4.9 bits per cell)\n\n"+
					"This is your weight matrix W.\n"+
					"Brighter = higher conductance\n\n"+
					"Click any cell to read its\n"+
					"exact conductance value.")
		case "IR Drop":
			ca.setEducationalContent("IR Drop Analysis",
				"Wire resistance causes voltage\n"+
					"drops along metal lines.\n\n"+
					"Red = high voltage drop (>5%)\n"+
					"Blue = low drop (<1%)\n\n"+
					"Impact: Cells far from drivers\n"+
					"compute with reduced voltage,\n"+
					"causing accuracy degradation.\n\n"+
					"Mitigation: Multiple distributed\n"+
					"voltage drivers.\n\n"+
					"Auto-computed after MVM.")
		case "Sneak Paths":
			ca.setEducationalContent("Sneak Path Analysis",
				"Unintended current paths\n"+
					"through passive crossbars.\n\n"+
					"Red = high parasitic current\n"+
					"Blue = minimal leakage\n\n"+
					"Impact: Reduces SNR, especially\n"+
					"in large arrays (>128x128).\n\n"+
					"Mitigation:\n"+
					"• 1T1R architecture (transistor)\n"+
					"• Selector devices (diode)\n"+
					"• Self-rectifying FeFETs\n\n"+
					"Auto-computed after MVM.")
		case "Input/Output":
			ca.setEducationalContent("MVM Vectors",
				"Matrix-Vector Multiplication\n"+
					"in a single analog step.\n\n"+
					"Top: Input voltages V (DAC)\n"+
					"Bottom: Output currents I (ADC)\n\n"+
					"Physics: I = G × V (Ohm's Law)\n"+
					"Result: I_row = Σ(G_ij × V_j)\n\n"+
					"All N² multiply-accumulate\n"+
					"operations happen in parallel\n"+
					"in ~1ns (speed of light).\n\n"+
					"Click 'Run MVM' to see it!")
		}
	}

	// Title and header
	titleLabel := widget.NewLabel("FeCIM Crossbar Array Visualization")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
	)

	// Right panel - improved layout with better spacing and grouping

	// Action buttons group
	actionLabel := widget.NewLabelWithStyle("Actions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	actionsGroup := container.NewVBox(
		actionLabel,
		ca.runMVMButton,
		ca.resetButton,
	)

	// Array settings group - collapsible style header
	settingsLabel := widget.NewLabelWithStyle("Array Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	settingsGroup := container.NewVBox(
		widget.NewSeparator(),
		settingsLabel,
		ca.arraySizeLabel,
		ca.arraySizeSlider,
	)

	// Noise/ADC group
	signalLabel := widget.NewLabelWithStyle("Signal Quality", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	signalGroup := container.NewVBox(
		widget.NewSeparator(),
		signalLabel,
		ca.noiseLabel,
		ca.noiseSlider,
		ca.adcBitsLabel,
		ca.adcBitsSlider,
	)

	// Architecture settings group (clarify 0T1R vs 1T1R)
	archLabel := widget.NewLabelWithStyle("Architecture", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	archGroup := container.NewVBox(
		widget.NewSeparator(),
		archLabel,
		ca.archSelect,
	)

	// Display settings group
	displayLabel := widget.NewLabelWithStyle("Display", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	displayGroup := container.NewVBox(
		widget.NewSeparator(),
		displayLabel,
		ca.colormapSelect,
	)

	// Combined controls with scroll for overflow
	controlsBox := container.NewVBox(
		actionsGroup,
		settingsGroup,
		archGroup,
		signalGroup,
		displayGroup,
	)
	controlsScroll := container.NewVScroll(controlsBox)
	controlsScroll.SetMinSize(fyne.NewSize(240, 250))

	// Stats section with header
	statsHeader := widget.NewLabelWithStyle("Analysis Results", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	statsSection := container.NewBorder(
		container.NewVBox(widget.NewSeparator(), statsHeader),
		nil, nil, nil,
		ca.statsLabel,
	)
	statsScroll := container.NewVScroll(statsSection)
	statsScroll.SetMinSize(fyne.NewSize(240, 120))

	// Use VSplit for controls and stats
	rightPanel := container.NewVSplit(
		controlsScroll,
		statsScroll,
	)
	rightPanel.SetOffset(0.6) // 60% controls, 40% stats

	// Left panel using simple labels (no custom widgets)
	leftPanel := container.NewVBox(
		ca.eduTitleLabel,
		widget.NewSeparator(),
		ca.eduContentLabel,
		widget.NewSeparator(),
		ca.keyStatLabel,
		ca.keyStatValue,
	)

	// Simple status footer with hover info
	// Wrap hoverInfoLabel in fixed-size container to prevent layout recalc on text change
	hoverInfoContainer := container.NewGridWrap(fyne.NewSize(450, 20), ca.hoverInfoLabel)
	simpleFooter := container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		hoverInfoContainer,
		widget.NewSeparator(),
		ca.infoLabel,
	)

	// Use HSplit for proportional 3-column layout
	// Left panel (15%) | Center tabs (70%) | Right panel (15%)
	ca.leftCenterSplit = container.NewHSplit(leftPanel, ca.tabs)
	ca.leftCenterSplit.SetOffset(0.15) // 15% left, 85% center+right

	ca.mainSplit = container.NewHSplit(ca.leftCenterSplit, rightPanel)
	ca.mainSplit.SetOffset(0.8) // 80% left+center, 20% right

	// Create responsive detector for breakpoint-based layout adjustments
	ca.responsiveDetector = sharedwidgets.NewResponsiveDetector(ca.onBreakpointChange)
	ca.currentBreakpoint = sharedwidgets.BreakpointXL // Default to desktop

	// Wrap with header and footer
	mainContent := container.NewBorder(
		header,       // top
		simpleFooter, // bottom
		nil,          // left
		nil,          // right
		ca.mainSplit, // center - the split panels
	)

	// Stack with responsive detector overlay
	return container.NewStack(mainContent, ca.responsiveDetector)
}

// setupControlCallbacks connects control panel events to actions.
func (ca *CrossbarApp) setupControlCallbacks() {
	ca.controlPanel.OnArraySizeChanged = func(size int) {
		ca.recreateArray(size, ca.config.NoiseLevel, ca.config.ADCBits)
	}

	ca.controlPanel.OnNoiseChanged = func(noise float64) {
		ca.config.NoiseLevel = noise
		ca.recreateArray(ca.config.Rows, noise, ca.config.ADCBits)
	}

	ca.controlPanel.OnADCBitsChanged = func(bits int) {
		ca.config.ADCBits = bits
		ca.recreateArray(ca.config.Rows, ca.config.NoiseLevel, bits)
	}

	ca.controlPanel.OnColormapChanged = func(colormap string) {
		ca.conductanceHeatmap.SetColormap(colormap)
	}

	ca.controlPanel.OnDemoModeChanged = ca.onDemoModeChanged
	ca.controlPanel.OnRunMVM = ca.runMVM
	ca.controlPanel.OnAnalyzeIR = ca.analyzeIRDrop
	ca.controlPanel.OnAnalyzeSneak = ca.analyzeSneakPaths
	ca.controlPanel.OnReset = ca.resetArray
}

// recreateArray creates a new array with updated parameters.
func (ca *CrossbarApp) recreateArray(size int, noise float64, adcBits int) {
	ca.config = &crossbar.Config{
		Rows:       size,
		Cols:       size,
		NoiseLevel: noise,
		ADCBits:    adcBits,
		DACBits:    8,
	}

	var err error
	ca.array, err = crossbar.NewArray(ca.config)
	if err != nil {
		ca.updateStatus("Error creating array")
		if ca.window != nil {
			dialog.ShowError(fmt.Errorf("failed to create crossbar array: %w", err), ca.window)
		}
		return
	}

	// Resize existing heatmaps instead of creating new ones
	// This preserves the widget references in the window layout
	ca.conductanceHeatmap.SetDimensions(size, size)
	ca.irDropHeatmap.SetDimensions(size, size)
	ca.sneakPathHeatmap.SetDimensions(size, size)

	ca.programRandomWeights()
	ca.updateConductanceDisplay()
	ca.updateInfoLabel()
	ca.setKeyStatValue(fmt.Sprintf("%d MACs", size*size))
	ca.updateStatus(fmt.Sprintf("Array resized to %dx%d (%d parallel MACs)", size, size, size*size))
}

// programRandomWeights fills the array with random weights quantized to 30 levels.
func (ca *CrossbarApp) programRandomWeights() {
	for i := 0; i < ca.config.Rows; i++ {
		for j := 0; j < ca.config.Cols; j++ {
			level := rand.Intn(30)
			weight := float64(level) / 29.0
			ca.array.ProgramWeight(i, j, weight)
		}
	}
}

// updateConductanceDisplay refreshes the conductance heatmap.
func (ca *CrossbarApp) updateConductanceDisplay() {
	matrix := ca.array.GetConductanceMatrix()
	fyne.Do(func() {
		ca.conductanceHeatmap.SetData(matrix)
	})
}

// updateStatus updates the status label.
func (ca *CrossbarApp) updateStatus(status string) {
	ca.statusLabel.SetText("Status: " + status)
}

// setEducationalContent updates the educational panel.
func (ca *CrossbarApp) setEducationalContent(title, content string) {
	ca.eduTitleLabel.SetText(title)
	ca.eduContentLabel.SetText(content)
}

// setKeyStatValue updates the key statistic display.
func (ca *CrossbarApp) setKeyStatValue(value string) {
	ca.keyStatValue.SetText(value)
}

// updateInfoLabel updates the info label with current config.
func (ca *CrossbarApp) updateInfoLabel() {
	ca.infoLabel.SetText(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))
}

// onBreakpointChange handles responsive layout adjustments.
func (ca *CrossbarApp) onBreakpointChange(bp sharedwidgets.Breakpoint, size fyne.Size) {
	ca.currentBreakpoint = bp

	// Adjust split offsets based on breakpoint
	switch bp {
	case sharedwidgets.BreakpointSM, sharedwidgets.BreakpointMD:
		// Small/Medium: Minimize side panels, maximize heatmap area
		// Left panel: collapse to 5% (minimal educational info)
		// Right panel: collapse to 10% (minimal controls)
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.05) // 5% left, 95% center+right
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.9) // 90% left+center, 10% right
		}

	case sharedwidgets.BreakpointLG:
		// Large: Balanced layout for laptops
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.12) // 12% left
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.85) // 85% left+center, 15% right
		}

	case sharedwidgets.BreakpointXL:
		// Extra Large: Desktop - original comfortable layout
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.15) // 15% left
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.8) // 80% left+center, 20% right
		}
	}
}
