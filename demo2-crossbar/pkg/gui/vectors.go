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

	v.Refresh()
}

// SetLabels sets the bar labels.
func (v *VectorBarChart) SetLabels(labels []string) {
	v.labels = labels
	v.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (v *VectorBarChart) CreateRenderer() fyne.WidgetRenderer {
	v.raster = canvas.NewRaster(v.generateImage)

	titleLabel := widget.NewLabel(v.title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	content := container.NewBorder(
		titleLabel, // top
		nil,        // bottom
		nil,        // left
		nil,        // right
		v.raster,   // center
	)

	return widget.NewSimpleRenderer(content)
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

// MVMVisualization shows the matrix-vector multiplication process.
type MVMVisualization struct {
	widget.BaseWidget

	inputVector  []float64
	outputVector []float64
	weights      [][]float64

	inputChart  *VectorBarChart
	outputChart *VectorBarChart
}

// NewMVMVisualization creates a new MVM visualization.
func NewMVMVisualization() *MVMVisualization {
	m := &MVMVisualization{
		inputChart:  NewVectorBarChart("Input Vector (V)", color.RGBA{100, 150, 255, 255}),
		outputChart: NewVectorBarChart("Output Vector (I)", color.RGBA{255, 200, 100, 255}),
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

// CreateRenderer implements fyne.Widget.
func (m *MVMVisualization) CreateRenderer() fyne.WidgetRenderer {
	// MVM operation symbol
	mvmLabel := widget.NewLabel("Crossbar W × V = I")
	mvmLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	mvmLabel.Alignment = fyne.TextAlignCenter

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

	content := container.NewVBox(
		m.inputChart,
		mvmLabel,
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

	c.Refresh()
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
	d.value = clamp(value, 0, 1)
	d.level = int(math.Round(d.value * 29))
	d.Refresh()
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
		widget.NewLabel("30 Discrete FeCIM Levels"),
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
