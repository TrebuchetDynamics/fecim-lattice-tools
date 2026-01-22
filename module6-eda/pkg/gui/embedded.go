// pkg/gui/embedded.go
// Embeddable version of the EDA app for the unified visualizer
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

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

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedEDAApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	// Create tabs
	tabContainer := container.NewAppTabs(
		container.NewTabItem("1. Compiler", tabs.MakeCompilerTab(app.state, window)),
		container.NewTabItem("2. Layout", tabs.MakeLayoutTab(app.state)),
		container.NewTabItem("3. Explorer", makePlaceholderTab("Design space explorer coming soon")),
		container.NewTabItem("4. Simulate", makePlaceholderTab("Simulation bridge coming soon")),
		container.NewTabItem("5. Export", tabs.MakeExportTab(app.state, window)),
		container.NewTabItem("6. Learn", makePlaceholderTab("Learning resources coming soon")),
	)
	tabContainer.SetTabLocation(container.TabLocationTop)

	// Add preview banner
	banner := widget.NewLabel("PREVIEW: Bridge to open-source EDA tools (ngspice, KLayout, CiMLoop)")
	banner.Alignment = fyne.TextAlignCenter

	app.content = container.NewBorder(banner, nil, nil, nil, tabContainer)
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
