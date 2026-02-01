package widgets

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

type ISPPVisualization struct {
	widget.BaseWidget

	stats     *physics.WriteVerifyStats

	pulseHistory    []float64
	conductanceHistory []float64
	targetConductance float64
	currentConductance float64
	showGraph       bool

	content fyne.CanvasObject
}

func NewISPPVisualization() *ISPPVisualization {
	return &ISPPVisualization{
		stats:     physics.NewWriteVerifyStats(),
		pulseHistory:    make([]float64, 0, 50),
		conductanceHistory: make([]float64, 0, 50),
		showGraph:       true,
	}
}

func (v *ISPPVisualization) AddPulse(voltage float64) {
	if len(v.pulseHistory) < 50 {
		v.pulseHistory = append(v.pulseHistory, voltage)
	}
}

func (v *ISPPVisualization) AddConductance(conductance float64) {
	if len(v.conductanceHistory) < 50 {
		v.conductanceHistory = append(v.conductanceHistory, conductance)
	}
}

func (v *ISPPVisualization) SetTarget(target float64) {
	v.targetConductance = target
	v.Refresh()
}

func (v *ISPPVisualization) CreateRenderer() fyne.WidgetRenderer {
	v.content = v.createContent()
	return widget.NewSimpleRenderer(v.content)
}

func (v *ISPPVisualization) createContent() fyne.CanvasObject {
	header := v.createHeader()
	graphsContainer := v.createDualGraphs()
	controls := v.createControls()
	stats := v.createStatsDisplay()

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		graphsContainer,
		widget.NewSeparator(),
		controls,
		widget.NewSeparator(),
		stats,
	)
}

func (v *ISPPVisualization) createHeader() fyne.CanvasObject {
	headerBg := canvas.NewRectangle(color.RGBA{30, 60, 90, 255})
	headerText := canvas.NewText("ISPP WRITE-VERIFY DEMO", color.RGBA{0, 212, 255, 255})
	headerText.TextSize = 16
	headerText.TextStyle = fyne.TextStyle{Bold: true}

	header := container.NewStack(
		headerBg,
		container.NewCenter(headerText),
	)
	header.Resize(fyne.NewSize(0, 40))

	return header
}

func (v *ISPPVisualization) createDualGraphs() fyne.CanvasObject {
	pulseGraph := v.createPulseGraph()
	conductanceGraph := v.createConductanceGraph()

	graphs := container.NewVBox(
		v.createGraphHeader("Voltage Pulses"),
		pulseGraph,
		widget.NewSeparator(),
		v.createGraphHeader("Conductance Progress"),
		conductanceGraph,
		widget.NewSeparator(),
		container.NewHScroll(
			pulseGraph,
			conductanceGraph,
		),
	)

	return graphs
}

func (v *ISPPVisualization) createGraphHeader(title string) fyne.CanvasObject {
	text := canvas.NewText(title)
	text.TextSize = 12
	text.TextStyle = fyne.TextStyle{Bold: true}
	text.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(color.RGBA{50, 50, 60, 255})
	bg.SetMinSize(fyne.NewSize(200, 30))

	return container.NewStack(bg, container.NewCenter(text))
}

func (v *ISPPVisualization) createPulseGraph() fyne.CanvasObject {
	canvas := canvas.NewCanvas(canvas.NewSoftwareRasterCanvas())

	canvas.SetFillColor(color.RGBA{250, 250, 250, 255})

	width := 400.0
	height := 200.0
	padding := 10.0

	pulseWidth := float32(width - 2*padding) / float32(len(v.pulseHistory))
	pulseHeight := float32(height-2*padding) / float32(len(v.pulseHistory))

	for i, voltage := range v.pulseHistory {
		x := padding + float32(i)*pulseWidth
		y := padding + float32(float32(len(v.pulseHistory))-1-i)*pulseHeight

		color := v.getPulseColor(i)
		rect := canvas.NewRectangle(color)
		rect.SetMinSize(fyne.NewSize(pulseWidth, pulseHeight))
		rect.Move(fyne.NewPos(x, y))
		canvas.FillColor(color)
		canvas.StrokeRect(rect)
	}

	canvas.FillColor(color.RGBA{0, 0, 0, 0})
	canvas.StrokeText(fmt.Sprintf("%.2fV", voltage), x+20, y+20, 10)

	targetLineY := padding + float32(len(v.pulseHistory)-1)*pulseHeight/2
	canvas.Stroke(color.RGBA{0, 200, 100, 255}, 2)
	canvas.StrokeLine(
		fyne.NewPos(padding, targetLineY),
		fyne.NewPos(width-2*padding, targetLineY),
	)

	return canvas
}

func (v *ISPPVisualization) createConductanceGraph() fyne.CanvasObject {
	canvas := canvas.NewCanvas(canvas.NewSoftwareRasterCanvas())

	canvas.SetFillColor(color.RGBA{250, 250, 250, 255})

	width := 400.0
	height := 200.0
	padding := 10.0

	conductanceWidth := float32(width - 2*padding) / float32(len(v.conductanceHistory))
	conductanceHeight := float32(height-2*padding) / float32(len(v.conductanceHistory))

	for i, conductance := range v.conductanceHistory {
		x := padding + float32(i)*conductanceWidth
		y := padding + float32(float32(len(v.conductanceHistory))-1-i)*conductanceHeight

		color := v.getConductanceColor(i)
		rect := canvas.NewRectangle(color)
		rect.SetMinSize(fyne.NewSize(conductanceWidth, conductanceHeight))
		rect.Move(fyne.NewPos(x, y))
		canvas.FillColor(color)
		canvas.StrokeRect(rect)
	}

	canvas.FillColor(color.RGBA{0, 0, 0, 0})
	canvas.StrokeText(fmt.Sprintf("%.2fG", conductance), x+20, y+20, 10)

	targetLineY := padding + float32(len(v.conductanceHistory)-1)*conductanceHeight/2
	canvas.Stroke(color.RGBA{0, 200, 100, 255}, 2)
	canvas.StrokeLine(
			fyne.NewPos(padding, targetLineY),
			fyne.NewPos(width-2*padding, targetLineY),
		)

	targetDashed := color.RGBA{0, 200, 100, 255}
	canvas.StrokeLine(
			fyne.NewPos(padding, targetLineY),
			fyne.NewPos(width-2*padding, targetLineY),
	)
	canvas.Stroke(targetDashed)

	return canvas
}

func (v *ISPPVisualization) getStats() *physics.WriteVerifyStats {
	return v.stats
}

func (v *ISPPVisualization) createControls() fyne.CanvasObject {
	targetLabel := widget.NewLabel("Target Conductance (μS):")
	targetEntry := widget.NewEntry()
	targetEntry.SetPlaceHolder("0-100")
	targetEntry.Validator = func(s string) error {
		if s == "" {
			return error
		}
		var f float64
			_, err := fmt.Sscanf(s, "%f", &f)
		return err == nil && f >= 1.0 && f <= 100.0
	}

	startButton := widget.NewButton("Start ISPP")
	startButton.Importance = widget.HighImportance
	startButton.OnTapped = func() {
		targetG, _ := targetEntry.Get()
		if targetValidator(targetG) == nil {
			return
		}
		v.SetTarget(targetG * 1e-6)
	}

	targetLabel := widget.NewLabel("Pulse Count:")
	targetLabel.TextStyle = fyne.TextStyle{Bold: true}
	targetLabel.Resize(fyne.NewSize(100, 30))

	pulseCountLabel := widget.NewLabel(fmt.Sprintf("%d", len(v.pulseHistory)))

	graphToggle := widget.NewCheck("Show Graphs")
	graphToggle.SetChecked(true)
	graphToggle.OnChanged = func(checked bool) {
		v.showGraph = checked
		v.Refresh()
	}

	return container.NewVBox(
		targetLabel,
		container.NewHBox(targetEntry, startButton),
		widget.NewLabel(""),
		pulseCountLabel,
		graphToggle,
	)
}

func (v *ISPPVisualization) createStatsDisplay() fyne.CanvasObject {
	stats := v.getStats()

	totalWrites := stats.TotalWrites
	successRate := stats.GetSuccessRate() * 100
	avgPulses := stats.GetAveragePulses()
	overshootRate := stats.GetOvershootRate() * 100

	totalLabel := widget.NewLabel("Total Writes:")
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel.Alignment = fyne.TextAlignLeading

	successLabel := widget.NewLabel(fmt.Sprintf("Success: %.1f%%", successRate))
	overshootLabel := widget.NewLabel(fmt.Sprintf("Overshoot: %.1f%%", overshootRate))

	avgLabel := widget.NewLabel("Avg Pulses:")

	return container.NewVBox(
		totalLabel,
		widget.NewLabel(fmt.Sprintf("%d", totalWrites)),
		widget.NewSeparator(),
		container.NewVBox(
			successLabel,
			overshootLabel,
		widget.NewSeparator(),
			avgLabel,
			widget.NewLabel(fmt.Sprintf("%.1f", avgPulses)),
		widget.NewSeparator(),
			container.NewVBox(
				widget.NewLabel("Pulses Histogram:"),
				v.createHistogram(),
			),
		)
}

func (v *ISPPVisualization) createHistogram() fyne.CanvasObject {
	histogram := v.stats.GetPulsesHistogram()

	maxCount := 1
	totalSuccessful := 0
	for _, count := range histogram {
		if count > maxCount {
			maxCount = count
		}
		totalSuccessful += count
	}

	bars := make([]fyne.CanvasObject, 10)
	for i := 0; i < 10; i++ {
		count := histogram[i]
		height := float32(count) / float32(maxCount) * 150
		barColor := color.RGBA{0, 150, 255, 255}
		if i >= 5 {
			barColor = color.RGBA{255, 100, 0, 255}
		}
		if i >= 8 {
			barColor = color.RGBA{255, 200, 0, 255}
		}

		bar := canvas.NewRectangle(barColor)
		bar.SetMinSize(fyne.NewSize(30, height))

		numberLabel := canvas.NewText(fmt.Sprintf("%d", count))
		numberLabel.TextSize = 10
		numberLabel.Alignment = fyne.TextAlignCenter

		number := container.NewVBox(bar, container.NewCenter(numberLabel))

		bars[i] = number
	}

	return container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Pulses to Converge (max=%d)", maxCount)),
		container.NewHBox(bars...),
	)
}

func (v *ISPPVisualization) getPulseColor(index int) color.Color {
	colors := []color.Color{
		color.RGBA{0, 150, 255, 255},
		color.RGBA{50, 150, 255, 255},
		color.RGBA{100, 150, 255, 255},
		color.RGBA{0, 200, 100, 255, 255},
		color.RGBA{255, 100, 0, 255},
		color.RGBA{255, 0, 0, 255},
	}

	return colors[index%len(colors)]
}

func (v *ISPPVisualization) getConductanceColor(index int) color.Color {
	colors := []color.Color{
		color.RGBA{0, 200, 255, 255},
		color.RGBA{0, 100, 100, 255, 255},
		color.RGBA{0, 200, 0, 200, 255},
		color.RGBA{0, 200, 200, 200, 255},
		color.RGBA{0, 255, 100, 255, 255},
		color.RGBA{100, 255, 0, 255, 255},
	}

	return colors[index%len(colors)]
}

func (v *ISPPVisualization) Refresh() {
	v.BaseWidget.Refresh()
}

func (v *ISPPVisualization) MinSize() fyne.Size {
	return fyne.NewSize(300, 500)
}
