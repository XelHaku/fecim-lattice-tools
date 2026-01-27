// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device simulation view that replaces separate WRITE/READ/COMPUTE modes.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ============================================================================
// UNIFIED DEVICE SIMULATION VIEW
// ============================================================================

// createUnifiedView creates the unified device simulation view
// Replaces the old mode-based createOperationsView()
func (ca *CircuitsApp) createUnifiedView() fyne.CanvasObject {
	// Initialize device state
	ca.deviceState = NewDeviceState(ca.arrayRows, ca.arrayCols, ca.tia, ca.adc)
	ca.operationsStatusLabel = widget.NewLabel("Ready")

	// 1. Signal chain header
	signalChainHeader := ca.createSignalChainHeader()

	// 2. DAC input section (top)
	dacSection := ca.createDACInputSection()

	// 3. Main visualization area (center)
	mainSection := ca.createMainSimSection()

	// 4. Action buttons (bottom)
	actionSection := ca.createUnifiedActionSection()

	return container.NewBorder(
		container.NewVBox(signalChainHeader, dacSection), // top
		actionSection, // bottom
		nil, nil,
		mainSection, // center
	)
}

// createSignalChainHeader creates the signal chain indicator
func (ca *CircuitsApp) createSignalChainHeader() fyne.CanvasObject {
	// Architecture toggle
	archToggle := ca.createArchitectureToggle()

	chainLabel := widget.NewLabelWithStyle(
		"SIGNAL CHAIN: DAC -> Array -> TIA -> ADC",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Operation classification label (updates based on configuration)
	ca.operationsModeHelp = widget.NewLabel("Configuration: Click cells or adjust voltages")
	ca.operationsModeHelp.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewHBox(
			chainLabel,
			layout.NewSpacer(),
			archToggle,
			layout.NewSpacer(),
			ca.operationsStatusLabel,
		),
		ca.operationsModeHelp,
		widget.NewSeparator(),
	)
}

// createDACInputSection creates the DAC input section with per-column voltage entries
func (ca *CircuitsApp) createDACInputSection() fyne.CanvasObject {
	// Per-column voltage entries (show first 8 columns)
	maxCols := min(8, ca.arrayCols)
	ca.unifiedDACEntries = make([]*widget.Entry, maxCols)
	ca.unifiedDACLabels = make([]*widget.Label, maxCols)

	entriesRow := container.NewHBox()
	for i := 0; i < maxCols; i++ {
		idx := i
		entry := widget.NewEntry()
		entry.SetText(fmt.Sprintf("%.2f", ca.deviceState.GetDACVoltage(i)))
		entry.OnChanged = func(s string) {
			ca.onDACVoltageChanged(idx, s)
		}
		ca.unifiedDACEntries[i] = entry

		label := widget.NewLabel(fmt.Sprintf("BL%d", i))
		label.TextStyle = fyne.TextStyle{Monospace: true}
		ca.unifiedDACLabels[i] = label

		col := container.NewVBox(label, entry)
		entriesRow.Add(col)
	}

	// Preset buttons
	presetRead := widget.NewButton("Read (0.5V)", func() {
		ca.setUnifiedDACPreset(DACReadPreset)
	})
	presetWrite := widget.NewButton("Write (1.5V)", func() {
		ca.setUnifiedDACPreset(DACWritePreset)
	})
	presetCompute := widget.NewButton("Input Vector", func() {
		ca.showInputVectorDialog()
	})
	presetRandom := widget.NewButton("Random", func() {
		ca.setUnifiedDACPreset(DACRandom)
	})

	// "Set All" entry for bulk voltage
	allEntry := widget.NewEntry()
	allEntry.SetPlaceHolder("Set all")
	allEntry.OnSubmitted = func(s string) {
		ca.setAllUnifiedDACVoltages(s)
	}

	presetsRow := container.NewHBox(
		widget.NewLabel("Presets:"),
		presetRead, presetWrite, presetCompute, presetRandom,
		layout.NewSpacer(),
		widget.NewLabel("All (V):"), allEntry,
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("DAC INPUTS (Voltage per column)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		entriesRow,
		presetsRow,
		widget.NewSeparator(),
	)
}

// createMainSimSection creates the main simulation visualization area
func (ca *CircuitsApp) createMainSimSection() fyne.CanvasObject {
	// Left: WL selector
	wlSelector := ca.createWLSelector()

	// Center: Array canvas
	arraySection := ca.createUnifiedArraySection()

	// Right: Output display
	outputDisplay := ca.createOutputDisplay()

	// Combine with HSplit for flexible sizing
	leftPanel := container.NewVBox(
		widget.NewLabelWithStyle("WORD LINES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		wlSelector,
	)

	rightPanel := container.NewVBox(
		widget.NewLabelWithStyle("OUTPUTS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		outputDisplay,
	)

	// Main content with array in center
	mainContent := container.NewBorder(
		nil, nil,
		leftPanel,
		rightPanel,
		arraySection,
	)

	return mainContent
}

// createWLSelector creates the word line selection controls
func (ca *CircuitsApp) createWLSelector() fyne.CanvasObject {
	// Row checkboxes (show first 8)
	maxRows := min(8, ca.arrayRows)
	ca.unifiedWLChecks = make([]*widget.Check, maxRows)

	checkboxes := container.NewVBox()
	for i := 0; i < maxRows; i++ {
		idx := i
		check := widget.NewCheck(fmt.Sprintf("WL%d", i), func(checked bool) {
			ca.onWLChanged(idx, checked)
		})
		// Initially only row 0 is checked
		check.SetChecked(i == 0)
		ca.unifiedWLChecks[i] = check
		checkboxes.Add(check)
	}

	// Mode buttons
	singleBtn := widget.NewButton("Single Row", func() {
		ca.setWLModeSingle()
	})
	singleBtn.Importance = widget.HighImportance

	allBtn := widget.NewButton("All Rows", func() {
		ca.setWLModeAll()
	})

	modeButtons := container.NewVBox(
		widget.NewSeparator(),
		singleBtn,
		allBtn,
	)

	return container.NewVBox(
		container.NewVScroll(checkboxes),
		modeButtons,
	)
}

// createUnifiedArraySection creates the array visualization section
func (ca *CircuitsApp) createUnifiedArraySection() fyne.CanvasObject {
	// Create tappable array canvas
	tappableArray := NewUnifiedTappableCanvas(ca, ca.drawUnifiedArray, ca.onUnifiedCellTapped)
	tappableArray.SetMinSize(fyne.NewSize(400, 350))
	ca.sharedArrayCanvas = tappableArray.raster

	// Cell info display
	ca.sharedCellInfoLabel = widget.NewLabel("Click a cell to select")

	// Array size info
	ca.sharedArrayInfoLabel = widget.NewLabel(fmt.Sprintf("Array: %dx%d | %d levels", ca.arrayRows, ca.arrayCols, ca.quantLevels))

	// Legend
	legendLabel := widget.NewLabel("Level: Low (blue) -> High (red) | Yellow = Selected")
	legendLabel.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		tappableArray,
		legendLabel,
		ca.sharedCellInfoLabel,
		ca.sharedArrayInfoLabel,
	)
}

// createOutputDisplay creates the output display section
func (ca *CircuitsApp) createOutputDisplay() fyne.CanvasObject {
	// Output labels for each row (show first 8)
	maxRows := min(8, ca.arrayRows)
	ca.unifiedOutputLabels = make([]*widget.Label, maxRows)

	outputList := container.NewVBox()
	for i := 0; i < maxRows; i++ {
		label := widget.NewLabel(fmt.Sprintf("y%d: --", i))
		label.TextStyle = fyne.TextStyle{Monospace: true}
		ca.unifiedOutputLabels[i] = label
		outputList.Add(label)
	}

	// TIA/ADC info
	tiaInfo := widget.NewLabel("TIA: 10k gain")
	adcInfo := widget.NewLabel("ADC: 5-bit")

	return container.NewVBox(
		container.NewVScroll(outputList),
		widget.NewSeparator(),
		tiaInfo,
		adcInfo,
	)
}

// createUnifiedActionSection creates the action buttons
func (ca *CircuitsApp) createUnifiedActionSection() fyne.CanvasObject {
	// Program button - only enabled when single row + high voltage
	programBtn := widget.NewButton("Program Cell", func() {
		ca.onUnifiedProgram()
	})
	programBtn.Importance = widget.HighImportance

	// Read button
	readBtn := widget.NewButton("Read/Sense", func() {
		ca.onUnifiedRead()
	})

	// Compute button
	computeBtn := widget.NewButton("Compute MVM", func() {
		ca.onUnifiedCompute()
	})

	// Animate button
	animateBtn := widget.NewButton("Animate", func() {
		ca.onUnifiedAnimate()
	})

	// Reset array button
	resetBtn := widget.NewButton("Reset Array", func() {
		ca.onUnifiedReset()
	})

	// Random array button
	randomBtn := widget.NewButton("Random Array", func() {
		ca.onUnifiedRandomArray()
	})

	return container.NewHBox(
		programBtn, readBtn, computeBtn,
		layout.NewSpacer(),
		animateBtn, randomBtn, resetBtn,
	)
}

// ============================================================================
// UNIFIED TAPPABLE CANVAS
// ============================================================================

// UnifiedTappableCanvas is a canvas.Raster that responds to taps for the unified view
type UnifiedTappableCanvas struct {
	widget.BaseWidget
	raster *canvas.Raster
	onTap  func(row, col int)
	ca     *CircuitsApp
}

func NewUnifiedTappableCanvas(ca *CircuitsApp, drawFunc func(w, h int) image.Image, onTap func(row, col int)) *UnifiedTappableCanvas {
	t := &UnifiedTappableCanvas{
		raster: canvas.NewRaster(drawFunc),
		onTap:  onTap,
		ca:     ca,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *UnifiedTappableCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.raster)
}

func (t *UnifiedTappableCanvas) SetMinSize(size fyne.Size) {
	t.raster.SetMinSize(size)
}

func (t *UnifiedTappableCanvas) Refresh() {
	t.raster.Refresh()
}

func (t *UnifiedTappableCanvas) Tapped(e *fyne.PointEvent) {
	size := t.raster.Size()

	t.ca.mu.RLock()
	rows := t.ca.arrayRows
	cols := t.ca.arrayCols
	t.ca.mu.RUnlock()

	w := int(size.Width)
	h := int(size.Height)

	// Same margin calculations as drawUnifiedArray
	topMargin := 50
	rightMargin := 20
	bottomMargin := 25
	leftMargin := 25

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows
	cellSize := min(cellW, cellH)
	if cellSize > 40 {
		cellSize = 40
	}
	if cellSize < 8 {
		cellSize = 8
	}

	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	col := (int(e.Position.X) - offsetX) / cellSize
	row := (int(e.Position.Y) - offsetY) / cellSize

	if row >= 0 && row < rows && col >= 0 && col < cols {
		t.onTap(row, col)
	}
}

func (t *UnifiedTappableCanvas) TappedSecondary(*fyne.PointEvent) {}

func (t *UnifiedTappableCanvas) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// ============================================================================
// UNIFIED ARRAY DRAWING
// ============================================================================

// drawUnifiedArray draws the unified array visualization
func (ca *CircuitsApp) drawUnifiedArray(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	levels := ca.quantLevels
	arch := ca.architecture
	animStep := ca.animationStep
	ca.mu.RUnlock()

	if ca.deviceState == nil {
		return img
	}

	// Draw gradient background
	bgTop := color.RGBA{12, 20, 35, 255}
	bgBottom := color.RGBA{8, 14, 28, 255}
	drawGradientRect(img, 0, 0, w, h, bgTop, bgBottom)

	if weights == nil || len(weights) == 0 {
		return img
	}

	// Calculate margins
	topMargin := 50
	rightMargin := 20
	bottomMargin := 25
	leftMargin := 25

	is1T1R := arch == sharedwidgets.Architecture1T1R
	is2T1R := arch == sharedwidgets.Architecture2T1R
	if is1T1R || is2T1R {
		leftMargin = 50
	}
	if is2T1R {
		bottomMargin = 50
	}

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows
	cellSize := min(cellW, cellH)
	if cellSize > 40 {
		cellSize = 40
	}
	if cellSize < 12 {
		cellSize = 12
	}

	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	// Store for click detection
	ca.mu.Lock()
	ca.sharedArrayCellSize = cellSize
	ca.sharedArrayOffsetX = offsetX
	ca.sharedArrayOffsetY = offsetY
	ca.mu.Unlock()

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()

	// Draw array background panel
	panelColor := color.RGBA{18, 28, 45, 255}
	drawRoundedRect(img, offsetX-6, offsetY-6, gridW+12, gridH+12, 8, panelColor)

	// Draw BIT LINES (vertical) - color based on DAC voltage
	for c := 0; c < cols; c++ {
		x := offsetX + c*cellSize + cellSize/2
		voltage := ca.deviceState.GetDACVoltage(c)

		var blCol color.RGBA
		if voltage >= VoltageThresholdWrite {
			blCol = color.RGBA{255, 100, 100, 255} // Red - write voltage
		} else if voltage > 0.1 {
			blCol = color.RGBA{100, 180, 255, 255} // Blue - read/compute voltage
		} else {
			blCol = color.RGBA{50, 60, 80, 150} // Dim - no signal
		}

		// Highlight selected column
		if c == selectedCol {
			blCol.R = uint8(min(int(blCol.R)+50, 255))
			blCol.G = uint8(min(int(blCol.G)+50, 255))
			blCol.B = uint8(min(int(blCol.B)+50, 255))
		}

		for y := offsetY - 20; y < offsetY+gridH+8; y++ {
			if y >= 0 && y < h {
				img.Set(x, y, blCol)
				if cellSize > 16 {
					img.Set(x+1, y, blCol)
				}
			}
		}
	}

	// Draw WORD LINES (horizontal) - color based on active state
	for r := 0; r < rows; r++ {
		y := offsetY + r*cellSize + cellSize/2
		isActive := ca.deviceState.IsRowActive(r)

		var wlCol color.RGBA
		if isActive {
			wlCol = color.RGBA{255, 180, 100, 255} // Bright orange - active
		} else {
			wlCol = color.RGBA{60, 50, 40, 150} // Dim - inactive
		}

		// Highlight selected row
		if r == selectedRow {
			wlCol.R = uint8(min(int(wlCol.R)+30, 255))
			wlCol.G = uint8(min(int(wlCol.G)+30, 255))
		}

		startX := offsetX - 15
		if is1T1R || is2T1R {
			startX = offsetX - 8
		}

		for x := startX; x < offsetX+gridW+15; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, wlCol)
				if cellSize > 16 {
					img.Set(x, y+1, wlCol)
				}
			}
		}
	}

	// Draw 1T1R/2T1R transistors
	if is1T1R || is2T1R {
		ca.drawRowTransistors(img, offsetX, offsetY, cellSize, rows, gridH, w, h)
	}
	if is2T1R {
		ca.drawColTransistors(img, offsetX, offsetY, cellSize, cols, gridW, gridH, w, h)
	}

	// Draw cells
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := offsetX + c*cellSize + 2
			y0 := offsetY + r*cellSize + 2
			cw := cellSize - 4
			ch := cellSize - 4

			level := weights[r][c]
			isSelected := r == selectedRow && c == selectedCol
			isActive := ca.deviceState.IsRowActive(r) && ca.deviceState.GetDACVoltage(c) > 0.01

			var cellColor color.RGBA
			if isSelected {
				cellColor = color.RGBA{255, 230, 100, 255}
			} else {
				cellColor = levelToColor(level, levels)
				// Dim inactive cells
				if !isActive {
					cellColor.R = uint8(float64(cellColor.R) * 0.4)
					cellColor.G = uint8(float64(cellColor.G) * 0.4)
					cellColor.B = uint8(float64(cellColor.B) * 0.4)
				}
			}

			// Animation highlight
			if animStep == 2 && isActive {
				cellColor.R = uint8(min(int(cellColor.R)+40, 255))
				cellColor.G = uint8(min(int(cellColor.G)+40, 255))
			}

			// Draw cell with 3D effect
			topColor := color.RGBA{
				uint8(min(int(cellColor.R)+35, 255)),
				uint8(min(int(cellColor.G)+35, 255)),
				uint8(min(int(cellColor.B)+35, 255)),
				255,
			}
			drawGradientRect(img, x0, y0, cw, ch, topColor, cellColor)

			// Border
			borderColor := color.RGBA{
				uint8(min(int(cellColor.R)+60, 255)),
				uint8(min(int(cellColor.G)+60, 255)),
				uint8(min(int(cellColor.B)+60, 255)),
				255,
			}
			drawRectBorder(img, x0, y0, cw, ch, borderColor)

			// Selected cell glow
			if isSelected {
				white := color.RGBA{255, 255, 255, 255}
				drawRectBorder(img, x0-1, y0-1, cw+2, ch+2, white)
				drawRectBorder(img, x0-2, y0-2, cw+4, ch+4, color.RGBA{255, 255, 180, 150})
			}
		}
	}

	// Draw DAC boxes (top)
	dacBoxH := 25
	dacBoxW := cellSize - 2
	if dacBoxW < 24 {
		dacBoxW = 24
	}
	dacY := offsetY - dacBoxH - 15

	for c := 0; c < min(8, cols); c++ {
		dacX := offsetX + c*cellSize + 1
		voltage := ca.deviceState.GetDACVoltage(c)
		highlighted := animStep == 1
		drawDACColumn(img, dacX, dacY, dacBoxW, dacBoxH, voltage, "", highlighted, false)
	}

	// Draw TIA+ADC boxes (right side)
	tiaBoxW := 28
	adcBoxW := 24
	tiaAdcBoxH := cellSize - 2
	if tiaAdcBoxH < 18 {
		tiaAdcBoxH = 18
	}
	tiaX := offsetX + gridW + 10

	for r := 0; r < min(8, rows); r++ {
		tiaY := offsetY + r*cellSize + 1
		current := ca.deviceState.GetRowCurrent(r)
		level := ca.deviceState.GetRowLevel(r)
		highlighted := animStep == 3
		dimmed := !ca.deviceState.IsRowActive(r)
		drawTIAADCRow(img, tiaX, tiaY, tiaBoxW, adcBoxW, tiaAdcBoxH, current, level, "", highlighted, dimmed, ca.tia, ca.adc)
	}

	// Draw labels
	drawSimpleText(img, "DAC", offsetX-25, dacY+dacBoxH/2-3, color.RGBA{170, 140, 220, 255})
	drawSimpleText(img, "TIA", tiaX, offsetY-10, color.RGBA{220, 180, 100, 255})
	drawSimpleText(img, "ADC", tiaX+tiaBoxW+4, offsetY-10, color.RGBA{130, 210, 170, 255})

	// Operation classification title
	opText := ca.deviceState.ClassifyOperation()
	var opColor color.RGBA
	switch {
	case opText == "PROGRAM":
		opColor = color.RGBA{255, 200, 100, 255}
	case opText == "READ":
		opColor = color.RGBA{100, 220, 255, 255}
	case opText == "COMPUTE (MVM)":
		opColor = color.RGBA{200, 150, 255, 255}
	default:
		opColor = color.RGBA{150, 150, 150, 255}
	}
	drawSimpleText(img, opText, 10, 8, opColor)

	// Architecture badge
	var archText string
	var archColor color.RGBA
	switch arch {
	case sharedwidgets.Architecture2T1R:
		archText = "2T1R"
		archColor = color.RGBA{100, 180, 220, 255}
	case sharedwidgets.Architecture1T1R:
		archText = "1T1R"
		archColor = color.RGBA{100, 220, 120, 255}
	default:
		archText = "PASSIVE"
		archColor = color.RGBA{220, 150, 100, 255}
	}
	drawSimpleText(img, archText, w-len(archText)*6-10, 8, archColor)

	return img
}

// drawRowTransistors draws the row transistors for 1T1R/2T1R architecture
func (ca *CircuitsApp) drawRowTransistors(img *image.RGBA, offsetX, offsetY, cellSize, rows, gridH, w, h int) {
	for r := 0; r < rows; r++ {
		ty := offsetY + r*cellSize + cellSize/2
		tx := offsetX - 28

		transistorOn := ca.deviceState.IsRowActive(r)

		var bodyCol, gateCol, channelCol color.RGBA
		if transistorOn {
			bodyCol = color.RGBA{60, 200, 80, 255}
			gateCol = color.RGBA{100, 255, 120, 255}
			channelCol = color.RGBA{80, 220, 100, 255}
		} else {
			bodyCol = color.RGBA{50, 50, 60, 255}
			gateCol = color.RGBA{70, 70, 80, 255}
			channelCol = color.RGBA{40, 40, 50, 255}
		}

		// Draw MOSFET body
		for dy := -6; dy <= 6; dy++ {
			for dx := 0; dx < 3; dx++ {
				px, py := tx+dx, ty+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, bodyCol)
				}
			}
		}

		// Draw gate
		gateX := tx - 5
		for dy := -8; dy <= 8; dy++ {
			py := ty + dy
			if gateX >= 0 && gateX < w && py >= 0 && py < h {
				img.Set(gateX, py, gateCol)
			}
		}

		// Draw channel
		for dx := 3; dx < 18; dx++ {
			px := tx + dx
			if px >= 0 && px < w {
				img.Set(px, ty, channelCol)
			}
		}

		// ON indicator
		if transistorOn {
			drawGlowCircle(img, tx+1, ty, 2, color.RGBA{150, 255, 150, 255}, color.RGBA{100, 200, 100, 80})
		}
	}
}

// drawColTransistors draws the column transistors for 2T1R architecture
func (ca *CircuitsApp) drawColTransistors(img *image.RGBA, offsetX, offsetY, cellSize, cols, gridW, gridH, w, h int) {
	for c := 0; c < cols; c++ {
		tx := offsetX + c*cellSize + cellSize/2
		ty := offsetY + gridH + 20

		// In 2T1R, column transistors are controlled by CSL
		// For simplicity, all column transistors are ON when computing
		transistorOn := ca.deviceState.GetWLMode() == WLAll || c == ca.deviceState.GetSelectedCol()

		var bodyCol, gateCol, channelCol color.RGBA
		if transistorOn {
			bodyCol = color.RGBA{60, 180, 200, 255}
			gateCol = color.RGBA{100, 220, 255, 255}
			channelCol = color.RGBA{80, 200, 220, 255}
		} else {
			bodyCol = color.RGBA{50, 50, 60, 255}
			gateCol = color.RGBA{70, 70, 80, 255}
			channelCol = color.RGBA{40, 40, 50, 255}
		}

		// Draw MOSFET body (horizontal)
		for dx := -6; dx <= 6; dx++ {
			for dy := 0; dy < 3; dy++ {
				px, py := tx+dx, ty+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, bodyCol)
				}
			}
		}

		// Draw gate
		gateY := ty + 5
		for dx := -8; dx <= 8; dx++ {
			px := tx + dx
			if px >= 0 && px < w && gateY >= 0 && gateY < h {
				img.Set(px, gateY, gateCol)
			}
		}

		// Draw channel
		for dy := -15; dy < 0; dy++ {
			py := ty + dy
			if tx >= 0 && tx < w && py >= 0 && py < h {
				img.Set(tx, py, channelCol)
			}
		}

		// ON indicator
		if transistorOn {
			drawGlowCircle(img, tx, ty+1, 2, color.RGBA{150, 220, 255, 255}, color.RGBA{100, 180, 200, 80})
		}
	}
}

// ============================================================================
// EVENT HANDLERS
// ============================================================================

// onDACVoltageChanged handles DAC voltage input changes
func (ca *CircuitsApp) onDACVoltageChanged(col int, voltageStr string) {
	voltage, err := strconv.ParseFloat(voltageStr, 64)
	if err != nil {
		return
	}

	// Clamp voltage to reasonable range
	if voltage < 0 {
		voltage = 0
	}
	if voltage > 2.0 {
		voltage = 2.0
	}

	ca.deviceState.SetDACVoltage(col, voltage)
	ca.recomputeAndRefresh()
}

// setUnifiedDACPreset applies a DAC preset
func (ca *CircuitsApp) setUnifiedDACPreset(preset DACMode) {
	switch preset {
	case DACReadPreset:
		ca.deviceState.SetDACPreset(DACReadPreset, VoltageThresholdRead)
	case DACWritePreset:
		ca.deviceState.SetDACPreset(DACWritePreset, VoltageMaxWrite)
	case DACRandom:
		// Random voltages between 0 and 1V
		for i := range ca.unifiedDACEntries {
			voltage := rand.Float64()
			ca.deviceState.SetDACVoltage(i, voltage)
		}
	}
	ca.updateDACEntries()
	ca.recomputeAndRefresh()
}

// setAllUnifiedDACVoltages sets all DAC voltages to the same value
func (ca *CircuitsApp) setAllUnifiedDACVoltages(voltageStr string) {
	voltage, err := strconv.ParseFloat(voltageStr, 64)
	if err != nil {
		return
	}
	ca.deviceState.SetAllDACVoltages(voltage)
	ca.updateDACEntries()
	ca.recomputeAndRefresh()
}

// showInputVectorDialog shows a dialog for entering input vector values
func (ca *CircuitsApp) showInputVectorDialog() {
	// For simplicity, set random input vector (0-255) converted to voltage
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = rand.Intn(256)
	}
	ca.mu.Unlock()

	// Convert to DAC voltages (0-255 -> 0-1V)
	params := make([]float64, len(ca.inputVector))
	for i, v := range ca.inputVector {
		params[i] = float64(v)
	}
	ca.deviceState.SetDACPreset(DACInputVector, params...)

	ca.updateDACEntries()
	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Input vector applied (0-255 -> 0-1V)")
}

// onWLChanged handles WL checkbox changes
func (ca *CircuitsApp) onWLChanged(row int, checked bool) {
	if ca.deviceState == nil {
		return
	}

	// Update the active rows pattern
	pattern := make([]bool, ca.arrayRows)
	for i := 0; i < len(ca.unifiedWLChecks) && i < ca.arrayRows; i++ {
		if ca.unifiedWLChecks[i] != nil {
			pattern[i] = ca.unifiedWLChecks[i].Checked
		}
	}
	ca.deviceState.SetWLCustom(pattern)

	ca.recomputeAndRefresh()
}

// setWLModeSingle sets WL mode to single (only selected row)
func (ca *CircuitsApp) setWLModeSingle() {
	selectedRow := ca.deviceState.GetSelectedRow()
	ca.deviceState.SetWLSingle(selectedRow)
	ca.updateWLCheckboxes()
	ca.recomputeAndRefresh()
}

// setWLModeAll sets WL mode to all rows active
func (ca *CircuitsApp) setWLModeAll() {
	ca.deviceState.SetWLAll()
	ca.updateWLCheckboxes()
	ca.recomputeAndRefresh()
}

// onUnifiedCellTapped handles cell selection
func (ca *CircuitsApp) onUnifiedCellTapped(row, col int) {
	ca.deviceState.SetSelectedCell(row, col)

	// If in single mode, also update WL
	if ca.deviceState.GetWLMode() == WLSingle {
		ca.deviceState.SetWLSingle(row)
		ca.updateWLCheckboxes()
	}

	ca.recomputeAndRefresh()
	ca.updateCellInfo()
}

// ============================================================================
// ACTION HANDLERS
// ============================================================================

// onUnifiedProgram programs the selected cell
func (ca *CircuitsApp) onUnifiedProgram() {
	// Check if we have write voltage
	selectedCol := ca.deviceState.GetSelectedCol()
	voltage := ca.deviceState.GetDACVoltage(selectedCol)

	if voltage < VoltageThresholdWrite {
		ca.operationsStatusLabel.SetText(fmt.Sprintf("Warning: Voltage %.2fV below write threshold (%.1fV)", voltage, VoltageThresholdWrite))
		return
	}

	selectedRow := ca.deviceState.GetSelectedRow()

	// Calculate target level from voltage
	targetLevel := int((voltage - VoltageThresholdWrite) / (VoltageMaxWrite - VoltageThresholdWrite) * float64(ca.quantLevels-1))
	if targetLevel < 0 {
		targetLevel = 0
	}
	if targetLevel >= ca.quantLevels {
		targetLevel = ca.quantLevels - 1
	}

	// Update array weight
	ca.mu.Lock()
	if selectedRow < len(ca.arrayWeights) && selectedCol < len(ca.arrayWeights[selectedRow]) {
		ca.arrayWeights[selectedRow][selectedCol] = targetLevel
	}
	ca.mu.Unlock()

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText(fmt.Sprintf("Programmed [%d,%d] = Level %d", selectedRow, selectedCol, targetLevel))
}

// onUnifiedRead reads the selected cell
func (ca *CircuitsApp) onUnifiedRead() {
	ca.recomputeAndRefresh()

	selectedRow := ca.deviceState.GetSelectedRow()
	level := ca.deviceState.GetRowLevel(selectedRow)
	current := ca.deviceState.GetRowCurrent(selectedRow)

	ca.operationsStatusLabel.SetText(fmt.Sprintf("Read [%d,*]: %.1fuA -> Level %d", selectedRow, current, level))
}

// onUnifiedCompute runs MVM computation
func (ca *CircuitsApp) onUnifiedCompute() {
	// Ensure all rows are active for MVM
	ca.deviceState.SetWLAll()
	ca.updateWLCheckboxes()

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Compute complete in ~20ns (parallel MVM)")
}

// onUnifiedAnimate animates the signal flow
func (ca *CircuitsApp) onUnifiedAnimate() {
	ca.mu.Lock()
	ca.animationActive = true
	ca.mu.Unlock()

	ca.operationsStatusLabel.SetText("Animating...")

	go func() {
		// Step 1: DAC
		ca.mu.Lock()
		ca.animationStep = 1
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 1: DAC conversion (5ns)")
		})
		ca.sleep(600)

		// Step 2: Array
		ca.mu.Lock()
		ca.animationStep = 2
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 2: Array MVM (5ns)")
		})
		ca.sleep(600)

		// Step 3: ADC
		ca.mu.Lock()
		ca.animationStep = 3
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 3: TIA+ADC conversion (10ns)")
		})
		ca.sleep(600)

		// Complete
		ca.mu.Lock()
		ca.animationStep = 0
		ca.animationActive = false
		ca.mu.Unlock()
		ca.recomputeAndRefresh()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Complete in ~20ns")
		})
	}()
}

// onUnifiedReset resets the array to random values
func (ca *CircuitsApp) onUnifiedReset() {
	// Reset DAC to read preset
	ca.deviceState.SetDACPreset(DACReadPreset, VoltageThresholdRead)
	ca.updateDACEntries()

	// Reset WL to single row 0
	ca.deviceState.SetWLSingle(0)
	ca.updateWLCheckboxes()

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Reset complete")
}

// onUnifiedRandomArray randomizes the array weights
func (ca *CircuitsApp) onUnifiedRandomArray() {
	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = rand.Intn(ca.quantLevels)
		}
	}
	ca.mu.Unlock()

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Array randomized")
}

// ============================================================================
// UI UPDATE HELPERS
// ============================================================================

// recomputeAndRefresh runs computation and updates all UI elements
func (ca *CircuitsApp) recomputeAndRefresh() {
	ca.mu.RLock()
	weights := ca.arrayWeights
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Run device simulation
	ca.deviceState.Compute(weights, levels)

	// Update output display
	ca.updateOutputDisplay()

	// Update cell info
	ca.updateCellInfo()

	// Update operation classification
	ca.updateOperationClassification()

	// Refresh array canvas
	ca.refreshUnifiedArray()
}

// refreshUnifiedArray refreshes the array canvas
func (ca *CircuitsApp) refreshUnifiedArray() {
	if ca.sharedArrayCanvas != nil {
		fyne.Do(func() {
			ca.sharedArrayCanvas.Refresh()
		})
	}
}

// updateDACEntries updates the DAC entry widgets
func (ca *CircuitsApp) updateDACEntries() {
	for i, entry := range ca.unifiedDACEntries {
		if entry != nil {
			voltage := ca.deviceState.GetDACVoltage(i)
			fyne.Do(func() {
				entry.SetText(fmt.Sprintf("%.2f", voltage))
			})
		}
	}
}

// updateWLCheckboxes updates the WL checkbox states
func (ca *CircuitsApp) updateWLCheckboxes() {
	for i, check := range ca.unifiedWLChecks {
		if check != nil {
			isActive := ca.deviceState.IsRowActive(i)
			fyne.Do(func() {
				check.SetChecked(isActive)
			})
		}
	}
}

// updateOutputDisplay updates the output labels
func (ca *CircuitsApp) updateOutputDisplay() {
	for i, label := range ca.unifiedOutputLabels {
		if label != nil {
			current := ca.deviceState.GetRowCurrent(i)
			tiaV := ca.deviceState.GetRowVoltage(i)
			level := ca.deviceState.GetRowLevel(i)
			sat := ca.deviceState.IsSaturated(i)
			active := ca.deviceState.IsRowActive(i)

			idx := i
			fyne.Do(func() {
				if !active {
					ca.unifiedOutputLabels[idx].SetText(fmt.Sprintf("y%d: (inactive)", idx))
				} else if sat {
					ca.unifiedOutputLabels[idx].SetText(fmt.Sprintf("y%d: %.1fuA->%.2fV->L%d SAT", idx, current, tiaV, level))
				} else {
					ca.unifiedOutputLabels[idx].SetText(fmt.Sprintf("y%d: %.1fuA->%.2fV->L%d", idx, current, tiaV, level))
				}
			})
		}
	}
}

// updateCellInfo updates the cell info display
func (ca *CircuitsApp) updateCellInfo() {
	if ca.sharedCellInfoLabel == nil {
		return
	}

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()

	ca.mu.RLock()
	var level int
	if selectedRow < len(ca.arrayWeights) && selectedCol < len(ca.arrayWeights[selectedRow]) {
		level = ca.arrayWeights[selectedRow][selectedCol]
	}
	levels := ca.quantLevels
	ca.mu.RUnlock()

	conductance := 1.0 + float64(level)/float64(levels-1)*99.0
	voltage := ca.deviceState.GetDACVoltage(selectedCol)

	fyne.Do(func() {
		ca.sharedCellInfoLabel.SetText(fmt.Sprintf("Cell [%d,%d]: Level %d | G=%.1fuS | BL=%.2fV", selectedRow, selectedCol, level, conductance, voltage))
	})
}

// updateOperationClassification updates the operation classification display
func (ca *CircuitsApp) updateOperationClassification() {
	if ca.operationsModeHelp == nil {
		return
	}

	opText := ca.deviceState.ClassifyOperation()
	arch := ca.architecture

	var helpText string
	switch opText {
	case "PROGRAM":
		if arch == sharedwidgets.Architecture2T1R {
			helpText = "PROGRAM: Row+Col transistors select single cell. Zero disturb to neighbors."
		} else if arch == sharedwidgets.Architecture1T1R {
			helpText = "PROGRAM: Row transistor gates selected row. Full write to target cell."
		} else {
			helpText = "PROGRAM: Passive array - partial voltages may disturb neighbors."
		}
	case "READ":
		if arch == sharedwidgets.Architecture2T1R {
			helpText = "READ: Dual transistor AND-gate selects single cell. Perfect isolation."
		} else if arch == sharedwidgets.Architecture1T1R {
			helpText = "READ: Row transistor isolates selected row. Clean sense current."
		} else {
			helpText = "READ: Passive array - sneak currents add noise to sense signal."
		}
	case "COMPUTE (MVM)":
		helpText = "COMPUTE: All WLs active. Full matrix-vector multiply in ~20ns."
	case "BULK PROGRAM (CAUTION)":
		helpText = "CAUTION: All rows active with write voltage. May cause unintended writes!"
	default:
		helpText = "Configure WL selection and DAC voltages to perform operations."
	}

	fyne.Do(func() {
		ca.operationsModeHelp.SetText(helpText)
	})
}

// ============================================================================
// ARCHITECTURE TOGGLE
// ============================================================================

// createArchitectureToggle creates the PASSIVE/1T1R/2T1R toggle buttons
func (ca *CircuitsApp) createArchitectureToggle() fyne.CanvasObject {
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
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture0T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.recomputeAndRefresh()
	}

	ca.arch1T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture1T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture1T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.recomputeAndRefresh()
	}

	ca.arch2T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture2T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture2T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.recomputeAndRefresh()
	}

	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)

	archLabel := widget.NewLabel("Array:")
	return container.NewHBox(archLabel, ca.archToggle)
}
