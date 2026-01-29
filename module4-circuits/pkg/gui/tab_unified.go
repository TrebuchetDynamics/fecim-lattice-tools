// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device simulation view that replaces separate WRITE/READ/COMPUTE modes.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	configphysics "fecim-lattice-tools/config/physics"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ============================================================================
// UNIFIED DEVICE SIMULATION VIEW
// ============================================================================

// createUnifiedView creates the unified device simulation view
// Replaces the old mode-based createOperationsView()
func (ca *CircuitsApp) createUnifiedView() fyne.CanvasObject {
	// Initialize device state
	ca.deviceState = NewDeviceState(ca.arrayRows, ca.arrayCols, ca.tia, ca.adc)
	ca.operationsStatusLabel = widget.NewLabel("Ready")

	// In passive (0T1R) mode, all WLs are always active - no transistor gating
	if ca.architecture == sharedwidgets.Architecture0T1R {
		ca.deviceState.SetPassiveMode(true)
	}

	// 1. Signal chain header
	signalChainHeader := ca.createSignalChainHeader()

	// 2. Mode bar at top (Mode-First UX)
	modeBar := ca.createModeBar()

	// 3. Mode-specific panels (initially hidden, shown based on mode)
	writePanelContent := ca.createEnhancedWriteModePanel()
	ca.writeModePanel = container.NewVBox(writePanelContent)
	ca.writeModePanel.Hide() // Hidden by default (READ mode)

	computePanelContent := ca.createComputeModePanel()
	ca.computeModePanel = container.NewVBox(computePanelContent)
	ca.computeModePanel.Hide() // Hidden by default (READ mode)

	// Stack the mode panels (only one visible at a time)
	modePanelStack := container.NewStack(ca.writeModePanel, ca.computeModePanel)

	// Initialize architecture-specific voltage panels
	ca.passiveVoltagePanel = ca.createPassiveVoltagePanel()
	ca.activeVoltagePanel = ca.createActiveVoltagePanel()
	ca.passiveVoltagePanel.Hide() // Hidden initially (1T1R default)
	archVoltageStack := container.NewStack(ca.passiveVoltagePanel, ca.activeVoltagePanel)

	// 4. DAC input section
	dacSection := ca.createDACInputSection()

	// Update DAC range mode label with current voltage range
	ca.updateDACRangeModeLabel()

	// 5. Main visualization area (center)
	mainSection := ca.createMainSimSection()

	// 6. Action buttons (bottom)
	actionSection := ca.createUnifiedActionSection()

	// Initialize button states for default READ mode
	ca.updateActionButtons()

	// Top section: signal chain header, mode bar, mode panels, architecture voltage panels, DAC presets
	topSection := container.NewVBox(
		signalChainHeader,
		modeBar,
		modePanelStack,
		archVoltageStack,
		dacSection,
	)

	return container.NewBorder(
		topSection,    // top
		actionSection, // bottom
		nil, nil,
		mainSection, // center
	)
}

// createSignalChainHeader creates the signal chain indicator
func (ca *CircuitsApp) createSignalChainHeader() fyne.CanvasObject {
	// Architecture toggle
	archToggle := ca.createArchitectureToggle()

	// Material selector
	materialSelector := ca.createMaterialSelector()

	// ADC bits selector
	adcBitsSelector := ca.createADCBitsSelector()

	chainLabel := widget.NewLabelWithStyle(
		"SIGNAL CHAIN: DAC -> Array -> TIA -> ADC",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Operation classification label (updates based on configuration)
	ca.operationsModeHelp = widget.NewLabel("Configuration: Click cells or adjust voltages")
	ca.operationsModeHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Circuit specs summary - shows current configuration
	adcBits := 5
	if ca.adc != nil {
		adcBits = ca.adc.Bits
	}
	adcLevels := 1 << adcBits
	circuitSpecsLabel := widget.NewLabel(fmt.Sprintf("ADC: %d-bit (%d levels, 0-%d)", adcBits, adcLevels, adcLevels-1))
	circuitSpecsLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewVBox(
		container.NewHBox(
			chainLabel,
			layout.NewSpacer(),
			materialSelector,
			adcBitsSelector,
			layout.NewSpacer(),
			archToggle,
			layout.NewSpacer(),
			circuitSpecsLabel,
		),
		ca.operationsModeHelp,
		widget.NewSeparator(),
	)
}

// createMaterialSelector creates the ferroelectric material selection dropdown with browse button
func (ca *CircuitsApp) createMaterialSelector() fyne.CanvasObject {
	materials := sharedphysics.AllMaterials()
	materialNames := make([]string, len(materials))
	for i, m := range materials {
		materialNames[i] = m.Name
	}

	selector := widget.NewSelect(materialNames, func(selected string) {
		// Find the material and set it
		for _, m := range materials {
			if m.Name == selected {
				ca.deviceState.SetMaterial(m)
				ca.updateDACRangeModeLabel() // Update mode indicator
				ca.recomputeAndRefresh()
				ca.operationsStatusLabel.SetText(fmt.Sprintf("Material: %s (Vc=%.2fV)", selected, m.CoerciveVoltage()))
				break
			}
		}
	})

	// Set default selection to FeCIM material
	selector.SetSelected("FeCIM HZO")

	// Material picker button for detailed view
	pickerBtn := sharedwidgets.CreateMaterialPickerButton(
		ca.window,
		"fecim_hzo", // Default material ID
		func(materialID string, mat *configphysics.Material) {
			if mat == nil {
				return
			}
			// Update dropdown to match selection
			selector.SetSelected(mat.Name)
			// The dropdown's OnChanged will handle the rest
		},
	)

	return container.NewHBox(widget.NewLabel("Material:"), selector, pickerBtn)
}

// createADCBitsSelector creates a dropdown to select ADC resolution (5-8 bits)
func (ca *CircuitsApp) createADCBitsSelector() fyne.CanvasObject {
	options := []string{"5-bit (32)", "6-bit (64)", "7-bit (128)", "8-bit (256)"}

	selector := widget.NewSelect(options, func(selected string) {
		var bits int
		switch selected {
		case "5-bit (32)":
			bits = 5
		case "6-bit (64)":
			bits = 6
		case "7-bit (128)":
			bits = 7
		case "8-bit (256)":
			bits = 8
		default:
			bits = 5
		}
		ca.deviceState.SetADCBits(bits)
		ca.recomputeAndRefresh()
		levels := 1 << bits
		ca.operationsStatusLabel.SetText(fmt.Sprintf("ADC: %d-bit (%d levels, 0-%d)", bits, levels, levels-1))
	})

	// Set default selection
	selector.SetSelected("5-bit (32)")

	return container.NewHBox(widget.NewLabel("ADC:"), selector)
}

// createDACInputSection creates the DAC status and manual control
func (ca *CircuitsApp) createDACInputSection() fyne.CanvasObject {
	// Initialize DAC entries array (used by updateDACEntries but not displayed)
	maxCols := min(8, ca.arrayCols)
	ca.unifiedDACEntries = make([]*widget.Entry, maxCols)
	ca.unifiedDACLabels = make([]*widget.Label, maxCols)

	// Range mode indicator - shows current DAC voltage range based on operation mode
	// Note: DAC range is set automatically by mode (READ/WRITE/COMPUTE)
	// Random input is available in COMPUTE mode panel
	ca.dacRangeLabel = widget.NewLabel("DAC: Read Range")
	ca.dacRangeLabel.TextStyle = fyne.TextStyle{Italic: true}

	// "Set All" entry for bulk voltage (manual override)
	allEntry := widget.NewEntry()
	allEntry.SetPlaceHolder("0.50")
	allEntry.OnSubmitted = func(s string) {
		ca.setAllUnifiedDACVoltages(s)
	}

	return container.NewHBox(
		ca.dacRangeLabel,
		layout.NewSpacer(),
		widget.NewLabel("Set All (V):"), allEntry,
	)
}

// updateDACRangeModeLabel updates the DAC range mode indicator based on operation mode
func (ca *CircuitsApp) updateDACRangeModeLabel() {
	if ca.dacRangeLabel == nil || ca.deviceState == nil {
		return
	}

	rangeMode := ca.deviceState.GetDACRangeMode()
	currentRange := ca.deviceState.GetCurrentVoltageRange()

	var text string
	if rangeMode == DACRangeWrite {
		text = fmt.Sprintf("DAC: Write (%.1f-%.1fV)", currentRange.Min, currentRange.Max)
	} else {
		text = fmt.Sprintf("DAC: Read (0-%.1fV)", currentRange.Max)
	}

	fyne.Do(func() {
		ca.dacRangeLabel.SetText(text)
	})
}

// createMainSimSection creates the main simulation visualization area
func (ca *CircuitsApp) createMainSimSection() fyne.CanvasObject {
	// WL checkboxes removed - row selection is done by clicking cells
	// WL state is determined automatically by mode and architecture:
	// - Passive (0T1R): All WLs always on
	// - 1T1R/2T1R READ/WRITE: Selected row only (via cell click)
	// - COMPUTE: All WLs on for MVM

	// Initialize empty WL checks array (some code may reference it)
	ca.unifiedWLChecks = make([]*widget.Check, 0)

	// Array canvas with DAC inputs at top, TIA/ADC outputs at right
	return ca.createUnifiedArraySection()
}

// setOperationMode sets the operation mode and configures WL/DAC accordingly
// READ: Single row, safe voltage (0-0.5V)
// WRITE: Single row, write voltage (1.2-1.5V on selected column)
// COMPUTE: All rows active, input vector (0-1V)
// NOTE: In passive mode (0T1R), all WLs are ALWAYS on - WL configuration is skipped
func (ca *CircuitsApp) setOperationMode(mode OpMode) {
	if ca.deviceState == nil {
		return
	}

	ca.deviceState.SetOperationMode(mode)

	// In passive mode, all WLs are always on - skip WL configuration
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R

	switch mode {
	case OpModeRead:
		// Single row active (only in 1T1R/2T1R)
		if !isPassive {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.deviceState.SetDACRangeMode(DACRangeRead)
		// Per VOLTAGE_RULES.md: Only selected column gets read voltage
		// Ground all columns first, then apply to selected column only
		readVoltage := ca.deviceState.GetReadRange().Max * 0.4 // ~0.2V safe read
		if readVoltage < 0.1 {
			readVoltage = 0.2
		}
		ca.deviceState.SetAllDACVoltages(0)
		ca.deviceState.SetDACVoltage(ca.deviceState.GetSelectedCol(), readVoltage)

	case OpModeWrite:
		// Single row active (only in 1T1R/2T1R)
		// DAC voltages stay at 0 (write requires explicit action to avoid accidents)
		if !isPassive {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.deviceState.SetDACRangeMode(DACRangeWrite)
		ca.deviceState.SetAllDACVoltages(0) // Safe: no voltage until explicit write

	case OpModeCompute:
		// All rows active for MVM
		if !isPassive {
			ca.deviceState.SetWLAll()
		}
		ca.deviceState.SetDACRangeMode(DACRangeRead)
		// Apply input vector as DAC voltages (MVM: I = G × V)
		// Input vector values (0-255) map to voltage range
		params := make([]float64, len(ca.inputVector))
		for i, v := range ca.inputVector {
			params[i] = float64(v)
		}
		ca.deviceState.SetDACPreset(DACInputVector, params...)
	}

	ca.updateModeButtons()
	ca.updateActionButtons() // Enable/disable action buttons based on mode
	ca.updateModePanels(mode) // Show/hide mode-specific panels
	ca.updateWLCheckboxes()
	ca.updateWLHelpLabel()
	ca.updateDACRangeModeLabel()
	ca.recomputeAndRefresh()
}

// updateWLHelpLabel is a no-op - WL UI has been removed
func (ca *CircuitsApp) updateWLHelpLabel() {
	// No-op: WL UI removed
}

// updateModeButtons updates the mode button highlighting
func (ca *CircuitsApp) updateModeButtons() {
	if ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()

	fyne.Do(func() {
		// Reset all to low importance
		if ca.modeReadBtn != nil {
			ca.modeReadBtn.Importance = widget.LowImportance
			ca.modeReadBtn.Refresh()
		}
		if ca.modeWriteBtn != nil {
			ca.modeWriteBtn.Importance = widget.LowImportance
			ca.modeWriteBtn.Refresh()
		}
		if ca.modeComputeBtn != nil {
			ca.modeComputeBtn.Importance = widget.LowImportance
			ca.modeComputeBtn.Refresh()
		}

		// Highlight active mode
		switch mode {
		case OpModeRead:
			if ca.modeReadBtn != nil {
				ca.modeReadBtn.Importance = widget.HighImportance
				ca.modeReadBtn.Refresh()
			}
		case OpModeWrite:
			if ca.modeWriteBtn != nil {
				ca.modeWriteBtn.Importance = widget.HighImportance
				ca.modeWriteBtn.Refresh()
			}
		case OpModeCompute:
			if ca.modeComputeBtn != nil {
				ca.modeComputeBtn.Importance = widget.HighImportance
				ca.modeComputeBtn.Refresh()
			}
		}
	})
}

// createUnifiedArraySection creates the array visualization section
func (ca *CircuitsApp) createUnifiedArraySection() fyne.CanvasObject {
	// Create tappable array canvas - larger size for better visualization
	tappableArray := NewUnifiedTappableCanvas(ca, ca.drawUnifiedArray, ca.onUnifiedCellTapped)
	tappableArray.SetMinSize(fyne.NewSize(850, 600)) // Large canvas for detailed visualization
	ca.sharedArrayCanvas = tappableArray.raster

	// Cell info display
	ca.sharedCellInfoLabel = widget.NewLabel("Click a cell to select")

	// Array size info with capacity calculation
	totalCells := ca.arrayRows * ca.arrayCols
	bitCapacity := float64(totalCells) * 4.9 // ~4.9 bits per 30-level cell
	ca.sharedArrayInfoLabel = widget.NewLabel(fmt.Sprintf("Array: %dx%d (%d cells) | %d levels (~%.0f bits)",
		ca.arrayRows, ca.arrayCols, totalCells, ca.quantLevels, bitCapacity))

	// Legend with energy info
	legendLabel := widget.NewLabel("States: Low G (blue) → High G (red) | Energy: READ ~45fJ, WRITE ~55fJ, MVM ~50fJ/cell")
	legendLabel.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		tappableArray,
		legendLabel,
		ca.sharedCellInfoLabel,
		ca.sharedArrayInfoLabel,
	)
}


// createUnifiedActionSection creates the action buttons
func (ca *CircuitsApp) createUnifiedActionSection() fyne.CanvasObject {
	// Program button - only enabled in WRITE mode
	ca.actionWriteCellBtn = widget.NewButton("Write Cell", func() {
		ca.onUnifiedProgram()
	})
	ca.actionWriteCellBtn.Importance = widget.HighImportance

	// Sense button - only enabled in READ mode
	ca.actionReadBtn = widget.NewButton("Sense Row", func() {
		ca.onUnifiedRead()
	})

	// Compute button - only enabled in COMPUTE mode
	ca.actionComputeBtn = widget.NewButton("Compute MVM", func() {
		ca.onUnifiedCompute()
	})

	// Animate button
	animateBtn := widget.NewButton("Animate", func() {
		ca.onUnifiedAnimate()
	})

	// H3 FIX: Undo button
	ca.undoHistoryBtn = widget.NewButton("Undo", func() {
		ca.onUndo()
	})
	ca.undoHistoryBtn.Disable() // Initially disabled (no history)

	// Reset array button
	resetBtn := widget.NewButton("Reset Array", func() {
		ca.onUnifiedReset()
	})

	// Random array button
	randomBtn := widget.NewButton("Random Array", func() {
		ca.onUnifiedRandomArray()
	})

	return container.NewHBox(
		ca.actionWriteCellBtn, ca.actionReadBtn, ca.actionComputeBtn,
		layout.NewSpacer(),
		ca.undoHistoryBtn, animateBtn, randomBtn, resetBtn,
	)
}

// updateActionButtons enables/disables action buttons based on current mode
func (ca *CircuitsApp) updateActionButtons() {
	if ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()

	fyne.Do(func() {
		// Write Cell: only in WRITE mode
		if ca.actionWriteCellBtn != nil {
			if mode == OpModeWrite {
				ca.actionWriteCellBtn.Enable()
				ca.actionWriteCellBtn.Importance = widget.HighImportance
			} else {
				ca.actionWriteCellBtn.Disable()
				ca.actionWriteCellBtn.Importance = widget.MediumImportance
			}
			ca.actionWriteCellBtn.Refresh()
		}

		// Read/Sense: only in READ mode
		if ca.actionReadBtn != nil {
			if mode == OpModeRead {
				ca.actionReadBtn.Enable()
			} else {
				ca.actionReadBtn.Disable()
			}
		}

		// Compute MVM: only in COMPUTE mode
		if ca.actionComputeBtn != nil {
			if mode == OpModeCompute {
				ca.actionComputeBtn.Enable()
			} else {
				ca.actionComputeBtn.Disable()
			}
		}
	})
}

// ============================================================================
// UNIFIED TAPPABLE CANVAS
// ============================================================================

// UnifiedTappableCanvas is a canvas.Raster that responds to taps for the unified view
type UnifiedTappableCanvas struct {
	widget.BaseWidget
	raster *canvas.Raster
	onTap  func(row, col int)
	ca     *CircuitsApp
}

func NewUnifiedTappableCanvas(ca *CircuitsApp, drawFunc func(w, h int) image.Image, onTap func(row, col int)) *UnifiedTappableCanvas {
	t := &UnifiedTappableCanvas{
		raster: canvas.NewRaster(drawFunc),
		onTap:  onTap,
		ca:     ca,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *UnifiedTappableCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.raster)
}

func (t *UnifiedTappableCanvas) SetMinSize(size fyne.Size) {
	t.raster.SetMinSize(size)
}

func (t *UnifiedTappableCanvas) Refresh() {
	t.raster.Refresh()
}

func (t *UnifiedTappableCanvas) Tapped(e *fyne.PointEvent) {
	size := t.raster.Size()

	t.ca.mu.RLock()
	rows := t.ca.arrayRows
	cols := t.ca.arrayCols
	t.ca.mu.RUnlock()

	w := int(size.Width)
	h := int(size.Height)

	// Same margin calculations as drawUnifiedArray
	topMargin := 50
	rightMargin := 20
	bottomMargin := 25
	leftMargin := 25

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows
	cellSize := min(cellW, cellH)
	if cellSize > 60 { // Match drawing function limits
		cellSize = 60
	}
	if cellSize < 16 { // Match drawing function limits
		cellSize = 16
	}

	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	col := (int(e.Position.X) - offsetX) / cellSize
	row := (int(e.Position.Y) - offsetY) / cellSize

	if row >= 0 && row < rows && col >= 0 && col < cols {
		t.onTap(row, col)
	}
}

func (t *UnifiedTappableCanvas) TappedSecondary(*fyne.PointEvent) {}

func (t *UnifiedTappableCanvas) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// ============================================================================
// UNIFIED ARRAY DRAWING
// ============================================================================

// drawUnifiedArray draws the unified array visualization
func (ca *CircuitsApp) drawUnifiedArray(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	levels := ca.quantLevels
	arch := ca.architecture
	animStep := ca.animationStep
	ca.mu.RUnlock()

	if ca.deviceState == nil {
		return img
	}

	// Draw gradient background
	bgTop := color.RGBA{12, 20, 35, 255}
	bgBottom := color.RGBA{8, 14, 28, 255}
	drawGradientRect(img, 0, 0, w, h, bgTop, bgBottom)

	if weights == nil || len(weights) == 0 {
		return img
	}

	// Calculate margins - increased for larger peripheral boxes
	topMargin := 65    // Increased from 50 for larger DAC boxes + column labels
	rightMargin := 130 // Increased from 20 for larger TIA+ADC boxes
	bottomMargin := 30 // Slightly increased
	leftMargin := 30   // Slightly increased

	is1T1R := arch == sharedwidgets.Architecture1T1R
	is2T1R := arch == sharedwidgets.Architecture2T1R
	if is1T1R || is2T1R {
		leftMargin = 55
	}
	if is2T1R {
		bottomMargin = 55
	}

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows
	cellSize := min(cellW, cellH)
	if cellSize > 70 { // Larger cells for better state visibility
		cellSize = 70
	}
	if cellSize < 18 { // Readable minimum
		cellSize = 18
	}

	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	// Store for click detection
	ca.mu.Lock()
	ca.sharedArrayCellSize = cellSize
	ca.sharedArrayOffsetX = offsetX
	ca.sharedArrayOffsetY = offsetY
	ca.mu.Unlock()

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()

	// Draw array background panel
	panelColor := color.RGBA{18, 28, 45, 255}
	drawRoundedRect(img, offsetX-6, offsetY-6, gridW+12, gridH+12, 8, panelColor)

	// Draw BIT LINES (vertical) - color based on DAC voltage
	writeThreshold := ca.deviceState.GetWriteRange().Min
	for c := 0; c < cols; c++ {
		x := offsetX + c*cellSize + cellSize/2
		voltage := ca.deviceState.GetDACVoltage(c)

		var blCol color.RGBA
		if voltage >= writeThreshold {
			blCol = color.RGBA{255, 100, 100, 255} // Red - write voltage
		} else if voltage > 0.1 {
			blCol = color.RGBA{100, 180, 255, 255} // Blue - read/compute voltage
		} else {
			blCol = color.RGBA{50, 60, 80, 150} // Dim - no signal
		}

		// Highlight selected column
		if c == selectedCol {
			blCol.R = uint8(min(int(blCol.R)+50, 255))
			blCol.G = uint8(min(int(blCol.G)+50, 255))
			blCol.B = uint8(min(int(blCol.B)+50, 255))
		}

		for y := offsetY - 20; y < offsetY+gridH+8; y++ {
			if y >= 0 && y < h {
				img.Set(x, y, blCol)
				if cellSize > 16 {
					img.Set(x+1, y, blCol)
				}
			}
		}
	}

	// Draw WORD LINES (horizontal) - color based on active state
	for r := 0; r < rows; r++ {
		y := offsetY + r*cellSize + cellSize/2
		isActive := ca.deviceState.IsRowActive(r)

		var wlCol color.RGBA
		if isActive {
			wlCol = color.RGBA{255, 180, 100, 255} // Bright orange - active
		} else {
			wlCol = color.RGBA{60, 50, 40, 150} // Dim - inactive
		}

		// Highlight selected row
		if r == selectedRow {
			wlCol.R = uint8(min(int(wlCol.R)+30, 255))
			wlCol.G = uint8(min(int(wlCol.G)+30, 255))
		}

		startX := offsetX - 15
		if is1T1R || is2T1R {
			startX = offsetX - 8
		}

		for x := startX; x < offsetX+gridW+15; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, wlCol)
				if cellSize > 16 {
					img.Set(x, y+1, wlCol)
				}
			}
		}
	}

	// Draw 1T1R/2T1R transistors
	if is1T1R || is2T1R {
		ca.drawRowTransistors(img, offsetX, offsetY, cellSize, rows, gridH, w, h)
	}
	if is2T1R {
		ca.drawColTransistors(img, offsetX, offsetY, cellSize, cols, gridW, gridH, w, h)
	}

	// Draw signal line labels (BL = Bit Line, WL = Word Line, SL = Source Line)
	// BL label at top of grid
	drawSimpleText(img, "BL", offsetX+gridW/2-6, offsetY-35, color.RGBA{100, 180, 255, 200})
	// WL label at left of grid
	drawSimpleText(img, "WL", offsetX-25, offsetY+gridH/2-3, color.RGBA{255, 180, 100, 200})
	// SL label at bottom for 2T1R
	if is2T1R {
		drawSimpleText(img, "SL", offsetX+gridW/2-6, offsetY+gridH+45, color.RGBA{100, 220, 255, 200})
	}

	// Draw row indices on left side of array
	for r := 0; r < rows; r++ {
		y := offsetY + r*cellSize + cellSize/2 - 3
		indexColor := color.RGBA{150, 150, 170, 200}
		if r == selectedRow {
			indexColor = color.RGBA{255, 220, 100, 255} // Highlight selected row
		}
		rowText := fmt.Sprintf("%d", r)
		drawSimpleText(img, rowText, 5, y, indexColor)
	}

	// Draw column indices below array (above DAC boxes position)
	for c := 0; c < cols; c++ {
		x := offsetX + c*cellSize + cellSize/2 - 3
		indexColor := color.RGBA{150, 150, 170, 200}
		if c == selectedCol {
			indexColor = color.RGBA{255, 220, 100, 255} // Highlight selected column
		}
		colText := fmt.Sprintf("%d", c)
		drawSimpleText(img, colText, x, offsetY+gridH+5, indexColor)
	}

	// Draw cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := offsetX + c*cellSize + 2
			y0 := offsetY + r*cellSize + 2
			cw := cellSize - 4
			ch := cellSize - 4

			level := weights[r][c]
			isSelected := r == selectedRow && c == selectedCol
			isActive := ca.deviceState.IsRowActive(r) && ca.deviceState.GetDACVoltage(c) > 0.01

			// Cell color based on level - always full brightness
			cellColor := levelToColor(level, levels)

			// Animation highlight (only during compute animation)
			if animStep == 2 && isActive {
				cellColor.R = uint8(min(int(cellColor.R)+40, 255))
				cellColor.G = uint8(min(int(cellColor.G)+40, 255))
			}

			// Draw cell with 3D effect
			topColor := color.RGBA{
				uint8(min(int(cellColor.R)+35, 255)),
				uint8(min(int(cellColor.G)+35, 255)),
				uint8(min(int(cellColor.B)+35, 255)),
				255,
			}
			drawGradientRect(img, x0, y0, cw, ch, topColor, cellColor)

			// Border
			borderColor := color.RGBA{
				uint8(min(int(cellColor.R)+60, 255)),
				uint8(min(int(cellColor.G)+60, 255)),
				uint8(min(int(cellColor.B)+60, 255)),
				255,
			}
			drawRectBorder(img, x0, y0, cw, ch, borderColor)

			// Draw state number and conductance in cell if large enough
			if cellSize >= 28 {
				// Calculate text color for contrast (light on dark, dark on light)
				brightness := (int(cellColor.R) + int(cellColor.G) + int(cellColor.B)) / 3
				var textColor color.RGBA
				if brightness > 140 {
					textColor = color.RGBA{0, 0, 0, 220} // Dark text on light bg
				} else {
					textColor = color.RGBA{255, 255, 255, 220} // Light text on dark bg
				}

				// For large cells (>= 45px), show both state and conductance
				if cellSize >= 45 {
					// Calculate conductance using material model
					var conductanceUS float64
					material := ca.deviceState.GetMaterial()
					if material != nil {
						conductanceUS = material.DiscreteLevel(level, levels) * 1e6 // S to µS
					} else {
						conductanceUS = 1.0 + float64(level)/float64(levels-1)*99.0
					}

					// Draw state level number (top half of cell)
					stateText := fmt.Sprintf("S%d", level)
					textX := x0 + cw/2 - len(stateText)*3
					textY := y0 + ch/3 - 3
					drawSimpleText(img, stateText, textX, textY, textColor)

					// Draw conductance value (bottom half of cell) - dimmer
					var gText string
					if conductanceUS < 10 {
						gText = fmt.Sprintf("%.1f", conductanceUS)
					} else {
						gText = fmt.Sprintf("%.0f", conductanceUS)
					}
					gTextX := x0 + cw/2 - len(gText)*3
					gTextY := y0 + ch*2/3 - 3
					dimTextColor := color.RGBA{textColor.R, textColor.G, textColor.B, 160}
					drawSimpleText(img, gText, gTextX, gTextY, dimTextColor)
				} else {
					// For medium cells, just show state number centered
					stateText := fmt.Sprintf("%d", level)
					textX := x0 + cw/2 - len(stateText)*3
					textY := y0 + ch/2 - 3
					drawSimpleText(img, stateText, textX, textY, textColor)
				}
			}

			// C1 FIX: Selected cell highlight with bright contrasting border
			if isSelected {
				// Bold yellow/gold border (3px thick)
				highlightColor := color.RGBA{255, 200, 0, 255}
				drawRectBorder(img, x0-1, y0-1, cw+2, ch+2, highlightColor)
				drawRectBorder(img, x0-2, y0-2, cw+4, ch+4, highlightColor)
				drawRectBorder(img, x0-3, y0-3, cw+6, ch+6, highlightColor)
				// Subtle white outer glow
				drawRectBorder(img, x0-4, y0-4, cw+8, ch+8, color.RGBA{255, 255, 255, 180})
			}
		}
	}

	// Draw DAC boxes (top) - larger for better visibility
	dacBoxH := 30  // Increased from 25
	dacBoxW := cellSize - 2
	if dacBoxW < 35 { // Increased from 24
		dacBoxW = 35
	}
	dacY := offsetY - dacBoxH - 18 // More spacing

	for c := 0; c < min(8, cols); c++ {
		dacX := offsetX + c*cellSize + 1
		voltage := ca.deviceState.GetDACVoltage(c)
		highlighted := animStep == 1
		// Show column number as label for clarity
		colLabel := fmt.Sprintf("C%d", c)
		drawDACColumn(img, dacX, dacY, dacBoxW, dacBoxH, voltage, colLabel, highlighted, false)
	}

	// Draw TIA+ADC boxes (right side) - larger for better visibility
	tiaBoxW := 70 // Increased from 50 for more data
	adcBoxW := 30 // Increased from 24
	tiaAdcBoxH := cellSize - 2
	if tiaAdcBoxH < 24 { // Increased from 18
		tiaAdcBoxH = 24
	}
	tiaX := offsetX + gridW + 12

	for r := 0; r < min(8, rows); r++ {
		tiaY := offsetY + r*cellSize + 1
		current := ca.deviceState.GetRowCurrent(r)
		level := ca.deviceState.GetRowLevel(r)
		highlighted := animStep == 3
		dimmed := !ca.deviceState.IsRowActive(r)
		// Show row number as label for clarity
		rowLabel := fmt.Sprintf("R%d", r)
		drawTIAADCRow(img, tiaX, tiaY, tiaBoxW, adcBoxW, tiaAdcBoxH, current, level, rowLabel, highlighted, dimmed, ca.tia, ca.adc)
	}

	// Draw labels
	drawSimpleText(img, "DAC", offsetX-25, dacY+dacBoxH/2-3, color.RGBA{170, 140, 220, 255})
	drawSimpleText(img, "TIA", tiaX, offsetY-10, color.RGBA{220, 180, 100, 255})
	drawSimpleText(img, "ADC", tiaX+tiaBoxW+4, offsetY-10, color.RGBA{130, 210, 170, 255})

	// Operation classification title with prominent badge
	opText := ca.deviceState.ClassifyOperation()
	var opColor, opBgColor color.RGBA
	switch {
	case opText == "WRITE":
		opColor = color.RGBA{255, 200, 100, 255}
		opBgColor = color.RGBA{80, 60, 30, 200}
	case opText == "READ":
		opColor = color.RGBA{100, 220, 255, 255}
		opBgColor = color.RGBA{30, 60, 80, 200}
	case opText == "COMPUTE (MVM)":
		opColor = color.RGBA{200, 150, 255, 255}
		opBgColor = color.RGBA{50, 40, 80, 200}
	default:
		opColor = color.RGBA{150, 150, 150, 255}
		opBgColor = color.RGBA{40, 40, 50, 200}
	}
	// Draw background badge for operation mode
	opBoxW := len(opText)*6 + 12
	drawRoundedRect(img, 5, 3, opBoxW, 16, 4, opBgColor)
	drawSimpleText(img, opText, 10, 8, opColor)

	// Architecture badge
	var archText string
	var archColor color.RGBA
	switch arch {
	case sharedwidgets.Architecture2T1R:
		archText = "2T1R"
		archColor = color.RGBA{100, 180, 220, 255}
	case sharedwidgets.Architecture1T1R:
		archText = "1T1R"
		archColor = color.RGBA{100, 220, 120, 255}
	default:
		archText = "PASSIVE"
		archColor = color.RGBA{220, 150, 100, 255}
	}
	drawSimpleText(img, archText, w-len(archText)*6-10, 8, archColor)

	// Draw energy/timing info in top-right corner (below architecture badge)
	mode := ca.deviceState.GetOperationMode()
	var energyText, timingText string
	var energyColor color.RGBA
	switch mode {
	case OpModeRead:
		energyText = "~45fJ"
		timingText = "65ns"
		energyColor = color.RGBA{100, 200, 255, 200}
	case OpModeWrite:
		energyText = "~55fJ"
		timingText = "170ns"
		energyColor = color.RGBA{255, 180, 100, 200}
	case OpModeCompute:
		// MVM energy scales with active cells
		activeCells := 0
		activeRows := 0
		activeCols := 0
		for r := 0; r < rows; r++ {
			if ca.deviceState.IsRowActive(r) {
				activeRows++
				for c := 0; c < cols; c++ {
					if ca.deviceState.GetDACVoltage(c) > 0.01 {
						activeCells++
					}
				}
			}
		}
		for c := 0; c < cols; c++ {
			if ca.deviceState.GetDACVoltage(c) > 0.01 {
				activeCols++
			}
		}
		energyFJ := 50 * activeCells // ~50fJ per cell
		energyText = fmt.Sprintf("~%dfJ", energyFJ)
		timingText = fmt.Sprintf("%d×%d=%d", activeRows, activeCols, activeCells)
		energyColor = color.RGBA{200, 150, 255, 200}
	default:
		energyText = ""
		timingText = ""
	}
	if energyText != "" {
		drawSimpleText(img, energyText, w-50, 22, energyColor)
		drawSimpleText(img, timingText, w-50, 34, color.RGBA{150, 150, 170, 180})
	}

	// Draw sneak path indicators for passive (0T1R) mode
	// Sneak currents flow through half-selected cells, causing read errors
	if arch == sharedwidgets.Architecture0T1R && ca.deviceState != nil {
		sneakColor := color.RGBA{255, 100, 100, 100} // Semi-transparent red for sneak paths
		selectedRow := ca.deviceState.GetSelectedRow()
		selectedCol := ca.deviceState.GetSelectedCol()
		voltage := ca.deviceState.GetDACVoltage(selectedCol)

		// Only show sneak paths when there's active voltage and in READ mode
		if voltage > 0.05 && ca.deviceState.GetOperationMode() == OpModeRead {
			// Draw diagonal sneak path indicators (L-shaped paths through half-selected cells)
			for r := 0; r < rows; r++ {
				if r == selectedRow {
					continue // Skip the target row
				}
				for c := 0; c < cols; c++ {
					if c == selectedCol {
						continue // Skip the target column
					}
					// Draw small diagonal indicator in corner of sneak path cells
					cx := offsetX + c*cellSize + cellSize - 6
					cy := offsetY + r*cellSize + 4
					// Draw small "S" marker for sneak path
					for dx := 0; dx < 4; dx++ {
						for dy := 0; dy < 4; dy++ {
							if (dx+dy)%2 == 0 { // Dashed pattern
								if cx+dx >= 0 && cx+dx < w && cy+dy >= 0 && cy+dy < h {
									img.Set(cx+dx, cy+dy, sneakColor)
								}
							}
						}
					}
				}
			}
			// Draw "SNEAK PATHS" warning label
			drawSimpleText(img, "0T1R: Sneak paths active", 10, h-15, color.RGBA{255, 150, 100, 200})
		}
	}

	// Draw current flow indicators during active operation
	if animStep >= 2 {
		// Draw current flow arrows on active bit lines (columns with voltage)
		for c := 0; c < cols; c++ {
			voltage := ca.deviceState.GetDACVoltage(c)
			if voltage > 0.05 {
				x := offsetX + c*cellSize + cellSize/2
				// Draw downward current arrow (electrons flow opposite to current)
				arrowColor := color.RGBA{100, 255, 150, 200}
				// Draw arrow shaft
				for y := offsetY - 10; y < offsetY+gridH+5; y += 8 {
					if y >= 0 && y < h {
						img.Set(x, y, arrowColor)
						img.Set(x-1, y, arrowColor)
						img.Set(x+1, y, arrowColor)
					}
				}
			}
		}

		// Draw current collection arrows flowing to TIA (horizontal on active rows)
		for r := 0; r < rows; r++ {
			if ca.deviceState.IsRowActive(r) {
				current := ca.deviceState.GetRowCurrent(r)
				if current > 0.1 {
					y := offsetY + r*cellSize + cellSize/2
					// Arrow intensity based on current magnitude
					intensity := uint8(min(int(current*2), 200))
					arrowColor := color.RGBA{255, 200, intensity, 180}
					// Draw rightward arrow to TIA
					for x := offsetX + gridW + 2; x < tiaX-2; x += 4 {
						if x >= 0 && x < w {
							img.Set(x, y, arrowColor)
							img.Set(x, y-1, arrowColor)
							img.Set(x, y+1, arrowColor)
						}
					}
				}
			}
		}
	}

	// Draw info badge showing selected cell's expected current
	if ca.deviceState != nil {
		selectedRow := ca.deviceState.GetSelectedRow()
		selectedCol := ca.deviceState.GetSelectedCol()
		voltage := ca.deviceState.GetDACVoltage(selectedCol)

		ca.mu.RLock()
		var level int
		if selectedRow < len(weights) && selectedCol < len(weights[selectedRow]) {
			level = weights[selectedRow][selectedCol]
		}
		ca.mu.RUnlock()

		if voltage > 0.05 && level > 0 {
			// Calculate expected current for selected cell
			var conductanceUS float64
			material := ca.deviceState.GetMaterial()
			if material != nil {
				conductanceUS = material.DiscreteLevel(level, levels) * 1e6
			} else {
				conductanceUS = 1.0 + float64(level)/float64(levels-1)*99.0
			}
			expectedCurrent := conductanceUS * voltage

			// Draw info near selected cell
			cellX := offsetX + selectedCol*cellSize + cellSize/2
			cellY := offsetY + selectedRow*cellSize - 12
			if cellY > 20 {
				infoText := fmt.Sprintf("%.1fuA", expectedCurrent)
				drawSimpleText(img, infoText, cellX-len(infoText)*3, cellY, color.RGBA{255, 255, 100, 220})
			}
		}
	}

	// Draw color legend in bottom-left corner
	legendY := h - 60
	legendX := 5

	// Title
	drawSimpleText(img, "Legend:", legendX, legendY, color.RGBA{200, 200, 220, 255})

	// Cell conductance gradient: Low G → High G
	legendY += 12
	drawSimpleText(img, "G:", legendX, legendY, color.RGBA{150, 150, 170, 200})
	// Draw gradient boxes
	boxW := 12
	for i := 0; i < 5; i++ {
		level := i * (levels - 1) / 4
		c := levelToColor(level, levels)
		drawRect(img, legendX+15+i*boxW, legendY-2, boxW-2, 10, c)
	}
	drawSimpleText(img, "Lo", legendX+15, legendY+10, color.RGBA{100, 150, 255, 200})
	drawSimpleText(img, "Hi", legendX+15+4*boxW, legendY+10, color.RGBA{255, 100, 100, 200})

	// DAC voltage zones
	legendY += 22
	drawSimpleText(img, "V:", legendX, legendY, color.RGBA{150, 150, 170, 200})
	// Read zone - blue
	drawRect(img, legendX+15, legendY-2, boxW, 10, color.RGBA{60, 140, 200, 255})
	drawSimpleText(img, "R", legendX+15+2, legendY+10, color.RGBA{100, 180, 255, 200})
	// Caution zone - yellow
	drawRect(img, legendX+15+boxW+2, legendY-2, boxW, 10, color.RGBA{200, 180, 60, 255})
	drawSimpleText(img, "!", legendX+15+boxW+4, legendY+10, color.RGBA{255, 220, 100, 200})
	// Write zone - red
	drawRect(img, legendX+15+2*(boxW+2), legendY-2, boxW, 10, color.RGBA{220, 100, 60, 255})
	drawSimpleText(img, "W", legendX+15+2*(boxW+2)+2, legendY+10, color.RGBA{255, 140, 100, 200})

	// Draw operation hint in bottom-right corner
	hintY := h - 20
	hintX := w - 200
	var hintText string
	hintColor := color.RGBA{120, 140, 160, 200}
	switch ca.deviceState.GetOperationMode() {
	case OpModeRead:
		hintText = "READ: Sense cell conductance"
		hintColor = color.RGBA{100, 180, 220, 200}
	case OpModeWrite:
		hintText = "WRITE: Program cell state"
		hintColor = color.RGBA{220, 160, 80, 200}
	case OpModeCompute:
		hintText = "MVM: y = W * x"
		hintColor = color.RGBA{180, 140, 220, 200}
	}
	if hintText != "" {
		drawSimpleText(img, hintText, hintX, hintY, hintColor)
	}

	return img
}

// drawRowTransistors draws the row transistors for 1T1R/2T1R architecture
// Enhanced with clearer MOSFET symbols and ON/OFF indicators
func (ca *CircuitsApp) drawRowTransistors(img *image.RGBA, offsetX, offsetY, cellSize, rows, gridH, w, h int) {
	for r := 0; r < rows; r++ {
		ty := offsetY + r*cellSize + cellSize/2
		tx := offsetX - 35 // Moved left for larger symbol

		transistorOn := ca.deviceState.IsRowActive(r)

		var bodyCol, gateCol, channelCol, terminalCol color.RGBA
		if transistorOn {
			bodyCol = color.RGBA{60, 200, 80, 255}    // Green body when ON
			gateCol = color.RGBA{100, 255, 120, 255}  // Bright green gate
			channelCol = color.RGBA{80, 220, 100, 255}
			terminalCol = color.RGBA{150, 255, 150, 255}
		} else {
			bodyCol = color.RGBA{60, 60, 70, 255}
			gateCol = color.RGBA{90, 90, 100, 255}
			channelCol = color.RGBA{50, 50, 60, 255}
			terminalCol = color.RGBA{100, 100, 110, 255}
		}

		// Draw MOSFET body (larger, 8x12 rectangle)
		for dy := -6; dy <= 6; dy++ {
			for dx := 0; dx < 5; dx++ {
				px, py := tx+dx, ty+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, bodyCol)
				}
			}
		}

		// Draw gate (thicker, 2px wide)
		gateX := tx - 4
		for dy := -8; dy <= 8; dy++ {
			py := ty + dy
			if gateX >= 0 && gateX+1 < w && py >= 0 && py < h {
				img.Set(gateX, py, gateCol)
				img.Set(gateX+1, py, gateCol)
			}
		}

		// Draw source terminal (top)
		for dx := -2; dx <= 2; dx++ {
			py := ty - 8
			px := tx + 2 + dx
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px, py-1, terminalCol)
			}
		}

		// Draw drain terminal (bottom)
		for dx := -2; dx <= 2; dx++ {
			py := ty + 8
			px := tx + 2 + dx
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px, py+1, terminalCol)
			}
		}

		// Draw channel (connecting to array)
		for dx := 5; dx < 25; dx++ {
			px := tx + dx
			if px >= 0 && px < w {
				img.Set(px, ty, channelCol)
				if transistorOn {
					img.Set(px, ty+1, channelCol) // Thicker when ON
				}
			}
		}

		// ON/OFF indicator with label
		if transistorOn {
			drawGlowCircle(img, tx+2, ty, 3, color.RGBA{150, 255, 150, 255}, color.RGBA{100, 200, 100, 100})
		} else {
			// Draw X for OFF state
			for d := -2; d <= 2; d++ {
				px1, py1 := tx+2+d, ty+d
				px2, py2 := tx+2+d, ty-d
				if px1 >= 0 && px1 < w && py1 >= 0 && py1 < h {
					img.Set(px1, py1, color.RGBA{150, 80, 80, 200})
				}
				if px2 >= 0 && px2 < w && py2 >= 0 && py2 < h {
					img.Set(px2, py2, color.RGBA{150, 80, 80, 200})
				}
			}
		}
	}
}

// drawColTransistors draws the column transistors for 2T1R architecture
// Enhanced with clearer MOSFET symbols (horizontal orientation)
func (ca *CircuitsApp) drawColTransistors(img *image.RGBA, offsetX, offsetY, cellSize, cols, gridW, gridH, w, h int) {
	for c := 0; c < cols; c++ {
		tx := offsetX + c*cellSize + cellSize/2
		ty := offsetY + gridH + 25 // Moved down slightly

		// In 2T1R, column transistors are controlled by CSL
		// For simplicity, all column transistors are ON when computing
		transistorOn := ca.deviceState.GetWLMode() == WLAll || c == ca.deviceState.GetSelectedCol()

		var bodyCol, gateCol, channelCol, terminalCol color.RGBA
		if transistorOn {
			bodyCol = color.RGBA{60, 180, 200, 255}    // Cyan body when ON
			gateCol = color.RGBA{100, 220, 255, 255}   // Bright cyan gate
			channelCol = color.RGBA{80, 200, 220, 255}
			terminalCol = color.RGBA{150, 230, 255, 255}
		} else {
			bodyCol = color.RGBA{60, 60, 70, 255}
			gateCol = color.RGBA{90, 90, 100, 255}
			channelCol = color.RGBA{50, 50, 60, 255}
			terminalCol = color.RGBA{100, 100, 110, 255}
		}

		// Draw MOSFET body (horizontal, larger)
		for dx := -6; dx <= 6; dx++ {
			for dy := 0; dy < 5; dy++ {
				px, py := tx+dx, ty+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, bodyCol)
				}
			}
		}

		// Draw gate (thicker)
		gateY := ty + 7
		for dx := -8; dx <= 8; dx++ {
			px := tx + dx
			if px >= 0 && px < w && gateY >= 0 && gateY+1 < h {
				img.Set(px, gateY, gateCol)
				img.Set(px, gateY+1, gateCol)
			}
		}

		// Draw left terminal
		for dy := -2; dy <= 2; dy++ {
			px := tx - 8
			py := ty + 2 + dy
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px-1, py, terminalCol)
			}
		}

		// Draw right terminal
		for dy := -2; dy <= 2; dy++ {
			px := tx + 8
			py := ty + 2 + dy
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px+1, py, terminalCol)
			}
		}

		// Draw channel (connecting to array above)
		for dy := -20; dy < 0; dy++ {
			py := ty + dy
			if tx >= 0 && tx < w && py >= 0 && py < h {
				img.Set(tx, py, channelCol)
				if transistorOn {
					img.Set(tx+1, py, channelCol) // Thicker when ON
				}
			}
		}

		// ON/OFF indicator
		if transistorOn {
			drawGlowCircle(img, tx, ty+2, 3, color.RGBA{150, 230, 255, 255}, color.RGBA{100, 180, 200, 100})
		} else {
			// Draw X for OFF state
			for d := -2; d <= 2; d++ {
				px1, py1 := tx+d, ty+2+d
				px2, py2 := tx+d, ty+2-d
				if px1 >= 0 && px1 < w && py1 >= 0 && py1 < h {
					img.Set(px1, py1, color.RGBA{150, 80, 80, 200})
				}
				if px2 >= 0 && px2 < w && py2 >= 0 && py2 < h {
					img.Set(px2, py2, color.RGBA{150, 80, 80, 200})
				}
			}
		}
	}
}

// ============================================================================
// EVENT HANDLERS
// ============================================================================

// onDACVoltageChanged handles DAC voltage input changes
func (ca *CircuitsApp) onDACVoltageChanged(col int, voltageStr string) {
	voltage, err := strconv.ParseFloat(voltageStr, 64)
	if err != nil {
		return
	}

	// Clamp voltage to reasonable range
	if voltage < 0 {
		voltage = 0
	}
	if voltage > 2.0 {
		voltage = 2.0
	}

	ca.deviceState.SetDACVoltage(col, voltage)
	ca.recomputeAndRefresh()
}

// setUnifiedDACPreset applies a DAC preset (called by mode changes)
func (ca *CircuitsApp) setUnifiedDACPreset(preset DACMode) {
	switch preset {
	case DACReadPreset:
		ca.deviceState.SetDACPreset(DACReadPreset)
	case DACWritePreset:
		ca.deviceState.SetDACPreset(DACWritePreset)
	}
	ca.updateDACRangeModeLabel()
	ca.updateDACEntries()
	ca.recomputeAndRefresh()
}

// setAllUnifiedDACVoltages sets all DAC voltages to the same value
func (ca *CircuitsApp) setAllUnifiedDACVoltages(voltageStr string) {
	voltage, err := strconv.ParseFloat(voltageStr, 64)
	if err != nil {
		return
	}
	ca.deviceState.SetAllDACVoltages(voltage)
	ca.updateDACEntries()
	ca.recomputeAndRefresh()
}

// onWLChanged is a no-op - WL checkboxes have been removed
// Row selection is now done via cell clicks in onUnifiedCellTapped
func (ca *CircuitsApp) onWLChanged(row int, checked bool) {
	// No-op: WL checkboxes removed from UI
}

// setWLModeSingle sets WL mode to single (only selected row)
func (ca *CircuitsApp) setWLModeSingle() {
	selectedRow := ca.deviceState.GetSelectedRow()
	ca.deviceState.SetWLSingle(selectedRow)
	ca.updateWLCheckboxes()
	ca.recomputeAndRefresh()
}

// setWLModeAll sets WL mode to all rows active
func (ca *CircuitsApp) setWLModeAll() {
	ca.deviceState.SetWLAll()
	ca.updateWLCheckboxes()
	ca.recomputeAndRefresh()
}

// onUnifiedCellTapped handles cell selection
// In READ/WRITE mode: selects row transistor (WL) and column transistor (BL)
func (ca *CircuitsApp) onUnifiedCellTapped(row, col int) {
	ca.deviceState.SetSelectedCell(row, col)

	mode := ca.deviceState.GetOperationMode()
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R

	// In READ/WRITE mode (non-passive): select ONLY this row (single transistor)
	if !isPassive && (mode == OpModeRead || mode == OpModeWrite) {
		ca.deviceState.SetWLSingle(row)
	}

	// Update target cell label in write mode panel
	ca.updateWriteTargetLabel()

	// Cell click only selects the cell - does NOT apply voltages
	// Voltages are only applied when user presses action buttons

	ca.recomputeAndRefresh()
	ca.updateCellInfo()
}

// ============================================================================
// ACTION HANDLERS
// ============================================================================

// onUnifiedProgram programs the selected cell using Write-ReadVerify loop
// This simulates ISPP (Incremental Step Pulse Programming) behavior
func (ca *CircuitsApp) onUnifiedProgram() {
	// Mode validation: only allowed in WRITE mode
	if ca.deviceState.GetOperationMode() != OpModeWrite {
		ca.operationsStatusLabel.SetText("Error: Switch to WRITE mode first")
		return
	}

	// Get target level directly from slider (the user's intent)
	targetLevel := int(ca.mfuxWriteLevelSlider.Value)
	if targetLevel < 0 {
		targetLevel = 0
	}
	if targetLevel >= ca.quantLevels {
		targetLevel = ca.quantLevels - 1
	}

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()

	// H3 FIX: Save current state to undo history before modifying
	ca.saveUndoHistory()

	// Run Write-ReadVerify loop in background goroutine
	go ca.runISPPWithAnimation(selectedRow, selectedCol, targetLevel)
}

// writeReadVerifyLoop performs animated Write-ReadVerify iterations
// Simulates ISPP: apply pulse, read back, adjust if needed, repeat until target reached
func (ca *CircuitsApp) writeReadVerifyLoop(row, col, targetLevel int, startVoltage float64) {
	const maxIterations = 5
	const iterationDelay = 300 * time.Millisecond

	writeRange := ca.deviceState.GetWriteRange()
	voltage := startVoltage
	currentLevel := 0

	// Get current level
	ca.mu.Lock()
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		currentLevel = ca.arrayWeights[row][col]
	}
	ca.mu.Unlock()

	// Check if using V/2 half-select (passive mode)
	isPassive := ca.deviceState.IsPassiveMode()

	for iteration := 1; iteration <= maxIterations; iteration++ {
		// === WRITE PHASE ===
		if isPassive {
			// V/2 half-select scheme for passive (0T1R) mode
			// WL = +V/2, BL = -V/2, target sees full V, half-selected see V/2
			halfV := voltage / 2.0
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("WRITE [%d,%d]: V/2 scheme WL=+%.2fV BL=-%.2fV (iter %d/%d)",
					row, col, halfV, halfV, iteration, maxIterations))
			})
			ca.deviceState.ApplyHalfSelectWrite(row, col, voltage)
		} else {
			// 1T1R/2T1R: transistor isolation, full voltage on BL
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("WRITE [%d,%d]: V=%.2fV (iter %d/%d)", row, col, voltage, iteration, maxIterations))
			})
			ca.deviceState.SetDACVoltage(col, voltage)
		}
		ca.recomputeAndRefresh()
		time.Sleep(iterationDelay / 2)

		// Simulate write: move current level toward target
		// In real hardware, the level change depends on voltage amplitude
		ca.mu.Lock()
		if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
			if currentLevel < targetLevel {
				// Increase by 1-2 levels per pulse (simulated partial switching)
				step := 1
				if targetLevel-currentLevel > 3 {
					step = 2
				}
				currentLevel += step
				if currentLevel > targetLevel {
					currentLevel = targetLevel
				}
			} else if currentLevel > targetLevel {
				// Decrease by 1-2 levels per pulse
				step := 1
				if currentLevel-targetLevel > 3 {
					step = 2
				}
				currentLevel -= step
				if currentLevel < targetLevel {
					currentLevel = targetLevel
				}
			}
			ca.arrayWeights[row][col] = currentLevel
		}
		ca.mu.Unlock()

		// === READ/VERIFY PHASE ===
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("VERIFY [%d,%d]: Read level %d (target %d)", row, col, currentLevel, targetLevel))
		})

		// Reset write voltages before applying read voltage
		// This clears V/2 biasing and puts array in safe read state
		ca.deviceState.ResetWriteVoltages()

		// Set DAC to read voltage for verification
		readVoltage := ca.deviceState.GetReadRange().Max * 0.5
		ca.deviceState.SetDACVoltage(col, readVoltage)
		ca.recomputeAndRefresh()
		time.Sleep(iterationDelay / 2)

		// Check if target reached
		// Write cycle: ~170ns per iteration, Energy: ~100fJ (pulse + verify)
		if currentLevel == targetLevel {
			totalTimeNs := iteration * 170
			totalEnergyFJ := iteration * 100
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("WRITE [%d,%d] = State %d | %d iter | ~%dns, ~%dfJ",
					row, col, targetLevel, iteration, totalTimeNs, totalEnergyFJ))
			})
			// Return all voltages to 0 (safe idle state)
			ca.deviceState.ResetWriteVoltages()
			ca.recomputeAndRefresh()
			return
		}

		// Adjust voltage for next iteration (ISPP: increment voltage if undershoot)
		if currentLevel < targetLevel {
			// Need higher voltage to switch more domains
			voltageStep := (writeRange.Max - writeRange.Min) / float64(ca.quantLevels*2)
			voltage += voltageStep
			if voltage > writeRange.Max {
				voltage = writeRange.Max
			}
		} else {
			// Need lower voltage (less aggressive write)
			voltageStep := (writeRange.Max - writeRange.Min) / float64(ca.quantLevels*2)
			voltage -= voltageStep
			if voltage < writeRange.Min {
				voltage = writeRange.Min
			}
		}
	}

	// Max iterations reached
	fyne.Do(func() {
		ca.operationsStatusLabel.SetText(fmt.Sprintf("PARTIAL [%d,%d] = State %d (target was %d, max iterations)",
			row, col, currentLevel, targetLevel))
	})
	// Return all voltages to 0 (safe idle state)
	ca.deviceState.ResetWriteVoltages()
	ca.recomputeAndRefresh()
}

// onUnifiedRead senses the selected cell per VOLTAGE_RULES.md:
// - Selected WL: Active (1T1R) or read voltage (0T1R)
// - Selected BL: Read voltage (0.1-0.5V)
// - Unselected BLs: 0V (grounded to minimize sneak paths)
func (ca *CircuitsApp) onUnifiedRead() {
	// Mode validation: only allowed in READ mode
	if ca.deviceState.GetOperationMode() != OpModeRead {
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Error: Switch to READ mode first")
		})
		return
	}

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()

	// Per VOLTAGE_RULES.md: Only selected row active
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R
	if !isPassive {
		ca.deviceState.SetWLSingle(selectedRow)
	}

	// Per VOLTAGE_RULES.md Section 3.1 and 4.1:
	// - Selected BL: Read voltage (0.2V typical)
	// - Unselected BLs: 0V (ground)
	readRange := ca.deviceState.GetReadRange()
	readVoltage := readRange.Max * 0.4 // ~0.2V for safe read
	if readVoltage < 0.1 {
		readVoltage = 0.2
	}

	// Ground all columns first, then apply read voltage only to selected column
	ca.deviceState.SetAllDACVoltages(0)
	ca.deviceState.SetDACVoltage(selectedCol, readVoltage)
	ca.deviceState.SetDACRangeMode(DACRangeRead)

	// Recompute with proper biasing
	ca.recomputeAndRefresh()

	// Get results for the selected cell
	current := ca.deviceState.GetRowCurrent(selectedRow)
	tiaVoltage := ca.deviceState.GetRowVoltage(selectedRow)
	adcLevel := ca.deviceState.GetRowLevel(selectedRow)

	// Get the cell's conductance for display
	ca.mu.RLock()
	level := 0
	if selectedRow < len(ca.arrayWeights) && selectedCol < len(ca.arrayWeights[selectedRow]) {
		level = ca.arrayWeights[selectedRow][selectedCol]
	}
	ca.mu.RUnlock()

	// Update status with single-cell sense result including energy/timing
	// Read cycle: ~65ns, Energy: DAC(15fJ) + TIA(5fJ) + ADC(25fJ) = ~45fJ
	fyne.Do(func() {
		ca.operationsStatusLabel.SetText(fmt.Sprintf("READ [%d,%d]: State=%d | I=%.1fµA → TIA=%.2fV → ADC=%d | ~65ns, ~45fJ",
			selectedRow, selectedCol, level, current, tiaVoltage, adcLevel))
	})
}

// onUnifiedCompute runs MVM computation with current input vector
func (ca *CircuitsApp) onUnifiedCompute() {
	// Mode validation: only allowed in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		ca.operationsStatusLabel.SetText("Error: Switch to COMPUTE mode first")
		return
	}

	// Ensure all rows are active for MVM
	ca.deviceState.SetWLAll()
	ca.updateWLCheckboxes()

	// Apply input vector to DAC (convert 0-255 to read range voltages)
	ca.mu.RLock()
	params := make([]float64, len(ca.inputVector))
	for i, v := range ca.inputVector {
		params[i] = float64(v)
	}
	ca.mu.RUnlock()
	ca.deviceState.SetDACPreset(DACInputVector, params...)

	ca.recomputeAndRefresh()

	// Save compute log for debugging
	// MVM: ~20ns (parallel), Energy: N × ~50fJ where N = active cells
	activeCells := ca.arrayRows * ca.arrayCols
	energyFJ := activeCells * 50 // ~50fJ per cell
	if err := SaveComputeLog(); err != nil {
		ca.operationsStatusLabel.SetText(fmt.Sprintf("MVM done (log error: %v)", err))
	} else {
		ca.operationsStatusLabel.SetText(fmt.Sprintf("MVM complete: %dx%d array | ~20ns, ~%dfJ total | saved log",
			ca.arrayRows, ca.arrayCols, energyFJ))
	}
}

// onUnifiedAnimate animates the signal flow
func (ca *CircuitsApp) onUnifiedAnimate() {
	ca.mu.Lock()
	ca.animationActive = true
	ca.mu.Unlock()

	ca.operationsStatusLabel.SetText("Animating...")

	go func() {
		// Step 1: DAC
		if ca.shouldStop() {
			return
		}
		ca.mu.Lock()
		ca.animationStep = 1
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 1: DAC conversion (5ns)")
		})
		if ca.sleep(600) {
			return // Interrupted
		}

		// Step 2: Array
		if ca.shouldStop() {
			return
		}
		ca.mu.Lock()
		ca.animationStep = 2
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 2: Array MVM (5ns)")
		})
		if ca.sleep(600) {
			return // Interrupted
		}

		// Step 3: ADC
		if ca.shouldStop() {
			return
		}
		ca.mu.Lock()
		ca.animationStep = 3
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 3: TIA+ADC conversion (10ns)")
		})
		if ca.sleep(600) {
			return // Interrupted
		}

		// Complete
		ca.mu.Lock()
		ca.animationStep = 0
		ca.animationActive = false
		ca.mu.Unlock()
		ca.recomputeAndRefresh()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Complete in ~20ns")
		})
	}()
}

// onUnifiedReset resets the array to random values
func (ca *CircuitsApp) onUnifiedReset() {
	// Clear undo history on reset (per code review recommendation)
	ca.mu.Lock()
	ca.undoHistory = nil
	ca.hasUndoHistory = false
	ca.mu.Unlock()

	fyne.Do(func() {
		if ca.undoHistoryBtn != nil {
			ca.undoHistoryBtn.Disable()
		}
	})

	// Reset DAC to read preset (uses material-derived voltage range)
	ca.deviceState.SetDACPreset(DACReadPreset)
	ca.updateDACRangeModeLabel()
	ca.updateDACEntries()

	// Reset WL based on operation mode (only in 1T1R/2T1R - passive keeps all on)
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R
	if !isPassive {
		if ca.deviceState.GetOperationMode() == OpModeCompute {
			ca.deviceState.SetWLAll() // COMPUTE needs all rows for MVM
		} else {
			ca.deviceState.SetWLSingle(0)
		}
	}
	ca.updateWLCheckboxes()

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Reset complete")
}

// onUnifiedRandomArray randomizes the array weights
func (ca *CircuitsApp) onUnifiedRandomArray() {
	// H3 FIX: Save current state to undo history before modifying
	ca.saveUndoHistory()

	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = rand.Intn(ca.quantLevels)
		}
	}
	ca.mu.Unlock()

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Array randomized")
}

// H3 FIX: saveUndoHistory saves the current array state for undo
func (ca *CircuitsApp) saveUndoHistory() {
	ca.mu.Lock()
	// Create a deep copy of current array state
	ca.undoHistory = make([][]int, len(ca.arrayWeights))
	for i := range ca.arrayWeights {
		ca.undoHistory[i] = make([]int, len(ca.arrayWeights[i]))
		copy(ca.undoHistory[i], ca.arrayWeights[i])
	}
	ca.hasUndoHistory = true
	ca.mu.Unlock() // Release lock before UI update to avoid deadlock

	// Enable undo button
	fyne.Do(func() {
		if ca.undoHistoryBtn != nil {
			ca.undoHistoryBtn.Enable()
		}
	})
}

// H3 FIX: onUndo restores the previous array state
func (ca *CircuitsApp) onUndo() {
	ca.mu.Lock()
	if !ca.hasUndoHistory || ca.undoHistory == nil {
		ca.mu.Unlock()
		return
	}

	// Restore array from history with defensive length check
	for i := range ca.arrayWeights {
		if i < len(ca.undoHistory) && len(ca.arrayWeights[i]) == len(ca.undoHistory[i]) {
			copy(ca.arrayWeights[i], ca.undoHistory[i])
		}
	}

	// Clear history (single-level undo only)
	ca.undoHistory = nil
	ca.hasUndoHistory = false
	ca.mu.Unlock() // Release lock before UI updates to avoid deadlock

	// Disable undo button
	fyne.Do(func() {
		if ca.undoHistoryBtn != nil {
			ca.undoHistoryBtn.Disable()
		}
		if ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText("Undo complete")
		}
	})

	ca.recomputeAndRefresh()
}

// ============================================================================
// UI UPDATE HELPERS
// ============================================================================

// recomputeAndRefresh runs computation and updates all UI elements
func (ca *CircuitsApp) recomputeAndRefresh() {
	ca.mu.RLock()
	weights := ca.arrayWeights
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Run device simulation
	ca.deviceState.Compute(weights, levels)

	// Update output display
	ca.updateOutputDisplay()

	// Update cell info
	ca.updateCellInfo()

	// Update operation classification
	ca.updateOperationClassification()

	// Refresh array canvas
	ca.refreshUnifiedArray()
}

// refreshUnifiedArray refreshes the array canvas
func (ca *CircuitsApp) refreshUnifiedArray() {
	if ca.sharedArrayCanvas != nil {
		fyne.Do(func() {
			ca.sharedArrayCanvas.Refresh()
		})
	}
}

// updateDACEntries is a no-op - DAC values are shown on the diagram
func (ca *CircuitsApp) updateDACEntries() {
	// DAC values are displayed in the array diagram's DAC boxes
}

// updateWLCheckboxes is a no-op - WL checkboxes have been removed
// WL state is now managed automatically based on mode and architecture
func (ca *CircuitsApp) updateWLCheckboxes() {
	// No-op: WL checkboxes removed from UI
	// Row selection is done via cell clicks
}

// updateWLCheckboxesForArchitecture is a no-op - WL UI has been removed
func (ca *CircuitsApp) updateWLCheckboxesForArchitecture() {
	// No-op: WL UI removed
}

// updateOutputDisplay is a no-op - outputs are shown on the diagram
func (ca *CircuitsApp) updateOutputDisplay() {
	// Outputs are displayed in the array diagram's TIA/ADC boxes
}

// updateCellInfo updates the cell info display with detailed circuit data
func (ca *CircuitsApp) updateCellInfo() {
	if ca.sharedCellInfoLabel == nil {
		return
	}

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()

	ca.mu.RLock()
	var level int
	if selectedRow < len(ca.arrayWeights) && selectedCol < len(ca.arrayWeights[selectedRow]) {
		level = ca.arrayWeights[selectedRow][selectedCol]
	}
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Use material's physics-based conductance calculation
	var conductanceUS float64
	material := ca.deviceState.GetMaterial()
	if material != nil {
		conductanceUS = material.DiscreteLevel(level, levels) * 1e6 // S to µS
	} else {
		conductanceUS = 1.0 + float64(level)/float64(levels-1)*99.0
	}

	voltage := ca.deviceState.GetDACVoltage(selectedCol)
	matName := ca.deviceState.GetMaterialName()

	// Calculate expected current I = G × V
	expectedCurrent := conductanceUS * voltage // µA

	// Get actual row output (includes all cells in row if active)
	rowCurrent := ca.deviceState.GetRowCurrent(selectedRow)
	rowVoltage := ca.deviceState.GetRowVoltage(selectedRow)
	adcLevel := ca.deviceState.GetRowLevel(selectedRow)
	isActive := ca.deviceState.IsRowActive(selectedRow)

	fyne.Do(func() {
		// Build detailed info string with signal chain data
		var infoStr string
		if isActive && voltage > 0.01 {
			// Show full signal chain: G → I → TIA → ADC
			infoStr = fmt.Sprintf("Cell [%d,%d]: State %d/%d | G=%.1fµS | BL=%.2fV → I=%.1fµA → TIA=%.2fV → ADC=%d | %s",
				selectedRow, selectedCol, level, levels-1, conductanceUS, voltage, expectedCurrent, rowVoltage, adcLevel, matName)
		} else {
			// Cell not being sensed
			infoStr = fmt.Sprintf("Cell [%d,%d]: State %d/%d | G=%.1fµS | (Row %s, BL=%.2fV) | %s",
				selectedRow, selectedCol, level, levels-1, conductanceUS,
				map[bool]string{true: "ON", false: "OFF"}[isActive], voltage, matName)
		}
		ca.sharedCellInfoLabel.SetText(infoStr)
	})

	// Also update array info label with total row current
	if ca.sharedArrayInfoLabel != nil {
		fyne.Do(func() {
			ca.sharedArrayInfoLabel.SetText(fmt.Sprintf("Array: %dx%d | %d levels | Row %d sum: I=%.1fµA",
				ca.arrayRows, ca.arrayCols, ca.quantLevels, selectedRow, rowCurrent))
		})
	}
}

// updateOperationClassification updates the operation classification display
func (ca *CircuitsApp) updateOperationClassification() {
	if ca.operationsModeHelp == nil || ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()
	arch := ca.architecture
	readRange := ca.deviceState.GetReadRange()
	writeRange := ca.deviceState.GetWriteRange()

	var helpText string
	switch mode {
	case OpModeRead:
		if arch == sharedwidgets.Architecture2T1R {
			helpText = fmt.Sprintf("READ: Single row, 0-%.1fV. 2T1R provides perfect isolation.", readRange.Max)
		} else if arch == sharedwidgets.Architecture1T1R {
			helpText = fmt.Sprintf("READ: Single row, 0-%.1fV. 1T1R transistor isolates selected row.", readRange.Max)
		} else {
			helpText = fmt.Sprintf("READ: 0-%.1fV. Passive array - sneak currents add 5-20%% error.", readRange.Max)
		}
	case OpModeWrite:
		if arch == sharedwidgets.Architecture2T1R {
			helpText = fmt.Sprintf("WRITE: Single row, %.1f-%.1fV. 2T1R selects single cell.", writeRange.Min, writeRange.Max)
		} else if arch == sharedwidgets.Architecture1T1R {
			helpText = fmt.Sprintf("WRITE: Single row, %.1f-%.1fV. 1T1R gates selected row.", writeRange.Min, writeRange.Max)
		} else {
			helpText = fmt.Sprintf("WRITE: %.1f-%.1fV. Passive: V/2 scheme reduces half-select disturb.", writeRange.Min, writeRange.Max)
		}
	case OpModeCompute:
		if arch == sharedwidgets.Architecture0T1R {
			helpText = fmt.Sprintf("COMPUTE: All rows, 0-%.1fV. Passive natural MVM mode (~75ns).", readRange.Max)
		} else {
			helpText = fmt.Sprintf("COMPUTE: All transistors ON, 0-%.1fV. Full MVM in ~75ns.", readRange.Max)
		}
	default:
		helpText = "Select a mode: READ, WRITE, or COMPUTE."
	}

	fyne.Do(func() {
		ca.operationsModeHelp.SetText(helpText)
	})
}

// H2 FIX: updateWriteTargetLabel updates the target cell display in write mode panel
func (ca *CircuitsApp) updateWriteTargetLabel() {
	if ca.mfuxWriteTargetLabel == nil || ca.deviceState == nil {
		return
	}

	row := ca.deviceState.GetSelectedRow()
	col := ca.deviceState.GetSelectedCol()

	fyne.Do(func() {
		ca.mfuxWriteTargetLabel.SetText(fmt.Sprintf("Target: Row %d, Col %d", row, col))
	})
}

// ============================================================================
// ARCHITECTURE TOGGLE
// ============================================================================

// createArchitectureToggle creates the PASSIVE/1T1R/2T1R toggle buttons
func (ca *CircuitsApp) createArchitectureToggle() fyne.CanvasObject {
	ca.archPassiveBtn = widget.NewButton("PASSIVE", nil)
	ca.arch1T1RBtn = widget.NewButton("1T1R", nil)
	ca.arch2T1RBtn = widget.NewButton("2T1R", nil)

	// Helper to update button styles based on selection
	updateArchButtons := func() {
		ca.archPassiveBtn.Importance = widget.LowImportance
		ca.arch1T1RBtn.Importance = widget.LowImportance
		ca.arch2T1RBtn.Importance = widget.LowImportance
		switch ca.architecture {
		case sharedwidgets.Architecture0T1R:
			ca.archPassiveBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture1T1R:
			ca.arch1T1RBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture2T1R:
			ca.arch2T1RBtn.Importance = widget.HighImportance
		default:
			ca.archPassiveBtn.Importance = widget.HighImportance
		}
		ca.archPassiveBtn.Refresh()
		ca.arch1T1RBtn.Refresh()
		ca.arch2T1RBtn.Refresh()
	}

	// Set initial state
	updateArchButtons()

	// Wire up callbacks
	ca.archPassiveBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture0T1R
		ca.mu.Unlock()
		updateArchButtons()
		// Passive mode: all WLs always active, cannot be changed
		ca.deviceState.SetPassiveMode(true)
		ca.updateWLCheckboxesForArchitecture()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	ca.arch1T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture1T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture1T1R
		ca.mu.Unlock()
		updateArchButtons()
		// 1T1R: disable passive mode, set WLs based on current operation mode
		ca.deviceState.SetPassiveMode(false)
		// Preserve WL state based on operation mode
		if ca.deviceState.GetOperationMode() == OpModeCompute {
			ca.deviceState.SetWLAll() // COMPUTE needs all rows for MVM
		} else {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.updateWLCheckboxesForArchitecture()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	ca.arch2T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture2T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture2T1R
		ca.mu.Unlock()
		updateArchButtons()
		// 2T1R: disable passive mode, set WLs based on current operation mode
		ca.deviceState.SetPassiveMode(false)
		// Preserve WL state based on operation mode
		if ca.deviceState.GetOperationMode() == OpModeCompute {
			ca.deviceState.SetWLAll() // COMPUTE needs all rows for MVM
		} else {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.updateWLCheckboxesForArchitecture()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)

	archLabel := widget.NewLabel("Array:")
	return container.NewHBox(archLabel, ca.archToggle)
}

// ============================================================================
// MODE-FIRST UX PANELS (Phase 1)
// ============================================================================

// createModeBar creates the top-level mode selection bar
// This replaces the mode buttons previously buried in createWLSelector()
func (ca *CircuitsApp) createModeBar() fyne.CanvasObject {
	ca.modeReadBtn = widget.NewButton("READ", func() {
		ca.setOperationMode(OpModeRead)
	})
	ca.modeWriteBtn = widget.NewButton("WRITE", func() {
		ca.setOperationMode(OpModeWrite)
	})
	ca.modeComputeBtn = widget.NewButton("COMPUTE", func() {
		ca.setOperationMode(OpModeCompute)
	})

	// Set initial highlight (READ mode by default)
	ca.modeReadBtn.Importance = widget.HighImportance

	return container.NewHBox(
		widget.NewLabelWithStyle("Mode:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.modeReadBtn,
		ca.modeWriteBtn,
		ca.modeComputeBtn,
		layout.NewSpacer(),
	)
}

// createWriteModePanel creates the write mode panel with level slider
// This addresses UX-004: No target level selector for WRITE mode
func (ca *CircuitsApp) createWriteModePanel() fyne.CanvasObject {
	// Slider: 0 to (quantLevels-1) - uses configured level count
	maxLevel := ca.quantLevels - 1
	midLevel := ca.quantLevels / 2
	ca.mfuxWriteLevelSlider = widget.NewSlider(0, float64(maxLevel))
	ca.mfuxWriteLevelSlider.Step = 1
	ca.mfuxWriteLevelSlider.Value = float64(midLevel)
	ca.mfuxWriteLevelSlider.OnChanged = func(v float64) {
		ca.onWriteLevelChanged(int(v))
	}

	ca.mfuxWriteLevelLabel = widget.NewLabel(fmt.Sprintf("Level: %d", midLevel))
	ca.mfuxWriteLevelLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteVoltageLabel = widget.NewLabel("Voltage: 1.00V")
	ca.mfuxWriteVoltageLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// H2 FIX: Add target cell display
	ca.mfuxWriteTargetLabel = widget.NewLabel("Target: Row 0, Col 0")
	ca.mfuxWriteTargetLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Layout: Title row with target cell, then slider with value labels
	titleLabel := widget.NewLabelWithStyle("Target Write Level:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	headerRow := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		ca.mfuxWriteTargetLabel,
	)

	sliderRow := container.NewBorder(nil, nil,
		ca.mfuxWriteLevelLabel,
		ca.mfuxWriteVoltageLabel,
		ca.mfuxWriteLevelSlider,
	)

	return container.NewVBox(
		headerRow,
		sliderRow,
	)
}

// onWriteLevelChanged handles write level slider changes
// Only updates UI labels - does NOT apply voltage to DAC
// Voltage is only applied when user presses "Write Cell" button
func (ca *CircuitsApp) onWriteLevelChanged(level int) {
	if ca.deviceState == nil {
		return
	}

	// Calculate voltage for display only (don't apply to DAC)
	voltage := ca.deviceState.CalculateVoltageForState(level, ca.quantLevels)

	fyne.Do(func() {
		if ca.mfuxWriteLevelLabel != nil {
			ca.mfuxWriteLevelLabel.SetText(fmt.Sprintf("Level: %d", level))
		}
		if ca.mfuxWriteVoltageLabel != nil {
			ca.mfuxWriteVoltageLabel.SetText(fmt.Sprintf("Voltage: %.2fV", voltage))
		}
	})
}

// createComputeModePanel creates the compute mode panel with input vector entries
// This addresses UX-005: Input vector entries not visible
func (ca *CircuitsApp) createComputeModePanel() fyne.CanvasObject {
	maxCols := min(8, ca.arrayCols)
	ca.mfuxInputVectorEntry = make([]*widget.Entry, maxCols)
	ca.mfuxInputVectorLabels = make([]*widget.Label, maxCols)

	entriesBox := container.NewHBox()
	for i := 0; i < maxCols; i++ {
		idx := i
		entry := widget.NewEntry()
		entry.SetPlaceHolder("0")
		entry.SetText("0")
		entry.OnChanged = func(s string) {
			ca.onInputVectorEntryChanged(idx, s)
		}
		ca.mfuxInputVectorEntry[i] = entry

		label := widget.NewLabel(fmt.Sprintf("x%d", i))
		label.TextStyle = fyne.TextStyle{Monospace: true}
		ca.mfuxInputVectorLabels[i] = label

		// Each column: label above entry
		col := container.NewVBox(label, entry)
		entriesBox.Add(col)
	}

	// Random button to populate with random values
	randomBtn := widget.NewButton("Random", func() {
		ca.randomizeInputVectorEntries()
	})

	// Clear button
	clearBtn := widget.NewButton("Clear", func() {
		ca.clearInputVectorEntries()
	})

	titleLabel := widget.NewLabelWithStyle("Input Vector (0-255):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	return container.NewVBox(
		titleLabel,
		entriesBox,
		container.NewHBox(randomBtn, clearBtn),
	)
}

// onInputVectorEntryChanged handles input vector entry changes
// Only applies DAC changes in COMPUTE mode to prevent state corruption
func (ca *CircuitsApp) onInputVectorEntryChanged(col int, valueStr string) {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return
	}

	// Clamp to valid range
	if value < 0 {
		value = 0
	}
	if value > 255 {
		value = 255
	}

	// Always store the value (for when user switches to COMPUTE mode)
	ca.mu.Lock()
	if col < len(ca.inputVector) {
		ca.inputVector[col] = value
	}
	ca.mu.Unlock()

	// Only apply to DAC if in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		return
	}

	// Convert all inputs to DAC voltages
	params := make([]float64, len(ca.inputVector))
	ca.mu.RLock()
	for i, v := range ca.inputVector {
		params[i] = float64(v)
	}
	ca.mu.RUnlock()

	ca.deviceState.SetDACPreset(DACInputVector, params...)
	ca.recomputeAndRefresh()
}

// randomizeInputVectorEntries fills entries with random 0-255 values
// Only applies to DAC if in COMPUTE mode
func (ca *CircuitsApp) randomizeInputVectorEntries() {
	// Generate random values and copy for UI update
	ca.mu.Lock()
	valuesCopy := make([]int, len(ca.inputVector))
	for i := range ca.inputVector {
		ca.inputVector[i] = rand.Intn(256)
		valuesCopy[i] = ca.inputVector[i]
	}
	ca.mu.Unlock()

	// Update entry widgets (no lock - use copy)
	fyne.Do(func() {
		for i, entry := range ca.mfuxInputVectorEntry {
			if entry != nil && i < len(valuesCopy) {
				entry.SetText(strconv.Itoa(valuesCopy[i]))
			}
		}
	})

	// Only apply to DAC if in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		return
	}

	// Apply to DAC
	params := make([]float64, len(valuesCopy))
	for i, v := range valuesCopy {
		params[i] = float64(v)
	}

	ca.deviceState.SetDACPreset(DACInputVector, params...)
	ca.recomputeAndRefresh()
}

// clearInputVectorEntries sets all entries to 0
// Only applies to DAC if in COMPUTE mode
func (ca *CircuitsApp) clearInputVectorEntries() {
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = 0
	}
	// IMPORTANT: Unlock BEFORE fyne.Do to prevent deadlock.
	// SetText triggers OnChanged which acquires ca.mu.
	ca.mu.Unlock()

	// Update entry widgets (no lock held - safe)
	fyne.Do(func() {
		for _, entry := range ca.mfuxInputVectorEntry {
			if entry != nil {
				entry.SetText("0")
			}
		}
	})

	// Only apply to DAC if in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		return
	}

	// Apply to DAC
	params := make([]float64, len(ca.inputVector))
	ca.deviceState.SetDACPreset(DACInputVector, params...)
	ca.recomputeAndRefresh()
}

// updateModePanels shows/hides mode-specific panels based on current mode
func (ca *CircuitsApp) updateModePanels(mode OpMode) {
	fyne.Do(func() {
		// Hide all panels first
		if ca.writeModePanel != nil {
			ca.writeModePanel.Hide()
		}
		if ca.computeModePanel != nil {
			ca.computeModePanel.Hide()
		}

		// Show relevant panel
		switch mode {
		case OpModeWrite:
			if ca.writeModePanel != nil {
				ca.writeModePanel.Show()
				// Update slider to reflect current selection
				if ca.mfuxWriteLevelSlider != nil {
					// Trigger an update to sync voltage display
					ca.onWriteLevelChanged(int(ca.mfuxWriteLevelSlider.Value))
				}
				// H2 FIX: Update target cell label when entering write mode
				ca.updateWriteTargetLabel()
			}
		case OpModeCompute:
			if ca.computeModePanel != nil {
				ca.computeModePanel.Show()
			}
		// OpModeRead: no special panel needed (clean view)
		}
	})
}

// ====================================================================================
// VOLTAGE RULES UI - 4-Phase Write Sequence, ISPP Animation, V/2 Overlay
// ====================================================================================

// AnimationFrameDelayMs is the animation frame delay for smooth updates
const AnimationFrameDelayMs = 50

// UI Colors for voltage visualization
var (
	colorFullVoltage   = color.RGBA{255, 200, 0, 255}   // Bright Gold for target cell
	colorHalfSelect    = color.RGBA{255, 165, 0, 255}   // Amber for V/2 cells
	colorZeroVoltage   = color.RGBA{50, 50, 60, 255}    // Dim Gray for inactive
	colorAscending     = color.RGBA{100, 220, 120, 255} // Green for ascending
	colorDescending    = color.RGBA{220, 100, 100, 255} // Red for descending
	colorPhaseActive   = color.RGBA{100, 200, 255, 255} // Cyan for active phase
	colorPhaseInactive = color.RGBA{80, 80, 90, 255}    // Dim for inactive phase
)

// drawWriteSequenceTimingDiagram draws the 4-phase timing diagram
// Shows: RESET -> HOLD -> WRITE -> HOLD with phase highlighting
func (ca *CircuitsApp) drawWriteSequenceTimingDiagram() fyne.CanvasObject {
	phaseInfo := ca.deviceState.GetWritePhaseInfo()

	// Phase labels with durations
	phases := []struct {
		name     string
		duration int
		phase    WritePhase
	}{
		{"RESET", PhaseResetDurationNs, PhaseReset},
		{"HOLD", PhaseHold1DurationNs, PhaseHold1},
		{"WRITE", PhaseWriteDurationNs, PhaseWrite},
		{"HOLD", PhaseHold2DurationNs, PhaseHold2},
	}

	phaseBoxes := container.NewHBox()

	for _, p := range phases {
		// Choose color based on current phase
		bgColor := colorPhaseInactive
		if phaseInfo.Active && phaseInfo.Phase == p.phase {
			bgColor = colorPhaseActive
		}

		// Create phase box with label and duration
		label := canvas.NewText(fmt.Sprintf("%s\n%dns", p.name, p.duration), color.White)
		label.Alignment = fyne.TextAlignCenter
		label.TextSize = 10

		bg := canvas.NewRectangle(bgColor)
		bg.SetMinSize(fyne.NewSize(60, 40))

		phaseBox := container.NewStack(bg, container.NewCenter(label))
		phaseBoxes.Add(phaseBox)
	}

	// Progress bar
	progress := widget.NewProgressBar()
	progress.SetValue(phaseInfo.Progress)
	progress.Min = 0
	progress.Max = 1

	// Voltage label showing current write voltage
	voltageLabel := widget.NewLabel(fmt.Sprintf("Write Voltage: %.2fV", phaseInfo.WriteVoltage))
	voltageLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewVBox(
		widget.NewLabelWithStyle("4-Phase Write Sequence", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		phaseBoxes,
		progress,
		voltageLabel,
	)
}

// animateWriteSequence runs the 4-phase write animation
// This is called from a goroutine
func (ca *CircuitsApp) animateWriteSequence() {
	for {
		// Check for stop signal
		if ca.shouldStop() {
			return
		}

		if ca.deviceState == nil {
			return
		}

		phaseInfo := ca.deviceState.GetWritePhaseInfo()
		if !phaseInfo.Active {
			return
		}

		// Get phase duration for timing
		duration := GetPhaseDuration(phaseInfo.Phase)

		// Update UI
		ca.updateWriteSequenceUI()

		// Delay proportional to phase duration (scaled for animation)
		animDelayMs := duration / 4
		if animDelayMs < AnimationFrameDelayMs {
			animDelayMs = AnimationFrameDelayMs
		}
		if ca.sleep(animDelayMs) {
			return // Interrupted
		}

		// Advance to next phase
		complete := ca.deviceState.AdvanceWritePhase()
		if complete {
			ca.updateWriteSequenceUI()
			return
		}
	}
}

// updateWriteSequenceUI refreshes the 4-phase timing display
func (ca *CircuitsApp) updateWriteSequenceUI() {
	fyne.Do(func() {
		if ca.writeSequencePanel != nil {
			// Rebuild the timing diagram with current state
			newDiagram := ca.drawWriteSequenceTimingDiagram()
			ca.writeSequencePanel.Objects = []fyne.CanvasObject{newDiagram}
			ca.writeSequencePanel.Refresh()
		}
	})
}

// runISPPWithAnimation runs the ISPP loop with visual feedback
// This ENHANCES the existing writeReadVerifyLoop() by adding:
// - 4-phase sequence animation within each iteration
// - Calibrated per-level voltage lookup
// - Hysteresis direction tracking
// - V/2 visualization for 0T1R mode
func (ca *CircuitsApp) runISPPWithAnimation(row, col, targetLevel int) {
	const iterationDelay = 300 * time.Millisecond

	// Get current level
	ca.mu.Lock()
	currentLevel := 0
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		currentLevel = ca.arrayWeights[row][col]
	}
	ca.mu.Unlock()

	// Determine direction
	direction := ca.deviceState.GetWriteDirection(row, col, currentLevel, targetLevel)
	ascending := direction == DirectionAscending

	// Enable V/2 visualization if in passive (0T1R) mode
	if ca.deviceState.IsPassiveMode() {
		voltage := ca.deviceState.GetVoltageForLevel(targetLevel, ascending)
		ca.deviceState.EnableHalfSelectVisualization(row, col, voltage)
		ca.updateHalfSelectVisualization()
	}

	// Start ISPP
	ca.deviceState.StartISPP(row, col, targetLevel, currentLevel)

	for {
		// Check for stop signal at start of each iteration
		if ca.shouldStop() {
			ca.deviceState.CancelISPP()
			return
		}

		isppStatus := ca.deviceState.GetISPPStatus()
		if !isppStatus.Active {
			break
		}

		// Start 4-phase write sequence for this iteration
		ca.deviceState.StartWriteSequence(row, col, targetLevel)
		go ca.animateWriteSequence()

		// Update ISPP UI
		ca.updateISPPUI()

		// Wait for 4-phase sequence to complete
		for {
			if ca.shouldStop() {
				ca.deviceState.CancelWriteSequence()
				ca.deviceState.CancelISPP()
				return
			}
			phaseInfo := ca.deviceState.GetWritePhaseInfo()
			if !phaseInfo.Active {
				break
			}
			if ca.sleep(AnimationFrameDelayMs) {
				ca.deviceState.CancelWriteSequence()
				ca.deviceState.CancelISPP()
				return
			}
		}

		// Simulate write: move toward target (same logic as existing writeReadVerifyLoop)
		ca.mu.Lock()
		if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
			if currentLevel < targetLevel {
				step := 1
				if targetLevel-currentLevel > 3 {
					step = 2
				}
				currentLevel += step
				if currentLevel > targetLevel {
					currentLevel = targetLevel
				}
			} else if currentLevel > targetLevel {
				step := 1
				if currentLevel-targetLevel > 3 {
					step = 2
				}
				currentLevel -= step
				if currentLevel < targetLevel {
					currentLevel = targetLevel
				}
			}
			ca.arrayWeights[row][col] = currentLevel
		}
		ca.mu.Unlock()

		// ISPP iteration with verification
		result := ca.deviceState.ISPPIterate(currentLevel)

		switch result {
		case ISPPResultVerified:
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("SUCCESS [%d,%d] = Level %d (%d iterations)",
					row, col, targetLevel, isppStatus.Iteration))
			})
			goto cleanup

		case ISPPResultOvershoot:
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("OVERSHOOT [%d,%d] - Resetting to saturation...", row, col))
			})
			ca.deviceState.HandleOvershoot(row, col)
			// Reset local currentLevel based on direction
			if direction == DirectionAscending {
				currentLevel = 0
			} else {
				currentLevel = 29
			}
			ca.mu.Lock()
			if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
				ca.arrayWeights[row][col] = currentLevel
			}
			ca.mu.Unlock()
			ca.recomputeAndRefresh()
			time.Sleep(iterationDelay)
			continue

		case ISPPResultMaxIterations:
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("PARTIAL [%d,%d] = Level %d (target was %d)",
					row, col, currentLevel, targetLevel))
			})
			goto cleanup
		}

		// Continue - update UI and wait
		time.Sleep(iterationDelay)
	}

cleanup:
	// Disable V/2 visualization
	ca.deviceState.DisableHalfSelectVisualization()
	ca.updateHalfSelectVisualization()

	// Record the write in hysteresis state
	ca.deviceState.RecordWrite(row, col, currentLevel)

	ca.recomputeAndRefresh()
}

// updateISPPUI refreshes the ISPP status display
func (ca *CircuitsApp) updateISPPUI() {
	isppStatus := ca.deviceState.GetISPPStatus()

	// Direction indicator
	dirStr := "^ Ascending"
	dirColor := colorAscending
	if isppStatus.Direction == DirectionDescending {
		dirStr = "v Descending"
		dirColor = colorDescending
	}

	fyne.Do(func() {
		if ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("ISPP [%d/%d]: Level %d -> %d | V=%.2fV | %s",
				isppStatus.Iteration, isppStatus.MaxIter,
				isppStatus.CurrentLevel, isppStatus.TargetLevel,
				isppStatus.Voltage, dirStr))
		}

		// Update direction indicator if we have one
		if ca.hysteresisDirectionLabel != nil {
			ca.hysteresisDirectionLabel.SetText(dirStr)
			// Note: Fyne doesn't support dynamic text colors easily, so we just update text
		}
		_ = dirColor // Use in future canvas-based indicator
	})
}

// drawHalfSelectOverlay draws V/2 voltage indicators on the array
// Target cell: Bright Gold (full voltage)
// Half-selected cells: Amber (V/2)
func (ca *CircuitsApp) drawHalfSelectOverlay(arrayCanvas fyne.CanvasObject) {
	hsState := ca.deviceState.GetHalfSelectState()
	if !hsState.Enabled {
		return
	}

	// This would overlay colored rectangles on the array canvas
	// For now, we update the array refresh to include V/2 coloring
	// The actual drawing happens in refreshArrayHeatmap or similar
}

// updateHalfSelectVisualization enables/disables V/2 overlay
func (ca *CircuitsApp) updateHalfSelectVisualization() {
	hsState := ca.deviceState.GetHalfSelectState()

	fyne.Do(func() {
		if ca.halfSelectIndicator != nil {
			if hsState.Enabled {
				ca.halfSelectIndicator.SetText(fmt.Sprintf("V/2 Bias Active | Full: %.2fV | Half: %.2fV",
					hsState.FullVoltage, hsState.HalfVoltage))
				ca.halfSelectIndicator.Show()
			} else {
				ca.halfSelectIndicator.Hide()
			}
		}
	})

	// Trigger array refresh to update cell colors
	ca.recomputeAndRefresh()
}

// getHalfSelectCellColor returns the color for a cell based on half-select state
// Used by array visualization to color cells during V/2 mode
func (ca *CircuitsApp) getHalfSelectCellColor(row, col int) (color.Color, bool) {
	hsState := ca.deviceState.GetHalfSelectState()
	if !hsState.Enabled {
		return nil, false
	}

	// Target cell gets full voltage color
	if row == hsState.SelectedRow && col == hsState.SelectedCol {
		return colorFullVoltage, true
	}

	// Check if half-selected
	if ca.deviceState.IsHalfSelected(row, col) {
		return colorHalfSelect, true
	}

	return nil, false
}

// createPassiveVoltagePanel creates the V/2 panel for 0T1R (passive) mode
func (ca *CircuitsApp) createPassiveVoltagePanel() fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle("Passive Crossbar (0T1R)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	infoText := `In passive crossbar mode, V/2 biasing is required:
- Target cell receives full write voltage
- Same-row and same-column cells receive V/2
- This prevents unintended state changes

Watch for disturb effects on half-selected cells.`

	infoLabel := widget.NewLabel(infoText)
	infoLabel.Wrapping = fyne.TextWrapWord

	// V/2 indicator (updated during write operations)
	ca.halfSelectIndicator = widget.NewLabel("V/2 Bias: Inactive")
	ca.halfSelectIndicator.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		infoLabel,
		ca.halfSelectIndicator,
	)
}

// createActiveVoltagePanel creates the panel for 1T1R/2T1R (active) mode
func (ca *CircuitsApp) createActiveVoltagePanel() fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle("Active Cell Access (1T1R/2T1R)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	infoText := `In active transistor mode:
- Transistors isolate non-selected cells
- Only the target cell sees write voltage
- No V/2 disturb effects
- Higher area overhead but cleaner writes`

	infoLabel := widget.NewLabel(infoText)
	infoLabel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		infoLabel,
	)
}

// updateArchitectureSpecificUI shows/hides panels based on architecture
func (ca *CircuitsApp) updateArchitectureSpecificUI() {
	isPassive := ca.deviceState.IsPassiveMode()

	fyne.Do(func() {
		if ca.passiveVoltagePanel != nil && ca.activeVoltagePanel != nil {
			if isPassive {
				ca.passiveVoltagePanel.Show()
				ca.activeVoltagePanel.Hide()
			} else {
				ca.passiveVoltagePanel.Hide()
				ca.activeVoltagePanel.Show()
			}
		}
	})
}

// updateHysteresisDirectionUI updates the direction indicator
func (ca *CircuitsApp) updateHysteresisDirectionUI(targetLevel int) {
	if ca.deviceState == nil {
		return
	}

	row := ca.deviceState.GetSelectedRow()
	col := ca.deviceState.GetSelectedCol()

	// Get current level
	ca.mu.RLock()
	currentLevel := 0
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		currentLevel = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	direction := ca.deviceState.GetWriteDirection(row, col, currentLevel, targetLevel)

	dirStr := "- Unknown"
	if direction == DirectionAscending {
		dirStr = "^ Ascending"
	} else if direction == DirectionDescending {
		dirStr = "v Descending"
	}

	fyne.Do(func() {
		if ca.hysteresisDirectionLabel != nil {
			ca.hysteresisDirectionLabel.SetText(dirStr)
		}
	})
}

// createEnhancedWriteModePanel creates an enhanced write panel with voltage rules UI
// This adds hysteresis direction and 4-phase sequence display to the existing write panel
func (ca *CircuitsApp) createEnhancedWriteModePanel() fyne.CanvasObject {
	// Existing write panel components
	maxLevel := ca.quantLevels - 1
	midLevel := ca.quantLevels / 2

	ca.mfuxWriteLevelSlider = widget.NewSlider(0, float64(maxLevel))
	ca.mfuxWriteLevelSlider.Step = 1
	ca.mfuxWriteLevelSlider.Value = float64(midLevel)
	ca.mfuxWriteLevelSlider.OnChanged = func(v float64) {
		ca.onWriteLevelChanged(int(v))
		ca.updateHysteresisDirectionUI(int(v))
	}

	// Level label with min/max indicators
	minLabel := widget.NewLabel("0")
	maxLabel := widget.NewLabel(fmt.Sprintf("%d", maxLevel))
	ca.mfuxWriteLevelLabel = widget.NewLabel(fmt.Sprintf("Level: %d", midLevel))
	ca.mfuxWriteLevelLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteVoltageLabel = widget.NewLabel("Voltage: 1.00V")
	ca.mfuxWriteVoltageLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteTargetLabel = widget.NewLabel("Target: Row 0, Col 0")
	ca.mfuxWriteTargetLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Hysteresis direction indicator
	ca.hysteresisDirectionLabel = widget.NewLabel("- Unknown")
	ca.hysteresisDirectionLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Layout
	titleLabel := widget.NewLabelWithStyle("Target Write Level:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	headerRow := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		ca.mfuxWriteTargetLabel,
		widget.NewSeparator(),
		ca.hysteresisDirectionLabel,
	)

	sliderWithMinMax := container.NewBorder(nil, nil,
		minLabel,
		maxLabel,
		ca.mfuxWriteLevelSlider,
	)

	valueRow := container.NewHBox(
		ca.mfuxWriteLevelLabel,
		layout.NewSpacer(),
		ca.mfuxWriteVoltageLabel,
	)

	// 4-phase sequence container (populated during write)
	ca.writeSequencePanel = container.NewStack()

	return container.NewVBox(
		headerRow,
		sliderWithMinMax,
		valueRow,
		widget.NewSeparator(),
		ca.writeSequencePanel,
	)
}
