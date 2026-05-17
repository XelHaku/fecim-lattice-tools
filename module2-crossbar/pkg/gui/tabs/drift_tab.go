//go:build legacy_fyne

// Package tabs provides individual tab components for the Demo 2 crossbar GUI.
package tabs

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/crossbar"
)

// DriftTab provides drift and variation analysis visualization.
type DriftTab struct {
	simulator   *crossbar.DriftSimulator
	driftChart  *fyne.Container
	statsLabel  *widget.Label
	statusLabel *widget.Label
	arraySize   int
}

// NewDriftTab creates a new drift analysis tab.
func NewDriftTab(arraySize int) *DriftTab {
	if arraySize < 1 {
		arraySize = 1
	}
	tab := &DriftTab{
		arraySize:   arraySize,
		statsLabel:  widget.NewLabel(""),
		statusLabel: widget.NewLabel("Ready"),
	}
	tab.initSimulator()
	return tab
}

func (t *DriftTab) initSimulator() {
	t.simulator = crossbar.NewDriftSimulator(t.arraySize, t.arraySize, 30)

	// Initialize with varied conductance levels
	for i := 0; i < t.arraySize; i++ {
		for j := 0; j < t.arraySize; j++ {
			level := (i*3 + j*5) % 30
			t.simulator.SetConductanceLevel(i, j, level)
		}
	}

	// Simulate initial time
	for step := 0; step < 50; step++ {
		t.simulator.SimulateTimeStep(200)
		t.simulator.RecordSnapshot()
	}
}

// Content returns the tab content.
func (t *DriftTab) Content() fyne.CanvasObject {
	// Drift chart container
	t.driftChart = container.NewWithoutLayout()
	t.driftChart.Resize(fyne.NewSize(500, 300))
	t.updateChart()

	chartScroll := container.NewScroll(t.driftChart)
	chartScroll.SetMinSize(fyne.NewSize(450, 250))

	// Stats panel
	t.updateStats()

	// Buttons
	sim1HourBtn := widget.NewButton("Simulate +1 Hour", func() {
		for step := 0; step < 18; step++ { // 18 * 200s = 1 hour
			t.simulator.SimulateTimeStep(200)
			t.simulator.RecordSnapshot()
		}
		t.updateChart()
		t.updateStats()
		t.statusLabel.SetText("Simulated +1 hour of drift")
	})

	sim1DayBtn := widget.NewButton("Simulate +1 Day", func() {
		for step := 0; step < 432; step++ { // 432 * 200s = 1 day
			t.simulator.SimulateTimeStep(200)
			if step%10 == 0 {
				t.simulator.RecordSnapshot()
			}
		}
		t.updateChart()
		t.updateStats()
		t.statusLabel.SetText("Simulated +1 day of drift")
	})

	resetBtn := widget.NewButton("Reset", func() {
		t.simulator.Reset()
		for step := 0; step < 50; step++ {
			t.simulator.SimulateTimeStep(200)
			t.simulator.RecordSnapshot()
		}
		t.updateChart()
		t.updateStats()
		t.statusLabel.SetText("Drift simulation reset")
	})

	// Technology comparison
	compareBtn := widget.NewButton("Compare Technologies", func() {
		t.showTechComparison()
	})

	// Wrap stats in scroll to prevent layout resize
	statsScroll := container.NewVScroll(t.statsLabel)
	statsScroll.SetMinSize(fyne.NewSize(240, 150))

	rightPanel := container.NewVBox(
		statsScroll,
		widget.NewSeparator(),
		widget.NewLabel("Time Simulation:"),
		sim1HourBtn,
		sim1DayBtn,
		widget.NewSeparator(),
		compareBtn,
		widget.NewSeparator(),
		resetBtn,
	)

	content := container.NewHSplit(
		container.NewBorder(
			widget.NewLabelWithStyle("Drift Over Time", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			t.statusLabel,
			nil, nil,
			chartScroll,
		),
		rightPanel,
	)
	content.SetOffset(0.6)

	return content
}

func (t *DriftTab) updateChart() {
	if t.driftChart == nil {
		return
	}
	t.driftChart.Objects = nil

	if len(t.simulator.DriftHistory) == 0 {
		return
	}

	// Find max drift for scaling
	maxDrift := 0.0
	for _, snap := range t.simulator.DriftHistory {
		if snap.MaxDrift > maxDrift {
			maxDrift = snap.MaxDrift
		}
	}
	if maxDrift == 0 {
		maxDrift = 1e-9
	}

	chartWidth := float32(450)
	chartHeight := float32(200)
	marginLeft := float32(50)

	// Draw axes
	yAxis := canvas.NewLine(color.White)
	yAxis.Position1 = fyne.NewPos(marginLeft, 10)
	yAxis.Position2 = fyne.NewPos(marginLeft, chartHeight)
	t.driftChart.Add(yAxis)

	xAxis := canvas.NewLine(color.White)
	xAxis.Position1 = fyne.NewPos(marginLeft, chartHeight)
	xAxis.Position2 = fyne.NewPos(chartWidth, chartHeight)
	t.driftChart.Add(xAxis)

	// Draw drift bars
	barWidth := (chartWidth - marginLeft) / float32(len(t.simulator.DriftHistory))
	for i, snap := range t.simulator.DriftHistory {
		// Max drift bar
		height := float32(snap.MaxDrift/maxDrift) * (chartHeight - 20)
		x := marginLeft + float32(i)*barWidth

		maxBar := canvas.NewRectangle(color.RGBA{255, 100, 100, 200})
		maxBar.Resize(fyne.NewSize(barWidth-1, height))
		maxBar.Move(fyne.NewPos(x, chartHeight-height))
		t.driftChart.Add(maxBar)

		// Avg drift bar
		avgHeight := float32(snap.AvgDrift/maxDrift) * (chartHeight - 20)
		avgBar := canvas.NewRectangle(color.RGBA{100, 200, 100, 200})
		avgBar.Resize(fyne.NewSize(barWidth-1, avgHeight))
		avgBar.Move(fyne.NewPos(x, chartHeight-avgHeight))
		t.driftChart.Add(avgBar)
	}

	// Labels
	yLabel := canvas.NewText(fmt.Sprintf("%.1e", maxDrift), color.White)
	yLabel.TextSize = 14
	yLabel.Move(fyne.NewPos(5, 10))
	t.driftChart.Add(yLabel)

	xLabel := canvas.NewText("Time", color.White)
	xLabel.TextSize = 14
	xLabel.Move(fyne.NewPos(chartWidth/2, chartHeight+10))
	t.driftChart.Add(xLabel)

	// Legend
	maxLegend := canvas.NewText("Max Drift", color.RGBA{255, 100, 100, 255})
	maxLegend.TextSize = 14
	maxLegend.Move(fyne.NewPos(chartWidth-100, 10))
	t.driftChart.Add(maxLegend)

	avgLegend := canvas.NewText("Avg Drift", color.RGBA{100, 200, 100, 255})
	avgLegend.TextSize = 14
	avgLegend.Move(fyne.NewPos(chartWidth-100, 25))
	t.driftChart.Add(avgLegend)

	t.driftChart.Refresh()
}

func (t *DriftTab) updateStats() {
	stats := t.simulator.GetStats()

	statsText := fmt.Sprintf(`Conductance Drift Analysis
==========================

Elapsed Time: %.1f seconds
Average Drift: %.4f%%
Maximum Drift: %.4f%%
Level Errors: %d (%.4f%%)
10-Year Retention: %.2f%%

FeCIM Drift Coefficient: %.4f
(est. ~50x better than RRAM)

Why It Matters:
---------------
Conductance drift causes stored
weights to change over time,
degrading inference accuracy.

FeCIM Advantage:
Ferroelectric polarization provides
extremely stable states with
minimal drift over time.`,
		stats.ElapsedTime,
		stats.AvgDriftPercent,
		stats.MaxDriftPercent,
		stats.NumLevelErrors,
		stats.LevelErrorRate,
		stats.RetentionPrediction,
		stats.TechnologyComparison.FeFETDrift)

	t.statsLabel.SetText(statsText)
	t.statsLabel.Wrapping = fyne.TextWrapWord
	t.statsLabel.TextStyle = fyne.TextStyle{Monospace: false}
}

func (t *DriftTab) showTechComparison() {
	// Run comparison
	results := crossbar.CompareTechnologies(t.arraySize, t.arraySize, 86400) // 1 day

	// Build comparison text
	compText := "Technology Comparison (24 hours)\n"
	compText += "================================\n\n"

	order := []string{"FeCIM (FeFET)", "Flash", "RRAM", "PCM"}
	for _, name := range order {
		stats, ok := results[name]
		if !ok {
			continue
		}
		compText += fmt.Sprintf("%-15s: Drift %.3f%%, Retention %.1f%%\n",
			name, stats.MaxDriftPercent, stats.RetentionPrediction)
	}

	compText += "\n"
	compText += "FeCIM achieves (model est.):\n"
	compText += "- ~50x lower drift than RRAM\n"
	compText += "- ~100x lower drift than PCM\n"
	compText += "- ~20x lower drift than Flash\n"
	compText += "\n10-year retention: >99.9% (modeled)"

	t.statsLabel.SetText(compText)
	t.statusLabel.SetText("Technology comparison complete")
}

// TechComparisonContent returns a separate view for technology comparison.
func (t *DriftTab) TechComparisonContent() fyne.CanvasObject {
	// Run comparison
	results := crossbar.CompareTechnologies(t.arraySize, t.arraySize, 86400) // 1 day

	// Create comparison bars
	bars := container.NewVBox()

	// Find max for scaling
	maxDrift := 0.0
	for _, stats := range results {
		if stats.MaxDriftPercent > maxDrift {
			maxDrift = stats.MaxDriftPercent
		}
	}

	order := []string{"FeCIM (FeFET)", "Flash", "RRAM", "PCM"}
	barColors := []color.Color{
		color.RGBA{0, 220, 150, 255}, // Green for FeCIM
		color.RGBA{255, 200, 0, 255}, // Yellow for Flash
		color.RGBA{255, 150, 0, 255}, // Orange for RRAM
		color.RGBA{255, 80, 80, 255}, // Red for PCM
	}

	for i, name := range order {
		stats, ok := results[name]
		if !ok {
			continue
		}

		// Label
		label := widget.NewLabel(fmt.Sprintf("%-15s Drift: %.3f%%  Retention: %.2f%%",
			name, stats.MaxDriftPercent, stats.RetentionPrediction))

		// Progress bar (scaled)
		bar := widget.NewProgressBar()
		bar.SetValue(stats.MaxDriftPercent / maxDrift)

		// Color indicator
		colorRect := canvas.NewRectangle(barColors[i])
		colorRect.SetMinSize(fyne.NewSize(20, 20))

		row := container.NewHBox(colorRect, label)
		bars.Add(row)
		bars.Add(bar)
		bars.Add(layout.NewSpacer())
	}

	return bars
}
