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

	// Initialize energy specs using constants from app.go
	// Source: docs/videos/ironlattice-youtube-script.md line 205
	// "CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. FeCIM: under 1 picojoule."
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      cpuEnergyPJPerMAC * 1000, // 1,000,000 fJ/MAC
		Source:        "Intel/AMD published specs",
		Verified:      true,
		SourceDetails: "Includes memory access energy (~640 pJ for DRAM fetch + ~3-5 pJ for MAC).",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      gpuEnergyPJPerMAC * 1000, // 100,000 fJ/MAC
		Source:        "NVIDIA H100 specifications",
		Verified:      true,
		SourceDetails: "H100 SXM: 700W TDP, ~3958 TFLOPS FP16. HBM access dominates.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      fecimEnergyPJPerMAC * 1000, // 1,000 fJ/MAC
		Source:        "Dr. Tour's presentation (NOT independently verified)",
		Verified:      false,
		SourceDetails: "Claimed: 'under 1 picojoule'. TRL 4 - lab only.",
	}

	return &EmbeddedComparisonApp{ComparisonApp: ca}
}

// BuildContent creates the UI content for embedding in a tab
func (e *EmbeddedComparisonApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.window = parentWindow

	content := e.createMainLayout()

	// Note: updateCalculations() is called via onWorkloadChanged when SetSelected triggers
	// No need for explicit call here - it causes duplicate calculation
	e.updateStatus("Ready. Select workload and adjust parameters.")

	return content
}

// Start begins any background processes (called when tab is selected)
func (e *EmbeddedComparisonApp) Start() {
	e.animMu.Lock()
	if e.running {
		e.animMu.Unlock()
		return
	}

	e.running = true
	e.paused = false
	e.animMu.Unlock()

	go e.animationLoop()

	debug.Println("EmbeddedComparisonApp: Animation started")
}

// Stop ends any background processes (called when tab is deselected)
func (e *EmbeddedComparisonApp) Stop() {
	e.animMu.Lock()
	e.running = false
	e.animMu.Unlock()

	debug.Println("EmbeddedComparisonApp: Animation stopped")
}
