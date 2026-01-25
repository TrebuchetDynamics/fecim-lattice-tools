package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ============================================================================
// TAB 2: READ MODE
// ============================================================================

func (ca *CircuitsApp) createReadTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**READ MODE**: Sense the conductance state of ferroelectric cells using low voltage (0.5V) to avoid disturbing the stored data. The TIA (transimpedance amplifier) converts the cell current to voltage, which is then digitized by the ADC for output.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Configuration section
	configSection := ca.createReadConfigSection()

	// Cell selection section
	cellSection := ca.createReadCellSection()

	// Data path visualization
	dataPathSection := ca.createReadDataPathSection()

	// Voltage zones visualization
	zoneSection := ca.createReadZoneSection()

	// Results section
	resultsSection := ca.createReadResultsSection()

	// Buttons
	readBtn := widget.NewButton("READ CELL", ca.onReadCell)
	readBtn.Importance = widget.HighImportance

	readAllBtn := widget.NewButton("READ ALL CELLS", ca.onReadAllCells)

	verifyBtn := widget.NewButton("VERIFY ARRAY", ca.onVerifyArray)

	ca.readStatusLabel = widget.NewLabel("Ready to read")

	buttonBox := container.NewHBox(
		readBtn,
		readAllBtn,
		verifyBtn,
		layout.NewSpacer(),
		ca.readStatusLabel,
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
		widget.NewLabel("VOLTAGE ZONES"),
		zoneSection,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("READ RESULTS"),
		resultsSection,
	)

	mainContent := container.NewHBox(
		container.NewPadded(leftPanel),
		widget.NewSeparator(),
		container.NewPadded(centerPanel),
		widget.NewSeparator(),
		container.NewPadded(rightPanel),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator(), mainContent),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		nil,
	)
}

func (ca *CircuitsApp) createReadConfigSection() fyne.CanvasObject {
	// Read voltage slider
	ca.readVoltageLabel = widget.NewLabel("Read Voltage: 0.5 V (non-destructive sensing)")
	ca.readVoltageSlider = widget.NewSlider(0.1, 1.5)
	ca.readVoltageSlider.Value = 0.5
	ca.readVoltageSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.readVoltage = v
		ca.mu.Unlock()
		ca.readVoltageLabel.SetText(fmt.Sprintf("Read Voltage: %.2f V (non-destructive sensing)", v))
		ca.refreshReadZone()
	}

	warningLabel := widget.NewLabel("SAFE ZONE: 0.1V - 1.0V")
	warningLabel.TextStyle = fyne.TextStyle{Bold: true}

	dangerLabel := widget.NewLabel("DANGER: > 2.0V (will modify cell!)")

	// Educational note about safe voltage
	safeHelp := widget.NewLabel("Read voltage (0.5V) is below Ec (~1.5V), ensuring non-destructive sensing.")
	safeHelp.TextStyle = fyne.TextStyle{Italic: true}

	// ADC resolution select with helper text
	adcOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	adcSelect := widget.NewSelect(adcOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		ca.mu.Lock()
		ca.adcBits = bits
		ca.mu.Unlock()
		ca.refreshReadCalculation()
	})
	adcSelect.SetSelected("8")
	adcHelp := widget.NewLabel("(bits of precision for digitizing analog current)")
	adcHelp.TextStyle = fyne.TextStyle{Italic: true}

	// TIA gain select with helper text
	tiaOptions := []string{"1", "10", "100"}
	tiaSelect := widget.NewSelect(tiaOptions, func(s string) {
		var gain float64
		fmt.Sscanf(s, "%f", &gain)
		ca.mu.Lock()
		ca.tiaGain = gain
		ca.mu.Unlock()
		ca.refreshReadCalculation()
	})
	tiaSelect.SetSelected("10")
	tiaHelp := widget.NewLabel("(Transimpedance: converts cell current to voltage)")
	tiaHelp.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		ca.readVoltageLabel,
		ca.readVoltageSlider,
		warningLabel,
		dangerLabel,
		safeHelp,
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("ADC Resolution:"),
			adcSelect,
			widget.NewLabel("bits"),
		),
		adcHelp,
		container.NewHBox(
			widget.NewLabel("TIA Gain:"),
			tiaSelect,
			widget.NewLabel("kOhm"),
		),
		tiaHelp,
	)
}

func (ca *CircuitsApp) createReadCellSection() fyne.CanvasObject {
	rowOptions := make([]string, ca.arrayRows)
	for i := range rowOptions {
		rowOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.readRowSelect = widget.NewSelect(rowOptions, func(s string) {
		var row int
		fmt.Sscanf(s, "%d", &row)
		ca.mu.Lock()
		ca.selectedRow = row
		ca.mu.Unlock()
	})
	ca.readRowSelect.SetSelected("3")

	colOptions := make([]string, ca.arrayCols)
	for i := range colOptions {
		colOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.readColSelect = widget.NewSelect(colOptions, func(s string) {
		var col int
		fmt.Sscanf(s, "%d", &col)
		ca.mu.Lock()
		ca.selectedCol = col
		ca.mu.Unlock()
	})
	ca.readColSelect.SetSelected("5")

	storedLabel := widget.NewLabel("Stored Level: -- (from previous WRITE)")

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Target Cell: Row"),
			ca.readRowSelect,
			widget.NewLabel("Col"),
			ca.readColSelect,
		),
		storedLabel,
	)
}

func (ca *CircuitsApp) createReadDataPathSection() fyne.CanvasObject {
	fefetBox := ca.createLabeledBox("FeFET", "Cell --,--", colorArrayCell)
	tiaBox := ca.createLabeledBox("TIA", "(I→V)", colorTIA)
	adcBox := ca.createLabeledBox("ADC", "8-bit", colorADC)
	digitalBox := ca.createLabeledBox("DIGITAL", "Output", colorPrimary)

	arrow1 := widget.NewLabel("→")
	arrow2 := widget.NewLabel("→")
	arrow3 := widget.NewLabel("→")

	ca.readDataPath = container.NewHBox(
		fefetBox,
		arrow1,
		tiaBox,
		arrow2,
		adcBox,
		arrow3,
		digitalBox,
	)

	helperText := widget.NewLabel("Data path: FeFET current → TIA voltage conversion → ADC digitization → Level")
	helperText.TextStyle = fyne.TextStyle{Italic: true}

	// Calculation box
	ca.readCalcLabel = widget.NewLabel(
		"I = G × V = -- µS × -- V = -- µA\n" +
			"V_tia = I × R = -- µA × -- kΩ = -- mV\n" +
			"ADC = (-- mV / 1000 mV) × 255 = --\n" +
			"Level = round(-- / 255 × (L-1)) = --",
	)
	// Use monospace for better alignment
	ca.readCalcLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewVBox(
		ca.readDataPath,
		helperText,
		widget.NewSeparator(),
		widget.NewLabel("Calculation:"),
		ca.readCalcLabel,
	)
}

func (ca *CircuitsApp) createReadZoneSection() fyne.CanvasObject {
	ca.readZoneCanvas = canvas.NewRaster(ca.drawReadZone)
	ca.readZoneCanvas.SetMinSize(fyne.NewSize(300, 200))
	return ca.readZoneCanvas
}

func (ca *CircuitsApp) drawReadZone(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	readV := ca.readVoltage
	ca.mu.RUnlock()

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 50
	marginRight := 20
	marginTop := 15
	marginBottom := 15
	plotH := h - marginTop - marginBottom
	plotW := w - marginLeft - marginRight

	writeZoneColor := color.RGBA{200, 50, 50, 180}
	readZoneColor := color.RGBA{50, 150, 50, 180}
	threshColor := color.RGBA{255, 200, 0, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	labelColor := color.RGBA{255, 255, 255, 255}
	axisColor := color.RGBA{150, 150, 150, 255}

	// Y-axis voltage scale (5V at top, 0V at bottom)
	maxVoltage := 5.0

	// Helper to convert voltage to Y position
	voltageToY := func(v float64) int {
		return marginTop + int((maxVoltage-v)/maxVoltage*float64(plotH))
	}

	// Write zone (> 2V) - red danger zone
	writeZoneTop := voltageToY(maxVoltage)
	writeZoneBottom := voltageToY(2.0)
	drawRect(img, marginLeft, writeZoneTop, plotW, writeZoneBottom-writeZoneTop, writeZoneColor)

	// Transition zone (1V - 2V) - neutral
	// (no special coloring, just background)

	// Read zone (< 1V) - green safe zone
	readZoneTop := voltageToY(1.0)
	readZoneBottom := voltageToY(0.0)
	drawRect(img, marginLeft, readZoneTop, plotW, readZoneBottom-readZoneTop, readZoneColor)

	// Threshold line (2V)
	thresholdY := voltageToY(2.0)
	for x := marginLeft; x < marginLeft+plotW; x++ {
		for dy := -1; dy <= 1; dy++ {
			if thresholdY+dy >= marginTop && thresholdY+dy < h-marginBottom {
				img.Set(x, thresholdY+dy, threshColor)
			}
		}
	}

	// Zone labels (right side of zones)
	drawSimpleText(img, "WRITE ZONE", marginLeft+10, writeZoneTop+15, labelColor)
	drawSimpleText(img, "> 2.0V DANGER", marginLeft+10, writeZoneTop+28, color.RGBA{255, 150, 150, 255})

	drawSimpleText(img, "2.0V THRESHOLD", marginLeft+plotW-110, thresholdY-10, threshColor)

	drawSimpleText(img, "READ ZONE", marginLeft+10, readZoneTop+15, labelColor)
	drawSimpleText(img, "< 1.0V SAFE", marginLeft+10, readZoneTop+28, color.RGBA{150, 255, 150, 255})

	// Y-axis with voltage scale
	for y := marginTop; y <= h-marginBottom; y++ {
		img.Set(marginLeft-1, y, axisColor)
	}

	// Voltage labels on Y-axis
	voltageMarkers := []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0}
	for _, v := range voltageMarkers {
		y := voltageToY(v)
		// Tick mark
		for dx := 0; dx < 5; dx++ {
			img.Set(marginLeft-5+dx, y, axisColor)
		}
		// Voltage label
		label := fmt.Sprintf("%.1fV", v)
		drawSimpleText(img, label, 5, y-3, axisColor)
	}

	// Current read voltage indicator line
	readY := voltageToY(readV)
	for x := marginLeft; x < marginLeft+plotW; x++ {
		for dy := -2; dy <= 2; dy++ {
			y := readY + dy
			if y >= marginTop && y < h-marginBottom {
				img.Set(x, y, cyanColor)
			}
		}
	}

	// Current voltage value label next to indicator
	voltageLabel := fmt.Sprintf("%.2fV", readV)
	drawSimpleText(img, voltageLabel, marginLeft+plotW-50, readY-10, cyanColor)

	// Arrow indicator on left side
	for i := 0; i < 8; i++ {
		img.Set(marginLeft-8+i, readY, cyanColor)
		if i < 4 {
			img.Set(marginLeft-8+i, readY-i, cyanColor)
			img.Set(marginLeft-8+i, readY+i, cyanColor)
		}
	}

	return img
}

func (ca *CircuitsApp) refreshReadZone() {
	if ca.readZoneCanvas != nil {
		fyne.Do(func() {
			ca.readZoneCanvas.Refresh()
		})
	}
}

func (ca *CircuitsApp) createReadResultsSection() fyne.CanvasObject {
	ca.readResultsLabel = widget.NewLabel(
		"Cell [--,--] Read Results\n" +
			"─────────────────────────\n" +
			"Programmed Level:    --\n" +
			"Read Current:        -- µA\n" +
			"TIA Voltage:         -- mV\n" +
			"ADC Raw:             -- / 255\n" +
			"Decoded Level:       --\n" +
			"Match:               --",
	)

	return ca.readResultsLabel
}

func (ca *CircuitsApp) onReadCell() {
	ca.mu.RLock()
	row := ca.selectedRow
	col := ca.selectedCol
	readV := ca.readVoltage
	tiaGain := ca.tiaGain
	levels := ca.quantLevels
	var storedLevel int
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		storedLevel = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	// Calculate conductance from stored level
	conductance := 1.0 + float64(storedLevel)/float64(levels-1)*99.0 // 1-100 µS

	// Calculate current: I = G × V
	current := conductance * readV // µA

	// TIA: V = I × R
	tiaVoltage := current * tiaGain // mV

	// ADC conversion (8-bit, 0-1V range)
	adcRaw := int(tiaVoltage / 1000.0 * 255.0)
	if adcRaw > 255 {
		adcRaw = 255
	}
	if adcRaw < 0 {
		adcRaw = 0
	}

	// Decode back to level
	decodedLevel := int(math.Round(float64(adcRaw) / 255.0 * float64(levels-1)))

	match := "CORRECT"
	if decodedLevel != storedLevel {
		match = fmt.Sprintf("MISMATCH (expected %d)", storedLevel)
	}

	ca.readResultsLabel.SetText(fmt.Sprintf(
		"Cell [%d,%d] Read Results\n"+
			"─────────────────────────\n"+
			"Programmed Level:    %d\n"+
			"Read Current:        %.1f µA\n"+
			"TIA Voltage:         %.0f mV\n"+
			"ADC Raw:             %d / 255\n"+
			"Decoded Level:       %d\n"+
			"Match:               %s",
		row, col, storedLevel, current, tiaVoltage, adcRaw, decodedLevel, match,
	))

	ca.readStatusLabel.SetText(fmt.Sprintf("Read cell [%d,%d]: Level %d", row, col, decodedLevel))

	// Update formula calculation display
	ca.readCalcLabel.SetText(fmt.Sprintf(
		"I     = G × V     = %.1f µS × %.2f V = %.1f µA\n"+
			"V_tia = I × R     = %.1f µA × %.0f kΩ = %.0f mV\n"+
			"ADC   = V_tia/Vref = %.0f / 1000 × 255 = %d\n"+
			"Level = ADC/Max   = %d / 255 × %d  = %d",
		conductance, readV, current,
		current, tiaGain, tiaVoltage,
		tiaVoltage, adcRaw,
		adcRaw, levels-1, decodedLevel,
	))
}

func (ca *CircuitsApp) onReadAllCells() {
	ca.readStatusLabel.SetText("Reading all cells...")

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	totalCells := rows * cols
	ca.mu.RUnlock()

	// Simulate reading with progress
	go func() {
		time.Sleep(100 * time.Millisecond) // Simulate work
		fyne.Do(func() {
			ca.readStatusLabel.SetText(fmt.Sprintf("Read complete: %d cells verified", totalCells))
		})
	}()
}

func (ca *CircuitsApp) onVerifyArray() {
	ca.readStatusLabel.SetText("Verifying array...")

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	levels := ca.quantLevels
	readV := ca.readVoltage
	tiaGain := ca.tiaGain
	ca.mu.RUnlock()

	// Perform verification in background
	go func() {
		errors := 0
		for r := 0; r < rows && r < len(weights); r++ {
			for c := 0; c < cols && c < len(weights[r]); c++ {
				storedLevel := weights[r][c]
				// Simulate read and decode
				conductance := 1.0 + float64(storedLevel)/float64(levels-1)*99.0
				current := conductance * readV
				tiaVoltage := current * tiaGain
				adcRaw := int(tiaVoltage / 1000.0 * 255.0)
				if adcRaw > 255 {
					adcRaw = 255
				}
				decodedLevel := int(math.Round(float64(adcRaw) / 255.0 * float64(levels-1)))
				if decodedLevel != storedLevel {
					errors++
				}
			}
		}

		totalCells := rows * cols
		fyne.Do(func() {
			if errors == 0 {
				ca.readStatusLabel.SetText(fmt.Sprintf("Verification complete: %d/%d cells OK", totalCells, totalCells))
			} else {
				ca.readStatusLabel.SetText(fmt.Sprintf("Verification complete: %d errors in %d cells", errors, totalCells))
			}
		})
	}()
}

// refreshReadCalculation updates the calculation display when TIA/ADC settings change
func (ca *CircuitsApp) refreshReadCalculation() {
	ca.mu.RLock()
	row := ca.selectedRow
	col := ca.selectedCol
	readV := ca.readVoltage
	tiaGain := ca.tiaGain
	levels := ca.quantLevels
	var storedLevel int
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		storedLevel = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	// Calculate conductance from stored level
	conductance := 1.0 + float64(storedLevel)/float64(levels-1)*99.0 // 1-100 µS

	// Calculate current: I = G × V
	current := conductance * readV // µA

	// TIA: V = I × R
	tiaVoltage := current * tiaGain // mV

	// ADC conversion (8-bit, 0-1V range)
	adcRaw := int(tiaVoltage / 1000.0 * 255.0)
	if adcRaw > 255 {
		adcRaw = 255
	}
	if adcRaw < 0 {
		adcRaw = 0
	}

	// Decode back to level
	decodedLevel := int(math.Round(float64(adcRaw) / 255.0 * float64(levels-1)))

	// Update formula calculation display
	if ca.readCalcLabel != nil {
		fyne.Do(func() {
			ca.readCalcLabel.SetText(fmt.Sprintf(
				"I     = G × V     = %.1f µS × %.2f V = %.1f µA\n"+
					"V_tia = I × R     = %.1f µA × %.0f kΩ = %.0f mV\n"+
					"ADC   = V_tia/Vref = %.0f / 1000 × 255 = %d\n"+
					"Level = ADC/Max   = %d / 255 × %d  = %d",
				conductance, readV, current,
				current, tiaGain, tiaVoltage,
				tiaVoltage, adcRaw,
				adcRaw, levels-1, decodedLevel,
			))
		})
	}
}
