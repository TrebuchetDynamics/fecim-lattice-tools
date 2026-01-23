// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains market analysis visualizations.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MarketSegment represents a market segment with growth data.
type MarketSegment struct {
	Name  string
	Y2025 float64 // Billion USD
	Y2030 float64 // Billion USD
	Color color.RGBA
}

// marketData holds the market opportunity data.
var marketData = []MarketSegment{
	{Name: "NAND Flash", Y2025: 78, Y2030: 98, Color: color.RGBA{200, 100, 100, 255}},
	{Name: "DRAM", Y2025: 143, Y2030: 220, Color: color.RGBA{100, 150, 200, 255}},
	{Name: "AI Semiconductor", Y2025: 163, Y2030: 403, Color: color.RGBA{100, 200, 150, 255}},
}

// MarketOpportunityChart shows the market opportunity visualization using Fyne widgets.
type MarketOpportunityChart struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	pulsePhase   float64
	minSize      fyne.Size

	container *fyne.Container
	totalText *canvas.Text
	bars2025  []*canvas.Rectangle
	bars2030  []*canvas.Rectangle
	values    []*widget.Label
}

// NewMarketOpportunityChart creates a new market chart.
func NewMarketOpportunityChart() *MarketOpportunityChart {
	m := &MarketOpportunityChart{
		minSize:  fyne.NewSize(350, 120),
		bars2025: make([]*canvas.Rectangle, len(marketData)),
		bars2030: make([]*canvas.Rectangle, len(marketData)),
		values:   make([]*widget.Label, len(marketData)),
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateAnimation advances the animation.
func (m *MarketOpportunityChart) UpdateAnimation(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.animProgress < 1.0 {
		m.animProgress += dt * 0.5
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
	m.totalText = canvas.NewText("$721B by 2030", color.RGBA{0, 212, 255, 255})
	m.totalText.TextSize = 24 // Increased size
	m.totalText.TextStyle = fyne.TextStyle{Bold: true}

	var segmentWidgets []fyne.CanvasObject
	maxVal := float32(450.0)
	barHeight := float32(50)

	for i, seg := range marketData {
		shortName := seg.Name
		if len(shortName) > 6 {
			shortName = shortName[:6]
		}
		segLabel := widget.NewLabel(shortName)

		darkColor := color.RGBA{seg.Color.R / 2, seg.Color.G / 2, seg.Color.B / 2, 255}
		m.bars2025[i] = canvas.NewRectangle(darkColor)
		m.bars2025[i].SetMinSize(fyne.NewSize(15, barHeight*float32(seg.Y2025)/maxVal))

		m.bars2030[i] = canvas.NewRectangle(seg.Color)
		m.bars2030[i].SetMinSize(fyne.NewSize(15, barHeight*float32(seg.Y2030)/maxVal))

		m.values[i] = widget.NewLabel(fmt.Sprintf("$%.0fB", seg.Y2030))

		barPair := container.NewHBox(m.bars2025[i], m.bars2030[i])
		segCol := container.NewVBox(segLabel, barPair, m.values[i])
		segmentWidgets = append(segmentWidgets, segCol)
	}

	barsRow := container.NewHBox(segmentWidgets...)

	citation := widget.NewLabel("Source: Gartner 2025 AI Semiconductor Forecast")
	citation.TextStyle = fyne.TextStyle{Italic: true}
	citation.Alignment = fyne.TextAlignCenter

	m.container = container.NewVBox(container.NewCenter(m.totalText), barsRow, citation)
	return widget.NewSimpleRenderer(m.container)
}

// Refresh updates the widget display.
func (m *MarketOpportunityChart) Refresh() {
	m.mu.RLock()
	progress := m.animProgress
	pulsePhase := m.pulsePhase
	m.mu.RUnlock()

	if m.totalText == nil {
		return
	}

	// Pulse total text when done
	if progress >= 1.0 {
		pulse := 0.7 + math.Sin(pulsePhase)*0.3
		m.totalText.Color = color.RGBA{
			0,
			uint8(212 * pulse),
			uint8(255 * pulse),
			255,
		}
	} else {
		m.totalText.Color = color.RGBA{0, 150, 200, 255}
	}

	// Update bar heights based on progress
	maxVal := float32(450.0)
	barHeight := float32(80)

	for i, seg := range marketData {
		bar2025Height := barHeight * float32(seg.Y2025) / maxVal * float32(progress)
		m.bars2025[i].SetMinSize(fyne.NewSize(25, max(2, bar2025Height)))

		bar2030Height := barHeight * float32(seg.Y2030) / maxVal * float32(progress)
		m.bars2030[i].SetMinSize(fyne.NewSize(25, max(2, bar2030Height)))

		m.values[i].SetText(fmt.Sprintf("$%.0fB", seg.Y2030*progress))

		canvas.Refresh(m.bars2025[i])
		canvas.Refresh(m.bars2030[i])
	}

	canvas.Refresh(m.totalText)
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
	return fyne.NewSize(350, 130)
}

// CreateRenderer implements fyne.Widget.
func (c *CompetitiveMatrix) CreateRenderer() fyne.WidgetRenderer {
	header := container.NewGridWithColumns(5,
		widget.NewLabel("Tech"),
		widget.NewLabel("Energy"),
		widget.NewLabel("Mem"),
		widget.NewLabel("CMOS"),
		widget.NewLabel("Scale"),
	)

	rows := container.NewVBox()
	for _, comp := range competitors {
		nameLabel := widget.NewLabel(comp.Name)
		if comp.Highlight {
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
		}
		row := container.NewGridWithColumns(5,
			nameLabel,
			widget.NewLabel(comp.Energy),
			createStatusLabel(comp.InMemory),
			createStatusLabel(comp.CMOS),
			createStatusLabel(comp.Scalable),
		)
		rows.Add(row)
	}

	content := container.NewVBox(header, rows)
	return widget.NewSimpleRenderer(content)
}

// createStatusLabel creates an icon showing status.
func createStatusLabel(status int) fyne.CanvasObject {
	var icon fyne.Resource
	switch status {
	case 0:
		icon = theme.CancelIcon()
	case 1:
		icon = theme.WarningIcon()
	case 2:
		icon = theme.ConfirmIcon()
	}
	return widget.NewIcon(icon)
}

// formatNumberMarket formats numbers with commas.
func formatNumberMarket(n float64) string {
	return fmt.Sprintf("%.0f", n)
}
