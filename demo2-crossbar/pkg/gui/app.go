// Package gui provides Fyne-based GUI components for crossbar visualization.
package gui

import (
	"fmt"
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

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
)

var debug *log.Logger
var logFile *os.File

func init() {
	// Create logs directory
	logsDir := "<local-path>"
	os.MkdirAll(logsDir, 0755)

	// Create log file with datetime
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-crossbar-demo02.log")

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

	// Live Slide components
	modeIndicator    *ModeIndicatorBox
	educationalPanel *EducationalPanel
	operationLog     *OperationLog
	ioDisplay        *InputOutputDisplay
	keyStat          *KeyStatBox

	// Status
	statusLabel *widget.Label
	infoLabel   *widget.Label

	// Current state
	lastInput  []float64
	lastOutput []float64

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
	ca.fyneApp = app.NewWithID("com.ironlattice.crossbar-demo")
	ca.fyneApp.Settings().SetTheme(theme.DarkTheme())

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
	debug.Println("App: Creating window")
	ca.window = ca.fyneApp.NewWindow("IronLattice Demo 2: Crossbar Array MVM")
	ca.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	debug.Println("App: Creating main layout")
	content := ca.createMainLayout()
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
	ca.conductanceHeatmap.SetColormap("ironlattice")
	ca.conductanceHeatmap.OnCellTapped = ca.onCellTapped

	ca.irDropHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.irDropHeatmap.SetColormap("coolwarm")

	ca.sneakPathHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.sneakPathHeatmap.SetColormap("plasma")

	// Create control panel
	ca.controlPanel = NewControlPanel()
	ca.setupControlCallbacks()

	// Create stats panel
	ca.statsPanel = NewStatsPanel("Analysis Results")

	// Create level indicator
	ca.levelIndicator = NewLevelIndicator()

	// Create Live Slide components
	ca.modeIndicator = NewModeIndicatorBox()
	ca.educationalPanel = NewEducationalPanel()
	ca.educationalPanel.SetIdleExplanation()
	ca.operationLog = NewOperationLog()
	ca.ioDisplay = NewInputOutputDisplay()
	ca.keyStat = NewKeyStatBox("N² Operations", fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols))

	// Create status labels
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	ca.infoLabel = widget.NewLabel(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))

	// Create tabbed heatmap view - use Max to fill available space
	tabs := container.NewAppTabs(
		container.NewTabItem("Conductance", container.NewMax(ca.conductanceHeatmap)),
		container.NewTabItem("IR Drop", container.NewMax(ca.irDropHeatmap)),
		container.NewTabItem("Sneak Paths", container.NewMax(ca.sneakPathHeatmap)),
	)

	// Title and header with Dr. Tour quote
	titleLabel := widget.NewLabel("IronLattice Crossbar Array Visualization")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	quoteLabel := widget.NewLabel("\"Compute in memory where the same device does memory and computation.\" — Dr. external research group")
	quoteLabel.Alignment = fyne.TextAlignCenter
	quoteLabel.TextStyle = fyne.TextStyle{Italic: true}

	header := container.NewVBox(
		titleLabel,
		quoteLabel,
		widget.NewSeparator(),
	)

	// Right panel (controls + stats) - compact, fixed width
	rightPanel := container.NewVBox(
		ca.controlPanel,
		widget.NewSeparator(),
		ca.statsPanel,
	)

	// Left panel (educational + key stat) - compact, fixed width
	leftPanel := container.NewVBox(
		ca.educationalPanel,
		widget.NewSeparator(),
		ca.keyStat,
	)

	// Simple status footer
	simpleFooter := container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		ca.infoLabel,
	)

	// Main content - simple 3-column layout with center heatmap
	mainContent := container.NewBorder(
		header,       // top
		simpleFooter, // bottom
		leftPanel,    // left
		rightPanel,   // right
		tabs,         // center
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

	ca.statsPanel.SetStats(fmt.Sprintf(
		"Selected Cell: [%d, %d]\n"+
			"Conductance: %.4f\n"+
			"Level: %d / 29\n"+
			"Quantized Value: %.4f",
		row, col, value, level, float64(level)/29.0,
	))

	ca.operationLog.Add(fmt.Sprintf("Read [%d,%d] → Level %d", row, col, level))
	ca.updateStatus(fmt.Sprintf("READ | Cell [%d,%d] = Level %d/30", row, col, level+1))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// runMVM performs matrix-vector multiplication.
func (ca *CrossbarApp) runMVM() {
	debug.Println("runMVM: Starting")

	// Update mode and educational panel
	debug.Println("runMVM: Setting mode to Compute")
	ca.modeIndicator.SetMode(DemoModeCompute)
	debug.Println("runMVM: Setting MVM explanation phase 1")
	ca.educationalPanel.SetMVMExplanation(1)
	debug.Println("runMVM: Updating status")
	ca.updateStatus("COMPUTE | Applying input voltages...")
	debug.Println("runMVM: Adding to operation log")
	ca.operationLog.Add("MVM: Generating input vector")

	// Create random input
	debug.Printf("runMVM: Creating input vector of size %d", ca.config.Cols)
	input := make([]float64, ca.config.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}
	ca.lastInput = input
	debug.Println("runMVM: Setting input display")
	ca.ioDisplay.SetInput(input)

	// Phase 2: Computing
	debug.Println("runMVM: Setting MVM explanation phase 2")
	ca.educationalPanel.SetMVMExplanation(2)
	ca.operationLog.Add("MVM: Computing I = G × V")

	// Perform MVM
	debug.Println("runMVM: Performing MVM computation")
	output, err := ca.array.MVM(input)
	if err != nil {
		debug.Printf("runMVM: Error - %v", err)
		ca.updateStatus(fmt.Sprintf("COMPUTE | Error: %v", err))
		ca.operationLog.AddWithResult("MVM", err.Error(), false)
		ca.modeIndicator.SetMode(DemoModeIdle)
		return
	}
	debug.Printf("runMVM: MVM complete, output size %d", len(output))
	ca.lastOutput = output
	debug.Println("runMVM: Setting output display")
	ca.ioDisplay.SetOutput(output)

	// Phase 3: Results
	debug.Println("runMVM: Setting MVM explanation phase 3")
	ca.educationalPanel.SetMVMExplanation(3)

	// Update stats
	var sumInput, sumOutput float64
	for _, v := range input {
		sumInput += v
	}
	for _, v := range output {
		sumOutput += v
	}

	reads, writes := ca.array.GetStats()
	macOps := ca.config.Rows * ca.config.Cols

	debug.Println("runMVM: Updating stats panel")
	ca.statsPanel.SetStats(fmt.Sprintf(
		"MVM Complete!\n"+
			"Input Sum: %.4f\n"+
			"Output Sum: %.4f\n"+
			"Total Reads: %d\n"+
			"Total Writes: %d\n"+
			"MAC Operations: %d",
		sumInput, sumOutput, reads, writes,
		macOps,
	))

	// Update key stat
	debug.Println("runMVM: Updating key stat")
	ca.keyStat.SetValue(fmt.Sprintf("%d MACs in 1 cycle", macOps))

	// Log completion
	debug.Println("runMVM: Logging completion")
	ca.operationLog.AddWithResult("MVM", fmt.Sprintf("%d ops", macOps), true)

	// Update status and return to idle
	debug.Println("runMVM: Final status update")
	ca.updateStatus(fmt.Sprintf("COMPUTE | Complete: %d parallel multiplications", macOps))
	debug.Println("runMVM: Setting mode to Idle")
	ca.modeIndicator.SetMode(DemoModeIdle)
	debug.Println("runMVM: Complete")
}

// analyzeIRDrop performs IR drop analysis.
func (ca *CrossbarApp) analyzeIRDrop() {
	// Update mode and educational panel
	ca.modeIndicator.SetMode(DemoModeIRDrop)
	ca.educationalPanel.SetIRDropExplanation()
	ca.updateStatus("IR DROP | Analyzing voltage drops...")
	ca.operationLog.Add("IR Drop: Starting analysis")

	// Use last input or create new one
	input := ca.lastInput
	if input == nil || len(input) != ca.config.Cols {
		input = make([]float64, ca.config.Cols)
		for i := range input {
			input[i] = rand.Float64()
		}
		ca.lastInput = input
		ca.ioDisplay.SetInput(input)
	}

	// Analyze IR drop
	params := crossbar.DefaultWireParams()
	analysis := ca.array.AnalyzeIRDrop(input, params)

	// Update IR drop heatmap
	irMap := analysis.GetIRDropMap()
	ca.irDropHeatmap.SetData(irMap)

	// Update stats
	ca.statsPanel.SetStats(fmt.Sprintf(
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
	ca.keyStat.SetValue(fmt.Sprintf("Max: %.1f%% drop", analysis.MaxIRDrop*100))
	ca.operationLog.AddWithResult("IR Drop", fmt.Sprintf("%.1f%% max", analysis.MaxIRDrop*100), analysis.MaxIRDrop < 0.1)

	ca.updateStatus(fmt.Sprintf("IR DROP | Complete: Max %.2f%% at [%d,%d]",
		analysis.MaxIRDrop*100, analysis.WorstCaseCell[0], analysis.WorstCaseCell[1]))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// analyzeSneakPaths performs sneak path analysis.
func (ca *CrossbarApp) analyzeSneakPaths() {
	// Update mode and educational panel
	ca.modeIndicator.SetMode(DemoModeSneakPath)
	ca.educationalPanel.SetSneakPathExplanation()
	ca.updateStatus("SNEAK | Analyzing parasitic paths...")
	ca.operationLog.Add("Sneak Path: Starting analysis")

	// Select center cell
	selectedRow := ca.config.Rows / 2
	selectedCol := ca.config.Cols / 2

	// Analyze sneak paths
	analysis := ca.array.AnalyzeSneakPaths(selectedRow, selectedCol)

	// Update sneak path heatmap
	sneakMap := analysis.GetSneakMap()
	ca.sneakPathHeatmap.SetData(sneakMap)

	// Highlight selected cell
	ca.sneakPathHeatmap.SetSelection(selectedRow, selectedCol)

	// Calculate signal-to-sneak ratio
	snr := 0.0
	if analysis.TotalSneak > 0 {
		snr = analysis.TotalSignal / analysis.TotalSneak
	}

	// Update stats
	ca.statsPanel.SetStats(fmt.Sprintf(
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
	ca.keyStat.SetValue(fmt.Sprintf("SNR: %.1f:1", snr))
	ca.operationLog.AddWithResult("Sneak Path", fmt.Sprintf("%.1f%% ratio", analysis.MaxSneakRatio*100), analysis.MaxSneakRatio < 0.05)

	ca.updateStatus(fmt.Sprintf("SNEAK | Complete: SNR %.1f:1 at [%d,%d]",
		snr, selectedRow, selectedCol))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// resetArray resets the array with new random weights.
func (ca *CrossbarApp) resetArray() {
	ca.modeIndicator.SetMode(DemoModeWrite)
	ca.updateStatus("WRITE | Programming random weights...")
	ca.operationLog.Add("Reset: Programming new weights")

	ca.programRandomWeights()
	ca.updateConductanceDisplay()
	ca.lastInput = nil
	ca.lastOutput = nil
	ca.ioDisplay.SetInput(nil)
	ca.ioDisplay.SetOutput(nil)
	ca.conductanceHeatmap.ClearSelection()
	ca.irDropHeatmap.ClearSelection()
	ca.sneakPathHeatmap.ClearSelection()
	ca.statsPanel.SetStats("Array reset with new random weights.\n\nSelect a cell or run an analysis.")

	// Update key stat
	ca.keyStat.SetValue(fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols))

	ca.operationLog.AddWithResult("Reset", fmt.Sprintf("%dx%d array", ca.config.Rows, ca.config.Cols), true)
	ca.educationalPanel.SetIdleExplanation()
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
		ca.operationLog.Add("Mode: Step-by-Step (manual)")
		ca.educationalPanel.SetContent("Step-by-Step Mode",
			"Click each button to see\nthe operation explained.\n\n"+
				"Recommended order:\n"+
				"1. Run MVM\n"+
				"2. Analyze IR Drop\n"+
				"3. Analyze Sneak Paths\n"+
				"4. Reset Array")
	case "Manual":
		ca.operationLog.Add("Mode: Manual")
		ca.educationalPanel.SetIdleExplanation()
	}
}

// startAutoDemoLoop starts the automatic demo loop.
func (ca *CrossbarApp) startAutoDemoLoop() {
	ca.autoDemo = true
	ca.autoDemoStep = 0
	ca.stopAutoDemo = make(chan bool)
	ca.autoDemoTimer = time.NewTicker(3 * time.Second)

	ca.operationLog.Add("Mode: Auto Demo started")
	ca.educationalPanel.SetContent("Auto Demo Mode",
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
		ca.operationLog.Add("Mode: Auto Demo stopped")
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
