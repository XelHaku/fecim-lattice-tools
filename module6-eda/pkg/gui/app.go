// pkg/gui/app.go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui/tabs"
)

// CreateMainWindow creates the main application window
func CreateMainWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Demo 6: FeCIM Design Suite (Preview)")
	w.Resize(fyne.NewSize(1200, 800))

	// Shared state
	state := &tabs.AppState{}

	// Create tab contents (pre-loaded to avoid layout cascades on Wayland/Sway)
	compilerContent := tabs.MakeCompilerTab(state, w)
	layoutContent := tabs.MakeLayoutTab(state)
	hdlContent := tabs.MakeHDLTab(state, w)       // Phase 3: HDL Generation
	explorerContent := makePlaceholderTab("Design space explorer coming soon")
	simulateContent := makePlaceholderTab("Simulation bridge coming soon")
	exportContent := tabs.MakeExportTab(state, w)
	learnContent := tabs.MakeLearnTab(state, w)   // Learning Center with OpenLane docs

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

func makePlaceholderTab(message string) fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel(message))
}
