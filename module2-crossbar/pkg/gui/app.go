// Package gui provides Fyne-based GUI components for crossbar visualization.
package gui

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
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

	"multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
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

var debug *log.Logger
var logFile *os.File

func init() {
	// Create logs directory
	logsDir := "<local-path>"
	os.MkdirAll(logsDir, 0755)

	// Create log file with datetime
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-crossbar-module02.log")

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

// CrossbarApp is the main application for the crossbar demo.
type CrossbarApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Core components
	array  *crossbar.Array
	config *crossbar.Config

	// GUI components
	conductanceHeatmap *CrossbarHeatmap
	irDropHeatmap      *CrossbarHeatmap
	sneakPathHeatmap   *CrossbarHeatmap

	controlPanel   *ControlPanel
	statsPanel     *StatsPanel
	levelIndicator *LevelIndicator

	// Enhanced widgets
	metricsPanel       *MetricsPanel
	comparisonBadge    *ComparisonBadge
	accuracyWaterfall  *AccuracyWaterfall
	beforeAfterToggle  *BeforeAfterToggle

	// Live Slide components
	modeIndicator    *ModeIndicatorBox
	educationalPanel *EducationalPanel
	operationLog     *OperationLog
	ioDisplay        *InputOutputDisplay
	keyStat          *KeyStatBox

	// Simple left panel labels (replacing custom widgets)
	eduTitleLabel   *widget.Label
	eduContentLabel *widget.Label
	keyStatLabel    *widget.Label
	keyStatValue    *widget.Label

	// Simple right panel widgets (replacing custom widgets)
	runMVMButton    *widget.Button
	resetButton     *widget.Button
	arraySizeLabel  *widget.Label
	arraySizeSlider *widget.Slider
	noiseLabel      *widget.Label
	noiseSlider     *widget.Slider
	adcBitsLabel    *widget.Label
	adcBitsSlider   *widget.Slider
	colormapSelect  *widget.Select
	statsLabel      *widget.Label

	// Status
	statusLabel    *widget.Label
	infoLabel      *widget.Label
	hoverInfoLabel *widget.Label

	// Current state
	lastInput          []float64
	lastOutput         []float64
	lastMVMResult      *crossbar.MVMResult
	lastIRDropAnalysis *crossbar.IRDropAnalysis
	lastSneakAnalysis  *crossbar.SneakPathAnalysis

	// Vector visualization
	mvmVis *MVMVisualization

	// Tabs container for programmatic switching
	tabs *container.AppTabs

	// Auto demo state
	autoDemo      bool
	autoDemoStep  int
	autoDemoTimer *time.Ticker
	stopAutoDemo  chan bool
}

// NewCrossbarApp creates and initializes the crossbar demo application.
func NewCrossbarApp() *CrossbarApp {
	ca := &CrossbarApp{}

	// Create Fyne app
	ca.fyneApp = app.NewWithID("com.fecim.crossbar-demo")
	ca.fyneApp.Settings().SetTheme(&feCIMTheme{})

	// Initialize with default config
	ca.config = &crossbar.Config{
		Rows:       64,
		Cols:       64,
		NoiseLevel: 0.02,
		ADCBits:    6,
		DACBits:    8,
	}

	// Create crossbar array
	var err error
	ca.array, err = crossbar.NewArray(ca.config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create crossbar array: %v", err))
	}

	// Program initial random weights
	ca.programRandomWeights()

	return ca
}

// Run starts the GUI application.
func (ca *CrossbarApp) Run() {
	ca.RunWithLayout(false) // Use standard layout by default
}

// RunEnhanced starts the GUI with enhanced features.
func (ca *CrossbarApp) RunEnhanced() {
	ca.RunWithLayout(true) // Use enhanced layout
}

// RunWithLayout starts the GUI application with specified layout.
func (ca *CrossbarApp) RunWithLayout(enhanced bool) {
	debug.Println("App: Creating window")
	windowTitle := "FeCIM Demo 2: Crossbar Array MVM"
	if enhanced {
		windowTitle += " (Enhanced)"
	}
	ca.window = ca.fyneApp.NewWindow(windowTitle)
	ca.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	debug.Println("App: Creating main layout")
	var content fyne.CanvasObject
	if enhanced {
		content = ca.createEnhancedMainLayout()
	} else {
		content = ca.createMainLayout()
	}
	debug.Println("App: Setting window content")
	ca.window.SetContent(content)

	// Initialize displays
	debug.Println("App: Updating conductance display")
	ca.updateConductanceDisplay()
	debug.Println("App: Updating status")
	ca.updateStatus("Ready. Program weights and run MVM operations.")

	debug.Println("App: ShowAndRun starting")
	ca.window.ShowAndRun()
	debug.Println("App: Window closed")
}

// createMainLayout builds the main application layout.
func (ca *CrossbarApp) createMainLayout() fyne.CanvasObject {
	// Create heatmaps
	ca.conductanceHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.conductanceHeatmap.SetColormap("fecim")
	ca.conductanceHeatmap.OnCellTapped = ca.onCellTapped
	ca.conductanceHeatmap.OnCellHover = ca.onCellHover

	ca.irDropHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.irDropHeatmap.SetColormap("viridis")
	ca.irDropHeatmap.OnCellTapped = ca.onIRDropCellTapped
	ca.irDropHeatmap.OnCellHover = ca.onIRDropCellHover

	ca.sneakPathHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.sneakPathHeatmap.SetColormap("plasma")
	ca.sneakPathHeatmap.OnCellTapped = ca.onSneakCellTapped
	ca.sneakPathHeatmap.OnCellHover = ca.onSneakCellHover

	// Create MVM visualization with bar charts
	ca.mvmVis = NewMVMVisualization()

	// Create level indicator and mode indicator
	ca.levelIndicator = NewLevelIndicator()
	ca.modeIndicator = NewModeIndicatorBox()

	// Create simple LEFT panel labels
	ca.eduTitleLabel = widget.NewLabelWithStyle("What You're Seeing", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ca.eduContentLabel = widget.NewLabel("CROSSBAR MVM\n\nClick a button to start\na demonstration.")
	ca.eduContentLabel.Wrapping = fyne.TextWrapOff
	ca.keyStatLabel = widget.NewLabel("N² Operations")
	ca.keyStatLabel.Alignment = fyne.TextAlignCenter
	ca.keyStatValue = widget.NewLabelWithStyle(fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create simple RIGHT panel widgets directly (no custom widgets)
	ca.runMVMButton = widget.NewButton("Run MVM", ca.runMVM)
	ca.runMVMButton.Importance = widget.HighImportance
	ca.resetButton = widget.NewButton("Reset Array", ca.resetArray)

	ca.arraySizeLabel = widget.NewLabel("Array Size: 64x64")
	ca.arraySizeSlider = widget.NewSlider(8, 128)
	ca.arraySizeSlider.Step = 8
	ca.arraySizeSlider.Value = 64
	ca.arraySizeSlider.OnChanged = func(v float64) {
		size := int(v)
		ca.arraySizeLabel.SetText(fmt.Sprintf("Array Size: %dx%d", size, size))
		ca.recreateArray(size, ca.config.NoiseLevel, ca.config.ADCBits)
	}

	ca.noiseLabel = widget.NewLabel("Noise: 2.0%")
	ca.noiseSlider = widget.NewSlider(0, 20)
	ca.noiseSlider.Step = 0.5
	ca.noiseSlider.Value = 2
	ca.noiseSlider.OnChanged = func(v float64) {
		ca.noiseLabel.SetText(fmt.Sprintf("Noise: %.1f%%", v))
		ca.config.NoiseLevel = v / 100.0
	}

	ca.adcBitsLabel = widget.NewLabel("ADC Bits: 6")
	ca.adcBitsSlider = widget.NewSlider(4, 10)
	ca.adcBitsSlider.Step = 1
	ca.adcBitsSlider.Value = 6
	ca.adcBitsSlider.OnChanged = func(v float64) {
		bits := int(v)
		ca.adcBitsLabel.SetText(fmt.Sprintf("ADC Bits: %d", bits))
		ca.config.ADCBits = bits
	}

	ca.colormapSelect = widget.NewSelect([]string{"fecim", "viridis", "plasma", "coolwarm"}, func(s string) {
		ca.conductanceHeatmap.SetColormap(s)
	})
	ca.colormapSelect.SetSelected("fecim")

	ca.statsLabel = widget.NewLabel("Analysis Results\n\nNo data yet.\nClick Run MVM to start.")
	ca.statsLabel.Wrapping = fyne.TextWrapOff
	ca.statsLabel.TextStyle = fyne.TextStyle{Monospace: true} // Fixed-width prevents resize

	// Create status labels
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	ca.infoLabel = widget.NewLabel(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))

	// Hover info label - shows cell info on mouse hover
	ca.hoverInfoLabel = widget.NewLabel("Hover over cells to see values")
	ca.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Create tabbed heatmap view - use Max to fill available space
	ca.tabs = container.NewAppTabs(
		container.NewTabItem("Conductance", container.NewMax(ca.conductanceHeatmap)),
		container.NewTabItem("IR Drop", container.NewMax(ca.irDropHeatmap)),
		container.NewTabItem("Sneak Paths", container.NewMax(ca.sneakPathHeatmap)),
		container.NewTabItem("Input/Output", container.NewMax(ca.mvmVis)),
	)

	// Update educational panel based on selected tab
	ca.tabs.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "Conductance":
			ca.setEducationalContent("Conductance Matrix",
				"Each cell = one FeFET\n\n"+
					"Color = conductance level\n"+
					"(0-29 discrete states)\n\n"+
					"This is your weight matrix W\n"+
					"for neural network inference.")
		case "IR Drop":
			ca.setEducationalContent("IR Drop Analysis",
				"Voltage drops along wires.\n\n"+
					"Red = high drop (bad)\n"+
					"Blue = low drop (good)\n\n"+
					"Cells far from drivers\n"+
					"see reduced voltage.\n\n"+
					"Auto-computed after\n"+
					"each MVM operation.")
		case "Sneak Paths":
			ca.setEducationalContent("Sneak Path Analysis",
				"Parasitic currents through\n"+
					"unselected cells.\n\n"+
					"Red = high sneak current\n"+
					"Blue = low sneak current\n\n"+
					"Bigger arrays = worse.\n"+
					"Use selector devices\n"+
					"to mitigate.")
		case "Input/Output":
			ca.setEducationalContent("MVM Vectors",
				"Top: Input voltages (V)\n"+
					"Bottom: Output currents (I)\n\n"+
					"I = W × V\n"+
					"(matrix-vector multiply)\n\n"+
					"Click 'Run MVM' to see\n"+
					"the computation.")
		}
	}

	// Title and header
	titleLabel := widget.NewLabel("FeCIM Crossbar Array Visualization")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
	)

	// Right panel - improved layout with better spacing and grouping

	// Action buttons group
	actionLabel := widget.NewLabelWithStyle("Actions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	actionsGroup := container.NewVBox(
		actionLabel,
		ca.runMVMButton,
		ca.resetButton,
	)

	// Array settings group - collapsible style header
	settingsLabel := widget.NewLabelWithStyle("Array Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	settingsGroup := container.NewVBox(
		widget.NewSeparator(),
		settingsLabel,
		ca.arraySizeLabel,
		ca.arraySizeSlider,
	)

	// Noise/ADC group
	signalLabel := widget.NewLabelWithStyle("Signal Quality", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	signalGroup := container.NewVBox(
		widget.NewSeparator(),
		signalLabel,
		ca.noiseLabel,
		ca.noiseSlider,
		ca.adcBitsLabel,
		ca.adcBitsSlider,
	)

	// Display settings group
	displayLabel := widget.NewLabelWithStyle("Display", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	displayGroup := container.NewVBox(
		widget.NewSeparator(),
		displayLabel,
		ca.colormapSelect,
	)

	// Combined controls with scroll for overflow
	controlsBox := container.NewVBox(
		actionsGroup,
		settingsGroup,
		signalGroup,
		displayGroup,
	)
	controlsScroll := container.NewVScroll(controlsBox)
	controlsScroll.SetMinSize(fyne.NewSize(180, 250))

	// Stats section with header
	statsHeader := widget.NewLabelWithStyle("Analysis Results", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	statsSection := container.NewBorder(
		container.NewVBox(widget.NewSeparator(), statsHeader),
		nil, nil, nil,
		ca.statsLabel,
	)
	statsScroll := container.NewVScroll(statsSection)
	statsScroll.SetMinSize(fyne.NewSize(180, 120))

	// Use VSplit for controls and stats
	rightPanel := container.NewVSplit(
		controlsScroll,
		statsScroll,
	)
	rightPanel.SetOffset(0.6) // 60% controls, 40% stats

	// Left panel using simple labels (no custom widgets)
	leftPanel := container.NewVBox(
		ca.eduTitleLabel,
		widget.NewSeparator(),
		ca.eduContentLabel,
		widget.NewSeparator(),
		ca.keyStatLabel,
		ca.keyStatValue,
	)

	// Simple status footer with hover info
	simpleFooter := container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		ca.hoverInfoLabel,
		widget.NewSeparator(),
		ca.infoLabel,
	)

	// Use HSplit for proportional 3-column layout
	// Left panel (15%) | Center tabs (70%) | Right panel (15%)
	leftCenterSplit := container.NewHSplit(leftPanel, ca.tabs)
	leftCenterSplit.SetOffset(0.15) // 15% left, 85% center+right

	mainSplit := container.NewHSplit(leftCenterSplit, rightPanel)
	mainSplit.SetOffset(0.8) // 80% left+center, 20% right

	// Wrap with header and footer
	mainContent := container.NewBorder(
		header,       // top
		simpleFooter, // bottom
		nil,          // left
		nil,          // right
		mainSplit,    // center - the split panels
	)

	return mainContent
}

// setupControlCallbacks connects control panel events to actions.
func (ca *CrossbarApp) setupControlCallbacks() {
	ca.controlPanel.OnArraySizeChanged = func(size int) {
		ca.recreateArray(size, ca.config.NoiseLevel, ca.config.ADCBits)
	}

	ca.controlPanel.OnNoiseChanged = func(noise float64) {
		ca.config.NoiseLevel = noise
		ca.recreateArray(ca.config.Rows, noise, ca.config.ADCBits)
	}

	ca.controlPanel.OnADCBitsChanged = func(bits int) {
		ca.config.ADCBits = bits
		ca.recreateArray(ca.config.Rows, ca.config.NoiseLevel, bits)
	}

	ca.controlPanel.OnColormapChanged = func(colormap string) {
		ca.conductanceHeatmap.SetColormap(colormap)
	}

	ca.controlPanel.OnDemoModeChanged = ca.onDemoModeChanged
	ca.controlPanel.OnRunMVM = ca.runMVM
	ca.controlPanel.OnAnalyzeIR = ca.analyzeIRDrop
	ca.controlPanel.OnAnalyzeSneak = ca.analyzeSneakPaths
	ca.controlPanel.OnReset = ca.resetArray
}

// recreateArray creates a new array with updated parameters.
func (ca *CrossbarApp) recreateArray(size int, noise float64, adcBits int) {
	ca.config = &crossbar.Config{
		Rows:       size,
		Cols:       size,
		NoiseLevel: noise,
		ADCBits:    adcBits,
		DACBits:    8,
	}

	var err error
	ca.array, err = crossbar.NewArray(ca.config)
	if err != nil {
		ca.updateStatus(fmt.Sprintf("Error: %v", err))
		return
	}

	// Resize existing heatmaps instead of creating new ones
	// This preserves the widget references in the window layout
	ca.conductanceHeatmap.SetDimensions(size, size)
	ca.irDropHeatmap.SetDimensions(size, size)
	ca.sneakPathHeatmap.SetDimensions(size, size)

	ca.programRandomWeights()
	ca.updateConductanceDisplay()
	ca.updateInfoLabel()
	ca.setKeyStatValue(fmt.Sprintf("%d MACs", size*size))
	ca.updateStatus(fmt.Sprintf("Array resized to %dx%d", size, size))
}

// programRandomWeights fills the array with random weights quantized to 30 levels.
func (ca *CrossbarApp) programRandomWeights() {
	for i := 0; i < ca.config.Rows; i++ {
		for j := 0; j < ca.config.Cols; j++ {
			level := rand.Intn(30)
			weight := float64(level) / 29.0
			ca.array.ProgramWeight(i, j, weight)
		}
	}
}

// updateConductanceDisplay refreshes the conductance heatmap.
func (ca *CrossbarApp) updateConductanceDisplay() {
	matrix := ca.array.GetConductanceMatrix()
	ca.conductanceHeatmap.SetData(matrix)
}

// updateStatus updates the status label.
func (ca *CrossbarApp) updateStatus(status string) {
	ca.statusLabel.SetText("Status: " + status)
}

// setEducationalContent updates the educational panel.
func (ca *CrossbarApp) setEducationalContent(title, content string) {
	ca.eduTitleLabel.SetText(title)
	ca.eduContentLabel.SetText(content)
}

// setKeyStatValue updates the key statistic display.
func (ca *CrossbarApp) setKeyStatValue(value string) {
	ca.keyStatValue.SetText(value)
}

// updateInfoLabel updates the info label with current config.
func (ca *CrossbarApp) updateInfoLabel() {
	ca.infoLabel.SetText(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))
}

// onCellTapped handles clicks on heatmap cells.
func (ca *CrossbarApp) onCellTapped(row, col int) {
	ca.modeIndicator.SetMode(DemoModeRead)

	matrix := ca.array.GetConductanceMatrix()
	value := matrix[row][col]
	level := int(value * 29)

	ca.levelIndicator.SetLevel(level)

	ca.statsLabel.SetText(fmt.Sprintf(
		"Selected Cell: [%d, %d]\n"+
			"Conductance: %.4f\n"+
			"Level: %d / 29\n"+
			"Quantized Value: %.4f",
		row, col, value, level, float64(level)/29.0,
	))

	ca.updateStatus(fmt.Sprintf("READ | Cell [%d,%d] = Level %d/30", row, col, level+1))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// onCellHover handles mouse hover over heatmap cells.
func (ca *CrossbarApp) onCellHover(row, col int, value float64) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells to see values")
		return
	}
	level := int(value * 29)
	ca.hoverInfoLabel.SetText(fmt.Sprintf("[%d,%d] G=%.4f L%d/29 Layer1", row, col, value, level))
}

// onIRDropCellTapped handles clicks on IR Drop heatmap.
func (ca *CrossbarApp) onIRDropCellTapped(row, col int) {
	ca.irDropHeatmap.SetSelection(row, col)
	// Get the displayed value from the heatmap data
	ca.statsLabel.SetText(fmt.Sprintf(
		"IR Drop Cell: [%d, %d]\n\n"+
			"Click Run MVM to update\n"+
			"IR Drop analysis.",
		row, col,
	))
}

// onIRDropCellHover handles hover on IR Drop heatmap.
func (ca *CrossbarApp) onIRDropCellHover(row, col int, value float64) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells")
		return
	}
	// Get conductance from array
	conductance := ca.array.GetConductanceMatrix()[row][col]
	level := int(conductance * 29)

	// Get voltage from analysis if available
	voltageStr := "N/A"
	if ca.lastIRDropAnalysis != nil && row < len(ca.lastIRDropAnalysis.EffectiveVoltage) &&
		col < len(ca.lastIRDropAnalysis.EffectiveVoltage[0]) {
		voltage := ca.lastIRDropAnalysis.EffectiveVoltage[row][col]
		voltageStr = fmt.Sprintf("%.3fV", voltage)
	}

	ca.hoverInfoLabel.SetText(fmt.Sprintf("[%d,%d] G=%.3f L%d V=%s Drop=%.1f%% Layer1",
		row, col, conductance, level, voltageStr, value*100))
}

// onSneakCellTapped handles clicks on Sneak Path heatmap.
func (ca *CrossbarApp) onSneakCellTapped(row, col int) {
	ca.sneakPathHeatmap.SetSelection(row, col)
	ca.statsLabel.SetText(fmt.Sprintf(
		"Sneak Path Cell: [%d, %d]\n\n"+
			"Click Run MVM to update\n"+
			"Sneak Path analysis.",
		row, col,
	))
}

// onSneakCellHover handles hover on Sneak Path heatmap.
func (ca *CrossbarApp) onSneakCellHover(row, col int, value float64) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells")
		return
	}
	// Get conductance from array
	conductance := ca.array.GetConductanceMatrix()[row][col]
	level := int(conductance * 29)

	// Get original sneak current (before sqrt transform)
	sneakCurrent := value * value // Undo sqrt transform

	ca.hoverInfoLabel.SetText(fmt.Sprintf("[%d,%d] G=%.3f L%d Sneak=%.4f Layer1",
		row, col, conductance, level, sneakCurrent))
}

// runMVM performs matrix-vector multiplication with animation.
func (ca *CrossbarApp) runMVM() {
	debug.Println("runMVM: Starting")

	// Disable button during computation
	ca.runMVMButton.Disable()

	// Create random input
	debug.Printf("runMVM: Creating input vector of size %d", ca.config.Cols)
	input := make([]float64, ca.config.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}
	ca.lastInput = input
	ca.mvmVis.SetInput(input)

	// Run animated MVM in goroutine
	go ca.runMVMAnimated(input)
}

// runMVMAnimated performs the MVM with visual animation.
func (ca *CrossbarApp) runMVMAnimated(input []float64) {
	// Phase 1: Input voltages applied (300ms)
	fyne.Do(func() {
		ca.modeIndicator.SetMode(DemoModeCompute)
		ca.updateStatus("COMPUTE | Phase 1: Applying input voltages...")

		// Highlight all columns to show input voltages
		cols := make([]int, ca.config.Cols)
		for i := range cols {
			cols[i] = i
		}
		ca.conductanceHeatmap.SetInputHighlight(cols)
		ca.conductanceHeatmap.SetAnimPhase(1, 0)
	})
	time.Sleep(300 * time.Millisecond)

	// Phase 2: Current flowing through cells (500ms animation)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 2: Current flowing through cells...")
	})

	// Animate wave propagation
	steps := 10
	for i := 0; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		fyne.Do(func() {
			ca.conductanceHeatmap.SetAnimPhase(2, progress)
		})
		time.Sleep(50 * time.Millisecond)
	}

	// Perform actual MVM computation
	output, err := ca.array.MVM(input)
	if err != nil {
		fyne.Do(func() {
			ca.updateStatus(fmt.Sprintf("COMPUTE | Error: %v", err))
			ca.modeIndicator.SetMode(DemoModeIdle)
			ca.conductanceHeatmap.ClearAnimation()
			ca.runMVMButton.Enable()
		})
		return
	}
	ca.lastOutput = output

	// Phase 3: Output currents collected (300ms)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 3: Collecting output currents...")
		ca.mvmVis.SetOutput(output)

		// Highlight all rows to show output currents
		rows := make([]int, ca.config.Rows)
		for i := range rows {
			rows[i] = i
		}
		ca.conductanceHeatmap.SetOutputHighlight(rows)
		ca.conductanceHeatmap.SetAnimPhase(3, 1)
	})
	time.Sleep(300 * time.Millisecond)

	// Finish and show results
	fyne.Do(func() {
		ca.conductanceHeatmap.ClearAnimation()

		// Calculate stats
		var sumInput, sumOutput float64
		for _, v := range input {
			sumInput += v
		}
		for _, v := range output {
			sumOutput += v
		}

		reads, writes := ca.array.GetStats()
		macOps := ca.config.Rows * ca.config.Cols

		ca.statsLabel.SetText(fmt.Sprintf(
			"MVM Complete!\n"+
				"Input Sum: %.4f\n"+
				"Output Sum: %.4f\n"+
				"Total Reads: %d\n"+
				"Total Writes: %d\n"+
				"MAC Operations: %d",
			sumInput, sumOutput, reads, writes, macOps,
		))

		ca.updateStatus(fmt.Sprintf("COMPUTE | Complete: %d parallel MACs", macOps))
		ca.modeIndicator.SetMode(DemoModeIdle)

		// Auto-run IR Drop and Sneak Path analysis
		ca.runIRDropAnalysis()
		ca.runSneakPathAnalysis()
	})

	// Cycle through all tabs (4 seconds each)
	tabNames := []string{"Conductance", "IR Drop", "Sneak Paths", "Input/Output"}
	for i, name := range tabNames {
		fyne.Do(func() {
			ca.tabs.SelectIndex(i)
			ca.updateStatus(fmt.Sprintf("Showing: %s", name))
		})
		time.Sleep(4 * time.Second)
	}

	// Return to Conductance tab and re-enable button
	fyne.Do(func() {
		ca.tabs.SelectIndex(0)
		ca.updateStatus("Ready for next MVM")
		ca.runMVMButton.Enable()
	})
	debug.Println("runMVM: Complete")
}

// runIRDropAnalysis updates IR drop heatmap silently (no tab switch).
func (ca *CrossbarApp) runIRDropAnalysis() {
	input := ca.lastInput
	if input == nil || len(input) != ca.config.Cols {
		input = make([]float64, ca.config.Cols)
		for i := range input {
			input[i] = rand.Float64()
		}
	}

	params := crossbar.DefaultWireParams()
	analysis := ca.array.AnalyzeIRDrop(input, params)
	ca.lastIRDropAnalysis = analysis // Store for hover info
	irMap := analysis.GetIRDropMap()
	ca.irDropHeatmap.SetData(irMap)
	ca.irDropHeatmap.SetSelection(analysis.WorstCaseCell[0], analysis.WorstCaseCell[1])
}

// runSneakPathAnalysis updates sneak path heatmap silently (no tab switch).
func (ca *CrossbarApp) runSneakPathAnalysis() {
	selectedRow := ca.config.Rows / 2
	selectedCol := ca.config.Cols / 2

	analysis := ca.array.AnalyzeSneakPaths(selectedRow, selectedCol)
	ca.lastSneakAnalysis = analysis // Store for hover info
	sneakMap := analysis.GetSneakMap()

	// Apply sqrt transformation for better visibility
	for i := range sneakMap {
		for j := range sneakMap[i] {
			sneakMap[i][j] = math.Sqrt(sneakMap[i][j])
		}
	}

	ca.sneakPathHeatmap.SetData(sneakMap)
	ca.sneakPathHeatmap.SetSelection(selectedRow, selectedCol)
}

// analyzeIRDrop performs IR drop analysis.
func (ca *CrossbarApp) analyzeIRDrop() {
	// Switch to IR Drop tab
	ca.tabs.SelectIndex(1)

	// Update mode and educational panel
	ca.modeIndicator.SetMode(DemoModeIRDrop)
	ca.setEducationalContent("Non-Ideality: IR Drop", "IR DROP ANALYSIS\n\nWire resistance causes\nvoltage drop along lines.\n\nCells far from drivers\nsee reduced voltage.\n\nThis affects accuracy:\n• Worst at corners\n• Mitigate with drivers")
	ca.updateStatus("IR DROP | Analyzing voltage drops...")

	// Use last input or create new one
	input := ca.lastInput
	if input == nil || len(input) != ca.config.Cols {
		input = make([]float64, ca.config.Cols)
		for i := range input {
			input[i] = rand.Float64()
		}
		ca.lastInput = input
	}

	// Analyze IR drop
	params := crossbar.DefaultWireParams()
	analysis := ca.array.AnalyzeIRDrop(input, params)

	// Update IR drop heatmap
	irMap := analysis.GetIRDropMap()
	debug.Printf("IR Drop map size: %dx%d, MaxIRDrop: %.6f", len(irMap), len(irMap[0]), analysis.MaxIRDrop)
	// Sample some values to see the range
	if len(irMap) > 0 && len(irMap[0]) > 0 {
		debug.Printf("IR Drop sample values: [0,0]=%.4f, [mid,mid]=%.4f", irMap[0][0], irMap[len(irMap)/2][len(irMap[0])/2])
	}
	ca.irDropHeatmap.SetData(irMap)

	// Update stats
	ca.statsLabel.SetText(fmt.Sprintf(
		"IR Drop Analysis\n"+
			"Max IR Drop: %.2f%%\n"+
			"Avg IR Drop: %.2f%%\n"+
			"Variance: %.6f\n"+
			"Worst Cell: [%d, %d]\n\n"+
			"Wire Parameters:\n"+
			"Word Line R: %.1f Ω/cell\n"+
			"Bit Line R: %.1f Ω/cell\n"+
			"Contact R: %.1f Ω",
		analysis.MaxIRDrop*100,
		analysis.AvgIRDrop*100,
		analysis.IRDropVariance,
		analysis.WorstCaseCell[0], analysis.WorstCaseCell[1],
		params.RwordLine, params.RbitLine, params.Rcontact,
	))

	// Highlight worst cell
	ca.irDropHeatmap.SetSelection(analysis.WorstCaseCell[0], analysis.WorstCaseCell[1])

	// Update key stat and log
	ca.setKeyStatValue(fmt.Sprintf("Max: %.1f%% drop", analysis.MaxIRDrop*100))

	ca.updateStatus(fmt.Sprintf("IR DROP | Complete: Max %.2f%% at [%d,%d]",
		analysis.MaxIRDrop*100, analysis.WorstCaseCell[0], analysis.WorstCaseCell[1]))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// analyzeSneakPaths performs sneak path analysis.
func (ca *CrossbarApp) analyzeSneakPaths() {
	// Switch to Sneak Paths tab
	ca.tabs.SelectIndex(2)

	// Update mode and educational panel
	ca.modeIndicator.SetMode(DemoModeSneakPath)
	ca.setEducationalContent("Non-Ideality: Sneak Paths", "SNEAK PATH ANALYSIS\n\nCurrent can flow through\nunintended paths in passive\ncrossbar arrays.\n\nMitigation strategies:\n• Selector devices\n• 1T1R architecture\n• Threshold switching")
	ca.updateStatus("SNEAK | Analyzing parasitic paths...")

	// Select center cell
	selectedRow := ca.config.Rows / 2
	selectedCol := ca.config.Cols / 2

	// Analyze sneak paths
	analysis := ca.array.AnalyzeSneakPaths(selectedRow, selectedCol)

	// Update sneak path heatmap with sqrt transform for better visibility
	sneakMap := analysis.GetSneakMap()
	debug.Printf("Sneak map size: %dx%d", len(sneakMap), len(sneakMap[0]))

	// Apply sqrt transformation to make small variations more visible
	for i := range sneakMap {
		for j := range sneakMap[i] {
			sneakMap[i][j] = math.Sqrt(sneakMap[i][j])
		}
	}

	if len(sneakMap) > 0 && len(sneakMap[0]) > 0 {
		debug.Printf("Sneak sample values (sqrt): [0,0]=%.4f, [mid,mid]=%.4f", sneakMap[0][0], sneakMap[len(sneakMap)/2][len(sneakMap[0])/2])
	}
	ca.sneakPathHeatmap.SetData(sneakMap)

	// Highlight selected cell
	ca.sneakPathHeatmap.SetSelection(selectedRow, selectedCol)

	// Calculate signal-to-sneak ratio
	snr := 0.0
	if analysis.TotalSneak > 0 {
		snr = analysis.TotalSignal / analysis.TotalSneak
	}

	// Update stats
	ca.statsLabel.SetText(fmt.Sprintf(
		"Sneak Path Analysis\n"+
			"Selected Cell: [%d, %d]\n"+
			"Signal Current: %.6f\n"+
			"Total Sneak: %.6f\n"+
			"Max Sneak/Signal: %.2f%%\n"+
			"Avg Sneak/Signal: %.2f%%\n"+
			"Signal/Sneak Ratio: %.1f:1\n\n"+
			"Impact Assessment:\n%s",
		selectedRow, selectedCol,
		analysis.TotalSignal,
		analysis.TotalSneak,
		analysis.MaxSneakRatio*100,
		analysis.AvgSneakRatio*100,
		snr,
		getImpactAssessment(analysis.MaxSneakRatio),
	))

	// Update key stat and log
	ca.setKeyStatValue(fmt.Sprintf("SNR: %.1f:1", snr))

	ca.updateStatus(fmt.Sprintf("SNEAK | Complete: SNR %.1f:1 at [%d,%d]",
		snr, selectedRow, selectedCol))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// resetArray resets the array with new random weights.
func (ca *CrossbarApp) resetArray() {
	ca.modeIndicator.SetMode(DemoModeWrite)
	ca.updateStatus("WRITE | Programming random weights...")

	ca.programRandomWeights()
	ca.updateConductanceDisplay()
	ca.lastInput = nil
	ca.lastOutput = nil
	ca.conductanceHeatmap.ClearSelection()
	ca.irDropHeatmap.ClearSelection()
	ca.sneakPathHeatmap.ClearSelection()
	ca.statsLabel.SetText("Array reset with new random weights.\n\nSelect a cell or run an analysis.")

	// Update key stat
	ca.setKeyStatValue(fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols))

	ca.setEducationalContent("What You're Seeing", "CROSSBAR MVM\n\nClick a button to start\na demonstration.")
	ca.updateStatus("● IDLE | Array reset with random weights")
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// getImpactAssessment returns a text assessment of sneak path severity.
func getImpactAssessment(maxRatio float64) string {
	if maxRatio < 0.01 {
		return "✓ Sneak paths negligible\n  Excellent cell isolation"
	} else if maxRatio < 0.05 {
		return "⚠ Moderate sneak paths\n  Consider selector devices"
	}
	return "✗ Significant sneak paths\n  1T1R or selector required"
}

// onDemoModeChanged handles demo mode selection changes.
func (ca *CrossbarApp) onDemoModeChanged(mode string) {
	// Stop any existing auto demo
	ca.stopAutoDemoLoop()

	switch mode {
	case "Auto Demo":
		ca.startAutoDemoLoop()
	case "Step-by-Step":
		ca.setEducationalContent("Step-by-Step Mode",
			"Click each button to see\nthe operation explained.\n\n"+
				"After 'Run MVM':\n"+
				"• IR Drop computed\n"+
				"• Sneak Paths computed\n"+
				"• Tabs auto-cycle")
	case "Manual":
		ca.setEducationalContent("What You're Seeing", "CROSSBAR MVM\n\nClick a button to start\na demonstration.")
	}
}

// startAutoDemoLoop starts the automatic demo loop.
func (ca *CrossbarApp) startAutoDemoLoop() {
	ca.autoDemo = true
	ca.autoDemoStep = 0
	ca.stopAutoDemo = make(chan bool)
	ca.autoDemoTimer = time.NewTicker(3 * time.Second)

	ca.setEducationalContent("Auto Demo Mode",
		"Watch the demo cycle through\nall operations automatically.\n\n"+
			"Operations:\n"+
			"1. MVM Computation\n"+
			"2. IR Drop Analysis\n"+
			"3. Sneak Path Analysis\n"+
			"4. Reset & Repeat")

	go ca.autoDemoLoop()
}

// stopAutoDemoLoop stops the automatic demo loop.
func (ca *CrossbarApp) stopAutoDemoLoop() {
	if ca.autoDemo {
		ca.autoDemo = false
		if ca.stopAutoDemo != nil {
			close(ca.stopAutoDemo)
		}
		if ca.autoDemoTimer != nil {
			ca.autoDemoTimer.Stop()
		}
	}
}

// autoDemoLoop runs the automatic demonstration.
func (ca *CrossbarApp) autoDemoLoop() {
	// Run first operation immediately
	ca.runAutoDemoStep()

	for {
		select {
		case <-ca.stopAutoDemo:
			return
		case <-ca.autoDemoTimer.C:
			if !ca.autoDemo {
				return
			}
			ca.runAutoDemoStep()
		}
	}
}

// runAutoDemoStep executes one step of the auto demo.
func (ca *CrossbarApp) runAutoDemoStep() {
	switch ca.autoDemoStep {
	case 0:
		fyne.Do(func() {
			ca.runMVM()
		})
	case 1:
		fyne.Do(func() {
			ca.analyzeIRDrop()
		})
	case 2:
		fyne.Do(func() {
			ca.analyzeSneakPaths()
		})
	case 3:
		fyne.Do(func() {
			ca.resetArray()
		})
	}

	ca.autoDemoStep = (ca.autoDemoStep + 1) % 4
}
