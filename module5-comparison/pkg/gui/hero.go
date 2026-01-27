// Package gui provides Fyne-based GUI components for architecture comparison.
// This file contains hero visualizations for the comparison demo.
// TECHNICAL BRIEFING DESIGN: Based on Dr. Tour's COSM 2025 presentation messaging.
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

// estimatedColor is used for unverified/estimated values (amber)
var estimatedColor = color.RGBA{255, 191, 0, 255}

// Technical briefing colors
var (
	heroTextColor   = color.RGBA{240, 244, 248, 255} // Off-white for maximum contrast
	heroCyanColor   = color.RGBA{0, 212, 255, 255}   // FeCIM cyan accent
	heroGreenColor  = color.RGBA{46, 204, 113, 255}  // Success green
	heroRedColor    = color.RGBA{231, 76, 60, 255}   // GPU baseline red
	heroAmberColor  = color.RGBA{243, 156, 18, 255}  // Warning/caution amber
	heroMutedColor  = color.RGBA{160, 180, 200, 255} // Secondary text
)

// Energy values in picojoules per MAC (from docs/videos/ironlattice-youtube-script.md)
// "CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. FeCIM: under 1 picojoule."
const (
	cpuEnergyPJ   = 1000.0 // 1000 pJ/MAC
	gpuEnergyPJ   = 100.0  // 100 pJ/MAC
	fecimEnergyPJ = 1.0    // ~1 pJ/MAC (conservative estimate for claimed "<1 pJ")
)

// AnimatedEnergyRace shows the HERO energy comparison - investor grade.
// HERO STATEMENT: "80-90% data center energy reduction"
type AnimatedEnergyRace struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64 // 0-1 for bar growth
	showWinner   bool
	pulsePhase   float64

	// Cached values to avoid redundant text formatting (prevents resize loops)
	// BUG-M5-004 FIX: Track raw progress values, not formatted strings
	lastGpuProgress   int // Percentage as int (0-100) to detect 1+ point changes
	lastFecimProgress int // Percentage as int (0-100)
	needsTextUpdate   bool

	// UI elements (cached for renderer reuse)
	container     *fyne.Container
	heroText      *canvas.Text
	heroSubtext   *canvas.Text
	gpuBar        *canvas.Rectangle
	fecimBar      *canvas.Rectangle
	gpuLabel      *canvas.Text
	fecimLabel    *canvas.Text
	statStrip     *canvas.Text
	renderer      fyne.WidgetRenderer
}

// NewAnimatedEnergyRace creates a new energy race visualization.
func NewAnimatedEnergyRace() *AnimatedEnergyRace {
	e := &AnimatedEnergyRace{}
	e.ExtendBaseWidget(e)
	return e
}

// SetLogScale enables/disables logarithmic scale (placeholder).
func (e *AnimatedEnergyRace) SetLogScale(log bool) {}

// UpdateAnimation advances the animation by dt seconds.
// BUG-M5-004 FIX: Check thresholds BEFORE formatting to avoid unnecessary recalculations
func (e *AnimatedEnergyRace) UpdateAnimation(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.animProgress < 1.0 {
		e.animProgress += dt * 0.4 // Slightly slower for dramatic effect
		if e.animProgress > 1.0 {
			e.animProgress = 1.0
			e.showWinner = true
		}
	}

	if e.showWinner {
		e.pulsePhase += dt * 2.0
	}

	// BUG-M5-004 FIX: Check if progress crossed a 1-point threshold before marking for update
	// Only mark needsTextUpdate when the integer percentage changes
	newGpuProgress := int(100 * e.animProgress)
	newFecimProgress := int(10 * e.animProgress)

	if newGpuProgress != e.lastGpuProgress || newFecimProgress != e.lastFecimProgress {
		e.needsTextUpdate = true
		e.lastGpuProgress = newGpuProgress
		e.lastFecimProgress = newFecimProgress
	}
}

// Reset resets the animation.
func (e *AnimatedEnergyRace) Reset() {
	e.mu.Lock()
	e.animProgress = 0
	e.showWinner = false
	e.pulsePhase = 0
	e.lastGpuProgress = -1   // Force text update on first frame
	e.lastFecimProgress = -1 // Force text update on first frame
	e.needsTextUpdate = true
	e.mu.Unlock()
	fyne.Do(func() {
		e.Refresh()
	})
}

// MinSize returns minimum size.
func (e *AnimatedEnergyRace) MinSize() fyne.Size {
	return fyne.NewSize(800, 400)
}

// CreateRenderer implements fyne.Widget.
// BUG-M5-003 FIX: Cache renderer and reuse canvas objects instead of recreating
func (e *AnimatedEnergyRace) CreateRenderer() fyne.WidgetRenderer {
	// Return cached renderer if already created (prevents recreation)
	if e.renderer != nil {
		return e.renderer
	}

	// === HERO SECTION: MASSIVE "80-90%" ===
	e.heroText = canvas.NewText("80-90%", heroTextColor)
	e.heroText.TextSize = 96 // MASSIVE for investor impact
	e.heroText.TextStyle = fyne.TextStyle{Bold: true}
	e.heroText.Alignment = fyne.TextAlignCenter

	e.heroSubtext = canvas.NewText("DATA CENTER ENERGY REDUCTION (PROJECTED)", heroCyanColor)
	e.heroSubtext.TextSize = 28
	e.heroSubtext.TextStyle = fyne.TextStyle{Bold: true}
	e.heroSubtext.Alignment = fyne.TextAlignCenter

	// Prominent TRL warning - CRIT-001 fix
	trlWarning := canvas.NewText("Laboratory estimates only - not independently verified", heroAmberColor)
	trlWarning.TextSize = 16
	trlWarning.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
	trlWarning.Alignment = fyne.TextAlignCenter

	heroSection := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(e.heroText),
		container.NewCenter(e.heroSubtext),
		container.NewCenter(trlWarning),
		layout.NewSpacer(),
	)

	// === BEFORE/AFTER COMPARISON: Simple, clean ===
	barWidth := float32(500)
	barHeight := float32(40)

	// "Today: GPU-based AI" label
	todayLabel := canvas.NewText("Today: GPU-based AI", heroMutedColor)
	todayLabel.TextSize = 14
	todayLabel.Alignment = fyne.TextAlignLeading

	// GPU bar - full width (100 units = 100%)
	e.gpuBar = canvas.NewRectangle(heroRedColor)
	e.gpuBar.SetMinSize(fyne.NewSize(barWidth, barHeight))
	e.gpuLabel = canvas.NewText("100 units power", heroTextColor)
	e.gpuLabel.TextSize = 14
	e.gpuLabel.Alignment = fyne.TextAlignTrailing

	gpuRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(180, barHeight), container.NewCenter(todayLabel)),
		e.gpuBar,
		container.NewGridWrap(fyne.NewSize(120, barHeight), container.NewCenter(e.gpuLabel)),
	)

	// Arrow
	arrowText := canvas.NewText("vs", heroMutedColor)
	arrowText.TextSize = 16
	arrowText.Alignment = fyne.TextAlignCenter
	arrowRow := container.NewCenter(arrowText)

	// "FeCIM: Compute-in-Memory" label
	fecimLabelText := canvas.NewText("FeCIM: Compute-in-Memory", heroMutedColor)
	fecimLabelText.TextSize = 14
	fecimLabelText.Alignment = fyne.TextAlignLeading

	// FeCIM bar - ~10% width (10 units = 90% reduction)
	e.fecimBar = canvas.NewRectangle(heroGreenColor)
	e.fecimBar.SetMinSize(fyne.NewSize(barWidth*0.1, barHeight))
	e.fecimLabel = canvas.NewText("~10 units power", heroTextColor)
	e.fecimLabel.TextSize = 14
	e.fecimLabel.Alignment = fyne.TextAlignTrailing

	fecimRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(180, barHeight), container.NewCenter(fecimLabelText)),
		e.fecimBar,
		container.NewGridWrap(fyne.NewSize(120, barHeight), container.NewCenter(e.fecimLabel)),
	)

	comparisonSection := container.NewVBox(
		gpuRow,
		arrowRow,
		fecimRow,
	)

	// === KEY STAT STRIP ===
	e.statStrip = canvas.NewText("1000x less than CPU  |  100x less than GPU  |  ~1 pJ per MAC", heroCyanColor)
	e.statStrip.TextSize = 16
	e.statStrip.TextStyle = fyne.TextStyle{Bold: true}
	e.statStrip.Alignment = fyne.TextAlignCenter

	// === DISCLAIMER ===
	disclaimer := canvas.NewText("* TRL 4 (Laboratory Validation) - Energy claims pending independent verification", estimatedColor)
	disclaimer.TextSize = 11
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}
	disclaimer.Alignment = fyne.TextAlignCenter

	// === ASSEMBLE ===
	e.container = container.NewVBox(
		heroSection,
		widget.NewSeparator(),
		container.NewPadded(comparisonSection),
		widget.NewSeparator(),
		container.NewCenter(e.statStrip),
		layout.NewSpacer(),
		container.NewCenter(disclaimer),
	)

	// Initialize cache values to force first update
	e.lastGpuProgress = -1
	e.lastFecimProgress = -1
	e.needsTextUpdate = true

	// Cache and return the renderer
	e.renderer = &energyRaceRenderer{widget: e, container: e.container}
	return e.renderer
}

// energyRaceRenderer is a custom renderer that properly implements Layout.
// BUG-M5-003 FIX: Proper WidgetRenderer with Layout() method
type energyRaceRenderer struct {
	widget    *AnimatedEnergyRace
	container *fyne.Container
}

func (r *energyRaceRenderer) Destroy() {}

func (r *energyRaceRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

func (r *energyRaceRenderer) MinSize() fyne.Size {
	return r.widget.MinSize()
}

func (r *energyRaceRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *energyRaceRenderer) Refresh() {
	r.widget.doRefresh()
}

// doRefresh performs the actual refresh logic (called by renderer).
// BUG-M5-004 FIX: Only format text when needsTextUpdate is true
func (e *AnimatedEnergyRace) doRefresh() {
	e.mu.RLock()
	progress := e.animProgress
	showWinner := e.showWinner
	pulsePhase := e.pulsePhase
	needsTextUpdate := e.needsTextUpdate
	e.mu.RUnlock()

	if e.gpuBar == nil {
		return
	}

	// Animate bar widths
	barWidth := float32(500)
	barHeight := float32(40)
	e.gpuBar.SetMinSize(fyne.NewSize(barWidth*float32(progress), barHeight))
	e.fecimBar.SetMinSize(fyne.NewSize(max(10, barWidth*0.1*float32(progress)), barHeight))

	// BUG-M5-004 FIX: Only format and update text when threshold was crossed
	if needsTextUpdate {
		var gpuText, fecimText string
		if progress > 0.9 {
			gpuText = "100 units power"
			fecimText = "~10 units power"
		} else {
			gpuText = fmt.Sprintf("%.0f units", 100*progress)
			fecimText = fmt.Sprintf("~%.0f units", 10*progress)
		}
		e.gpuLabel.Text = gpuText
		e.fecimLabel.Text = fecimText
		canvas.Refresh(e.gpuLabel)
		canvas.Refresh(e.fecimLabel)

		// Clear the flag
		e.mu.Lock()
		e.needsTextUpdate = false
		e.mu.Unlock()
	}

	// Pulse the hero text when animation complete
	if showWinner {
		pulse := 0.85 + math.Sin(pulsePhase)*0.15
		e.heroText.Color = color.RGBA{
			uint8(240 * pulse),
			uint8(244 * pulse),
			uint8(248 * pulse),
			255,
		}
		canvas.Refresh(e.heroText)
	}

	canvas.Refresh(e.gpuBar)
	canvas.Refresh(e.fecimBar)
	e.container.Refresh()
}

// Refresh triggers a widget refresh via the base widget.
func (e *AnimatedEnergyRace) Refresh() {
	e.BaseWidget.Refresh()
}

// PhasedStrategyDiagram shows the de-risking phased market entry strategy.
// CRITICAL FOR INVESTORS: Shows NAND first, then DRAM, then full CIM
type PhasedStrategyDiagram struct {
	widget.BaseWidget

	mu           sync.RWMutex
	animProgress float64
	pulsePhase   float64

	container *fyne.Container
}

// NewPhasedStrategyDiagram creates a new phased strategy visualization.
func NewPhasedStrategyDiagram() *PhasedStrategyDiagram {
	p := &PhasedStrategyDiagram{}
	p.ExtendBaseWidget(p)
	return p
}

// UpdateAnimation advances the animation.
func (p *PhasedStrategyDiagram) UpdateAnimation(dt float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.animProgress < 1.0 {
		p.animProgress += dt * 0.3
		if p.animProgress > 1.0 {
			p.animProgress = 1.0
		}
	}
	p.pulsePhase += dt * 2.0
}

// Reset resets the animation.
func (p *PhasedStrategyDiagram) Reset() {
	p.mu.Lock()
	p.animProgress = 0
	p.pulsePhase = 0
	p.mu.Unlock()
	fyne.Do(func() {
		p.Refresh()
	})
}

// MinSize returns minimum size.
func (p *PhasedStrategyDiagram) MinSize() fyne.Size {
	return fyne.NewSize(700, 200)
}

// CreateRenderer implements fyne.Widget.
func (p *PhasedStrategyDiagram) CreateRenderer() fyne.WidgetRenderer {
	// Phase 1: NAND Replacement
	phase1Title := canvas.NewText("PHASE 1", heroCyanColor)
	phase1Title.TextSize = 14
	phase1Title.TextStyle = fyne.TextStyle{Bold: true}
	phase1Name := canvas.NewText("NAND Replacement", heroTextColor)
	phase1Name.TextSize = 18
	phase1Name.TextStyle = fyne.TextStyle{Bold: true}
	phase1Desc := widget.NewLabel("Immediate revenue\nNo software changes\nDrop-in compatible")
	phase1Desc.Alignment = fyne.TextAlignCenter
	phase1Box := container.NewVBox(
		container.NewCenter(phase1Title),
		container.NewCenter(phase1Name),
		phase1Desc,
	)

	// Arrow 1
	arrow1 := canvas.NewText("->", heroMutedColor)
	arrow1.TextSize = 24

	// Phase 2: DRAM Replacement
	phase2Title := canvas.NewText("PHASE 2", heroCyanColor)
	phase2Title.TextSize = 14
	phase2Title.TextStyle = fyne.TextStyle{Bold: true}
	phase2Name := canvas.NewText("DRAM Replacement", heroTextColor)
	phase2Name.TextSize = 18
	phase2Name.TextStyle = fyne.TextStyle{Bold: true}
	phase2Desc := widget.NewLabel("Eliminate refresh power\nHigher density\nNon-volatile")
	phase2Desc.Alignment = fyne.TextAlignCenter
	phase2Box := container.NewVBox(
		container.NewCenter(phase2Title),
		container.NewCenter(phase2Name),
		phase2Desc,
	)

	// Arrow 2
	arrow2 := canvas.NewText("->", heroMutedColor)
	arrow2.TextSize = 24

	// Phase 3: Full CIM
	phase3Title := canvas.NewText("PHASE 3", heroGreenColor)
	phase3Title.TextSize = 14
	phase3Title.TextStyle = fyne.TextStyle{Bold: true}
	phase3Name := canvas.NewText("Full CIM", heroTextColor)
	phase3Name.TextSize = 18
	phase3Name.TextStyle = fyne.TextStyle{Bold: true}
	phase3Desc := widget.NewLabel("80-90% energy reduction\nTransform data centers\nAI acceleration")
	phase3Desc.Alignment = fyne.TextAlignCenter
	phase3Box := container.NewVBox(
		container.NewCenter(phase3Title),
		container.NewCenter(phase3Name),
		phase3Desc,
	)

	// Assemble phases horizontally
	p.container = container.NewHBox(
		layout.NewSpacer(),
		phase1Box,
		container.NewCenter(arrow1),
		phase2Box,
		container.NewCenter(arrow2),
		phase3Box,
		layout.NewSpacer(),
	)

	return widget.NewSimpleRenderer(p.container)
}

// Refresh updates the widget display.
func (p *PhasedStrategyDiagram) Refresh() {
	if p.container != nil {
		p.container.Refresh()
	}
}

// AnalogStatesComparison placeholder for compatibility.
type AnalogStatesComparison struct {
	widget.BaseWidget
	mu           sync.RWMutex
	animProgress float64
}

// NewAnalogStatesComparison creates a new analog states comparison.
func NewAnalogStatesComparison() *AnalogStatesComparison {
	a := &AnalogStatesComparison{}
	a.ExtendBaseWidget(a)
	return a
}

// UpdateAnimation advances the animation.
func (a *AnalogStatesComparison) UpdateAnimation(dt float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.animProgress < 1.0 {
		a.animProgress += dt * 0.5
		if a.animProgress > 1.0 {
			a.animProgress = 1.0
		}
	}
}

// Reset resets the animation.
func (a *AnalogStatesComparison) Reset() {
	a.mu.Lock()
	a.animProgress = 0
	a.mu.Unlock()
}

// MinSize returns minimum size.
func (a *AnalogStatesComparison) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

// CreateRenderer implements fyne.Widget.
func (a *AnalogStatesComparison) CreateRenderer() fyne.WidgetRenderer {
	text := widget.NewLabel("30 Analog States = ~4.9 bits/cell")
	text.Alignment = fyne.TextAlignCenter
	return widget.NewSimpleRenderer(container.NewCenter(text))
}

// Packet represents a data packet (kept for compatibility).
type Packet struct {
	x, y   float64
	vx     float64
	active bool
}
