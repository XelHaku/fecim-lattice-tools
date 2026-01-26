// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified OPERATIONS view that consolidates WRITE, READ, and COMPUTE modes.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
)

// OperationMode represents the current operation mode in the unified view
type OperationMode int

const (
	ModeWrite OperationMode = iota
	ModeRead
	ModeCompute
)

// ============================================================================
// TAPPABLE ARRAY CANVAS WIDGET
// ============================================================================

// TappableArrayCanvas is a canvas.Raster that responds to taps
type TappableArrayCanvas struct {
	widget.BaseWidget
	raster *canvas.Raster
	onTap  func(row, col int)
	ca     *CircuitsApp
}

func NewTappableArrayCanvas(ca *CircuitsApp, drawFunc func(w, h int) image.Image, onTap func(row, col int)) *TappableArrayCanvas {
	t := &TappableArrayCanvas{
		raster: canvas.NewRaster(drawFunc),
		onTap:  onTap,
		ca:     ca,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *TappableArrayCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.raster)
}

func (t *TappableArrayCanvas) SetMinSize(size fyne.Size) {
	t.raster.SetMinSize(size)
}

func (t *TappableArrayCanvas) Refresh() {
	t.raster.Refresh()
}

func (t *TappableArrayCanvas) Tapped(e *fyne.PointEvent) {
	// Get current widget/raster size
	size := t.raster.Size()

	t.ca.mu.RLock()
	rows := t.ca.arrayRows
	cols := t.ca.arrayCols
	t.ca.mu.RUnlock()

	// Recalculate cell geometry using same logic as drawSharedArray
	w := int(size.Width)
	h := int(size.Height)

	// Use same asymmetric margins as drawSharedArray
	topMargin := 70
	rightMargin := 70
	bottomMargin := 30
	leftMargin := 30

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows

	// Cell size calculations (COMPLETE - matching drawSharedArray exactly)
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
	if cellSize <= 0 {
		return
	}

	// Calculate grid size and offset
	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	// Convert click position to cell coordinates
	col := (int(e.Position.X) - offsetX) / cellSize
	row := (int(e.Position.Y) - offsetY) / cellSize

	// Bounds check
	if row >= 0 && row < rows && col >= 0 && col < cols {
		t.onTap(row, col)
	}
}

func (t *TappableArrayCanvas) TappedSecondary(*fyne.PointEvent) {}

// Cursor returns a pointer cursor to indicate the array is clickable
func (t *TappableArrayCanvas) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// ============================================================================
// UNIFIED OPERATIONS VIEW
// ============================================================================

// createOperationsView creates the unified OPERATIONS view with mode selector
func (ca *CircuitsApp) createOperationsView() fyne.CanvasObject {
	// Initialize operations-specific fields
	ca.currentMode = ModeWrite
	ca.operationsStatusLabel = widget.NewLabel("Ready")

	// 1. Create mode selector (segmented buttons using radio group)
	modeSelector := ca.createModeSelector()

	// 2. Create shared array section (left panel, always visible)
	arraySection := ca.createSharedArraySection()

	// 3. Create mode-specific panels (stacked, visibility toggled)
	ca.createWriteModePanel()
	ca.createReadModePanel()
	ca.createComputeModePanel()

	// Stack all mode panels (visibility toggled based on selection)
	modeStack := container.NewStack(
		ca.writeConfigPanel,
		ca.readConfigPanel,
		ca.computeConfigPanel,
	)

	// Initialize visibility: show write, hide others
	ca.writeConfigPanel.Show()
	ca.readConfigPanel.Hide()
	ca.computeConfigPanel.Hide()

	// 4. Create action buttons (changes per mode)
	actionButtons := ca.createOperationsButtons()

	// Layout: left panel (array), right panel (mode-specific content)
	rightPanel := container.NewVScroll(modeStack)

	mainContent := container.NewHSplit(
		arraySection,
		rightPanel,
	)
	mainContent.SetOffset(0.55) // Array gets 55% width (more space for integrated DAC/ADC)

	return container.NewBorder(
		modeSelector,
		actionButtons,
		nil, nil,
		mainContent,
	)
}

// createModeSelector creates the WRITE/READ/COMPUTE mode toggle
func (ca *CircuitsApp) createModeSelector() fyne.CanvasObject {
	modeRadio := widget.NewRadioGroup([]string{"WRITE", "READ", "COMPUTE"}, func(mode string) {
		ca.onModeChanged(mode)
	})
	modeRadio.Horizontal = true
	modeRadio.SetSelected("WRITE")

	modeHelp := widget.NewLabel("")
	modeHelp.TextStyle = fyne.TextStyle{Italic: true}
	ca.operationsModeHelp = modeHelp
	ca.updateModeHelp()

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Mode:"),
			modeRadio,
			layout.NewSpacer(),
			ca.operationsStatusLabel,
		),
		modeHelp,
		widget.NewSeparator(),
	)
}

// createSharedArraySection creates the array panel that's always visible
func (ca *CircuitsApp) createSharedArraySection() fyne.CanvasObject {
	// Create tappable array canvas
	tappableArray := NewTappableArrayCanvas(ca, ca.drawSharedArray, ca.onArrayCellTapped)
	// Larger size to accommodate integrated DAC (top) and ADC (right) boxes
	tappableArray.SetMinSize(fyne.NewSize(480, 420))
	ca.sharedArrayCanvas = tappableArray.raster // Keep reference for refresh

	// Color legend
	legendLabel := widget.NewLabel("Level: Low (blue) -> High (red) | Yellow = Selected | Click to select")
	legendLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Cell info display
	ca.sharedCellInfoLabel = widget.NewLabel("Click a cell to select")

	// Array size info
	ca.sharedArrayInfoLabel = widget.NewLabel(fmt.Sprintf("Array: %dx%d | %d levels", ca.arrayRows, ca.arrayCols, ca.quantLevels))

	titleLabel := widget.NewLabelWithStyle("CROSSBAR ARRAY", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create placeholder container for compute input row (will be populated later)
	ca.computeInputRowContainer = container.NewVBox()
	ca.computeInputRowContainer.Hide() // Initially hidden

	return container.NewVBox(
		titleLabel,
		ca.computeInputRowContainer, // Input row appears here in COMPUTE mode
		tappableArray,
		legendLabel,
		ca.sharedCellInfoLabel,
		ca.sharedArrayInfoLabel,
	)
}

// drawSharedArray draws the shared array visualization with click interaction
func (ca *CircuitsApp) drawSharedArray(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	selectedRow := ca.selectedRow
	selectedCol := ca.selectedCol
	levels := ca.quantLevels
	mode := ca.currentMode
	animStep := ca.animationStep
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

	// Calculate asymmetric margins for integrated DAC (top) and ADC (right)
	topMargin := 70   // Space for DAC boxes + labels above grid
	rightMargin := 70 // Space for ADC boxes + labels right of grid
	bottomMargin := 30
	leftMargin := 30

	// Calculate available area for grid
	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	// Calculate cell size (use square cells)
	cellW := availableW / cols
	cellH := availableH / rows
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

	// Calculate grid dimensions
	gridW := cols * cellSize
	gridH := rows * cellSize

	// Calculate offset to center grid within available area
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	// Store cell geometry for click detection
	ca.mu.Lock()
	ca.sharedArrayCellSize = cellSize
	ca.sharedArrayOffsetX = offsetX
	ca.sharedArrayOffsetY = offsetY
	ca.mu.Unlock()

	// Draw column lines (for compute mode - input voltages)
	if mode == ModeCompute {
		lineColor := color.RGBA{100, 100, 150, 150}
		for c := 0; c < cols; c++ {
			x := offsetX + c*cellSize + cellSize/2
			for y := offsetY - 15; y < offsetY+gridH+15; y++ {
				if y >= 0 && y < h {
					img.Set(x, y, lineColor)
				}
			}
		}
	}

	// Draw row lines (for compute mode - output currents)
	if mode == ModeCompute {
		lineColor := color.RGBA{150, 100, 100, 150}
		for r := 0; r < rows; r++ {
			y := offsetY + r*cellSize + cellSize/2
			for x := offsetX - 15; x < offsetX+gridW+15; x++ {
				if x >= 0 && x < w {
					img.Set(x, y, lineColor)
				}
			}
		}
	}

	// Draw cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := offsetX + c*cellSize
			y0 := offsetY + r*cellSize

			level := weights[r][c]
			intensity := float64(level) / float64(levels-1)

			// Check if this is the selected cell
			isSelected := r == selectedRow && c == selectedCol

			// Color based on level (blue to red gradient)
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

			// Draw border around selected cell
			if isSelected {
				borderColor := color.RGBA{255, 255, 255, 255}
				borderWidth := 3
				drawRect(img, x0, y0, cellSize, borderWidth, borderColor)
				drawRect(img, x0, y0+cellSize-borderWidth, cellSize, borderWidth, borderColor)
				drawRect(img, x0, y0, borderWidth, cellSize, borderColor)
				drawRect(img, x0+cellSize-borderWidth, y0, borderWidth, cellSize, borderColor)
			}

			// Highlight cells during array animation (Step 2)
			if mode == ModeCompute && animStep == 2 {
				// Add a bright cyan overlay/border to show computation in progress
				overlayColor := color.RGBA{0, 255, 255, 100} // Semi-transparent cyan
				// Draw brighter border around each cell
				drawRect(img, x0, y0, cellSize, 2, overlayColor)
				drawRect(img, x0, y0+cellSize-2, cellSize, 2, overlayColor)
				drawRect(img, x0, y0, 2, cellSize, overlayColor)
				drawRect(img, x0+cellSize-2, y0, 2, cellSize, overlayColor)
			}
		}
	}

	// Draw mode-specific overlays
	switch mode {
	case ModeWrite:
		// Show target level indicator on selected cell
		if selectedRow < rows && selectedCol < cols {
			x0 := offsetX + selectedCol*cellSize
			y0 := offsetY + selectedRow*cellSize
			// Draw small arrow pointing to cell
			arrowColor := color.RGBA{255, 255, 0, 255}
			for i := 0; i < 8; i++ {
				img.Set(x0-10+i, y0+cellSize/2, arrowColor)
			}
		}

	case ModeRead:
		// Show read probe indicator
		if selectedRow < rows && selectedCol < cols {
			x0 := offsetX + selectedCol*cellSize + cellSize/2
			y0 := offsetY + selectedRow*cellSize + cellSize/2
			// Draw probe circle
			probeColor := color.RGBA{0, 255, 255, 200}
			radius := cellSize / 3
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					if dx*dx+dy*dy <= radius*radius && dx*dx+dy*dy >= (radius-2)*(radius-2) {
						px, py := x0+dx, y0+dy
						if px >= 0 && px < w && py >= 0 && py < h {
							img.Set(px, py, probeColor)
						}
					}
				}
			}
		}

	case ModeCompute:
		// Show input/output indicators
		// Input arrows at top of columns
		arrowColor := color.RGBA{150, 100, 200, 255}
		for c := 0; c < min(8, cols); c++ {
			x := offsetX + c*cellSize + cellSize/2
			for i := 0; i < 10; i++ {
				y := offsetY - 20 + i
				if y >= 0 {
					img.Set(x, y, arrowColor)
					if i > 5 {
						img.Set(x-1, y, arrowColor)
						img.Set(x+1, y, arrowColor)
					}
				}
			}
		}

		// Output arrows at right of rows
		outputColor := color.RGBA{100, 200, 150, 255}
		for r := 0; r < min(8, rows); r++ {
			y := offsetY + r*cellSize + cellSize/2
			xStart := offsetX + gridW + 5
			for i := 0; i < 10; i++ {
				x := xStart + i
				if x < w {
					img.Set(x, y, outputColor)
					if i > 5 {
						img.Set(x, y-1, outputColor)
						img.Set(x, y+1, outputColor)
					}
				}
			}
		}

		// Draw DAC boxes at TOP of each column (integrated visualization)
		dacBoxHeight := 25
		dacBoxWidth := cellSize - 4
		dacY := offsetY - dacBoxHeight - 10
		dacColor := color.RGBA{100, 80, 180, 255} // Purple for DACs
		if animStep == 1 {
			dacColor = color.RGBA{255, 255, 100, 255} // Bright yellow when animating DAC step
		}

		// Define input label color for column labels (light blue for inputs)
		inputLabelColor := color.RGBA{100, 150, 255, 255}

		// OPTIMIZATION: Copy input vector data once before loop to avoid RLock per iteration
		dacColCount := min(8, cols)
		inputVectorCopy := make([]int, dacColCount)
		ca.mu.RLock()
		copy(inputVectorCopy, ca.inputVector[:dacColCount])
		ca.mu.RUnlock()

		for c := 0; c < dacColCount; c++ {
			dacX := offsetX + c*cellSize + 2

			// Draw DAC box
			drawRect(img, dacX, dacY, dacBoxWidth, dacBoxHeight, dacColor)

			// Draw border
			borderColor := color.RGBA{150, 130, 220, 255}
			drawRectBorder(img, dacX, dacY, dacBoxWidth, dacBoxHeight, borderColor)

			// Use pre-copied input value (no lock needed)
			inputVal := inputVectorCopy[c]
			voltage := float64(inputVal) / 255.0

			// Show voltage value
			voltageText := fmt.Sprintf("%.1fV", voltage)
			textX := dacX + dacBoxWidth/2 - len(voltageText)*3
			textY := dacY + dacBoxHeight/2 - 3
			drawSimpleText(img, voltageText, textX, textY, color.RGBA{255, 255, 255, 255})

			// Draw column label above DAC box
			labelX := offsetX + c*cellSize + cellSize/2 - 6
			labelY := dacY - 12
			drawSimpleText(img, fmt.Sprintf("x%d", c), labelX, labelY, inputLabelColor)
		}

		// Draw ADC boxes at RIGHT of each row (integrated visualization)
		adcBoxWidth := 45
		adcBoxHeight := cellSize - 4
		adcX := offsetX + gridW + 8
		adcColor := color.RGBA{80, 150, 100, 255} // Green for ADCs
		if animStep == 3 {
			adcColor = color.RGBA{100, 255, 150, 255} // Bright green when animating ADC step
		}

		// Define output label color for row labels (light orange for outputs)
		outputLabelColor := color.RGBA{255, 180, 100, 255}

		// OPTIMIZATION: Copy output vector data once before loop to avoid RLock per iteration
		adcRowCount := min(8, rows)
		outputVectorCopy := make([]float64, adcRowCount)
		ca.mu.RLock()
		copy(outputVectorCopy, ca.outputVector[:min(adcRowCount, len(ca.outputVector))])
		ca.mu.RUnlock()

		for r := 0; r < adcRowCount; r++ {
			adcY := offsetY + r*cellSize + 2

			// Draw ADC box
			drawRect(img, adcX, adcY, adcBoxWidth, adcBoxHeight, adcColor)

			// Draw border
			borderColor := color.RGBA{130, 200, 150, 255}
			drawRectBorder(img, adcX, adcY, adcBoxWidth, adcBoxHeight, borderColor)

			// Use pre-copied output value (no lock needed)
			outputVal := outputVectorCopy[r]

			// Show ADC level (after TIA+ADC conversion)
			tiaVoltage := ca.tia.Convert(outputVal * 1e-6)
			adcLevel := ca.adc.Convert(tiaVoltage)

			levelText := fmt.Sprintf("L%d", adcLevel)
			textX := adcX + adcBoxWidth/2 - len(levelText)*3
			textY := adcY + adcBoxHeight/2 - 3
			drawSimpleText(img, levelText, textX, textY, color.RGBA{255, 255, 255, 255})

			// Draw row label to right of ADC box
			labelX := adcX + adcBoxWidth + 5
			labelY := offsetY + r*cellSize + cellSize/2 - 3
			drawSimpleText(img, fmt.Sprintf("y%d", r), labelX, labelY, outputLabelColor)
		}
	}

	return img
}

// refreshSharedArray refreshes the shared array canvas
func (ca *CircuitsApp) refreshSharedArray() {
	if ca.sharedArrayCanvas != nil {
		fyne.Do(func() {
			ca.sharedArrayCanvas.Refresh()
		})
	}
}

// onModeChanged handles mode switching
func (ca *CircuitsApp) onModeChanged(mode string) {
	ca.mu.Lock()
	switch mode {
	case "WRITE":
		ca.currentMode = ModeWrite
	case "READ":
		ca.currentMode = ModeRead
	case "COMPUTE":
		ca.currentMode = ModeCompute
	}
	ca.mu.Unlock()

	// Update visible panels
	ca.updateOperationsPanels()
	ca.updateModeHelp()
	ca.refreshSharedArray()
	ca.updateSharedCellInfo()

	// Auto-compute when entering COMPUTE mode
	if mode == "COMPUTE" {
		ca.computeAndUpdateAll()
	}
}

// updateOperationsPanels shows/hides panels based on current mode
func (ca *CircuitsApp) updateOperationsPanels() {
	ca.mu.RLock()
	mode := ca.currentMode
	ca.mu.RUnlock()

	// Toggle panel visibility
	if ca.writeConfigPanel != nil {
		if mode == ModeWrite {
			ca.writeConfigPanel.Show()
		} else {
			ca.writeConfigPanel.Hide()
		}
	}

	if ca.readConfigPanel != nil {
		if mode == ModeRead {
			ca.readConfigPanel.Show()
		} else {
			ca.readConfigPanel.Hide()
		}
	}

	if ca.computeConfigPanel != nil {
		if mode == ModeCompute {
			ca.computeConfigPanel.Show()
		} else {
			ca.computeConfigPanel.Hide()
		}
	}

	// Toggle input row visibility in array section
	if ca.computeInputRowContainer != nil {
		if mode == ModeCompute {
			ca.computeInputRowContainer.Show()
		} else {
			ca.computeInputRowContainer.Hide()
		}
	}

	// Update action buttons
	ca.updateOperationsButtons()
}

// updateModeHelp updates the mode description text
func (ca *CircuitsApp) updateModeHelp() {
	if ca.operationsModeHelp == nil {
		return
	}

	ca.mu.RLock()
	mode := ca.currentMode
	ca.mu.RUnlock()

	var helpText string
	switch mode {
	case ModeWrite:
		helpText = "WRITE: Program cells using DAC voltage pulses (2-5V). Select cell and target level, then PROGRAM."
	case ModeRead:
		helpText = "READ: Sense cell conductance with low voltage (0.5V). TIA converts current to voltage for ADC."
	case ModeCompute:
		helpText = "COMPUTE: Matrix-vector multiply in ~20ns. Input voltages x conductances, summed by KCL."
	}

	fyne.Do(func() {
		ca.operationsModeHelp.SetText(helpText)
	})
}

// updateSharedCellInfo updates the cell info display
func (ca *CircuitsApp) updateSharedCellInfo() {
	if ca.sharedCellInfoLabel == nil {
		return
	}

	ca.mu.RLock()
	row := ca.selectedRow
	col := ca.selectedCol
	mode := ca.currentMode
	levels := ca.quantLevels
	var level int
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		level = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	conductance := 1.0 + float64(level)/float64(levels-1)*99.0

	var infoText string
	switch mode {
	case ModeWrite:
		infoText = fmt.Sprintf("Cell [%d,%d]: Level %d | Target: %d | G=%.1f uS", row, col, level, ca.targetLevel, conductance)
	case ModeRead:
		infoText = fmt.Sprintf("Cell [%d,%d]: Level %d | G=%.1f uS | Ready to read", row, col, level, conductance)
	case ModeCompute:
		infoText = fmt.Sprintf("Cell [%d,%d]: Level %d | G=%.1f uS | Weight in MVM", row, col, level, conductance)
	}

	fyne.Do(func() {
		ca.sharedCellInfoLabel.SetText(infoText)
	})
}

// onArrayCellTapped handles cell selection via click
func (ca *CircuitsApp) onArrayCellTapped(row, col int) {
	ca.mu.Lock()
	ca.selectedRow = row
	ca.selectedCol = col
	ca.mu.Unlock()

	ca.refreshSharedArray()
	ca.updateSharedCellInfo()
	ca.updateOpsWriteDataPath()
}

// ============================================================================
// WRITE MODE PANEL
// ============================================================================

// createWriteModePanel creates the write mode configuration panel
func (ca *CircuitsApp) createWriteModePanel() {
	// Target level slider
	ca.opsWriteLevelLabel = widget.NewLabel(fmt.Sprintf("Target Level: %d (0-%d)", ca.targetLevel, ca.quantLevels-1))
	ca.opsWriteLevelSlider = widget.NewSlider(0, float64(ca.quantLevels-1))
	ca.opsWriteLevelSlider.Value = float64(ca.targetLevel)
	ca.opsWriteLevelSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.targetLevel = int(v)
		ca.mu.Unlock()
		ca.opsWriteLevelLabel.SetText(fmt.Sprintf("Target Level: %d (0-%d)", ca.targetLevel, ca.quantLevels-1))
		ca.updateOpsWriteDataPath()
		ca.refreshOpsWritePulse()
		ca.updateSharedCellInfo()
	}

	// Voltage range entries
	vMinEntry := widget.NewEntry()
	vMinEntry.SetText(fmt.Sprintf("%.1f", ca.vMin))
	vMinEntry.OnChanged = func(s string) {
		var v float64
		fmt.Sscanf(s, "%f", &v)
		ca.mu.Lock()
		ca.vMin = v
		ca.mu.Unlock()
		ca.updateOpsWriteDataPath()
	}

	vMaxEntry := widget.NewEntry()
	vMaxEntry.SetText(fmt.Sprintf("%.1f", ca.vMax))
	vMaxEntry.OnChanged = func(s string) {
		var v float64
		fmt.Sscanf(s, "%f", &v)
		ca.mu.Lock()
		ca.vMax = v
		ca.mu.Unlock()
		ca.updateOpsWriteDataPath()
	}

	// Data path visualization
	ca.opsWriteDigitalLabel = widget.NewLabel(fmt.Sprintf("Level:%d\n%05b", ca.targetLevel, ca.targetLevel))
	ca.opsWriteDACLabel = widget.NewLabel("3.55V")
	ca.opsWriteFeFETLabel = widget.NewLabel(fmt.Sprintf("[%d,%d]\n--uS", ca.selectedRow, ca.selectedCol))

	digitalBox := ca.createLabeledBoxWithLabel("DIGITAL", ca.opsWriteDigitalLabel, sharedtheme.ColorPrimary)
	dacBox := ca.createLabeledBoxWithLabel("DAC", ca.opsWriteDACLabel, sharedtheme.ColorAccent)
	fefetBox := ca.createLabeledBoxWithLabel("FeFET", ca.opsWriteFeFETLabel, sharedtheme.ColorInfo)

	dataPath := container.NewHBox(
		digitalBox, widget.NewLabel("->"),
		dacBox, widget.NewLabel("->"),
		fefetBox,
	)

	// Pulse visualization
	ca.opsWritePulseCanvas = canvas.NewRaster(ca.drawOpsWritePulse)
	ca.opsWritePulseCanvas.SetMinSize(fyne.NewSize(350, 120))

	// Config section
	configSection := container.NewVBox(
		widget.NewLabelWithStyle("TARGET LEVEL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.opsWriteLevelLabel,
		ca.opsWriteLevelSlider,
		widget.NewLabel("Each level = stable polarization state (~4.9 bits/cell)"),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("VOLTAGE RANGE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(
			widget.NewLabel("Vmin:"), vMinEntry, widget.NewLabel("V"),
			widget.NewLabel("Vmax:"), vMaxEntry, widget.NewLabel("V"),
		),
		widget.NewLabel("Write voltage must exceed Ec (~1.5 MV/cm)"),
	)

	// Data path section
	dataPathSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("DATA PATH", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dataPath,
		widget.NewLabel("Digital level -> DAC voltage -> FeFET polarization"),
	)

	// Pulse section
	pulseSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("PROGRAMMING PULSE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.opsWritePulseCanvas,
	)

	ca.writeConfigPanel = container.NewVBox(
		configSection,
		dataPathSection,
		pulseSection,
	)

	// Initialize data path values
	ca.updateOpsWriteDataPath()
}

// updateOpsWriteDataPath updates the write mode data path display
func (ca *CircuitsApp) updateOpsWriteDataPath() {
	ca.mu.RLock()
	level := ca.targetLevel
	row := ca.selectedRow
	col := ca.selectedCol
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	ca.mu.RUnlock()

	voltage := vMin + float64(level)/float64(levels-1)*(vMax-vMin)
	conductance := 1.0 + float64(level)/float64(levels-1)*99.0

	if ca.opsWriteDigitalLabel != nil {
		fyne.Do(func() {
			ca.opsWriteDigitalLabel.SetText(fmt.Sprintf("Level:%d\n%05b", level, level))
		})
	}
	if ca.opsWriteDACLabel != nil {
		fyne.Do(func() {
			ca.opsWriteDACLabel.SetText(fmt.Sprintf("%.2fV", voltage))
		})
	}
	if ca.opsWriteFeFETLabel != nil {
		fyne.Do(func() {
			ca.opsWriteFeFETLabel.SetText(fmt.Sprintf("[%d,%d]\n%.1fuS", row, col, conductance))
		})
	}
}

// drawOpsWritePulse draws the programming pulse waveform
func (ca *CircuitsApp) drawOpsWritePulse(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := color.RGBA{0, 40, 80, 255}

	ca.mu.RLock()
	level := ca.targetLevel
	vMin := ca.vMin
	vMax := ca.vMax
	levels := ca.quantLevels
	ca.mu.RUnlock()

	voltage := vMin + float64(level)/float64(levels-1)*(vMax-vMin)

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw axes
	marginLeft := 40
	marginBottom := 20
	marginTop := 15
	marginRight := 20

	plotW := w - marginLeft - marginRight
	plotH := h - marginTop - marginBottom
	axisColor := color.RGBA{200, 200, 200, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	fillColor := color.RGBA{0, 100, 150, 200}
	threshColor := color.RGBA{255, 200, 0, 255}

	// Y-axis
	for y := marginTop; y < h-marginBottom; y++ {
		img.Set(marginLeft, y, axisColor)
	}

	// X-axis
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

	// Threshold line (dashed)
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

		// Thick line
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

	// Labels
	drawSimpleText(img, fmt.Sprintf("%.1fV", vMax), 5, marginTop+3, axisColor)
	drawSimpleText(img, fmt.Sprintf("%.1fV", vMin), 5, yThreshold+3, axisColor)
	drawSimpleText(img, "0V", 15, y0V-8, axisColor)
	drawSimpleText(img, "Time", w-50, h-10, axisColor)

	return img
}

// refreshOpsWritePulse refreshes the write pulse canvas
func (ca *CircuitsApp) refreshOpsWritePulse() {
	if ca.opsWritePulseCanvas != nil {
		fyne.Do(func() {
			ca.opsWritePulseCanvas.Refresh()
		})
	}
}

// ============================================================================
// READ MODE PANEL
// ============================================================================

// createReadModePanel creates the read mode configuration panel
func (ca *CircuitsApp) createReadModePanel() {
	// Read voltage slider
	ca.opsReadVoltageLabel = widget.NewLabel(fmt.Sprintf("Read Voltage: %.2f V", ca.readVoltage))
	ca.opsReadVoltageSlider = widget.NewSlider(0.1, 0.5)  // Max 0.5V for non-disturbing read
	ca.opsReadVoltageSlider.Value = ca.readVoltage
	ca.opsReadVoltageSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.readVoltage = v
		ca.mu.Unlock()
		ca.opsReadVoltageLabel.SetText(fmt.Sprintf("Read Voltage: %.2f V", v))
		ca.refreshOpsReadZone()
	}

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

	// ADC bits select
	adcOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	adcSelect := widget.NewSelect(adcOptions, func(s string) {
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		ca.mu.Lock()
		ca.adcBits = bits
		ca.mu.Unlock()
	})
	adcSelect.SetSelected("8")

	// Voltage zone visualization
	ca.opsReadZoneCanvas = canvas.NewRaster(ca.drawOpsReadZone)
	ca.opsReadZoneCanvas.SetMinSize(fyne.NewSize(250, 150))

	// Data path visualization
	fefetBox := ca.createLabeledBox("FeFET", "Cell", sharedtheme.ColorInfo)
	tiaBox := ca.createLabeledBox("TIA", "I->V", sharedtheme.ColorWarning)
	adcBox := ca.createLabeledBox("ADC", "8-bit", sharedtheme.ColorSuccess)
	digitalBox := ca.createLabeledBox("DIGITAL", "Level", sharedtheme.ColorPrimary)

	dataPath := container.NewHBox(
		fefetBox, widget.NewLabel("->"),
		tiaBox, widget.NewLabel("->"),
		adcBox, widget.NewLabel("->"),
		digitalBox,
	)

	// Results display
	ca.opsReadResultsLabel = widget.NewLabel(
		"Cell [--,--] Read Results\n" +
			"Programmed Level: --\n" +
			"Read Current:     -- uA\n" +
			"TIA Voltage:      -- mV\n" +
			"ADC Raw:          --\n" +
			"Decoded Level:    --\n" +
			"Match:            --",
	)
	ca.opsReadResultsLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Config section
	configSection := container.NewVBox(
		widget.NewLabelWithStyle("READ PARAMETERS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.opsReadVoltageLabel,
		ca.opsReadVoltageSlider,
		widget.NewLabel("SAFE: ≤0.5V (non-disturbing) | CAUTION: >0.5V may disturb polarization"),
		container.NewHBox(
			widget.NewLabel("TIA Gain:"), tiaSelect, widget.NewLabel("kOhm"),
		),
		container.NewHBox(
			widget.NewLabel("ADC Bits:"), adcSelect,
		),
	)

	// Zone visualization
	zoneSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("VOLTAGE ZONES", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.opsReadZoneCanvas,
	)

	// Data path section
	dataPathSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("DATA PATH", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dataPath,
		widget.NewLabel("FeFET current -> TIA voltage -> ADC -> Level"),
	)

	// Results section
	resultsSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("READ RESULTS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.opsReadResultsLabel,
	)

	ca.readConfigPanel = container.NewVBox(
		configSection,
		zoneSection,
		dataPathSection,
		resultsSection,
	)
}

// drawOpsReadZone draws the read voltage zone visualization
func (ca *CircuitsApp) drawOpsReadZone(w, h int) image.Image {
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

	marginLeft := 40
	marginRight := 15
	marginTop := 10
	marginBottom := 10
	plotH := h - marginTop - marginBottom
	plotW := w - marginLeft - marginRight

	writeZoneColor := color.RGBA{200, 50, 50, 180}
	readZoneColor := color.RGBA{50, 150, 50, 180}
	threshColor := color.RGBA{255, 200, 0, 255}
	cyanColor := color.RGBA{0, 255, 255, 255}
	labelColor := color.RGBA{255, 255, 255, 255}
	axisColor := color.RGBA{150, 150, 150, 255}

	maxVoltage := 5.0

	voltageToY := func(v float64) int {
		return marginTop + int((maxVoltage-v)/maxVoltage*float64(plotH))
	}

	// Write zone (> 2V)
	writeZoneTop := voltageToY(maxVoltage)
	writeZoneBottom := voltageToY(2.0)
	drawRect(img, marginLeft, writeZoneTop, plotW, writeZoneBottom-writeZoneTop, writeZoneColor)

	// Read zone (< 1V)
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

	// Zone labels
	drawSimpleText(img, "WRITE", marginLeft+5, writeZoneTop+12, labelColor)
	drawSimpleText(img, "READ", marginLeft+5, readZoneTop+12, labelColor)

	// Y-axis
	for y := marginTop; y <= h-marginBottom; y++ {
		img.Set(marginLeft-1, y, axisColor)
	}

	// Voltage markers
	for _, v := range []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0} {
		y := voltageToY(v)
		for dx := 0; dx < 4; dx++ {
			img.Set(marginLeft-4+dx, y, axisColor)
		}
		drawSimpleText(img, fmt.Sprintf("%.0fV", v), 5, y-3, axisColor)
	}

	// Current read voltage indicator
	readY := voltageToY(readV)
	for x := marginLeft; x < marginLeft+plotW; x++ {
		for dy := -2; dy <= 2; dy++ {
			y := readY + dy
			if y >= marginTop && y < h-marginBottom {
				img.Set(x, y, cyanColor)
			}
		}
	}

	// Arrow indicator
	for i := 0; i < 6; i++ {
		img.Set(marginLeft-6+i, readY, cyanColor)
	}

	return img
}

// refreshOpsReadZone refreshes the read zone canvas
func (ca *CircuitsApp) refreshOpsReadZone() {
	if ca.opsReadZoneCanvas != nil {
		fyne.Do(func() {
			ca.opsReadZoneCanvas.Refresh()
		})
	}
}

// ============================================================================
// COMPUTE MODE PANEL
// ============================================================================

// createComputeModePanel creates the compute mode configuration panel
func (ca *CircuitsApp) createComputeModePanel() {
	// Input vector entries
	ca.opsComputeInputs = make([]*widget.Entry, ca.arrayCols)
	ca.opsComputeVoltageLabels = make([]*widget.Label, ca.arrayCols)

	// Create horizontal input row with compact entries
	inputRow := container.NewHBox()
	maxDisplay := min(8, ca.arrayCols)
	for i := 0; i < maxDisplay; i++ {
		ca.opsComputeInputs[i] = widget.NewEntry()
		ca.opsComputeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))
		ca.opsComputeInputs[i].Resize(fyne.NewSize(45, 30)) // Compact width

		idx := i
		ca.opsComputeInputs[i].OnChanged = func(s string) {
			var v int
			fmt.Sscanf(s, "%d", &v)
			if v > 255 {
				v = 255
			}
			ca.mu.Lock()
			ca.inputVector[idx] = v
			ca.mu.Unlock()
			if ca.opsComputeVoltageLabels[idx] != nil {
				ca.opsComputeVoltageLabels[idx].SetText(fmt.Sprintf("%.2fV", float64(v)/255.0))
			}
			// Auto-compute on input change
			ca.computeAndUpdateAll()
		}

		// Compact column: label on top, entry below
		ca.opsComputeVoltageLabels[i] = widget.NewLabel(fmt.Sprintf("%.2fV", float64(ca.inputVector[i])/255.0))
		ca.opsComputeVoltageLabels[i].TextStyle = fyne.TextStyle{Monospace: true}

		col := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("x%d", i)),
			ca.opsComputeInputs[i],
		)
		inputRow.Add(col)
	}

	// Output display
	ca.opsComputeOutputLabels = make([]*widget.Label, 8)
	outputGrid := container.NewGridWithColumns(2)
	for i := 0; i < 8; i++ {
		ca.opsComputeOutputLabels[i] = widget.NewLabel(fmt.Sprintf("y%d: --", i))
		outputGrid.Add(ca.opsComputeOutputLabels[i])
	}

	// Math breakdown
	ca.opsComputeMathLabel = widget.NewLabel(
		"I0 = G00*V0 + G01*V1 + ... (KCL sum)\n" +
			"All rows computed simultaneously!\n" +
			"Total latency: ~20ns",
	)
	ca.opsComputeMathLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Create Random Bits button
	randomBitsBtn := widget.NewButton("RANDOM BITS", func() {
		ca.mu.Lock()
		for i := range ca.inputVector {
			ca.inputVector[i] = rand.Intn(256)
		}
		ca.mu.Unlock()
		ca.updateOpsComputeInputs()
		ca.computeAndUpdateAll()
	})

	// Compact input section header (will be shown above array)
	inputHeader := container.NewHBox(
		widget.NewLabelWithStyle("INPUT VECTOR (0-255)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		randomBitsBtn,
	)

	physicsNote := widget.NewLabel("0.3-0.5V COMPUTE-safe (well below Vc)")
	physicsNote.TextStyle = fyne.TextStyle{Italic: true}

	// Populate the input row container that will appear above the array
	ca.computeInputRowContainer.Objects = []fyne.CanvasObject{
		widget.NewSeparator(),
		inputHeader,
		inputRow,
		physicsNote,
		widget.NewSeparator(),
	}

	// Output section
	// Ideal crossbar disclaimer
	idealDisclaimer := widget.NewLabel(
		"IDEAL CROSSBAR: No IR drop or sneak paths (see Module 2)")
	idealDisclaimer.TextStyle = fyne.TextStyle{Italic: true}

	outputSection := container.NewVBox(
		widget.NewLabelWithStyle("OUTPUT VECTOR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("I_row -> TIA (10k) -> ADC (5-bit):"),
		outputGrid,
		idealDisclaimer,
	)

	// Math section
	mathSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("MATH (Row 0 Breakdown)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.opsComputeMathLabel,
	)

	// Performance info
	perfSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("PERFORMANCE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("DAC: 5ns | Array settle: 5ns | ADC: 10ns"),
		widget.NewLabel("TOTAL: ~20ns for full MVM!"),
		widget.NewLabel("GPU equivalent: ~1000 cycles"),
	)

	ca.computeConfigPanel = container.NewVBox(
		outputSection,
		mathSection,
		perfSection,
	)
}

// updateOpsComputeInputs updates the compute input display
func (ca *CircuitsApp) updateOpsComputeInputs() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	for i := 0; i < min(8, len(ca.opsComputeInputs)); i++ {
		if ca.opsComputeInputs[i] != nil {
			fyne.Do(func() {
				ca.opsComputeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))
			})
		}
		if ca.opsComputeVoltageLabels[i] != nil {
			voltage := float64(ca.inputVector[i]) / 255.0
			fyne.Do(func() {
				ca.opsComputeVoltageLabels[i].SetText(fmt.Sprintf("%.2fV", voltage))
			})
		}
	}
}

// computeAndUpdateAll performs MVM and updates all output displays
// Called by: input changes, RANDOM BITS, mode selector, COMPUTE button
// IMPORTANT: Does NOT call updateOpsComputeInputs() to prevent Entry->OnChanged recursion
func (ca *CircuitsApp) computeAndUpdateAll() {
	// 1. MVM computation
	ca.mu.Lock()
	rows := min(8, ca.arrayRows)
	cols := min(8, ca.arrayCols)

	for r := 0; r < rows && r < len(ca.arrayWeights); r++ {
		sum := 0.0
		for c := 0; c < cols && c < len(ca.arrayWeights[r]); c++ {
			conductance := 1.0 + float64(ca.arrayWeights[r][c])/29.0*99.0
			voltage := float64(ca.inputVector[c]) / 255.0
			sum += conductance * voltage
		}
		ca.outputVector[r] = sum
	}
	ca.mu.Unlock()

	// 2. Update output labels with TIA/ADC conversion
	ca.mu.RLock()
	for i := 0; i < 8 && i < len(ca.outputVector); i++ {
		if ca.opsComputeOutputLabels[i] != nil {
			rawCurrent := ca.outputVector[i]
			tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6)
			adcLevel := ca.adc.Convert(tiaVoltage)
			isSaturated := rawCurrent > 100.0

			idx := i
			current := rawCurrent
			tiaV := tiaVoltage
			level := adcLevel
			sat := isSaturated
			fyne.Do(func() {
				// Show full pipeline: Current -> TIA Voltage -> ADC Level
				satSuffix := ""
				if sat {
					satSuffix = " SAT"
				}
				ca.opsComputeOutputLabels[idx].SetText(
					fmt.Sprintf("y%d: %.1fuA -> %.2fV -> L%d%s", idx, current, tiaV, level, satSuffix))
			})
		}
	}
	ca.mu.RUnlock()

	// 3. Update math breakdown
	ca.updateOpsComputeMath()

	// 4. Update data path displays
	ca.updateOpsComputeInputDataPath()
	ca.updateOpsComputeOutputDataPath()
}

// updateOpsComputeInputDataPath updates the input data path display
func (ca *CircuitsApp) updateOpsComputeInputDataPath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.inputVector) == 0 {
		return
	}

	// Show summary of input vector (first value as example)
	digitalVal := ca.inputVector[0]
	voltage := float64(digitalVal) / 255.0

	if ca.opsComputeInputDigitalLabel != nil {
		fyne.Do(func() {
			ca.opsComputeInputDigitalLabel.SetText(fmt.Sprintf("x0: %d\n0b%08b", digitalVal, digitalVal))
		})
	}
	if ca.opsComputeInputDACLabel != nil {
		fyne.Do(func() {
			ca.opsComputeInputDACLabel.SetText(fmt.Sprintf("%.2fV", voltage))
		})
	}
}

// updateOpsComputeOutputDataPath updates the output data path display
func (ca *CircuitsApp) updateOpsComputeOutputDataPath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.outputVector) == 0 {
		return
	}

	// Show y0 as the example
	rawCurrent := ca.outputVector[0] // uA

	// TIA conversion (saturates at 100 uA -> 1.0V output)
	tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6) // uA to A

	// ADC conversion (5-bit: 0V->0, 1V->31)
	adcLevel := ca.adc.Convert(tiaVoltage)

	// Check for TIA saturation
	isSaturated := rawCurrent > 100.0

	satSuffix := ""
	if isSaturated {
		satSuffix = " (SAT)"
	}

	if ca.opsComputeOutputCurrentLabel != nil {
		fyne.Do(func() {
			ca.opsComputeOutputCurrentLabel.SetText(fmt.Sprintf("%.1f uA%s", rawCurrent, satSuffix))
		})
	}
	if ca.opsComputeOutputTIALabel != nil {
		fyne.Do(func() {
			ca.opsComputeOutputTIALabel.SetText(fmt.Sprintf("%.3f V%s", tiaVoltage, satSuffix))
		})
	}
	if ca.opsComputeOutputADCLabel != nil {
		fyne.Do(func() {
			ca.opsComputeOutputADCLabel.SetText(fmt.Sprintf("Level %d%s", adcLevel, satSuffix))
		})
	}
}

// ============================================================================
// ACTION BUTTONS
// ============================================================================

// createOperationsButtons creates the mode-specific action buttons
func (ca *CircuitsApp) createOperationsButtons() fyne.CanvasObject {
	// Write mode buttons
	ca.opsProgramBtn = widget.NewButton("PROGRAM CELL", ca.onOpsProgram)
	ca.opsProgramBtn.Importance = widget.HighImportance
	ca.opsProgramRandomBtn = widget.NewButton("RANDOM ARRAY", ca.onOpsProgramRandom)

	// Read mode buttons
	ca.opsReadBtn = widget.NewButton("READ CELL", ca.onOpsRead)
	ca.opsReadBtn.Importance = widget.HighImportance
	ca.opsVerifyBtn = widget.NewButton("VERIFY ARRAY", ca.onOpsVerify)

	// Compute mode buttons
	ca.opsComputeBtn = widget.NewButton("COMPUTE", ca.onOpsCompute)
	ca.opsComputeBtn.Importance = widget.HighImportance
	ca.opsAnimateBtn = widget.NewButton("ANIMATE", ca.onOpsAnimate)
	ca.opsResetBtn = widget.NewButton("RESET", ca.onOpsReset)

	// Create button containers for each mode
	ca.opsWriteButtons = container.NewHBox(ca.opsProgramBtn, ca.opsProgramRandomBtn)
	ca.opsReadButtons = container.NewHBox(ca.opsReadBtn, ca.opsVerifyBtn)
	ca.opsComputeButtons = container.NewHBox(ca.opsComputeBtn, ca.opsAnimateBtn, ca.opsResetBtn)

	// Stack all button sets
	buttonStack := container.NewStack(
		ca.opsWriteButtons,
		ca.opsReadButtons,
		ca.opsComputeButtons,
	)

	// Initialize visibility
	ca.opsWriteButtons.Show()
	ca.opsReadButtons.Hide()
	ca.opsComputeButtons.Hide()

	return container.NewHBox(
		buttonStack,
		layout.NewSpacer(),
		ca.operationsStatusLabel,
	)
}

// updateOperationsButtons shows/hides action buttons based on mode
func (ca *CircuitsApp) updateOperationsButtons() {
	ca.mu.RLock()
	mode := ca.currentMode
	ca.mu.RUnlock()

	if ca.opsWriteButtons != nil {
		if mode == ModeWrite {
			ca.opsWriteButtons.Show()
		} else {
			ca.opsWriteButtons.Hide()
		}
	}

	if ca.opsReadButtons != nil {
		if mode == ModeRead {
			ca.opsReadButtons.Show()
		} else {
			ca.opsReadButtons.Hide()
		}
	}

	if ca.opsComputeButtons != nil {
		if mode == ModeCompute {
			ca.opsComputeButtons.Show()
		} else {
			ca.opsComputeButtons.Hide()
		}
	}
}

// ============================================================================
// ACTION HANDLERS
// ============================================================================

// onOpsProgram programs the selected cell
func (ca *CircuitsApp) onOpsProgram() {
	ca.mu.Lock()
	row := ca.selectedRow
	col := ca.selectedCol
	level := ca.targetLevel

	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		ca.arrayWeights[row][col] = level
	}
	ca.mu.Unlock()

	ca.refreshSharedArray()
	ca.updateSharedCellInfo()
	ca.operationsStatusLabel.SetText(fmt.Sprintf("Programmed [%d,%d] = Level %d", row, col, level))
}

// onOpsProgramRandom programs the entire array with random values
func (ca *CircuitsApp) onOpsProgramRandom() {
	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = rand.Intn(ca.quantLevels)
		}
	}
	ca.mu.Unlock()

	ca.refreshSharedArray()
	ca.updateSharedCellInfo()
	ca.operationsStatusLabel.SetText("Programmed array with random values")
}

// onOpsRead reads the selected cell
func (ca *CircuitsApp) onOpsRead() {
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
	conductance := 1.0 + float64(storedLevel)/float64(levels-1)*99.0

	// Current: I = G * V
	current := conductance * readV

	// TIA: V = I * R
	tiaVoltage := current * tiaGain

	// ADC conversion
	adcRaw := int(tiaVoltage / 1000.0 * 255.0)
	if adcRaw > 255 {
		adcRaw = 255
	}

	// Decode back to level
	decodedLevel := int(math.Round(float64(adcRaw) / 255.0 * float64(levels-1)))

	match := "MATCH"
	if decodedLevel != storedLevel {
		match = fmt.Sprintf("MISMATCH (exp %d)", storedLevel)
	}

	// Update results display
	if ca.opsReadResultsLabel != nil {
		fyne.Do(func() {
			ca.opsReadResultsLabel.SetText(fmt.Sprintf(
				"Cell [%d,%d] Read Results\n"+
					"Programmed Level: %d\n"+
					"Read Current:     %.1f uA\n"+
					"TIA Voltage:      %.0f mV\n"+
					"ADC Raw:          %d\n"+
					"Decoded Level:    %d\n"+
					"Match:            %s",
				row, col, storedLevel, current, tiaVoltage, adcRaw, decodedLevel, match,
			))
		})
	}

	ca.operationsStatusLabel.SetText(fmt.Sprintf("Read [%d,%d]: Level %d", row, col, decodedLevel))
}

// onOpsVerify verifies all cells in the array
func (ca *CircuitsApp) onOpsVerify() {
	ca.operationsStatusLabel.SetText("Verifying array...")

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	levels := ca.quantLevels
	readV := ca.readVoltage
	tiaGain := ca.tiaGain
	ca.mu.RUnlock()

	go func() {
		errors := 0
		for r := 0; r < rows && r < len(weights); r++ {
			for c := 0; c < cols && c < len(weights[r]); c++ {
				storedLevel := weights[r][c]
				conductance := 1.0 + float64(storedLevel)/float64(levels-1)*99.0
				current := conductance * readV
				tiaVoltage := current * tiaGain
				adcRaw := int(tiaVoltage / 1000.0 * 255.0)
				if adcRaw > 255 {
					adcRaw = 255
				}
				decodedLevel := int(math.Round(float64(adcRaw) / 255.0 * float64(levels-1)))
				if decodedLevel != storedLevel {
					errors++
				}
			}
		}

		totalCells := rows * cols
		fyne.Do(func() {
			if errors == 0 {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("Verify OK: %d/%d cells", totalCells, totalCells))
			} else {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("Verify: %d errors in %d cells", errors, totalCells))
			}
		})
	}()
}

// onOpsCompute performs matrix-vector multiplication
func (ca *CircuitsApp) onOpsCompute() {
	ca.computeAndUpdateAll()
	ca.operationsStatusLabel.SetText("Compute complete in ~20ns")
}

// updateOpsComputeMath updates the math breakdown display
func (ca *CircuitsApp) updateOpsComputeMath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.arrayWeights) == 0 || len(ca.arrayWeights[0]) == 0 {
		return
	}

	cols := min(4, len(ca.arrayWeights[0]))
	mathText := "I0 = "
	var terms []string
	totalCurrent := 0.0

	for c := 0; c < cols; c++ {
		conductance := 1.0 + float64(ca.arrayWeights[0][c])/29.0*99.0
		voltage := float64(ca.inputVector[c]) / 255.0
		current := conductance * voltage
		totalCurrent += current
		terms = append(terms, fmt.Sprintf("%.0f*%.2f", conductance, voltage))
	}

	mathText += terms[0]
	for i := 1; i < len(terms); i++ {
		mathText += " + " + terms[i]
	}
	mathText += " + ...\n"
	mathText += fmt.Sprintf("   = %.1f uA\n", ca.outputVector[0])
	mathText += "ALL ROWS IN PARALLEL!"

	if ca.opsComputeMathLabel != nil {
		fyne.Do(func() {
			ca.opsComputeMathLabel.SetText(mathText)
		})
	}
}

// onOpsAnimate animates the compute process step by step with visual feedback
func (ca *CircuitsApp) onOpsAnimate() {
	ca.mu.Lock()
	ca.animationActive = true
	ca.mu.Unlock()

	ca.operationsStatusLabel.SetText("Animating...")

	go func() {
		// Step 1: DAC highlight
		ca.mu.Lock()
		ca.animationStep = 1
		ca.mu.Unlock()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 1: DAC conversion (5ns)")
		})
		time.Sleep(600 * time.Millisecond)

		// Step 2: Array highlight
		ca.mu.Lock()
		ca.animationStep = 2
		ca.mu.Unlock()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 2: Array MVM (5ns)")
		})
		time.Sleep(600 * time.Millisecond)

		// Step 3: ADC highlight
		ca.mu.Lock()
		ca.animationStep = 3
		ca.mu.Unlock()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 3: ADC conversion (10ns)")
		})
		time.Sleep(600 * time.Millisecond)

		// Complete
		ca.mu.Lock()
		ca.animationStep = 0
		ca.animationActive = false
		ca.mu.Unlock()
		ca.computeAndUpdateAll()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Compute complete in ~20ns")
		})
	}()
}

// onOpsReset resets the compute state
func (ca *CircuitsApp) onOpsReset() {
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = 0
	}
	for i := range ca.outputVector {
		ca.outputVector[i] = 0
	}
	ca.mu.Unlock()

	ca.updateOpsComputeInputs()
	for i := 0; i < 8 && i < len(ca.opsComputeOutputLabels); i++ {
		if ca.opsComputeOutputLabels[i] != nil {
			fyne.Do(func() {
				ca.opsComputeOutputLabels[i].SetText(fmt.Sprintf("y%d: --", i))
			})
		}
	}

	if ca.opsComputeMathLabel != nil {
		fyne.Do(func() {
			ca.opsComputeMathLabel.SetText(
				"I0 = G00*V0 + G01*V1 + ... (KCL sum)\n" +
					"All rows computed simultaneously!\n" +
					"Total latency: ~20ns",
			)
		})
	}

	ca.operationsStatusLabel.SetText("Reset complete")
}
