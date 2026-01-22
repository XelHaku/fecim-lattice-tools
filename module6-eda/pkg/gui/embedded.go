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
	// Create tab contents
	compilerContent := tabs.MakeCompilerTab(app.state, window)
	layoutContent := tabs.MakeLayoutTab(app.state)
	explorerContent := makePlaceholderTab("Design space explorer coming soon")
	simulateContent := makePlaceholderTab("Simulation bridge coming soon")
	exportContent := tabs.MakeExportTab(app.state, window)
	learnContent := makePlaceholderTab("Learning resources coming soon")

	// View selector (replaces nested tabs to save space)
	viewSelector := widget.NewSelect(
		[]string{"Compiler", "Layout", "Explorer", "Simulate", "Export", "Learn"},
		nil,
	)
	viewSelector.SetSelected("Compiler")

	// Content container
	contentContainer := container.NewMax(compilerContent)

	// Update view based on selection
	viewSelector.OnChanged = func(view string) {
		switch view {
		case "Compiler":
			contentContainer.Objects[0] = compilerContent
		case "Layout":
			contentContainer.Objects[0] = layoutContent
		case "Explorer":
			contentContainer.Objects[0] = explorerContent
		case "Simulate":
			contentContainer.Objects[0] = simulateContent
		case "Export":
			contentContainer.Objects[0] = exportContent
		case "Learn":
			contentContainer.Objects[0] = learnContent
		}
		contentContainer.Refresh()
	}

	// Header with inline view selector
	banner := widget.NewLabel("PREVIEW: Bridge to open-source EDA tools (ngspice, KLayout, CiMLoop)")
	banner.Alignment = fyne.TextAlignCenter

	headerRow := container.NewHBox(
		widget.NewLabel("View:"),
		viewSelector,
		widget.NewSeparator(),
		banner,
	)

	header := container.NewVBox(
		headerRow,
		widget.NewSeparator(),
	)

	app.content = container.NewBorder(header, nil, nil, nil, contentContainer)
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
