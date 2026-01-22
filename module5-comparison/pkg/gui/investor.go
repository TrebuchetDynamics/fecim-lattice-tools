// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains investor-focused visualizations.
package gui

import (
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// PhasedStrategyDiagram shows the commercialization strategy phases.
type PhasedStrategyDiagram struct {
	widget.BaseWidget

	mu           sync.RWMutex
	currentPhase int     // 0-2 for highlighted phase
	animProgress float64 // Arrow animation progress
	raster       *canvas.Raster
	minSize      fyne.Size
}

// NewPhasedStrategyDiagram creates a new strategy diagram.
func NewPhasedStrategyDiagram() *PhasedStrategyDiagram {
	p := &PhasedStrategyDiagram{
		minSize: fyne.NewSize(500, 120),
	}
	p.ExtendBaseWidget(p)
	return p
}

// SetPhase sets the highlighted phase.
func (p *PhasedStrategyDiagram) SetPhase(phase int) {
	p.mu.Lock()
	p.currentPhase = phase % 3
	p.mu.Unlock()
	p.Refresh()
}

// UpdateAnimation advances the animation.
func (p *PhasedStrategyDiagram) UpdateAnimation(dt float64) {
	p.mu.Lock()
	p.animProgress += dt * 0.5
	if p.animProgress > 3.0 {
		p.animProgress = 0
	}
	p.mu.Unlock()
}

// MinSize returns minimum size.
func (p *PhasedStrategyDiagram) MinSize() fyne.Size {
	return p.minSize
}

// CreateRenderer implements fyne.Widget.
func (p *PhasedStrategyDiagram) CreateRenderer() fyne.WidgetRenderer {
	p.raster = canvas.NewRaster(p.generateImage)
	return widget.NewSimpleRenderer(p.raster)
}

// generateImage creates the strategy diagram.
func (p *PhasedStrategyDiagram) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 300 || h < 80 {
		return img
	}

	p.mu.RLock()
	currentPhase := p.currentPhase
	animProgress := p.animProgress
	p.mu.RUnlock()

	// Title
	drawTextSimple(img, "COMMERCIALIZATION STRATEGY", w/2-100, 8, color.RGBA{0, 212, 255, 255}, 12)

	// Phase data
	phases := []struct {
		title    string
		subtitle string
		benefit  string
	}{
		{"PHASE 1", "NAND Flash", "Drop-in compatible"},
		{"PHASE 2", "DRAM", "No refresh needed"},
		{"PHASE 3", "Full CIM", "80-90% savings"},
	}

	// Layout
	boxWidth := (w - 100) / 3
	boxHeight := 50
	startY := 30
	spacing := 40

	for i, phase := range phases {
		boxX := 30 + i*(boxWidth+spacing)

		// Determine if this phase is highlighted
		isHighlighted := i == currentPhase
		isAnimated := int(animProgress) == i

		// Box colors
		var borderColor, fillColor color.RGBA
		if isHighlighted || isAnimated {
			borderColor = color.RGBA{0, 212, 255, 255}
			fillColor = color.RGBA{0, 80, 130, 255}
		} else {
			borderColor = color.RGBA{80, 100, 130, 255}
			fillColor = color.RGBA{40, 50, 70, 255}
		}

		// Draw box
		drawBoxFilled(img, boxX, startY, boxWidth, boxHeight, borderColor, fillColor)

		// Phase title
		titleColor := borderColor
		drawTextSimple(img, phase.title, boxX+boxWidth/2-25, startY+8, titleColor, 10)

		// Subtitle
		drawTextSimple(img, phase.subtitle, boxX+boxWidth/2-30, startY+22, color.RGBA{200, 200, 200, 255}, 9)

		// Benefit (below box)
		drawTextSimple(img, phase.benefit, boxX+5, startY+boxHeight+5, color.RGBA{100, 200, 150, 255}, 8)

		// Arrow to next phase
		if i < 2 {
			arrowX := boxX + boxWidth + 5
			arrowY := startY + boxHeight/2

			// Arrow color based on animation
			arrowAlpha := uint8(150)
			if int(animProgress) == i {
				// Pulsing arrow
				pulse := math.Sin(animProgress*math.Pi*2) * 0.5 + 0.5
				arrowAlpha = uint8(150 + pulse*105)
			}
			arrowColor := color.RGBA{0, 212, 255, arrowAlpha}

			// Draw arrow line
			for ax := 0; ax < spacing-10; ax++ {
				img.Set(arrowX+ax, arrowY, arrowColor)
				img.Set(arrowX+ax, arrowY+1, arrowColor)
			}

			// Arrow head
			for ay := -4; ay <= 4; ay++ {
				headX := arrowX + spacing - 15 + abs(ay)
				img.Set(headX, arrowY+ay, arrowColor)
			}
		}
	}

	return img
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// AnalogStatesComparison shows binary vs FeCIM memory comparison.
type AnalogStatesComparison struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // Counter animation
	raster       *canvas.Raster
	minSize      fyne.Size
}

// NewAnalogStatesComparison creates a new analog states comparison.
func NewAnalogStatesComparison() *AnalogStatesComparison {
	a := &AnalogStatesComparison{
		minSize: fyne.NewSize(400, 150),
	}
	a.ExtendBaseWidget(a)
	return a
}

// UpdateAnimation advances the animation.
func (a *AnalogStatesComparison) UpdateAnimation(dt float64) {
	a.mu.Lock()
	a.animProgress += dt
	a.mu.Unlock()
}

// MinSize returns minimum size.
func (a *AnalogStatesComparison) MinSize() fyne.Size {
	return a.minSize
}

// CreateRenderer implements fyne.Widget.
func (a *AnalogStatesComparison) CreateRenderer() fyne.WidgetRenderer {
	a.raster = canvas.NewRaster(a.generateImage)
	return widget.NewSimpleRenderer(a.raster)
}

// generateImage creates the analog states comparison.
func (a *AnalogStatesComparison) generateImage(w, h int) image.Image {
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

	a.mu.RLock()
	animProgress := a.animProgress
	a.mu.RUnlock()

	midX := w / 2

	// LEFT SIDE: Binary Memory
	drawTextSimple(img, "BINARY MEMORY", 20, 15, color.RGBA{200, 100, 100, 255}, 11)

	// Draw 2-cell grid (0 and 1)
	cellSize := 35
	cellY := 40
	for i := 0; i < 2; i++ {
		cellX := 30 + i*(cellSize+5)
		var cellColor color.RGBA
		if i == 0 {
			cellColor = color.RGBA{40, 40, 40, 255} // Black for 0
		} else {
			cellColor = color.RGBA{255, 255, 255, 255} // White for 1
		}
		// Draw cell
		for dy := 0; dy < cellSize; dy++ {
			for dx := 0; dx < cellSize; dx++ {
				img.Set(cellX+dx, cellY+dy, cellColor)
			}
		}
		// Border
		borderColor := color.RGBA{100, 100, 100, 255}
		for dx := 0; dx < cellSize; dx++ {
			img.Set(cellX+dx, cellY, borderColor)
			img.Set(cellX+dx, cellY+cellSize-1, borderColor)
		}
		for dy := 0; dy < cellSize; dy++ {
			img.Set(cellX, cellY+dy, borderColor)
			img.Set(cellX+cellSize-1, cellY+dy, borderColor)
		}
		// Label
		label := "0"
		if i == 1 {
			label = "1"
		}
		labelColor := color.RGBA{100, 100, 100, 255}
		if i == 0 {
			labelColor = color.RGBA{200, 200, 200, 255}
		}
		drawTextSimple(img, label, cellX+cellSize/2-4, cellY+cellSize/2-6, labelColor, 12)
	}

	// Binary stats
	drawTextSimple(img, "2 states", 30, cellY+cellSize+10, color.RGBA{180, 180, 180, 255}, 10)
	drawTextSimple(img, "1 bit/cell", 30, cellY+cellSize+25, color.RGBA{150, 150, 150, 255}, 9)

	// RIGHT SIDE: FeCIM Memory
	drawTextSimple(img, "FeCIM MEMORY", midX+20, 15, color.RGBA{100, 200, 150, 255}, 11)

	// Draw 30-cell gradient (horizontal bar divided into cells)
	gradWidth := w - midX - 40
	gradHeight := 35
	gradY := 40
	cellWidth := gradWidth / 30

	for i := 0; i < 30; i++ {
		cellX := midX + 20 + i*cellWidth
		// Color gradient from blue to red
		t := float64(i) / 29.0
		var cellColor color.RGBA
		if t < 0.5 {
			t2 := t * 2
			cellColor = color.RGBA{
				uint8(80 + t2*175),
				uint8(120 + t2*135),
				255,
				255,
			}
		} else {
			t2 := (t - 0.5) * 2
			cellColor = color.RGBA{
				255,
				uint8(255 - t2*175),
				uint8(255 - t2*175),
				255,
			}
		}
		// Draw cell
		for dy := 0; dy < gradHeight; dy++ {
			for dx := 0; dx < cellWidth; dx++ {
				if cellX+dx < w-20 {
					img.Set(cellX+dx, gradY+dy, cellColor)
				}
			}
		}
	}

	// Border around gradient
	borderColor := color.RGBA{0, 212, 255, 255}
	for dx := 0; dx < gradWidth; dx++ {
		img.Set(midX+20+dx, gradY, borderColor)
		img.Set(midX+20+dx, gradY+gradHeight-1, borderColor)
	}
	for dy := 0; dy < gradHeight; dy++ {
		img.Set(midX+20, gradY+dy, borderColor)
		img.Set(midX+20+gradWidth-1, gradY+dy, borderColor)
	}

	// Labels
	drawTextSimple(img, "1", midX+25, gradY+gradHeight+5, color.RGBA{150, 150, 255, 255}, 8)
	drawTextSimple(img, "30", midX+gradWidth-5, gradY+gradHeight+5, color.RGBA{255, 150, 150, 255}, 8)

	// FeCIM stats with animated counter
	drawTextSimple(img, "30 states", midX+20, gradY+gradHeight+20, color.RGBA{180, 180, 180, 255}, 10)

	// Animated bits counter
	bitsTarget := 4.9
	bitsDisplay := bitsTarget
	if animProgress < 2.0 {
		bitsDisplay = bitsTarget * (animProgress / 2.0)
	}
	bitsText := formatNumberWithDecimal(bitsDisplay, 1) + " bits/cell"
	drawTextSimple(img, bitsText, midX+20, gradY+gradHeight+35, color.RGBA{100, 200, 150, 255}, 10)

	// Bottom tagline
	pulse := 0.7 + math.Sin(animProgress*2)*0.3
	taglineColor := color.RGBA{
		uint8(float64(0) * pulse),
		uint8(float64(212) * pulse),
		uint8(float64(255) * pulse),
		255,
	}
	drawTextSimple(img, "Same silicon, 5x more information", w/2-100, h-15, taglineColor, 11)

	return img
}

// WeebitNanoCard shows the Weebit Nano precedent.
type WeebitNanoCard struct {
	widget.BaseWidget
}

// NewWeebitNanoCard creates a new Weebit card.
func NewWeebitNanoCard() *WeebitNanoCard {
	w := &WeebitNanoCard{}
	w.ExtendBaseWidget(w)
	return w
}

// MinSize returns minimum size.
func (w *WeebitNanoCard) MinSize() fyne.Size {
	return fyne.NewSize(280, 180)
}

// CreateRenderer implements fyne.Widget.
func (w *WeebitNanoCard) CreateRenderer() fyne.WidgetRenderer {
	// Title with icon
	title := widget.NewLabelWithStyle("PRECEDENT: WEEBIT NANO", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Quote
	quote := widget.NewLabel("\"This company Weebit—this is another memory that came out of my lab... it's selling now on the market with three big customers.\"")
	quote.Wrapping = fyne.TextWrapWord
	quote.TextStyle = fyne.TextStyle{Italic: true}

	// Attribution
	attribution := widget.NewLabel("— Dr. external research group")
	attribution.Alignment = fyne.TextAlignTrailing

	// Checkmarks
	check1 := widget.NewLabel("✓ Started at TRL 4 (like FeCIM today)")
	check2 := widget.NewLabel("✓ Now partnered with major foundries")
	check3 := widget.NewLabel("✓ Proven commercialization path")

	checks := container.NewVBox(check1, check2, check3)

	// Stock note
	stockNote := widget.NewLabel("ASX: WBT")
	stockNote.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		quote,
		attribution,
		widget.NewSeparator(),
		checks,
		stockNote,
	)

	return widget.NewSimpleRenderer(content)
}
