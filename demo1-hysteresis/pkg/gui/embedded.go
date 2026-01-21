// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"
	"multilayer-ferroelectric-cim-visualizer/demo1-hysteresis/pkg/ferroelectric"
)

// EmbeddedApp holds the state for an embedded demo instance
type EmbeddedApp struct {
	*App
}

// NewEmbeddedApp creates a new embedded GUI application (for use in unified visualizer)
func NewEmbeddedApp() *EmbeddedApp {
	materials := []*ferroelectric.HZOMaterial{
		ferroelectric.DefaultHZO(),
		ferroelectric.OptimizedHZO(),
		ferroelectric.FeCIMMaterial(),
	}

	mat := materials[0]
	preisach := ferroelectric.NewMayergoyzPreisach(mat, 30)

	app := &App{
		material:       mat,
		preisach:       preisach,
		materials:      materials,
		matIndex:       0,
		maxHistory:     500,
		eHistory:       make([]float64, 0, 500),
		pHistory:       make([]float64, 0, 500),
		autoMode:       true,
		waveform:       WaveformSine,
		frequency:      0.5, // 0.5 Hz default
		rwTargetLevel:  25,
		rwStepDelay:    2.5, // 2.5 seconds between random level changes
		wrdTargetLevel: 28,  // Start high for dramatic first write
		maxLogLines:    12,
		logEntries:     make([]string, 0, 12),
		lastLogPhase:   -1,
	}

	return &EmbeddedApp{App: app}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.mainWindow = parentWindow

	// Create UI components
	content := e.createUI()

	return content
}

// Start begins the simulation loop (call after BuildContent)
func (e *EmbeddedApp) Start() {
	e.running = true
	go e.simulationLoop()
}

// Stop ends the simulation loop
func (e *EmbeddedApp) Stop() {
	e.running = false
}
