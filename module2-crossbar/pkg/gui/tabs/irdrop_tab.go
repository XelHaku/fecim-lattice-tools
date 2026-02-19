// Package tabs provides individual tab components for the Demo 2 crossbar GUI.
package tabs

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/crossbar"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// IRDropTab provides IR drop analysis visualization.
type IRDropTab struct {
	simulator   *crossbar.IRDropSimulator
	heatmap     *fyne.Container
	legend      *sharedwidgets.ColorLegend
	statsLabel  *widget.Label
	statusLabel *widget.Label
	arraySize   int
}

// NewIRDropTab creates a new IR drop analysis tab.
func NewIRDropTab(arraySize int) *IRDropTab {
	tab := &IRDropTab{
		arraySize:   arraySize,
		statsLabel:  widget.NewLabel(""),
		statusLabel: widget.NewLabel("Ready"),
	}
	tab.initSimulator()
	return tab
}

func (t *IRDropTab) initSimulator() {
	t.simulator = crossbar.NewIRDropSimulator(t.arraySize, t.arraySize)

	// Set up input voltages
	for i := 0; i < t.arraySize; i++ {
		t.simulator.SetInputVoltage(i, 0.3+0.2*float64(i%5)/4.0)
	}

	// Set up conductances (non-uniform pattern)
	for i := 0; i < t.arraySize; i++ {
		for j := 0; j < t.arraySize; j++ {
			distFromCenter := float64((i-t.arraySize/2)*(i-t.arraySize/2) + (j-t.arraySize/2)*(j-t.arraySize/2))
			g := 50e-6 + 30e-6*distFromCenter/float64(t.arraySize*t.arraySize/2)
			t.simulator.SetConductance(i, j, g)
		}
	}

	t.simulator.Simulate(100)
}

// Content returns the tab content.
func (t *IRDropTab) Content() fyne.CanvasObject {
	// Heatmap container
	t.heatmap = container.NewWithoutLayout()
	t.heatmap.Resize(fyne.NewSize(400, 400))

	// Color legend (vertical)
	t.legend = sharedwidgets.NewColorLegend(0, 1, "mV", true, sharedwidgets.GreenToRedColor)

	t.updateHeatmap()

	heatmapScroll := container.NewScroll(t.heatmap)
	heatmapScroll.SetMinSize(fyne.NewSize(350, 350))

	// Add legend to left of heatmap
	heatmapWithLegend := container.NewBorder(nil, nil, t.legend, nil, heatmapScroll)

	// Stats panel
	t.updateStats()

	// Buttons
	mitigate2xBtn := widget.NewButton("Apply 2x Wider Lines", func() {
		mitigation := crossbar.IRDropMitigation{
			UseWidenedLines:   true,
			LineWidthIncrease: 2.0,
		}
		t.simulator.ApplyMitigation(mitigation)
		t.updateHeatmap()
		t.updateStats()
		t.statusLabel.SetText("Mitigation applied: 2x wider metal lines")
	})

	mitigate4xBtn := widget.NewButton("Apply 4x Wider Lines", func() {
		mitigation := crossbar.IRDropMitigation{
			UseWidenedLines:   true,
			LineWidthIncrease: 4.0,
		}
		t.simulator.ApplyMitigation(mitigation)
		t.updateHeatmap()
		t.updateStats()
		t.statusLabel.SetText("Mitigation applied: 4x wider metal lines")
	})

	resetBtn := widget.NewButton("Reset", func() {
		t.initSimulator()
		t.updateHeatmap()
		t.updateStats()
		t.statusLabel.SetText("IR drop simulation reset")
	})

	// Wrap stats in scroll to prevent layout resize
	statsScroll := container.NewVScroll(t.statsLabel)
	statsScroll.SetMinSize(fyne.NewSize(240, 150))

	rightPanel := container.NewVBox(
		statsScroll,
		widget.NewSeparator(),
		widget.NewLabel("Mitigation Strategies:"),
		mitigate2xBtn,
		mitigate4xBtn,
		widget.NewSeparator(),
		resetBtn,
	)

	content := container.NewHSplit(
		container.NewBorder(
			widget.NewLabelWithStyle("IR Drop Heatmap", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			t.statusLabel,
			nil, nil,
			heatmapWithLegend,
		),
		rightPanel,
	)
	content.SetOffset(0.6)

	return content
}

func (t *IRDropTab) updateHeatmap() {
	if t.heatmap == nil {
		return
	}
	t.heatmap.Objects = nil

	cellSize := float32(20)
	padding := float32(2)

	maxDrop := t.simulator.GetMaxIRDrop()
	if maxDrop == 0 {
		maxDrop = 1
	}

	// Update legend range
	if t.legend != nil {
		t.legend.SetRange(0, maxDrop*1000) // Convert to mV
	}

	for i := 0; i < t.arraySize; i++ {
		for j := 0; j < t.arraySize; j++ {
			drop := t.simulator.IRDropMap[i][j]
			normalized := drop / maxDrop

			// Color from green (low) to red (high)
			r := uint8(normalized * 255)
			g := uint8((1 - normalized) * 200)
			b := uint8(50)

			rect := canvas.NewRectangle(color.RGBA{r, g, b, 255})
			rect.Resize(fyne.NewSize(cellSize, cellSize))
			rect.Move(fyne.NewPos(float32(j)*(cellSize+padding)+30, float32(i)*(cellSize+padding)+30))
			t.heatmap.Add(rect)
		}
	}

	t.heatmap.Refresh()
}

func (t *IRDropTab) updateStats() {
	stats := t.simulator.GetStats()
	severity := t.getIRSeverity(stats.MaxOutputError)

	statsText := fmt.Sprintf(`IR Drop Analysis
================

Max IR Drop: %.3f mV
Avg IR Drop: %.3f mV
Max Output Error: %.2f%%
Avg Output Error: %.2f%%
Worst Cell: (%d, %d)

Severity: %s

Why It Matters:
---------------
IR drop causes voltage variations
along metal lines, leading to
inaccurate MVM results.

FeCIM Advantage:
Low-power operation reduces IR drop
impact compared to DRAM/SRAM.`,
		stats.MaxIRDrop*1000,
		stats.AvgIRDrop*1000,
		stats.MaxOutputError,
		stats.AvgOutputError,
		stats.WorstCellRow, stats.WorstCellCol,
		severity)

	t.statsLabel.SetText(statsText)
	t.statsLabel.Wrapping = fyne.TextWrapWord
	t.statsLabel.TextStyle = fyne.TextStyle{Monospace: true}
}

func (t *IRDropTab) getIRSeverity(maxError float64) string {
	if maxError < 1 {
		return "Excellent (<1% error)"
	} else if maxError < 5 {
		return "Good (<5% error)"
	} else if maxError < 10 {
		return "Acceptable (<10% error)"
	}
	return "Needs Mitigation (>10% error)"
}
