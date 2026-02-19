// Package gui provides custom widgets for crossbar visualization.
// M15: Device-to-device variation modeling GUI with yield prediction.
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

	"fecim-lattice-tools/shared/crossbar"
)

// VariationModelGUI displays device-to-device variation modeling with yield prediction.
type VariationModelGUI struct {
	widget.BaseWidget

	// Configuration
	progSigma float64 // Programming error sigma (%)
	readSigma float64 // Read noise sigma (%)
	arrayRows int
	arrayCols int

	// Results
	yieldPrediction float64
	errorStats      *VariationStats

	// UI components
	progSigmaSlider *widget.Slider
	readSigmaSlider *widget.Slider
	progSigmaLabel  *widget.Label
	readSigmaLabel  *widget.Label
	yieldLabel      *widget.Label
	statsLabel      *widget.Label
	distributionViz *canvas.Raster
}

// VariationStats holds statistics about device variation.
type VariationStats struct {
	MeanError      float64
	StdError       float64
	MaxError       float64
	Cells3Sigma    float64 // % of cells within 3-sigma
	Cells6Sigma    float64 // % of cells within 6-sigma
	LevelErrorRate float64 // % of cells with level errors
	PredictedYield float64 // Manufacturing yield prediction
}

// NewVariationModelGUI creates a new variation modeling GUI widget.
func NewVariationModelGUI() *VariationModelGUI {
	w := &VariationModelGUI{
		progSigma: 5.0, // 5% programming error default
		readSigma: 1.0, // 1% read noise default
		arrayRows: 64,
		arrayCols: 64,
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetArraySize sets the array size for yield calculations.
func (w *VariationModelGUI) SetArraySize(rows, cols int) {
	w.arrayRows = rows
	w.arrayCols = cols
	w.recalculate()
}

// recalculate computes variation statistics and yield.
func (w *VariationModelGUI) recalculate() {
	totalCells := float64(w.arrayRows * w.arrayCols)

	// Combined sigma (programming + read noise add in quadrature)
	combinedSigma := math.Sqrt(w.progSigma*w.progSigma + w.readSigma*w.readSigma)

	// Statistics for Gaussian distribution
	// 3-sigma: 99.73% of values within ±3σ
	// 6-sigma: 99.99966% of values within ±6σ
	cells3Sigma := 99.73
	cells6Sigma := 99.99966

	// Level error rate estimation
	// Assuming 30 levels, level spacing = 100%/29 ≈ 3.45%
	// Error causes level change when error > level_spacing/2
	levelSpacing := 100.0 / 29.0
	threshold := levelSpacing / 2.0

	// Probability of error exceeding threshold (two-tailed)
	// P(|X| > threshold) = 2 * (1 - CDF(threshold/sigma))
	// Using approximation: erfc(x/sqrt(2))/2
	if combinedSigma > 0 {
		z := threshold / combinedSigma
		// Approximate error function for probability calculation
		levelErrorProb := 2.0 * (1.0 - normalCDF(z))
		w.errorStats = &VariationStats{
			MeanError:      0, // Gaussian centered at 0
			StdError:       combinedSigma,
			MaxError:       3 * combinedSigma, // 3-sigma typical max
			Cells3Sigma:    cells3Sigma,
			Cells6Sigma:    cells6Sigma,
			LevelErrorRate: levelErrorProb * 100,
		}
	} else {
		w.errorStats = &VariationStats{
			MeanError:      0,
			StdError:       0,
			MaxError:       0,
			Cells3Sigma:    100,
			Cells6Sigma:    100,
			LevelErrorRate: 0,
		}
	}

	// Yield prediction
	// Simple model: yield = (1 - defect_rate)^num_cells
	// Defect = cell with error > 3σ (catastrophic failure)
	defectRate := (100.0 - cells3Sigma) / 100.0
	w.yieldPrediction = math.Pow(1-defectRate, totalCells) * 100

	// Adjust for level errors (soft failures)
	softFailRate := w.errorStats.LevelErrorRate / 100.0
	functionalYield := math.Pow(1-softFailRate*0.1, totalCells) * 100 // 10% weight for soft fails
	w.errorStats.PredictedYield = math.Min(w.yieldPrediction, functionalYield)
}

// normalCDF approximates the standard normal CDF.
func normalCDF(x float64) float64 {
	// Approximation using error function
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// CreateRenderer implements fyne.Widget.
func (w *VariationModelGUI) CreateRenderer() fyne.WidgetRenderer {
	// Title
	title := widget.NewLabelWithStyle(
		"Device Variation Modeling",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Programming error slider
	w.progSigmaLabel = widget.NewLabel(fmt.Sprintf("Programming σ: %.1f%%", w.progSigma))
	w.progSigmaSlider = widget.NewSlider(0, 20)
	w.progSigmaSlider.Value = w.progSigma
	w.progSigmaSlider.Step = 0.5
	w.progSigmaSlider.OnChanged = func(v float64) {
		w.progSigma = v
		w.progSigmaLabel.SetText(fmt.Sprintf("Programming σ: %.1f%%", v))
		w.recalculate()
		w.updateDisplay()
	}

	// Read noise slider
	w.readSigmaLabel = widget.NewLabel(fmt.Sprintf("Read noise σ: %.1f%%", w.readSigma))
	w.readSigmaSlider = widget.NewSlider(0, 10)
	w.readSigmaSlider.Value = w.readSigma
	w.readSigmaSlider.Step = 0.1
	w.readSigmaSlider.OnChanged = func(v float64) {
		w.readSigma = v
		w.readSigmaLabel.SetText(fmt.Sprintf("Read noise σ: %.1f%%", v))
		w.recalculate()
		w.updateDisplay()
	}

	// Results labels
	w.yieldLabel = widget.NewLabelWithStyle(
		"Predicted Yield: --",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	w.statsLabel = widget.NewLabel("Adjust sliders to see variation impact")
	w.statsLabel.Wrapping = fyne.TextWrapWord

	// Distribution visualization
	w.distributionViz = canvas.NewRaster(func(width, height int) image.Image {
		return w.generateDistributionImage(width, height)
	})
	w.distributionViz.SetMinSize(fyne.NewSize(280, 100))

	// Technology comparison
	techCompareLabel := widget.NewLabel(
		"FeCIM typical: σ_prog=3-5%, σ_read=0.5-1%\n" +
			"RRAM typical: σ_prog=10-20%, σ_read=2-5%",
	)
	techCompareLabel.TextStyle = fyne.TextStyle{Italic: true}
	techCompareLabel.Wrapping = fyne.TextWrapWord

	// Initial calculation
	w.recalculate()

	// Build content
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		w.progSigmaLabel,
		w.progSigmaSlider,
		w.readSigmaLabel,
		w.readSigmaSlider,
		widget.NewSeparator(),
		w.distributionViz,
		widget.NewSeparator(),
		w.yieldLabel,
		w.statsLabel,
		widget.NewSeparator(),
		techCompareLabel,
	)

	// Update display with initial values
	w.updateDisplay()

	return widget.NewSimpleRenderer(content)
}

// updateDisplay updates all display elements.
func (w *VariationModelGUI) updateDisplay() {
	if w.errorStats == nil {
		return
	}

	// Update yield label with color coding
	yieldText := fmt.Sprintf("Predicted Yield: %.2f%%", w.errorStats.PredictedYield)
	if w.yieldLabel != nil {
		w.yieldLabel.SetText(yieldText)
	}

	// Update stats label
	if w.statsLabel != nil {
		statsText := fmt.Sprintf(
			"Combined σ: %.2f%%\n"+
				"Level error rate: %.2f%%\n"+
				"3-sigma coverage: %.2f%%\n"+
				"Array size: %dx%d (%d cells)",
			w.errorStats.StdError,
			w.errorStats.LevelErrorRate,
			w.errorStats.Cells3Sigma,
			w.arrayRows, w.arrayCols,
			w.arrayRows*w.arrayCols,
		)
		w.statsLabel.SetText(statsText)
	}

	// Refresh distribution visualization
	if w.distributionViz != nil {
		w.distributionViz.Refresh()
	}
}

// generateDistributionImage draws a Gaussian distribution visualization.
func (w *VariationModelGUI) generateDistributionImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w.errorStats == nil || w.errorStats.StdError == 0 {
		return img
	}

	sigma := w.errorStats.StdError
	if sigma == 0 {
		sigma = 1 // Avoid division by zero
	}
	maxX := 4 * sigma // Show ±4σ range

	// Draw Gaussian curve
	curveColor := color.RGBA{0, 200, 255, 255}
	fillColor := color.RGBA{0, 100, 150, 150}

	for x := 0; x < width; x++ {
		// Map x to -4σ to +4σ
		xVal := (float64(x)/float64(width)*2 - 1) * maxX

		// Gaussian PDF
		yNorm := math.Exp(-0.5 * (xVal / sigma) * (xVal / sigma))

		// Scale to height
		yPix := height - 1 - int(yNorm*float64(height-20))
		if yPix < 0 {
			yPix = 0
		}

		// Draw fill from bottom to curve
		for py := height - 1; py >= yPix; py-- {
			img.Set(x, py, fillColor)
		}

		// Draw curve point
		img.Set(x, yPix, curveColor)
		if yPix+1 < height {
			img.Set(x, yPix+1, curveColor)
		}
	}

	// Draw ±1σ, ±2σ, ±3σ markers
	for i, sigmaMultiple := range []float64{1, 2, 3} {
		markerColor := color.RGBA{255, 255, uint8(100 + i*50), 200}
		xFrac := sigmaMultiple / maxX
		xPos := int((xFrac + 0.5) * float64(width))
		xNeg := int((-xFrac + 0.5) * float64(width))

		for y := 0; y < height; y += 3 { // Dashed line
			if xPos >= 0 && xPos < width {
				img.Set(xPos, y, markerColor)
			}
			if xNeg >= 0 && xNeg < width {
				img.Set(xNeg, y, markerColor)
			}
		}
	}

	return img
}

// MinSize returns the minimum size.
func (w *VariationModelGUI) MinSize() fyne.Size {
	return fyne.NewSize(300, 400)
}

// RunVariationAnalysis runs variation analysis with given configuration.
func (w *VariationModelGUI) RunVariationAnalysis(array *crossbar.Array) {
	if array == nil {
		return
	}

	cfg := array.GetConfig()
	w.arrayRows = cfg.Rows
	w.arrayCols = cfg.Cols
	w.recalculate()
	w.updateDisplay()
}
