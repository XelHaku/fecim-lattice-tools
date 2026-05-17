//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for crossbar visualization.
// tabbed_app.go provides the enhanced 4-tab interface combining MVM and non-idealities.
package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module2-crossbar/pkg/gui/tabs"
	"fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/logging"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// TabbedCrossbarApp is the enhanced crossbar demo with 4 tabs.
type TabbedCrossbarApp struct {
	fyneApp fyne.App
	window  fyne.Window
	array   *crossbar.Array

	// Tabs
	idealTab   *tabs.IdealTab
	irdropTab  *tabs.IRDropTab
	sneakTab   *tabs.SneakTab
	driftTab   *tabs.DriftTab
	fecapTab   *tabs.FeCAPTab
	stack3dTab *tabs.Stack3DTab

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
	var err error
	app.array, err = crossbar.NewArray(cfg)
	if err != nil {
		logging.Printf("[module2-crossbar] failed to initialize default array config, using minimal fallback: %v", err)
		app.array, err = crossbar.NewArray(&crossbar.Config{Rows: 1, Cols: 1, NoiseLevel: 0, ADCBits: 6, DACBits: 8})
		if err != nil {
			logging.Printf("[module2-crossbar] failed to initialize fallback array: %v", err)
		}
	}

	// Initialize tabs
	app.idealTab = tabs.NewIdealTab(app.array, func() {
		// Callback when array changes
	})
	app.irdropTab = tabs.NewIRDropTab(16)
	app.sneakTab = tabs.NewSneakTab(16)
	app.driftTab = tabs.NewDriftTab(16)
	app.fecapTab = tabs.NewFeCAPTab(8) // 8x8 for responsive display
	app.stack3dTab = tabs.NewStack3DTab()

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
	// Pre-load all content views (avoids layout cascades on Wayland/Sway)
	idealContent := app.idealTab.Content()
	irdropContent := app.irdropTab.Content()
	sneakContent := app.sneakTab.Content()
	driftContent := app.driftTab.Content()
	fecapContent := app.fecapTab.Content()
	stack3dContent := app.stack3dTab.Content()

	// All content in a Stack - we'll hide/show instead of swapping
	allViews := []fyne.CanvasObject{idealContent, irdropContent, sneakContent, driftContent, fecapContent, stack3dContent}

	// Button group for view selection
	idealBtn := widget.NewButton("Conductance", func() {})
	irDropBtn := widget.NewButton("IR Drop", func() {})
	sneakBtn := widget.NewButton("Sneak Paths", func() {})
	driftBtn := widget.NewButton("Input/Output", func() {})
	fecapBtn := widget.NewButton("FeCAP Mode", func() {})
	stack3dBtn := widget.NewButton("3D Stack", func() {})

	// Track current view to avoid redundant updates
	currentView := -1

	updateView := func(view int) {
		if view == currentView {
			return // No change needed
		}
		currentView = view

		// Hide all views first
		for _, v := range allViews {
			v.Hide()
		}

		// Show selected view and update buttons
		allViews[view].Show()

		// Update button importance (without calling Refresh - Fyne handles it)
		buttons := []*widget.Button{idealBtn, irDropBtn, sneakBtn, driftBtn, fecapBtn, stack3dBtn}
		statusTexts := []string{
			"Ideal MVM - No non-idealities",
			"IR Drop - Voltage drop along metal lines",
			"Sneak Paths - Parasitic current paths",
			"Drift - Conductance change over time",
			"FeCAP Mode - Charge-domain MVM (Q = C×V, no sneak paths or IR drop)",
			"3D Stack - Multi-layer isometric visualization (up to 512 layers)",
		}

		for i, btn := range buttons {
			if i == view {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.MediumImportance
			}
		}
		app.statusLabel.SetText(statusTexts[view])
	}

	idealBtn.OnTapped = func() { updateView(0) }
	irDropBtn.OnTapped = func() { updateView(1) }
	sneakBtn.OnTapped = func() { updateView(2) }
	driftBtn.OnTapped = func() { updateView(3) }
	fecapBtn.OnTapped = func() { updateView(4) }
	stack3dBtn.OnTapped = func() { updateView(5) }

	toolWidgets := sharedwidgets.NewToolValidationWidgets(sharedwidgets.ToolValidationOptions{
		Window:          app.window,
		ButtonLabel:     "Validate Tools",
		DialogTitle:     "Tool Validation",
		StatusLabelMode: sharedwidgets.ToolStatusSymbolWithName,
		MessageStyle:    sharedwidgets.ToolMessageUnicodeSkip,
		DialogFooter: "Install commands:\n" +
			"• CrossSim: git clone https://github.com/sandialabs/cross-sim && pip install -e ./cross-sim\n" +
			"• BadCrossbar: pip install badcrossbar",
	})

	title := canvas.NewText("FeCIM Crossbar Array Visualization", color.White)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}

	viewLabel := widget.NewLabel("What You're Seeing")

	buttonBar := container.NewHBox(
		viewLabel,
		idealBtn,
		irDropBtn,
		sneakBtn,
		driftBtn,
		fecapBtn,
		stack3dBtn,
		layout.NewSpacer(),
		toolWidgets.CrossSimStatus,
		toolWidgets.BadCrossbarStatus,
		toolWidgets.Button,
	)

	header := container.NewVBox(
		buttonBar,
		widget.NewSeparator(),
	)

	// Content container using Stack - all views layered, visibility toggled
	contentContainer := container.NewStack(allViews...)

	// Set initial view (shows first, hides others)
	updateView(0)

	return container.NewBorder(
		header,
		app.statusLabel,
		nil,
		nil,
		contentContainer,
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
