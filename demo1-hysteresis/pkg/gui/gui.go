// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// Uses fyne.io/fyne/v2 for cross-platform native GUI with proper graphics.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ironlattice-vis/demo1-hysteresis/pkg/ferroelectric"
)

// Colors - IronLattice theme
var (
	colorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan
	colorSecondary  = color.RGBA{255, 107, 107, 255} // Coral red
	colorAccent     = color.RGBA{78, 205, 196, 255}  // Teal
	colorWarning    = color.RGBA{255, 230, 109, 255} // Yellow
	colorBackground = color.RGBA{26, 26, 46, 255}    // Dark blue
	colorGrid       = color.RGBA{60, 60, 80, 128}    // Grid lines
	colorAxis       = color.RGBA{150, 150, 150, 255} // Axis lines
	colorPositive   = color.RGBA{255, 100, 100, 255} // Positive polarization
	colorNegative   = color.RGBA{100, 150, 255, 255} // Negative polarization
)

// App holds the main application state
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window

	// Physics
	material  *ferroelectric.HZOMaterial
	preisach  *ferroelectric.MayergoyzPreisach
	materials []*ferroelectric.HZOMaterial
	matIndex  int

	// Simulation state
	mu            sync.RWMutex
	electricField float64
	polarization  float64
	normalizedP   float64
	discreteLevel int

	// History for plotting
	eHistory   []float64
	pHistory   []float64
	maxHistory int

	// UI state
	running   bool
	paused    bool
	autoMode  bool
	waveform  WaveformType
	frequency float64
	simTime   float64

	// UI components
	plot           *PEPlot
	levelIndicator *LevelIndicator
	eFieldSlider   *widget.Slider
	eFieldLabel    *widget.Label
	pLabel         *widget.Label
	levelLabel     *widget.Label
	materialSelect *widget.Select
	waveformSelect *widget.Select
	statusLabel    *widget.Label
	pauseBtn       *widget.Button
}

// WaveformType represents the input waveform
type WaveformType int

const (
	WaveformManual WaveformType = iota
	WaveformSine
	WaveformTriangle
	WaveformSquare
)

func (w WaveformType) String() string {
	switch w {
	case WaveformManual:
		return "Manual"
	case WaveformSine:
		return "Sine Wave"
	case WaveformTriangle:
		return "Triangle Wave"
	case WaveformSquare:
		return "Square Wave"
	default:
		return "Unknown"
	}
}

// NewApp creates a new GUI application
func NewApp() *App {
	materials := []*ferroelectric.HZOMaterial{
		ferroelectric.DefaultHZO(),
		ferroelectric.OptimizedHZO(),
		ferroelectric.IronLatticeMaterial(),
	}

	mat := materials[0]
	preisach := ferroelectric.NewMayergoyzPreisach(mat, 30)

	return &App{
		material:   mat,
		preisach:   preisach,
		materials:  materials,
		matIndex:   0,
		maxHistory: 500,
		eHistory:   make([]float64, 0, 500),
		pHistory:   make([]float64, 0, 500),
		autoMode:   true,
		waveform:   WaveformSine,
		frequency:  1.0, // 1 Hz for smooth animation
	}
}

// Run starts the GUI application
func Run() error {
	a := NewApp()
	return a.run()
}

func (a *App) run() error {
	a.fyneApp = app.New()
	a.fyneApp.Settings().SetTheme(&ironLatticeTheme{})

	a.mainWindow = a.fyneApp.NewWindow("IronLattice Hysteresis Visualizer - Demo 1")
	a.mainWindow.Resize(fyne.NewSize(1200, 800))

	// Create UI components
	content := a.createUI()
	a.mainWindow.SetContent(content)

	// Start simulation loop
	a.running = true
	go a.simulationLoop()

	a.mainWindow.ShowAndRun()
	a.running = false
	return nil
}

func (a *App) createUI() fyne.CanvasObject {
	// Create P-E plot
	a.plot = NewPEPlot(a.material.Ec*2.5, a.material.Ps*1.2)
	a.plot.SetMinSize(fyne.NewSize(600, 500))

	// Create level indicator
	a.levelIndicator = NewLevelIndicator()
	a.levelIndicator.SetMinSize(fyne.NewSize(80, 500))

	// Create controls panel
	controls := a.createControlsPanel()

	// Create info panel
	info := a.createInfoPanel()

	// Layout: [Plot | Level | Controls/Info]
	plotContainer := container.NewBorder(
		widget.NewLabelWithStyle("P-E Hysteresis Loop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		a.plot,
	)

	levelContainer := container.NewBorder(
		widget.NewLabelWithStyle("30 Levels", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%.1f bits/cell", math.Log2(30))),
		nil, nil,
		a.levelIndicator,
	)

	rightPanel := container.NewVBox(
		controls,
		widget.NewSeparator(),
		info,
	)

	mainLayout := container.NewHBox(
		plotContainer,
		widget.NewSeparator(),
		levelContainer,
		widget.NewSeparator(),
		rightPanel,
	)

	// Status bar at bottom
	a.statusLabel = widget.NewLabel("Running...")
	statusBar := container.NewHBox(
		layout.NewSpacer(),
		a.statusLabel,
		layout.NewSpacer(),
	)

	return container.NewBorder(
		a.createHeader(),
		statusBar,
		nil, nil,
		mainLayout,
	)
}

func (a *App) createHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"IronLattice Ferroelectric Hysteresis Visualization",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	subtitle := widget.NewLabelWithStyle(
		"\"It's got 30 discrete states. So it's not 0-1-0-1.\" — Dr. external research group",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	return container.NewVBox(title, subtitle, widget.NewSeparator())
}

func (a *App) createControlsPanel() fyne.CanvasObject {
	// E-field slider
	a.eFieldSlider = widget.NewSlider(-2, 2)
	a.eFieldSlider.Step = 0.01
	a.eFieldSlider.Value = 0
	a.eFieldSlider.OnChanged = func(v float64) {
		if a.waveform == WaveformManual {
			a.mu.Lock()
			a.electricField = v * a.material.Ec
			a.mu.Unlock()
		}
	}
	a.eFieldLabel = widget.NewLabel("E-field: 0.00 MV/cm")

	// Waveform selector
	waveforms := []string{"Manual", "Sine Wave", "Triangle Wave", "Square Wave"}
	a.waveformSelect = widget.NewSelect(waveforms, func(s string) {
		switch s {
		case "Manual":
			a.waveform = WaveformManual
			a.autoMode = false
			a.eFieldSlider.Enable()
		case "Sine Wave":
			a.waveform = WaveformSine
			a.autoMode = true
			a.eFieldSlider.Disable()
		case "Triangle Wave":
			a.waveform = WaveformTriangle
			a.autoMode = true
			a.eFieldSlider.Disable()
		case "Square Wave":
			a.waveform = WaveformSquare
			a.autoMode = true
			a.eFieldSlider.Disable()
		}
	})
	a.waveformSelect.SetSelected("Sine Wave")

	// Material selector
	matNames := []string{"Default HZO", "Optimized Superlattice", "IronLattice HZO"}
	a.materialSelect = widget.NewSelect(matNames, func(s string) {
		var idx int
		switch s {
		case "Default HZO":
			idx = 0
		case "Optimized Superlattice":
			idx = 1
		case "IronLattice HZO":
			idx = 2
		}
		a.mu.Lock()
		a.matIndex = idx
		a.material = a.materials[idx]
		a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 30)
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.plot.SetBounds(a.material.Ec*2.5, a.material.Ps*1.2)
		a.mu.Unlock()
	})
	a.materialSelect.SetSelected("Default HZO")

	// Pause/Resume button
	a.pauseBtn = widget.NewButton("Pause", func() {
		a.paused = !a.paused
		if a.paused {
			a.pauseBtn.SetText("Resume")
		} else {
			a.pauseBtn.SetText("Pause")
		}
	})

	// Reset button
	resetBtn := widget.NewButton("Reset", func() {
		a.mu.Lock()
		a.preisach.Reset()
		a.electricField = 0
		a.polarization = 0
		a.normalizedP = 0
		a.discreteLevel = 15
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.eFieldSlider.SetValue(0)
		a.mu.Unlock()
	})

	// Frequency slider
	freqSlider := widget.NewSlider(0.1, 5.0)
	freqSlider.Step = 0.1
	freqSlider.Value = 1.0
	freqLabel := widget.NewLabel("Frequency: 1.0 Hz")
	freqSlider.OnChanged = func(v float64) {
		a.frequency = v
		freqLabel.SetText(fmt.Sprintf("Frequency: %.1f Hz", v))
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("Controls", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Material:"),
		a.materialSelect,
		widget.NewSeparator(),
		widget.NewLabel("Waveform:"),
		a.waveformSelect,
		widget.NewSeparator(),
		widget.NewLabel("E-field (×Ec):"),
		a.eFieldSlider,
		a.eFieldLabel,
		widget.NewSeparator(),
		freqLabel,
		freqSlider,
		widget.NewSeparator(),
		container.NewHBox(a.pauseBtn, resetBtn),
	)
}

func (a *App) createInfoPanel() fyne.CanvasObject {
	a.pLabel = widget.NewLabel("P: 0.00 µC/cm²")
	a.levelLabel = widget.NewLabel("Level: 15/30")

	return container.NewVBox(
		widget.NewLabelWithStyle("Current State", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		a.eFieldLabel,
		a.pLabel,
		a.levelLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Material Parameters", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Pr: %.1f µC/cm²", a.material.Pr*100)),
		widget.NewLabel(fmt.Sprintf("Ps: %.1f µC/cm²", a.material.Ps*100)),
		widget.NewLabel(fmt.Sprintf("Ec: %.2f MV/cm", a.material.Ec/1e8)),
		widget.NewLabel(fmt.Sprintf("τ: %.1f ns", a.material.Tau*1e9)),
		widget.NewLabel(fmt.Sprintf("Endurance: %.0e", a.material.EnduranceCycles)),
	)
}

func (a *App) simulationLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	lastTime := time.Now()

	for a.running {
		<-ticker.C

		if a.paused {
			continue
		}

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		a.simTime += dt

		a.mu.Lock()

		// Generate E-field based on waveform
		if a.autoMode && a.waveform != WaveformManual {
			Emax := a.material.Ec * 2
			phase := 2 * math.Pi * a.frequency * a.simTime

			switch a.waveform {
			case WaveformSine:
				a.electricField = Emax * math.Sin(phase)
			case WaveformTriangle:
				p := math.Mod(phase, 2*math.Pi) / (2 * math.Pi)
				if p < 0.25 {
					a.electricField = Emax * (4 * p)
				} else if p < 0.75 {
					a.electricField = Emax * (2 - 4*p)
				} else {
					a.electricField = Emax * (4*p - 4)
				}
			case WaveformSquare:
				if math.Sin(phase) >= 0 {
					a.electricField = Emax
				} else {
					a.electricField = -Emax
				}
			}
		}

		// Update physics
		a.polarization = a.preisach.Update(a.electricField)
		a.normalizedP = a.preisach.NormalizedPolarization()
		a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * 29))
		if a.discreteLevel < 0 {
			a.discreteLevel = 0
		}
		if a.discreteLevel > 29 {
			a.discreteLevel = 29
		}

		// Record history
		a.eHistory = append(a.eHistory, a.electricField)
		a.pHistory = append(a.pHistory, a.polarization)
		if len(a.eHistory) > a.maxHistory {
			a.eHistory = a.eHistory[1:]
			a.pHistory = a.pHistory[1:]
		}

		// Copy data for UI update
		eField := a.electricField
		pol := a.polarization
		level := a.discreteLevel
		eHist := make([]float64, len(a.eHistory))
		pHist := make([]float64, len(a.pHistory))
		copy(eHist, a.eHistory)
		copy(pHist, a.pHistory)

		a.mu.Unlock()

		// Update UI (must be on main thread)
		a.updateUI(eField, pol, level, eHist, pHist)
	}
}

func (a *App) updateUI(eField, pol float64, level int, eHist, pHist []float64) {
	// Update labels
	a.eFieldLabel.SetText(fmt.Sprintf("E-field: %.3f MV/cm", eField/1e8))
	a.pLabel.SetText(fmt.Sprintf("P: %.2f µC/cm²", pol*100))
	a.levelLabel.SetText(fmt.Sprintf("Level: %d/30", level+1))

	// Update slider position for auto modes
	if a.autoMode {
		a.eFieldSlider.SetValue(eField / a.material.Ec)
	}

	// Update status
	if a.paused {
		a.statusLabel.SetText("⏸ Paused")
	} else {
		frac := a.preisach.GetSwitchedFraction() * 100
		a.statusLabel.SetText(fmt.Sprintf("● Running | t=%.2fs | Switched: %.1f%%", a.simTime, frac))
	}

	// Update plot
	a.plot.SetData(eHist, pHist, eField, pol)
	a.plot.Refresh()

	// Update level indicator
	a.levelIndicator.SetLevel(level)
	a.levelIndicator.Refresh()
}

// ============================================================
// Custom P-E Plot Widget
// ============================================================

// PEPlot is a custom widget for drawing P-E hysteresis curves
type PEPlot struct {
	widget.BaseWidget

	mu       sync.RWMutex
	eData    []float64
	pData    []float64
	currentE float64
	currentP float64
	eMax     float64
	pMax     float64
	minSize  fyne.Size
}

// NewPEPlot creates a new P-E plot widget
func NewPEPlot(eMax, pMax float64) *PEPlot {
	p := &PEPlot{
		eMax:    eMax,
		pMax:    pMax,
		minSize: fyne.NewSize(400, 300),
	}
	p.ExtendBaseWidget(p)
	return p
}

func (p *PEPlot) SetMinSize(size fyne.Size) {
	p.minSize = size
}

func (p *PEPlot) MinSize() fyne.Size {
	return p.minSize
}

func (p *PEPlot) SetBounds(eMax, pMax float64) {
	p.mu.Lock()
	p.eMax = eMax
	p.pMax = pMax
	p.mu.Unlock()
}

func (p *PEPlot) SetData(eData, pData []float64, currentE, currentP float64) {
	p.mu.Lock()
	p.eData = eData
	p.pData = pData
	p.currentE = currentE
	p.currentP = currentP
	p.mu.Unlock()
}

func (p *PEPlot) CreateRenderer() fyne.WidgetRenderer {
	return &peplotRenderer{plot: p}
}

type peplotRenderer struct {
	plot    *PEPlot
	objects []fyne.CanvasObject
}

func (r *peplotRenderer) MinSize() fyne.Size {
	return r.plot.minSize
}

func (r *peplotRenderer) Layout(size fyne.Size) {
	// Layout is handled in Refresh
}

func (r *peplotRenderer) Refresh() {
	r.plot.mu.RLock()
	defer r.plot.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.plot.Size()

	// Background
	bg := canvas.NewRectangle(colorBackground)
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Margins
	margin := float32(40)
	plotW := size.Width - 2*margin
	plotH := size.Height - 2*margin

	// Grid lines
	for i := 0; i <= 10; i++ {
		t := float32(i) / 10.0

		// Vertical grid line
		x := margin + t*plotW
		vLine := canvas.NewLine(colorGrid)
		vLine.Position1 = fyne.NewPos(x, margin)
		vLine.Position2 = fyne.NewPos(x, margin+plotH)
		vLine.StrokeWidth = 1
		r.objects = append(r.objects, vLine)

		// Horizontal grid line
		y := margin + t*plotH
		hLine := canvas.NewLine(colorGrid)
		hLine.Position1 = fyne.NewPos(margin, y)
		hLine.Position2 = fyne.NewPos(margin+plotW, y)
		hLine.StrokeWidth = 1
		r.objects = append(r.objects, hLine)
	}

	// Axes
	centerX := margin + plotW/2
	centerY := margin + plotH/2

	xAxis := canvas.NewLine(colorAxis)
	xAxis.Position1 = fyne.NewPos(margin, centerY)
	xAxis.Position2 = fyne.NewPos(margin+plotW, centerY)
	xAxis.StrokeWidth = 2
	r.objects = append(r.objects, xAxis)

	yAxis := canvas.NewLine(colorAxis)
	yAxis.Position1 = fyne.NewPos(centerX, margin)
	yAxis.Position2 = fyne.NewPos(centerX, margin+plotH)
	yAxis.StrokeWidth = 2
	r.objects = append(r.objects, yAxis)

	// Axis labels
	eLabel := canvas.NewText(fmt.Sprintf("E (MV/cm) [±%.1f]", r.plot.eMax/1e8), colorAxis)
	eLabel.TextSize = 12
	eLabel.Move(fyne.NewPos(margin+plotW-80, centerY+5))
	r.objects = append(r.objects, eLabel)

	pLabel := canvas.NewText(fmt.Sprintf("P (µC/cm²) [±%.1f]", r.plot.pMax*100), colorAxis)
	pLabel.TextSize = 12
	pLabel.Move(fyne.NewPos(centerX+5, margin))
	r.objects = append(r.objects, pLabel)

	// Plot the hysteresis data
	if len(r.plot.eData) > 1 {
		for i := 1; i < len(r.plot.eData); i++ {
			// Map data to screen coordinates
			x1 := margin + plotW/2 + float32(r.plot.eData[i-1]/r.plot.eMax)*plotW/2
			y1 := centerY - float32(r.plot.pData[i-1]/r.plot.pMax)*plotH/2
			x2 := margin + plotW/2 + float32(r.plot.eData[i]/r.plot.eMax)*plotW/2
			y2 := centerY - float32(r.plot.pData[i]/r.plot.pMax)*plotH/2

			// Color based on age (fade effect)
			age := float64(len(r.plot.eData)-i) / float64(len(r.plot.eData))
			alpha := uint8(255 * (1 - age*0.7))

			var lineColor color.RGBA
			if r.plot.pData[i] >= 0 {
				lineColor = color.RGBA{colorPositive.R, colorPositive.G, colorPositive.B, alpha}
			} else {
				lineColor = color.RGBA{colorNegative.R, colorNegative.G, colorNegative.B, alpha}
			}

			line := canvas.NewLine(lineColor)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			line.StrokeWidth = 2
			r.objects = append(r.objects, line)
		}
	}

	// Current position marker
	markerX := margin + plotW/2 + float32(r.plot.currentE/r.plot.eMax)*plotW/2
	markerY := centerY - float32(r.plot.currentP/r.plot.pMax)*plotH/2

	marker := canvas.NewCircle(colorWarning)
	marker.Resize(fyne.NewSize(12, 12))
	marker.Move(fyne.NewPos(markerX-6, markerY-6))
	r.objects = append(r.objects, marker)

	markerInner := canvas.NewCircle(colorBackground)
	markerInner.Resize(fyne.NewSize(6, 6))
	markerInner.Move(fyne.NewPos(markerX-3, markerY-3))
	r.objects = append(r.objects, markerInner)
}

func (r *peplotRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *peplotRenderer) Destroy() {}

// ============================================================
// Custom Level Indicator Widget
// ============================================================

// LevelIndicator shows the 30 discrete states
type LevelIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	level   int
	minSize fyne.Size
}

// NewLevelIndicator creates a new level indicator
func NewLevelIndicator() *LevelIndicator {
	l := &LevelIndicator{
		level:   15,
		minSize: fyne.NewSize(60, 400),
	}
	l.ExtendBaseWidget(l)
	return l
}

func (l *LevelIndicator) SetMinSize(size fyne.Size) {
	l.minSize = size
}

func (l *LevelIndicator) MinSize() fyne.Size {
	return l.minSize
}

func (l *LevelIndicator) SetLevel(level int) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

func (l *LevelIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &levelRenderer{indicator: l}
}

type levelRenderer struct {
	indicator *LevelIndicator
	objects   []fyne.CanvasObject
}

func (r *levelRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *levelRenderer) Layout(size fyne.Size) {}

func (r *levelRenderer) Refresh() {
	r.indicator.mu.RLock()
	level := r.indicator.level
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.indicator.Size()

	// Background
	bg := canvas.NewRectangle(color.RGBA{30, 30, 40, 255})
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Draw 30 level segments
	margin := float32(5)
	barW := size.Width - 2*margin
	totalH := size.Height - 2*margin
	segH := totalH / 30
	gap := float32(2)

	for i := 0; i < 30; i++ {
		y := margin + float32(29-i)*segH

		var segColor color.RGBA
		if i == level {
			// Current level - bright white/yellow
			segColor = colorWarning
		} else if i < level {
			// Below current - gradient blue to cyan
			t := float64(i) / 29.0
			segColor = color.RGBA{
				uint8(50 + t*150),
				uint8(50 + t*200),
				255,
				200,
			}
		} else {
			// Above current - gradient pink to red
			t := float64(i) / 29.0
			segColor = color.RGBA{
				255,
				uint8(200 - t*150),
				uint8(200 - t*150),
				200,
			}
		}

		seg := canvas.NewRectangle(segColor)
		seg.Resize(fyne.NewSize(barW, segH-gap))
		seg.Move(fyne.NewPos(margin, y))
		r.objects = append(r.objects, seg)

		// Level number for every 5th level
		if i%5 == 0 || i == 29 {
			label := canvas.NewText(fmt.Sprintf("%d", i+1), colorAxis)
			label.TextSize = 10
			label.Move(fyne.NewPos(margin+barW+2, y))
			r.objects = append(r.objects, label)
		}
	}
}

func (r *levelRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *levelRenderer) Destroy() {}

// ============================================================
// Custom Theme
// ============================================================

type ironLatticeTheme struct{}

func (t *ironLatticeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{20, 20, 35, 255}
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{40, 40, 60, 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *ironLatticeTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *ironLatticeTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *ironLatticeTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
