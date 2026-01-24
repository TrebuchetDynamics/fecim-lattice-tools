// pkg/gui/tabs/cell_builder_tab.go
package tabs

import (
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/export"
)

// MakeCellBuilderTab creates Tab 1: Cell Builder
// Allows user to design fecim_bitcell and generate LEF/LIB/V files
func MakeCellBuilderTab() fyne.CanvasObject {
	// Input fields for cell configuration
	nameEntry := widget.NewEntry()
	nameEntry.SetText("fecim_bitcell")

	widthEntry := widget.NewEntry()
	widthEntry.SetText("0.460")

	heightEntry := widget.NewEntry()
	heightEntry.SetText("2.720")

	cellTypeSelect := widget.NewSelect([]string{"passive", "1t1r"}, nil)
	cellTypeSelect.SetSelected("passive")

	riseEntry := widget.NewEntry()
	riseEntry.SetText("0.1")

	fallEntry := widget.NewEntry()
	fallEntry.SetText("0.1")

	capEntry := widget.NewEntry()
	capEntry.SetText("0.002")

	leakageEntry := widget.NewEntry()
	leakageEntry.SetText("0.001")

	// Preview boxes for generated files
	lefPreview := widget.NewMultiLineEntry()
	lefPreview.Wrapping = fyne.TextWrapOff

	libPreview := widget.NewMultiLineEntry()
	libPreview.Wrapping = fyne.TextWrapOff

	verilogPreview := widget.NewMultiLineEntry()
	verilogPreview.Wrapping = fyne.TextWrapOff

	// Status label
	statusLabel := widget.NewLabel("⏸️ Not generated")

	// Helper to parse config from inputs
	getCellConfig := func() config.CellConfig {
		width, _ := strconv.ParseFloat(widthEntry.Text, 64)
		height, _ := strconv.ParseFloat(heightEntry.Text, 64)
		rise, _ := strconv.ParseFloat(riseEntry.Text, 64)
		fall, _ := strconv.ParseFloat(fallEntry.Text, 64)
		cap, _ := strconv.ParseFloat(capEntry.Text, 64)
		leakage, _ := strconv.ParseFloat(leakageEntry.Text, 64)

		return config.CellConfig{
			Name:         nameEntry.Text,
			Width:        width,
			Height:       height,
			CellType:     cellTypeSelect.Selected,
			Technology:   "sky130",
			RiseTime:     rise,
			FallTime:     fall,
			InputCap:     cap,
			LeakagePower: leakage,
		}
	}

	// Generate all files button
	generateAllBtn := widget.NewButton("Generate All Files", func() {
		cfg := getCellConfig()

		// Generate content
		lefContent := export.GenerateLEF(cfg)
		libContent := export.GenerateLiberty(cfg)
		verilogContent := export.GenerateCellVerilog(cfg)

		// Update previews
		lefPreview.SetText(lefContent)
		libPreview.SetText(libContent)
		verilogPreview.SetText(verilogContent)

		// Write files to disk
		dir := "cells/fecim_bitcell"
		os.MkdirAll(dir, 0755)
		os.WriteFile(dir+"/fecim_bitcell.lef", []byte(lefContent), 0644)
		os.WriteFile(dir+"/fecim_bitcell.lib", []byte(libContent), 0644)
		os.WriteFile(dir+"/fecim_bitcell.v", []byte(verilogContent), 0644)

		statusLabel.SetText("✅ All files generated in " + dir + "/")
	})

	// Individual generate buttons
	genLEFBtn := widget.NewButton("Generate LEF", func() {
		cfg := getCellConfig()
		content := export.GenerateLEF(cfg)
		lefPreview.SetText(content)
		statusLabel.SetText("✅ LEF generated (preview only)")
	})

	genLIBBtn := widget.NewButton("Generate LIB", func() {
		cfg := getCellConfig()
		content := export.GenerateLiberty(cfg)
		libPreview.SetText(content)
		statusLabel.SetText("✅ LIB generated (preview only)")
	})

	genVBtn := widget.NewButton("Generate V", func() {
		cfg := getCellConfig()
		content := export.GenerateCellVerilog(cfg)
		verilogPreview.SetText(content)
		statusLabel.SetText("✅ Verilog generated (preview only)")
	})

	// Configuration form
	configForm := widget.NewForm(
		widget.NewFormItem("Cell Name", nameEntry),
		widget.NewFormItem("Width (μm)", widthEntry),
		widget.NewFormItem("Height (μm)", heightEntry),
		widget.NewFormItem("Cell Type", cellTypeSelect),
		widget.NewFormItem("Rise Time (ns)", riseEntry),
		widget.NewFormItem("Fall Time (ns)", fallEntry),
		widget.NewFormItem("Input Cap (pF)", capEntry),
		widget.NewFormItem("Leakage (nW)", leakageEntry),
	)

	// Calculate area display
	areaLabel := widget.NewLabel(fmt.Sprintf("Area: %.4f μm²", 0.46*2.72))
	widthEntry.OnChanged = func(s string) {
		w, _ := strconv.ParseFloat(s, 64)
		h, _ := strconv.ParseFloat(heightEntry.Text, 64)
		areaLabel.SetText(fmt.Sprintf("Area: %.4f μm²", w*h))
	}
	heightEntry.OnChanged = func(s string) {
		w, _ := strconv.ParseFloat(widthEntry.Text, 64)
		h, _ := strconv.ParseFloat(s, 64)
		areaLabel.SetText(fmt.Sprintf("Area: %.4f μm²", w*h))
	}

	// Button row
	buttonRow := container.NewHBox(
		genLEFBtn,
		genLIBBtn,
		genVBtn,
	)

	// Left panel
	leftPanel := container.NewVBox(
		widget.NewLabel("Cell Configuration"),
		widget.NewSeparator(),
		configForm,
		areaLabel,
		widget.NewSeparator(),
		buttonRow,
		generateAllBtn,
		widget.NewSeparator(),
		statusLabel,
	)

	// Preview tabs
	previewTabs := container.NewAppTabs(
		container.NewTabItem("LEF", container.NewScroll(lefPreview)),
		container.NewTabItem("LIB", container.NewScroll(libPreview)),
		container.NewTabItem("Verilog", container.NewScroll(verilogPreview)),
	)

	// Split layout
	return container.NewHSplit(leftPanel, previewTabs)
}
