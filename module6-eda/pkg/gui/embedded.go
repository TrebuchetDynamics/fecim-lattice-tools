// pkg/gui/embedded.go
// Embeddable version of the EDA app for the unified visualizer
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/gui/tabs"
	"fecim-lattice-tools/shared/logging"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

var log = logging.NewLogger("eda")

// EmbeddedEDAApp is the embeddable version of the EDA app
type EmbeddedEDAApp struct {
	sharedwidgets.EmbeddedAppBase
}

// NewEmbeddedEDAApp creates a new embedded EDA app instance
func NewEmbeddedEDAApp() *EmbeddedEDAApp {
	logging.GlobalDebug("[EDA] NewEmbeddedEDAApp created")
	return &EmbeddedEDAApp{}
}

// CreateModuleContent creates the embedded module6 content
func CreateModuleContent(window fyne.Window) fyne.CanvasObject {
	logging.GlobalInfo("[EDA] Creating module content")

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

	logging.GlobalDebug("[EDA] Module content created with %dx%d array config", arrayConfig.Rows, arrayConfig.Cols)

	// All 4 views, matching the standalone EDA GUI (app.go)
	appTabs := container.NewAppTabs(
		container.NewTabItem("1. Builder & Validation", tabs.MakeBuilderValidationTab(arrayConfig, window)),
		container.NewTabItem("2. Layout Visualizer", tabs.MakeLayoutVisualizerTab(arrayConfig, window)),
		container.NewTabItem("3. Learn", tabs.MakeLearnTab(nil, window)),
		container.NewTabItem("4. Flow Scripts", tabs.MakeFlowScriptsTab(arrayConfig, window)),
	)
	appTabs.SetTabLocation(container.TabLocationTop)
	// Re-render tab content when switching tabs. This is required because the
	// tab content widgets are constructed (and SetText called) before they have
	// been laid out in a window. The OnSelected callback guarantees Fyne
	// performs a full layout+paint pass on the newly-visible content.
	appTabs.OnSelected = func(tab *container.TabItem) {
		if tab != nil && tab.Content != nil {
			tab.Content.Refresh()
		}
	}
	return appTabs
}

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedEDAApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	// Use CreateModuleContent with full Learn tab
	app.EmbeddedAppBase.Init(fyneApp, window)
	content := CreateModuleContent(window)
	app.SetContent(content)
	return content
}

// Start is called when this demo tab is selected
func (app *EmbeddedEDAApp) Start() {
	logging.GlobalInfo("[EDA] Module started")
	app.EmbeddedAppBase.Start()
	// No background processes to start
}

// Stop is called when this demo tab is deselected
func (app *EmbeddedEDAApp) Stop() {
	logging.GlobalInfo("[EDA] Module stopped")
	// No background processes to stop
	app.EmbeddedAppBase.Stop()
}
