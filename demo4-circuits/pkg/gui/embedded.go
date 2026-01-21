// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"
	"multilayer-ferroelectric-cim-visualizer/demo4-circuits/pkg/peripherals"
)

// EmbeddedCircuitsApp holds the state for an embedded demo instance
type EmbeddedCircuitsApp struct {
	*CircuitsApp
}

// NewEmbeddedCircuitsApp creates a new embedded circuits app (for use in unified visualizer)
func NewEmbeddedCircuitsApp() *EmbeddedCircuitsApp {
	ca := &CircuitsApp{
		currentLevel: 15,
	}

	// Initialize peripheral components
	ca.dac = peripherals.DefaultDAC()
	ca.adc = peripherals.DefaultADC()
	ca.tia = peripherals.DefaultTIA()
	ca.pump = peripherals.DefaultChargePump()

	return &EmbeddedCircuitsApp{CircuitsApp: ca}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedCircuitsApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create UI components
	content := e.createMainLayout()

	// Initialize displays
	e.updateValues()
	e.updateStatus("Ready. Select a circuit or run a cycle.")

	return content
}

// Start begins any background processes
func (e *EmbeddedCircuitsApp) Start() {
	// Start auto demo if it was set
	if e.demoModeSelect != nil && e.demoModeSelect.Selected == "Auto Demo" {
		e.startAutoDemoLoop()
	}
}

// Stop ends any background processes
func (e *EmbeddedCircuitsApp) Stop() {
	e.stopAutoDemoLoop()
}
