// pkg/gui/tabs/def_export_tab.go
package tabs

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/validation"
)

// MakeDEFExportTab creates Tab 4: DEF Export
// Generates and visualizes physical placement DEF file
func MakeDEFExportTab(cfg *config.ArrayConfig) fyne.CanvasObject {
	// DEF preview
	defPreview := widget.NewMultiLineEntry()
	defPreview.Wrapping = fyne.TextWrapOff

	// Layout visualization (simple text-based)
	layoutViz := widget.NewLabel("Generate DEF to see layout")
	layoutViz.TextStyle.Monospace = true

	statusLabel := widget.NewLabel("⏸️ Not generated")

	// Generate button
	generateBtn := widget.NewButton("Generate DEF", func() {
		content := generateDEF(*cfg)
		defPreview.SetText(content)

		filename := fmt.Sprintf("output/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
		os.MkdirAll("output", 0755)
		os.WriteFile(filename, []byte(content), 0644)

		// Update visualization
		vizText := makeLayoutVisualization(cfg)
		layoutViz.SetText(vizText)

		statusLabel.SetText("✅ " + filename)
	})

	// Validate button
	validateBtn := widget.NewButton("Validate DEF", func() {
		filename := fmt.Sprintf("output/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
		err := validation.ValidateDEF(filename)
		if err != nil {
			statusLabel.SetText("❌ " + err.Error())
		} else {
			stats, _ := validation.GetDEFStats(filename)
			statusLabel.SetText(fmt.Sprintf("✅ Valid DEF - %d components", stats["component_count"]))
		}
	})

	// Toolbar
	toolbar := container.NewHBox(generateBtn, validateBtn)

	// Left: Layout visualization
	leftPanel := container.NewVBox(
		widget.NewLabel("Layout View"),
		widget.NewSeparator(),
		container.NewScroll(layoutViz),
	)

	// Right: DEF text preview
	rightPanel := container.NewVBox(
		widget.NewLabel("DEF Preview"),
		widget.NewSeparator(),
		container.NewScroll(defPreview),
	)

	// Top bar
	topBar := container.NewVBox(
		widget.NewLabel("Physical Placement (DEF)"),
		widget.NewSeparator(),
		toolbar,
		statusLabel,
		widget.NewSeparator(),
	)

	splitView := container.NewHSplit(leftPanel, rightPanel)
	splitView.SetOffset(0.3)

	return container.NewBorder(topBar, nil, nil, nil, splitView)
}

// generateDEF generates DEF content (since we haven't updated the export package yet)
func generateDEF(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	
	dbu := 1000 // Database units per micron
	cellWidthDBU := int(cfg.CellWidth * float64(dbu))
	cellHeightDBU := int(cfg.CellHeight * float64(dbu))
	
	margin := 1000
	dieWidth := cfg.Cols*cellWidthDBU + 2*margin
	dieHeight := cfg.Rows*cellHeightDBU + 2*margin
	
	var content strings.Builder
	content.WriteString(`VERSION 5.8 ;
DIVIDERCHAR "/" ;
BUSBITCHARS "[]" ;
`)
	content.WriteString(fmt.Sprintf("DESIGN %s ;\n", designName))
	content.WriteString(fmt.Sprintf("UNITS DISTANCE MICRONS %d ;\n\n", dbu))
	content.WriteString(fmt.Sprintf("DIEAREA ( 0 0 ) ( %d %d ) ;\n\n", dieWidth, dieHeight))
	
	// Components
	totalCells := cfg.Rows * cfg.Cols
	content.WriteString(fmt.Sprintf("COMPONENTS %d ;\n", totalCells))
	
	for row := 0; row < cfg.Rows; row++ {
		for col := 0; col < cfg.Cols; col++ {
			x := margin + col*cellWidthDBU
			y := margin + row*cellHeightDBU
			content.WriteString(fmt.Sprintf("    - cell_%d_%d fecim_bitcell + FIXED ( %d %d ) N ;\n", row, col, x, y))
		}
	}
	content.WriteString("END COMPONENTS\n\n")
	
	// Pins (simplified)
	numPins := cfg.Rows + cfg.Cols + 2
	content.WriteString(fmt.Sprintf("PINS %d ;\n", numPins))
	content.WriteString("    - VPWR + NET VPWR + DIRECTION INOUT + USE POWER ;\n")
	content.WriteString("    - VGND + NET VGND + DIRECTION INOUT + USE GROUND ;\n")
	for i := 0; i < cfg.Rows; i++ {
		content.WriteString(fmt.Sprintf("    - WL[%d] + NET WL[%d] + DIRECTION INPUT + USE SIGNAL ;\n", i, i))
	}
	for i := 0; i < cfg.Cols; i++ {
		content.WriteString(fmt.Sprintf("    - BL[%d] + NET BL[%d] + DIRECTION OUTPUT + USE SIGNAL ;\n", i, i))
	}
	content.WriteString("END PINS\n\n")
	
	content.WriteString("END DESIGN\n")
	return content.String()
}

func makeLayoutVisualization(cfg *config.ArrayConfig) string {
	var viz strings.Builder
	
	rows := cfg.Rows
	cols := cfg.Cols
	if rows > 12 {
		rows = 12
	}
	if cols > 8 {
		cols = 8
	}
	
	viz.WriteString(fmt.Sprintf("FeCIM Crossbar %dx%d Layout\n\n", cfg.Rows, cfg.Cols))
	
	for r := 0; r < rows; r++ {
		viz.WriteString(fmt.Sprintf("WL[%d] ", r))
		for c := 0; c < cols; c++ {
			viz.WriteString("┌─┐")
		}
		viz.WriteString("\n     ")
		for c := 0; c < cols; c++ {
			viz.WriteString("│ │")
		}
		viz.WriteString("\n     ")
		for c := 0; c < cols; c++ {
			viz.WriteString("└─┘")
		}
		viz.WriteString("\n")
	}
	
	if cfg.Rows > 12 {
		viz.WriteString(fmt.Sprintf("     ... (%d more rows)\n", cfg.Rows-12))
	}
	
	viz.WriteString("     ")
	for c := 0; c < cols; c++ {
		viz.WriteString(fmt.Sprintf("BL%d ", c))
	}
	viz.WriteString("\n")
	
	return viz.String()
}
