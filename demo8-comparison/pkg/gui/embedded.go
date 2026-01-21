// Package gui provides Fyne-based GUI components for architecture comparison.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"
)

// EmbeddedComparisonApp holds the state for an embedded demo instance
type EmbeddedComparisonApp struct {
	*ComparisonApp
}

// NewEmbeddedComparisonApp creates a new embedded comparison app (for use in unified visualizer)
func NewEmbeddedComparisonApp() *EmbeddedComparisonApp {
	ca := &ComparisonApp{
		currentWorkload:   "MNIST",
		currentInferences: 10000,
	}

	// Initialize energy specs with HONEST numbers and sources
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      1000, // ~1000 fJ/MAC
		Source:        "Intel/AMD published specs",
		Verified:      true,
		SourceDetails: "Includes memory access energy. Intel Xeon specs, AMD EPYC specs.",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      100, // ~100 fJ/MAC
		Source:        "NVIDIA H100 specifications",
		Verified:      true,
		SourceDetails: "H100 SXM: 700W TDP, ~3958 TFLOPS FP16. ~177 fJ/FLOP.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      10, // ~1-10 fJ/MAC (claimed)
		Source:        "Dr. Tour's presentation (NOT independently verified)",
		Verified:      false,
		SourceDetails: "Claimed: '10M× lower energy than NAND'. TRL 4 - lab only.",
	}

	return &EmbeddedComparisonApp{ComparisonApp: ca}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedComparisonApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	// Create UI components
	content := e.createMainLayout()

	// Initialize calculations
	e.updateCalculations()
	e.updateStatus("Ready. Select workload and adjust parameters.")

	return content
}

// Start begins any background processes
func (e *EmbeddedComparisonApp) Start() {
	// No continuous processes
}

// Stop ends any background processes
func (e *EmbeddedComparisonApp) Stop() {
	// No cleanup needed
}
