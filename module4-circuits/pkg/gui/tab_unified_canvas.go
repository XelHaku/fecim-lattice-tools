//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the tappable canvas widget for the unified view.
package gui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
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
		layout := buildUnifiedArrayLayout(int(size.Width), int(size.Height), rows, cols, arch, zoom)
		cellSize = layout.cellSize
		offsetX = layout.offsetX
		offsetY = layout.offsetY
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
