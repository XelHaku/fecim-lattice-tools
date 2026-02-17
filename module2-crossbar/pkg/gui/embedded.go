// Package gui provides Fyne-based GUI components for crossbar visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fmt"
	"math/rand"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
	"fyne.io/fyne/v2"
)

// EmbeddedCrossbarApp holds the state for an embedded crossbar demo instance
type EmbeddedCrossbarApp struct {
	*CrossbarApp
	sharedwidgets.EmbeddedAppBase
	initialized bool
}

// NewEmbeddedCrossbarApp creates a new embedded crossbar GUI application.
// The crossbar array is lazily initialized when the module is first opened.
func NewEmbeddedCrossbarApp() (*EmbeddedCrossbarApp, error) {
	ca := &CrossbarApp{
		selectedRow:   -1, // No selection initially
		selectedCol:   -1,
		tabHasNewData: make(map[string]bool),
		temperatureK:  300.0,
	}

	// Initialize with default config (array created lazily)
	ca.config = &crossbar.Config{
		Rows:       64,
		Cols:       64,
		NoiseLevel: 0.02,
		ADCBits:    6,
		DACBits:    8,
	}

	return &EmbeddedCrossbarApp{CrossbarApp: ca}, nil
}

// initArray lazily initializes the crossbar array on first use
func (e *EmbeddedCrossbarApp) initArray() error {
	if e.initialized {
		return nil
	}

	var err error
	e.array, err = crossbar.NewArray(e.config)
	if err != nil {
		return fmt.Errorf("failed to create crossbar array: %w", err)
	}

	// Program initial random weights
	for i := 0; i < e.config.Rows; i++ {
		for j := 0; j < e.config.Cols; j++ {
			level := rand.Intn(30)
			weight := float64(level) / 29.0
			e.array.ProgramWeight(i, j, weight)
		}
	}

	e.initialized = true
	return nil
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance must be provided by the parent
func (e *EmbeddedCrossbarApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.EmbeddedAppBase.Init(fyneApp, parentWindow)
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Lazily initialize the crossbar array on first BuildContent call
	if err := e.initArray(); err != nil {
		// Return error placeholder if init fails
		return fyne.NewContainerWithLayout(nil)
	}

	// Create enhanced layout (embedded version always uses enhanced features)
	content := e.createEnhancedMainLayout()
	e.SetContent(content)

	// Initialize displays
	e.updateConductanceDisplay()
	e.updateStatus("Ready. Program weights and run MVM operations.")

	// Set first-load onboarding content (same as standalone mode)
	e.setEducationalContent("Quick Guide",
		"Welcome to Crossbar!\n\n"+
			"Quick Start:\n"+
			"1. Hover over cells\n"+
			"2. Click for details\n"+
			"3. Use controls (right)\n"+
			"4. Explore tabs\n\n"+
			"Key Concepts:\n"+
			"• 30 levels/cell\n"+
			"• MVM = W × V\n"+
			"• Parallel compute")

	return content
}

// BuildContentStandard creates standard UI content for embedding (no enhanced features)
func (e *EmbeddedCrossbarApp) BuildContentStandard(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Lazily initialize the crossbar array on first BuildContent call
	if err := e.initArray(); err != nil {
		return fyne.NewContainerWithLayout(nil)
	}

	// Create standard layout
	content := e.createMainLayout()

	// Initialize displays
	e.updateConductanceDisplay()
	e.updateStatus("Ready. Program weights and run MVM operations.")

	return content
}

// Start initializes anything that needs to run after UI is visible
func (e *EmbeddedCrossbarApp) Start() {
	e.EmbeddedAppBase.Start()
	// Auto-run MVM once on first visit to populate all analysis data
	if !e.hasRunInitialMVM {
		e.hasRunInitialMVM = true
		// Run enhanced MVM to populate IR drop, sneak path, and other analysis
		// Use instant (non-animated) version for initial load
		e.runEnhancedMVMInstant()
	}
}

// Stop cleans up any running processes
func (e *EmbeddedCrossbarApp) Stop() {
	e.stopAutoDemoLoop()
	e.EmbeddedAppBase.Stop()
}
