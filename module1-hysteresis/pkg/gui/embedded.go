// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// This file provides BuildContent for embedding in the unified visualizer.
package gui

import (
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// EmbeddedApp holds the state for an embedded demo instance
type EmbeddedApp struct {
	*App
	sharedwidgets.EmbeddedAppBase

	loopMu  sync.Mutex
	stopCh  chan struct{}
	simWG   sync.WaitGroup
	calibWG sync.WaitGroup
}

// SetPhysicsEngine switches the active physics engine for the embedded module.
// Note: this mutates simulation state and resets history (see App.setPhysicsEngine).
func (e *EmbeddedApp) SetPhysicsEngine(engine PhysicsEngine) {
	if e == nil || e.App == nil {
		return
	}
	e.App.setPhysicsEngine(engine)
}

// SetPhysicsEngineName is a convenience wrapper for automation.
// Accepts: "preisach"|"p" (default) and "lk"|"l-k"|"landau".
func (e *EmbeddedApp) SetPhysicsEngineName(name string) {
	if e == nil || e.App == nil {
		return
	}
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "lk", "l-k", "landau", "landau-k", "landau-khalatnikov":
		e.App.setPhysicsEngine(PhysicsLandau)
	default:
		e.App.setPhysicsEngine(PhysicsPreisach)
	}
}

// NewEmbeddedApp creates a new embedded GUI application (for use in unified visualizer)
func NewEmbeddedApp() *EmbeddedApp {
	materials := ferroelectric.AllMaterials()
	mat, matIndex := defaultMaterialSelection(materials)
	numLevels := physics.DefaultLevels
	preisach := ferroelectric.NewPreisachModel(mat)

	// Refactoring: Initialize managers
	calibManager := algo.NewCalibrationManager(numLevels)
	writeController := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, calibManager)

	app := &App{
		calibManager:       calibManager,
		writeController:    writeController,
		material:           mat,
		preisach:           preisach,
		materials:          materials,
		matIndex:           matIndex,
		numLevels:          numLevels,
		calibrationUp:      make([]float64, numLevels),
		calibrationDown:    make([]float64, numLevels),
		maxHistory:         50000,
		eHistory:           make([]float64, 0, 2000),
		pHistory:           make([]float64, 0, 2000),
		lastHistorySample:  -1,
		autoMode:           true,
		waveform:           WaveformSine,
		physicsEngine:      PhysicsPreisach,
		frequency:          0.5, // 0.5 Hz default
		timeScale:          1.0,
		wrdTargetLevel:     28, // Start high for dramatic first write
		wrdNextTargetLevel: 0,
		wrdSkipPrep:        true,
		wrdRangeFrac:       rangeFracForMaterial(mat),
		wrdGuardFrac:       0.15,
		maxLogLines:        12,
		logEntries:         make([]string, 0, 12),
		lastLogPhase:       -1,
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

	e.EmbeddedAppBase.Init(fyneApp, parentWindow)

	e.fyneApp = fyneApp
	e.mainWindow = parentWindow

	// Create UI components
	content := e.createUI()
	e.SetContent(content)

	return content
}

// Start begins the simulation loop (call after BuildContent)
func (e *EmbeddedApp) Start() {
	e.EmbeddedAppBase.Start()

	e.loopMu.Lock()
	if e.stopCh != nil {
		e.loopMu.Unlock()
		return
	}
	stopCh := make(chan struct{})
	e.stopCh = stopCh
	e.running.Store(true)
	e.loopMu.Unlock()

	e.startDataLogger()

	// Pre-warm equation SVG cache so the dialog opens instantly.
	go widgets.PrefetchEquationAssets()

	// Load or run calibration at startup (ensures calibration files exist)
	if os.Getenv("FECIM_DISABLE_STARTUP_CALIBRATION") != "1" {
		e.calibWG.Add(1)
		go func() {
			defer e.calibWG.Done()

			timer := time.NewTimer(100 * time.Millisecond) // Let UI settle
			defer timer.Stop()
			select {
			case <-stopCh:
				return
			case <-timer.C:
			}

			select {
			case <-stopCh:
				return
			default:
			}

			e.mu.Lock()
			defer e.mu.Unlock()

			// Stop() may be requested while we're waiting for the timer or acquiring
			// locks; bail out before doing expensive calibration or disk writes.
			select {
			case <-stopCh:
				return
			default:
			}
			if !e.running.Load() {
				return
			}

			if !e.loadCalibration() {
				// No valid saved calibration - run immediately
				log.Printf("Running calibration for %s at startup...", e.material.Name)
				e.calibrateLevelsAtTemperature(300)
				select {
				case <-stopCh:
					return
				default:
				}
				if !e.running.Load() {
					return
				}
				if err := e.saveCalibration(); err != nil {
					log.Printf("Warning: failed to save calibration: %v", err)
				}
			}
		}()
	}

	e.simWG.Add(1)
	go func() {
		defer e.simWG.Done()
		e.simulationLoop()
	}()
}

// Stop ends the simulation loop
func (e *EmbeddedApp) Stop() {
	var stopCh chan struct{}
	e.loopMu.Lock()
	stopCh = e.stopCh
	if stopCh == nil {
		e.loopMu.Unlock()
		e.EmbeddedAppBase.Stop()
		return
	}
	e.stopCh = nil
	e.running.Store(false)
	close(stopCh)
	e.loopMu.Unlock()

	e.simWG.Wait()
	e.calibWG.Wait()

	e.stopDataLogger()

	// Save calibration for next session (disabled in tests via env gate).
	e.mu.Lock()
	if err := e.saveCalibration(); err != nil {
		if log != nil {
			log.Printf("Warning: failed to save calibration: %v", err)
		}
	}
	e.mu.Unlock()

	e.EmbeddedAppBase.Stop()
}
