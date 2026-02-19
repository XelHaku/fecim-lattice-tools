// Package gui provides custom widgets for crossbar visualization.
// M08: Side-by-side sneak path comparison (PASSIVE vs 1T1R).
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/crossbar"
)

// SneakCompareWidget shows side-by-side comparison of sneak paths
// between PASSIVE (0T1R) and 1T1R architectures.
type SneakCompareWidget struct {
	widget.BaseWidget

	// Heatmaps for each architecture
	passiveHeatmap    *CrossbarHeatmap
	active1T1RHeatmap *CrossbarHeatmap

	// Data storage
	passiveSneakData    [][]float64
	active1T1RSneakData [][]float64

	// Labels and stats
	passiveLabel    *widget.Label
	activeLabel     *widget.Label
	statsLabel      *widget.Label
	comparisonLabel *widget.Label

	// Analysis results
	passiveAnalysis *crossbar.SneakPathAnalysis
	activeAnalysis  *crossbar.SneakPathAnalysis

	// Array reference
	array      *crossbar.Array
	rows, cols int

	// Callbacks
	OnCellTapped func(row, col int, isPassive bool)
}

// NewSneakCompareWidget creates a new sneak path comparison widget.
func NewSneakCompareWidget(rows, cols int) *SneakCompareWidget {
	w := &SneakCompareWidget{
		rows:              rows,
		cols:              cols,
		passiveHeatmap:    NewCrossbarHeatmap(rows, cols),
		active1T1RHeatmap: NewCrossbarHeatmap(rows, cols),
	}

	// Use plasma colormap for sneak paths (good contrast)
	w.passiveHeatmap.SetColormap("plasma")
	w.active1T1RHeatmap.SetColormap("plasma")

	// Wire up callbacks
	w.passiveHeatmap.OnCellTapped = func(row, col int) {
		w.syncSelection(row, col)
		if w.OnCellTapped != nil {
			w.OnCellTapped(row, col, true)
		}
	}
	w.active1T1RHeatmap.OnCellTapped = func(row, col int) {
		w.syncSelection(row, col)
		if w.OnCellTapped != nil {
			w.OnCellTapped(row, col, false)
		}
	}

	w.ExtendBaseWidget(w)
	return w
}

// syncSelection syncs selection across both heatmaps.
func (w *SneakCompareWidget) syncSelection(row, col int) {
	w.passiveHeatmap.SetSelection(row, col)
	w.active1T1RHeatmap.SetSelection(row, col)
}

// SetArray sets the crossbar array for analysis.
func (w *SneakCompareWidget) SetArray(array *crossbar.Array) {
	w.array = array
}

// SetDimensions changes the dimensions of both heatmaps.
func (w *SneakCompareWidget) SetDimensions(rows, cols int) {
	w.rows = rows
	w.cols = cols
	w.passiveHeatmap.SetDimensions(rows, cols)
	w.active1T1RHeatmap.SetDimensions(rows, cols)
}

// RunComparison performs sneak path analysis for both architectures.
func (w *SneakCompareWidget) RunComparison(array *crossbar.Array, input []float64) {
	w.array = array

	// Analyze with PASSIVE (0T1R) architecture
	passiveOpts := &crossbar.MVMOptions{
		EnableIRDrop:     false,
		EnableSneakPaths: true,
		EnableDrift:      false,
		Architecture:     "0T1R",
	}
	passiveResult, err := array.MVMWithNonIdealities(input, passiveOpts)
	if err == nil && passiveResult != nil {
		w.passiveAnalysis = passiveResult.SneakPathAnalysis
	}

	// Analyze with 1T1R architecture
	activeOpts := &crossbar.MVMOptions{
		EnableIRDrop:     false,
		EnableSneakPaths: true,
		EnableDrift:      false,
		Architecture:     "1T1R",
	}
	activeResult, err := array.MVMWithNonIdealities(input, activeOpts)
	if err == nil && activeResult != nil {
		w.activeAnalysis = activeResult.SneakPathAnalysis
	}

	// Convert to heatmap data
	w.passiveSneakData = w.analysisToHeatmapData(w.passiveAnalysis)
	w.active1T1RSneakData = w.analysisToHeatmapData(w.activeAnalysis)

	// Use same scale for fair comparison
	maxSneak := w.findMaxSneak(w.passiveSneakData, w.active1T1RSneakData)
	w.passiveHeatmap.SetFixedScale(0, maxSneak)
	w.active1T1RHeatmap.SetFixedScale(0, maxSneak)

	// Update heatmaps
	w.passiveHeatmap.SetData(w.passiveSneakData)
	w.active1T1RHeatmap.SetData(w.active1T1RSneakData)

	// Update stats
	w.updateStats()
}

// analysisToHeatmapData converts sneak analysis to 2D heatmap data.
func (w *SneakCompareWidget) analysisToHeatmapData(analysis *crossbar.SneakPathAnalysis) [][]float64 {
	if analysis == nil {
		// Return empty data
		data := make([][]float64, w.rows)
		for i := range data {
			data[i] = make([]float64, w.cols)
		}
		return data
	}
	return analysis.SneakCurrents
}

// findMaxSneak finds the maximum sneak value across both datasets.
func (w *SneakCompareWidget) findMaxSneak(data1, data2 [][]float64) float64 {
	maxVal := 0.0
	for _, data := range [][][]float64{data1, data2} {
		for i := range data {
			for j := range data[i] {
				if data[i][j] > maxVal {
					maxVal = data[i][j]
				}
			}
		}
	}
	if maxVal == 0 {
		maxVal = 1e-9 // Avoid division by zero
	}
	return maxVal
}

// updateStats updates the comparison statistics.
func (w *SneakCompareWidget) updateStats() {
	if w.statsLabel == nil {
		return
	}

	passiveTotal := 0.0
	activeTotal := 0.0

	if w.passiveAnalysis != nil {
		passiveTotal = w.passiveAnalysis.TotalSneak
	}
	if w.activeAnalysis != nil {
		activeTotal = w.activeAnalysis.TotalSneak
	}

	// Calculate reduction
	reduction := 0.0
	if passiveTotal > 0 {
		reduction = (passiveTotal - activeTotal) / passiveTotal * 100
	}

	statsText := fmt.Sprintf(
		"PASSIVE (0T1R): %.3f µA total sneak\n"+
			"1T1R (Active): %.3f µA total sneak\n"+
			"Reduction: %.1f%%",
		passiveTotal*1e6,
		activeTotal*1e6,
		reduction,
	)
	w.statsLabel.SetText(statsText)

	// Update comparison label with key insight
	if w.comparisonLabel != nil {
		if reduction > 99 {
			w.comparisonLabel.SetText("1T1R eliminates virtually all sneak paths")
		} else if reduction > 90 {
			w.comparisonLabel.SetText("1T1R provides excellent sneak reduction")
		} else if reduction > 50 {
			w.comparisonLabel.SetText("1T1R provides moderate sneak reduction")
		} else {
			w.comparisonLabel.SetText("Compare sneak current distributions")
		}
	}
}

// CreateRenderer implements fyne.Widget.
func (w *SneakCompareWidget) CreateRenderer() fyne.WidgetRenderer {
	// Create labels
	w.passiveLabel = widget.NewLabelWithStyle(
		"PASSIVE (0T1R)",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	w.passiveLabel.Truncation = fyne.TextTruncateEllipsis
	w.activeLabel = widget.NewLabelWithStyle(
		"1T1R (Active)",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	w.activeLabel.Truncation = fyne.TextTruncateEllipsis

	w.statsLabel = widget.NewLabel("Run MVM to compare architectures")
	w.statsLabel.TextStyle = fyne.TextStyle{Monospace: true}
	w.statsLabel.Wrapping = fyne.TextWrapWord

	w.comparisonLabel = widget.NewLabel("Select cells to compare sneak paths")
	w.comparisonLabel.Alignment = fyne.TextAlignCenter

	// Create panes
	leftPane := container.NewBorder(
		w.passiveLabel, nil, nil, nil,
		w.passiveHeatmap,
	)
	rightPane := container.NewBorder(
		w.activeLabel, nil, nil, nil,
		w.active1T1RHeatmap,
	)

	// Split view
	splitView := container.NewHSplit(leftPane, rightPane)
	splitView.SetOffset(0.5)

	// Info panel at bottom
	infoPanel := container.NewVBox(
		widget.NewSeparator(),
		w.statsLabel,
		w.comparisonLabel,
	)

	content := container.NewBorder(nil, infoPanel, nil, nil, splitView)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns the minimum size.
func (w *SneakCompareWidget) MinSize() fyne.Size {
	return fyne.NewSize(500, 350)
}
