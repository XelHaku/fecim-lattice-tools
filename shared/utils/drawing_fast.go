// Package utils provides shared drawing utilities for canvas rendering.
// This file contains optimized drawing primitives that use direct pixel manipulation
// instead of the slower img.Set() interface method.
package utils

import (
	"image"
	"image/color"
)

// DrawRectFast fills a rectangular region using direct pixel manipulation.
// This is significantly faster than DrawRect for large rectangles.
// The color must be an RGBA color for this optimization.
func DrawRectFast(img *image.RGBA, x, y, rectW, rectH int, c color.RGBA) {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	// Clip to image bounds
	x0 := x
	y0 := y
	x1 := x + rectW
	y1 := y + rectH

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > imgW {
		x1 = imgW
	}
	if y1 > imgH {
		y1 = imgH
	}

	if x0 >= x1 || y0 >= y1 {
		return // Nothing to draw
	}

	stride := img.Stride
	pix := img.Pix

	// Pre-compute the color bytes
	r, g, b, a := c.R, c.G, c.B, c.A

	for py := y0; py < y1; py++ {
		rowStart := py * stride
		for px := x0; px < x1; px++ {
			offset := rowStart + px*4
			pix[offset] = r
			pix[offset+1] = g
			pix[offset+2] = b
			pix[offset+3] = a
		}
	}
}

// DrawRectBorderFast draws only the border using direct pixel manipulation.
func DrawRectBorderFast(img *image.RGBA, x, y, rectW, rectH int, c color.RGBA) {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	stride := img.Stride
	pix := img.Pix

	r, g, b, a := c.R, c.G, c.B, c.A

	// Helper to set a pixel with bounds check
	setPixel := func(px, py int) {
		if px >= 0 && px < imgW && py >= 0 && py < imgH {
			offset := py*stride + px*4
			pix[offset] = r
			pix[offset+1] = g
			pix[offset+2] = b
			pix[offset+3] = a
		}
	}

	// Top edge
	for px := x; px < x+rectW; px++ {
		setPixel(px, y)
	}
	// Bottom edge
	bottomY := y + rectH - 1
	for px := x; px < x+rectW; px++ {
		setPixel(px, bottomY)
	}
	// Left edge (excluding corners already drawn)
	for py := y + 1; py < y+rectH-1; py++ {
		setPixel(x, py)
	}
	// Right edge (excluding corners already drawn)
	rightX := x + rectW - 1
	for py := y + 1; py < y+rectH-1; py++ {
		setPixel(rightX, py)
	}
}

// FillImageFast fills the entire image with a single color.
// Much faster than iterating with img.Set().
func FillImageFast(img *image.RGBA, c color.RGBA) {
	pix := img.Pix
	r, g, b, a := c.R, c.G, c.B, c.A

	// Fill first row
	bounds := img.Bounds()
	width := bounds.Dx()
	stride := img.Stride

	// Fill first row
	for x := 0; x < width; x++ {
		offset := x * 4
		pix[offset] = r
		pix[offset+1] = g
		pix[offset+2] = b
		pix[offset+3] = a
	}

	// Copy first row to remaining rows
	height := bounds.Dy()
	rowBytes := width * 4
	for y := 1; y < height; y++ {
		copy(pix[y*stride:y*stride+rowBytes], pix[0:rowBytes])
	}
}

// DrawHLineFast draws a horizontal line using direct pixel manipulation.
func DrawHLineFast(img *image.RGBA, x, y, length int, c color.RGBA) {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	if y < 0 || y >= imgH {
		return
	}

	x0 := x
	x1 := x + length
	if x0 < 0 {
		x0 = 0
	}
	if x1 > imgW {
		x1 = imgW
	}
	if x0 >= x1 {
		return
	}

	stride := img.Stride
	pix := img.Pix
	r, g, b, a := c.R, c.G, c.B, c.A

	rowStart := y * stride
	for px := x0; px < x1; px++ {
		offset := rowStart + px*4
		pix[offset] = r
		pix[offset+1] = g
		pix[offset+2] = b
		pix[offset+3] = a
	}
}

// DrawVLineFast draws a vertical line using direct pixel manipulation.
func DrawVLineFast(img *image.RGBA, x, y, length int, c color.RGBA) {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	if x < 0 || x >= imgW {
		return
	}

	y0 := y
	y1 := y + length
	if y0 < 0 {
		y0 = 0
	}
	if y1 > imgH {
		y1 = imgH
	}
	if y0 >= y1 {
		return
	}

	stride := img.Stride
	pix := img.Pix
	r, g, b, a := c.R, c.G, c.B, c.A
	xOffset := x * 4

	for py := y0; py < y1; py++ {
		offset := py*stride + xOffset
		pix[offset] = r
		pix[offset+1] = g
		pix[offset+2] = b
		pix[offset+3] = a
	}
}

// DrawGridFast draws a grid pattern efficiently.
// Useful for crossbar visualization and digit canvas grids.
func DrawGridFast(img *image.RGBA, cellW, cellH, rows, cols int, c color.RGBA) {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	// Draw vertical lines
	for col := 0; col <= cols; col++ {
		x := col * cellW
		if x >= 0 && x < imgW {
			DrawVLineFast(img, x, 0, imgH, c)
		}
	}

	// Draw horizontal lines
	for row := 0; row <= rows; row++ {
		y := row * cellH
		if y >= 0 && y < imgH {
			DrawHLineFast(img, 0, y, imgW, c)
		}
	}
}

// BlendPixelFast blends a color onto an existing pixel with alpha blending.
func BlendPixelFast(img *image.RGBA, x, y int, c color.RGBA) {
	bounds := img.Bounds()
	if x < 0 || x >= bounds.Dx() || y < 0 || y >= bounds.Dy() {
		return
	}

	stride := img.Stride
	pix := img.Pix
	offset := y*stride + x*4

	if c.A == 255 {
		// Fully opaque - just overwrite
		pix[offset] = c.R
		pix[offset+1] = c.G
		pix[offset+2] = c.B
		pix[offset+3] = 255
		return
	}

	if c.A == 0 {
		return // Fully transparent - no change
	}

	// Alpha blending
	alpha := float64(c.A) / 255.0
	invAlpha := 1.0 - alpha

	pix[offset] = uint8(float64(c.R)*alpha + float64(pix[offset])*invAlpha)
	pix[offset+1] = uint8(float64(c.G)*alpha + float64(pix[offset+1])*invAlpha)
	pix[offset+2] = uint8(float64(c.B)*alpha + float64(pix[offset+2])*invAlpha)
	pix[offset+3] = 255
}

// DrawGradientRectFast draws a vertical gradient rectangle efficiently.
func DrawGradientRectFast(img *image.RGBA, x, y, rectW, rectH int, topColor, bottomColor color.RGBA) {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	// Clip to bounds
	x0 := x
	y0 := y
	x1 := x + rectW
	y1 := y + rectH

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > imgW {
		x1 = imgW
	}
	if y1 > imgH {
		y1 = imgH
	}

	if x0 >= x1 || y0 >= y1 {
		return
	}

	stride := img.Stride
	pix := img.Pix
	height := float64(rectH)

	for py := y0; py < y1; py++ {
		t := float64(py-y) / height
		invT := 1.0 - t

		r := uint8(float64(topColor.R)*invT + float64(bottomColor.R)*t)
		g := uint8(float64(topColor.G)*invT + float64(bottomColor.G)*t)
		b := uint8(float64(topColor.B)*invT + float64(bottomColor.B)*t)
		a := uint8(float64(topColor.A)*invT + float64(bottomColor.A)*t)

		rowStart := py * stride
		for px := x0; px < x1; px++ {
			offset := rowStart + px*4
			pix[offset] = r
			pix[offset+1] = g
			pix[offset+2] = b
			pix[offset+3] = a
		}
	}
}
