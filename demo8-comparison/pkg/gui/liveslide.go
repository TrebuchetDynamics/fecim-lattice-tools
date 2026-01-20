// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ComparisonMode represents the current demo mode.
type ComparisonMode int

const (
	ComparisonModeIdle ComparisonMode = iota
	ComparisonModeCalculating
	ComparisonModeComparing
)

func (m ComparisonMode) String() string {
	switch m {
	case ComparisonModeIdle:
		return "IDLE"
	case ComparisonModeCalculating:
		return "CALCULATING"
	case ComparisonModeComparing:
		return "COMPARING"
	default:
		return "UNKNOWN"
	}
}

// ComparisonModeIndicator shows the current mode.
type ComparisonModeIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	mode    ComparisonMode
	minSize fyne.Size
}

// NewComparisonModeIndicator creates a new mode indicator.
func NewComparisonModeIndicator() *ComparisonModeIndicator {
	m := &ComparisonModeIndicator{
		mode:    ComparisonModeIdle,
		minSize: fyne.NewSize(120, 40),
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetMode updates the current mode.
func (m *ComparisonModeIndicator) SetMode(mode ComparisonMode) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	m.Refresh()
}

// MinSize returns the minimum size.
func (m *ComparisonModeIndicator) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *ComparisonModeIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &comparisonModeRenderer{indicator: m}
}

type comparisonModeRenderer struct {
	indicator *ComparisonModeIndicator
	objects   []fyne.CanvasObject
}

func (r *comparisonModeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *comparisonModeRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *comparisonModeRenderer) Refresh() {
	r.indicator.mu.RLock()
	mode := r.indicator.mode
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.indicator.Size()

	var bgColor, borderColor color.RGBA
	var modeText string

	switch mode {
	case ComparisonModeIdle:
		bgColor = color.RGBA{60, 60, 80, 255}
		borderColor = color.RGBA{100, 100, 130, 255}
		modeText = "IDLE"
	case ComparisonModeCalculating:
		bgColor = color.RGBA{80, 120, 50, 255}
		borderColor = color.RGBA{140, 200, 100, 255}
		modeText = "CALCULATING"
	case ComparisonModeComparing:
		bgColor = color.RGBA{50, 80, 150, 255}
		borderColor = color.RGBA{100, 150, 255, 255}
		modeText = "COMPARING"
	}

	// Border
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background
	padding := float32(2)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Mode text
	text := canvas.NewText(modeText, color.White)
	fontSize := size.Height * 0.35
	if fontSize > 14 {
		fontSize = 14
	}
	if fontSize < 10 {
		fontSize = 10
	}
	text.TextSize = fontSize
	text.TextStyle = fyne.TextStyle{Bold: true}
	textWidth := float32(len(modeText)) * fontSize * 0.6
	text.Move(fyne.NewPos((size.Width-textWidth)/2, (size.Height-fontSize)/2))
	r.objects = append(r.objects, text)
}

func (r *comparisonModeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *comparisonModeRenderer) Destroy() {}

// ComparisonEducationalPanel shows explanations.
type ComparisonEducationalPanel struct {
	widget.BaseWidget

	mu      sync.RWMutex
	title   string
	content string
	minSize fyne.Size
}

// NewComparisonEducationalPanel creates a new educational panel.
func NewComparisonEducationalPanel() *ComparisonEducationalPanel {
	e := &ComparisonEducationalPanel{
		title:   "Why CIM Wins",
		content: "Compute-in-memory eliminates\nthe memory bottleneck.",
		minSize: fyne.NewSize(200, 200),
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetContent updates the content.
func (e *ComparisonEducationalPanel) SetContent(title, content string) {
	e.mu.Lock()
	e.title = title
	e.content = content
	e.mu.Unlock()
	e.Refresh()
}

// SetComparison sets comparison explanation.
func (e *ComparisonEducationalPanel) SetComparison(cpuRatio, gpuRatio float64) {
	content := "THE MEMORY WALL\n\n" +
		"Traditional CPUs/GPUs:\n" +
		"  Data moves between\n" +
		"  memory and processor.\n" +
		"  This wastes energy.\n\n" +
		"Compute-in-Memory:\n" +
		"  Computation happens\n" +
		"  WHERE data lives.\n" +
		"  No movement = no waste.\n\n" +
		fmt.Sprintf("FeCIM vs CPU: %.0f× less power*\n", cpuRatio) +
		fmt.Sprintf("FeCIM vs GPU: %.0f× less power*\n", gpuRatio) +
		"\n* If claims hold (TRL 4)"
	e.SetContent("Why CIM Wins", content)
}

// MinSize returns the minimum size.
func (e *ComparisonEducationalPanel) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *ComparisonEducationalPanel) CreateRenderer() fyne.WidgetRenderer {
	e.mu.RLock()
	title := e.title
	content := e.content
	e.mu.RUnlock()

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	contentLabel := widget.NewLabel(content)
	contentLabel.Wrapping = fyne.TextWrapWord

	box := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		contentLabel,
	)

	return widget.NewSimpleRenderer(box)
}

// ComparisonOperationLog shows timestamped operations.
type ComparisonOperationLog struct {
	widget.BaseWidget

	mu         sync.RWMutex
	entries    []string
	maxEntries int
	startTime  time.Time
	minSize    fyne.Size

	titleLabel   *widget.Label
	contentLabel *widget.Label
}

// NewComparisonOperationLog creates a new operation log.
func NewComparisonOperationLog() *ComparisonOperationLog {
	o := &ComparisonOperationLog{
		maxEntries: 8,
		startTime:  time.Now(),
		minSize:    fyne.NewSize(200, 150),
		entries:    make([]string, 0, 8),
	}
	o.titleLabel = widget.NewLabelWithStyle("Calculation Log", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	o.contentLabel = widget.NewLabel("Ready for calculations...")
	o.contentLabel.Wrapping = fyne.TextWrapWord
	o.ExtendBaseWidget(o)
	return o
}

// Add adds a new log entry.
func (o *ComparisonOperationLog) Add(entry string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	elapsed := time.Since(o.startTime).Seconds()
	timestamped := fmt.Sprintf("t=%.1fs >> %s", elapsed, entry)
	o.entries = append(o.entries, timestamped)

	if len(o.entries) > o.maxEntries {
		o.entries = o.entries[1:]
	}

	o.updateContent()
}

func (o *ComparisonOperationLog) updateContent() {
	if len(o.entries) == 0 {
		o.contentLabel.SetText("Ready for calculations...")
		return
	}
	o.contentLabel.SetText(strings.Join(o.entries, "\n"))
}

// MinSize returns the minimum size.
func (o *ComparisonOperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *ComparisonOperationLog) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		o.contentLabel,
	)
	return widget.NewSimpleRenderer(box)
}
