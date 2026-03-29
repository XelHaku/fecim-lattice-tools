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

	// Create 2 tabs - consolidated architecture
	return container.NewAppTabs(
		container.NewTabItem("1. Builder & Validation", tabs.MakeBuilderValidationTab(arrayConfig, window)),
		container.NewTabItem("2. Learn", tabs.MakeLearnTab(nil, window)),
	)
}

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedEDAApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	return app.EmbeddedAppBase.BuildOrReuseContent(fyneApp, window, func() fyne.CanvasObject {
		return CreateModuleContent(window)
	})
}

// Start is called when this demo tab is selected
func (app *EmbeddedEDAApp) Start() {
	app.EmbeddedAppBase.Start()
	logging.GlobalInfo("[EDA] Module started")
	// No background processes to start
}

// Stop is called when this demo tab is deselected
func (app *EmbeddedEDAApp) Stop() {
	logging.GlobalInfo("[EDA] Module stopped")
	// No background processes to stop
	app.EmbeddedAppBase.Stop()
}
