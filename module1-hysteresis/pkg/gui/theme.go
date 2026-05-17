//go:build legacy_fyne

// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
package gui

import (
	sharedtheme "fecim-lattice-tools/shared/theme"
	"fyne.io/fyne/v2"
)

// Exported colors - accessible from widgets subpackage
var (
	ColorPrimary    = sharedtheme.ColorPrimary
	ColorSecondary  = sharedtheme.ColorSecondary
	ColorAccent     = sharedtheme.ColorAccent
	ColorWarning    = sharedtheme.ColorWarning
	ColorBackground = sharedtheme.ColorBackground
	ColorGrid       = sharedtheme.ColorGrid
	ColorAxis       = sharedtheme.ColorAxis
	ColorPositive   = sharedtheme.ColorPositive
	ColorNegative   = sharedtheme.ColorNegative
)

// Legacy unexported aliases for backward compatibility within gui package
var (
	colorPrimary    = ColorPrimary
	colorSecondary  = ColorSecondary
	colorAccent     = ColorAccent
	colorWarning    = ColorWarning
	colorBackground = ColorBackground
	colorGrid       = ColorGrid
	colorAxis       = ColorAxis
	colorPositive   = ColorPositive
	colorNegative   = ColorNegative
)

// ============================================================
// Fixed Width Layout
// ============================================================

// fixedWidthLayout is a custom layout that enforces a fixed width
type fixedWidthLayout struct {
	width float32
}

func (l *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if o.Visible() {
			minH = fyne.Max(minH, o.MinSize().Height)
		}
	}
	return fyne.NewSize(l.width, minH)
}

func (l *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(l.width, size.Height))
		o.Move(fyne.NewPos(0, 0))
	}
}

// fixedMinWidthLayout enforces a minimum width but allows expansion
type fixedMinWidthLayout struct {
	minWidth float32
}

func (l *fixedMinWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if o.Visible() {
			minH = fyne.Max(minH, o.MinSize().Height)
		}
	}
	return fyne.NewSize(l.minWidth, minH)
}

func (l *fixedMinWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		// Only use the given width, but respect child's MinSize for height
		// This prevents vertical stretching
		childMinSize := o.MinSize()
		o.Resize(fyne.NewSize(size.Width, childMinSize.Height))
		o.Move(fyne.NewPos(0, 0))
	}
}
