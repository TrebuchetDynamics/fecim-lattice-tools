package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ============================================================================
// TAB 6: SPECIFICATIONS
// ============================================================================

func (ca *CircuitsApp) createSpecsTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**SPECIFICATIONS**: Detailed electrical and physical parameters for all peripheral components (DAC, ADC, TIA) and FeFET cells. Includes array configuration, conversion times, power consumption, and device characteristics.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Array configuration
	arraySection := ca.createSpecArraySection()

	// DAC specs
	dacSection := ca.createSpecDACSection()

	// ADC specs
	adcSection := ca.createSpecADCSection()

	// TIA specs
	tiaSection := ca.createSpecTIASection()

	// FeFET cell specs
	fefetSection := ca.createSpecFeFETSection()

	// System summary
	summarySection := ca.createSpecSummarySection()

	// Buttons
	exportBtn := widget.NewButton("EXPORT SPECS", ca.onExportSpecs)
	compareBtn := widget.NewButton("COMPARE TO GPU", ca.onCompareToGPU)

	ca.specStatusLabel = widget.NewLabel("System specifications")

	buttonBox := container.NewHBox(
		exportBtn,
		compareBtn,
		layout.NewSpacer(),
		ca.specStatusLabel,
	)

	// Layout in a grid with improved visual hierarchy
	arrayHeader := widget.NewLabelWithStyle("ARRAY CONFIGURATION", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dacHeader := widget.NewLabelWithStyle("DAC SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	adcHeader := widget.NewLabelWithStyle("ADC SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tiaHeader := widget.NewLabelWithStyle("TIA SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fefetHeader := widget.NewLabelWithStyle("FeFET CELL SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	summaryHeader := widget.NewLabelWithStyle("SYSTEM SUMMARY", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	leftCol := container.NewVBox(
		arrayHeader,
		layout.NewSpacer(), // Small spacing after header
		arraySection,
		layout.NewSpacer(), // Spacing before separator
		widget.NewSeparator(),
		layout.NewSpacer(), // Spacing after separator
		dacHeader,
		layout.NewSpacer(),
		dacSection,
		layout.NewSpacer(),
		widget.NewSeparator(),
		layout.NewSpacer(),
		adcHeader,
		layout.NewSpacer(),
		adcSection,
	)

	rightCol := container.NewVBox(
		tiaHeader,
		layout.NewSpacer(),
		tiaSection,
		layout.NewSpacer(),
		widget.NewSeparator(),
		layout.NewSpacer(),
		fefetHeader,
		layout.NewSpacer(),
		fefetSection,
		layout.NewSpacer(),
		widget.NewSeparator(),
		layout.NewSpacer(),
		summaryHeader,
		layout.NewSpacer(),
		summarySection,
	)

	mainContent := container.NewHBox(
		container.NewPadded(leftCol),
		widget.NewSeparator(),
		container.NewPadded(rightCol),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		container.NewVScroll(mainContent),
	)
}

func (ca *CircuitsApp) createSpecArraySection() fyne.CanvasObject {
	sizeOptions := []string{"8", "16", "32", "64", "128"}
	ca.specArraySizeSelect = widget.NewSelect(sizeOptions, func(s string) {
		// Update the summary when size changes
		ca.updateSpecSummary()
	})
	ca.specArraySizeSelect.SetSelected("32")

	levelOptions := []string{"2", "4", "8", "16", "30", "32", "64", "128", "256"}
	ca.specQuantLevelSelect = widget.NewSelect(levelOptions, nil)
	ca.specQuantLevelSelect.SetSelected("30")

	// Calculate storage
	cells := 32 * 32
	bitsPerCell := math.Log2(30)
	totalBits := float64(cells) * bitsPerCell

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Array Size:"), ca.specArraySizeSelect, widget.NewLabel("×"), ca.specArraySizeSelect, widget.NewLabel(fmt.Sprintf("= %d cells", cells))),
		widget.NewLabel(""), // Spacing
		container.NewHBox(widget.NewLabel("Quantization:"), ca.specQuantLevelSelect, widget.NewLabel(fmt.Sprintf("levels (~%.1f bits/cell)", bitsPerCell))),
		widget.NewLabel(""), // Spacing
		widget.NewLabel(fmt.Sprintf("Total Storage: %d × %.1f = %.0f bits", cells, bitsPerCell, totalBits)),
	)
}

func (ca *CircuitsApp) createSpecDACSection() fyne.CanvasObject {
	dacBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	ca.specDACBitsSelect = widget.NewSelect(dacBitsOptions, nil)
	ca.specDACBitsSelect.SetSelected("8")

	specs := `Count:             32 (one per column)
Resolution:        8 bits (256 levels)
Output Range:      0V to 1.0V (read), 2V to 5V (write)
Conversion Time:   5 ns (digital to analog latency)
Power per DAC:     0.1 mW (static + dynamic)
Total DAC Power:   3.2 mW (for 32 DACs)
INL:               < 0.5 LSB (integral nonlinearity)
DNL:               < 0.5 LSB (differential nonlinearity)
Rise/Fall Time:    2-5 ns (signal edge transitions)`

	helpText := widget.NewLabel("DAC converts digital level (0-29) to precise analog voltage for programming FeFET cells")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Resolution:"), ca.specDACBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel(""), // Spacing
		widget.NewLabel(specs),
		widget.NewLabel(""), // Spacing
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecADCSection() fyne.CanvasObject {
	adcBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	ca.specADCBitsSelect = widget.NewSelect(adcBitsOptions, nil)
	ca.specADCBitsSelect.SetSelected("8")

	specs := `Count:             32 (one per row)
Resolution:        8 bits (256 levels)
Input Range:       0V to 1.0V (after TIA conversion)
Conversion Time:   10 ns (analog to digital latency)
Power per ADC:     0.5 mW (conversion energy)
Total ADC Power:   16 mW (for 32 ADCs)
ENOB:              7.5 bits (effective resolution with noise)
SNR:               46 dB (signal-to-noise ratio)
Sample Rate:       100 MSPS (samples per second)`

	helpText := widget.NewLabel("ADC digitizes analog current from TIA, converting continuous values to discrete digital levels")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Resolution:"), ca.specADCBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel(""), // Spacing
		widget.NewLabel(specs),
		widget.NewLabel(""), // Spacing
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecTIASection() fyne.CanvasObject {
	tiaGainOptions := []string{"1", "10", "100"}
	ca.specTIAGainSelect = widget.NewSelect(tiaGainOptions, nil)
	ca.specTIAGainSelect.SetSelected("10")

	specs := `Count:             32 (one per row)
Gain (R_f):        10 kOhm (transimpedance gain)
Bandwidth:         100 MHz (frequency response)
Input Current:     0 to 100 µA (cell current range)
Output Voltage:    0 to 1.0 V (V_out = I_in × R_f)
Noise:             < 1 µA RMS (input-referred noise)
Response Time:     ~2 ns (settling time)`

	helpText := widget.NewLabel("TIA (Transimpedance Amplifier) converts tiny FeFET currents to measurable voltages for ADC")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Gain:"), ca.specTIAGainSelect, widget.NewLabel("kOhm")),
		widget.NewLabel(""), // Spacing
		widget.NewLabel(specs),
		widget.NewLabel(""), // Spacing
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecFeFETSection() fyne.CanvasObject {
	grid := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Material:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("HfZrO2 (HZO)"),

		widget.NewLabelWithStyle("Thickness:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("10 nm (ferroelectric layer)"),

		widget.NewLabelWithStyle("Levels:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("30 discrete states (~4.9 bits/cell)"),

		widget.NewLabelWithStyle("Conductance:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("1 µS to 100 µS (programmable range)"),

		widget.NewLabelWithStyle("Read Voltage:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("0.5 V (non-destructive, below write threshold)"),

		widget.NewLabelWithStyle("Write Voltage:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("2.0 V to 5.0 V (exceeds coercive field Ec)"),

		widget.NewLabelWithStyle("Write Time:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("50 ns (pulse duration for polarization switching)"),

		widget.NewLabelWithStyle("Endurance:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("10^12 cycles (write/erase lifetime)"),

		widget.NewLabelWithStyle("Retention:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("10 years (data persistence without power)"),

		widget.NewLabelWithStyle("Cell Size:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("~0.01 µm² (width × height in silicon area)"),
	)

	helpText := widget.NewLabel("Note: Rise/fall times typically 2-10 ns; capacitance 0.1-10 pF; leakage < 1 nW per cell")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		grid,
		widget.NewLabel(""), // Spacing
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecSummarySection() fyne.CanvasObject {
	// Calculate initial summary based on default size (32x32)
	size := 32
	cells := size * size
	throughput := float64(cells) / 20.0 // MACs per ns = GOPS

	summary := fmt.Sprintf(`Component       | Count | Power   | Area     | Latency
----------------|-------|---------|----------|--------
FeFET Array     | %d | 0.1 mW  | 0.01 mm² | 5 ns
DACs            | %d    | 3.2 mW  | 0.02 mm² | 5 ns
TIAs            | %d    | 1.6 mW  | 0.01 mm² | 2 ns
ADCs            | %d    | 16 mW   | 0.04 mm² | 10 ns
Control         | 1     | 0.5 mW  | 0.01 mm² | 2 ns
----------------|-------|---------|----------|--------
TOTAL           |       | 21.4 mW | 0.09 mm² | 20 ns

Throughput:     %d MACs / 20ns = %.1f GOPS
Efficiency:     %.1f GOPS / 21.4 mW = %d GOPS/W`,
		cells, size, size, size,
		cells, throughput, throughput, int(throughput*1000/21.4))

	ca.specSummaryLabel = widget.NewLabel(summary)
	return ca.specSummaryLabel
}

func (ca *CircuitsApp) updateSpecSummary() {
	if ca.specSummaryLabel == nil || ca.specArraySizeSelect == nil {
		return
	}

	// Get current array size
	var size int
	fmt.Sscanf(ca.specArraySizeSelect.Selected, "%d", &size)
	if size == 0 {
		size = 32 // default
	}

	cells := size * size
	throughput := float64(cells) / 20.0 // MACs per ns = GOPS

	summary := fmt.Sprintf(`Component       | Count | Power   | Area     | Latency
----------------|-------|---------|----------|--------
FeFET Array     | %d | 0.1 mW  | 0.01 mm² | 5 ns
DACs            | %d    | 3.2 mW  | 0.02 mm² | 5 ns
TIAs            | %d    | 1.6 mW  | 0.01 mm² | 2 ns
ADCs            | %d    | 16 mW   | 0.04 mm² | 10 ns
Control         | 1     | 0.5 mW  | 0.01 mm² | 2 ns
----------------|-------|---------|----------|--------
TOTAL           |       | 21.4 mW | 0.09 mm² | 20 ns

Throughput:     %d MACs / 20ns = %.1f GOPS
Efficiency:     %.1f GOPS / 21.4 mW = %d GOPS/W`,
		cells, size, size, size,
		cells, throughput, throughput, int(throughput*1000/21.4))

	ca.specSummaryLabel.SetText(summary)
}

func (ca *CircuitsApp) onExportSpecs() {
	// Show "Export not implemented" message in status
	fyne.Do(func() {
		ca.specStatusLabel.SetText("Export not implemented - copy specs from display or take screenshot")
	})
}

func (ca *CircuitsApp) onCompareToGPU() {
	// Show comparison summary in status label
	fyne.Do(func() {
		ca.specStatusLabel.SetText("FeFET vs GPU: 25x faster (20ns vs 500ns), 2000x more efficient (2392 vs ~5 GOPS/W), 100x smaller area")
	})
}
