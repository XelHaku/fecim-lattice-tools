// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/peripherals"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// EmbeddedCircuitsApp holds the state for an embedded demo instance
type EmbeddedCircuitsApp struct {
	*CircuitsApp
	sharedwidgets.EmbeddedAppBase
}

// NewEmbeddedCircuitsApp creates a new embedded circuits app (for use in unified visualizer)
func NewEmbeddedCircuitsApp() *EmbeddedCircuitsApp {
	ca := &CircuitsApp{
		arrayRows:       DefaultSize,
		arrayCols:       DefaultSize,
		quantLevels:     FeCIMLevels,
		dacBits:         DefaultDACBits,
		adcBits:         DefaultADCBits,
		vMin:            2.0,
		vMax:            5.0,
		pulseWidth:      50.0,
		readVoltage:     0.5,
		tiaGain:         10.0,
		selectedRow:     3,
		selectedCol:     5,
		targetLevel:     15,
		architecture:    sharedwidgets.Architecture0T1R, // Default to passive for educational demo
		readOverlayMode: "Off",
	}

	// Initialize peripheral components
	ca.dac = peripherals.DefaultDAC()
	ca.adc = peripherals.DefaultADC()
	ca.tia = peripherals.DefaultTIA()
	ca.pump = peripherals.DefaultChargePump()

	// Initialize array - all cells start at mid-level (neutral state)
	midLevel := ca.quantLevels / 2
	ca.arrayWeights = make([][]int, ca.arrayRows)
	ca.halfSelectResidue = make([][]float64, ca.arrayRows)
	for i := range ca.arrayWeights {
		ca.arrayWeights[i] = make([]int, ca.arrayCols)
		ca.halfSelectResidue[i] = make([]float64, ca.arrayCols)
		for j := range ca.arrayWeights[i] {
			ca.arrayWeights[i][j] = midLevel
		}
	}

	ca.inputVector = make([]int, ca.arrayCols)
	ca.outputVector = make([]float64, ca.arrayRows)
	// Input vector starts at 0

	return &EmbeddedCircuitsApp{CircuitsApp: ca}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedCircuitsApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.EmbeddedAppBase.Init(fyneApp, parentWindow)
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create main tabbed layout (same as standalone)
	content := e.createMainLayout()
	e.SetContent(content)

	return content
}

// Start begins any background processes when the tab is selected
func (e *EmbeddedCircuitsApp) Start() {
	e.EmbeddedAppBase.Start()
	// Reset stop state so goroutines can run again
	e.mu.Lock()
	if e.stopped {
		e.stopped = false
		e.stopChan = make(chan struct{})
	}
	e.mu.Unlock()

	// Update action button states for current mode
	e.updateActionButtons()

	// Refresh all canvases when tab is selected
	e.refreshWriteArray()
	e.refreshWritePulse()
	e.refreshReadZone()
	e.refreshTimingDiagrams()
	e.refreshUnifiedArray()
	sharedwidgets.SafeDo(func() {
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
	// Signal all goroutines to stop
	e.mu.Lock()
	if !e.stopped && e.stopChan != nil {
		e.stopped = true
		close(e.stopChan)
	}
	e.animationActive = false
	e.animationStep = 0
	e.mu.Unlock()

	// Cancel any ongoing device state operations
	if e.deviceState != nil {
		e.deviceState.CancelWriteSequence()
		e.deviceState.CancelISPP()
	}
	e.EmbeddedAppBase.Stop()
}
