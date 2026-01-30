// Package widgets provides reusable Fyne widget components.
package widgets

import (
	"fyne.io/fyne/v2"
)

// ResponsiveGridLayout adapts both column count AND item height based on breakpoints.
// This provides a more responsive experience than GridWrapLayout by:
// - Adjusting card heights for better readability on small screens
// - Capping columns per breakpoint to prevent too many columns on narrow screens
// - Capping max item width to prevent oversized cards on 4K displays
// - Centering the grid when items hit max width
type ResponsiveGridLayout struct {
	MinItemWidth float32 // Minimum card width (default: 280)
	MaxItemWidth float32 // Maximum card width (default: 450) - prevents huge cards on 4K

	// Height per breakpoint - taller cards on smaller screens for readability
	ItemHeightSM float32 // <=576px: default 180px (single column, taller cards)
	ItemHeightMD float32 // 577-768px: default 160px
	ItemHeightLG float32 // 769-992px: default 150px
	ItemHeightXL float32 // >992px: default 140px

	RowSpacing float32
	ColSpacing float32

	cache LayoutCache
}

// NewResponsiveGridLayout creates a new responsive grid layout with sensible defaults.
func NewResponsiveGridLayout() *ResponsiveGridLayout {
	return &ResponsiveGridLayout{
		MinItemWidth: 300,
		MaxItemWidth: 550,  // Moderate max width
		ItemHeightSM: 200,  // UI-002: Increased from 180 to 200
		ItemHeightMD: 200,  // UI-002: Increased from 180 to 200
		ItemHeightLG: 220,  // UI-002: Increased from 200 to 220
		ItemHeightXL: 230,  // UI-002: Increased from 210 to 230
		RowSpacing:   24,   // UI-002: Increased from 20 to 24px
		ColSpacing:   24,   // UI-002: Increased from 20 to 24px
	}
}

// getBreakpointHeight returns the item height for the current breakpoint.
func (g *ResponsiveGridLayout) getBreakpointHeight(width float32) float32 {
	bp := GetBreakpoint(width)
	switch bp {
	case BreakpointSM:
		return g.ItemHeightSM
	case BreakpointMD:
		return g.ItemHeightMD
	case BreakpointLG:
		return g.ItemHeightLG
	default:
		return g.ItemHeightXL
	}
}

// getMaxColumns returns the maximum columns allowed for the current breakpoint.
// M13: <1024px=1-col, 1024-1600px=2-col, >1600px=3-col
func (g *ResponsiveGridLayout) getMaxColumns(width float32) int {
	bp := GetBreakpoint(width)
	switch bp {
	case BreakpointSM, BreakpointMD, BreakpointLG:
		return 1 // Single column on mobile/tablet/small laptop (<1024px)
	case BreakpointXL:
		return 2 // Two columns on desktop (1024-1600px)
	case BreakpointXXL:
		return 3 // Three columns on ultra-wide (>1600px)
	default:
		return 2 // Default to 2 columns
	}
}

// calculateLayout computes the grid parameters for a given width.
func (g *ResponsiveGridLayout) calculateLayout(width float32) (cols int, itemWidth, itemHeight, offsetX float32) {
	// Get breakpoint-specific values
	itemHeight = g.getBreakpointHeight(width)
	maxCols := g.getMaxColumns(width)

	// Calculate how many columns fit based on minimum width
	cols = int((width + g.ColSpacing) / (g.MinItemWidth + g.ColSpacing))
	if cols < 1 {
		cols = 1
	}
	// Cap columns by breakpoint
	if cols > maxCols {
		cols = maxCols
	}

	// Calculate actual item width to fill space evenly
	itemWidth = (width - float32(cols-1)*g.ColSpacing) / float32(cols)

	// Cap item width at maximum to prevent huge cards on 4K
	if itemWidth > g.MaxItemWidth {
		itemWidth = g.MaxItemWidth
		// Calculate offset to center the grid
		totalGridWidth := float32(cols)*itemWidth + float32(cols-1)*g.ColSpacing
		offsetX = (width - totalGridWidth) / 2
	} else {
		offsetX = 0
	}

	return cols, itemWidth, itemHeight, offsetX
}

// MinSize returns the minimum size needed to display all items.
func (g *ResponsiveGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(g.MinItemWidth, g.ItemHeightSM)
	}

	// Count visible objects
	visibleCount := 0
	for _, obj := range objects {
		if obj.Visible() {
			visibleCount++
		}
	}

	if visibleCount == 0 {
		return fyne.NewSize(g.MinItemWidth, g.ItemHeightSM)
	}

	// Minimum size: single column, all rows with SM height (tallest)
	height := float32(visibleCount)*g.ItemHeightSM + float32(visibleCount-1)*g.RowSpacing
	return fyne.NewSize(g.MinItemWidth, height)
}

// Layout positions all objects in a responsive wrapping grid.
func (g *ResponsiveGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	// Skip if size hasn't changed (prevents oscillation)
	if !g.cache.ShouldLayout(size) {
		return
	}

	cols, itemWidth, itemHeight, offsetX := g.calculateLayout(size.Width)

	x := offsetX
	y := float32(0)
	col := 0

	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}

		obj.Resize(fyne.NewSize(itemWidth, itemHeight))
		obj.Move(fyne.NewPos(x, y))

		col++
		if col >= cols {
			col = 0
			x = offsetX
			y += itemHeight + g.RowSpacing
		} else {
			x += itemWidth + g.ColSpacing
		}
	}

	g.cache.MarkLayout(size)
}
