// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified OPERATIONS view that consolidates WRITE, READ, and COMPUTE modes.
package gui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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

	// CRITICAL: Hide ALL panels first to prevent leftover UI
	if ca.writeConfigPanel != nil {
		ca.writeConfigPanel.Hide()
	}
	if ca.readConfigPanel != nil {
		ca.readConfigPanel.Hide()
	}
	if ca.computeConfigPanel != nil {
		ca.computeConfigPanel.Hide()
	}

	// THEN show only the selected panel
	switch mode {
	case ModeWrite:
		if ca.writeConfigPanel != nil {
			ca.writeConfigPanel.Show()
		}
	case ModeRead:
		if ca.readConfigPanel != nil {
			ca.readConfigPanel.Show()
		}
	case ModeCompute:
		if ca.computeConfigPanel != nil {
			ca.computeConfigPanel.Show()
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
		helpText = "WRITE: Program cells using DAC voltage pulses (1.2-1.5V). Select cell and target level, then PROGRAM."
	case ModeRead:
		helpText = "READ: Sense cell conductance with low voltage (≤0.5V). TIA converts current to voltage for ADC."
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
// ACTION BUTTONS
// ============================================================================

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

	// CRITICAL: Hide ALL button sets first to prevent leftover UI
	if ca.opsWriteButtons != nil {
		ca.opsWriteButtons.Hide()
	}
	if ca.opsReadButtons != nil {
		ca.opsReadButtons.Hide()
	}
	if ca.opsComputeButtons != nil {
		ca.opsComputeButtons.Hide()
	}

	// THEN show only the selected button set
	switch mode {
	case ModeWrite:
		if ca.opsWriteButtons != nil {
			ca.opsWriteButtons.Show()
		}
	case ModeRead:
		if ca.opsReadButtons != nil {
			ca.opsReadButtons.Show()
		}
	case ModeCompute:
		if ca.opsComputeButtons != nil {
			ca.opsComputeButtons.Show()
		}
	}
}

