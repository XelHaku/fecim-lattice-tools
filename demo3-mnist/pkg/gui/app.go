// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/crossbar"
	"multilayer-ferroelectric-cim-visualizer/demo3-mnist/pkg/training"
)

var debug *log.Logger
var logFile *os.File

func init() {
	// Create logs directory
	logsDir := "<local-path>"
	os.MkdirAll(logsDir, 0755)

	// Create log file with datetime
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-mnist-demo03.log")

	var err error
	logFile, err = os.Create(logPath)
	if err != nil {
		// Fallback to stdout if file creation fails
		debug = log.New(os.Stdout, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
		debug.Printf("Failed to create log file: %v, using stdout", err)
		return
	}

	// Write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	debug = log.New(multiWriter, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
	debug.Printf("Logging to: %s", logPath)
}

// FeCIM theme colors - same as demo1
var (
	colorBackground = color.RGBA{0, 50, 100, 255}  // FeCIM blue #003264
	colorPrimary    = color.RGBA{0, 212, 255, 255} // Cyan
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground // FeCIM blue #003264
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255} // Slightly lighter blue
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255} // Darker blue for inputs
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255} // Separator lines
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

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
	modeIndicator      *MNISTModeIndicator
	educationalPanel   *MNISTEducationalPanel
	operationLog       *MNISTOperationLog
	predictionDisplay  *PredictionDisplay
	keyStat            *MNISTKeyStat
	demoModeSelect     *widget.Select

	// Labels
	statusLabel     *widget.Label
	predictionLabel *widget.Label
	confidenceLabel *widget.Label
	hoverInfoLabel  *widget.Label
	infoLabel       *widget.Label

	// Test data for confusion matrix
	testImages [][]float64
	testLabels []int

	// Data directory
	dataDir string

	// Auto demo state
	autoDemo      bool
	autoDemoTimer *time.Ticker
	stopAutoDemo  chan bool
}

// NewMNISTApp creates and initializes the MNIST demo application.
func NewMNISTApp() *MNISTApp {
	debug.Println("NewMNISTApp: Creating application")
	ma := &MNISTApp{}

	// Create Fyne app
	ma.fyneApp = app.NewWithID("com.fecim.mnist-demo")
	ma.fyneApp.Settings().SetTheme(&feCIMTheme{})
	debug.Println("NewMNISTApp: Applied FeCIM theme")

	// Find data directory
	ma.dataDir = findDataDir()

	// Create crossbar arrays for layers
	// Layer 1: hidden x 784 (transposed for MVM)
	layer1Config := &crossbar.Config{
		Rows:       128, // hidden size
		Cols:       784, // input size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer1, _ := crossbar.NewArray(layer1Config)

	// Layer 2: 10 x hidden
	layer2Config := &crossbar.Config{
		Rows:       10,  // output size
		Cols:       128, // hidden size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer2, _ := crossbar.NewArray(layer2Config)

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

// findDataDir locates the demo3-mnist/data directory.
func findDataDir() string {
	// Try common locations
	paths := []string{
		"data",
		"demo3-mnist/data",
		"../data",
		"../../demo3-mnist/data",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "data" // Default
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

	// Initialize
	debug.Println("App: Updating status")
	ma.updateStatus("Ready. Draw a digit or load test data.")

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
	ma.keyStat = NewMNISTKeyStat("Target Accuracy", "87%")

	// Demo mode selector
	ma.demoModeSelect = widget.NewSelect(
		[]string{"Manual", "Auto Demo", "Step-by-Step"},
		ma.onDemoModeChanged,
	)
	ma.demoModeSelect.SetSelected("Manual")

	// Status labels
	ma.statusLabel = widget.NewLabel("● IDLE | Ready to draw or load test data")
	ma.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	ma.predictionLabel = widget.NewLabel("Prediction: -")
	ma.predictionLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	ma.confidenceLabel = widget.NewLabel("Confidence: -")

	// Hover info label - shows activation info on mouse hover
	ma.hoverInfoLabel = widget.NewLabel("Hover over neurons to see values")
	ma.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Info label with network specs
	ma.infoLabel = widget.NewLabel("Network: 784→128→10 | Levels: 30 | Target: 87%")

	// Control buttons
	clearBtn := widget.NewButton("Clear Canvas", func() {
		ma.digitCanvas.Clear()
		ma.predictionDisplay.SetPrediction(-1, 0)
		ma.operationLog.Add("Canvas cleared")
		ma.updateStatus("● IDLE | Canvas cleared")
	})

	randomBtn := widget.NewButton("Random Test", func() {
		ma.loadRandomTestDigit()
	})

	evalBtn := widget.NewButton("Evaluate All", func() {
		ma.evaluateNetwork()
	})
	evalBtn.Importance = widget.HighImportance

	loadTestBtn := widget.NewButton("Load Test Data", func() {
		ma.loadTestData()
	})

	// Buttons container
	buttonBox := container.NewHBox(
		clearBtn,
		randomBtn,
		loadTestBtn,
		evalBtn,
	)

	// Title and header with Dr. Tour quote
	titleLabel := widget.NewLabel("FeCIM MNIST Neural Network")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	quoteLabel := widget.NewLabel("\"We're at 87% validation here\" — Dr. external research group")
	quoteLabel.Alignment = fyne.TextAlignCenter
	quoteLabel.TextStyle = fyne.TextStyle{Italic: true}

	specsLabel := widget.NewLabel("Architecture: 784 -> 128 -> 10 | Target: 87% (88% theoretical max) | 30 Levels")
	specsLabel.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		titleLabel,
		quoteLabel,
		specsLabel,
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
			widget.NewLabel("Demo Mode:"),
			ma.demoModeSelect,
			widget.NewSeparator(),
			canvasLabel,
		),
		container.NewVBox(
			predictionBox,
			widget.NewSeparator(),
			buttonBox,
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

	// Tabs for different views - use Max for full expansion
	drawTab := container.NewTabItem("Draw & Predict", container.NewMax(mainSplit))

	metricsTab := container.NewTabItem("Evaluation Metrics", container.NewMax(metricsSplit))

	tabs := container.NewAppTabs(drawTab, metricsTab)

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

	// Main content
	mainContent := container.NewBorder(
		header, // top
		footer, // bottom
		nil,    // left
		nil,    // right
		tabs,   // center
	)

	return mainContent
}

// onDigitChanged handles canvas drawing updates with animated phases (like demo2 MVM).
func (ma *MNISTApp) onDigitChanged(pixels []float64) {
	debug.Println("onDigitChanged: Starting inference")

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

	// Try to load from IDX files
	testImagesPath := filepath.Join(ma.dataDir, "t10k-images-idx3-ubyte")
	testLabelsPath := filepath.Join(ma.dataDir, "t10k-labels-idx1-ubyte")
	debug.Printf("loadTestData: Trying to load from %s", testImagesPath)

	images, labels, err := loadMNISTData(testImagesPath, testLabelsPath)
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
	ma.hoverInfoLabel.SetText(fmt.Sprintf("Evaluated: %d MACs total", totalMACs))

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
	ma.statusLabel.SetText("Status: " + status)
}

// loadMNISTData loads MNIST data from IDX files.
func loadMNISTData(imagesPath, labelsPath string) ([][]float64, []int, error) {
	// Read images
	imagesData, err := os.ReadFile(imagesPath)
	if err != nil {
		return nil, nil, err
	}

	// Read labels
	labelsData, err := os.ReadFile(labelsPath)
	if err != nil {
		return nil, nil, err
	}

	// Parse images (IDX3 format)
	if len(imagesData) < 16 {
		return nil, nil, fmt.Errorf("images file too small")
	}

	// Skip 16-byte header
	numImages := int(imagesData[4])<<24 | int(imagesData[5])<<16 | int(imagesData[6])<<8 | int(imagesData[7])
	numRows := int(imagesData[8])<<24 | int(imagesData[9])<<16 | int(imagesData[10])<<8 | int(imagesData[11])
	numCols := int(imagesData[12])<<24 | int(imagesData[13])<<16 | int(imagesData[14])<<8 | int(imagesData[15])

	if numRows != 28 || numCols != 28 {
		return nil, nil, fmt.Errorf("unexpected image dimensions: %dx%d", numRows, numCols)
	}

	imageSize := 28 * 28
	images := make([][]float64, numImages)
	for i := 0; i < numImages; i++ {
		images[i] = make([]float64, imageSize)
		offset := 16 + i*imageSize
		for j := 0; j < imageSize && offset+j < len(imagesData); j++ {
			images[i][j] = float64(imagesData[offset+j]) / 255.0
		}
	}

	// Parse labels (IDX1 format)
	if len(labelsData) < 8 {
		return nil, nil, fmt.Errorf("labels file too small")
	}

	numLabels := int(labelsData[4])<<24 | int(labelsData[5])<<16 | int(labelsData[6])<<8 | int(labelsData[7])
	labels := make([]int, numLabels)
	for i := 0; i < numLabels && 8+i < len(labelsData); i++ {
		labels[i] = int(labelsData[8+i])
	}

	return images, labels, nil
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
			for i := 0; i < 16; i++ {
				y := 6 + i
				x := 8 + i/2
				if y < 28 && x < 28 {
					images[labels[0]][y*28+x] = 0.8 + rand.Float64()*0.2
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
	ma.autoDemo = true
	ma.stopAutoDemo = make(chan bool)
	ma.autoDemoTimer = time.NewTicker(2 * time.Second)

	ma.operationLog.Add("Mode: Auto Demo started")
	ma.educationalPanel.SetContent("Auto Demo Mode",
		"Watch the demo cycle\nthrough test digits.\n\n"+
			"The network will:\n"+
			"1. Load random digit\n"+
			"2. Run inference\n"+
			"3. Show prediction\n"+
			"4. Repeat\n\n"+
			"Target: 87% accuracy")

	go ma.autoDemoLoop()
}

// stopAutoDemoLoop stops the automatic demo loop.
func (ma *MNISTApp) stopAutoDemoLoop() {
	if ma.autoDemo {
		ma.autoDemo = false
		if ma.stopAutoDemo != nil {
			close(ma.stopAutoDemo)
		}
		if ma.autoDemoTimer != nil {
			ma.autoDemoTimer.Stop()
		}
		ma.operationLog.Add("Mode: Auto Demo stopped")
	}
}

// autoDemoLoop runs the automatic demonstration.
func (ma *MNISTApp) autoDemoLoop() {
	// Run first operation immediately
	fyne.Do(func() {
		ma.loadRandomTestDigit()
	})

	for {
		select {
		case <-ma.stopAutoDemo:
			return
		case <-ma.autoDemoTimer.C:
			if !ma.autoDemo {
				return
			}
			fyne.Do(func() {
				ma.loadRandomTestDigit()
			})
		}
	}
}
