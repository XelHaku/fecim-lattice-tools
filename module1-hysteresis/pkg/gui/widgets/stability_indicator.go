//go:build legacy_fyne

// Package widgets provides GUI components for the hysteresis demo.
// This file implements M12: State stability indicator with color-coded warnings.
package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// StabilityLevel indicates the stability category of a memory state.
type StabilityLevel int

const (
	// StabilityHigh indicates highly stable states (levels 1-5, 26-30 are edge states)
	StabilityHigh StabilityLevel = iota
	// StabilityMedium indicates moderately stable states (levels 6-10, 21-25)
	StabilityMedium
	// StabilityLow indicates less stable states (levels 11-20 - middle range)
	StabilityLow
)

// StabilityIndicator shows color-coded stability warnings for memory states.
// Per M12: Levels near saturation (1-5, 26-30) are most stable,
// middle levels (11-20) are least stable due to noise margins.
type StabilityIndicator struct {
	widget.BaseWidget
	currentLevel int
	numLevels    int
	stability    StabilityLevel
	showWarning  bool
}

// NewStabilityIndicator creates a new stability indicator widget.
func NewStabilityIndicator() *StabilityIndicator {
	s := &StabilityIndicator{
		currentLevel: 15,
		numLevels:    30,
		stability:    StabilityLow,
		showWarning:  true,
	}
	s.ExtendBaseWidget(s)
	return s
}

// SetLevel updates the current level and recalculates stability.
func (s *StabilityIndicator) SetLevel(level, numLevels int) {
	s.currentLevel = level
	s.numLevels = numLevels
	s.stability = s.calculateStability(level, numLevels)
	s.Refresh()
}

// SetShowWarning enables or disables the warning display.
func (s *StabilityIndicator) SetShowWarning(show bool) {
	s.showWarning = show
	s.Refresh()
}

// GetStability returns the current stability level.
func (s *StabilityIndicator) GetStability() StabilityLevel {
	return s.stability
}

// calculateStability determines the stability category based on level position.
// Physics rationale:
// - Edge states (near ±Ps) have clear separation from adjacent levels
// - Middle states have smaller noise margins and are more susceptible to drift
// - For 30 levels: 1-5 and 26-30 are most stable (near saturation)
func (s *StabilityIndicator) calculateStability(level, numLevels int) StabilityLevel {
	if numLevels <= 0 {
		return StabilityMedium
	}

	// Calculate position as fraction from center (0.0 = center, 1.0 = edge)
	center := float64(numLevels+1) / 2.0
	distFromCenter := float64(level) - center
	if distFromCenter < 0 {
		distFromCenter = -distFromCenter
	}
	normalizedDist := distFromCenter / center // 0.0 = center, 1.0 = edge

	// Edge states (top/bottom 20%) are most stable
	if normalizedDist > 0.7 {
		return StabilityHigh
	}
	// States between 20-40% from edge have medium stability
	if normalizedDist > 0.3 {
		return StabilityMedium
	}
	// Center states (middle 60%) are least stable
	return StabilityLow
}

// CreateRenderer implements fyne.Widget.
func (s *StabilityIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &stabilityIndicatorRenderer{
		indicator: s,
	}
}

// MinSize returns the minimum size for the stability indicator.
func (s *StabilityIndicator) MinSize() fyne.Size {
	return fyne.NewSize(140, 24)
}

// stabilityIndicatorRenderer handles rendering of the stability indicator.
type stabilityIndicatorRenderer struct {
	indicator *StabilityIndicator
}

// getStabilityColor returns the color for the current stability level.
func (r *stabilityIndicatorRenderer) getStabilityColor() color.Color {
	switch r.indicator.stability {
	case StabilityHigh:
		return color.RGBA{50, 205, 50, 255} // Green - stable
	case StabilityMedium:
		return color.RGBA{255, 200, 0, 255} // Yellow/amber - moderate
	case StabilityLow:
		return color.RGBA{255, 100, 100, 255} // Red/pink - edge case
	default:
		return color.RGBA{150, 150, 150, 255}
	}
}

// getStabilityText returns the text description for the current stability level.
func (r *stabilityIndicatorRenderer) getStabilityText() string {
	level := r.indicator.currentLevel
	switch r.indicator.stability {
	case StabilityHigh:
		return fmt.Sprintf("L%d ● STABLE", level)
	case StabilityMedium:
		return fmt.Sprintf("L%d ◐ MODERATE", level)
	case StabilityLow:
		return fmt.Sprintf("L%d ○ LOW MARGIN", level)
	default:
		return fmt.Sprintf("L%d", level)
	}
}

// getStabilityTooltip returns detailed tooltip text for the current stability.
func (r *stabilityIndicatorRenderer) getStabilityTooltip() string {
	level := r.indicator.currentLevel
	numLevels := r.indicator.numLevels

	switch r.indicator.stability {
	case StabilityHigh:
		return fmt.Sprintf("Level %d/%d: High stability\nNear saturation polarization\nLarge noise margin to adjacent levels", level, numLevels)
	case StabilityMedium:
		return fmt.Sprintf("Level %d/%d: Moderate stability\nIntermediate polarization region\nAdequate noise margin", level, numLevels)
	case StabilityLow:
		return fmt.Sprintf("Level %d/%d: Low margin\nCenter of polarization range\nSmaller separation from adjacent levels\nMore susceptible to drift/noise", level, numLevels)
	default:
		return fmt.Sprintf("Level %d/%d", level, numLevels)
	}
}

// Layout positions the objects within the widget.
func (r *stabilityIndicatorRenderer) Layout(size fyne.Size) {
	// Layout handled by container
}

// MinSize returns the minimum size.
func (r *stabilityIndicatorRenderer) MinSize() fyne.Size {
	return r.indicator.MinSize()
}

// Refresh updates the renderer.
func (r *stabilityIndicatorRenderer) Refresh() {
	// Handled by Objects() returning new content
}

// Objects returns the objects to render.
func (r *stabilityIndicatorRenderer) Objects() []fyne.CanvasObject {
	if !r.indicator.showWarning {
		return []fyne.CanvasObject{}
	}

	stabilityColor := r.getStabilityColor()
	stabilityText := r.getStabilityText()

	// Background with stability color
	bg := canvas.NewRectangle(color.RGBA{30, 30, 40, 255})
	bg.SetMinSize(r.indicator.MinSize())

	// Colored indicator bar on left
	indicatorBar := canvas.NewRectangle(stabilityColor)
	indicatorBar.SetMinSize(fyne.NewSize(4, 20))

	// Status text
	statusText := canvas.NewText(stabilityText, stabilityColor)
	statusText.TextSize = 14
	statusText.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewHBox(
		indicatorBar,
		container.NewCenter(statusText),
	)

	return []fyne.CanvasObject{
		container.NewStack(bg, container.NewPadded(content)),
	}
}

// Destroy cleans up resources.
func (r *stabilityIndicatorRenderer) Destroy() {}
