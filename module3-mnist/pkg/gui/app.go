// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/module3-mnist/pkg/training"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/themes"
	"fecim-lattice-tools/shared/utils"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

var debug *logging.Logger

func init() {
	// Use shared logging - writes to single fecim.log file
	debug = logging.NewLogger("mnist-gui")
}

// FeCIM theme colors - use shared themes package
var (
	colorBackground = themes.DarkBackground // FeCIM blue #003264
	colorPrimary    = themes.DarkPrimary    // Cyan
)

// MNISTApp is the main application for the MNIST demo.
type MNISTApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Neural network
	network *training.MNISTNetwork

	// GUI components
	digitCanvas     *DigitCanvas
	layerView       *LayerActivationView
	confusionMatrix *ConfusionMatrix
	metricsPanel    *MetricsPanel
	classStatsPanel *ClassStatsPanel
	outputChart     *OutputBarChart

	// Live Slide components
	modeIndicator     *MNISTModeIndicator
	educationalPanel  *MNISTEducationalPanel
	operationLog      *MNISTOperationLog
	predictionDisplay *PredictionDisplay
	keyStat           *MNISTKeyStat
	demoModeSelect    *widget.Select

	// Labels
	statusLabel     *widget.Label
	statusBar       *sharedwidgets.StatusBar
	predictionLabel *widget.Label
	confidenceLabel *widget.Label
	hoverInfoLabel  *widget.Label
	infoLabel       *widget.Label

	// Test data for confusion matrix
	testImages [][]float64
	testLabels []int

	// Data directory
	dataDir string

	// Auto demo state (protected by autoDemoMu)
	autoDemoMu     sync.Mutex
	autoDemo       bool
	autoDemoTimer  *time.Ticker
	autoDemoCtx    context.Context
	autoDemoCancel context.CancelFunc
}

// NewMNISTApp creates and initializes the MNIST demo application.
func NewMNISTApp() *MNISTApp {
	debug.Println("NewMNISTApp: Creating application")
	ma := &MNISTApp{}

	// Create Fyne app with theme manager
	ma.fyneApp = app.NewWithID("com.fecim.mnist-demo")
	themeManager := themes.NewManager(ma.fyneApp)
	themeManager.LoadPreference()
	debug.Println("NewMNISTApp: Applied FeCIM theme")

	// Find data directory
	ma.dataDir = utils.FindModuleDataDir("module3-mnist", "pretrained_weights.json")
	if ma.dataDir == "" {
		ma.dataDir = "module3-mnist/data" // Default fallback
	}

	// Create crossbar arrays for layers
	// Layer 1: hidden x 784 (transposed for MVM)
	layer1Config := &crossbar.Config{
		Rows:       128, // hidden size
		Cols:       784, // input size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer1, err := crossbar.NewArray(layer1Config)
	if err != nil {
		debug.Printf("NewMNISTApp: Failed to create layer 1 crossbar: %v", err)
		return nil
	}

	// Layer 2: 10 x hidden
	layer2Config := &crossbar.Config{
		Rows:       10,  // output size
		Cols:       128, // hidden size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer2, err := crossbar.NewArray(layer2Config)
	if err != nil {
		debug.Printf("NewMNISTApp: Failed to create layer 2 crossbar: %v", err)
		return nil
	}

	// Create network
	ma.network = training.NewMNISTNetwork(layer1, layer2)

	// Try to load pretrained weights
	weightsPath := filepath.Join(ma.dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		if err := ma.network.LoadWeights(weightsPath); err == nil {
			debug.Printf("NewMNISTApp: Loaded pretrained weights from %s", weightsPath)
		} else {
			debug.Printf("NewMNISTApp: Failed to load weights: %v", err)
		}
	} else {
		debug.Printf("NewMNISTApp: No pretrained weights found at %s", weightsPath)
	}

	debug.Println("NewMNISTApp: Initialization complete")
	return ma
}

// Run starts the GUI application.
func (ma *MNISTApp) Run() {
	debug.Println("App: Creating window")
	ma.window = ma.fyneApp.NewWindow("FeCIM Demo 3: MNIST Neural Network")
	ma.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	debug.Println("App: Creating main layout")
	content := ma.createMainLayout()
	debug.Println("App: Setting window content")
	ma.window.SetContent(content)

	// Setup keyboard shortcuts
	ma.setupKeyboard()

	// Initialize
	debug.Println("App: Updating status")
	ma.updateStatus("Ready. Draw a digit or load test data. Press ? for shortcuts.")

	debug.Println("App: ShowAndRun starting")
	ma.window.ShowAndRun()
	debug.Println("App: Window closed")
}

// createMainLayout builds the main application layout.
func (ma *MNISTApp) createMainLayout() fyne.CanvasObject {
	// Create components
	ma.digitCanvas = NewDigitCanvas()
	ma.digitCanvas.OnDigitChanged = ma.onDigitChanged

	ma.layerView = NewLayerActivationView()
	ma.outputChart = NewOutputBarChart()

	ma.confusionMatrix = NewConfusionMatrix()
	ma.confusionMatrix.OnCellTapped = ma.onConfusionCellTapped

	ma.metricsPanel = NewMetricsPanel()
	ma.classStatsPanel = NewClassStatsPanel()

	// Create Live Slide components
	ma.modeIndicator = NewMNISTModeIndicator()
	ma.educationalPanel = NewMNISTEducationalPanel()
	ma.educationalPanel.SetIdleExplanation()
	ma.operationLog = NewMNISTOperationLog()
	ma.predictionDisplay = NewPredictionDisplay()
	ma.keyStat = NewMNISTKeyStat("Literature", "96-98% (lit.)")

	// Demo mode selector
	ma.demoModeSelect = widget.NewSelect(
		[]string{"Manual", "Auto Demo", "Step-by-Step"},
		ma.onDemoModeChanged,
	)
	ma.demoModeSelect.SetSelected("Manual")

	// Status labels
	ma.statusLabel = widget.NewLabel("● IDLE | Ready to draw or load test data")
	ma.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	ma.statusBar = sharedwidgets.NewStatusBarWithLabel(ma.statusLabel, "Status: ")

	ma.predictionLabel = widget.NewLabel("Prediction: -")
	ma.predictionLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	ma.confidenceLabel = widget.NewLabel("Confidence: -")

	// Hover info label - shows activation info on mouse hover
	ma.hoverInfoLabel = widget.NewLabel("Hover over neurons to see values")
	ma.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Info label with network specs - distinguish verified vs simulated
	ma.infoLabel = widget.NewLabel("Network: 784→128→10 | Levels: 30 (claim) | Literature: 96-98% (non-FeCIM) | Demo: projected")

	// Control buttons - organized into groups
	clearBtn := widget.NewButton("Clear", func() {
		ma.digitCanvas.Clear()
		ma.predictionDisplay.SetPrediction(-1, 0)
		ma.operationLog.Add("Canvas cleared")
		ma.updateStatus("● IDLE | Canvas cleared")
	})

	randomBtn := widget.NewButton("Random", func() {
		ma.loadRandomTestDigit()
	})

	var evalBtn *widget.Button
	var loadTestBtn *widget.Button

	evalBtn = widget.NewButton("Evaluate", func() {
		evalBtn.Disable()
		loadTestBtn.Disable()
		evalBtn.SetText("Evaluating...")
		go func() {
			ma.evaluateNetwork()
			fyne.Do(func() {
				evalBtn.Enable()
				loadTestBtn.Enable()
				evalBtn.SetText("Evaluate")
			})
		}()
	})
	evalBtn.Importance = widget.HighImportance

	loadTestBtn = widget.NewButton("Load Data", func() {
		loadTestBtn.Disable()
		evalBtn.Disable()
		loadTestBtn.SetText("Loading...")
		go func() {
			ma.loadTestData()
			fyne.Do(func() {
				loadTestBtn.Enable()
				evalBtn.Enable()
				loadTestBtn.SetText("Load Data")
			})
		}()
	})

	// Buttons in 2x2 grid for compactness
	buttonGrid := container.NewGridWithColumns(2,
		clearBtn,
		randomBtn,
		loadTestBtn,
		evalBtn,
	)

	// View selector buttons (replaces tabs to save vertical space)
	var contentContainer *fyne.Container
	drawBtn := widget.NewButton("Draw & Predict", func() {})
	metricsBtn := widget.NewButton("Evaluation Metrics", func() {})

	// Title and header with inline view selector
	titleLabel := widget.NewLabel("FeCIM MNIST Neural Network")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	viewSelectorRow := container.NewHBox(
		widget.NewLabel("View:"),
		drawBtn,
		metricsBtn,
		layout.NewSpacer(),
		widget.NewLabel("784 -> 128 -> 10 | 30 Levels (claim) | Accuracy varies with config"),
	)

	header := container.NewVBox(
		viewSelectorRow,
		widget.NewSeparator(),
	)

	// Left panel: Drawing canvas + prediction display
	canvasLabel := widget.NewLabel("Draw a Digit (0-9)")
	canvasLabel.TextStyle = fyne.TextStyle{Bold: true}
	canvasLabel.Alignment = fyne.TextAlignCenter

	predictionBox := container.NewVBox(
		widget.NewSeparator(),
		ma.predictionDisplay,
		ma.predictionLabel,
		ma.confidenceLabel,
	)

	// Left panel with scroll for proper sizing
	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Mode", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			ma.demoModeSelect,
			widget.NewSeparator(),
			canvasLabel,
		),
		container.NewVBox(
			predictionBox,
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Actions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			buttonGrid,
		),
		nil, nil,
		container.NewCenter(ma.digitCanvas),
	)

	// Center panel: Layer activations with proper expansion
	activationLabel := widget.NewLabel("Network Activations")
	activationLabel.TextStyle = fyne.TextStyle{Bold: true}
	activationLabel.Alignment = fyne.TextAlignCenter

	// Use VSplit for center panel to allow proper resizing
	centerSplit := container.NewVSplit(
		container.NewMax(ma.layerView),
		container.NewMax(ma.outputChart),
	)
	centerSplit.SetOffset(0.6) // 60% layer view, 40% output chart

	centerPanel := container.NewBorder(
		container.NewVBox(activationLabel, widget.NewSeparator()),
		nil, nil, nil,
		centerSplit,
	)

	// Right panel (educational): Educational + Log + Key Stat
	educationalRight := container.NewBorder(
		ma.educationalPanel,
		ma.keyStat,
		nil, nil,
		container.NewVBox(
			widget.NewSeparator(),
			ma.operationLog,
			widget.NewSeparator(),
		),
	)

	// Metrics panel for evaluation tab - use VSplit for resizing
	metricsSplit := container.NewVSplit(
		container.NewMax(ma.confusionMatrix),
		container.NewVBox(
			widget.NewSeparator(),
			ma.metricsPanel,
			widget.NewSeparator(),
			ma.classStatsPanel,
		),
	)
	metricsSplit.SetOffset(0.5)

	// Use HSplit for proportional 3-column layout (like demo2)
	// Left panel (20%) | Center panel (55%) | Right panel (25%)
	leftCenterSplit := container.NewHSplit(
		leftPanel,
		centerPanel,
	)
	leftCenterSplit.SetOffset(0.22) // 22% left, 78% center+right

	mainSplit := container.NewHSplit(
		leftCenterSplit,
		educationalRight,
	)
	mainSplit.SetOffset(0.78) // 78% left+center, 22% right

	// Footer with mode indicator and status (like demo2)
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			ma.modeIndicator,
			widget.NewSeparator(),
			ma.statusLabel,
			layout.NewSpacer(),
			ma.hoverInfoLabel,
			widget.NewSeparator(),
			ma.infoLabel,
		),
	)

	// Content views - pre-created to avoid layout cascades on Wayland/Sway
	drawView := container.NewMax(mainSplit)
	metricsView := container.NewMax(metricsSplit)

	// Content container using Stack - all views layered, visibility toggled
	contentContainer = container.NewStack(drawView, metricsView)

	// Track current view to avoid redundant updates
	currentView := -1

	// Setup view switching logic using Hide/Show (avoids layout cascades)
	updateView := func(view int) {
		if view == currentView {
			return // No change needed
		}
		currentView = view

		if view == 0 {
			drawBtn.Importance = widget.HighImportance
			metricsBtn.Importance = widget.MediumImportance
			metricsView.Hide()
			drawView.Show()
		} else {
			drawBtn.Importance = widget.MediumImportance
			metricsBtn.Importance = widget.HighImportance
			drawView.Hide()
			metricsView.Show()
		}
	}

	drawBtn.OnTapped = func() { updateView(0) }
	metricsBtn.OnTapped = func() { updateView(1) }

	// Set initial view
	updateView(0)

	// Main content
	mainContent := container.NewBorder(
		header,           // top
		footer,           // bottom
		nil,              // left
		nil,              // right
		contentContainer, // center
	)

	return mainContent
}

// onDigitChanged handles canvas drawing updates with animated phases (like demo2 MVM).
func (ma *MNISTApp) onDigitChanged(pixels []float64) {
	debug.Println("onDigitChanged: Starting inference")

	pixels = preprocessIfUserInput(ma.digitCanvas, pixels)

	// Run animated inference in goroutine for smooth UI
	go ma.runInferenceAnimated(pixels)
}

// runInferenceAnimated performs inference with visual animation phases (like demo2 MVM).
func (ma *MNISTApp) runInferenceAnimated(pixels []float64) {
	// Phase 1: Input processing (200ms)
	fyne.Do(func() {
		ma.modeIndicator.SetMode(MNISTModeInference)
		ma.updateStatus("INFERENCE | Phase 1: Processing input pixels...")
		ma.educationalPanel.SetInferenceExplanation(1)
		ma.hoverInfoLabel.SetText("Phase 1: Input → 784 neurons")
	})
	time.Sleep(200 * time.Millisecond)

	// Phase 2: Hidden layer computation (300ms)
	fyne.Do(func() {
		ma.updateStatus("INFERENCE | Phase 2: Hidden layer MVM (784→128)...")
		ma.educationalPanel.SetInferenceExplanation(2)
		ma.hoverInfoLabel.SetText("Phase 2: MVM 100,352 MACs")
	})

	// Get layer activations
	input, hidden, probs := ma.network.GetLayerActivations(pixels)
	debug.Printf("onDigitChanged: Got activations - input:%d hidden:%d probs:%d",
		len(input), len(hidden), len(probs))
	time.Sleep(300 * time.Millisecond)

	// Phase 3: Output layer computation (200ms)
	fyne.Do(func() {
		ma.updateStatus("INFERENCE | Phase 3: Output layer MVM (128→10)...")
		ma.layerView.SetActivations(input, hidden, probs)
		ma.hoverInfoLabel.SetText("Phase 3: MVM 1,280 MACs")
	})
	time.Sleep(200 * time.Millisecond)

	// Phase 4: Final prediction (display results)
	fyne.Do(func() {
		ma.educationalPanel.SetInferenceExplanation(3)
		ma.outputChart.SetValues(probs)

		// Get final prediction
		pred, conf := ma.network.Predict(pixels)
		debug.Printf("onDigitChanged: Prediction=%d Confidence=%.2f%%", pred, conf*100)

		ma.predictionLabel.SetText(fmt.Sprintf("Prediction: %d", pred))
		ma.confidenceLabel.SetText(fmt.Sprintf("Confidence: %.1f%%", conf*100))
		ma.predictionDisplay.SetPrediction(pred, conf)

		// Log and update status
		ma.operationLog.AddPrediction(pred, conf)
		totalMACs := 784*128 + 128*10
		ma.updateStatus(fmt.Sprintf("INFERENCE | Prediction: %d (%.1f%% confidence) | %d MACs", pred, conf*100, totalMACs))
		ma.hoverInfoLabel.SetText(fmt.Sprintf("Complete: %d parallel MACs", totalMACs))
		ma.modeIndicator.SetMode(MNISTModeIdle)
	})
	debug.Println("onDigitChanged: Complete")
}

// loadRandomTestDigit loads a random digit from test data.
func (ma *MNISTApp) loadRandomTestDigit() {
	debug.Println("loadRandomTestDigit: Starting")
	ma.modeIndicator.SetMode(MNISTModeLoading)
	ma.operationLog.Add("Loading random test digit")

	if len(ma.testImages) == 0 {
		debug.Println("loadRandomTestDigit: No test data, loading...")
		ma.loadTestData()
		if len(ma.testImages) == 0 {
			debug.Println("loadRandomTestDigit: Failed to load test data")
			ma.updateStatus("● IDLE | No test data available")
			ma.modeIndicator.SetMode(MNISTModeIdle)
			return
		}
	}

	idx := rand.Intn(len(ma.testImages))
	pixels := ma.testImages[idx]
	label := ma.testLabels[idx]
	debug.Printf("loadRandomTestDigit: Selected digit #%d (label: %d)", idx, label)

	ma.digitCanvas.SetPixels(pixels)
	ma.onDigitChanged(pixels)

	ma.operationLog.Add(fmt.Sprintf("Loaded digit #%d (label: %d)", idx, label))
	ma.updateStatus(fmt.Sprintf("LOADING | Loaded test digit #%d (true label: %d)", idx, label))
}

// loadTestData loads MNIST test data.
func (ma *MNISTApp) loadTestData() {
	debug.Println("loadTestData: Starting")
	ma.modeIndicator.SetMode(MNISTModeLoading)
	ma.updateStatus("LOADING | Loading test data...")
	ma.operationLog.Add("Loading MNIST test data")

	debug.Printf("loadTestData: Trying to load MNIST data from %s", ma.dataDir)
	images, labels, err := mnist.LoadMNIST(ma.dataDir, false)
	if err != nil {
		// Fall back to generating synthetic data for demo
		debug.Printf("loadTestData: MNIST not found (%v), using synthetic data", err)
		ma.operationLog.Add("Using synthetic data (MNIST not found)")
		ma.updateStatus("LOADING | Using synthetic test data")
		ma.testImages, ma.testLabels = generateSyntheticData(100)
		ma.modeIndicator.SetMode(MNISTModeIdle)
		return
	}

	// Limit to 1000 samples for speed
	if len(images) > 1000 {
		ma.testImages = images[:1000]
		ma.testLabels = labels[:1000]
	} else {
		ma.testImages = images
		ma.testLabels = labels
	}

	debug.Printf("loadTestData: Loaded %d samples", len(ma.testImages))
	ma.operationLog.Add(fmt.Sprintf("Loaded %d samples", len(ma.testImages)))
	ma.updateStatus(fmt.Sprintf("● IDLE | Loaded %d test samples", len(ma.testImages)))
	ma.modeIndicator.SetMode(MNISTModeIdle)
}

// evaluateNetwork runs evaluation on test data.
func (ma *MNISTApp) evaluateNetwork() {
	debug.Println("evaluateNetwork: Starting")
	ma.modeIndicator.SetMode(MNISTModeEvaluating)
	ma.educationalPanel.SetEvaluationExplanation()
	ma.operationLog.Add("Starting full evaluation")

	if len(ma.testImages) == 0 {
		debug.Println("evaluateNetwork: No test data, loading...")
		ma.loadTestData()
		if len(ma.testImages) == 0 {
			debug.Println("evaluateNetwork: Failed to load test data")
			ma.updateStatus("● IDLE | No test data for evaluation")
			ma.modeIndicator.SetMode(MNISTModeIdle)
			return
		}
	}

	debug.Printf("evaluateNetwork: Testing on %d samples", len(ma.testImages))
	ma.updateStatus(fmt.Sprintf("EVALUATING | Testing on %d samples...", len(ma.testImages)))

	// Compute confusion matrix
	confMatrix := ma.network.ComputeConfusionMatrix(ma.testImages, ma.testLabels)
	debug.Println("evaluateNetwork: Confusion matrix computed")

	// Convert to [10][10]int
	var matrix [10][10]int
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			matrix[i][j] = confMatrix[i][j]
		}
	}
	ma.confusionMatrix.SetMatrix(matrix)

	// Compute metrics
	precision, recall, f1 := ma.network.GetPerClassMetrics(confMatrix)

	var precArr, recArr, f1Arr [10]float64
	for i := 0; i < 10; i++ {
		precArr[i] = precision[i]
		recArr[i] = recall[i]
		f1Arr[i] = f1[i]
	}

	accuracy := ma.confusionMatrix.GetAccuracy()
	debug.Printf("evaluateNetwork: Accuracy=%.2f%%", accuracy*100)
	ma.metricsPanel.SetMetrics(precArr, recArr, f1Arr, accuracy)

	// Update key stat
	ma.keyStat.SetValue(fmt.Sprintf("%.1f%%", accuracy*100))

	// Update hover info with evaluation summary
	totalMACs := len(ma.testImages) * (784*128 + 128*10)
	fyne.Do(func() {
		ma.hoverInfoLabel.SetText(fmt.Sprintf("Evaluated: %d MACs total", totalMACs))
	})

	// Log result
	ma.operationLog.Add(fmt.Sprintf("Accuracy: %.1f%% on %d samples", accuracy*100, len(ma.testImages)))
	ma.updateStatus(fmt.Sprintf("● IDLE | Evaluation complete: %.1f%% accuracy", accuracy*100))
	ma.modeIndicator.SetMode(MNISTModeIdle)
	debug.Println("evaluateNetwork: Complete")
}

// onConfusionCellTapped handles clicks on the confusion matrix.
func (ma *MNISTApp) onConfusionCellTapped(actual, predicted, count int) {
	precision, recall, f1 := ma.confusionMatrix.GetClassMetrics(actual)

	// Estimate TP based on click location
	tp := 0
	fp := 0
	fn := 0
	if actual == predicted {
		tp = count
	} else {
		fp = count // Misclassification
	}

	ma.classStatsPanel.SetClass(actual, precision, recall, f1, tp, fp, fn, count)

	ma.updateStatus(fmt.Sprintf("Cell [%d,%d]: %d samples (actual=%d, predicted=%d)",
		actual, predicted, count, actual, predicted))
}

// updateStatus updates the status label.
func (ma *MNISTApp) updateStatus(status string) {
	if ma.statusBar == nil {
		if ma.statusLabel == nil {
			return
		}
		ma.statusBar = sharedwidgets.NewStatusBarWithLabel(ma.statusLabel, "Status: ")
	}
	ma.statusBar.Update(status)
}

// generateSyntheticData creates simple synthetic digit patterns for demo.
func generateSyntheticData(count int) ([][]float64, []int) {
	images := make([][]float64, count)
	labels := make([]int, count)

	for i := 0; i < count; i++ {
		digit := rand.Intn(10)
		labels[i] = digit
		images[i] = make([]float64, 784)

		// Generate simple digit-like pattern
		switch digit {
		case 0: // Circle
			for y := 8; y < 20; y++ {
				for x := 8; x < 20; x++ {
					dx := float64(x) - 14
					dy := float64(y) - 14
					dist := dx*dx + dy*dy
					if dist > 25 && dist < 49 {
						images[i][y*28+x] = 0.8 + rand.Float64()*0.2
					}
				}
			}
		case 1: // Vertical line
			for y := 6; y < 22; y++ {
				images[i][y*28+14] = 0.8 + rand.Float64()*0.2
				images[i][y*28+13] = 0.5 + rand.Float64()*0.3
			}
		case 7: // Diagonal line from top-left
			for j := 0; j < 16; j++ {
				y := 6 + j
				x := 8 + j/2
				if y < 28 && x < 28 {
					images[i][y*28+x] = 0.8 + rand.Float64()*0.2
				}
			}
		default: // Random blob for other digits
			cx := 10 + rand.Intn(8)
			cy := 10 + rand.Intn(8)
			for y := 0; y < 28; y++ {
				for x := 0; x < 28; x++ {
					dx := float64(x - cx)
					dy := float64(y - cy)
					dist := dx*dx + dy*dy
					if dist < 36 {
						images[i][y*28+x] = 0.5 + rand.Float64()*0.5
					}
				}
			}
		}

		// Add noise
		for j := 0; j < 784; j++ {
			if rand.Float64() < 0.02 {
				images[i][j] = rand.Float64() * 0.3
			}
		}
	}

	return images, labels
}

// onDemoModeChanged handles demo mode selection changes.
func (ma *MNISTApp) onDemoModeChanged(mode string) {
	// Stop any existing auto demo
	ma.stopAutoDemoLoop()

	switch mode {
	case "Auto Demo":
		ma.startAutoDemoLoop()
	case "Step-by-Step":
		ma.operationLog.Add("Mode: Step-by-Step (manual)")
		ma.educationalPanel.SetContent("Step-by-Step Mode",
			"Follow along manually:\n\n"+
				"1. Draw a digit OR\n"+
				"   click Random Test\n\n"+
				"2. Watch the network\n"+
				"   layers activate\n\n"+
				"3. See the prediction\n"+
				"   and confidence\n\n"+
				"4. Click Evaluate All\n"+
				"   for full accuracy test")
	case "Manual":
		ma.operationLog.Add("Mode: Manual")
		ma.educationalPanel.SetIdleExplanation()
	}
}

// startAutoDemoLoop starts the automatic demo loop.
func (ma *MNISTApp) startAutoDemoLoop() {
	ma.autoDemoMu.Lock()
	defer ma.autoDemoMu.Unlock()

	ma.autoDemo = true
	ma.autoDemoCtx, ma.autoDemoCancel = context.WithCancel(context.Background())
	ma.autoDemoTimer = time.NewTicker(2 * time.Second)

	ma.operationLog.Add("Mode: Auto Demo started")
	ma.educationalPanel.SetContent("Auto Demo Mode",
		"Watch the demo cycle\nthrough test digits.\n\n"+
			"The network will:\n"+
			"1. Load random digit\n"+
			"2. Run inference\n"+
			"3. Show prediction\n"+
			"4. Repeat\n\n"+
			"Accuracy varies")

	go ma.autoDemoLoop(ma.autoDemoCtx)
}

// stopAutoDemoLoop stops the automatic demo loop.
func (ma *MNISTApp) stopAutoDemoLoop() {
	ma.autoDemoMu.Lock()
	defer ma.autoDemoMu.Unlock()

	if !ma.autoDemo {
		return
	}

	ma.autoDemo = false
	// Cancel context to signal goroutine - goroutine handles its own cleanup
	if ma.autoDemoCancel != nil {
		ma.autoDemoCancel()
		ma.autoDemoCancel = nil
	}
	// Note: ticker is stopped by the goroutine via defer, not here
	// This prevents race conditions between stopping and the goroutine reading
	ma.operationLog.Add("Mode: Auto Demo stopped")
}

// autoDemoLoop runs the automatic demonstration.
// The goroutine owns the ticker and cleans it up on exit to prevent memory leaks.
func (ma *MNISTApp) autoDemoLoop(ctx context.Context) {
	// Goroutine owns the ticker cleanup to prevent race with stopAutoDemoLoop
	ma.autoDemoMu.Lock()
	ticker := ma.autoDemoTimer
	ma.autoDemoMu.Unlock()

	if ticker == nil {
		return
	}
	defer func() {
		ticker.Stop()
		ma.autoDemoMu.Lock()
		ma.autoDemoTimer = nil
		ma.autoDemoMu.Unlock()
	}()

	// Run first operation immediately
	fyne.Do(func() {
		ma.loadRandomTestDigit()
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ma.autoDemoMu.Lock()
			running := ma.autoDemo
			ma.autoDemoMu.Unlock()

			if !running {
				return
			}
			fyne.Do(func() {
				ma.loadRandomTestDigit()
			})
		}
	}
}
