//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains drawing utilities for timing diagrams.
package gui

import (
	"image"
	"image/color"
)

// drawThickHorizontalLine draws a horizontal line from x1 to x2 at y with specified thickness.
// This is used by timing diagram functions to render continuous waveforms instead of sparse dots.
func drawThickHorizontalLine(img *image.RGBA, x1, x2, y int, thickness int, c color.Color) {
	h := img.Bounds().Dy()
	halfT := thickness / 2
	for px := x1; px <= x2; px++ {
		for dy := -halfT; dy <= halfT; dy++ {
			py := y + dy
			if py >= 0 && py < h {
				img.Set(px, py, c)
			}
		}
	}
}
