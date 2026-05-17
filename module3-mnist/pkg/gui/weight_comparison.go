//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for MNIST visualization.
// weight_comparison.go implements FP vs Quantized weight visualization
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// WeightComparisonWidget shows FP and Quantized weights side-by-side with error visualization.
type WeightComparisonWidget struct {
	widget.BaseWidget

	mu sync.RWMutex

	// Weight data
	fpWeights    [][]float64
	quantWeights [][]float64
	wMin, wMax   float64

	// Display mode
	showMode   int // 0=FP, 1=Quantized, 2=Difference
	layerIndex int // 0=Layer1, 1=Layer2

	// Statistics
	meanError   float64
	maxError    float64
	errorStdDev float64

	// Visual components
	titleLabel *widget.Label
	statsLabel *widget.Label
	raster     *canvas.Raster
}

// NewWeightComparisonWidget creates a new weight comparison widget.
func NewWeightComparisonWidget() *WeightComparisonWidget {
	wcw := &WeightComparisonWidget{
		showMode:   2, // Default to difference view
		layerIndex: 0,
	}
	wcw.titleLabel = widget.NewLabelWithStyle("Weight Comparison (FP vs Quantized)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	wcw.statsLabel = widget.NewLabel("Load weights to see comparison")
	wcw.ExtendBaseWidget(wcw)
	return wcw
}

// SetWeights updates both FP and quantized weight arrays.
func (wcw *WeightComparisonWidget) SetWeights(fpWeights, quantWeights [][]float64) {
	wcw.mu.Lock()
	defer wcw.mu.Unlock()

	wcw.fpWeights = fpWeights
	wcw.quantWeights = quantWeights

	if len(fpWeights) == 0 || len(fpWeights[0]) == 0 {
		return
	}

	// Calculate statistics
	wcw.calculateStats()

	if os.Getenv("FECIM_FYNE_TEST") == "1" {
		// Avoid fyne.Do() in the test driver (no render loop); stats correctness is validated elsewhere.
		wcw.updateStatsLabel()
		return
	}

	fyne.Do(func() {
		wcw.updateStatsLabel()
		wcw.Refresh()
	})
}

// SetShowMode changes display mode: 0=FP, 1=Quantized, 2=Difference
func (wcw *WeightComparisonWidget) SetShowMode(mode int) {
	wcw.mu.Lock()
	wcw.showMode = mode
	wcw.mu.Unlock()
	fyne.Do(func() {
		wcw.Refresh()
	})
}

// SetLayerIndex changes which layer to display.
func (wcw *WeightComparisonWidget) SetLayerIndex(index int) {
	wcw.mu.Lock()
	wcw.layerIndex = index
	wcw.mu.Unlock()
	fyne.Do(func() {
		wcw.Refresh()
	})
}

// calculateStats computes error statistics between FP and quantized weights.
func (wcw *WeightComparisonWidget) calculateStats() {
	if len(wcw.fpWeights) == 0 || len(wcw.quantWeights) == 0 {
		return
	}

	var sumError, sumSqError float64
	wcw.maxError = 0
	wcw.wMin, wcw.wMax = wcw.fpWeights[0][0], wcw.fpWeights[0][0]
	count := 0

	rows := len(wcw.fpWeights)
	cols := len(wcw.fpWeights[0])

	for i := 0; i < rows && i < len(wcw.quantWeights); i++ {
		for j := 0; j < cols && j < len(wcw.quantWeights[i]); j++ {
			fp := wcw.fpWeights[i][j]
			quant := wcw.quantWeights[i][j]

			// Track range
			if fp < wcw.wMin {
				wcw.wMin = fp
			}
			if fp > wcw.wMax {
				wcw.wMax = fp
			}

			// Calculate error
			err := math.Abs(fp - quant)
			sumError += err
			sumSqError += err * err
			if err > wcw.maxError {
				wcw.maxError = err
			}
			count++
		}
	}

	if count > 0 {
		wcw.meanError = sumError / float64(count)
		variance := sumSqError/float64(count) - wcw.meanError*wcw.meanError
		if variance > 0 {
			wcw.errorStdDev = math.Sqrt(variance)
		}
	}
}

func (wcw *WeightComparisonWidget) updateStatsLabel() {
	modeNames := []string{"FP (Float32)", "Quantized (30 Levels, claim)", "Difference (Error)"}
	modeName := modeNames[wcw.showMode]

	if len(wcw.fpWeights) == 0 {
		wcw.statsLabel.SetText("No weights loaded")
		return
	}

	rows := len(wcw.fpWeights)
	cols := 0
	if rows > 0 {
		cols = len(wcw.fpWeights[0])
	}

	// MED-003 fix: Add context about what error means
	wRange := wcw.wMax - wcw.wMin
	if wRange == 0 {
		wRange = 1
	}
	errorPctOfRange := (wcw.meanError / wRange) * 100

	wcw.statsLabel.SetText(fmt.Sprintf("%s | %dx%d | Mean Error: %.4f (~%.1f%% of weight range) | Max: %.4f | 30 levels (conference-claim baseline) ≈ near-ideal accuracy",
		modeName, rows, cols, wcw.meanError, errorPctOfRange, wcw.maxError))
}

// MinSize returns the minimum size for the widget.
func (wcw *WeightComparisonWidget) MinSize() fyne.Size {
	return fyne.NewSize(400, 250)
}

// CreateRenderer implements fyne.Widget.
func (wcw *WeightComparisonWidget) CreateRenderer() fyne.WidgetRenderer {
	wcw.raster = canvas.NewRaster(wcw.generateImage)

	// Mode selector
	modeSelect := widget.NewSelect([]string{"FP (Float32)", "Quantized", "Difference"}, func(s string) {
		switch s {
		case "FP (Float32)":
			wcw.SetShowMode(0)
		case "Quantized":
			wcw.SetShowMode(1)
		case "Difference":
			wcw.SetShowMode(2)
		}
	})
	modeSelect.SetSelected("Difference")

	controlRow := container.NewHBox(
		widget.NewLabel("View:"),
		modeSelect,
	)

	content := container.NewBorder(
		container.NewVBox(
			wcw.titleLabel,
			controlRow,
			widget.NewSeparator(),
		),
		wcw.statsLabel,
		nil, nil,
		container.NewMax(wcw.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage creates the weight visualization.
func (wcw *WeightComparisonWidget) generateImage(w, h int) image.Image {
	if w < 10 {
		w = 400
	}
	if h < 10 {
		h = 180
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 30, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	wcw.mu.RLock()
	fpWeights := wcw.fpWeights
	quantWeights := wcw.quantWeights
	showMode := wcw.showMode
	wMin, wMax := wcw.wMin, wcw.wMax
	maxError := wcw.maxError
	wcw.mu.RUnlock()

	if len(fpWeights) == 0 || len(fpWeights[0]) == 0 {
		drawSimpleText(img, "No weights loaded", w/2-60, h/2-8, color.RGBA{100, 100, 120, 255})
		return img
	}

	rows := len(fpWeights)
	cols := len(fpWeights[0])

	// Calculate cell size
	cellW := float64(w) / float64(cols)
	cellH := float64(h) / float64(rows)

	wRange := wMax - wMin
	if wRange == 0 {
		wRange = 1
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Map pixel to weight matrix
			col := int(float64(x) / cellW)
			row := int(float64(y) / cellH)

			if row >= rows {
				row = rows - 1
			}
			if col >= cols {
				col = cols - 1
			}

			var c color.RGBA
			switch showMode {
			case 0: // FP weights
				val := fpWeights[row][col]
				normalized := (val - wMin) / wRange
				c = wcw.blueWhiteRedColor(normalized)
			case 1: // Quantized weights
				if row < len(quantWeights) && col < len(quantWeights[row]) {
					val := quantWeights[row][col]
					normalized := (val - wMin) / wRange
					c = wcw.blueWhiteRedColor(normalized)
				}
			case 2: // Difference (error)
				if row < len(quantWeights) && col < len(quantWeights[row]) {
					err := math.Abs(fpWeights[row][col] - quantWeights[row][col])
					// Normalize error - use black to red gradient
					normalized := 0.0
					if maxError > 0 {
						normalized = err / maxError
					}
					c = wcw.errorColor(normalized)
				}
			}

			img.Set(x, y, c)
		}
	}

	return img
}

// blueWhiteRedColor returns a blue-white-red color for normalized value [0,1].
func (wcw *WeightComparisonWidget) blueWhiteRedColor(normalized float64) color.RGBA {
	if normalized < 0.5 {
		// Blue to white
		t := normalized * 2
		r := uint8(t * 255)
		g := uint8(t * 255)
		b := uint8(255)
		return color.RGBA{r, g, b, 255}
	}
	// White to red
	t := (normalized - 0.5) * 2
	r := uint8(255)
	g := uint8((1 - t) * 255)
	b := uint8((1 - t) * 255)
	return color.RGBA{r, g, b, 255}
}

// errorColor returns a black-to-red color for error visualization.
func (wcw *WeightComparisonWidget) errorColor(normalized float64) color.RGBA {
	// Black (no error) to Yellow (medium error) to Red (high error)
	if normalized < 0.5 {
		// Black to yellow
		t := normalized * 2
		r := uint8(t * 255)
		g := uint8(t * 200)
		b := uint8(0)
		return color.RGBA{r, g, b, 255}
	}
	// Yellow to red
	t := (normalized - 0.5) * 2
	r := uint8(255)
	g := uint8((1 - t) * 200)
	b := uint8(0)
	return color.RGBA{r, g, b, 255}
}

// DualWeightHeatmap shows FP and Quantized weights side-by-side.
type DualWeightHeatmap struct {
	widget.BaseWidget

	mu sync.RWMutex

	// Weight data for both layers
	fpW1, fpW2       [][]float64
	quantW1, quantW2 [][]float64

	// Current layer
	layerIndex int // 0 or 1

	// Visual
	raster *canvas.Raster
}

// NewDualWeightHeatmap creates a new dual weight heatmap.
func NewDualWeightHeatmap() *DualWeightHeatmap {
	dwh := &DualWeightHeatmap{}
	dwh.ExtendBaseWidget(dwh)
	return dwh
}

// SetWeights sets weights for both layers.
func (dwh *DualWeightHeatmap) SetWeights(fpW1, fpW2, quantW1, quantW2 [][]float64) {
	dwh.mu.Lock()
	dwh.fpW1 = fpW1
	dwh.fpW2 = fpW2
	dwh.quantW1 = quantW1
	dwh.quantW2 = quantW2
	dwh.mu.Unlock()
	fyne.Do(func() {
		dwh.Refresh()
	})
}

// SetLayer changes which layer to display.
func (dwh *DualWeightHeatmap) SetLayer(index int) {
	dwh.mu.Lock()
	dwh.layerIndex = index
	dwh.mu.Unlock()
	fyne.Do(func() {
		dwh.Refresh()
	})
}

// MinSize returns minimum size.
func (dwh *DualWeightHeatmap) MinSize() fyne.Size {
	return fyne.NewSize(500, 150)
}

// CreateRenderer implements fyne.Widget.
func (dwh *DualWeightHeatmap) CreateRenderer() fyne.WidgetRenderer {
	dwh.raster = canvas.NewRaster(dwh.generateImage)

	title := widget.NewLabelWithStyle("FP (left) vs Quantized (right)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewBorder(
		title,
		nil, nil, nil,
		container.NewMax(dwh.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage draws FP on left, Quantized on right.
func (dwh *DualWeightHeatmap) generateImage(w, h int) image.Image {
	if w < 20 {
		w = 500
	}
	if h < 10 {
		h = 130
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{25, 30, 45, 255})
		}
	}

	dwh.mu.RLock()
	var fpW, quantW [][]float64
	if dwh.layerIndex == 0 {
		fpW = dwh.fpW1
		quantW = dwh.quantW1
	} else {
		fpW = dwh.fpW2
		quantW = dwh.quantW2
	}
	dwh.mu.RUnlock()

	if len(fpW) == 0 {
		return img
	}

	// Split view: left half for FP, right half for quantized
	halfW := w / 2
	dividerX := halfW

	// Find global range
	wMin, wMax := fpW[0][0], fpW[0][0]
	for _, row := range fpW {
		for _, v := range row {
			if v < wMin {
				wMin = v
			}
			if v > wMax {
				wMax = v
			}
		}
	}
	wRange := wMax - wMin
	if wRange == 0 {
		wRange = 1
	}

	rows := len(fpW)
	cols := len(fpW[0])

	// Draw FP weights (left)
	dwh.drawHeatmap(img, 0, 0, halfW-2, h, fpW, rows, cols, wMin, wRange)

	// Draw divider
	for y := 0; y < h; y++ {
		img.Set(dividerX-1, y, color.RGBA{100, 100, 120, 255})
		img.Set(dividerX, y, color.RGBA{100, 100, 120, 255})
	}

	// Draw quantized weights (right)
	if len(quantW) > 0 {
		dwh.drawHeatmap(img, halfW+2, 0, halfW-2, h, quantW, rows, cols, wMin, wRange)
	}

	return img
}

func (dwh *DualWeightHeatmap) drawHeatmap(img *image.RGBA, startX, startY, w, h int, weights [][]float64, rows, cols int, wMin, wRange float64) {
	cellW := float64(w) / float64(cols)
	cellH := float64(h) / float64(rows)

	for y := startY; y < startY+h; y++ {
		for x := startX; x < startX+w; x++ {
			col := int(float64(x-startX) / cellW)
			row := int(float64(y-startY) / cellH)

			if row >= rows {
				row = rows - 1
			}
			if col >= cols {
				col = cols - 1
			}

			val := weights[row][col]
			normalized := (val - wMin) / wRange

			// Blue-white-red colormap
			var c color.RGBA
			if normalized < 0.5 {
				t := normalized * 2
				c = color.RGBA{uint8(t * 255), uint8(t * 255), 255, 255}
			} else {
				t := (normalized - 0.5) * 2
				c = color.RGBA{255, uint8((1 - t) * 255), uint8((1 - t) * 255), 255}
			}

			img.Set(x, y, c)
		}
	}
}
