//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// Module 4: Peripheral Circuits - Unified 3-view design
// OPERATIONS (Write/Read/Compute modes), COMPARISON (benchmarks), REFERENCE (timing diagrams + specs)
package gui

import (
	"image"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	crossbar "fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/peripherals"
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/presets"
	sharedtheme "fecim-lattice-tools/shared/theme"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Constants
const (
	FeCIMLevels    = physics.DefaultLevels // Demo baseline (conference claim; from shared/physics)
	MaxArraySize   = 128                   // Maximum array dimension
	DefaultSize    = 8                     // Default array size
	DefaultDACBits = 4                     // Default DAC resolution (4-bit optimal per 2024-2025 literature)
	DefaultADCBits = 6                     // Default ADC resolution (6-bit: 64 codes for 30 conductance levels)
)

// CircuitsApp is the main application for the peripheral circuits demo.
type CircuitsApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Peripheral components
	dac  *peripherals.DAC
	adc  *peripherals.ADC
	tia  *peripherals.TIA
	pump *peripherals.ChargePump

	// Configuration state
	mu                 sync.RWMutex
	uiUpdateMu         sync.Mutex
	lastUIUpdate       time.Time
	pendingUIUpd       bool
	uiUpdateTimer      *time.Timer
	suspendUIRefresh   bool
	arrayRows          int
	arrayCols          int
	quantLevels        int
	dacBits            int
	adcBits            int
	vMin               float64 // Min write voltage
	vMax               float64 // Max write voltage
	pulseWidth         float64 // ns
	readVoltage        float64 // Read voltage (safe zone)
	tiaGain            float64 // TIA gain (kOhm)
	selectedRow        int
	selectedCol        int
	targetLevel        int
	arrayWeights       [][]int                      // Current programmed levels
	halfSelectResidue  [][]float64                  // Fractional disturb accumulation for half-selected cells
	inputVector        []int                        // Input vector for compute
	writeDisturbEngine *crossbar.WriteDisturbEngine // Module 2 cumulative stress model for V/2 disturb
	outputVector       []float64
	architecture       string // "1T1R" or "0T1R" - affects row selection behavior

	// Tab-specific GUI components
	// Tab 1: Write
	writeRowSelect    *widget.Select
	writeColSelect    *widget.Select
	writeLevelSlider  *widget.Slider
	writeLevelLabel   *widget.Label
	writeArrayCanvas  *canvas.Raster
	writeDataPath     *fyne.Container
	writeDigitalLabel *widget.Label // Label for digital box value
	writeDACLabel     *widget.Label // Label for DAC box value
	writeFeFETLabel   *widget.Label // Label for FeFET box value
	writePulseCanvas  *canvas.Raster
	writeMappingLabel *widget.Label
	writeStatusLabel  *widget.Label

	// Tab 2: Read
	readRowSelect     *widget.Select
	readColSelect     *widget.Select
	readVoltageSlider *widget.Slider
	readVoltageLabel  *widget.Label
	readDataPath      *fyne.Container
	readZoneCanvas    *canvas.Raster
	readResultsLabel  *widget.Label
	readStatusLabel   *widget.Label
	readCalcLabel     *widget.Label // Added for dynamic calculation display

	// Tab 3: Compute
	computeInputs        []*widget.Entry
	computeVoltageLabels []*widget.Label
	computeArrayCanvas   *canvas.Raster
	computeOutputLabels  []*widget.Label
	computeMathLabel     *widget.Label
	computeStatusLabel   *widget.Label

	// Tab 4: Comparison
	compArchCanvas   *canvas.Raster
	compTimingCanvas *canvas.Raster
	compEnergyCanvas *canvas.Raster
	compTableLabels  []*widget.Label
	compStatusLabel  *widget.Label
	compArraySize    int // Current array size for comparison (8, 16, 32, or 64)

	// Tab 5: Timing
	timingOpSelect      *widget.Select
	timingWriteCanvas   *canvas.Raster
	timingReadCanvas    *canvas.Raster
	timingComputeCanvas *canvas.Raster
	timingStatusLabel   *widget.Label

	// Tab 6: Specs
	specArraySizeSelect  *widget.Select
	specQuantLevelSelect *widget.Select
	specDACBitsSelect    *widget.Select
	specADCBitsSelect    *widget.Select
	specTIAGainSelect    *widget.Select
	specSummaryLabels    []*widget.Label
	specSummaryLabel     *fyne.Container
	specStatusLabel      *widget.Label

	// Tab 7: Reference (unified TIMING + SPECS + VOLTAGE RULES)
	refTimingSection       fyne.CanvasObject
	refSpecsSection        fyne.CanvasObject
	refVoltageRulesSection fyne.CanvasObject

	// Main tabs
	mainTabs *container.AppTabs

	// ============================================================================
	// UNIFIED OPERATIONS VIEW (tab_operations.go)
	// ============================================================================
	currentMode           OperationMode
	operationsStatusLabel *widget.Label
	operationsModeHelp    *widget.Label

	// Shared array section
	sharedArrayCanvas        *canvas.Raster
	sharedCellInfoLabel      *widget.Label
	sharedCellDisplayToggle  *widget.Button
	showCurrentInCellInfo    bool
	sharedArrayInfoLabel     *widget.Label
	sharedArrayCellSize      int // For click detection
	sharedArrayOffsetX       int
	sharedArrayOffsetY       int
	computeInputRowContainer *fyne.Container // Input row shown above array in COMPUTE mode

	// Cached static layer for unified array rendering (redraw on resize/layout changes only)
	unifiedStaticCache       *image.RGBA
	unifiedStaticCacheKey    string
	unifiedStaticRedrawCount int

	// Write mode panel widgets (ops prefix to distinguish from legacy)
	writeConfigPanel     *fyne.Container
	opsWriteLevelSlider  *widget.Slider
	opsWriteLevelLabel   *widget.Label
	opsWriteDigitalLabel *widget.Label
	opsWriteDACLabel     *widget.Label
	opsWriteFeFETLabel   *widget.Label
	opsWritePulseCanvas  *canvas.Raster

	// Read mode panel widgets
	readConfigPanel      *fyne.Container
	opsReadVoltageSlider *widget.Slider
	opsReadVoltageLabel  *widget.Label
	opsReadZoneCanvas    *canvas.Raster
	opsReadResultsLabel  *widget.Label

	// Compute mode panel widgets
	computeConfigPanel      *fyne.Container
	opsComputeInputs        []*widget.Entry
	opsComputeVoltageLabels []*widget.Label
	opsComputeOutputLabels  []*widget.Label
	opsComputeMathLabel     *widget.Label

	// Compute mode INPUT data path labels
	opsComputeInputDigitalLabel *widget.Label // Shows "x0: 128\n0b10000000"
	opsComputeInputDACLabel     *widget.Label // Shows "0.50V"

	// Compute mode OUTPUT data path labels
	opsComputeOutputCurrentLabel *widget.Label // Shows "50.0 uA" or "160.0 uA (SAT)"
	opsComputeOutputTIALabel     *widget.Label // Shows "0.500 V" or "1.000 V (SAT)"
	opsComputeOutputADCLabel     *widget.Label // Shows "Level 16" or "Level 31 (SAT)"

	// Operations action buttons
	opsProgramBtn       *widget.Button
	opsProgramRandomBtn *widget.Button
	opsReadBtn          *widget.Button
	opsVerifyBtn        *widget.Button
	opsComputeBtn       *widget.Button
	opsAnimateBtn       *widget.Button
	opsResetBtn         *widget.Button
	opsWriteButtons     *fyne.Container
	opsReadButtons      *fyne.Container
	opsComputeButtons   *fyne.Container

	// Animation state
	animationStep     int // 0=none, 1=DAC, 2=Array, 3=ADC
	animationActive   bool
	programmingActive bool
	stopChan          chan struct{} // Closed to signal all goroutines to stop
	stopped           bool          // True when Stop() has been called
	backgroundWG      sync.WaitGroup

	// Zoom state
	zoomLevel  float64 // 1.0 = 100%, range 0.5 to 3.0
	zoomSlider *widget.Slider
	zoomLabel  *widget.Label

	// UI update flags (prevent recursive callbacks)
	updatingWLCheckboxes bool // True when programmatically updating WL checkboxes

	// Architecture toggle (1T1R vs 0T1R vs 2T1R)
	archPassiveBtn *widget.Button
	arch1T1RBtn    *widget.Button
	arch2T1RBtn    *widget.Button
	archToggle     *fyne.Container

	// Coupling model selector (Ideal / Tier A / Tier B)
	couplingIdealBtn   *widget.Button // legacy test hooks
	couplingApproxBtn  *widget.Button // legacy test hooks
	couplingToggle     *fyne.Container
	couplingTierSelect *widget.Select

	// ============================================================================
	// UNIFIED DEVICE SIMULATION VIEW (tab_unified.go)
	// ============================================================================
	deviceState *DeviceState

	// Unified view DAC input widgets
	unifiedDACEntries []*widget.Entry
	unifiedDACLabels  []*widget.Label

	// Unified view DAC range indicator
	dacRangeLabel *widget.Label

	// Material selection button (like module 1)
	materialBtn *widget.Button

	// Technology node selection (CMOS scaling)
	techNodeSelect      *widget.Select
	selectedTechNodeNm  float64
	refCellFootprintLbl *widget.Label
	refCellDensityLbl   *widget.Label

	// Unified view WL selector widgets
	unifiedWLChecks    []*widget.Check
	unifiedWLHelpLabel *widget.Label // Shows "Checked = Active" or passive mode explanation

	// Unified view mode buttons (READ/WRITE/COMPUTE)
	modeReadBtn    *widget.Button
	modeWriteBtn   *widget.Button
	modeComputeBtn *widget.Button

	// Unified view output labels
	unifiedOutputLabels []*widget.Label

	// READ observability overlay on array canvas (Off / Vcell / Icell)
	readOverlaySelect *widget.Select
	readOverlayMode   string

	// ============================================================================
	// MODE-FIRST UX PANELS (Phase 1)
	// ============================================================================

	// Write mode panel with level slider (mfux = Mode-First UX prefix)
	writeModePanel        *fyne.Container
	mfuxWriteLevelSlider  *widget.Slider
	mfuxWriteLevelLabel   *widget.Label
	mfuxWriteVoltageLabel *widget.Label
	mfuxWriteTargetLabel  *widget.Label // H2 FIX: Shows "Target: Row X, Col Y"
	isppEngineSelect      *widget.Select
	isppMethodSelect      *widget.Select

	// Voltage rules UI widgets
	writeSequencePanel       *fyne.Container   // 4-phase timing diagram container
	halfSelectIndicator      *widget.Label     // V/2 bias status indicator
	hysteresisDirectionLabel *widget.Label     // Direction indicator (^/v)
	passiveVoltagePanel      fyne.CanvasObject // 0T1R voltage panel
	activeVoltagePanel       fyne.CanvasObject // 1T1R/2T1R voltage panel

	// Compute mode panel with input vector entries
	computeModePanel      *fyne.Container
	computeInputTitle     *widget.Label   // Title label (updated on resize)
	computeInputContainer *fyne.Container // Container for input entries (rebuilt on resize)
	mfuxInputVectorEntry  []*widget.Entry
	mfuxInputVectorLabels []*widget.Label
	computeRandomBtn      *widget.Button
	computeClearBtn       *widget.Button

	// Sense panel widgets (READ/COMPUTE)
	sensePanel           *fyne.Container
	senseTitleLabel      *widget.Label
	senseRowLabel        *widget.Label
	senseCurrentLabel    *widget.Label
	senseVoltageLabel    *widget.Label
	senseCodeLabel       *widget.Label
	senseSNRLabel        *widget.Label
	senseSaturationLabel *widget.Label
	senseRangeLabel      *widget.Label
	senseLSBLabel        *widget.Label
	senseENOBLabel       *widget.Label
	senseEffSNRLabel     *widget.Label
	senseRfEntry         *widget.Entry
	senseAdcVminEntry    *widget.Entry
	senseAdcVmaxEntry    *widget.Entry
	sensePresetSelect    *widget.Select
	sensePresetUpdating  bool

	// H3 FIX: Undo history for array changes
	undoHistory    [][]int // Previous array state
	undoHistoryBtn *widget.Button
	hasUndoHistory bool

	// Action buttons (stored for mode-based enable/disable)
	actionWriteCellBtn   *widget.Button
	actionComputeBtn     *widget.Button
	actionRandomArrayBtn *widget.Button
	actionResetArrayBtn  *widget.Button
	actionFitBtn         *widget.Button
}

// NewCircuitsApp creates and initializes the circuits demo application.
func NewCircuitsApp() *CircuitsApp {
	ca := &CircuitsApp{
		arrayRows:             DefaultSize,
		arrayCols:             DefaultSize,
		quantLevels:           FeCIMLevels,
		dacBits:               DefaultDACBits,
		adcBits:               DefaultADCBits,
		vMin:                  1.2, // Minimum write voltage (just above Vc for 10nm HZO)
		vMax:                  1.5, // Maximum write voltage
		pulseWidth:            50.0,
		readVoltage:           0.5,
		tiaGain:               10.0,
		selectedRow:           3,
		selectedCol:           5,
		targetLevel:           15,
		compArraySize:         8,                              // Start with 8x8 array for comparison
		architecture:          sharedwidgets.Architecture0T1R, // Default to passive for educational demo
		zoomLevel:             1.0,                            // Default zoom 100%
		readOverlayMode:       "Off",
		showCurrentInCellInfo: false,
		selectedTechNodeNm:    130,
	}

	// Create Fyne app
	ca.fyneApp = app.NewWithID("com.fecim.circuits-demo")
	ca.fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

	// Initialize peripheral components
	ca.dac = peripherals.DefaultDAC()
	ca.adc = peripherals.DefaultADC()
	ca.tia = peripherals.DefaultTIA()
	ca.pump = peripherals.DefaultChargePump()

	// Initialize array
	ca.initializeArray()

	// Initialize stop channel for goroutine lifecycle
	ca.stopChan = make(chan struct{})

	// Register with global preset manager
	presets.Global().RegisterProvider(NewCircuitsPresetProvider(ca))

	return ca
}

// initializeArray sets up the weight array with all cells at mid-level
func (ca *CircuitsApp) initializeArray() {
	midLevel := ca.quantLevels / 2
	ca.arrayWeights = make([][]int, ca.arrayRows)
	ca.halfSelectResidue = make([][]float64, ca.arrayRows)
	for i := range ca.arrayWeights {
		ca.arrayWeights[i] = make([]int, ca.arrayCols)
		ca.halfSelectResidue[i] = make([]float64, ca.arrayCols)
		// All cells start at mid-level (neutral state)
		for j := range ca.arrayWeights[i] {
			ca.arrayWeights[i][j] = midLevel
		}
	}

	ca.inputVector = make([]int, ca.arrayCols)
	ca.outputVector = make([]float64, ca.arrayRows)
	// Input vector starts at 0
}

// Run starts the GUI application.
func (ca *CircuitsApp) Run() {
	ca.window = ca.fyneApp.NewWindow("FeCIM Demo 4: Peripheral Circuits")
	ca.window.Resize(fyne.NewSize(1400, 900))

	// Create main tabbed layout
	content := ca.createMainLayout()
	ca.window.SetContent(content)

	// Setup keyboard shortcuts
	ca.setupKeyboard()

	ca.window.ShowAndRun()
}

// createMainLayout builds the main application layout.
// FOCUS-92: keep Module 4 in OPERATIONS-only mode (remove dead View dropdown).
func (ca *CircuitsApp) createMainLayout() fyne.CanvasObject {
	operationsContent := ca.createUnifiedView()

	// Initialize comparison/reference tabs so their fields (compStatusLabel, etc.)
	// are non-nil even though we only show OPERATIONS view (FOCUS-92).
	_ = ca.createComparisonTab()
	_ = ca.createReferenceTab()

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("OPERATIONS"),
			layout.NewSpacer(),
			widget.NewLabel("DAC -> FeFET -> TIA -> ADC"),
		),
		widget.NewSeparator(),
	)

	footerLabel := widget.NewLabel("FeCIM Ferroelectric Compute-in-Memory | Based on Published Research")
	footerLabel.Alignment = fyne.TextAlignCenter
	footer := container.NewVBox(widget.NewSeparator(), footerLabel)

	return container.NewBorder(header, footer, nil, nil, operationsContent)
}

// ============================================================================
// HELPER METHODS FOR UI COMPONENTS
// ============================================================================

// createLabeledBox creates a styled box with title and value text (static value)
func (ca *CircuitsApp) createLabeledBox(title, value string, bgColor color.Color) *fyne.Container {
	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	titleLbl.Alignment = fyne.TextAlignCenter

	valueLbl := widget.NewLabel(value)
	valueLbl.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(100, 60))
	bg.CornerRadius = 5

	content := container.NewVBox(titleLbl, valueLbl)

	return container.NewStack(bg, container.NewCenter(content))
}

// createLabeledBoxWithLabel creates a styled box with title and a widget.Label for dynamic updates
func (ca *CircuitsApp) createLabeledBoxWithLabel(title string, valueLbl *widget.Label, bgColor color.Color) *fyne.Container {
	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	titleLbl.Alignment = fyne.TextAlignCenter

	valueLbl.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(100, 60))
	bg.CornerRadius = 5

	content := container.NewVBox(titleLbl, valueLbl)

	return container.NewStack(bg, container.NewCenter(content))
}

// ============================================================================
// REFRESH METHODS FOR LEGACY COMPATIBILITY
// ============================================================================

// refreshWriteArray refreshes the write mode array canvas (legacy)
func (ca *CircuitsApp) refreshWriteArray() {
	if ca.writeArrayCanvas != nil {
		sharedwidgets.SafeDo(func() {
			ca.writeArrayCanvas.Refresh()
		})
	}
}

// refreshWritePulse refreshes the write pulse visualization (legacy)
func (ca *CircuitsApp) refreshWritePulse() {
	if ca.writePulseCanvas != nil {
		sharedwidgets.SafeDo(func() {
			ca.writePulseCanvas.Refresh()
		})
	}
}

// refreshReadZone refreshes the read zone visualization (legacy)
func (ca *CircuitsApp) refreshReadZone() {
	if ca.readZoneCanvas != nil {
		sharedwidgets.SafeDo(func() {
			ca.readZoneCanvas.Refresh()
		})
	}
}
