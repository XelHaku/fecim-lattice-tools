// Package gui provides custom widgets for crossbar visualization.
// M14: Conductance drift time-dependent visualization.
package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

// DriftTimeScale represents a time scale for drift visualization.
type DriftTimeScale struct {
	Label    string
	Seconds  float64
	DriftPct float64 // Predicted drift percentage
	LevelErr float64 // Level error rate percentage
}

// DriftVisualization shows conductance drift at different time scales.
// M14: Visualizes 1s, 1hr, 1day, 1year retention/drift.
type DriftVisualization struct {
	widget.BaseWidget

	// Time scales to display
	timeScales []DriftTimeScale

	// Simulator for drift calculations
	simulator *crossbar.DriftSimulator

	// UI components
	scaleLabels []*widget.Label
	driftBars   []*canvas.Rectangle
	driftLabels []*canvas.Text
	statusLabel *widget.Label
	techCompare *widget.Label
	modelInfo   *widget.Label
}

// NewDriftVisualization creates a new drift visualization widget.
func NewDriftVisualization() *DriftVisualization {
	w := &DriftVisualization{
		timeScales: []DriftTimeScale{
			{Label: "1 second", Seconds: 1},
			{Label: "1 hour", Seconds: 3600},
			{Label: "1 day", Seconds: 86400},
			{Label: "1 week", Seconds: 604800},
			{Label: "1 year", Seconds: 31536000},
			{Label: "10 years", Seconds: 315360000},
		},
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetSimulator sets the drift simulator and recalculates predictions.
func (w *DriftVisualization) SetSimulator(sim *crossbar.DriftSimulator) {
	w.simulator = sim
	w.calculateDriftPredictions()
	w.Refresh()
}

// calculateDriftPredictions calculates drift for each time scale.
func (w *DriftVisualization) calculateDriftPredictions() {
	if w.simulator == nil {
		return
	}

	// Create temporary simulators for each time scale
	for i := range w.timeScales {
		tempSim := crossbar.NewDriftSimulator(8, 8, 30)
		tempSim.DriftCoeff = w.simulator.DriftCoeff
		tempSim.Temperature = w.simulator.Temperature

		// Simulate to the target time
		numSteps := 100
		dt := w.timeScales[i].Seconds / float64(numSteps)
		for step := 0; step < numSteps; step++ {
			tempSim.SimulateTimeStep(dt)
		}

		stats := tempSim.GetStats()
		w.timeScales[i].DriftPct = stats.AvgDriftPercent
		w.timeScales[i].LevelErr = stats.LevelErrorRate
	}
}

// RunSimulation runs drift simulation with default parameters.
func (w *DriftVisualization) RunSimulation() {
	if w.simulator == nil {
		w.simulator = crossbar.NewDriftSimulator(8, 8, 30)
	}
	w.calculateDriftPredictions()
	w.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (w *DriftVisualization) CreateRenderer() fyne.WidgetRenderer {
	// Title
	title := widget.NewLabelWithStyle(
		"Conductance Drift Over Time",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Create bars and labels for each time scale
	w.scaleLabels = make([]*widget.Label, len(w.timeScales))
	w.driftBars = make([]*canvas.Rectangle, len(w.timeScales))
	w.driftLabels = make([]*canvas.Text, len(w.timeScales))

	rows := make([]fyne.CanvasObject, 0, len(w.timeScales))

	for i, scale := range w.timeScales {
		// Time scale label
		w.scaleLabels[i] = widget.NewLabel(scale.Label)
		w.scaleLabels[i].TextStyle = fyne.TextStyle{Monospace: true}

		// Drift bar (will be sized based on drift %)
		w.driftBars[i] = canvas.NewRectangle(w.getDriftColor(scale.DriftPct))
		w.driftBars[i].SetMinSize(fyne.NewSize(10, 16))

		// Drift value label
		w.driftLabels[i] = canvas.NewText(
			fmt.Sprintf("%.3f%% drift", scale.DriftPct),
			color.White,
		)
		w.driftLabels[i].TextSize = 14

		row := container.NewHBox(
			container.NewGridWrap(fyne.NewSize(80, 20), w.scaleLabels[i]),
			w.driftBars[i],
			w.driftLabels[i],
		)
		rows = append(rows, row)
	}

	// Status and info labels
	w.statusLabel = widget.NewLabel("Click 'Run Simulation' to see drift predictions")
	w.statusLabel.Alignment = fyne.TextAlignCenter

	w.techCompare = widget.NewLabel("FeCIM: ~50× better retention than RRAM")
	w.techCompare.Alignment = fyne.TextAlignCenter
	w.techCompare.TextStyle = fyne.TextStyle{Italic: true}

	// Model info with assumption warning
	w.modelInfo = widget.NewLabel("⚠️ Drift coefficient: ASSUMED (derived from retention studies)")
	w.modelInfo.Alignment = fyne.TextAlignCenter
	w.modelInfo.TextStyle = fyne.TextStyle{Italic: true}

	// Run button
	runBtn := widget.NewButton("Run Simulation", func() {
		w.RunSimulation()
	})

	// Build content
	barsContainer := container.NewVBox(rows...)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		barsContainer,
		widget.NewSeparator(),
		w.statusLabel,
		w.techCompare,
		w.modelInfo,
		runBtn,
	)

	return widget.NewSimpleRenderer(content)
}

// getDriftColor returns a color based on drift severity.
func (w *DriftVisualization) getDriftColor(driftPct float64) color.Color {
	// Green (low) to yellow to red (high)
	if driftPct < 0.1 {
		return color.RGBA{50, 200, 50, 255} // Green - excellent
	} else if driftPct < 1.0 {
		return color.RGBA{150, 200, 50, 255} // Yellow-green - good
	} else if driftPct < 5.0 {
		return color.RGBA{255, 200, 0, 255} // Yellow - moderate
	} else if driftPct < 10.0 {
		return color.RGBA{255, 150, 0, 255} // Orange - concerning
	}
	return color.RGBA{255, 80, 80, 255} // Red - high drift
}

// Refresh updates the display with current data.
func (w *DriftVisualization) Refresh() {
	if w.driftBars == nil || len(w.driftBars) != len(w.timeScales) {
		w.BaseWidget.Refresh()
		return
	}

	// Update bars and labels
	maxDrift := 10.0 // 10% max for scaling
	for i, scale := range w.timeScales {
		// Update bar color
		w.driftBars[i].FillColor = w.getDriftColor(scale.DriftPct)

		// Update bar width based on drift (capped at max)
		drift := scale.DriftPct
		if drift > maxDrift {
			drift = maxDrift
		}
		barWidth := float32(drift/maxDrift*150) + 10 // Min 10px
		w.driftBars[i].SetMinSize(fyne.NewSize(barWidth, 16))
		w.driftBars[i].Refresh()

		// Update label
		if scale.LevelErr > 0 {
			w.driftLabels[i].Text = fmt.Sprintf("%.3f%% drift (%.2f%% errors)",
				scale.DriftPct, scale.LevelErr)
		} else {
			w.driftLabels[i].Text = fmt.Sprintf("%.4f%% drift", scale.DriftPct)
		}
		w.driftLabels[i].Refresh()
	}

	// Update status
	if w.statusLabel != nil && w.simulator != nil {
		stats := w.simulator.GetStats()
		w.statusLabel.SetText(fmt.Sprintf(
			"10-year retention: %.2f%% | Current drift: %.4f%%",
			stats.RetentionPrediction, stats.AvgDriftPercent))

		// Update tech comparison
		if w.techCompare != nil {
			w.techCompare.SetText(fmt.Sprintf(
				"FeCIM: %.0f× better than RRAM, %.0f× better than PCM",
				stats.TechnologyComparison.FeFETAdvantage,
				stats.TechnologyComparison.PCMDrift/stats.TechnologyComparison.FeFETDrift))
		}

		// Update model info
		if w.modelInfo != nil {
			info := w.simulator.GetDriftModelInfo()
			if info.IsAssumed {
				w.modelInfo.SetText(fmt.Sprintf("⚠️ Model: %s (coeff=%.4f)",
					info.ModelName, info.Coefficient))
			} else {
				w.modelInfo.SetText(fmt.Sprintf("Model: %s (coeff=%.4f)",
					info.ModelName, info.Coefficient))
			}
		}
	}

	w.BaseWidget.Refresh()
}

// MinSize returns the minimum size.
func (w *DriftVisualization) MinSize() fyne.Size {
	return fyne.NewSize(300, 300)
}
