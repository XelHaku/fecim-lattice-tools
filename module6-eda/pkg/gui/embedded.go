// pkg/gui/embedded.go
// Embeddable version of the EDA app for the unified visualizer
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui/tabs"
)

// EmbeddedEDAApp is the embeddable version of the EDA app
type EmbeddedEDAApp struct {
	content fyne.CanvasObject
}

// NewEmbeddedEDAApp creates a new embedded EDA app instance
func NewEmbeddedEDAApp() *EmbeddedEDAApp {
	return &EmbeddedEDAApp{}
}

// CreateModuleContent creates the embedded module6 content
func CreateModuleContent(window fyne.Window) fyne.CanvasObject {
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

	// Create 2 tabs - consolidated architecture
	return container.NewAppTabs(
		container.NewTabItem("1. Builder & Validation", tabs.MakeBuilderValidationTab(arrayConfig, window)),
		container.NewTabItem("2. Learn", tabs.MakeLearnTab(nil, window)),
	)
}

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedEDAApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	// Use CreateModuleContent with full Learn tab
	app.content = CreateModuleContent(window)
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
