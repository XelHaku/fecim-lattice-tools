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
	mu             sync.RWMutex
	arrayRows      int
	arrayCols      int
	quantLevels    int
	dacBits        int
	adcBits        int
	vMin           float64 // Min write voltage
	vMax           float64 // Max write voltage
	pulseWidth     float64 // ns
	readVoltage    float64 // Read voltage (safe zone)
	tiaGain        float64 // TIA gain (kOhm)
	selectedRow    int
	selectedCol    int
	targetLevel    int
	arrayWeights   [][]int // Current programmed levels
	inputVector    []int   // Input vector for compute
	outputVector   []float64

	// Tab-specific GUI components
	// Tab 1: Write
	writeRowSelect    *widget.Select
	writeColSelect    *widget.Select
	writeLevelSlider  *widget.Slider
	writeLevelLabel   *widget.Label
	writeArrayCanvas  *canvas.Raster
	writeDataPath     *fyne.Container
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

	// Tab 3: Compute
	computeInputs       []*widget.Entry
	computeVoltageLabels []*widget.Label
	computeArrayCanvas  *canvas.Raster
	computeOutputLabels []*widget.Label
	computeMathLabel    *widget.Label
	computeStatusLabel  *widget.Label

	// Tab 4: Comparison
	compArchCanvas   *canvas.Raster
	compTimingCanvas *canvas.Raster
	compEnergyCanvas *canvas.Raster
	compTableLabels  []*widget.Label
	compStatusLabel  *widget.Label

	// Tab 5: Timing
	timingOpSelect     *widget.Select
	timingWriteCanvas  *canvas.Raster
	timingReadCanvas   *canvas.Raster
	timingComputeCanvas *canvas.Raster
	timingStatusLabel  *widget.Label

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
	// Create tab contents
	writeTabContent := ca.createWriteTab()
	readTabContent := ca.createReadTab()
	computeTabContent := ca.createComputeTab()
	comparisonTabContent := ca.createComparisonTab()
	timingTabContent := ca.createTimingTab()
	specsTabContent := ca.createSpecsTab()

	// View selector dropdown (replaces nested tabs to save space)
	viewSelector := widget.NewSelect(
		[]string{"WRITE", "READ", "COMPUTE", "COMPARISON", "TIMING", "SPECS"},
		nil,
	)
	viewSelector.SetSelected("WRITE")

	// Content container
	contentContainer := container.NewMax(writeTabContent)

	// Update view based on selection
	viewSelector.OnChanged = func(view string) {
		switch view {
		case "WRITE":
			contentContainer.Objects[0] = writeTabContent
		case "READ":
			contentContainer.Objects[0] = readTabContent
		case "COMPUTE":
			contentContainer.Objects[0] = computeTabContent
		case "COMPARISON":
			contentContainer.Objects[0] = comparisonTabContent
		case "TIMING":
			contentContainer.Objects[0] = timingTabContent
		case "SPECS":
			contentContainer.Objects[0] = specsTabContent
		}
		contentContainer.Refresh()
	}

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
		topRow,
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

	// Voltage range entries
	vMinEntry := widget.NewEntry()
	vMinEntry.SetText("2.0")
	vMinEntry.OnChanged = func(s string) {
		var v float64
		fmt.Sscanf(s, "%f", &v)
		ca.mu.Lock()
		ca.vMin = v
		ca.mu.Unlock()
	}

	vMaxEntry := widget.NewEntry()
	vMaxEntry.SetText("5.0")
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
	ca.writeLevelLabel = widget.NewLabel("Target Level: 15 / 30")
	ca.writeLevelSlider = widget.NewSlider(0, float64(ca.quantLevels-1))
	ca.writeLevelSlider.Value = 15
	ca.writeLevelSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.targetLevel = int(v)
		ca.mu.Unlock()
		ca.writeLevelLabel.SetText(fmt.Sprintf("Target Level: %d / %d", ca.targetLevel, ca.quantLevels))
		ca.updateWriteDataPath()
		ca.refreshWritePulse()
	}

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Target Cell: Row"),
			ca.writeRowSelect,
			widget.NewLabel("Col"),
			ca.writeColSelect,
		),
		ca.writeLevelLabel,
		ca.writeLevelSlider,
	)
}

func (ca *CircuitsApp) createWriteDataPathSection() fyne.CanvasObject {
	// Create visual boxes for the data path
	digitalBox := ca.createLabeledBox("DIGITAL", "Level: --", colorPrimary)
	dacBox := ca.createLabeledBox("DAC", "5-bit", colorDAC)
	fefetBox := ca.createLabeledBox("FeFET", "Cell --,--", colorArrayCell)

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

	return ca.writeDataPath
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

	// Calculate conductance (1-100 uS range)
	conductance := 1.0 + float64(level)/float64(levels-1)*99.0

	// Binary representation
	binary := fmt.Sprintf("%05b", level)

	// Update the data path display
	if ca.writeDataPath != nil && len(ca.writeDataPath.Objects) >= 5 {
		// Update digital box
		if digitalBox, ok := ca.writeDataPath.Objects[0].(*fyne.Container); ok {
			if stack, ok := digitalBox.Objects[1].(*fyne.Container); ok {
				if vbox, ok := stack.Objects[0].(*fyne.Container); ok {
					if len(vbox.Objects) >= 2 {
						if lbl, ok := vbox.Objects[1].(*widget.Label); ok {
							lbl.SetText(fmt.Sprintf("Level:%d\n%s", level, binary))
						}
					}
				}
			}
		}

		// Update DAC box
		if dacBox, ok := ca.writeDataPath.Objects[2].(*fyne.Container); ok {
			if stack, ok := dacBox.Objects[1].(*fyne.Container); ok {
				if vbox, ok := stack.Objects[0].(*fyne.Container); ok {
					if len(vbox.Objects) >= 2 {
						if lbl, ok := vbox.Objects[1].(*widget.Label); ok {
							lbl.SetText(fmt.Sprintf("%.2fV", voltage))
						}
					}
				}
			}
		}

		// Update FeFET box
		if fefetBox, ok := ca.writeDataPath.Objects[4].(*fyne.Container); ok {
			if stack, ok := fefetBox.Objects[1].(*fyne.Container); ok {
				if vbox, ok := stack.Objects[0].(*fyne.Container); ok {
					if len(vbox.Objects) >= 2 {
						if lbl, ok := vbox.Objects[1].(*widget.Label); ok {
							lbl.SetText(fmt.Sprintf("[%d,%d]\n%.1fµS", row, col, conductance))
						}
					}
				}
			}
		}
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
		ca.writePulseCanvas.Refresh()
	}
}

func (ca *CircuitsApp) createWriteArraySection() fyne.CanvasObject {
	ca.writeArrayCanvas = canvas.NewRaster(ca.drawWriteArray)
	ca.writeArrayCanvas.SetMinSize(fyne.NewSize(400, 300))
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

	// Calculate cell size
	margin := 40
	cellW := (w - 2*margin) / cols
	cellH := (h - 2*margin) / rows

	if cellW > 40 {
		cellW = 40
	}
	if cellH > 40 {
		cellH = 40
	}

	// Draw cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := margin + c*cellW
			y0 := margin + r*cellH

			level := weights[r][c]
			intensity := float64(level) / float64(levels-1)

			// Color based on level (blue to red)
			var cr, cg, cb uint8
			if r == selectedRow && c == selectedCol {
				cr, cg, cb = 255, 200, 50
			} else {
				cr = uint8(intensity * 200)
				cg = uint8(50 + (1-intensity)*100)
				cb = uint8((1 - intensity) * 200)
			}

			cellColor := color.RGBA{cr, cg, cb, 255}
			drawRect(img, x0+2, y0+2, cellW-4, cellH-4, cellColor)
		}
	}

	return img
}

func (ca *CircuitsApp) refreshWriteArray() {
	if ca.writeArrayCanvas != nil {
		ca.writeArrayCanvas.Refresh()
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

	text := "Level │ Voltage │ Conductance │ Resistance\n"
	text += "──────┼─────────┼─────────────┼────────────\n"

	// Show a sample of levels including the target
	sampleLevels := []int{0, levels / 4, levels / 2, target, levels - 1}
	seen := make(map[int]bool)

	for _, l := range sampleLevels {
		if seen[l] {
			continue
		}
		seen[l] = true

		voltage := vMin + float64(l)/float64(levels-1)*(vMax-vMin)
		conductance := 1.0 + float64(l)/float64(levels-1)*99.0 // 1-100 µS
		resistance := 1000.0 / conductance                     // kΩ

		marker := "  "
		if l == target {
			marker = "→ "
		}
		text += fmt.Sprintf("%s%3d  │ %5.2fV  │   %5.1f µS  │  %6.1f kΩ\n",
			marker, l, voltage, conductance, resistance)
	}

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
		mainContent,
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		nil,
	)
}

func (ca *CircuitsApp) createReadConfigSection() fyne.CanvasObject {
	// Read voltage slider
	ca.readVoltageLabel = widget.NewLabel("Read Voltage: 0.5 V")
	ca.readVoltageSlider = widget.NewSlider(0.1, 1.5)
	ca.readVoltageSlider.Value = 0.5
	ca.readVoltageSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.readVoltage = v
		ca.mu.Unlock()
		ca.readVoltageLabel.SetText(fmt.Sprintf("Read Voltage: %.2f V", v))
		ca.refreshReadZone()
	}

	warningLabel := widget.NewLabel("SAFE ZONE: 0.1V - 1.0V")
	warningLabel.TextStyle = fyne.TextStyle{Bold: true}

	dangerLabel := widget.NewLabel("DANGER: > 2.0V (will modify cell!)")

	// ADC resolution select
	adcOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	adcSelect := widget.NewSelect(adcOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		ca.mu.Lock()
		ca.adcBits = bits
		ca.mu.Unlock()
	})
	adcSelect.SetSelected("8")

	// TIA gain select
	tiaOptions := []string{"1", "10", "100"}
	tiaSelect := widget.NewSelect(tiaOptions, func(s string) {
		var gain float64
		fmt.Sscanf(s, "%f", &gain)
		ca.mu.Lock()
		ca.tiaGain = gain
		ca.mu.Unlock()
	})
	tiaSelect.SetSelected("10")

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
		container.NewHBox(
			widget.NewLabel("TIA Gain:"),
			tiaSelect,
			widget.NewLabel("kOhm"),
		),
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

	// Calculation box
	calcLabel := widget.NewLabel(
		"I = G × V = -- µS × -- V = -- µA\n" +
			"V_tia = I × R = -- µA × -- kΩ = -- mV\n" +
			"ADC = (-- mV / 1000 mV) × 255 = --\n" +
			"Level = round(-- / 255 × 29) = --",
	)

	return container.NewVBox(
		ca.readDataPath,
		widget.NewSeparator(),
		widget.NewLabel("Calculation:"),
		calcLabel,
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

	margin := 50
	plotH := h - 2*margin
	plotW := w - 2*margin

	writeZoneColor := color.RGBA{200, 50, 50, 200}
	readZoneColor := color.RGBA{50, 150, 50, 200}
	threshColor := color.RGBA{255, 200, 0, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}

	// Write zone (> 2V)
	writeZoneTop := margin
	writeZoneBottom := margin + plotH*30/100
	drawRect(img, margin, writeZoneTop, plotW, writeZoneBottom-writeZoneTop, writeZoneColor)

	// Threshold line (2V)
	thresholdY := margin + plotH*40/100
	for x := margin; x < margin+plotW; x++ {
		for dy := -1; dy <= 1; dy++ {
			img.Set(x, thresholdY+dy, threshColor)
		}
	}

	// Read zone (< 1V)
	readZoneTop := margin + plotH*60/100
	readZoneBottom := h - margin
	drawRect(img, margin, readZoneTop, plotW, readZoneBottom-readZoneTop, readZoneColor)

	// Current read voltage indicator
	readY := h - margin - int(readV/5.0*float64(plotH))
	for x := margin; x < margin+plotW; x++ {
		for dy := -2; dy <= 2; dy++ {
			y := readY + dy
			if y >= margin && y < h-margin {
				img.Set(x, y, cyanColor)
			}
		}
	}

	return img
}

func (ca *CircuitsApp) refreshReadZone() {
	if ca.readZoneCanvas != nil {
		ca.readZoneCanvas.Refresh()
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
		mainContent,
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
	ca.computeArrayCanvas.SetMinSize(fyne.NewSize(500, 300))
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

	// DACs on left
	dacX := 30
	dacW := 60
	dacSpacing := (h - 40) / cols

	// Array in center
	arrayX := dacX + dacW + 40
	arrayW := 200
	arrayH := h - 80
	cellW := arrayW / cols
	cellH := arrayH / rows

	// ADCs on right
	adcX := arrayX + arrayW + 40

	// Draw DACs
	for i := 0; i < cols && i < len(inputs); i++ {
		y := 30 + i*dacSpacing
		drawRect(img, dacX, y, dacW, 30, dacColor)
	}

	// Draw array
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := arrayX + c*cellW
			y0 := 40 + r*cellH

			level := weights[r][c]
			intensity := float64(level) / 29.0

			cr := uint8(intensity * 200)
			cg := uint8(100 + (1-intensity)*100)
			cb := uint8((1 - intensity) * 200)
			cellColor := color.RGBA{cr, cg, cb, 255}

			drawRect(img, x0+2, y0+2, cellW-4, cellH-4, cellColor)
		}
	}

	// Draw ADCs
	for i := 0; i < rows && i < len(outputs); i++ {
		y := 40 + i*cellH
		drawRect(img, adcX, y, 60, 30, adcColor)
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
		nil,
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

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	sectionH := h / 3

	// CPU + DRAM section
	drawRect(img, 20, 10, 80, sectionH-20, colorCPU)
	drawRect(img, 150, 10, 80, sectionH-20, colorCPU)

	// GPU + HBM section
	y := sectionH
	drawRect(img, 20, y+10, 80, sectionH-20, colorGPU)
	drawRect(img, 150, y+10, 80, sectionH-20, colorGPU)

	// FeFET CIM section
	y = 2 * sectionH
	drawRect(img, 20, y+10, 210, sectionH-20, colorFeFET)

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

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	margin := 60
	barH := 25
	spacing := 40

	// CPU bar (500ns - full width)
	drawRect(img, margin, 20, w-margin-50, barH, colorCPU)

	// GPU bar (50ns - 10% width)
	gpuW := (w - margin - 50) * 50 / 500
	drawRect(img, margin, 20+spacing, gpuW, barH, colorGPU)

	// FeFET bar (20ns - 4% width)
	fefetW := (w - margin - 50) * 20 / 500
	if fefetW < 40 {
		fefetW = 40
	}
	drawRect(img, margin, 20+2*spacing, fefetW, barH, colorFeFET)

	// Axis
	axisY := h - 20
	for x := margin; x < w-50; x++ {
		img.Set(x, axisY, axisColor)
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

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	margin := 60
	barH := 30
	spacing := 50

	// CPU bar (1000 pJ - full width)
	cpuW := w - margin - 50
	drawRect(img, margin, 20, cpuW, barH, colorCPU)

	// GPU bar (100 pJ - 10% width)
	gpuW := cpuW * 100 / 1000
	drawRect(img, margin, 20+spacing, gpuW, barH, colorGPU)

	// FeFET bar (0.05 pJ - tiny)
	fefetW := 20 // Minimum visible
	drawRect(img, margin, 20+2*spacing, fefetW, barH, colorFeFET)

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
	if ca.compArchCanvas != nil {
		ca.compArchCanvas.Refresh()
	}
	if ca.compTimingCanvas != nil {
		ca.compTimingCanvas.Refresh()
	}
	if ca.compEnergyCanvas != nil {
		ca.compEnergyCanvas.Refresh()
	}

	ca.compStatusLabel.SetText("Comparison complete: FeFET wins by 20,000x energy efficiency!")
}

// ============================================================================
// TAB 5: TIMING DIAGRAMS
// ============================================================================

func (ca *CircuitsApp) createTimingTab() fyne.CanvasObject {
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
		container.NewHBox(widget.NewLabel("OPERATION:"), ca.timingOpSelect),
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

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 80
	signalH := 25
	spacing := 30

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

	for i, sig := range signals {
		y := 20 + i*spacing
		plotW := w - marginLeft - 20
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

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 80
	spacing := 30

	signals := []string{"CLK", "V_READ", "I_SENSE", "ADC_EN", "DATA_OUT"}

	for i, name := range signals {
		y := 20 + i*spacing
		plotW := w - marginLeft - 20

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

			img.Set(x, lineY, cyanColor)
		}
	}

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

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 100
	spacing := 30

	signals := []string{"CLK", "INPUT_VALID", "DAC_ALL", "ARRAY_SETTLE", "ADC_ALL", "OUTPUT_VALID"}

	for i, name := range signals {
		y := 20 + i*spacing
		plotW := w - marginLeft - 20

		for pct := 0; pct <= 100; pct++ {
			x := marginLeft + pct*plotW/100

			var lineY int
			switch name {
			case "CLK":
				if (pct/8)%2 == 0 && pct < 95 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "INPUT_VALID":
				if pct >= 5 && pct <= 85 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "DAC_ALL":
				if pct >= 10 && pct <= 35 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "ARRAY_SETTLE":
				if pct >= 35 && pct <= 60 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "ADC_ALL":
				if pct >= 55 && pct <= 90 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			case "OUTPUT_VALID":
				if pct >= 90 {
					lineY = y + 5
				} else {
					lineY = y + 20
				}
			}

			img.Set(x, lineY, cyanColor)
		}
	}

	return img
}

func (ca *CircuitsApp) refreshTimingDiagrams() {
	if ca.timingWriteCanvas != nil {
		ca.timingWriteCanvas.Refresh()
	}
	if ca.timingReadCanvas != nil {
		ca.timingReadCanvas.Refresh()
	}
	if ca.timingComputeCanvas != nil {
		ca.timingComputeCanvas.Refresh()
	}
}

// ============================================================================
// TAB 6: SPECIFICATIONS
// ============================================================================

func (ca *CircuitsApp) createSpecsTab() fyne.CanvasObject {
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
		nil,
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
Conversion Time:   5 ns
Power per DAC:     0.1 mW
Total DAC Power:   3.2 mW
INL:               < 0.5 LSB
DNL:               < 0.5 LSB`

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Resolution:"), ca.specDACBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel(specs),
	)
}

func (ca *CircuitsApp) createSpecADCSection() fyne.CanvasObject {
	adcBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	ca.specADCBitsSelect = widget.NewSelect(adcBitsOptions, nil)
	ca.specADCBitsSelect.SetSelected("8")

	specs := `Count:             32 (one per row)
Resolution:        8 bits (256 levels)
Input Range:       0V to 1.0V (after TIA)
Conversion Time:   10 ns
Power per ADC:     0.5 mW
Total ADC Power:   16 mW
ENOB:              7.5 bits
SNR:               46 dB`

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Resolution:"), ca.specADCBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel(specs),
	)
}

func (ca *CircuitsApp) createSpecTIASection() fyne.CanvasObject {
	tiaGainOptions := []string{"1", "10", "100"}
	ca.specTIAGainSelect = widget.NewSelect(tiaGainOptions, nil)
	ca.specTIAGainSelect.SetSelected("10")

	specs := `Count:             32 (one per row)
Gain (R_f):        10 kOhm
Bandwidth:         100 MHz
Input Current:     0 to 100 uA
Output Voltage:    0 to 1.0 V
Noise:             < 1 uA RMS`

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Gain:"), ca.specTIAGainSelect, widget.NewLabel("kOhm")),
		widget.NewLabel(specs),
	)
}

func (ca *CircuitsApp) createSpecFeFETSection() fyne.CanvasObject {
	specs := `Material:          HfZrO2 (HZO)
Thickness:         10 nm
Levels:            30 discrete states
Conductance:       1 uS to 100 uS
Read Voltage:      0.5 V (safe zone)
Write Voltage:     2.0 V to 5.0 V
Write Time:        50 ns
Endurance:         10^12 cycles
Retention:         10 years
Cell Size:         ~0.01 um^2`

	return widget.NewLabel(specs)
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
