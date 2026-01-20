// Package gui provides Fyne-based GUI components for crossbar visualization.
package gui

import (
	"fmt"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
)

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
	ca.window = ca.fyneApp.NewWindow("IronLattice Demo 2: Crossbar Array MVM")
	ca.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	content := ca.createMainLayout()
	ca.window.SetContent(content)

	// Initialize displays
	ca.updateConductanceDisplay()
	ca.updateStatus("Ready. Program weights and run MVM operations.")

	ca.window.ShowAndRun()
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

	// Create tabbed heatmap view
	tabs := container.NewAppTabs(
		container.NewTabItem("Conductance", container.NewPadded(ca.conductanceHeatmap)),
		container.NewTabItem("IR Drop", container.NewPadded(ca.irDropHeatmap)),
		container.NewTabItem("Sneak Paths", container.NewPadded(ca.sneakPathHeatmap)),
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

	// Footer with status bar showing mode and current action
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			ca.modeIndicator,
			widget.NewSeparator(),
			ca.statusLabel,
			layout.NewSpacer(),
			ca.infoLabel,
		),
	)

	// Right panel (controls + I/O display + stats)
	rightPanel := container.NewVBox(
		ca.controlPanel,
		widget.NewSeparator(),
		ca.ioDisplay,
		widget.NewSeparator(),
		ca.statsPanel,
		widget.NewSeparator(),
		ca.levelIndicator,
	)

	// Left panel (educational + log + key stat)
	leftPanel := container.NewVBox(
		ca.educationalPanel,
		widget.NewSeparator(),
		ca.operationLog,
		widget.NewSeparator(),
		ca.keyStat,
	)

	// Main content with left panel for educational content
	mainContent := container.NewBorder(
		header,                           // top
		footer,                           // bottom
		container.NewPadded(leftPanel),   // left
		container.NewPadded(rightPanel),  // right
		tabs,                             // center
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

	// Recreate heatmaps
	ca.conductanceHeatmap = NewCrossbarHeatmap(size, size)
	ca.conductanceHeatmap.SetColormap("ironlattice")
	ca.conductanceHeatmap.OnCellTapped = ca.onCellTapped

	ca.irDropHeatmap = NewCrossbarHeatmap(size, size)
	ca.irDropHeatmap.SetColormap("coolwarm")

	ca.sneakPathHeatmap = NewCrossbarHeatmap(size, size)
	ca.sneakPathHeatmap.SetColormap("plasma")

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

	ca.updateStatus(fmt.Sprintf("Cell [%d,%d] selected - Level %d", row, col, level))
}

// runMVM performs matrix-vector multiplication.
func (ca *CrossbarApp) runMVM() {
	ca.updateStatus("Running MVM operation...")

	// Create random input
	input := make([]float64, ca.config.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}
	ca.lastInput = input

	// Perform MVM
	output, err := ca.array.MVM(input)
	if err != nil {
		ca.updateStatus(fmt.Sprintf("MVM Error: %v", err))
		return
	}
	ca.lastOutput = output

	// Update stats
	var sumInput, sumOutput float64
	for _, v := range input {
		sumInput += v
	}
	for _, v := range output {
		sumOutput += v
	}

	reads, writes := ca.array.GetStats()

	ca.statsPanel.SetStats(fmt.Sprintf(
		"MVM Complete!\n"+
		"Input Sum: %.4f\n"+
		"Output Sum: %.4f\n"+
		"Total Reads: %d\n"+
		"Total Writes: %d\n"+
		"MAC Operations: %d",
		sumInput, sumOutput, reads, writes,
		ca.config.Rows*ca.config.Cols,
	))

	ca.updateStatus("MVM complete - Ready for analysis")
}

// analyzeIRDrop performs IR drop analysis.
func (ca *CrossbarApp) analyzeIRDrop() {
	ca.updateStatus("Analyzing IR drop...")

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

	ca.updateStatus(fmt.Sprintf("IR drop analysis complete - Max drop: %.2f%%", analysis.MaxIRDrop*100))
}

// analyzeSneakPaths performs sneak path analysis.
func (ca *CrossbarApp) analyzeSneakPaths() {
	ca.updateStatus("Analyzing sneak paths...")

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

	ca.updateStatus(fmt.Sprintf("Sneak path analysis complete - Max ratio: %.2f%%", analysis.MaxSneakRatio*100))
}

// resetArray resets the array with new random weights.
func (ca *CrossbarApp) resetArray() {
	ca.programRandomWeights()
	ca.updateConductanceDisplay()
	ca.lastInput = nil
	ca.lastOutput = nil
	ca.conductanceHeatmap.ClearSelection()
	ca.irDropHeatmap.ClearSelection()
	ca.sneakPathHeatmap.ClearSelection()
	ca.statsPanel.SetStats("Array reset with new random weights.\n\nSelect a cell or run an analysis.")
	ca.updateStatus("Array reset with random weights")
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
