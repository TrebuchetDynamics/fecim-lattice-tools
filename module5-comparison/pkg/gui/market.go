// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains market analysis visualizations.
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

// MarketSegment represents a market segment with growth data.
type MarketSegment struct {
	Name   string
	Y2025  float64 // Billion USD
	Y2030  float64 // Billion USD
	Color  color.RGBA
}

// marketData holds the market opportunity data.
var marketData = []MarketSegment{
	{Name: "NAND Flash", Y2025: 78, Y2030: 98, Color: color.RGBA{200, 100, 100, 255}},
	{Name: "DRAM", Y2025: 143, Y2030: 220, Color: color.RGBA{100, 150, 200, 255}},
	{Name: "AI Semiconductor", Y2025: 163, Y2030: 403, Color: color.RGBA{100, 200, 150, 255}},
}

// MarketOpportunityChart shows the market opportunity visualization.
type MarketOpportunityChart struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	pulsePhase   float64
	raster       *canvas.Raster
	minSize      fyne.Size
}

// NewMarketOpportunityChart creates a new market chart.
func NewMarketOpportunityChart() *MarketOpportunityChart {
	m := &MarketOpportunityChart{
		minSize: fyne.NewSize(500, 200),
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateAnimation advances the animation.
func (m *MarketOpportunityChart) UpdateAnimation(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.animProgress < 1.0 {
		m.animProgress += dt * 0.5 // 2 seconds to fill
		if m.animProgress > 1.0 {
			m.animProgress = 1.0
		}
	}

	m.pulsePhase += dt * 2.0
}

// Reset resets the animation.
func (m *MarketOpportunityChart) Reset() {
	m.mu.Lock()
	m.animProgress = 0
	m.pulsePhase = 0
	m.mu.Unlock()
	m.Refresh()
}

// MinSize returns minimum size.
func (m *MarketOpportunityChart) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *MarketOpportunityChart) CreateRenderer() fyne.WidgetRenderer {
	m.raster = canvas.NewRaster(m.generateImage)
	return widget.NewSimpleRenderer(m.raster)
}

// generateImage creates the market chart.
func (m *MarketOpportunityChart) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 300 || h < 150 {
		return img
	}

	m.mu.RLock()
	progress := m.animProgress
	pulsePhase := m.pulsePhase
	m.mu.RUnlock()

	// Layout
	padding := 20
	labelWidth := 100
	chartWidth := w - 2*padding - labelWidth
	barGroupWidth := chartWidth / len(marketData)
	maxVal := 450.0 // Max Y value for scaling

	// Title
	drawTextSimple(img, "MARKET OPPORTUNITY ($B)", w/2-100, 12, color.RGBA{0, 212, 255, 255}, 14)

	// Calculate totals
	total2025 := 0.0
	total2030 := 0.0
	for _, seg := range marketData {
		total2025 += seg.Y2025
		total2030 += seg.Y2030
	}

	// Draw bars for each segment
	chartStartX := padding + labelWidth
	chartStartY := padding + 35
	chartHeight := h - chartStartY - 50

	for i, seg := range marketData {
		groupX := chartStartX + i*barGroupWidth

		// 2025 bar
		bar2025Height := int(float64(chartHeight) * (seg.Y2025 / maxVal) * progress)
		bar2025X := groupX + 10
		bar2025Y := chartStartY + chartHeight - bar2025Height
		barWidth := (barGroupWidth - 30) / 2

		// Draw 2025 bar
		darkColor := color.RGBA{seg.Color.R / 2, seg.Color.G / 2, seg.Color.B / 2, 255}
		for dy := 0; dy < bar2025Height; dy++ {
			for dx := 0; dx < barWidth; dx++ {
				img.Set(bar2025X+dx, bar2025Y+dy, darkColor)
			}
		}

		// 2030 bar
		bar2030Height := int(float64(chartHeight) * (seg.Y2030 / maxVal) * progress)
		bar2030X := groupX + 10 + barWidth + 5
		bar2030Y := chartStartY + chartHeight - bar2030Height

		// Draw 2030 bar
		for dy := 0; dy < bar2030Height; dy++ {
			for dx := 0; dx < barWidth; dx++ {
				img.Set(bar2030X+dx, bar2030Y+dy, seg.Color)
			}
		}

		// Growth arrow
		if bar2025Height > 0 && bar2030Height > 0 {
			arrowColor := color.RGBA{100, 255, 150, 200}
			// Arrow line
			for ay := bar2025Y; ay > bar2030Y; ay -= 3 {
				img.Set(bar2025X+barWidth/2, ay, arrowColor)
			}
			// Arrow head
			for ax := -3; ax <= 3; ax++ {
				img.Set(bar2030X+barWidth/2+ax, bar2030Y+5, arrowColor)
			}
		}

		// Segment label (below)
		labelY := chartStartY + chartHeight + 5
		drawTextSimple(img, seg.Name, groupX+5, labelY, color.RGBA{180, 180, 180, 255}, 9)

		// Values on bars
		if progress >= 1.0 {
			val2025 := int(seg.Y2025 * progress)
			val2030 := int(seg.Y2030 * progress)
			drawTextSimple(img, "$"+formatNumber(float64(val2025))+"B", bar2025X, bar2025Y-12, darkColor, 8)
			drawTextSimple(img, "$"+formatNumber(float64(val2030))+"B", bar2030X, bar2030Y-12, seg.Color, 8)
		}
	}

	// Year labels
	drawTextSimple(img, "2025", chartStartX+10, chartStartY+chartHeight+20, color.RGBA{150, 150, 150, 255}, 10)
	drawTextSimple(img, "2030", chartStartX+chartWidth-50, chartStartY+chartHeight+20, color.RGBA{200, 200, 200, 255}, 10)

	// Total headline with pulse
	if progress >= 1.0 {
		pulse := 0.7 + math.Sin(pulsePhase)*0.3
		headlineColor := color.RGBA{
			uint8(float64(0) * pulse),
			uint8(float64(212) * pulse),
			uint8(float64(255) * pulse),
			255,
		}

		// Hero number
		totalText := "$711B by 2030"
		drawTextSimple(img, totalText, padding, chartStartY+chartHeight/2, headlineColor, 16)

		// Subtext
		drawTextSimple(img, "FeCIM can", padding+5, chartStartY+chartHeight/2+20, color.RGBA{150, 150, 150, 255}, 10)
		drawTextSimple(img, "address ALL", padding+5, chartStartY+chartHeight/2+35, color.RGBA{100, 200, 150, 255}, 10)
	}

	return img
}

// Competitor represents a competitor in the matrix.
type Competitor struct {
	Name      string
	Energy    string
	InMemory  int // 0=no, 1=partial, 2=yes
	CMOS      int
	Scalable  int
	Highlight bool
}

// competitors data for the competitive matrix.
var competitors = []Competitor{
	{"FeCIM", "1-10 fJ*", 2, 2, 2, true},
	{"Google TPU", "~100 fJ", 0, 2, 2, false},
	{"Intel Loihi 2", "~10 fJ", 2, 0, 0, false},
	{"Mythic AI", "~5 fJ", 2, 0, 1, false},
}

// CompetitiveMatrix shows competitive comparison.
type CompetitiveMatrix struct {
	widget.BaseWidget
}

// NewCompetitiveMatrix creates a new competitive matrix.
func NewCompetitiveMatrix() *CompetitiveMatrix {
	c := &CompetitiveMatrix{}
	c.ExtendBaseWidget(c)
	return c
}

// MinSize returns minimum size.
func (c *CompetitiveMatrix) MinSize() fyne.Size {
	return fyne.NewSize(500, 180)
}

// CreateRenderer implements fyne.Widget.
func (c *CompetitiveMatrix) CreateRenderer() fyne.WidgetRenderer {
	// Build the table using Fyne widgets
	header := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Technology", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Energy/MAC", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("In-Memory", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("CMOS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Scalable", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	rows := container.NewVBox()
	for _, comp := range competitors {
		nameLabel := widget.NewLabel(comp.Name)
		if comp.Highlight {
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
		}

		energyLabel := widget.NewLabel(comp.Energy)

		row := container.NewGridWithColumns(5,
			nameLabel,
			energyLabel,
			createStatusLabel(comp.InMemory),
			createStatusLabel(comp.CMOS),
			createStatusLabel(comp.Scalable),
		)
		rows.Add(row)
	}

	// Disclaimer
	disclaimer := widget.NewLabel("* TRL 4 - Lab validation only")
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		widget.NewLabelWithStyle("Competitive Comparison", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		header,
		widget.NewSeparator(),
		rows,
		widget.NewSeparator(),
		disclaimer,
	)

	return widget.NewSimpleRenderer(content)
}

// createStatusLabel creates a label showing status (checkmark, cross, or partial).
func createStatusLabel(status int) *widget.Label {
	var text string
	switch status {
	case 0:
		text = "✗"
	case 1:
		text = "◐"
	case 2:
		text = "✓"
	}
	label := widget.NewLabel(text)
	label.Alignment = fyne.TextAlignCenter
	return label
}
