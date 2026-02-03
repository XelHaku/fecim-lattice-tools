// Package widgets provides reusable GUI widgets for the hysteresis module.
package widgets

import (
	"fmt"
	"image/color"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// LevelIndicator displays the current N-level state of the ferroelectric cell.
// Implements fyne.Tappable to allow interactive level selection.
type LevelIndicator struct {
	widget.BaseWidget

	mu        sync.RWMutex
	level     int
	numLevels int // Number of discrete levels (default 30)
	minSize   fyne.Size

	// Interactive mode callback - called when user clicks a level
	OnLevelClicked func(targetLevel int)

	// Interactive mode state - controls visual feedback for clickability
	interactive bool // True when clicking is enabled (Manual mode)

	// Target level highlighting (for Write/Read Demo)
	targetLevel     int
	highlightTarget bool
	targetMode      TargetMode

	// Animation state for pulsing highlight
	pulseAnim     *fyne.Animation
	pulseProgress float32 // 0.0 to 1.0, used for pulsing effect

	// L04: Polarization bar pulsing (always on to show current level)
	polBarAnim     *fyne.Animation
	polBarProgress float32
}

// TargetMode identifies how to render the target highlight.
type TargetMode int

const (
	TargetModeNone TargetMode = iota
	TargetModeWrite
	TargetModeVerify
	TargetModeManual
)

// NewLevelIndicator creates a new level indicator with default 30 levels
func NewLevelIndicator() *LevelIndicator {
	l := &LevelIndicator{
		level:     15,
		numLevels: 30,
		minSize:   fyne.NewSize(60, 400),
	}
	l.ExtendBaseWidget(l)
	// L04: Animation started in CreateRenderer (requires running app)
	return l
}

// SetNumLevels changes the number of discrete levels displayed
func (l *LevelIndicator) SetNumLevels(n int) {
	if n < 2 {
		n = 2
	}
	if n > 256 {
		n = 256
	}
	l.mu.Lock()
	l.numLevels = n
	// Clamp current level to new range
	if l.level >= n {
		l.level = n - 1
	}
	l.mu.Unlock()
	l.Refresh()
}

// NumLevels returns the current number of levels
func (l *LevelIndicator) NumLevels() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.numLevels
}

func (l *LevelIndicator) SetMinSize(size fyne.Size) {
	l.minSize = size
}

func (l *LevelIndicator) MinSize() fyne.Size {
	return l.minSize
}

func (l *LevelIndicator) SetLevel(level int) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// SetInteractive sets whether the level indicator is clickable.
// When true, shows "CLICK TO SET" hint; when false, shows "AUTO" indicator.
func (l *LevelIndicator) SetInteractive(interactive bool) {
	l.mu.Lock()
	changed := l.interactive != interactive
	l.interactive = interactive
	l.mu.Unlock()
	if changed {
		l.Refresh()
	}
}

// IsInteractive returns whether clicking is enabled.
func (l *LevelIndicator) IsInteractive() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.interactive
}

// SetTargetLevel sets the target level to highlight (for Write/Read Demo).
// When highlight is true, starts a pulsing animation that auto-refreshes.
func (l *LevelIndicator) SetTargetLevel(level int, highlight bool) {
	l.SetTargetLevelMode(level, highlight, TargetModeWrite)
}

// SetTargetLevelMode sets the target level and rendering mode for clearer feedback.
func (l *LevelIndicator) SetTargetLevelMode(level int, highlight bool, mode TargetMode) {
	l.mu.Lock()
	wasHighlighting := l.highlightTarget
	modeChanged := l.targetLevel != level || l.targetMode != mode || l.highlightTarget != highlight
	l.targetLevel = level
	l.highlightTarget = highlight
	if highlight {
		l.targetMode = mode
	} else {
		l.targetMode = TargetModeNone
	}
	l.mu.Unlock()

	// Manage animation lifecycle
	if highlight && !wasHighlighting {
		// Start pulsing animation
		l.startPulseAnimation()
	} else if !highlight && wasHighlighting {
		// Stop pulsing animation
		l.stopPulseAnimation()
	}
	if modeChanged {
		l.Refresh()
	}
}

// startPulseAnimation starts the continuous pulse animation for target highlight.
func (l *LevelIndicator) startPulseAnimation() {
	// Stop any existing animation first
	l.stopPulseAnimation()

	// Create a looping animation at ~30 FPS equivalent (completes full cycle in 600ms)
	// The animation callback updates pulseProgress and triggers refresh
	l.pulseAnim = fyne.NewAnimation(600*time.Millisecond, func(progress float32) {
		l.mu.Lock()
		l.pulseProgress = progress
		l.mu.Unlock()
		l.Refresh()
	})
	l.pulseAnim.RepeatCount = fyne.AnimationRepeatForever
	l.pulseAnim.AutoReverse = true
	l.pulseAnim.Start()
}

// stopPulseAnimation stops the pulse animation.
func (l *LevelIndicator) stopPulseAnimation() {
	if l.pulseAnim != nil {
		l.pulseAnim.Stop()
		l.pulseAnim = nil
	}
	l.mu.Lock()
	l.pulseProgress = 0
	l.mu.Unlock()
}

// L04: startPolBarAnimation starts the continuous polarization bar pulsing.
func (l *LevelIndicator) startPolBarAnimation() {
	l.polBarAnim = fyne.NewAnimation(800*time.Millisecond, func(progress float32) {
		l.mu.Lock()
		l.polBarProgress = progress
		l.mu.Unlock()
		l.Refresh()
	})
	l.polBarAnim.RepeatCount = fyne.AnimationRepeatForever
	l.polBarAnim.AutoReverse = true
	l.polBarAnim.Start()
}

// Tapped implements fyne.Tappable - allows clicking to select a level
func (l *LevelIndicator) Tapped(e *fyne.PointEvent) {
	if l.OnLevelClicked == nil {
		return
	}
	size := l.Size()
	if size.Height <= 0 {
		return
	}

	l.mu.RLock()
	numLevels := l.numLevels
	l.mu.RUnlock()

	if numLevels < 2 {
		numLevels = 30
	}

	// Match the renderer's layout EXACTLY
	// See layoutWithSize() for the reference implementation
	marginH := float32(50)
	marginBottom := float32(35)
	totalH := size.Height - marginH - marginBottom
	if totalH <= 0 {
		return
	}

	centerY := marginH + totalH/2
	pMaxScale := float32(1.2)
	levelRangeH := totalH / pMaxScale
	levelTop := centerY - levelRangeH/2
	segH := levelRangeH / float32(numLevels)

	// Calculate which segment was clicked
	// Renderer draws: y = levelTop + (numLevels-1-i)*segH where i is 0 to numLevels-1, level = i+1
	// So level N (i=N-1) is at top, level 1 (i=0) is at bottom
	relY := e.Position.Y - levelTop

	// Find segment index from top (0 = level N, N-1 = level 1)
	segFromTop := int(relY / segH)

	// Clamp to valid range
	if segFromTop < 0 {
		segFromTop = 0
	}
	if segFromTop > numLevels-1 {
		segFromTop = numLevels - 1
	}

	// Convert to level: top segment (0) = level N, bottom segment (N-1) = level 1
	targetLevel := numLevels - segFromTop

	l.OnLevelClicked(targetLevel)
}

func (l *LevelIndicator) CreateRenderer() fyne.WidgetRenderer {
	// L04: Start polarization bar animation when renderer is created (app is running)
	if l.polBarAnim == nil {
		l.startPolBarAnimation()
	}
	return &levelRenderer{indicator: l}
}

// Ensure LevelIndicator implements Tappable
var _ fyne.Tappable = (*LevelIndicator)(nil)

type levelRenderer struct {
	indicator *LevelIndicator
	objects   []fyne.CanvasObject
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *levelRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *levelRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("levelRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *levelRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("levelRenderer", r.indicator.Size())
	size := r.indicator.Size()
	// Always re-layout on Refresh for this dynamic widget (level changes)
	r.layoutWithSize(size)
}

func (r *levelRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.indicator.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.indicator.mu.RLock()
	level := r.indicator.level
	numLevels := r.indicator.numLevels
	targetLevel := r.indicator.targetLevel
	highlightTarget := r.indicator.highlightTarget
	targetMode := r.indicator.targetMode
	pulseProgress := r.indicator.pulseProgress
	polBarProgress := r.indicator.polBarProgress
	interactive := r.indicator.interactive
	r.indicator.mu.RUnlock()

	if numLevels < 2 {
		numLevels = 30 // Fallback
	}

	r.objects = r.objects[:0]

	// Suppress unused variable warnings (vars used later in drawing loop)
	_ = targetLevel
	_ = highlightTarget
	_ = targetMode
	_ = pulseProgress

	// Allow level indicator to expand to match plot height
	// Only constrain width to keep it compact
	minSize := r.indicator.minSize
	if size.Width > minSize.Width*1.5 {
		size.Width = minSize.Width * 1.5
	}
	// Don't constrain height - let it match the plot

	// Background with subtle border
	border := canvas.NewRectangle(color.RGBA{0, 100, 180, 255})
	border.Resize(size)
	r.objects = append(r.objects, border)

	bg := canvas.NewRectangle(color.RGBA{0, 40, 80, 255})
	bg.Resize(fyne.NewSize(size.Width-4, size.Height-4))
	bg.Move(fyne.NewPos(2, 2))
	r.objects = append(r.objects, bg)

	// Title at top - shows number of levels
	title := canvas.NewText(fmt.Sprintf("%d LEVELS", numLevels), color.RGBA{0, 212, 255, 255})
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(6, 8))
	r.objects = append(r.objects, title)

	// Bits label - calculate log2(numLevels)
	bitsPerCell := math.Log2(float64(numLevels))
	bitsLabel := canvas.NewText(fmt.Sprintf("%.1f bits", bitsPerCell), color.RGBA{180, 180, 180, 255})
	bitsLabel.TextSize = 12
	bitsLabel.Move(fyne.NewPos(10, 26))
	r.objects = append(r.objects, bitsLabel)

	// Target mode label for clarity
	if highlightTarget && targetLevel > 0 {
		modeText := ""
		switch targetMode {
		case TargetModeWrite:
			modeText = "WRITE"
		case TargetModeVerify:
			modeText = "VERIFY"
		case TargetModeManual:
			modeText = "MANUAL"
		default:
			modeText = "TARGET"
		}
		label := canvas.NewText(fmt.Sprintf("Target L%d (%s)", targetLevel, modeText), color.RGBA{255, 200, 120, 255})
		label.TextSize = 11
		label.Move(fyne.NewPos(8, 42))
		r.objects = append(r.objects, label)
	}

	// Draw 30 level segments
	// Match P-E plot margins for proper Y-axis alignment
	// The plot uses marginTop=50, marginBottom=35
	// However, the plot's legend (at marginTop+10) and +Ec/+Pr labels cause
	// the visual data area to appear shifted down by ~15px
	// Shift level bars down to align with plot's visual ±Pr positions
	visualOffset := float32(15)                // Offset to align with plot's +Pr/-Pr visual positions
	marginH := float32(50) + visualOffset      // Increased top margin shifts levels down
	marginBottom := float32(35) - visualOffset // Decreased bottom margin compensates
	marginW := float32(6)
	labelW := float32(28)
	barW := size.Width - 2*marginW - labelW
	totalH := size.Height - marginH - marginBottom

	// Calculate center Y (same as plot's centerY)
	centerY := marginH + totalH/2

	// The plot shows P from -Ps*1.2 to +Ps*1.2
	// Levels 1-N should map to -Ps to +Ps (the inner 1/1.2 = 83% of plot range)
	pMaxScale := float32(1.2)
	// The actual Y range for the N levels (±Ps portion of the plot)
	levelRangeH := totalH / pMaxScale // ~83% of totalH
	segH := levelRangeH / float32(numLevels)
	gap := float32(1)
	if numLevels > 64 {
		gap = 0 // No gaps for many levels
	}

	// Y positions for level range (screen coords: y increases downward)
	// levelTop = where level 30 (+Ps) should be = centerY - levelRangeH/2
	// levelBottom = where level 1 (-Ps) should be = centerY + levelRangeH/2
	levelTop := centerY - levelRangeH/2

	// Color constants
	colorCurrent := color.RGBA{50, 255, 100, 255} // Green for current level
	colorTarget := color.RGBA{255, 220, 0, 255}   // Yellow for verify/target
	colorTargetWrite := color.RGBA{255, 140, 0, 255}
	colorTargetManual := color.RGBA{140, 200, 255, 255}
	colorAxis := color.RGBA{150, 180, 200, 255}

	for i := 0; i < numLevels; i++ {
		// Level i=0 is level 1 (bottom, -Ps), i=numLevels-1 is level numLevels (top, +Ps)
		// Invert: level N at top, level 1 at bottom
		// y = levelTop + (numLevels-1-i) * segH
		y := levelTop + float32(numLevels-1-i)*segH

		// Color gradient (blue to red)
		t := float64(i) / float64(numLevels-1)
		var segColor color.RGBA
		if i == level {
			// Current level - bright GREEN
			segColor = colorCurrent
		} else if t < 0.5 {
			t2 := t * 2
			segColor = color.RGBA{
				uint8(80 + t2*175),
				uint8(120 + t2*135),
				255,
				180,
			}
		} else {
			t2 := (t - 0.5) * 2
			segColor = color.RGBA{
				255,
				uint8(255 - t2*175),
				uint8(255 - t2*175),
				180,
			}
		}

		// Target level gets pulsing YELLOW border (if highlighted)
		if highlightTarget && i == (targetLevel-1) {
			targetColor := colorTarget
			switch targetMode {
			case TargetModeWrite:
				targetColor = colorTargetWrite
			case TargetModeManual:
				targetColor = colorTargetManual
			case TargetModeVerify:
				targetColor = colorTarget
			}
			// Pulsing effect using animation-driven pulseProgress (0.0 to 1.0)
			// Convert to alpha: pulses between 100 and 255
			pulseAlpha := uint8(100 + 155*pulseProgress)

			// Outer pulse glow - yellow
			targetGlow := canvas.NewRectangle(color.RGBA{targetColor.R, targetColor.G, targetColor.B, pulseAlpha})
			targetGlow.Resize(fyne.NewSize(barW+10, segH+8))
			targetGlow.Move(fyne.NewPos(marginW-5, y-4))
			r.objects = append(r.objects, targetGlow)

			// Inner border - solid yellow
			targetBorder := canvas.NewRectangle(targetColor)
			targetBorder.Resize(fyne.NewSize(barW+4, segH+2))
			targetBorder.Move(fyne.NewPos(marginW-2, y-1))
			r.objects = append(r.objects, targetBorder)
		}

		// Current level gets GREEN glow + L04: 16px pulsing polarization bar
		if i == level {
			// Outer glow
			glow := canvas.NewRectangle(color.RGBA{colorCurrent.R, colorCurrent.G, colorCurrent.B, 100})
			glow.Resize(fyne.NewSize(barW+6, segH+4))
			glow.Move(fyne.NewPos(marginW-3, y-2))
			r.objects = append(r.objects, glow)

			// L04: 16px pulsing polarization bar indicator
			polBarAlpha := uint8(150 + 105*polBarProgress) // Pulses 150-255
			polBar := canvas.NewRectangle(color.RGBA{colorCurrent.R, colorCurrent.G, colorCurrent.B, polBarAlpha})
			polBar.Resize(fyne.NewSize(16, segH+8)) // 16px wide bar
			polBar.Move(fyne.NewPos(size.Width-20, y-4))
			r.objects = append(r.objects, polBar)
		}

		seg := canvas.NewRectangle(segColor)
		seg.Resize(fyne.NewSize(barW, segH-gap))
		seg.Move(fyne.NewPos(marginW, y))
		r.objects = append(r.objects, seg)

		// Level number for key levels and current level
		// Show every Nth level where N scales with total levels
		labelInterval := numLevels / 6
		if labelInterval < 1 {
			labelInterval = 1
		}
		if i%labelInterval == 0 || i == numLevels-1 || i == level {
			labelColor := colorAxis
			fontSize := float32(11)
			if i == level {
				labelColor = colorCurrent // Green for current level
				fontSize = 12
			}
			label := canvas.NewText(fmt.Sprintf("%d", i+1), labelColor)
			label.TextSize = fontSize
			if i == level {
				label.TextStyle = fyne.TextStyle{Bold: true}
			}
			label.Move(fyne.NewPos(marginW+barW+4, y+(segH-gap)/2-6))
			r.objects = append(r.objects, label)
		}
	}

	// Mode indicator at bottom - shows clickability state
	modeY := size.Height - 28
	if interactive {
		// Clickable mode - show "CLICK" hint with pulsing cyan background
		modeBg := canvas.NewRectangle(color.RGBA{0, 180, 220, 150})
		modeBg.Resize(fyne.NewSize(size.Width-8, 22))
		modeBg.Move(fyne.NewPos(4, modeY))
		r.objects = append(r.objects, modeBg)

		modeText := canvas.NewText("CLICK TO SET", color.RGBA{255, 255, 255, 255})
		modeText.TextSize = 10
		modeText.TextStyle = fyne.TextStyle{Bold: true}
		modeText.Move(fyne.NewPos(8, modeY+4))
		r.objects = append(r.objects, modeText)
	} else {
		// Auto mode - show "AUTO" with dimmed background
		modeBg := canvas.NewRectangle(color.RGBA{60, 60, 80, 180})
		modeBg.Resize(fyne.NewSize(size.Width-8, 22))
		modeBg.Move(fyne.NewPos(4, modeY))
		r.objects = append(r.objects, modeBg)

		modeText := canvas.NewText("AUTO", color.RGBA{150, 150, 170, 255})
		modeText.TextSize = 10
		modeText.Move(fyne.NewPos(20, modeY+4))
		r.objects = append(r.objects, modeText)
	}

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
}

func (r *levelRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *levelRenderer) Destroy() {}
