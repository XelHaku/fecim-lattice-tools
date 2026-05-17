//go:build legacy_fyne

package gui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedtheme "fecim-lattice-tools/shared/theme"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ============================================================================
// TAB 4: COMPARISON (FeFET vs GPU vs CPU)
// ============================================================================

type compMetrics struct {
	TimeNS   float64
	EnergyPJ float64
	TOPSW    float64
}

const (
	compBaselineArraySize = 8.0
	compBaselineMACs      = compBaselineArraySize * compBaselineArraySize
)

func compMetricsForArray(size int, baseline compMetrics) compMetrics {
	if size <= 0 {
		size = int(compBaselineArraySize)
	}
	scale := float64(size*size) / compBaselineMACs
	timeNS := baseline.TimeNS * scale
	energyPJ := baseline.EnergyPJ * scale
	opsPerMVM := 2.0 * float64(size*size)
	opsPerSecond := opsPerMVM / (timeNS * 1e-9)
	powerW := (energyPJ * 1e-12) / (timeNS * 1e-9)
	topsw := 0.0
	if powerW > 0 {
		topsw = (opsPerSecond / 1e12) / powerW
	}
	return compMetrics{TimeNS: timeNS, EnergyPJ: energyPJ, TOPSW: topsw}
}

func (ca *CircuitsApp) createComparisonTab() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**COMPARISON**: Compare FeFET crossbar architecture against traditional von Neumann systems (CPU/GPU). FeFET performs computation in-memory using analog physics (Ohm's law), avoiding the memory bottleneck that limits conventional digital systems. **Context**: numbers here emphasize single-operation latency/energy; GPUs recover throughput with large batched workloads at higher total power.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Architecture comparison
	archSection := ca.createCompArchSection()

	// Timing comparison
	timingSection := ca.createCompTimingSection()

	// Energy comparison
	energySection := ca.createCompEnergySection()

	// Live comparison table
	tableSection := ca.createCompTableSection()

	// Buttons
	runBtn := widget.NewButton("Run Comparison", ca.onRunComparison)
	runBtn.Importance = widget.HighImportance

	animateBtn := widget.NewButton("Animate", ca.onAnimateComparison)
	scaleBtn := widget.NewButton("Scale Up", ca.onScaleUpComparison)

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

	comparisonBody := container.NewVBox(topRow, widget.NewSeparator(), bottomRow)
	comparisonBodyScroll := container.NewVScroll(container.NewHScroll(comparisonBody))

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		comparisonBodyScroll,
	)
}

func (ca *CircuitsApp) createCompArchSection() fyne.CanvasObject {
	ca.compArchCanvas = canvas.NewRaster(ca.drawCompArch)
	ca.compArchCanvas.SetMinSize(fyne.NewSize(400, 200))
	return ca.compArchCanvas
}

func (ca *CircuitsApp) drawCompArch(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := sharedtheme.ColorBackground
	labelColor := sharedtheme.ColorText
	arrowColor := sharedtheme.ColorWarning

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	sectionH := h / 3
	boxW := 80
	boxH := sectionH - 25

	// Row 1: CPU + DRAM section
	cpuX, cpuY := 30, 12
	dramX := cpuX + boxW + 70

	drawRect(img, cpuX, cpuY, boxW, boxH, sharedtheme.ColorError)
	drawSimpleText(img, "CPU", cpuX+25, cpuY+boxH/2-3, labelColor)

	drawRect(img, dramX, cpuY, boxW, boxH, sharedtheme.ColorError)
	drawSimpleText(img, "DRAM", dramX+20, cpuY+boxH/2-3, labelColor)

	// Arrow between CPU and DRAM
	arrowY := cpuY + boxH/2
	for x := cpuX + boxW + 5; x < dramX-5; x++ {
		img.Set(x, arrowY, arrowColor)
		img.Set(x, arrowY-1, arrowColor)
	}
	// Arrowhead
	for i := 0; i < 6; i++ {
		img.Set(dramX-5-i, arrowY-i, arrowColor)
		img.Set(dramX-5-i, arrowY+i, arrowColor)
	}
	drawSimpleText(img, "Data Bus", cpuX+boxW+15, arrowY-12, arrowColor)

	// Row 2: GPU + HBM section
	gpuY := sectionH + 8
	drawRect(img, cpuX, gpuY, boxW, boxH, sharedtheme.ColorSuccess)
	drawSimpleText(img, "GPU", cpuX+25, gpuY+boxH/2-3, labelColor)

	drawRect(img, dramX, gpuY, boxW, boxH, sharedtheme.ColorSuccess)
	drawSimpleText(img, "HBM", dramX+25, gpuY+boxH/2-3, labelColor)

	// Arrow between GPU and HBM
	arrowY = gpuY + boxH/2
	for x := cpuX + boxW + 5; x < dramX-5; x++ {
		img.Set(x, arrowY, arrowColor)
		img.Set(x, arrowY-1, arrowColor)
	}
	for i := 0; i < 6; i++ {
		img.Set(dramX-5-i, arrowY-i, arrowColor)
		img.Set(dramX-5-i, arrowY+i, arrowColor)
	}
	drawSimpleText(img, "Data Bus", cpuX+boxW+15, arrowY-12, arrowColor)

	// Row 3: FeFET CIM section (unified)
	fefetY := 2*sectionH + 5
	fefetW := dramX + boxW - cpuX
	drawRect(img, cpuX, fefetY, fefetW, boxH, sharedtheme.ColorPrimary)
	drawSimpleText(img, "FeFET CIM", cpuX+fefetW/2-35, fefetY+boxH/2-10, labelColor)
	drawSimpleText(img, "No Data Movement", cpuX+fefetW/2-55, fefetY+boxH/2+5, sharedtheme.ColorAccent)

	// Right side labels (clamped to avoid clipping on narrow widths)
	rightX := max(dramX+boxW+8, w-90)
	drawSimpleText(img, "Von Neumann", rightX, cpuY+boxH/2-3, sharedtheme.ColorError)
	drawSimpleText(img, "Near Memory", rightX, gpuY+boxH/2-3, sharedtheme.ColorSuccess)
	drawSimpleText(img, "In Memory", rightX, fefetY+boxH/2-3, sharedtheme.ColorPrimary)

	return img
}

func (ca *CircuitsApp) createCompTimingSection() fyne.CanvasObject {
	ca.compTimingCanvas = canvas.NewRaster(ca.drawCompTiming)
	ca.compTimingCanvas.SetMinSize(fyne.NewSize(400, 150))
	return ca.compTimingCanvas
}

func (ca *CircuitsApp) drawCompTiming(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := sharedtheme.ColorBackground
	axisColor := sharedtheme.ColorAxis
	labelColor := sharedtheme.ColorText
	valueColor := sharedtheme.ColorWarning

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 60
	marginRight := 80
	barH := 25
	spacing := 35
	if h < 150 {
		barH = max(16, h/8)
		spacing = max(barH+6, h/4)
	}
	maxBarW := max(40, w-marginLeft-marginRight)

	// CPU bar (500ns - full width)
	cpuY := 15
	cpuW := maxBarW
	drawSimpleText(img, "CPU", 10, cpuY+8, sharedtheme.ColorError)
	drawRect(img, marginLeft, cpuY, cpuW, barH, sharedtheme.ColorError)
	drawSimpleText(img, "500ns", marginLeft+cpuW+5, cpuY+8, valueColor)

	// GPU bar (50ns - 10% width)
	gpuY := cpuY + spacing
	gpuW := maxBarW * 50 / 500
	if gpuW < 30 {
		gpuW = 30
	}
	drawSimpleText(img, "GPU", 10, gpuY+8, sharedtheme.ColorSuccess)
	drawRect(img, marginLeft, gpuY, gpuW, barH, sharedtheme.ColorSuccess)
	drawSimpleText(img, "50ns", marginLeft+gpuW+5, gpuY+8, valueColor)

	// FeFET bar (76ns - ~15% width)
	fefetY := gpuY + spacing
	fefetW := maxBarW * 76 / 500
	if fefetW < 20 {
		fefetW = 20
	}
	drawSimpleText(img, "FeFET", 5, fefetY+8, sharedtheme.ColorPrimary)
	drawRect(img, marginLeft, fefetY, fefetW, barH, sharedtheme.ColorPrimary)
	drawSimpleText(img, "76ns", marginLeft+fefetW+5, fefetY+8, valueColor)

	// Speedup annotation (clamped to avoid clipping on narrow widths)
	speedupX := max(marginLeft, w-110)
	drawSimpleText(img, "≈6.6x vs CPU", speedupX, fefetY+8, sharedtheme.ColorAccent)

	// X-axis
	axisY := h - 25
	for x := marginLeft; x < w-marginRight; x++ {
		img.Set(x, axisY, axisColor)
	}

	// Axis label
	drawSimpleText(img, "Time (ns)", w/2-30, axisY+10, labelColor)

	// Scale markers
	scaleMarkers := []struct {
		pct   int
		label string
	}{
		{0, "0"},
		{50, "250"},
		{100, "500"},
	}

	for _, sm := range scaleMarkers {
		x := marginLeft + sm.pct*maxBarW/100
		for dy := 0; dy < 5; dy++ {
			img.Set(x, axisY+dy, axisColor)
		}
		drawSimpleText(img, sm.label, x-len(sm.label)*3, axisY+10, axisColor)
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
	bgColor := sharedtheme.ColorBackground
	axisColor := sharedtheme.ColorAxis
	labelColor := sharedtheme.ColorText
	valueColor := sharedtheme.ColorWarning

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	marginLeft := 60
	marginRight := 100
	barH := 28
	spacing := 45
	if h < 200 {
		barH = max(16, h/9)
		spacing = max(barH+8, h/4)
	}
	maxBarW := max(40, w-marginLeft-marginRight)

	// CPU bar (64,000 pJ - full width)
	cpuY := 20
	cpuW := maxBarW
	drawSimpleText(img, "CPU", 10, cpuY+8, sharedtheme.ColorError)
	drawRect(img, marginLeft, cpuY, cpuW, barH, sharedtheme.ColorError)
	drawSimpleText(img, "64000 pJ", marginLeft+cpuW+5, cpuY+8, valueColor)

	// GPU bar (6,400 pJ - 10% width)
	gpuY := cpuY + spacing
	gpuW := maxBarW * 6400 / 64000
	if gpuW < 30 {
		gpuW = 30
	}
	drawSimpleText(img, "GPU", 10, gpuY+8, sharedtheme.ColorSuccess)
	drawRect(img, marginLeft, gpuY, gpuW, barH, sharedtheme.ColorSuccess)
	drawSimpleText(img, "6400 pJ", marginLeft+gpuW+5, gpuY+8, valueColor)

	// FeFET bar (~2.9 pJ - tiny, need minimum visible)
	fefetY := gpuY + spacing
	fefetW := maxBarW * 29 / 64000 // 2.9 pJ scaled
	if fefetW < 8 {
		fefetW = 8 // Minimum visible
	}
	drawSimpleText(img, "FeFET", 5, fefetY+8, sharedtheme.ColorPrimary)
	drawRect(img, marginLeft, fefetY, fefetW, barH, sharedtheme.ColorPrimary)
	drawSimpleText(img, "2.9 pJ", marginLeft+fefetW+5, fefetY+8, valueColor)

	// Energy annotation — removed "10-100x savings" claim (flagged as unverified in honesty-audit.md)
	savingsX := max(marginLeft, w-120)
	drawSimpleText(img, "model estimate", savingsX, fefetY+8, sharedtheme.ColorAccent)

	// FeFET energy breakdown (below the bar)
	breakdownY := fefetY + barH + 12
	drawSimpleText(img, "Breakdown: DAC (35%) + Array (45%) + TIA (10%) + ADC (10%)", 10, breakdownY, color.RGBA{180, 180, 200, 255})

	// X-axis
	axisY := h - 30
	for x := marginLeft; x < w-marginRight; x++ {
		img.Set(x, axisY, axisColor)
	}

	// Axis label
	drawSimpleText(img, "Energy per 8x8 MVM", w/2-60, axisY+12, labelColor)

	// Scale note (log scale would be better but linear for illustration)
	drawSimpleText(img, "[Linear scale - FeFET bar scaled up for visibility]", 10, h-12, sharedtheme.ColorTextDim)

	return img
}

func (ca *CircuitsApp) createCompTableSection() fyne.CanvasObject {
	// Create table labels
	ca.compTableLabels = make([]*widget.Label, 16)
	cpu, gpu, fefet := computeComparisonMetrics(8)

	headers := []string{"", "Time", "Energy", "GOPS"}
	cpuRow := []string{cpu.Label, metricLatency(cpu.LatencyNS), metricEnergy(cpu.EnergyPJ), metricGOPS(cpu.GOPS)}
	gpuRow := []string{gpu.Label, metricLatency(gpu.LatencyNS), metricEnergy(gpu.EnergyPJ), metricGOPS(gpu.GOPS)}
	fefetRow := []string{fefet.Label, metricLatency(fefet.LatencyNS), metricEnergy(fefet.EnergyPJ), metricGOPS(fefet.GOPS)}

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
	contextLabel := widget.NewLabel("GPU row shown as sequential-equivalent work for one MVM. FeCIM completes in 1 cycle. Note: GPUs compensate with massive parallelism (batching) at higher power.")
	contextLabel.Wrapping = fyne.TextWrapWord

	tableContent := container.NewVBox(
		arraySizeLabel,
		contextLabel,
		widget.NewSeparator(),
		grid,
	)

	return container.NewVScroll(container.NewHScroll(tableContent))
}

func (ca *CircuitsApp) onRunComparison() {
	logAction("comparison_run")
	ca.compStatusLabel.SetText("Running comparison for 8×8 MVM...")

	// Refresh canvases
	sharedwidgets.SafeDo(func() {
		if ca.compArchCanvas != nil {
			ca.compArchCanvas.Refresh()
		}
		if ca.compTimingCanvas != nil {
			ca.compTimingCanvas.Refresh()
		}
		if ca.compEnergyCanvas != nil {
			ca.compEnergyCanvas.Refresh()
		}
	})

	ca.compStatusLabel.SetText("Comparison complete: FeFET advantage shown for per-operation latency/energy; GPUs can improve throughput with batching at higher power.")
}

func (ca *CircuitsApp) onAnimateComparison() {
	logAction("comparison_animate")
	// Animate the comparison showing CPU vs GPU vs FeFET timing step by step
	ca.compStatusLabel.SetText("Animating comparison...")

	steps := []string{
		"Step 1: CPU loads data from DRAM (250ns)...",
		"Step 2: CPU computes MVM (250ns)...",
		"Step 3: GPU loads data from HBM (25ns)...",
		"Step 4: GPU computes MVM (25ns)...",
		"Step 5: FeFET performs in-memory compute (76ns)...",
		"Animation complete: FeFET ≈6.6x faster than CPU (latency model)",
	}

	ca.launchBackground(func() {
		for i, step := range steps {
			if ca.shouldStop() {
				return
			}
			sharedwidgets.SafeDo(func() {
				ca.compStatusLabel.SetText(step)
			})
			if i < len(steps)-1 {
				// Pause between animation steps
				if ca.sleep(500) {
					return // Interrupted
				}
			}
		}
	})
}

func (ca *CircuitsApp) onScaleUpComparison() {
	logAction("comparison_scale_up")
	// Cycle through array sizes and update comparison values
	sizes := []int{8, 16, 32, 64}

	// Find next size in cycle
	currentSize := ca.compArraySize
	for i, size := range sizes {
		if size == currentSize {
			currentSize = sizes[(i+1)%len(sizes)]
			break
		}
	}
	ca.compArraySize = currentSize

	cpu, gpu, fefet := computeComparisonMetrics(currentSize)

	// Update table labels
	sharedwidgets.SafeDo(func() {
		ca.compTableLabels[5].SetText(metricLatency(cpu.LatencyNS))
		ca.compTableLabels[6].SetText(metricEnergy(cpu.EnergyPJ))
		ca.compTableLabels[7].SetText(metricGOPS(cpu.GOPS))

		ca.compTableLabels[9].SetText(metricLatency(gpu.LatencyNS))
		ca.compTableLabels[10].SetText(metricEnergy(gpu.EnergyPJ))
		ca.compTableLabels[11].SetText(metricGOPS(gpu.GOPS))

		ca.compTableLabels[13].SetText(metricLatency(fefet.LatencyNS))
		ca.compTableLabels[14].SetText(metricEnergy(fefet.EnergyPJ))
		ca.compTableLabels[15].SetText(metricGOPS(fefet.GOPS))

		ca.compStatusLabel.SetText(fmt.Sprintf("Scaled to %d×%d (%d MACs (Ops)): FeFET %.3f GOPS, %.3f pJ/op",
			currentSize, currentSize, currentSize*currentSize, fefet.GOPS, fefet.EnergyOpPJ))
	})
}
