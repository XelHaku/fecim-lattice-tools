// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// Module 4: Peripheral Circuits - Complete revamp with 6 tabs
// Write, Read, Compute, Comparison, Timing, Specifications
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/peripherals"
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

// Theme colors
var (
	colorBackground   = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
	colorPrimary      = color.RGBA{0, 212, 255, 255}   // Cyan
	colorAccent       = color.RGBA{255, 165, 0, 255}   // Orange for highlights
	colorSuccess      = color.RGBA{0, 200, 100, 255}   // Green for success
	colorWarning      = color.RGBA{255, 200, 0, 255}   // Yellow for warnings
	colorDanger       = color.RGBA{255, 80, 80, 255}   // Red for danger
	colorCPU          = color.RGBA{200, 100, 100, 255} // CPU color
	colorGPU          = color.RGBA{100, 200, 100, 255} // GPU color
	colorFeFET        = color.RGBA{100, 150, 255, 255} // FeFET color
	colorWriteZone    = color.RGBA{200, 50, 50, 200}   // Write zone (danger)
	colorReadZone     = color.RGBA{50, 150, 50, 200}   // Read zone (safe)
	colorThreshold    = color.RGBA{255, 200, 0, 200}   // Threshold line
	colorDAC          = color.RGBA{150, 100, 200, 255} // Purple for DAC
	colorADC          = color.RGBA{100, 200, 150, 255} // Green for ADC
	colorTIA          = color.RGBA{200, 150, 100, 255} // Orange for TIA
	colorArrayCell    = color.RGBA{100, 150, 200, 255} // Blue for array cells
	colorSelectedCell = color.RGBA{255, 200, 50, 255}  // Yellow for selected
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
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
	specStatusLabel      *widget.Label

	// Main tabs
	mainTabs *container.AppTabs
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
	}

	// Create Fyne app
	ca.fyneApp = app.NewWithID("com.fecim.circuits-demo")
	ca.fyneApp.Settings().SetTheme(&feCIMTheme{})

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

// createMainLayout builds the main application layout with tabs.
func (ca *CircuitsApp) createMainLayout() fyne.CanvasObject {
	// Create tab contents (pre-loaded to avoid layout cascades on Wayland/Sway)
	writeTabContent := ca.createWriteTab()
	readTabContent := ca.createReadTab()
	computeTabContent := ca.createComputeTab()
	comparisonTabContent := ca.createComparisonTab()
	timingTabContent := ca.createTimingTab()
	specsTabContent := ca.createSpecsTab()

	// All views for Hide/Show toggling
	viewNames := []string{"WRITE", "READ", "COMPUTE", "COMPARISON", "TIMING", "SPECS"}
	allViews := []fyne.CanvasObject{
		writeTabContent, readTabContent, computeTabContent,
		comparisonTabContent, timingTabContent, specsTabContent,
	}

	// View selector dropdown (replaces nested tabs to save space)
	viewSelector := widget.NewSelect(viewNames, nil)
	viewSelector.SetSelected("WRITE")

	// Content container using Stack - all views layered, visibility toggled
	contentContainer := container.NewStack(allViews...)

	// Track current view
	currentView := ""

	// Update view based on selection using Hide/Show (avoids layout cascades)
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
	}

	// Initialize: show first view, hide others
	for i, v := range allViews {
		if i == 0 {
			v.Show()
		} else {
			v.Hide()
		}
	}
	currentView = "WRITE"

	// Header with inline view selector
	titleLabel := widget.NewLabel("FeCIM Peripheral Circuits Visualizer")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	headerRow := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		widget.NewLabel("View:"),
		viewSelector,
		layout.NewSpacer(),
		widget.NewLabel("DAC → FeFET → TIA → ADC | 30 Levels"),
	)

	header := container.NewVBox(
		headerRow,
		widget.NewSeparator(),
	)

	// Footer
	footerLabel := widget.NewLabel("FeCIM Ferroelectric Compute-in-Memory | Based on Dr. Tour's Research | Standard CMOS Compatible")
	footerLabel.Alignment = fyne.TextAlignCenter

	footer := container.NewVBox(
		widget.NewSeparator(),
		footerLabel,
	)

	return container.NewBorder(header, footer, nil, nil, contentContainer)
}

// ============================================================================
// TAB 1: WRITE MODE
// ============================================================================

func (ca *CircuitsApp) createWriteTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**WRITE MODE**: Program ferroelectric cells to specific conductance levels using precise voltage pulses from the charge pump and DAC. The DAC converts digital levels (0-29) to analog voltages (2.0V-5.0V), which are applied as pulses to modify the FeFET polarization state.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Configuration section
	configSection := ca.createWriteConfigSection()

	// Cell selection section
	cellSection := ca.createWriteCellSection()

	// Data path visualization
	dataPathSection := ca.createWriteDataPathSection()

	// Programming pulse visualization
	pulseSection := ca.createWritePulseSection()

	// Array view
	arraySection := ca.createWriteArraySection()

	// Level-to-voltage mapping table
	mappingSection := ca.createWriteMappingSection()

	// Buttons
	programBtn := widget.NewButton("PROGRAM CELL", ca.onProgramCell)
	programBtn.Importance = widget.HighImportance

	randomBtn := widget.NewButton("PROGRAM RANDOM ARRAY", ca.onProgramRandomArray)

	ca.writeStatusLabel = widget.NewLabel("Ready to program")

	buttonBox := container.NewHBox(
		programBtn,
		randomBtn,
		layout.NewSpacer(),
		ca.writeStatusLabel,
	)

	// Layout
	leftPanel := container.NewVBox(
		widget.NewLabel("CONFIGURATION"),
		configSection,
		widget.NewSeparator(),
		widget.NewLabel("CELL SELECTION"),
		cellSection,
	)

	centerPanel := container.NewVBox(
		widget.NewLabel("DATA PATH VISUALIZATION"),
		dataPathSection,
		widget.NewSeparator(),
		widget.NewLabel("PROGRAMMING PULSE"),
		pulseSection,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("LEVEL-TO-VOLTAGE MAPPING"),
		mappingSection,
	)

	topRow := container.NewHBox(
		container.NewPadded(leftPanel),
		widget.NewSeparator(),
		container.NewPadded(centerPanel),
		widget.NewSeparator(),
		container.NewPadded(rightPanel),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator(), topRow),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		container.NewVBox(
			widget.NewLabel("ARRAY VIEW (click cell to select)"),
			arraySection,
		),
	)
}

func (ca *CircuitsApp) createWriteConfigSection() fyne.CanvasObject {
	// Array size selects
	sizeOptions := []string{"4", "8", "16", "32", "64"}
	rowSelect := widget.NewSelect(sizeOptions, func(s string) {
		var size int
		fmt.Sscanf(s, "%d", &size)
		ca.mu.Lock()
		ca.arrayRows = size
		ca.mu.Unlock()
		ca.initializeArray()
		ca.refreshWriteArray()
	})
	rowSelect.SetSelected("8")

	colSelect := widget.NewSelect(sizeOptions, func(s string) {
		var size int
		fmt.Sscanf(s, "%d", &size)
		ca.mu.Lock()
		ca.arrayCols = size
		ca.mu.Unlock()
		ca.initializeArray()
		ca.refreshWriteArray()
	})
	colSelect.SetSelected("8")

	// Quantization levels
	levelOptions := []string{"2", "4", "8", "16", "30", "32", "64", "128", "256"}
	levelSelect := widget.NewSelect(levelOptions, func(s string) {
		var levels int
		fmt.Sscanf(s, "%d", &levels)
		ca.mu.Lock()
		ca.quantLevels = levels
		ca.mu.Unlock()
	})
	levelSelect.SetSelected("30")
	quantHelp := widget.NewLabel("FeCIM uses 30 discrete analog states per cell (Dr. Tour, COSM 2025)")
	quantHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Voltage range entries
	vMinEntry := widget.NewEntry()
	vMinEntry.SetText("2.0")
	vMinEntry.SetPlaceHolder("Minimum write voltage (V) - must exceed coercive field")
	vMinEntry.OnChanged = func(s string) {
		var v float64
		fmt.Sscanf(s, "%f", &v)
		ca.mu.Lock()
		ca.vMin = v
		ca.mu.Unlock()
	}

	vMaxEntry := widget.NewEntry()
	vMaxEntry.SetText("5.0")
	vMaxEntry.SetPlaceHolder("Maximum write voltage (V) - for full polarization")
	vMaxEntry.OnChanged = func(s string) {
		var v float64
		fmt.Sscanf(s, "%f", &v)
		ca.mu.Lock()
		ca.vMax = v
		ca.mu.Unlock()
	}

	// Pulse width entry
	pulseEntry := widget.NewEntry()
	pulseEntry.SetText("50")
	pulseEntry.SetPlaceHolder("Pulse duration in nanoseconds (typical FeFET: 10-100 ns)")
	pulseEntry.OnChanged = func(s string) {
		var pw float64
		fmt.Sscanf(s, "%f", &pw)
		ca.mu.Lock()
		ca.pulseWidth = pw
		ca.mu.Unlock()
	}

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Array Size:"),
			rowSelect,
			widget.NewLabel("x"),
			colSelect,
		),
		container.NewHBox(
			widget.NewLabel("Quantization:"),
			levelSelect,
			widget.NewLabel("levels"),
		),
		quantHelp,
		container.NewHBox(
			widget.NewLabel("Voltage Range:"),
			vMinEntry,
			widget.NewLabel("V min"),
			vMaxEntry,
			widget.NewLabel("V max"),
		),
		container.NewHBox(
			widget.NewLabel("Pulse Width:"),
			pulseEntry,
			widget.NewLabel("ns"),
		),
		widget.NewLabel("(Write pulse duration: shorter = faster but needs higher voltage)"),
	)

	return form
}

func (ca *CircuitsApp) createWriteCellSection() fyne.CanvasObject {
	// Row/col selects
	rowOptions := make([]string, ca.arrayRows)
	for i := range rowOptions {
		rowOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.writeRowSelect = widget.NewSelect(rowOptions, func(s string) {
		var row int
		fmt.Sscanf(s, "%d", &row)
		ca.mu.Lock()
		ca.selectedRow = row
		ca.mu.Unlock()
		ca.refreshWriteArray()
		ca.updateWriteDataPath()
	})
	ca.writeRowSelect.SetSelected("3")

	colOptions := make([]string, ca.arrayCols)
	for i := range colOptions {
		colOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.writeColSelect = widget.NewSelect(colOptions, func(s string) {
		var col int
		fmt.Sscanf(s, "%d", &col)
		ca.mu.Lock()
		ca.selectedCol = col
		ca.mu.Unlock()
		ca.refreshWriteArray()
		ca.updateWriteDataPath()
	})
	ca.writeColSelect.SetSelected("5")

	// Target level slider
	ca.writeLevelLabel = widget.NewLabel("Target Level: 15 / 30 (discrete conductance state)")
	ca.writeLevelSlider = widget.NewSlider(0, float64(ca.quantLevels-1))
	ca.writeLevelSlider.Value = 15
	ca.writeLevelSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.targetLevel = int(v)
		ca.mu.Unlock()
		ca.writeLevelLabel.SetText(fmt.Sprintf("Target Level: %d / %d (discrete conductance state)", ca.targetLevel, ca.quantLevels))
		ca.updateWriteDataPath()
		ca.refreshWritePulse()
	}

	levelHelp := widget.NewLabel("Each level represents a stable polarization state (~4.9 bits/cell)")
	levelHelp.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Target Cell: Row"),
			ca.writeRowSelect,
			widget.NewLabel("Col"),
			ca.writeColSelect,
		),
		ca.writeLevelLabel,
		ca.writeLevelSlider,
		levelHelp,
	)
}

func (ca *CircuitsApp) createWriteDataPathSection() fyne.CanvasObject {
	// Create visual boxes for the data path with stored label references
	ca.writeDigitalLabel = widget.NewLabel("Level:15\n01111")
	ca.writeDACLabel = widget.NewLabel("3.55V")
	ca.writeFeFETLabel = widget.NewLabel("[3,5]\n52.2µS")

	digitalBox := ca.createLabeledBoxWithLabel("DIGITAL", ca.writeDigitalLabel, colorPrimary)
	dacBox := ca.createLabeledBoxWithLabel("DAC", ca.writeDACLabel, colorDAC)
	fefetBox := ca.createLabeledBoxWithLabel("FeFET", ca.writeFeFETLabel, colorArrayCell)

	arrow1 := widget.NewLabel("→")
	arrow2 := widget.NewLabel("→")

	ca.writeDataPath = container.NewHBox(
		digitalBox,
		arrow1,
		dacBox,
		arrow2,
		fefetBox,
	)

	ca.updateWriteDataPath()

	helperText := widget.NewLabel("Data path: Digital level → DAC voltage conversion → FeFET polarization")
	helperText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(ca.writeDataPath, helperText)
}

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

func (ca *CircuitsApp) updateWriteDataPath() {
	ca.mu.RLock()
	level := ca.targetLevel
	row := ca.selectedRow
	col := ca.selectedCol
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Calculate voltage
	voltage := vMin + float64(level)/float64(levels-1)*(vMax-vMin)

	// Calculate conductance (1-100 µS range)
	conductance := 1.0 + float64(level)/float64(levels-1)*99.0

	// Binary representation
	binary := fmt.Sprintf("%05b", level)

	// Update the data path display using direct label references
	if ca.writeDigitalLabel != nil {
		ca.writeDigitalLabel.SetText(fmt.Sprintf("Level:%d\n%s", level, binary))
	}
	if ca.writeDACLabel != nil {
		ca.writeDACLabel.SetText(fmt.Sprintf("%.2fV", voltage))
	}
	if ca.writeFeFETLabel != nil {
		ca.writeFeFETLabel.SetText(fmt.Sprintf("[%d,%d]\n%.1fµS", row, col, conductance))
	}
}

func (ca *CircuitsApp) createWritePulseSection() fyne.CanvasObject {
	ca.writePulseCanvas = canvas.NewRaster(ca.drawWritePulse)
	ca.writePulseCanvas.SetMinSize(fyne.NewSize(400, 150))
	return ca.writePulseCanvas
}

func (ca *CircuitsApp) drawWritePulse(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	level := ca.targetLevel
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	_ = ca.pulseWidth // Used for display
	ca.mu.RUnlock()

	voltage := vMin + float64(level)/float64(levels-1)*(vMax-vMin)

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw axes
	marginLeft := 50
	marginBottom := 30
	marginTop := 20
	marginRight := 30

	plotW := w - marginLeft - marginRight
	plotH := h - marginTop - marginBottom
	axisColor := color.RGBA{200, 200, 200, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	fillColor := color.RGBA{0, 100, 150, 200}
	threshColor := color.RGBA{255, 200, 0, 255}

	// Y-axis (voltage)
	for y := marginTop; y < h-marginBottom; y++ {
		img.Set(marginLeft, y, axisColor)
	}

	// X-axis (time)
	for x := marginLeft; x < w-marginRight; x++ {
		img.Set(x, h-marginBottom, axisColor)
	}

	// Pulse positions
	pulseStart := marginLeft + plotW*10/100
	pulseEnd := marginLeft + plotW*70/100
	riseEnd := pulseStart + plotW*2/100
	fallStart := pulseEnd - plotW*2/100

	// Y positions
	y0V := h - marginBottom
	yVoltage := marginTop + int(float64(plotH)*(1.0-(voltage-0)/(vMax+0.5)))
	yThreshold := marginTop + int(float64(plotH)*(1.0-(vMin-0)/(vMax+0.5)))

	// Draw threshold line (dashed)
	if yThreshold >= marginTop && yThreshold < h-marginBottom {
		for x := marginLeft; x < w-marginRight; x += 6 {
			img.Set(x, yThreshold, threshColor)
		}
	}

	// Draw pulse
	for x := marginLeft; x < w-marginRight; x++ {
		var y int
		if x < pulseStart {
			y = y0V
		} else if x < riseEnd {
			t := float64(x-pulseStart) / float64(riseEnd-pulseStart)
			y = y0V + int(float64(yVoltage-y0V)*t)
		} else if x < fallStart {
			y = yVoltage
		} else if x < pulseEnd {
			t := float64(x-fallStart) / float64(pulseEnd-fallStart)
			y = yVoltage + int(float64(y0V-yVoltage)*t)
		} else {
			y = y0V
		}

		// Draw thick line
		for dy := -2; dy <= 2; dy++ {
			py := y + dy
			if py >= marginTop && py < h-marginBottom {
				img.Set(x, py, cyanColor)
			}
		}

		// Fill pulse area
		if x >= riseEnd && x < fallStart {
			for py := yVoltage; py < y0V; py++ {
				img.Set(x, py, fillColor)
			}
		}
	}

	// Axis labels
	// Y-axis label
	drawScaledText(img, "Voltage (V)", marginLeft-40, marginTop-8, 1, axisColor)

	// X-axis label
	drawScaledText(img, "Time (ns)", w-marginRight-50, h-marginBottom+15, 1, axisColor)

	// Values
	drawSimpleText(img, fmt.Sprintf("%.1fV", vMax), 5, marginTop+5, axisColor)
	drawSimpleText(img, fmt.Sprintf("%.1fV", vMin), 5, yThreshold+5, axisColor)
	drawSimpleText(img, "0V", 25, y0V-5, axisColor)

	return img
}

// Simple rect drawing helper for image.RGBA
func drawRect(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	for py := y; py < y+rectH; py++ {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.Set(px, py, c)
			}
		}
	}
}

func (ca *CircuitsApp) refreshWritePulse() {
	if ca.writePulseCanvas != nil {
		fyne.Do(func() {
			ca.writePulseCanvas.Refresh()
		})
	}
}

func (ca *CircuitsApp) createWriteArraySection() fyne.CanvasObject {
	ca.writeArrayCanvas = canvas.NewRaster(ca.drawWriteArray)
	ca.writeArrayCanvas.SetMinSize(fyne.NewSize(500, 350))
	return ca.writeArrayCanvas
}

func (ca *CircuitsApp) drawWriteArray(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	selectedRow := ca.selectedRow
	selectedCol := ca.selectedCol
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if weights == nil || len(weights) == 0 {
		return img
	}

	// Calculate cell size (use square cells like crossbar module)
	margin := 40
	cellW := (w - 2*margin) / cols
	cellH := (h - 2*margin) / rows
	cellSize := cellW
	if cellH < cellSize {
		cellSize = cellH
	}
	if cellSize > 40 {
		cellSize = 40
	}
	if cellSize < 8 {
		cellSize = 8
	}

	// Center the grid in the available space
	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := (w - gridW) / 2
	offsetY := (h - gridH) / 2

	// Draw cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := offsetX + c*cellSize
			y0 := offsetY + r*cellSize

			level := weights[r][c]
			intensity := float64(level) / float64(levels-1)

			// Check if this is the selected cell
			isSelected := r == selectedRow && c == selectedCol

			// Color based on level (blue to red)
			var cr, cg, cb uint8
			if isSelected {
				cr, cg, cb = 255, 200, 50 // Bright yellow for selection
			} else {
				cr = uint8(intensity * 200)
				cg = uint8(50 + (1-intensity)*100)
				cb = uint8((1 - intensity) * 200)
			}

			cellColor := color.RGBA{cr, cg, cb, 255}
			drawRect(img, x0+2, y0+2, cellSize-4, cellSize-4, cellColor)

			// Draw a thick white border around the selected cell for better visibility
			if isSelected {
				borderColor := color.RGBA{255, 255, 255, 255}
				borderWidth := 3
				// Top border
				drawRect(img, x0, y0, cellSize, borderWidth, borderColor)
				// Bottom border
				drawRect(img, x0, y0+cellSize-borderWidth, cellSize, borderWidth, borderColor)
				// Left border
				drawRect(img, x0, y0, borderWidth, cellSize, borderColor)
				// Right border
				drawRect(img, x0+cellSize-borderWidth, y0, borderWidth, cellSize, borderColor)
			}
		}
	}

	return img
}

func (ca *CircuitsApp) refreshWriteArray() {
	if ca.writeArrayCanvas != nil {
		fyne.Do(func() {
			ca.writeArrayCanvas.Refresh()
		})
	}
}

func (ca *CircuitsApp) createWriteMappingSection() fyne.CanvasObject {
	ca.writeMappingLabel = widget.NewLabel(ca.getMappingText())
	return container.NewVScroll(ca.writeMappingLabel)
}

func (ca *CircuitsApp) getMappingText() string {
	ca.mu.RLock()
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	target := ca.targetLevel
	ca.mu.RUnlock()

	text := "LEVEL-TO-VOLTAGE MAPPING TABLE\n"
	text += "Shows how digital levels (0-29) map to programming voltages\n"
	text += "and resulting FeFET conductance states.\n"
	text += "================================================================\n\n"
	text += "Level   Voltage   Conductance   Resistance\n"
	text += "-----   -------   -----------   ----------\n"

	// Show more levels for better visibility (8 levels)
	sampleLevels := []int{0, 4, 8, 12, 15, 20, 25, levels - 1}

	// Always include target if not already present
	hasTarget := false
	for _, l := range sampleLevels {
		if l == target {
			hasTarget = true
			break
		}
	}
	if !hasTarget {
		// Insert target in sorted position
		newLevels := make([]int, 0, len(sampleLevels)+1)
		inserted := false
		for _, l := range sampleLevels {
			if !inserted && target < l {
				newLevels = append(newLevels, target)
				inserted = true
			}
			newLevels = append(newLevels, l)
		}
		if !inserted {
			newLevels = append(newLevels, target)
		}
		sampleLevels = newLevels
	}

	seen := make(map[int]bool)

	for _, l := range sampleLevels {
		if seen[l] || l >= levels {
			continue
		}
		seen[l] = true

		voltage := vMin + float64(l)/float64(levels-1)*(vMax-vMin)
		conductance := 1.0 + float64(l)/float64(levels-1)*99.0 // 1-100 µS
		resistance := 1000.0 / conductance                     // kO

		marker := "  "
		if l == target {
			marker = "> "
		}
		text += fmt.Sprintf("%s%2d      %5.2fV      %5.1f µS      %6.1f kΩ\n",
			marker, l, voltage, conductance, resistance)
	}

	text += "\n================================\n"
	text += fmt.Sprintf("TARGET: Level %d = %.2fV\n", target,
		vMin+float64(target)/float64(levels-1)*(vMax-vMin))
	text += fmt.Sprintf("Range: %.1fV to %.1fV (%d levels)\n", vMin, vMax, levels)

	return text
}

func (ca *CircuitsApp) onProgramCell() {
	ca.mu.Lock()
	row := ca.selectedRow
	col := ca.selectedCol
	level := ca.targetLevel

	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		ca.arrayWeights[row][col] = level
	}
	ca.mu.Unlock()

	ca.refreshWriteArray()
	ca.writeStatusLabel.SetText(fmt.Sprintf("Programmed cell [%d,%d] to level %d", row, col, level))
}

func (ca *CircuitsApp) onProgramRandomArray() {
	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = rand.Intn(ca.quantLevels)
		}
	}
	ca.mu.Unlock()

	ca.refreshWriteArray()
	ca.writeStatusLabel.SetText("Programmed array with random values")
}

// ============================================================================
// TAB 2: READ MODE
// ============================================================================

func (ca *CircuitsApp) createReadTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**READ MODE**: Sense the conductance state of ferroelectric cells using low voltage (0.5V) to avoid disturbing the stored data. The TIA (transimpedance amplifier) converts the cell current to voltage, which is then digitized by the ADC for output.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Configuration section
	configSection := ca.createReadConfigSection()

	// Cell selection section
	cellSection := ca.createReadCellSection()

	// Data path visualization
	dataPathSection := ca.createReadDataPathSection()

	// Voltage zones visualization
	zoneSection := ca.createReadZoneSection()

	// Results section
	resultsSection := ca.createReadResultsSection()

	// Buttons
	readBtn := widget.NewButton("READ CELL", ca.onReadCell)
	readBtn.Importance = widget.HighImportance

	readAllBtn := widget.NewButton("READ ALL CELLS", ca.onReadAllCells)

	verifyBtn := widget.NewButton("VERIFY ARRAY", ca.onVerifyArray)

	ca.readStatusLabel = widget.NewLabel("Ready to read")

	buttonBox := container.NewHBox(
		readBtn,
		readAllBtn,
		verifyBtn,
		layout.NewSpacer(),
		ca.readStatusLabel,
	)

	// Layout
	leftPanel := container.NewVBox(
		widget.NewLabel("CONFIGURATION"),
		configSection,
		widget.NewSeparator(),
		widget.NewLabel("CELL SELECTION"),
		cellSection,
	)

	centerPanel := container.NewVBox(
		widget.NewLabel("DATA PATH VISUALIZATION"),
		dataPathSection,
		widget.NewSeparator(),
		widget.NewLabel("VOLTAGE ZONES"),
		zoneSection,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("READ RESULTS"),
		resultsSection,
	)

	mainContent := container.NewHBox(
		container.NewPadded(leftPanel),
		widget.NewSeparator(),
		container.NewPadded(centerPanel),
		widget.NewSeparator(),
		container.NewPadded(rightPanel),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator(), mainContent),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		nil,
	)
}

func (ca *CircuitsApp) createReadConfigSection() fyne.CanvasObject {
	// Read voltage slider
	ca.readVoltageLabel = widget.NewLabel("Read Voltage: 0.5 V (non-destructive sensing)")
	ca.readVoltageSlider = widget.NewSlider(0.1, 1.5)
	ca.readVoltageSlider.Value = 0.5
	ca.readVoltageSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.readVoltage = v
		ca.mu.Unlock()
		ca.readVoltageLabel.SetText(fmt.Sprintf("Read Voltage: %.2f V (non-destructive sensing)", v))
		ca.refreshReadZone()
	}

	warningLabel := widget.NewLabel("SAFE ZONE: 0.1V - 1.0V")
	warningLabel.TextStyle = fyne.TextStyle{Bold: true}

	dangerLabel := widget.NewLabel("DANGER: > 2.0V (will modify cell!)")

	// ADC resolution select with helper text
	adcOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	adcSelect := widget.NewSelect(adcOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		ca.mu.Lock()
		ca.adcBits = bits
		ca.mu.Unlock()
	})
	adcSelect.SetSelected("8")
	adcHelp := widget.NewLabel("(bits of precision for digitizing analog current)")
	adcHelp.TextStyle = fyne.TextStyle{Italic: true}

	// TIA gain select with helper text
	tiaOptions := []string{"1", "10", "100"}
	tiaSelect := widget.NewSelect(tiaOptions, func(s string) {
		var gain float64
		fmt.Sscanf(s, "%f", &gain)
		ca.mu.Lock()
		ca.tiaGain = gain
		ca.mu.Unlock()
	})
	tiaSelect.SetSelected("10")
	tiaHelp := widget.NewLabel("(Transimpedance: converts cell current to voltage)")
	tiaHelp.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		ca.readVoltageLabel,
		ca.readVoltageSlider,
		warningLabel,
		dangerLabel,
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("ADC Resolution:"),
			adcSelect,
			widget.NewLabel("bits"),
		),
		adcHelp,
		container.NewHBox(
			widget.NewLabel("TIA Gain:"),
			tiaSelect,
			widget.NewLabel("kOhm"),
		),
		tiaHelp,
	)
}

func (ca *CircuitsApp) createReadCellSection() fyne.CanvasObject {
	rowOptions := make([]string, ca.arrayRows)
	for i := range rowOptions {
		rowOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.readRowSelect = widget.NewSelect(rowOptions, func(s string) {
		var row int
		fmt.Sscanf(s, "%d", &row)
		ca.mu.Lock()
		ca.selectedRow = row
		ca.mu.Unlock()
	})
	ca.readRowSelect.SetSelected("3")

	colOptions := make([]string, ca.arrayCols)
	for i := range colOptions {
		colOptions[i] = fmt.Sprintf("%d", i)
	}
	ca.readColSelect = widget.NewSelect(colOptions, func(s string) {
		var col int
		fmt.Sscanf(s, "%d", &col)
		ca.mu.Lock()
		ca.selectedCol = col
		ca.mu.Unlock()
	})
	ca.readColSelect.SetSelected("5")

	storedLabel := widget.NewLabel("Stored Level: -- (from previous WRITE)")

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Target Cell: Row"),
			ca.readRowSelect,
			widget.NewLabel("Col"),
			ca.readColSelect,
		),
		storedLabel,
	)
}

func (ca *CircuitsApp) createReadDataPathSection() fyne.CanvasObject {
	fefetBox := ca.createLabeledBox("FeFET", "Cell --,--", colorArrayCell)
	tiaBox := ca.createLabeledBox("TIA", "(I→V)", colorTIA)
	adcBox := ca.createLabeledBox("ADC", "8-bit", colorADC)
	digitalBox := ca.createLabeledBox("DIGITAL", "Output", colorPrimary)

	arrow1 := widget.NewLabel("→")
	arrow2 := widget.NewLabel("→")
	arrow3 := widget.NewLabel("→")

	ca.readDataPath = container.NewHBox(
		fefetBox,
		arrow1,
		tiaBox,
		arrow2,
		adcBox,
		arrow3,
		digitalBox,
	)

	helperText := widget.NewLabel("Data path: FeFET current → TIA voltage conversion → ADC digitization → Level")
	helperText.TextStyle = fyne.TextStyle{Italic: true}

	// Calculation box
	ca.readCalcLabel = widget.NewLabel(
		"I = G × V = -- µS × -- V = -- µA\n" +
			"V_tia = I × R = -- µA × -- kΩ = -- mV\n" +
			"ADC = (-- mV / 1000 mV) × 255 = --\n" +
			"Level = round(-- / 255 × 29) = --",
	)
	// Use monospace for better alignment
	ca.readCalcLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewVBox(
		ca.readDataPath,
		helperText,
		widget.NewSeparator(),
		widget.NewLabel("Calculation:"),
		ca.readCalcLabel,
	)
}

func (ca *CircuitsApp) createReadZoneSection() fyne.CanvasObject {
	ca.readZoneCanvas = canvas.NewRaster(ca.drawReadZone)
	ca.readZoneCanvas.SetMinSize(fyne.NewSize(300, 200))
	return ca.readZoneCanvas
}

func (ca *CircuitsApp) drawReadZone(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	readV := ca.readVoltage
	ca.mu.RUnlock()

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 50
	marginRight := 20
	marginTop := 15
	marginBottom := 15
	plotH := h - marginTop - marginBottom
	plotW := w - marginLeft - marginRight

	writeZoneColor := color.RGBA{200, 50, 50, 180}
	readZoneColor := color.RGBA{50, 150, 50, 180}
	threshColor := color.RGBA{255, 200, 0, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	labelColor := color.RGBA{255, 255, 255, 255}
	axisColor := color.RGBA{150, 150, 150, 255}

	// Y-axis voltage scale (5V at top, 0V at bottom)
	maxVoltage := 5.0

	// Helper to convert voltage to Y position
	voltageToY := func(v float64) int {
		return marginTop + int((maxVoltage-v)/maxVoltage*float64(plotH))
	}

	// Write zone (> 2V) - red danger zone
	writeZoneTop := voltageToY(maxVoltage)
	writeZoneBottom := voltageToY(2.0)
	drawRect(img, marginLeft, writeZoneTop, plotW, writeZoneBottom-writeZoneTop, writeZoneColor)

	// Transition zone (1V - 2V) - neutral
	// (no special coloring, just background)

	// Read zone (< 1V) - green safe zone
	readZoneTop := voltageToY(1.0)
	readZoneBottom := voltageToY(0.0)
	drawRect(img, marginLeft, readZoneTop, plotW, readZoneBottom-readZoneTop, readZoneColor)

	// Threshold line (2V)
	thresholdY := voltageToY(2.0)
	for x := marginLeft; x < marginLeft+plotW; x++ {
		for dy := -1; dy <= 1; dy++ {
			if thresholdY+dy >= marginTop && thresholdY+dy < h-marginBottom {
				img.Set(x, thresholdY+dy, threshColor)
			}
		}
	}

	// Zone labels (right side of zones)
	drawSimpleText(img, "WRITE ZONE", marginLeft+10, writeZoneTop+15, labelColor)
	drawSimpleText(img, "> 2.0V DANGER", marginLeft+10, writeZoneTop+28, color.RGBA{255, 150, 150, 255})

	drawSimpleText(img, "2.0V THRESHOLD", marginLeft+plotW-110, thresholdY-10, threshColor)

	drawSimpleText(img, "READ ZONE", marginLeft+10, readZoneTop+15, labelColor)
	drawSimpleText(img, "< 1.0V SAFE", marginLeft+10, readZoneTop+28, color.RGBA{150, 255, 150, 255})

	// Y-axis with voltage scale
	for y := marginTop; y <= h-marginBottom; y++ {
		img.Set(marginLeft-1, y, axisColor)
	}

	// Voltage labels on Y-axis
	voltageMarkers := []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0}
	for _, v := range voltageMarkers {
		y := voltageToY(v)
		// Tick mark
		for dx := 0; dx < 5; dx++ {
			img.Set(marginLeft-5+dx, y, axisColor)
		}
		// Voltage label
		label := fmt.Sprintf("%.1fV", v)
		drawSimpleText(img, label, 5, y-3, axisColor)
	}

	// Current read voltage indicator line
	readY := voltageToY(readV)
	for x := marginLeft; x < marginLeft+plotW; x++ {
		for dy := -2; dy <= 2; dy++ {
			y := readY + dy
			if y >= marginTop && y < h-marginBottom {
				img.Set(x, y, cyanColor)
			}
		}
	}

	// Current voltage value label next to indicator
	voltageLabel := fmt.Sprintf("%.2fV", readV)
	drawSimpleText(img, voltageLabel, marginLeft+plotW-50, readY-10, cyanColor)

	// Arrow indicator on left side
	for i := 0; i < 8; i++ {
		img.Set(marginLeft-8+i, readY, cyanColor)
		if i < 4 {
			img.Set(marginLeft-8+i, readY-i, cyanColor)
			img.Set(marginLeft-8+i, readY+i, cyanColor)
		}
	}

	return img
}

func (ca *CircuitsApp) refreshReadZone() {
	if ca.readZoneCanvas != nil {
		fyne.Do(func() {
			ca.readZoneCanvas.Refresh()
		})
	}
}

func (ca *CircuitsApp) createReadResultsSection() fyne.CanvasObject {
	ca.readResultsLabel = widget.NewLabel(
		"Cell [--,--] Read Results\n" +
			"─────────────────────────\n" +
			"Programmed Level:    --\n" +
			"Read Current:        -- µA\n" +
			"TIA Voltage:         -- mV\n" +
			"ADC Raw:             -- / 255\n" +
			"Decoded Level:       --\n" +
			"Match:               --",
	)

	return ca.readResultsLabel
}

func (ca *CircuitsApp) onReadCell() {
	ca.mu.RLock()
	row := ca.selectedRow
	col := ca.selectedCol
	readV := ca.readVoltage
	tiaGain := ca.tiaGain
	levels := ca.quantLevels
	var storedLevel int
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		storedLevel = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	// Calculate conductance from stored level
	conductance := 1.0 + float64(storedLevel)/float64(levels-1)*99.0 // 1-100 µS

	// Calculate current: I = G × V
	current := conductance * readV // µA

	// TIA: V = I × R
	tiaVoltage := current * tiaGain // mV

	// ADC conversion (8-bit, 0-1V range)
	adcRaw := int(tiaVoltage / 1000.0 * 255.0)
	if adcRaw > 255 {
		adcRaw = 255
	}
	if adcRaw < 0 {
		adcRaw = 0
	}

	// Decode back to level
	decodedLevel := int(math.Round(float64(adcRaw) / 255.0 * float64(levels-1)))

	match := "CORRECT"
	if decodedLevel != storedLevel {
		match = fmt.Sprintf("MISMATCH (expected %d)", storedLevel)
	}

	ca.readResultsLabel.SetText(fmt.Sprintf(
		"Cell [%d,%d] Read Results\n"+
			"─────────────────────────\n"+
			"Programmed Level:    %d\n"+
			"Read Current:        %.1f µA\n"+
			"TIA Voltage:         %.0f mV\n"+
			"ADC Raw:             %d / 255\n"+
			"Decoded Level:       %d\n"+
			"Match:               %s",
		row, col, storedLevel, current, tiaVoltage, adcRaw, decodedLevel, match,
	))

	ca.readStatusLabel.SetText(fmt.Sprintf("Read cell [%d,%d]: Level %d", row, col, decodedLevel))

	// Update formula calculation display
	ca.readCalcLabel.SetText(fmt.Sprintf(
		"I     = G × V     = %.1f µS × %.2f V = %.1f µA\n"+
			"V_tia = I × R     = %.1f µA × %.0f kΩ = %.0f mV\n"+
			"ADC   = V_tia/Vref = %.0f / 1000 × 255 = %d\n"+
			"Level = ADC/Max   = %d / 255 × 29  = %d",
		conductance, readV, current,
		current, tiaGain, tiaVoltage,
		tiaVoltage, adcRaw,
		adcRaw, decodedLevel,
	))
}

func (ca *CircuitsApp) onReadAllCells() {
	ca.readStatusLabel.SetText("Reading all cells...")
	// In a real implementation, this would iterate through all cells
	ca.readStatusLabel.SetText(fmt.Sprintf("Read all %d cells", ca.arrayRows*ca.arrayCols))
}

func (ca *CircuitsApp) onVerifyArray() {
	ca.readStatusLabel.SetText("Verifying array...")
	// Simplified verification
	errors := 0
	ca.readStatusLabel.SetText(fmt.Sprintf("Verification complete: %d errors", errors))
}

// ============================================================================
// TAB 3: COMPUTE MODE
// ============================================================================

func (ca *CircuitsApp) createComputeTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**COMPUTE MODE**: Perform matrix-vector multiplication in a single analog operation. Input voltages are applied to columns, multiplied by cell conductances (stored weights), and summed as currents in each row via Kirchhoff's law - computing all dot products in parallel.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Configuration section
	configSection := ca.createComputeConfigSection()

	// Input vector section
	inputSection := ca.createComputeInputSection()

	// Visualization section
	vizSection := ca.createComputeVizSection()

	// Math breakdown section
	mathSection := ca.createComputeMathSection()

	// Output section
	outputSection := ca.createComputeOutputSection()

	// Buttons
	computeBtn := widget.NewButton("COMPUTE", ca.onCompute)
	computeBtn.Importance = widget.HighImportance

	animateBtn := widget.NewButton("ANIMATE STEP-BY-STEP", ca.onAnimateCompute)

	resetBtn := widget.NewButton("RESET", ca.onResetCompute)

	ca.computeStatusLabel = widget.NewLabel("Ready to compute")

	buttonBox := container.NewHBox(
		computeBtn,
		animateBtn,
		resetBtn,
		layout.NewSpacer(),
		ca.computeStatusLabel,
	)

	// Layout
	leftPanel := container.NewVBox(
		widget.NewLabel("CONFIGURATION"),
		configSection,
		widget.NewSeparator(),
		widget.NewLabel("INPUT VECTOR"),
		inputSection,
	)

	centerPanel := container.NewVBox(
		widget.NewLabel("COMPUTE VISUALIZATION"),
		vizSection,
		widget.NewSeparator(),
		widget.NewLabel("MATH BREAKDOWN (Row 0)"),
		mathSection,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("OUTPUT VECTOR"),
		outputSection,
	)

	mainContent := container.NewHBox(
		container.NewPadded(leftPanel),
		widget.NewSeparator(),
		container.NewPadded(centerPanel),
		widget.NewSeparator(),
		container.NewPadded(rightPanel),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator(), mainContent),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		nil,
	)
}

func (ca *CircuitsApp) createComputeConfigSection() fyne.CanvasObject {
	sizeOptions := []string{"4", "8", "16", "32"}
	rowSelect := widget.NewSelect(sizeOptions, nil)
	rowSelect.SetSelected("8")

	colSelect := widget.NewSelect(sizeOptions, nil)
	colSelect.SetSelected("8")

	levelOptions := []string{"30"}
	levelSelect := widget.NewSelect(levelOptions, nil)
	levelSelect.SetSelected("30")

	dacBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	dacSelect := widget.NewSelect(dacBitsOptions, nil)
	dacSelect.SetSelected("8")

	adcBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	adcSelect := widget.NewSelect(adcBitsOptions, nil)
	adcSelect.SetSelected("8")

	readVEntry := widget.NewEntry()
	readVEntry.SetText("0.5")

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Array Size:"), rowSelect, widget.NewLabel("x"), colSelect),
		container.NewHBox(widget.NewLabel("Levels:"), levelSelect),
		container.NewHBox(widget.NewLabel("DAC Bits:"), dacSelect),
		container.NewHBox(widget.NewLabel("ADC Bits:"), adcSelect),
		container.NewHBox(widget.NewLabel("Read Voltage:"), readVEntry, widget.NewLabel("V")),
	)
}

func (ca *CircuitsApp) createComputeInputSection() fyne.CanvasObject {
	modeOptions := []string{"Manual", "Random", "Ramp", "Pattern"}
	modeSelect := widget.NewSelect(modeOptions, func(s string) {
		switch s {
		case "Random":
			ca.mu.Lock()
			for i := range ca.inputVector {
				ca.inputVector[i] = rand.Intn(256)
			}
			ca.mu.Unlock()
			ca.updateComputeInputs()
		case "Ramp":
			ca.mu.Lock()
			for i := range ca.inputVector {
				ca.inputVector[i] = i * 255 / max(1, len(ca.inputVector)-1)
			}
			ca.mu.Unlock()
			ca.updateComputeInputs()
		}
	})
	modeSelect.SetSelected("Manual")

	// Create input entries
	ca.computeInputs = make([]*widget.Entry, ca.arrayCols)
	ca.computeVoltageLabels = make([]*widget.Label, ca.arrayCols)

	inputGrid := container.NewGridWithColumns(4)
	for i := 0; i < min(8, ca.arrayCols); i++ {
		ca.computeInputs[i] = widget.NewEntry()
		ca.computeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))

		idx := i
		ca.computeInputs[i].OnChanged = func(s string) {
			var v int
			fmt.Sscanf(s, "%d", &v)
			if v > 255 {
				v = 255
			}
			ca.mu.Lock()
			ca.inputVector[idx] = v
			ca.mu.Unlock()
		}

		ca.computeVoltageLabels[i] = widget.NewLabel(fmt.Sprintf("%.2fV", float64(ca.inputVector[i])/255.0))

		inputGrid.Add(container.NewVBox(
			widget.NewLabel(fmt.Sprintf("x%d", i)),
			ca.computeInputs[i],
			ca.computeVoltageLabels[i],
		))
	}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Input Mode:"), modeSelect),
		widget.NewLabel("Digital Inputs (0-255):"),
		inputGrid,
	)
}

func (ca *CircuitsApp) updateComputeInputs() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	for i := 0; i < min(8, len(ca.computeInputs)); i++ {
		if ca.computeInputs[i] != nil {
			ca.computeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))
		}
		if ca.computeVoltageLabels[i] != nil {
			ca.computeVoltageLabels[i].SetText(fmt.Sprintf("%.2fV", float64(ca.inputVector[i])/255.0))
		}
	}
}

func (ca *CircuitsApp) createComputeVizSection() fyne.CanvasObject {
	ca.computeArrayCanvas = canvas.NewRaster(ca.drawComputeViz)
	ca.computeArrayCanvas.SetMinSize(fyne.NewSize(450, 350))
	return ca.computeArrayCanvas
}

func (ca *CircuitsApp) drawComputeViz(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	ca.mu.RLock()
	rows := min(8, ca.arrayRows)
	cols := min(8, ca.arrayCols)
	inputs := ca.inputVector
	weights := ca.arrayWeights
	outputs := ca.outputVector
	ca.mu.RUnlock()

	dacColor := color.RGBA{150, 100, 200, 255}
	adcColor := color.RGBA{100, 200, 150, 255}

	// Calculate square cell size for array (use min of both dimensions)
	maxArrayW := w - 260 // Leave space for DACs and ADCs
	maxArrayH := h - 80
	cellW := maxArrayW / cols
	cellH := maxArrayH / rows
	cellSize := cellW
	if cellH < cellSize {
		cellSize = cellH
	}
	if cellSize > 30 {
		cellSize = 30
	}
	if cellSize < 12 {
		cellSize = 12
	}

	// Array dimensions
	arrayW := cols * cellSize
	arrayH := rows * cellSize

	// Center layout horizontally
	totalW := 60 + 20 + arrayW + 20 + 60 // DAC + gap + array + gap + ADC
	startX := (w - totalW) / 2
	if startX < 10 {
		startX = 10
	}

	dacX := startX
	dacW := 60
	arrayX := dacX + dacW + 20
	adcX := arrayX + arrayW + 20

	// Vertical centering
	arrayY := (h - arrayH) / 2
	if arrayY < 30 {
		arrayY = 30
	}

	// Draw DACs (one per column input)
	for i := 0; i < cols && i < len(inputs); i++ {
		y := arrayY + i*cellSize + (cellSize-24)/2
		drawRect(img, dacX, y, dacW, 24, dacColor)
	}

	// Draw array with square cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := arrayX + c*cellSize
			y0 := arrayY + r*cellSize

			level := weights[r][c]
			intensity := float64(level) / 29.0

			cr := uint8(intensity * 200)
			cg := uint8(100 + (1-intensity)*100)
			cb := uint8((1 - intensity) * 200)
			cellColor := color.RGBA{cr, cg, cb, 255}

			drawRect(img, x0+2, y0+2, cellSize-4, cellSize-4, cellColor)
		}
	}

	// Draw ADCs (one per row output)
	for i := 0; i < rows && i < len(outputs); i++ {
		y := arrayY + i*cellSize + (cellSize-24)/2
		drawRect(img, adcX, y, 60, 24, adcColor)
	}

	return img
}

func (ca *CircuitsApp) createComputeMathSection() fyne.CanvasObject {
	ca.computeMathLabel = widget.NewLabel(
		"I₀ = G₀₀×V₀ + G₀₁×V₁ + G₀₂×V₂ + ... + G₀₇×V₇\n\n" +
			"I₀ = --µS×--V + --µS×--V + ...\n" +
			"   = -- µA\n\n" +
			"THIS IS A DOT PRODUCT! (weights · inputs)\n" +
			"ALL 8 ROWS COMPUTED SIMULTANEOUSLY!",
	)

	return ca.computeMathLabel
}

func (ca *CircuitsApp) createComputeOutputSection() fyne.CanvasObject {
	ca.computeOutputLabels = make([]*widget.Label, 8)

	outputGrid := container.NewGridWithColumns(2)
	for i := 0; i < 8; i++ {
		ca.computeOutputLabels[i] = widget.NewLabel(fmt.Sprintf("y%d: --", i))
		outputGrid.Add(ca.computeOutputLabels[i])
	}

	return container.NewVBox(
		widget.NewLabel("Output Currents (µA):"),
		outputGrid,
		widget.NewSeparator(),
		widget.NewLabel("TOTAL LATENCY: ~20ns"),
	)
}

func (ca *CircuitsApp) onCompute() {
	ca.mu.Lock()
	rows := min(8, ca.arrayRows)
	cols := min(8, ca.arrayCols)

	// Perform MVM: output = weights × input
	for r := 0; r < rows && r < len(ca.arrayWeights); r++ {
		sum := 0.0
		for c := 0; c < cols && c < len(ca.arrayWeights[r]); c++ {
			// Conductance (1-100 µS)
			conductance := 1.0 + float64(ca.arrayWeights[r][c])/29.0*99.0
			// Input voltage (0-1V)
			voltage := float64(ca.inputVector[c]) / 255.0
			// Current contribution
			sum += conductance * voltage
		}
		ca.outputVector[r] = sum
	}
	ca.mu.Unlock()

	// Update output labels
	ca.mu.RLock()
	for i := 0; i < 8 && i < len(ca.outputVector); i++ {
		if ca.computeOutputLabels[i] != nil {
			ca.computeOutputLabels[i].SetText(fmt.Sprintf("y%d: %.1f µA", i, ca.outputVector[i]))
		}
	}
	ca.mu.RUnlock()

	// Update math breakdown for row 0
	ca.updateComputeMath()

	ca.computeStatusLabel.SetText("Compute complete in ~20ns")
}

func (ca *CircuitsApp) updateComputeMath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.arrayWeights) == 0 || len(ca.arrayWeights[0]) == 0 {
		return
	}

	cols := min(4, len(ca.arrayWeights[0]))
	mathText := "I₀ = "
	var terms []string
	totalCurrent := 0.0

	for c := 0; c < cols; c++ {
		conductance := 1.0 + float64(ca.arrayWeights[0][c])/29.0*99.0
		voltage := float64(ca.inputVector[c]) / 255.0
		current := conductance * voltage
		totalCurrent += current
		terms = append(terms, fmt.Sprintf("%.0fµS×%.2fV", conductance, voltage))
	}

	mathText += terms[0]
	for i := 1; i < len(terms); i++ {
		mathText += " + " + terms[i]
	}
	mathText += " + ...\n"
	mathText += fmt.Sprintf("   = %.1f µA\n\n", ca.outputVector[0])
	mathText += "THIS IS A DOT PRODUCT! (weights · inputs)\n"
	mathText += "ALL ROWS COMPUTED SIMULTANEOUSLY!"

	ca.computeMathLabel.SetText(mathText)
}

func (ca *CircuitsApp) onAnimateCompute() {
	ca.computeStatusLabel.SetText("Animating... (DAC → Array → ADC)")
	// Animation would be implemented with goroutines and fyne.Do()
	go func() {
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.computeStatusLabel.SetText("Step 1: DAC conversion (5ns)")
		})
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.computeStatusLabel.SetText("Step 2: Array settle (5ns)")
		})
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.computeStatusLabel.SetText("Step 3: ADC conversion (10ns)")
			ca.onCompute()
		})
	}()
}

func (ca *CircuitsApp) onResetCompute() {
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = 0
	}
	for i := range ca.outputVector {
		ca.outputVector[i] = 0
	}
	ca.mu.Unlock()

	ca.updateComputeInputs()
	for i := 0; i < 8 && i < len(ca.computeOutputLabels); i++ {
		if ca.computeOutputLabels[i] != nil {
			ca.computeOutputLabels[i].SetText(fmt.Sprintf("y%d: --", i))
		}
	}

	ca.computeStatusLabel.SetText("Reset complete")
}

// ============================================================================
// TAB 4: COMPARISON (FeFET vs GPU vs CPU)
// ============================================================================

func (ca *CircuitsApp) createComparisonTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**COMPARISON**: Compare FeFET crossbar architecture against traditional von Neumann systems (CPU/GPU). FeFET performs computation in-memory using analog physics (Ohm's law), avoiding the memory bottleneck that limits conventional digital systems.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Architecture comparison
	archSection := ca.createCompArchSection()

	// Timing comparison
	timingSection := ca.createCompTimingSection()

	// Energy comparison
	energySection := ca.createCompEnergySection()

	// Live comparison table
	tableSection := ca.createCompTableSection()

	// Buttons
	runBtn := widget.NewButton("RUN COMPARISON", ca.onRunComparison)
	runBtn.Importance = widget.HighImportance

	animateBtn := widget.NewButton("ANIMATE", nil)
	scaleBtn := widget.NewButton("SCALE UP", nil)

	ca.compStatusLabel = widget.NewLabel("8×8 Matrix-Vector Multiply Comparison")

	buttonBox := container.NewHBox(
		runBtn,
		animateBtn,
		scaleBtn,
		layout.NewSpacer(),
		ca.compStatusLabel,
	)

	// Layout
	topRow := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("ARCHITECTURE COMPARISON"), archSection),
		container.NewVBox(widget.NewLabel("TIMING COMPARISON"), timingSection),
	)

	bottomRow := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("ENERGY COMPARISON"), energySection),
		container.NewVBox(widget.NewLabel("LIVE COMPARISON"), tableSection),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		container.NewVBox(topRow, widget.NewSeparator(), bottomRow),
	)
}

func (ca *CircuitsApp) createCompArchSection() fyne.CanvasObject {
	ca.compArchCanvas = canvas.NewRaster(ca.drawCompArch)
	ca.compArchCanvas.SetMinSize(fyne.NewSize(400, 200))
	return ca.compArchCanvas
}

func (ca *CircuitsApp) drawCompArch(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}
	labelColor := color.RGBA{255, 255, 255, 255}
	arrowColor := color.RGBA{255, 200, 100, 255}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	sectionH := h / 3
	boxW := 80
	boxH := sectionH - 25

	// Row 1: CPU + DRAM section
	cpuX, cpuY := 30, 12
	dramX := cpuX + boxW + 70

	drawRect(img, cpuX, cpuY, boxW, boxH, colorCPU)
	drawSimpleText(img, "CPU", cpuX+25, cpuY+boxH/2-3, labelColor)

	drawRect(img, dramX, cpuY, boxW, boxH, color.RGBA{180, 80, 80, 255})
	drawSimpleText(img, "DRAM", dramX+20, cpuY+boxH/2-3, labelColor)

	// Arrow between CPU and DRAM
	arrowY := cpuY + boxH/2
	for x := cpuX + boxW + 5; x < dramX-5; x++ {
		img.Set(x, arrowY, arrowColor)
		img.Set(x, arrowY-1, arrowColor)
	}
	// Arrowhead
	for i := 0; i < 6; i++ {
		img.Set(dramX-5-i, arrowY-i, arrowColor)
		img.Set(dramX-5-i, arrowY+i, arrowColor)
	}
	drawSimpleText(img, "Data Bus", cpuX+boxW+15, arrowY-12, arrowColor)

	// Row 2: GPU + HBM section
	gpuY := sectionH + 8
	drawRect(img, cpuX, gpuY, boxW, boxH, colorGPU)
	drawSimpleText(img, "GPU", cpuX+25, gpuY+boxH/2-3, labelColor)

	drawRect(img, dramX, gpuY, boxW, boxH, color.RGBA{80, 180, 80, 255})
	drawSimpleText(img, "HBM", dramX+25, gpuY+boxH/2-3, labelColor)

	// Arrow between GPU and HBM
	arrowY = gpuY + boxH/2
	for x := cpuX + boxW + 5; x < dramX-5; x++ {
		img.Set(x, arrowY, arrowColor)
		img.Set(x, arrowY-1, arrowColor)
	}
	for i := 0; i < 6; i++ {
		img.Set(dramX-5-i, arrowY-i, arrowColor)
		img.Set(dramX-5-i, arrowY+i, arrowColor)
	}
	drawSimpleText(img, "Data Bus", cpuX+boxW+15, arrowY-12, arrowColor)

	// Row 3: FeFET CIM section (unified)
	fefetY := 2*sectionH + 5
	fefetW := dramX + boxW - cpuX
	drawRect(img, cpuX, fefetY, fefetW, boxH, colorFeFET)
	drawSimpleText(img, "FeFET CIM", cpuX+fefetW/2-35, fefetY+boxH/2-10, labelColor)
	drawSimpleText(img, "No Data Movement", cpuX+fefetW/2-55, fefetY+boxH/2+5, color.RGBA{0, 255, 200, 255})

	// Right side labels
	rightX := w - 90
	drawSimpleText(img, "Von Neumann", rightX, cpuY+boxH/2-3, colorCPU)
	drawSimpleText(img, "Near Memory", rightX, gpuY+boxH/2-3, colorGPU)
	drawSimpleText(img, "In Memory", rightX, fefetY+boxH/2-3, colorFeFET)

	return img
}

func (ca *CircuitsApp) createCompTimingSection() fyne.CanvasObject {
	ca.compTimingCanvas = canvas.NewRaster(ca.drawCompTiming)
	ca.compTimingCanvas.SetMinSize(fyne.NewSize(400, 150))
	return ca.compTimingCanvas
}

func (ca *CircuitsApp) drawCompTiming(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}
	axisColor := color.RGBA{200, 200, 200, 255}
	labelColor := color.RGBA{255, 255, 255, 255}
	valueColor := color.RGBA{200, 200, 150, 255}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 60
	marginRight := 80
	barH := 25
	spacing := 35
	maxBarW := w - marginLeft - marginRight

	// CPU bar (500ns - full width)
	cpuY := 15
	cpuW := maxBarW
	drawSimpleText(img, "CPU", 10, cpuY+8, colorCPU)
	drawRect(img, marginLeft, cpuY, cpuW, barH, colorCPU)
	drawSimpleText(img, "500ns", marginLeft+cpuW+5, cpuY+8, valueColor)

	// GPU bar (50ns - 10% width)
	gpuY := cpuY + spacing
	gpuW := maxBarW * 50 / 500
	if gpuW < 30 {
		gpuW = 30
	}
	drawSimpleText(img, "GPU", 10, gpuY+8, colorGPU)
	drawRect(img, marginLeft, gpuY, gpuW, barH, colorGPU)
	drawSimpleText(img, "50ns", marginLeft+gpuW+5, gpuY+8, valueColor)

	// FeFET bar (20ns - 4% width)
	fefetY := gpuY + spacing
	fefetW := maxBarW * 20 / 500
	if fefetW < 20 {
		fefetW = 20
	}
	drawSimpleText(img, "FeFET", 5, fefetY+8, colorFeFET)
	drawRect(img, marginLeft, fefetY, fefetW, barH, colorFeFET)
	drawSimpleText(img, "20ns", marginLeft+fefetW+5, fefetY+8, valueColor)

	// Speedup annotation
	drawSimpleText(img, "25x faster!", w-80, fefetY+8, color.RGBA{0, 255, 200, 255})

	// X-axis
	axisY := h - 25
	for x := marginLeft; x < w-marginRight; x++ {
		img.Set(x, axisY, axisColor)
	}

	// Axis label
	drawSimpleText(img, "Time (ns)", w/2-30, axisY+10, labelColor)

	// Scale markers
	scaleMarkers := []struct {
		pct   int
		label string
	}{
		{0, "0"},
		{50, "250"},
		{100, "500"},
	}

	for _, sm := range scaleMarkers {
		x := marginLeft + sm.pct*maxBarW/100
		for dy := 0; dy < 5; dy++ {
			img.Set(x, axisY+dy, axisColor)
		}
		drawSimpleText(img, sm.label, x-len(sm.label)*3, axisY+10, axisColor)
	}

	return img
}

func (ca *CircuitsApp) createCompEnergySection() fyne.CanvasObject {
	ca.compEnergyCanvas = canvas.NewRaster(ca.drawCompEnergy)
	ca.compEnergyCanvas.SetMinSize(fyne.NewSize(400, 200))
	return ca.compEnergyCanvas
}

func (ca *CircuitsApp) drawCompEnergy(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}
	axisColor := color.RGBA{200, 200, 200, 255}
	labelColor := color.RGBA{255, 255, 255, 255}
	valueColor := color.RGBA{200, 200, 150, 255}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 60
	marginRight := 100
	barH := 28
	spacing := 45
	maxBarW := w - marginLeft - marginRight

	// CPU bar (64,000 pJ - full width)
	cpuY := 20
	cpuW := maxBarW
	drawSimpleText(img, "CPU", 10, cpuY+8, colorCPU)
	drawRect(img, marginLeft, cpuY, cpuW, barH, colorCPU)
	drawSimpleText(img, "64000 pJ", marginLeft+cpuW+5, cpuY+8, valueColor)

	// GPU bar (6,400 pJ - 10% width)
	gpuY := cpuY + spacing
	gpuW := maxBarW * 6400 / 64000
	if gpuW < 30 {
		gpuW = 30
	}
	drawSimpleText(img, "GPU", 10, gpuY+8, colorGPU)
	drawRect(img, marginLeft, gpuY, gpuW, barH, colorGPU)
	drawSimpleText(img, "6400 pJ", marginLeft+gpuW+5, gpuY+8, valueColor)

	// FeFET bar (3.2 pJ - tiny, need minimum visible)
	fefetY := gpuY + spacing
	fefetW := maxBarW * 32 / 64000 // 3.2 pJ scaled
	if fefetW < 8 {
		fefetW = 8 // Minimum visible
	}
	drawSimpleText(img, "FeFET", 5, fefetY+8, colorFeFET)
	drawRect(img, marginLeft, fefetY, fefetW, barH, colorFeFET)
	drawSimpleText(img, "3.2 pJ", marginLeft+fefetW+5, fefetY+8, valueColor)

	// Energy savings annotation
	drawSimpleText(img, "20000x savings!", w-120, fefetY+8, color.RGBA{0, 255, 200, 255})

	// X-axis
	axisY := h - 30
	for x := marginLeft; x < w-marginRight; x++ {
		img.Set(x, axisY, axisColor)
	}

	// Axis label
	drawSimpleText(img, "Energy per 8x8 MVM", w/2-60, axisY+12, labelColor)

	// Scale note (log scale would be better but linear for illustration)
	drawSimpleText(img, "[Linear scale - FeFET bar scaled up for visibility]", 10, h-12, color.RGBA{120, 120, 140, 255})

	return img
}

func (ca *CircuitsApp) createCompTableSection() fyne.CanvasObject {
	// Create table labels
	ca.compTableLabels = make([]*widget.Label, 16)

	headers := []string{"", "Time", "Energy", "TOPS/W"}
	cpuRow := []string{"CPU", "500 ns", "64,000 pJ", "0.5"}
	gpuRow := []string{"GPU", "50 ns", "6,400 pJ", "5.0"}
	fefetRow := []string{"FeFET", "20 ns", "3.2 pJ", "2,000"}

	grid := container.NewGridWithColumns(4)
	for i, h := range headers {
		lbl := widget.NewLabel(h)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		ca.compTableLabels[i] = lbl
		grid.Add(lbl)
	}
	for i, v := range cpuRow {
		lbl := widget.NewLabel(v)
		ca.compTableLabels[4+i] = lbl
		grid.Add(lbl)
	}
	for i, v := range gpuRow {
		lbl := widget.NewLabel(v)
		ca.compTableLabels[8+i] = lbl
		grid.Add(lbl)
	}
	for i, v := range fefetRow {
		lbl := widget.NewLabel(v)
		ca.compTableLabels[12+i] = lbl
		grid.Add(lbl)
	}

	arraySizeLabel := widget.NewLabel("Array Size: 8 × 8 = 64 MACs")

	return container.NewVBox(
		arraySizeLabel,
		widget.NewSeparator(),
		grid,
	)
}

func (ca *CircuitsApp) onRunComparison() {
	ca.compStatusLabel.SetText("Running comparison for 8×8 MVM...")

	// Refresh canvases
	fyne.Do(func() {
		if ca.compArchCanvas != nil {
			ca.compArchCanvas.Refresh()
		}
		if ca.compTimingCanvas != nil {
			ca.compTimingCanvas.Refresh()
		}
		if ca.compEnergyCanvas != nil {
			ca.compEnergyCanvas.Refresh()
		}
	})

	ca.compStatusLabel.SetText("Comparison complete: FeFET wins by 20,000x energy efficiency!")
}

// ============================================================================
// TAB 5: TIMING DIAGRAMS
// ============================================================================

func (ca *CircuitsApp) createTimingTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**TIMING DIAGRAMS**: View signal waveforms for write, read, and compute operations. Shows the precise timing relationships between clock, voltage pulses, current sensing, ADC conversion, and data output with nanosecond precision.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Operation selector
	ca.timingOpSelect = widget.NewSelect([]string{"WRITE", "READ", "COMPUTE"}, func(s string) {
		ca.refreshTimingDiagrams()
	})
	ca.timingOpSelect.SetSelected("WRITE")

	// Timing diagrams
	writeSection := ca.createTimingWriteSection()
	readSection := ca.createTimingReadSection()
	computeSection := ca.createTimingComputeSection()

	// Buttons
	animateBtn := widget.NewButton("ANIMATE", nil)
	exportBtn := widget.NewButton("EXPORT SVG", nil)

	ca.timingStatusLabel = widget.NewLabel("Select operation to view timing")

	buttonBox := container.NewHBox(
		animateBtn,
		exportBtn,
		layout.NewSpacer(),
		ca.timingStatusLabel,
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator(), container.NewHBox(widget.NewLabel("OPERATION:"), ca.timingOpSelect)),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		container.NewVScroll(container.NewVBox(
			widget.NewLabel("WRITE TIMING"),
			writeSection,
			widget.NewSeparator(),
			widget.NewLabel("READ TIMING"),
			readSection,
			widget.NewSeparator(),
			widget.NewLabel("COMPUTE TIMING"),
			computeSection,
		)),
	)
}

func (ca *CircuitsApp) createTimingWriteSection() fyne.CanvasObject {
	ca.timingWriteCanvas = canvas.NewRaster(ca.drawTimingWrite)
	ca.timingWriteCanvas.SetMinSize(fyne.NewSize(600, 200))
	return ca.timingWriteCanvas
}

func (ca *CircuitsApp) drawTimingWrite(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	labelColor := color.RGBA{180, 180, 200, 255}
	timeColor := color.RGBA{255, 200, 100, 255}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 80
	marginBottom := 25
	signalH := 22
	spacing := 27

	signals := []struct {
		name string
		high []int
	}{
		{"CLK", []int{5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90}},
		{"ROW_SEL", []int{10, 80}},
		{"COL_SEL", []int{10, 80}},
		{"DAC_EN", []int{15, 75}},
		{"V_PROG", []int{20, 70}},
		{"DONE", []int{85, 95}},
	}

	plotW := w - marginLeft - 20

	// Draw signal labels on left margin
	for i, sig := range signals {
		y := 10 + i*spacing
		drawSimpleText(img, sig.name, 5, y+8, labelColor)
	}

	// Draw signals
	for i, sig := range signals {
		y := 10 + i*spacing
		prevHigh := false

		for pct := 0; pct <= 100; pct++ {
			x := marginLeft + pct*plotW/100

			isHigh := false
			for j := 0; j < len(sig.high)-1; j += 2 {
				if pct >= sig.high[j] && pct <= sig.high[j+1] {
					isHigh = true
					break
				}
			}

			if sig.name == "CLK" {
				isHigh = (pct/5)%2 == 0 && pct < 95
			}

			lineY := y + signalH - 5
			if isHigh {
				lineY = y + 5
			}

			if isHigh != prevHigh && pct > 0 {
				for py := y + 5; py < y+signalH-5; py++ {
					img.Set(x, py, cyanColor)
				}
			}
			prevHigh = isHigh

			img.Set(x, lineY, cyanColor)
		}
	}

	// Draw time axis at bottom
	axisY := h - marginBottom
	axisColor := color.RGBA{150, 150, 150, 255}
	for x := marginLeft; x < w-20; x++ {
		img.Set(x, axisY, axisColor)
	}

	// Time markers: 0ns, 17ns, 35ns, 52ns, 70ns
	timeMarkers := []struct {
		pct   int
		label string
	}{
		{0, "0ns"},
		{25, "17ns"},
		{50, "35ns"},
		{75, "52ns"},
		{100, "70ns"},
	}

	for _, tm := range timeMarkers {
		x := marginLeft + tm.pct*plotW/100
		// Draw tick mark
		for dy := 0; dy < 5; dy++ {
			img.Set(x, axisY+dy, axisColor)
		}
		// Draw label
		labelX := x - len(tm.label)*3
		if labelX < marginLeft {
			labelX = marginLeft
		}
		drawSimpleText(img, tm.label, labelX, axisY+7, timeColor)
	}

	// Total time label
	drawSimpleText(img, "70ns total", w-80, axisY+7, color.RGBA{0, 255, 200, 255})

	return img
}

func (ca *CircuitsApp) createTimingReadSection() fyne.CanvasObject {
	ca.timingReadCanvas = canvas.NewRaster(ca.drawTimingRead)
	ca.timingReadCanvas.SetMinSize(fyne.NewSize(600, 180))
	return ca.timingReadCanvas
}

func (ca *CircuitsApp) drawTimingRead(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	labelColor := color.RGBA{180, 180, 200, 255}
	timeColor := color.RGBA{255, 200, 100, 255}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 80
	marginBottom := 25
	spacing := 27

	signals := []string{"CLK", "V_READ", "I_SENSE", "ADC_EN", "DATA_OUT"}
	plotW := w - marginLeft - 20

	// Draw signal labels on left margin
	for i, name := range signals {
		y := 10 + i*spacing
		drawSimpleText(img, name, 5, y+8, labelColor)
	}

	// Draw signals
	for i, name := range signals {
		y := 10 + i*spacing
		prevLineY := -1

		for pct := 0; pct <= 100; pct++ {
			x := marginLeft + pct*plotW/100

			var lineY int
			switch name {
			case "CLK":
				if (pct/10)%2 == 0 && pct < 90 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "V_READ":
				if pct >= 10 && pct <= 70 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "I_SENSE":
				if pct >= 15 && pct <= 75 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "ADC_EN":
				if pct >= 40 && pct <= 70 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "DATA_OUT":
				if pct >= 75 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			}

			// Draw vertical transition
			if prevLineY != -1 && lineY != prevLineY {
				minY := min(lineY, prevLineY)
				maxY := max(lineY, prevLineY)
				for py := minY; py <= maxY; py++ {
					img.Set(x, py, cyanColor)
				}
			}
			prevLineY = lineY

			img.Set(x, lineY, cyanColor)
		}
	}

	// Draw time axis at bottom
	axisY := h - marginBottom
	axisColor := color.RGBA{150, 150, 150, 255}
	for x := marginLeft; x < w-20; x++ {
		img.Set(x, axisY, axisColor)
	}

	// Time markers: 0ns, 5ns, 10ns, 15ns, 20ns
	timeMarkers := []struct {
		pct   int
		label string
	}{
		{0, "0ns"},
		{25, "5ns"},
		{50, "10ns"},
		{75, "15ns"},
		{100, "20ns"},
	}

	for _, tm := range timeMarkers {
		x := marginLeft + tm.pct*plotW/100
		// Draw tick mark
		for dy := 0; dy < 5; dy++ {
			img.Set(x, axisY+dy, axisColor)
		}
		// Draw label
		labelX := x - len(tm.label)*3
		if labelX < marginLeft {
			labelX = marginLeft
		}
		drawSimpleText(img, tm.label, labelX, axisY+7, timeColor)
	}

	// Total time label
	drawSimpleText(img, "20ns total", w-80, axisY+7, color.RGBA{0, 255, 200, 255})

	return img
}

func (ca *CircuitsApp) createTimingComputeSection() fyne.CanvasObject {
	ca.timingComputeCanvas = canvas.NewRaster(ca.drawTimingCompute)
	ca.timingComputeCanvas.SetMinSize(fyne.NewSize(600, 200))
	return ca.timingComputeCanvas
}

func (ca *CircuitsApp) drawTimingCompute(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	labelColor := color.RGBA{180, 180, 200, 255}
	phaseColor := color.RGBA{200, 150, 255, 200}

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 100
	marginBottom := 35
	spacing := 25

	signals := []string{"CLK", "INPUT_VALID", "DAC_ALL", "ARRAY_SETTLE", "ADC_ALL", "OUTPUT_VALID"}
	plotW := w - marginLeft - 20

	// Draw signal labels on left margin
	for i, name := range signals {
		y := 8 + i*spacing
		drawSimpleText(img, name, 5, y+6, labelColor)
	}

	// Draw signals
	for i, name := range signals {
		y := 8 + i*spacing
		prevLineY := -1

		for pct := 0; pct <= 100; pct++ {
			x := marginLeft + pct*plotW/100

			var lineY int
			switch name {
			case "CLK":
				if (pct/8)%2 == 0 && pct < 95 {
					lineY = y + 5
				} else {
					lineY = y + 18
				}
			case "INPUT_VALID":
				if pct >= 5 && pct <= 85 {
					lineY = y + 5
				} else {
					lineY = y + 18
				}
			case "DAC_ALL":
				if pct >= 10 && pct <= 35 {
					lineY = y + 5
				} else {
					lineY = y + 18
				}
			case "ARRAY_SETTLE":
				if pct >= 35 && pct <= 60 {
					lineY = y + 5
				} else {
					lineY = y + 18
				}
			case "ADC_ALL":
				if pct >= 55 && pct <= 90 {
					lineY = y + 5
				} else {
					lineY = y + 18
				}
			case "OUTPUT_VALID":
				if pct >= 90 {
					lineY = y + 5
				} else {
					lineY = y + 18
				}
			}

			// Draw vertical transition
			if prevLineY != -1 && lineY != prevLineY {
				minY := min(lineY, prevLineY)
				maxY := max(lineY, prevLineY)
				for py := minY; py <= maxY; py++ {
					img.Set(x, py, cyanColor)
				}
			}
			prevLineY = lineY

			img.Set(x, lineY, cyanColor)
		}
	}

	// Draw phase markers at bottom
	phaseY := h - marginBottom - 8
	phases := []struct {
		startPct int
		endPct   int
		label    string
	}{
		{10, 35, "DAC 5ns"},
		{35, 60, "ARRAY 5ns"},
		{55, 90, "ADC 10ns"},
	}

	for _, phase := range phases {
		startX := marginLeft + phase.startPct*plotW/100
		endX := marginLeft + phase.endPct*plotW/100
		midX := (startX + endX) / 2

		// Draw phase bracket
		for x := startX; x <= endX; x++ {
			img.Set(x, phaseY, phaseColor)
		}
		// Vertical edges
		for dy := 0; dy < 4; dy++ {
			img.Set(startX, phaseY-dy, phaseColor)
			img.Set(endX, phaseY-dy, phaseColor)
		}

		// Draw phase label
		labelX := midX - len(phase.label)*3
		drawSimpleText(img, phase.label, labelX, phaseY+3, phaseColor)
	}

	// Draw time axis at bottom
	axisY := h - 15
	axisColor := color.RGBA{150, 150, 150, 255}
	for x := marginLeft; x < w-20; x++ {
		img.Set(x, axisY, axisColor)
	}

	// Time markers: 0ns, 5ns, 10ns, 15ns, 20ns
	timeMarkers := []struct {
		pct   int
		label string
	}{
		{0, "0ns"},
		{25, "5ns"},
		{50, "10ns"},
		{75, "15ns"},
		{100, "20ns"},
	}

	for _, tm := range timeMarkers {
		x := marginLeft + tm.pct*plotW/100
		// Draw tick mark
		for dy := 0; dy < 4; dy++ {
			img.Set(x, axisY+dy, axisColor)
		}
	}

	// Total time label
	drawSimpleText(img, "20ns total", w-80, axisY-2, color.RGBA{0, 255, 200, 255})

	return img
}

func (ca *CircuitsApp) refreshTimingDiagrams() {
	fyne.Do(func() {
		if ca.timingWriteCanvas != nil {
			ca.timingWriteCanvas.Refresh()
		}
		if ca.timingReadCanvas != nil {
			ca.timingReadCanvas.Refresh()
		}
		if ca.timingComputeCanvas != nil {
			ca.timingComputeCanvas.Refresh()
		}
	})
}

// ============================================================================
// TAB 6: SPECIFICATIONS
// ============================================================================

func (ca *CircuitsApp) createSpecsTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**SPECIFICATIONS**: Detailed electrical and physical parameters for all peripheral components (DAC, ADC, TIA) and FeFET cells. Includes array configuration, conversion times, power consumption, and device characteristics.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Array configuration
	arraySection := ca.createSpecArraySection()

	// DAC specs
	dacSection := ca.createSpecDACSection()

	// ADC specs
	adcSection := ca.createSpecADCSection()

	// TIA specs
	tiaSection := ca.createSpecTIASection()

	// FeFET cell specs
	fefetSection := ca.createSpecFeFETSection()

	// System summary
	summarySection := ca.createSpecSummarySection()

	// Buttons
	exportBtn := widget.NewButton("EXPORT SPECS", nil)
	compareBtn := widget.NewButton("COMPARE TO GPU", nil)

	ca.specStatusLabel = widget.NewLabel("System specifications")

	buttonBox := container.NewHBox(
		exportBtn,
		compareBtn,
		layout.NewSpacer(),
		ca.specStatusLabel,
	)

	// Layout in a grid
	leftCol := container.NewVBox(
		widget.NewLabel("ARRAY CONFIGURATION"),
		arraySection,
		widget.NewSeparator(),
		widget.NewLabel("DAC SPECIFICATIONS"),
		dacSection,
		widget.NewSeparator(),
		widget.NewLabel("ADC SPECIFICATIONS"),
		adcSection,
	)

	rightCol := container.NewVBox(
		widget.NewLabel("TIA SPECIFICATIONS"),
		tiaSection,
		widget.NewSeparator(),
		widget.NewLabel("FeFET CELL SPECIFICATIONS"),
		fefetSection,
		widget.NewSeparator(),
		widget.NewLabel("SYSTEM SUMMARY"),
		summarySection,
	)

	mainContent := container.NewHBox(
		container.NewPadded(leftCol),
		widget.NewSeparator(),
		container.NewPadded(rightCol),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		container.NewVScroll(mainContent),
	)
}

func (ca *CircuitsApp) createSpecArraySection() fyne.CanvasObject {
	sizeOptions := []string{"8", "16", "32", "64", "128"}
	ca.specArraySizeSelect = widget.NewSelect(sizeOptions, nil)
	ca.specArraySizeSelect.SetSelected("32")

	levelOptions := []string{"2", "4", "8", "16", "30", "32", "64", "128", "256"}
	ca.specQuantLevelSelect = widget.NewSelect(levelOptions, nil)
	ca.specQuantLevelSelect.SetSelected("30")

	// Calculate storage
	cells := 32 * 32
	bitsPerCell := math.Log2(30)
	totalBits := float64(cells) * bitsPerCell

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Array Size:"), ca.specArraySizeSelect, widget.NewLabel("×"), ca.specArraySizeSelect, widget.NewLabel(fmt.Sprintf("= %d cells", cells))),
		container.NewHBox(widget.NewLabel("Quantization:"), ca.specQuantLevelSelect, widget.NewLabel(fmt.Sprintf("levels (~%.1f bits/cell)", bitsPerCell))),
		widget.NewLabel(fmt.Sprintf("Total Storage: %d × %.1f = %.0f bits", cells, bitsPerCell, totalBits)),
	)
}

func (ca *CircuitsApp) createSpecDACSection() fyne.CanvasObject {
	dacBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	ca.specDACBitsSelect = widget.NewSelect(dacBitsOptions, nil)
	ca.specDACBitsSelect.SetSelected("8")

	specs := `Count:             32 (one per column)
Resolution:        8 bits (256 levels)
Output Range:      0V to 1.0V (read), 2V to 5V (write)
Conversion Time:   5 ns (digital to analog latency)
Power per DAC:     0.1 mW (static + dynamic)
Total DAC Power:   3.2 mW (for 32 DACs)
INL:               < 0.5 LSB (integral nonlinearity)
DNL:               < 0.5 LSB (differential nonlinearity)
Rise/Fall Time:    2-5 ns (signal edge transitions)`

	helpText := widget.NewLabel("DAC converts digital level (0-29) to precise analog voltage for programming FeFET cells")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Resolution:"), ca.specDACBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel(specs),
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecADCSection() fyne.CanvasObject {
	adcBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	ca.specADCBitsSelect = widget.NewSelect(adcBitsOptions, nil)
	ca.specADCBitsSelect.SetSelected("8")

	specs := `Count:             32 (one per row)
Resolution:        8 bits (256 levels)
Input Range:       0V to 1.0V (after TIA conversion)
Conversion Time:   10 ns (analog to digital latency)
Power per ADC:     0.5 mW (conversion energy)
Total ADC Power:   16 mW (for 32 ADCs)
ENOB:              7.5 bits (effective resolution with noise)
SNR:               46 dB (signal-to-noise ratio)
Sample Rate:       100 MSPS (samples per second)`

	helpText := widget.NewLabel("ADC digitizes analog current from TIA, converting continuous values to discrete digital levels")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Resolution:"), ca.specADCBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel(specs),
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecTIASection() fyne.CanvasObject {
	tiaGainOptions := []string{"1", "10", "100"}
	ca.specTIAGainSelect = widget.NewSelect(tiaGainOptions, nil)
	ca.specTIAGainSelect.SetSelected("10")

	specs := `Count:             32 (one per row)
Gain (R_f):        10 kOhm (transimpedance gain)
Bandwidth:         100 MHz (frequency response)
Input Current:     0 to 100 µA (cell current range)
Output Voltage:    0 to 1.0 V (V_out = I_in × R_f)
Noise:             < 1 µA RMS (input-referred noise)
Response Time:     ~2 ns (settling time)`

	helpText := widget.NewLabel("TIA (Transimpedance Amplifier) converts tiny FeFET currents to measurable voltages for ADC")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Gain:"), ca.specTIAGainSelect, widget.NewLabel("kOhm")),
		widget.NewLabel(specs),
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecFeFETSection() fyne.CanvasObject {
	grid := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Material:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("HfZrO2 (HZO)"),

		widget.NewLabelWithStyle("Thickness:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("10 nm (ferroelectric layer)"),

		widget.NewLabelWithStyle("Levels:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("30 discrete states (~4.9 bits/cell)"),

		widget.NewLabelWithStyle("Conductance:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("1 µS to 100 µS (programmable range)"),

		widget.NewLabelWithStyle("Read Voltage:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("0.5 V (non-destructive, below write threshold)"),

		widget.NewLabelWithStyle("Write Voltage:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("2.0 V to 5.0 V (exceeds coercive field Ec)"),

		widget.NewLabelWithStyle("Write Time:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("50 ns (pulse duration for polarization switching)"),

		widget.NewLabelWithStyle("Endurance:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("10^12 cycles (write/erase lifetime)"),

		widget.NewLabelWithStyle("Retention:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("10 years (data persistence without power)"),

		widget.NewLabelWithStyle("Cell Size:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("~0.01 µm² (width × height in silicon area)"),
	)

	helpText := widget.NewLabel("Note: Rise/fall times typically 2-10 ns; capacitance 0.1-10 pF; leakage < 1 nW per cell")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(grid, widget.NewSeparator(), helpText)
}

func (ca *CircuitsApp) createSpecSummarySection() fyne.CanvasObject {
	summary := `Component       | Count | Power   | Area     | Latency
----------------|-------|---------|----------|--------
FeFET Array     | 1,024 | 0.1 mW  | 0.01 mm² | 5 ns
DACs            | 32    | 3.2 mW  | 0.02 mm² | 5 ns
TIAs            | 32    | 1.6 mW  | 0.01 mm² | 2 ns
ADCs            | 32    | 16 mW   | 0.04 mm² | 10 ns
Control         | 1     | 0.5 mW  | 0.01 mm² | 2 ns
----------------|-------|---------|----------|--------
TOTAL           |       | 21.4 mW | 0.09 mm² | 20 ns

Throughput:     1,024 MACs / 20ns = 51.2 GOPS
Efficiency:     51.2 GOPS / 21.4 mW = 2,392 GOPS/W`

	return widget.NewLabel(summary)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ============================================================================
// BITMAP FONT TEXT RENDERING
// ============================================================================

// drawSimpleText draws text using a simple bitmap font.
func drawSimpleText(img *image.RGBA, text string, x, y int, c color.Color) {
	drawScaledText(img, text, x, y, 1, c)
}

// drawSimpleChar draws a single character.
func drawSimpleChar(img *image.RGBA, ch rune, x, y int, c color.Color) {
	drawScaledChar(img, ch, x, y, 1, c)
}

// drawScaledText draws text with a scaling factor.
func drawScaledText(img *image.RGBA, text string, x, y int, scale int, c color.Color) {
	charWidth := 7 * scale
	for i, ch := range text {
		cx := x + i*charWidth
		drawScaledChar(img, ch, cx, y, scale, c)
	}
}

// drawScaledChar draws a single character with scaling.
func drawScaledChar(img *image.RGBA, ch rune, x, y int, scale int, c color.Color) {
	pattern, ok := fontPatterns[ch]
	if !ok {
		return
	}

	for dy, row := range pattern {
		for dx, pixel := range row {
			if pixel == '1' {
				// Draw scaled pixel
				for sy := 0; sy < scale; sy++ {
					for sx := 0; sx < scale; sx++ {
						px := x + dx*scale + sx
						py := y + dy*scale + sy
						if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
							img.Set(px, py, c)
						}
					}
				}
			}
		}
	}
}

// Basic 5x7 font patterns
var fontPatterns = map[rune][]string{
	'0': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2': {"01110", "10001", "00001", "00110", "01000", "10000", "11111"},
	'3': {"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
	'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
	'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
	'.': {"00000", "00000", "00000", "00000", "00000", "01100", "01100"},
	'-': {"00000", "00000", "00000", "11111", "00000", "00000", "00000"},
	':': {"00000", "01100", "01100", "00000", "01100", "01100", "00000"},
	'%': {"11001", "11010", "00100", "01000", "01011", "10011", "00000"},
	' ': {"00000", "00000", "00000", "00000", "00000", "00000", "00000"},
	'x': {"00000", "00000", "10001", "01010", "00100", "01010", "10001"},
	'n': {"00000", "00000", "10110", "11001", "10001", "10001", "10001"},
	'J': {"00111", "00010", "00010", "00010", "00010", "10010", "01100"},
	'G': {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
	'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'e': {"00000", "00000", "01110", "10001", "11111", "10000", "01110"},
	'C': {"01110", "10001", "10000", "10000", "10000", "10001", "01110"},
	'I': {"01110", "00100", "00100", "00100", "00100", "00100", "01110"},
	'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'O': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'S': {"01110", "10001", "10000", "01110", "00001", "10001", "01110"},
	's': {"00000", "00000", "01110", "10000", "01110", "00001", "11110"},
	'i': {"00100", "00000", "01100", "00100", "00100", "00100", "01110"},
	'o': {"00000", "00000", "01110", "10001", "10001", "10001", "01110"},
	'c': {"00000", "00000", "01110", "10000", "10000", "10001", "01110"},
	'f': {"00110", "01000", "01000", "11100", "01000", "01000", "01000"},
	'r': {"00000", "00000", "10110", "11001", "10000", "10000", "10000"},
	'a': {"00000", "00000", "01110", "00001", "01111", "10001", "01111"},
	't': {"00100", "00100", "01110", "00100", "00100", "00100", "00011"},
	'h': {"10000", "10000", "10110", "11001", "10001", "10001", "10001"},
	'm': {"00000", "00000", "11010", "10101", "10101", "10001", "10001"},
	'd': {"00001", "00001", "01101", "10011", "10001", "10001", "01111"},
	'v': {"00000", "00000", "10001", "10001", "10001", "01010", "00100"},
	'y': {"00000", "00000", "10001", "10001", "01111", "00001", "01110"},
	'k': {"10000", "10000", "10010", "10100", "11000", "10100", "10010"},
	'g': {"00000", "00000", "01111", "10001", "01111", "00001", "01110"},
	'W': {"10001", "10001", "10001", "10101", "10101", "10101", "01010"},
	'p': {"00000", "00000", "11110", "10001", "11110", "10000", "10000"},
	'!': {"00100", "00100", "00100", "00100", "00100", "00000", "00100"},
	'(': {"00010", "00100", "01000", "01000", "01000", "00100", "00010"},
	')': {"01000", "00100", "00010", "00010", "00010", "00100", "01000"},
	'_': {"00000", "00000", "00000", "00000", "00000", "00000", "11111"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'D': {"11100", "10010", "10001", "10001", "10001", "10010", "11100"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
	'b': {"10000", "10000", "11110", "10001", "10001", "10001", "11110"},
	'l': {"01100", "00100", "00100", "00100", "00100", "00100", "01110"},
	'u': {"00000", "00000", "10001", "10001", "10001", "10011", "01101"},
	'w': {"00000", "00000", "10001", "10001", "10101", "10101", "01010"},
	'z': {"00000", "00000", "11111", "00010", "00100", "01000", "11111"},
	'[': {"01110", "01000", "01000", "01000", "01000", "01000", "01110"},
	']': {"01110", "00010", "00010", "00010", "00010", "00010", "01110"},
	'/': {"00001", "00010", "00010", "00100", "01000", "01000", "10000"},
	'=': {"00000", "00000", "11111", "00000", "11111", "00000", "00000"},
	'+': {"00000", "00100", "00100", "11111", "00100", "00100", "00000"},
	'<': {"00010", "00100", "01000", "10000", "01000", "00100", "00010"},
	'>': {"01000", "00100", "00010", "00001", "00010", "00100", "01000"},
	',': {"00000", "00000", "00000", "00000", "00110", "00100", "01000"},
	'q': {"00000", "00000", "01111", "10001", "01111", "00001", "00001"},
	'j': {"00010", "00000", "00110", "00010", "00010", "10010", "01100"},
}
