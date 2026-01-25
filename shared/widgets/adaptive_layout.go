// Package widgets provides reusable Fyne widget components.
package widgets

import (
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// AdaptiveLayout provides responsive layout that adapts to screen size.
// It maintains both a desktop layout (using splits) and a mobile layout (using tabs),
// reparenting the same widget instances between them to preserve state.
//
// The breakpoint detection is handled directly in the widget's renderer to avoid
// refresh cascades that occur when using a separate detector widget.
//
// Usage:
//
//	zones := []fyne.CanvasObject{drawingZone, resultsZone, controlsZone, weightsZone}
//	tabLabels := []string{"Draw", "Results", "Config", "Weights"}
//	adaptive := NewAdaptiveLayout(zones, tabLabels)
//	adaptive.SetDesktopLayout(func(zones []fyne.CanvasObject) fyne.CanvasObject {
//	    // Return your HSplit/VSplit layout
//	})
//	content := adaptive.Content()
type AdaptiveLayout struct {
	widget.BaseWidget

	// zones holds the actual content widgets that get reparented
	zones []fyne.CanvasObject

	// tabLabels for mobile tab navigation
	tabLabels []string

	// Desktop layout builder - called to create the desktop arrangement
	desktopLayoutBuilder func(zones []fyne.CanvasObject) fyne.CanvasObject

	// Containers for different modes
	desktopContainer fyne.CanvasObject
	mobileContainer  *container.AppTabs

	// Current state
	currentBreakpoint Breakpoint
	isMobile          bool
	initialized       bool
	lastSize          fyne.Size

	// The main content container (created once, reused)
	contentContainer *fyne.Container

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Re-entrancy guard to prevent infinite loops
	switching atomic.Bool

	// Callbacks
	OnBreakpointChange func(bp Breakpoint)
}

// NewAdaptiveLayout creates a new adaptive layout with the given zones and tab labels.
// The zones are the actual content widgets that will be reparented between layouts.
// tabLabels are the names shown in the mobile tab navigation.
func NewAdaptiveLayout(zones []fyne.CanvasObject, tabLabels []string) *AdaptiveLayout {
	if len(zones) == 0 {
		zones = []fyne.CanvasObject{}
	}
	if len(tabLabels) < len(zones) {
		// Pad with default labels
		for i := len(tabLabels); i < len(zones); i++ {
			tabLabels = append(tabLabels, "Tab "+string(rune('1'+i)))
		}
	}

	a := &AdaptiveLayout{
		zones:             zones,
		tabLabels:         tabLabels,
		currentBreakpoint: BreakpointXL,
		isMobile:          false,
		initialized:       false,
	}
	a.ExtendBaseWidget(a)

	// Create mobile tabs container (but don't populate yet)
	a.buildMobileTabs()

	// Create the main content container ONCE - this is returned by Content()
	// Initially empty, will be populated when SetDesktopLayout is called
	a.contentContainer = container.NewStack()

	return a
}

// SetDesktopLayout sets the function that builds the desktop layout from zones.
// This function receives the zones and should return a container (typically HSplit/VSplit).
func (a *AdaptiveLayout) SetDesktopLayout(builder func(zones []fyne.CanvasObject) fyne.CanvasObject) {
	a.mu.Lock()
	a.desktopLayoutBuilder = builder
	a.mu.Unlock()

	// Build initial desktop layout if builder is set
	if builder != nil {
		// Build the desktop container (without holding the lock during builder call)
		desktopContent := builder(a.zones)

		a.mu.Lock()
		a.desktopContainer = desktopContent
		a.initialized = true

		// Initialize content container with desktop layout
		// Breakpoint detection is handled in the renderer's Layout method
		a.contentContainer.Objects = []fyne.CanvasObject{a.desktopContainer}
		a.mu.Unlock()
	}
}

// buildMobileTabs creates the tab container for mobile layout.
// NOTE: Tabs are created with placeholder content. Actual zones are reparented
// when switchToMobile is called to avoid dual-parenting issues.
func (a *AdaptiveLayout) buildMobileTabs() {
	tabs := make([]*container.TabItem, len(a.zones))
	for i := range a.zones {
		label := a.tabLabels[i]
		// Create tabs with empty placeholder - zones will be reparented on switch
		tabs[i] = container.NewTabItem(label, container.NewMax())
	}

	a.mobileContainer = container.NewAppTabs(tabs...)
	a.mobileContainer.SetTabLocation(container.TabLocationTop)
}

// Content returns the main container that adapts to screen size.
// This returns the same container instance every time (important for Fyne).
func (a *AdaptiveLayout) Content() fyne.CanvasObject {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.contentContainer
}

// onBreakpointChange handles breakpoint transitions (with re-entrancy check).
func (a *AdaptiveLayout) onBreakpointChange(bp Breakpoint, size fyne.Size) {
	// Prevent re-entrant calls that could cause infinite loops
	if !a.switching.CompareAndSwap(false, true) {
		return // Already switching, skip this call
	}
	defer a.switching.Store(false)

	a.handleBreakpointSwitch(bp, size)
}

// handleBreakpointSwitch performs the actual breakpoint switch.
// Caller must ensure proper re-entrancy protection.
func (a *AdaptiveLayout) handleBreakpointSwitch(bp Breakpoint, size fyne.Size) {
	a.mu.Lock()
	if !a.initialized {
		a.mu.Unlock()
		return
	}

	oldBp := a.currentBreakpoint
	a.currentBreakpoint = bp
	callback := a.OnBreakpointChange

	// Determine if we should be in mobile mode
	// SM and MD are considered mobile (< 768px)
	shouldBeMobile := bp == BreakpointSM || bp == BreakpointMD
	needsSwitch := shouldBeMobile != a.isMobile
	a.mu.Unlock()

	// Only switch if needed
	if needsSwitch {
		if shouldBeMobile {
			a.switchToMobileInternal()
		} else {
			a.switchToDesktopInternal()
		}
	}

	// Fire callback if breakpoint changed (outside of any locks)
	if callback != nil && bp != oldBp {
		callback(bp)
	}
}

// switchToMobile reparents zones to the tab container (public API).
func (a *AdaptiveLayout) switchToMobile() {
	if !a.switching.CompareAndSwap(false, true) {
		return
	}
	defer a.switching.Store(false)
	a.switchToMobileInternal()
}

// switchToMobileInternal performs the actual switch without re-entrancy check.
func (a *AdaptiveLayout) switchToMobileInternal() {
	a.mu.Lock()
	if a.isMobile {
		a.mu.Unlock()
		return // Already mobile
	}
	a.isMobile = true

	// Reparent zones to tabs
	for i, zone := range a.zones {
		if i < len(a.mobileContainer.Items) {
			a.mobileContainer.Items[i].Content = container.NewMax(zone)
		}
	}

	// Update content container to show mobile layout
	a.contentContainer.Objects = []fyne.CanvasObject{a.mobileContainer}
	cc := a.contentContainer
	a.mu.Unlock()

	// Refresh the content container (outside lock)
	// Note: This is already called from within fyne.Do context from Layout
	cc.Refresh()
}

// switchToDesktop reparents zones to the split containers (public API).
func (a *AdaptiveLayout) switchToDesktop() {
	if !a.switching.CompareAndSwap(false, true) {
		return
	}
	defer a.switching.Store(false)
	a.switchToDesktopInternal()
}

// switchToDesktopInternal performs the actual switch without re-entrancy check.
func (a *AdaptiveLayout) switchToDesktopInternal() {
	// Get builder without holding lock
	a.mu.RLock()
	if !a.isMobile {
		a.mu.RUnlock()
		return // Already desktop
	}
	builder := a.desktopLayoutBuilder
	zones := a.zones
	a.mu.RUnlock()

	// Build desktop container without holding the lock
	var desktopContent fyne.CanvasObject
	if builder != nil {
		desktopContent = builder(zones)
	}

	// Now update state with lock
	a.mu.Lock()
	a.isMobile = false
	if desktopContent != nil {
		a.desktopContainer = desktopContent
	}

	// Update content container to show desktop layout
	if a.desktopContainer != nil {
		a.contentContainer.Objects = []fyne.CanvasObject{a.desktopContainer}
	}
	cc := a.contentContainer
	a.mu.Unlock()

	// Refresh the content container (outside lock)
	// Note: This is already called from within fyne.Do context from Layout
	cc.Refresh()
}

// IsMobile returns true if currently in mobile layout.
func (a *AdaptiveLayout) IsMobile() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isMobile
}

// CurrentBreakpoint returns the current breakpoint.
func (a *AdaptiveLayout) CurrentBreakpoint() Breakpoint {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentBreakpoint
}

// GetZone returns a zone by index for external access.
func (a *AdaptiveLayout) GetZone(index int) fyne.CanvasObject {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if index < 0 || index >= len(a.zones) {
		return nil
	}
	return a.zones[index]
}

// SetZone replaces a zone at the given index.
func (a *AdaptiveLayout) SetZone(index int, zone fyne.CanvasObject) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if index < 0 || index >= len(a.zones) {
		return
	}

	a.zones[index] = zone

	// Update in current container
	if a.isMobile {
		if index < len(a.mobileContainer.Items) {
			a.mobileContainer.Items[index].Content = container.NewMax(zone)
			a.mobileContainer.Refresh()
		}
	} else if a.desktopLayoutBuilder != nil {
		// Rebuild desktop layout
		a.desktopContainer = a.desktopLayoutBuilder(a.zones)
		a.contentContainer.Objects = []fyne.CanvasObject{a.desktopContainer}
		a.contentContainer.Refresh()
	}
}

// SelectTab selects a tab in mobile mode by index.
func (a *AdaptiveLayout) SelectTab(index int) {
	a.mu.RLock()
	isMobile := a.isMobile
	tabs := a.mobileContainer
	a.mu.RUnlock()

	if isMobile && tabs != nil && index >= 0 && index < len(tabs.Items) {
		tabs.SelectIndex(index)
	}
}

// CreateRenderer implements fyne.Widget.
func (a *AdaptiveLayout) CreateRenderer() fyne.WidgetRenderer {
	a.mu.RLock()
	content := a.contentContainer
	a.mu.RUnlock()

	return &adaptiveLayoutRenderer{
		layout:  a,
		content: content,
	}
}

// adaptiveLayoutRenderer is the renderer for AdaptiveLayout.
type adaptiveLayoutRenderer struct {
	layout      *AdaptiveLayout
	content     *fyne.Container
	lastSize    fyne.Size
	initialized bool
}

func (r *adaptiveLayoutRenderer) Layout(size fyne.Size) {
	if r.content != nil {
		r.content.Resize(size)
	}

	// Handle breakpoint detection directly in the renderer
	// Only process valid sizes
	if size.Width <= 0 || size.Height <= 0 {
		return
	}

	newBreakpoint := GetBreakpoint(size.Width)

	// On first layout, just initialize without triggering any callbacks
	if !r.initialized {
		r.layout.mu.Lock()
		r.layout.currentBreakpoint = newBreakpoint
		r.layout.lastSize = size
		r.layout.mu.Unlock()
		r.lastSize = size
		r.initialized = true
		return
	}

	// Skip if size hasn't changed significantly (prevents oscillation)
	if r.lastSize == size {
		return
	}
	r.lastSize = size

	// Skip if already switching to prevent recursion
	if r.layout.switching.Load() {
		return
	}

	// Check if we need to switch modes (only on breakpoint change)
	r.layout.mu.RLock()
	currentBp := r.layout.currentBreakpoint
	r.layout.mu.RUnlock()

	// Only trigger mode switch when breakpoint actually changes
	if newBreakpoint != currentBp {
		// Set switching flag BEFORE triggering the switch
		if !r.layout.switching.CompareAndSwap(false, true) {
			return // Already switching
		}
		// Schedule the switch for after current layout completes
		fyne.Do(func() {
			defer r.layout.switching.Store(false)
			r.layout.handleBreakpointSwitch(newBreakpoint, size)
		})
	}
}

func (r *adaptiveLayoutRenderer) MinSize() fyne.Size {
	// Return a reasonable minimum size
	return fyne.NewSize(320, 240)
}

func (r *adaptiveLayoutRenderer) Refresh() {
	if r.content != nil {
		r.content.Refresh()
	}
}

func (r *adaptiveLayoutRenderer) Objects() []fyne.CanvasObject {
	if r.content != nil {
		return []fyne.CanvasObject{r.content}
	}
	return nil
}

func (r *adaptiveLayoutRenderer) Destroy() {}

// ResponsiveSplit is a helper that adjusts HSplit/VSplit offset based on breakpoint.
type ResponsiveSplit struct {
	Split *container.Split

	// Offsets for different breakpoints
	SmOffset float64
	MdOffset float64
	LgOffset float64
	XlOffset float64
}

// NewResponsiveSplit creates a split with breakpoint-aware offsets.
func NewResponsiveSplit(split *container.Split, smOffset, mdOffset, lgOffset, xlOffset float64) *ResponsiveSplit {
	return &ResponsiveSplit{
		Split:    split,
		SmOffset: smOffset,
		MdOffset: mdOffset,
		LgOffset: lgOffset,
		XlOffset: xlOffset,
	}
}

// ApplyBreakpoint sets the split offset for the given breakpoint.
func (rs *ResponsiveSplit) ApplyBreakpoint(bp Breakpoint) {
	if rs.Split == nil {
		return
	}

	var offset float64
	switch bp {
	case BreakpointSM:
		offset = rs.SmOffset
	case BreakpointMD:
		offset = rs.MdOffset
	case BreakpointLG:
		offset = rs.LgOffset
	default:
		offset = rs.XlOffset
	}

	rs.Split.SetOffset(offset)
}

// GridWrapLayout is a layout that wraps items to the next row when they exceed width.
// This provides responsive grid behavior similar to CSS flexbox with wrap.
type GridWrapLayout struct {
	MinItemWidth float32
	ItemHeight   float32
	RowSpacing   float32
	ColSpacing   float32
}

// NewGridWrapLayout creates a new grid wrap layout with specified item dimensions.
func NewGridWrapLayout(minItemWidth, itemHeight, rowSpacing, colSpacing float32) *GridWrapLayout {
	return &GridWrapLayout{
		MinItemWidth: minItemWidth,
		ItemHeight:   itemHeight,
		RowSpacing:   rowSpacing,
		ColSpacing:   colSpacing,
	}
}

// MinSize returns the minimum size needed to display all items.
func (g *GridWrapLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(g.MinItemWidth, g.ItemHeight)
	}

	// Minimum: 1 column, all rows
	visibleCount := 0
	for _, obj := range objects {
		if obj.Visible() {
			visibleCount++
		}
	}

	if visibleCount == 0 {
		return fyne.NewSize(g.MinItemWidth, g.ItemHeight)
	}

	height := float32(visibleCount)*g.ItemHeight + float32(visibleCount-1)*g.RowSpacing
	return fyne.NewSize(g.MinItemWidth, height)
}

// Layout positions all objects in a wrapping grid.
func (g *GridWrapLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	// Calculate how many columns fit
	availableWidth := size.Width
	cols := int((availableWidth + g.ColSpacing) / (g.MinItemWidth + g.ColSpacing))
	if cols < 1 {
		cols = 1
	}

	// Calculate actual item width to fill space evenly
	itemWidth := (availableWidth - float32(cols-1)*g.ColSpacing) / float32(cols)

	x := float32(0)
	y := float32(0)
	col := 0

	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}

		obj.Resize(fyne.NewSize(itemWidth, g.ItemHeight))
		obj.Move(fyne.NewPos(x, y))

		col++
		if col >= cols {
			col = 0
			x = 0
			y += g.ItemHeight + g.RowSpacing
		} else {
			x += itemWidth + g.ColSpacing
		}
	}
}
