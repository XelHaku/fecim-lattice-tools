// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the tappable canvas widget for the unified view.
package gui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ============================================================================
// UNIFIED TAPPABLE CANVAS
// ============================================================================

// UnifiedTappableCanvas is a canvas.Raster that responds to taps for the unified view
type UnifiedTappableCanvas struct {
	widget.BaseWidget
	raster *canvas.Raster
	onTap  func(row, col int)
	ca     *CircuitsApp
}

func NewUnifiedTappableCanvas(ca *CircuitsApp, drawFunc func(w, h int) image.Image, onTap func(row, col int)) *UnifiedTappableCanvas {
	t := &UnifiedTappableCanvas{
		raster: canvas.NewRaster(drawFunc),
		onTap:  onTap,
		ca:     ca,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *UnifiedTappableCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.raster)
}

func (t *UnifiedTappableCanvas) SetMinSize(size fyne.Size) {
	t.raster.SetMinSize(size)
}

func (t *UnifiedTappableCanvas) Refresh() {
	t.raster.Refresh()
}

func (t *UnifiedTappableCanvas) Tapped(e *fyne.PointEvent) {
	size := t.raster.Size()

	t.ca.mu.RLock()
	rows := t.ca.arrayRows
	cols := t.ca.arrayCols
	arch := t.ca.architecture
	zoom := t.ca.zoomLevel
	// Use stored offsets from the drawing function for precise click detection
	cellSize := t.ca.sharedArrayCellSize
	offsetX := t.ca.sharedArrayOffsetX
	offsetY := t.ca.sharedArrayOffsetY
	t.ca.mu.RUnlock()

	if zoom == 0 {
		zoom = 1.0
	}

	// If stored values not set yet, calculate them (fallback)
	if cellSize == 0 {
		w := int(size.Width)
		h := int(size.Height)

		// MUST match drawUnifiedArray margins exactly
		topMargin := 65
		rightMargin := 130
		bottomMargin := 30
		leftMargin := 30

		is1T1R := arch == sharedwidgets.Architecture1T1R
		is2T1R := arch == sharedwidgets.Architecture2T1R
		if is1T1R || is2T1R {
			leftMargin = 55
		}
		if is2T1R {
			bottomMargin = 55
		}

		availableW := w - leftMargin - rightMargin
		availableH := h - topMargin - bottomMargin

		// Scale max/min cell size based on array dimensions AND zoom (must match drawUnifiedArray)
		maxCellSize := int(float64(70) * zoom)
		minCellSize := int(float64(18) * zoom)
		if cols > 32 || rows > 32 {
			maxCellSize = int(float64(30) * zoom)
			minCellSize = int(float64(8) * zoom)
		} else if cols > 16 || rows > 16 {
			maxCellSize = int(float64(40) * zoom)
			minCellSize = int(float64(12) * zoom)
		}

		cellW := availableW / cols
		cellH := availableH / rows
		cellSize = min(cellW, cellH)

		// Apply cell size limits (scaled by zoom)
		if cellSize > maxCellSize {
			cellSize = maxCellSize
		}
		if cellSize < minCellSize {
			cellSize = minCellSize
		}

		gridW := cols * cellSize
		gridH := rows * cellSize
		offsetX = leftMargin + (availableW-gridW)/2
		offsetY = topMargin + (availableH-gridH)/2
	}

	col := (int(e.Position.X) - offsetX) / cellSize
	row := (int(e.Position.Y) - offsetY) / cellSize

	if row >= 0 && row < rows && col >= 0 && col < cols {
		t.onTap(row, col)
	}
}

func (t *UnifiedTappableCanvas) TappedSecondary(*fyne.PointEvent) {}

func (t *UnifiedTappableCanvas) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
