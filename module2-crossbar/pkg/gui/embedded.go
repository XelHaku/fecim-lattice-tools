// Package gui provides Fyne-based GUI components for crossbar visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fmt"
	"math/rand"

	"fyne.io/fyne/v2"
	"multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
)

// EmbeddedCrossbarApp holds the state for an embedded crossbar demo instance
type EmbeddedCrossbarApp struct {
	*CrossbarApp
}

// NewEmbeddedCrossbarApp creates a new embedded crossbar GUI application
func NewEmbeddedCrossbarApp() *EmbeddedCrossbarApp {
	ca := &CrossbarApp{
		selectedRow: -1, // No selection initially
		selectedCol: -1,
	}

	// Initialize with default config
	ca.config = &crossbar.Config{
		Rows:       64,
		Cols:       64,
		NoiseLevel: 0.02,
		ADCBits:    6,
		DACBits:    8,
	}

	// Create crossbar array
	var err error
	ca.array, err = crossbar.NewArray(ca.config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create crossbar array: %v", err))
	}

	// Program initial random weights
	for i := 0; i < ca.config.Rows; i++ {
		for j := 0; j < ca.config.Cols; j++ {
			level := rand.Intn(30)
			weight := float64(level) / 29.0
			ca.array.ProgramWeight(i, j, weight)
		}
	}

	return &EmbeddedCrossbarApp{CrossbarApp: ca}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance must be provided by the parent
func (e *EmbeddedCrossbarApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create enhanced layout (embedded version always uses enhanced features)
	content := e.createEnhancedMainLayout()

	// Initialize displays
	e.updateConductanceDisplay()
	e.updateStatus("Ready. Program weights and run MVM operations.")

	return content
}

// BuildContentStandard creates standard UI content for embedding (no enhanced features)
func (e *EmbeddedCrossbarApp) BuildContentStandard(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create standard layout
	content := e.createMainLayout()

	// Initialize displays
	e.updateConductanceDisplay()
	e.updateStatus("Ready. Program weights and run MVM operations.")

	return content
}

// Start initializes anything that needs to run after UI is visible
func (e *EmbeddedCrossbarApp) Start() {
	// Nothing to start for crossbar demo - it's event-driven
}

// Stop cleans up any running processes
func (e *EmbeddedCrossbarApp) Stop() {
	e.stopAutoDemoLoop()
}
