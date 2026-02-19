// pkg/gui/app.go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/gui/tabs"
)

// CreateMainWindow creates the main application window
func CreateMainWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Module 6: FeCIM Design Suite - EDA")
	w.Resize(fyne.NewSize(1600, 1000))

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
	builderContent := tabs.MakeBuilderValidationTab(arrayConfig, w) // Tab 1
	exportViewerContent := tabs.MakeExportViewerTab(arrayConfig, w) // Tab 2
	layoutVisualizerContent := tabs.MakeLayoutVisualizerTab(arrayConfig, w)
	learnContent := tabs.MakeLearnTab(nil, w) // Tab 4

	// View names for selector
	viewNames := []string{
		"1. Builder & Validation",
		"2. Export Viewer",
		"3. Layout Visualizer",
		"4. Learn",
	}

	allViews := []fyne.CanvasObject{
		builderContent,
		exportViewerContent,
		layoutVisualizerContent,
		learnContent,
	}

	// View selector dropdown
	viewSelector := widget.NewSelect(viewNames, nil)
	viewSelector.SetSelected("1. Builder & Validation")

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
	currentView = "1. Builder & Validation"

	// Header with inline view selector
	banner := widget.NewLabel("Educational outputs for OpenLane/SKY130 (simulation-only; not tapeout-ready)")
	banner.Alignment = fyne.TextAlignCenter
	banner.Truncation = fyne.TextTruncateEllipsis

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

	// Setup keyboard shortcuts
	SetupKeyboard(w, viewSelector)

	return w
}
