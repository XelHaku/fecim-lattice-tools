// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains hero visualizations for the technical briefing.
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

// AnimatedEnergyRace shows animated energy comparison bars.
type AnimatedEnergyRace struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	cpuEnergy    float64 // Target: 1000 fJ
	gpuEnergy    float64 // Target: 100 fJ
	fecimEnergy  float64 // Target: 10 fJ
	showWinner   bool    // Pulse FeCIM when true
	logScale     bool    // Use logarithmic scale
	pulsePhase   float64 // For winner pulse animation
	minSize      fyne.Size

	// UI elements
	raster        *canvas.Raster
	titleText     *canvas.Text
	cpuLabel      *canvas.Text
	gpuLabel      *canvas.Text
	fecimLabel    *canvas.Text
	cpuValue      *canvas.Text
	gpuValue      *canvas.Text
	fecimValue    *canvas.Text
	headlineText  *canvas.Text
	sourceText    *canvas.Text
}

// NewAnimatedEnergyRace creates a new energy race visualization.
func NewAnimatedEnergyRace() *AnimatedEnergyRace {
	e := &AnimatedEnergyRace{
		cpuEnergy:   1000,
		gpuEnergy:   100,
		fecimEnergy: 10,
		logScale:    false,
		minSize:     fyne.NewSize(550, 180),
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetLogScale enables/disables logarithmic scale.
func (e *AnimatedEnergyRace) SetLogScale(log bool) {
	e.mu.Lock()
	e.logScale = log
	e.mu.Unlock()
	e.Refresh()
}

// UpdateAnimation advances the animation by dt seconds.
func (e *AnimatedEnergyRace) UpdateAnimation(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.animProgress < 1.0 {
		e.animProgress += dt * 0.5
		if e.animProgress > 1.0 {
			e.animProgress = 1.0
			e.showWinner = true
		}
	}

	if e.showWinner {
		e.pulsePhase += dt * 3.0
	}
}

// Reset resets the animation to start.
func (e *AnimatedEnergyRace) Reset() {
	e.mu.Lock()
	e.animProgress = 0
	e.showWinner = false
	e.pulsePhase = 0
	e.mu.Unlock()
	e.Refresh()
}

// MinSize returns minimum size.
func (e *AnimatedEnergyRace) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *AnimatedEnergyRace) CreateRenderer() fyne.WidgetRenderer {
	e.raster = canvas.NewRaster(e.generateBars)

	e.titleText = canvas.NewText("ENERGY PER MAC OPERATION", color.RGBA{0, 212, 255, 255})
	e.titleText.TextSize = 14
	e.titleText.TextStyle = fyne.TextStyle{Bold: true}

	e.cpuLabel = canvas.NewText("CPU+DRAM", color.RGBA{200, 200, 200, 255})
	e.cpuLabel.TextSize = 11

	e.gpuLabel = canvas.NewText("GPU+HBM", color.RGBA{200, 200, 200, 255})
	e.gpuLabel.TextSize = 11

	e.fecimLabel = canvas.NewText("FeCIM", color.RGBA{200, 200, 200, 255})
	e.fecimLabel.TextSize = 11

	e.cpuValue = canvas.NewText("1000 fJ", color.RGBA{200, 100, 100, 255})
	e.cpuValue.TextSize = 11

	e.gpuValue = canvas.NewText("100 fJ", color.RGBA{200, 180, 100, 255})
	e.gpuValue.TextSize = 11

	e.fecimValue = canvas.NewText("10 fJ", color.RGBA{100, 200, 150, 255})
	e.fecimValue.TextSize = 11

	e.headlineText = canvas.NewText("100x LESS ENERGY", color.RGBA{0, 212, 255, 255})
	e.headlineText.TextSize = 12
	e.headlineText.TextStyle = fyne.TextStyle{Bold: true}

	e.sourceText = canvas.NewText("* FeCIM: Dr. Tour claims (TRL 4, not verified)", color.RGBA{150, 150, 150, 255})
	e.sourceText.TextSize = 9

	return &energyRaceRenderer{widget: e}
}

type energyRaceRenderer struct {
	widget *AnimatedEnergyRace
}

func (r *energyRaceRenderer) MinSize() fyne.Size {
	return r.widget.minSize
}

func (r *energyRaceRenderer) Layout(size fyne.Size) {
	r.widget.raster.Resize(size)

	// Title at top center
	r.widget.titleText.Move(fyne.NewPos(size.Width/2-100, 5))

	// Bar labels on left
	barHeight := (size.Height - 60) / 3
	startY := float32(35)

	r.widget.cpuLabel.Move(fyne.NewPos(10, startY+barHeight/2-6))
	r.widget.gpuLabel.Move(fyne.NewPos(10, startY+barHeight+10+barHeight/2-6))
	r.widget.fecimLabel.Move(fyne.NewPos(10, startY+2*(barHeight+10)+barHeight/2-6))

	// Value labels on right
	r.widget.cpuValue.Move(fyne.NewPos(size.Width-70, startY+barHeight/2-6))
	r.widget.gpuValue.Move(fyne.NewPos(size.Width-70, startY+barHeight+10+barHeight/2-6))
	r.widget.fecimValue.Move(fyne.NewPos(size.Width-70, startY+2*(barHeight+10)+barHeight/2-6))

	// Headline at bottom center
	r.widget.headlineText.Move(fyne.NewPos(size.Width/2-70, size.Height-35))

	// Source at bottom
	r.widget.sourceText.Move(fyne.NewPos(10, size.Height-15))
}

func (r *energyRaceRenderer) Refresh() {
	r.widget.mu.RLock()
	progress := r.widget.animProgress
	showWinner := r.widget.showWinner
	pulsePhase := r.widget.pulsePhase
	r.widget.mu.RUnlock()

	// Update value labels
	r.widget.cpuValue.Text = fmt.Sprintf("%.0f fJ", 1000*progress)
	r.widget.gpuValue.Text = fmt.Sprintf("%.0f fJ", 100*progress)
	r.widget.fecimValue.Text = fmt.Sprintf("%.1f fJ", 10*progress)

	// Pulse headline
	if showWinner {
		pulse := 0.7 + math.Sin(pulsePhase)*0.3
		r.widget.headlineText.Color = color.RGBA{
			0,
			uint8(212 * pulse),
			uint8(255 * pulse),
			255,
		}
	} else {
		r.widget.headlineText.Color = color.RGBA{0, 0, 0, 0} // Hidden
	}

	r.widget.raster.Refresh()
	canvas.Refresh(r.widget.titleText)
	canvas.Refresh(r.widget.cpuValue)
	canvas.Refresh(r.widget.gpuValue)
	canvas.Refresh(r.widget.fecimValue)
	canvas.Refresh(r.widget.headlineText)
}

func (r *energyRaceRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.widget.raster,
		r.widget.titleText,
		r.widget.cpuLabel,
		r.widget.gpuLabel,
		r.widget.fecimLabel,
		r.widget.cpuValue,
		r.widget.gpuValue,
		r.widget.fecimValue,
		r.widget.headlineText,
		r.widget.sourceText,
	}
}

func (r *energyRaceRenderer) Destroy() {}

// generateBars creates just the bar graphics (no text).
func (e *AnimatedEnergyRace) generateBars(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 100 || h < 80 {
		return img
	}

	e.mu.RLock()
	progress := e.animProgress
	showWinner := e.showWinner
	pulsePhase := e.pulsePhase
	e.mu.RUnlock()

	labelWidth := 80
	valueWidth := 80
	barAreaWidth := w - labelWidth - valueWidth - 20
	barHeight := (h - 60) / 3
	barSpacing := 10
	startX := labelWidth + 5
	startY := 35

	bars := []struct {
		energy float64
		color  color.RGBA
	}{
		{1000, color.RGBA{200, 100, 100, 255}},
		{100, color.RGBA{200, 180, 100, 255}},
		{10, color.RGBA{100, 200, 150, 255}},
	}

	maxEnergy := 1000.0

	for i, bar := range bars {
		y := startY + i*(barHeight+barSpacing)

		barWidth := int(float64(barAreaWidth) * (bar.energy / maxEnergy) * progress)

		// Track
		trackColor := color.RGBA{40, 50, 70, 255}
		for dy := 0; dy < barHeight; dy++ {
			for dx := 0; dx < barAreaWidth; dx++ {
				img.Set(startX+dx, y+dy, trackColor)
			}
		}

		// Bar
		barColor := bar.color
		if i == 2 && showWinner {
			pulse := math.Sin(pulsePhase) * 0.3
			barColor = color.RGBA{
				uint8(min(255, int(float64(bar.color.R)*(1+pulse)))),
				uint8(min(255, int(float64(bar.color.G)*(1+pulse)))),
				uint8(min(255, int(float64(bar.color.B)*(1+pulse)))),
				255,
			}
		}
		for dy := 0; dy < barHeight; dy++ {
			for dx := 0; dx < barWidth; dx++ {
				img.Set(startX+dx, y+dy, barColor)
			}
		}
	}

	return img
}

// MemoryWallAnimation shows data movement visualization.
type MemoryWallAnimation struct {
	widget.BaseWidget

	mu            sync.RWMutex
	packets       []Packet
	dataMovements int
	simTime       float64
	minSize       fyne.Size

	raster         *canvas.Raster
	vonNeumannText *canvas.Text
	cimText        *canvas.Text
	cpuText        *canvas.Text
	dramText       *canvas.Text
	computeText1   *canvas.Text
	computeText2   *canvas.Text
	vsText         *canvas.Text
	counterText    *canvas.Text
	wasteText      *canvas.Text
	zeroMoveText   *canvas.Text
	zeroWasteText  *canvas.Text
}

// Packet represents a data packet moving between CPU and memory.
type Packet struct {
	x, y   float64
	vx     float64
	active bool
}

// NewMemoryWallAnimation creates a new memory wall visualization.
func NewMemoryWallAnimation() *MemoryWallAnimation {
	m := &MemoryWallAnimation{
		packets: make([]Packet, 0, 10),
		minSize: fyne.NewSize(500, 130),
	}
	for i := 0; i < 5; i++ {
		m.packets = append(m.packets, Packet{
			x:      float64(50 + i*30),
			y:      float64(50 + (i%3)*15),
			vx:     100 + float64(i*20),
			active: true,
		})
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateAnimation advances the animation.
func (m *MemoryWallAnimation) UpdateAnimation(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.simTime += dt

	for i := range m.packets {
		if !m.packets[i].active {
			continue
		}
		m.packets[i].x += m.packets[i].vx * dt
		if m.packets[i].x > 180 {
			m.packets[i].vx = -m.packets[i].vx
			m.dataMovements++
		} else if m.packets[i].x < 50 {
			m.packets[i].vx = -m.packets[i].vx
			m.dataMovements++
		}
	}
}

// Reset resets the animation.
func (m *MemoryWallAnimation) Reset() {
	m.mu.Lock()
	m.simTime = 0
	m.dataMovements = 0
	for i := range m.packets {
		m.packets[i].x = float64(50 + i*30)
		m.packets[i].vx = 100 + float64(i*20)
	}
	m.mu.Unlock()
	m.Refresh()
}

// MinSize returns minimum size.
func (m *MemoryWallAnimation) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *MemoryWallAnimation) CreateRenderer() fyne.WidgetRenderer {
	m.raster = canvas.NewRaster(m.generateGraphics)

	m.vonNeumannText = canvas.NewText("VON NEUMANN", color.RGBA{200, 100, 100, 255})
	m.vonNeumannText.TextSize = 11
	m.vonNeumannText.TextStyle = fyne.TextStyle{Bold: true}

	m.cimText = canvas.NewText("COMPUTE-IN-MEMORY", color.RGBA{100, 200, 150, 255})
	m.cimText.TextSize = 11
	m.cimText.TextStyle = fyne.TextStyle{Bold: true}

	m.cpuText = canvas.NewText("CPU", color.RGBA{255, 255, 255, 255})
	m.cpuText.TextSize = 10

	m.dramText = canvas.NewText("DRAM", color.RGBA{255, 255, 255, 255})
	m.dramText.TextSize = 10

	m.computeText1 = canvas.NewText("COMPUTE", color.RGBA{255, 255, 255, 255})
	m.computeText1.TextSize = 10

	m.computeText2 = canvas.NewText("HERE", color.RGBA{255, 255, 255, 255})
	m.computeText2.TextSize = 10

	m.vsText = canvas.NewText("VS", color.RGBA{0, 212, 255, 255})
	m.vsText.TextSize = 12
	m.vsText.TextStyle = fyne.TextStyle{Bold: true}

	m.counterText = canvas.NewText("Data Moves: 0", color.RGBA{255, 200, 100, 255})
	m.counterText.TextSize = 10

	m.wasteText = canvas.NewText("(ENERGY WASTE)", color.RGBA{255, 100, 100, 200})
	m.wasteText.TextSize = 9

	m.zeroMoveText = canvas.NewText("Zero Data Movement", color.RGBA{100, 255, 150, 255})
	m.zeroMoveText.TextSize = 10

	m.zeroWasteText = canvas.NewText("= Zero Waste", color.RGBA{100, 200, 150, 200})
	m.zeroWasteText.TextSize = 9

	return &memoryWallRenderer{widget: m}
}

type memoryWallRenderer struct {
	widget *MemoryWallAnimation
}

func (r *memoryWallRenderer) MinSize() fyne.Size {
	return r.widget.minSize
}

func (r *memoryWallRenderer) Layout(size fyne.Size) {
	r.widget.raster.Resize(size)
	midX := size.Width / 2

	r.widget.vonNeumannText.Move(fyne.NewPos(30, 5))
	r.widget.cimText.Move(fyne.NewPos(midX+30, 5))

	r.widget.cpuText.Move(fyne.NewPos(45, 45))
	r.widget.dramText.Move(fyne.NewPos(170, 45))

	r.widget.computeText1.Move(fyne.NewPos(midX+75, 40))
	r.widget.computeText2.Move(fyne.NewPos(midX+85, 55))

	r.widget.vsText.Move(fyne.NewPos(midX-15, size.Height/2-8))

	r.widget.counterText.Move(fyne.NewPos(30, size.Height-30))
	r.widget.wasteText.Move(fyne.NewPos(30, size.Height-15))

	r.widget.zeroMoveText.Move(fyne.NewPos(midX+50, size.Height-30))
	r.widget.zeroWasteText.Move(fyne.NewPos(midX+70, size.Height-15))
}

func (r *memoryWallRenderer) Refresh() {
	r.widget.mu.RLock()
	dataMovements := r.widget.dataMovements
	r.widget.mu.RUnlock()

	r.widget.counterText.Text = fmt.Sprintf("Data Moves: %d", dataMovements)
	r.widget.raster.Refresh()
	canvas.Refresh(r.widget.counterText)
}

func (r *memoryWallRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.widget.raster,
		r.widget.vonNeumannText,
		r.widget.cimText,
		r.widget.cpuText,
		r.widget.dramText,
		r.widget.computeText1,
		r.widget.computeText2,
		r.widget.vsText,
		r.widget.counterText,
		r.widget.wasteText,
		r.widget.zeroMoveText,
		r.widget.zeroWasteText,
	}
}

func (r *memoryWallRenderer) Destroy() {}

// generateGraphics creates the memory wall graphics (boxes and packets, no text).
func (m *MemoryWallAnimation) generateGraphics(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 200 || h < 80 {
		return img
	}

	m.mu.RLock()
	packets := make([]Packet, len(m.packets))
	copy(packets, m.packets)
	m.mu.RUnlock()

	midX := w / 2

	// CPU Box
	cpuX, cpuY := 30, 35
	cpuW, cpuH := 60, 35
	drawBoxFilledHero(img, cpuX, cpuY, cpuW, cpuH, color.RGBA{200, 100, 100, 255}, color.RGBA{100, 50, 50, 255})

	// DRAM Box
	memX, memY := 160, 35
	memW, memH := 60, 35
	drawBoxFilledHero(img, memX, memY, memW, memH, color.RGBA{100, 100, 200, 255}, color.RGBA{50, 50, 100, 255})

	// Data bus
	busY := cpuY + cpuH/2
	for x := cpuX + cpuW; x < memX; x++ {
		img.Set(x, busY, color.RGBA{100, 100, 100, 255})
		img.Set(x, busY+1, color.RGBA{100, 100, 100, 255})
	}

	// Packets
	for _, p := range packets {
		if !p.active {
			continue
		}
		px := int(p.x)
		py := int(p.y)
		packetColor := color.RGBA{255, 100, 100, 255}
		glowColor := color.RGBA{255, 50, 50, 100}
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				if px+dx > 0 && px+dx < midX-30 && py+dy > 0 && py+dy < h {
					img.Set(px+dx, py+dy, glowColor)
				}
			}
		}
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if px+dx > 0 && px+dx < midX-30 && py+dy > 0 && py+dy < h {
					img.Set(px+dx, py+dy, packetColor)
				}
			}
		}
	}

	// Divider
	divColor := color.RGBA{0, 100, 150, 255}
	for y := 10; y < h-10; y++ {
		img.Set(midX-10, y, divColor)
	}

	// CIM Box
	cimX, cimY := midX+50, 35
	cimW, cimH := 100, 45
	drawBoxFilledHero(img, cimX, cimY, cimW, cimH, color.RGBA{100, 200, 150, 255}, color.RGBA{50, 100, 75, 255})

	return img
}

func drawBoxFilledHero(img *image.RGBA, x, y, width, height int, borderColor, fillColor color.RGBA) {
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

// formatNumberHero formats numbers with commas.
func formatNumberHero(n float64) string {
	intVal := int(n)
	if intVal == 0 {
		return "0"
	}

	negative := intVal < 0
	if negative {
		intVal = -intVal
	}

	result := ""
	digits := 0
	for intVal > 0 {
		if digits > 0 && digits%3 == 0 {
			result = "," + result
		}
		result = string(rune('0'+intVal%10)) + result
		intVal /= 10
		digits++
	}

	if negative {
		result = "-" + result
	}

	return result
}

// SimplifiedEnergyRace is a simpler version using pure Fyne widgets
type SimplifiedEnergyRace struct {
	widget.BaseWidget
	container *fyne.Container
}

// NewSimplifiedEnergyRace creates energy comparison using Fyne widgets
func NewSimplifiedEnergyRace() *SimplifiedEnergyRace {
	s := &SimplifiedEnergyRace{}

	title := widget.NewLabelWithStyle("ENERGY PER MAC OPERATION", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create progress bars
	cpuBar := widget.NewProgressBar()
	cpuBar.Max = 1000
	cpuBar.SetValue(1000)

	gpuBar := widget.NewProgressBar()
	gpuBar.Max = 1000
	gpuBar.SetValue(100)

	fecimBar := widget.NewProgressBar()
	fecimBar.Max = 1000
	fecimBar.SetValue(10)

	cpuRow := container.NewBorder(nil, nil,
		widget.NewLabel("CPU+DRAM"),
		widget.NewLabel("1000 fJ"),
		cpuBar,
	)

	gpuRow := container.NewBorder(nil, nil,
		widget.NewLabel("GPU+HBM"),
		widget.NewLabel("100 fJ"),
		gpuBar,
	)

	fecimRow := container.NewBorder(nil, nil,
		widget.NewLabel("FeCIM"),
		widget.NewLabel("10 fJ*"),
		fecimBar,
	)

	headline := widget.NewLabelWithStyle("100x LESS ENERGY", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	source := widget.NewLabel("* FeCIM: Dr. Tour claims (TRL 4, not verified)")
	source.TextStyle = fyne.TextStyle{Italic: true}

	s.container = container.NewVBox(
		title,
		cpuRow,
		gpuRow,
		fecimRow,
		headline,
		source,
	)

	s.ExtendBaseWidget(s)
	return s
}

func (s *SimplifiedEnergyRace) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.container)
}

func (s *SimplifiedEnergyRace) MinSize() fyne.Size {
	return fyne.NewSize(400, 150)
}
