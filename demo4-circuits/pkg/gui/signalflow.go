// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
package gui

import (
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// SignalFlowDiagram visualizes the DAC → Cell → ADC signal path.
type SignalFlowDiagram struct {
	widget.BaseWidget

	mu sync.RWMutex

	// Current values
	digitalIn  int     // 0-29
	analogV    float64 // Voltage
	current    float64 // Current through cell
	analogOut  float64 // TIA output voltage
	digitalOut int     // 0-29

	// Animation state
	activeStage int // 0=idle, 1=DAC, 2=Pump, 3=Cell, 4=TIA, 5=ADC

	raster  *canvas.Raster
	minSize fyne.Size
}

// NewSignalFlowDiagram creates a new signal flow visualization.
func NewSignalFlowDiagram() *SignalFlowDiagram {
	s := &SignalFlowDiagram{
		digitalIn:   15,
		analogV:     0.75,
		current:     50e-6,
		analogOut:   0.5,
		digitalOut:  15,
		activeStage: 0,
		minSize:     fyne.NewSize(600, 200),
	}
	s.ExtendBaseWidget(s)
	return s
}

// SetDigitalInput sets the input digital level.
func (s *SignalFlowDiagram) SetDigitalInput(level int) {
	s.mu.Lock()
	s.digitalIn = level
	s.mu.Unlock()
	s.Refresh()
}

// SetActiveStage highlights a stage in the signal flow.
func (s *SignalFlowDiagram) SetActiveStage(stage int) {
	s.mu.Lock()
	s.activeStage = stage
	s.mu.Unlock()
	s.Refresh()
}

// SetValues updates all values in the signal chain.
func (s *SignalFlowDiagram) SetValues(digitalIn int, analogV, current, analogOut float64, digitalOut int) {
	s.mu.Lock()
	s.digitalIn = digitalIn
	s.analogV = analogV
	s.current = current
	s.analogOut = analogOut
	s.digitalOut = digitalOut
	s.mu.Unlock()
	s.Refresh()
}

// MinSize returns the minimum size.
func (s *SignalFlowDiagram) MinSize() fyne.Size {
	return s.minSize
}

// CreateRenderer implements fyne.Widget.
func (s *SignalFlowDiagram) CreateRenderer() fyne.WidgetRenderer {
	s.raster = canvas.NewRaster(s.generateImage)
	return widget.NewSimpleRenderer(s.raster)
}

func (s *SignalFlowDiagram) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	s.mu.RLock()
	activeStage := s.activeStage
	s.mu.RUnlock()

	// Define component positions
	boxW := w / 7
	boxH := h / 3
	centerY := h / 2

	// Colors
	boxColor := color.RGBA{60, 60, 80, 255}
	activeColor := color.RGBA{100, 180, 255, 255}
	arrowColor := color.RGBA{150, 150, 170, 255}
	textColor := color.RGBA{200, 200, 220, 255}

	// Draw components: [Digital In] -> [DAC] -> [Pump] -> [Cell] -> [TIA] -> [ADC] -> [Digital Out]
	positions := []struct {
		x    int
		name string
	}{
		{boxW / 2, "IN"},
		{boxW * 3 / 2, "DAC"},
		{boxW * 5 / 2, "PUMP"},
		{boxW * 7 / 2, "CELL"},
		{boxW * 9 / 2, "TIA"},
		{boxW * 11 / 2, "ADC"},
		{boxW * 13 / 2, "OUT"},
	}

	// Draw boxes
	for i, pos := range positions {
		col := boxColor
		if i == activeStage {
			col = activeColor
		}

		// Draw box
		x0 := pos.x - boxW/3
		y0 := centerY - boxH/2
		for y := y0; y < y0+boxH; y++ {
			for x := x0; x < x0+boxW*2/3; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, col)
				}
			}
		}

		// Draw label (simple text approximation)
		s.drawText(img, pos.name, pos.x-len(pos.name)*3, centerY-5, textColor)
	}

	// Draw arrows between components
	for i := 0; i < len(positions)-1; i++ {
		x0 := positions[i].x + boxW/3
		x1 := positions[i+1].x - boxW/3
		y := centerY

		// Arrow line
		for x := x0; x < x1; x++ {
			img.Set(x, y, arrowColor)
			img.Set(x, y-1, arrowColor)
		}

		// Arrow head
		for dy := -4; dy <= 4; dy++ {
			dx := 4 - abs(dy)
			img.Set(x1-dx, y+dy, arrowColor)
		}
	}

	return img
}

func (s *SignalFlowDiagram) drawText(img *image.RGBA, text string, x, y int, col color.Color) {
	// Simple text drawing (just a placeholder - real text would use a font)
	for i, ch := range text {
		s.drawChar(img, ch, x+i*6, y, col)
	}
}

func (s *SignalFlowDiagram) drawChar(img *image.RGBA, ch rune, x, y int, col color.Color) {
	// Very simple 5x7 character approximation
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	// Just draw a simple pattern for each character
	patterns := map[rune][]string{
		'I': {"  #  ", "  #  ", "  #  ", "  #  ", "  #  "},
		'N': {"#   #", "##  #", "# # #", "#  ##", "#   #"},
		'O': {" ### ", "#   #", "#   #", "#   #", " ### "},
		'U': {"#   #", "#   #", "#   #", "#   #", " ### "},
		'T': {"#####", "  #  ", "  #  ", "  #  ", "  #  "},
		'D': {"#### ", "#   #", "#   #", "#   #", "#### "},
		'A': {" ### ", "#   #", "#####", "#   #", "#   #"},
		'C': {" ####", "#    ", "#    ", "#    ", " ####"},
		'P': {"#### ", "#   #", "#### ", "#    ", "#    "},
		'M': {"#   #", "## ##", "# # #", "#   #", "#   #"},
		'E': {"#####", "#    ", "###  ", "#    ", "#####"},
		'L': {"#    ", "#    ", "#    ", "#    ", "#####"},
	}

	pattern, ok := patterns[ch]
	if !ok {
		return
	}

	for dy, row := range pattern {
		for dx, c := range row {
			if c == '#' {
				px := x + dx
				py := y + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, col)
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TimingDiagram shows write/read cycle timing.
type TimingDiagram struct {
	widget.BaseWidget

	mu sync.RWMutex

	// Which cycle to show
	showWrite bool
	phase     int // Current phase in animation

	raster  *canvas.Raster
	minSize fyne.Size
}

// NewTimingDiagram creates a new timing diagram.
func NewTimingDiagram() *TimingDiagram {
	t := &TimingDiagram{
		showWrite: true,
		phase:     0,
		minSize:   fyne.NewSize(500, 200),
	}
	t.ExtendBaseWidget(t)
	return t
}

// SetWriteCycle sets the diagram to show write cycle.
func (t *TimingDiagram) SetWriteCycle() {
	t.mu.Lock()
	t.showWrite = true
	t.mu.Unlock()
	t.Refresh()
}

// SetReadCycle sets the diagram to show read cycle.
func (t *TimingDiagram) SetReadCycle() {
	t.mu.Lock()
	t.showWrite = false
	t.mu.Unlock()
	t.Refresh()
}

// SetPhase sets the current phase for animation.
func (t *TimingDiagram) SetPhase(phase int) {
	t.mu.Lock()
	t.phase = phase
	t.mu.Unlock()
	t.Refresh()
}

// MinSize returns the minimum size.
func (t *TimingDiagram) MinSize() fyne.Size {
	return t.minSize
}

// CreateRenderer implements fyne.Widget.
func (t *TimingDiagram) CreateRenderer() fyne.WidgetRenderer {
	t.raster = canvas.NewRaster(t.generateImage)
	return widget.NewSimpleRenderer(t.raster)
}

func (t *TimingDiagram) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	t.mu.RLock()
	showWrite := t.showWrite
	phase := t.phase
	t.mu.RUnlock()

	// Signal colors
	clkColor := color.RGBA{100, 200, 255, 255}
	sigColor := color.RGBA{100, 255, 150, 255}
	activeColor := color.RGBA{255, 200, 100, 255}
	textColor := color.RGBA{180, 180, 200, 255}
	gridColor := color.RGBA{40, 40, 50, 255}

	// Layout
	labelWidth := 60
	signalHeight := h / 5
	signalWidth := w - labelWidth - 20

	// Draw grid lines
	for x := labelWidth; x < w-20; x += signalWidth / 8 {
		for y := 0; y < h; y++ {
			img.Set(x, y, gridColor)
		}
	}

	// Define signals based on write/read cycle
	var signals []struct {
		name     string
		waveform []int // 0=low, 1=high, 2=rising, 3=falling
	}

	if showWrite {
		signals = []struct {
			name     string
			waveform []int
		}{
			{"CLK", []int{0, 1, 0, 1, 0, 1, 0, 1}},
			{"WREN", []int{1, 1, 1, 1, 1, 1, 0, 0}},
			{"VDAC", []int{0, 2, 1, 1, 1, 1, 3, 0}},
			{"VPUMP", []int{0, 0, 2, 1, 1, 3, 0, 0}},
		}
	} else {
		signals = []struct {
			name     string
			waveform []int
		}{
			{"CLK", []int{0, 1, 0, 1, 0, 1, 0, 1}},
			{"RDEN", []int{1, 1, 1, 1, 1, 1, 0, 0}},
			{"VREAD", []int{0, 2, 1, 1, 3, 0, 0, 0}},
			{"ADC", []int{0, 0, 0, 2, 1, 1, 3, 0}},
		}
	}

	// Draw each signal
	for i, sig := range signals {
		y0 := 10 + i*signalHeight
		y1 := y0 + signalHeight - 10

		// Draw label
		for dx, ch := range sig.name {
			t.drawChar(img, ch, 5+dx*6, y0+signalHeight/3, textColor)
		}

		// Draw waveform
		segWidth := signalWidth / len(sig.waveform)
		for j, level := range sig.waveform {
			x0 := labelWidth + j*segWidth
			x1 := x0 + segWidth

			col := sigColor
			if j == phase%len(sig.waveform) {
				col = activeColor
			}

			switch level {
			case 0: // Low
				for x := x0; x < x1; x++ {
					img.Set(x, y1, col)
				}
			case 1: // High
				for x := x0; x < x1; x++ {
					img.Set(x, y0, col)
				}
			case 2: // Rising
				for x := x0; x < x1; x++ {
					progress := float64(x-x0) / float64(x1-x0)
					y := y1 - int(progress*float64(y1-y0))
					img.Set(x, y, col)
				}
			case 3: // Falling
				for x := x0; x < x1; x++ {
					progress := float64(x-x0) / float64(x1-x0)
					y := y0 + int(progress*float64(y1-y0))
					img.Set(x, y, col)
				}
			}

			// Connect to previous
			if j > 0 {
				prevLevel := sig.waveform[j-1]
				if prevLevel == 0 && level == 1 {
					for y := y0; y <= y1; y++ {
						img.Set(x0, y, col)
					}
				} else if prevLevel == 1 && level == 0 {
					for y := y0; y <= y1; y++ {
						img.Set(x0, y, col)
					}
				}
			}
		}

		// Draw clock if first signal
		if i == 0 {
			for x := labelWidth; x < w-20; x++ {
				img.Set(x, y0-2, clkColor)
			}
		}
	}

	return img
}

func (t *TimingDiagram) drawChar(img *image.RGBA, ch rune, x, y int, col color.Color) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	patterns := map[rune][]string{
		'C': {" ####", "#    ", "#    ", "#    ", " ####"},
		'L': {"#    ", "#    ", "#    ", "#    ", "#####"},
		'K': {"#   #", "#  # ", "###  ", "#  # ", "#   #"},
		'W': {"#   #", "#   #", "# # #", "## ##", "#   #"},
		'R': {"#### ", "#   #", "#### ", "#  # ", "#   #"},
		'E': {"#####", "#    ", "###  ", "#    ", "#####"},
		'N': {"#   #", "##  #", "# # #", "#  ##", "#   #"},
		'V': {"#   #", "#   #", "#   #", " # # ", "  #  "},
		'D': {"#### ", "#   #", "#   #", "#   #", "#### "},
		'A': {" ### ", "#   #", "#####", "#   #", "#   #"},
		'P': {"#### ", "#   #", "#### ", "#    ", "#    "},
		'U': {"#   #", "#   #", "#   #", "#   #", " ### "},
		'M': {"#   #", "## ##", "# # #", "#   #", "#   #"},
	}

	pattern, ok := patterns[ch]
	if !ok {
		return
	}

	for dy, row := range pattern {
		for dx, c := range row {
			if c == '#' {
				px := x + dx
				py := y + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, col)
				}
			}
		}
	}
}
