// pkg/gui/tabs/verilog_export_tab.go
package tabs

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/export"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/validation"
)

// MakeVerilogExportTab creates Tab 3: Verilog Export
// Generates and validates structural Verilog netlist
func MakeVerilogExportTab(cfg *config.ArrayConfig) fyne.CanvasObject {
	// Preview area
	preview := widget.NewMultiLineEntry()
	preview.Wrapping = fyne.TextWrapOff

	// Statistics labels
	statsLabel := widget.NewLabel("Stats: -")
	statusLabel := widget.NewLabel("⏸️ Not generated")

	// Generate button
	generateBtn := widget.NewButton("Generate Verilog", func() {
		content := export.GenerateArrayVerilog(*cfg)
		preview.SetText(content)

		instances := cfg.Rows * cfg.Cols
		lines := strings.Count(content, "\n")
		size := float64(len(content)) / 1024

		statsLabel.SetText(fmt.Sprintf(
			"Stats: Instances: %d | Lines: %d | Size: %.1f KB",
			instances, lines, size,
		))

		filename := fmt.Sprintf("output/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
		os.MkdirAll("output", 0755)
		os.WriteFile(filename, []byte(content), 0644)
		statusLabel.SetText("✅ " + filename)
	})

	// Validate button (requires yosys)
	validateBtn := widget.NewButton("Validate with Yosys", func() {
		arrayPath := fmt.Sprintf("output/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
		cellPath := "cells/fecim_bitcell/fecim_bitcell.v"

		err := validation.ValidateVerilogWithCell(arrayPath, cellPath)
		if err != nil {
			statusLabel.SetText("❌ " + err.Error())
		} else {
			statusLabel.SetText("✅ Yosys validation passed")
		}
	})

	// Copy button
	copyBtn := widget.NewButton("Copy to Clipboard", func() {
		// Note: Fyne clipboard access requires window context
		// This is a simplified version
		statusLabel.SetText("ℹ️ Copy functionality requires window context")
	})

	// Save button (redundant with generate, but for clarity)
	saveBtn := widget.NewButton("Save File", func() {
		if preview.Text == "" {
			statusLabel.SetText("❌ Generate Verilog first")
			return
		}
		filename := fmt.Sprintf("output/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
		os.MkdirAll("output", 0755)
		os.WriteFile(filename, []byte(preview.Text), 0644)
		statusLabel.SetText("✅ Saved to " + filename)
	})

	// Toolbar
	toolbar := container.NewHBox(
		generateBtn,
		validateBtn,
		copyBtn,
		saveBtn,
	)

	// Top section
	topSection := container.NewVBox(
		widget.NewLabel("Verilog Netlist Generator"),
		widget.NewSeparator(),
		toolbar,
		statsLabel,
		statusLabel,
		widget.NewSeparator(),
	)

	return container.NewBorder(
		topSection,
		nil, nil, nil,
		container.NewScroll(preview),
	)
}
