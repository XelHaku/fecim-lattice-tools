// Package gui provides Fyne-based GUI for non-idealities visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"
	"multilayer-ferroelectric-cim-visualizer/demo7-nonidealities/pkg/nonidealities"
)

// EmbeddedNonIdealitiesApp holds the state for an embedded demo instance
type EmbeddedNonIdealitiesApp struct {
	*NonIdealitiesApp
}

// NewEmbeddedNonIdealitiesApp creates a new embedded non-idealities app (for use in unified visualizer)
func NewEmbeddedNonIdealitiesApp() *EmbeddedNonIdealitiesApp {
	na := &NonIdealitiesApp{
		arraySize: 16,
	}
	return &EmbeddedNonIdealitiesApp{NonIdealitiesApp: na}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedNonIdealitiesApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Initialize simulators
	e.initSimulators()

	// Create UI components
	content := e.createMainLayout()

	// Initial update
	e.updateIRDrop()
	e.updateSneakPaths()
	e.updateDrift()

	return content
}

// initSimulators initializes the simulation engines
func (e *EmbeddedNonIdealitiesApp) initSimulators() {
	// IR Drop simulator
	e.irSim = nonidealities.NewIRDropSimulator(e.arraySize, e.arraySize)
	for i := 0; i < e.arraySize; i++ {
		e.irSim.SetInputVoltage(i, 0.3+0.2*float64(i%5)/4.0)
	}
	for i := 0; i < e.arraySize; i++ {
		for j := 0; j < e.arraySize; j++ {
			distFromCenter := float64((i-e.arraySize/2)*(i-e.arraySize/2) + (j-e.arraySize/2)*(j-e.arraySize/2))
			g := 50e-6 + 30e-6*distFromCenter/float64(e.arraySize*e.arraySize/2)
			e.irSim.SetConductance(i, j, g)
		}
	}
	e.irSim.Simulate(100)

	// Sneak path analyzer
	e.sneakSim = nonidealities.NewSneakPathAnalyzer(e.arraySize, e.arraySize)
	for i := 0; i < e.arraySize; i++ {
		for j := 0; j < e.arraySize; j++ {
			g := (10 + float64((i*7+j*11)%80)) * 1e-6
			e.sneakSim.SetConductance(i, j, g)
		}
	}
	e.sneakSim.AnalyzeTarget(e.arraySize/2, e.arraySize/2, 0.5)

	// Drift simulator
	e.driftSim = nonidealities.NewDriftSimulator(e.arraySize, e.arraySize, 30)
	for i := 0; i < e.arraySize; i++ {
		for j := 0; j < e.arraySize; j++ {
			level := (i*3 + j*5) % 30
			e.driftSim.SetConductanceLevel(i, j, level)
		}
	}
	// Simulate some time
	for step := 0; step < 50; step++ {
		e.driftSim.SimulateTimeStep(200)
		e.driftSim.RecordSnapshot()
	}
}

// Start begins any background processes
func (e *EmbeddedNonIdealitiesApp) Start() {
	// No continuous simulation needed
}

// Stop ends any background processes
func (e *EmbeddedNonIdealitiesApp) Stop() {
	// No cleanup needed
}
