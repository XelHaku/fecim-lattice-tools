// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains market analysis visualizations.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MarketSegment represents a market segment with growth data.
type MarketSegment struct {
	Name  string
	Y2024 float64 // Billion USD (historical baseline)
	Y2026 float64 // Billion USD (near-term forecast)
	Y2030 float64 // Billion USD (long-term forecast)
	Color color.RGBA
}

// marketData holds the market opportunity data (in billions USD).
// Sources: Gartner 2026 AI Semiconductor Forecast, WSTS Semiconductor Market Statistics 2024
var marketData = []MarketSegment{
	{Name: "NAND Flash", Y2024: 72, Y2026: 85, Y2030: 98, Color: color.RGBA{200, 100, 100, 255}},
	{Name: "DRAM", Y2024: 130, Y2026: 165, Y2030: 220, Color: color.RGBA{100, 150, 200, 255}},
	{Name: "AI Semiconductor", Y2024: 140, Y2026: 220, Y2030: 403, Color: color.RGBA{100, 200, 150, 255}},
}

// MarketOpportunityChart shows the market opportunity visualization using Fyne widgets.
type MarketOpportunityChart struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	pulsePhase   float64
	minSize      fyne.Size

	// Cached values to avoid redundant SetText calls (prevents resize loops)
	lastValues []string

	container *fyne.Container
	totalText *canvas.Text
	bars2024  []*canvas.Rectangle
	bars2026  []*canvas.Rectangle
	bars2030  []*canvas.Rectangle
	values    []*widget.Label
}

// NewMarketOpportunityChart creates a new market chart.
func NewMarketOpportunityChart() *MarketOpportunityChart {
	m := &MarketOpportunityChart{
		minSize:    fyne.NewSize(350, 120),
		bars2024:   make([]*canvas.Rectangle, len(marketData)),
		bars2026:   make([]*canvas.Rectangle, len(marketData)),
		bars2030:   make([]*canvas.Rectangle, len(marketData)),
		values:     make([]*widget.Label, len(marketData)),
		lastValues: make([]string, len(marketData)),
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
	fyne.Do(func() {
		m.Refresh()
	})
}

// MinSize returns minimum size.
func (m *MarketOpportunityChart) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *MarketOpportunityChart) CreateRenderer() fyne.WidgetRenderer {
	// Total market calculation: $98B + $220B + $403B = $721B
	m.totalText = canvas.NewText("$721B Market by 2030", color.RGBA{0, 212, 255, 255})
	m.totalText.TextSize = 24
	m.totalText.TextStyle = fyne.TextStyle{Bold: true}

	var segmentWidgets []fyne.CanvasObject
	maxVal := float32(450.0)
	barHeight := float32(50)

	// Create Y-axis with scale markers
	yAxisLabel := widget.NewLabel("USD\n(Billions)")
	yAxisLabel.Alignment = fyne.TextAlignCenter

	// Y-axis scale markers: 0, 200B, 400B
	scale400 := widget.NewLabel("400")
	scale400.Alignment = fyne.TextAlignTrailing
	scale200 := widget.NewLabel("200")
	scale200.Alignment = fyne.TextAlignTrailing
	scale0 := widget.NewLabel("0")
	scale0.Alignment = fyne.TextAlignTrailing

	yAxisScales := container.NewVBox(
		scale400,
		container.NewPadded(), // spacer
		scale200,
		container.NewPadded(), // spacer
		scale0,
	)

	yAxis := container.NewHBox(yAxisLabel, yAxisScales)

	for i, seg := range marketData {
		shortName := seg.Name
		if len(shortName) > 6 {
			shortName = shortName[:6]
		}
		segLabel := widget.NewLabel(shortName)

		darkestColor := color.RGBA{seg.Color.R / 3, seg.Color.G / 3, seg.Color.B / 3, 255}
		m.bars2024[i] = canvas.NewRectangle(darkestColor)
		m.bars2024[i].SetMinSize(fyne.NewSize(12, barHeight*float32(seg.Y2024)/maxVal))

		mediumColor := color.RGBA{seg.Color.R * 2 / 3, seg.Color.G * 2 / 3, seg.Color.B * 2 / 3, 255}
		m.bars2026[i] = canvas.NewRectangle(mediumColor)
		m.bars2026[i].SetMinSize(fyne.NewSize(12, barHeight*float32(seg.Y2026)/maxVal))

		m.bars2030[i] = canvas.NewRectangle(seg.Color)
		m.bars2030[i].SetMinSize(fyne.NewSize(12, barHeight*float32(seg.Y2030)/maxVal))

		m.values[i] = widget.NewLabel(fmt.Sprintf("$%.0fB", seg.Y2030))

		barGroup := container.NewHBox(m.bars2024[i], m.bars2026[i], m.bars2030[i])
		segCol := container.NewVBox(segLabel, barGroup, m.values[i])
		segmentWidgets = append(segmentWidgets, segCol)
	}

	barsRow := container.NewHBox(yAxis, container.NewHBox(segmentWidgets...))

	citation := widget.NewLabel("Source: Gartner 2026 AI Semiconductor Forecast")
	citation.TextStyle = fyne.TextStyle{Italic: true}
	citation.Alignment = fyne.TextAlignCenter

	m.container = container.NewVBox(container.NewCenter(m.totalText), barsRow, citation)
	return widget.NewSimpleRenderer(m.container)
}

// Refresh updates the widget display.
func (m *MarketOpportunityChart) Refresh() {
	m.mu.RLock()
	progress := m.animProgress
	pulsePhase := m.pulsePhase
	m.mu.RUnlock()

	if m.totalText == nil {
		return
	}

	// Pulse total text when done
	if progress >= 1.0 {
		pulse := 0.7 + math.Sin(pulsePhase)*0.3
		m.totalText.Color = color.RGBA{
			0,
			uint8(212 * pulse),
			uint8(255 * pulse),
			255,
		}
	} else {
		m.totalText.Color = color.RGBA{0, 150, 200, 255}
	}

	// Update bar heights based on progress
	maxVal := float32(450.0)
	barHeight := float32(80)

	for i, seg := range marketData {
		bar2024Height := barHeight * float32(seg.Y2024) / maxVal * float32(progress)
		m.bars2024[i].SetMinSize(fyne.NewSize(20, max(2, bar2024Height)))

		bar2026Height := barHeight * float32(seg.Y2026) / maxVal * float32(progress)
		m.bars2026[i].SetMinSize(fyne.NewSize(20, max(2, bar2026Height)))

		bar2030Height := barHeight * float32(seg.Y2030) / maxVal * float32(progress)
		m.bars2030[i].SetMinSize(fyne.NewSize(20, max(2, bar2030Height)))

		// Use caching to avoid redundant SetText calls
		newText := fmt.Sprintf("$%.0fB", seg.Y2030*progress)
		if newText != m.lastValues[i] {
			m.values[i].SetText(newText)
			m.lastValues[i] = newText
		}

		canvas.Refresh(m.bars2024[i])
		canvas.Refresh(m.bars2026[i])
		canvas.Refresh(m.bars2030[i])
	}

	canvas.Refresh(m.totalText)
}

// Competitor represents a competitor in the matrix.
type Competitor struct {
	Name        string
	Energy      string
	InMemory    int // 0=no, 1=partial, 2=yes
	CMOS        int
	Scalable    int
	Highlight   bool
	IsEstimated bool // True if energy value is estimated
}

// competitors data for the competitive matrix.
// Energy values sourced from published specifications where available.
var competitors = []Competitor{
	{"FeCIM", "~1 pJ*", 2, 2, 2, true, true},            // Dr. Tour claim (unverified)
	{"Google TPU v5", "~100 pJ", 0, 2, 2, false, false}, // Google published specs
	{"Intel Loihi 2", "~10 pJ", 2, 0, 0, false, false},  // Intel published specs (non-CMOS fab)
	{"IBM Analog AI", "~10 pJ", 2, 1, 1, false, true},   // Research prototype estimate
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
	return fyne.NewSize(350, 130)
}

// CreateRenderer implements fyne.Widget.
func (c *CompetitiveMatrix) CreateRenderer() fyne.WidgetRenderer {
	// Amber color for estimated values
	estimatedAmber := color.RGBA{255, 191, 0, 255}

	header := container.NewGridWithColumns(5,
		widget.NewLabel("Tech"),
		widget.NewLabel("Energy"),
		widget.NewLabel("Memory Type"),
		widget.NewLabel("CMOS Compatible"),
		widget.NewLabel("Scalable"),
	)

	rows := container.NewVBox()
	for _, comp := range competitors {
		nameLabel := widget.NewLabel(comp.Name)
		if comp.Highlight {
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
		}

		// Render energy value in amber if estimated
		var energyWidget fyne.CanvasObject
		if comp.IsEstimated {
			energyText := canvas.NewText(comp.Energy, estimatedAmber)
			energyText.TextSize = 14
			energyWidget = energyText
		} else {
			energyWidget = widget.NewLabel(comp.Energy)
		}

		row := container.NewGridWithColumns(5,
			nameLabel,
			energyWidget,
			createStatusLabel(comp.InMemory),
			createStatusLabel(comp.CMOS),
			createStatusLabel(comp.Scalable),
		)
		rows.Add(row)
	}

	// Add legend below the table
	legend := widget.NewLabel("✓ = Yes | ⚠ = Partial | ✗ = No")
	legend.TextStyle = fyne.TextStyle{Italic: true}
	legend.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(header, rows, legend)
	return widget.NewSimpleRenderer(content)
}

// createStatusLabel creates an icon showing status.
func createStatusLabel(status int) fyne.CanvasObject {
	var icon fyne.Resource
	switch status {
	case 0:
		icon = theme.CancelIcon()
	case 1:
		icon = theme.WarningIcon()
	case 2:
		icon = theme.ConfirmIcon()
	}
	return widget.NewIcon(icon)
}

// formatNumberMarket formats numbers with commas.
func formatNumberMarket(n float64) string {
	return fmt.Sprintf("%.0f", n)
}
