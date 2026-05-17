//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ExampleGlossaryIntegration demonstrates how to add glossary to module menus.
//
// Usage in any module's GUI:
//
//	// In your menu creation:
//	helpMenu := fyne.NewMenu("Help",
//	    widgets.CreateHelpMenuItems(window)...,
//	)
//
//	// Or as standalone buttons:
//	glossaryBtn := widgets.CreateGlossaryButton(window)
//	referencesBtn := widgets.CreateReferencesButton(window)
//
//	// Or popup specific term:
//	widgets.ShowGlossary("FeCIM", window)
//
//	// Or programmatic lookup:
//	definition := widgets.QuickTermLookup("Ec")
func ExampleGlossaryIntegration() {
	myApp := app.New()
	window := myApp.NewWindow("Glossary Example")

	// Method 1: Full glossary widget embedded
	glossaryWidget := NewGlossaryWidget()

	// Method 2: Help menu with glossary and references
	helpMenuItems := CreateHelpMenuItems(window)
	helpMenu := fyne.NewMenu("Help", helpMenuItems...)
	mainMenu := fyne.NewMainMenu(helpMenu)
	window.SetMainMenu(mainMenu)

	// Method 3: Toolbar buttons
	glossaryBtn := CreateGlossaryButton(window)
	referencesBtn := CreateReferencesButton(window)
	toolbar := container.NewHBox(glossaryBtn, referencesBtn)

	// Method 4: Popup specific term on button click
	termExampleBtn := widget.NewButton("What is FeCIM?", func() {
		ShowGlossary("FeCIM", window)
	})

	// Method 5: Programmatic access
	lookupBtn := widget.NewButton("Lookup Ec", func() {
		def := QuickTermLookup("Ec")
		widget.NewLabel("Definition: " + def)
	})

	// Layout example
	content := container.NewBorder(
		toolbar,
		container.NewHBox(termExampleBtn, lookupBtn),
		nil, nil,
		glossaryWidget,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}

// Example: Add to existing module (e.g., module1-hysteresis)
//
// In module1-hysteresis/pkg/gui/gui.go:
//
//	import "fecim-lattice-tools/shared/widgets"
//
//	func (g *GUI) BuildContent() fyne.CanvasObject {
//	    // ... existing code ...
//
//	    // Add help menu
//	    helpMenu := fyne.NewMenu("Help",
//	        widgets.CreateHelpMenuItems(g.window)...,
//	    )
//	    mainMenu := fyne.NewMainMenu(
//	        g.createFileMenu(),
//	        g.createViewMenu(),
//	        helpMenu,
//	    )
//	    g.window.SetMainMenu(mainMenu)
//
//	    // ... rest of existing code ...
//	}
//
// Or add glossary button to toolbar:
//
//	toolbar := container.NewHBox(
//	    // ... existing buttons ...
//	    widget.NewSeparator(),
//	    widgets.CreateGlossaryButton(g.window),
//	)
