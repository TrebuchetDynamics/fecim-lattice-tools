// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var debug *log.Logger
var logFile *os.File

func init() {
	logsDir := "<local-path>"
	os.MkdirAll(logsDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-comparison-module05.log")

	var err error
	logFile, err = os.Create(logPath)
	if err != nil {
		debug = log.New(os.Stdout, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
		debug.Printf("Failed to create log file: %v, using stdout", err)
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	debug = log.New(multiWriter, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
	debug.Printf("Logging to: %s", logPath)
}

// FeCIM theme colors
var (
	colorBackground = color.RGBA{0, 50, 100, 255}
	colorPrimary    = color.RGBA{0, 212, 255, 255}
)

type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// EnergySpec holds energy per MAC specifications with sources.
type EnergySpec struct {
	Name          string
	EnergyFJ      float64 // femtojoules per MAC
	Source        string
	Verified      bool
	SourceDetails string
}

// ComparisonApp is the main application for architecture comparison.
type ComparisonApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Energy specs (honest numbers with sources)
	cpuSpec   EnergySpec
	gpuSpec   EnergySpec
	fecimSpec EnergySpec

	// Animation state
	running          bool
	paused           bool
	simTime          float64
	presentationMode PresentationMode
	currentPhase     AutoDemoPhase
	phaseTimer       float64

	// GUI components - Original
	energyChart      *EnergyBarChart
	archDiagram      *ArchitectureDiagram
	calculator       *DataCenterCalculator
	verifiedTable    *VerifiedClaimsTable
	educationalPanel *ComparisonEducationalPanel
	operationLog     *ComparisonOperationLog
	modeIndicator    *ComparisonModeIndicator

	// GUI components - New hero visualizations
	energyRace       *AnimatedEnergyRace
	memoryWall       *MemoryWallAnimation
	marketChart      *MarketOpportunityChart
	competitiveMatrix *CompetitiveMatrix
	phasedStrategy   *PhasedStrategyDiagram
	analogStates     *AnalogStatesComparison
	weebitCard       *WeebitNanoCard
	dcTransformation *DataCenterTransformation

	// Controls
	workloadSelect   *widget.Select
	inferencesSlider *widget.Slider
	inferencesLabel  *widget.Label
	modeSelect       *widget.Select
	pauseBtn         *widget.Button

	// Status
	statusLabel *widget.Label

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
	ca.fyneApp.Settings().SetTheme(&feCIMTheme{})

	// Initialize energy specs with HONEST numbers and sources
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      1000, // ~1000 fJ/MAC
		Source:        "Intel/AMD published specs",
		Verified:      true,
		SourceDetails: "Includes memory access energy. Intel Xeon specs, AMD EPYC specs.",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      100, // ~100 fJ/MAC
		Source:        "NVIDIA H100 specifications",
		Verified:      true,
		SourceDetails: "H100 SXM: 700W TDP, ~3958 TFLOPS FP16. ~177 fJ/FLOP.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      10, // ~1-10 fJ/MAC (claimed)
		Source:        "Dr. Tour's presentation (NOT independently verified)",
		Verified:      false,
		SourceDetails: "Claimed: '10M× lower energy than NAND'. TRL 4 - lab only.",
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
	ca.running = true
	go ca.animationLoop()

	debug.Println("App: ShowAndRun starting")
	ca.window.ShowAndRun()
	ca.running = false
}

// animationLoop runs the main animation at 60 FPS.
func (ca *ComparisonApp) animationLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	lastTime := time.Now()

	for ca.running {
		<-ticker.C

		if ca.paused {
			lastTime = time.Now()
			continue
		}

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		ca.simTime += dt

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

		// Handle auto-demo mode phase transitions
		if ca.presentationMode == PresentationModeAuto {
			ca.phaseTimer += dt
			phaseDuration := ca.currentPhase.PhaseDuration().Seconds()
			if ca.phaseTimer >= phaseDuration {
				ca.phaseTimer = 0
				ca.currentPhase = AutoDemoPhase((int(ca.currentPhase) + 1) % int(AutoDemoPhaseCount))
				ca.onPhaseChanged()
			}
		}

		// Refresh UI on main thread
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
	debug.Printf("Auto-demo phase changed to: %s", ca.currentPhase.String())

	// Update educational panel for new phase
	if ca.educationalPanel != nil {
		ca.educationalPanel.SetPhase(ca.currentPhase)
	}

	// Update phased strategy diagram
	if ca.phasedStrategy != nil {
		ca.phasedStrategy.SetPhase(int(ca.currentPhase) % 3)
	}

	// Reset animations for certain phases
	switch ca.currentPhase {
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

// updateStatusForMode updates the status based on current mode.
func (ca *ComparisonApp) updateStatusForMode() {
	if ca.statusLabel == nil {
		return
	}

	switch ca.presentationMode {
	case PresentationModeAuto:
		remaining := ca.currentPhase.PhaseDuration().Seconds() - ca.phaseTimer
		ca.statusLabel.SetText(fmt.Sprintf("Auto Demo: %s (%.0fs remaining)", ca.currentPhase.String(), remaining))
	case PresentationModeInvestor:
		ca.statusLabel.SetText("Mode: Technical Briefing")
	case PresentationModeEngineer:
		ca.statusLabel.SetText("Mode: Technical Deep-Dive")
	default:
		// Manual mode - keep existing status
	}
}

// SetPresentationMode sets the current presentation mode.
func (ca *ComparisonApp) SetPresentationMode(mode PresentationMode) {
	ca.presentationMode = mode
	ca.phaseTimer = 0
	ca.currentPhase = AutoDemoPhaseEnergyRace

	// Update educational panel
	if ca.educationalPanel != nil {
		ca.educationalPanel.SetPresentationMode(mode)
	}

	// Reset animations
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
	// Create original components
	ca.energyChart = NewEnergyBarChart()
	ca.archDiagram = NewArchitectureDiagram()
	ca.calculator = NewDataCenterCalculator()
	ca.verifiedTable = NewVerifiedClaimsTable()
	ca.educationalPanel = NewComparisonEducationalPanel()
	ca.operationLog = NewComparisonOperationLog()
	ca.modeIndicator = NewComparisonModeIndicator()

	// Create new hero visualizations
	ca.energyRace = NewAnimatedEnergyRace()
	ca.memoryWall = NewMemoryWallAnimation()
	ca.marketChart = NewMarketOpportunityChart()
	ca.competitiveMatrix = NewCompetitiveMatrix()
	ca.phasedStrategy = NewPhasedStrategyDiagram()
	ca.analogStates = NewAnalogStatesComparison()
	ca.weebitCard = NewWeebitNanoCard()
	ca.dcTransformation = NewDataCenterTransformation()

	// Set initial energy values
	ca.energyChart.SetValues(ca.cpuSpec, ca.gpuSpec, ca.fecimSpec)

	// Mode selector
	ca.modeSelect = widget.NewSelect(
		[]string{"Manual", "Auto Demo", "Investor", "Engineer"},
		func(s string) {
			mode := PresentationModeFromString(s)
			ca.SetPresentationMode(mode)
		},
	)
	ca.modeSelect.SetSelected("Manual")

	// Pause button
	ca.pauseBtn = widget.NewButton("Pause", func() {
		ca.paused = !ca.paused
		if ca.paused {
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

	// Status
	ca.statusLabel = widget.NewLabel("Status: Ready")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Calculate button
	calcBtn := widget.NewButton("Calculate", func() {
		ca.updateCalculations()
	})
	calcBtn.Importance = widget.HighImportance

	// Header with mode selector
	titleLabel := widget.NewLabel("FeCIM Architecture Comparison")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	modeLabel := widget.NewLabel("Mode:")
	header := container.NewVBox(
		container.NewHBox(
			titleLabel,
			layout.NewSpacer(),
			modeLabel,
			ca.modeSelect,
			ca.pauseBtn,
		),
		widget.NewSeparator(),
	)

	// Left panel: Controls + Verified Claims + Weebit Card
	controlsLabel := widget.NewLabel("Configuration")
	controlsLabel.TextStyle = fyne.TextStyle{Bold: true}

	leftPanel := container.NewVBox(
		controlsLabel,
		widget.NewSeparator(),
		widget.NewLabel("Workload:"),
		ca.workloadSelect,
		widget.NewSeparator(),
		ca.inferencesLabel,
		ca.inferencesSlider,
		widget.NewSeparator(),
		calcBtn,
		widget.NewSeparator(),
		ca.verifiedTable,
		widget.NewSeparator(),
		ca.weebitCard,
	)

	// Center panel: Hero visualizations + Charts
	heroSection := container.NewVBox(
		widget.NewLabelWithStyle("Energy Race", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ca.energyRace,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Memory Wall Problem", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ca.memoryWall,
	)

	marketSection := container.NewVBox(
		widget.NewSeparator(),
		ca.marketChart,
	)

	competitiveSection := container.NewHBox(
		container.NewVBox(ca.competitiveMatrix),
		container.NewVBox(ca.phasedStrategy),
	)

	calculatorSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Data Center Calculator", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ca.calculator,
		ca.dcTransformation,
	)

	analogSection := container.NewVBox(
		widget.NewSeparator(),
		ca.analogStates,
	)

	centerPanel := container.NewVBox(
		heroSection,
		marketSection,
		competitiveSection,
		calculatorSection,
		analogSection,
	)

	// Right panel: Educational + Log
	rightPanel := container.NewVBox(
		ca.educationalPanel,
		widget.NewSeparator(),
		ca.operationLog,
	)

	// Footer
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			ca.modeIndicator,
			widget.NewSeparator(),
			ca.statusLabel,
			layout.NewSpacer(),
			widget.NewLabel("TRL 4 | Lab Validation Only | Sources in footnotes"),
		),
	)

	// Main layout using HSplit
	leftCenterSplit := container.NewHSplit(
		container.NewScroll(container.NewPadded(leftPanel)),
		container.NewScroll(centerPanel),
	)
	leftCenterSplit.SetOffset(0.20)

	mainSplit := container.NewHSplit(
		leftCenterSplit,
		container.NewScroll(container.NewPadded(rightPanel)),
	)
	mainSplit.SetOffset(0.75)

	mainContent := container.NewBorder(
		header,
		footer,
		nil,
		nil,
		mainSplit,
	)

	return mainContent
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

	// Get MACs for workload
	macs := ca.getWorkloadMACs()

	// Calculate energy per inference (µJ)
	cpuEnergy := float64(macs) * ca.cpuSpec.EnergyFJ / 1e9 // fJ to µJ
	gpuEnergy := float64(macs) * ca.gpuSpec.EnergyFJ / 1e9
	fecimEnergy := float64(macs) * ca.fecimSpec.EnergyFJ / 1e9

	// Calculate power for target inferences/sec (W)
	cpuPower := cpuEnergy * ca.currentInferences / 1e6 // µJ * inf/s = µW, /1e6 = W
	gpuPower := gpuEnergy * ca.currentInferences / 1e6
	fecimPower := fecimEnergy * ca.currentInferences / 1e6

	// Monthly cost at $0.10/kWh
	hoursPerMonth := 730.0
	cpuCost := cpuPower / 1000 * hoursPerMonth * 0.10
	gpuCost := gpuPower / 1000 * hoursPerMonth * 0.10
	fecimCost := fecimPower / 1000 * hoursPerMonth * 0.10

	// Update calculator
	ca.calculator.SetResults(
		ca.currentWorkload,
		macs,
		ca.currentInferences,
		cpuEnergy, gpuEnergy, fecimEnergy,
		cpuPower, gpuPower, fecimPower,
		cpuCost, gpuCost, fecimCost,
	)

	// Update educational panel
	ca.educationalPanel.SetComparison(
		cpuPower/fecimPower,
		gpuPower/fecimPower,
	)

	// Log
	ca.operationLog.Add(fmt.Sprintf("Calculated: %.0f MACs × %.0f inf/s", float64(macs), ca.currentInferences))
	ca.operationLog.Add(fmt.Sprintf("  CPU: %.1fW, GPU: %.1fW, FeCIM: %.2fW", cpuPower, gpuPower, fecimPower))

	ca.modeIndicator.SetMode(ComparisonModeCalculating)
	ca.updateStatus(fmt.Sprintf("Calculated for %s @ %.0f inf/s", ca.currentWorkload, ca.currentInferences))

	// Reset mode after brief delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.modeIndicator.SetMode(ComparisonModeIdle)
		})
	}()
}

// getWorkloadMACs returns MACs for the current workload.
func (ca *ComparisonApp) getWorkloadMACs() int {
	switch ca.currentWorkload {
	case "MNIST":
		return 101632 // 784*128 + 128*10
	case "ResNet-50":
		return 4000000000 // ~4B MACs
	case "BERT-Base":
		return 11000000000 // ~11B MACs
	case "GPT-2":
		return 35000000000 // ~35B MACs
	case "LLM-70B":
		return 140000000000000 // ~140T MACs
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
