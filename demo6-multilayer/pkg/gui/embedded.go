// Package gui provides Fyne-based GUI for the 3D multilayer stack visualization.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"
	"multilayer-ferroelectric-cim-visualizer/demo6-multilayer/pkg/multilayer"
)

// EmbeddedMultilayerApp holds the state for an embedded demo instance
type EmbeddedMultilayerApp struct {
	*MultilayerApp
}

// NewEmbeddedMultilayerApp creates a new embedded multilayer app (for use in unified visualizer)
func NewEmbeddedMultilayerApp() *EmbeddedMultilayerApp {
	ma := &MultilayerApp{}
	// Initialize with demo stack
	ma.stack = multilayer.SmallStack()
	return &EmbeddedMultilayerApp{MultilayerApp: ma}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedMultilayerApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create UI components
	content := e.createMainLayout()

	// Initialize views
	e.updateStackView()
	e.updateMetrics()

	return content
}

// Start begins any background processes
func (e *EmbeddedMultilayerApp) Start() {
	// No continuous simulation needed for multilayer view
}

// Stop ends any background processes
func (e *EmbeddedMultilayerApp) Stop() {
	// No cleanup needed
}
