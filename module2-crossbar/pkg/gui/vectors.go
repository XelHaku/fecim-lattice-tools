//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for crossbar visualization.
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

	"fecim-lattice-tools/shared/mathutil"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// VectorBarChart displays a vector as a bar chart.
type VectorBarChart struct {
	widget.BaseWidget

	values   []float64
	labels   []string
	title    string
	barColor color.Color
	maxVal   float64
	minVal   float64
	unit     string // Y-axis unit label (e.g., "V", "µA", "mS")

	raster *canvas.Raster
}

// NewVectorBarChart creates a new vector bar chart.
func NewVectorBarChart(title string, barColor color.Color) *VectorBarChart {
	v := &VectorBarChart{
		title:    title,
		barColor: barColor,
		maxVal:   1.0,
		minVal:   0.0,
	}
	v.ExtendBaseWidget(v)
	return v
}

// SetValues updates the chart data.
func (v *VectorBarChart) SetValues(values []float64) {
	v.values = values
	v.minVal = 0
	v.maxVal = 0

	for _, val := range values {
		if val > v.maxVal {
			v.maxVal = val
		}
		if val < v.minVal {
			v.minVal = val
		}
	}

	if v.maxVal <= v.minVal {
		v.maxVal = v.minVal + 1
	}

	// Skip refresh during startup stabilization to prevent resize oscillation
	if sharedwidgets.IsStartupStabilizing() {
		return
	}

	fyne.Do(func() {
		v.Refresh()
	})
}

// SetLabels sets the bar labels.
func (v *VectorBarChart) SetLabels(labels []string) {
	v.labels = labels
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		v.Refresh()
	})
}

// SetUnit sets the Y-axis unit label.
func (v *VectorBarChart) SetUnit(unit string) {
	v.unit = unit
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		v.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (v *VectorBarChart) CreateRenderer() fyne.WidgetRenderer {
	v.raster = canvas.NewRaster(v.generateImage)

	titleLabel := widget.NewLabel(v.title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	// Y-axis labels showing min/max values and unit
	maxLabel := widget.NewLabel("")
	maxLabel.TextStyle = fyne.TextStyle{Monospace: false}
	minLabel := widget.NewLabel("")
	minLabel.TextStyle = fyne.TextStyle{Monospace: false}

	yAxisLabels := container.NewVBox(
		maxLabel,
		widget.NewLabel(""), // spacer
		minLabel,
	)

	chartWithAxis := container.NewBorder(
		nil,         // top
		nil,         // bottom
		yAxisLabels, // left - Y-axis labels
		nil,         // right
		v.raster,    // center
	)

	content := container.NewBorder(
		titleLabel,    // top
		nil,           // bottom
		nil,           // left
		nil,           // right
		chartWithAxis, // center
	)

	return &vectorBarChartRenderer{
		chart:    v,
		content:  content,
		maxLabel: maxLabel,
		minLabel: minLabel,
	}
}

// MinSize returns the minimum size.
func (v *VectorBarChart) MinSize() fyne.Size {
	return fyne.NewSize(300, 150)
}

// generateImage creates the bar chart image.
func (v *VectorBarChart) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{40, 40, 50, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(v.values) == 0 {
		return img
	}

	// Calculate bar dimensions
	padding := 30
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding
	barWidth := chartWidth / len(v.values)
	if barWidth < 2 {
		barWidth = 2
	}

	// Draw axis
	axisColor := color.RGBA{100, 100, 100, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, h-padding, axisColor)
	}
	for y := padding; y < h-padding; y++ {
		img.Set(padding, y, axisColor)
	}

	// Draw bars
	for i, val := range v.values {
		// Normalize value
		normVal := (val - v.minVal) / (v.maxVal - v.minVal)
		barHeight := int(normVal * float64(chartHeight))

		x0 := padding + i*barWidth + 1
		x1 := padding + (i+1)*barWidth - 1
		y0 := h - padding - barHeight
		y1 := h - padding

		// Draw bar
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, v.barColor)
				}
			}
		}
	}

	return img
}

// vectorBarChartRenderer is a custom renderer for VectorBarChart with Y-axis labels.
type vectorBarChartRenderer struct {
	chart       *VectorBarChart
	content     fyne.CanvasObject
	maxLabel    *widget.Label
	minLabel    *widget.Label
	cache       sharedwidgets.LayoutCache // Shared utility for safe layout
	lastMaxText string                    // Cache to avoid redundant SetText calls
	lastMinText string
}

func (r *vectorBarChartRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("vectorBarChartRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.content.Resize(size)
	r.cache.MarkLayout(size)
}

func (r *vectorBarChartRenderer) MinSize() fyne.Size {
	return r.chart.MinSize()
}

func (r *vectorBarChartRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("vectorBarChartRenderer", r.chart.Size())
	// Update Y-axis labels with current min/max values and unit - only if changed
	unit := r.chart.unit
	if r.chart.maxVal != 0 || r.chart.minVal != 0 {
		maxText := fmt.Sprintf("%.2f%s", r.chart.maxVal, unit)
		minText := fmt.Sprintf("%.2f%s", r.chart.minVal, unit)
		if maxText != r.lastMaxText {
			r.maxLabel.SetText(maxText)
			r.lastMaxText = maxText
		}
		if minText != r.lastMinText {
			r.minLabel.SetText(minText)
			r.lastMinText = minText
		}
	}
	r.content.Refresh()
}

func (r *vectorBarChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content}
}

func (r *vectorBarChartRenderer) Destroy() {}

// MiniMatrixView displays a small heatmap of the conductance matrix.
type MiniMatrixView struct {
	widget.BaseWidget

	data   [][]float64
	rows   int
	cols   int
	raster *canvas.Raster
}

// NewMiniMatrixView creates a new mini matrix visualization.
func NewMiniMatrixView() *MiniMatrixView {
	m := &MiniMatrixView{}
	m.ExtendBaseWidget(m)
	return m
}

// SetData updates the matrix data.
func (m *MiniMatrixView) SetData(data [][]float64) {
	m.data = data
	if len(data) > 0 {
		m.rows = len(data)
		m.cols = len(data[0])
	}
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		m.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (m *MiniMatrixView) CreateRenderer() fyne.WidgetRenderer {
	m.raster = canvas.NewRaster(m.generateImage)
	return widget.NewSimpleRenderer(m.raster)
}

// MinSize returns minimum size.
func (m *MiniMatrixView) MinSize() fyne.Size {
	return fyne.NewSize(150, 100)
}

// generateImage creates the mini matrix heatmap.
func (m *MiniMatrixView) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{35, 35, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(m.data) == 0 || m.rows == 0 || m.cols == 0 {
		// Draw placeholder text area
		placeholderColor := color.RGBA{60, 60, 70, 255}
		for y := h / 4; y < 3*h/4; y++ {
			for x := w / 4; x < 3*w/4; x++ {
				img.Set(x, y, placeholderColor)
			}
		}
		return img
	}

	// Limit to max 32x32 for performance
	displayRows := m.rows
	displayCols := m.cols
	if displayRows > 32 {
		displayRows = 32
	}
	if displayCols > 32 {
		displayCols = 32
	}

	// Calculate cell size
	padding := 10
	cellW := float64(w-2*padding) / float64(displayCols)
	cellH := float64(h-2*padding) / float64(displayRows)
	cellSize := math.Min(cellW, cellH)
	if cellSize < 2 {
		cellSize = 2
	}

	// Find min/max for normalization
	minVal := math.Inf(1)
	maxVal := math.Inf(-1)
	for i := 0; i < displayRows && i < len(m.data); i++ {
		for j := 0; j < displayCols && j < len(m.data[i]); j++ {
			if m.data[i][j] < minVal {
				minVal = m.data[i][j]
			}
			if m.data[i][j] > maxVal {
				maxVal = m.data[i][j]
			}
		}
	}
	if maxVal <= minVal {
		maxVal = minVal + 1
	}

	// Draw cells
	for i := 0; i < displayRows && i < len(m.data); i++ {
		for j := 0; j < displayCols && j < len(m.data[i]); j++ {
			normVal := (m.data[i][j] - minVal) / (maxVal - minVal)
			cellColor := fecimColor(normVal)

			x0 := int(float64(padding) + float64(j)*cellSize)
			y0 := int(float64(padding) + float64(i)*cellSize)
			x1 := int(float64(padding) + float64(j+1)*cellSize - 1)
			y1 := int(float64(padding) + float64(i+1)*cellSize - 1)

			for y := y0; y < y1 && y < h; y++ {
				for x := x0; x < x1 && x < w; x++ {
					img.Set(x, y, cellColor)
				}
			}
		}
	}

	// Draw border
	borderColor := color.RGBA{100, 120, 140, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, padding, borderColor)
		img.Set(x, h-padding-1, borderColor)
	}
	for y := padding; y < h-padding; y++ {
		img.Set(padding, y, borderColor)
		img.Set(w-padding-1, y, borderColor)
	}

	return img
}

// MVMVisualization shows the matrix-vector multiplication process.
type MVMVisualization struct {
	widget.BaseWidget

	inputVector  []float64
	outputVector []float64
	weights      [][]float64

	inputChart  *VectorBarChart
	outputChart *VectorBarChart
	miniMatrix  *MiniMatrixView
}

// NewMVMVisualization creates a new MVM visualization.
func NewMVMVisualization() *MVMVisualization {
	inputChart := NewVectorBarChart("Input Vector (V)", color.RGBA{100, 150, 255, 255})
	inputChart.SetUnit(" V") // Voltage unit

	outputChart := NewVectorBarChart("Output Vector (I)", color.RGBA{255, 200, 100, 255})
	outputChart.SetUnit(" µA") // Current unit

	miniMatrix := NewMiniMatrixView()

	m := &MVMVisualization{
		inputChart:  inputChart,
		outputChart: outputChart,
		miniMatrix:  miniMatrix,
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetInput updates the input vector visualization.
func (m *MVMVisualization) SetInput(input []float64) {
	m.inputVector = input
	m.inputChart.SetValues(input)
}

// SetOutput updates the output vector visualization.
func (m *MVMVisualization) SetOutput(output []float64) {
	m.outputVector = output
	m.outputChart.SetValues(output)
}

// SetWeights updates the conductance matrix visualization.
func (m *MVMVisualization) SetWeights(weights [][]float64) {
	m.weights = weights
	if m.miniMatrix != nil {
		m.miniMatrix.SetData(weights)
	}
}

// CreateRenderer implements fyne.Widget.
func (m *MVMVisualization) CreateRenderer() fyne.WidgetRenderer {
	// Input label
	inputLabel := canvas.NewText("Input Vector (V)", color.RGBA{150, 180, 255, 255})
	inputLabel.TextSize = 14
	inputLabel.Alignment = fyne.TextAlignCenter

	// MVM operation symbol
	mvmLabel := widget.NewLabel("W × V = I")
	mvmLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: false}
	mvmLabel.Alignment = fyne.TextAlignCenter

	// Matrix label
	matrixLabel := canvas.NewText("Conductance Matrix (W)", color.RGBA{180, 180, 180, 255})
	matrixLabel.TextSize = 14
	matrixLabel.Alignment = fyne.TextAlignCenter

	// Output label
	outputLabel := canvas.NewText("Output Current (I)", color.RGBA{255, 200, 150, 255})
	outputLabel.TextSize = 14
	outputLabel.Alignment = fyne.TextAlignCenter

	// Stats
	statsLabel := widget.NewLabel("")
	if len(m.inputVector) > 0 && len(m.outputVector) > 0 {
		var sumIn, sumOut float64
		for _, v := range m.inputVector {
			sumIn += v
		}
		for _, v := range m.outputVector {
			sumOut += v
		}
		statsLabel.SetText(fmt.Sprintf("Sum(V)=%.2f  Sum(I)=%.2f", sumIn, sumOut))
	}
	statsLabel.Alignment = fyne.TextAlignCenter

	// Layout: Input → Matrix → Output (vertical flow)
	content := container.NewVBox(
		inputLabel,
		m.inputChart,
		mvmLabel,
		matrixLabel,
		container.NewCenter(m.miniMatrix),
		outputLabel,
		m.outputChart,
		statsLabel,
	)

	return widget.NewSimpleRenderer(content)
}

// ComparisonChart shows ideal vs actual output comparison.
type ComparisonChart struct {
	widget.BaseWidget

	ideal  []float64
	actual []float64
	title  string
	rmse   float64

	raster *canvas.Raster
}

// NewComparisonChart creates a comparison chart.
func NewComparisonChart(title string) *ComparisonChart {
	c := &ComparisonChart{
		title: title,
	}
	c.ExtendBaseWidget(c)
	return c
}

// SetData updates the comparison data.
func (c *ComparisonChart) SetData(ideal, actual []float64) {
	c.ideal = ideal
	c.actual = actual

	// Calculate RMSE
	if len(ideal) > 0 && len(ideal) == len(actual) {
		var sumSq float64
		for i := range ideal {
			diff := ideal[i] - actual[i]
			sumSq += diff * diff
		}
		c.rmse = math.Sqrt(sumSq / float64(len(ideal)))
	}

	fyne.Do(func() {
		c.Refresh()
	})
}

// GetRMSE returns the root mean square error.
func (c *ComparisonChart) GetRMSE() float64 {
	return c.rmse
}

// CreateRenderer implements fyne.Widget.
func (c *ComparisonChart) CreateRenderer() fyne.WidgetRenderer {
	c.raster = canvas.NewRaster(c.generateImage)

	titleLabel := widget.NewLabel(c.title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	rmseLabel := widget.NewLabel(fmt.Sprintf("RMSE: %.6f", c.rmse))
	rmseLabel.Alignment = fyne.TextAlignCenter

	content := container.NewBorder(
		titleLabel,
		rmseLabel,
		nil,
		nil,
		c.raster,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (c *ComparisonChart) MinSize() fyne.Size {
	return fyne.NewSize(400, 200)
}

// generateImage creates the comparison chart image.
func (c *ComparisonChart) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(c.ideal) == 0 || len(c.actual) == 0 {
		return img
	}

	// Find range
	maxVal := 0.0
	for _, v := range c.ideal {
		if v > maxVal {
			maxVal = v
		}
	}
	for _, v := range c.actual {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	// Calculate dimensions
	padding := 40
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding
	barWidth := chartWidth / len(c.ideal) / 2

	idealColor := color.RGBA{100, 150, 255, 255}
	actualColor := color.RGBA{255, 150, 100, 255}

	// Draw bars
	for i := 0; i < len(c.ideal); i++ {
		// Ideal bar
		normIdeal := c.ideal[i] / maxVal
		idealHeight := int(normIdeal * float64(chartHeight))
		x0 := padding + i*barWidth*2
		x1 := x0 + barWidth - 1

		for y := h - padding - idealHeight; y < h-padding; y++ {
			for x := x0; x < x1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, idealColor)
				}
			}
		}

		// Actual bar
		if i < len(c.actual) {
			normActual := c.actual[i] / maxVal
			actualHeight := int(normActual * float64(chartHeight))
			x0 = padding + i*barWidth*2 + barWidth
			x1 = x0 + barWidth - 1

			for y := h - padding - actualHeight; y < h-padding; y++ {
				for x := x0; x < x1; x++ {
					if x >= 0 && x < w && y >= 0 && y < h {
						img.Set(x, y, actualColor)
					}
				}
			}
		}
	}

	// Draw legend
	legendY := 10
	for x := padding; x < padding+20; x++ {
		img.Set(x, legendY, idealColor)
		img.Set(x, legendY+1, idealColor)
	}
	for x := padding + 60; x < padding+80; x++ {
		img.Set(x, legendY, actualColor)
		img.Set(x, legendY+1, actualColor)
	}

	return img
}

// DiscreteLevel30Indicator shows the 30 FeCIM levels graphically.
type DiscreteLevel30Indicator struct {
	widget.BaseWidget

	value float64 // 0-1 normalized
	level int     // 0-29

	raster *canvas.Raster
}

// NewDiscreteLevel30Indicator creates a new indicator.
func NewDiscreteLevel30Indicator() *DiscreteLevel30Indicator {
	d := &DiscreteLevel30Indicator{}
	d.ExtendBaseWidget(d)
	return d
}

// SetValue sets the value (0-1) and calculates the discrete level.
func (d *DiscreteLevel30Indicator) SetValue(value float64) {
	d.value = mathutil.Clamp(value, 0, 1)
	d.level = int(math.Round(d.value * 29))
	fyne.Do(func() {
		d.Refresh()
	})
}

// GetLevel returns the current discrete level (0-29).
func (d *DiscreteLevel30Indicator) GetLevel() int {
	return d.level
}

// CreateRenderer implements fyne.Widget.
func (d *DiscreteLevel30Indicator) CreateRenderer() fyne.WidgetRenderer {
	d.raster = canvas.NewRaster(d.generateImage)

	levelLabel := widget.NewLabel(fmt.Sprintf("Level: %d/29 (%.3f)", d.level, d.value))
	levelLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		widget.NewLabel("30 Discrete FeCIM Levels (claim)"),
		d.raster,
		levelLabel,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (d *DiscreteLevel30Indicator) MinSize() fyne.Size {
	return fyne.NewSize(320, 60)
}

// generateImage creates the level indicator image.
func (d *DiscreteLevel30Indicator) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw 30 level bars
	padding := 10
	barWidth := (w - 2*padding) / 30
	barHeight := h - 2*padding

	for i := 0; i < 30; i++ {
		x0 := padding + i*barWidth
		x1 := x0 + barWidth - 1
		y0 := padding
		y1 := y0 + barHeight

		var barColor color.RGBA
		if i <= d.level {
			// Active level - use FeCIM colormap
			t := float64(i) / 29.0
			barColor = fecimColor(t)
		} else {
			// Inactive level - dim
			barColor = color.RGBA{60, 60, 70, 255}
		}

		// Highlight current level
		if i == d.level {
			// Draw border
			borderColor := color.RGBA{255, 255, 255, 255}
			for x := x0; x < x1; x++ {
				img.Set(x, y0, borderColor)
				img.Set(x, y1-1, borderColor)
			}
			for y := y0; y < y1; y++ {
				img.Set(x0, y, borderColor)
				img.Set(x1-1, y, borderColor)
			}
			y0++
			y1--
			x0++
			x1--
		}

		// Draw bar
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				img.Set(x, y, barColor)
			}
		}
	}

	return img
}
