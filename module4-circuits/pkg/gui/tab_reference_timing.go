// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the timing diagram section of the REFERENCE tab.
package gui

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
)

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
