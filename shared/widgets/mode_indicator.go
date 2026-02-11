// Package widgets provides shared widget utilities for Fyne GUI development.
// This file implements a reusable ModeIndicator widget for displaying
// operation modes with colored backgrounds.
package widgets

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ModeStyle defines the visual style for a mode.
type ModeStyle struct {
	BackgroundColor color.RGBA
	BorderColor     color.RGBA
	Text            string
}

// ModeIndicator is a reusable widget that displays a mode name with
// a colored background. It's designed for "Live Slide" style demos.
type ModeIndicator struct {
	widget.BaseWidget

	mu       sync.RWMutex
	mode     int
	styles   map[int]ModeStyle
	minSize  fyne.Size
	fallback ModeStyle
}

// ModeIndicatorConfig holds configuration for creating a ModeIndicator.
type ModeIndicatorConfig struct {
	MinSize      fyne.Size
	DefaultStyle ModeStyle
	Styles       map[int]ModeStyle
}

// NewModeIndicator creates a new mode indicator widget.
func NewModeIndicator(config ModeIndicatorConfig) *ModeIndicator {
	if config.MinSize.Width <= 0 {
		config.MinSize.Width = 100
	}
	if config.MinSize.Height <= 0 {
		config.MinSize.Height = 40
	}
	if config.DefaultStyle.Text == "" {
		config.DefaultStyle = ModeStyle{
			BackgroundColor: color.RGBA{40, 40, 60, 255},
			BorderColor:     color.RGBA{80, 80, 120, 255},
			Text:            "IDLE",
		}
	}

	m := &ModeIndicator{
		mode:     0,
		styles:   config.Styles,
		minSize:  config.MinSize,
		fallback: config.DefaultStyle,
	}
	if m.styles == nil {
		m.styles = make(map[int]ModeStyle)
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetMode updates the current mode.
func (m *ModeIndicator) SetMode(mode int) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	fyne.Do(func() {
		m.Refresh()
	})
}

// GetMode returns the current mode.
func (m *ModeIndicator) GetMode() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// SetStyle sets or updates the style for a specific mode.
func (m *ModeIndicator) SetStyle(mode int, style ModeStyle) {
	m.mu.Lock()
	m.styles[mode] = style
	m.mu.Unlock()
}

// GetCurrentStyle returns the style for the current mode.
func (m *ModeIndicator) GetCurrentStyle() ModeStyle {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if style, ok := m.styles[m.mode]; ok {
		return style
	}
	return m.fallback
}

// MinSize returns the minimum size for the widget.
func (m *ModeIndicator) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *ModeIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &modeIndicatorRenderer{indicator: m}
}

type modeIndicatorRenderer struct {
	indicator *ModeIndicator
	objects   []fyne.CanvasObject
	cache     LayoutCache
}

func (r *modeIndicatorRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *modeIndicatorRenderer) Layout(size fyne.Size) {
	DebugLayoutCall("modeIndicatorRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
	r.cache.MarkLayout(size)
}

func (r *modeIndicatorRenderer) Refresh() {
	DebugRefreshCall("modeIndicatorRenderer", r.indicator.Size())
	size := r.indicator.Size()
	// Always re-layout on Refresh for this dynamic widget (mode changes)
	r.layoutWithSize(size)
}

func (r *modeIndicatorRenderer) layoutWithSize(size fyne.Size) {
	// Skip layout with invalid sizes
	if size.Width <= 0 || size.Height <= 0 {
		return
	}

	r.objects = r.objects[:0]
	style := r.indicator.GetCurrentStyle()

	// Background
	bg := canvas.NewRectangle(style.BackgroundColor)
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Border
	border := canvas.NewRectangle(style.BorderColor)
	border.StrokeWidth = 2
	border.FillColor = color.Transparent
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Mode text - scale with widget size
	fontSize := size.Height * 0.3
	if fontSize < 14 {
		fontSize = 10
	}
	if fontSize > 16 {
		fontSize = 16
	}
	modeText := canvas.NewText(style.Text, color.White)
	modeText.TextSize = fontSize
	modeText.TextStyle = fyne.TextStyle{Bold: true}
	textWidth := float32(len(style.Text)) * fontSize * 0.55
	modeText.Move(fyne.NewPos((size.Width-textWidth)/2, (size.Height-fontSize)/2))
	r.objects = append(r.objects, modeText)
}

func (r *modeIndicatorRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *modeIndicatorRenderer) Destroy() {}

// DefaultModeStyles returns a set of common mode styles.
func DefaultModeStyles() map[int]ModeStyle {
	return map[int]ModeStyle{
		0: { // Idle
			BackgroundColor: color.RGBA{40, 40, 60, 255},
			BorderColor:     color.RGBA{80, 80, 120, 255},
			Text:            "IDLE",
		},
		1: { // Processing/Working
			BackgroundColor: color.RGBA{80, 120, 50, 255},
			BorderColor:     color.RGBA{140, 200, 100, 255},
			Text:            "PROCESSING",
		},
		2: { // Complete/Success
			BackgroundColor: color.RGBA{50, 150, 80, 255},
			BorderColor:     color.RGBA{100, 220, 130, 255},
			Text:            "COMPLETE",
		},
		3: { // Error/Warning
			BackgroundColor: color.RGBA{180, 80, 50, 255},
			BorderColor:     color.RGBA{255, 140, 100, 255},
			Text:            "ERROR",
		},
	}
}
