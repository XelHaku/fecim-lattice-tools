// Package gui provides Fyne-based GUI components for MNIST visualization.
// accuracy_sweep_panel.go provides the accuracy-vs-parameter sweep analysis panel.
package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	neural "fecim-lattice-tools/shared/neural"
)

// AccuracySweepPanel shows accuracy-vs-levels, accuracy-vs-noise, and accuracy-vs-ADC results.
type AccuracySweepPanel struct {
	app          *DualModeApp
	statusLabel  *widget.Label
	resultsLabel *widget.Label
	content      fyne.CanvasObject
}

// NewAccuracySweepPanel creates the sweep panel bound to the given app.
func NewAccuracySweepPanel(app *DualModeApp) *AccuracySweepPanel {
	p := &AccuracySweepPanel{
		app:          app,
		statusLabel:  widget.NewLabel("Choose a sweep to run (uses up to 500-image subset)."),
		resultsLabel: widget.NewLabel(""),
	}
	p.resultsLabel.Wrapping = fyne.TextWrapWord

	btnLevels := widget.NewButton("Sweep Levels (2..30)", func() { go p.runLevelSweep() })
	btnADC := widget.NewButton("Sweep ADC Bits (3..10)", func() { go p.runADCSweep() })
	btnNoise := widget.NewButton("Sweep Noise (0%..20%)", func() { go p.runNoiseSweep() })

	buttons := container.NewGridWithColumns(3, btnLevels, btnADC, btnNoise)

	p.content = container.NewVBox(
		widget.NewLabelWithStyle("Accuracy Sweep Analysis", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Evaluates CIM accuracy while varying one parameter at a time."),
		buttons,
		widget.NewSeparator(),
		p.statusLabel,
		container.NewVScroll(p.resultsLabel),
	)
	return p
}

// Content returns the Fyne canvas object for this panel.
func (p *AccuracySweepPanel) Content() fyne.CanvasObject {
	return p.content
}

// getTestSubset returns up to maxN images/labels from the app's loaded dataset.
// Falls back to synthetic patterns if no dataset is loaded.
func (p *AccuracySweepPanel) getTestSubset(maxN int) ([][]float64, []int) {
	imgs := p.app.testImages()
	lbls := p.app.testLabels()
	if len(imgs) > 0 && len(imgs) == len(lbls) {
		if len(imgs) > maxN {
			return imgs[:maxN], lbls[:maxN]
		}
		return imgs, lbls
	}

	// Synthetic fallback: simple patterns with one non-zero input per class
	result := make([][]float64, maxN)
	labels := make([]int, maxN)
	for i := range result {
		result[i] = make([]float64, 784)
		digit := i % 10
		result[i][digit] = 1.0
		labels[i] = digit
	}
	return result, labels
}

func (p *AccuracySweepPanel) setStatus(s string) {
	fyne.Do(func() { p.statusLabel.SetText(s) })
}

func (p *AccuracySweepPanel) setResults(s string) {
	fyne.Do(func() { p.resultsLabel.SetText(s) })
}

func (p *AccuracySweepPanel) runLevelSweep() {
	p.setStatus("Running level sweep...")
	p.setResults("")

	imgs, lbls := p.getTestSubset(500)
	levels := []int{2, 4, 8, 16, 30}

	net := p.app.network()
	results, err := neural.SweepQuantizationLevels(net, imgs, lbls, levels)
	if err != nil {
		p.setStatus(fmt.Sprintf("Level sweep error: %v", err))
		return
	}

	var sb strings.Builder
	sb.WriteString("Levels  Accuracy  Bar\n")
	sb.WriteString("─────────────────────────────\n")
	for _, r := range results {
		bar := sweepProgressBar(r.Accuracy, 20)
		sb.WriteString(fmt.Sprintf("%3d lvl  %5.1f%%  %s\n", r.NumLevels, r.Accuracy*100, bar))
	}
	p.setResults(sb.String())
	p.setStatus("Level sweep complete.")
}

func (p *AccuracySweepPanel) runADCSweep() {
	p.setStatus("Running ADC bit sweep...")
	p.setResults("")

	imgs, lbls := p.getTestSubset(500)
	bits := []int{3, 4, 6, 8, 10}

	net := p.app.network()
	results, err := neural.SweepADCBits(net, imgs, lbls, bits)
	if err != nil {
		p.setStatus(fmt.Sprintf("ADC sweep error: %v", err))
		return
	}

	var sb strings.Builder
	sb.WriteString("ADC Bits  Accuracy  Bar\n")
	sb.WriteString("───────────────────────────────\n")
	for _, r := range results {
		bar := sweepProgressBar(r.Accuracy, 20)
		sb.WriteString(fmt.Sprintf(" %2d-bit   %5.1f%%  %s\n", r.ADCBits, r.Accuracy*100, bar))
	}
	p.setResults(sb.String())
	p.setStatus("ADC sweep complete.")
}

func (p *AccuracySweepPanel) runNoiseSweep() {
	p.setStatus("Running noise sweep...")
	p.setResults("")

	imgs, lbls := p.getTestSubset(500)
	noiseLevels := []float64{0.0, 0.02, 0.05, 0.10, 0.15, 0.20}

	net := p.app.network()
	results, err := neural.SweepNoiseLevel(net, imgs, lbls, noiseLevels)
	if err != nil {
		p.setStatus(fmt.Sprintf("Noise sweep error: %v", err))
		return
	}

	var sb strings.Builder
	sb.WriteString("Noise σ   Accuracy  Bar\n")
	sb.WriteString("────────────────────────────────\n")
	for _, r := range results {
		bar := sweepProgressBar(r.Accuracy, 20)
		sb.WriteString(fmt.Sprintf(" %4.0f%%    %5.1f%%  %s\n", r.NoiseLevel*100, r.Accuracy*100, bar))
	}
	p.setResults(sb.String())
	p.setStatus("Noise sweep complete.")
}

// sweepProgressBar renders a simple ASCII progress bar for accuracy visualization.
func sweepProgressBar(fraction float64, width int) string {
	filled := int(fraction * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}
