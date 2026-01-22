// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file implements the "Live Slide" pattern components for Demo 4.
package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// CircuitMode represents the current demo mode.
type CircuitMode int

const (
	CircuitModeIdle CircuitMode = iota
	CircuitModeDAC
	CircuitModeADC
	CircuitModeTIA
	CircuitModePump
	CircuitModeWrite
	CircuitModeRead
)

func (m CircuitMode) String() string {
	switch m {
	case CircuitModeIdle:
		return "IDLE"
	case CircuitModeDAC:
		return "DAC"
	case CircuitModeADC:
		return "ADC"
	case CircuitModeTIA:
		return "TIA"
	case CircuitModePump:
		return "PUMP"
	case CircuitModeWrite:
		return "WRITE"
	case CircuitModeRead:
		return "READ"
	default:
		return "UNKNOWN"
	}
}

// CircuitModeIndicator shows the current mode with a colored background.
type CircuitModeIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	mode    CircuitMode
	minSize fyne.Size
}

// NewCircuitModeIndicator creates a new mode indicator.
func NewCircuitModeIndicator() *CircuitModeIndicator {
	m := &CircuitModeIndicator{
		mode:    CircuitModeIdle,
		minSize: fyne.NewSize(120, 50),
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetMode updates the current mode.
func (m *CircuitModeIndicator) SetMode(mode CircuitMode) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	m.Refresh()
}

// GetMode returns the current mode.
func (m *CircuitModeIndicator) GetMode() CircuitMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// MinSize returns the minimum size.
func (m *CircuitModeIndicator) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *CircuitModeIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &circuitModeRenderer{indicator: m}
}

type circuitModeRenderer struct {
	indicator *CircuitModeIndicator
	objects   []fyne.CanvasObject
}

func (r *circuitModeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *circuitModeRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *circuitModeRenderer) Refresh() {
	r.indicator.mu.RLock()
	mode := r.indicator.mode
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.indicator.Size()

	// Colors based on mode
	var bgColor, borderColor color.RGBA
	var modeText string

	switch mode {
	case CircuitModeIdle:
		bgColor = color.RGBA{60, 60, 80, 255}
		borderColor = color.RGBA{100, 100, 130, 255}
		modeText = "░░ IDLE ░░"
	case CircuitModeDAC:
		bgColor = color.RGBA{80, 50, 150, 255}
		borderColor = color.RGBA{140, 100, 220, 255}
		modeText = "▼▼ DAC ▼▼"
	case CircuitModeADC:
		bgColor = color.RGBA{50, 150, 80, 255}
		borderColor = color.RGBA{100, 220, 130, 255}
		modeText = "▲▲ ADC ▲▲"
	case CircuitModeTIA:
		bgColor = color.RGBA{180, 120, 50, 255}
		borderColor = color.RGBA{255, 180, 100, 255}
		modeText = "⚡ TIA ⚡"
	case CircuitModePump:
		bgColor = color.RGBA{180, 50, 50, 255}
		borderColor = color.RGBA{255, 100, 100, 255}
		modeText = "⬆⬆ PUMP ⬆⬆"
	case CircuitModeWrite:
		bgColor = color.RGBA{200, 80, 80, 255}
		borderColor = color.RGBA{255, 130, 130, 255}
		modeText = "██ WRITE ██"
	case CircuitModeRead:
		bgColor = color.RGBA{50, 120, 180, 255}
		borderColor = color.RGBA{100, 180, 255, 255}
		modeText = "░░ READ ░░"
	}

	// Border
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background
	padding := float32(3)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Mode text
	text := canvas.NewText(modeText, color.White)
	text.TextSize = 14
	text.TextStyle = fyne.TextStyle{Bold: true}
	textWidth := float32(len(modeText) * 8)
	text.Move(fyne.NewPos((size.Width-textWidth)/2, (size.Height-20)/2))
	r.objects = append(r.objects, text)
}

func (r *circuitModeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *circuitModeRenderer) Destroy() {}

// CircuitEducationalPanel shows context-sensitive explanations.
type CircuitEducationalPanel struct {
	widget.BaseWidget

	mu      sync.RWMutex
	title   string
	content string
	minSize fyne.Size
}

// NewCircuitEducationalPanel creates a new educational panel.
func NewCircuitEducationalPanel() *CircuitEducationalPanel {
	e := &CircuitEducationalPanel{
		title:   "What You're Seeing",
		content: "Select a circuit to learn\nabout peripheral systems.",
		minSize: fyne.NewSize(200, 200),
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetContent updates the educational content.
func (e *CircuitEducationalPanel) SetContent(title, content string) {
	e.mu.Lock()
	e.title = title
	e.content = content
	e.mu.Unlock()
	e.Refresh()
}

// SetDACExplanation sets content for DAC.
func (e *CircuitEducationalPanel) SetDACExplanation() {
	content := "DAC (Digital-to-Analog)\n\n" +
		"Converts digital level (0-29)\n" +
		"to analog voltage.\n\n" +
		"1. Receive 5-bit code\n" +
		"2. Resistor ladder divides\n" +
		"   reference voltage\n" +
		"3. Output: 0V to 1.5V\n\n" +
		"Settling time: ~10ns\n" +
		"Energy: ~1 fJ"
	e.SetContent("DAC: Write Path", content)
}

// SetADCExplanation sets content for ADC.
func (e *CircuitEducationalPanel) SetADCExplanation() {
	content := "ADC (Analog-to-Digital)\n\n" +
		"Converts analog voltage\n" +
		"back to digital level.\n\n" +
		"1. Sample input voltage\n" +
		"2. SAR: Compare/divide\n" +
		"3. Output: 5-bit code\n\n" +
		"Conversion time: ~50ns\n" +
		"ENOB: ~4.8 bits"
	e.SetContent("ADC: Read Path", content)
}

// SetTIAExplanation sets content for TIA.
func (e *CircuitEducationalPanel) SetTIAExplanation() {
	content := "TIA (Transimpedance Amp)\n\n" +
		"Converts cell current to\n" +
		"voltage for ADC input.\n\n" +
		"I_cell → V_out\n" +
		"Gain: 10-100 kOhm\n\n" +
		"Bandwidth: ~100 MHz\n" +
		"Dynamic range: 60+ dB"
	e.SetContent("TIA: Current Sense", content)
}

// SetPumpExplanation sets content for charge pump.
func (e *CircuitEducationalPanel) SetPumpExplanation() {
	content := "Charge Pump\n\n" +
		"Boosts 1V CMOS supply\n" +
		"to 1.5V for FeFET write.\n\n" +
		"Dickson topology:\n" +
		"Capacitor chain pumps\n" +
		"charge to higher voltage.\n\n" +
		"Efficiency: ~80%\n" +
		"Rise time: ~1 us"
	e.SetContent("Charge Pump: Boost", content)
}

// SetWriteCycleExplanation sets content for write cycle.
func (e *CircuitEducationalPanel) SetWriteCycleExplanation(phase int) {
	var content string
	switch phase {
	case 1:
		content = "WRITE CYCLE (1/4)\n\n" +
			"1. Digital → DAC\n" +
			"   Level 0-29 converted\n" +
			"   to analog voltage"
	case 2:
		content = "WRITE CYCLE (2/4)\n\n" +
			"2. Voltage → Charge Pump\n" +
			"   Boosted to write level\n" +
			"   (requires > Vcc)"
	case 3:
		content = "WRITE CYCLE (3/4)\n\n" +
			"3. Program → FeFET\n" +
			"   Polarization set by\n" +
			"   applied voltage"
	case 4:
		content = "WRITE CYCLE (4/4)\n\n" +
			"4. Verify → Read back\n" +
			"   Confirm programmed\n" +
			"   level is correct"
	default:
		content = "WRITE CYCLE\n\n" +
			"Digital → Voltage →\n" +
			"Program → Verify\n\n" +
			"Full path through\n" +
			"peripheral circuits."
	}
	e.SetContent("Write Operation", content)
}

// SetReadCycleExplanation sets content for read cycle.
func (e *CircuitEducationalPanel) SetReadCycleExplanation(phase int) {
	var content string
	switch phase {
	case 1:
		content = "READ CYCLE (1/4)\n\n" +
			"1. Apply V_read\n" +
			"   Small read voltage\n" +
			"   (non-destructive)"
	case 2:
		content = "READ CYCLE (2/4)\n\n" +
			"2. Current → TIA\n" +
			"   Cell current sensed\n" +
			"   and amplified"
	case 3:
		content = "READ CYCLE (3/4)\n\n" +
			"3. Voltage → ADC\n" +
			"   TIA output digitized\n" +
			"   to 5-bit code"
	case 4:
		content = "READ CYCLE (4/4)\n\n" +
			"4. Output → Level\n" +
			"   Digital value ready\n" +
			"   for computation"
	default:
		content = "READ CYCLE\n\n" +
			"Apply → Sense →\n" +
			"Convert → Output\n\n" +
			"Non-destructive read\n" +
			"through TIA + ADC."
	}
	e.SetContent("Read Operation", content)
}

// SetIdleExplanation sets content for idle state.
func (e *CircuitEducationalPanel) SetIdleExplanation() {
	content := "PERIPHERAL CIRCUITS\n\n" +
		"\"Works on a standard\n" +
		"CMOS line\"\n\n" +
		"— Dr. external research group\n\n" +
		"Standard circuits enable\n" +
		"ferroelectric CIM in\n" +
		"existing fabs."
	e.SetContent("What You're Seeing", content)
}

// MinSize returns the minimum size.
func (e *CircuitEducationalPanel) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *CircuitEducationalPanel) CreateRenderer() fyne.WidgetRenderer {
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

// CircuitOperationLog shows timestamped operation history.
type CircuitOperationLog struct {
	widget.BaseWidget

	mu         sync.RWMutex
	entries    []string
	maxEntries int
	startTime  time.Time
	minSize    fyne.Size

	titleLabel   *widget.Label
	contentLabel *widget.Label
}

// NewCircuitOperationLog creates a new operation log.
func NewCircuitOperationLog() *CircuitOperationLog {
	o := &CircuitOperationLog{
		maxEntries: 10,
		startTime:  time.Now(),
		minSize:    fyne.NewSize(200, 180),
		entries:    make([]string, 0, 10),
	}
	o.titleLabel = widget.NewLabelWithStyle("Operation Log", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	o.contentLabel = widget.NewLabel("Waiting for operations...")
	o.contentLabel.Wrapping = fyne.TextWrapWord
	o.ExtendBaseWidget(o)
	return o
}

// Add adds a new log entry with timestamp.
func (o *CircuitOperationLog) Add(entry string) {
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

// Clear clears all log entries.
func (o *CircuitOperationLog) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.entries = o.entries[:0]
	o.startTime = time.Now()
	o.updateContent()
}

func (o *CircuitOperationLog) updateContent() {
	if len(o.entries) == 0 {
		o.contentLabel.SetText("Waiting for operations...")
		return
	}
	o.contentLabel.SetText(strings.Join(o.entries, "\n"))
}

// MinSize returns the minimum size.
func (o *CircuitOperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *CircuitOperationLog) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		o.contentLabel,
	)
	return widget.NewSimpleRenderer(box)
}

// CircuitKeyStat displays a key statistic prominently.
type CircuitKeyStat struct {
	widget.BaseWidget

	mu      sync.RWMutex
	label   string
	value   string
	minSize fyne.Size
}

// NewCircuitKeyStat creates a new key stat box.
func NewCircuitKeyStat(label, value string) *CircuitKeyStat {
	k := &CircuitKeyStat{
		label:   label,
		value:   value,
		minSize: fyne.NewSize(150, 60),
	}
	k.ExtendBaseWidget(k)
	return k
}

// SetValue updates the statistic value.
func (k *CircuitKeyStat) SetValue(value string) {
	k.mu.Lock()
	k.value = value
	k.mu.Unlock()
	k.Refresh()
}

// MinSize returns the minimum size.
func (k *CircuitKeyStat) MinSize() fyne.Size {
	return k.minSize
}

// CreateRenderer implements fyne.Widget.
func (k *CircuitKeyStat) CreateRenderer() fyne.WidgetRenderer {
	k.mu.RLock()
	label := k.label
	value := k.value
	k.mu.RUnlock()

	labelWidget := widget.NewLabel(label)
	labelWidget.Alignment = fyne.TextAlignCenter

	valueWidget := widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	box := container.NewVBox(labelWidget, valueWidget)
	return widget.NewSimpleRenderer(box)
}
