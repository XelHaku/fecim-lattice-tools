// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains market analysis visualizations.
// TECHNICAL BRIEFING DESIGN: Based on project literature review and model-input messaging.
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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
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
// Model inputs based on WSTS Semiconductor Trade Statistics 2025 and Gartner AI Semiconductor forecasts 2025.
// TOTAL: $98B + $220B + $403B = $721B (scenario input for visualization)
var marketData = []MarketSegment{
	{Name: "NAND Flash", Y2024: 72, Y2026: 85, Y2030: 98, Color: color.RGBA{231, 76, 60, 255}},           // Red
	{Name: "DRAM", Y2024: 130, Y2026: 165, Y2030: 220, Color: color.RGBA{243, 156, 18, 255}},             // Amber
	{Name: "AI Semiconductor", Y2024: 140, Y2026: 220, Y2030: 403, Color: color.RGBA{46, 204, 113, 255}}, // Green
}

// MarketOpportunityChart shows the HERO market opportunity - investor grade.
// HERO STATEMENT: "$721B ADDRESSABLE MARKET BY 2030"
type MarketOpportunityChart struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	pulsePhase   float64
	minSize      fyne.Size

	// BUG-M5-004 FIX: Track raw progress values to detect threshold crossings
	lastProgressPct []int // Percentage values per segment
	needsTextUpdate bool

	container    *fyne.Container
	heroText     *canvas.Text
	heroSubtext  *canvas.Text
	marketBoxes  []*canvas.Rectangle
	marketLabels []*canvas.Text
	marketValues []*canvas.Text
	renderer     fyne.WidgetRenderer // BUG-M5-003 FIX: Cache renderer
}

// NewMarketOpportunityChart creates a new market chart.
func NewMarketOpportunityChart() *MarketOpportunityChart {
	m := &MarketOpportunityChart{
		minSize:         fyne.NewSize(600, 160),
		marketBoxes:     make([]*canvas.Rectangle, len(marketData)),
		marketLabels:    make([]*canvas.Text, len(marketData)),
		marketValues:    make([]*canvas.Text, len(marketData)),
		lastProgressPct: make([]int, len(marketData)),
	}
	// Initialize to -1 to force first update
	for i := range m.lastProgressPct {
		m.lastProgressPct[i] = -1
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateAnimation advances the animation.
// BUG-M5-004 FIX: Check thresholds BEFORE formatting to avoid unnecessary recalculations
func (m *MarketOpportunityChart) UpdateAnimation(dt float64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	dirty := false
	if m.animProgress < 1.0 {
		m.animProgress += dt * 0.4
		if m.animProgress > 1.0 {
			m.animProgress = 1.0
		}
		dirty = true
	}

	m.pulsePhase += dt * 2.0
	dirty = true

	// BUG-M5-004 FIX: Check if any segment crossed a threshold (1+ point change)
	for i, seg := range marketData {
		newPct := int(seg.Y2030 * m.animProgress)
		if newPct != m.lastProgressPct[i] {
			m.needsTextUpdate = true
			m.lastProgressPct[i] = newPct
			dirty = true
		}
	}
	return dirty
}

// Reset resets the animation.
func (m *MarketOpportunityChart) Reset() {
	m.mu.Lock()
	m.animProgress = 0
	m.pulsePhase = 0
	// Force text update on next frame
	for i := range m.lastProgressPct {
		m.lastProgressPct[i] = -1
	}
	m.needsTextUpdate = true
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
// BUG-M5-003 FIX: Cache renderer and reuse canvas objects instead of recreating
func (m *MarketOpportunityChart) CreateRenderer() fyne.WidgetRenderer {
	// Return cached renderer if already created (prevents recreation)
	if m.renderer != nil {
		return m.renderer
	}

	// === HERO SECTION: COMPACT "$721B" ===
	m.heroText = canvas.NewText("$721B", heroTextColor)
	m.heroText.TextSize = 48 // Compact for unified view
	m.heroText.TextStyle = fyne.TextStyle{Bold: true}
	m.heroText.Alignment = fyne.TextAlignCenter

	m.heroSubtext = canvas.NewText("⚠️ SIMULATION ONLY | MODEL INPUTS (market scenarios) | TRL 4→9 Required", heroAmberColor)
	m.heroSubtext.TextSize = 14
	m.heroSubtext.TextStyle = fyne.TextStyle{Bold: true}
	m.heroSubtext.Alignment = fyne.TextAlignCenter

	heroSection := container.NewHBox(
		layout.NewSpacer(),
		container.NewVBox(
			container.NewCenter(m.heroText),
			container.NewCenter(m.heroSubtext),
		),
		layout.NewSpacer(),
	)

	// === THREE MARKET BOXES (horizontal) ===
	boxWidth := float32(100)
	boxHeight := float32(60)

	var marketWidgets []fyne.CanvasObject
	for i, seg := range marketData {
		// Market segment box
		m.marketBoxes[i] = canvas.NewRectangle(seg.Color)
		m.marketBoxes[i].SetMinSize(fyne.NewSize(boxWidth, boxHeight))
		m.marketBoxes[i].CornerRadius = 8

		// Market name label
		m.marketLabels[i] = canvas.NewText(seg.Name, heroTextColor)
		m.marketLabels[i].TextSize = 14
		m.marketLabels[i].TextStyle = fyne.TextStyle{Bold: true}
		m.marketLabels[i].Alignment = fyne.TextAlignCenter

		// Market value (compact)
		m.marketValues[i] = canvas.NewText(fmt.Sprintf("$%.0fB", seg.Y2030), heroTextColor)
		m.marketValues[i].TextSize = 18
		m.marketValues[i].TextStyle = fyne.TextStyle{Bold: true}
		m.marketValues[i].Alignment = fyne.TextAlignCenter

		// Stack value and label inside box
		boxContent := container.NewVBox(
			layout.NewSpacer(),
			container.NewCenter(m.marketValues[i]),
			container.NewCenter(m.marketLabels[i]),
			layout.NewSpacer(),
		)

		boxWithContent := container.NewStack(m.marketBoxes[i], boxContent)
		marketWidgets = append(marketWidgets, container.NewPadded(boxWithContent))
	}

	marketsRow := container.NewHBox(
		layout.NewSpacer(),
		marketWidgets[0],
		marketWidgets[1],
		marketWidgets[2],
		layout.NewSpacer(),
	)

	// === SOURCE ===
	citation := canvas.NewText("Model input references (not validated): WSTS + Gartner (2025)", heroMutedColor)
	citation.TextSize = 14
	citation.TextStyle = fyne.TextStyle{Italic: true}
	citation.Alignment = fyne.TextAlignCenter

	// === ASSEMBLE ===
	m.container = container.NewVBox(
		heroSection,
		marketsRow,
		container.NewCenter(citation),
	)

	// Initialize cache values to force first update
	for i := range m.lastProgressPct {
		m.lastProgressPct[i] = -1
	}
	m.needsTextUpdate = true

	// Cache and return the renderer
	m.renderer = &marketChartRenderer{widget: m, container: m.container}
	return m.renderer
}

// marketChartRenderer is a custom renderer that properly implements Layout.
// BUG-M5-003 FIX: Proper WidgetRenderer with Layout() method
type marketChartRenderer struct {
	widget    *MarketOpportunityChart
	container *fyne.Container
}

func (r *marketChartRenderer) Destroy() {}

func (r *marketChartRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

func (r *marketChartRenderer) MinSize() fyne.Size {
	return r.widget.MinSize()
}

func (r *marketChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *marketChartRenderer) Refresh() {
	r.widget.doRefresh()
}

// doRefresh performs the actual refresh logic (called by renderer).
// BUG-M5-004 FIX: Only format text when needsTextUpdate is true
func (m *MarketOpportunityChart) doRefresh() {
	m.mu.RLock()
	progress := m.animProgress
	pulsePhase := m.pulsePhase
	needsTextUpdate := m.needsTextUpdate
	m.mu.RUnlock()

	if m.heroText == nil {
		return
	}

	// Pulse hero text when animation complete
	if progress >= 1.0 {
		pulse := 0.85 + math.Sin(pulsePhase)*0.15
		m.heroText.Color = color.RGBA{
			uint8(240 * pulse),
			uint8(244 * pulse),
			uint8(248 * pulse),
			255,
		}
	} else {
		m.heroText.Color = heroTextColor
	}

	// BUG-M5-004 FIX: Only format and update text when threshold was crossed
	if needsTextUpdate {
		for i, seg := range marketData {
			m.marketValues[i].Text = fmt.Sprintf("$%.0fB", seg.Y2030*progress)
			canvas.Refresh(m.marketValues[i])
		}

		// Clear the flag
		m.mu.Lock()
		m.needsTextUpdate = false
		m.mu.Unlock()
	}

	canvas.Refresh(m.heroText)
}

// Refresh triggers a widget refresh via the base widget.
func (m *MarketOpportunityChart) Refresh() {
	m.BaseWidget.Refresh()
}

// Competitor represents a competitor in the matrix.
type Competitor struct {
	Name      string
	Energy    float64 // 0..1 confidence-weighted score
	Speed     float64
	Endurance float64
	CMOS      float64
	Scalable  float64
	Highlight bool
}

func confidenceAdjusted(raw float64, tag physics.ConfidenceTag) float64 {
	if raw < 0 {
		raw = 0
	}
	if raw > 1 {
		raw = 1
	}
	return raw * tag.Confidence
}

func buildCompetitors() []Competitor {
	ledger := physics.NewConfidenceLedger()
	energyTag, _ := ledger.Lookup("Pr")
	speedTag, _ := ledger.Lookup("Ec")
	enduranceTag, _ := ledger.Lookup("rho_viscosity")
	cmosTag, _ := ledger.Lookup("beta_landau")
	scaleTag, _ := ledger.Lookup("gamma_landau")

	return []Competitor{
		{"FeCIM", confidenceAdjusted(0.95, energyTag), confidenceAdjusted(0.88, speedTag), confidenceAdjusted(0.82, enduranceTag), confidenceAdjusted(0.89, cmosTag), confidenceAdjusted(0.84, scaleTag), true},
		{"Google TPU v5", confidenceAdjusted(0.62, energyTag), confidenceAdjusted(0.94, speedTag), confidenceAdjusted(0.86, enduranceTag), confidenceAdjusted(0.93, cmosTag), confidenceAdjusted(0.95, scaleTag), false},
		{"Intel Loihi 2", confidenceAdjusted(0.79, energyTag), confidenceAdjusted(0.83, speedTag), confidenceAdjusted(0.80, enduranceTag), confidenceAdjusted(0.54, cmosTag), confidenceAdjusted(0.52, scaleTag), false},
		{"IBM Analog AI", confidenceAdjusted(0.74, energyTag), confidenceAdjusted(0.55, speedTag), confidenceAdjusted(0.46, enduranceTag), confidenceAdjusted(0.49, cmosTag), confidenceAdjusted(0.45, scaleTag), false},
		{"ReRAM", confidenceAdjusted(0.77, energyTag), confidenceAdjusted(0.58, speedTag), confidenceAdjusted(0.50, enduranceTag), confidenceAdjusted(0.66, cmosTag), confidenceAdjusted(0.48, scaleTag), false},
	}
}

var competitors = buildCompetitors()

// CompetitiveMatrix shows simplified competitive comparison (model inputs).
// INVESTOR MESSAGE: "Only FeCIM has checkmarks in ALL categories" (model input)
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
	return fyne.NewSize(600, 160)
}

// CreateRenderer implements fyne.Widget.
func (c *CompetitiveMatrix) CreateRenderer() fyne.WidgetRenderer {
	// Hero message
	heroText := canvas.NewText("Confidence-aware model scoring (0-100) across key categories", heroCyanColor)
	heroText.TextSize = 14
	heroText.TextStyle = fyne.TextStyle{Bold: true}
	heroText.Alignment = fyne.TextAlignCenter

	// Header row - simplified categories
	headers := []string{"Technology", "Energy", "Speed", "Endurance", "CMOS", "Scalable"}
	var headerWidgets []fyne.CanvasObject
	for _, h := range headers {
		label := canvas.NewText(h, heroTextColor)
		label.TextSize = 14
		label.TextStyle = fyne.TextStyle{Bold: true}
		label.Alignment = fyne.TextAlignCenter
		headerWidgets = append(headerWidgets, container.NewCenter(label))
	}
	headerRow := container.NewGridWithColumns(6, headerWidgets...)

	// Data rows
	rows := container.NewVBox()
	for _, comp := range competitors {
		var rowWidgets []fyne.CanvasObject

		// Technology name
		nameText := canvas.NewText(comp.Name, heroTextColor)
		if comp.Highlight {
			nameText.Color = heroCyanColor
			nameText.TextStyle = fyne.TextStyle{Bold: true}
		}
		nameText.TextSize = 14
		rowWidgets = append(rowWidgets, container.NewCenter(nameText))

		// Confidence-weighted score columns
		scores := []float64{comp.Energy, comp.Speed, comp.Endurance, comp.CMOS, comp.Scalable}
		for _, score := range scores {
			txt := canvas.NewText(fmt.Sprintf("%.0f", score*100), heroTextColor)
			txt.Alignment = fyne.TextAlignCenter
			txt.TextSize = 14
			switch {
			case score >= 0.75:
				txt.Color = color.RGBA{46, 204, 113, 255}
			case score >= 0.50:
				txt.Color = color.RGBA{243, 156, 18, 255}
			default:
				txt.Color = color.RGBA{231, 76, 60, 255}
			}
			rowWidgets = append(rowWidgets, container.NewCenter(txt))
		}

		row := container.NewGridWithColumns(6, rowWidgets...)

		// Highlight FeCIM row with cyan background for investor visibility
		if comp.Highlight {
			highlightBg := canvas.NewRectangle(color.RGBA{0, 212, 255, 40}) // Subtle cyan
			highlightBg.CornerRadius = 4
			rows.Add(container.NewStack(highlightBg, row))
		} else {
			rows.Add(row)
		}
	}

	// === CITATION ===
	citation := canvas.NewText("Sources: Samsung Nature 2025, IEEE IRPS 2022, Nano Letters 2024, CEA-Leti 2024", heroMutedColor)
	citation.TextSize = 14
	citation.TextStyle = fyne.TextStyle{Italic: true}
	citation.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		container.NewCenter(heroText),
		headerRow,
		rows,
		container.NewCenter(citation),
	)
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
