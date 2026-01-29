// Package gui provides Fyne-based GUI components for MNIST visualization.
// This file provides package-local font rendering wrappers that delegate to shared/utils.
package gui

import (
	"image"
	"image/color"

	"fecim-lattice-tools/shared/utils"
)

// drawSimpleText draws text using a simple bitmap font.
func drawSimpleText(img *image.RGBA, text string, x, y int, c color.Color) {
	utils.DrawSimpleText(img, text, x, y, c)
}

// drawScaledText draws text with a scaling factor.
func drawScaledText(img *image.RGBA, text string, x, y int, scale int, c color.Color) {
	utils.DrawScaledText(img, text, x, y, scale, c)
}

// drawScaledChar draws a single character with scaling.
func drawScaledChar(img *image.RGBA, ch rune, x, y int, scale int, c color.Color) {
	utils.DrawScaledChar(img, ch, x, y, scale, c)
}
