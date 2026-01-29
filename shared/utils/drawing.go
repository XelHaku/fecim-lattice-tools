// Package utils provides shared drawing utilities for canvas rendering.
// This file contains drawing primitives used across multiple modules.
package utils

import (
	"image"
	"image/color"
)

// DrawRect fills a rectangular region in an RGBA image with the specified color.
// It performs boundary checks to ensure all pixels are within image bounds.
func DrawRect(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	for py := y; py < y+rectH; py++ {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.Set(px, py, c)
			}
		}
	}
}

// DrawRectBorder draws only the border (outline) of a rectangular region.
// It performs boundary checks to ensure all pixels are within image bounds.
func DrawRectBorder(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	// Top edge
	if y >= 0 && y < maxY {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < maxX {
				img.Set(px, y, c)
			}
		}
	}
	// Bottom edge
	bottomY := y + rectH - 1
	if bottomY >= 0 && bottomY < maxY {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < maxX {
				img.Set(px, bottomY, c)
			}
		}
	}
	// Left edge
	if x >= 0 && x < maxX {
		for py := y; py < y+rectH; py++ {
			if py >= 0 && py < maxY {
				img.Set(x, py, c)
			}
		}
	}
	// Right edge
	rightX := x + rectW - 1
	if rightX >= 0 && rightX < maxX {
		for py := y; py < y+rectH; py++ {
			if py >= 0 && py < maxY {
				img.Set(rightX, py, c)
			}
		}
	}
}

// DrawRoundedRect draws a filled rectangle with rounded corners.
// cornerRadius specifies the radius of the corner rounding.
func DrawRoundedRect(img *image.RGBA, x, y, rectW, rectH, cornerRadius int, c color.Color) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	for py := y; py < y+rectH; py++ {
		for px := x; px < x+rectW; px++ {
			if px < 0 || px >= maxX || py < 0 || py >= maxY {
				continue
			}

			// Check if pixel is in corner region
			inCorner := false
			var cx, cy int

			// Top-left corner
			if px < x+cornerRadius && py < y+cornerRadius {
				cx, cy = x+cornerRadius, y+cornerRadius
				inCorner = true
			}
			// Top-right corner
			if px >= x+rectW-cornerRadius && py < y+cornerRadius {
				cx, cy = x+rectW-cornerRadius-1, y+cornerRadius
				inCorner = true
			}
			// Bottom-left corner
			if px < x+cornerRadius && py >= y+rectH-cornerRadius {
				cx, cy = x+cornerRadius, y+rectH-cornerRadius-1
				inCorner = true
			}
			// Bottom-right corner
			if px >= x+rectW-cornerRadius && py >= y+rectH-cornerRadius {
				cx, cy = x+rectW-cornerRadius-1, y+rectH-cornerRadius-1
				inCorner = true
			}

			if inCorner {
				dx := px - cx
				dy := py - cy
				if dx*dx+dy*dy > cornerRadius*cornerRadius {
					continue // Outside rounded corner
				}
			}

			img.Set(px, py, c)
		}
	}
}

// DrawGradientRect draws a rectangle with vertical gradient from topColor to bottomColor.
func DrawGradientRect(img *image.RGBA, x, y, rectW, rectH int, topColor, bottomColor color.RGBA) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	for py := y; py < y+rectH; py++ {
		if py < 0 || py >= maxY {
			continue
		}
		// Calculate interpolation factor
		t := float64(py-y) / float64(rectH)

		// Interpolate colors
		r := uint8(float64(topColor.R)*(1-t) + float64(bottomColor.R)*t)
		g := uint8(float64(topColor.G)*(1-t) + float64(bottomColor.G)*t)
		b := uint8(float64(topColor.B)*(1-t) + float64(bottomColor.B)*t)
		a := uint8(float64(topColor.A)*(1-t) + float64(bottomColor.A)*t)

		rowColor := color.RGBA{r, g, b, a}

		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < maxX {
				img.Set(px, py, rowColor)
			}
		}
	}
}

// DrawGlowCircle draws a circle with a soft glow effect.
func DrawGlowCircle(img *image.RGBA, cx, cy, radius int, centerColor, glowColor color.RGBA) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	glowRadius := radius + 3

	for dy := -glowRadius; dy <= glowRadius; dy++ {
		for dx := -glowRadius; dx <= glowRadius; dx++ {
			px, py := cx+dx, cy+dy
			if px < 0 || px >= maxX || py < 0 || py >= maxY {
				continue
			}

			dist := dx*dx + dy*dy

			if dist <= radius*radius {
				// Inner solid circle
				img.Set(px, py, centerColor)
			} else if dist <= glowRadius*glowRadius {
				// Glow region - fade out
				t := float64(dist-radius*radius) / float64(glowRadius*glowRadius-radius*radius)
				alpha := uint8(float64(glowColor.A) * (1 - t))
				if alpha > 10 {
					// Blend with existing pixel
					existing := img.RGBAAt(px, py)
					blendAlpha := float64(alpha) / 255.0
					r := uint8(float64(glowColor.R)*blendAlpha + float64(existing.R)*(1-blendAlpha))
					g := uint8(float64(glowColor.G)*blendAlpha + float64(existing.G)*(1-blendAlpha))
					b := uint8(float64(glowColor.B)*blendAlpha + float64(existing.B)*(1-blendAlpha))
					img.Set(px, py, color.RGBA{r, g, b, 255})
				}
			}
		}
	}
}

// LevelToColor converts a FeCIM level to a color using blue-white-red gradient.
// Level 0: Blue (low conductance) - full brightness
// Mid level: White (neutral)
// Max level: Red (high conductance) - full brightness
func LevelToColor(level, maxLevel int) color.RGBA {
	if maxLevel <= 1 {
		return color.RGBA{255, 255, 255, 255} // White for single level
	}

	// Normalize level to 0.0 - 1.0 range
	t := float64(level) / float64(maxLevel-1)

	var r, g, b uint8

	if t < 0.5 {
		// Blue to White: level 0 -> mid
		// t=0: Blue (0, 0, 255), t=0.5: White (255, 255, 255)
		s := t * 2.0        // Scale to 0-1 for this half
		r = uint8(s * 255)  // 0 -> 255
		g = uint8(s * 255)  // 0 -> 255
		b = 255             // Always 255 (full blue)
	} else {
		// White to Red: mid -> max level
		// t=0.5: White (255, 255, 255), t=1: Red (255, 0, 0)
		s := (t - 0.5) * 2.0   // Scale to 0-1 for this half
		r = 255                // Always 255 (full red)
		g = uint8(255 - s*255) // 255 -> 0
		b = uint8(255 - s*255) // 255 -> 0
	}

	return color.RGBA{r, g, b, 255}
}

// DrawThickLine draws a line with specified thickness.
func DrawThickLine(img *image.RGBA, x1, y1, x2, y2, thickness int, c color.Color) {
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

	// Simple Bresenham-like line drawing with thickness
	dx := x2 - x1
	dy := y2 - y1

	steps := max(abs(dx), abs(dy))
	if steps == 0 {
		steps = 1
	}

	xInc := float64(dx) / float64(steps)
	yInc := float64(dy) / float64(steps)

	halfT := thickness / 2

	x := float64(x1)
	y := float64(y1)

	for i := 0; i <= steps; i++ {
		px := int(x)
		py := int(y)

		// Draw thick point
		for ty := -halfT; ty <= halfT; ty++ {
			for tx := -halfT; tx <= halfT; tx++ {
				ppx, ppy := px+tx, py+ty
				if ppx >= 0 && ppx < maxX && ppy >= 0 && ppy < maxY {
					img.Set(ppx, ppy, c)
				}
			}
		}

		x += xInc
		y += yInc
	}
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
