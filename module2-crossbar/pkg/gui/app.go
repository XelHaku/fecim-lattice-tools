// Package gui provides Fyne-based GUI components for crossbar visualization.
package gui

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
	"multilayer-ferroelectric-cim-visualizer/shared/logging"
	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// Package-level logger using shared logging infrastructure
var debug *logging.Logger

func init() {
	debug = logging.NewLogger("crossbar-app")
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

	// Color legends for each heatmap
	condLegend  *sharedwidgets.ColorLegend
	irLegend    *sharedwidgets.ColorLegend
	sneakLegend *sharedwidgets.ColorLegend

	controlPanel   *ControlPanel
	statsPanel     *StatsPanel
	levelIndicator *LevelIndicator

	// Enhanced widgets
	metricsPanel      *MetricsPanel
	comparisonBadge   *ComparisonBadge
	accuracyWaterfall *AccuracyWaterfall
	beforeAfterToggle *BeforeAfterToggle

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

	// Track colormap per tab
	condColormap  string
	irColormap    string
	sneakColormap string

	// Architecture selector (Dr. Tour: clarify 0T1R vs 1T1R)
	archSelect   *widget.Select
	architecture string // "1T1R (Transistor)" or "0T1R (Passive)"

	// Status
	statusLabel    *widget.Label
	infoLabel      *widget.Label
	hoverInfoLabel *widget.Label

	// Mutex for protecting state accessed by multiple goroutines
	// Protects: lastInput, lastOutput, lastMVMResult, lastIRDropAnalysis,
	// lastSneakAnalysis, selectedRow, selectedCol
	stateMu sync.RWMutex

	// Current state (protected by stateMu)
	lastInput          []float64
	lastOutput         []float64
	lastMVMResult      *crossbar.MVMResult
	lastIRDropAnalysis *crossbar.IRDropAnalysis
	lastSneakAnalysis  *crossbar.SneakPathAnalysis

	// Persistent cell selection (synced across all heatmaps, protected by stateMu)
	selectedRow int
	selectedCol int

	// Vector visualization
	mvmVis *MVMVisualization

	// Tabs container for programmatic switching
	tabs *container.AppTabs

	// Auto demo cancellation context (replaces chan bool)
	autoCtx    context.Context
	autoCancel context.CancelFunc

	// Auto demo state
	autoDemo      bool
	autoDemoStep  int
	autoDemoTimer *time.Ticker

	// First visit flag for auto-run MVM
	hasRunInitialMVM bool

	// Responsive layout support
	responsiveDetector  *sharedwidgets.ResponsiveDetector
	leftCenterSplit     *container.Split
	mainSplit           *container.Split
	currentBreakpoint   sharedwidgets.Breakpoint
}

// NewCrossbarApp creates and initializes the crossbar demo application.
// Returns an error if the crossbar array cannot be created.
func NewCrossbarApp() (*CrossbarApp, error) {
	ca := &CrossbarApp{}

	// Create Fyne app
	ca.fyneApp = app.NewWithID("com.fecim.crossbar-demo")
	ca.fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

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
		return nil, fmt.Errorf("failed to create crossbar array: %w", err)
	}

	// Program initial random weights
	ca.programRandomWeights()

	return ca, nil
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
	ca.updateStatus("Ready | Array initialized with random weights. Click 'Run MVM' to start!")

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
	ca.eduContentLabel = widget.NewLabel("Matrix-Vector Multiplication\n" +
		"using FeFET crossbar arrays.\n\n" +
		"Click 'Run MVM' to see the\n" +
		"computation in action!\n\n" +
		"The array computes I = W × V\n" +
		"using physics (Ohm's Law)\n" +
		"instead of digital logic.\n\n" +
		"All operations happen in\n" +
		"parallel - no sequential ALU!")
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

	// Architecture selector: 0T1R (passive) vs 1T1R (with access transistor)
	// Dr. Tour recommendation: clarify sneak path behavior depends on architecture
	ca.archSelect = widget.NewSelect([]string{"1T1R (Transistor)", "0T1R (Passive)"}, func(s string) {
		ca.stateMu.Lock()
		ca.architecture = s
		ca.stateMu.Unlock()
		// Update educational content based on architecture
		if s == "1T1R (Transistor)" {
			ca.setEducationalContent("1T1R Architecture",
				"1T1R = One Transistor per FeFET\n\n"+
					"How it works:\n"+
					"Transistor acts as controlled\n"+
					"switch, isolating unselected cells.\n\n"+
					"Advantages:\n"+
					"✓ Zero sneak paths\n"+
					"✓ Linear I-V characteristics\n"+
					"✓ Industry standard (SRAM-like)\n\n"+
					"Tradeoffs:\n"+
					"✗ 50% area overhead\n"+
					"✗ More complex fabrication\n\n"+
					"Best for: High-precision inference\n"+
					"(vision, language models)")
		} else {
			ca.setEducationalContent("0T1R Architecture",
				"0T1R = Passive Crossbar (no transistor)\n\n"+
					"How it works:\n"+
					"Direct connection between wires.\n"+
					"FeFET is the only device.\n\n"+
					"Advantages:\n"+
					"✓ Highest density (4F² per cell)\n"+
					"✓ Simpler fabrication\n"+
					"✓ Lower cost\n\n"+
					"Tradeoffs:\n"+
					"✗ Sneak paths (2-15% SNR loss)\n"+
					"✗ Requires selector device OR\n"+
					"    self-rectifying FeFET\n\n"+
					"FeFET advantage: Natural\n"+
					"rectification in HfO₂-ZrO₂!")
		}
	})
	ca.archSelect.SetSelected("1T1R (Transistor)")

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
	ca.hoverInfoLabel.Wrapping = fyne.TextWrapOff
	ca.hoverInfoLabel.Truncation = fyne.TextTruncateEllipsis

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
					"Conductance G (1-100 µS)\n"+
					"stored as 30 discrete levels\n"+
					"(~4.9 bits per cell)\n\n"+
					"This is your weight matrix W.\n"+
					"Brighter = higher conductance\n\n"+
					"Click any cell to read its\n"+
					"exact conductance value.")
		case "IR Drop":
			ca.setEducationalContent("IR Drop Analysis",
				"Wire resistance causes voltage\n"+
					"drops along metal lines.\n\n"+
					"Red = high voltage drop (>5%)\n"+
					"Blue = low drop (<1%)\n\n"+
					"Impact: Cells far from drivers\n"+
					"compute with reduced voltage,\n"+
					"causing accuracy degradation.\n\n"+
					"Mitigation: Multiple distributed\n"+
					"voltage drivers.\n\n"+
					"Auto-computed after MVM.")
		case "Sneak Paths":
			ca.setEducationalContent("Sneak Path Analysis",
				"Unintended current paths\n"+
					"through passive crossbars.\n\n"+
					"Red = high parasitic current\n"+
					"Blue = minimal leakage\n\n"+
					"Impact: Reduces SNR, especially\n"+
					"in large arrays (>128x128).\n\n"+
					"Mitigation:\n"+
					"• 1T1R architecture (transistor)\n"+
					"• Selector devices (diode)\n"+
					"• Self-rectifying FeFETs\n\n"+
					"Auto-computed after MVM.")
		case "Input/Output":
			ca.setEducationalContent("MVM Vectors",
				"Matrix-Vector Multiplication\n"+
					"in a single analog step.\n\n"+
					"Top: Input voltages V (DAC)\n"+
					"Bottom: Output currents I (ADC)\n\n"+
					"Physics: I = G × V (Ohm's Law)\n"+
					"Result: I_row = Σ(G_ij × V_j)\n\n"+
					"All N² multiply-accumulate\n"+
					"operations happen in parallel\n"+
					"in ~1ns (speed of light).\n\n"+
					"Click 'Run MVM' to see it!")
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

	// Architecture settings group (Dr. Tour: clarify 0T1R vs 1T1R)
	archLabel := widget.NewLabelWithStyle("Architecture", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	archGroup := container.NewVBox(
		widget.NewSeparator(),
		archLabel,
		ca.archSelect,
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
		archGroup,
		signalGroup,
		displayGroup,
	)
	controlsScroll := container.NewVScroll(controlsBox)
	controlsScroll.SetMinSize(fyne.NewSize(240, 250))

	// Stats section with header
	statsHeader := widget.NewLabelWithStyle("Analysis Results", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	statsSection := container.NewBorder(
		container.NewVBox(widget.NewSeparator(), statsHeader),
		nil, nil, nil,
		ca.statsLabel,
	)
	statsScroll := container.NewVScroll(statsSection)
	statsScroll.SetMinSize(fyne.NewSize(240, 120))

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
	// Wrap hoverInfoLabel in fixed-size container to prevent layout recalc on text change
	hoverInfoContainer := container.NewGridWrap(fyne.NewSize(450, 20), ca.hoverInfoLabel)
	simpleFooter := container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		hoverInfoContainer,
		widget.NewSeparator(),
		ca.infoLabel,
	)

	// Use HSplit for proportional 3-column layout
	// Left panel (15%) | Center tabs (70%) | Right panel (15%)
	ca.leftCenterSplit = container.NewHSplit(leftPanel, ca.tabs)
	ca.leftCenterSplit.SetOffset(0.15) // 15% left, 85% center+right

	ca.mainSplit = container.NewHSplit(ca.leftCenterSplit, rightPanel)
	ca.mainSplit.SetOffset(0.8) // 80% left+center, 20% right

	// Create responsive detector for breakpoint-based layout adjustments
	ca.responsiveDetector = sharedwidgets.NewResponsiveDetector(ca.onBreakpointChange)
	ca.currentBreakpoint = sharedwidgets.BreakpointXL // Default to desktop

	// Wrap with header and footer
	mainContent := container.NewBorder(
		header,       // top
		simpleFooter, // bottom
		nil,          // left
		nil,          // right
		ca.mainSplit, // center - the split panels
	)

	// Stack with responsive detector overlay
	return container.NewStack(mainContent, ca.responsiveDetector)
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
		ca.updateStatus("Error creating array")
		if ca.window != nil {
			dialog.ShowError(fmt.Errorf("failed to create crossbar array: %w", err), ca.window)
		}
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
	ca.updateStatus(fmt.Sprintf("Array resized to %dx%d (%d parallel MACs)", size, size, size*size))
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
	fyne.Do(func() {
		ca.conductanceHeatmap.SetData(matrix)
	})
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

// onBreakpointChange handles responsive layout adjustments.
func (ca *CrossbarApp) onBreakpointChange(bp sharedwidgets.Breakpoint, size fyne.Size) {
	ca.currentBreakpoint = bp

	// Adjust split offsets based on breakpoint
	switch bp {
	case sharedwidgets.BreakpointSM, sharedwidgets.BreakpointMD:
		// Small/Medium: Minimize side panels, maximize heatmap area
		// Left panel: collapse to 5% (minimal educational info)
		// Right panel: collapse to 10% (minimal controls)
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.05) // 5% left, 95% center+right
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.9) // 90% left+center, 10% right
		}

	case sharedwidgets.BreakpointLG:
		// Large: Balanced layout for laptops
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.12) // 12% left
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.85) // 85% left+center, 15% right
		}

	case sharedwidgets.BreakpointXL:
		// Extra Large: Desktop - original comfortable layout
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.15) // 15% left
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.8) // 80% left+center, 20% right
		}
	}
}
