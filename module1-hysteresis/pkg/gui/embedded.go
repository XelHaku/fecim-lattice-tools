// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"time"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
)

// EmbeddedApp holds the state for an embedded demo instance
type EmbeddedApp struct {
	*App
}

// NewEmbeddedApp creates a new embedded GUI application (for use in unified visualizer)
func NewEmbeddedApp() *EmbeddedApp {
	materials := ferroelectric.AllMaterials()

	mat := materials[0]
	numLevels := 30         // Default: FeCIM's 30 discrete analog states
	// preisachGridSize := 200 // DEPRECATED
	preisach := ferroelectric.NewPreisachModel(mat)

	// Refactoring: Initialize managers
	calibManager := algo.NewCalibrationManager(numLevels)
	writeController := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, calibManager)

	app := &App{
		calibManager:    calibManager,
		writeController: writeController,
		material:        mat,
		preisach:        preisach,
		materials:       materials,
		matIndex:        0,
		numLevels:       numLevels,
		calibrationUp:   make([]float64, numLevels),
		calibrationDown: make([]float64, numLevels),
		maxHistory:      2000,
		eHistory:        make([]float64, 0, 2000),
		pHistory:        make([]float64, 0, 2000),
		autoMode:        true,
		waveform:        WaveformSine,
		frequency:       0.5, // 0.5 Hz default
		wrdTargetLevel:  28,  // Start high for dramatic first write
		maxLogLines:     12,
		logEntries:      make([]string, 0, 12),
		lastLogPhase:    -1,
		// isppCalc:        physics.NewISPPCalculator(preisach.GetEffectiveEc(), numLevels),
	}
	
	// Initialize L-K Solver and Adaptive ISPP
	app.lkSolver = physics.NewLKSolver()
	app.lkSolver.ConfigureFromMaterial(mat) // Load material-specific params (K_dep, etc.)
	app.adaptiveISPP = physics.NewAdaptiveISPP(app.lkSolver, mat)

	return &EmbeddedApp{App: app}
}

// BuildContent creates the UI content for embedding in a tab
// The fyne.App instance and window must be provided by the parent
func (e *EmbeddedApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	// Initialize logger here (after EnableFileLogging() has been called in main)
	if log == nil {
		log = logging.NewLogger("hysteresis")
	}

	e.fyneApp = fyneApp
	e.mainWindow = parentWindow

	// Create UI components
	content := e.createUI()

	return content
}

// Start begins the simulation loop (call after BuildContent)
func (e *EmbeddedApp) Start() {
	e.running = true

	// Load or run calibration at startup (ensures calibration files exist)
	go func() {
		time.Sleep(100 * time.Millisecond) // Let UI settle
		e.mu.Lock()
		if !e.loadCalibration() {
			// No valid saved calibration - run immediately
			log.Printf("Running calibration for %s at startup...", e.material.Name)
			e.calibrateLevelsAtTemperature(300)
			if err := e.saveCalibration(); err != nil {
				log.Printf("Warning: failed to save calibration: %v", err)
			}
		}
		e.mu.Unlock()
	}()

	go e.simulationLoop()
}

// Stop ends the simulation loop
func (e *EmbeddedApp) Stop() {
	e.running = false

	// Save calibration for next session
	e.mu.Lock()
	if err := e.saveCalibration(); err != nil {
		log.Printf("Warning: failed to save calibration: %v", err)
	}
	e.mu.Unlock()
}
