// Package gui provides UI components for the circuits module.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"time"
)

// drawRect fills a rectangular region in an RGBA image with the specified color.
// It performs boundary checks to ensure all pixels are within image bounds.
func drawRect(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	for py := y; py < y+rectH; py++ {
		for px := x; px < x+rectW; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.Set(px, py, c)
			}
		}
	}
}

// drawRectBorder draws only the border (outline) of a rectangular region.
// It performs boundary checks to ensure all pixels are within image bounds.
func drawRectBorder(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
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

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sleep pauses execution for the specified number of milliseconds.
// Used for animation timing. Returns true if interrupted by stop signal.
func (ca *CircuitsApp) sleep(milliseconds int) bool {
	select {
	case <-ca.stopChan:
		return true // Interrupted
	case <-time.After(time.Duration(milliseconds) * time.Millisecond):
		return false // Normal completion
	}
}

// shouldStop returns true if Stop() has been called
func (ca *CircuitsApp) shouldStop() bool {
	select {
	case <-ca.stopChan:
		return true
	default:
		return false
	}
}

// drawRoundedRect draws a filled rectangle with rounded corners.
// cornerRadius specifies the radius of the corner rounding.
func drawRoundedRect(img *image.RGBA, x, y, rectW, rectH, cornerRadius int, c color.Color) {
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

// drawGradientRect draws a rectangle with vertical gradient from topColor to bottomColor.
func drawGradientRect(img *image.RGBA, x, y, rectW, rectH int, topColor, bottomColor color.RGBA) {
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

// drawGlowCircle draws a circle with a soft glow effect.
func drawGlowCircle(img *image.RGBA, cx, cy, radius int, centerColor, glowColor color.RGBA) {
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

// levelToColor converts a FeCIM level to a color using blue-white-red gradient.
// Level 0: Blue (low conductance) - full brightness
// Mid level: White (neutral)
// Max level: Red (high conductance) - full brightness
func levelToColor(level, maxLevel int) color.RGBA {
	if maxLevel <= 1 {
		return color.RGBA{255, 255, 255, 255} // White for single level
	}

	// Normalize level to 0.0 - 1.0 range
	t := float64(level) / float64(maxLevel-1)

	var r, g, b uint8

	if t < 0.5 {
		// Blue to White: level 0 -> mid
		// t=0: Blue (0, 0, 255), t=0.5: White (255, 255, 255)
		s := t * 2.0 // Scale to 0-1 for this half
		r = uint8(s * 255)       // 0 -> 255
		g = uint8(s * 255)       // 0 -> 255
		b = 255                  // Always 255 (full blue)
	} else {
		// White to Red: mid -> max level
		// t=0.5: White (255, 255, 255), t=1: Red (255, 0, 0)
		s := (t - 0.5) * 2.0 // Scale to 0-1 for this half
		r = 255              // Always 255 (full red)
		g = uint8(255 - s*255) // 255 -> 0
		b = uint8(255 - s*255) // 255 -> 0
	}

	return color.RGBA{r, g, b, 255}
}

// drawThickLine draws a line with specified thickness.
func drawThickLine(img *image.RGBA, x1, y1, x2, y2, thickness int, c color.Color) {
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

// ============================================================================
// PERIPHERAL COMPONENT DRAWING HELPERS
// Single source of truth for DAC, TIA, ADC visualization across all modes
// ============================================================================

// PeripheralStyle defines the visual style for a peripheral component
type PeripheralStyle struct {
	TopColor    color.RGBA
	BottomColor color.RGBA
	BorderColor color.RGBA
	TextColor   color.RGBA
	Highlighted bool
	Dimmed      bool
}

// DACStyle returns the style for a DAC box
func DACStyle(highlighted, dimmed bool) PeripheralStyle {
	style := PeripheralStyle{
		TopColor:    color.RGBA{130, 90, 190, 255},
		BottomColor: color.RGBA{80, 50, 140, 255},
		BorderColor: color.RGBA{170, 140, 220, 255},
		TextColor:   color.RGBA{255, 255, 255, 255},
		Highlighted: highlighted,
		Dimmed:      dimmed,
	}
	if highlighted {
		style.TopColor = color.RGBA{255, 255, 120, 255}
		style.BottomColor = color.RGBA{220, 200, 80, 255}
		style.BorderColor = color.RGBA{255, 255, 180, 255}
	}
	if dimmed {
		style.TopColor = color.RGBA{80, 60, 120, 200}
		style.BottomColor = color.RGBA{50, 35, 90, 200}
		style.BorderColor = color.RGBA{100, 80, 140, 200}
		style.TextColor = color.RGBA{180, 180, 180, 255}
	}
	return style
}

// TIAStyle returns the style for a TIA box
func TIAStyle(highlighted, dimmed bool) PeripheralStyle {
	style := PeripheralStyle{
		TopColor:    color.RGBA{190, 140, 70, 255},
		BottomColor: color.RGBA{140, 100, 40, 255},
		BorderColor: color.RGBA{220, 180, 100, 255},
		TextColor:   color.RGBA{255, 255, 255, 255},
		Highlighted: highlighted,
		Dimmed:      dimmed,
	}
	if highlighted {
		style.TopColor = color.RGBA{255, 220, 100, 255}
		style.BottomColor = color.RGBA{220, 180, 60, 255}
		style.BorderColor = color.RGBA{255, 240, 150, 255}
	}
	if dimmed {
		style.TopColor = color.RGBA{120, 90, 50, 200}
		style.BottomColor = color.RGBA{90, 65, 30, 200}
		style.BorderColor = color.RGBA{140, 110, 70, 200}
		style.TextColor = color.RGBA{180, 180, 180, 255}
	}
	return style
}

// ADCStyle returns the style for an ADC box
func ADCStyle(highlighted, dimmed bool) PeripheralStyle {
	style := PeripheralStyle{
		TopColor:    color.RGBA{70, 170, 130, 255},
		BottomColor: color.RGBA{40, 120, 90, 255},
		BorderColor: color.RGBA{130, 210, 170, 255},
		TextColor:   color.RGBA{255, 255, 255, 255},
		Highlighted: highlighted,
		Dimmed:      dimmed,
	}
	if highlighted {
		style.TopColor = color.RGBA{100, 255, 170, 255}
		style.BottomColor = color.RGBA{70, 200, 130, 255}
		style.BorderColor = color.RGBA{150, 255, 200, 255}
	}
	if dimmed {
		style.TopColor = color.RGBA{50, 110, 85, 200}
		style.BottomColor = color.RGBA{30, 80, 60, 200}
		style.BorderColor = color.RGBA{85, 140, 115, 200}
		style.TextColor = color.RGBA{180, 180, 180, 255}
	}
	return style
}

// drawPeripheralBox draws a peripheral component box with label
func drawPeripheralBox(img *image.RGBA, x, y, w, h int, style PeripheralStyle, text string) {
	drawGradientRect(img, x, y, w, h, style.TopColor, style.BottomColor)
	drawRectBorder(img, x, y, w, h, style.BorderColor)

	// Center text in box
	textX := x + w/2 - len(text)*3
	textY := y + h/2 - 3
	drawSimpleText(img, text, textX, textY, style.TextColor)
}

// drawDACColumn draws a DAC box above a column with voltage display
func drawDACColumn(img *image.RGBA, x, y, w, h int, voltage float64, label string, highlighted, dimmed bool) {
	style := DACStyle(highlighted, dimmed)
	vText := fmt.Sprintf("%.2f", voltage)
	drawPeripheralBox(img, x, y, w, h, style, vText)

	// Draw label above
	if label != "" {
		labelColor := color.RGBA{140, 170, 255, 255}
		if dimmed {
			labelColor = color.RGBA{100, 120, 180, 200}
		}
		drawSimpleText(img, label, x+w/2-len(label)*3, y-12, labelColor)
	}
}

// drawTIAADCRow draws TIA+ADC boxes to the right of a row with current/level display
func drawTIAADCRow(img *image.RGBA, x, y, tiaW, adcW, h int, current float64, level int, label string, highlighted, dimmed bool, tia interface{ Convert(float64) float64 }, adc interface{ Convert(float64) int }) {
	// TIA box - shows both input current and output voltage
	tiaStyle := TIAStyle(highlighted, dimmed)
	tiaV := 0.0
	if tia != nil {
		tiaV = tia.Convert(current * 1e-6) // uA to A
	}
	// Format: "12µA→.12" (compact: current in µA, arrow, voltage without leading 0)
	var tiaText string
	if current < 10 {
		tiaText = fmt.Sprintf("%.1fµ→%.2f", current, tiaV)
	} else if current < 100 {
		tiaText = fmt.Sprintf("%.0fµ→%.2f", current, tiaV)
	} else {
		tiaText = fmt.Sprintf("%.0fµ→%.2f", current, tiaV)
	}
	drawPeripheralBox(img, x, y, tiaW, h, tiaStyle, tiaText)

	// ADC box (to the right of TIA) - shows decoded state (0-29 for 30 analog states)
	adcStyle := ADCStyle(highlighted, dimmed)
	adcLevel := level
	if adc != nil && tia != nil {
		adcLevel = adc.Convert(tiaV)
	}
	adcText := fmt.Sprintf("%d", adcLevel)
	drawPeripheralBox(img, x+tiaW+2, y, adcW, h, adcStyle, adcText)

	// Draw label to the right
	if label != "" {
		labelColor := color.RGBA{255, 190, 140, 255}
		if dimmed {
			labelColor = color.RGBA{180, 140, 110, 200}
		}
		drawSimpleText(img, label, x+tiaW+adcW+8, y+h/2-3, labelColor)
	}
}

// drawPeripheralLabels draws the "DAC" and "TIA+ADC" labels at fixed positions
func drawPeripheralLabels(img *image.RGBA, dacLabelX, dacLabelY, adcLabelX, adcLabelY int, showDAC, showADC bool) {
	if showDAC {
		drawSimpleText(img, "DAC", dacLabelX, dacLabelY, color.RGBA{170, 140, 220, 255})
	}
	if showADC {
		drawSimpleText(img, "TIA", adcLabelX, adcLabelY, color.RGBA{220, 180, 100, 255})
		drawSimpleText(img, "ADC", adcLabelX+30, adcLabelY, color.RGBA{130, 210, 170, 255})
	}
}
