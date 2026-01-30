// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains array drawing functions for the unified view.
package gui

import (
	"fmt"
	"image"
	"image/color"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

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

	// Calculate margins - increased for larger peripheral boxes
	topMargin := 65    // Increased from 50 for larger DAC boxes + column labels
	rightMargin := 130 // Increased from 20 for larger TIA+ADC boxes
	bottomMargin := 30 // Slightly increased
	leftMargin := 30   // Slightly increased

	is1T1R := arch == sharedwidgets.Architecture1T1R
	is2T1R := arch == sharedwidgets.Architecture2T1R
	if is1T1R || is2T1R {
		leftMargin = 55
	}
	if is2T1R {
		bottomMargin = 55
	}

	availableW := w - leftMargin - rightMargin
	availableH := h - topMargin - bottomMargin

	cellW := availableW / cols
	cellH := availableH / rows
	cellSize := min(cellW, cellH)

	// Scale max/min cell size based on array dimensions
	maxCellSize := 70 // Default for small arrays
	minCellSize := 18 // Default minimum

	// For larger arrays, reduce cell size to fit
	if cols > 32 || rows > 32 {
		maxCellSize = 30
		minCellSize = 8
	} else if cols > 16 || rows > 16 {
		maxCellSize = 40
		minCellSize = 12
	}

	if cellSize > maxCellSize {
		cellSize = maxCellSize
	}
	if cellSize < minCellSize {
		cellSize = minCellSize
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
	writeThreshold := ca.deviceState.GetWriteRange().Min
	for c := 0; c < cols; c++ {
		x := offsetX + c*cellSize + cellSize/2
		voltage := ca.deviceState.GetDACVoltage(c)

		var blCol color.RGBA
		if voltage >= writeThreshold {
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

	// Draw signal line labels (BL = Bit Line, WL = Word Line, SL = Source Line)
	// BL label at top of grid
	drawSimpleText(img, "BL", offsetX+gridW/2-6, offsetY-35, color.RGBA{100, 180, 255, 200})
	// WL label at left of grid
	drawSimpleText(img, "WL", offsetX-25, offsetY+gridH/2-3, color.RGBA{255, 180, 100, 200})
	// SL label at bottom for 2T1R
	if is2T1R {
		drawSimpleText(img, "SL", offsetX+gridW/2-6, offsetY+gridH+45, color.RGBA{100, 220, 255, 200})
	}

	// Draw row indices on left side of array
	// For large arrays, only show every Nth index to avoid overlap
	rowLabelStep := 1
	if rows > 64 {
		rowLabelStep = 16
	} else if rows > 32 {
		rowLabelStep = 8
	} else if rows > 16 {
		rowLabelStep = 4
	} else if rows > 8 {
		rowLabelStep = 2
	}

	for r := 0; r < rows; r++ {
		// Only draw label at intervals or if selected
		if r%rowLabelStep != 0 && r != selectedRow {
			continue
		}
		y := offsetY + r*cellSize + cellSize/2 - 3
		indexColor := color.RGBA{150, 150, 170, 200}
		if r == selectedRow {
			indexColor = color.RGBA{255, 220, 100, 255} // Highlight selected row
		}
		rowText := fmt.Sprintf("%d", r)
		drawSimpleText(img, rowText, 5, y, indexColor)
	}

	// Draw column indices below array (above DAC boxes position)
	// For large arrays, only show every Nth index to avoid overlap
	colLabelStep := 1
	if cols > 64 {
		colLabelStep = 16
	} else if cols > 32 {
		colLabelStep = 8
	} else if cols > 16 {
		colLabelStep = 4
	} else if cols > 8 {
		colLabelStep = 2
	}

	for c := 0; c < cols; c++ {
		// Only draw label at intervals or if selected
		if c%colLabelStep != 0 && c != selectedCol {
			continue
		}
		x := offsetX + c*cellSize + cellSize/2 - 3
		indexColor := color.RGBA{150, 150, 170, 200}
		if c == selectedCol {
			indexColor = color.RGBA{255, 220, 100, 255} // Highlight selected column
		}
		colText := fmt.Sprintf("%d", c)
		drawSimpleText(img, colText, x, offsetY+gridH+5, indexColor)
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

			// Cell color based on level - always full brightness
			cellColor := levelToColor(level, levels)

			// Animation highlight (only during compute animation)
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

			// Draw state number and conductance in cell if large enough
			if cellSize >= 28 {
				// Calculate text color for contrast (light on dark, dark on light)
				brightness := (int(cellColor.R) + int(cellColor.G) + int(cellColor.B)) / 3
				var textColor color.RGBA
				if brightness > 140 {
					textColor = color.RGBA{0, 0, 0, 220} // Dark text on light bg
				} else {
					textColor = color.RGBA{255, 255, 255, 220} // Light text on dark bg
				}

				// For large cells (>= 45px), show both state and conductance
				if cellSize >= 45 {
					// Calculate conductance using material model
					var conductanceUS float64
					material := ca.deviceState.GetMaterial()
					if material != nil {
						conductanceUS = material.DiscreteLevel(level, levels) * 1e6 // S to uS
					} else {
						conductanceUS = 1.0 + float64(level)/float64(levels-1)*99.0
					}

					// Draw state level number (top half of cell)
					stateText := fmt.Sprintf("S%d", level)
					textX := x0 + cw/2 - len(stateText)*3
					textY := y0 + ch/3 - 3
					drawSimpleText(img, stateText, textX, textY, textColor)

					// Draw conductance value (bottom half of cell) - dimmer
					var gText string
					if conductanceUS < 10 {
						gText = fmt.Sprintf("%.1f", conductanceUS)
					} else {
						gText = fmt.Sprintf("%.0f", conductanceUS)
					}
					gTextX := x0 + cw/2 - len(gText)*3
					gTextY := y0 + ch*2/3 - 3
					dimTextColor := color.RGBA{textColor.R, textColor.G, textColor.B, 160}
					drawSimpleText(img, gText, gTextX, gTextY, dimTextColor)
				} else {
					// For medium cells, just show state number centered
					stateText := fmt.Sprintf("%d", level)
					textX := x0 + cw/2 - len(stateText)*3
					textY := y0 + ch/2 - 3
					drawSimpleText(img, stateText, textX, textY, textColor)
				}
			}

			// C1 FIX: Selected cell highlight with bright contrasting border
			if isSelected {
				// Bold yellow/gold border (3px thick)
				highlightColor := color.RGBA{255, 200, 0, 255}
				drawRectBorder(img, x0-1, y0-1, cw+2, ch+2, highlightColor)
				drawRectBorder(img, x0-2, y0-2, cw+4, ch+4, highlightColor)
				drawRectBorder(img, x0-3, y0-3, cw+6, ch+6, highlightColor)
				// Subtle white outer glow
				drawRectBorder(img, x0-4, y0-4, cw+8, ch+8, color.RGBA{255, 255, 255, 180})
			}
		}
	}

	// Draw DAC boxes (top) - scale based on array size
	dacBoxH := 30
	dacBoxW := cellSize - 2
	// Scale DAC box size based on array dimensions
	if cols > 32 {
		dacBoxH = 18
		dacBoxW = cellSize - 1
	} else if cols > 16 {
		dacBoxH = 22
	}
	if dacBoxW < 20 {
		dacBoxW = 20
	}
	dacY := offsetY - dacBoxH - 12

	// Show ALL DAC boxes (scaled to fit)
	for c := 0; c < cols; c++ {
		dacX := offsetX + c*cellSize + 1
		voltage := ca.deviceState.GetDACVoltage(c)
		highlighted := animStep == 1
		// Only show label for every Nth column to avoid clutter
		colLabel := ""
		if cols <= 16 || c%4 == 0 {
			colLabel = fmt.Sprintf("C%d", c)
		}
		drawDACColumn(img, dacX, dacY, dacBoxW, dacBoxH, voltage, colLabel, highlighted, false)
	}

	// Draw TIA+ADC boxes (right side) - scale based on array size
	tiaBoxW := 70
	adcBoxW := 30
	tiaAdcBoxH := cellSize - 2
	// Scale TIA/ADC box size based on array dimensions
	if rows > 32 {
		tiaBoxW = 45
		adcBoxW = 20
		tiaAdcBoxH = cellSize - 1
	} else if rows > 16 {
		tiaBoxW = 55
		adcBoxW = 25
	}
	if tiaAdcBoxH < 16 {
		tiaAdcBoxH = 16
	}
	tiaX := offsetX + gridW + 8

	// Show ALL TIA+ADC boxes (scaled to fit)
	for r := 0; r < rows; r++ {
		tiaY := offsetY + r*cellSize + 1
		current := ca.deviceState.GetRowCurrent(r)
		level := ca.deviceState.GetRowLevel(r)
		highlighted := animStep == 3
		dimmed := !ca.deviceState.IsRowActive(r)
		// Only show label for every Nth row to avoid clutter
		rowLabel := ""
		if rows <= 16 || r%4 == 0 {
			rowLabel = fmt.Sprintf("R%d", r)
		}
		drawTIAADCRow(img, tiaX, tiaY, tiaBoxW, adcBoxW, tiaAdcBoxH, current, level, rowLabel, highlighted, dimmed, ca.tia, ca.adc)
	}

	// Draw labels
	drawSimpleText(img, "DAC", offsetX-25, dacY+dacBoxH/2-3, color.RGBA{170, 140, 220, 255})
	drawSimpleText(img, "TIA", tiaX, offsetY-10, color.RGBA{220, 180, 100, 255})
	drawSimpleText(img, "ADC", tiaX+tiaBoxW+4, offsetY-10, color.RGBA{130, 210, 170, 255})

	// Draw voltage safety gauge (horizontal bar showing operating zone)
	// Position it to the right of the array, below TIA/ADC labels
	gaugeX := tiaX
	gaugeY := offsetY + gridH + 15
	gaugeW := 100
	gaugeH := 12

	// Get material's coercive voltage for zone boundaries
	writeRange := ca.deviceState.GetWriteRange()
	vcThreshold := writeRange.Min // Vc threshold
	maxV := 3.0                   // Hardware max

	// Draw gauge background zones
	readZoneW := int(float64(gaugeW) * (0.5 * vcThreshold / maxV))
	cautionZoneW := int(float64(gaugeW) * (vcThreshold / maxV)) - readZoneW
	writeZoneW := gaugeW - readZoneW - cautionZoneW

	// Read zone (safe) - blue
	drawRect(img, gaugeX, gaugeY, readZoneW, gaugeH, color.RGBA{60, 140, 200, 200})
	// Caution zone - yellow
	drawRect(img, gaugeX+readZoneW, gaugeY, cautionZoneW, gaugeH, color.RGBA{200, 180, 60, 200})
	// Write zone - orange/red
	drawRect(img, gaugeX+readZoneW+cautionZoneW, gaugeY, writeZoneW, gaugeH, color.RGBA{220, 100, 60, 200})

	// Draw current voltage indicator
	selectedVoltage := ca.deviceState.GetDACVoltage(ca.deviceState.GetSelectedCol())
	if selectedVoltage > 0 {
		indicatorX := gaugeX + int(float64(gaugeW)*(selectedVoltage/maxV))
		if indicatorX > gaugeX+gaugeW-2 {
			indicatorX = gaugeX + gaugeW - 2
		}
		// Draw triangle indicator
		for dy := 0; dy < 5; dy++ {
			for dx := -dy; dx <= dy; dx++ {
				px, py := indicatorX+dx, gaugeY-2-dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}

	// Gauge labels
	drawSimpleText(img, "V:", gaugeX-15, gaugeY+3, color.RGBA{150, 150, 170, 180})
	drawSimpleText(img, "0", gaugeX, gaugeY+gaugeH+2, color.RGBA{100, 150, 200, 150})
	drawSimpleText(img, "Vc", gaugeX+readZoneW+cautionZoneW/2-6, gaugeY+gaugeH+2, color.RGBA{255, 200, 100, 180})
	drawSimpleText(img, fmt.Sprintf("%.1f", maxV), gaugeX+gaugeW-12, gaugeY+gaugeH+2, color.RGBA{255, 100, 100, 150})

	// Operation classification title with prominent badge
	opText := ca.deviceState.ClassifyOperation()
	var opColor, opBgColor color.RGBA
	switch {
	case opText == "WRITE":
		opColor = color.RGBA{255, 200, 100, 255}
		opBgColor = color.RGBA{80, 60, 30, 200}
	case opText == "READ":
		opColor = color.RGBA{100, 220, 255, 255}
		opBgColor = color.RGBA{30, 60, 80, 200}
	case opText == "COMPUTE (MVM)":
		opColor = color.RGBA{200, 150, 255, 255}
		opBgColor = color.RGBA{50, 40, 80, 200}
	default:
		opColor = color.RGBA{150, 150, 150, 255}
		opBgColor = color.RGBA{40, 40, 50, 200}
	}
	// Draw background badge for operation mode
	opBoxW := len(opText)*6 + 12
	drawRoundedRect(img, 5, 3, opBoxW, 16, 4, opBgColor)
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

	// Draw energy/timing info in top-right corner (below architecture badge)
	mode := ca.deviceState.GetOperationMode()
	var energyText, timingText string
	var energyColor color.RGBA
	switch mode {
	case OpModeRead:
		energyText = "~45fJ"
		timingText = "65ns"
		energyColor = color.RGBA{100, 200, 255, 200}
	case OpModeWrite:
		energyText = "~55fJ"
		timingText = "170ns"
		energyColor = color.RGBA{255, 180, 100, 200}
	case OpModeCompute:
		// MVM energy scales with active cells
		activeCells := 0
		activeRows := 0
		activeCols := 0
		for r := 0; r < rows; r++ {
			if ca.deviceState.IsRowActive(r) {
				activeRows++
				for c := 0; c < cols; c++ {
					if ca.deviceState.GetDACVoltage(c) > 0.01 {
						activeCells++
					}
				}
			}
		}
		for c := 0; c < cols; c++ {
			if ca.deviceState.GetDACVoltage(c) > 0.01 {
				activeCols++
			}
		}
		energyFJ := 50 * activeCells // ~50fJ per cell
		energyText = fmt.Sprintf("~%dfJ", energyFJ)
		timingText = fmt.Sprintf("%dx%d=%d", activeRows, activeCols, activeCells)
		energyColor = color.RGBA{200, 150, 255, 200}
	default:
		energyText = ""
		timingText = ""
	}
	if energyText != "" {
		drawSimpleText(img, energyText, w-50, 22, energyColor)
		drawSimpleText(img, timingText, w-50, 34, color.RGBA{150, 150, 170, 180})
	}

	// Draw sneak path indicators for passive (0T1R) mode
	// Sneak currents flow through half-selected cells, causing read errors
	if arch == sharedwidgets.Architecture0T1R && ca.deviceState != nil {
		selectedRow := ca.deviceState.GetSelectedRow()
		selectedCol := ca.deviceState.GetSelectedCol()
		voltage := ca.deviceState.GetDACVoltage(selectedCol)

		// Only show sneak paths when there's active voltage
		if voltage > 0.05 {
			sneakPathColor := color.RGBA{255, 100, 100, 60} // Red tint for sneak path cells

			// Highlight half-selected cells (same row OR same column as target)
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					if r == selectedRow && c == selectedCol {
						continue // Skip the target cell itself
					}

					x0 := offsetX + c*cellSize + 2
					y0 := offsetY + r*cellSize + 2
					cw := cellSize - 4
					ch := cellSize - 4

					if r == selectedRow || c == selectedCol {
						// Half-selected cell (V/2 voltage) - orange tint
						for dy := 0; dy < ch; dy++ {
							for dx := 0; dx < cw; dx++ {
								px, py := x0+dx, y0+dy
								if px >= 0 && px < w && py >= 0 && py < h {
									// Blend with existing color
									existing := img.RGBAAt(px, py)
									blended := color.RGBA{
										uint8(min(int(existing.R)+30, 255)),
										uint8(min(int(existing.G)+15, 255)),
										existing.B,
										255,
									}
									img.Set(px, py, blended)
								}
							}
						}
						// Draw "V/2" label for larger cells
						if cellSize >= 30 {
							drawSimpleText(img, "V/2", x0+cw/2-9, y0+ch-8, color.RGBA{255, 200, 100, 200})
						}
					} else {
						// Sneak path cell (receives current via L-path) - subtle red tint
						for dy := 0; dy < ch; dy += 2 {
							for dx := 0; dx < cw; dx += 2 {
								px, py := x0+dx, y0+dy
								if px >= 0 && px < w && py >= 0 && py < h {
									img.Set(px, py, sneakPathColor)
								}
							}
						}
					}
				}
			}
			// Draw warning label with sneak path count
			sneakCount := (rows - 1) * (cols - 1)
			warnText := fmt.Sprintf("0T1R: %d sneak paths, %d half-select", sneakCount, rows+cols-2)
			drawSimpleText(img, warnText, 10, h-15, color.RGBA{255, 150, 100, 200})
		}
	}

	// Draw current flow indicators during active operation
	if animStep >= 2 {
		// Draw current flow arrows on active bit lines (columns with voltage)
		for c := 0; c < cols; c++ {
			voltage := ca.deviceState.GetDACVoltage(c)
			if voltage > 0.05 {
				x := offsetX + c*cellSize + cellSize/2
				// Draw downward current arrow (electrons flow opposite to current)
				arrowColor := color.RGBA{100, 255, 150, 200}
				// Draw arrow shaft
				for y := offsetY - 10; y < offsetY+gridH+5; y += 8 {
					if y >= 0 && y < h {
						img.Set(x, y, arrowColor)
						img.Set(x-1, y, arrowColor)
						img.Set(x+1, y, arrowColor)
					}
				}
			}
		}

		// Draw current collection arrows flowing to TIA (horizontal on active rows)
		for r := 0; r < rows; r++ {
			if ca.deviceState.IsRowActive(r) {
				current := ca.deviceState.GetRowCurrent(r)
				if current > 0.1 {
					y := offsetY + r*cellSize + cellSize/2
					// Arrow intensity based on current magnitude
					intensity := uint8(min(int(current*2), 200))
					arrowColor := color.RGBA{255, 200, intensity, 180}
					// Draw rightward arrow to TIA
					for x := offsetX + gridW + 2; x < tiaX-2; x += 4 {
						if x >= 0 && x < w {
							img.Set(x, y, arrowColor)
							img.Set(x, y-1, arrowColor)
							img.Set(x, y+1, arrowColor)
						}
					}
				}
			}
		}
	}

	// Draw info badge showing selected cell's expected current
	if ca.deviceState != nil {
		selectedRow := ca.deviceState.GetSelectedRow()
		selectedCol := ca.deviceState.GetSelectedCol()
		voltage := ca.deviceState.GetDACVoltage(selectedCol)

		ca.mu.RLock()
		var level int
		if selectedRow < len(weights) && selectedCol < len(weights[selectedRow]) {
			level = weights[selectedRow][selectedCol]
		}
		ca.mu.RUnlock()

		if voltage > 0.05 && level > 0 {
			// Calculate expected current for selected cell
			var conductanceUS float64
			material := ca.deviceState.GetMaterial()
			if material != nil {
				conductanceUS = material.DiscreteLevel(level, levels) * 1e6
			} else {
				conductanceUS = 1.0 + float64(level)/float64(levels-1)*99.0
			}
			expectedCurrent := conductanceUS * voltage

			// Draw info near selected cell
			cellX := offsetX + selectedCol*cellSize + cellSize/2
			cellY := offsetY + selectedRow*cellSize - 12
			if cellY > 20 {
				infoText := fmt.Sprintf("%.1fuA", expectedCurrent)
				drawSimpleText(img, infoText, cellX-len(infoText)*3, cellY, color.RGBA{255, 255, 100, 220})
			}
		}
	}

	// Draw color legend in top-left corner (below operation badge)
	legendX := 8
	legendY := 28 // Below operation badge
	legendW := 100
	legendH := 55

	// Draw semi-transparent background for legend
	legendBg := color.RGBA{15, 20, 35, 190} // Dark blue-gray, ~75% opacity
	for py := legendY - 5; py < legendY+legendH; py++ {
		for px := legendX - 3; px < legendX+legendW; px++ {
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, legendBg)
			}
		}
	}

	// Title
	drawSimpleText(img, "Legend:", legendX, legendY, color.RGBA{200, 200, 220, 255})

	// Cell conductance gradient: Low G -> High G
	legendY += 12
	drawSimpleText(img, "G:", legendX, legendY, color.RGBA{150, 150, 170, 200})
	// Draw gradient boxes
	boxW := 12
	for i := 0; i < 5; i++ {
		level := i * (levels - 1) / 4
		c := levelToColor(level, levels)
		drawRect(img, legendX+15+i*boxW, legendY-2, boxW-2, 10, c)
	}
	drawSimpleText(img, "Lo", legendX+15, legendY+10, color.RGBA{100, 150, 255, 200})
	drawSimpleText(img, "Hi", legendX+15+4*boxW, legendY+10, color.RGBA{255, 100, 100, 200})

	// DAC voltage zones
	legendY += 22
	drawSimpleText(img, "V:", legendX, legendY, color.RGBA{150, 150, 170, 200})
	// Read zone - blue
	drawRect(img, legendX+15, legendY-2, boxW, 10, color.RGBA{60, 140, 200, 255})
	drawSimpleText(img, "R", legendX+15+2, legendY+10, color.RGBA{100, 180, 255, 200})
	// Caution zone - yellow
	drawRect(img, legendX+15+boxW+2, legendY-2, boxW, 10, color.RGBA{200, 180, 60, 255})
	drawSimpleText(img, "!", legendX+15+boxW+4, legendY+10, color.RGBA{255, 220, 100, 200})
	// Write zone - red
	drawRect(img, legendX+15+2*(boxW+2), legendY-2, boxW, 10, color.RGBA{220, 100, 60, 255})
	drawSimpleText(img, "W", legendX+15+2*(boxW+2)+2, legendY+10, color.RGBA{255, 140, 100, 200})

	// Draw energy breakdown bar chart in bottom-center
	// C10: System Power Breakdown - Array ~45%, ADC/DAC ~40%, Peripherals ~15%
	energyBarX := w/2 - 60
	energyBarY := h - 65
	drawSimpleText(img, "System Power:", energyBarX-20, energyBarY, color.RGBA{180, 180, 200, 200})
	// Show percentage breakdown (C10 requirement)
	drawSimpleText(img, "Array~45% | ADC/DAC~40% | Periph~15%", energyBarX-20, energyBarY+12, color.RGBA{255, 191, 0, 180})
	// DAC: 15fJ (purple), TIA: 5fJ (orange), ADC: 25fJ (green), Pump: 10fJ (gray)
	barY := energyBarY + 24
	barH := 8
	// Scale: 1fJ = 1px, max ~55fJ total
	// DAC bar (purple)
	drawRect(img, energyBarX, barY, 15, barH, color.RGBA{130, 90, 190, 255})
	// TIA bar (orange)
	drawRect(img, energyBarX+15, barY, 5, barH, color.RGBA{190, 140, 70, 255})
	// ADC bar (green) - largest consumer
	drawRect(img, energyBarX+20, barY, 25, barH, color.RGBA{70, 170, 130, 255})
	// Pump bar (gray)
	drawRect(img, energyBarX+45, barY, 10, barH, color.RGBA{100, 100, 120, 255})
	// Labels
	drawSimpleText(img, "DAC", energyBarX, barY+12, color.RGBA{130, 90, 190, 180})
	drawSimpleText(img, "TIA", energyBarX+15, barY+12, color.RGBA{190, 140, 70, 180})
	drawSimpleText(img, "ADC", energyBarX+30, barY+12, color.RGBA{70, 170, 130, 180})
	drawSimpleText(img, "55fJ", energyBarX+55, barY, color.RGBA{150, 150, 170, 180})

	// Draw operation hint in bottom-right corner
	hintY := h - 20
	hintX := w - 200
	var hintText string
	hintColor := color.RGBA{120, 140, 160, 200}
	switch ca.deviceState.GetOperationMode() {
	case OpModeRead:
		hintText = "READ: Sense cell conductance"
		hintColor = color.RGBA{100, 180, 220, 200}
	case OpModeWrite:
		hintText = "WRITE: Program cell state"
		hintColor = color.RGBA{220, 160, 80, 200}
	case OpModeCompute:
		hintText = "MVM: y = W * x"
		hintColor = color.RGBA{180, 140, 220, 200}
	}
	if hintText != "" {
		drawSimpleText(img, hintText, hintX, hintY, hintColor)
	}

	return img
}

// drawRowTransistors draws the row transistors for 1T1R/2T1R architecture
// Enhanced with clearer MOSFET symbols and ON/OFF indicators
func (ca *CircuitsApp) drawRowTransistors(img *image.RGBA, offsetX, offsetY, cellSize, rows, gridH, w, h int) {
	for r := 0; r < rows; r++ {
		ty := offsetY + r*cellSize + cellSize/2
		tx := offsetX - 35 // Moved left for larger symbol

		transistorOn := ca.deviceState.IsRowActive(r)

		var bodyCol, gateCol, channelCol, terminalCol color.RGBA
		if transistorOn {
			bodyCol = color.RGBA{60, 200, 80, 255}    // Green body when ON
			gateCol = color.RGBA{100, 255, 120, 255}  // Bright green gate
			channelCol = color.RGBA{80, 220, 100, 255}
			terminalCol = color.RGBA{150, 255, 150, 255}
		} else {
			bodyCol = color.RGBA{60, 60, 70, 255}
			gateCol = color.RGBA{90, 90, 100, 255}
			channelCol = color.RGBA{50, 50, 60, 255}
			terminalCol = color.RGBA{100, 100, 110, 255}
		}

		// Draw MOSFET body (larger, 8x12 rectangle)
		for dy := -6; dy <= 6; dy++ {
			for dx := 0; dx < 5; dx++ {
				px, py := tx+dx, ty+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, bodyCol)
				}
			}
		}

		// Draw gate (thicker, 2px wide)
		gateX := tx - 4
		for dy := -8; dy <= 8; dy++ {
			py := ty + dy
			if gateX >= 0 && gateX+1 < w && py >= 0 && py < h {
				img.Set(gateX, py, gateCol)
				img.Set(gateX+1, py, gateCol)
			}
		}

		// Draw source terminal (top)
		for dx := -2; dx <= 2; dx++ {
			py := ty - 8
			px := tx + 2 + dx
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px, py-1, terminalCol)
			}
		}

		// Draw drain terminal (bottom)
		for dx := -2; dx <= 2; dx++ {
			py := ty + 8
			px := tx + 2 + dx
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px, py+1, terminalCol)
			}
		}

		// Draw channel (connecting to array)
		for dx := 5; dx < 25; dx++ {
			px := tx + dx
			if px >= 0 && px < w {
				img.Set(px, ty, channelCol)
				if transistorOn {
					img.Set(px, ty+1, channelCol) // Thicker when ON
				}
			}
		}

		// ON/OFF indicator with label
		if transistorOn {
			drawGlowCircle(img, tx+2, ty, 3, color.RGBA{150, 255, 150, 255}, color.RGBA{100, 200, 100, 100})
		} else {
			// Draw X for OFF state
			for d := -2; d <= 2; d++ {
				px1, py1 := tx+2+d, ty+d
				px2, py2 := tx+2+d, ty-d
				if px1 >= 0 && px1 < w && py1 >= 0 && py1 < h {
					img.Set(px1, py1, color.RGBA{150, 80, 80, 200})
				}
				if px2 >= 0 && px2 < w && py2 >= 0 && py2 < h {
					img.Set(px2, py2, color.RGBA{150, 80, 80, 200})
				}
			}
		}
	}
}

// drawColTransistors draws the column transistors for 2T1R architecture
// Enhanced with clearer MOSFET symbols (horizontal orientation)
func (ca *CircuitsApp) drawColTransistors(img *image.RGBA, offsetX, offsetY, cellSize, cols, gridW, gridH, w, h int) {
	for c := 0; c < cols; c++ {
		tx := offsetX + c*cellSize + cellSize/2
		ty := offsetY + gridH + 25 // Moved down slightly

		// In 2T1R, column transistors are controlled by CSL
		// For simplicity, all column transistors are ON when computing
		transistorOn := ca.deviceState.GetWLMode() == WLAll || c == ca.deviceState.GetSelectedCol()

		var bodyCol, gateCol, channelCol, terminalCol color.RGBA
		if transistorOn {
			bodyCol = color.RGBA{60, 180, 200, 255}    // Cyan body when ON
			gateCol = color.RGBA{100, 220, 255, 255}   // Bright cyan gate
			channelCol = color.RGBA{80, 200, 220, 255}
			terminalCol = color.RGBA{150, 230, 255, 255}
		} else {
			bodyCol = color.RGBA{60, 60, 70, 255}
			gateCol = color.RGBA{90, 90, 100, 255}
			channelCol = color.RGBA{50, 50, 60, 255}
			terminalCol = color.RGBA{100, 100, 110, 255}
		}

		// Draw MOSFET body (horizontal, larger)
		for dx := -6; dx <= 6; dx++ {
			for dy := 0; dy < 5; dy++ {
				px, py := tx+dx, ty+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, bodyCol)
				}
			}
		}

		// Draw gate (thicker)
		gateY := ty + 7
		for dx := -8; dx <= 8; dx++ {
			px := tx + dx
			if px >= 0 && px < w && gateY >= 0 && gateY+1 < h {
				img.Set(px, gateY, gateCol)
				img.Set(px, gateY+1, gateCol)
			}
		}

		// Draw left terminal
		for dy := -2; dy <= 2; dy++ {
			px := tx - 8
			py := ty + 2 + dy
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px-1, py, terminalCol)
			}
		}

		// Draw right terminal
		for dy := -2; dy <= 2; dy++ {
			px := tx + 8
			py := ty + 2 + dy
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, terminalCol)
				img.Set(px+1, py, terminalCol)
			}
		}

		// Draw channel (connecting to array above)
		for dy := -20; dy < 0; dy++ {
			py := ty + dy
			if tx >= 0 && tx < w && py >= 0 && py < h {
				img.Set(tx, py, channelCol)
				if transistorOn {
					img.Set(tx+1, py, channelCol) // Thicker when ON
				}
			}
		}

		// ON/OFF indicator
		if transistorOn {
			drawGlowCircle(img, tx, ty+2, 3, color.RGBA{150, 230, 255, 255}, color.RGBA{100, 180, 200, 100})
		} else {
			// Draw X for OFF state
			for d := -2; d <= 2; d++ {
				px1, py1 := tx+d, ty+2+d
				px2, py2 := tx+d, ty+2-d
				if px1 >= 0 && px1 < w && py1 >= 0 && py1 < h {
					img.Set(px1, py1, color.RGBA{150, 80, 80, 200})
				}
				if px2 >= 0 && px2 < w && py2 >= 0 && py2 < h {
					img.Set(px2, py2, color.RGBA{150, 80, 80, 200})
				}
			}
		}
	}
}
