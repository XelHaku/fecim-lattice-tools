//go:build legacy_fyne

package gui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Breakpoint constants define responsive breakpoints (width in px)
// Mobile: < 600, Tablet: 600-900, Desktop: 900-1200, Wide: > 1200
const (
	BreakpointTablet  float32 = 600  // Tablet >= 600
	BreakpointDesktop float32 = 900  // Desktop >= 900
	BreakpointWide    float32 = 1200 // Wide > 1200
)

// LayoutMode represents the current responsive layout configuration
type LayoutMode int

const (
	LayoutMobile LayoutMode = iota
	LayoutTablet
	LayoutDesktop
	LayoutWide
)

// LayoutManager manages responsive layout for the documentation viewer
type LayoutManager struct {
	currentMode    LayoutMode
	sidebarVisible bool
	tocVisible     bool

	// UI components to manage
	sidebar   fyne.CanvasObject
	content   fyne.CanvasObject
	toc       fyne.CanvasObject
	topBar    fyne.CanvasObject
	bottomBar fyne.CanvasObject // Mobile only

	// Container that holds the responsive layout
	container *fyne.Container
	root      *fyne.Container

	// Callbacks
	onSidebarToggle func(visible bool)
	onTocToggle     func(visible bool)

	mu sync.Mutex
}

// NewLayoutManager creates a new responsive layout manager
func NewLayoutManager() *LayoutManager {
	return &LayoutManager{
		currentMode:    LayoutDesktop,
		sidebarVisible: true,
		tocVisible:     true,
	}
}

// SetComponents sets the UI components to manage
func (lm *LayoutManager) SetComponents(sidebar, content, toc, topBar fyne.CanvasObject) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.sidebar = sidebar
	lm.content = content
	lm.toc = toc
	lm.topBar = topBar
}

// SetBottomBar sets the mobile-only bottom action bar
func (lm *LayoutManager) SetBottomBar(bottomBar fyne.CanvasObject) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.bottomBar = bottomBar
}

// SetSidebarToggleCallback sets the callback for sidebar toggle events
func (lm *LayoutManager) SetSidebarToggleCallback(callback func(visible bool)) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.onSidebarToggle = callback
}

// SetTocToggleCallback sets the callback for ToC toggle events
func (lm *LayoutManager) SetTocToggleCallback(callback func(visible bool)) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.onTocToggle = callback
}

// OnResize handles window resize events
func (lm *LayoutManager) OnResize(size fyne.Size) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	width := size.Width

	var newMode LayoutMode
	switch {
	case width > BreakpointWide:
		newMode = LayoutWide
	case width >= BreakpointDesktop:
		newMode = LayoutDesktop
	case width >= BreakpointTablet:
		newMode = LayoutTablet
	default:
		newMode = LayoutMobile
	}

	if newMode != lm.currentMode {
		lm.currentMode = newMode
		lm.applyLayoutMode()
	}
}

// GetCurrentMode returns the current layout mode
func (lm *LayoutManager) GetCurrentMode() LayoutMode {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	return lm.currentMode
}

// IsSidebarVisible returns whether the sidebar is currently visible
func (lm *LayoutManager) IsSidebarVisible() bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	return lm.sidebarVisible
}

// IsTocVisible returns whether the ToC is currently visible
func (lm *LayoutManager) IsTocVisible() bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	return lm.tocVisible
}

// ToggleSidebar manually toggles sidebar visibility
func (lm *LayoutManager) ToggleSidebar() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.sidebarVisible = !lm.sidebarVisible
	lm.applyLayoutMode()

	if lm.onSidebarToggle != nil {
		lm.onSidebarToggle(lm.sidebarVisible)
	}
}

// ToggleToc manually toggles ToC visibility
func (lm *LayoutManager) ToggleToc() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.tocVisible = !lm.tocVisible
	lm.applyLayoutMode()

	if lm.onTocToggle != nil {
		lm.onTocToggle(lm.tocVisible)
	}
}

// SetTocVisible explicitly controls ToC visibility and reapplies layout.
func (lm *LayoutManager) SetTocVisible(visible bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.tocVisible == visible {
		return
	}

	lm.tocVisible = visible
	lm.applyLayoutMode()

	if lm.onTocToggle != nil {
		lm.onTocToggle(lm.tocVisible)
	}
}

// BuildLayout creates the layout container based on current mode
func (lm *LayoutManager) BuildLayout() *fyne.Container {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.applyLayoutMode()
	if lm.root == nil {
		lm.root = container.New(&resizeWatcherLayout{onResize: lm.OnResize}, lm.container)
	}
	return lm.root
}

// resizeWatcherLayout forwards size changes to the layout manager.
type resizeWatcherLayout struct {
	onResize func(fyne.Size)
}

func (r *resizeWatcherLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if r.onResize != nil {
		r.onResize(size)
	}
	if len(objects) == 0 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
}

func (r *resizeWatcherLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}

// applyLayoutMode rebuilds the layout based on current mode
// Must be called with lm.mu held
func (lm *LayoutManager) applyLayoutMode() {
	var newContainer *fyne.Container

	switch lm.currentMode {
	case LayoutMobile:
		newContainer = lm.buildMobileLayout()
	case LayoutTablet:
		newContainer = lm.buildTabletLayout()
	case LayoutDesktop:
		newContainer = lm.buildDesktopLayout()
	case LayoutWide:
		newContainer = lm.buildWideLayout()
	}

	if lm.container != nil {
		lm.container.Objects = newContainer.Objects
		lm.container.Layout = newContainer.Layout
		lm.container.Refresh()
	} else {
		lm.container = newContainer
	}
}

// buildMobileLayout creates the mobile layout (< 600px)
// Full-width content, sidebar and ToC as overlays
func (lm *LayoutManager) buildMobileLayout() *fyne.Container {
	var mainContent fyne.CanvasObject = lm.content

	// Overlay sidebar if visible
	if lm.sidebarVisible && lm.sidebar != nil {
		// Create a semi-transparent backdrop
		backdrop := widget.NewCard("", "", lm.sidebar)
		mainContent = container.NewStack(
			lm.content,
			backdrop,
		)
	}

	// Build with top and bottom bars
	if lm.bottomBar != nil {
		return container.NewBorder(lm.topBar, lm.bottomBar, nil, nil, mainContent)
	}

	return container.NewBorder(lm.topBar, nil, nil, nil, mainContent)
}

// buildTabletLayout creates the tablet layout (600-900px)
// Fixed sidebar (28%) + content (72%), no ToC to avoid overflow
func (lm *LayoutManager) buildTabletLayout() *fyne.Container {
	var mainContent fyne.CanvasObject = lm.content

	if lm.sidebar != nil && lm.sidebarVisible {
		split := container.NewHSplit(lm.sidebar, lm.content)
		split.SetOffset(0.28) // 28% sidebar, 72% content
		mainContent = split
	}

	// Skip ToC at tablet width to avoid cramped layout (< 900px)
	// ToC is only shown in desktop (>= 900px) and wide (> 1200px) modes

	return container.NewBorder(lm.topBar, nil, nil, nil, mainContent)
}

// buildDesktopLayout creates the desktop layout (900-1200px)
// Sidebar (25%) + wider content (75%)
func (lm *LayoutManager) buildDesktopLayout() *fyne.Container {
	var mainContent fyne.CanvasObject = lm.content

	if lm.sidebar != nil && lm.sidebarVisible {
		split := container.NewHSplit(lm.sidebar, lm.content)
		split.SetOffset(0.25) // 25% sidebar, 75% content
		mainContent = split
	}

	// Add ToC if visible and available
	if lm.toc != nil && lm.tocVisible {
		finalSplit := container.NewHSplit(mainContent, lm.toc)
		finalSplit.SetOffset(0.80) // 80% main, 20% ToC
		mainContent = finalSplit
	}

	return container.NewBorder(lm.topBar, nil, nil, nil, mainContent)
}

// buildWideLayout creates the wide layout (> 1200px)
// Sidebar (20%) + content (60%) + ToC (20%)
func (lm *LayoutManager) buildWideLayout() *fyne.Container {
	var mainContent fyne.CanvasObject = lm.content

	// Add sidebar if visible
	if lm.sidebar != nil && lm.sidebarVisible {
		leftSplit := container.NewHSplit(lm.sidebar, lm.content)
		leftSplit.SetOffset(0.25) // 25% sidebar, 75% right side
		mainContent = leftSplit
	}

	// Add ToC if visible and available
	if lm.toc != nil && lm.tocVisible {
		finalSplit := container.NewHSplit(mainContent, lm.toc)

		// Calculate offset based on whether sidebar is visible
		if lm.sidebarVisible {
			finalSplit.SetOffset(0.80) // 80% left (sidebar + content), 20% ToC
		} else {
			finalSplit.SetOffset(0.80) // 80% content, 20% ToC
		}

		mainContent = finalSplit
	}

	return container.NewBorder(lm.topBar, nil, nil, nil, mainContent)
}

// GetModeString returns a human-readable string for the current layout mode
func (lm *LayoutManager) GetModeString() string {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	switch lm.currentMode {
	case LayoutMobile:
		return "Mobile"
	case LayoutTablet:
		return "Tablet"
	case LayoutDesktop:
		return "Desktop"
	case LayoutWide:
		return "Wide"
	default:
		return "Unknown"
	}
}

// GetBreakpointForWidth returns the layout mode for a given width
func GetBreakpointForWidth(width float32) LayoutMode {
	switch {
	case width > BreakpointWide:
		return LayoutWide
	case width >= BreakpointDesktop:
		return LayoutDesktop
	case width >= BreakpointTablet:
		return LayoutTablet
	default:
		return LayoutMobile
	}
}
