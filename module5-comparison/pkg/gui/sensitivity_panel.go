//go:build legacy_fyne

// This file implements an interactive sensitivity analysis panel showing
// which input parameters have the largest impact on the comparison output.
package gui

import (
	"fmt"
	"image/color"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

// maxTornadoBars limits the tornado chart to the most impactful parameters.
const maxTornadoBars = 6

// SensitivityPanel displays input parameter provenance and a tornado chart.
type SensitivityPanel struct {
	widget.BaseWidget
	run        ScenarioRun
	outputName string
}

// NewSensitivityPanel creates a panel for the given run and output metric.
func NewSensitivityPanel(run ScenarioRun, outputName string) *SensitivityPanel {
	sp := &SensitivityPanel{run: run, outputName: outputName}
	sp.ExtendBaseWidget(sp)
	return sp
}

// SetRun updates the panel with a new scenario run and refreshes.
func (sp *SensitivityPanel) SetRun(run ScenarioRun, outputName string) {
	sp.run = run
	sp.outputName = outputName
	sp.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (sp *SensitivityPanel) CreateRenderer() fyne.WidgetRenderer {
	content := sp.buildContent()
	return widget.NewSimpleRenderer(content)
}

func (sp *SensitivityPanel) buildContent() fyne.CanvasObject {
	headerBg := canvas.NewRectangle(color.RGBA{30, 50, 70, 255})
	headerText := canvas.NewText("SENSITIVITY ANALYSIS", color.RGBA{180, 210, 255, 255})
	headerText.TextSize = 14
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.Alignment = fyne.TextAlignCenter
	header := container.NewStack(headerBg, container.NewCenter(headerText))

	if len(sp.run.Inputs) == 0 {
		placeholder := widget.NewLabel("No scenario data. Run a calculation first.")
		return container.NewVBox(header, placeholder)
	}

	paramSection := sp.buildParameterList()
	tornadoSection := sp.buildTornadoChart()

	caveat := widget.NewLabelWithStyle(
		"Tornado bars show +/-10% perturbation impact on "+sp.outputName+". All values are model inputs (TRL4).",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	caveat.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		paramSection,
		widget.NewSeparator(),
		tornadoSection,
		widget.NewSeparator(),
		caveat,
	)
}

func (sp *SensitivityPanel) buildParameterList() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Input Parameters", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	keys := make([]string, 0, len(sp.run.Inputs))
	for k := range sp.run.Inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := container.NewVBox()
	for _, k := range keys {
		tv := sp.run.Inputs[k]
		rows.Add(buildParamRow(tv))
	}

	scrolled := container.NewScroll(rows)
	scrolled.SetMinSize(fyne.NewSize(0, 120))
	return container.NewVBox(title, scrolled)
}

func buildParamRow(tv TaggedValue) fyne.CanvasObject {
	indicator := canvas.NewRectangle(provenanceColor(tv.Tag.Provenance))
	indicator.SetMinSize(fyne.NewSize(10, 10))

	nameLabel := widget.NewLabel(tv.Name)
	nameLabel.Truncation = fyne.TextTruncateEllipsis

	valueLabel := widget.NewLabel(fmt.Sprintf("%.4g %s", tv.Value, tv.Units))
	valueLabel.TextStyle = fyne.TextStyle{Bold: true}

	tagLabel := widget.NewLabel(fmt.Sprintf("[%s c=%.0f%%]", tv.Tag.Provenance, tv.Tag.Confidence*100))

	return container.NewHBox(indicator, nameLabel, layout.NewSpacer(), valueLabel, tagLabel)
}

func (sp *SensitivityPanel) buildTornadoChart() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		fmt.Sprintf("Tornado Chart: %s", sp.outputName),
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
	)

	impacts := SensitivityRanking(sp.run, sp.outputName)
	if len(impacts) == 0 {
		return container.NewVBox(title, widget.NewLabel("No sensitivity data for this output."))
	}

	limit := maxTornadoBars
	if len(impacts) < limit {
		limit = len(impacts)
	}
	impacts = impacts[:limit]

	maxImpact := impacts[0].ImpactPct
	if maxImpact <= 0 {
		maxImpact = 1 // guard against div-by-zero
	}

	const maxBarWidth float32 = 260
	rows := container.NewVBox()
	for _, imp := range impacts {
		rows.Add(buildTornadoRow(imp, maxImpact, maxBarWidth))
	}

	return container.NewVBox(title, rows)
}

func buildTornadoRow(imp SensitivityImpact, maxImpact float64, maxBarWidth float32) fyne.CanvasObject {
	nameLabel := widget.NewLabel(imp.InputName)
	nameLabel.Truncation = fyne.TextTruncateEllipsis

	barWidth := float32(imp.ImpactPct/maxImpact) * maxBarWidth
	if barWidth < 4 {
		barWidth = 4
	}

	barColor := tornadoBarColor(imp.ImpactPct)
	bar := canvas.NewRectangle(barColor)
	bar.SetMinSize(fyne.NewSize(barWidth, 16))

	impactLabel := widget.NewLabel(fmt.Sprintf("%.1f%%", imp.ImpactPct))
	impactLabel.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewHBox(
		container.NewGridWrap(fyne.NewSize(200, 20), nameLabel),
		bar,
		impactLabel,
	)
}

func provenanceColor(p physics.Provenance) color.Color {
	switch p {
	case physics.ProvenanceMeasured:
		return color.RGBA{60, 180, 60, 255} // green
	case physics.ProvenanceCalibrated:
		return color.RGBA{100, 160, 220, 255} // blue
	case physics.ProvenanceEstimated:
		return color.RGBA{220, 180, 60, 255} // amber
	case physics.ProvenancePlaceholder:
		return color.RGBA{220, 80, 80, 255} // red
	default:
		return color.RGBA{160, 160, 160, 255} // gray
	}
}

func tornadoBarColor(impactPct float64) color.Color {
	if impactPct > 8 {
		return color.RGBA{220, 80, 80, 255} // red - high sensitivity
	}
	if impactPct > 3 {
		return color.RGBA{220, 160, 60, 255} // amber - moderate
	}
	return color.RGBA{100, 160, 220, 255} // blue - low
}

// MinSize returns the minimum size for the sensitivity panel.
func (sp *SensitivityPanel) MinSize() fyne.Size {
	return fyne.NewSize(400, 350)
}
