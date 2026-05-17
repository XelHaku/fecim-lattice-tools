//go:build legacy_fyne

// Package gui provides custom widgets for crossbar visualization.
package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BeforeAfterToggle shows side-by-side comparison of ideal vs actual.
type BeforeAfterToggle struct {
	widget.BaseWidget

	idealData  [][]float64
	actualData [][]float64
	mode       string // "split", "before", "after", "diff"

	leftHeatmap  *CrossbarHeatmap
	rightHeatmap *CrossbarHeatmap
	toggleGroup  *widget.RadioGroup

	// Statistical metrics display
	statsLabel   *widget.Label
	leftLabel    *widget.Label
	rightLabel   *widget.Label
	legendLabel  *widget.Label
	legendWidget fyne.CanvasObject

	// Callbacks for cell interactions
	// isIdeal indicates if the tap/hover is on the ideal (left) or actual (right) heatmap
	OnCellTapped func(row, col int, isIdeal bool)
	OnCellHover  func(row, col int, value float64, isIdeal bool)
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

	// Wire up cell tap callbacks
	b.leftHeatmap.OnCellTapped = func(row, col int) {
		if b.OnCellTapped != nil {
			b.OnCellTapped(row, col, true) // true = ideal/left
		}
	}
	b.rightHeatmap.OnCellTapped = func(row, col int) {
		if b.OnCellTapped != nil {
			b.OnCellTapped(row, col, false) // false = actual/right
		}
	}

	// Wire up cell hover callbacks
	b.leftHeatmap.OnCellHover = func(row, col int, value float64) {
		if b.OnCellHover != nil {
			b.OnCellHover(row, col, value, true)
		}
	}
	b.rightHeatmap.OnCellHover = func(row, col int, value float64) {
		if b.OnCellHover != nil {
			b.OnCellHover(row, col, value, false)
		}
	}

	b.ExtendBaseWidget(b)
	return b
}

// SetData updates the comparison data.
func (b *BeforeAfterToggle) SetData(ideal, actual [][]float64) {
	b.idealData = ideal
	b.actualData = actual
	b.updateDisplay()
	b.updateStatsLabel()
}

// SetMode changes the display mode.
func (b *BeforeAfterToggle) SetMode(mode string) {
	b.mode = mode
	b.updateDisplay()
	b.updateLegend()
}

// SetDimensions changes the dimensions of both internal heatmaps and reinitializes data.
func (b *BeforeAfterToggle) SetDimensions(rows, cols int) {
	b.leftHeatmap.SetDimensions(rows, cols)
	b.rightHeatmap.SetDimensions(rows, cols)
	// Clear cached data since dimensions changed
	b.idealData = nil
	b.actualData = nil
}

// updateDisplay refreshes the heatmaps based on current mode.
func (b *BeforeAfterToggle) updateDisplay() {
	if b.idealData == nil || b.actualData == nil {
		return
	}

	switch b.mode {
	case "split":
		b.leftHeatmap.SetColormap("fecim")
		b.rightHeatmap.SetColormap("fecim")
		b.leftHeatmap.SetData(b.idealData)
		b.rightHeatmap.SetData(b.actualData)
		if b.leftLabel != nil {
			b.leftLabel.SetText("Ideal (No Non-Idealities)")
		}
		if b.rightLabel != nil {
			b.rightLabel.SetText("Actual (With Non-Idealities)")
		}
	case "before":
		b.leftHeatmap.SetColormap("fecim")
		b.rightHeatmap.SetColormap("fecim")
		b.leftHeatmap.SetData(b.idealData)
		b.rightHeatmap.SetData(b.idealData)
		if b.leftLabel != nil {
			b.leftLabel.SetText("Ideal (No Non-Idealities)")
		}
		if b.rightLabel != nil {
			b.rightLabel.SetText("Ideal (No Non-Idealities)")
		}
	case "after":
		b.leftHeatmap.SetColormap("fecim")
		b.rightHeatmap.SetColormap("fecim")
		b.leftHeatmap.SetData(b.actualData)
		b.rightHeatmap.SetData(b.actualData)
		if b.leftLabel != nil {
			b.leftLabel.SetText("Actual (With Non-Idealities)")
		}
		if b.rightLabel != nil {
			b.rightLabel.SetText("Actual (With Non-Idealities)")
		}
	case "diff":
		// Use diverging colormap (blue-white-red) for difference view
		b.leftHeatmap.SetColormap("diverging")
		b.rightHeatmap.SetColormap("diverging")
		diff := b.computeSignedDifference()
		b.leftHeatmap.SetData(diff)
		b.rightHeatmap.SetData(diff)
		if b.leftLabel != nil {
			b.leftLabel.SetText("Difference (Actual - Ideal)")
		}
		if b.rightLabel != nil {
			b.rightLabel.SetText("Difference (Actual - Ideal)")
		}
	}
}

// computeDifference computes the absolute difference map.
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

// computeSignedDifference computes the signed difference map for diverging colormap.
// Returns values in range [-1, 1] where negative means ideal > actual, positive means actual > ideal.
func (b *BeforeAfterToggle) computeSignedDifference() [][]float64 {
	rows := len(b.idealData)
	if rows == 0 {
		return nil
	}
	cols := len(b.idealData[0])

	// First pass: find max absolute difference for normalization
	maxAbsDiff := 0.0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols && j < len(b.idealData[i]) && j < len(b.actualData[i]); j++ {
			absDiff := math.Abs(b.actualData[i][j] - b.idealData[i][j])
			if absDiff > maxAbsDiff {
				maxAbsDiff = absDiff
			}
		}
	}

	// Avoid division by zero
	if maxAbsDiff == 0 {
		maxAbsDiff = 1
	}

	// Second pass: compute normalized signed differences
	diff := make([][]float64, rows)
	for i := range diff {
		diff[i] = make([]float64, cols)
		for j := range diff[i] {
			if j < len(b.idealData[i]) && j < len(b.actualData[i]) {
				// Normalize to [-1, 1] range
				diff[i][j] = (b.actualData[i][j] - b.idealData[i][j]) / maxAbsDiff
			}
		}
	}
	return diff
}

// computeStats calculates RMSE, MAE, and Max difference between ideal and actual.
func (b *BeforeAfterToggle) computeStats() (rmse, mae, maxDiff float64) {
	if b.idealData == nil || b.actualData == nil {
		return 0, 0, 0
	}

	rows := len(b.idealData)
	if rows == 0 {
		return 0, 0, 0
	}
	cols := len(b.idealData[0])
	n := float64(rows * cols)

	var sumSq, sumAbs float64
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			diff := b.idealData[i][j] - b.actualData[i][j]
			absDiff := math.Abs(diff)
			sumSq += diff * diff
			sumAbs += absDiff
			if absDiff > maxDiff {
				maxDiff = absDiff
			}
		}
	}

	rmse = math.Sqrt(sumSq / n)
	mae = sumAbs / n
	return rmse, mae, maxDiff
}

// updateStatsLabel updates the statistics display.
func (b *BeforeAfterToggle) updateStatsLabel() {
	if b.statsLabel == nil {
		return
	}

	rmse, mae, maxDiff := b.computeStats()

	// Convert to level units (0-29 scale)
	rmseLevel := rmse * 29
	maeLevel := mae * 29
	maxDiffLevel := maxDiff * 29

	statsText := fmt.Sprintf("RMSE: %.3f (%.2f levels) | MAE: %.3f (%.2f levels) | Max Δ: %.3f (%.2f levels)",
		rmse, rmseLevel, mae, maeLevel, maxDiff, maxDiffLevel)
	b.statsLabel.SetText(statsText)
}

// updateLegend updates the legend display based on the current mode.
func (b *BeforeAfterToggle) updateLegend() {
	if b.legendLabel == nil {
		return
	}

	if b.mode == "diff" {
		b.legendLabel.SetText("Color Scale: Blue (Actual < Ideal) ← White (No Difference) → Red (Actual > Ideal)")
		b.legendLabel.Show()
	} else {
		b.legendLabel.Hide()
	}
}

// CreateRenderer implements fyne.Widget.
func (b *BeforeAfterToggle) CreateRenderer() fyne.WidgetRenderer {
	b.toggleGroup = widget.NewRadioGroup(
		[]string{"Split View", "Ideal Only", "Actual Only", "Difference"},
		func(value string) {
			mode := map[string]string{
				"Split View":  "split",
				"Ideal Only":  "before",
				"Actual Only": "after",
				"Difference":  "diff",
			}[value]
			b.SetMode(mode)
		},
	)
	b.toggleGroup.SetSelected("Split View")

	// Statistics label for RMSE, MAE, Max Diff
	b.statsLabel = widget.NewLabel("Statistics: Run comparison to see metrics")
	b.statsLabel.TextStyle = fyne.TextStyle{Monospace: false}
	b.statsLabel.Alignment = fyne.TextAlignCenter

	// Legend for difference view
	b.legendLabel = widget.NewLabel("")
	b.legendLabel.TextStyle = fyne.TextStyle{Italic: true}
	b.legendLabel.Alignment = fyne.TextAlignCenter
	b.legendLabel.Hide() // Hidden by default

	// Create dynamic labels that update based on mode
	b.leftLabel = widget.NewLabelWithStyle("Ideal (No Non-Idealities)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	b.rightLabel = widget.NewLabelWithStyle("Actual (With Non-Idealities)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	leftPane := container.NewBorder(b.leftLabel, nil, nil, nil, b.leftHeatmap)
	rightPane := container.NewBorder(b.rightLabel, nil, nil, nil, b.rightHeatmap)

	splitView := container.NewHSplit(leftPane, rightPane)
	splitView.SetOffset(0.5)

	// Top controls with toggle, stats, and legend
	topControls := container.NewVBox(
		b.toggleGroup,
		b.statsLabel,
		b.legendLabel,
	)

	content := container.NewBorder(topControls, nil, nil, nil, splitView)

	return widget.NewSimpleRenderer(content)
}
