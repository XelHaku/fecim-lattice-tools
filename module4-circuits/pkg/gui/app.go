// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// Module 4: Peripheral Circuits - Unified 3-view design
// OPERATIONS (Write/Read/Compute modes), COMPARISON (benchmarks), REFERENCE (timing diagrams + specs)
package gui

import (
	"fmt"
	"image/color"
	"math/rand"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/peripherals"
	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// Constants
const (
	FeCIMLevels    = 30  // Always 30 levels for FeCIM
	MaxArraySize   = 128 // Maximum array dimension
	DefaultSize    = 8   // Default array size
	DefaultDACBits = 8   // Default DAC resolution
	DefaultADCBits = 8   // Default ADC resolution
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
	mu           sync.RWMutex
	arrayRows    int
	arrayCols    int
	quantLevels  int
	dacBits      int
	adcBits      int
	vMin         float64 // Min write voltage
	vMax         float64 // Max write voltage
	pulseWidth   float64 // ns
	readVoltage  float64 // Read voltage (safe zone)
	tiaGain      float64 // TIA gain (kOhm)
	selectedRow  int
	selectedCol  int
	targetLevel  int
	arrayWeights [][]int // Current programmed levels
	inputVector  []int   // Input vector for compute
	outputVector []float64

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
	specSummaryLabel     *widget.Label
	specStatusLabel      *widget.Label

	// Tab 7: Reference (unified TIMING + SPECS)
	refTimingSection fyne.CanvasObject
	refSpecsSection  fyne.CanvasObject

	// Main tabs
	mainTabs *container.AppTabs

	// ============================================================================
	// UNIFIED OPERATIONS VIEW (tab_operations.go)
	// ============================================================================
	currentMode           OperationMode
	operationsStatusLabel *widget.Label
	operationsModeHelp    *widget.Label

	// Shared array section
	sharedArrayCanvas    *canvas.Raster
	sharedCellInfoLabel  *widget.Label
	sharedArrayInfoLabel *widget.Label
	sharedArrayCellSize  int // For click detection
	sharedArrayOffsetX   int
	sharedArrayOffsetY   int

	// Write mode panel widgets (ops prefix to distinguish from legacy)
	writeConfigPanel      *fyne.Container
	opsWriteLevelSlider   *widget.Slider
	opsWriteLevelLabel    *widget.Label
	opsWriteDigitalLabel  *widget.Label
	opsWriteDACLabel      *widget.Label
	opsWriteFeFETLabel    *widget.Label
	opsWritePulseCanvas   *canvas.Raster

	// Read mode panel widgets
	readConfigPanel        *fyne.Container
	opsReadVoltageSlider   *widget.Slider
	opsReadVoltageLabel    *widget.Label
	opsReadZoneCanvas      *canvas.Raster
	opsReadResultsLabel    *widget.Label

	// Compute mode panel widgets
	computeConfigPanel       *fyne.Container
	opsComputeInputs         []*widget.Entry
	opsComputeVoltageLabels  []*widget.Label
	opsComputeOutputLabels   []*widget.Label
	opsComputeMathLabel      *widget.Label

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
}

// NewCircuitsApp creates and initializes the circuits demo application.
func NewCircuitsApp() *CircuitsApp {
	ca := &CircuitsApp{
		arrayRows:   DefaultSize,
		arrayCols:   DefaultSize,
		quantLevels: FeCIMLevels,
		dacBits:     DefaultDACBits,
		adcBits:     DefaultADCBits,
		vMin:        2.0,
		vMax:        5.0,
		pulseWidth:  50.0,
		readVoltage: 0.5,
		tiaGain:     10.0,
		selectedRow: 3,
		selectedCol: 5,
		targetLevel: 15,
		compArraySize: 8, // Start with 8x8 array for comparison
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

	return ca
}

// initializeArray sets up the weight array with random values
func (ca *CircuitsApp) initializeArray() {
	ca.arrayWeights = make([][]int, ca.arrayRows)
	for i := range ca.arrayWeights {
		ca.arrayWeights[i] = make([]int, ca.arrayCols)
		for j := range ca.arrayWeights[i] {
			ca.arrayWeights[i][j] = rand.Intn(ca.quantLevels)
		}
	}

	ca.inputVector = make([]int, ca.arrayCols)
	ca.outputVector = make([]float64, ca.arrayRows)
	for j := range ca.inputVector {
		ca.inputVector[j] = rand.Intn(256)
	}
}

// Run starts the GUI application.
func (ca *CircuitsApp) Run() {
	ca.window = ca.fyneApp.NewWindow("FeCIM Demo 4: Peripheral Circuits")
	ca.window.Resize(fyne.NewSize(1400, 900))

	// Create main tabbed layout
	content := ca.createMainLayout()
	ca.window.SetContent(content)

	ca.window.ShowAndRun()
}

// createMainLayout builds the main application layout with 3 views.
func (ca *CircuitsApp) createMainLayout() fyne.CanvasObject {
	// Create tab contents (pre-loaded to avoid layout cascades on Wayland/Sway)
	operationsContent := ca.createOperationsView()     // NEW: from tab_operations.go
	comparisonContent := ca.createComparisonTab()       // KEEP: from tab_comparison.go
	referenceContent := ca.createReferenceTab()         // NEW: from tab_reference.go

	// All views for Hide/Show toggling
	viewNames := []string{"OPERATIONS", "COMPARISON", "REFERENCE"}
	allViews := []fyne.CanvasObject{
		operationsContent, comparisonContent, referenceContent,
	}

	// View selector dropdown
	viewSelector := widget.NewSelect(viewNames, nil)
	viewSelector.SetSelected("OPERATIONS")

	// Content container using Stack
	contentContainer := container.NewStack(allViews...)

	// Track current view
	currentView := ""

	// Update view based on selection
	viewSelector.OnChanged = func(view string) {
		sharedwidgets.DebugInteraction(fmt.Sprintf("circuits viewSelector changed to '%s'", view))
		if view == currentView {
			return
		}
		currentView = view

		// Hide all views, then show selected
		for i, v := range allViews {
			if viewNames[i] == view {
				v.Show()
			} else {
				v.Hide()
			}
		}

		// Refresh canvases when specific views shown
		if view == "OPERATIONS" {
			ca.refreshSharedArray()
		} else if view == "REFERENCE" {
			ca.refreshTimingDiagrams()
		}
	}

	// Initialize: show first view, hide others
	for i, v := range allViews {
		if i == 0 {
			v.Show()
		} else {
			v.Hide()
		}
	}
	currentView = "OPERATIONS"

	// Header with inline view selector
	titleLabel := widget.NewLabel("FeCIM Peripheral Circuits Visualizer")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	headerRow := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		widget.NewLabel("View:"),
		viewSelector,
		layout.NewSpacer(),
		widget.NewLabel("3 Views | DAC -> FeFET -> TIA -> ADC"),
	)

	header := container.NewVBox(
		headerRow,
		widget.NewSeparator(),
	)

	// Footer
	footerLabel := widget.NewLabel("FeCIM Ferroelectric Compute-in-Memory | Based on Dr. Tour's Research")
	footerLabel.Alignment = fyne.TextAlignCenter

	footer := container.NewVBox(
		widget.NewSeparator(),
		footerLabel,
	)

	return container.NewBorder(header, footer, nil, nil, contentContainer)
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
		fyne.Do(func() {
			ca.writeArrayCanvas.Refresh()
		})
	}
}

// refreshWritePulse refreshes the write pulse visualization (legacy)
func (ca *CircuitsApp) refreshWritePulse() {
	if ca.writePulseCanvas != nil {
		fyne.Do(func() {
			ca.writePulseCanvas.Refresh()
		})
	}
}

// refreshReadZone refreshes the read zone visualization (legacy)
func (ca *CircuitsApp) refreshReadZone() {
	if ca.readZoneCanvas != nil {
		fyne.Do(func() {
			ca.readZoneCanvas.Refresh()
		})
	}
}

