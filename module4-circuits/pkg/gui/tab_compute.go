package gui

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ============================================================================
// TAB 3: COMPUTE MODE
// ============================================================================

func (ca *CircuitsApp) createComputeTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**COMPUTE MODE**: Perform matrix-vector multiplication in a single analog operation. Input voltages are applied to columns, multiplied by cell conductances (stored weights). Kirchhoff's Current Law (KCL) sums all column currents at each row line, computing all dot products in parallel in ~20ns.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Configuration section
	configSection := ca.createComputeConfigSection()

	// Input vector section
	inputSection := ca.createComputeInputSection()

	// Visualization section
	vizSection := ca.createComputeVizSection()

	// Math breakdown section
	mathSection := ca.createComputeMathSection()

	// Output section
	outputSection := ca.createComputeOutputSection()

	// Buttons
	computeBtn := widget.NewButton("COMPUTE", ca.onCompute)
	computeBtn.Importance = widget.HighImportance

	animateBtn := widget.NewButton("ANIMATE STEP-BY-STEP", ca.onAnimateCompute)

	resetBtn := widget.NewButton("RESET", ca.onResetCompute)

	ca.computeStatusLabel = widget.NewLabel("Ready to compute")

	buttonBox := container.NewHBox(
		computeBtn,
		animateBtn,
		resetBtn,
		layout.NewSpacer(),
		ca.computeStatusLabel,
	)

	// Layout
	leftPanel := container.NewVBox(
		widget.NewLabel("CONFIGURATION"),
		configSection,
		widget.NewSeparator(),
		widget.NewLabel("INPUT VECTOR"),
		inputSection,
	)

	centerPanel := container.NewVBox(
		widget.NewLabel("COMPUTE VISUALIZATION"),
		vizSection,
		widget.NewSeparator(),
		widget.NewLabel("MATH BREAKDOWN (Row 0)"),
		mathSection,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("OUTPUT VECTOR"),
		outputSection,
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

func (ca *CircuitsApp) createComputeConfigSection() fyne.CanvasObject {
	sizeOptions := []string{"4", "8", "16", "32"}
	rowSelect := widget.NewSelect(sizeOptions, nil)
	rowSelect.SetSelected("8")

	colSelect := widget.NewSelect(sizeOptions, nil)
	colSelect.SetSelected("8")

	levelOptions := []string{"30"}
	levelSelect := widget.NewSelect(levelOptions, nil)
	levelSelect.SetSelected("30")

	dacBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	dacSelect := widget.NewSelect(dacBitsOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		ca.mu.Lock()
		ca.dacBits = bits
		ca.mu.Unlock()
	})
	dacSelect.SetSelected("8")

	adcBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	adcSelect := widget.NewSelect(adcBitsOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		ca.mu.Lock()
		ca.adcBits = bits
		ca.mu.Unlock()
	})
	adcSelect.SetSelected("8")

	readVEntry := widget.NewEntry()
	readVEntry.SetText("0.5")

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Array Size:"), rowSelect, widget.NewLabel("x"), colSelect),
		container.NewHBox(widget.NewLabel("Levels:"), levelSelect),
		container.NewHBox(widget.NewLabel("DAC Bits:"), dacSelect),
		container.NewHBox(widget.NewLabel("ADC Bits:"), adcSelect),
		container.NewHBox(widget.NewLabel("Read Voltage:"), readVEntry, widget.NewLabel("V")),
	)
}

func (ca *CircuitsApp) createComputeInputSection() fyne.CanvasObject {
	modeOptions := []string{"Manual", "Random", "Ramp", "Pattern"}
	modeSelect := widget.NewSelect(modeOptions, func(s string) {
		switch s {
		case "Random":
			ca.mu.Lock()
			for i := range ca.inputVector {
				ca.inputVector[i] = rand.Intn(256)
			}
			ca.mu.Unlock()
			ca.updateComputeInputs()
		case "Ramp":
			ca.mu.Lock()
			for i := range ca.inputVector {
				ca.inputVector[i] = i * 255 / max(1, len(ca.inputVector)-1)
			}
			ca.mu.Unlock()
			ca.updateComputeInputs()
		}
	})
	modeSelect.SetSelected("Manual")

	// Create input entries
	ca.computeInputs = make([]*widget.Entry, ca.arrayCols)
	ca.computeVoltageLabels = make([]*widget.Label, ca.arrayCols)

	inputGrid := container.NewGridWithColumns(4)
	maxDisplay := min(8, ca.arrayCols)
	for i := 0; i < maxDisplay; i++ {
		ca.computeInputs[i] = widget.NewEntry()
		ca.computeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))

		idx := i
		ca.computeInputs[i].OnChanged = func(s string) {
			var v int
			fmt.Sscanf(s, "%d", &v)
			if v > 255 {
				v = 255
			}
			ca.mu.Lock()
			ca.inputVector[idx] = v
			ca.mu.Unlock()
		}

		ca.computeVoltageLabels[i] = widget.NewLabel(fmt.Sprintf("%.2fV", float64(ca.inputVector[i])/255.0))

		inputGrid.Add(container.NewVBox(
			widget.NewLabel(fmt.Sprintf("x%d", i)),
			ca.computeInputs[i],
			ca.computeVoltageLabels[i],
		))
	}

	// Add indicator for remaining columns if array is larger
	if ca.arrayCols > 8 {
		moreLabel := widget.NewLabel(fmt.Sprintf("+ %d more...", ca.arrayCols-8))
		moreLabel.TextStyle = fyne.TextStyle{Italic: true}
		inputGrid.Add(moreLabel)
	}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Input Mode:"), modeSelect),
		widget.NewLabel("Digital Inputs (0-255):"),
		inputGrid,
	)
}

func (ca *CircuitsApp) updateComputeInputs() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	for i := 0; i < min(8, len(ca.computeInputs)); i++ {
		if ca.computeInputs[i] != nil {
			ca.computeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))
		}
		if ca.computeVoltageLabels[i] != nil {
			ca.computeVoltageLabels[i].SetText(fmt.Sprintf("%.2fV", float64(ca.inputVector[i])/255.0))
		}
	}
}

func (ca *CircuitsApp) createComputeVizSection() fyne.CanvasObject {
	ca.computeArrayCanvas = canvas.NewRaster(ca.drawComputeViz)
	ca.computeArrayCanvas.SetMinSize(fyne.NewSize(450, 350))
	return ca.computeArrayCanvas
}

func (ca *CircuitsApp) drawComputeViz(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	ca.mu.RLock()
	rows := min(8, ca.arrayRows)
	cols := min(8, ca.arrayCols)
	inputs := ca.inputVector
	weights := ca.arrayWeights
	outputs := ca.outputVector
	ca.mu.RUnlock()

	dacColor := color.RGBA{150, 100, 200, 255}
	adcColor := color.RGBA{100, 200, 150, 255}

	// Calculate square cell size for array (use min of both dimensions)
	maxArrayW := w - 260 // Leave space for DACs and ADCs
	maxArrayH := h - 80
	cellW := maxArrayW / cols
	cellH := maxArrayH / rows
	cellSize := cellW
	if cellH < cellSize {
		cellSize = cellH
	}
	if cellSize > 30 {
		cellSize = 30
	}
	if cellSize < 12 {
		cellSize = 12
	}

	// Array dimensions
	arrayW := cols * cellSize
	arrayH := rows * cellSize

	// Center layout horizontally
	totalW := 60 + 20 + arrayW + 20 + 60 // DAC + gap + array + gap + ADC
	startX := (w - totalW) / 2
	if startX < 10 {
		startX = 10
	}

	dacX := startX
	dacW := 60
	arrayX := dacX + dacW + 20
	adcX := arrayX + arrayW + 20

	// Vertical centering
	arrayY := (h - arrayH) / 2
	if arrayY < 30 {
		arrayY = 30
	}

	// Draw DACs (one per column input)
	for i := 0; i < cols && i < len(inputs); i++ {
		y := arrayY + i*cellSize + (cellSize-24)/2
		drawRect(img, dacX, y, dacW, 24, dacColor)
	}

	// Draw array with square cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := arrayX + c*cellSize
			y0 := arrayY + r*cellSize

			level := weights[r][c]
			intensity := float64(level) / 29.0

			cr := uint8(intensity * 200)
			cg := uint8(100 + (1-intensity)*100)
			cb := uint8((1 - intensity) * 200)
			cellColor := color.RGBA{cr, cg, cb, 255}

			drawRect(img, x0+2, y0+2, cellSize-4, cellSize-4, cellColor)
		}
	}

	// Draw ADCs (one per row output)
	for i := 0; i < rows && i < len(outputs); i++ {
		y := arrayY + i*cellSize + (cellSize-24)/2
		drawRect(img, adcX, y, 60, 24, adcColor)
	}

	return img
}

func (ca *CircuitsApp) createComputeMathSection() fyne.CanvasObject {
	ca.computeMathLabel = widget.NewLabel(
		"I₀ = G₀₀×V₀ + G₀₁×V₁ + G₀₂×V₂ + ... + G₀₇×V₇\n\n" +
			"I₀ = --µS×--V + --µS×--V + ...\n" +
			"   = -- µA\n\n" +
			"THIS IS A DOT PRODUCT! (weights · inputs)\n" +
			"ALL 8 ROWS COMPUTED SIMULTANEOUSLY!",
	)

	return ca.computeMathLabel
}

func (ca *CircuitsApp) createComputeOutputSection() fyne.CanvasObject {
	ca.computeOutputLabels = make([]*widget.Label, 8)

	outputGrid := container.NewGridWithColumns(2)
	for i := 0; i < 8; i++ {
		ca.computeOutputLabels[i] = widget.NewLabel(fmt.Sprintf("y%d: --", i))
		outputGrid.Add(ca.computeOutputLabels[i])
	}

	return container.NewVBox(
		widget.NewLabel("Output Currents (µA):"),
		outputGrid,
		widget.NewSeparator(),
		widget.NewLabel("TOTAL LATENCY: ~20ns"),
	)
}

func (ca *CircuitsApp) onCompute() {
	ca.mu.Lock()
	rows := min(8, ca.arrayRows)
	cols := min(8, ca.arrayCols)

	// Perform MVM: output = weights × input
	for r := 0; r < rows && r < len(ca.arrayWeights); r++ {
		sum := 0.0
		for c := 0; c < cols && c < len(ca.arrayWeights[r]); c++ {
			// Conductance (1-100 µS)
			conductance := 1.0 + float64(ca.arrayWeights[r][c])/29.0*99.0
			// Input voltage (0-1V)
			voltage := float64(ca.inputVector[c]) / 255.0
			// Current contribution
			sum += conductance * voltage
		}
		ca.outputVector[r] = sum
	}
	ca.mu.Unlock()

	// Update output labels
	ca.mu.RLock()
	for i := 0; i < 8 && i < len(ca.outputVector); i++ {
		if ca.computeOutputLabels[i] != nil {
			ca.computeOutputLabels[i].SetText(fmt.Sprintf("y%d: %.1f µA", i, ca.outputVector[i]))
		}
	}
	ca.mu.RUnlock()

	// Update math breakdown for row 0
	ca.updateComputeMath()

	ca.computeStatusLabel.SetText("Compute complete in ~20ns")
}

func (ca *CircuitsApp) updateComputeMath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.arrayWeights) == 0 || len(ca.arrayWeights[0]) == 0 {
		return
	}

	cols := min(6, len(ca.arrayWeights[0]))
	mathText := "I₀ = "
	var terms []string
	totalCurrent := 0.0

	for c := 0; c < cols; c++ {
		conductance := 1.0 + float64(ca.arrayWeights[0][c])/29.0*99.0
		voltage := float64(ca.inputVector[c]) / 255.0
		current := conductance * voltage
		totalCurrent += current
		terms = append(terms, fmt.Sprintf("%.0fµS×%.2fV", conductance, voltage))
	}

	mathText += terms[0]
	for i := 1; i < len(terms); i++ {
		mathText += " + " + terms[i]
	}
	mathText += " + ...\n"
	mathText += fmt.Sprintf("   = %.1f µA\n\n", ca.outputVector[0])
	mathText += "THIS IS A DOT PRODUCT! (weights · inputs)\n"
	mathText += "ALL ROWS COMPUTED SIMULTANEOUSLY!"

	ca.computeMathLabel.SetText(mathText)
}

func (ca *CircuitsApp) onAnimateCompute() {
	ca.computeStatusLabel.SetText("Animating... (DAC → Array → ADC)")
	// Animation would be implemented with goroutines and fyne.Do()
	go func() {
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.computeStatusLabel.SetText("Step 1: DAC conversion (5ns)")
		})
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.computeStatusLabel.SetText("Step 2: Array settle (5ns)")
		})
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.computeStatusLabel.SetText("Step 3: ADC conversion (10ns)")
			ca.onCompute()
		})
	}()
}

func (ca *CircuitsApp) onResetCompute() {
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = 0
	}
	for i := range ca.outputVector {
		ca.outputVector[i] = 0
	}
	ca.mu.Unlock()

	ca.updateComputeInputs()
	for i := 0; i < 8 && i < len(ca.computeOutputLabels); i++ {
		if ca.computeOutputLabels[i] != nil {
			ca.computeOutputLabels[i].SetText(fmt.Sprintf("y%d: --", i))
		}
	}

	ca.computeStatusLabel.SetText("Reset complete")
}
