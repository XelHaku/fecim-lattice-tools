// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import (
	"fmt"
	"image/color"
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
	ma := &MNISTApp{}

	// Create Fyne app
	ma.fyneApp = app.NewWithID("com.fecim.mnist-demo")
	ma.fyneApp.Settings().SetTheme(&feCIMTheme{})

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
			fmt.Println("Loaded pretrained weights from", weightsPath)
		}
	}

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
	ma.window = ma.fyneApp.NewWindow("FeCIM Demo 3: MNIST Neural Network")
	ma.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	content := ma.createMainLayout()
	ma.window.SetContent(content)

	// Initialize
	ma.updateStatus("Ready. Draw a digit or load test data.")

	ma.window.ShowAndRun()
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

	leftPanel := container.NewVBox(
		widget.NewLabel("Demo Mode:"),
		ma.demoModeSelect,
		widget.NewSeparator(),
		canvasLabel,
		container.NewCenter(ma.digitCanvas),
		predictionBox,
		widget.NewSeparator(),
		buttonBox,
	)

	// Center panel: Layer activations
	activationLabel := widget.NewLabel("Network Activations")
	activationLabel.TextStyle = fyne.TextStyle{Bold: true}
	activationLabel.Alignment = fyne.TextAlignCenter

	centerPanel := container.NewVBox(
		activationLabel,
		widget.NewSeparator(),
		ma.layerView,
		widget.NewSeparator(),
		ma.outputChart,
	)

	// Right panel (educational): Educational + Log + Key Stat
	educationalRight := container.NewVBox(
		ma.educationalPanel,
		widget.NewSeparator(),
		ma.operationLog,
		widget.NewSeparator(),
		ma.keyStat,
	)

	// Metrics panel for evaluation tab
	metricsRight := container.NewVBox(
		ma.confusionMatrix,
		widget.NewSeparator(),
		ma.metricsPanel,
		widget.NewSeparator(),
		ma.classStatsPanel,
	)

	// Tabs for different views
	drawTab := container.NewTabItem("Draw & Predict",
		container.NewBorder(
			nil, nil,
			container.NewPadded(leftPanel),
			container.NewPadded(educationalRight),
			container.NewPadded(centerPanel),
		),
	)

	metricsTab := container.NewTabItem("Evaluation Metrics",
		container.NewPadded(metricsRight),
	)

	tabs := container.NewAppTabs(drawTab, metricsTab)

	// Footer with mode indicator and status
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			ma.modeIndicator,
			widget.NewSeparator(),
			ma.statusLabel,
			layout.NewSpacer(),
			widget.NewLabel("FeCIM Ferroelectric CIM | 30 Discrete Levels"),
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

// onDigitChanged handles canvas drawing updates.
func (ma *MNISTApp) onDigitChanged(pixels []float64) {
	ma.modeIndicator.SetMode(MNISTModeInference)
	ma.educationalPanel.SetInferenceExplanation(1)

	// Run inference
	input, hidden, probs := ma.network.GetLayerActivations(pixels)

	ma.educationalPanel.SetInferenceExplanation(2)

	// Update visualization
	ma.layerView.SetActivations(input, hidden, probs)
	ma.outputChart.SetValues(probs)

	ma.educationalPanel.SetInferenceExplanation(3)

	// Update prediction
	pred, conf := ma.network.Predict(pixels)
	ma.predictionLabel.SetText(fmt.Sprintf("Prediction: %d", pred))
	ma.confidenceLabel.SetText(fmt.Sprintf("Confidence: %.1f%%", conf*100))
	ma.predictionDisplay.SetPrediction(pred, conf)

	// Log and update status
	ma.operationLog.AddPrediction(pred, conf)
	ma.updateStatus(fmt.Sprintf("INFERENCE | Prediction: %d (%.1f%% confidence)", pred, conf*100))
	ma.modeIndicator.SetMode(MNISTModeIdle)
}

// loadRandomTestDigit loads a random digit from test data.
func (ma *MNISTApp) loadRandomTestDigit() {
	ma.modeIndicator.SetMode(MNISTModeLoading)
	ma.operationLog.Add("Loading random test digit")

	if len(ma.testImages) == 0 {
		ma.loadTestData()
		if len(ma.testImages) == 0 {
			ma.updateStatus("● IDLE | No test data available")
			ma.modeIndicator.SetMode(MNISTModeIdle)
			return
		}
	}

	idx := rand.Intn(len(ma.testImages))
	pixels := ma.testImages[idx]
	label := ma.testLabels[idx]

	ma.digitCanvas.SetPixels(pixels)
	ma.onDigitChanged(pixels)

	ma.operationLog.Add(fmt.Sprintf("Loaded digit #%d (label: %d)", idx, label))
	ma.updateStatus(fmt.Sprintf("LOADING | Loaded test digit #%d (true label: %d)", idx, label))
}

// loadTestData loads MNIST test data.
func (ma *MNISTApp) loadTestData() {
	ma.modeIndicator.SetMode(MNISTModeLoading)
	ma.updateStatus("LOADING | Loading test data...")
	ma.operationLog.Add("Loading MNIST test data")

	// Try to load from IDX files
	testImagesPath := filepath.Join(ma.dataDir, "t10k-images-idx3-ubyte")
	testLabelsPath := filepath.Join(ma.dataDir, "t10k-labels-idx1-ubyte")

	images, labels, err := loadMNISTData(testImagesPath, testLabelsPath)
	if err != nil {
		// Fall back to generating synthetic data for demo
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

	ma.operationLog.Add(fmt.Sprintf("Loaded %d samples", len(ma.testImages)))
	ma.updateStatus(fmt.Sprintf("● IDLE | Loaded %d test samples", len(ma.testImages)))
	ma.modeIndicator.SetMode(MNISTModeIdle)
}

// evaluateNetwork runs evaluation on test data.
func (ma *MNISTApp) evaluateNetwork() {
	ma.modeIndicator.SetMode(MNISTModeEvaluating)
	ma.educationalPanel.SetEvaluationExplanation()
	ma.operationLog.Add("Starting full evaluation")

	if len(ma.testImages) == 0 {
		ma.loadTestData()
		if len(ma.testImages) == 0 {
			ma.updateStatus("● IDLE | No test data for evaluation")
			ma.modeIndicator.SetMode(MNISTModeIdle)
			return
		}
	}

	ma.updateStatus(fmt.Sprintf("EVALUATING | Testing on %d samples...", len(ma.testImages)))

	// Compute confusion matrix
	confMatrix := ma.network.ComputeConfusionMatrix(ma.testImages, ma.testLabels)

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
	ma.metricsPanel.SetMetrics(precArr, recArr, f1Arr, accuracy)

	// Update key stat
	ma.keyStat.SetValue(fmt.Sprintf("%.1f%%", accuracy*100))

	// Log result
	ma.operationLog.Add(fmt.Sprintf("Accuracy: %.1f%% on %d samples", accuracy*100, len(ma.testImages)))
	ma.updateStatus(fmt.Sprintf("● IDLE | Evaluation complete: %.1f%% accuracy", accuracy*100))
	ma.modeIndicator.SetMode(MNISTModeIdle)
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
