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
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/core"
)

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
	fpPredLabel  *widget.Label
	fpConfBar    *widget.ProgressBar
	cimPredLabel *widget.Label
	cimConfBar   *widget.ProgressBar
	agreementLabel *widget.Label
	energyLabel  *widget.Label
	fpProbBars   [10]*widget.ProgressBar
	cimProbBars  [10]*widget.ProgressBar

	// Weight visualization
	weightHeatmap *canvas.Raster
	weightLayer   int // 0 = layer 1, 1 = layer 2
	weightDimLabel *widget.Label
	weightRangeLabel *widget.Label
	weightLevelsLabel *widget.Label

	// Quick test
	testButton *widget.Button
	testResultLabel *widget.Label

	// Test data
	testImages [][]float64
	testLabels []int

	// Status
	statusLabel *widget.Label
	initialized bool // true after UI is fully built

	// Last inference result for refresh
	lastPixels []float64

	// Guided Tour
	tour *GuidedTour
}

// NewDualModeApp creates a new dual-mode MNIST application.
func NewDualModeApp() *DualModeApp {
	app := &DualModeApp{
		dataDir: findDataDir(),
	}

	// Create network
	app.network = core.NewDualModeNetwork(784, 128, 10)

	// Load pretrained weights
	weightsPath := filepath.Join(app.dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		if err := app.network.LoadWeights(weightsPath); err != nil {
			fmt.Printf("Warning: Failed to load weights: %v\n", err)
		}
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
	footer := container.NewVBox(
		widget.NewSeparator(),
		app.statusLabel,
	)

	// Arrange zones in 2x2 grid
	topRow := container.NewGridWithColumns(2, zone1, zone2)
	bottomRow := container.NewGridWithColumns(2, zone3, zone4)

	mainContent := container.NewBorder(
		header,
		footer,
		nil, nil,
		container.NewVBox(
			topRow,
			widget.NewSeparator(),
			bottomRow,
		),
	)

	// Mark as initialized and trigger initial setup
	app.initialized = true
	app.changeHiddenSize(128) // Initialize with default hidden size

	return mainContent
}

// createHeader creates the title and info header.
func (app *DualModeApp) createHeader() fyne.CanvasObject {
	title := widget.NewLabel("MNIST FeCIM Demo - FP vs CIM Comparison")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("\"We're at 87% validation here... theoretical is 88%.\" - Dr. external research group")
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	specs := widget.NewLabel("Architecture: 784 -> 128 -> 10 | FeCIM: 30 Discrete Levels | Target: 87%")
	specs.Alignment = fyne.TextAlignCenter

	// Tour and info buttons
	tourBtn := widget.NewButton("Start Guided Tour", func() {
		if app.tour == nil {
			app.tour = NewGuidedTour(app)
		}
		app.tour.Start()
	})
	tourBtn.Importance = widget.HighImportance

	why30Btn := widget.NewButton("Why 30?", func() {
		ShowWhy30LevelsDialog(app.window)
	})

	realityBtn := widget.NewButton("Hardware Reality", func() {
		ShowHardwareRealityDialog(app.window)
	})

	failuresBtn := widget.NewButton("Failure Modes", func() {
		ShowFailureModesDialog(app.window)
	})

	aboutBtn := widget.NewButton("About", func() {
		ShowAboutDialog(app.window)
	})

	buttonRow := container.NewHBox(
		tourBtn,
		widget.NewSeparator(),
		why30Btn,
		realityBtn,
		failuresBtn,
		aboutBtn,
	)

	return container.NewVBox(
		title,
		subtitle,
		specs,
		container.NewCenter(buttonRow),
		widget.NewSeparator(),
	)
}

// createDrawingZone creates the drawing canvas zone (Zone 1).
func (app *DualModeApp) createDrawingZone() fyne.CanvasObject {
	label := widget.NewLabel("Draw Digit Here")
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	app.digitCanvas = NewDigitCanvas()
	app.digitCanvas.OnDigitChanged = app.onDigitChanged

	clearBtn := widget.NewButton("Clear", func() {
		app.digitCanvas.Clear()
		app.resetResults()
	})

	randomBtn := widget.NewButton("Random Sample", func() {
		app.loadRandomSample()
	})

	buttons := container.NewHBox(clearBtn, randomBtn)

	return container.NewVBox(
		label,
		container.NewCenter(app.digitCanvas),
		buttons,
	)
}

// createResultsZone creates the FP vs CIM results zone (Zone 2).
func (app *DualModeApp) createResultsZone() fyne.CanvasObject {
	label := widget.NewLabel("Inference Results")
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	// FP results
	fpLabel := widget.NewLabel("Digital (FP)")
	fpLabel.TextStyle = fyne.TextStyle{Bold: true}
	app.fpPredLabel = widget.NewLabel("Prediction: -")
	app.fpConfBar = widget.NewProgressBar()

	// CIM results
	cimLabel := widget.NewLabel("FeCIM (Analog)")
	cimLabel.TextStyle = fyne.TextStyle{Bold: true}
	app.cimPredLabel = widget.NewLabel("Prediction: -")
	app.cimConfBar = widget.NewProgressBar()

	// Agreement
	app.agreementLabel = widget.NewLabel("")

	// Side by side
	fpBox := container.NewVBox(fpLabel, app.fpPredLabel, app.fpConfBar)
	cimBox := container.NewVBox(cimLabel, app.cimPredLabel, app.cimConfBar)
	comparison := container.NewGridWithColumns(2, fpBox, cimBox)

	// Probability distribution bars
	probLabel := widget.NewLabel("Class Probabilities (FP top, CIM bottom)")
	probLabel.TextStyle = fyne.TextStyle{Monospace: true}

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

	return container.NewVBox(
		label,
		comparison,
		app.agreementLabel,
		widget.NewSeparator(),
		probLabel,
		probGrid,
		widget.NewSeparator(),
		app.energyLabel,
	)
}

// createControlsZone creates the hardware control panel (Zone 3).
func (app *DualModeApp) createControlsZone() fyne.CanvasObject {
	label := widget.NewLabel("Hardware Configuration")
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	// Levels slider
	levelsTitle := widget.NewLabel("Weight Levels (FeCIM = 30):")
	app.levelsLabel = widget.NewLabel("30")
	app.levelsSlider = widget.NewSlider(1, 30)
	app.levelsSlider.Step = 1
	app.levelsSlider.Value = 30
	app.levelsSlider.OnChanged = func(v float64) {
		app.levelsLabel.SetText(fmt.Sprintf("%d", int(v)))
		app.network.SetNumLevels(int(v))
		app.updateWeightHeatmap()
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	levelsRow := container.NewBorder(nil, nil, levelsTitle, app.levelsLabel, app.levelsSlider)

	// Noise slider
	noiseTitle := widget.NewLabel("Noise Level (sigma/mu):")
	app.noiseLabel = widget.NewLabel("0.01")
	app.noiseSlider = widget.NewSlider(0.0, 0.20)
	app.noiseSlider.Step = 0.01
	app.noiseSlider.Value = 0.01
	app.noiseSlider.OnChanged = func(v float64) {
		app.noiseLabel.SetText(fmt.Sprintf("%.2f", v))
		app.network.SetNoiseLevel(v)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	noiseRow := container.NewBorder(nil, nil, noiseTitle, app.noiseLabel, app.noiseSlider)

	// ADC/DAC selects
	adcLabel := widget.NewLabel("ADC Bits:")
	dacLabel := widget.NewLabel("DAC Bits:")

	bitOptions := []string{"3", "4", "5", "6", "7", "8", "10", "12", "16"}
	app.adcSelect = widget.NewSelect(bitOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		app.network.SetADCBits(bits)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
	app.adcSelect.SetSelected("8")

	app.dacSelect = widget.NewSelect(bitOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		app.network.SetDACBits(bits)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
	app.dacSelect.SetSelected("8")

	adcDacRow := container.NewGridWithColumns(4, adcLabel, app.adcSelect, dacLabel, app.dacSelect)

	// Hidden size selector
	hiddenLabel := widget.NewLabel("Hidden Size:")
	app.hiddenSelect = widget.NewSelect([]string{"64", "128", "256"}, func(s string) {
		var size int
		fmt.Sscanf(s, "%d", &size)
		app.changeHiddenSize(size)
	})
	app.hiddenSelect.SetSelected("128")

	// Preset buttons
	presetLabel := widget.NewLabel("Failure Mode Presets:")
	idealBtn := widget.NewButton("Ideal", func() { app.applyPreset(30, 0.01, 8, 8) })
	quantCliffBtn := widget.NewButton("Quant Cliff", func() { app.applyPreset(2, 0.01, 8, 8) })
	noisyBtn := widget.NewButton("Noisy", func() { app.applyPreset(30, 0.15, 6, 8) })
	brokenBtn := widget.NewButton("Broken ADC", func() { app.applyPreset(30, 0.01, 3, 8) })

	presetRow := container.NewGridWithColumns(4, idealBtn, quantCliffBtn, noisyBtn, brokenBtn)

	// Quick test
	app.testResultLabel = widget.NewLabel("")
	app.testButton = widget.NewButton("Run Quick Test (200 samples)", func() {
		app.runQuickTest()
	})

	hiddenRow := container.NewHBox(hiddenLabel, app.hiddenSelect)

	return container.NewVBox(
		label,
		widget.NewSeparator(),
		levelsRow,
		noiseRow,
		adcDacRow,
		hiddenRow,
		widget.NewSeparator(),
		presetLabel,
		presetRow,
		widget.NewSeparator(),
		app.testButton,
		app.testResultLabel,
	)
}

// createWeightZone creates the weight visualization zone (Zone 4).
func (app *DualModeApp) createWeightZone() fyne.CanvasObject {
	label := widget.NewLabel("Crossbar Weight Map")
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	// Layer selector
	layerSelect := widget.NewRadioGroup(
		[]string{"Input -> Hidden (784x128)", "Hidden -> Output (128x10)"},
		func(s string) {
			if s == "Input -> Hidden (784x128)" {
				app.weightLayer = 0
			} else {
				app.weightLayer = 1
			}
			app.updateWeightHeatmap()
		},
	)
	layerSelect.SetSelected("Input -> Hidden (784x128)")

	// Info labels
	app.weightDimLabel = widget.NewLabel("")
	app.weightRangeLabel = widget.NewLabel("")
	app.weightLevelsLabel = widget.NewLabel("")

	// Create heatmap raster
	app.weightHeatmap = canvas.NewRaster(app.drawWeightHeatmap)
	app.weightHeatmap.SetMinSize(fyne.NewSize(256, 128))

	// Color legend
	legendLabel := widget.NewLabel("Blue = negative, White = zero, Red = positive")
	legendLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewVBox(
		label,
		layerSelect,
		app.weightDimLabel,
		app.weightRangeLabel,
		app.weightLevelsLabel,
		container.NewCenter(app.weightHeatmap),
		legendLabel,
	)
}

// onDigitChanged handles canvas drawing updates.
func (app *DualModeApp) onDigitChanged(pixels []float64) {
	app.lastPixels = pixels
	app.runInference(pixels)
}

// runInference runs dual-path inference and updates the UI.
func (app *DualModeApp) runInference(pixels []float64) {
	result := app.network.Infer(pixels)

	fyne.Do(func() {
		// Update FP results
		app.fpPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.FPPrediction, result.FPConfidence*100))
		app.fpConfBar.SetValue(result.FPConfidence)

		// Update CIM results
		app.cimPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.CIMPrediction, result.CIMConfidence*100))
		app.cimConfBar.SetValue(result.CIMConfidence)

		// Update agreement
		if result.Agree {
			app.agreementLabel.SetText("PREDICTIONS MATCH")
		} else {
			app.agreementLabel.SetText(fmt.Sprintf("DISAGREEMENT (KL=%.3f)", result.Disagreement))
		}

		// Update probability bars
		for i := 0; i < 10; i++ {
			app.fpProbBars[i].SetValue(result.FPProbabilities[i])
			app.cimProbBars[i].SetValue(result.CIMProbabilities[i])
		}

		// Update energy
		gpuEnergy := result.EnergyUsed * 10000 // Estimated 10,000x for GPU
		app.energyLabel.SetText(fmt.Sprintf("Energy: %.2f uJ (FeCIM) vs %.0f mJ (GPU) = %.0fx savings",
			result.EnergyUsed, gpuEnergy/1000, 10000.0))

		// Update status
		app.statusLabel.SetText(fmt.Sprintf("FP: %d (%.1f%%) | CIM: %d (%.1f%%) | %s",
			result.FPPrediction, result.FPConfidence*100,
			result.CIMPrediction, result.CIMConfidence*100,
			map[bool]string{true: "MATCH", false: "MISMATCH"}[result.Agree]))
	})
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
}

// applyPreset sets hardware parameters to a failure mode preset.
func (app *DualModeApp) applyPreset(levels int, noise float64, adcBits, dacBits int) {
	app.levelsSlider.SetValue(float64(levels))
	app.noiseSlider.SetValue(noise)
	app.adcSelect.SetSelected(fmt.Sprintf("%d", adcBits))
	app.dacSelect.SetSelected(fmt.Sprintf("%d", dacBits))

	app.network.SetNumLevels(levels)
	app.network.SetNoiseLevel(noise)
	app.network.SetADCBits(adcBits)
	app.network.SetDACBits(dacBits)

	app.updateWeightHeatmap()

	if len(app.lastPixels) > 0 {
		app.runInference(app.lastPixels)
	}
}

// loadRandomSample loads a random test sample.
func (app *DualModeApp) loadRandomSample() {
	if len(app.testImages) == 0 {
		app.loadTestData()
		if len(app.testImages) == 0 {
			app.statusLabel.SetText("No test data available")
			return
		}
	}

	idx := int(time.Now().UnixNano() % int64(len(app.testImages)))
	pixels := app.testImages[idx]
	label := app.testLabels[idx]

	app.digitCanvas.SetPixels(pixels)
	app.statusLabel.SetText(fmt.Sprintf("Loaded sample #%d (true label: %d)", idx, label))
	app.onDigitChanged(pixels)
}

// loadTestData loads MNIST test data.
func (app *DualModeApp) loadTestData() {
	testImagesPath := filepath.Join(app.dataDir, "t10k-images-idx3-ubyte")
	testLabelsPath := filepath.Join(app.dataDir, "t10k-labels-idx1-ubyte")

	images, labels, err := loadMNISTData(testImagesPath, testLabelsPath)
	if err != nil {
		app.testImages, app.testLabels = generateSyntheticData(200)
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
	app.testButton.Disable()
	app.testResultLabel.SetText("Testing...")

	go func() {
		if len(app.testImages) == 0 {
			app.loadTestData()
		}

		n := 200
		if n > len(app.testImages) {
			n = len(app.testImages)
		}

		fpCorrect := 0
		cimCorrect := 0
		agreements := 0

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
		}

		fpAcc := float64(fpCorrect) / float64(n) * 100
		cimAcc := float64(cimCorrect) / float64(n) * 100
		agreeRate := float64(agreements) / float64(n) * 100

		fyne.Do(func() {
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
		app.weightHeatmap.Refresh()
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
		app.statusLabel.SetText(fmt.Sprintf("Error loading weights: %v", err))
		return
	}

	// Reset results and update heatmap
	app.resetResults()
	app.updateWeightHeatmap()
	app.statusLabel.SetText(fmt.Sprintf("Loaded network with hidden size %d", size))
}
