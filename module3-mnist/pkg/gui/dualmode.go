// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode.go provides the dual-mode (FP vs CIM) comparison GUI.
package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/core"
	"multilayer-ferroelectric-cim-visualizer/shared/logging"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// Constants for MNIST network configuration
const (
	// Network architecture
	MNISTInputSize   = 784 // 28x28 pixel images
	MNISTHiddenSize  = 128 // Default hidden layer size
	MNISTOutputSize  = 10  // Digits 0-9
	MNISTTotalMACs   = 101632 // Total multiply-accumulate operations per inference

	// FeCIM hardware parameters
	FeCIMDefaultLevels = 30    // Standard 30-level quantization
	FeCIMDefaultNoise  = 0.01  // 1% standard deviation (typical production)
	FeCIMDefaultADC    = 8     // 8-bit ADC resolution
	FeCIMDefaultDAC    = 8     // 8-bit DAC resolution

	// Energy efficiency
	FeCIMEnergyPerMAC = 50e-15  // 50 fJ/MAC (femtojoules)
	GPUEnergyPerMAC   = 500e-12 // 500 pJ/MAC (with DRAM access)
	EnergyRatioGPU    = 10000   // GPU uses 10,000x more energy

	// Accuracy targets
	TargetHardwareAccuracy = 0.87 // 87% measured on real FeCIM hardware
	TargetFP32Accuracy     = 0.98 // 98% theoretical with Float32
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

	// Network
	network *core.DualModeNetwork
	dataDir string

	// Drawing
	digitCanvas *DigitCanvas

	// Control panel components
	levelsSlider *widget.Slider
	levelsLabel  *widget.Label
	noiseSlider  *widget.Slider
	noiseLabel   *widget.Label
	adcSelect    *widget.Select
	dacSelect    *widget.Select
	hiddenSelect *widget.Select

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

	// Test data
	testImages [][]float64
	testLabels []int

	// Status
	statusLabel *widget.Label
	initialized bool // true after UI is fully built

	// Last inference result for refresh
	lastPixels []float64

	// QAT (Quantization-Aware Training) weight tracking
	currentQATLevel int // Currently loaded QAT weights level (10, 20, 29, 30, 31)

	// Guided Tour
	tour *GuidedTour

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
		dataDir:         findDataDir(),
		currentQATLevel: FeCIMDefaultLevels, // Default QAT level (30)
	}

	// Create network
	app.network = core.NewDualModeNetwork(MNISTInputSize, MNISTHiddenSize, MNISTOutputSize)

	// Load pretrained weights (default 30-level QAT weights)
	weightsPath := filepath.Join(app.dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		if err := app.network.LoadWeights(weightsPath); err != nil {
			mnistLog.Printf("Warning: Failed to load weights from %s: %v", weightsPath, err)
		}
	} else {
		mnistLog.Printf("Note: No pretrained weights found at %s, using random initialization", weightsPath)
	}

	return app
}

// BuildContent creates the UI content for embedding.
func (app *DualModeApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	fmt.Println("[MNIST] BuildContent: start")
	app.fyneApp = fyneApp
	app.window = parentWindow

	// Create main layout
	fmt.Println("[MNIST] BuildContent: calling createMainLayout...")
	content := app.createMainLayout()
	fmt.Println("[MNIST] BuildContent: done (deferred init to Start)")
	// NOTE: updateWeightHeatmap and changeHiddenSize deferred to Start() to avoid fyne.Do() deadlock

	return content
}

// Start initializes anything that needs to run after UI is visible.
func (app *DualModeApp) Start() {
	// Initialize network display now that UI is ready (deferred from BuildContent to avoid fyne.Do() deadlock)
	fmt.Println("[MNIST] Start: initializing network display...")

	// Set layout offsets (deferred from createMainLayout to avoid layout cascade)
	fyne.Do(func() {
		if app.leftSplit != nil {
			app.leftSplit.SetOffset(0.50) // 50% drawing, 50% controls (increased canvas space)
		}
		if app.rightSplit != nil {
			app.rightSplit.SetOffset(0.55) // 55% results, 45% weights
		}
		if app.mainSplit != nil {
			app.mainSplit.SetOffset(0.40) // 40% left, 60% right (increased left column width)
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
					fmt.Println("[MNIST] Start: done")
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
	fmt.Println("[MNIST] createMainLayout: start")
	// Status label must be created first (used by callbacks in controls zone)
	app.statusLabel = widget.NewLabel("Ready. Draw a digit or click 'Random' to load a test sample from the MNIST dataset.")

	// Header
	fmt.Println("[MNIST] createMainLayout: creating header...")
	header := app.createHeader()

	// Zone 1: Drawing canvas (top-left)
	fmt.Println("[MNIST] createMainLayout: creating zone1 (drawing)...")
	zone1 := app.createDrawingZone()

	// Zone 2: Results (top-right)
	fmt.Println("[MNIST] createMainLayout: creating zone2 (results)...")
	zone2 := app.createResultsZone()

	// Zone 3: Controls (bottom-left)
	fmt.Println("[MNIST] createMainLayout: creating zone3 (controls)...")
	zone3 := app.createControlsZone()

	// Zone 4: Weight visualization (bottom-right)
	fmt.Println("[MNIST] createMainLayout: creating zone4 (weights)...")
	zone4 := app.createWeightZone()

	// Status footer
	footer := app.statusLabel

	// Create AdaptiveLayout for responsive design
	fmt.Println("[MNIST] createMainLayout: creating adaptive layout...")
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

	fmt.Println("[MNIST] createMainLayout: creating border...")
	mainContent := container.NewBorder(
		header,
		footer,
		nil, nil,
		app.adaptiveLayout.Content(),
	)

	// Mark as initialized - but defer changeHiddenSize to Start() to avoid fyne.Do() deadlock
	fmt.Println("[MNIST] createMainLayout: setting initialized=true...")
	app.initialized = true
	// NOTE: changeHiddenSize(128) moved to Start() method to avoid deadlock

	fmt.Println("[MNIST] createMainLayout: returning")
	return mainContent
}

// createHeader creates the title and info header.
func (app *DualModeApp) createHeader() fyne.CanvasObject {
	title := widget.NewLabel(fmt.Sprintf("MNIST FeCIM | %d→%d→%d | %d Levels | %.0f%% Target",
		MNISTInputSize, MNISTHiddenSize, MNISTOutputSize,
		FeCIMDefaultLevels, TargetHardwareAccuracy*100))
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

	// Guided Tour button
	tourBtn := widget.NewButton("Tour", func() {
		mnistLog.Button("GuidedTour")
		if app.tour == nil {
			app.tour = NewGuidedTour(app)
		}
		app.tour.Start()
	})
	tourBtn.Importance = widget.HighImportance

	// Info buttons
	why30Btn := widget.NewButton("Why 30?", func() {
		ShowWhy30LevelsDialog(app.window)
	})

	realityBtn := widget.NewButton("HW Reality", func() {
		ShowHardwareRealityDialog(app.window)
	})

	failuresBtn := widget.NewButton("Failures", func() {
		ShowFailureModesDialog(app.window)
	})

	aboutBtn := widget.NewButton("About", func() {
		ShowAboutDialog(app.window)
	})

	buttonRow := container.NewHBox(
		quickDemoBtn,
		widget.NewSeparator(),
		tourBtn,
		why30Btn,
		realityBtn,
		failuresBtn,
		aboutBtn,
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

	// Brush size selector
	brushSelect := widget.NewSelect([]string{"Thin", "Medium (Recommended)", "Thick"}, func(s string) {
		switch s {
		case "Thin":
			app.digitCanvas.SetBrushSize(BrushThin)
		case "Medium (Recommended)":
			app.digitCanvas.SetBrushSize(BrushMedium)
		case "Thick":
			app.digitCanvas.SetBrushSize(BrushThick)
		}
	})
	brushSelect.SetSelected("Medium (Recommended)")

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

	// Bottom row: brush selector
	brushRow := container.NewHBox(widget.NewLabel("Brush:"), brushSelect)

	return container.NewVBox(
		topRow,
		container.NewCenter(app.digitCanvas),
		brushRow,
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
