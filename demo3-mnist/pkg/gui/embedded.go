// Package gui provides Fyne-based GUI components for MNIST visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/crossbar"
	"multilayer-ferroelectric-cim-visualizer/demo3-mnist/pkg/training"
)

// EmbeddedMNISTApp holds the state for an embedded MNIST demo instance
type EmbeddedMNISTApp struct {
	*MNISTApp
}

// NewEmbeddedMNISTApp creates a new embedded MNIST GUI application
func NewEmbeddedMNISTApp() *EmbeddedMNISTApp {
	ma := &MNISTApp{}

	// Find data directory
	ma.dataDir = findDataDir()

	// Create crossbar arrays for layers
	// Layer 1: hidden x 784 (transposed for MVM)
	layer1Config := &crossbar.Config{
		Rows:       128, // hidden size
		Cols:       784, // input size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer1, _ := crossbar.NewArray(layer1Config)

	// Layer 2: 10 x hidden
	layer2Config := &crossbar.Config{
		Rows:       10,  // output size
		Cols:       128, // hidden size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer2, _ := crossbar.NewArray(layer2Config)

	// Create network
	ma.network = training.NewMNISTNetwork(layer1, layer2)

	// Try to load pretrained weights
	weightsPath := filepath.Join(ma.dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		_ = ma.network.LoadWeights(weightsPath)
	}

	return &EmbeddedMNISTApp{MNISTApp: ma}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance must be provided by the parent
func (e *EmbeddedMNISTApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create main layout
	content := e.createMainLayout()

	// Initialize
	e.updateStatus("Ready. Draw a digit or load test data.")

	return content
}

// Start initializes anything that needs to run after UI is visible
func (e *EmbeddedMNISTApp) Start() {
	// Nothing to start - MNIST demo is event-driven
}

// Stop cleans up any running processes
func (e *EmbeddedMNISTApp) Stop() {
	e.stopAutoDemoLoop()
}
