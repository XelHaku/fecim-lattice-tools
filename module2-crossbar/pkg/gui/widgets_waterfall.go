// Package gui provides custom widgets for crossbar visualization.
package gui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	sharedutils "multilayer-ferroelectric-cim-visualizer/shared/utils"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// AccuracyWaterfall displays stepwise accuracy degradation.
type AccuracyWaterfall struct {
	widget.BaseWidget

	steps          []WaterfallStep
	targetAccuracy float64

	raster *canvas.Raster

	// Y-axis labels (0%, 20%, 40%, 60%, 80%, 100%)
	yAxisLabels []*canvas.Text
	// Bar value labels (accuracy % above each bar)
	barLabels []*canvas.Text
	// Step labels (X-axis labels)
	stepLabels []*canvas.Text
	// Target label
	targetLabel *canvas.Text
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
		steps:          []WaterfallStep{},
		targetAccuracy: 87.0, // Dr. Tour's reported 87%
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetSteps updates the waterfall steps.
func (w *AccuracyWaterfall) SetSteps(steps []WaterfallStep) {
	w.steps = steps
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		w.Refresh()
	})
}

// SetTarget sets the target accuracy line.
func (w *AccuracyWaterfall) SetTarget(target float64) {
	w.targetAccuracy = target
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		w.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (w *AccuracyWaterfall) CreateRenderer() fyne.WidgetRenderer {
	w.raster = canvas.NewRaster(w.generateImage)

	// Create Y-axis labels (0%, 20%, 40%, 60%, 80%, 100%)
	labelColor := color.RGBA{180, 180, 180, 255}
	w.yAxisLabels = make([]*canvas.Text, 6)
	for i := 0; i <= 5; i++ {
		pct := i * 20
		label := canvas.NewText(fmt.Sprintf("%d%%", pct), labelColor)
		label.TextSize = 10
		label.Alignment = fyne.TextAlignTrailing
		w.yAxisLabels[i] = label
	}

	// Create target label
	w.targetLabel = canvas.NewText(fmt.Sprintf("Target: %.0f%%", w.targetAccuracy), color.RGBA{255, 200, 0, 255})
	w.targetLabel.TextSize = 10
	w.targetLabel.Alignment = fyne.TextAlignLeading

	// Bar and step labels will be created/updated when data is set
	w.barLabels = make([]*canvas.Text, 0)
	w.stepLabels = make([]*canvas.Text, 0)

	return &waterfallRenderer{
		waterfall: w,
	}
}

// MinSize returns minimum size.
func (w *AccuracyWaterfall) MinSize() fyne.Size {
	return fyne.NewSize(350, 250)
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
	for i := 0; i <= 5; i++ {
		acc := minAcc + float64(i)*(maxAcc-minAcc)/5.0
		y := marginTop + chartHeight - int(float64(chartHeight)*(acc-minAcc)/(maxAcc-minAcc))

		// Grid line
		for x := marginLeft; x < marginLeft+chartWidth; x++ {
			img.Set(x, y, gridColor)
		}
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

	// Draw waterfall bars - proper step-down visualization
	stepWidth := chartWidth / (len(w.steps) + 1)
	barWidth := stepWidth - 20
	if barWidth < 20 {
		barWidth = 20
	}

	for i, step := range w.steps {
		x := marginLeft + (i+1)*stepWidth - barWidth/2

		// Calculate Y positions
		currentY := marginTop + chartHeight - int(float64(chartHeight)*(step.Accuracy-minAcc)/(maxAcc-minAcc))

		// Get previous accuracy (100% for first step)
		prevAcc := 100.0
		if i > 0 {
			prevAcc = w.steps[i-1].Accuracy
		}
		prevY := marginTop + chartHeight - int(float64(chartHeight)*(prevAcc-minAcc)/(maxAcc-minAcc))

		// Draw the "remaining" bar (from current accuracy to bottom) in muted color
		mutedColor := color.RGBA{step.Color.R / 3, step.Color.G / 3, step.Color.B / 3, 255}
		for bx := x; bx < x+barWidth && bx < width-marginRight; bx++ {
			for by := currentY; by < marginTop+chartHeight; by++ {
				img.Set(bx, by, mutedColor)
			}
		}

		// Draw the "loss" portion (from previous to current) in bright color
		if step.Loss > 0 {
			for bx := x; bx < x+barWidth && bx < width-marginRight; bx++ {
				for by := prevY; by < currentY; by++ {
					img.Set(bx, by, step.Color)
				}
			}
		} else {
			// First bar - draw full bar to current accuracy
			for bx := x; bx < x+barWidth && bx < width-marginRight; bx++ {
				for by := currentY; by < marginTop+chartHeight; by++ {
					img.Set(bx, by, step.Color)
				}
			}
		}

		// Draw connector line from previous bar to this bar's top
		if i > 0 {
			connectorColor := color.RGBA{150, 150, 150, 200}
			prevX := marginLeft + i*stepWidth + barWidth/2
			for cx := prevX; cx < x; cx++ {
				img.Set(cx, prevY, connectorColor)
				img.Set(cx, prevY+1, connectorColor)
			}
		}

		// Draw border
		borderColor := color.RGBA{200, 200, 200, 255}
		for bx := x; bx < x+barWidth && bx < width-marginRight; bx++ {
			img.Set(bx, currentY, borderColor)
			img.Set(bx, marginTop+chartHeight-1, borderColor)
		}
		for by := currentY; by < marginTop+chartHeight; by++ {
			img.Set(x, by, borderColor)
			if x+barWidth-1 < width-marginRight {
				img.Set(x+barWidth-1, by, borderColor)
			}
		}

		// Draw accuracy value text (simplified as small marker)
		markerColor := color.RGBA{255, 255, 255, 255}
		for dx := -2; dx <= 2; dx++ {
			img.Set(x+barWidth/2+dx, currentY-3, markerColor)
		}
	}

	// Axis labels with units
	labelColor := color.RGBA{180, 180, 180, 255}
	// Y-axis label (rotated text would be ideal, but using horizontal for now)
	sharedutils.DrawSimpleText(img, "Accuracy (%)", 5, 10, labelColor)
	// X-axis label
	xLabelX := marginLeft + chartWidth/2 - 70 // Center the label
	sharedutils.DrawSimpleText(img, "Degradation Stage", xLabelX, height-20, labelColor)

	return img
}

// waterfallRenderer is a custom renderer for AccuracyWaterfall with labels.
type waterfallRenderer struct {
	waterfall *AccuracyWaterfall
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *waterfallRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("waterfallRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.cache.MarkLayout(size)
	r.waterfall.raster.Resize(size)

	// Layout Y-axis labels
	marginLeft := float32(80)
	marginTop := float32(30)
	marginBottom := float32(40)
	chartHeight := size.Height - marginTop - marginBottom

	for i, label := range r.waterfall.yAxisLabels {
		pct := float64(i) * 20.0
		y := marginTop + chartHeight*(1.0-float32(pct)/100.0) - 5
		label.Move(fyne.NewPos(5, y))
	}

	// Layout target label
	if r.waterfall.targetAccuracy > 0 {
		targetY := marginTop + chartHeight*(1.0-float32(r.waterfall.targetAccuracy)/100.0) - 5
		r.waterfall.targetLabel.Move(fyne.NewPos(marginLeft+5, targetY-12))
	}

	// Layout bar value labels
	if len(r.waterfall.steps) > 0 && len(r.waterfall.barLabels) == len(r.waterfall.steps) {
		marginRight := float32(20)
		chartWidth := size.Width - marginLeft - marginRight
		stepWidth := chartWidth / float32(len(r.waterfall.steps)+1)

		for i, label := range r.waterfall.barLabels {
			x := marginLeft + float32(i+1)*stepWidth - 15
			y := marginTop + chartHeight*(1.0-float32(r.waterfall.steps[i].Accuracy)/100.0) - 18
			if y < marginTop-10 {
				y = marginTop - 10
			}
			label.Move(fyne.NewPos(x, y))
		}
	}

	// Layout step labels
	if len(r.waterfall.steps) > 0 && len(r.waterfall.stepLabels) == len(r.waterfall.steps) {
		marginRight := float32(20)
		chartWidth := size.Width - marginLeft - marginRight
		stepWidth := chartWidth / float32(len(r.waterfall.steps)+1)

		for i, label := range r.waterfall.stepLabels {
			x := marginLeft + float32(i+1)*stepWidth - 25
			y := size.Height - marginBottom + 5
			label.Move(fyne.NewPos(x, y))
		}
	}
}

func (r *waterfallRenderer) MinSize() fyne.Size {
	return r.waterfall.MinSize()
}

func (r *waterfallRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("waterfallRenderer", r.waterfall.Size())
	// Update bar labels based on current steps
	w := r.waterfall

	// Recreate bar labels if needed
	if len(w.barLabels) != len(w.steps) {
		w.barLabels = make([]*canvas.Text, len(w.steps))
		for i, step := range w.steps {
			label := canvas.NewText(fmt.Sprintf("%.1f%%", step.Accuracy), color.White)
			label.TextSize = 10
			label.Alignment = fyne.TextAlignCenter
			w.barLabels[i] = label
		}
	} else {
		for i, step := range w.steps {
			w.barLabels[i].Text = fmt.Sprintf("%.1f%%", step.Accuracy)
		}
	}

	// Recreate step labels if needed
	if len(w.stepLabels) != len(w.steps) {
		w.stepLabels = make([]*canvas.Text, len(w.steps))
		for i, step := range w.steps {
			// Use full label name without truncation
			label := canvas.NewText(step.Label, color.RGBA{150, 150, 150, 255})
			label.TextSize = 10
			label.Alignment = fyne.TextAlignCenter
			w.stepLabels[i] = label
		}
	}

	// Update target label
	w.targetLabel.Text = fmt.Sprintf("Target: %.0f%%", w.targetAccuracy)

	w.raster.Refresh()
}

func (r *waterfallRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{r.waterfall.raster}

	// Add Y-axis labels
	for _, label := range r.waterfall.yAxisLabels {
		objects = append(objects, label)
	}

	// Add target label
	objects = append(objects, r.waterfall.targetLabel)

	// Add bar labels
	for _, label := range r.waterfall.barLabels {
		objects = append(objects, label)
	}

	// Add step labels
	for _, label := range r.waterfall.stepLabels {
		objects = append(objects, label)
	}

	return objects
}

func (r *waterfallRenderer) Destroy() {}
