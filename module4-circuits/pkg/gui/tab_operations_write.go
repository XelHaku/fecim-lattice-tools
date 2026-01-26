// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the WRITE mode panel and related functions.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
)

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

	// DAC tooltip
	dacTooltipBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Digital-to-Analog Converter (DAC)",
			"Converts digital level values to analog voltage pulses.\n\n"+
				"For WRITE: 1.2-1.5V to program FeFET polarization.\n"+
				"For COMPUTE: 0.3-0.5V for MVM operations.", ca.window)
	})
	dacTooltipBtn.Importance = widget.LowImportance

	dataPath := container.NewHBox(
		digitalBox, widget.NewLabel("->"),
		dacBox, dacTooltipBtn, widget.NewLabel("->"),
		fefetBox,
	)

	// Pulse visualization
	ca.opsWritePulseCanvas = canvas.NewRaster(ca.drawOpsWritePulse)
	ca.opsWritePulseCanvas.SetMinSize(fyne.NewSize(350, 120))

	// Ec tooltip for write section
	ecWriteTooltipBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Coercive Field (Ec)",
			"The electric field required to switch ferroelectric polarization.\n\n"+
				"For HfO₂-ZrO₂: Ec ≈ 1.0-1.5 MV/cm.\n\n"+
				"Write voltage must exceed Ec to reprogram the cell.", ca.window)
	})
	ecWriteTooltipBtn.Importance = widget.LowImportance

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
		container.NewHBox(
			widget.NewLabel("Write voltage must exceed Ec (~1.5 MV/cm)"),
			ecWriteTooltipBtn,
		),
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

	// Axis labels with units
	drawSimpleText(img, "Voltage (V)", 5, 5, axisColor)     // Y-axis label at top
	drawSimpleText(img, "Time (ns)", w-60, h-10, axisColor) // X-axis label at bottom

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
// WRITE MODE ACTION HANDLERS
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
