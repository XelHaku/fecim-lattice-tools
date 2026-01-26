// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"math/rand"

	"fyne.io/fyne/v2"

	"multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/peripherals"
)

// EmbeddedCircuitsApp holds the state for an embedded demo instance
type EmbeddedCircuitsApp struct {
	*CircuitsApp
}

// NewEmbeddedCircuitsApp creates a new embedded circuits app (for use in unified visualizer)
func NewEmbeddedCircuitsApp() *EmbeddedCircuitsApp {
	ca := &CircuitsApp{
		arrayRows:   DefaultSize,
		arrayCols:   DefaultSize,
		quantLevels: FeCIMLevels,
		dacBits:     DefaultDACBits,
		adcBits:     DefaultADCBits,
		vMin:        2.0,
		vMax:        5.0,
		pulseWidth:  50.0,
		readVoltage: 0.5,
		tiaGain:     10.0,
		selectedRow: 3,
		selectedCol: 5,
		targetLevel: 15,
	}

	// Initialize peripheral components
	ca.dac = peripherals.DefaultDAC()
	ca.adc = peripherals.DefaultADC()
	ca.tia = peripherals.DefaultTIA()
	ca.pump = peripherals.DefaultChargePump()

	// Initialize array
	ca.arrayWeights = make([][]int, ca.arrayRows)
	for i := range ca.arrayWeights {
		ca.arrayWeights[i] = make([]int, ca.arrayCols)
		for j := range ca.arrayWeights[i] {
			ca.arrayWeights[i][j] = rand.Intn(ca.quantLevels)
		}
	}

	ca.inputVector = make([]int, ca.arrayCols)
	ca.outputVector = make([]float64, ca.arrayRows)
	for j := range ca.inputVector {
		ca.inputVector[j] = rand.Intn(256)
	}

	return &EmbeddedCircuitsApp{CircuitsApp: ca}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedCircuitsApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create main tabbed layout (same as standalone)
	content := e.createMainLayout()

	return content
}

// Start begins any background processes when the tab is selected
func (e *EmbeddedCircuitsApp) Start() {
	// Refresh all canvases when tab is selected
	e.refreshWriteArray()
	e.refreshWritePulse()
	e.refreshReadZone()
	e.refreshTimingDiagrams()
	e.refreshSharedArray()
	fyne.Do(func() {
		if e.computeArrayCanvas != nil {
			e.computeArrayCanvas.Refresh()
		}
		if e.compArchCanvas != nil {
			e.compArchCanvas.Refresh()
		}
		if e.compTimingCanvas != nil {
			e.compTimingCanvas.Refresh()
		}
		if e.compEnergyCanvas != nil {
			e.compEnergyCanvas.Refresh()
		}
	})
}

// Stop ends any background processes when the tab is deselected
func (e *EmbeddedCircuitsApp) Stop() {
	// Nothing to stop in current implementation
}
