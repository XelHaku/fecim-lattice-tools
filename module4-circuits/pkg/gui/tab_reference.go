// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified REFERENCE tab combining TIMING and SPECS sections.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
)

// ============================================================================
// UNIFIED REFERENCE TAB: TIMING DIAGRAMS + SPECIFICATIONS
// ============================================================================

// createReferenceTab creates the unified REFERENCE view
func (ca *CircuitsApp) createReferenceTab() fyne.CanvasObject {
	// Create sections FIRST (before selector triggers callback)
	timingSection := ca.createReferenceTimingSection()
	specsSection := ca.createReferenceSpecsSection()

	// Assign to struct BEFORE SetSelected triggers callback
	ca.refTimingSection = timingSection
	ca.refSpecsSection = specsSection
	specsSection.Hide()

	// Section selector (callback now safe - sections are assigned)
	sectionSelect := widget.NewSelect([]string{"TIMING DIAGRAMS", "SPECIFICATIONS"}, func(s string) {
		ca.onReferenceSectionChanged(s)
	})
	sectionSelect.SetSelected("TIMING DIAGRAMS")

	contentStack := container.NewStack(timingSection, specsSection)

	header := container.NewHBox(
		widget.NewLabel("Reference:"),
		sectionSelect,
		layout.NewSpacer(),
	)

	return container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil, nil, nil,
		contentStack,
	)
}

func (ca *CircuitsApp) onReferenceSectionChanged(section string) {
	// Safety check - sections may not be initialized yet
	if ca.refTimingSection == nil || ca.refSpecsSection == nil {
		return
	}
	if section == "TIMING DIAGRAMS" {
		ca.refTimingSection.Show()
		ca.refSpecsSection.Hide()
	} else {
		ca.refTimingSection.Hide()
		ca.refSpecsSection.Show()
	}
}

// ============================================================================
// REFERENCE TIMING SECTION
// ============================================================================

func (ca *CircuitsApp) createReferenceTimingSection() fyne.CanvasObject {
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
	animateBtn := widget.NewButton("ANIMATE", ca.onAnimateTiming)
	exportBtn := widget.NewButton("EXPORT SVG", ca.onExportTimingSVG)

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
	bgColor := sharedtheme.ColorDarkBG
	signalColor := sharedtheme.ColorCyan
	labelColor := sharedtheme.ColorTextSecondary
	timeColor := sharedtheme.ColorOrange

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
		prevX := marginLeft

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
				// Draw horizontal line from prevX to x before transition
				drawThickHorizontalLine(img, prevX, x, lineY, 3, signalColor)
				// Draw vertical transition
				for py := y + 5; py < y+signalH-5; py++ {
					img.Set(x, py, signalColor)
				}
				prevX = x
			}
			prevHigh = isHigh
		}
		// Draw final horizontal segment
		finalX := marginLeft + plotW
		finalLineY := y + signalH - 5
		if prevHigh {
			finalLineY = y + 5
		}
		drawThickHorizontalLine(img, prevX, finalX, finalLineY, 3, signalColor)
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
	drawSimpleText(img, "70ns total", w-80, axisY+7, sharedtheme.ColorGreen)

	return img
}

func (ca *CircuitsApp) createTimingReadSection() fyne.CanvasObject {
	ca.timingReadCanvas = canvas.NewRaster(ca.drawTimingRead)
	ca.timingReadCanvas.SetMinSize(fyne.NewSize(600, 180))
	return ca.timingReadCanvas
}

func (ca *CircuitsApp) drawTimingRead(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := sharedtheme.ColorDarkBG
	signalColor := sharedtheme.ColorCyan
	labelColor := sharedtheme.ColorTextSecondary
	timeColor := sharedtheme.ColorOrange

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
		prevX := marginLeft

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

			// Draw horizontal line segment before transition
			if prevX < x {
				yToDraw := lineY
				if prevLineY != -1 {
					yToDraw = prevLineY
				}
				drawThickHorizontalLine(img, prevX, x, yToDraw, 3, signalColor)
			}

			// Draw vertical transition
			if prevLineY != -1 && lineY != prevLineY {
				minY := min(lineY, prevLineY)
				maxY := max(lineY, prevLineY)
				for py := minY; py <= maxY; py++ {
					img.Set(x, py, signalColor)
				}
				prevX = x
			}
			prevLineY = lineY
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
	drawSimpleText(img, "20ns total", w-80, axisY+7, sharedtheme.ColorGreen)

	return img
}

func (ca *CircuitsApp) createTimingComputeSection() fyne.CanvasObject {
	ca.timingComputeCanvas = canvas.NewRaster(ca.drawTimingCompute)
	ca.timingComputeCanvas.SetMinSize(fyne.NewSize(600, 200))
	return ca.timingComputeCanvas
}

func (ca *CircuitsApp) drawTimingCompute(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := sharedtheme.ColorDarkBG
	signalColor := sharedtheme.ColorCyan
	labelColor := sharedtheme.ColorTextSecondary
	phaseColor := sharedtheme.ColorPurple

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
		prevX := marginLeft

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

			// Draw horizontal line segment before transition
			if prevX < x {
				yToDraw := lineY
				if prevLineY != -1 {
					yToDraw = prevLineY
				}
				drawThickHorizontalLine(img, prevX, x, yToDraw, 3, signalColor)
			}

			// Draw vertical transition
			if prevLineY != -1 && lineY != prevLineY {
				minY := min(lineY, prevLineY)
				maxY := max(lineY, prevLineY)
				for py := minY; py <= maxY; py++ {
					img.Set(x, py, signalColor)
				}
				prevX = x
			}
			prevLineY = lineY
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
	drawSimpleText(img, "20ns total", w-80, axisY-2, sharedtheme.ColorGreen)

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

func (ca *CircuitsApp) onAnimateTiming() {
	// Animate signal phases step by step with status updates
	selectedOp := ca.timingOpSelect.Selected

	var steps []string
	switch selectedOp {
	case "WRITE":
		steps = []string{
			"Phase 1: CLK rising edge (0ns)...",
			"Phase 2: ROW_SEL and COL_SEL active (10ns)...",
			"Phase 3: DAC_EN enables voltage conversion (15ns)...",
			"Phase 4: V_PROG writes polarization state (20-70ns)...",
			"Phase 5: DONE signal asserted (85ns)...",
			"Write complete: Total 70ns",
		}
	case "READ":
		steps = []string{
			"Phase 1: CLK rising edge (0ns)...",
			"Phase 2: V_READ applied to cell (10ns)...",
			"Phase 3: I_SENSE measures current (15ns)...",
			"Phase 4: ADC_EN converts analog to digital (40-70ns)...",
			"Phase 5: DATA_OUT valid (75ns)...",
			"Read complete: Total 20ns",
		}
	case "COMPUTE":
		steps = []string{
			"Phase 1: INPUT_VALID asserted (5ns)...",
			"Phase 2: DAC_ALL converts inputs to voltages (10-35ns)...",
			"Phase 3: ARRAY_SETTLE - currents accumulate via Kirchhoff's law (35-60ns)...",
			"Phase 4: ADC_ALL digitizes summed currents (55-90ns)...",
			"Phase 5: OUTPUT_VALID - MVM result ready (90ns)...",
			"Compute complete: Total 20ns for full MVM",
		}
	default:
		steps = []string{"Select an operation to animate"}
	}

	ca.timingStatusLabel.SetText("Animating " + selectedOp + " timing...")

	go func() {
		for i, step := range steps {
			fyne.Do(func() {
				ca.timingStatusLabel.SetText(step)
			})
			if i < len(steps)-1 {
				// Pause between animation steps
				ca.sleep(600)
			}
		}
	}()
}

func (ca *CircuitsApp) onExportTimingSVG() {
	// Show "Export SVG not implemented - use screenshot" message in status
	fyne.Do(func() {
		ca.timingStatusLabel.SetText("Export SVG not implemented - use screenshot (Cmd+Shift+4 on macOS, PrtSc on Linux)")
	})
}

// ============================================================================
// REFERENCE SPECS SECTION
// ============================================================================

func (ca *CircuitsApp) createReferenceSpecsSection() fyne.CanvasObject {
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
	exportBtn := widget.NewButton("EXPORT SPECS", ca.onExportSpecs)
	compareBtn := widget.NewButton("COMPARE TO GPU", ca.onCompareToGPU)

	ca.specStatusLabel = widget.NewLabel("System specifications")

	buttonBox := container.NewHBox(
		exportBtn,
		compareBtn,
		layout.NewSpacer(),
		ca.specStatusLabel,
	)

	// Layout in a grid with improved visual hierarchy
	arrayHeader := widget.NewLabelWithStyle("ARRAY CONFIGURATION", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dacHeader := widget.NewLabelWithStyle("DAC SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	adcHeader := widget.NewLabelWithStyle("ADC SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tiaHeader := widget.NewLabelWithStyle("TIA SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fefetHeader := widget.NewLabelWithStyle("FeFET CELL SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	summaryHeader := widget.NewLabelWithStyle("SYSTEM SUMMARY", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	leftCol := container.NewVBox(
		arrayHeader,
		layout.NewSpacer(), // Small spacing after header
		arraySection,
		layout.NewSpacer(), // Spacing before separator
		widget.NewSeparator(),
		layout.NewSpacer(), // Spacing after separator
		dacHeader,
		layout.NewSpacer(),
		dacSection,
		layout.NewSpacer(),
		widget.NewSeparator(),
		layout.NewSpacer(),
		adcHeader,
		layout.NewSpacer(),
		adcSection,
	)

	rightCol := container.NewVBox(
		tiaHeader,
		layout.NewSpacer(),
		tiaSection,
		layout.NewSpacer(),
		widget.NewSeparator(),
		layout.NewSpacer(),
		fefetHeader,
		layout.NewSpacer(),
		fefetSection,
		layout.NewSpacer(),
		widget.NewSeparator(),
		layout.NewSpacer(),
		summaryHeader,
		layout.NewSpacer(),
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
	ca.specArraySizeSelect = widget.NewSelect(sizeOptions, func(s string) {
		// Update the summary when size changes
		ca.updateSpecSummary()
	})
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
		widget.NewLabel(""), // Spacing
		container.NewHBox(widget.NewLabel("Quantization:"), ca.specQuantLevelSelect, widget.NewLabel(fmt.Sprintf("levels (~%.1f bits/cell)", bitsPerCell))),
		widget.NewLabel(""), // Spacing
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
		widget.NewLabel(""), // Spacing
		widget.NewLabel(specs),
		widget.NewLabel(""), // Spacing
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
		widget.NewLabel(""), // Spacing
		widget.NewLabel(specs),
		widget.NewLabel(""), // Spacing
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
		widget.NewLabel(""), // Spacing
		widget.NewLabel(specs),
		widget.NewLabel(""), // Spacing
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

	return container.NewVBox(
		grid,
		widget.NewLabel(""), // Spacing
		widget.NewSeparator(),
		helpText,
	)
}

func (ca *CircuitsApp) createSpecSummarySection() fyne.CanvasObject {
	// Calculate initial summary based on default size (32x32)
	size := 32
	cells := size * size
	throughput := float64(cells) / 20.0 // MACs per ns = GOPS

	summary := fmt.Sprintf(`Component       | Count | Power   | Area     | Latency
----------------|-------|---------|----------|--------
FeFET Array     | %d | 0.1 mW  | 0.01 mm² | 5 ns
DACs            | %d    | 3.2 mW  | 0.02 mm² | 5 ns
TIAs            | %d    | 1.6 mW  | 0.01 mm² | 2 ns
ADCs            | %d    | 16 mW   | 0.04 mm² | 10 ns
Control         | 1     | 0.5 mW  | 0.01 mm² | 2 ns
----------------|-------|---------|----------|--------
TOTAL           |       | 21.4 mW | 0.09 mm² | 20 ns

Throughput:     %d MACs / 20ns = %.1f GOPS
Efficiency:     %.1f GOPS / 21.4 mW = %d GOPS/W`,
		cells, size, size, size,
		cells, throughput, throughput, int(throughput*1000/21.4))

	ca.specSummaryLabel = widget.NewLabel(summary)
	return ca.specSummaryLabel
}

func (ca *CircuitsApp) updateSpecSummary() {
	if ca.specSummaryLabel == nil || ca.specArraySizeSelect == nil {
		return
	}

	// Get current array size
	var size int
	fmt.Sscanf(ca.specArraySizeSelect.Selected, "%d", &size)
	if size == 0 {
		size = 32 // default
	}

	cells := size * size
	throughput := float64(cells) / 20.0 // MACs per ns = GOPS

	summary := fmt.Sprintf(`Component       | Count | Power   | Area     | Latency
----------------|-------|---------|----------|--------
FeFET Array     | %d | 0.1 mW  | 0.01 mm² | 5 ns
DACs            | %d    | 3.2 mW  | 0.02 mm² | 5 ns
TIAs            | %d    | 1.6 mW  | 0.01 mm² | 2 ns
ADCs            | %d    | 16 mW   | 0.04 mm² | 10 ns
Control         | 1     | 0.5 mW  | 0.01 mm² | 2 ns
----------------|-------|---------|----------|--------
TOTAL           |       | 21.4 mW | 0.09 mm² | 20 ns

Throughput:     %d MACs / 20ns = %.1f GOPS
Efficiency:     %.1f GOPS / 21.4 mW = %d GOPS/W`,
		cells, size, size, size,
		cells, throughput, throughput, int(throughput*1000/21.4))

	ca.specSummaryLabel.SetText(summary)
}

func (ca *CircuitsApp) onExportSpecs() {
	// Show "Export not implemented" message in status
	fyne.Do(func() {
		ca.specStatusLabel.SetText("Export not implemented - copy specs from display or take screenshot")
	})
}

func (ca *CircuitsApp) onCompareToGPU() {
	// Show comparison summary in status label
	fyne.Do(func() {
		ca.specStatusLabel.SetText("FeFET vs GPU: 25x faster (20ns vs 500ns), 2000x more efficient (2392 vs ~5 GOPS/W), 100x smaller area")
	})
}
