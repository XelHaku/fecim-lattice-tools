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

	sharedwidgets "fecim-lattice-tools/shared/widgets"
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

// createModeSelector creates the WRITE/READ/COMPUTE mode toggle with architecture selector
func (ca *CircuitsApp) createModeSelector() fyne.CanvasObject {
	modeRadio := widget.NewRadioGroup([]string{"WRITE", "READ", "COMPUTE"}, func(mode string) {
		ca.onModeChanged(mode)
	})
	modeRadio.Horizontal = true
	modeRadio.SetSelected("WRITE")

	modeHelp := widget.NewLabel("")
	modeHelp.TextStyle = fyne.TextStyle{Italic: true}
	ca.operationsModeHelp = modeHelp

	// Create architecture toggle (1T1R vs 0T1R)
	archToggle := ca.createArchitectureToggle()

	ca.updateModeHelp()

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Mode:"),
			modeRadio,
			layout.NewSpacer(),
			archToggle,
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

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := ca.arrayWeights
	selectedRow := ca.selectedRow
	selectedCol := ca.selectedCol
	levels := ca.quantLevels
	mode := ca.currentMode
	animStep := ca.animationStep
	arch := ca.architecture
	ca.mu.RUnlock()

	// Draw gradient background
	bgTop := color.RGBA{12, 20, 35, 255}
	bgBottom := color.RGBA{8, 14, 28, 255}
	drawGradientRect(img, 0, 0, w, h, bgTop, bgBottom)

	if weights == nil || len(weights) == 0 {
		return img
	}

	// Calculate margins - more space for transistors and labels
	topMargin := 80
	rightMargin := 80
	bottomMargin := 35
	leftMargin := 50
	is1T1R := arch == sharedwidgets.Architecture1T1R
	is2T1R := arch == sharedwidgets.Architecture2T1R
	if is1T1R || is2T1R {
		leftMargin = 75 // More space for row MOSFET symbols
	}
	if is2T1R {
		bottomMargin = 75 // More space for column MOSFET symbols below array
	}

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows
	cellSize := cellW
	if cellH < cellSize {
		cellSize = cellH
	}
	if cellSize > 42 {
		cellSize = 42
	}
	if cellSize < 12 {
		cellSize = 12
	}

	gridW := cols * cellSize
	gridH := rows * cellSize
	offsetX := leftMargin + (availableW-gridW)/2
	offsetY := topMargin + (availableH-gridH)/2

	ca.mu.Lock()
	ca.sharedArrayCellSize = cellSize
	ca.sharedArrayOffsetX = offsetX
	ca.sharedArrayOffsetY = offsetY
	ca.mu.Unlock()

	// ========== PASSIVE MODE: Draw sneak path visualization ==========
	if !is1T1R && !is2T1R && (mode == ModeRead || mode == ModeWrite) {
		// Show faded sneak paths through unselected cells
		sneakColor := color.RGBA{80, 40, 40, 60}
		for r := 0; r < rows; r++ {
			if r == selectedRow {
				continue
			}
			// Horizontal sneak indication
			y := offsetY + r*cellSize + cellSize/2
			for x := offsetX; x < offsetX+gridW; x++ {
				if x >= 0 && x < w {
					img.Set(x, y, sneakColor)
				}
			}
		}
		for c := 0; c < cols; c++ {
			if c == selectedCol {
				continue
			}
			// Vertical sneak indication
			x := offsetX + c*cellSize + cellSize/2
			for y := offsetY; y < offsetY+gridH; y++ {
				if y >= 0 && y < h {
					img.Set(x, y, sneakColor)
				}
			}
		}
	}

	// ========== Draw array background panel ==========
	panelColor := color.RGBA{18, 28, 45, 255}
	drawRoundedRect(img, offsetX-6, offsetY-6, gridW+12, gridH+12, 8, panelColor)

	// ========== Highlight selected row (1T1R/2T1R WRITE/READ) ==========
	if (is1T1R || is2T1R) && (mode == ModeWrite || mode == ModeRead) {
		rowHighlight := color.RGBA{40, 60, 40, 255}
		drawRect(img, offsetX-4, offsetY+selectedRow*cellSize-2, gridW+8, cellSize+4, rowHighlight)
	}

	// ========== Highlight selected column (2T1R WRITE/READ) ==========
	if is2T1R && (mode == ModeWrite || mode == ModeRead) {
		colHighlight := color.RGBA{40, 60, 60, 255}
		drawRect(img, offsetX+selectedCol*cellSize-2, offsetY-4, cellSize+4, gridH+8, colHighlight)
	}

	// ========== Highlight selected column (WRITE/READ) ==========
	if mode == ModeWrite || mode == ModeRead {
		colHighlight := color.RGBA{40, 40, 60, 200}
		drawRect(img, offsetX+selectedCol*cellSize-2, offsetY-4, cellSize+4, gridH+8, colHighlight)
	}

	// ========== Draw BIT LINES (vertical) ==========
	for c := 0; c < cols; c++ {
		x := offsetX + c*cellSize + cellSize/2
		isSelectedCol := (c == selectedCol)

		var blCol color.RGBA
		if mode == ModeCompute {
			blCol = color.RGBA{100, 130, 200, 255}
		} else if isSelectedCol {
			blCol = color.RGBA{100, 180, 255, 255} // Bright blue for selected
		} else {
			blCol = color.RGBA{50, 70, 100, 150}
		}

		// Draw from top to column transistors (if 2T1R) or just past grid
		endY := offsetY + gridH + 8
		if is2T1R {
			endY = offsetY + gridH + 20 // Extend to column transistors
		}
		for y := offsetY - 25; y < endY; y++ {
			if y >= 0 && y < h {
				img.Set(x, y, blCol)
				if cellSize > 16 {
					img.Set(x+1, y, blCol)
				}
			}
		}
	}

	// ========== Draw WORD LINES (horizontal) ==========
	for r := 0; r < rows; r++ {
		y := offsetY + r*cellSize + cellSize/2
		isSelectedRow := (r == selectedRow)

		var wlCol color.RGBA
		if mode == ModeCompute {
			wlCol = color.RGBA{200, 130, 100, 255}
		} else if isSelectedRow && (mode == ModeWrite || mode == ModeRead) {
			wlCol = color.RGBA{255, 180, 100, 255} // Bright orange for selected
		} else {
			wlCol = color.RGBA{100, 70, 50, 150}
		}

		startX := offsetX - 20
		if is1T1R || is2T1R {
			startX = offsetX - 8 // Start after transistor
		}

		for x := startX; x < offsetX+gridW+25; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, wlCol)
				if cellSize > 16 {
					img.Set(x, y+1, wlCol)
				}
			}
		}
	}

	// ========== Draw 1T1R/2T1R ROW MOSFET transistors (left side) ==========
	if is1T1R || is2T1R {
		for r := 0; r < rows; r++ {
			ty := offsetY + r*cellSize + cellSize/2
			tx := offsetX - 28

			var transistorOn bool
			switch mode {
			case ModeWrite, ModeRead:
				transistorOn = (r == selectedRow)
			case ModeCompute:
				transistorOn = true // All row transistors ON for MVM
			}

			// MOSFET symbol colors
			var bodyCol, gateCol, channelCol color.RGBA
			if transistorOn {
				bodyCol = color.RGBA{60, 200, 80, 255}    // Green body
				gateCol = color.RGBA{100, 255, 120, 255}  // Bright green gate
				channelCol = color.RGBA{80, 220, 100, 255}
			} else {
				bodyCol = color.RGBA{50, 50, 60, 255}    // Gray body
				gateCol = color.RGBA{70, 70, 80, 255}    // Gray gate
				channelCol = color.RGBA{40, 40, 50, 255}
			}

			// Draw MOSFET body (vertical bar)
			for dy := -8; dy <= 8; dy++ {
				for dx := 0; dx < 4; dx++ {
					px, py := tx+dx, ty+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						img.Set(px, py, bodyCol)
					}
				}
			}

			// Draw gate (vertical line with gap)
			gateX := tx - 6
			for dy := -10; dy <= 10; dy++ {
				py := ty + dy
				if gateX >= 0 && gateX < w && py >= 0 && py < h {
					img.Set(gateX, py, gateCol)
					img.Set(gateX+1, py, gateCol)
				}
			}

			// Draw channel (horizontal connection to WL)
			for dx := 4; dx < 20; dx++ {
				px := tx + dx
				if px >= 0 && px < w {
					img.Set(px, ty, channelCol)
					img.Set(px, ty+1, channelCol)
				}
			}

			// Draw source/drain terminals
			termCol := channelCol
			// Source (top)
			for dy := -12; dy <= -8; dy++ {
				px := tx + 2
				py := ty + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, termCol)
				}
			}
			// Drain (bottom)
			for dy := 8; dy <= 12; dy++ {
				px := tx + 2
				py := ty + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, termCol)
				}
			}

			// Draw ON/OFF indicator
			if transistorOn {
				// Small glow
				drawGlowCircle(img, tx+2, ty, 3,
					color.RGBA{150, 255, 150, 255},
					color.RGBA{100, 200, 100, 80})
			}
		}

		// Labels
		drawSimpleText(img, "WL", offsetX-55, offsetY-12, color.RGBA{150, 200, 150, 255})
		if is2T1R {
			drawSimpleText(img, "Row T", offsetX-55, offsetY+gridH+8, color.RGBA{120, 160, 120, 200})
		} else {
			drawSimpleText(img, "Gate", offsetX-55, offsetY+gridH+8, color.RGBA{120, 160, 120, 200})
		}
	}

	// ========== Draw 2T1R COLUMN MOSFET transistors (bottom side) ==========
	if is2T1R {
		for c := 0; c < cols; c++ {
			tx := offsetX + c*cellSize + cellSize/2
			ty := offsetY + gridH + 28

			var transistorOn bool
			switch mode {
			case ModeWrite, ModeRead:
				transistorOn = (c == selectedCol)
			case ModeCompute:
				transistorOn = true // All column transistors ON for MVM
			}

			// MOSFET symbol colors (cyan/teal for column transistors)
			var bodyCol, gateCol, channelCol color.RGBA
			if transistorOn {
				bodyCol = color.RGBA{60, 180, 200, 255}    // Cyan body
				gateCol = color.RGBA{100, 220, 255, 255}   // Bright cyan gate
				channelCol = color.RGBA{80, 200, 220, 255}
			} else {
				bodyCol = color.RGBA{50, 50, 60, 255}    // Gray body
				gateCol = color.RGBA{70, 70, 80, 255}    // Gray gate
				channelCol = color.RGBA{40, 40, 50, 255}
			}

			// Draw MOSFET body (horizontal bar)
			for dx := -8; dx <= 8; dx++ {
				for dy := 0; dy < 4; dy++ {
					px, py := tx+dx, ty+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						img.Set(px, py, bodyCol)
					}
				}
			}

			// Draw gate (horizontal line with gap)
			gateY := ty + 6
			for dx := -10; dx <= 10; dx++ {
				px := tx + dx
				if px >= 0 && px < w && gateY >= 0 && gateY < h {
					img.Set(px, gateY, gateCol)
					img.Set(px, gateY+1, gateCol)
				}
			}

			// Draw channel (vertical connection to BL above)
			for dy := -20; dy < 0; dy++ {
				py := ty + dy
				if tx >= 0 && tx < w && py >= 0 && py < h {
					img.Set(tx, py, channelCol)
					img.Set(tx+1, py, channelCol)
				}
			}

			// Draw source/drain terminals
			termCol := channelCol
			// Left terminal
			for dx := -12; dx <= -8; dx++ {
				px := tx + dx
				py := ty + 2
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, termCol)
				}
			}
			// Right terminal
			for dx := 8; dx <= 12; dx++ {
				px := tx + dx
				py := ty + 2
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, termCol)
				}
			}

			// Draw ON/OFF indicator
			if transistorOn {
				// Small glow
				drawGlowCircle(img, tx, ty+2, 3,
					color.RGBA{150, 220, 255, 255},
					color.RGBA{100, 180, 200, 80})
			}
		}

		// CSL label
		drawSimpleText(img, "CSL", offsetX-28, offsetY+gridH+30, color.RGBA{100, 180, 200, 255})
		drawSimpleText(img, "Col T", offsetX+gridW+10, offsetY+gridH+30, color.RGBA{100, 160, 180, 200})
	}

	// ========== Draw cells ==========
	for r := 0; r < rows && r < len(weights); r++ {
		for c := 0; c < cols && c < len(weights[r]); c++ {
			x0 := offsetX + c*cellSize + 2
			y0 := offsetY + r*cellSize + 2
			cw := cellSize - 4
			ch := cellSize - 4

			level := weights[r][c]
			isSelected := r == selectedRow && c == selectedCol

			// Determine if this cell is "active" based on mode
			isActive := false
			switch mode {
			case ModeWrite, ModeRead:
				isActive = isSelected
			case ModeCompute:
				isActive = true
			}

			// Get cell color
			var cellColor color.RGBA
			if isSelected {
				cellColor = color.RGBA{255, 230, 100, 255}
			} else {
				cellColor = levelToColor(level, levels)
				// Dim inactive cells in WRITE/READ mode
				if !isActive && (mode == ModeWrite || mode == ModeRead) {
					cellColor.R = uint8(float64(cellColor.R) * 0.5)
					cellColor.G = uint8(float64(cellColor.G) * 0.5)
					cellColor.B = uint8(float64(cellColor.B) * 0.5)
				}
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

			// Compute animation
			if mode == ModeCompute && animStep == 2 {
				drawRectBorder(img, x0, y0, cw, ch, color.RGBA{0, 255, 255, 100})
			}
		}
	}

	// ========== Draw current flow arrows ==========
	if mode == ModeCompute || ((mode == ModeWrite || mode == ModeRead) && (is1T1R || is2T1R)) {
		arrowCol := color.RGBA{255, 255, 100, 200}

		if mode == ModeCompute {
			// Input arrows (down from DAC)
			for c := 0; c < min(cols, 8); c++ {
				ax := offsetX + c*cellSize + cellSize/2
				ay := offsetY - 18
				// Arrow pointing down
				for i := 0; i < 8; i++ {
					img.Set(ax, ay+i, arrowCol)
				}
				for i := 0; i < 4; i++ {
					img.Set(ax-i, ay+8-i, arrowCol)
					img.Set(ax+i, ay+8-i, arrowCol)
				}
			}

			// Output arrows (right to ADC)
			for r := 0; r < min(rows, 8); r++ {
				ax := offsetX + gridW + 8
				ay := offsetY + r*cellSize + cellSize/2
				// Arrow pointing right
				for i := 0; i < 8; i++ {
					img.Set(ax+i, ay, arrowCol)
				}
				for i := 0; i < 4; i++ {
					img.Set(ax+8-i, ay-i, arrowCol)
					img.Set(ax+8-i, ay+i, arrowCol)
				}
			}
		} else if mode == ModeWrite || mode == ModeRead {
			// Single cell current path
			cx := offsetX + selectedCol*cellSize + cellSize/2
			cy := offsetY + selectedRow*cellSize + cellSize/2

			// Vertical arrow to cell
			for i := 0; i < 12; i++ {
				img.Set(cx, offsetY-15+i, arrowCol)
			}
			// Horizontal arrow from cell
			for i := 0; i < 12; i++ {
				img.Set(offsetX+gridW+5+i, cy, arrowCol)
			}
		}
	}

	// ========== COMPUTE MODE: DAC and ADC boxes ==========
	if mode == ModeCompute {
		dacBoxH := 30
		dacBoxW := cellSize - 2
		dacY := offsetY - dacBoxH - 18

		dacColCount := min(8, cols)
		inputVectorCopy := make([]int, dacColCount)
		ca.mu.RLock()
		copy(inputVectorCopy, ca.inputVector[:dacColCount])
		ca.mu.RUnlock()

		for c := 0; c < dacColCount; c++ {
			dacX := offsetX + c*cellSize + 1

			dacTop := color.RGBA{130, 90, 190, 255}
			dacBot := color.RGBA{80, 50, 140, 255}
			if animStep == 1 {
				dacTop = color.RGBA{255, 255, 120, 255}
				dacBot = color.RGBA{220, 200, 80, 255}
			}

			drawGradientRect(img, dacX, dacY, dacBoxW, dacBoxH, dacTop, dacBot)
			drawRectBorder(img, dacX, dacY, dacBoxW, dacBoxH, color.RGBA{170, 140, 220, 255})

			// Voltage display
			voltage := float64(inputVectorCopy[c]) / 255.0
			vText := fmt.Sprintf("%.2f", voltage)
			drawSimpleText(img, vText, dacX+dacBoxW/2-len(vText)*3, dacY+dacBoxH/2-3, color.RGBA{255, 255, 255, 255})

			// Column label
			drawSimpleText(img, fmt.Sprintf("x%d", c), offsetX+c*cellSize+cellSize/2-6, dacY-12, color.RGBA{140, 170, 255, 255})
		}

		// DAC label
		drawSimpleText(img, "DAC", offsetX-28, dacY+dacBoxH/2-3, color.RGBA{170, 140, 220, 255})

		// ADC boxes
		adcBoxW := 52
		adcBoxH := cellSize - 2
		adcX := offsetX + gridW + 15

		adcRowCount := min(8, rows)
		outputVectorCopy := make([]float64, adcRowCount)
		ca.mu.RLock()
		copy(outputVectorCopy, ca.outputVector[:min(adcRowCount, len(ca.outputVector))])
		ca.mu.RUnlock()

		for r := 0; r < adcRowCount; r++ {
			adcY := offsetY + r*cellSize + 1

			adcTop := color.RGBA{70, 170, 130, 255}
			adcBot := color.RGBA{40, 120, 90, 255}
			if animStep == 3 {
				adcTop = color.RGBA{100, 255, 170, 255}
				adcBot = color.RGBA{70, 200, 130, 255}
			}

			drawGradientRect(img, adcX, adcY, adcBoxW, adcBoxH, adcTop, adcBot)
			drawRectBorder(img, adcX, adcY, adcBoxW, adcBoxH, color.RGBA{130, 210, 170, 255})

			// Level display
			tiaV := ca.tia.Convert(outputVectorCopy[r] * 1e-6)
			lvl := ca.adc.Convert(tiaV)
			lText := fmt.Sprintf("L%d", lvl)
			drawSimpleText(img, lText, adcX+adcBoxW/2-len(lText)*3, adcY+adcBoxH/2-3, color.RGBA{255, 255, 255, 255})

			// Row label
			drawSimpleText(img, fmt.Sprintf("y%d", r), adcX+adcBoxW+6, offsetY+r*cellSize+cellSize/2-3, color.RGBA{255, 190, 140, 255})
		}

		// ADC label
		drawSimpleText(img, "TIA+ADC", adcX, offsetY-12, color.RGBA{130, 210, 170, 255})
	}

	// ========== Mode title and architecture indicator ==========
	var titleText string
	var titleColor color.RGBA
	switch mode {
	case ModeWrite:
		titleText = "WRITE"
		titleColor = color.RGBA{255, 200, 100, 255}
	case ModeRead:
		titleText = "READ"
		titleColor = color.RGBA{100, 220, 255, 255}
	case ModeCompute:
		titleText = "COMPUTE (MVM)"
		titleColor = color.RGBA{200, 150, 255, 255}
	}

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

	// Draw title bar
	drawSimpleText(img, titleText, 10, 8, titleColor)
	drawSimpleText(img, archText, w-len(archText)*6-10, 8, archColor)

	// ========== Legend (bottom) ==========
	legendY := h - 18
	legendColor := color.RGBA{120, 130, 150, 255}

	switch arch {
	case sharedwidgets.Architecture2T1R:
		if mode == ModeCompute {
			drawSimpleText(img, "2T1R: All row+col transistors ON | Full MVM", 10, legendY, legendColor)
		} else {
			drawSimpleText(img, "2T1R: Row(green)+Col(cyan) = AND gate | Single cell selected", 10, legendY, legendColor)
		}
	case sharedwidgets.Architecture1T1R:
		if mode == ModeCompute {
			drawSimpleText(img, "All transistors ON | Full matrix active", 10, legendY, legendColor)
		} else {
			drawSimpleText(img, "Row transistor ON (green) | Others isolated", 10, legendY, legendColor)
		}
	default:
		if mode == ModeCompute {
			drawSimpleText(img, "Passive array | Sneak paths affect accuracy", 10, legendY, legendColor)
		} else {
			drawSimpleText(img, "Passive array | Red lines show sneak paths", 10, legendY, legendColor)
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

// updateModeHelp updates the mode description text with architecture-aware context
func (ca *CircuitsApp) updateModeHelp() {
	if ca.operationsModeHelp == nil {
		return
	}

	ca.mu.RLock()
	mode := ca.currentMode
	arch := ca.architecture
	ca.mu.RUnlock()

	is1T1R := arch == sharedwidgets.Architecture1T1R
	is2T1R := arch == sharedwidgets.Architecture2T1R

	var helpText string
	switch mode {
	case ModeWrite:
		if is2T1R {
			helpText = "WRITE: Row transistor (WL●) AND column transistor (CSL●) select ONLY target cell. Zero disturb to neighbors."
		} else if is1T1R {
			helpText = "WRITE: Transistor gates ONLY selected row (green●). Full write pulse to target cell, others isolated."
		} else {
			helpText = "WRITE: Passive array - partial voltages affect neighboring rows (sneak paths ~5-20% error)."
		}
	case ModeRead:
		if is2T1R {
			helpText = "READ: Dual transistor AND-gate selects ONLY target cell (WL●+CSL●). Perfect isolation, zero noise."
		} else if is1T1R {
			helpText = "READ: Transistor isolates selected row (green●). Clean sense current from target cell only."
		} else {
			helpText = "READ: Passive array - sneak currents add ~5-20% noise to sense signal."
		}
	case ModeCompute:
		if is2T1R {
			helpText = "COMPUTE: ALL CSL transistors ON (column gates●), WL selects active rows. Clean parallel MVM."
		} else if is1T1R {
			helpText = "COMPUTE: ALL transistors ON (all green●) for full MVM. Sneak-free parallel computation."
		} else {
			helpText = "COMPUTE: Passive MVM - sneak paths cause ~5-20% output error. Still functional for AI inference."
		}
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

// ============================================================================
// ARCHITECTURE TOGGLE (1T1R vs 0T1R)
// ============================================================================

// createArchitectureToggle creates the PASSIVE/1T1R/2T1R toggle buttons
// 1T1R: Transistor gates each row - only selected row active (write/read) or all rows (compute)
// 2T1R: Dual transistors (row + column) - individual cell addressing with AND-gate selection
// 0T1R: Passive crossbar - sneak paths affect accuracy
func (ca *CircuitsApp) createArchitectureToggle() fyne.CanvasObject {
	// Create toggle buttons (same pattern as Module 2)
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
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture0T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.refreshSharedArray()
		ca.updateModeHelp()
	}

	ca.arch1T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture1T1R {
			return // Already selected
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture1T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.refreshSharedArray()
		ca.updateModeHelp()
	}

	ca.arch2T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture2T1R {
			return // Already selected
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture2T1R
		ca.mu.Unlock()
		updateArchButtons()
		ca.refreshSharedArray()
		ca.updateModeHelp()
	}

	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)

	archLabel := widget.NewLabel("Array:")
	return container.NewHBox(archLabel, ca.archToggle)
}

