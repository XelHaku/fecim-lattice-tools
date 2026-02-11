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

// estimatedColor is used for model-input values (amber)
var estimatedColor = color.RGBA{255, 191, 0, 255}

// Technical briefing colors
var (
	heroTextColor  = color.RGBA{240, 244, 248, 255} // Off-white for maximum contrast
	heroCyanColor  = color.RGBA{0, 212, 255, 255}   // FeCIM cyan accent
	heroGreenColor = color.RGBA{46, 204, 113, 255}  // Success green
	heroRedColor   = color.RGBA{231, 76, 60, 255}   // GPU baseline red
	heroAmberColor = color.RGBA{243, 156, 18, 255}  // Warning/caution amber
	heroMutedColor = color.RGBA{160, 180, 200, 255} // Secondary text
)

// Energy values in picojoules per MAC (model inputs from docs/videos/ironlattice-youtube-script.md)
// "CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. FeCIM: under 1 picojoule."
const (
	cpuEnergyPJ   = 1000.0 // 1000 pJ/MAC
	gpuEnergyPJ   = 100.0  // 100 pJ/MAC
	fecimEnergyPJ = 1.0    // ~1 pJ/MAC (conservative estimate for claimed "<1 pJ")
)

// AnimatedEnergyRace shows the HERO energy comparison - investor grade.
// HERO STATEMENT: "80-90% data center energy reduction" (model input headline)
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
	container   *fyne.Container
	heroText    *canvas.Text
	heroSubtext *canvas.Text
	gpuBar      *canvas.Rectangle
	fecimBar    *canvas.Rectangle
	gpuLabel    *canvas.Text
	fecimLabel  *canvas.Text
	statStrip   *canvas.Text
	renderer    fyne.WidgetRenderer
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
	return fyne.NewSize(600, 180)
}

// CreateRenderer implements fyne.Widget.
// BUG-M5-003 FIX: Cache renderer and reuse canvas objects instead of recreating
func (e *AnimatedEnergyRace) CreateRenderer() fyne.WidgetRenderer {
	// Return cached renderer if already created (prevents recreation)
	if e.renderer != nil {
		return e.renderer
	}

	// === HERO SECTION: COMPACT "80-90%" ===
	e.heroText = canvas.NewText("80-90%", heroTextColor)
	e.heroText.TextSize = 48 // Compact for unified view
	e.heroText.TextStyle = fyne.TextStyle{Bold: true}
	e.heroText.Alignment = fyne.TextAlignCenter

	e.heroSubtext = canvas.NewText("⚠️ SIMULATION ONLY | MODEL INPUTS | TRL 4 Lab Context", heroAmberColor)
	e.heroSubtext.TextSize = 14
	e.heroSubtext.TextStyle = fyne.TextStyle{Bold: true}
	e.heroSubtext.Alignment = fyne.TextAlignCenter

	heroSection := container.NewVBox(
		container.NewCenter(e.heroText),
		container.NewCenter(e.heroSubtext),
	)

	// === BEFORE/AFTER COMPARISON: Prominent bars for investor impact ===
	barWidth := float32(500)
	barHeight := float32(32)

	// "Today: GPU AI" label (shortened to fit)
	todayLabel := canvas.NewText("Today: GPU AI", heroMutedColor)
	todayLabel.TextSize = 14
	todayLabel.Alignment = fyne.TextAlignLeading

	// GPU bar - full width (100 units = 100%)
	e.gpuBar = canvas.NewRectangle(heroRedColor)
	e.gpuBar.SetMinSize(fyne.NewSize(barWidth, barHeight))
	e.gpuLabel = canvas.NewText("100 units", heroTextColor)
	e.gpuLabel.TextSize = 14
	e.gpuLabel.Alignment = fyne.TextAlignTrailing

	gpuRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(160, barHeight), container.NewCenter(todayLabel)),
		e.gpuBar,
		container.NewGridWrap(fyne.NewSize(120, barHeight), container.NewCenter(e.gpuLabel)),
	)

	// "FeCIM" label (shortened to fit)
	fecimLabelText := canvas.NewText("FeCIM", heroMutedColor)
	fecimLabelText.TextSize = 14
	fecimLabelText.Alignment = fyne.TextAlignLeading

	// FeCIM bar - ~10% width (10 units = 90% reduction)
	e.fecimBar = canvas.NewRectangle(heroGreenColor)
	e.fecimBar.SetMinSize(fyne.NewSize(barWidth*0.1, barHeight))
	e.fecimLabel = canvas.NewText("~10 units", heroTextColor)
	e.fecimLabel.TextSize = 14
	e.fecimLabel.Alignment = fyne.TextAlignTrailing

	fecimRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(160, barHeight), container.NewCenter(fecimLabelText)),
		e.fecimBar,
		container.NewGridWrap(fyne.NewSize(120, barHeight), container.NewCenter(e.fecimLabel)),
	)

	comparisonSection := container.NewVBox(
		gpuRow,
		fecimRow,
	)

	// === KEY STAT STRIP ===
	e.statStrip = canvas.NewText("Model inputs: 1000× vs CPU  |  100× vs GPU  |  ~1 pJ/MAC (TRL 4)", heroCyanColor)
	e.statStrip.TextSize = 14
	e.statStrip.TextStyle = fyne.TextStyle{Bold: true}
	e.statStrip.Alignment = fyne.TextAlignCenter

	// === SYSTEM POWER BREAKDOWN (C10: Per Dr. Tour critique) ===
	// Shows total system power distribution, not just array power
	powerBreakdown := canvas.NewText("Model assumption: System Power = Array ~45% | ADC/DAC ~40% | Peripherals ~15%", heroAmberColor)
	powerBreakdown.TextSize = 14
	powerBreakdown.TextStyle = fyne.TextStyle{Italic: true}
	powerBreakdown.Alignment = fyne.TextAlignCenter

	// === CITATION ===
	citation := canvas.NewText("Model input references (not validated): Samsung Nature 2025 range, NVIDIA H100 datasheets, Intel/AMD datasheets", heroMutedColor)
	citation.TextSize = 14
	citation.TextStyle = fyne.TextStyle{Italic: true}
	citation.Alignment = fyne.TextAlignCenter

	// === ASSEMBLE ===
	e.container = container.NewVBox(
		heroSection,
		comparisonSection,
		container.NewCenter(e.statStrip),
		container.NewCenter(powerBreakdown),
		container.NewCenter(citation),
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

	// Animate bar widths - prominent for investor impact
	barWidth := float32(500)
	barHeight := float32(32)

	// Prepare text updates if needed
	var gpuText, fecimText string
	if needsTextUpdate {
		if progress > 0.9 {
			gpuText = "100 units"
			fecimText = "~10 units"
		} else {
			gpuText = fmt.Sprintf("%.0f units", 100*progress)
			fecimText = fmt.Sprintf("~%.0f units", 10*progress)
		}
	}

	// Prepare hero text color if needed
	var heroColor color.RGBA
	if showWinner {
		pulse := 0.85 + math.Sin(pulsePhase)*0.15
		heroColor = color.RGBA{
			uint8(240 * pulse),
			uint8(244 * pulse),
			uint8(248 * pulse),
			255,
		}
	}

	// All UI updates must be on main thread
	fyne.Do(func() {
		e.gpuBar.SetMinSize(fyne.NewSize(barWidth*float32(progress), barHeight))
		e.fecimBar.SetMinSize(fyne.NewSize(max(10, barWidth*0.1*float32(progress)), barHeight))

		if needsTextUpdate {
			e.gpuLabel.Text = gpuText
			e.fecimLabel.Text = fecimText
			canvas.Refresh(e.gpuLabel)
			canvas.Refresh(e.fecimLabel)
		}

		if showWinner {
			e.heroText.Color = heroColor
			canvas.Refresh(e.heroText)
		}

		canvas.Refresh(e.gpuBar)
		canvas.Refresh(e.fecimBar)
		e.container.Refresh()
	})

	// Clear the flag after UI update
	if needsTextUpdate {
		e.mu.Lock()
		e.needsTextUpdate = false
		e.mu.Unlock()
	}
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
	return fyne.NewSize(600, 120)
}

// CreateRenderer implements fyne.Widget.
func (p *PhasedStrategyDiagram) CreateRenderer() fyne.WidgetRenderer {
	// Phase 1: NAND Replacement
	phase1Title := canvas.NewText("PHASE 1", heroCyanColor)
	phase1Title.TextSize = 14
	phase1Title.TextStyle = fyne.TextStyle{Bold: true}
	phase1Name := canvas.NewText("NAND Replacement", heroTextColor)
	phase1Name.TextSize = 14
	phase1Name.TextStyle = fyne.TextStyle{Bold: true}
	phase1Desc := canvas.NewText("No software changes", heroMutedColor)
	phase1Desc.TextSize = 14
	phase1Box := container.NewVBox(
		container.NewCenter(phase1Title),
		container.NewCenter(phase1Name),
		container.NewCenter(phase1Desc),
	)

	// Arrow 1
	arrow1 := canvas.NewText("→", heroMutedColor)
	arrow1.TextSize = 16

	// Phase 2: DRAM Replacement
	phase2Title := canvas.NewText("PHASE 2", heroCyanColor)
	phase2Title.TextSize = 14
	phase2Title.TextStyle = fyne.TextStyle{Bold: true}
	phase2Name := canvas.NewText("DRAM Replacement", heroTextColor)
	phase2Name.TextSize = 14
	phase2Name.TextStyle = fyne.TextStyle{Bold: true}
	phase2Desc := canvas.NewText("Higher density", heroMutedColor)
	phase2Desc.TextSize = 14
	phase2Box := container.NewVBox(
		container.NewCenter(phase2Title),
		container.NewCenter(phase2Name),
		container.NewCenter(phase2Desc),
	)

	// Arrow 2
	arrow2 := canvas.NewText("→", heroMutedColor)
	arrow2.TextSize = 16

	// Phase 3: Full CIM
	phase3Title := canvas.NewText("PHASE 3", heroGreenColor)
	phase3Title.TextSize = 14
	phase3Title.TextStyle = fyne.TextStyle{Bold: true}
	phase3Name := canvas.NewText("Full CIM", heroTextColor)
	phase3Name.TextSize = 14
	phase3Name.TextStyle = fyne.TextStyle{Bold: true}
	phase3Desc := canvas.NewText("80-90% model reduction", heroMutedColor)
	phase3Desc.TextSize = 14
	phase3Box := container.NewVBox(
		container.NewCenter(phase3Title),
		container.NewCenter(phase3Name),
		container.NewCenter(phase3Desc),
	)

	// === CITATION ===
	citation := canvas.NewText("Strategy based on WSTS/Gartner market analysis", heroMutedColor)
	citation.TextSize = 14
	citation.TextStyle = fyne.TextStyle{Italic: true}
	citation.Alignment = fyne.TextAlignCenter

	// Assemble phases horizontally
	phasesRow := container.NewHBox(
		layout.NewSpacer(),
		phase1Box,
		container.NewCenter(arrow1),
		phase2Box,
		container.NewCenter(arrow2),
		phase3Box,
		layout.NewSpacer(),
	)

	p.container = container.NewVBox(
		phasesRow,
		container.NewCenter(citation),
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
	text := widget.NewLabel("30 Analog States (model input; conference claim) ≈ 4.9 bits/cell")
	text.Alignment = fyne.TextAlignCenter
	return widget.NewSimpleRenderer(container.NewCenter(text))
}

// Packet represents a data packet (kept for compatibility).
type Packet struct {
	x, y   float64
	vx     float64
	active bool
}
