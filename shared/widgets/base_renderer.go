// Package widgets provides shared widget utilities for Fyne GUI development.
package widgets

import (
	"fyne.io/fyne/v2"
)

// ConstrainedWidget provides a base for widgets that should not grow beyond their MinSize.
// Embed this in your widget to get size-constrained behavior.
type ConstrainedWidget struct {
	minSize fyne.Size
}

// NewConstrainedWidget creates a new constrained widget with the given minimum size.
func NewConstrainedWidget(minSize fyne.Size) *ConstrainedWidget {
	return &ConstrainedWidget{minSize: minSize}
}

// SetMinSize sets the minimum (and maximum) size for this widget.
func (c *ConstrainedWidget) SetMinSize(size fyne.Size) {
	c.minSize = size
}

// GetMinSize returns the configured minimum size.
func (c *ConstrainedWidget) GetMinSize() fyne.Size {
	return c.minSize
}

// ConstrainedSize returns the minimum size, clamped to not exceed minSize.
// Use this in Layout() to ensure the widget doesn't grow.
func (c *ConstrainedWidget) ConstrainedSize(allocated fyne.Size) fyne.Size {
	return ConstrainSize(allocated, c.minSize)
}

// BaseRendererHelper provides common patterns for widget renderers.
// Use these helper methods to ensure consistent layout behavior.
type BaseRendererHelper struct {
	widgetName string
}

// NewBaseRendererHelper creates a new renderer helper with the widget name for debugging.
func NewBaseRendererHelper(widgetName string) *BaseRendererHelper {
	return &BaseRendererHelper{widgetName: widgetName}
}

// LogLayout logs a Layout() call if debug mode is enabled.
// Returns true if there might be a layout loop (suspiciously high call count).
func (h *BaseRendererHelper) LogLayout(size fyne.Size) bool {
	return DebugLayoutCall(h.widgetName, size)
}

// LogRefresh logs a Refresh() call if debug mode is enabled.
func (h *BaseRendererHelper) LogRefresh(widgetSize fyne.Size) {
	DebugRefreshCall(h.widgetName, widgetSize)
}

// LogMinSize logs a MinSize() call if debug mode is enabled.
func (h *BaseRendererHelper) LogMinSize(minSize fyne.Size) {
	DebugMinSizeCall(h.widgetName, minSize)
}

// LayoutHelpers provides common layout calculations.
type LayoutHelpers struct{}

// CenterObject centers an object within a container size.
func (LayoutHelpers) CenterObject(obj fyne.CanvasObject, containerSize fyne.Size) {
	objSize := obj.Size()
	if objSize.Width == 0 || objSize.Height == 0 {
		objSize = obj.MinSize()
	}
	pos := CenterInSize(objSize, containerSize)
	obj.Move(pos)
}

// ResizeAndPosition resizes an object and moves it to the specified position.
func (LayoutHelpers) ResizeAndPosition(obj fyne.CanvasObject, size fyne.Size, pos fyne.Position) {
	obj.Resize(size)
	obj.Move(pos)
}

// FillContainer resizes an object to fill the container and positions it at origin.
func (LayoutHelpers) FillContainer(obj fyne.CanvasObject, containerSize fyne.Size) {
	obj.Resize(containerSize)
	obj.Move(fyne.NewPos(0, 0))
}

// SafeLayoutPattern provides the recommended layout/refresh pattern.
// Use this to avoid Layout->Refresh cycles.
//
// Example usage in a renderer:
//
//	func (r *myRenderer) Layout(size fyne.Size) {
//	    r.layoutWithSize(size) // Do the actual layout
//	}
//
//	func (r *myRenderer) Refresh() {
//	    // Update any text/colors first
//	    r.updateContent()
//	    // Then re-layout with current size
//	    r.layoutWithSize(r.widget.Size())
//	}
//
//	func (r *myRenderer) layoutWithSize(size fyne.Size) {
//	    // Position all objects based on size parameter
//	    // NEVER call Refresh() from here
//	}
//
// Key rules:
// 1. Layout() uses the SIZE PARAMETER, not widget.Size()
// 2. Refresh() can use widget.Size() to get current allocated size
// 3. layoutWithSize() never calls Refresh() on self or parent
// 4. Only call child.Refresh() when updating child content, not layout
type SafeLayoutPattern struct{}

// VerifyLayoutPattern checks if a renderer follows the safe layout pattern.
// This is a documentation helper, not an actual runtime check.
func (SafeLayoutPattern) VerifyLayoutPattern() string {
	return `
Safe Layout Pattern Checklist:
[ ] Layout(size) uses the size parameter, not widget.Size()
[ ] Refresh() does not call Layout() directly
[ ] layoutWithSize() is a shared helper for both Layout() and Refresh()
[ ] No Refresh() calls to self or parent from within layoutWithSize()
[ ] Child Refresh() calls only update content, not trigger re-layout
[ ] MinSize() returns a constant or cached value, not computed from layout
`
}

// LayoutCache tracks the last layout size to avoid redundant layout operations.
// Use this in renderers to prevent layout cascade bugs.
type LayoutCache struct {
	LastSize fyne.Size
	HasLayout bool
}

// ShouldLayout returns true if layout is needed (size changed or first layout).
// Also validates that size is positive - returns false for invalid sizes.
func (c *LayoutCache) ShouldLayout(size fyne.Size) bool {
	// Guard against invalid sizes (negative or zero) - critical for Wayland stability
	if size.Width <= 0 || size.Height <= 0 {
		return false
	}
	// Skip if size hasn't changed
	if c.HasLayout && size.Width == c.LastSize.Width && size.Height == c.LastSize.Height {
		return false
	}
	return true
}

// MarkLayout marks the layout as done with the given size.
// Call this after successfully performing layout.
func (c *LayoutCache) MarkLayout(size fyne.Size) {
	c.LastSize = size
	c.HasLayout = true
}

// ValidateSize returns true if the size is valid for layout (positive dimensions).
// Use this at the start of Layout() to early-exit on invalid sizes.
func ValidateSize(size fyne.Size) bool {
	return size.Width > 0 && size.Height > 0
}

// SafeResize resizes a canvas object only if the size is valid.
// Returns true if resize was performed.
func SafeResize(obj fyne.CanvasObject, size fyne.Size) bool {
	if size.Width <= 0 || size.Height <= 0 {
		return false
	}
	obj.Resize(size)
	return true
}
