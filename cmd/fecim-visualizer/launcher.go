package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// DemoInfo holds information about a demo
type DemoInfo struct {
	Number      int
	Title       string
	Subtitle    string
	Description string
	Ready       bool
}

// GetDemos returns all demo information (6 consolidated demos)
func GetDemos() []DemoInfo {
	return []DemoInfo{
		{
			Number:      1,
			Title:       "Hysteresis",
			Subtitle:    "The Memory Cell",
			Description: "How the memory cell works: visualize ferroelectric polarization switching with 30 discrete states",
			Ready:       true,
		},
		{
			Number:      2,
			Title:       "Crossbar+",
			Subtitle:    "MVM + Non-Idealities",
			Description: "How we compute in memory: matrix-vector multiplication with IR drop, sneak paths, and drift analysis",
			Ready:       true,
		},
		{
			Number:      3,
			Title:       "MNIST",
			Subtitle:    "The AI Brain",
			Description: "What we can build: draw digits and watch the neural network classify them at 87% accuracy (FP vs CIM)",
			Ready:       true,
		},
		{
			Number:      4,
			Title:       "Circuits",
			Subtitle:    "The Chip System",
			Description: "How it fits in a real chip: DAC, ADC, TIA, and CMOS-compatible peripheral design",
			Ready:       true,
		},
		{
			Number:      5,
			Title:       "Comparison",
			Subtitle:    "Why FeCIM Wins",
			Description: "The business case: energy efficiency, competitive matrix, data center savings calculator",
			Ready:       true,
		},
		{
			Number:      6,
			Title:       "EDA",
			Subtitle:    "Design Suite",
			Description: "Bridge to open-source EDA: weight compiler, layout visualization, SPICE export for ngspice/KLayout",
			Ready:       true,
		},
	}
}

// DemoCard creates a card widget for a demo
type DemoCard struct {
	widget.BaseWidget
	info     DemoInfo
	onTapped func()
	minSize  fyne.Size
}

// NewDemoCard creates a new demo card
func NewDemoCard(info DemoInfo, onTapped func()) *DemoCard {
	card := &DemoCard{
		info:     info,
		onTapped: onTapped,
		minSize:  fyne.NewSize(280, 180),
	}
	card.ExtendBaseWidget(card)
	return card
}

func (c *DemoCard) MinSize() fyne.Size {
	return c.minSize
}

func (c *DemoCard) Tapped(*fyne.PointEvent) {
	if c.info.Ready && c.onTapped != nil {
		c.onTapped()
	}
}

func (c *DemoCard) TappedSecondary(*fyne.PointEvent) {}

func (c *DemoCard) CreateRenderer() fyne.WidgetRenderer {
	return &demoCardRenderer{card: c}
}

type demoCardRenderer struct {
	card    *DemoCard
	objects []fyne.CanvasObject
}

func (r *demoCardRenderer) MinSize() fyne.Size {
	return r.card.minSize
}

func (r *demoCardRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *demoCardRenderer) Refresh() {
	r.objects = r.objects[:0]
	size := r.card.Size()
	info := r.card.info

	// Background color based on ready state
	var bgColor, borderColor, textColor, descColor color.RGBA
	if info.Ready {
		bgColor = color.RGBA{0, 60, 120, 255}       // Darker blue for ready
		borderColor = color.RGBA{0, 212, 255, 255}  // Cyan border
		textColor = color.RGBA{255, 255, 255, 255}  // White for titles
		descColor = color.RGBA{200, 220, 255, 255}  // Light blue-white for descriptions (high contrast)
	} else {
		bgColor = color.RGBA{30, 40, 50, 200} // Gray for coming soon
		borderColor = color.RGBA{80, 90, 100, 255}
		textColor = color.RGBA{120, 130, 140, 255}
		descColor = color.RGBA{100, 110, 120, 255}
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

	// Scale elements based on card size
	scale := size.Height / 180.0
	if scale < 1 {
		scale = 1
	}

	// Demo number badge
	badgeSize := float32(36) * scale
	if badgeSize > 50 {
		badgeSize = 50
	}
	badgeX := float32(12)
	badgeY := float32(12)

	badgeBg := canvas.NewCircle(borderColor)
	badgeBg.Resize(fyne.NewSize(badgeSize, badgeSize))
	badgeBg.Move(fyne.NewPos(badgeX, badgeY))
	r.objects = append(r.objects, badgeBg)

	numTextSize := float32(20) * scale
	if numTextSize > 28 {
		numTextSize = 28
	}
	numText := canvas.NewText(string('0'+byte(info.Number)), bgColor)
	numText.TextSize = numTextSize
	numText.TextStyle = fyne.TextStyle{Bold: true}
	numText.Move(fyne.NewPos(badgeX+badgeSize/2-numTextSize/4, badgeY+badgeSize/2-numTextSize/2))
	r.objects = append(r.objects, numText)

	// Title
	titleSize := float32(18) * scale
	if titleSize > 26 {
		titleSize = 26
	}
	title := canvas.NewText(info.Title, textColor)
	title.TextSize = titleSize
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(badgeX+badgeSize+12, 14))
	r.objects = append(r.objects, title)

	// Subtitle
	subtitleSize := float32(13) * scale
	if subtitleSize > 18 {
		subtitleSize = 18
	}
	subtitle := canvas.NewText(info.Subtitle, color.RGBA{0, 212, 255, 255}) // Cyan for subtitle
	subtitle.TextSize = subtitleSize
	subtitle.Move(fyne.NewPos(badgeX+badgeSize+12, 14+titleSize+4))
	r.objects = append(r.objects, subtitle)

	// Status badge
	var statusText string
	var statusColor color.RGBA
	if info.Ready {
		statusText = "READY"
		statusColor = color.RGBA{100, 255, 150, 255} // Bright green
	} else {
		statusText = "COMING SOON"
		statusColor = color.RGBA{150, 150, 150, 255}
	}
	status := canvas.NewText(statusText, statusColor)
	status.TextSize = 11
	status.TextStyle = fyne.TextStyle{Bold: true}
	status.Move(fyne.NewPos(size.Width-75, 14))
	r.objects = append(r.objects, status)

	// Description (wrapped manually)
	desc := info.Description

	// Simple word wrap - split into lines
	descSize := float32(13) * scale
	if descSize > 16 {
		descSize = 16
	}
	maxWidth := size.Width - 28
	lineY := badgeY + badgeSize + 16
	lineHeight := descSize + 4

	words := splitWords(desc)
	line := ""
	for _, word := range words {
		testLine := line + word + " "
		testText := canvas.NewText(testLine, descColor)
		testText.TextSize = descSize
		if testText.MinSize().Width > maxWidth && line != "" {
			// Write current line
			lineText := canvas.NewText(line, descColor)
			lineText.TextSize = descSize
			lineText.Move(fyne.NewPos(14, lineY))
			r.objects = append(r.objects, lineText)
			lineY += lineHeight
			line = word + " "
		} else {
			line = testLine
		}
	}
	if line != "" {
		lineText := canvas.NewText(line, descColor)
		lineText.TextSize = descSize
		lineText.Move(fyne.NewPos(14, lineY))
		r.objects = append(r.objects, lineText)
	}

	// Hover hint for ready cards
	if info.Ready {
		hint := canvas.NewText("Click to open →", color.RGBA{0, 212, 255, 200})
		hint.TextSize = 11
		hint.TextStyle = fyne.TextStyle{Bold: true}
		hint.Move(fyne.NewPos(size.Width-95, size.Height-22))
		r.objects = append(r.objects, hint)
	}
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

func (r *demoCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *demoCardRenderer) Destroy() {}

// CreateLauncherContent creates the launcher tab content
func CreateLauncherContent(onDemoSelected func(demoNum int)) fyne.CanvasObject {
	demos := GetDemos()

	// Title - larger and more prominent
	titleLabel := widget.NewLabelWithStyle(
		"FeCIM Visualization Suite",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Subtitle with narrative
	subtitleLabel := widget.NewLabelWithStyle(
		"6 Demos: Physics → Compute → Application → System → Business → Design",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	header := container.NewVBox(
		titleLabel,
		subtitleLabel,
	)

	// Create demo cards
	cards := make([]fyne.CanvasObject, len(demos))
	for i, demo := range demos {
		d := demo // Capture for closure
		cards[i] = NewDemoCard(d, func() {
			if onDemoSelected != nil {
				onDemoSelected(d.Number)
			}
		})
	}

	// Use GridWrap layout for dynamic sizing - 3 columns, 2 rows
	// This will expand to fill available space
	grid := container.New(layout.NewGridLayoutWithRows(2),
		container.New(layout.NewGridLayoutWithColumns(3), cards[0], cards[1], cards[2]),
		container.New(layout.NewGridLayoutWithColumns(3), cards[3], cards[4], cards[5]),
	)

	// Key metrics in footer - single line
	metricsLabel := widget.NewLabelWithStyle(
		"30 Levels (4.9 bits) | 87% MNIST | 10M× vs NAND | 1000× vs DRAM | TRL 4  —  Click any card to explore",
		fyne.TextAlignCenter,
		fyne.TextStyle{Monospace: true},
	)

	footer := container.NewVBox(
		metricsLabel,
	)

	// Use border layout - grid expands to fill center
	return container.NewBorder(
		header,
		footer,
		nil, nil,
		container.NewPadded(grid),
	)
}
