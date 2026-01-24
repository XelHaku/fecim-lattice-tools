// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode.go provides the dual-mode (FP vs CIM) comparison GUI.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/core"
	"multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/mnist"
	"multilayer-ferroelectric-cim-visualizer/shared/logging"
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
	weightHeatmap     *canvas.Raster
	weightLayer       int // 0 = layer 1, 1 = layer 2
	weightDimLabel    *widget.Label
	weightRangeLabel  *widget.Label
	weightLevelsLabel *widget.Label

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
	inferencePhaseLabel *widget.Label  // Shows current inference phase
	quickDemoRunning    bool           // True when quick demo is active
	quickDemoStopChan   chan struct{}  // Channel to stop quick demo
	animationEnabled    bool           // Enable/disable inference animation

	// P2 Enhancement: Weight Comparison
	weightComparisonWidget *WeightComparisonWidget
	dualWeightHeatmap      *DualWeightHeatmap
}

// NewDualModeApp creates a new dual-mode MNIST application.
func NewDualModeApp() *DualModeApp {
	app := &DualModeApp{
		dataDir:         findDataDir(),
		currentQATLevel: 30, // Default QAT level
	}

	// Create network
	app.network = core.NewDualModeNetwork(784, 128, 10)

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
	app.fyneApp = fyneApp
	app.window = parentWindow

	// Create main layout
	content := app.createMainLayout()

	// Initialize weight heatmap
	app.updateWeightHeatmap()

	return content
}

// Start initializes anything that needs to run after UI is visible.
func (app *DualModeApp) Start() {
	// Nothing to start - event-driven
}

// Stop cleans up resources.
func (app *DualModeApp) Stop() {
	// Nothing to stop
}

// createMainLayout builds the 4-zone layout per the plan.
func (app *DualModeApp) createMainLayout() fyne.CanvasObject {
	// Status label must be created first (used by callbacks in controls zone)
	app.statusLabel = widget.NewLabel("Ready. Draw a digit or click 'Random Sample'.")

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

	// Arrange zones using expandable splits to fill available space
	// Left column: Drawing (top) + Controls (bottom)
	leftSplit := container.NewVSplit(zone1, zone3)
	leftSplit.SetOffset(0.35) // 35% drawing, 65% controls

	// Right column: Results (top) + Weights (bottom)
	rightSplit := container.NewVSplit(zone2, zone4)
	rightSplit.SetOffset(0.55) // 55% results, 45% weights

	// Main horizontal split
	mainSplit := container.NewHSplit(leftSplit, rightSplit)
	mainSplit.SetOffset(0.35) // 35% left, 65% right

	mainContent := container.NewBorder(
		header,
		footer,
		nil, nil,
		mainSplit,
	)

	// Mark as initialized and trigger initial setup
	app.initialized = true
	app.changeHiddenSize(128) // Initialize with default hidden size

	return mainContent
}

// createHeader creates the title and info header.
func (app *DualModeApp) createHeader() fyne.CanvasObject {
	title := widget.NewLabel("MNIST FeCIM | 784→128→10 | 30 Levels | 87%")
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
	label := widget.NewLabel("Results")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// FP results (legacy - kept for compatibility)
	fpLabel := widget.NewLabel("FP")
	fpLabel.TextStyle = fyne.TextStyle{Bold: true}
	app.fpPredLabel = widget.NewLabel("-")
	app.fpConfBar = widget.NewProgressBar()

	// CIM results (legacy - kept for compatibility)
	cimLabel := widget.NewLabel("CIM")
	cimLabel.TextStyle = fyne.TextStyle{Bold: true}
	app.cimPredLabel = widget.NewLabel("-")
	app.cimConfBar = widget.NewProgressBar()

	// Agreement
	app.agreementLabel = widget.NewLabel("")

	// Side by side (legacy compact view)
	fpBox := container.NewVBox(fpLabel, app.fpPredLabel, app.fpConfBar)
	cimBox := container.NewVBox(cimLabel, app.cimPredLabel, app.cimConfBar)
	comparison := container.NewGridWithColumns(2, fpBox, cimBox)

	// Probability distribution bars (legacy - 5 per row for 0-9)
	probGrid := container.NewGridWithColumns(5)
	for i := 0; i < 10; i++ {
		app.fpProbBars[i] = widget.NewProgressBar()
		app.cimProbBars[i] = widget.NewProgressBar()
		probGrid.Add(container.NewVBox(
			widget.NewLabel(fmt.Sprintf("%d", i)),
			app.fpProbBars[i],
			app.cimProbBars[i],
		))
	}

	// Energy
	app.energyLabel = widget.NewLabel("Energy: -")

	// P1.2: Enhanced Comparison Card
	app.comparisonCard = NewComparisonCard()

	// P1.2: Dual Probability Chart with divergence highlighting
	app.dualProbabilityChart = NewDualProbabilityChart()

	// Create tabbed view: Enhanced (new) vs Classic (legacy)
	// Use Border layout so probability chart expands to fill space
	enhancedTab := container.NewBorder(
		app.comparisonCard, // Top: comparison card (fixed height)
		nil, nil, nil,
		container.NewMax(app.dualProbabilityChart), // Center: probability chart (expands)
	)

	classicTab := container.NewBorder(
		container.NewVBox(
			container.NewHBox(label, layout.NewSpacer(), app.agreementLabel, app.energyLabel),
			comparison,
		),
		nil, nil, nil,
		container.NewMax(container.NewVBox(probGrid)), // Probability grid expands
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("Enhanced", enhancedTab),
		container.NewTabItem("Classic", classicTab),
	)

	return tabs
}

// createControlsZone creates the hardware control panel (Zone 3).
func (app *DualModeApp) createControlsZone() fyne.CanvasObject {
	label := widget.NewLabel("Hardware Config")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Levels slider (2-31, covers all QAT-trained levels)
	levelsTitle := widget.NewLabel("Levels:")
	app.levelsLabel = widget.NewLabel("30")
	app.levelsSlider = widget.NewSlider(2, 31)
	app.levelsSlider.Step = 1
	app.levelsSlider.Value = 30
	app.levelsSlider.OnChanged = func(v float64) {
		mnistLog.SliderChange("Levels", v)
		levels := int(v)
		app.levelsLabel.SetText(fmt.Sprintf("%d", levels))

		// Check if we should load level-specific QAT weights
		bestLevel := core.GetBestMatchingWeightsLevel(levels)
		app.tryLoadQATWeights(bestLevel)

		app.network.SetNumLevels(levels)
		app.updateWeightHeatmap()
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	levelsRow := container.NewBorder(nil, nil,
		container.NewHBox(levelsTitle, widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			dialog.ShowInformation("Quantization Levels", "Number of discrete conductance states in the ferroelectric device.\n30 levels = ~4.9 bits of precision (FeCIM standard).p", app.window)
		})),
		app.levelsLabel, app.levelsSlider)

	// Noise slider
	noiseTitle := widget.NewLabel("Noise:")
	app.noiseLabel = widget.NewLabel("0.01")
	app.noiseSlider = widget.NewSlider(0.0, 0.20)
	app.noiseSlider.Step = 0.01
	app.noiseSlider.Value = 0.01
	app.noiseSlider.OnChanged = func(v float64) {
		mnistLog.SliderChange("Noise", v)
		app.noiseLabel.SetText(fmt.Sprintf("%.2f", v))
		app.network.SetNoiseLevel(v)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	noiseRow := container.NewBorder(nil, nil,
		container.NewHBox(noiseTitle, widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			dialog.ShowInformation("Read Noise", "Gaussian noise added to analog read operations.\nSimulates thermal noise and device variability.\n0.01 = 1% noise (standard).", app.window)
		})),
		app.noiseLabel, app.noiseSlider)

	// ADC/DAC + Hidden selects on one row
	bitOptions := []string{"3", "4", "5", "6", "7", "8", "10", "12", "16"}
	app.adcSelect = widget.NewSelect(bitOptions, func(s string) {
		mnistLog.Selection("ADC", s)
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		app.network.SetADCBits(bits)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
	app.adcSelect.SetSelected("8")

	app.dacSelect = widget.NewSelect(bitOptions, func(s string) {
		mnistLog.Selection("DAC", s)
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		app.network.SetDACBits(bits)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
	app.dacSelect.SetSelected("8")

	app.hiddenSelect = widget.NewSelect([]string{"64", "128", "256"}, func(s string) {
		mnistLog.Selection("Hidden", s)
		var size int
		fmt.Sscanf(s, "%d", &size)
		app.changeHiddenSize(size)
	})
	app.hiddenSelect.SetSelected("128")

	selectsRow := container.NewGridWithColumns(6,
		widget.NewLabel("ADC:"), app.adcSelect,
		widget.NewLabel("DAC:"), app.dacSelect,
		widget.NewLabel("Hidden:"), app.hiddenSelect,
	)

	// Preset buttons (first row: standard presets)
	idealBtn := widget.NewButton("Ideal", func() { mnistLog.Button("Preset:Ideal"); app.applyPresetWithMode(30, 0.01, 8, 8, false) })
	quantCliffBtn := widget.NewButton("QuantCliff", func() { mnistLog.Button("Preset:QuantCliff"); app.applyPresetWithMode(2, 0.01, 8, 8, false) })
	noisyBtn := widget.NewButton("Noisy", func() { mnistLog.Button("Preset:Noisy"); app.applyPresetWithMode(30, 0.15, 6, 8, false) })
	brokenBtn := widget.NewButton("BrokenADC", func() { mnistLog.Button("Preset:BrokenADC"); app.applyPresetWithMode(30, 0.01, 3, 8, false) })

	// Tour Mode button (single-layer 784→10, matching Dr. Tour's architecture)
	// Achieved 83% accuracy with 30-level quantization (Tour claimed 87% with 88% theoretical max)
	tourBtn := widget.NewButton("Tour", func() { mnistLog.Button("Preset:Tour"); app.applyPresetWithMode(30, 0.01, 8, 8, true) })
	tourBtn.Importance = widget.HighImportance // Highlight this button

	presetRow := container.NewGridWithColumns(5, idealBtn, quantCliffBtn, noisyBtn, brokenBtn, tourBtn)

	// Quick test with progress bar
	app.testResultLabel = widget.NewLabel("")
	app.testProgressBar = widget.NewProgressBar()
	app.testProgressBar.Hide() // Hidden until test starts
	app.testButton = widget.NewButton("Test (200)", func() {
		mnistLog.Button("Test200")
		app.runQuickTest()
	})

	// P1.1: Quantization Visualization Widget
	app.quantizationWidget = NewQuantizationWidget()

	// P1.3: Energy Tracking Widget
	app.energyWidget = NewEnergyWidget(784, 128, 10)

	// Controls header (compact, fixed height)
	controlsHeader := container.NewVBox(
		container.NewHBox(label, layout.NewSpacer(), app.testButton, app.testProgressBar, app.testResultLabel),
		levelsRow,
		noiseRow,
		selectsRow,
		presetRow,
	)

	// P1 widgets that expand to fill space (shown in Config tab)
	p1ExpandingContent := container.NewVSplit(
		app.quantizationWidget,
		app.energyWidget,
	)
	p1ExpandingContent.SetOffset(0.5) // Equal split

	// Config tab: controls at top, P1 widgets fill remaining space
	// This eliminates the need for a separate Quantization tab and uses the space better
	configTab := container.NewBorder(
		controlsHeader,
		nil, nil, nil,
		p1ExpandingContent,
	)

	return configTab
}

// createWeightZone creates the weight visualization zone (Zone 4).
func (app *DualModeApp) createWeightZone() fyne.CanvasObject {
	label := widget.NewLabel("Weights")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Layer selector as horizontal radio
	layerSelect := widget.NewRadioGroup(
		[]string{"Layer1 (784x128)", "Layer2 (128x10)"},
		func(s string) {
			if s == "Layer1 (784x128)" {
				app.weightLayer = 0
			} else {
				app.weightLayer = 1
			}
			app.updateWeightHeatmap()
			app.updateWeightComparison()
		},
	)
	layerSelect.Horizontal = true
	layerSelect.SetSelected("Layer1 (784x128)")

	// Info labels (combined into one)
	app.weightDimLabel = widget.NewLabel("")
	app.weightRangeLabel = widget.NewLabel("")
	app.weightLevelsLabel = widget.NewLabel("")

	// Create heatmap raster (quantized weights)
	app.weightHeatmap = canvas.NewRaster(app.drawWeightHeatmap)
	app.weightHeatmap.SetMinSize(fyne.NewSize(256, 128))

	// Create FP vs Quantized comparison widget
	app.weightComparisonWidget = NewWeightComparisonWidget()

	// Create dual heatmap (side-by-side FP vs Quantized)
	app.dualWeightHeatmap = NewDualWeightHeatmap()

	zoomBtn := widget.NewButtonWithIcon("", theme.ZoomFitIcon(), func() {
		app.showZoomedHeatmap()
	})

	// Header with controls
	header := container.NewVBox(
		container.NewHBox(label, layout.NewSpacer(), layerSelect, zoomBtn),
		container.NewHBox(app.weightDimLabel, app.weightRangeLabel, app.weightLevelsLabel),
	)

	// Create tabbed view for different weight visualizations
	quantizedTab := container.NewMax(app.weightHeatmap)
	comparisonTab := container.NewMax(app.weightComparisonWidget)
	sideBySideTab := container.NewMax(app.dualWeightHeatmap)

	weightTabs := container.NewAppTabs(
		container.NewTabItem("Quantized", quantizedTab),
		container.NewTabItem("FP vs Quant", comparisonTab),
		container.NewTabItem("Side-by-Side", sideBySideTab),
	)

	return container.NewBorder(
		header,
		nil, nil, nil,
		weightTabs,
	)
}

// updateWeightComparison updates the FP vs Quantized comparison widgets.
func (app *DualModeApp) updateWeightComparison() {
	fpW1, fpW2, _, _ := app.network.GetFPWeights()
	quantW1, quantW2, _, _ := app.network.GetQuantWeights()

	// Update comparison widget based on selected layer
	if app.weightComparisonWidget != nil {
		if app.weightLayer == 0 {
			app.weightComparisonWidget.SetWeights(fpW1, quantW1)
		} else {
			app.weightComparisonWidget.SetWeights(fpW2, quantW2)
		}
	}

	// Update dual heatmap
	if app.dualWeightHeatmap != nil {
		app.dualWeightHeatmap.SetWeights(fpW1, fpW2, quantW1, quantW2)
		app.dualWeightHeatmap.SetLayer(app.weightLayer)
	}
}

// onDigitChanged handles canvas drawing updates.
func (app *DualModeApp) onDigitChanged(pixels []float64) {
	app.lastPixels = pixels
	if app.animationEnabled {
		go app.runInferenceAnimated(pixels)
	} else {
		app.runInference(pixels)
	}
}

// runInference runs dual-path inference and updates the UI.
func (app *DualModeApp) runInference(pixels []float64) {
	result := app.network.Infer(pixels)

	// Get quantized weights for P1.1 visualization
	quantWeights, _, _, _ := app.network.GetQuantWeights()

	fyne.Do(func() {
		// Update FP results (legacy)
		app.fpPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.FPPrediction, result.FPConfidence*100))
		app.fpConfBar.SetValue(result.FPConfidence)

		// Update CIM results (legacy)
		app.cimPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.CIMPrediction, result.CIMConfidence*100))
		app.cimConfBar.SetValue(result.CIMConfidence)

		// Update agreement (legacy)
		if result.Agree {
			app.agreementLabel.SetText("PREDICTIONS MATCH")
		} else {
			app.agreementLabel.SetText(fmt.Sprintf("DISAGREEMENT (KL=%.3f)", result.Disagreement))
		}

		// Update probability bars (legacy)
		for i := 0; i < 10; i++ {
			app.fpProbBars[i].SetValue(result.FPProbabilities[i])
			app.cimProbBars[i].SetValue(result.CIMProbabilities[i])
		}

		// Update energy (legacy)
		gpuEnergy := result.EnergyUsed * 10000 // Estimated 10,000x for GPU
		app.energyLabel.SetText(fmt.Sprintf("Energy: %.2f uJ (FeCIM) vs %.0f mJ (GPU) = %.0fx savings",
			result.EnergyUsed, gpuEnergy/1000, 10000.0))

		// Update status
		app.statusLabel.SetText(fmt.Sprintf("FP: %d (%.1f%%) | CIM: %d (%.1f%%) | %s",
			result.FPPrediction, result.FPConfidence*100,
			result.CIMPrediction, result.CIMConfidence*100,
			map[bool]string{true: "MATCH", false: "MISMATCH"}[result.Agree]))

		// === P1 ENHANCEMENTS ===

		// P1.1: Update quantization visualization with sample weights
		if app.quantizationWidget != nil && len(quantWeights) > 0 {
			app.quantizationWidget.SetNumLevels(app.network.GetNumLevels())
			app.quantizationWidget.UpdateWithWeights(quantWeights, 5) // Show 5 sample weights
		}

		// P1.2: Update enhanced comparison card
		if app.comparisonCard != nil {
			compResult := &ComparisonResult{
				FPPrediction:     result.FPPrediction,
				FPConfidence:     result.FPConfidence,
				FPProbabilities:  result.FPProbabilities,
				CIMPrediction:    result.CIMPrediction,
				CIMConfidence:    result.CIMConfidence,
				CIMProbabilities: result.CIMProbabilities,
				Match:            result.Agree,
				ConfidenceDelta:  result.FPConfidence - result.CIMConfidence,
				EnergyFeCIM:      result.EnergyUsed * 1e6,         // Convert to nJ
				EnergyGPU:        result.EnergyUsed * 1e6 * 10000, // 10,000x for GPU
				EnergyRatio:      10000.0,
			}
			if compResult.ConfidenceDelta < 0 {
				compResult.ConfidenceDelta = -compResult.ConfidenceDelta
			}
			app.comparisonCard.SetResult(compResult)
		}

		// P1.2: Update dual probability chart
		if app.dualProbabilityChart != nil {
			app.dualProbabilityChart.SetProbabilities(
				result.FPProbabilities,
				result.CIMProbabilities,
				result.FPPrediction,
				result.CIMPrediction,
			)
		}

		// P1.3: Record inference in energy widget
		if app.energyWidget != nil {
			app.energyWidget.RecordInference()
		}
	})
}

// runInferenceAnimated performs inference with visual animation phases.
// This shows the data flow through the network for educational purposes.
func (app *DualModeApp) runInferenceAnimated(pixels []float64) {
	// Phase 1: Input Processing (150ms)
	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Phase 1: Processing input (784 pixels)...")
		}
		app.statusLabel.SetText("INFERENCE | Phase 1: Input → 784 neurons")
	})
	time.Sleep(150 * time.Millisecond)

	// Phase 2: Layer 1 MVM (200ms)
	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Phase 2: Layer 1 MVM (784×128 = 100,352 MACs)...")
		}
		app.statusLabel.SetText("INFERENCE | Phase 2: MVM 784→128 (100,352 parallel MACs)")
	})
	time.Sleep(200 * time.Millisecond)

	// Phase 3: Layer 2 MVM (150ms)
	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Phase 3: Layer 2 MVM (128×10 = 1,280 MACs)...")
		}
		app.statusLabel.SetText("INFERENCE | Phase 3: MVM 128→10 (1,280 parallel MACs)")
	})
	time.Sleep(150 * time.Millisecond)

	// Phase 4: Result - run actual inference and display
	result := app.network.Infer(pixels)
	quantWeights, _, _, _ := app.network.GetQuantWeights()

	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Complete! 101,632 MACs in ~5µs")
		}

		// Update all displays (same as runInference)
		app.updateResultDisplays(result, quantWeights)

		// Show dramatic match/mismatch feedback
		if result.Agree {
			app.statusLabel.SetText(fmt.Sprintf("MATCH | FP: %d | CIM: %d | Confidence: %.1f%% | 10,000x energy savings",
				result.FPPrediction, result.CIMPrediction, result.CIMConfidence*100))
		} else {
			app.statusLabel.SetText(fmt.Sprintf("MISMATCH | FP: %d vs CIM: %d | Check hardware config!",
				result.FPPrediction, result.CIMPrediction))
		}
	})
}

// updateResultDisplays updates all UI elements with inference results.
// Extracted from runInference to avoid duplication.
func (app *DualModeApp) updateResultDisplays(result *core.InferenceResult, quantWeights [][]float64) {
	// Update FP results (legacy)
	app.fpPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.FPPrediction, result.FPConfidence*100))
	app.fpConfBar.SetValue(result.FPConfidence)

	// Update CIM results (legacy)
	app.cimPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.CIMPrediction, result.CIMConfidence*100))
	app.cimConfBar.SetValue(result.CIMConfidence)

	// Update agreement (legacy)
	if result.Agree {
		app.agreementLabel.SetText("PREDICTIONS MATCH")
	} else {
		app.agreementLabel.SetText(fmt.Sprintf("DISAGREEMENT (KL=%.3f)", result.Disagreement))
	}

	// Update probability bars (legacy)
	for i := 0; i < 10; i++ {
		app.fpProbBars[i].SetValue(result.FPProbabilities[i])
		app.cimProbBars[i].SetValue(result.CIMProbabilities[i])
	}

	// Update energy (legacy)
	gpuEnergy := result.EnergyUsed * 10000
	app.energyLabel.SetText(fmt.Sprintf("Energy: %.2f uJ (FeCIM) vs %.0f mJ (GPU) = %.0fx savings",
		result.EnergyUsed, gpuEnergy/1000, 10000.0))

	// P1 Enhancements
	if app.quantizationWidget != nil && len(quantWeights) > 0 {
		app.quantizationWidget.SetNumLevels(app.network.GetNumLevels())
		app.quantizationWidget.UpdateWithWeights(quantWeights, 5)
	}

	if app.comparisonCard != nil {
		compResult := &ComparisonResult{
			FPPrediction:     result.FPPrediction,
			FPConfidence:     result.FPConfidence,
			FPProbabilities:  result.FPProbabilities,
			CIMPrediction:    result.CIMPrediction,
			CIMConfidence:    result.CIMConfidence,
			CIMProbabilities: result.CIMProbabilities,
			Match:            result.Agree,
			ConfidenceDelta:  result.FPConfidence - result.CIMConfidence,
			EnergyFeCIM:      result.EnergyUsed * 1e6,
			EnergyGPU:        result.EnergyUsed * 1e6 * 10000,
			EnergyRatio:      10000.0,
		}
		if compResult.ConfidenceDelta < 0 {
			compResult.ConfidenceDelta = -compResult.ConfidenceDelta
		}
		app.comparisonCard.SetResult(compResult)
	}

	if app.dualProbabilityChart != nil {
		app.dualProbabilityChart.SetProbabilities(
			result.FPProbabilities,
			result.CIMProbabilities,
			result.FPPrediction,
			result.CIMPrediction,
		)
	}

	if app.energyWidget != nil {
		app.energyWidget.RecordInference()
	}
}

// StartQuickDemo runs an automated 30-second demonstration of FeCIM's key insight.
// Shows: Ideal (30 levels) → Success → Break (2 levels) → Failure → Restore
func (app *DualModeApp) StartQuickDemo() {
	if app.quickDemoRunning {
		return
	}

	app.quickDemoRunning = true
	app.quickDemoStopChan = make(chan struct{})
	app.animationEnabled = true

	go func() {
		defer func() {
			app.quickDemoRunning = false
			app.animationEnabled = false
		}()

		// Step 1: Introduction (2s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 1/5: Welcome to FeCIM - Watch the magic of 30 analog levels!")
		})
		if app.waitOrStop(2 * time.Second) {
			return
		}

		// Step 2: Load sample and show ideal prediction (3s)
		fyne.Do(func() {
			app.applyPresetWithMode(30, 0.01, 8, 8, false)
			app.statusLabel.SetText("QUICK DEMO | Step 2/5: Loading test digit with 30 levels (ideal)...")
		})
		if app.waitOrStop(500 * time.Millisecond) {
			return
		}
		fyne.Do(func() {
			app.loadRandomSample()
		})
		if app.waitOrStop(2500 * time.Millisecond) {
			return
		}

		// Step 3: Show success with 30 levels (3s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 3/5: 30 LEVELS = HIGH ACCURACY! FP and CIM predictions match.")
		})
		if app.waitOrStop(3 * time.Second) {
			return
		}

		// Step 4: Break it with 2 levels (4s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 4/5: Now watch what happens with only 2 levels (binary)...")
		})
		if app.waitOrStop(1 * time.Second) {
			return
		}
		fyne.Do(func() {
			app.applyPresetWithMode(2, 0.01, 8, 8, false)
		})
		if app.waitOrStop(500 * time.Millisecond) {
			return
		}
		// Re-run inference with same digit
		if len(app.lastPixels) > 0 {
			fyne.Do(func() {
				app.runInference(app.lastPixels)
			})
		}
		if app.waitOrStop(2500 * time.Millisecond) {
			return
		}

		// Step 5: Show failure explanation (3s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 5/5: 2 LEVELS = FAILURE! Binary weights cannot represent the network.")
		})
		if app.waitOrStop(3 * time.Second) {
			return
		}

		// Restore ideal settings (2s)
		fyne.Do(func() {
			app.applyPresetWithMode(30, 0.01, 8, 8, false)
			app.statusLabel.SetText("DEMO COMPLETE | Key insight: 30 levels enable 87% accuracy with 10,000x energy savings!")
		})
		if app.waitOrStop(500 * time.Millisecond) {
			return
		}
		// Re-run inference with restored settings
		if len(app.lastPixels) > 0 {
			fyne.Do(func() {
				app.runInference(app.lastPixels)
			})
		}
	}()
}

// StopQuickDemo stops the running quick demo.
func (app *DualModeApp) StopQuickDemo() {
	if app.quickDemoRunning && app.quickDemoStopChan != nil {
		close(app.quickDemoStopChan)
	}
}

// waitOrStop waits for duration or returns true if demo was stopped.
func (app *DualModeApp) waitOrStop(d time.Duration) bool {
	select {
	case <-app.quickDemoStopChan:
		return true
	case <-time.After(d):
		return false
	}
}

// resetResults clears the result displays.
func (app *DualModeApp) resetResults() {
	app.lastPixels = nil
	app.fpPredLabel.SetText("Prediction: -")
	app.fpConfBar.SetValue(0)
	app.cimPredLabel.SetText("Prediction: -")
	app.cimConfBar.SetValue(0)
	app.agreementLabel.SetText("")
	app.energyLabel.SetText("Energy: -")
	for i := 0; i < 10; i++ {
		app.fpProbBars[i].SetValue(0)
		app.cimProbBars[i].SetValue(0)
	}
	app.statusLabel.SetText("Ready. Draw a digit or click 'Random Sample'.")

	// Clear P1 widgets
	if app.quantizationWidget != nil {
		app.quantizationWidget.Clear()
	}
	if app.comparisonCard != nil {
		app.comparisonCard.Clear()
	}
	if app.dualProbabilityChart != nil {
		app.dualProbabilityChart.Clear()
	}
	// Note: Energy widget is not cleared to keep session totals
}

// applyPreset sets hardware parameters to a failure mode preset.
func (app *DualModeApp) applyPreset(levels int, noise float64, adcBits, dacBits int) {
	app.applyPresetWithMode(levels, noise, adcBits, dacBits, false)
}

// applyPresetWithMode sets hardware parameters with optional Tour Mode (single-layer).
func (app *DualModeApp) applyPresetWithMode(levels int, noise float64, adcBits, dacBits int, singleLayer bool) {
	// Update network parameters (thread-safe, not UI)
	app.network.SetNumLevels(levels)
	app.network.SetNoiseLevel(noise)
	app.network.SetADCBits(adcBits)
	app.network.SetDACBits(dacBits)
	app.network.SetSingleLayer(singleLayer)

	// Update UI elements on main thread
	fyne.Do(func() {
		app.levelsSlider.SetValue(float64(levels))
		app.noiseSlider.SetValue(noise)
		app.adcSelect.SetSelected(fmt.Sprintf("%d", adcBits))
		app.dacSelect.SetSelected(fmt.Sprintf("%d", dacBits))

		// Update status to indicate Tour Mode
		if singleLayer {
			app.statusLabel.SetText("Tour Mode: Single-layer (784→10) ~83% trained accuracy")
		}

		app.updateWeightHeatmap()

		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
}

// loadRandomSample loads a random test sample.
func (app *DualModeApp) loadRandomSample() {
	if len(app.testImages) == 0 {
		app.loadTestData()
		if len(app.testImages) == 0 {
			fyne.Do(func() {
				app.statusLabel.SetText("No test data available")
			})
			return
		}
	}

	idx := int(time.Now().UnixNano() % int64(len(app.testImages)))
	pixels := app.testImages[idx]
	label := app.testLabels[idx]

	fyne.Do(func() {
		app.digitCanvas.SetPixels(pixels)
		app.statusLabel.SetText(fmt.Sprintf("Loaded sample #%d (true label: %d)", idx, label))
		app.onDigitChanged(pixels)
	})
}

// loadTestData loads MNIST test data.
func (app *DualModeApp) loadTestData() {
	images, labels, err := mnist.LoadMNIST(app.dataDir, false) // false = test set
	if err != nil {
		mnistLog.Printf("Failed to load MNIST test data: %v, using synthetic data", err)
		app.testImages, app.testLabels = generateSyntheticData(200)
		// Notify user that we're using synthetic data
		fyne.Do(func() {
			app.statusLabel.SetText("Using synthetic test data (MNIST not found)")
		})
		return
	}

	if len(images) > 1000 {
		app.testImages = images[:1000]
		app.testLabels = labels[:1000]
	} else {
		app.testImages = images
		app.testLabels = labels
	}
}

// runQuickTest runs inference on 200 samples and reports accuracy.
func (app *DualModeApp) runQuickTest() {
	fyne.Do(func() {
		app.testButton.Disable()
		app.testResultLabel.SetText("")
		app.testProgressBar.Show()
		app.testProgressBar.SetValue(0)
	})

	go func() {
		if len(app.testImages) == 0 {
			fyne.Do(func() {
				app.testResultLabel.SetText("Loading test data...")
			})
			app.loadTestData()
		}

		n := 200
		if n > len(app.testImages) {
			n = len(app.testImages)
		}

		fpCorrect := 0
		cimCorrect := 0
		agreements := 0

		// Update progress every 10 samples to avoid UI overload
		updateInterval := 10

		for i := 0; i < n; i++ {
			result := app.network.Infer(app.testImages[i])
			if result.FPPrediction == app.testLabels[i] {
				fpCorrect++
			}
			if result.CIMPrediction == app.testLabels[i] {
				cimCorrect++
			}
			if result.Agree {
				agreements++
			}

			// Update progress bar at intervals
			if (i+1)%updateInterval == 0 || i == n-1 {
				progress := float64(i+1) / float64(n)
				current := i + 1
				fyne.Do(func() {
					app.testProgressBar.SetValue(progress)
					app.testResultLabel.SetText(fmt.Sprintf("Testing %d/%d...", current, n))
				})
			}
		}

		fpAcc := float64(fpCorrect) / float64(n) * 100
		cimAcc := float64(cimCorrect) / float64(n) * 100
		agreeRate := float64(agreements) / float64(n) * 100

		fyne.Do(func() {
			app.testProgressBar.Hide()
			app.testResultLabel.SetText(fmt.Sprintf(
				"FP: %.1f%% | CIM: %.1f%% | Agreement: %.1f%% | Target: 87%%",
				fpAcc, cimAcc, agreeRate))
			app.testButton.Enable()
		})
	}()
}

// drawWeightHeatmap draws the weight heatmap.
func (app *DualModeApp) drawWeightHeatmap(w, h int) image.Image {
	_, w2, _, _ := app.network.GetQuantWeights()

	var weights [][]float64
	if app.weightLayer == 0 {
		weights, _, _, _ = app.network.GetQuantWeights()
	} else {
		weights = w2
	}

	if len(weights) == 0 || len(weights[0]) == 0 {
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}

	rows := len(weights)
	cols := len(weights[0])

	// Find range
	wMin, wMax := weights[0][0], weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < wMin {
				wMin = weights[i][j]
			}
			if weights[i][j] > wMax {
				wMax = weights[i][j]
			}
		}
	}

	// Create image with proper scaling
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	scaleX := float64(cols) / float64(w)
	scaleY := float64(rows) / float64(h)

	wRange := wMax - wMin
	if wRange == 0 {
		wRange = 1
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Map to weight matrix
			row := int(float64(y) * scaleY)
			col := int(float64(x) * scaleX)

			if row >= rows {
				row = rows - 1
			}
			if col >= cols {
				col = cols - 1
			}

			val := weights[row][col]
			normalized := (val - wMin) / wRange

			// Blue-White-Red colormap
			c := weightToColor(normalized)
			img.Set(x, y, c)
		}
	}

	return img
}

// weightToColor converts normalized value [0,1] to color.
func weightToColor(normalized float64) color.Color {
	if normalized < 0.5 {
		// Blue to white
		t := normalized * 2
		r := uint8(t * 255)
		g := uint8(t * 255)
		b := uint8(255)
		return color.RGBA{r, g, b, 255}
	}
	// White to red
	t := (normalized - 0.5) * 2
	r := uint8(255)
	g := uint8((1 - t) * 255)
	b := uint8((1 - t) * 255)
	return color.RGBA{r, g, b, 255}
}

// updateWeightHeatmap refreshes the weight visualization.
func (app *DualModeApp) updateWeightHeatmap() {
	if !app.initialized {
		return
	}
	if app.weightHeatmap != nil {
		fyne.Do(func() {
			app.weightHeatmap.Refresh()
		})
	}

	// Update info labels
	var weights [][]float64
	if app.weightLayer == 0 {
		weights, _, _, _ = app.network.GetQuantWeights()
		if len(weights) > 0 && len(weights[0]) > 0 {
			app.weightDimLabel.SetText(fmt.Sprintf("Dimensions: %d rows x %d cols", len(weights), len(weights[0])))
		}
	} else {
		_, weights, _, _ = app.network.GetQuantWeights()
		if len(weights) > 0 && len(weights[0]) > 0 {
			app.weightDimLabel.SetText(fmt.Sprintf("Dimensions: %d rows x %d cols", len(weights), len(weights[0])))
		}
	}

	if len(weights) > 0 && len(weights[0]) > 0 {
		// Find range
		wMin, wMax := weights[0][0], weights[0][0]
		distinctMap := make(map[float64]bool)
		for i := range weights {
			for j := range weights[i] {
				if weights[i][j] < wMin {
					wMin = weights[i][j]
				}
				if weights[i][j] > wMax {
					wMax = weights[i][j]
				}
				distinctMap[weights[i][j]] = true
			}
		}
		app.weightRangeLabel.SetText(fmt.Sprintf("Range: [%.3f, %.3f]", wMin, wMax))
		app.weightLevelsLabel.SetText(fmt.Sprintf("Distinct levels: %d (FeCIM max: 30)", len(distinctMap)))
	}

	// Also update FP vs Quantized comparison
	app.updateWeightComparison()
}

// changeHiddenSize changes the network hidden layer size.
// This requires loading different pretrained weights.
func (app *DualModeApp) changeHiddenSize(size int) {
	// Skip during initialization - will be called again after UI is ready
	if !app.initialized {
		return
	}

	// Map size to weight file
	weightsFile := fmt.Sprintf("pretrained_30_h%d.json", size)
	weightsPath := filepath.Join(app.dataDir, weightsFile)

	// Check if file exists, fallback to default
	if _, err := os.Stat(weightsPath); os.IsNotExist(err) {
		// Try default weights file
		weightsPath = filepath.Join(app.dataDir, "pretrained_weights.json")
		app.statusLabel.SetText(fmt.Sprintf("Note: Using default weights (h%d weights not found)", size))
	}

	// Create new network with specified hidden size
	app.network = core.NewDualModeNetwork(784, size, 10)

	// Load weights
	if err := app.network.LoadWeights(weightsPath); err != nil {
		app.statusLabel.SetText("Error loading weights")
		if app.window != nil {
			dialog.ShowError(fmt.Errorf("failed to load weights from %s: %w", weightsPath, err), app.window)
		}
		return
	}

	// Reset results and update heatmap
	app.resetResults()
	app.updateWeightHeatmap()
	app.statusLabel.SetText(fmt.Sprintf("Loaded network with hidden size %d", size))
}

// tryLoadQATWeights attempts to load QAT weights optimized for the given level.
// Only reloads if the optimal weights are different from currently loaded.
func (app *DualModeApp) tryLoadQATWeights(targetLevel int) {
	// Check if we already have optimal weights loaded
	if app.currentQATLevel == targetLevel {
		return
	}

	// Find the weights file for this level
	weightsPath := core.GetWeightsFilename(app.dataDir, targetLevel)

	// Check if the file exists
	if _, err := os.Stat(weightsPath); os.IsNotExist(err) {
		// No level-specific weights, keep current
		return
	}

	// Load the new weights
	if err := app.network.LoadWeights(weightsPath); err != nil {
		// Failed to load, keep current
		return
	}

	// Update tracking
	app.currentQATLevel = targetLevel

	// Update status
	if app.statusLabel != nil {
		app.statusLabel.SetText(fmt.Sprintf("Loaded QAT weights for %d levels", targetLevel))
	}
}

// showZoomedHeatmap opens a new window with a larger view of the weight heatmap.
func (app *DualModeApp) showZoomedHeatmap() {
	w := app.fyneApp.NewWindow("Weight Matrix Detail")
	w.Resize(fyne.NewSize(800, 600))

	// Create a new raster that uses the same drawing function
	// We wrap it to ensure it uses the current App state (which it does via method receiver)
	heatmap := canvas.NewRaster(app.drawWeightHeatmap)

	// Add a close button
	closeBtn := widget.NewButton("Close", func() {
		w.Close()
	})

	content := container.NewBorder(
		nil,
		container.NewHBox(layout.NewSpacer(), closeBtn),
		nil, nil,
		container.NewMax(heatmap),
	)

	w.SetContent(content)
	w.Show()
}
