// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains hero visualizations for the comparison demo.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// estimatedColor is used for unverified/estimated values (amber)
var estimatedColor = color.RGBA{255, 191, 0, 255}

// Energy values in picojoules per MAC (from docs/videos/ironlattice-youtube-script.md)
// "CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. FeCIM: under 1 picojoule."
const (
	cpuEnergyPJ   = 1000.0 // 1000 pJ/MAC
	gpuEnergyPJ   = 100.0  // 100 pJ/MAC
	fecimEnergyPJ = 1.0    // ~1 pJ/MAC (conservative estimate for claimed "<1 pJ")
)

// AnimatedEnergyRace shows animated energy comparison bars.
type AnimatedEnergyRace struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	showWinner   bool
	pulsePhase   float64

	// Cached values to avoid redundant SetText calls (prevents resize loops)
	lastCpuText   string
	lastGpuText   string
	lastFecimText string

	// UI elements
	container    *fyne.Container
	cpuBar       *canvas.Rectangle
	gpuBar       *canvas.Rectangle
	fecimBar     *canvas.Rectangle
	cpuValue     *widget.Label
	gpuValue     *widget.Label
	fecimValue   *widget.Label
	headlineText *canvas.Text
}

// NewAnimatedEnergyRace creates a new energy race visualization.
func NewAnimatedEnergyRace() *AnimatedEnergyRace {
	e := &AnimatedEnergyRace{}
	e.ExtendBaseWidget(e)
	return e
}

// SetLogScale enables/disables logarithmic scale (placeholder).
func (e *AnimatedEnergyRace) SetLogScale(log bool) {}

// UpdateAnimation advances the animation by dt seconds.
func (e *AnimatedEnergyRace) UpdateAnimation(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.animProgress < 1.0 {
		e.animProgress += dt * 0.5
		if e.animProgress > 1.0 {
			e.animProgress = 1.0
			e.showWinner = true
		}
	}

	if e.showWinner {
		e.pulsePhase += dt * 3.0
	}
}

// Reset resets the animation.
func (e *AnimatedEnergyRace) Reset() {
	e.mu.Lock()
	e.animProgress = 0
	e.showWinner = false
	e.pulsePhase = 0
	e.mu.Unlock()
	fyne.Do(func() {
		e.Refresh()
	})
}

// MinSize returns minimum size.
func (e *AnimatedEnergyRace) MinSize() fyne.Size {
	return fyne.NewSize(400, 160)
}

// CreateRenderer implements fyne.Widget.
func (e *AnimatedEnergyRace) CreateRenderer() fyne.WidgetRenderer {
	barHeight := float32(26)   // Increased from 18 to 26 for better visibility
	trackWidth := float32(400) // Reference width for CPU (1000 pJ)

	// LINEAR SCALE: CPU=100%, GPU=10%, FeCIM=0.1%
	labelWidth := float32(80)
	valueWidth := float32(80) // Increased for larger text

	// CPU row - full width (1000 pJ reference)
	cpuIcon := canvas.NewText("🖥️", color.Black)
	cpuIcon.TextSize = 18
	cpuLabel := widget.NewLabel("CPU+DRAM")
	cpuLabel.TextStyle = fyne.TextStyle{Bold: true}
	cpuLabel.Importance = widget.HighImportance // Makes text larger/bolder
	cpuLabelBox := container.NewHBox(cpuIcon, cpuLabel)
	cpuLabelContainer := container.NewGridWrap(fyne.NewSize(labelWidth+20, barHeight), cpuLabelBox)
	e.cpuBar = canvas.NewRectangle(color.RGBA{231, 76, 60, 255}) // Strong red (#E74C3C)
	e.cpuBar.SetMinSize(fyne.NewSize(trackWidth, barHeight))
	e.cpuValue = widget.NewLabel("1000 pJ")
	e.cpuValue.TextStyle = fyne.TextStyle{Bold: true}
	e.cpuValue.Importance = widget.HighImportance // Makes text larger/bolder
	cpuValueBox := container.NewGridWrap(fyne.NewSize(valueWidth, barHeight), e.cpuValue)
	cpuRow := container.NewHBox(cpuLabelContainer, e.cpuBar, cpuValueBox)

	// GPU row - 10% width (100 pJ = 10× less)
	gpuIcon := canvas.NewText("🎮", color.Black)
	gpuIcon.TextSize = 18
	gpuLabel := widget.NewLabel("GPU+HBM")
	gpuLabel.TextStyle = fyne.TextStyle{Bold: true}
	gpuLabel.Importance = widget.HighImportance
	gpuLabelBox := container.NewHBox(gpuIcon, gpuLabel)
	gpuLabelContainer := container.NewGridWrap(fyne.NewSize(labelWidth+20, barHeight), gpuLabelBox)
	e.gpuBar = canvas.NewRectangle(color.RGBA{243, 156, 18, 255}) // Amber/Orange (#F39C12)
	e.gpuBar.SetMinSize(fyne.NewSize(trackWidth*0.1, barHeight)) // 10% of CPU
	e.gpuValue = widget.NewLabel("100 pJ")
	e.gpuValue.TextStyle = fyne.TextStyle{Bold: true}
	e.gpuValue.Importance = widget.HighImportance
	gpuValueBox := container.NewGridWrap(fyne.NewSize(valueWidth, barHeight), e.gpuValue)
	gpuRow := container.NewHBox(gpuLabelContainer, e.gpuBar, gpuValueBox)

	// FeCIM row - 0.1% width (1 pJ = 1000× less) - minimum 8px visible
	fecimIcon := canvas.NewText("💾", color.Black)
	fecimIcon.TextSize = 18
	fecimLabel := widget.NewLabel("FeCIM")
	fecimLabel.TextStyle = fyne.TextStyle{Bold: true}
	fecimLabel.Importance = widget.HighImportance
	fecimAsterisk := canvas.NewText("*", estimatedColor)
	fecimAsterisk.TextSize = 14
	fecimAsterisk.TextStyle = fyne.TextStyle{Bold: true}
	fecimLabelBox := container.NewHBox(fecimIcon, fecimLabel, fecimAsterisk)
	fecimLabelContainer := container.NewGridWrap(fyne.NewSize(labelWidth+20, barHeight), fecimLabelBox)
	e.fecimBar = canvas.NewRectangle(color.RGBA{46, 204, 113, 255}) // Bright green (#2ECC71)
	e.fecimBar.SetMinSize(fyne.NewSize(max(8, trackWidth*0.001), barHeight)) // 0.1% of CPU, min 8px
	e.fecimValue = widget.NewLabel("~1 pJ")
	e.fecimValue.TextStyle = fyne.TextStyle{Bold: true}
	e.fecimValue.Importance = widget.HighImportance
	fecimValueBox := container.NewGridWrap(fyne.NewSize(valueWidth, barHeight), e.fecimValue)
	fecimRow := container.NewHBox(fecimLabelContainer, e.fecimBar, fecimValueBox)

	// Headline - larger and visible from start (dimmed), then pulses bright
	e.headlineText = canvas.NewText("1000× LESS ENERGY*", color.RGBA{0, 85, 102, 128}) // Dimmed cyan, visible from start
	e.headlineText.TextSize = 38 // Increased from 32 to 38
	e.headlineText.TextStyle = fyne.TextStyle{Bold: true}
	e.headlineText.Alignment = fyne.TextAlignCenter

	// Scale note - more prominent
	scaleNote := canvas.NewText("Scale: Linear (1000:100:1 ratio)", color.RGBA{100, 100, 100, 255})
	scaleNote.TextSize = 11
	scaleNote.TextStyle = fyne.TextStyle{Italic: true}

	// Legend for estimated indicator
	estimatedNote := canvas.NewText("* = Estimated (TRL 4)", estimatedColor)
	estimatedNote.TextSize = 10
	estimatedNote.TextStyle = fyne.TextStyle{Italic: true}

	e.container = container.NewVBox(
		cpuRow,
		gpuRow,
		fecimRow,
		container.NewCenter(e.headlineText),
		container.NewHBox(container.NewCenter(scaleNote), layout.NewSpacer(), estimatedNote),
	)

	return widget.NewSimpleRenderer(e.container)
}

// Refresh updates the widget display.
func (e *AnimatedEnergyRace) Refresh() {
	e.mu.RLock()
	progress := e.animProgress
	showWinner := e.showWinner
	pulsePhase := e.pulsePhase
	e.mu.RUnlock()

	if e.cpuBar == nil {
		return
	}

	// LINEAR scale bar widths: CPU=100%, GPU=10%, FeCIM=0.1%
	barHeight := float32(26) // Updated from 18 to 26
	trackWidth := float32(400)
	e.cpuBar.SetMinSize(fyne.NewSize(trackWidth*float32(progress), barHeight))
	e.gpuBar.SetMinSize(fyne.NewSize(trackWidth*0.1*float32(progress), barHeight))
	e.fecimBar.SetMinSize(fyne.NewSize(max(8, trackWidth*0.001*float32(progress)), barHeight)) // Updated min from 4 to 8

	// Update value labels - show final values after animation
	// Use caching to avoid redundant SetText calls that trigger layout recalculations
	var cpuText, gpuText, fecimText string
	if progress > 0.9 {
		cpuText = "1000 pJ"
		gpuText = "100 pJ"
		fecimText = "~1 pJ"
	} else {
		cpuText = fmt.Sprintf("%.0f pJ", cpuEnergyPJ*progress)
		gpuText = fmt.Sprintf("%.0f pJ", gpuEnergyPJ*progress)
		fecimText = fmt.Sprintf("%.1f pJ", fecimEnergyPJ*progress)
	}
	if cpuText != e.lastCpuText {
		e.cpuValue.SetText(cpuText)
		e.lastCpuText = cpuText
	}
	if gpuText != e.lastGpuText {
		e.gpuValue.SetText(gpuText)
		e.lastGpuText = gpuText
	}
	if fecimText != e.lastFecimText {
		e.fecimValue.SetText(fecimText)
		e.lastFecimText = fecimText
	}

	// Headline visibility and pulse - now dimmed at start, bright pulse when complete
	if showWinner {
		// Bright pulsing cyan when animation complete
		pulse := 0.8 + math.Sin(pulsePhase)*0.2
		e.headlineText.Color = color.RGBA{0, uint8(212 * pulse), uint8(255 * pulse), 255}
	} else {
		// Dimmed cyan, visible from start
		e.headlineText.Color = color.RGBA{0, 85, 102, 128}
	}

	canvas.Refresh(e.cpuBar)
	canvas.Refresh(e.gpuBar)
	canvas.Refresh(e.fecimBar)
	canvas.Refresh(e.headlineText)
	e.container.Refresh()
}

// MemoryWallAnimation shows data movement visualization.
type MemoryWallAnimation struct {
	widget.BaseWidget

	mu            sync.RWMutex
	dataMovements int
	simTime       float64
	pulsePhase    float64

	container   *fyne.Container
	counterText *widget.Label
	arrowText   *canvas.Text
}

// NewMemoryWallAnimation creates a new memory wall visualization.
func NewMemoryWallAnimation() *MemoryWallAnimation {
	m := &MemoryWallAnimation{}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateAnimation advances the animation.
func (m *MemoryWallAnimation) UpdateAnimation(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.simTime += dt
	m.pulsePhase += dt * 3.0

	// Simulate data movements
	if int(m.simTime*10)%3 == 0 {
		m.dataMovements++
	}
}

// Reset resets the animation.
func (m *MemoryWallAnimation) Reset() {
	m.mu.Lock()
	m.simTime = 0
	m.dataMovements = 0
	m.pulsePhase = 0
	m.mu.Unlock()
	fyne.Do(func() {
		m.Refresh()
	})
}

// MinSize returns minimum size.
func (m *MemoryWallAnimation) MinSize() fyne.Size {
	return fyne.NewSize(400, 80)
}

// CreateRenderer implements fyne.Widget.
func (m *MemoryWallAnimation) CreateRenderer() fyne.WidgetRenderer {
	// Von Neumann side
	cpuBox := canvas.NewRectangle(color.RGBA{180, 80, 80, 255})
	cpuBox.SetMinSize(fyne.NewSize(40, 30))
	cpuLabel := widget.NewLabel("CPU")

	m.arrowText = canvas.NewText("<->", color.RGBA{255, 200, 100, 255})
	m.arrowText.TextSize = 14

	memBox := canvas.NewRectangle(color.RGBA{80, 80, 180, 255})
	memBox.SetMinSize(fyne.NewSize(40, 30))
	memLabel := widget.NewLabel("MEM")

	vonNeumann := container.NewHBox(
		container.NewStack(cpuBox, container.NewCenter(cpuLabel)),
		m.arrowText,
		container.NewStack(memBox, container.NewCenter(memLabel)),
	)
	m.counterText = widget.NewLabel("Moves: 0")

	// VS divider
	vsText := canvas.NewText("VS", color.RGBA{0, 212, 255, 255})
	vsText.TextSize = 14

	// CIM side
	cimBox := canvas.NewRectangle(color.RGBA{80, 180, 120, 255})
	cimBox.SetMinSize(fyne.NewSize(70, 30))
	cimLabel := widget.NewLabel("CIM")
	cimStack := container.NewStack(cimBox, container.NewCenter(cimLabel))
	zeroLabel := widget.NewLabel("Zero Movement")

	m.container = container.NewHBox(
		container.NewVBox(container.NewCenter(vonNeumann), m.counterText),
		layout.NewSpacer(),
		vsText,
		layout.NewSpacer(),
		container.NewVBox(container.NewCenter(cimStack), zeroLabel),
	)

	return widget.NewSimpleRenderer(m.container)
}

// Refresh updates the widget display.
func (m *MemoryWallAnimation) Refresh() {
	m.mu.RLock()
	dataMovements := m.dataMovements
	pulsePhase := m.pulsePhase
	m.mu.RUnlock()

	if m.counterText != nil {
		m.counterText.SetText(fmt.Sprintf("Moves: %d", dataMovements))
	}

	if m.arrowText != nil {
		pulse := 0.5 + math.Sin(pulsePhase)*0.5
		m.arrowText.Color = color.RGBA{255, uint8(150 + 100*pulse), uint8(50 + 50*pulse), 255}
		canvas.Refresh(m.arrowText)
	}
}

// Packet represents a data packet (kept for compatibility).
type Packet struct {
	x, y   float64
	vx     float64
	active bool
}
