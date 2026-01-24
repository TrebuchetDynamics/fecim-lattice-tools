// pkg/gui/tabs/array_builder_tab.go
package tabs

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
)

// MakeArrayBuilderTab creates Tab 2: Array Builder
// Allows user to configure array dimensions and mode
func MakeArrayBuilderTab(cfg *config.ArrayConfig) fyne.CanvasObject {
	// Dimension inputs
	rowsEntry := widget.NewEntry()
	rowsEntry.SetText(fmt.Sprintf("%d", cfg.Rows))

	colsEntry := widget.NewEntry()
	colsEntry.SetText(fmt.Sprintf("%d", cfg.Cols))

	// Mode selection
	modeSelect := widget.NewSelect(
		[]string{"storage", "memory", "compute"},
		func(s string) { cfg.Mode = s },
	)
	modeSelect.SetSelected(cfg.Mode)

	// Architecture selection
	archSelect := widget.NewSelect(
		[]string{"passive", "1t1r"},
		func(s string) { cfg.Architecture = s },
	)
	archSelect.SetSelected(cfg.Architecture)

	// Statistics labels
	totalLabel := widget.NewLabel(fmt.Sprintf("Total Cells: %d", cfg.Rows*cfg.Cols))
	areaLabel := widget.NewLabel(fmt.Sprintf("Array Area: %.2f μm²", float64(cfg.Rows*cfg.Cols)*cfg.CellWidth*cfg.CellHeight))
	wlLengthLabel := widget.NewLabel(fmt.Sprintf("WL Length: %.2f μm", float64(cfg.Cols)*cfg.CellWidth))
	blLengthLabel := widget.NewLabel(fmt.Sprintf("BL Length: %.2f μm", float64(cfg.Rows)*cfg.CellHeight))

	// Update statistics
	updateStats := func() {
		rows, _ := strconv.Atoi(rowsEntry.Text)
		cols, _ := strconv.Atoi(colsEntry.Text)
		cfg.Rows = rows
		cfg.Cols = cols

		total := rows * cols
		area := float64(total) * cfg.CellWidth * cfg.CellHeight
		wlLength := float64(cols) * cfg.CellWidth
		blLength := float64(rows) * cfg.CellHeight

		totalLabel.SetText(fmt.Sprintf("Total Cells: %d", total))
		areaLabel.SetText(fmt.Sprintf("Array Area: %.2f μm²", area))
		wlLengthLabel.SetText(fmt.Sprintf("WL Length: %.2f μm", wlLength))
		blLengthLabel.SetText(fmt.Sprintf("BL Length: %.2f μm", blLength))
	}

	rowsEntry.OnChanged = func(s string) { updateStats() }
	colsEntry.OnChanged = func(s string) { updateStats() }

	applyBtn := widget.NewButton("Apply Configuration", func() {
		updateStats()
	})

	// Mode descriptions
	modeDesc := widget.NewLabel("")
	modeDesc.Wrapping = fyne.TextWrapWord
	updateModeDesc := func(mode string) {
		switch mode {
		case "storage":
			modeDesc.SetText("NAND Flash Replacement\n• High-density non-volatile storage\n• 30 levels/cell (~4.9 bits)\n• 10+ year retention")
		case "memory":
			modeDesc.SetText("DRAM Replacement\n• High-speed zero-refresh memory\n• ~10ns access time\n• Non-volatile operation")
		case "compute":
			modeDesc.SetText("AI Accelerator\n• Analog compute-in-memory\n• Matrix-vector multiplication\n• Optional pre-loaded weights")
		}
	}
	modeSelect.OnChanged = func(s string) {
		cfg.Mode = s
		updateModeDesc(s)
	}
	updateModeDesc(cfg.Mode)

	// Configuration form
	configForm := widget.NewForm(
		widget.NewFormItem("Rows", rowsEntry),
		widget.NewFormItem("Columns", colsEntry),
		widget.NewFormItem("Mode", modeSelect),
		widget.NewFormItem("Architecture", archSelect),
	)

	// Left panel
	leftPanel := container.NewVBox(
		widget.NewLabel("Array Dimensions"),
		widget.NewSeparator(),
		configForm,
		applyBtn,
		widget.NewSeparator(),
		widget.NewLabel("Statistics"),
		totalLabel,
		areaLabel,
		wlLengthLabel,
		blLengthLabel,
	)

	// Right panel with mode description
	rightPanel := container.NewVBox(
		widget.NewLabel("Operation Mode"),
		widget.NewSeparator(),
		modeDesc,
		widget.NewSeparator(),
		widget.NewLabel("Array Visualization"),
		makeSimpleArrayGrid(cfg),
	)

	return container.NewHSplit(leftPanel, rightPanel)
}

// makeSimpleArrayGrid creates a simple text representation of the array
func makeSimpleArrayGrid(cfg *config.ArrayConfig) fyne.CanvasObject {
	gridText := fmt.Sprintf("    BL[0]  BL[1]  BL[2]  BL[3]\n")
	gridText += fmt.Sprintf("      │      │      │      │\n")
	
	rows := cfg.Rows
	if rows > 8 {
		rows = 8 // Limit display to 8 rows
	}
	
	for i := 0; i < rows; i++ {
		gridText += fmt.Sprintf("WL[%d] ●──────●──────●──────●\n", i)
		gridText += fmt.Sprintf("      │      │      │      │\n")
	}
	
	if cfg.Rows > 8 {
		gridText += fmt.Sprintf("      ... (%d more rows)\n", cfg.Rows-8)
	}

	label := widget.NewLabel(gridText)
	label.TextStyle.Monospace = true
	return label
}
