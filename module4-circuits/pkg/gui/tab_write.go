package gui

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ============================================================================
// TAB 1: WRITE MODE
// ============================================================================

func (ca *CircuitsApp) createWriteTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**WRITE MODE**: Program ferroelectric cells to specific conductance levels using precise voltage pulses from the charge pump and DAC. The DAC converts digital levels (0-29) to analog voltages (2.0V-5.0V), which are applied as pulses to modify the FeFET polarization state.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Configuration section
	configSection := ca.createWriteConfigSection()

	// Cell selection section
	cellSection := ca.createWriteCellSection()

	// Data path visualization
	dataPathSection := ca.createWriteDataPathSection()

	// Programming pulse visualization
	pulseSection := ca.createWritePulseSection()

	// Array view
	arraySection := ca.createWriteArraySection()

	// Level-to-voltage mapping table
	mappingSection := ca.createWriteMappingSection()

	// Buttons
	programBtn := widget.NewButton("PROGRAM CELL", ca.onProgramCell)
	programBtn.Importance = widget.HighImportance

	randomBtn := widget.NewButton("PROGRAM RANDOM ARRAY", ca.onProgramRandomArray)

	ca.writeStatusLabel = widget.NewLabel("Ready to program")

	buttonBox := container.NewHBox(
		programBtn,
		randomBtn,
		layout.NewSpacer(),
		ca.writeStatusLabel,
	)

	// Layout
	leftPanel := container.NewVBox(
		widget.NewLabel("CONFIGURATION"),
		configSection,
		widget.NewSeparator(),
		widget.NewLabel("CELL SELECTION"),
		cellSection,
	)

	centerPanel := container.NewVBox(
		widget.NewLabel("DATA PATH VISUALIZATION"),
		dataPathSection,
		widget.NewSeparator(),
		widget.NewLabel("PROGRAMMING PULSE"),
		pulseSection,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("LEVEL-TO-VOLTAGE MAPPING"),
		mappingSection,
	)

	topRow := container.NewHBox(
		container.NewPadded(leftPanel),
		widget.NewSeparator(),
		container.NewPadded(centerPanel),
		widget.NewSeparator(),
		container.NewPadded(rightPanel),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator(), topRow),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		container.NewVBox(
			widget.NewLabel("ARRAY VIEW (click cell to select)"),
			arraySection,
		),
	)
}

func (ca *CircuitsApp) createWriteConfigSection() fyne.CanvasObject {
	// Array size selects
	sizeOptions := []string{"4", "8", "16", "32", "64"}
	rowSelect := widget.NewSelect(sizeOptions, func(s string) {
		var size int
		fmt.Sscanf(s, "%d", &size)
		ca.mu.Lock()
		ca.arrayRows = size
		ca.mu.Unlock()
		ca.initializeArray()
		ca.refreshWriteArray()
		ca.refreshCellSelectOptions()
	})
	rowSelect.SetSelected("8")

	colSelect := widget.NewSelect(sizeOptions, func(s string) {
		var size int
		fmt.Sscanf(s, "%d", &size)
		ca.mu.Lock()
		ca.arrayCols = size
		ca.mu.Unlock()
		ca.initializeArray()
		ca.refreshWriteArray()
		ca.refreshCellSelectOptions()
	})
	colSelect.SetSelected("8")

	// Quantization levels
	levelOptions := []string{"2", "4", "8", "16", "30", "32", "64", "128", "256"}
	levelSelect := widget.NewSelect(levelOptions, func(s string) {
		var levels int
		fmt.Sscanf(s, "%d", &levels)
		ca.mu.Lock()
		ca.quantLevels = levels
		ca.mu.Unlock()
	})
	levelSelect.SetSelected("30")
	quantHelp := widget.NewLabel("FeCIM uses 30 discrete analog states per cell (Dr. Tour, COSM 2025)")
	quantHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Voltage range entries
	vMinEntry := widget.NewEntry()
	vMinEntry.SetText("2.0")
	vMinEntry.SetPlaceHolder("Minimum write voltage (V) - must exceed coercive field")
	vMinEntry.OnChanged = func(s string) {
		var v float64
		fmt.Sscanf(s, "%f", &v)
		ca.mu.Lock()
		ca.vMin = v
		ca.mu.Unlock()
	}

	vMaxEntry := widget.NewEntry()
	vMaxEntry.SetText("5.0")
	vMaxEntry.SetPlaceHolder("Maximum write voltage (V) - for full polarization")
	vMaxEntry.OnChanged = func(s string) {
		var v float64
		fmt.Sscanf(s, "%f", &v)
		ca.mu.Lock()
		ca.vMax = v
		ca.mu.Unlock()
	}

	// Educational note about coercive field
	ecHelp := widget.NewLabel("Note: Ec (Coercive Field) ≈ 1.0-1.5 MV/cm for HZO. Write voltage must exceed Ec to switch polarization.")
	ecHelp.TextStyle = fyne.TextStyle{Italic: true}
	ecHelp.Wrapping = fyne.TextWrapWord

	// Pulse width entry
	pulseEntry := widget.NewEntry()
	pulseEntry.SetText("50")
	pulseEntry.SetPlaceHolder("Pulse duration in nanoseconds (typical FeFET: 10-100 ns)")
	pulseEntry.OnChanged = func(s string) {
		var pw float64
		fmt.Sscanf(s, "%f", &pw)
		ca.mu.Lock()
		ca.pulseWidth = pw
		ca.mu.Unlock()
		ca.refreshWritePulse()
	}

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Array Size:"),
			rowSelect,
			widget.NewLabel("x"),
			colSelect,
		),
		container.NewHBox(
			widget.NewLabel("Quantization:"),
			levelSelect,
			widget.NewLabel("levels"),
		),
		quantHelp,
		container.NewHBox(
			widget.NewLabel("Voltage Range:"),
			vMinEntry,
			widget.NewLabel("V min"),
			vMaxEntry,
			widget.NewLabel("V max"),
		),
		ecHelp,
		container.NewHBox(
			widget.NewLabel("Pulse Width:"),
			pulseEntry,
			widget.NewLabel("ns"),
		),
		widget.NewLabel("(Write pulse duration: shorter = faster but needs higher voltage)"),
	)

	return form
}

func (ca *CircuitsApp) createWriteCellSection() fyne.CanvasObject {
	// Row/col selects
	rowOptions := make([]string, ca.arrayRows)
	for i := range rowOptions {
		rowOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.writeRowSelect = widget.NewSelect(rowOptions, func(s string) {
		var row int
		fmt.Sscanf(s, "%d", &row)
		ca.mu.Lock()
		ca.selectedRow = row
		ca.mu.Unlock()
		ca.refreshWriteArray()
		ca.updateWriteDataPath()
	})
	ca.writeRowSelect.SetSelected("3")

	colOptions := make([]string, ca.arrayCols)
	for i := range colOptions {
		colOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.writeColSelect = widget.NewSelect(colOptions, func(s string) {
		var col int
		fmt.Sscanf(s, "%d", &col)
		ca.mu.Lock()
		ca.selectedCol = col
		ca.mu.Unlock()
		ca.refreshWriteArray()
		ca.updateWriteDataPath()
	})
	ca.writeColSelect.SetSelected("5")

	// Target level slider
	ca.writeLevelLabel = widget.NewLabel("Target Level: 15 / 30 (discrete conductance state)")
	ca.writeLevelSlider = widget.NewSlider(0, float64(ca.quantLevels-1))
	ca.writeLevelSlider.Value = 15
	ca.writeLevelSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.targetLevel = int(v)
		ca.mu.Unlock()
		ca.writeLevelLabel.SetText(fmt.Sprintf("Target Level: %d / %d (discrete conductance state)", ca.targetLevel, ca.quantLevels))
		ca.updateWriteDataPath()
		ca.refreshWritePulse()
		// Update mapping table to show current target level
		ca.writeMappingLabel.SetText(ca.getMappingText())
	}

	levelHelp := widget.NewLabel("Each level represents a stable polarization state (~4.9 bits/cell)")
	levelHelp.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Target Cell: Row"),
			ca.writeRowSelect,
			widget.NewLabel("Col"),
			ca.writeColSelect,
		),
		ca.writeLevelLabel,
		ca.writeLevelSlider,
		levelHelp,
	)
}

func (ca *CircuitsApp) createWriteDataPathSection() fyne.CanvasObject {
	// Create visual boxes for the data path with stored label references
	ca.writeDigitalLabel = widget.NewLabel("Level:15\n01111")
	ca.writeDACLabel = widget.NewLabel("3.55V")
	ca.writeFeFETLabel = widget.NewLabel("[3,5]\n52.2µS")

	digitalBox := ca.createLabeledBoxWithLabel("DIGITAL", ca.writeDigitalLabel, colorPrimary)
	dacBox := ca.createLabeledBoxWithLabel("DAC", ca.writeDACLabel, colorDAC)
	fefetBox := ca.createLabeledBoxWithLabel("FeFET", ca.writeFeFETLabel, colorArrayCell)

	arrow1 := widget.NewLabel("→")
	arrow2 := widget.NewLabel("→")

	ca.writeDataPath = container.NewHBox(
		digitalBox,
		arrow1,
		dacBox,
		arrow2,
		fefetBox,
	)

	ca.updateWriteDataPath()

	helperText := widget.NewLabel("Data path: Digital level → DAC voltage conversion → FeFET polarization")
	helperText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(ca.writeDataPath, helperText)
}

func (ca *CircuitsApp) createLabeledBox(title, value string, bgColor color.Color) *fyne.Container {
	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	titleLbl.Alignment = fyne.TextAlignCenter

	valueLbl := widget.NewLabel(value)
	valueLbl.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(100, 60))
	bg.CornerRadius = 5

	content := container.NewVBox(titleLbl, valueLbl)

	return container.NewStack(bg, container.NewCenter(content))
}

func (ca *CircuitsApp) createLabeledBoxWithLabel(title string, valueLbl *widget.Label, bgColor color.Color) *fyne.Container {
	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	titleLbl.Alignment = fyne.TextAlignCenter

	valueLbl.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(100, 60))
	bg.CornerRadius = 5

	content := container.NewVBox(titleLbl, valueLbl)

	return container.NewStack(bg, container.NewCenter(content))
}

func (ca *CircuitsApp) updateWriteDataPath() {
	ca.mu.RLock()
	level := ca.targetLevel
	row := ca.selectedRow
	col := ca.selectedCol
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Calculate voltage
	voltage := vMin + float64(level)/float64(levels-1)*(vMax-vMin)

	// Calculate conductance (1-100 µS range)
	conductance := 1.0 + float64(level)/float64(levels-1)*99.0

	// Binary representation
	binary := fmt.Sprintf("%05b", level)

	// Update the data path display using direct label references
	if ca.writeDigitalLabel != nil {
		ca.writeDigitalLabel.SetText(fmt.Sprintf("Level:%d\n%s", level, binary))
	}
	if ca.writeDACLabel != nil {
		ca.writeDACLabel.SetText(fmt.Sprintf("%.2fV", voltage))
	}
	if ca.writeFeFETLabel != nil {
		ca.writeFeFETLabel.SetText(fmt.Sprintf("[%d,%d]\n%.1fµS", row, col, conductance))
	}
}

func (ca *CircuitsApp) createWritePulseSection() fyne.CanvasObject {
	ca.writePulseCanvas = canvas.NewRaster(ca.drawWritePulse)
	ca.writePulseCanvas.SetMinSize(fyne.NewSize(400, 150))
	return ca.writePulseCanvas
}

func (ca *CircuitsApp) drawWritePulse(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	level := ca.targetLevel
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	_ = ca.pulseWidth // Used for display
	ca.mu.RUnlock()

	voltage := vMin + float64(level)/float64(levels-1)*(vMax-vMin)

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw axes
	marginLeft := 50
	marginBottom := 30
	marginTop := 20
	marginRight := 30

	plotW := w - marginLeft - marginRight
	plotH := h - marginTop - marginBottom
	axisColor := color.RGBA{200, 200, 200, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	fillColor := color.RGBA{0, 100, 150, 200}
	threshColor := color.RGBA{255, 200, 0, 255}

	// Y-axis (voltage)
	for y := marginTop; y < h-marginBottom; y++ {
		img.Set(marginLeft, y, axisColor)
	}

	// X-axis (time)
	for x := marginLeft; x < w-marginRight; x++ {
		img.Set(x, h-marginBottom, axisColor)
	}

	// Pulse positions
	pulseStart := marginLeft + plotW*10/100
	pulseEnd := marginLeft + plotW*70/100
	riseEnd := pulseStart + plotW*2/100
	fallStart := pulseEnd - plotW*2/100

	// Y positions
	y0V := h - marginBottom
	yVoltage := marginTop + int(float64(plotH)*(1.0-(voltage-0)/(vMax+0.5)))
	yThreshold := marginTop + int(float64(plotH)*(1.0-(vMin-0)/(vMax+0.5)))

	// Draw threshold line (dashed)
	if yThreshold >= marginTop && yThreshold < h-marginBottom {
		for x := marginLeft; x < w-marginRight; x += 6 {
			img.Set(x, yThreshold, threshColor)
		}
	}

	// Draw pulse
	for x := marginLeft; x < w-marginRight; x++ {
		var y int
		if x < pulseStart {
			y = y0V
		} else if x < riseEnd {
			t := float64(x-pulseStart) / float64(riseEnd-pulseStart)
			y = y0V + int(float64(yVoltage-y0V)*t)
		} else if x < fallStart {
			y = yVoltage
		} else if x < pulseEnd {
			t := float64(x-fallStart) / float64(pulseEnd-fallStart)
			y = yVoltage + int(float64(y0V-yVoltage)*t)
		} else {
			y = y0V
		}

		// Draw thick line
		for dy := -2; dy <= 2; dy++ {
			py := y + dy
			if py >= marginTop && py < h-marginBottom {
				img.Set(x, py, cyanColor)
			}
		}

		// Fill pulse area
		if x >= riseEnd && x < fallStart {
			for py := yVoltage; py < y0V; py++ {
				img.Set(x, py, fillColor)
			}
		}
	}

	// Axis labels
	// Y-axis label
	drawScaledText(img, "Voltage (V)", marginLeft-40, marginTop-8, 1, axisColor)

	// X-axis label
	drawScaledText(img, "Time (ns)", w-marginRight-50, h-marginBottom+15, 1, axisColor)

	// Values
	drawSimpleText(img, fmt.Sprintf("%.1fV", vMax), 5, marginTop+5, axisColor)
	drawSimpleText(img, fmt.Sprintf("%.1fV", vMin), 5, yThreshold+5, axisColor)
	drawSimpleText(img, "0V", 25, y0V-5, axisColor)

	return img
}

func (ca *CircuitsApp) refreshWritePulse() {
	if ca.writePulseCanvas != nil {
		fyne.Do(func() {
			ca.writePulseCanvas.Refresh()
		})
	}
}

func (ca *CircuitsApp) createWriteArraySection() fyne.CanvasObject {
	ca.writeArrayCanvas = canvas.NewRaster(ca.drawWriteArray)
	ca.writeArrayCanvas.SetMinSize(fyne.NewSize(500, 350))
	return ca.writeArrayCanvas
}

func (ca *CircuitsApp) drawWriteArray(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	selectedRow := ca.selectedRow
	selectedCol := ca.selectedCol
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if weights == nil || len(weights) == 0 {
		return img
	}

	// Calculate cell size (use square cells like crossbar module)
	margin := 40
	cellW := (w - 2*margin) / cols
	cellH := (h - 2*margin) / rows
	cellSize := cellW
	if cellH < cellSize {
		cellSize = cellH
	}
	if cellSize > 40 {
		cellSize = 40
	}
	if cellSize < 8 {
		cellSize = 8
	}

	// Center the grid in the available space
	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := (w - gridW) / 2
	offsetY := (h - gridH) / 2

	// Draw cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := offsetX + c*cellSize
			y0 := offsetY + r*cellSize

			level := weights[r][c]
			intensity := float64(level) / float64(levels-1)

			// Check if this is the selected cell
			isSelected := r == selectedRow && c == selectedCol

			// Color based on level (blue to red)
			var cr, cg, cb uint8
			if isSelected {
				cr, cg, cb = 255, 200, 50 // Bright yellow for selection
			} else {
				cr = uint8(intensity * 200)
				cg = uint8(50 + (1-intensity)*100)
				cb = uint8((1 - intensity) * 200)
			}

			cellColor := color.RGBA{cr, cg, cb, 255}
			drawRect(img, x0+2, y0+2, cellSize-4, cellSize-4, cellColor)

			// Draw a thick white border around the selected cell for better visibility
			if isSelected {
				borderColor := color.RGBA{255, 255, 255, 255}
				borderWidth := 3
				// Top border
				drawRect(img, x0, y0, cellSize, borderWidth, borderColor)
				// Bottom border
				drawRect(img, x0, y0+cellSize-borderWidth, cellSize, borderWidth, borderColor)
				// Left border
				drawRect(img, x0, y0, borderWidth, cellSize, borderColor)
				// Right border
				drawRect(img, x0+cellSize-borderWidth, y0, borderWidth, cellSize, borderColor)
			}
		}
	}

	return img
}

func (ca *CircuitsApp) refreshWriteArray() {
	if ca.writeArrayCanvas != nil {
		fyne.Do(func() {
			ca.writeArrayCanvas.Refresh()
		})
	}
}

func (ca *CircuitsApp) createWriteMappingSection() fyne.CanvasObject {
	ca.writeMappingLabel = widget.NewLabel(ca.getMappingText())
	ca.writeMappingLabel.TextStyle = fyne.TextStyle{Monospace: true}
	return container.NewVScroll(ca.writeMappingLabel)
}

func (ca *CircuitsApp) getMappingText() string {
	ca.mu.RLock()
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	target := ca.targetLevel
	ca.mu.RUnlock()

	text := "LEVEL-TO-VOLTAGE MAPPING TABLE\n"
	text += "Shows how digital levels (0-29) map to programming voltages\n"
	text += "and resulting FeFET conductance states.\n"
	text += "================================================================\n\n"
	text += "Level   Voltage   Conductance   Resistance\n"
	text += "-----   -------   -----------   ----------\n"

	// Show more levels for better visibility (8 levels)
	sampleLevels := []int{0, 4, 8, 12, 15, 20, 25, levels - 1}

	// Always include target if not already present
	hasTarget := false
	for _, l := range sampleLevels {
		if l == target {
			hasTarget = true
			break
		}
	}
	if !hasTarget {
		// Insert target in sorted position
		newLevels := make([]int, 0, len(sampleLevels)+1)
		inserted := false
		for _, l := range sampleLevels {
			if !inserted && target < l {
				newLevels = append(newLevels, target)
				inserted = true
			}
			newLevels = append(newLevels, l)
		}
		if !inserted {
			newLevels = append(newLevels, target)
		}
		sampleLevels = newLevels
	}

	seen := make(map[int]bool)

	for _, l := range sampleLevels {
		if seen[l] || l >= levels {
			continue
		}
		seen[l] = true

		voltage := vMin + float64(l)/float64(levels-1)*(vMax-vMin)
		conductance := 1.0 + float64(l)/float64(levels-1)*99.0 // 1-100 µS
		resistance := 1000.0 / conductance                     // kO

		marker := "  "
		if l == target {
			marker = "> "
		}
		text += fmt.Sprintf("%s%2d      %5.2fV      %5.1f µS      %6.1f kΩ\n",
			marker, l, voltage, conductance, resistance)
	}

	text += "\n================================\n"
	text += fmt.Sprintf("TARGET: Level %d = %.2fV\n", target,
		vMin+float64(target)/float64(levels-1)*(vMax-vMin))
	text += fmt.Sprintf("Range: %.1fV to %.1fV (%d levels)\n", vMin, vMax, levels)

	return text
}

func (ca *CircuitsApp) onProgramCell() {
	ca.mu.Lock()
	row := ca.selectedRow
	col := ca.selectedCol
	level := ca.targetLevel

	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		ca.arrayWeights[row][col] = level
	}
	ca.mu.Unlock()

	ca.refreshWriteArray()
	ca.writeStatusLabel.SetText(fmt.Sprintf("Programmed cell [%d,%d] to level %d", row, col, level))
}

func (ca *CircuitsApp) onProgramRandomArray() {
	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = rand.Intn(ca.quantLevels)
		}
	}
	ca.mu.Unlock()

	ca.refreshWriteArray()
	ca.writeStatusLabel.SetText("Programmed array with random values")
}

// refreshCellSelectOptions updates row/col dropdown options when array size changes
func (ca *CircuitsApp) refreshCellSelectOptions() {
	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	selectedRow := ca.selectedRow
	selectedCol := ca.selectedCol
	ca.mu.RUnlock()

	// Update row select options
	if ca.writeRowSelect != nil {
		rowOptions := make([]string, rows)
		for i := range rowOptions {
			rowOptions[i] = fmt.Sprintf("%d", i)
		}
		ca.writeRowSelect.Options = rowOptions

		// Reset selection if out of bounds
		if selectedRow >= rows {
			ca.mu.Lock()
			ca.selectedRow = 0
			ca.mu.Unlock()
			ca.writeRowSelect.SetSelected("0")
		} else {
			ca.writeRowSelect.SetSelected(fmt.Sprintf("%d", selectedRow))
		}
	}

	// Update col select options
	if ca.writeColSelect != nil {
		colOptions := make([]string, cols)
		for i := range colOptions {
			colOptions[i] = fmt.Sprintf("%d", i)
		}
		ca.writeColSelect.Options = colOptions

		// Reset selection if out of bounds
		if selectedCol >= cols {
			ca.mu.Lock()
			ca.selectedCol = 0
			ca.mu.Unlock()
			ca.writeColSelect.SetSelected("0")
		} else {
			ca.writeColSelect.SetSelected(fmt.Sprintf("%d", selectedCol))
		}
	}

	// Refresh the array display
	ca.refreshWriteArray()
}
