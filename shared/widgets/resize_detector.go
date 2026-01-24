// Package widgets provides reusable Fyne widget components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ResizeCallback is called when the widget's size changes.
type ResizeCallback func(size fyne.Size)

// ResizeDetector is a transparent widget that fires callbacks when resized.
// Use this to detect window size changes and implement responsive layouts.
//
// Usage:
//
//	detector := widgets.NewResizeDetector(func(size fyne.Size) {
//	    if size.Width < 768 {
//	        // Switch to mobile layout
//	    } else {
//	        // Switch to desktop layout
//	    }
//	})
//	content := container.NewStack(mainContent, detector)
type ResizeDetector struct {
	widget.BaseWidget

	// OnResize is called when the widget is resized.
	// The callback receives the new size in device-independent pixels (dp).
	OnResize ResizeCallback

	// lastSize tracks the previous size to detect changes
	lastSize fyne.Size
}

// NewResizeDetector creates a new resize detector with the given callback.
func NewResizeDetector(onResize ResizeCallback) *ResizeDetector {
	rd := &ResizeDetector{
		OnResize: onResize,
	}
	rd.ExtendBaseWidget(rd)
	return rd
}

// CreateRenderer implements fyne.Widget.
func (rd *ResizeDetector) CreateRenderer() fyne.WidgetRenderer {
	// Create a fully transparent rectangle as the visual element
	rect := canvas.NewRectangle(nil) // nil color = fully transparent
	return &resizeDetectorRenderer{
		detector: rd,
		rect:     rect,
	}
}

// resizeDetectorRenderer is the renderer for ResizeDetector.
type resizeDetectorRenderer struct {
	detector *ResizeDetector
	rect     *canvas.Rectangle
}

// Layout is called when the widget needs to be laid out.
// This is where we detect resize events.
func (r *resizeDetectorRenderer) Layout(size fyne.Size) {
	DebugLayoutCall("resizeDetectorRenderer", size)
	// Resize the internal rectangle to fill the container
	r.rect.Resize(size)

	// Check if size has changed
	if size != r.detector.lastSize {
		r.detector.lastSize = size

		// Fire the callback if set
		if r.detector.OnResize != nil {
			r.detector.OnResize(size)
		}
	}
}

// MinSize returns the minimum size (zero, as this is a passive detector).
func (r *resizeDetectorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

// Refresh refreshes the renderer.
func (r *resizeDetectorRenderer) Refresh() {
	DebugRefreshCall("resizeDetectorRenderer", r.detector.Size())
	r.rect.Refresh()
}

// Objects returns the canvas objects for rendering.
func (r *resizeDetectorRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect}
}

// Destroy cleans up the renderer.
func (r *resizeDetectorRenderer) Destroy() {}

// Breakpoint represents a responsive design breakpoint.
type Breakpoint int

const (
	// BreakpointSM is for small devices (mobile phones) - width <= 576dp
	BreakpointSM Breakpoint = iota
	// BreakpointMD is for medium devices (tablets) - width <= 768dp
	BreakpointMD
	// BreakpointLG is for large devices (small laptops) - width <= 992dp
	BreakpointLG
	// BreakpointXL is for extra large devices (desktops) - width > 992dp
	BreakpointXL
)

// Breakpoint thresholds in device-independent pixels
const (
	ThresholdSM = 576
	ThresholdMD = 768
	ThresholdLG = 992
)

// GetBreakpoint returns the current breakpoint for a given width.
func GetBreakpoint(width float32) Breakpoint {
	switch {
	case width <= ThresholdSM:
		return BreakpointSM
	case width <= ThresholdMD:
		return BreakpointMD
	case width <= ThresholdLG:
		return BreakpointLG
	default:
		return BreakpointXL
	}
}

// BreakpointName returns a human-readable name for the breakpoint.
func BreakpointName(bp Breakpoint) string {
	switch bp {
	case BreakpointSM:
		return "SM (Mobile)"
	case BreakpointMD:
		return "MD (Tablet)"
	case BreakpointLG:
		return "LG (Laptop)"
	case BreakpointXL:
		return "XL (Desktop)"
	default:
		return "Unknown"
	}
}

// ResponsiveCallback is called when the breakpoint changes.
type ResponsiveCallback func(newBreakpoint Breakpoint, size fyne.Size)

// ResponsiveDetector is a resize detector that fires callbacks only when
// the breakpoint changes (not on every resize).
type ResponsiveDetector struct {
	widget.BaseWidget

	// OnBreakpointChange is called when the breakpoint changes.
	OnBreakpointChange ResponsiveCallback

	// currentBreakpoint tracks the current breakpoint
	currentBreakpoint Breakpoint
	lastSize          fyne.Size
	initialized       bool
}

// NewResponsiveDetector creates a new responsive detector.
func NewResponsiveDetector(onBreakpointChange ResponsiveCallback) *ResponsiveDetector {
	rd := &ResponsiveDetector{
		OnBreakpointChange: onBreakpointChange,
		currentBreakpoint:  BreakpointXL, // Default to desktop
	}
	rd.ExtendBaseWidget(rd)
	return rd
}

// CreateRenderer implements fyne.Widget.
func (rd *ResponsiveDetector) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(nil)
	return &responsiveDetectorRenderer{
		detector: rd,
		rect:     rect,
	}
}

// CurrentBreakpoint returns the current breakpoint.
func (rd *ResponsiveDetector) CurrentBreakpoint() Breakpoint {
	return rd.currentBreakpoint
}

// responsiveDetectorRenderer is the renderer for ResponsiveDetector.
type responsiveDetectorRenderer struct {
	detector *ResponsiveDetector
	rect     *canvas.Rectangle
}

// Layout is called when the widget needs to be laid out.
func (r *responsiveDetectorRenderer) Layout(size fyne.Size) {
	DebugLayoutCall("responsiveDetectorRenderer", size)
	r.rect.Resize(size)

	// Check if breakpoint has changed
	newBreakpoint := GetBreakpoint(size.Width)

	if !r.detector.initialized || newBreakpoint != r.detector.currentBreakpoint {
		r.detector.currentBreakpoint = newBreakpoint
		r.detector.lastSize = size
		r.detector.initialized = true

		// Fire the callback if set
		if r.detector.OnBreakpointChange != nil {
			r.detector.OnBreakpointChange(newBreakpoint, size)
		}
	}
}

// MinSize returns the minimum size.
func (r *responsiveDetectorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

// Refresh refreshes the renderer.
func (r *responsiveDetectorRenderer) Refresh() {
	DebugRefreshCall("responsiveDetectorRenderer", r.detector.Size())
	r.rect.Refresh()
}

// Objects returns the canvas objects.
func (r *responsiveDetectorRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect}
}

// Destroy cleans up the renderer.
func (r *responsiveDetectorRenderer) Destroy() {}
