// Package gui provides Fyne-based GUI components for crossbar visualization.
// tabbed_app.go provides the enhanced 4-tab interface combining MVM and non-idealities.
package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/crossbar"
	"multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/gui/tabs"
)

// TabbedCrossbarApp is the enhanced crossbar demo with 4 tabs.
type TabbedCrossbarApp struct {
	fyneApp fyne.App
	window  fyne.Window
	array   *crossbar.Array

	// Tabs
	idealTab  *tabs.IdealTab
	irdropTab *tabs.IRDropTab
	sneakTab  *tabs.SneakTab
	driftTab  *tabs.DriftTab

	// UI
	tabContainer *container.AppTabs
	statusLabel  *widget.Label
}

// NewTabbedCrossbarApp creates a new tabbed crossbar app.
func NewTabbedCrossbarApp() *TabbedCrossbarApp {
	app := &TabbedCrossbarApp{
		statusLabel: widget.NewLabel("Ready - Select a tab to explore"),
	}

	// Initialize crossbar array
	cfg := &crossbar.Config{
		Rows:       16,
		Cols:       16,
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	app.array, _ = crossbar.NewArray(cfg)

	// Initialize tabs
	app.idealTab = tabs.NewIdealTab(app.array, func() {
		// Callback when array changes
	})
	app.irdropTab = tabs.NewIRDropTab(16)
	app.sneakTab = tabs.NewSneakTab(16)
	app.driftTab = tabs.NewDriftTab(16)

	return app
}

// BuildContent creates the UI content (for embedding in unified visualizer).
func (app *TabbedCrossbarApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	app.fyneApp = fyneApp
	app.window = parentWindow

	return app.createMainLayout()
}

// Run starts the standalone application.
func (app *TabbedCrossbarApp) Run() {
	app.window.SetContent(app.createMainLayout())
	app.window.Resize(fyne.NewSize(1200, 800))
	app.window.ShowAndRun()
}

func (app *TabbedCrossbarApp) createMainLayout() fyne.CanvasObject {
	// Header
	header := app.createHeader()

	// Create tabs
	app.tabContainer = container.NewAppTabs(
		container.NewTabItem("1. Ideal MVM", app.idealTab.Content()),
		container.NewTabItem("2. IR Drop Analysis", app.irdropTab.Content()),
		container.NewTabItem("3. Sneak Paths", app.sneakTab.Content()),
		container.NewTabItem("4. Drift & Variation", app.driftTab.Content()),
	)
	app.tabContainer.SetTabLocation(container.TabLocationTop)

	// Tab change callback
	app.tabContainer.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "1. Ideal MVM":
			app.statusLabel.SetText("Ideal MVM - No non-idealities")
		case "2. IR Drop Analysis":
			app.statusLabel.SetText("IR Drop - Voltage drop along metal lines")
		case "3. Sneak Paths":
			app.statusLabel.SetText("Sneak Paths - Parasitic current paths")
		case "4. Drift & Variation":
			app.statusLabel.SetText("Drift - Conductance change over time")
		}
	}

	return container.NewBorder(
		header,
		app.statusLabel,
		nil,
		nil,
		app.tabContainer,
	)
}

func (app *TabbedCrossbarApp) createHeader() fyne.CanvasObject {
	title := canvas.NewText("Demo 2: Crossbar MVM + Non-Idealities", color.White)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Ideal Operation | IR Drop | Sneak Paths | Conductance Drift")
	subtitle.Alignment = fyne.TextAlignCenter

	quote := widget.NewLabel(`"We handle real-world challenges better than competition" — FeCIM Advantage`)
	quote.Alignment = fyne.TextAlignCenter
	quote.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(subtitle),
		container.NewCenter(quote),
		widget.NewSeparator(),
	)
}

// Start is called when the tab is selected in unified visualizer.
func (app *TabbedCrossbarApp) Start() {
	// Nothing to start - tabs are static
}

// Stop is called when the tab is deselected in unified visualizer.
func (app *TabbedCrossbarApp) Stop() {
	// Nothing to stop
}

// EmbeddedTabbedCrossbarApp wraps TabbedCrossbarApp for the unified visualizer.
type EmbeddedTabbedCrossbarApp struct {
	*TabbedCrossbarApp
}

// NewEmbeddedTabbedCrossbarApp creates a new embedded tabbed crossbar app.
func NewEmbeddedTabbedCrossbarApp() *EmbeddedTabbedCrossbarApp {
	return &EmbeddedTabbedCrossbarApp{
		TabbedCrossbarApp: NewTabbedCrossbarApp(),
	}
}

// BuildContent creates the UI content for embedding.
func (e *EmbeddedTabbedCrossbarApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	return e.TabbedCrossbarApp.BuildContent(fyneApp, parentWindow)
}

// Start is called when the tab is selected.
func (e *EmbeddedTabbedCrossbarApp) Start() {
	e.TabbedCrossbarApp.Start()
}

// Stop is called when the tab is deselected.
func (e *EmbeddedTabbedCrossbarApp) Stop() {
	e.TabbedCrossbarApp.Stop()
}

// feCIMTheme for consistent branding (used if running standalone)
type feCIMTabbedTheme struct{}

func (t *feCIMTabbedTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	colorBackground := color.RGBA{0, 50, 100, 255}
	colorPrimary := color.RGBA{0, 212, 255, 255}

	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTabbedTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTabbedTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTabbedTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
