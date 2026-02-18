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
}

// NewEmbeddedMNISTApp creates a new embedded MNIST GUI application
func NewEmbeddedMNISTApp() *EmbeddedMNISTApp {
	ma := &MNISTApp{}

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
		fmt.Printf("Error: failed to create layer 1 crossbar: %v\n", err)
		return nil
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
		fmt.Printf("Error: failed to create layer 2 crossbar: %v\n", err)
		return nil
	}

	// Create network
	ma.network = training.NewMNISTNetwork(layer1, layer2)

	// Try to load pretrained weights
	weightsPath := filepath.Join(ma.dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		if err := ma.network.LoadWeights(weightsPath); err != nil {
			fmt.Printf("Warning: failed to load pretrained weights from %s: %v\n", weightsPath, err)
		}
	}

	return &EmbeddedMNISTApp{MNISTApp: ma}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance must be provided by the parent
func (e *EmbeddedMNISTApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.EmbeddedAppBase.Init(fyneApp, parentWindow)
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create main layout
	content := e.createMainLayout()
	e.SetContent(content)

	// Initialize
	e.updateStatus("Ready. Draw a digit or load test data.")

	return content
}

// Start initializes anything that needs to run after UI is visible
func (e *EmbeddedMNISTApp) Start() {
	e.EmbeddedAppBase.Start()
	// Nothing to start - MNIST demo is event-driven
}

// Stop cleans up any running processes
func (e *EmbeddedMNISTApp) Stop() {
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
	e.EmbeddedAppBase.Init(fyneApp, parentWindow)
	content := e.DualModeApp.BuildContent(fyneApp, parentWindow)
	e.SetContent(content)
	return content
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
