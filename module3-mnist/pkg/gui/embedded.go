//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for MNIST visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/module3-mnist/pkg/training"
	"fecim-lattice-tools/shared/canvas"
	"fecim-lattice-tools/shared/crossbar"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// EmbeddedMNISTApp holds the state for an embedded MNIST demo instance
type EmbeddedMNISTApp struct {
	*MNISTApp
	sharedwidgets.EmbeddedAppBase
	initErr        error
	startupWarning string
}

func (e *EmbeddedMNISTApp) bindHost(fyneApp fyne.App, parentWindow fyne.Window) {
	e.fyneApp = fyneApp
	e.window = parentWindow
}

// NewEmbeddedMNISTApp creates a new embedded MNIST GUI application
func NewEmbeddedMNISTApp() *EmbeddedMNISTApp {
	embedded := &EmbeddedMNISTApp{MNISTApp: &MNISTApp{}}
	if err := embedded.initialize(); err != nil {
		embedded.initErr = err
	}
	return embedded
}

func (e *EmbeddedMNISTApp) initialize() error {
	ma := e.MNISTApp

	// Find data directory
	ma.dataDir = utils.FindModuleDataDir("module3-mnist", "pretrained_weights.json")
	if ma.dataDir == "" {
		ma.dataDir = "module3-mnist/data" // Default fallback
	}

	// Create crossbar arrays for layers
	// Layer 1: hidden x 784 (transposed for MVM)
	layer1Config := &crossbar.Config{
		Rows:       128, // hidden size
		Cols:       784, // input size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer1, err := crossbar.NewArray(layer1Config)
	if err != nil {
		return fmt.Errorf("create MNIST hidden-layer crossbar: %w", err)
	}

	// Layer 2: 10 x hidden
	layer2Config := &crossbar.Config{
		Rows:       10,  // output size
		Cols:       128, // hidden size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer2, err := crossbar.NewArray(layer2Config)
	if err != nil {
		return fmt.Errorf("create MNIST output-layer crossbar: %w", err)
	}

	// Create network
	ma.network = training.NewMNISTNetwork(layer1, layer2)

	// Try to load pretrained weights
	weightsPath := filepath.Join(ma.dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		if err := ma.network.LoadWeights(weightsPath); err != nil {
			e.startupWarning = fmt.Sprintf("Pretrained weights unavailable: %v", err)
		}
	}

	return nil
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance must be provided by the parent
func (e *EmbeddedMNISTApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	return e.EmbeddedAppBase.BuildOrReuseContentWithHostSync(fyneApp, parentWindow, e.bindHost, func() fyne.CanvasObject {
		if e.initErr != nil {
			return sharedwidgets.NewModuleErrorContent("MNIST", e.initErr)
		}

		content := e.createMainLayout()

		status := "Ready. Draw a digit or load test data."
		if e.startupWarning != "" {
			status = e.startupWarning
		}
		e.updateStatus(status)

		return content
	})
}

// Start initializes anything that needs to run after UI is visible
func (e *EmbeddedMNISTApp) Start() {
	if e.initErr != nil {
		return
	}
	e.EmbeddedAppBase.Start()
	// Nothing to start - MNIST demo is event-driven
}

// Stop cleans up any running processes
func (e *EmbeddedMNISTApp) Stop() {
	if e.initErr != nil {
		return
	}
	e.stopAutoDemoLoop()
	e.EmbeddedAppBase.Stop()
}

// EmbeddedDualModeApp holds the state for the dual-mode (FP vs CIM) MNIST demo.
type EmbeddedDualModeApp struct {
	*DualModeApp
	sharedwidgets.EmbeddedAppBase
}

// NewEmbeddedDualModeApp creates a new embedded dual-mode MNIST GUI application.
func NewEmbeddedDualModeApp() *EmbeddedDualModeApp {
	return &EmbeddedDualModeApp{DualModeApp: NewDualModeApp()}
}

// BuildContent creates the UI content for embedding in a tab.
// The fyne.App instance must be provided by the parent.
func (e *EmbeddedDualModeApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	return e.EmbeddedAppBase.BuildOrReuseContent(fyneApp, parentWindow, func() fyne.CanvasObject {
		return e.DualModeApp.BuildContent(fyneApp, parentWindow)
	})
}

// Start initializes anything that needs to run after UI is visible.
func (e *EmbeddedDualModeApp) Start() {
	e.EmbeddedAppBase.Start()
	e.DualModeApp.Start()
}

// Stop cleans up any running processes.
func (e *EmbeddedDualModeApp) Stop() {
	e.DualModeApp.Stop()
	e.EmbeddedAppBase.Stop()
}
