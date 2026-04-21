// Package gui provides Fyne-based GUI components for architecture comparison.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// EmbeddedComparisonApp holds the state for an embedded demo instance
type EmbeddedComparisonApp struct {
	*ComparisonApp
	sharedwidgets.EmbeddedAppBase
}

func (e *EmbeddedComparisonApp) bindHost(fyneApp fyne.App, parentWindow fyne.Window) {
	e.fyneApp = fyneApp
	e.window = parentWindow
}

// NewEmbeddedComparisonApp creates a new embedded comparison app (for use in unified visualizer)
func NewEmbeddedComparisonApp() *EmbeddedComparisonApp {
	ca := &ComparisonApp{
		currentWorkload:   "GPT-2",
		currentInferences: 10000,
	}

	// Initialize energy specs using constants from app.go (model inputs).
	// See docs/4-research/honesty-audit.md for the current claim boundary.
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      cpuEnergyPJPerMAC * 1000, // 1,000,000 fJ/MAC
		Source:        "Model input (public CPU/DRAM datasheets)",
		Verified:      false,
		SourceDetails: "Model input: includes memory access energy (~640 pJ for DRAM fetch + ~3-5 pJ for MAC).",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      gpuEnergyPJPerMAC * 1000, // 100,000 fJ/MAC
		Source:        "Model input (public GPU/HBM datasheets)",
		Verified:      false,
		SourceDetails: "Model input: H100 SXM 700W TDP, ~3958 TFLOPS FP16; HBM access dominates.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      fecimEnergyPJPerMAC * 1000, // 1,000 fJ/MAC
		Source:        "Model input (assumed simulation baseline)",
		Verified:      false,
		SourceDetails: "Model input: 1.0 picojoule per MAC (assumed simulation baseline).",
	}

	return &EmbeddedComparisonApp{ComparisonApp: ca}
}

// RegisterKeyboard re-registers the comparison module's keyboard handler on the
// shared canvas. Called by the unified app when this tab becomes active.
func (e *EmbeddedComparisonApp) RegisterKeyboard() {
	if e.ComparisonApp != nil && e.window != nil {
		e.setupKeyboard()
	}
}

// BuildContent creates the UI content for embedding in a tab
func (e *EmbeddedComparisonApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	return e.EmbeddedAppBase.BuildOrReuseContentWithHostSync(fyneApp, parentWindow, e.bindHost, func() fyne.CanvasObject {
		content := e.createMainLayout()

		// Note: updateCalculations() is called via onWorkloadChanged when SetSelected triggers
		// No need for explicit call here - it causes duplicate calculation
		e.updateStatus("Ready. Select workload and adjust parameters.")

		// Add a Sensitivity tab alongside the main comparison view
		sensitivityPanel := NewSensitivityPanel(e.lastRun, "tco_reduction_pct")
		tabs := container.NewAppTabs(
			container.NewTabItem("Comparison", content),
			container.NewTabItem("Sensitivity", sensitivityPanel),
		)
		return tabs
	})
}

// Start begins any background processes (called when tab is selected)
func (e *EmbeddedComparisonApp) Start() {
	e.EmbeddedAppBase.Start()
	e.animMu.Lock()
	if e.running {
		e.animMu.Unlock()
		return
	}

	e.running = true
	e.paused = false
	e.animWG.Add(1)
	e.animMu.Unlock()

	// Auto-calculate on start to show real savings (not $0M)
	e.updateCalculations()

	go func() {
		defer e.animWG.Done()
		e.animationLoop()
	}()

	debug.Println("EmbeddedComparisonApp: Animation started")
}

// Stop ends any background processes (called when tab is deselected)
func (e *EmbeddedComparisonApp) Stop() {
	e.animMu.Lock()
	e.running = false
	e.animMu.Unlock()

	e.animWG.Wait()
	debug.Println("EmbeddedComparisonApp: Animation stopped")
	e.EmbeddedAppBase.Stop()
}
