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
	// Create tab contents (pre-loaded to avoid layout cascades on Wayland/Sway)
	compilerContent := tabs.MakeCompilerTab(app.state, window)
	layoutContent := tabs.MakeLayoutTab(app.state)
	hdlContent := tabs.MakeHDLTab(app.state, window)       // Phase 3: HDL Generation
	explorerContent := makePlaceholderTab("Design space explorer coming soon")
	simulateContent := makePlaceholderTab("Simulation bridge coming soon")
	exportContent := tabs.MakeExportTab(app.state, window)
	learnContent := tabs.MakeLearnTab(app.state, window)   // Learning Center with OpenLane docs

	// All views for Hide/Show toggling
	viewNames := []string{"Compiler", "Layout", "HDL", "Explorer", "Simulate", "Export", "Learn"}
	allViews := []fyne.CanvasObject{
		compilerContent, layoutContent, hdlContent,
		explorerContent, simulateContent, exportContent, learnContent,
	}

	// View selector (replaces nested tabs to save space)
	viewSelector := widget.NewSelect(viewNames, nil)
	viewSelector.SetSelected("Compiler")

	// Content container using Stack - all views layered, visibility toggled
	contentContainer := container.NewStack(allViews...)

	// Track current view
	currentView := ""

	// Update view based on selection using Hide/Show (avoids layout cascades)
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
	currentView = "Compiler"

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
