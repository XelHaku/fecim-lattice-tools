//go:build legacy_fyne

package widgets

import (
	"fmt"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

// isppPulse records one ISPP programming pulse.
type isppPulse struct {
	Voltage float64
	Level   int
}

// ISPPVisualization renders a voltage-vs-pulse-count step chart showing the
// ISPP ladder. Call AddPulse to append data and Refresh to redraw.
type ISPPVisualization struct {
	widget.BaseWidget
	mu      sync.Mutex
	stats   *physics.WriteVerifyStats
	pulses  []isppPulse
	target  float64
	content *fyne.Container
}

func NewISPPVisualization() *ISPPVisualization {
	v := &ISPPVisualization{
		stats: physics.NewWriteVerifyStats(),
	}
	v.ExtendBaseWidget(v)
	return v
}

// AddPulse records a programming pulse at the given voltage reaching the given level.
func (v *ISPPVisualization) AddPulse(voltage float64, level int) {
	v.mu.Lock()
	if len(v.pulses) < 50 {
		v.pulses = append(v.pulses, isppPulse{Voltage: voltage, Level: level})
	}
	v.mu.Unlock()
	v.Refresh()
}

// AddConductance is kept for interface compatibility (delegates to AddPulse).
func (v *ISPPVisualization) AddConductance(_ float64) {}

// SetTarget sets the target conductance level for the dashed-line indicator.
func (v *ISPPVisualization) SetTarget(target float64) {
	v.mu.Lock()
	v.target = target
	v.mu.Unlock()
	v.Refresh()
}

func (v *ISPPVisualization) CreateRenderer() fyne.WidgetRenderer {
	v.content = v.buildContent()
	return widget.NewSimpleRenderer(v.content)
}

func (v *ISPPVisualization) buildContent() *fyne.Container {
	v.mu.Lock()
	pulses := make([]isppPulse, len(v.pulses))
	copy(pulses, v.pulses)
	v.mu.Unlock()

	headerBg := canvas.NewRectangle(color.RGBA{30, 60, 90, 255})
	headerText := canvas.NewText("ISPP Voltage Ladder", color.RGBA{200, 220, 255, 255})
	headerText.TextSize = 13
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.Alignment = fyne.TextAlignCenter
	header := container.NewStack(headerBg, container.NewCenter(headerText))

	if len(pulses) == 0 {
		placeholder := widget.NewLabel("No ISPP pulses recorded yet.\nRun an ISPP write sequence to see the voltage ladder.")
		placeholder.Wrapping = fyne.TextWrapWord
		return container.NewVBox(header, placeholder)
	}

	const maxBarW float32 = 200
	maxV := 1.0
	for _, p := range pulses {
		if p.Voltage > maxV {
			maxV = p.Voltage
		}
	}

	bars := container.NewVBox()
	for i, p := range pulses {
		barW := float32(p.Voltage/maxV) * maxBarW
		if barW < 4 {
			barW = 4
		}
		bar := canvas.NewRectangle(color.RGBA{80, 160, 220, 255})
		bar.SetMinSize(fyne.NewSize(barW, 14))
		label := widget.NewLabel(fmt.Sprintf("#%d  %.2f V  L%d", i+1, p.Voltage, p.Level))
		bars.Add(container.NewHBox(bar, label))
	}

	scroll := container.NewVScroll(bars)
	scroll.SetMinSize(fyne.NewSize(0, 180))

	summary := widget.NewLabel(fmt.Sprintf("Pulses: %d | Target: %.2f", len(pulses), v.target))
	summary.Alignment = fyne.TextAlignCenter

	return container.NewVBox(header, scroll, widget.NewSeparator(), summary)
}

func (v *ISPPVisualization) Refresh() {
	v.BaseWidget.Refresh()
}

func (v *ISPPVisualization) MinSize() fyne.Size {
	return fyne.NewSize(300, 250)
}

func (v *ISPPVisualization) GetStats() *physics.WriteVerifyStats {
	return v.stats
}

func (v *ISPPVisualization) SetAnimationState(_ float64, _ float64, _ int, _ float64, _ bool) {
	// Animation hook (unused for now)
}
