package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// DemoInfo holds information about a demo
type DemoInfo struct {
	Number      int
	Title       string
	Subtitle    string
	Description string
	Icon        string // Unicode icon for the demo
	Ready       bool
	WIP         bool // Work in progress - show indicator but still accessible
}

// GetDemos returns all demo information (6 consolidated demos)
func GetDemos() []DemoInfo {
	return []DemoInfo{
		{
			Number:      1,
			Title:       "Hysteresis",
			Subtitle:    "P-E Curve Physics",
			Description: "Explore the ferroelectric memory effect: how HZO superlattice stores 30 analog states through polarization switching",
			Icon:        "~",
			Ready:       true,
		},
		{
			Number:      2,
			Title:       "Crossbar+",
			Subtitle:    "Compute-in-Memory Array",
			Description: "See matrix-vector multiply in action with real non-idealities: IR drop, sneak paths, and conductance drift",
			Icon:        "#",
			Ready:       true,
		},
		{
			Number:      3,
			Title:       "MNIST",
			Subtitle:    "Neural Network Demo",
			Description: "Draw handwritten digits and watch FeCIM classify them at 87% accuracy (88% theoretical max)",
			Icon:        "9",
			Ready:       true,
		},
		{
			Number:      4,
			Title:       "Circuits",
			Subtitle:    "Peripheral Electronics",
			Description: "Design the analog interface: DAC inputs, TIA sensing, ADC readout for CMOS integration",
			Icon:        "V",
			Ready:       true,
		},
		{
			Number:      5,
			Title:       "Comparison",
			Subtitle:    "Technology Benchmarks",
			Description: "Compare FeCIM vs NAND, DRAM, ReRAM, and competing CIM: 10M× energy savings, 1M× faster",
			Icon:        "$",
			Ready:       true,
		},
		{
			Number:      6,
			Title:       "EDA",
			Subtitle:    "Chip Layout Tools",
			Description: "Build crossbar arrays for OpenLane tapeout: GDS export, DRC checks, SPICE netlist generation",
			Icon:        "L",
			Ready:       true,
			WIP:         true,
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
		minSize:  fyne.NewSize(320, 160),
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
	cache   sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *demoCardRenderer) MinSize() fyne.Size {
	return r.card.minSize
}

func (r *demoCardRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("demoCardRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *demoCardRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("demoCardRenderer", r.card.Size())
	size := r.card.Size()
	// Always rebuild if objects are empty (first render) or size changed
	if len(r.objects) == 0 || r.cache.ShouldLayout(size) {
		r.layoutWithSize(size)
		if size.Width > 0 && size.Height > 0 {
			r.cache.MarkLayout(size)
		}
	}
}

func (r *demoCardRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.card.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.objects = r.objects[:0]
	info := r.card.info

	// Colors
	cyanColor := color.RGBA{0, 212, 255, 255}
	var bgColor, borderColor, headerBgColor, textColor, subtitleColor, descColor, numberBgColor color.RGBA

	if info.Ready {
		borderColor = cyanColor
		bgColor = color.RGBA{0, 45, 90, 255}
		headerBgColor = color.RGBA{0, 55, 110, 255}
		textColor = color.RGBA{255, 255, 255, 255}
		subtitleColor = cyanColor
		descColor = color.RGBA{180, 200, 220, 255}
		numberBgColor = color.RGBA{0, 80, 160, 255}
	} else {
		borderColor = color.RGBA{80, 90, 100, 255}
		bgColor = color.RGBA{30, 40, 50, 200}
		headerBgColor = color.RGBA{35, 45, 55, 200}
		textColor = color.RGBA{120, 130, 140, 255}
		subtitleColor = color.RGBA{100, 110, 120, 255}
		descColor = color.RGBA{100, 110, 120, 255}
		numberBgColor = color.RGBA{50, 60, 70, 255}
	}

	borderWidth := float32(2)
	cornerRadius := float32(8)
	headerHeight := float32(50)

	// Outer border with rounded corners
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	border.CornerRadius = cornerRadius
	r.objects = append(r.objects, border)

	// Main background
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-borderWidth*2, size.Height-borderWidth*2))
	bg.Move(fyne.NewPos(borderWidth, borderWidth))
	bg.CornerRadius = cornerRadius - 1
	r.objects = append(r.objects, bg)

	// Header background
	headerBg := canvas.NewRectangle(headerBgColor)
	headerBg.Resize(fyne.NewSize(size.Width-borderWidth*2, headerHeight))
	headerBg.Move(fyne.NewPos(borderWidth, borderWidth))
	r.objects = append(r.objects, headerBg)

	// Number badge - compact circle
	badgeSize := float32(36)
	badgeX := float32(14)
	badgeY := float32(7) + borderWidth

	badgeBg := canvas.NewCircle(numberBgColor)
	badgeBg.Resize(fyne.NewSize(badgeSize, badgeSize))
	badgeBg.Move(fyne.NewPos(badgeX, badgeY))
	r.objects = append(r.objects, badgeBg)

	// Number text
	numText := canvas.NewText(string('0'+byte(info.Number)), textColor)
	numText.TextSize = 20
	numText.TextStyle = fyne.TextStyle{Bold: true}
	numText.Move(fyne.NewPos(badgeX+badgeSize/2-6, badgeY+badgeSize/2-12))
	r.objects = append(r.objects, numText)

	// Title
	title := canvas.NewText(info.Title, textColor)
	title.TextSize = 22
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(badgeX+badgeSize+12, 12))
	r.objects = append(r.objects, title)

	// Subtitle
	subtitle := canvas.NewText(info.Subtitle, subtitleColor)
	subtitle.TextSize = 13
	subtitle.Move(fyne.NewPos(badgeX+badgeSize+12, 36))
	r.objects = append(r.objects, subtitle)

	// Status indicator - WIP badge or green dot
	if info.Ready {
		if info.WIP {
			// Work In Progress badge
			wipWidth := float32(70)
			wipHeight := float32(18)
			wipBg := canvas.NewRectangle(color.RGBA{255, 165, 0, 255})
			wipBg.Resize(fyne.NewSize(wipWidth, wipHeight))
			wipBg.Move(fyne.NewPos(size.Width-wipWidth-8, 8))
			wipBg.CornerRadius = 3
			r.objects = append(r.objects, wipBg)

			wipText := canvas.NewText("WIP", color.RGBA{0, 0, 0, 255})
			wipText.TextSize = 11
			wipText.TextStyle = fyne.TextStyle{Bold: true}
			wipText.Move(fyne.NewPos(size.Width-wipWidth-8+24, 11))
			r.objects = append(r.objects, wipText)
		} else {
			// Green dot for ready
			dotSize := float32(10)
			statusDot := canvas.NewCircle(color.RGBA{100, 255, 150, 255})
			statusDot.Resize(fyne.NewSize(dotSize, dotSize))
			statusDot.Move(fyne.NewPos(size.Width-dotSize-10, 10))
			r.objects = append(r.objects, statusDot)
		}
	}

	// Description - wrapped text below header
	desc := info.Description
	descSize := float32(13)
	maxWidth := size.Width - 32
	lineY := headerHeight + borderWidth + 12
	lineHeight := descSize + 5

	words := splitWords(desc)
	line := ""
	for _, word := range words {
		testLine := line + word + " "
		testText := canvas.NewText(testLine, descColor)
		testText.TextSize = descSize
		if testText.MinSize().Width > maxWidth && line != "" {
			lineText := canvas.NewText(line, descColor)
			lineText.TextSize = descSize
			lineText.Move(fyne.NewPos(16, lineY))
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
		lineText.Move(fyne.NewPos(16, lineY))
		r.objects = append(r.objects, lineText)
	}

	// Click hint at bottom right
	hintText := canvas.NewText("Click to explore", color.RGBA{100, 130, 160, 200})
	hintText.TextSize = 11
	hintText.Move(fyne.NewPos(size.Width-100, size.Height-22))
	r.objects = append(r.objects, hintText)

	r.cache.MarkLayout(size)
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

	// Create header with branding
	titleText := canvas.NewText("FeCIM Lattice Tools", color.RGBA{255, 255, 255, 255})
	titleText.TextSize = 28
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	subtitleText := canvas.NewText("Ferroelectric Compute-in-Memory Educational Suite", color.RGBA{0, 212, 255, 255})
	subtitleText.TextSize = 16

	taglineText := canvas.NewText("\"Compute in memory where the same device does the memory and the computation.\" — Dr. external research group", color.RGBA{180, 200, 220, 200})
	taglineText.TextSize = 13
	taglineText.TextStyle = fyne.TextStyle{Italic: true}

	header := container.NewVBox(
		container.NewCenter(titleText),
		container.NewCenter(subtitleText),
		container.NewCenter(taglineText),
		widget.NewSeparator(),
	)

	// Grid layout - 3 columns, 2 rows
	grid := container.New(layout.NewGridLayoutWithRows(2),
		container.New(layout.NewGridLayoutWithColumns(3), cards[0], cards[1], cards[2]),
		container.New(layout.NewGridLayoutWithColumns(3), cards[3], cards[4], cards[5]),
	)

	// Key metrics in footer - split into two lines for readability
	line1 := canvas.NewText("30 Analog States  |  87% MNIST Accuracy  |  10M× Lower Energy vs NAND  |  1000× vs DRAM  |  TRL 4", color.RGBA{0, 212, 255, 230})
	line1.TextSize = 13
	line1.Alignment = fyne.TextAlignCenter

	line2 := canvas.NewText("1. Physics  →  2. Compute  →  3. Application  →  4. System  →  5. Business  →  6. Design", color.RGBA{150, 170, 190, 200})
	line2.TextSize = 12
	line2.Alignment = fyne.TextAlignCenter

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(line1),
		container.NewCenter(line2),
	)

	// Use border layout with header and footer
	return container.NewBorder(
		container.NewPadded(header),
		container.NewPadded(footer),
		nil, nil,
		container.NewPadded(grid),
	)
}
