// Package gui provides custom widgets for crossbar visualization.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ColorLegend displays a vertical color legend for heatmaps.
type ColorLegend struct {
	widget.BaseWidget

	minLabel  string
	maxLabel  string
	colormap  string
	unit      string
	levels    int
	showTicks bool

	raster *canvas.Raster
}

// NewColorLegend creates a new color legend widget.
func NewColorLegend(minLabel, maxLabel, unit string, levels int) *ColorLegend {
	l := &ColorLegend{
		minLabel:  minLabel,
		maxLabel:  maxLabel,
		unit:      unit,
		levels:    levels,
		colormap:  "viridis",
		showTicks: levels == 30, // Show tick marks for 30-level FeCIM
	}
	l.ExtendBaseWidget(l)
	return l
}

// SetColormap changes the colormap.
func (l *ColorLegend) SetColormap(name string) {
	l.colormap = name
	l.Refresh()
}

// SetLabels updates the min/max labels.
func (l *ColorLegend) SetLabels(minLabel, maxLabel string) {
	l.minLabel = minLabel
	l.maxLabel = maxLabel
	l.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (l *ColorLegend) CreateRenderer() fyne.WidgetRenderer {
	l.raster = canvas.NewRaster(l.generateImage)

	minText := canvas.NewText(l.minLabel, color.White)
	minText.TextSize = 10
	minText.Alignment = fyne.TextAlignLeading

	maxText := canvas.NewText(l.maxLabel, color.White)
	maxText.TextSize = 10
	maxText.Alignment = fyne.TextAlignLeading

	unitText := canvas.NewText(l.unit, color.RGBA{200, 200, 200, 255})
	unitText.TextSize = 9
	unitText.Alignment = fyne.TextAlignCenter

	return &colorLegendRenderer{
		legend:   l,
		raster:   l.raster,
		minText:  minText,
		maxText:  maxText,
		unitText: unitText,
	}
}

// MinSize returns the minimum size.
func (l *ColorLegend) MinSize() fyne.Size {
	return fyne.NewSize(60, 100)
}

// generateImage creates the legend gradient.
func (l *ColorLegend) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Draw gradient bar
	barX := 10
	barWidth := 20
	barY := 30
	barHeight := h - 60

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Gradient bar
	for y := 0; y < barHeight; y++ {
		t := 1.0 - float64(y)/float64(barHeight) // Top = 1.0 (max), bottom = 0.0 (min)
		var c color.RGBA
		switch l.colormap {
		case "viridis":
			c = viridisColor(t)
		case "plasma":
			c = plasmaColor(t)
		case "coolwarm":
			c = coolwarmColor(t)
		case "fecim":
			c = fecimColor(t)
		default:
			c = viridisColor(t)
		}

		for x := barX; x < barX+barWidth; x++ {
			img.Set(x, barY+y, c)
		}
	}

	// Border around bar
	borderColor := color.RGBA{150, 150, 150, 255}
	for x := barX - 1; x < barX+barWidth+1; x++ {
		img.Set(x, barY-1, borderColor)
		img.Set(x, barY+barHeight, borderColor)
	}
	for y := barY; y < barY+barHeight; y++ {
		img.Set(barX-1, y, borderColor)
		img.Set(barX+barWidth, y, borderColor)
	}

	// Tick marks for 30 levels
	if l.showTicks && l.levels == 30 {
		tickColor := color.RGBA{200, 200, 200, 255}
		for level := 0; level < 30; level++ {
			// Every 5th level
			if level%5 == 0 {
				t := float64(level) / 29.0
				y := barY + int(float64(barHeight)*(1.0-t))
				// Draw tick mark
				for x := barX + barWidth; x < barX+barWidth+5; x++ {
					img.Set(x, y, tickColor)
				}
			}
		}
	}

	return img
}

type colorLegendRenderer struct {
	legend   *ColorLegend
	raster   *canvas.Raster
	minText  *canvas.Text
	maxText  *canvas.Text
	unitText *canvas.Text
}

func (r *colorLegendRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
	r.maxText.Move(fyne.NewPos(35, 5))
	r.unitText.Move(fyne.NewPos(0, 20))
	r.minText.Move(fyne.NewPos(35, size.Height-15))
}

func (r *colorLegendRenderer) MinSize() fyne.Size {
	return r.legend.MinSize()
}

func (r *colorLegendRenderer) Refresh() {
	r.minText.Text = r.legend.minLabel
	r.maxText.Text = r.legend.maxLabel
	r.unitText.Text = r.legend.unit
	r.raster.Refresh()
	r.minText.Refresh()
	r.maxText.Refresh()
	r.unitText.Refresh()
}

func (r *colorLegendRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster, r.minText, r.maxText, r.unitText}
}

func (r *colorLegendRenderer) Destroy() {}

// MetricsPanel displays accuracy and energy metrics.
type MetricsPanel struct {
	widget.BaseWidget

	// Accuracy metrics
	idealAccuracy  float64
	actualAccuracy float64
	accuracyDelta  float64

	// Energy metrics
	fecimEnergy   float64
	gpuEnergy     float64
	efficiency    float64

	// Performance
	macOps  int
	latency float64

	// Labels
	labels map[string]*widget.Label
}

// NewMetricsPanel creates a new metrics panel.
func NewMetricsPanel() *MetricsPanel {
	m := &MetricsPanel{
		labels: make(map[string]*widget.Label),
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateMetrics updates all metrics.
func (m *MetricsPanel) UpdateMetrics(idealAcc, actualAcc, fecimE, gpuE float64, macs int, lat float64) {
	m.idealAccuracy = idealAcc
	m.actualAccuracy = actualAcc
	m.accuracyDelta = idealAcc - actualAcc
	m.fecimEnergy = fecimE
	m.gpuEnergy = gpuE
	m.efficiency = gpuE / fecimE
	m.macOps = macs
	m.latency = lat
	m.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (m *MetricsPanel) CreateRenderer() fyne.WidgetRenderer {
	// Create labels
	headerLabel := widget.NewLabelWithStyle("Live Metrics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	accHeader := widget.NewLabelWithStyle("Accuracy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	idealLabel := widget.NewLabel("Ideal: --")
	actualLabel := widget.NewLabel("Actual: --")
	deltaLabel := widget.NewLabel("Δ: --")

	energyHeader := widget.NewLabelWithStyle("Energy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fecimELabel := widget.NewLabel("FeCIM: --")
	gpuELabel := widget.NewLabel("GPU: --")
	effLabel := widget.NewLabel("Efficiency: --")

	perfHeader := widget.NewLabelWithStyle("Performance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	macLabel := widget.NewLabel("MACs: --")
	latLabel := widget.NewLabel("Latency: --")

	m.labels["ideal"] = idealLabel
	m.labels["actual"] = actualLabel
	m.labels["delta"] = deltaLabel
	m.labels["fecim"] = fecimELabel
	m.labels["gpu"] = gpuELabel
	m.labels["eff"] = effLabel
	m.labels["mac"] = macLabel
	m.labels["lat"] = latLabel

	content := container.NewVBox(
		headerLabel,
		widget.NewSeparator(),
		accHeader,
		idealLabel,
		actualLabel,
		deltaLabel,
		widget.NewSeparator(),
		energyHeader,
		fecimELabel,
		gpuELabel,
		effLabel,
		widget.NewSeparator(),
		perfHeader,
		macLabel,
		latLabel,
	)

	return widget.NewSimpleRenderer(content)
}

// Refresh updates the display.
func (m *MetricsPanel) Refresh() {
	if len(m.labels) == 0 {
		m.BaseWidget.Refresh()
		return
	}

	// Update accuracy labels
	if m.idealAccuracy > 0 {
		m.labels["ideal"].SetText(fmt.Sprintf("Ideal: %.1f%%", m.idealAccuracy))
		m.labels["actual"].SetText(fmt.Sprintf("Actual: %.1f%%", m.actualAccuracy))
		deltaStr := fmt.Sprintf("Δ: %.1f%%", m.accuracyDelta)
		if m.accuracyDelta > 0 {
			deltaStr = fmt.Sprintf("Δ: -%.1f%%", math.Abs(m.accuracyDelta))
		}
		m.labels["delta"].SetText(deltaStr)
	}

	// Update energy labels
	if m.fecimEnergy > 0 {
		m.labels["fecim"].SetText(fmt.Sprintf("FeCIM: %.2f pJ", m.fecimEnergy))
		m.labels["gpu"].SetText(fmt.Sprintf("GPU: %.0f pJ", m.gpuEnergy))
		m.labels["eff"].SetText(fmt.Sprintf("%.0f× better", m.efficiency))
	}

	// Update performance labels
	if m.macOps > 0 {
		m.labels["mac"].SetText(fmt.Sprintf("%d MACs", m.macOps))
		m.labels["lat"].SetText(fmt.Sprintf("%.0f ns", m.latency))
	}

	m.BaseWidget.Refresh()
}

// ComparisonBadge displays a visual comparison.
type ComparisonBadge struct {
	widget.BaseWidget

	metric      string
	fecimValue  string
	gpuValue    string
	improvement string

	raster *canvas.Raster
}

// NewComparisonBadge creates a new comparison badge.
func NewComparisonBadge(metric string) *ComparisonBadge {
	b := &ComparisonBadge{
		metric:      metric,
		fecimValue:  "--",
		gpuValue:    "--",
		improvement: "--",
	}
	b.ExtendBaseWidget(b)
	return b
}

// UpdateValues updates the comparison values.
func (b *ComparisonBadge) UpdateValues(fecimVal, gpuVal string, improvement string) {
	b.fecimValue = fecimVal
	b.gpuValue = gpuVal
	b.improvement = improvement
	b.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (b *ComparisonBadge) CreateRenderer() fyne.WidgetRenderer {
	b.raster = canvas.NewRaster(b.generateImage)
	return widget.NewSimpleRenderer(b.raster)
}

// MinSize returns minimum size.
func (b *ComparisonBadge) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

// generateImage creates the badge image.
func (b *ComparisonBadge) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{20, 40, 60, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Border
	borderColor := color.RGBA{0, 212, 255, 255} // FeCIM cyan
	for x := 0; x < w; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, 1, borderColor)
		img.Set(x, h-1, borderColor)
		img.Set(x, h-2, borderColor)
	}
	for y := 0; y < h; y++ {
		img.Set(0, y, borderColor)
		img.Set(1, y, borderColor)
		img.Set(w-1, y, borderColor)
		img.Set(w-2, y, borderColor)
	}

	return img
}

// AccuracyWaterfall displays stepwise accuracy degradation.
type AccuracyWaterfall struct {
	widget.BaseWidget

	steps       []WaterfallStep
	targetAccuracy float64

	raster *canvas.Raster
}

// WaterfallStep represents one degradation step.
type WaterfallStep struct {
	Label    string
	Accuracy float64
	Loss     float64
	Color    color.RGBA
}

// NewAccuracyWaterfall creates a new waterfall chart.
func NewAccuracyWaterfall() *AccuracyWaterfall {
	w := &AccuracyWaterfall{
		steps: []WaterfallStep{},
		targetAccuracy: 87.0, // Dr. Tour's reported 87%
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetSteps updates the waterfall steps.
func (w *AccuracyWaterfall) SetSteps(steps []WaterfallStep) {
	w.steps = steps
	w.Refresh()
}

// SetTarget sets the target accuracy line.
func (w *AccuracyWaterfall) SetTarget(target float64) {
	w.targetAccuracy = target
	w.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (w *AccuracyWaterfall) CreateRenderer() fyne.WidgetRenderer {
	w.raster = canvas.NewRaster(w.generateImage)
	return widget.NewSimpleRenderer(w.raster)
}

// MinSize returns minimum size.
func (w *AccuracyWaterfall) MinSize() fyne.Size {
	return fyne.NewSize(300, 200)
}

// generateImage creates the waterfall chart.
func (w *AccuracyWaterfall) generateImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(w.steps) == 0 {
		return img
	}

	// Chart area
	marginLeft := 80
	marginRight := 20
	marginTop := 30
	marginBottom := 40
	chartWidth := width - marginLeft - marginRight
	chartHeight := height - marginTop - marginBottom

	// Y-axis scale (0-100%)
	maxAcc := 100.0
	minAcc := 0.0

	// Draw Y-axis grid and labels
	gridColor := color.RGBA{60, 60, 70, 255}
	_ = color.RGBA{180, 180, 180, 255} // labelColor - reserved for future label rendering
	for i := 0; i <= 5; i++ {
		acc := minAcc + float64(i)*(maxAcc-minAcc)/5.0
		y := marginTop + chartHeight - int(float64(chartHeight)*(acc-minAcc)/(maxAcc-minAcc))

		// Grid line
		for x := marginLeft; x < marginLeft+chartWidth; x++ {
			img.Set(x, y, gridColor)
		}

		// TODO: Draw label text (requires font rendering - simplified for now)
	}

	// Draw target accuracy line
	if w.targetAccuracy > 0 {
		targetY := marginTop + chartHeight - int(float64(chartHeight)*(w.targetAccuracy-minAcc)/(maxAcc-minAcc))
		targetColor := color.RGBA{255, 200, 0, 255}
		for x := marginLeft; x < marginLeft+chartWidth; x += 4 {
			img.Set(x, targetY, targetColor)
			img.Set(x, targetY+1, targetColor)
		}
	}

	// Draw waterfall bars
	stepWidth := chartWidth / (len(w.steps) + 1)
	for i, step := range w.steps {
		x := marginLeft + (i+1)*stepWidth - stepWidth/2
		barWidth := stepWidth - 10
		y := marginTop + chartHeight - int(float64(chartHeight)*(step.Accuracy-minAcc)/(maxAcc-minAcc))

		// Draw bar
		for bx := x; bx < x+barWidth; bx++ {
			for by := y; by < marginTop+chartHeight; by++ {
				img.Set(bx, by, step.Color)
			}
		}

		// Draw border
		borderColor := color.RGBA{200, 200, 200, 255}
		for bx := x; bx < x+barWidth; bx++ {
			img.Set(bx, y, borderColor)
		}
		for by := y; by < marginTop+chartHeight; by++ {
			img.Set(x, by, borderColor)
			img.Set(x+barWidth-1, by, borderColor)
		}
	}

	return img
}

// BeforeAfterToggle shows side-by-side comparison of ideal vs actual.
type BeforeAfterToggle struct {
	widget.BaseWidget

	idealData  [][]float64
	actualData [][]float64
	mode       string // "split", "before", "after", "diff"

	leftHeatmap  *CrossbarHeatmap
	rightHeatmap *CrossbarHeatmap
	toggleGroup  *widget.RadioGroup
}

// NewBeforeAfterToggle creates a new comparison widget.
func NewBeforeAfterToggle(rows, cols int) *BeforeAfterToggle {
	b := &BeforeAfterToggle{
		mode:         "split",
		leftHeatmap:  NewCrossbarHeatmap(rows, cols),
		rightHeatmap: NewCrossbarHeatmap(rows, cols),
	}

	b.leftHeatmap.SetColormap("fecim")
	b.rightHeatmap.SetColormap("fecim")

	b.ExtendBaseWidget(b)
	return b
}

// SetData updates the comparison data.
func (b *BeforeAfterToggle) SetData(ideal, actual [][]float64) {
	b.idealData = ideal
	b.actualData = actual
	b.updateDisplay()
}

// SetMode changes the display mode.
func (b *BeforeAfterToggle) SetMode(mode string) {
	b.mode = mode
	b.updateDisplay()
}

// updateDisplay refreshes the heatmaps based on current mode.
func (b *BeforeAfterToggle) updateDisplay() {
	if b.idealData == nil || b.actualData == nil {
		return
	}

	switch b.mode {
	case "split":
		b.leftHeatmap.SetData(b.idealData)
		b.rightHeatmap.SetData(b.actualData)
	case "before":
		b.leftHeatmap.SetData(b.idealData)
		b.rightHeatmap.SetData(b.idealData)
	case "after":
		b.leftHeatmap.SetData(b.actualData)
		b.rightHeatmap.SetData(b.actualData)
	case "diff":
		diff := b.computeDifference()
		b.leftHeatmap.SetData(diff)
		b.rightHeatmap.SetData(diff)
	}
}

// computeDifference computes the difference map.
func (b *BeforeAfterToggle) computeDifference() [][]float64 {
	rows := len(b.idealData)
	if rows == 0 {
		return nil
	}
	cols := len(b.idealData[0])

	diff := make([][]float64, rows)
	for i := range diff {
		diff[i] = make([]float64, cols)
		for j := range diff[i] {
			diff[i][j] = math.Abs(b.idealData[i][j] - b.actualData[i][j])
		}
	}
	return diff
}

// CreateRenderer implements fyne.Widget.
func (b *BeforeAfterToggle) CreateRenderer() fyne.WidgetRenderer {
	b.toggleGroup = widget.NewRadioGroup(
		[]string{"Split View", "Ideal Only", "Actual Only", "Difference"},
		func(value string) {
			mode := map[string]string{
				"Split View":   "split",
				"Ideal Only":   "before",
				"Actual Only":  "after",
				"Difference":   "diff",
			}[value]
			b.SetMode(mode)
		},
	)
	b.toggleGroup.SetSelected("Split View")

	leftLabel := widget.NewLabelWithStyle("Ideal (No Non-Idealities)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	rightLabel := widget.NewLabelWithStyle("Actual (With Non-Idealities)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	leftPane := container.NewBorder(leftLabel, nil, nil, nil, b.leftHeatmap)
	rightPane := container.NewBorder(rightLabel, nil, nil, nil, b.rightHeatmap)

	splitView := container.NewHSplit(leftPane, rightPane)
	splitView.SetOffset(0.5)

	content := container.NewBorder(b.toggleGroup, nil, nil, nil, splitView)

	return widget.NewSimpleRenderer(content)
}
