// pkg/gui/app.go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"demo6-eda/pkg/gui/tabs"
)

// CreateMainWindow creates the main application window
func CreateMainWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Demo 6: FeCIM Design Suite (Preview)")
	w.Resize(fyne.NewSize(1200, 800))

	// Shared state
	state := &tabs.AppState{}

	// Create tabs
	tabContainer := container.NewAppTabs(
		container.NewTabItem("1. Compiler", tabs.MakeCompilerTab(state, w)),
		container.NewTabItem("2. Layout", tabs.MakeLayoutTab(state)),
		container.NewTabItem("3. Explorer", makePlaceholderTab("Design space explorer coming soon")),
		container.NewTabItem("4. Simulate", makePlaceholderTab("Simulation bridge coming soon")),
		container.NewTabItem("5. Export", tabs.MakeExportTab(state, w)),
		container.NewTabItem("6. Learn", makePlaceholderTab("Learning resources coming soon")),
	)
	tabContainer.SetTabLocation(container.TabLocationTop)

	// Add preview banner
	banner := widget.NewLabel("PREVIEW: Bridge to open-source EDA tools (ngspice, KLayout, CiMLoop)")
	banner.Alignment = fyne.TextAlignCenter

	content := container.NewBorder(banner, nil, nil, nil, tabContainer)
	w.SetContent(content)

	return w
}

func makePlaceholderTab(message string) fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel(message))
}
