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
)

// ConfusionMatrix displays an interactive confusion matrix.
type ConfusionMatrix struct {
	widget.BaseWidget

	// 10x10 matrix (actual x predicted)
	matrix [10][10]int
	total  int

	// Selection
	selectedRow int
	selectedCol int
	focused     bool

	// Rendering
	raster *canvas.Raster

	// Callback
	OnCellTapped func(actual, predicted int, count int)
}

// NewConfusionMatrix creates a new confusion matrix widget.
func NewConfusionMatrix() *ConfusionMatrix {
	cm := &ConfusionMatrix{
		selectedRow: -1,
		selectedCol: -1,
	}
	cm.ExtendBaseWidget(cm)
	return cm
}

// SetMatrix updates the confusion matrix data.
func (cm *ConfusionMatrix) SetMatrix(matrix [10][10]int) {
	cm.matrix = matrix
	cm.total = 0
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			cm.total += matrix[i][j]
		}
	}
	fyne.Do(func() {
		cm.Refresh()
	})
}

// GetAccuracy returns the overall accuracy.
func (cm *ConfusionMatrix) GetAccuracy() float64 {
	if cm.total == 0 {
		return 0
	}
	correct := 0
	for i := 0; i < 10; i++ {
		correct += cm.matrix[i][i]
	}
	return float64(correct) / float64(cm.total)
}

// GetClassMetrics returns precision, recall, and F1 for a class.
func (cm *ConfusionMatrix) GetClassMetrics(class int) (precision, recall, f1 float64) {
	if class < 0 || class >= 10 {
		return 0, 0, 0
	}

	// True positives
	tp := float64(cm.matrix[class][class])

	// False positives (predicted class but actually something else)
	fp := 0.0
	for i := 0; i < 10; i++ {
		if i != class {
			fp += float64(cm.matrix[i][class])
		}
	}

	// False negatives (actually class but predicted something else)
	fn := 0.0
	for j := 0; j < 10; j++ {
		if j != class {
			fn += float64(cm.matrix[class][j])
		}
	}

	// Precision
	if tp+fp > 0 {
		precision = tp / (tp + fp)
	}

	// Recall
	if tp+fn > 0 {
		recall = tp / (tp + fn)
	}

	// F1
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return
}

// CreateRenderer implements fyne.Widget.
func (cm *ConfusionMatrix) CreateRenderer() fyne.WidgetRenderer {
	cm.raster = canvas.NewRaster(cm.generateImage)

	titleLabel := widget.NewLabel("Confusion Matrix (Actual vs Predicted)")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	accuracyLabel := widget.NewLabel(fmt.Sprintf("Accuracy (modeled): %.1f%%", cm.GetAccuracy()*100))
	accuracyLabel.Alignment = fyne.TextAlignCenter

	// Use Max container for raster to fill available space
	content := container.NewBorder(
		titleLabel,
		accuracyLabel,
		nil,
		nil,
		container.NewMax(cm.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size - smaller for flexible layout.
func (cm *ConfusionMatrix) MinSize() fyne.Size {
	return fyne.NewSize(250, 250)
}

// Tapped handles tap events.
func (cm *ConfusionMatrix) Tapped(e *fyne.PointEvent) {
	size := cm.Size()
	padding := float32(30)
	cellSize := (size.Width - 2*padding) / 10

	// Calculate cell
	col := int((e.Position.X - padding) / cellSize)
	row := int((e.Position.Y - padding) / cellSize)

	if row >= 0 && row < 10 && col >= 0 && col < 10 {
		cm.selectCell(row, col, true)
	}
}

func (cm *ConfusionMatrix) selectCell(row, col int, triggerCallback bool) {
	if row < 0 || row > 9 || col < 0 || col > 9 {
		return
	}
	cm.selectedRow = row
	cm.selectedCol = col
	fyne.Do(func() {
		cm.Refresh()
	})
	if triggerCallback && cm.OnCellTapped != nil {
		cm.OnCellTapped(row, col, cm.matrix[row][col])
	}
}

// FocusGained implements fyne.Focusable.
func (cm *ConfusionMatrix) FocusGained() {
	cm.focused = true
	if cm.selectedRow < 0 || cm.selectedCol < 0 {
		cm.selectedRow, cm.selectedCol = 0, 0
	}
	cm.Refresh()
}

// FocusLost implements fyne.Focusable.
func (cm *ConfusionMatrix) FocusLost() {
	cm.focused = false
	cm.Refresh()
}

// TypedRune implements fyne.Focusable.
func (cm *ConfusionMatrix) TypedRune(_ rune) {}

// TypedKey enables keyboard navigation with arrow keys.
func (cm *ConfusionMatrix) TypedKey(ev *fyne.KeyEvent) {
	if ev == nil {
		return
	}
	if cm.selectedRow < 0 || cm.selectedCol < 0 {
		cm.selectedRow, cm.selectedCol = 0, 0
	}
	row, col := cm.selectedRow, cm.selectedCol
	switch ev.Name {
	case fyne.KeyLeft:
		if col > 0 {
			col--
		}
	case fyne.KeyRight:
		if col < 9 {
			col++
		}
	case fyne.KeyUp:
		if row > 0 {
			row--
		}
	case fyne.KeyDown:
		if row < 9 {
			row++
		}
	case fyne.KeyHome:
		row, col = 0, 0
	case fyne.KeyEnd:
		row, col = 9, 9
	case fyne.KeyReturn, fyne.KeySpace:
		cm.selectCell(cm.selectedRow, cm.selectedCol, true)
		return
	default:
		return
	}
	cm.selectCell(row, col, false)
}

var _ fyne.Tappable = (*ConfusionMatrix)(nil)
var _ fyne.Focusable = (*ConfusionMatrix)(nil)

// generateImage creates the confusion matrix image.
func (cm *ConfusionMatrix) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	padding := 30
	cellSize := (w - 2*padding) / 10

	// Find max for color scaling
	maxVal := 1
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if cm.matrix[i][j] > maxVal {
				maxVal = cm.matrix[i][j]
			}
		}
	}

	// Draw cells
	for row := 0; row < 10; row++ {
		for col := 0; col < 10; col++ {
			value := cm.matrix[row][col]
			intensity := float64(value) / float64(maxVal)

			x0 := padding + col*cellSize
			y0 := padding + row*cellSize
			x1 := x0 + cellSize - 1
			y1 := y0 + cellSize - 1

			// Color: green for diagonal, red-orange for off-diagonal
			var c color.RGBA
			if row == col {
				// Diagonal - correct predictions (green)
				g := uint8(50 + intensity*205)
				c = color.RGBA{30, g, 80, 255}
			} else {
				// Off-diagonal - errors (red-orange)
				r := uint8(50 + intensity*200)
				c = color.RGBA{r, 40, 40, 255}
			}

			// Draw cell
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					if x >= 0 && x < w && y >= 0 && y < h {
						img.Set(x, y, c)
					}
				}
			}

			// Highlight selected cell
			if row == cm.selectedRow && col == cm.selectedCol {
				borderColor := color.RGBA{255, 255, 100, 255}
				for x := x0; x < x1; x++ {
					if y0 >= 0 && y0 < h {
						img.Set(x, y0, borderColor)
					}
					if y1-1 >= 0 && y1-1 < h {
						img.Set(x, y1-1, borderColor)
					}
				}
				for y := y0; y < y1; y++ {
					if x0 >= 0 && x0 < w {
						img.Set(x0, y, borderColor)
					}
					if x1-1 >= 0 && x1-1 < w {
						img.Set(x1-1, y, borderColor)
					}
				}
			}
		}
	}

	// Draw grid lines
	// ACCESSIBILITY FIX: Increased contrast from (60,60,70) to (85,90,100) - 3.5:1 ratio
	gridColor := color.RGBA{85, 90, 100, 255}
	for i := 0; i <= 10; i++ {
		x := padding + i*cellSize
		y := padding + i*cellSize
		for j := padding; j < h-padding; j++ {
			if x >= 0 && x < w {
				img.Set(x, j, gridColor)
			}
		}
		for j := padding; j < w-padding; j++ {
			if y >= 0 && y < h {
				img.Set(j, y, gridColor)
			}
		}
	}

	return img
}

// MetricsPanel displays per-class precision, recall, and F1 scores.
type MetricsPanel struct {
	widget.BaseWidget

	// Per-class metrics
	precision [10]float64
	recall    [10]float64
	f1        [10]float64

	// Overall metrics
	accuracy float64
	avgF1    float64

	// Rendering
	raster *canvas.Raster
}

// NewMetricsPanel creates a new metrics panel.
func NewMetricsPanel() *MetricsPanel {
	mp := &MetricsPanel{}
	mp.ExtendBaseWidget(mp)
	return mp
}

// SetMetrics updates all metrics.
func (mp *MetricsPanel) SetMetrics(precision, recall, f1 [10]float64, accuracy float64) {
	mp.precision = precision
	mp.recall = recall
	mp.f1 = f1
	mp.accuracy = accuracy

	// Calculate average F1
	sum := 0.0
	for _, v := range f1 {
		sum += v
	}
	mp.avgF1 = sum / 10.0

	fyne.Do(func() {
		mp.Refresh()
	})
}

// SetFromConfusionMatrix calculates metrics from a confusion matrix.
func (mp *MetricsPanel) SetFromConfusionMatrix(cm *ConfusionMatrix) {
	var precision, recall, f1 [10]float64
	for i := 0; i < 10; i++ {
		precision[i], recall[i], f1[i] = cm.GetClassMetrics(i)
	}
	mp.SetMetrics(precision, recall, f1, cm.GetAccuracy())
}

// CreateRenderer implements fyne.Widget.
func (mp *MetricsPanel) CreateRenderer() fyne.WidgetRenderer {
	mp.raster = canvas.NewRaster(mp.generateImage)

	titleLabel := widget.NewLabel("Per-Class Metrics (Precision/Recall/F1)")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	summaryLabel := widget.NewLabel(fmt.Sprintf(
		"Accuracy (modeled): %.1f%% | Avg F1: %.3f",
		mp.accuracy*100, mp.avgF1,
	))
	summaryLabel.Alignment = fyne.TextAlignCenter

	// Use Max container for raster to fill available space
	content := container.NewBorder(
		titleLabel,
		summaryLabel,
		nil,
		nil,
		container.NewMax(mp.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size - smaller for flexible layout.
func (mp *MetricsPanel) MinSize() fyne.Size {
	return fyne.NewSize(300, 150)
}

// generateImage creates the metrics visualization.
func (mp *MetricsPanel) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	padding := 40
	barHeight := (h - 2*padding) / 10
	maxBarWidth := w - 2*padding - 60 // Leave space for labels

	// Colors for each metric
	precColor := color.RGBA{100, 180, 255, 255} // Blue
	recColor := color.RGBA{255, 180, 100, 255}  // Orange
	f1Color := color.RGBA{100, 255, 180, 255}   // Green

	for i := 0; i < 10; i++ {
		y0 := padding + i*barHeight
		barThickness := (barHeight - 4) / 3

		// Draw precision bar
		precWidth := int(mp.precision[i] * float64(maxBarWidth))
		for y := y0; y < y0+barThickness; y++ {
			for x := padding + 20; x < padding+20+precWidth; x++ {
				if x < w && y < h {
					img.Set(x, y, precColor)
				}
			}
		}

		// Draw recall bar
		recWidth := int(mp.recall[i] * float64(maxBarWidth))
		for y := y0 + barThickness + 1; y < y0+2*barThickness+1; y++ {
			for x := padding + 20; x < padding+20+recWidth; x++ {
				if x < w && y < h {
					img.Set(x, y, recColor)
				}
			}
		}

		// Draw F1 bar
		f1Width := int(mp.f1[i] * float64(maxBarWidth))
		for y := y0 + 2*barThickness + 2; y < y0+3*barThickness+2; y++ {
			for x := padding + 20; x < padding+20+f1Width; x++ {
				if x < w && y < h {
					img.Set(x, y, f1Color)
				}
			}
		}

		// Draw class number indicator
		classY := y0 + barHeight/2
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= 3 {
					img.Set(padding+8+dx, classY+dy, color.RGBA{200, 200, 200, 255})
				}
			}
		}
	}

	// Draw legend at bottom
	legendY := h - 15
	legendColors := []color.RGBA{precColor, recColor, f1Color}
	legendX := []int{padding, padding + 100, padding + 200}

	for i, c := range legendColors {
		for dx := 0; dx < 15; dx++ {
			for dy := -2; dy <= 2; dy++ {
				img.Set(legendX[i]+dx, legendY+dy, c)
			}
		}
	}

	return img
}

// ClassStatsPanel shows detailed stats for a selected class.
type ClassStatsPanel struct {
	widget.BaseWidget

	selectedClass int
	precision     float64
	recall        float64
	f1            float64
	support       int
	tp, fp, fn    int

	textLabel *widget.Label
}

// NewClassStatsPanel creates a new class stats panel.
func NewClassStatsPanel() *ClassStatsPanel {
	csp := &ClassStatsPanel{
		selectedClass: -1,
	}
	csp.ExtendBaseWidget(csp)
	return csp
}

// SetClass updates the displayed class statistics.
func (csp *ClassStatsPanel) SetClass(class int, precision, recall, f1 float64, tp, fp, fn, support int) {
	csp.selectedClass = class
	csp.precision = precision
	csp.recall = recall
	csp.f1 = f1
	csp.tp = tp
	csp.fp = fp
	csp.fn = fn
	csp.support = support
	fyne.Do(func() {
		csp.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (csp *ClassStatsPanel) CreateRenderer() fyne.WidgetRenderer {
	titleLabel := widget.NewLabel("Class Statistics")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	csp.textLabel = widget.NewLabel("Select a class to view details")
	csp.textLabel.TextStyle = fyne.TextStyle{Monospace: false}

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		csp.textLabel,
	)

	return widget.NewSimpleRenderer(content)
}

// Refresh updates the display.
func (csp *ClassStatsPanel) Refresh() {
	if csp.textLabel == nil {
		return
	}

	if csp.selectedClass < 0 {
		csp.textLabel.SetText("Select a class to view details")
	} else {
		csp.textLabel.SetText(fmt.Sprintf(
			"Class: %d\n\n"+
				"Precision: %.3f (%.1f%%)\n"+
				"Recall:    %.3f (%.1f%%)\n"+
				"F1 Score:  %.3f\n\n"+
				"True Positives:  %d\n"+
				"False Positives: %d\n"+
				"False Negatives: %d\n"+
				"Support: %d",
			csp.selectedClass,
			csp.precision, csp.precision*100,
			csp.recall, csp.recall*100,
			csp.f1,
			csp.tp, csp.fp, csp.fn, csp.support,
		))
	}

	csp.BaseWidget.Refresh()
}
