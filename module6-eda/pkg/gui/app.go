// pkg/gui/app.go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui/tabs"
)

// CreateMainWindow creates the main application window
func CreateMainWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Module 6: FeCIM Design Suite - EDA")
	w.Resize(fyne.NewSize(1400, 900))

	// Shared array configuration (used across tabs 2-7)
	arrayConfig := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	// Create tab contents
	cellBuilderContent := tabs.MakeCellBuilderTab()                   // Tab 1
	arrayBuilderContent := tabs.MakeArrayBuilderTab(arrayConfig)     // Tab 2
	verilogContent := tabs.MakeVerilogExportTab(arrayConfig)         // Tab 3
	defContent := tabs.MakeDEFExportTab(arrayConfig)                 // Tab 4
	validationContent := tabs.MakeValidationTab(arrayConfig)         // Tab 5
	learnContent := tabs.MakeLearnTab(&tabs.AppState{}, w)           // Tab 6 (existing)
	exportAllContent := tabs.MakeExportAllTab(arrayConfig)           // Tab 7

	// View names for selector
	viewNames := []string{
		"1. Cell Builder",
		"2. Array Builder",
		"3. Verilog Export",
		"4. DEF Export",
		"5. Validation",
		"6. Learn",
		"7. Export All",
	}
	
	allViews := []fyne.CanvasObject{
		cellBuilderContent,
		arrayBuilderContent,
		verilogContent,
		defContent,
		validationContent,
		learnContent,
		exportAllContent,
	}

	// View selector dropdown
	viewSelector := widget.NewSelect(viewNames, nil)
	viewSelector.SetSelected("1. Cell Builder")

	// Content container using Stack
	contentContainer := container.NewStack(allViews...)

	// Track current view
	currentView := ""

	// Update view based on selection
	viewSelector.OnChanged = func(view string) {
		if view == currentView {
			return
		}
		currentView = view

		// Hide all views, then show selected
		for i, v := range allViews {
			if viewNames[i] == view {
				v.Show()
			} else {
				v.Hide()
			}
		}
	}

	// Initialize: show first view, hide others
	for i, v := range allViews {
		if i == 0 {
			v.Show()
		} else {
			v.Hide()
		}
	}
	currentView = "1. Cell Builder"

	// Header with inline view selector
	banner := widget.NewLabel("Generate fabrication-ready files for OpenLane/SKY130")
	banner.Alignment = fyne.TextAlignCenter

	headerRow := container.NewHBox(
		widget.NewLabel("View:"),
		viewSelector,
		layout.NewSpacer(),
		banner,
	)

	header := container.NewVBox(
		headerRow,
		widget.NewSeparator(),
	)

	content := container.NewBorder(header, nil, nil, nil, contentContainer)
	w.SetContent(content)

	return w
}

