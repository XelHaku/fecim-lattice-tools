// Package gui provides Fyne-based GUI for thermal analysis visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"
	"multilayer-ferroelectric-cim-visualizer/demo5-thermal/pkg/thermal"
)

// EmbeddedThermalApp holds the state for an embedded demo instance
type EmbeddedThermalApp struct {
	*ThermalApp
}

// NewEmbeddedThermalApp creates a new embedded thermal app (for use in unified visualizer)
func NewEmbeddedThermalApp() *EmbeddedThermalApp {
	ta := &ThermalApp{
		selectedLayer:  1, // Middle layer (crossbar)
		powerLevel:     1.0,
		simulationMode: "FeCIM (Low Power)",
		logEntries:     make([]string, 0, 20),
	}
	return &EmbeddedThermalApp{ThermalApp: ta}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedThermalApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Initialize simulation
	e.multiSim = thermal.DefaultMultiLayerSim()
	e.singleSim = e.multiSim.Layers[e.selectedLayer]
	e.setPowerDistribution("FeCIM (Low Power)")

	// Create UI components
	content := e.createMainLayout()

	// Initial updates
	e.updateHeatmap()
	e.updateStats()

	return content
}

// Start begins the simulation loop
func (e *EmbeddedThermalApp) Start() {
	// Optionally start simulation automatically
}

// Stop ends the simulation loop
func (e *EmbeddedThermalApp) Stop() {
	e.stopSimulation()
}
