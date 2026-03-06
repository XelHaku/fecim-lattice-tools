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

// SneakTab provides sneak path analysis visualization.
type SneakTab struct {
	simulator   *crossbar.SneakPathAnalyzer
	heatmap     *fyne.Container
	legend      *sharedwidgets.ColorLegend
	statsLabel  *widget.Label
	statusLabel *widget.Label
	arraySize   int
}

// NewSneakTab creates a new sneak path analysis tab.
func NewSneakTab(arraySize int) *SneakTab {
	if arraySize < 1 {
		arraySize = 1
	}
	tab := &SneakTab{
		arraySize:   arraySize,
		statsLabel:  widget.NewLabel(""),
		statusLabel: widget.NewLabel("Ready"),
	}
	tab.initSimulator()
	return tab
}

func (t *SneakTab) initSimulator() {
	t.simulator = crossbar.NewSneakPathAnalyzer(t.arraySize, t.arraySize)

	// Set up conductances (varied pattern)
	for i := 0; i < t.arraySize; i++ {
		for j := 0; j < t.arraySize; j++ {
			g := (10 + float64((i*7+j*11)%80)) * 1e-6
			t.simulator.SetConductance(i, j, g)
		}
	}

	// Analyze center cell
	t.simulator.AnalyzeTarget(t.arraySize/2, t.arraySize/2, 0.5)
}

// Content returns the tab content.
func (t *SneakTab) Content() fyne.CanvasObject {
	// Heatmap container
	t.heatmap = container.NewWithoutLayout()
	t.heatmap.Resize(fyne.NewSize(400, 400))

	// Color legend (vertical) - blue to yellow for sneak current
	t.legend = sharedwidgets.NewColorLegend(0, 1, "µA", true, sharedwidgets.BlueToYellowColor)

	t.updateHeatmap()

	heatmapScroll := container.NewScroll(t.heatmap)
	heatmapScroll.SetMinSize(fyne.NewSize(350, 350))

	// Add legend to left of heatmap
	heatmapWithLegend := container.NewBorder(nil, nil, t.legend, nil, heatmapScroll)

	// Stats panel
	t.updateStats()

	// Buttons
	selector100Btn := widget.NewButton("Apply Selector (100:1)", func() {
		mitigation := crossbar.SneakMitigation{
			UseSelector:   true,
			SelectorOnOff: 100,
		}
		t.simulator.AnalyzeWithMitigation(t.arraySize/2, t.arraySize/2, 0.5, mitigation)
		t.updateHeatmap()
		t.updateStats()
		t.statusLabel.SetText("Mitigation applied: selector (100:1)")
	})

	selector1000Btn := widget.NewButton("Apply Selector (1000:1)", func() {
		mitigation := crossbar.SneakMitigation{
			UseSelector:   true,
			SelectorOnOff: 1000,
		}
		t.simulator.AnalyzeWithMitigation(t.arraySize/2, t.arraySize/2, 0.5, mitigation)
		t.updateHeatmap()
		t.updateStats()
		t.statusLabel.SetText("Mitigation applied: selector (1000:1)")
	})

	halfSelectBtn := widget.NewButton("Apply Half-Select Scheme", func() {
		mitigation := crossbar.SneakMitigation{
			UseHalfSelect: true,
		}
		t.simulator.AnalyzeWithMitigation(t.arraySize/2, t.arraySize/2, 0.5, mitigation)
		t.updateHeatmap()
		t.updateStats()
		t.statusLabel.SetText("Mitigation applied: half-select scheme")
	})

	resetBtn := widget.NewButton("Reset", func() {
		t.initSimulator()
		t.updateHeatmap()
		t.updateStats()
		t.statusLabel.SetText("Sneak path simulation reset")
	})

	// Wrap stats in scroll to prevent layout resize
	statsScroll := container.NewVScroll(t.statsLabel)
	statsScroll.SetMinSize(fyne.NewSize(240, 150))

	rightPanel := container.NewVBox(
		statsScroll,
		widget.NewSeparator(),
		widget.NewLabel("Mitigation Strategies:"),
		selector100Btn,
		selector1000Btn,
		halfSelectBtn,
		widget.NewSeparator(),
		resetBtn,
	)

	content := container.NewHSplit(
		container.NewBorder(
			widget.NewLabelWithStyle("Sneak Current Map (X = target)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			t.statusLabel,
			nil, nil,
			heatmapWithLegend,
		),
		rightPanel,
	)
	content.SetOffset(0.6)

	return content
}

func (t *SneakTab) updateHeatmap() {
	if t.heatmap == nil {
		return
	}
	t.heatmap.Objects = nil

	cellSize := float32(20)
	padding := float32(2)

	// Find max sneak current
	maxSneak := 0.0
	for i := 0; i < t.arraySize; i++ {
		for j := 0; j < t.arraySize; j++ {
			if t.simulator.SneakCurrents[i][j] > maxSneak {
				maxSneak = t.simulator.SneakCurrents[i][j]
			}
		}
	}
	if maxSneak == 0 {
		maxSneak = 1
	}

	// Update legend range (convert to µA)
	if t.legend != nil {
		t.legend.SetRange(0, maxSneak*1e6)
	}

	colorPrimary := color.RGBA{0, 212, 255, 255}

	for i := 0; i < t.arraySize; i++ {
		for j := 0; j < t.arraySize; j++ {
			var rectColor color.Color

			if i == t.simulator.TargetRow && j == t.simulator.TargetCol {
				// Target cell - cyan
				rectColor = colorPrimary
			} else {
				sneak := t.simulator.SneakCurrents[i][j]
				normalized := sneak / maxSneak

				// Color from blue (low) to yellow (high) with minimum brightness
				// Ensures visibility on dark theme backgrounds
				r := uint8(50 + normalized*205)  // 50-255: visible at all levels
				g := uint8(80 + normalized*140)  // 80-220: visible at all levels
				b := uint8(200 - normalized*150) // 200-50: blue fades as yellow increases
				rectColor = color.RGBA{r, g, b, 255}
			}

			rect := canvas.NewRectangle(rectColor)
			rect.Resize(fyne.NewSize(cellSize, cellSize))
			rect.Move(fyne.NewPos(float32(j)*(cellSize+padding)+30, float32(i)*(cellSize+padding)+30))
			t.heatmap.Add(rect)
		}
	}

	// Mark target
	targetX := float32(t.simulator.TargetCol)*(cellSize+padding) + 30 + cellSize/2 - 5
	targetY := float32(t.simulator.TargetRow)*(cellSize+padding) + 30 + cellSize/2 - 6
	targetMark := canvas.NewText("X", color.White)
	targetMark.TextSize = 14
	targetMark.TextStyle = fyne.TextStyle{Bold: true}
	targetMark.Move(fyne.NewPos(targetX, targetY))
	t.heatmap.Add(targetMark)

	t.heatmap.Refresh()
}

func (t *SneakTab) updateStats() {
	stats := t.simulator.GetStats(0.5)
	severity := t.getSneakSeverity(stats.SignalToNoiseRatio)

	statsText := fmt.Sprintf(`Sneak Path Analysis
===================

Target Cell: (%d, %d)
Target Current: %.3f uA
Total Sneak Current: %.3f uA
Sneak Ratio: %.2f%%
Number of Paths: %d
Signal-to-Noise: %.1f dB

Severity: %s

Why It Matters:
---------------
Sneak paths are parasitic current
paths through unselected cells that
corrupt the target cell reading.

1T1R vs 1R Architecture:
- 1R (simple): High sneak currents
- 1T1R (transistor): Low sneak currents
- FeCIM can use either approach`,
		t.simulator.TargetRow, t.simulator.TargetCol,
		stats.TargetCurrent*1e6,
		stats.TotalSneakCurrent*1e6,
		stats.SneakRatio*100,
		stats.NumSneakPaths,
		stats.SignalToNoiseRatio,
		severity)

	t.statsLabel.SetText(statsText)
	t.statsLabel.Wrapping = fyne.TextWrapWord
	t.statsLabel.TextStyle = fyne.TextStyle{Monospace: false}
}

func (t *SneakTab) getSneakSeverity(snr float64) string {
	if snr > 30 {
		return "Excellent (SNR > 30dB)"
	} else if snr > 20 {
		return "Good (SNR > 20dB)"
	} else if snr > 10 {
		return "Acceptable (SNR > 10dB)"
	}
	return "Needs Selector (SNR < 10dB)"
}
