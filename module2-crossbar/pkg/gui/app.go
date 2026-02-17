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

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/shared/logging"
	sharedtheme "fecim-lattice-tools/shared/theme"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Package-level logger using shared logging infrastructure
// Lazy-initialized to ensure it's created after EnableFileLogging() is called
var debug *logging.Logger
var debugOnce sync.Once

func getDebug() *logging.Logger {
	debugOnce.Do(func() {
		debug = logging.NewLogger("crossbar-app")
	})
	return debug
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
	resetButton       *widget.Button
	arraySizeSelect   *widget.Select // Dropdown for array size
	arraySizeLabel    *widget.Label  // Label for slider display
	arraySizeSlider   *widget.Slider // Slider for array size
	noiseLabel        *widget.Label
	noiseSlider       *widget.Slider
	adcBitsLabel      *widget.Label
	adcBitsSlider     *widget.Slider
	temperatureLabel  *widget.Label
	temperatureSlider *widget.Slider
	colormapSelect    *widget.Select
	statsLabel        *widget.Label

	// Track colormap per tab
	condColormap  string
	irColormap    string
	sneakColormap string

	// Architecture toggle (Dr. Tour: clarify 0T1R vs 1T1R vs 2T1R)
	archToggle     *fyne.Container // Container with toggle buttons
	archPassiveBtn *widget.Button
	arch1T1RBtn    *widget.Button
	arch2T1RBtn    *widget.Button
	architecture   string // "1T1R (Transistor)", "0T1R (Passive)", or "2T1R (Dual Transistor)"

	// Status
	statusLabel    *widget.Label
	statusBar      *sharedwidgets.StatusBar
	infoLabel      *widget.Label
	hoverInfoLabel *widget.Label

	// Mutex for protecting state accessed by multiple goroutines
	// Protects: lastInput, lastOutput, lastMVMResult, lastIRDropAnalysis,
	// lastSneakAnalysis, selectedRow, selectedCol, temperatureK
	stateMu sync.RWMutex

	// Current state (protected by stateMu)
	lastInput          []float64
	lastOutput         []float64
	lastMVMResult      *crossbar.MVMResult
	lastIRDropAnalysis *crossbar.IRDropAnalysis
	lastSneakAnalysis  *crossbar.SneakPathAnalysis
	temperatureK       float64

	// Baseline values from 0T1R (passive) for legend scaling
	// These provide consistent reference for comparing architectures
	baselineMaxIRDrop float64 // Max IR drop % from passive array
	baselineMaxSneak  float64 // Max sneak ratio from passive array

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

	// M5 UX fix: Track if MVM is running to disable controls during animation
	isMVMRunning bool

	// Responsive layout support
	responsiveDetector *sharedwidgets.ResponsiveDetector
	leftCenterSplit    *container.Split
	mainSplit          *container.Split
	currentBreakpoint  sharedwidgets.Breakpoint

	// Tab badge state for accessibility/discoverability (C2 fix)
	// Tracks which tabs have new/unseen data
	tabHasNewData map[string]bool
}

// NewCrossbarApp creates and initializes the crossbar demo application.
// Returns an error if the crossbar array cannot be created.
func NewCrossbarApp() (*CrossbarApp, error) {
	ca := &CrossbarApp{
		tabHasNewData: make(map[string]bool),
		temperatureK:  300.0,
	}

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
	getDebug().Println("App: Creating window")
	windowTitle := "FeCIM Demo 2: Crossbar Array MVM"
	if enhanced {
		windowTitle += " (Enhanced)"
	}
	ca.window = ca.fyneApp.NewWindow(windowTitle)
	ca.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	getDebug().Println("App: Creating main layout")
	var content fyne.CanvasObject
	if enhanced {
		content = ca.createEnhancedMainLayout()
	} else {
		content = ca.createMainLayout()
	}
	getDebug().Println("App: Setting window content")
	ca.window.SetContent(content)

	// Initialize displays
	getDebug().Println("App: Updating conductance display")
	ca.updateConductanceDisplay()
	getDebug().Println("App: Updating status")
	ca.updateStatus("Ready | Array initialized with random weights. Click 'Run MVM' to start!")

	// Setup keyboard shortcuts
	ca.setupKeyboard()

	// m5 UX fix: Set first-load onboarding content
	ca.setEducationalContent("Getting Started",
		"Welcome to Crossbar MVM!\n\n"+
			"Quick Start:\n"+
			"1. Hover over cells to see\n"+
			"   conductance values\n"+
			"2. Click cells for details\n"+
			"3. Adjust controls on right\n"+
			"4. Explore IR Drop & Sneak\n"+
			"   Path analysis tabs\n\n"+
			"Key Concepts:\n"+
			"• 30 analog levels/cell (baseline; conference claim)\n"+
			"• MVM = Matrix × Vector\n"+
			"• All rows compute in 1 step\n\n"+
			"Try: Switch tabs to see\n"+
			"different non-idealities!\n\n"+
			"Press ? for keyboard shortcuts")

	getDebug().Println("App: ShowAndRun starting")
	ca.window.ShowAndRun()
	getDebug().Println("App: Window closed")
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

	// Create color legends for each heatmap
	// Conductance displayed as discrete level (0-29) per FeCIM 30-level spec
	ca.condLegend = sharedwidgets.NewColorLegendWithColormap(0, 29, "Level", true, "fecim")
	ca.irLegend = sharedwidgets.NewColorLegendWithColormap(0, 10, "%", true, "viridis")    // Typical IR drop range ~1-10%
	ca.sneakLegend = sharedwidgets.NewColorLegendWithColormap(0, 100, "%", true, "plasma") // Sneak ratio: 0-100% of signal

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
	ca.resetButton = widget.NewButton("Reset", ca.resetArray)
	ca.resetButton.Importance = widget.MediumImportance

	// Array size slider - create without callback to avoid triggering
	// recreateArray before UI is initialized
	ca.arraySizeLabel = widget.NewLabel("Array Size: 64×64")
	ca.arraySizeSlider = widget.NewSlider(8, 128)
	ca.arraySizeSlider.Step = 8
	ca.arraySizeSlider.Value = 64
	ca.arraySizeSlider.OnChanged = func(v float64) {
		size := int(v)
		ca.arraySizeLabel.SetText(fmt.Sprintf("Array Size: %d×%d", size, size))
		ca.recreateArray(size, ca.config.NoiseLevel, ca.config.ADCBits)
	}

	ca.noiseLabel = widget.NewLabel("2.0%")
	ca.noiseSlider = widget.NewSlider(0, 50)
	ca.noiseSlider.Step = 0.5
	ca.noiseSlider.Value = 2
	ca.noiseSlider.OnChanged = func(v float64) {
		ca.noiseLabel.SetText(fmt.Sprintf("%.1f%%", v))
		ca.config.NoiseLevel = v / 100.0
		ca.runEnhancedMVMInstant()
	}

	ca.adcBitsLabel = widget.NewLabel("6")
	ca.adcBitsSlider = widget.NewSlider(4, 10)
	ca.adcBitsSlider.Step = 1
	ca.adcBitsSlider.Value = 6
	ca.adcBitsSlider.OnChanged = func(v float64) {
		bits := int(v)
		ca.adcBitsLabel.SetText(fmt.Sprintf("ADC Bits: %d", bits))
		ca.config.ADCBits = bits
		ca.runEnhancedMVMInstant()
	}

	// Temperature slider (Kelvin)
	ca.temperatureLabel = widget.NewLabel(ca.formatTemperatureLabel(ca.currentTemperatureK()))
	ca.temperatureLabel.Wrapping = fyne.TextWrapOff
	ca.temperatureSlider = widget.NewSlider(77, 450)
	ca.temperatureSlider.Step = 5
	ca.temperatureSlider.Value = ca.currentTemperatureK()
	ca.temperatureSlider.OnChanged = func(v float64) {
		ca.temperatureLabel.SetText(ca.formatTemperatureLabel(v))
		ca.setTemperatureK(v)
		ca.stateMu.Lock()
		ca.baselineMaxIRDrop = 0
		ca.stateMu.Unlock()
		ca.updateInfoLabel()
		ca.runEnhancedMVMInstant()
	}

	ca.colormapSelect = widget.NewSelect([]string{"fecim", "viridis", "plasma", "coolwarm"}, func(s string) {
		ca.conductanceHeatmap.SetColormap(s)
		ca.condLegend.SetColormap(s)
	})
	ca.colormapSelect.SetSelected("fecim")

	// Architecture toggle: 0T1R (passive) vs 1T1R (with access transistor) vs 2T1R (dual transistor)
	// Dr. Tour recommendation: clarify sneak path behavior depends on architecture
	ca.architecture = sharedwidgets.Architecture0T1R // Default to passive

	// Create toggle buttons
	ca.archPassiveBtn = widget.NewButton("PASSIVE", nil)
	ca.arch1T1RBtn = widget.NewButton("1T1R", nil)
	ca.arch2T1RBtn = widget.NewButton("2T1R", nil)

	// Helper to update button styles based on selection
	updateArchButtons := func() {
		ca.archPassiveBtn.Importance = widget.LowImportance
		ca.arch1T1RBtn.Importance = widget.LowImportance
		ca.arch2T1RBtn.Importance = widget.LowImportance
		switch ca.architecture {
		case sharedwidgets.Architecture0T1R:
			ca.archPassiveBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture1T1R:
			ca.arch1T1RBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture2T1R:
			ca.arch2T1RBtn.Importance = widget.HighImportance
		default:
			ca.archPassiveBtn.Importance = widget.HighImportance
		}
		ca.archPassiveBtn.Refresh()
		ca.arch1T1RBtn.Refresh()
		ca.arch2T1RBtn.Refresh()
	}

	// Set initial state
	updateArchButtons()

	// Wire up callbacks
	ca.archPassiveBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			return // Already selected
		}
		ca.stateMu.Lock()
		ca.architecture = sharedwidgets.Architecture0T1R
		ca.stateMu.Unlock()
		updateArchButtons()
		title, content := sharedwidgets.ArchitectureInfo(sharedwidgets.Architecture0T1R)
		ca.setEducationalContent(title, content)
		ca.runEnhancedMVMWithCurrentInput()
	}

	ca.arch1T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture1T1R {
			return // Already selected
		}
		ca.stateMu.Lock()
		ca.architecture = sharedwidgets.Architecture1T1R
		ca.stateMu.Unlock()
		updateArchButtons()
		title, content := sharedwidgets.ArchitectureInfo(sharedwidgets.Architecture1T1R)
		ca.setEducationalContent(title, content)
		ca.runEnhancedMVMWithCurrentInput()
	}

	ca.arch2T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture2T1R {
			return // Already selected
		}
		ca.stateMu.Lock()
		ca.architecture = sharedwidgets.Architecture2T1R
		ca.stateMu.Unlock()
		updateArchButtons()
		title, content := sharedwidgets.ArchitectureInfo(sharedwidgets.Architecture2T1R)
		ca.setEducationalContent(title, content)
		ca.runEnhancedMVMWithCurrentInput()
	}

	// Create horizontal container for toggle (3 buttons)
	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)

	ca.statsLabel = widget.NewLabel("Analysis Results\n\nNo data yet.\nClick Run MVM to start.")
	ca.statsLabel.Wrapping = fyne.TextWrapOff
	ca.statsLabel.TextStyle = fyne.TextStyle{Monospace: true} // Fixed-width prevents resize

	// Create status labels
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	ca.statusBar = sharedwidgets.NewStatusBarWithLabel(ca.statusLabel, "Status: ")

	ca.infoLabel = widget.NewLabel("")
	ca.updateInfoLabel()

	// Hover info label - shows cell info on mouse hover
	ca.hoverInfoLabel = widget.NewLabel("Hover over cells to see values")
	ca.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}
	ca.hoverInfoLabel.Wrapping = fyne.TextWrapOff
	ca.hoverInfoLabel.Truncation = fyne.TextTruncateEllipsis

	// Create tabbed heatmap view with legends on the right
	conductanceWithLegend := container.NewBorder(nil, nil, nil, ca.condLegend, ca.conductanceHeatmap)
	irDropWithLegend := container.NewBorder(nil, nil, nil, ca.irLegend, ca.irDropHeatmap)
	sneakPathWithLegend := container.NewBorder(nil, nil, nil, ca.sneakLegend, ca.sneakPathHeatmap)

	ca.tabs = container.NewAppTabs(
		container.NewTabItem("Conductance", conductanceWithLegend),
		container.NewTabItem("IR Drop", irDropWithLegend),
		container.NewTabItem("Sneak Paths", sneakPathWithLegend),
		container.NewTabItem("Input/Output", container.NewMax(ca.mvmVis)),
	)

	// Update educational panel based on selected tab
	ca.tabs.OnSelected = func(tab *container.TabItem) {
		// Clear badge when tab is viewed (C2 accessibility fix)
		baseName := ca.getBaseTabName(tab.Text)
		ca.clearTabBadge(baseName)

		switch baseName {
		case "Conductance":
			ca.setEducationalContent("Conductance Matrix",
				"Each cell = one FeFET\n\n"+
					"Conductance G (1-100 µS)\n"+
					"stored as 30 discrete levels (claim)\n"+
					"(conference claim baseline)\n\n"+
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
	aboutScienceBtn := sharedwidgets.CreateAboutScienceButton(ca.window)
	aboutScienceBtn.Importance = widget.LowImportance

	header := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), titleLabel, layout.NewSpacer(), aboutScienceBtn),
		widget.NewSeparator(),
	)

	// Right panel - improved layout with better spacing and grouping

	// Export button
	exportButton := widget.NewButton("Export", func() { ca.exportData() })
	exportButton.Importance = widget.MediumImportance

	// Array size row - slider with label
	arraySizeRow := container.NewBorder(nil, nil, widget.NewLabel("Array:"), ca.arraySizeLabel, ca.arraySizeSlider)

	// Architecture toggle
	archLabel := widget.NewLabelWithStyle("Architecture", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Signal quality - inline labels
	noiseTooltip := sharedwidgets.TooltipContent{
		Title:       "Read/Write Noise",
		Description: "Simulates random variations in conductance programming and read operations.",
		Physics:     "Real analog hardware suffers from thermal noise and device variability. This parameter adds Gaussian noise (σ) to write and read operations to test network robustness.",
		Range:       "0% - 50% (Typical: 1-5%)",
	}
	noiseRow := sharedwidgets.SliderWithTooltip("Noise:", ca.noiseSlider, ca.noiseLabel, noiseTooltip, ca.window)
	adcRow := container.NewBorder(nil, nil, widget.NewLabel("ADC:"), ca.adcBitsLabel, ca.adcBitsSlider)
	tempRow := container.NewBorder(nil, nil, widget.NewLabel("Temp:"), ca.temperatureLabel, ca.temperatureSlider)

	// Colormap row
	colormapRow := container.NewBorder(nil, nil, widget.NewLabel("Color:"), nil, ca.colormapSelect)

	// Action buttons row
	actionButtons := container.NewGridWithColumns(2, ca.resetButton, exportButton)

	// M4 UX fix: Remove scroll from controls section to avoid nested scroll issues
	// Controls are fixed-height, only stats need scrolling
	controlsBox := container.NewVBox(
		arraySizeRow,
		widget.NewSeparator(),
		archLabel,
		ca.archToggle,
		widget.NewSeparator(),
		noiseRow,
		adcRow,
		tempRow,
		widget.NewSeparator(),
		colormapRow,
		widget.NewSeparator(),
		actionButtons,
	)

	// Stats section with header (scrollable)
	statsHeader := widget.NewLabelWithStyle("Analysis Results", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	statsSection := container.NewBorder(
		container.NewVBox(widget.NewSeparator(), statsHeader),
		nil, nil, nil,
		ca.statsLabel,
	)
	statsScroll := container.NewVScroll(statsSection)
	statsScroll.SetMinSize(fyne.NewSize(240, 120))

	// Use Border layout: controls at top (fixed), stats fills rest (scrollable)
	// This eliminates nested scroll containers that compete for scroll events
	rightPanel := container.NewBorder(
		controlsBox, // top - fixed controls
		nil,         // bottom
		nil,         // left
		nil,         // right
		statsScroll, // center - scrollable stats
	)

	// Left panel using simple labels (no custom widgets)
	// M2 UX fix: Wrap educational content in fixed-size container to prevent layout shifts
	// when content changes between tabs (BUG-M2-004 mitigation)
	eduContentWrapper := container.NewGridWrap(fyne.NewSize(200, 280), ca.eduContentLabel)
	leftPanelContent := container.NewVBox(
		ca.eduTitleLabel,
		widget.NewSeparator(),
		eduContentWrapper,
		widget.NewSeparator(),
		ca.keyStatLabel,
		ca.keyStatValue,
	)
	leftPanel := container.NewVScroll(leftPanelContent)

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
		ca.condLegend.SetColormap(colormap)
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

	// Reset baseline values so they get recomputed for new array size
	ca.stateMu.Lock()
	ca.baselineMaxIRDrop = 0
	ca.baselineMaxSneak = 0
	ca.selectedRow = -1 // Reset selection - old indices may be out of bounds
	ca.selectedCol = -1
	ca.stateMu.Unlock()

	// Resize existing heatmaps instead of creating new ones
	// This preserves the widget references in the window layout
	ca.conductanceHeatmap.SetDimensions(size, size)
	ca.irDropHeatmap.SetDimensions(size, size)
	ca.sneakPathHeatmap.SetDimensions(size, size)

	// Resize the Ideal vs Actual comparison widget if it exists
	if ca.beforeAfterToggle != nil {
		ca.beforeAfterToggle.SetDimensions(size, size)
	}

	ca.programRandomWeights()
	ca.updateConductanceDisplay()
	ca.updateInfoLabel()
	ca.setKeyStatValue(fmt.Sprintf("%d MACs", size*size))
	ca.updateStatus(fmt.Sprintf("Array resized to %dx%d (%d parallel MACs)", size, size, size*size))

	// Auto-run MVM after resize
	ca.runEnhancedMVMInstant()
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

	// Conductance values are normalized [0,1] mapping to 30 discrete levels (0-29)
	// Legend shows the level range (0-29) per FeCIM spec
	fyne.Do(func() {
		ca.conductanceHeatmap.SetData(matrix)
		ca.condLegend.SetRange(0, float64(crossbar.DefaultQuantizationLevels-1))
	})
}

// updateStatus updates the status label.
func (ca *CrossbarApp) updateStatus(status string) {
	sharedwidgets.EnsureStatusBar(&ca.statusBar, ca.statusLabel, "Status: ", status)
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

func (ca *CrossbarApp) currentTemperatureK() float64 {
	ca.stateMu.RLock()
	tempK := ca.temperatureK
	ca.stateMu.RUnlock()
	if tempK <= 0 {
		tempK = 300.0
	}
	return tempK
}

func (ca *CrossbarApp) setTemperatureK(tempK float64) {
	if tempK <= 0 {
		tempK = 300.0
	}
	ca.stateMu.Lock()
	ca.temperatureK = tempK
	ca.stateMu.Unlock()
}

func (ca *CrossbarApp) formatTemperatureLabel(tempK float64) string {
	tempC := tempK - 273.15
	return fmt.Sprintf("%.0f K (%.0f°C)", tempK, tempC)
}

func (ca *CrossbarApp) applyTemperatureToWireParams(params *crossbar.WireParams) {
	if params == nil {
		return
	}
	tempK := ca.currentTemperatureK()
	if tempK != 300.0 {
		tempEffects := crossbar.NewTemperatureEffects(tempK)
		params.RwordLine = tempEffects.AdjustedWireResistance(params.RwordLine)
		params.RbitLine = tempEffects.AdjustedWireResistance(params.RbitLine)
	}
}

// updateInfoLabel updates the info label with current config.
func (ca *CrossbarApp) updateInfoLabel() {
	tempK := ca.currentTemperatureK()
	ca.infoLabel.SetText(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 (claim) | Noise: %.1f%% | ADC: %d bits | Temp: %.0fK",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits, tempK,
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

// Tab badge constants for accessibility (C2 fix)
const (
	tabBadgeNew = " ●" // Indicator for tabs with new/unseen data
)

// baseTabNames maps tab indices to their base names (without badges)
var baseTabNames = []string{"Conductance", "IR Drop", "Sneak Paths", "Input/Output"}

// setTabBadge marks a tab as having new data (adds visual indicator).
// This improves discoverability by showing users when analysis tabs have been updated.
func (ca *CrossbarApp) setTabBadge(tabName string) {
	if ca.tabs == nil {
		return
	}

	ca.tabHasNewData[tabName] = true

	fyne.Do(func() {
		for _, tab := range ca.tabs.Items {
			// Find the tab by checking if its text starts with the base name
			baseName := ca.getBaseTabName(tab.Text)
			if baseName == tabName && !ca.hasBadge(tab.Text) {
				tab.Text = tabName + tabBadgeNew
				ca.tabs.Refresh()
				break
			}
		}
	})
}

// clearTabBadge removes the new data indicator from a tab.
func (ca *CrossbarApp) clearTabBadge(tabName string) {
	if ca.tabs == nil {
		return
	}

	baseName := ca.getBaseTabName(tabName)
	ca.tabHasNewData[baseName] = false

	fyne.Do(func() {
		for _, tab := range ca.tabs.Items {
			if ca.getBaseTabName(tab.Text) == baseName {
				tab.Text = baseName
				ca.tabs.Refresh()
				break
			}
		}
	})
}

// getBaseTabName strips any badge suffix from a tab name.
func (ca *CrossbarApp) getBaseTabName(tabText string) string {
	for _, base := range baseTabNames {
		if len(tabText) >= len(base) && tabText[:len(base)] == base {
			return base
		}
	}
	// For enhanced mode tabs
	enhancedTabs := []string{"Ideal vs Actual", "Accuracy Analysis"}
	for _, base := range enhancedTabs {
		if len(tabText) >= len(base) && tabText[:len(base)] == base {
			return base
		}
	}
	return tabText
}

// hasBadge checks if a tab name already has a badge.
func (ca *CrossbarApp) hasBadge(tabText string) bool {
	return len(tabText) > len(tabBadgeNew) && tabText[len(tabText)-len(tabBadgeNew):] == tabBadgeNew
}
