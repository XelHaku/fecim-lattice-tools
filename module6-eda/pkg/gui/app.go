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

	// Create tab contents
	compilerContent := tabs.MakeCompilerTab(state, w)
	layoutContent := tabs.MakeLayoutTab(state)
	explorerContent := makePlaceholderTab("Design space explorer coming soon")
	simulateContent := makePlaceholderTab("Simulation bridge coming soon")
	exportContent := tabs.MakeExportTab(state, w)
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
