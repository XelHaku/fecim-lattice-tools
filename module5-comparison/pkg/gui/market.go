// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains market analysis visualizations.
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

// MarketSegment represents a market segment with growth data.
type MarketSegment struct {
	Name  string
	Y2025 float64 // Billion USD
	Y2030 float64 // Billion USD
	Color color.RGBA
}

// marketData holds the market opportunity data.
var marketData = []MarketSegment{
	{Name: "NAND Flash", Y2025: 78, Y2030: 98, Color: color.RGBA{200, 100, 100, 255}},
	{Name: "DRAM", Y2025: 143, Y2030: 220, Color: color.RGBA{100, 150, 200, 255}},
	{Name: "AI Semiconductor", Y2025: 163, Y2030: 403, Color: color.RGBA{100, 200, 150, 255}},
}

// MarketOpportunityChart shows the market opportunity visualization.
type MarketOpportunityChart struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	pulsePhase   float64
	minSize      fyne.Size

	raster       *canvas.Raster
	titleText    *canvas.Text
	totalText    *canvas.Text
	subtextLine1 *canvas.Text
	subtextLine2 *canvas.Text
	yearLabels   []*canvas.Text
	segLabels    []*canvas.Text
}

// NewMarketOpportunityChart creates a new market chart.
func NewMarketOpportunityChart() *MarketOpportunityChart {
	m := &MarketOpportunityChart{
		minSize:    fyne.NewSize(500, 180),
		yearLabels: make([]*canvas.Text, 2),
		segLabels:  make([]*canvas.Text, len(marketData)),
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateAnimation advances the animation.
func (m *MarketOpportunityChart) UpdateAnimation(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.animProgress < 1.0 {
		m.animProgress += dt * 0.5
		if m.animProgress > 1.0 {
			m.animProgress = 1.0
		}
	}

	m.pulsePhase += dt * 2.0
}

// Reset resets the animation.
func (m *MarketOpportunityChart) Reset() {
	m.mu.Lock()
	m.animProgress = 0
	m.pulsePhase = 0
	m.mu.Unlock()
	m.Refresh()
}

// MinSize returns minimum size.
func (m *MarketOpportunityChart) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *MarketOpportunityChart) CreateRenderer() fyne.WidgetRenderer {
	m.raster = canvas.NewRaster(m.generateBars)

	m.titleText = canvas.NewText("MARKET OPPORTUNITY ($B)", color.RGBA{0, 212, 255, 255})
	m.titleText.TextSize = 13
	m.titleText.TextStyle = fyne.TextStyle{Bold: true}

	m.totalText = canvas.NewText("$711B by 2030", color.RGBA{0, 212, 255, 255})
	m.totalText.TextSize = 14
	m.totalText.TextStyle = fyne.TextStyle{Bold: true}

	m.subtextLine1 = canvas.NewText("FeCIM can", color.RGBA{150, 150, 150, 255})
	m.subtextLine1.TextSize = 10

	m.subtextLine2 = canvas.NewText("address ALL", color.RGBA{100, 200, 150, 255})
	m.subtextLine2.TextSize = 10

	m.yearLabels[0] = canvas.NewText("2025", color.RGBA{150, 150, 150, 255})
	m.yearLabels[0].TextSize = 10
	m.yearLabels[1] = canvas.NewText("2030", color.RGBA{200, 200, 200, 255})
	m.yearLabels[1].TextSize = 10

	for i, seg := range marketData {
		m.segLabels[i] = canvas.NewText(seg.Name, color.RGBA{180, 180, 180, 255})
		m.segLabels[i].TextSize = 9
	}

	return &marketChartRenderer{widget: m}
}

type marketChartRenderer struct {
	widget *MarketOpportunityChart
}

func (r *marketChartRenderer) MinSize() fyne.Size {
	return r.widget.minSize
}

func (r *marketChartRenderer) Layout(size fyne.Size) {
	r.widget.raster.Resize(size)

	// Title at top center
	r.widget.titleText.Move(fyne.NewPos(size.Width/2-100, 5))

	// Total and subtext on left
	r.widget.totalText.Move(fyne.NewPos(10, size.Height/2-10))
	r.widget.subtextLine1.Move(fyne.NewPos(15, size.Height/2+15))
	r.widget.subtextLine2.Move(fyne.NewPos(15, size.Height/2+30))

	// Year labels at bottom
	chartStartX := float32(100)
	chartWidth := size.Width - 120
	r.widget.yearLabels[0].Move(fyne.NewPos(chartStartX+10, size.Height-20))
	r.widget.yearLabels[1].Move(fyne.NewPos(chartStartX+chartWidth-40, size.Height-20))

	// Segment labels
	barGroupWidth := chartWidth / float32(len(marketData))
	for i := range marketData {
		r.widget.segLabels[i].Move(fyne.NewPos(chartStartX+float32(i)*barGroupWidth+5, size.Height-35))
	}
}

func (r *marketChartRenderer) Refresh() {
	r.widget.mu.RLock()
	progress := r.widget.animProgress
	pulsePhase := r.widget.pulsePhase
	r.widget.mu.RUnlock()

	// Pulse total text
	if progress >= 1.0 {
		pulse := 0.7 + math.Sin(pulsePhase)*0.3
		r.widget.totalText.Color = color.RGBA{
			0,
			uint8(212 * pulse),
			uint8(255 * pulse),
			255,
		}
	} else {
		r.widget.totalText.Color = color.RGBA{0, 0, 0, 0} // Hidden until done
	}

	r.widget.raster.Refresh()
	canvas.Refresh(r.widget.totalText)
}

func (r *marketChartRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{
		r.widget.raster,
		r.widget.titleText,
		r.widget.totalText,
		r.widget.subtextLine1,
		r.widget.subtextLine2,
	}
	for _, lbl := range r.widget.yearLabels {
		objects = append(objects, lbl)
	}
	for _, lbl := range r.widget.segLabels {
		objects = append(objects, lbl)
	}
	return objects
}

func (r *marketChartRenderer) Destroy() {}

// generateBars creates just the bar graphics.
func (m *MarketOpportunityChart) generateBars(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 200 || h < 100 {
		return img
	}

	m.mu.RLock()
	progress := m.animProgress
	m.mu.RUnlock()

	labelWidth := 100
	chartWidth := w - labelWidth - 20
	barGroupWidth := chartWidth / len(marketData)
	maxVal := 450.0

	chartStartX := labelWidth
	chartStartY := 30
	chartHeight := h - chartStartY - 45

	for i, seg := range marketData {
		groupX := chartStartX + i*barGroupWidth

		bar2025Height := int(float64(chartHeight) * (seg.Y2025 / maxVal) * progress)
		bar2025X := groupX + 10
		bar2025Y := chartStartY + chartHeight - bar2025Height
		barWidth := (barGroupWidth - 30) / 2

		darkColor := color.RGBA{seg.Color.R / 2, seg.Color.G / 2, seg.Color.B / 2, 255}
		for dy := 0; dy < bar2025Height; dy++ {
			for dx := 0; dx < barWidth; dx++ {
				img.Set(bar2025X+dx, bar2025Y+dy, darkColor)
			}
		}

		bar2030Height := int(float64(chartHeight) * (seg.Y2030 / maxVal) * progress)
		bar2030X := groupX + 10 + barWidth + 5
		bar2030Y := chartStartY + chartHeight - bar2030Height

		for dy := 0; dy < bar2030Height; dy++ {
			for dx := 0; dx < barWidth; dx++ {
				img.Set(bar2030X+dx, bar2030Y+dy, seg.Color)
			}
		}

		// Growth arrow
		if bar2025Height > 0 && bar2030Height > 0 {
			arrowColor := color.RGBA{100, 255, 150, 200}
			for ay := bar2025Y; ay > bar2030Y; ay -= 3 {
				img.Set(bar2025X+barWidth/2, ay, arrowColor)
			}
			for ax := -3; ax <= 3; ax++ {
				img.Set(bar2030X+barWidth/2+ax, bar2030Y+5, arrowColor)
			}
		}
	}

	return img
}

// Competitor represents a competitor in the matrix.
type Competitor struct {
	Name      string
	Energy    string
	InMemory  int // 0=no, 1=partial, 2=yes
	CMOS      int
	Scalable  int
	Highlight bool
}

// competitors data for the competitive matrix.
var competitors = []Competitor{
	{"FeCIM", "1-10 fJ*", 2, 2, 2, true},
	{"Google TPU", "~100 fJ", 0, 2, 2, false},
	{"Intel Loihi 2", "~10 fJ", 2, 0, 0, false},
	{"Mythic AI", "~5 fJ", 2, 0, 1, false},
}

// CompetitiveMatrix shows competitive comparison.
type CompetitiveMatrix struct {
	widget.BaseWidget
}

// NewCompetitiveMatrix creates a new competitive matrix.
func NewCompetitiveMatrix() *CompetitiveMatrix {
	c := &CompetitiveMatrix{}
	c.ExtendBaseWidget(c)
	return c
}

// MinSize returns minimum size.
func (c *CompetitiveMatrix) MinSize() fyne.Size {
	return fyne.NewSize(400, 160)
}

// CreateRenderer implements fyne.Widget.
func (c *CompetitiveMatrix) CreateRenderer() fyne.WidgetRenderer {
	header := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Technology", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Energy/MAC", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("In-Memory", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("CMOS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Scalable", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	rows := container.NewVBox()
	for _, comp := range competitors {
		nameLabel := widget.NewLabel(comp.Name)
		if comp.Highlight {
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
		}

		energyLabel := widget.NewLabel(comp.Energy)

		row := container.NewGridWithColumns(5,
			nameLabel,
			energyLabel,
			createStatusLabel(comp.InMemory),
			createStatusLabel(comp.CMOS),
			createStatusLabel(comp.Scalable),
		)
		rows.Add(row)
	}

	disclaimer := widget.NewLabel("* TRL 4 - Lab validation only")
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		widget.NewLabelWithStyle("Competitive Comparison", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		header,
		widget.NewSeparator(),
		rows,
		widget.NewSeparator(),
		disclaimer,
	)

	return widget.NewSimpleRenderer(content)
}

// createStatusLabel creates a label showing status (checkmark, cross, or partial).
func createStatusLabel(status int) *widget.Label {
	var text string
	switch status {
	case 0:
		text = "X"
	case 1:
		text = "~"
	case 2:
		text = "Y"
	}
	label := widget.NewLabel(text)
	label.Alignment = fyne.TextAlignCenter
	return label
}

// formatNumberMarket formats numbers with commas.
func formatNumberMarket(n float64) string {
	return fmt.Sprintf("%.0f", n)
}
