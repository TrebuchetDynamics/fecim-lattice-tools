// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/shared/logging"
	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
)

// Package-level logger using shared logging infrastructure
var debug *logging.Logger

func init() {
	debug = logging.NewLogger("comparison-app")
}

// Energy specs - sourced from docs/videos/ironlattice-youtube-script.md line 205:
// "CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. FeCIM: under 1 picojoule."
const (
	cpuEnergyPJPerMAC   = 1000.0 // 1000 pJ/MAC
	gpuEnergyPJPerMAC   = 100.0  // 100 pJ/MAC
	fecimEnergyPJPerMAC = 1.0    // ~1 pJ/MAC (conservative for claimed "<1 pJ")
)

// EnergySpec holds energy per MAC specifications with sources.
type EnergySpec struct {
	Name          string
	EnergyFJ      float64 // femtojoules per MAC (1 pJ = 1000 fJ)
	Source        string
	Verified      bool
	SourceDetails string
}

// ComparisonApp is the main application for architecture comparison.
type ComparisonApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Energy specs
	cpuSpec   EnergySpec
	gpuSpec   EnergySpec
	fecimSpec EnergySpec

	// Animation state (protected by animMu)
	animMu           sync.RWMutex
	running          bool
	paused           bool
	simTime          float64
	presentationMode PresentationMode
	currentPhase     AutoDemoPhase
	phaseTimer       float64

	// GUI components - Hero visualizations
	energyRace        *AnimatedEnergyRace
	memoryWall        *MemoryWallAnimation
	marketChart       *MarketOpportunityChart
	competitiveMatrix *CompetitiveMatrix
	phasedStrategy    *PhasedStrategyDiagram
	analogStates      *AnalogStatesComparison
	dcTransformation  *DataCenterTransformation

	// GUI components - Calculator
	calculator *DataCenterCalculator

	// GUI components - Educational
	educationalPanel *ComparisonEducationalPanel
	operationLog     *ComparisonOperationLog
	modeIndicator    *ComparisonModeIndicator

	// Controls
	workloadSelect   *widget.Select
	inferencesSlider *widget.Slider
	inferencesLabel  *widget.Label
	modeSelect       *widget.Select
	pauseBtn         *widget.Button

	// Status
	statusLabel     *widget.Label
	lastStatusText  string // Cache to avoid redundant SetText calls

	// Current settings
	currentWorkload   string
	currentInferences float64
}

// NewComparisonApp creates the comparison demo application.
func NewComparisonApp() *ComparisonApp {
	debug.Println("NewComparisonApp: Creating application")
	ca := &ComparisonApp{
		currentWorkload:   "MNIST",
		currentInferences: 10000,
	}

	ca.fyneApp = app.NewWithID("com.fecim.comparison-demo")
	ca.fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

	// Initialize energy specs (convert pJ to fJ: multiply by 1000)
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      cpuEnergyPJPerMAC * 1000, // 1,000,000 fJ/MAC
		Source:        "Intel/AMD published specs",
		Verified:      true,
		SourceDetails: "Includes memory access energy (~640 pJ for DRAM fetch + ~3-5 pJ for MAC).",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      gpuEnergyPJPerMAC * 1000, // 100,000 fJ/MAC
		Source:        "NVIDIA H100 specifications",
		Verified:      true,
		SourceDetails: "H100 SXM: 700W TDP, ~3958 TFLOPS FP16. HBM access dominates.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      fecimEnergyPJPerMAC * 1000, // 1,000 fJ/MAC
		Source:        "Dr. Tour COSM 2025 (NOT independently verified)",
		Verified:      false,
		SourceDetails: "Claimed: 'under 1 picojoule per MAC'. TRL 4 - laboratory validation only.",
	}

	debug.Println("NewComparisonApp: Initialization complete")
	return ca
}

// Run starts the GUI application.
func (ca *ComparisonApp) Run() {
	debug.Println("App: Creating window")
	ca.window = ca.fyneApp.NewWindow("FeCIM Demo 5: Architecture Comparison")
	ca.window.Resize(fyne.NewSize(1400, 900))

	content := ca.createMainLayout()
	ca.window.SetContent(content)

	ca.updateCalculations()
	ca.updateStatus("Ready. Select workload and adjust parameters.")

	// Start animation loop
	ca.animMu.Lock()
	ca.running = true
	ca.animMu.Unlock()
	go ca.animationLoop()

	debug.Println("App: ShowAndRun starting")
	ca.window.ShowAndRun()
	ca.animMu.Lock()
	ca.running = false
	ca.animMu.Unlock()
}

// animationLoop runs the main animation at 30 FPS (reduced from 60 to prevent resize loops on tiling WMs).
func (ca *ComparisonApp) animationLoop() {
	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()

	lastTime := time.Now()

	for {
		<-ticker.C

		// Check if we should stop
		ca.animMu.RLock()
		running := ca.running
		paused := ca.paused
		ca.animMu.RUnlock()

		if !running {
			return
		}

		if paused {
			lastTime = time.Now()
			continue
		}

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()

		ca.animMu.Lock()
		ca.simTime += dt
		ca.animMu.Unlock()

		// Update animated widgets
		if ca.energyRace != nil {
			ca.energyRace.UpdateAnimation(dt)
		}
		if ca.memoryWall != nil {
			ca.memoryWall.UpdateAnimation(dt)
		}
		if ca.marketChart != nil {
			ca.marketChart.UpdateAnimation(dt)
		}
		if ca.phasedStrategy != nil {
			ca.phasedStrategy.UpdateAnimation(dt)
		}
		if ca.analogStates != nil {
			ca.analogStates.UpdateAnimation(dt)
		}
		if ca.dcTransformation != nil {
			ca.dcTransformation.UpdateAnimation(dt)
		}

		// Handle auto-demo mode
		ca.animMu.Lock()
		if ca.presentationMode == PresentationModeAuto {
			ca.phaseTimer += dt
			phaseDuration := ca.currentPhase.PhaseDuration().Seconds()
			if ca.phaseTimer >= phaseDuration {
				ca.phaseTimer = 0
				ca.currentPhase = AutoDemoPhase((int(ca.currentPhase) + 1) % int(AutoDemoPhaseCount))
				ca.animMu.Unlock()
				ca.onPhaseChanged()
			} else {
				ca.animMu.Unlock()
			}
		} else {
			ca.animMu.Unlock()
		}

		// Refresh UI
		fyne.Do(func() {
			if ca.energyRace != nil {
				ca.energyRace.Refresh()
			}
			if ca.memoryWall != nil {
				ca.memoryWall.Refresh()
			}
			if ca.marketChart != nil {
				ca.marketChart.Refresh()
			}
			if ca.phasedStrategy != nil {
				ca.phasedStrategy.Refresh()
			}
			if ca.analogStates != nil {
				ca.analogStates.Refresh()
			}
			if ca.dcTransformation != nil {
				ca.dcTransformation.Refresh()
			}
			ca.updateStatusForMode()
		})
	}
}

// onPhaseChanged handles auto-demo phase transitions.
func (ca *ComparisonApp) onPhaseChanged() {
	ca.animMu.RLock()
	phase := ca.currentPhase
	ca.animMu.RUnlock()

	debug.Printf("Auto-demo phase changed to: %s", phase.String())

	if ca.educationalPanel != nil {
		ca.educationalPanel.SetPhase(phase)
	}

	if ca.phasedStrategy != nil {
		ca.phasedStrategy.SetPhase(int(phase) % 3)
	}

	switch phase {
	case AutoDemoPhaseEnergyRace:
		if ca.energyRace != nil {
			ca.energyRace.Reset()
		}
	case AutoDemoPhaseMarket:
		if ca.marketChart != nil {
			ca.marketChart.Reset()
		}
	}
}

// updateStatusForMode updates status based on current mode.
func (ca *ComparisonApp) updateStatusForMode() {
	if ca.statusLabel == nil {
		return
	}

	ca.animMu.RLock()
	mode := ca.presentationMode
	phase := ca.currentPhase
	timer := ca.phaseTimer
	ca.animMu.RUnlock()

	var newText string
	switch mode {
	case PresentationModeAuto:
		remaining := phase.PhaseDuration().Seconds() - timer
		newText = fmt.Sprintf("Auto Demo: %s (%.0fs remaining)", phase.String(), remaining)
	case PresentationModeInvestor:
		newText = "Mode: Technical Briefing"
	case PresentationModeEngineer:
		newText = "Mode: Technical Deep-Dive"
	case PresentationModeManual:
		// In manual mode, preserve the current status text (don't override calculation status)
		return
	}

	// Use caching to avoid redundant SetText calls that trigger layout recalculations
	if newText != "" && newText != ca.lastStatusText {
		ca.statusLabel.SetText(newText)
		ca.lastStatusText = newText
	}
}

// SetPresentationMode sets the current presentation mode.
func (ca *ComparisonApp) SetPresentationMode(mode PresentationMode) {
	ca.animMu.Lock()
	ca.presentationMode = mode
	ca.phaseTimer = 0
	ca.currentPhase = AutoDemoPhaseEnergyRace
	ca.animMu.Unlock()

	if ca.educationalPanel != nil {
		ca.educationalPanel.SetPresentationMode(mode)
	}

	if ca.energyRace != nil {
		ca.energyRace.Reset()
	}
	if ca.marketChart != nil {
		ca.marketChart.Reset()
	}

	debug.Printf("Presentation mode set to: %s", mode.String())
}

// createMainLayout builds the main application layout.
func (ca *ComparisonApp) createMainLayout() fyne.CanvasObject {
	// Create widgets
	ca.energyRace = NewAnimatedEnergyRace()
	ca.memoryWall = NewMemoryWallAnimation()
	ca.marketChart = NewMarketOpportunityChart()
	ca.competitiveMatrix = NewCompetitiveMatrix()
	ca.phasedStrategy = NewPhasedStrategyDiagram()
	ca.analogStates = NewAnalogStatesComparison()
	ca.dcTransformation = NewDataCenterTransformation()
	ca.calculator = NewDataCenterCalculator()
	ca.educationalPanel = NewComparisonEducationalPanel()
	ca.operationLog = NewComparisonOperationLog()
	ca.modeIndicator = NewComparisonModeIndicator()

	// Mode selector
	ca.modeSelect = widget.NewSelect(
		[]string{"Manual", "Auto Demo", "Investor", "Engineer"},
		func(s string) {
			ca.SetPresentationMode(PresentationModeFromString(s))
		},
	)
	ca.modeSelect.SetSelected("Manual")

	// Pause button
	ca.pauseBtn = widget.NewButton("Pause", func() {
		ca.animMu.Lock()
		ca.paused = !ca.paused
		paused := ca.paused
		ca.animMu.Unlock()
		if paused {
			ca.pauseBtn.SetText("Resume")
		} else {
			ca.pauseBtn.SetText("Pause")
		}
	})

	// Workload selector
	ca.workloadSelect = widget.NewSelect(
		[]string{"MNIST", "ResNet-50", "BERT-Base", "GPT-2", "LLM-70B"},
		ca.onWorkloadChanged,
	)
	ca.workloadSelect.SetSelected("MNIST")

	// Inferences slider
	ca.inferencesLabel = widget.NewLabel("Inferences/sec: 10,000")
	ca.inferencesSlider = widget.NewSlider(100, 100000)
	ca.inferencesSlider.Value = 10000
	ca.inferencesSlider.OnChanged = func(v float64) {
		ca.currentInferences = v
		ca.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", v))
		ca.updateCalculations()
	}

	// Calculate button
	var calcBtn *widget.Button
	calcBtn = widget.NewButton("Calculate", func() {
		calcBtn.Disable()
		calcBtn.SetText("Calculating...")
		go func() {
			ca.updateCalculations()
			fyne.Do(func() {
				calcBtn.Enable()
				calcBtn.SetText("Calculate")
			})
		}()
	})
	calcBtn.Importance = widget.HighImportance

	// Status
	ca.statusLabel = widget.NewLabel("Status: Ready")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// === HEADER ===
	titleLabel := widget.NewLabel("FeCIM Architecture Comparison")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	header := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		widget.NewLabel("Mode:"),
		ca.modeSelect,
		ca.pauseBtn,
	)

	// === LEFT PANEL - Narrow controls (~12%) ===
	leftConfigLabel := widget.NewLabelWithStyle("Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	leftPanel := container.NewVBox(
		leftConfigLabel,
		widget.NewSeparator(),
		widget.NewLabel("Workload:"),
		ca.workloadSelect,
		widget.NewSeparator(),
		ca.inferencesLabel,
		ca.inferencesSlider,
		widget.NewSeparator(),
		calcBtn,
	)

	// === CENTER PANEL - Expanded with better hierarchy (~68%) ===
	// HERO: Energy Race (full width, larger)
	heroEnergyRace := widget.NewCard(
		"Energy per MAC Operation",
		fmt.Sprintf("CPU: %d pJ | GPU: %d pJ | FeCIM: %.1f pJ",
			int(cpuEnergyPJPerMAC), int(gpuEnergyPJPerMAC), fecimEnergyPJPerMAC),
		container.NewPadded(ca.energyRace),
	)

	// Row 1: Calculator + Memory Wall
	row1 := container.NewGridWithColumns(2,
		widget.NewCard("Data Center Calculator", "Interactive power and cost estimates",
			container.NewPadded(container.NewVBox(ca.calculator, ca.dcTransformation))),
		widget.NewCard("Memory Wall Problem", "Data movement dominates energy",
			container.NewPadded(ca.memoryWall)),
	)

	// Row 2: Market + Competitive
	row2 := container.NewGridWithColumns(2,
		widget.NewCard("Market Opportunity", "AI semiconductor growth projection",
			container.NewPadded(ca.marketChart)),
		widget.NewCard("Competitive Landscape", "FeCIM vs alternatives",
			container.NewPadded(ca.competitiveMatrix)),
	)

	// Row 3: Strategy + Analog States
	row3 := container.NewGridWithColumns(2,
		widget.NewCard("Commercialization Strategy", "Phased market entry plan",
			container.NewPadded(ca.phasedStrategy)),
		widget.NewCard("Analog States", "30 programmable levels per cell",
			container.NewPadded(ca.analogStates)),
	)

	// Verified Claims - Collapsible accordion at bottom
	verifiedClaimsAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Verified Claims & Technology Status", ca.createVerifiedClaimsWidget()),
	)

	// Scroll indicator to help users discover content below the fold
	scrollHintLabel := widget.NewLabelWithStyle("▼ Scroll down for Data Center Calculator, Market Analysis, and more ▼", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	scrollHintLabel.Importance = widget.LowImportance

	centerPanel := container.NewVBox(
		heroEnergyRace,
		scrollHintLabel,
		widget.NewSeparator(),
		row1,
		widget.NewSeparator(),
		row2,
		widget.NewSeparator(),
		row3,
		widget.NewSeparator(),
		verifiedClaimsAccordion,
	)

	// === RIGHT PANEL - Educational + Log (~20%) ===
	sourcesLink := widget.NewHyperlink("View Sources & References", nil)
	sourcesLink.OnTapped = func() {
		ca.operationLog.Add("Sources: See docs/videos/COSM_2025_AI_Hardware_Breakthrough/")
	}

	operationLogAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Operation Log", ca.operationLog),
	)

	rightPanel := container.NewVBox(
		ca.educationalPanel,
		widget.NewSeparator(),
		operationLogAccordion,
		layout.NewSpacer(),
		sourcesLink,
	)

	// === FOOTER ===
	disclaimerIcon := widget.NewLabel("⚠")
	disclaimerIcon.TextStyle = fyne.TextStyle{Bold: true}

	disclaimerLabel := widget.NewLabel("Technology Readiness Level 4 | Laboratory Validation | Energy claims pending independent verification")
	disclaimerLabel.TextStyle = fyne.TextStyle{Bold: true}

	disclaimerContainer := container.NewHBox(
		disclaimerIcon,
		disclaimerLabel,
	)

	footer := container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		disclaimerContainer,
	)

	// === MAIN LAYOUT ===
	// Left panel: ~12% (150px fixed), Center: ~68%, Right: ~20% (250px fixed)
	leftScroll := container.NewScroll(container.NewPadded(leftPanel))
	leftScroll.SetMinSize(fyne.NewSize(150, 0))

	centerScroll := container.NewScroll(container.NewPadded(centerPanel))

	rightScroll := container.NewScroll(container.NewPadded(rightPanel))
	rightScroll.SetMinSize(fyne.NewSize(250, 0))

	mainContent := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), footer),
		leftScroll,
		rightScroll,
		centerScroll,
	)

	return mainContent
}

// createVerifiedClaimsWidget creates a compact verified/claimed section.
// Dr. Tour recommendation: Show explicit energy numbers with units and citations
func (ca *ComparisonApp) createVerifiedClaimsWidget() fyne.CanvasObject {
	verifiedLabel := widget.NewLabelWithStyle("VERIFIED:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	verifiedItems := widget.NewLabel("• 30 analog levels\n• 87% MNIST accuracy\n• CMOS compatible\n• Non-volatile")

	// Explicit energy numbers with units (Dr. Tour recommendation)
	energyLabel := widget.NewLabelWithStyle("ENERGY/MAC (pJ):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	energyItems := widget.NewLabel(fmt.Sprintf(
		"• CPU+DRAM: %d pJ ✓\n• GPU+HBM: %d pJ ✓\n• FeCIM: ~%.1f pJ (TRL4)",
		int(cpuEnergyPJPerMAC), int(gpuEnergyPJPerMAC), fecimEnergyPJPerMAC))

	claimedLabel := widget.NewLabelWithStyle("CLAIMED (not verified):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	claimedItems := widget.NewLabel(fmt.Sprintf(
		"• %d× less vs CPU\n• %d× less vs GPU\n• 80-90%% DC savings",
		int(cpuEnergyPJPerMAC/fecimEnergyPJPerMAC),
		int(gpuEnergyPJPerMAC/fecimEnergyPJPerMAC)))

	trlLabel := widget.NewLabelWithStyle("Status: TRL 4 (Lab only)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true})

	return container.NewVBox(
		verifiedLabel,
		verifiedItems,
		widget.NewSeparator(),
		energyLabel,
		energyItems,
		widget.NewSeparator(),
		claimedLabel,
		claimedItems,
		widget.NewSeparator(),
		trlLabel,
	)
}

// onWorkloadChanged handles workload selection.
func (ca *ComparisonApp) onWorkloadChanged(workload string) {
	ca.currentWorkload = workload
	ca.operationLog.Add(fmt.Sprintf("Workload: %s", workload))
	ca.updateCalculations()
}

// updateCalculations recalculates all values.
func (ca *ComparisonApp) updateCalculations() {
	debug.Printf("updateCalculations: workload=%s, inferences=%.0f", ca.currentWorkload, ca.currentInferences)

	macs := ca.getWorkloadMACs()

	// Calculate energy per inference (µJ) = MACs × fJ/MAC / 1e9
	cpuEnergy := float64(macs) * ca.cpuSpec.EnergyFJ / 1e9
	gpuEnergy := float64(macs) * ca.gpuSpec.EnergyFJ / 1e9
	fecimEnergy := float64(macs) * ca.fecimSpec.EnergyFJ / 1e9

	// Calculate power (W) = µJ/inf × inf/s / 1e6
	cpuPower := cpuEnergy * ca.currentInferences / 1e6
	gpuPower := gpuEnergy * ca.currentInferences / 1e6
	fecimPower := fecimEnergy * ca.currentInferences / 1e6

	// Monthly cost at $0.10/kWh
	hoursPerMonth := 730.0
	cpuCost := cpuPower / 1000 * hoursPerMonth * 0.10
	gpuCost := gpuPower / 1000 * hoursPerMonth * 0.10
	fecimCost := fecimPower / 1000 * hoursPerMonth * 0.10

	// Update calculator
	ca.calculator.SetResults(
		ca.currentWorkload, macs, ca.currentInferences,
		cpuEnergy, gpuEnergy, fecimEnergy,
		cpuPower, gpuPower, fecimPower,
		cpuCost, gpuCost, fecimCost,
	)

	// Update educational panel
	ca.educationalPanel.SetComparison(cpuPower/fecimPower, gpuPower/fecimPower)

	// Update transformation widget
	ca.dcTransformation.SetValues(gpuPower, fecimPower)

	// Log
	ca.operationLog.Add(fmt.Sprintf("Calculated: %.0f MACs × %.0f inf/s", float64(macs), ca.currentInferences))

	ca.modeIndicator.SetMode(ComparisonModeCalculating)
	ca.updateStatus(fmt.Sprintf("Calculated for %s @ %.0f inf/s", ca.currentWorkload, ca.currentInferences))

	go func() {
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			if ca.modeIndicator != nil {
				ca.modeIndicator.SetMode(ComparisonModeIdle)
			}
		})
	}()
}

// getWorkloadMACs returns MACs per inference for common neural network workloads.
// Sources: Published architecture specifications and measured inference costs.
func (ca *ComparisonApp) getWorkloadMACs() int {
	switch ca.currentWorkload {
	case "MNIST":
		// Simple 2-layer MLP: 784 input → 128 hidden → 10 output
		return 101632 // (784×128) + (128×10)
	case "ResNet-50":
		// Deep residual network for image classification
		return 4000000000 // ~4 GMACs
	case "BERT-Base":
		// Transformer for NLP (sequence length 512)
		return 11000000000 // ~11 GMACs
	case "GPT-2":
		// Large language model (117M parameters)
		return 35000000000 // ~35 GMACs
	case "LLM-70B":
		// Llama-2-70B or similar large model
		return 140000000000000 // ~140 TMACs
	default:
		return 101632
	}
}

// updateStatus updates the status label.
func (ca *ComparisonApp) updateStatus(status string) {
	if ca.statusLabel == nil {
		return
	}
	ca.statusLabel.SetText("Status: " + status)
}
