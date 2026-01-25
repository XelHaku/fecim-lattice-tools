// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains investor-focused visualizations.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// PhasedStrategyDiagram shows the commercialization strategy phases using Fyne widgets.
type PhasedStrategyDiagram struct {
	widget.BaseWidget

	mu           sync.RWMutex
	currentPhase int
	animProgress float64
	minSize      fyne.Size

	container  *fyne.Container
	phaseBoxes []*canvas.Rectangle
	arrows     []*canvas.Text
}

// NewPhasedStrategyDiagram creates a new strategy diagram.
func NewPhasedStrategyDiagram() *PhasedStrategyDiagram {
	p := &PhasedStrategyDiagram{
		minSize:    fyne.NewSize(350, 80),
		phaseBoxes: make([]*canvas.Rectangle, 3),
		arrows:     make([]*canvas.Text, 2),
	}
	p.ExtendBaseWidget(p)
	return p
}

// SetPhase sets the highlighted phase.
func (p *PhasedStrategyDiagram) SetPhase(phase int) {
	p.mu.Lock()
	p.currentPhase = phase % 3
	p.mu.Unlock()
	fyne.Do(func() {
		p.Refresh()
	})
}

// UpdateAnimation advances the animation.
func (p *PhasedStrategyDiagram) UpdateAnimation(dt float64) {
	p.mu.Lock()
	p.animProgress += dt * 0.5
	if p.animProgress > 3.0 {
		p.animProgress = 0
	}
	p.mu.Unlock()
}

// MinSize returns minimum size.
func (p *PhasedStrategyDiagram) MinSize() fyne.Size {
	return p.minSize
}

// CreateRenderer implements fyne.Widget.
func (p *PhasedStrategyDiagram) CreateRenderer() fyne.WidgetRenderer {
	phases := []struct {
		name  string
		sub   string
		color color.RGBA
	}{
		{"P1", "NAND", color.RGBA{180, 80, 80, 255}},
		{"P2", "DRAM", color.RGBA{80, 120, 180, 255}},
		{"P3", "CIM", color.RGBA{80, 180, 120, 255}},
	}

	var phaseWidgets []fyne.CanvasObject

	for i, phase := range phases {
		p.phaseBoxes[i] = canvas.NewRectangle(phase.color)
		p.phaseBoxes[i].SetMinSize(fyne.NewSize(50, 35))

		phaseLabel := widget.NewLabel(phase.name + "\n" + phase.sub)
		phaseStack := container.NewStack(p.phaseBoxes[i], container.NewCenter(phaseLabel))
		phaseWidgets = append(phaseWidgets, phaseStack)

		if i < 2 {
			p.arrows[i] = canvas.NewText("→", color.RGBA{0, 212, 255, 255})
			p.arrows[i].TextSize = 16
			phaseWidgets = append(phaseWidgets, p.arrows[i])
		}
	}

	p.container = container.NewHBox(phaseWidgets...)
	return widget.NewSimpleRenderer(p.container)
}

// Refresh updates the widget display.
func (p *PhasedStrategyDiagram) Refresh() {
	p.mu.RLock()
	currentPhase := p.currentPhase
	animProgress := p.animProgress
	p.mu.RUnlock()

	if p.phaseBoxes[0] == nil {
		return
	}

	// Highlight current phase
	for i := 0; i < 3; i++ {
		if i == currentPhase || int(animProgress) == i {
			p.phaseBoxes[i].StrokeWidth = 3
			p.phaseBoxes[i].StrokeColor = color.RGBA{0, 255, 255, 255}
		} else {
			p.phaseBoxes[i].StrokeWidth = 1
			p.phaseBoxes[i].StrokeColor = color.RGBA{80, 100, 130, 255}
		}
		canvas.Refresh(p.phaseBoxes[i])
	}

	// Pulse arrows
	for i := 0; i < 2; i++ {
		if int(animProgress) == i {
			pulse := 0.5 + math.Sin(animProgress*math.Pi*4)*0.5
			p.arrows[i].Color = color.RGBA{
				0,
				uint8(212 + 43*pulse),
				255,
				255,
			}
		} else {
			p.arrows[i].Color = color.RGBA{0, 212, 255, 200}
		}
		canvas.Refresh(p.arrows[i])
	}
}

// AnalogStatesComparison shows binary vs FeCIM memory comparison using Fyne widgets.
type AnalogStatesComparison struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64
	minSize      fyne.Size

	container    *fyne.Container
	fecimBits    *widget.Label
	taglineText  *canvas.Text
	gradientRects []*canvas.Rectangle
}

// NewAnalogStatesComparison creates a new analog states comparison.
func NewAnalogStatesComparison() *AnalogStatesComparison {
	a := &AnalogStatesComparison{
		minSize:       fyne.NewSize(350, 100),
		gradientRects: make([]*canvas.Rectangle, 10), // Use 10 cells for visual
	}
	a.ExtendBaseWidget(a)
	return a
}

// UpdateAnimation advances the animation.
func (a *AnalogStatesComparison) UpdateAnimation(dt float64) {
	a.mu.Lock()
	a.animProgress += dt
	a.mu.Unlock()
}

// MinSize returns minimum size.
func (a *AnalogStatesComparison) MinSize() fyne.Size {
	return a.minSize
}

// CreateRenderer implements fyne.Widget.
func (a *AnalogStatesComparison) CreateRenderer() fyne.WidgetRenderer {
	// Left: Binary memory - compact
	cell0 := canvas.NewRectangle(color.RGBA{40, 40, 40, 255})
	cell0.SetMinSize(fyne.NewSize(25, 25))
	cell1 := canvas.NewRectangle(color.RGBA{255, 255, 255, 255})
	cell1.SetMinSize(fyne.NewSize(25, 25))
	binaryCells := container.NewHBox(cell0, cell1)

	binaryCol := container.NewVBox(
		widget.NewLabel("Binary"),
		container.NewCenter(binaryCells),
		widget.NewLabel("2 states"),
	)

	// Middle: VS
	vsText := canvas.NewText("VS", color.RGBA{0, 212, 255, 255})
	vsText.TextSize = 14
	vsText.TextStyle = fyne.TextStyle{Bold: true}

	// Right: FeCIM gradient - compact
	var gradientWidgets []fyne.CanvasObject
	for i := 0; i < 10; i++ {
		t := float64(i) / 9.0
		var cellColor color.RGBA
		if t < 0.5 {
			t2 := t * 2
			cellColor = color.RGBA{
				uint8(80 + t2*175),
				uint8(120 + t2*135),
				255,
				255,
			}
		} else {
			t2 := (t - 0.5) * 2
			cellColor = color.RGBA{
				255,
				uint8(255 - t2*175),
				uint8(255 - t2*175),
				255,
			}
		}
		a.gradientRects[i] = canvas.NewRectangle(cellColor)
		a.gradientRects[i].SetMinSize(fyne.NewSize(12, 20))
		gradientWidgets = append(gradientWidgets, a.gradientRects[i])
	}
	gradientRow := container.NewHBox(gradientWidgets...)

	a.fecimBits = widget.NewLabel("30 states")
	a.fecimBits.Alignment = fyne.TextAlignCenter

	fecimCol := container.NewVBox(
		widget.NewLabel("FeCIM"),
		gradientRow,
		a.fecimBits,
	)

	mainRow := container.NewHBox(
		binaryCol,
		layout.NewSpacer(),
		container.NewCenter(vsText),
		layout.NewSpacer(),
		fecimCol,
	)

	// Tagline - emphasize information density advantage
	a.taglineText = canvas.NewText("15× information density vs binary", color.RGBA{0, 212, 255, 255})
	a.taglineText.TextSize = 10
	a.taglineText.TextStyle = fyne.TextStyle{Bold: true}

	a.container = container.NewVBox(
		mainRow,
		container.NewCenter(a.taglineText),
	)

	return widget.NewSimpleRenderer(a.container)
}

// Refresh updates the widget display.
func (a *AnalogStatesComparison) Refresh() {
	a.mu.RLock()
	animProgress := a.animProgress
	a.mu.RUnlock()

	if a.taglineText == nil {
		return
	}

	// Pulse tagline
	pulse := 0.7 + math.Sin(animProgress*2)*0.3
	a.taglineText.Color = color.RGBA{
		0,
		uint8(212 * pulse),
		uint8(255 * pulse),
		255,
	}

	// Animated bits display - show information content
	bitsTarget := 4.9 // log2(30) ≈ 4.91 bits
	bitsDisplay := bitsTarget
	if animProgress < 2.0 {
		bitsDisplay = bitsTarget * (animProgress / 2.0)
	}
	a.fecimBits.SetText(fmt.Sprintf("30 states → %.1f bits/cell", bitsDisplay))

	canvas.Refresh(a.taglineText)
}

// WeebitNanoCard shows the Weebit Nano precedent.
type WeebitNanoCard struct {
	widget.BaseWidget
}

// NewWeebitNanoCard creates a new Weebit card.
func NewWeebitNanoCard() *WeebitNanoCard {
	wc := &WeebitNanoCard{}
	wc.ExtendBaseWidget(wc)
	return wc
}

// MinSize returns minimum size.
func (wc *WeebitNanoCard) MinSize() fyne.Size {
	return fyne.NewSize(250, 160)
}

// CreateRenderer implements fyne.Widget.
func (wc *WeebitNanoCard) CreateRenderer() fyne.WidgetRenderer {
	title := widget.NewLabelWithStyle("COMMERCIALIZATION PRECEDENT", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	quote := widget.NewLabel("\"Weebit Nano—another memory technology from my lab—is now selling on the market with three major customers.\"")
	quote.Wrapping = fyne.TextWrapWord
	quote.TextStyle = fyne.TextStyle{Italic: true}

	attribution := widget.NewLabel("— Dr. external research group, COSM 2025")
	attribution.Alignment = fyne.TextAlignTrailing

	check1 := widget.NewLabel("✓ Started at TRL 4 (like FeCIM)")
	check2 := widget.NewLabel("✓ Partnered with production foundries")
	check3 := widget.NewLabel("✓ Validated commercialization pathway")

	checks := container.NewVBox(check1, check2, check3)

	stockNote := widget.NewLabel("ASX: WBT")
	stockNote.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		quote,
		attribution,
		widget.NewSeparator(),
		checks,
		stockNote,
	)

	return widget.NewSimpleRenderer(content)
}
