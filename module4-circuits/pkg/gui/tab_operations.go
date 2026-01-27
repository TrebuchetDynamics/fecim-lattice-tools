// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified OPERATIONS view that consolidates WRITE, READ, and COMPUTE modes.
package gui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// OperationMode represents the current operation mode in the unified view
type OperationMode int

const (
	ModeWrite OperationMode = iota
	ModeRead
	ModeCompute
)

// ============================================================================
// TAPPABLE ARRAY CANVAS WIDGET
// ============================================================================

// TappableArrayCanvas is a canvas.Raster that responds to taps
type TappableArrayCanvas struct {
	widget.BaseWidget
	raster *canvas.Raster
	onTap  func(row, col int)
	ca     *CircuitsApp
}

func NewTappableArrayCanvas(ca *CircuitsApp, drawFunc func(w, h int) image.Image, onTap func(row, col int)) *TappableArrayCanvas {
	t := &TappableArrayCanvas{
		raster: canvas.NewRaster(drawFunc),
		onTap:  onTap,
		ca:     ca,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *TappableArrayCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.raster)
}

func (t *TappableArrayCanvas) SetMinSize(size fyne.Size) {
	t.raster.SetMinSize(size)
}

func (t *TappableArrayCanvas) Refresh() {
	t.raster.Refresh()
}

func (t *TappableArrayCanvas) Tapped(e *fyne.PointEvent) {
	// Get current widget/raster size
	size := t.raster.Size()

	t.ca.mu.RLock()
	rows := t.ca.arrayRows
	cols := t.ca.arrayCols
	t.ca.mu.RUnlock()

	// Recalculate cell geometry using same logic as drawSharedArray
	w := int(size.Width)
	h := int(size.Height)

	// Use same asymmetric margins as drawSharedArray
	topMargin := 70
	rightMargin := 70
	bottomMargin := 30
	leftMargin := 30

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows

	// Cell size calculations (COMPLETE - matching drawSharedArray exactly)
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
	if cellSize <= 0 {
		return
	}

	// Calculate grid size and offset
	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	// Convert click position to cell coordinates
	col := (int(e.Position.X) - offsetX) / cellSize
	row := (int(e.Position.Y) - offsetY) / cellSize

	// Bounds check
	if row >= 0 && row < rows && col >= 0 && col < cols {
		t.onTap(row, col)
	}
}

func (t *TappableArrayCanvas) TappedSecondary(*fyne.PointEvent) {}

// Cursor returns a pointer cursor to indicate the array is clickable
func (t *TappableArrayCanvas) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// ============================================================================
// UNIFIED OPERATIONS VIEW
// ============================================================================

// createOperationsView creates the unified OPERATIONS view with mode selector
func (ca *CircuitsApp) createOperationsView() fyne.CanvasObject {
	// Initialize operations-specific fields
	ca.currentMode = ModeWrite
	ca.operationsStatusLabel = widget.NewLabel("Ready")

	// 1. Create mode selector (segmented buttons using radio group)
	modeSelector := ca.createModeSelector()

	// 2. Create shared array section (left panel, always visible)
	arraySection := ca.createSharedArraySection()

	// 3. Create mode-specific panels (stacked, visibility toggled)
	ca.createWriteModePanel()
	ca.createReadModePanel()
	ca.createComputeModePanel()

	// Stack all mode panels (visibility toggled based on selection)
	modeStack := container.NewStack(
		ca.writeConfigPanel,
		ca.readConfigPanel,
		ca.computeConfigPanel,
	)

	// Initialize visibility: show write, hide others
	ca.writeConfigPanel.Show()
	ca.readConfigPanel.Hide()
	ca.computeConfigPanel.Hide()

	// 4. Create action buttons (changes per mode)
	actionButtons := ca.createOperationsButtons()

	// Layout: left panel (array), right panel (mode-specific content)
	rightPanel := container.NewVScroll(modeStack)

	mainContent := container.NewHSplit(
		arraySection,
		rightPanel,
	)
	mainContent.SetOffset(0.55) // Array gets 55% width (more space for integrated DAC/ADC)

	return container.NewBorder(
		modeSelector,
		actionButtons,
		nil, nil,
		mainContent,
	)
}

// createModeSelector creates the WRITE/READ/COMPUTE mode toggle with architecture selector
func (ca *CircuitsApp) createModeSelector() fyne.CanvasObject {
	modeRadio := widget.NewRadioGroup([]string{"WRITE", "READ", "COMPUTE"}, func(mode string) {
		ca.onModeChanged(mode)
	})
	modeRadio.Horizontal = true
	modeRadio.SetSelected("WRITE")

	modeHelp := widget.NewLabel("")
	modeHelp.TextStyle = fyne.TextStyle{Italic: true}
	ca.operationsModeHelp = modeHelp

	// Create architecture toggle (1T1R vs 0T1R)
	archToggle := ca.createArchitectureToggle()

	ca.updateModeHelp()

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Mode:"),
			modeRadio,
			layout.NewSpacer(),
			archToggle,
			layout.NewSpacer(),
			ca.operationsStatusLabel,
		),
		modeHelp,
		widget.NewSeparator(),
	)
}

// createSharedArraySection creates the array panel that's always visible
func (ca *CircuitsApp) createSharedArraySection() fyne.CanvasObject {
	// Create tappable array canvas
	tappableArray := NewTappableArrayCanvas(ca, ca.drawSharedArray, ca.onArrayCellTapped)
	// Larger size to accommodate integrated DAC (top) and ADC (right) boxes
	tappableArray.SetMinSize(fyne.NewSize(480, 420))
	ca.sharedArrayCanvas = tappableArray.raster // Keep reference for refresh

	// Color legend
	legendLabel := widget.NewLabel("Level: Low (blue) -> High (red) | Yellow = Selected | Click to select")
	legendLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Cell info display
	ca.sharedCellInfoLabel = widget.NewLabel("Click a cell to select")

	// Array size info
	ca.sharedArrayInfoLabel = widget.NewLabel(fmt.Sprintf("Array: %dx%d | %d levels", ca.arrayRows, ca.arrayCols, ca.quantLevels))

	titleLabel := widget.NewLabelWithStyle("CROSSBAR ARRAY", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create placeholder container for compute input row (will be populated later)
	ca.computeInputRowContainer = container.NewVBox()
	ca.computeInputRowContainer.Hide() // Initially hidden

	return container.NewVBox(
		titleLabel,
		ca.computeInputRowContainer, // Input row appears here in COMPUTE mode
		tappableArray,
		legendLabel,
		ca.sharedCellInfoLabel,
		ca.sharedArrayInfoLabel,
	)
}

// drawSharedArray draws the shared array visualization with click interaction
func (ca *CircuitsApp) drawSharedArray(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	selectedRow := ca.selectedRow
	selectedCol := ca.selectedCol
	levels := ca.quantLevels
	mode := ca.currentMode
	animStep := ca.animationStep
	arch := ca.architecture
	ca.mu.RUnlock()

	// Draw gradient background (dark blue to darker blue)
	bgTop := color.RGBA{15, 25, 45, 255}
	bgBottom := color.RGBA{5, 15, 35, 255}
	drawGradientRect(img, 0, 0, w, h, bgTop, bgBottom)

	if weights == nil || len(weights) == 0 {
		return img
	}

	// Calculate asymmetric margins for integrated DAC (top) and ADC (right)
	topMargin := 75   // Space for DAC boxes + labels above grid
	rightMargin := 75 // Space for ADC boxes + labels right of grid
	bottomMargin := 30
	leftMargin := 45 // Extra space for 1T1R transistors

	// Adjust left margin for 1T1R mode
	if arch == sharedwidgets.Architecture1T1R {
		leftMargin = 60
	}

	// Calculate available area for grid
	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	// Calculate cell size (use square cells)
	cellW := availableW / cols
	cellH := availableH / rows
	cellSize := cellW
	if cellH < cellSize {
		cellSize = cellH
	}
	if cellSize > 40 {
		cellSize = 40
	}
	if cellSize < 10 {
		cellSize = 10
	}

	// Calculate grid dimensions
	gridW := cols * cellSize
	gridH := rows * cellSize

	// Calculate offset to center grid within available area
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	// Store cell geometry for click detection
	ca.mu.Lock()
	ca.sharedArrayCellSize = cellSize
	ca.sharedArrayOffsetX = offsetX
	ca.sharedArrayOffsetY = offsetY
	ca.mu.Unlock()

	// ========== Draw subtle grid background ==========
	gridBgColor := color.RGBA{20, 35, 55, 255}
	drawRoundedRect(img, offsetX-4, offsetY-4, gridW+8, gridH+8, 6, gridBgColor)

	// Draw subtle grid lines (always visible)
	gridLineColor := color.RGBA{35, 50, 70, 255}
	// Vertical lines
	for c := 0; c <= cols; c++ {
		x := offsetX + c*cellSize
		for y := offsetY; y < offsetY+gridH; y++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, gridLineColor)
			}
		}
	}
	// Horizontal lines
	for r := 0; r <= rows; r++ {
		y := offsetY + r*cellSize
		for x := offsetX; x < offsetX+gridW; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, gridLineColor)
			}
		}
	}

	// ========== Draw wire connections (column BLs and row WLs) ==========
	// Bit Lines (vertical, for input voltages in compute mode)
	blColor := color.RGBA{60, 80, 120, 200}
	if mode == ModeCompute {
		blColor = color.RGBA{100, 120, 180, 255} // Brighter in compute mode
	}
	for c := 0; c < cols; c++ {
		x := offsetX + c*cellSize + cellSize/2
		// Draw from DAC area to bottom of grid
		for y := offsetY - 20; y < offsetY+gridH+5; y++ {
			if y >= 0 && y < h {
				img.Set(x, y, blColor)
				if cellSize > 15 {
					img.Set(x+1, y, blColor) // Thicker wire
				}
			}
		}
	}

	// Word Lines (horizontal, for row selection)
	wlColor := color.RGBA{120, 80, 60, 200}
	if mode == ModeCompute {
		wlColor = color.RGBA{180, 120, 100, 255} // Brighter in compute mode
	}
	for r := 0; r < rows; r++ {
		y := offsetY + r*cellSize + cellSize/2
		startX := offsetX - 15
		if arch == sharedwidgets.Architecture1T1R {
			startX = offsetX - 30
		}
		// Draw from left edge to ADC area
		for x := startX; x < offsetX+gridW+20; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, wlColor)
				if cellSize > 15 {
					img.Set(x, y+1, wlColor) // Thicker wire
				}
			}
		}
	}

	// ========== Draw 1T1R transistor indicators ==========
	if arch == sharedwidgets.Architecture1T1R {
		transistorX := offsetX - 35

		for r := 0; r < rows; r++ {
			transistorY := offsetY + r*cellSize + cellSize/2

			// Determine transistor state
			var transistorOn bool
			switch mode {
			case ModeWrite, ModeRead:
				transistorOn = (r == selectedRow)
			case ModeCompute:
				transistorOn = true
			}

			// Draw transistor with glow effect when ON
			if transistorOn {
				// Glow effect for ON state
				drawGlowCircle(img, transistorX, transistorY, 6,
					color.RGBA{80, 255, 80, 255},  // Bright green center
					color.RGBA{50, 200, 50, 150})  // Green glow
			} else {
				// Simple gray circle for OFF state
				offColor := color.RGBA{50, 55, 65, 255}
				for dy := -5; dy <= 5; dy++ {
					for dx := -5; dx <= 5; dx++ {
						if dx*dx+dy*dy <= 25 {
							px, py := transistorX+dx, transistorY+dy
							if px >= 0 && px < w && py >= 0 && py < h {
								img.Set(px, py, offColor)
							}
						}
					}
				}
			}

			// Draw gate terminal (vertical bar)
			gateColor := color.RGBA{100, 100, 120, 255}
			if transistorOn {
				gateColor = color.RGBA{150, 255, 150, 255}
			}
			gateX := transistorX - 10
			for dy := -6; dy <= 6; dy++ {
				py := transistorY + dy
				if gateX >= 0 && gateX < w && py >= 0 && py < h {
					img.Set(gateX, py, gateColor)
					img.Set(gateX+1, py, gateColor)
				}
			}
		}

		// Draw "WL" label
		wlColor := color.RGBA{130, 180, 130, 255}
		drawSimpleText(img, "WL", transistorX-15, offsetY-20, wlColor)
	}

	// ========== Draw cells with improved colors ==========
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := offsetX + c*cellSize + 1
			y0 := offsetY + r*cellSize + 1
			cw := cellSize - 2
			ch := cellSize - 2

			level := weights[r][c]
			isSelected := r == selectedRow && c == selectedCol

			// Get cell color using improved gradient
			var cellColor color.RGBA
			if isSelected {
				cellColor = color.RGBA{255, 220, 80, 255} // Bright gold for selection
			} else {
				cellColor = levelToColor(level, levels)
			}

			// Draw cell with gradient effect (lighter at top)
			topColor := color.RGBA{
				uint8(min(int(cellColor.R)+30, 255)),
				uint8(min(int(cellColor.G)+30, 255)),
				uint8(min(int(cellColor.B)+30, 255)),
				255,
			}
			drawGradientRect(img, x0, y0, cw, ch, topColor, cellColor)

			// Draw cell border
			borderColor := color.RGBA{
				uint8(min(int(cellColor.R)+50, 255)),
				uint8(min(int(cellColor.G)+50, 255)),
				uint8(min(int(cellColor.B)+50, 255)),
				255,
			}
			drawRectBorder(img, x0, y0, cw, ch, borderColor)

			// Highlight selected cell with glow
			if isSelected {
				// Draw white border
				white := color.RGBA{255, 255, 255, 255}
				drawRectBorder(img, x0-1, y0-1, cw+2, ch+2, white)
				drawRectBorder(img, x0-2, y0-2, cw+4, ch+4, color.RGBA{255, 255, 200, 180})
			}

			// Animation highlight in compute mode
			if mode == ModeCompute && animStep == 2 {
				overlayColor := color.RGBA{0, 255, 255, 80}
				drawRectBorder(img, x0, y0, cw, ch, overlayColor)
				drawRectBorder(img, x0+1, y0+1, cw-2, ch-2, overlayColor)
			}
		}
	}

	// ========== Draw mode-specific overlays ==========
	switch mode {
	case ModeWrite:
		// Show write target arrow
		if selectedRow < rows && selectedCol < cols {
			arrowX := offsetX + selectedCol*cellSize - 12
			arrowY := offsetY + selectedRow*cellSize + cellSize/2
			arrowColor := color.RGBA{255, 200, 50, 255}
			// Draw arrow shape
			for i := 0; i < 10; i++ {
				img.Set(arrowX+i, arrowY, arrowColor)
				img.Set(arrowX+i, arrowY-1, arrowColor)
				img.Set(arrowX+i, arrowY+1, arrowColor)
			}
			// Arrow head
			for j := 0; j < 4; j++ {
				img.Set(arrowX+10+j, arrowY-j, arrowColor)
				img.Set(arrowX+10+j, arrowY+j, arrowColor)
			}
		}

	case ModeRead:
		// Show read probe with glow
		if selectedRow < rows && selectedCol < cols {
			probeX := offsetX + selectedCol*cellSize + cellSize/2
			probeY := offsetY + selectedRow*cellSize + cellSize/2
			drawGlowCircle(img, probeX, probeY, cellSize/4,
				color.RGBA{0, 255, 255, 255},
				color.RGBA{0, 200, 200, 100})
		}

	case ModeCompute:
		// ========== Draw DAC boxes (top) ==========
		dacBoxHeight := 28
		dacBoxWidth := cellSize - 2
		dacY := offsetY - dacBoxHeight - 15

		dacColCount := min(8, cols)
		inputVectorCopy := make([]int, dacColCount)
		ca.mu.RLock()
		copy(inputVectorCopy, ca.inputVector[:dacColCount])
		ca.mu.RUnlock()

		for c := 0; c < dacColCount; c++ {
			dacX := offsetX + c*cellSize + 1

			// DAC gradient colors (purple theme)
			dacTopColor := color.RGBA{140, 100, 200, 255}
			dacBottomColor := color.RGBA{90, 60, 160, 255}
			if animStep == 1 {
				dacTopColor = color.RGBA{255, 255, 150, 255}
				dacBottomColor = color.RGBA{255, 220, 100, 255}
			}

			drawGradientRect(img, dacX, dacY, dacBoxWidth, dacBoxHeight, dacTopColor, dacBottomColor)
			drawRectBorder(img, dacX, dacY, dacBoxWidth, dacBoxHeight, color.RGBA{180, 150, 230, 255})

			// Voltage value
			inputVal := inputVectorCopy[c]
			voltage := float64(inputVal) / 255.0
			voltageText := fmt.Sprintf("%.2f", voltage)
			textX := dacX + dacBoxWidth/2 - len(voltageText)*3
			textY := dacY + dacBoxHeight/2 - 3
			drawSimpleText(img, voltageText, textX, textY, color.RGBA{255, 255, 255, 255})

			// Column label
			labelX := offsetX + c*cellSize + cellSize/2 - 6
			labelY := dacY - 14
			drawSimpleText(img, fmt.Sprintf("x%d", c), labelX, labelY, color.RGBA{150, 180, 255, 255})
		}

		// "DAC" label
		drawSimpleText(img, "DAC", offsetX-25, dacY+dacBoxHeight/2-3, color.RGBA{180, 150, 230, 255})

		// ========== Draw ADC/TIA boxes (right) ==========
		adcBoxWidth := 50
		adcBoxHeight := cellSize - 2
		adcX := offsetX + gridW + 12

		adcRowCount := min(8, rows)
		outputVectorCopy := make([]float64, adcRowCount)
		ca.mu.RLock()
		copy(outputVectorCopy, ca.outputVector[:min(adcRowCount, len(ca.outputVector))])
		ca.mu.RUnlock()

		for r := 0; r < adcRowCount; r++ {
			adcY := offsetY + r*cellSize + 1

			// ADC gradient colors (green/teal theme)
			adcTopColor := color.RGBA{80, 180, 140, 255}
			adcBottomColor := color.RGBA{50, 130, 100, 255}
			if animStep == 3 {
				adcTopColor = color.RGBA{120, 255, 180, 255}
				adcBottomColor = color.RGBA{80, 220, 140, 255}
			}

			drawGradientRect(img, adcX, adcY, adcBoxWidth, adcBoxHeight, adcTopColor, adcBottomColor)
			drawRectBorder(img, adcX, adcY, adcBoxWidth, adcBoxHeight, color.RGBA{140, 220, 180, 255})

			// ADC level
			outputVal := outputVectorCopy[r]
			tiaVoltage := ca.tia.Convert(outputVal * 1e-6)
			adcLevel := ca.adc.Convert(tiaVoltage)

			levelText := fmt.Sprintf("L%d", adcLevel)
			textX := adcX + adcBoxWidth/2 - len(levelText)*3
			textY := adcY + adcBoxHeight/2 - 3
			drawSimpleText(img, levelText, textX, textY, color.RGBA{255, 255, 255, 255})

			// Row label
			labelX := adcX + adcBoxWidth + 6
			labelY := offsetY + r*cellSize + cellSize/2 - 3
			drawSimpleText(img, fmt.Sprintf("y%d", r), labelX, labelY, color.RGBA{255, 200, 150, 255})
		}

		// "TIA+ADC" label
		labelY := offsetY - 15
		drawSimpleText(img, "TIA+ADC", adcX, labelY, color.RGBA{140, 220, 180, 255})
	}

	// ========== Draw title based on mode ==========
	var titleText string
	var titleColor color.RGBA
	switch mode {
	case ModeWrite:
		titleText = "WRITE MODE"
		titleColor = color.RGBA{255, 200, 100, 255}
	case ModeRead:
		titleText = "READ MODE"
		titleColor = color.RGBA{100, 255, 255, 255}
	case ModeCompute:
		titleText = "COMPUTE MODE"
		titleColor = color.RGBA{200, 150, 255, 255}
	}
	titleX := offsetX + gridW/2 - len(titleText)*3
	titleY := 8
	drawSimpleText(img, titleText, titleX, titleY, titleColor)

	return img
}

// refreshSharedArray refreshes the shared array canvas
func (ca *CircuitsApp) refreshSharedArray() {
	if ca.sharedArrayCanvas != nil {
		fyne.Do(func() {
			ca.sharedArrayCanvas.Refresh()
		})
	}
}

// onModeChanged handles mode switching
func (ca *CircuitsApp) onModeChanged(mode string) {
	ca.mu.Lock()
	switch mode {
	case "WRITE":
		ca.currentMode = ModeWrite
	case "READ":
		ca.currentMode = ModeRead
	case "COMPUTE":
		ca.currentMode = ModeCompute
	}
	ca.mu.Unlock()

	// Update visible panels
	ca.updateOperationsPanels()
	ca.updateModeHelp()
	ca.refreshSharedArray()
	ca.updateSharedCellInfo()

	// Auto-compute when entering COMPUTE mode
	if mode == "COMPUTE" {
		ca.computeAndUpdateAll()
	}
}

// updateOperationsPanels shows/hides panels based on current mode
func (ca *CircuitsApp) updateOperationsPanels() {
	ca.mu.RLock()
	mode := ca.currentMode
	ca.mu.RUnlock()

	// CRITICAL: Hide ALL panels first to prevent leftover UI
	if ca.writeConfigPanel != nil {
		ca.writeConfigPanel.Hide()
	}
	if ca.readConfigPanel != nil {
		ca.readConfigPanel.Hide()
	}
	if ca.computeConfigPanel != nil {
		ca.computeConfigPanel.Hide()
	}

	// THEN show only the selected panel
	switch mode {
	case ModeWrite:
		if ca.writeConfigPanel != nil {
			ca.writeConfigPanel.Show()
		}
	case ModeRead:
		if ca.readConfigPanel != nil {
			ca.readConfigPanel.Show()
		}
	case ModeCompute:
		if ca.computeConfigPanel != nil {
			ca.computeConfigPanel.Show()
		}
	}

	// Toggle input row visibility in array section
	if ca.computeInputRowContainer != nil {
		if mode == ModeCompute {
			ca.computeInputRowContainer.Show()
		} else {
			ca.computeInputRowContainer.Hide()
		}
	}

	// Update action buttons
	ca.updateOperationsButtons()
}

// updateModeHelp updates the mode description text with architecture-aware context
func (ca *CircuitsApp) updateModeHelp() {
	if ca.operationsModeHelp == nil {
		return
	}

	ca.mu.RLock()
	mode := ca.currentMode
	arch := ca.architecture
	ca.mu.RUnlock()

	is1T1R := arch == sharedwidgets.Architecture1T1R

	var helpText string
	switch mode {
	case ModeWrite:
		if is1T1R {
			helpText = "WRITE: Transistor gates ONLY selected row (green●). Full write pulse to target cell, others isolated."
		} else {
			helpText = "WRITE: Passive array - partial voltages affect neighboring rows (sneak paths ~5-20% error)."
		}
	case ModeRead:
		if is1T1R {
			helpText = "READ: Transistor isolates selected row (green●). Clean sense current from target cell only."
		} else {
			helpText = "READ: Passive array - sneak currents add ~5-20% noise to sense signal."
		}
	case ModeCompute:
		if is1T1R {
			helpText = "COMPUTE: ALL transistors ON (all green●) for full MVM. Sneak-free parallel computation."
		} else {
			helpText = "COMPUTE: Passive MVM - sneak paths cause ~5-20% output error. Still functional for AI inference."
		}
	}

	fyne.Do(func() {
		ca.operationsModeHelp.SetText(helpText)
	})
}

// updateSharedCellInfo updates the cell info display
func (ca *CircuitsApp) updateSharedCellInfo() {
	if ca.sharedCellInfoLabel == nil {
		return
	}

	ca.mu.RLock()
	row := ca.selectedRow
	col := ca.selectedCol
	mode := ca.currentMode
	levels := ca.quantLevels
	var level int
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		level = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	conductance := 1.0 + float64(level)/float64(levels-1)*99.0

	var infoText string
	switch mode {
	case ModeWrite:
		infoText = fmt.Sprintf("Cell [%d,%d]: Level %d | Target: %d | G=%.1f uS", row, col, level, ca.targetLevel, conductance)
	case ModeRead:
		infoText = fmt.Sprintf("Cell [%d,%d]: Level %d | G=%.1f uS | Ready to read", row, col, level, conductance)
	case ModeCompute:
		infoText = fmt.Sprintf("Cell [%d,%d]: Level %d | G=%.1f uS | Weight in MVM", row, col, level, conductance)
	}

	fyne.Do(func() {
		ca.sharedCellInfoLabel.SetText(infoText)
	})
}

// onArrayCellTapped handles cell selection via click
func (ca *CircuitsApp) onArrayCellTapped(row, col int) {
	ca.mu.Lock()
	ca.selectedRow = row
	ca.selectedCol = col
	ca.mu.Unlock()

	ca.refreshSharedArray()
	ca.updateSharedCellInfo()
	ca.updateOpsWriteDataPath()
}

// ============================================================================
// ACTION BUTTONS
// ============================================================================

// ============================================================================
// ACTION BUTTONS
// ============================================================================

// createOperationsButtons creates the mode-specific action buttons
func (ca *CircuitsApp) createOperationsButtons() fyne.CanvasObject {
	// Write mode buttons
	ca.opsProgramBtn = widget.NewButton("PROGRAM CELL", ca.onOpsProgram)
	ca.opsProgramBtn.Importance = widget.HighImportance
	ca.opsProgramRandomBtn = widget.NewButton("RANDOM ARRAY", ca.onOpsProgramRandom)

	// Read mode buttons
	ca.opsReadBtn = widget.NewButton("READ CELL", ca.onOpsRead)
	ca.opsReadBtn.Importance = widget.HighImportance
	ca.opsVerifyBtn = widget.NewButton("VERIFY ARRAY", ca.onOpsVerify)

	// Compute mode buttons
	ca.opsComputeBtn = widget.NewButton("COMPUTE", ca.onOpsCompute)
	ca.opsComputeBtn.Importance = widget.HighImportance
	ca.opsAnimateBtn = widget.NewButton("ANIMATE", ca.onOpsAnimate)
	ca.opsResetBtn = widget.NewButton("RESET", ca.onOpsReset)

	// Create button containers for each mode
	ca.opsWriteButtons = container.NewHBox(ca.opsProgramBtn, ca.opsProgramRandomBtn)
	ca.opsReadButtons = container.NewHBox(ca.opsReadBtn, ca.opsVerifyBtn)
	ca.opsComputeButtons = container.NewHBox(ca.opsComputeBtn, ca.opsAnimateBtn, ca.opsResetBtn)

	// Stack all button sets
	buttonStack := container.NewStack(
		ca.opsWriteButtons,
		ca.opsReadButtons,
		ca.opsComputeButtons,
	)

	// Initialize visibility
	ca.opsWriteButtons.Show()
	ca.opsReadButtons.Hide()
	ca.opsComputeButtons.Hide()

	return container.NewHBox(
		buttonStack,
		layout.NewSpacer(),
		ca.operationsStatusLabel,
	)
}

// updateOperationsButtons shows/hides action buttons based on mode
func (ca *CircuitsApp) updateOperationsButtons() {
	ca.mu.RLock()
	mode := ca.currentMode
	ca.mu.RUnlock()

	// CRITICAL: Hide ALL button sets first to prevent leftover UI
	if ca.opsWriteButtons != nil {
		ca.opsWriteButtons.Hide()
	}
	if ca.opsReadButtons != nil {
		ca.opsReadButtons.Hide()
	}
	if ca.opsComputeButtons != nil {
		ca.opsComputeButtons.Hide()
	}

	// THEN show only the selected button set
	switch mode {
	case ModeWrite:
		if ca.opsWriteButtons != nil {
			ca.opsWriteButtons.Show()
		}
	case ModeRead:
		if ca.opsReadButtons != nil {
			ca.opsReadButtons.Show()
		}
	case ModeCompute:
		if ca.opsComputeButtons != nil {
			ca.opsComputeButtons.Show()
		}
	}
}

// ============================================================================
// ARCHITECTURE TOGGLE (1T1R vs 0T1R)
// ============================================================================

// createArchitectureToggle creates the PASSIVE/1T1R toggle buttons
// 1T1R: Transistor gates each row - only selected row active (write/read) or all rows (compute)
// 0T1R: Passive crossbar - sneak paths affect accuracy
func (ca *CircuitsApp) createArchitectureToggle() fyne.CanvasObject {
	// Create toggle buttons (same pattern as Module 2)
	ca.archPassiveBtn = widget.NewButton("PASSIVE", nil)
	ca.arch1T1RBtn = widget.NewButton("1T1R GATE", nil)

	// Helper to update button styles based on selection
	updateArchButtons := func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			ca.archPassiveBtn.Importance = widget.HighImportance
			ca.arch1T1RBtn.Importance = widget.LowImportance
		} else {
			ca.archPassiveBtn.Importance = widget.LowImportance
			ca.arch1T1RBtn.Importance = widget.HighImportance
		}
		ca.archPassiveBtn.Refresh()
		ca.arch1T1RBtn.Refresh()
	}

	// Set initial state
	updateArchButtons()

	// Wire up callbacks
	ca.archPassiveBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			return // Already selected
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture0T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.refreshSharedArray()
		ca.updateModeHelp()
	}

	ca.arch1T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture1T1R {
			return // Already selected
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture1T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.refreshSharedArray()
		ca.updateModeHelp()
	}

	ca.archToggle = container.NewGridWithColumns(2, ca.archPassiveBtn, ca.arch1T1RBtn)

	archLabel := widget.NewLabel("Array:")
	return container.NewHBox(archLabel, ca.archToggle)
}

