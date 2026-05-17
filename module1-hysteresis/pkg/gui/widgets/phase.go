//go:build legacy_fyne

// Package widgets provides reusable GUI widgets for the hysteresis module.
package widgets

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// PhaseIndicator displays the current state machine phase prominently.
// Shows phases: PROGRAM, VERIFY, RESULT with color-coded backgrounds.
type PhaseIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	phase   int    // Current phase (0-3 for Write/Read Demo)
	mode    string // "wrd" for Write/Read Demo, "manual" for Manual mode
	minSize fyne.Size
}

// Phase constants for Write/Read Demo (3 phases)
const (
	PhaseProgram = 0
	PhaseVerify  = 1
	PhaseResult  = 2
	NumWRDPhases = 3
)

// Manual mode phases (4 phases)
const (
	ManualReset     = 0
	ManualHoldReset = 1
	ManualWrite     = 2
	ManualHold      = 3
	NumManualPhases = 4
)

// Phase colors for Write/Read Demo
var wrdPhaseColors = map[int]color.RGBA{
	PhaseProgram: {255, 100, 100, 255}, // Red for program
	PhaseVerify:  {100, 180, 255, 255}, // Blue for verify
	PhaseResult:  {255, 200, 100, 255}, // Yellow for result
}

// Phase colors for Manual mode
var manualPhaseColors = map[int]color.RGBA{
	ManualReset:     {180, 100, 220, 255}, // Purple for reset
	ManualHoldReset: {100, 150, 200, 255}, // Blue-gray for settle
	ManualWrite:     {255, 100, 100, 255}, // Red for write
	ManualHold:      {100, 200, 100, 255}, // Green for hold
}

// Phase names for Write/Read Demo
var wrdPhaseNames = map[int]string{
	PhaseProgram: "PROGRAM",
	PhaseVerify:  "VERIFY",
	PhaseResult:  "RESULT",
}

// Phase names for Manual mode
var manualPhaseNames = map[int]string{
	ManualReset:     "RESET",
	ManualHoldReset: "SETTLE",
	ManualWrite:     "WRITE",
	ManualHold:      "HOLD",
}

// NewPhaseIndicator creates a new phase indicator widget.
func NewPhaseIndicator() *PhaseIndicator {
	p := &PhaseIndicator{
		phase:   -1, // No phase initially
		mode:    "",
		minSize: fyne.NewSize(140, 50),
	}
	p.ExtendBaseWidget(p)
	return p
}

// SetPhase sets the current phase and triggers a refresh.
func (p *PhaseIndicator) SetPhase(phase int, mode string) {
	p.mu.Lock()
	changed := p.phase != phase || p.mode != mode
	p.phase = phase
	p.mode = mode
	p.mu.Unlock()
	if changed {
		p.Refresh()
	}
}

// SetMinSize sets the minimum size of the widget.
func (p *PhaseIndicator) SetMinSize(size fyne.Size) {
	p.minSize = size
}

// CurrentPhase returns the active phase and mode for diagnostics/tests.
func (p *PhaseIndicator) CurrentPhase() (int, string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.phase, p.mode
}

// MinSize returns the minimum size.
func (p *PhaseIndicator) MinSize() fyne.Size {
	return p.minSize
}

// CreateRenderer implements fyne.Widget.
func (p *PhaseIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &phaseRenderer{indicator: p}
}

type phaseRenderer struct {
	indicator *PhaseIndicator
	objects   []fyne.CanvasObject
}

func (r *phaseRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *phaseRenderer) Layout(size fyne.Size) {
	r.layoutWithSize(size)
}

func (r *phaseRenderer) Refresh() {
	r.layoutWithSize(r.indicator.Size())
}

func (r *phaseRenderer) layoutWithSize(size fyne.Size) {
	if size.Width <= 0 || size.Height <= 0 {
		size = r.indicator.minSize
	}

	r.indicator.mu.RLock()
	phase := r.indicator.phase
	mode := r.indicator.mode
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]

	// Background
	bgColor := color.RGBA{30, 40, 60, 255}
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	if mode == "" || phase < 0 {
		// No active state machine - show idle state
		idleText := canvas.NewText("IDLE", color.RGBA{120, 120, 140, 255})
		idleText.TextSize = 14
		idleText.Move(fyne.NewPos(size.Width/2-15, size.Height/2-8))
		r.objects = append(r.objects, idleText)
		return
	}

	// Determine phases based on mode
	var numPhases int
	var phaseNames map[int]string
	var phaseColors map[int]color.RGBA

	if mode == "wrd" {
		numPhases = NumWRDPhases
		phaseNames = wrdPhaseNames
		phaseColors = wrdPhaseColors
	} else if mode == "manual" {
		numPhases = NumManualPhases
		phaseNames = manualPhaseNames
		phaseColors = manualPhaseColors
	} else {
		// Default to WRD
		numPhases = NumWRDPhases
		phaseNames = wrdPhaseNames
		phaseColors = wrdPhaseColors
	}

	// Draw phase indicator boxes
	boxWidth := (size.Width - 8) / float32(numPhases)
	boxHeight := float32(20)
	boxY := float32(4)

	for i := 0; i < numPhases; i++ {
		boxX := 4 + float32(i)*boxWidth

		// Box color - highlight current phase
		boxColor := color.RGBA{50, 60, 80, 255}
		textColor := color.RGBA{100, 110, 130, 255}
		if i == phase {
			boxColor = phaseColors[i]
			textColor = color.RGBA{255, 255, 255, 255}
		} else if i < phase {
			// Completed phases shown dimmer
			c := phaseColors[i]
			boxColor = color.RGBA{c.R / 3, c.G / 3, c.B / 3, 180}
			textColor = color.RGBA{150, 150, 160, 255}
		}

		// Phase box
		box := canvas.NewRectangle(boxColor)
		box.Resize(fyne.NewSize(boxWidth-2, boxHeight))
		box.Move(fyne.NewPos(boxX, boxY))
		r.objects = append(r.objects, box)

		// Phase label
		labelText := phaseNames[i]
		label := canvas.NewText(labelText, textColor)
		label.TextSize = 14
		if i == phase {
			label.TextStyle = fyne.TextStyle{Bold: true}
		}
		// Center the text in the box
		labelX := boxX + (boxWidth-2)/2 - float32(len(labelText))*2.5
		label.Move(fyne.NewPos(labelX, boxY+5))
		r.objects = append(r.objects, label)
	}

	// Progress bar showing current phase progress (visual indicator)
	progressY := boxY + boxHeight + 3
	progressBg := canvas.NewRectangle(color.RGBA{40, 50, 70, 255})
	progressBg.Resize(fyne.NewSize(size.Width-8, 3))
	progressBg.Move(fyne.NewPos(4, progressY))
	r.objects = append(r.objects, progressBg)

	// Progress fill - shows which phases are complete
	if phase >= 0 && phase < numPhases {
		progressWidth := float32(phase+1) / float32(numPhases) * (size.Width - 8)
		progressFill := canvas.NewRectangle(phaseColors[phase])
		progressFill.Resize(fyne.NewSize(progressWidth, 3))
		progressFill.Move(fyne.NewPos(4, progressY))
		r.objects = append(r.objects, progressFill)
	}

	// Current phase description
	descY := progressY + 8
	descText := ""
	if mode == "wrd" {
		switch phase {
		case PhaseProgram:
			descText = "Applying ISPP pulses..."
		case PhaseVerify:
			descText = "Verifying at E=0"
		case PhaseResult:
			descText = "Result stable at E=0"
		}
	} else if mode == "manual" {
		switch phase {
		case ManualReset:
			descText = "Saturating to known state..."
		case ManualHoldReset:
			descText = "Settling at E=0"
		case ManualWrite:
			descText = "Applying calibrated E-field..."
		case ManualHold:
			descText = "Done! Data persists at E=0"
		}
	}
	desc := canvas.NewText(descText, color.RGBA{180, 190, 210, 255})
	desc.TextSize = 14
	desc.Move(fyne.NewPos(6, descY))
	r.objects = append(r.objects, desc)
}

func (r *phaseRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *phaseRenderer) Destroy() {}
