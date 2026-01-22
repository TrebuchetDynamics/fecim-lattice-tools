// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains hero visualizations for the technical briefing.
package gui

import (
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// AnimatedEnergyRace shows animated energy comparison bars.
type AnimatedEnergyRace struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	cpuEnergy    float64 // Target: 1000 fJ
	gpuEnergy    float64 // Target: 100 fJ
	fecimEnergy  float64 // Target: 10 fJ
	showWinner   bool    // Pulse FeCIM when true
	logScale     bool    // Use logarithmic scale
	pulsePhase   float64 // For winner pulse animation
	raster       *canvas.Raster
	minSize      fyne.Size
}

// NewAnimatedEnergyRace creates a new energy race visualization.
func NewAnimatedEnergyRace() *AnimatedEnergyRace {
	e := &AnimatedEnergyRace{
		cpuEnergy:   1000,
		gpuEnergy:   100,
		fecimEnergy: 10,
		logScale:    false,
		minSize:     fyne.NewSize(600, 200),
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetLogScale enables/disables logarithmic scale.
func (e *AnimatedEnergyRace) SetLogScale(log bool) {
	e.mu.Lock()
	e.logScale = log
	e.mu.Unlock()
	e.Refresh()
}

// UpdateAnimation advances the animation by dt seconds.
func (e *AnimatedEnergyRace) UpdateAnimation(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Animate progress (2 seconds to complete)
	if e.animProgress < 1.0 {
		e.animProgress += dt * 0.5 // 2 seconds to fill
		if e.animProgress > 1.0 {
			e.animProgress = 1.0
			e.showWinner = true
		}
	}

	// Animate winner pulse
	if e.showWinner {
		e.pulsePhase += dt * 3.0 // Pulse frequency
	}
}

// Reset resets the animation to start.
func (e *AnimatedEnergyRace) Reset() {
	e.mu.Lock()
	e.animProgress = 0
	e.showWinner = false
	e.pulsePhase = 0
	e.mu.Unlock()
	e.Refresh()
}

// MinSize returns minimum size.
func (e *AnimatedEnergyRace) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *AnimatedEnergyRace) CreateRenderer() fyne.WidgetRenderer {
	e.raster = canvas.NewRaster(e.generateImage)
	return widget.NewSimpleRenderer(e.raster)
}

// generateImage creates the energy race visualization.
func (e *AnimatedEnergyRace) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 200 || h < 100 {
		return img
	}

	e.mu.RLock()
	progress := e.animProgress
	showWinner := e.showWinner
	pulsePhase := e.pulsePhase
	logScale := e.logScale
	e.mu.RUnlock()

	// Layout
	padding := 20
	labelWidth := 80
	barAreaWidth := w - 2*padding - labelWidth - 100 // Leave room for value labels
	barHeight := (h - 2*padding - 60) / 3            // 3 bars with title space
	barSpacing := 10
	startX := padding + labelWidth
	startY := padding + 40 // Room for title

	// Title
	drawTextSimple(img, "ENERGY PER MAC OPERATION", w/2-120, 15, color.RGBA{0, 212, 255, 255}, 16)

	// Draw bars
	bars := []struct {
		name   string
		energy float64
		color  color.RGBA
	}{
		{"CPU+DRAM", 1000, color.RGBA{200, 100, 100, 255}},
		{"GPU+HBM", 100, color.RGBA{200, 180, 100, 255}},
		{"FeCIM", 10, color.RGBA{100, 200, 150, 255}},
	}

	maxEnergy := 1000.0

	for i, bar := range bars {
		y := startY + i*(barHeight+barSpacing)

		// Label
		drawTextSimple(img, bar.name, padding, y+barHeight/2-6, color.RGBA{200, 200, 200, 255}, 12)

		// Calculate bar width
		var barWidth int
		if logScale {
			// Logarithmic scale
			logMax := math.Log10(maxEnergy)
			logVal := math.Log10(bar.energy)
			barWidth = int(float64(barAreaWidth) * (logVal / logMax) * progress)
		} else {
			// Linear scale
			barWidth = int(float64(barAreaWidth) * (bar.energy / maxEnergy) * progress)
		}

		// Draw bar background (track)
		trackColor := color.RGBA{40, 50, 70, 255}
		for dy := 0; dy < barHeight; dy++ {
			for dx := 0; dx < barAreaWidth; dx++ {
				img.Set(startX+dx, y+dy, trackColor)
			}
		}

		// Draw bar
		barColor := bar.color
		if i == 2 && showWinner {
			// Pulse effect for FeCIM
			pulse := math.Sin(pulsePhase) * 0.3
			barColor = color.RGBA{
				uint8(min(255, int(float64(bar.color.R)*(1+pulse)))),
				uint8(min(255, int(float64(bar.color.G)*(1+pulse)))),
				uint8(min(255, int(float64(bar.color.B)*(1+pulse)))),
				255,
			}
		}
		for dy := 0; dy < barHeight; dy++ {
			for dx := 0; dx < barWidth; dx++ {
				img.Set(startX+dx, y+dy, barColor)
			}
		}

		// Value label
		valX := startX + barAreaWidth + 10
		valText := ""
		currentEnergy := bar.energy * progress
		if currentEnergy >= 100 {
			valText = formatFloat(currentEnergy, 0) + " fJ"
		} else {
			valText = formatFloat(currentEnergy, 1) + " fJ"
		}
		drawTextSimple(img, valText, valX, y+barHeight/2-6, bar.color, 12)
	}

	// Show "100× LESS ENERGY" headline when animation complete
	if showWinner && progress >= 1.0 {
		// Draw headline with pulse
		pulse := 0.7 + math.Sin(pulsePhase)*0.3
		headlineColor := color.RGBA{
			uint8(float64(0) * pulse),
			uint8(float64(212) * pulse),
			uint8(float64(255) * pulse),
			255,
		}

		// Background box for headline
		boxX := w/2 - 100
		boxY := h - 35
		boxW := 200
		boxH := 25
		boxColor := color.RGBA{0, 50, 100, 200}
		for dy := 0; dy < boxH; dy++ {
			for dx := 0; dx < boxW; dx++ {
				img.Set(boxX+dx, boxY+dy, boxColor)
			}
		}

		drawTextSimple(img, "100× LESS ENERGY", boxX+20, boxY+5, headlineColor, 14)
	}

	// Source note
	drawTextSimple(img, "* FeCIM: Dr. Tour claims (TRL 4, not verified)", padding, h-15, color.RGBA{150, 150, 150, 255}, 10)

	return img
}

// MemoryWallAnimation shows data movement visualization.
type MemoryWallAnimation struct {
	widget.BaseWidget

	mu            sync.RWMutex
	packets       []Packet
	dataMovements int     // Counter
	simTime       float64 // Animation time
	raster        *canvas.Raster
	minSize       fyne.Size
}

// Packet represents a data packet moving between CPU and memory.
type Packet struct {
	x, y   float64
	vx     float64
	active bool
}

// NewMemoryWallAnimation creates a new memory wall visualization.
func NewMemoryWallAnimation() *MemoryWallAnimation {
	m := &MemoryWallAnimation{
		packets: make([]Packet, 0, 10),
		minSize: fyne.NewSize(500, 150),
	}
	// Initialize packets
	for i := 0; i < 5; i++ {
		m.packets = append(m.packets, Packet{
			x:      float64(50 + i*30),
			y:      float64(50 + (i%3)*15),
			vx:     100 + float64(i*20),
			active: true,
		})
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateAnimation advances the animation.
func (m *MemoryWallAnimation) UpdateAnimation(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.simTime += dt

	// Move packets back and forth
	for i := range m.packets {
		if !m.packets[i].active {
			continue
		}

		m.packets[i].x += m.packets[i].vx * dt

		// Bounce at boundaries
		if m.packets[i].x > 180 {
			m.packets[i].vx = -m.packets[i].vx
			m.dataMovements++
		} else if m.packets[i].x < 50 {
			m.packets[i].vx = -m.packets[i].vx
			m.dataMovements++
		}
	}
}

// Reset resets the animation.
func (m *MemoryWallAnimation) Reset() {
	m.mu.Lock()
	m.simTime = 0
	m.dataMovements = 0
	for i := range m.packets {
		m.packets[i].x = float64(50 + i*30)
		m.packets[i].vx = 100 + float64(i*20)
	}
	m.mu.Unlock()
	m.Refresh()
}

// MinSize returns minimum size.
func (m *MemoryWallAnimation) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *MemoryWallAnimation) CreateRenderer() fyne.WidgetRenderer {
	m.raster = canvas.NewRaster(m.generateImage)
	return widget.NewSimpleRenderer(m.raster)
}

// generateImage creates the memory wall visualization.
func (m *MemoryWallAnimation) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 300 || h < 100 {
		return img
	}

	m.mu.RLock()
	packets := make([]Packet, len(m.packets))
	copy(packets, m.packets)
	dataMovements := m.dataMovements
	m.mu.RUnlock()

	midX := w / 2

	// LEFT SIDE: Von Neumann Architecture
	// Title
	drawTextSimple(img, "VON NEUMANN", 30, 15, color.RGBA{200, 100, 100, 255}, 12)

	// CPU Box
	cpuX, cpuY := 30, 40
	cpuW, cpuH := 60, 40
	drawBoxFilled(img, cpuX, cpuY, cpuW, cpuH, color.RGBA{200, 100, 100, 255}, color.RGBA{100, 50, 50, 255})
	drawTextSimple(img, "CPU", cpuX+15, cpuY+15, color.RGBA{255, 255, 255, 255}, 12)

	// Memory Box
	memX, memY := 160, 40
	memW, memH := 60, 40
	drawBoxFilled(img, memX, memY, memW, memH, color.RGBA{100, 100, 200, 255}, color.RGBA{50, 50, 100, 255})
	drawTextSimple(img, "DRAM", memX+10, memY+15, color.RGBA{255, 255, 255, 255}, 12)

	// Draw data bus
	busY := cpuY + cpuH/2
	for x := cpuX + cpuW; x < memX; x++ {
		img.Set(x, busY, color.RGBA{100, 100, 100, 255})
		img.Set(x, busY+1, color.RGBA{100, 100, 100, 255})
	}

	// Draw packets (data moving)
	for _, p := range packets {
		if !p.active {
			continue
		}
		px := int(p.x)
		py := int(p.y)
		// Red packet with glow (energy waste)
		packetColor := color.RGBA{255, 100, 100, 255}
		glowColor := color.RGBA{255, 50, 50, 100}
		// Glow
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				if px+dx > 0 && px+dx < midX-30 && py+dy > 0 && py+dy < h {
					img.Set(px+dx, py+dy, glowColor)
				}
			}
		}
		// Packet
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if px+dx > 0 && px+dx < midX-30 && py+dy > 0 && py+dy < h {
					img.Set(px+dx, py+dy, packetColor)
				}
			}
		}
	}

	// Data movement counter
	counterText := formatNumber(float64(dataMovements))
	drawTextSimple(img, "Data Moves: "+counterText, 30, h-30, color.RGBA{255, 200, 100, 255}, 11)
	drawTextSimple(img, "(ENERGY WASTE)", 30, h-15, color.RGBA{255, 100, 100, 200}, 9)

	// DIVIDER
	divColor := color.RGBA{0, 100, 150, 255}
	for y := 10; y < h-10; y++ {
		img.Set(midX-10, y, divColor)
	}
	drawTextSimple(img, "VS", midX-15, h/2-6, color.RGBA{0, 212, 255, 255}, 12)

	// RIGHT SIDE: FeCIM Architecture
	// Title
	drawTextSimple(img, "COMPUTE-IN-MEMORY", midX+30, 15, color.RGBA{100, 200, 150, 255}, 12)

	// Combined CIM Box
	cimX, cimY := midX + 50, 40
	cimW, cimH := 120, 50
	drawBoxFilled(img, cimX, cimY, cimW, cimH, color.RGBA{100, 200, 150, 255}, color.RGBA{50, 100, 75, 255})
	drawTextSimple(img, "COMPUTE", cimX+30, cimY+10, color.RGBA{255, 255, 255, 255}, 11)
	drawTextSimple(img, "HERE", cimX+42, cimY+25, color.RGBA{255, 255, 255, 255}, 11)

	// No movement indicator
	drawTextSimple(img, "Zero Data Movement", midX+50, h-30, color.RGBA{100, 255, 150, 255}, 11)
	drawTextSimple(img, "= Zero Waste", midX+70, h-15, color.RGBA{100, 200, 150, 200}, 10)

	return img
}

// Helper functions

func drawTextSimple(img *image.RGBA, text string, x, y int, c color.RGBA, fontSize int) {
	// Simple text drawing - each character is a box
	// This is a placeholder for proper font rendering
	charWidth := fontSize / 2
	for i, ch := range text {
		if ch == ' ' {
			continue
		}
		cx := x + i*charWidth
		// Draw a simple representation
		for dy := 0; dy < fontSize; dy++ {
			for dx := 0; dx < charWidth-1; dx++ {
				if cy := y + dy; cy >= 0 && cy < img.Bounds().Dy() {
					if ccx := cx + dx; ccx >= 0 && ccx < img.Bounds().Dx() {
						// Simple pattern based on character
						if dy > 1 && dy < fontSize-2 && dx > 0 && dx < charWidth-2 {
							img.Set(ccx, cy, c)
						}
					}
				}
			}
		}
	}
}

func drawBoxFilled(img *image.RGBA, x, y, width, height int, borderColor, fillColor color.RGBA) {
	// Fill
	for dy := 2; dy < height-2; dy++ {
		for dx := 2; dx < width-2; dx++ {
			img.Set(x+dx, y+dy, fillColor)
		}
	}
	// Border
	for dx := 0; dx < width; dx++ {
		img.Set(x+dx, y, borderColor)
		img.Set(x+dx, y+1, borderColor)
		img.Set(x+dx, y+height-1, borderColor)
		img.Set(x+dx, y+height-2, borderColor)
	}
	for dy := 0; dy < height; dy++ {
		img.Set(x, y+dy, borderColor)
		img.Set(x+1, y+dy, borderColor)
		img.Set(x+width-1, y+dy, borderColor)
		img.Set(x+width-2, y+dy, borderColor)
	}
}

func formatFloat(f float64, decimals int) string {
	if decimals == 0 {
		return formatNumber(f)
	}
	format := "%." + string(rune('0'+decimals)) + "f"
	result := ""
	switch decimals {
	case 1:
		result = formatNumberWithDecimal(f, 1)
	case 2:
		result = formatNumberWithDecimal(f, 2)
	default:
		result = formatNumber(f)
	}
	_ = format // unused, keeping for clarity
	return result
}

func formatNumberWithDecimal(n float64, decimals int) string {
	intPart := int(n)
	fracPart := n - float64(intPart)

	intStr := formatNumber(float64(intPart))

	if decimals == 0 {
		return intStr
	}

	// Format decimal part
	fracVal := int(fracPart * math.Pow(10, float64(decimals)))
	fracStr := ""
	for i := 0; i < decimals; i++ {
		fracStr = string(rune('0'+fracVal%10)) + fracStr
		fracVal /= 10
	}

	return intStr + "." + fracStr
}

func formatNumber(n float64) string {
	// Simple number formatting with commas
	intVal := int(n)
	if intVal == 0 {
		return "0"
	}

	negative := intVal < 0
	if negative {
		intVal = -intVal
	}

	// Build string from right to left
	result := ""
	digits := 0
	for intVal > 0 {
		if digits > 0 && digits%3 == 0 {
			result = "," + result
		}
		result = string(rune('0'+intVal%10)) + result
		intVal /= 10
		digits++
	}

	if negative {
		result = "-" + result
	}

	return result
}
