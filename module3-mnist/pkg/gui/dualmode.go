// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode.go provides the dual-mode (FP vs CIM) comparison GUI.
package gui

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/shared/logging"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Constants for MNIST network configuration
const (
	// Network architecture
	MNISTInputSize  = 784    // 28x28 pixel images
	MNISTHiddenSize = 128    // Default hidden layer size
	MNISTOutputSize = 10     // Digits 0-9
	MNISTTotalMACs  = 101632 // Total MACs per inference: (784×128) + (128×10) = 101632 (reference value)

	// FeCIM hardware parameters
	FeCIMDefaultLevels = 30   // Default levels (can be changed in UI)
	FeCIMDefaultNoise  = 0.01 // 1% standard deviation (typical production)
	FeCIMDefaultADC    = 8    // 8-bit ADC resolution
	FeCIMDefaultDAC    = 8    // 8-bit DAC resolution

	// Energy efficiency (Samsung Nature 2025: 25-100× vs NAND)
	FeCIMEnergyPerMAC = 50e-15 // 50 fJ/MAC (femtojoules)
	GPUEnergyPerMAC   = 5e-12  // 5 pJ/MAC estimate
	EnergyRatioGPU    = 100    // Verified: 25-100× (Samsung Nature 2025)

	// Accuracy reference (peer-reviewed baselines, not targets)
	// Note: Accuracy varies with noise, levels, and architecture
	// Peer-reviewed: 96.6% (Nature Commun. 2023), 98.24% (ScienceDirect 2025)
	TargetFP32Accuracy = 0.98 // 98% theoretical with Float32
)

// Package-level logger for MNIST GUI (named mnistLog to avoid conflict with app.go's debug logger)
var mnistLog *logging.Logger

func init() {
	mnistLog = logging.NewLogger("mnist")
}

// DualModeApp provides dual-path inference visualization (FP vs CIM).
type DualModeApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Network controller (ARCH-001: extracted from DualModeApp)
	networkCtrl *NetworkController

	// Drawing
	digitCanvas *DigitCanvas

	// Control panel components
	levelsSelect *widget.Select
	levelsLabel  *widget.Label
	noiseSlider  *widget.Slider
	noiseLabel   *widget.Label

	// Result panel components
	fpPredLabel    *widget.Label
	fpConfBar      *widget.ProgressBar
	cimPredLabel   *widget.Label
	cimConfBar     *widget.ProgressBar
	agreementLabel *widget.Label
	energyLabel    *widget.Label
	fpProbBars     [10]*widget.ProgressBar
	cimProbBars    [10]*widget.ProgressBar

	// Weight visualization
	weightHeatmap          *canvas.Raster
	hoverableWeightHeatmap *HoverableHeatmap
	weightLayer            int // 0 = layer 1, 1 = layer 2
	weightDimLabel         *widget.Label
	weightRangeLabel       *widget.Label
	weightLevelsLabel      *widget.Label

	// Quick test
	testButton      *widget.Button
	testResultLabel *widget.Label
	testProgressBar *widget.ProgressBar

	// Status
	statusLabel *widget.Label
	initialized bool // true after UI is fully built

	// Last inference result for refresh
	lastPixels []float64

	// P1 Enhancement Widgets
	quantizationWidget   *QuantizationWidget   // P1.1: Shows weight quantization
	energyWidget         *EnergyWidget         // P1.3: Energy tracking
	comparisonCard       *ComparisonCard       // P1.2: Enhanced FP vs CIM comparison
	dualProbabilityChart *DualProbabilityChart // P1.2: Probability divergence chart

	// P2 Enhancements: Animation & Quick Demo
	inferencePhaseLabel *widget.Label // Shows current inference phase
	quickDemoRunning    bool          // True when quick demo is active
	quickDemoStopChan   chan struct{} // Channel to stop quick demo
	animationEnabled    bool          // Enable/disable inference animation

	// P2 Enhancement: Weight Comparison
	weightComparisonWidget *WeightComparisonWidget
	dualWeightHeatmap      *DualWeightHeatmap

	// Weight zone collapse state
	weightsCollapsed bool
	weightsToggleBtn *widget.Button
	weightsContent   *fyne.Container

	// Layout containers (stored to defer SetOffset)
	leftSplit  *container.Split
	rightSplit *container.Split
	mainSplit  *container.Split

	// Adaptive layout for responsive design
	adaptiveLayout *sharedwidgets.AdaptiveLayout
}

// NewDualModeApp creates a new dual-mode MNIST application.
func NewDualModeApp() *DualModeApp {
	app := &DualModeApp{
		// Create network controller (ARCH-001: network management extracted)
		networkCtrl: NewNetworkController(MNISTInputSize, MNISTHiddenSize, MNISTOutputSize),
	}

	return app
}

// ======================================================================
// Network Controller Accessors (ARCH-001: backward compatibility layer)
// These methods delegate to networkCtrl for existing code compatibility.
// ======================================================================

// network returns the underlying DualModeNetwork for backward compatibility.
// Prefer using networkCtrl methods directly for new code.
func (app *DualModeApp) network() *core.DualModeNetwork {
	return app.networkCtrl.Network()
}

// dataDir returns the data directory path.
func (app *DualModeApp) dataDir() string {
	return app.networkCtrl.DataDir()
}

// testImages returns the cached test images.
func (app *DualModeApp) testImages() [][]float64 {
	images, _ := app.networkCtrl.GetTestData()
	return images
}

// testLabels returns the cached test labels.
func (app *DualModeApp) testLabels() []int {
	_, labels := app.networkCtrl.GetTestData()
	return labels
}

// currentQATLevel returns the current QAT level.
func (app *DualModeApp) currentQATLevel() int {
	return app.networkCtrl.CurrentQATLevel()
}

// setCurrentQATLevel sets the current QAT level.
func (app *DualModeApp) setCurrentQATLevel(level int) {
	app.networkCtrl.SetCurrentQATLevel(level)
}

// hasWarnedMissingLevel checks if a warning was shown for a missing level.
func (app *DualModeApp) hasWarnedMissingLevel(level int) bool {
	return app.networkCtrl.HasWarnedMissingLevel(level)
}

// setWarnedMissingLevel marks a level as warned.
func (app *DualModeApp) setWarnedMissingLevel(level int) {
	app.networkCtrl.SetWarnedMissingLevel(level)
}

// clearWarnedMissingLevel clears the warning flag for a level.
func (app *DualModeApp) clearWarnedMissingLevel(level int) {
	app.networkCtrl.ClearWarnedMissingLevel(level)
}

// BuildContent creates the UI content for embedding.
func (app *DualModeApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	app.fyneApp = fyneApp
	app.window = parentWindow

	// Create main layout
	content := app.createMainLayout()
	// NOTE: updateWeightHeatmap and changeHiddenSize deferred to Start() to avoid fyne.Do() deadlock

	return content
}

// Start initializes anything that needs to run after UI is visible.
func (app *DualModeApp) Start() {
	// Initialize network display now that UI is ready (deferred from BuildContent to avoid fyne.Do() deadlock)

	// Set layout offsets (deferred from createMainLayout to avoid layout cascade)
	fyne.Do(func() {
		if app.leftSplit != nil {
			app.leftSplit.SetOffset(0.60) // 60% drawing, 40% controls (increased canvas space)
		}
		if app.rightSplit != nil {
			app.rightSplit.SetOffset(0.55) // 55% results, 45% weights
		}
		if app.mainSplit != nil {
			app.mainSplit.SetOffset(0.50) // 50% left, 50% right (balanced layout)
		}
	})

	// Show loading progress
	fyne.Do(func() {
		app.statusLabel.SetText("Loading MNIST module...")
	})

	// Use WaitGroup to coordinate background operations
	var wg sync.WaitGroup
	wg.Add(2)

	// Channel to signal completion and handle errors
	done := make(chan error, 2)

	// Start background goroutines for heavy operations
	go func() {
		defer wg.Done()

		// Background: changeHiddenSize with loading feedback
		done <- nil               // Signal changeHiddenSize started
		app.changeHiddenSize(128) // Call original function without error channel
	}()

	// Background: updateWeightHeatmap with loading feedback
	go func() {
		defer wg.Done()
		app.updateWeightHeatmapWithProgress(done)
	}()

	// Wait for background operations to complete
	go func() {
		wg.Wait()
		close(done)
	}()

	// Handle completion and update UI
	go func() {
		for err := range done {
			if err != nil {
				fyne.Do(func() {
					app.statusLabel.SetText(fmt.Sprintf("Error loading weights: %v", err))
					if app.window != nil {
						dialog.ShowError(fmt.Errorf("Failed to load MNIST weights: %w", err), app.window)
					}
				})
			} else {
				fyne.Do(func() {
					app.statusLabel.SetText(fmt.Sprintf("Loaded network with hidden size %d", 128))
				})
			}
		}
	}()
}

// Stop cleans up resources.
func (app *DualModeApp) Stop() {
	// Nothing to stop
}

// createMainLayout builds the 4-zone responsive layout.
// Uses AdaptiveLayout to switch between desktop (splits) and mobile (tabs) layouts.
func (app *DualModeApp) createMainLayout() fyne.CanvasObject {
	// Status label must be created first (used by callbacks in controls zone)
	app.statusLabel = widget.NewLabel("Ready. Draw a digit or click 'Random' to load a test sample from the MNIST dataset.")

	// Header
	header := app.createHeader()

	// Zone 1: Drawing canvas (top-left)
	zone1 := app.createDrawingZone()

	// Zone 2: Results (top-right)
	zone2 := app.createResultsZone()

	// Zone 3: Controls (bottom-left)
	zone3 := app.createControlsZone()

	// Zone 4: Weight visualization (bottom-right)
	zone4 := app.createWeightZone()

	// Status footer
	footer := app.statusLabel

	// Create AdaptiveLayout for responsive design
	zones := []fyne.CanvasObject{zone1, zone2, zone3, zone4}
	tabLabels := []string{"Draw", "Results", "Config", "Weights"}
	app.adaptiveLayout = sharedwidgets.NewAdaptiveLayout(zones, tabLabels)

	// Set desktop layout builder - creates the split-based layout
	// NOTE: SetOffset is called in Start() after UI is visible to avoid layout cascade
	app.adaptiveLayout.SetDesktopLayout(func(zones []fyne.CanvasObject) fyne.CanvasObject {
		// Left column: Drawing (top) + Controls (bottom)
		leftSplit := container.NewVSplit(zones[0], zones[2])

		// Right column: Results (top) + Weights (bottom)
		rightSplit := container.NewVSplit(zones[1], zones[3])

		// Main horizontal split
		mainSplit := container.NewHSplit(leftSplit, rightSplit)

		// Store references for later access
		app.leftSplit = leftSplit
		app.rightSplit = rightSplit
		app.mainSplit = mainSplit

		return mainSplit
	})

	// Set breakpoint change callback for logging
	app.adaptiveLayout.OnBreakpointChange = func(bp sharedwidgets.Breakpoint) {
		mnistLog.Printf("Breakpoint changed to: %s", sharedwidgets.BreakpointName(bp))
	}

	mainContent := container.NewBorder(
		header,
		footer,
		nil, nil,
		app.adaptiveLayout.Content(),
	)

	// Mark as initialized - but defer changeHiddenSize to Start() to avoid fyne.Do() deadlock
	app.initialized = true
	// NOTE: changeHiddenSize(128) moved to Start() method to avoid deadlock

	return mainContent
}

// createHeader creates the title and info header.
func (app *DualModeApp) createHeader() fyne.CanvasObject {
	// Header with peer-reviewed accuracy context (87% claim removed)
	title := widget.NewLabel("FeCIM MNIST Demo | Peer-reviewed: 96.6-98.24%")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// Quick Demo button - prominent call to action
	quickDemoBtn := widget.NewButton("Quick Demo", func() {
		mnistLog.Button("QuickDemo")
		if app.quickDemoRunning {
			app.StopQuickDemo()
		} else {
			app.StartQuickDemo()
		}
	})
	quickDemoBtn.Importance = widget.WarningImportance // Orange/yellow - stands out

	// Consolidated Info button - opens tabbed dialog with all info
	infoBtn := widget.NewButton("ℹ Info", func() {
		mnistLog.Button("Info")
		ShowInfoDialog(app.window)
	})

	buttonRow := container.NewHBox(
		quickDemoBtn,
		widget.NewSeparator(),
		infoBtn,
	)

	return container.NewHBox(title, layout.NewSpacer(), buttonRow)
}

// createDrawingZone creates the drawing canvas zone (Zone 1).
func (app *DualModeApp) createDrawingZone() fyne.CanvasObject {
	label := widget.NewLabel("Draw Digit")
	label.TextStyle = fyne.TextStyle{Bold: true}

	app.digitCanvas = NewDigitCanvas()
	app.digitCanvas.OnDigitChanged = app.onDigitChanged

	// Pixel count display
	pixelCountLabel := widget.NewLabel("Pixels: 0/784")
	pixelCountLabel.TextStyle = fyne.TextStyle{Monospace: true}
	app.digitCanvas.OnPixelCountChanged = func(count, total int) {
		fyne.Do(func() {
			pixelCountLabel.SetText(fmt.Sprintf("Pixels: %d/%d", count, total))
		})
	}

	// Set default brush size (medium)
	app.digitCanvas.SetBrushSize(BrushMedium)

	clearBtn := widget.NewButton("Clear", func() {
		mnistLog.Button("Clear")
		app.digitCanvas.Clear()
		app.resetResults()
	})

	randomBtn := widget.NewButton("Random", func() {
		mnistLog.Button("Random")
		app.loadRandomSample()
	})

	// Top row: label + pixel count + buttons
	topRow := container.NewHBox(label, layout.NewSpacer(), pixelCountLabel, clearBtn, randomBtn)

	return container.NewVBox(
		topRow,
		container.NewCenter(app.digitCanvas),
	)
}

// createResultsZone creates the FP vs CIM results zone (Zone 2).
// Enhanced with P1 widgets for better visualization.
func (app *DualModeApp) createResultsZone() fyne.CanvasObject {
	// Initialize legacy widgets (kept for compatibility with updateResultDisplays)
	app.fpPredLabel = widget.NewLabel("-")
	app.fpConfBar = widget.NewProgressBar()
	app.cimPredLabel = widget.NewLabel("-")
	app.cimConfBar = widget.NewProgressBar()
	app.agreementLabel = widget.NewLabel("")
	app.energyLabel = widget.NewLabel("Energy: -")

	// Initialize probability bars (legacy)
	for i := 0; i < 10; i++ {
		app.fpProbBars[i] = widget.NewProgressBar()
		app.cimProbBars[i] = widget.NewProgressBar()
	}

	// P1.2: Enhanced Comparison Card
	app.comparisonCard = NewComparisonCard()

	// P1.2: Dual Probability Chart with divergence highlighting
	app.dualProbabilityChart = NewDualProbabilityChart()

	// Use Border layout so probability chart expands to fill space
	resultsLayout := container.NewBorder(
		app.comparisonCard, // Top: comparison card (fixed height)
		nil, nil, nil,
		container.NewMax(app.dualProbabilityChart), // Center: probability chart (expands)
	)

	return resultsLayout
}
