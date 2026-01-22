// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains investor-focused visualizations.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// PhasedStrategyDiagram shows the commercialization strategy phases.
type PhasedStrategyDiagram struct {
	widget.BaseWidget

	mu           sync.RWMutex
	currentPhase int
	animProgress float64
	minSize      fyne.Size

	raster      *canvas.Raster
	titleText   *canvas.Text
	phaseLabels []*canvas.Text
	subtitles   []*canvas.Text
	benefits    []*canvas.Text
}

// NewPhasedStrategyDiagram creates a new strategy diagram.
func NewPhasedStrategyDiagram() *PhasedStrategyDiagram {
	p := &PhasedStrategyDiagram{
		minSize:     fyne.NewSize(450, 100),
		phaseLabels: make([]*canvas.Text, 3),
		subtitles:   make([]*canvas.Text, 3),
		benefits:    make([]*canvas.Text, 3),
	}
	p.ExtendBaseWidget(p)
	return p
}

// SetPhase sets the highlighted phase.
func (p *PhasedStrategyDiagram) SetPhase(phase int) {
	p.mu.Lock()
	p.currentPhase = phase % 3
	p.mu.Unlock()
	p.Refresh()
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
	p.raster = canvas.NewRaster(p.generateBoxes)

	p.titleText = canvas.NewText("COMMERCIALIZATION STRATEGY", color.RGBA{0, 212, 255, 255})
	p.titleText.TextSize = 11
	p.titleText.TextStyle = fyne.TextStyle{Bold: true}

	phases := []string{"PHASE 1", "PHASE 2", "PHASE 3"}
	subs := []string{"NAND Flash", "DRAM", "Full CIM"}
	bens := []string{"Drop-in compatible", "No refresh needed", "80-90% savings"}

	for i := 0; i < 3; i++ {
		p.phaseLabels[i] = canvas.NewText(phases[i], color.RGBA{0, 212, 255, 255})
		p.phaseLabels[i].TextSize = 9
		p.phaseLabels[i].TextStyle = fyne.TextStyle{Bold: true}

		p.subtitles[i] = canvas.NewText(subs[i], color.RGBA{200, 200, 200, 255})
		p.subtitles[i].TextSize = 9

		p.benefits[i] = canvas.NewText(bens[i], color.RGBA{100, 200, 150, 255})
		p.benefits[i].TextSize = 8
	}

	return &phasedStrategyRenderer{widget: p}
}

type phasedStrategyRenderer struct {
	widget *PhasedStrategyDiagram
}

func (r *phasedStrategyRenderer) MinSize() fyne.Size {
	return r.widget.minSize
}

func (r *phasedStrategyRenderer) Layout(size fyne.Size) {
	r.widget.raster.Resize(size)

	r.widget.titleText.Move(fyne.NewPos(size.Width/2-90, 3))

	boxWidth := (size.Width - 80) / 3
	spacing := float32(30)
	startY := float32(25)

	for i := 0; i < 3; i++ {
		boxX := float32(20) + float32(i)*(boxWidth+spacing)
		r.widget.phaseLabels[i].Move(fyne.NewPos(boxX+boxWidth/2-25, startY+5))
		r.widget.subtitles[i].Move(fyne.NewPos(boxX+boxWidth/2-25, startY+18))
		r.widget.benefits[i].Move(fyne.NewPos(boxX+5, startY+45))
	}
}

func (r *phasedStrategyRenderer) Refresh() {
	r.widget.mu.RLock()
	currentPhase := r.widget.currentPhase
	r.widget.mu.RUnlock()

	for i := 0; i < 3; i++ {
		if i == currentPhase {
			r.widget.phaseLabels[i].Color = color.RGBA{0, 255, 255, 255}
		} else {
			r.widget.phaseLabels[i].Color = color.RGBA{0, 212, 255, 255}
		}
		canvas.Refresh(r.widget.phaseLabels[i])
	}

	r.widget.raster.Refresh()
}

func (r *phasedStrategyRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{r.widget.raster, r.widget.titleText}
	for i := 0; i < 3; i++ {
		objects = append(objects, r.widget.phaseLabels[i], r.widget.subtitles[i], r.widget.benefits[i])
	}
	return objects
}

func (r *phasedStrategyRenderer) Destroy() {}

// generateBoxes creates the strategy boxes and arrows.
func (p *PhasedStrategyDiagram) generateBoxes(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 200 || h < 60 {
		return img
	}

	p.mu.RLock()
	currentPhase := p.currentPhase
	animProgress := p.animProgress
	p.mu.RUnlock()

	boxWidth := (w - 80) / 3
	boxHeight := 40
	startY := 20
	spacing := 30

	for i := 0; i < 3; i++ {
		boxX := 20 + i*(boxWidth+spacing)

		isHighlighted := i == currentPhase
		isAnimated := int(animProgress) == i

		var borderColor, fillColor color.RGBA
		if isHighlighted || isAnimated {
			borderColor = color.RGBA{0, 212, 255, 255}
			fillColor = color.RGBA{0, 80, 130, 255}
		} else {
			borderColor = color.RGBA{80, 100, 130, 255}
			fillColor = color.RGBA{40, 50, 70, 255}
		}

		drawBoxFilledInvestor(img, boxX, startY, boxWidth, boxHeight, borderColor, fillColor)

		// Arrow to next phase
		if i < 2 {
			arrowX := boxX + boxWidth + 5
			arrowY := startY + boxHeight/2

			arrowAlpha := uint8(150)
			if int(animProgress) == i {
				pulse := math.Sin(animProgress*math.Pi*2)*0.5 + 0.5
				arrowAlpha = uint8(150 + pulse*105)
			}
			arrowColor := color.RGBA{0, 212, 255, arrowAlpha}

			for ax := 0; ax < spacing-10; ax++ {
				img.Set(arrowX+ax, arrowY, arrowColor)
				img.Set(arrowX+ax, arrowY+1, arrowColor)
			}

			for ay := -4; ay <= 4; ay++ {
				headX := arrowX + spacing - 15 + absInvestor(ay)
				img.Set(headX, arrowY+ay, arrowColor)
			}
		}
	}

	return img
}

func absInvestor(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawBoxFilledInvestor(img *image.RGBA, x, y, width, height int, borderColor, fillColor color.RGBA) {
	for dy := 2; dy < height-2; dy++ {
		for dx := 2; dx < width-2; dx++ {
			img.Set(x+dx, y+dy, fillColor)
		}
	}
	for dx := 0; dx < width; dx++ {
		img.Set(x+dx, y, borderColor)
		img.Set(x+dx, y+1, borderColor)
		img.Set(x+dx, y+height-1, borderColor)
		img.Set(x+dx, y+height-2, borderColor)
	}
	for dy := 0; dy < height; dy++ {
		img.Set(x, y+dy, borderColor)
		img.Set(x+1, y+dy, borderColor)
		img.Set(x+width-1, y+dy, borderColor)
		img.Set(x+width-2, y+dy, borderColor)
	}
}

// AnalogStatesComparison shows binary vs FeCIM memory comparison.
type AnalogStatesComparison struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64
	minSize      fyne.Size

	raster       *canvas.Raster
	binaryTitle  *canvas.Text
	fecimTitle   *canvas.Text
	binaryStats1 *canvas.Text
	binaryStats2 *canvas.Text
	fecimStats1  *canvas.Text
	fecimStats2  *canvas.Text
	taglineText  *canvas.Text
	label0       *canvas.Text
	label1       *canvas.Text
	labelStart   *canvas.Text
	labelEnd     *canvas.Text
}

// NewAnalogStatesComparison creates a new analog states comparison.
func NewAnalogStatesComparison() *AnalogStatesComparison {
	a := &AnalogStatesComparison{
		minSize: fyne.NewSize(400, 130),
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
	a.raster = canvas.NewRaster(a.generateGraphics)

	a.binaryTitle = canvas.NewText("BINARY MEMORY", color.RGBA{200, 100, 100, 255})
	a.binaryTitle.TextSize = 10
	a.binaryTitle.TextStyle = fyne.TextStyle{Bold: true}

	a.fecimTitle = canvas.NewText("FeCIM MEMORY", color.RGBA{100, 200, 150, 255})
	a.fecimTitle.TextSize = 10
	a.fecimTitle.TextStyle = fyne.TextStyle{Bold: true}

	a.binaryStats1 = canvas.NewText("2 states", color.RGBA{180, 180, 180, 255})
	a.binaryStats1.TextSize = 9

	a.binaryStats2 = canvas.NewText("1 bit/cell", color.RGBA{150, 150, 150, 255})
	a.binaryStats2.TextSize = 9

	a.fecimStats1 = canvas.NewText("30 states", color.RGBA{180, 180, 180, 255})
	a.fecimStats1.TextSize = 9

	a.fecimStats2 = canvas.NewText("4.9 bits/cell", color.RGBA{100, 200, 150, 255})
	a.fecimStats2.TextSize = 9

	a.taglineText = canvas.NewText("Same silicon, 5x more information", color.RGBA{0, 212, 255, 255})
	a.taglineText.TextSize = 10
	a.taglineText.TextStyle = fyne.TextStyle{Bold: true}

	a.label0 = canvas.NewText("0", color.RGBA{200, 200, 200, 255})
	a.label0.TextSize = 10

	a.label1 = canvas.NewText("1", color.RGBA{50, 50, 50, 255})
	a.label1.TextSize = 10

	a.labelStart = canvas.NewText("1", color.RGBA{150, 150, 255, 255})
	a.labelStart.TextSize = 8

	a.labelEnd = canvas.NewText("30", color.RGBA{255, 150, 150, 255})
	a.labelEnd.TextSize = 8

	return &analogStatesRenderer{widget: a}
}

type analogStatesRenderer struct {
	widget *AnalogStatesComparison
}

func (r *analogStatesRenderer) MinSize() fyne.Size {
	return r.widget.minSize
}

func (r *analogStatesRenderer) Layout(size fyne.Size) {
	r.widget.raster.Resize(size)
	midX := size.Width / 2

	r.widget.binaryTitle.Move(fyne.NewPos(20, 5))
	r.widget.fecimTitle.Move(fyne.NewPos(midX+20, 5))

	r.widget.label0.Move(fyne.NewPos(42, 42))
	r.widget.label1.Move(fyne.NewPos(82, 42))

	cellY := float32(25)
	cellSize := float32(30)
	r.widget.binaryStats1.Move(fyne.NewPos(30, cellY+cellSize+8))
	r.widget.binaryStats2.Move(fyne.NewPos(30, cellY+cellSize+22))

	r.widget.labelStart.Move(fyne.NewPos(midX+22, cellY+cellSize+5))
	r.widget.labelEnd.Move(fyne.NewPos(size.Width-35, cellY+cellSize+5))

	r.widget.fecimStats1.Move(fyne.NewPos(midX+20, cellY+cellSize+18))
	r.widget.fecimStats2.Move(fyne.NewPos(midX+20, cellY+cellSize+32))

	r.widget.taglineText.Move(fyne.NewPos(size.Width/2-100, size.Height-18))
}

func (r *analogStatesRenderer) Refresh() {
	r.widget.mu.RLock()
	animProgress := r.widget.animProgress
	r.widget.mu.RUnlock()

	// Pulse tagline
	pulse := 0.7 + math.Sin(animProgress*2)*0.3
	r.widget.taglineText.Color = color.RGBA{
		0,
		uint8(212 * pulse),
		uint8(255 * pulse),
		255,
	}

	// Animated bits display
	bitsTarget := 4.9
	bitsDisplay := bitsTarget
	if animProgress < 2.0 {
		bitsDisplay = bitsTarget * (animProgress / 2.0)
	}
	r.widget.fecimStats2.Text = fmt.Sprintf("%.1f bits/cell", bitsDisplay)

	r.widget.raster.Refresh()
	canvas.Refresh(r.widget.taglineText)
	canvas.Refresh(r.widget.fecimStats2)
}

func (r *analogStatesRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.widget.raster,
		r.widget.binaryTitle,
		r.widget.fecimTitle,
		r.widget.label0,
		r.widget.label1,
		r.widget.binaryStats1,
		r.widget.binaryStats2,
		r.widget.labelStart,
		r.widget.labelEnd,
		r.widget.fecimStats1,
		r.widget.fecimStats2,
		r.widget.taglineText,
	}
}

func (r *analogStatesRenderer) Destroy() {}

// generateGraphics creates the analog states graphics.
func (a *AnalogStatesComparison) generateGraphics(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 150 || h < 80 {
		return img
	}

	midX := w / 2

	// Binary cells
	cellSize := 30
	cellY := 25
	for i := 0; i < 2; i++ {
		cellX := 30 + i*(cellSize+5)
		var cellColor color.RGBA
		if i == 0 {
			cellColor = color.RGBA{40, 40, 40, 255}
		} else {
			cellColor = color.RGBA{255, 255, 255, 255}
		}
		for dy := 0; dy < cellSize; dy++ {
			for dx := 0; dx < cellSize; dx++ {
				img.Set(cellX+dx, cellY+dy, cellColor)
			}
		}
		borderColor := color.RGBA{100, 100, 100, 255}
		for dx := 0; dx < cellSize; dx++ {
			img.Set(cellX+dx, cellY, borderColor)
			img.Set(cellX+dx, cellY+cellSize-1, borderColor)
		}
		for dy := 0; dy < cellSize; dy++ {
			img.Set(cellX, cellY+dy, borderColor)
			img.Set(cellX+cellSize-1, cellY+dy, borderColor)
		}
	}

	// FeCIM gradient bar
	gradWidth := w - midX - 40
	gradHeight := 30
	gradY := 25
	cellWidth := gradWidth / 30
	if cellWidth < 1 {
		cellWidth = 1
	}

	for i := 0; i < 30; i++ {
		cellX := midX + 20 + i*cellWidth
		t := float64(i) / 29.0
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
		for dy := 0; dy < gradHeight; dy++ {
			for dx := 0; dx < cellWidth; dx++ {
				if cellX+dx < w-20 {
					img.Set(cellX+dx, gradY+dy, cellColor)
				}
			}
		}
	}

	// Border
	borderColor := color.RGBA{0, 212, 255, 255}
	for dx := 0; dx < gradWidth; dx++ {
		if midX+20+dx < w {
			img.Set(midX+20+dx, gradY, borderColor)
			img.Set(midX+20+dx, gradY+gradHeight-1, borderColor)
		}
	}
	for dy := 0; dy < gradHeight; dy++ {
		img.Set(midX+20, gradY+dy, borderColor)
		if midX+20+gradWidth-1 < w {
			img.Set(midX+20+gradWidth-1, gradY+dy, borderColor)
		}
	}

	return img
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
	title := widget.NewLabelWithStyle("PRECEDENT: WEEBIT NANO", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	quote := widget.NewLabel("\"This company Weebit—this is another memory that came out of my lab... it's selling now on the market with three big customers.\"")
	quote.Wrapping = fyne.TextWrapWord
	quote.TextStyle = fyne.TextStyle{Italic: true}

	attribution := widget.NewLabel("— Dr. external research group")
	attribution.Alignment = fyne.TextAlignTrailing

	check1 := widget.NewLabel("Y Started at TRL 4 (like FeCIM)")
	check2 := widget.NewLabel("Y Now partnered with foundries")
	check3 := widget.NewLabel("Y Proven commercialization path")

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
