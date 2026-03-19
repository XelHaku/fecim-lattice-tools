// Package gui provides Fyne-based GUI components for MNIST visualization.
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
)

// LayerActivationView displays neural network layer activations.
type LayerActivationView struct {
	widget.BaseWidget

	// Layer data
	inputLayer  []float64 // 784 values (28x28)
	hiddenLayer []float64 // variable size
	outputLayer []float64 // 10 values

	// Rasters for each layer
	inputRaster  *canvas.Raster
	hiddenRaster *canvas.Raster
	outputRaster *canvas.Raster

	// Labels
	predictionLabel *widget.Label
	confidenceLabel *widget.Label
}

// NewLayerActivationView creates a new layer activation visualization.
func NewLayerActivationView() *LayerActivationView {
	lav := &LayerActivationView{
		inputLayer:  make([]float64, 784),
		hiddenLayer: make([]float64, 128),
		outputLayer: make([]float64, 10),
	}
	lav.ExtendBaseWidget(lav)
	return lav
}

// SetInput sets the input layer (28x28 = 784 values).
func (lav *LayerActivationView) SetInput(input []float64) {
	lav.inputLayer = input
	fyne.Do(func() {
		lav.Refresh()
	})
}

// SetHidden sets the hidden layer activations.
func (lav *LayerActivationView) SetHidden(hidden []float64) {
	lav.hiddenLayer = hidden
	fyne.Do(func() {
		lav.Refresh()
	})
}

// SetOutput sets the output layer (10 class probabilities).
func (lav *LayerActivationView) SetOutput(output []float64) {
	lav.outputLayer = output
	fyne.Do(func() {
		lav.Refresh()
	})
}

// SetActivations sets all layer activations at once.
func (lav *LayerActivationView) SetActivations(input, hidden, output []float64) {
	lav.inputLayer = input
	lav.hiddenLayer = hidden
	lav.outputLayer = output
	fyne.Do(func() {
		lav.Refresh()
	})
}

// GetPrediction returns the predicted class and confidence.
func (lav *LayerActivationView) GetPrediction() (int, float64) {
	if len(lav.outputLayer) == 0 {
		return -1, 0
	}

	maxIdx := 0
	maxVal := lav.outputLayer[0]
	for i, v := range lav.outputLayer {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx, maxVal
}

// CreateRenderer implements fyne.Widget.
func (lav *LayerActivationView) CreateRenderer() fyne.WidgetRenderer {
	// Create rasters for each layer - they will resize with their container
	lav.inputRaster = canvas.NewRaster(lav.generateInputImage)
	lav.hiddenRaster = canvas.NewRaster(lav.generateHiddenImage)
	lav.outputRaster = canvas.NewRaster(lav.generateOutputImage)

	// Labels
	inputLabel := widget.NewLabel("Input (28×28)")
	inputLabel.TextStyle = fyne.TextStyle{Bold: true}
	inputLabel.Alignment = fyne.TextAlignCenter

	hiddenLabel := widget.NewLabel("Hidden (128)")
	hiddenLabel.TextStyle = fyne.TextStyle{Bold: true}
	hiddenLabel.Alignment = fyne.TextAlignCenter

	outputLabel := widget.NewLabel("Output (10)")
	outputLabel.TextStyle = fyne.TextStyle{Bold: true}
	outputLabel.Alignment = fyne.TextAlignCenter

	// Prediction labels
	lav.predictionLabel = widget.NewLabel("Prediction: -")
	lav.predictionLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: false}
	lav.predictionLabel.Alignment = fyne.TextAlignCenter

	lav.confidenceLabel = widget.NewLabel("Confidence: -")
	lav.confidenceLabel.Alignment = fyne.TextAlignCenter

	// Layout each layer with Border+Max for proper expansion
	inputBox := container.NewBorder(
		inputLabel, nil, nil, nil,
		container.NewMax(lav.inputRaster),
	)

	hiddenBox := container.NewBorder(
		hiddenLabel, nil, nil, nil,
		container.NewMax(lav.hiddenRaster),
	)

	outputBox := container.NewBorder(
		outputLabel,
		container.NewVBox(
			widget.NewSeparator(),
			lav.predictionLabel,
			lav.confidenceLabel,
		),
		nil, nil,
		container.NewMax(lav.outputRaster),
	)

	// Use HSplit for proper proportional layout that resizes
	leftSplit := container.NewHSplit(inputBox, hiddenBox)
	leftSplit.SetOffset(0.45)

	mainSplit := container.NewHSplit(leftSplit, outputBox)
	mainSplit.SetOffset(0.65)

	return widget.NewSimpleRenderer(mainSplit)
}

// MinSize returns minimum size - smaller to allow flexible resizing.
func (lav *LayerActivationView) MinSize() fyne.Size {
	return fyne.NewSize(400, 150)
}

// generateInputImage creates the 28x28 input visualization.
func (lav *LayerActivationView) generateInputImage(w, h int) image.Image {
	// Use actual widget size for responsive layout
	if w < 10 {
		w = 140
	}
	if h < 10 {
		h = 140
	}
	size := w
	if h < w {
		size = h
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(lav.inputLayer) < 784 {
		return img
	}

	// Center the 28x28 grid
	offsetX := (w - size) / 2
	offsetY := (h - size) / 2

	// Draw 28x28 pixels
	cellSize := size / 28
	for py := 0; py < 28; py++ {
		for px := 0; px < 28; px++ {
			value := lav.inputLayer[py*28+px]
			if value > 0 {
				intensity := uint8(mathutil.Clamp(value, 0, 1) * 255)
				c := color.RGBA{
					R: uint8(float64(intensity) * 0.7),
					G: intensity,
					B: intensity,
					A: 255,
				}

				x0 := offsetX + px*cellSize
				y0 := offsetY + py*cellSize
				for y := y0; y < y0+cellSize; y++ {
					for x := x0; x < x0+cellSize; x++ {
						if x >= 0 && x < w && y >= 0 && y < h {
							img.Set(x, y, c)
						}
					}
				}
			}
		}
	}

	return img
}

// generateHiddenImage creates the hidden layer activation visualization.
func (lav *LayerActivationView) generateHiddenImage(w, h int) image.Image {
	// Use actual widget dimensions for responsive layout
	if w < 10 {
		w = 160
	}
	if h < 10 {
		h = 140
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(lav.hiddenLayer) == 0 {
		return img
	}

	// Arrange neurons in a grid
	neurons := len(lav.hiddenLayer)
	cols := int(math.Ceil(math.Sqrt(float64(neurons))))
	rows := (neurons + cols - 1) / cols

	padding := 5
	cellW := (w - 2*padding) / cols
	cellH := (h - 2*padding) / rows
	if cellW < 2 {
		cellW = 2
	}
	if cellH < 2 {
		cellH = 2
	}

	// Find max for normalization
	maxVal := 0.0
	for _, v := range lav.hiddenLayer {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	// Draw neurons
	for i, value := range lav.hiddenLayer {
		col := i % cols
		row := i / cols

		normVal := value / maxVal
		intensity := uint8(mathutil.Clamp(normVal, 0, 1) * 255)

		// Use orange-yellow gradient for hidden layer
		c := color.RGBA{
			R: intensity,
			G: uint8(float64(intensity) * 0.7),
			B: uint8(float64(intensity) * 0.2),
			A: 255,
		}

		x0 := padding + col*cellW
		y0 := padding + row*cellH
		for y := y0; y < y0+cellH-1 && y < h; y++ {
			for x := x0; x < x0+cellW-1 && x < w; x++ {
				img.Set(x, y, c)
			}
		}
	}

	return img
}

// generateOutputImage creates the output layer bar chart.
func (lav *LayerActivationView) generateOutputImage(w, h int) image.Image {
	// Use actual widget dimensions for responsive layout
	if w < 10 {
		w = 200
	}
	if h < 10 {
		h = 140
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(lav.outputLayer) != 10 {
		return img
	}

	// Find prediction
	maxIdx := 0
	maxVal := lav.outputLayer[0]
	for i, v := range lav.outputLayer {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	// Update prediction labels
	if lav.predictionLabel != nil {
		lav.predictionLabel.SetText(fmt.Sprintf("Prediction: %d", maxIdx))
	}
	if lav.confidenceLabel != nil {
		lav.confidenceLabel.SetText(fmt.Sprintf("Confidence: %.1f%%", maxVal*100))
	}

	// Draw bar chart with responsive dimensions
	padding := 15
	barWidth := (w - 2*padding) / 10
	chartHeight := h - 2*padding

	// Draw bars for each class
	for i := 0; i < 10; i++ {
		value := lav.outputLayer[i]
		barHeight := int(mathutil.Clamp(value, 0, 1) * float64(chartHeight))

		x0 := padding + i*barWidth + 1
		x1 := padding + (i+1)*barWidth - 1
		y0 := h - padding - barHeight
		y1 := h - padding

		// Color: green for prediction, coral for others
		var c color.RGBA
		if i == maxIdx {
			c = color.RGBA{100, 255, 150, 255} // Green for predicted
		} else {
			c = color.RGBA{255, 127, 80, 200} // Coral for others
		}

		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, c)
				}
			}
		}
	}

	// Draw axis
	// ACCESSIBILITY FIX: Increased contrast from (80,80,90) to (100,105,120)
	axisColor := color.RGBA{100, 105, 120, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, h-padding, axisColor)
	}
	for y := padding; y < h-padding; y++ {
		img.Set(padding, y, axisColor)
	}

	return img
}

// OutputBarChart provides a standalone output visualization.
type OutputBarChart struct {
	widget.BaseWidget

	values      []float64
	labels      []string
	predicted   int
	selectedBar int
	focused     bool
	raster      *canvas.Raster
}

// NewOutputBarChart creates a new output bar chart.
func NewOutputBarChart() *OutputBarChart {
	obc := &OutputBarChart{
		values:      make([]float64, 10),
		labels:      []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		predicted:   -1,
		selectedBar: 0,
	}
	obc.ExtendBaseWidget(obc)
	return obc
}

// SetValues updates the chart with new probabilities.
func (obc *OutputBarChart) SetValues(values []float64) {
	obc.values = values

	// Find prediction
	obc.predicted = 0
	maxVal := 0.0
	for i, v := range values {
		if v > maxVal {
			maxVal = v
			obc.predicted = i
		}
	}

	fyne.Do(func() {
		obc.Refresh()
	})
}

// GetPrediction returns the predicted class.
func (obc *OutputBarChart) GetPrediction() int {
	return obc.predicted
}

// FocusGained implements fyne.Focusable.
func (obc *OutputBarChart) FocusGained() {
	obc.focused = true
	obc.Refresh()
}

// FocusLost implements fyne.Focusable.
func (obc *OutputBarChart) FocusLost() {
	obc.focused = false
	obc.Refresh()
}

// TypedRune implements fyne.Focusable.
func (obc *OutputBarChart) TypedRune(_ rune) {}

// TypedKey enables keyboard navigation across class bars.
func (obc *OutputBarChart) TypedKey(ev *fyne.KeyEvent) {
	if ev == nil || len(obc.values) == 0 {
		return
	}
	switch ev.Name {
	case fyne.KeyLeft:
		if obc.selectedBar > 0 {
			obc.selectedBar--
		}
	case fyne.KeyRight:
		if obc.selectedBar < len(obc.values)-1 {
			obc.selectedBar++
		}
	case fyne.KeyHome:
		obc.selectedBar = 0
	case fyne.KeyEnd:
		obc.selectedBar = len(obc.values) - 1
	default:
		return
	}
	obc.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (obc *OutputBarChart) CreateRenderer() fyne.WidgetRenderer {
	obc.raster = canvas.NewRaster(obc.generateImage)

	titleLabel := widget.NewLabel("Class Probabilities (0-9)")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	// Use Max container for raster to fill available space
	content := container.NewBorder(
		titleLabel,
		nil,
		nil,
		nil,
		container.NewMax(obc.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size - smaller for flexible layout.
func (obc *OutputBarChart) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

// generateImage creates the bar chart image.
func (obc *OutputBarChart) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(obc.values) == 0 {
		return img
	}

	// Calculate dimensions
	padding := 30
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding
	barWidth := chartWidth / len(obc.values)

	// Draw bars
	for i, val := range obc.values {
		barHeight := int(mathutil.Clamp(val, 0, 1) * float64(chartHeight))

		x0 := padding + i*barWidth + 2
		x1 := padding + (i+1)*barWidth - 2
		y0 := h - padding - barHeight
		y1 := h - padding

		// Color based on prediction and keyboard selection
		var c color.RGBA
		if i == obc.selectedBar {
			c = color.RGBA{255, 180, 60, 255} // Orange for focused/selected bar
		} else if i == obc.predicted {
			c = color.RGBA{0, 230, 180, 255} // Cyan for predicted
		} else {
			c = color.RGBA{100, 100, 120, 255} // Gray for others
		}

		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, c)
				}
			}
		}
	}

	// Draw axis
	// ACCESSIBILITY FIX: Increased contrast from (80,80,90) to (100,105,120)
	axisColor := color.RGBA{100, 105, 120, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, h-padding, axisColor)
	}

	if obc.focused {
		focusColor := color.RGBA{255, 165, 0, 255}
		for x := 0; x < w; x++ {
			img.Set(x, 0, focusColor)
			img.Set(x, h-1, focusColor)
		}
		for y := 0; y < h; y++ {
			img.Set(0, y, focusColor)
			img.Set(w-1, y, focusColor)
		}
	}

	return img
}

var _ fyne.Focusable = (*OutputBarChart)(nil)
