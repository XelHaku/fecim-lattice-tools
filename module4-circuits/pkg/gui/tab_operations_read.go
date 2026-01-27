// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the READ mode panel and related functions.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedtheme "fecim-lattice-tools/shared/theme"
)

// ============================================================================
// READ MODE PANEL
// ============================================================================

// createReadModePanel creates the read mode configuration panel
func (ca *CircuitsApp) createReadModePanel() {
	// Read voltage slider
	ca.opsReadVoltageLabel = widget.NewLabel(fmt.Sprintf("Read Voltage: %.2f V", ca.readVoltage))
	ca.opsReadVoltageSlider = widget.NewSlider(0.1, 0.5) // Max 0.5V for non-disturbing read
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

	// Data path visualization with tooltips
	fefetBox := ca.createLabeledBox("FeFET", "Cell", sharedtheme.ColorInfo)
	tiaBox := ca.createLabeledBox("TIA", "I->V", sharedtheme.ColorWarning)
	adcBox := ca.createLabeledBox("ADC", "8-bit", sharedtheme.ColorSuccess)
	digitalBox := ca.createLabeledBox("DIGITAL", "Level", sharedtheme.ColorPrimary)

	// TIA tooltip
	tiaTooltipBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Transimpedance Amplifier (TIA)",
			"Converts output current from FeFET to voltage for ADC.\n\n"+
				"Gain: R_feedback (1kΩ - 100kΩ typical).\n\n"+
				"Critical for sensing low currents in READ operations.", ca.window)
	})
	tiaTooltipBtn.Importance = widget.LowImportance

	// ADC tooltip
	adcTooltipBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Analog-to-Digital Converter (ADC)",
			"Converts TIA voltage back to digital level values.\n\n"+
				"Resolution: 4-12 bits typical.\n\n"+
				"For READ: Recovers stored level from sensed current.", ca.window)
	})
	adcTooltipBtn.Importance = widget.LowImportance

	dataPath := container.NewHBox(
		fefetBox, widget.NewLabel("->"),
		tiaBox, tiaTooltipBtn, widget.NewLabel("->"),
		adcBox, adcTooltipBtn, widget.NewLabel("->"),
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
		widget.NewLabel("SAFE: ≤0.5V (empirical threshold) | CAUTION: >0.5V may disturb polarization"),
		widget.NewLabel("Note: 0.5V threshold is simulation default; actual values are device-specific"),
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
// READ MODE ACTION HANDLERS
// ============================================================================

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
