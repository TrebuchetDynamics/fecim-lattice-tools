// pkg/gui/embedded.go
// Embeddable version of the EDA app for the unified visualizer
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui/tabs"
)

// EmbeddedEDAApp is the embeddable version of the EDA app
type EmbeddedEDAApp struct {
	state   *tabs.AppState
	content fyne.CanvasObject
}

// NewEmbeddedEDAApp creates a new embedded EDA app instance
func NewEmbeddedEDAApp() *EmbeddedEDAApp {
	return &EmbeddedEDAApp{
		state: &tabs.AppState{},
	}
}

// CreateModuleContent creates the embedded module6 content
func CreateModuleContent() fyne.CanvasObject {
	// Shared array configuration
	arrayConfig := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	// Create 7 tabs matching new architecture
	return container.NewAppTabs(
		container.NewTabItem("1. Cell Builder", tabs.MakeCellBuilderTab()),
		container.NewTabItem("2. Array Builder", tabs.MakeArrayBuilderTab(arrayConfig)),
		container.NewTabItem("3. Verilog Export", tabs.MakeVerilogExportTab(arrayConfig)),
		container.NewTabItem("4. DEF Export", tabs.MakeDEFExportTab(arrayConfig)),
		container.NewTabItem("5. Validation", tabs.MakeValidationTab(arrayConfig)),
		container.NewTabItem("6. Learn", createLearnTabStub()),
		container.NewTabItem("7. Export All", tabs.MakeExportAllTab(arrayConfig)),
	)
}

// createLearnTabStub creates a simplified Learn tab for embedded mode
func createLearnTabStub() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Learn tab - see standalone module6-eda app"))
}

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedEDAApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	// Use the simplified CreateModuleContent for embedded view
	app.content = CreateModuleContent()
	return app.content
}



// Start is called when this demo tab is selected
func (app *EmbeddedEDAApp) Start() {
	// No background processes to start
}

// Stop is called when this demo tab is deselected
func (app *EmbeddedEDAApp) Stop() {
	// No background processes to stop
}
