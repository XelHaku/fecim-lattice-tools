//go:build legacy_fyne

package widgets

import (
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MicLevel is a widget that displays a microphone icon with level indicator.
// The icon color changes from gray to green proportional to sound level.
type MicLevel struct {
	widget.BaseWidget

	mu           sync.RWMutex
	level        int // 0-100
	peak         int // 0-100
	isMonitoring bool
	isRecording  bool

	// Visual components
	icon      *canvas.Image
	levelBar  *canvas.Rectangle
	peakBar   *canvas.Rectangle
	container *fyne.Container

	// Colors
	colorSilent    color.Color
	colorLow       color.Color
	colorMedium    color.Color
	colorHigh      color.Color
	colorRecording color.Color
}

// NewMicLevel creates a new microphone level indicator widget.
func NewMicLevel() *MicLevel {
	ml := &MicLevel{
		level:          0,
		peak:           0,
		colorSilent:    color.NRGBA{R: 128, G: 128, B: 128, A: 255}, // Gray
		colorLow:       color.NRGBA{R: 0, G: 180, B: 0, A: 255},     // Green
		colorMedium:    color.NRGBA{R: 200, G: 200, B: 0, A: 255},   // Yellow
		colorHigh:      color.NRGBA{R: 220, G: 50, B: 50, A: 255},   // Red
		colorRecording: color.NRGBA{R: 255, G: 0, B: 0, A: 255},     // Bright red
	}
	ml.ExtendBaseWidget(ml)
	return ml
}

// SetLevel sets the current audio level (0-100).
func (ml *MicLevel) SetLevel(level int) {
	ml.mu.Lock()
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	ml.level = level
	ml.mu.Unlock()

	ml.Refresh()
}

// SetPeak sets the peak audio level (0-100).
func (ml *MicLevel) SetPeak(peak int) {
	ml.mu.Lock()
	if peak < 0 {
		peak = 0
	}
	if peak > 100 {
		peak = 100
	}
	ml.peak = peak
	ml.mu.Unlock()

	ml.Refresh()
}

// SetLevelAndPeak sets both level and peak in one call.
func (ml *MicLevel) SetLevelAndPeak(level, peak int) {
	ml.mu.Lock()
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	if peak < 0 {
		peak = 0
	}
	if peak > 100 {
		peak = 100
	}
	ml.level = level
	ml.peak = peak
	ml.mu.Unlock()

	ml.Refresh()
}

// SetMonitoring sets whether audio monitoring is active.
func (ml *MicLevel) SetMonitoring(monitoring bool) {
	ml.mu.Lock()
	ml.isMonitoring = monitoring
	ml.mu.Unlock()
	ml.Refresh()
}

// SetRecording sets whether audio recording is active.
func (ml *MicLevel) SetRecording(recording bool) {
	ml.mu.Lock()
	ml.isRecording = recording
	ml.mu.Unlock()
	ml.Refresh()
}

// Level returns the current audio level.
func (ml *MicLevel) Level() int {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return ml.level
}

// Peak returns the current peak level.
func (ml *MicLevel) Peak() int {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return ml.peak
}

// levelColor returns the appropriate color for the given level.
func (ml *MicLevel) levelColor(level int) color.Color {
	if level <= 0 {
		return ml.colorSilent
	}
	if level < 50 {
		// Interpolate between silent (gray) and low (green)
		t := float64(level) / 50.0
		return interpolateColor(ml.colorSilent, ml.colorLow, t)
	}
	if level < 80 {
		// Interpolate between low (green) and medium (yellow)
		t := float64(level-50) / 30.0
		return interpolateColor(ml.colorLow, ml.colorMedium, t)
	}
	// Interpolate between medium (yellow) and high (red)
	t := float64(level-80) / 20.0
	return interpolateColor(ml.colorMedium, ml.colorHigh, t)
}

// interpolateColor linearly interpolates between two colors.
func interpolateColor(c1, c2 color.Color, t float64) color.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return color.NRGBA{
		R: uint8(float64(r1>>8)*(1-t) + float64(r2>>8)*t),
		G: uint8(float64(g1>>8)*(1-t) + float64(g2>>8)*t),
		B: uint8(float64(b1>>8)*(1-t) + float64(b2>>8)*t),
		A: uint8(float64(a1>>8)*(1-t) + float64(a2>>8)*t),
	}
}

// CreateRenderer creates the renderer for the widget.
func (ml *MicLevel) CreateRenderer() fyne.WidgetRenderer {
	ml.mu.RLock()
	level := ml.level
	peak := ml.peak
	isRecording := ml.isRecording
	ml.mu.RUnlock()

	// Create mic icon using theme resource
	micIcon := canvas.NewImageFromResource(theme.VolumeUpIcon())
	micIcon.FillMode = canvas.ImageFillContain
	micIcon.SetMinSize(fyne.NewSize(24, 24))

	// Create level bar (vertical bar next to icon)
	levelBar := canvas.NewRectangle(ml.levelColor(level))
	levelBar.SetMinSize(fyne.NewSize(4, 24))

	// Create peak indicator
	peakBar := canvas.NewRectangle(ml.colorHigh)
	peakBar.SetMinSize(fyne.NewSize(4, 2))

	// Recording indicator (red dot)
	recordDot := canvas.NewCircle(ml.colorRecording)
	if !isRecording {
		recordDot.FillColor = color.Transparent
	}

	return &micLevelRenderer{
		widget:    ml,
		micIcon:   micIcon,
		levelBar:  levelBar,
		peakBar:   peakBar,
		recordDot: recordDot,
		level:     level,
		peak:      peak,
	}
}

// micLevelRenderer implements the fyne.WidgetRenderer interface.
type micLevelRenderer struct {
	widget    *MicLevel
	micIcon   *canvas.Image
	levelBar  *canvas.Rectangle
	peakBar   *canvas.Rectangle
	recordDot *canvas.Circle
	level     int
	peak      int
}

func (r *micLevelRenderer) Layout(size fyne.Size) {
	// Icon on the left
	iconSize := fyne.NewSize(24, 24)
	r.micIcon.Resize(iconSize)
	r.micIcon.Move(fyne.NewPos(0, (size.Height-iconSize.Height)/2))

	// Level bar next to icon
	barWidth := float32(4)
	barMaxHeight := size.Height - 4

	r.widget.mu.RLock()
	level := r.widget.level
	peak := r.widget.peak
	r.widget.mu.RUnlock()

	// Level bar height proportional to level
	levelHeight := barMaxHeight * float32(level) / 100
	r.levelBar.Resize(fyne.NewSize(barWidth, levelHeight))
	r.levelBar.Move(fyne.NewPos(iconSize.Width+2, size.Height-2-levelHeight))

	// Peak indicator
	peakY := size.Height - 2 - (barMaxHeight * float32(peak) / 100)
	r.peakBar.Resize(fyne.NewSize(barWidth, 2))
	r.peakBar.Move(fyne.NewPos(iconSize.Width+2, peakY))

	// Recording dot in top-right of icon
	r.recordDot.Resize(fyne.NewSize(6, 6))
	r.recordDot.Move(fyne.NewPos(iconSize.Width-6, 0))
}

func (r *micLevelRenderer) MinSize() fyne.Size {
	return fyne.NewSize(34, 28) // Icon + level bar
}

func (r *micLevelRenderer) Refresh() {
	r.widget.mu.RLock()
	level := r.widget.level
	peak := r.widget.peak
	isRecording := r.widget.isRecording
	isMonitoring := r.widget.isMonitoring
	r.widget.mu.RUnlock()

	r.level = level
	r.peak = peak

	// Update level bar color
	r.levelBar.FillColor = r.widget.levelColor(level)
	r.levelBar.Refresh()

	// Update peak bar visibility
	if peak > 0 {
		r.peakBar.FillColor = r.widget.colorHigh
	} else {
		r.peakBar.FillColor = color.Transparent
	}
	r.peakBar.Refresh()

	// Update recording dot
	if isRecording {
		r.recordDot.FillColor = r.widget.colorRecording
	} else {
		r.recordDot.FillColor = color.Transparent
	}
	r.recordDot.Refresh()

	// Tint the mic icon based on monitoring state
	if isMonitoring && level > 0 {
		// Icon tint not directly supported, but we can update the visual
	}

	// Re-layout for level changes
	r.Layout(r.widget.Size())

	canvas.Refresh(r.widget)
}

func (r *micLevelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.micIcon, r.levelBar, r.peakBar, r.recordDot}
}

func (r *micLevelRenderer) Destroy() {}

// =============================================================================
// MicLevelController connects AudioMonitor to MicLevel widget
// =============================================================================

// MicLevelController connects an AudioMonitor to a MicLevel widget.
type MicLevelController struct {
	widget  *MicLevel
	monitor interface {
		Level() int
		PeakLevel() int
		IsRunning() bool
		OnLevelChange(callback func(level, peak int))
		Start() error
		Stop()
	}

	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewMicLevelController creates a controller that updates the widget from the monitor.
func NewMicLevelController(widget *MicLevel, monitor interface {
	Level() int
	PeakLevel() int
	IsRunning() bool
	OnLevelChange(callback func(level, peak int))
	Start() error
	Stop()
}) *MicLevelController {
	return &MicLevelController{
		widget:  widget,
		monitor: monitor,
	}
}

// Start begins updating the widget from the monitor.
func (c *MicLevelController) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	// Set up callback from monitor
	c.monitor.OnLevelChange(func(level, peak int) {
		fyne.Do(func() {
			c.widget.SetLevelAndPeak(level, peak)
		})
	})

	// Start the monitor if not running
	if !c.monitor.IsRunning() {
		if err := c.monitor.Start(); err != nil {
			return err
		}
	}

	c.widget.SetMonitoring(true)
	c.running = true
	c.stopChan = make(chan struct{})

	// Start a fallback polling loop in case callbacks don't work
	go c.pollLoop()

	return nil
}

// Stop stops updating the widget.
func (c *MicLevelController) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}

	c.running = false
	c.widget.SetMonitoring(false)
	c.widget.SetLevelAndPeak(0, 0)
}

// pollLoop polls the monitor for level updates as a fallback.
func (c *MicLevelController) pollLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			if c.monitor.IsRunning() {
				level := c.monitor.Level()
				peak := c.monitor.PeakLevel()
				fyne.Do(func() {
					c.widget.SetLevelAndPeak(level, peak)
				})
			}
		}
	}
}

// SetRecording sets the recording state on the widget.
func (c *MicLevelController) SetRecording(recording bool) {
	c.widget.SetRecording(recording)
}

// =============================================================================
// Simple Mic Button with Level
// =============================================================================

// MicButton is a button with integrated mic level indicator.
type MicButton struct {
	widget.BaseWidget

	micLevel *MicLevel
	button   *widget.Button
	onTap    func()
}

// NewMicButton creates a new mic button with level indicator.
func NewMicButton(onTap func()) *MicButton {
	mb := &MicButton{
		micLevel: NewMicLevel(),
		onTap:    onTap,
	}
	mb.button = widget.NewButton("", func() {
		if mb.onTap != nil {
			mb.onTap()
		}
	})
	mb.button.Importance = widget.LowImportance
	mb.ExtendBaseWidget(mb)
	return mb
}

// MicLevel returns the mic level widget for external control.
func (mb *MicButton) MicLevel() *MicLevel {
	return mb.micLevel
}

// SetEnabled enables or disables the button.
func (mb *MicButton) SetEnabled(enabled bool) {
	if enabled {
		mb.button.Enable()
	} else {
		mb.button.Disable()
	}
}

// CreateRenderer creates the renderer for the widget.
func (mb *MicButton) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(mb.button, container.NewCenter(mb.micLevel))
	return widget.NewSimpleRenderer(c)
}
