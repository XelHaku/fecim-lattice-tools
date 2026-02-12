// Package gui provides UI components for the circuits module.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"fecim-lattice-tools/shared/utils"
)

// drawRect fills a rectangular region in an RGBA image with the specified color.
// It performs boundary checks to ensure all pixels are within image bounds.
func drawRect(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	utils.DrawRect(img, x, y, rectW, rectH, c)
}

// drawRectBorder draws only the border (outline) of a rectangular region.
// It performs boundary checks to ensure all pixels are within image bounds.
func drawRectBorder(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
	utils.DrawRectBorder(img, x, y, rectW, rectH, c)
}

// sleep pauses execution for the specified number of milliseconds.
// Used for animation timing. Returns true if interrupted by stop signal.
// Uses time.NewTimer instead of time.After to avoid memory leaks when interrupted.
func (ca *CircuitsApp) sleep(milliseconds int) bool {
	timer := time.NewTimer(time.Duration(milliseconds) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ca.stopChan:
		return true // Interrupted
	case <-timer.C:
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
	utils.DrawRoundedRect(img, x, y, rectW, rectH, cornerRadius, c)
}

// drawGradientRect draws a rectangle with vertical gradient from topColor to bottomColor.
func drawGradientRect(img *image.RGBA, x, y, rectW, rectH int, topColor, bottomColor color.RGBA) {
	utils.DrawGradientRect(img, x, y, rectW, rectH, topColor, bottomColor)
}

// drawGlowCircle draws a circle with a soft glow effect.
func drawGlowCircle(img *image.RGBA, cx, cy, radius int, centerColor, glowColor color.RGBA) {
	utils.DrawGlowCircle(img, cx, cy, radius, centerColor, glowColor)
}

// levelToColor converts a FeCIM level to a color using blue-white-red gradient.
// Level 0: Blue (low conductance) - full brightness
// Mid level: White (neutral)
// Max level: Red (high conductance) - full brightness
func levelToColor(level, maxLevel int) color.RGBA {
	return utils.LevelToColor(level, maxLevel)
}

// drawThickLine draws a line with specified thickness.
func drawThickLine(img *image.RGBA, x1, y1, x2, y2, thickness int, c color.Color) {
	utils.DrawThickLine(img, x1, y1, x2, y2, thickness, c)
}

// drawSimpleText draws text using a simple bitmap font.
func drawSimpleText(img *image.RGBA, text string, x, y int, c color.Color) {
	utils.DrawSimpleText(img, text, x, y, c)
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
	utils.DrawSimpleText(img, text, textX, textY, style.TextColor)
}

// drawDACColumn draws a DAC box above a column with voltage display
// Colors indicate voltage zone: blue/green = read (safe), orange/red = write (programming)
func drawDACColumn(img *image.RGBA, x, y, w, h int, voltage float64, label string, highlighted, dimmed bool) {
	style := DACStyle(highlighted, dimmed)

	// Voltage zone color coding (based on typical FeCIM thresholds)
	// Read zone: 0-0.5V (safe sensing, below coercive voltage)
	// Caution zone: 0.5-0.8V (approaching Vc)
	// Write zone: >0.8V (above Vc, will program cell)
	if !highlighted && !dimmed {
		if voltage < 0.05 {
			// No voltage - dim gray
			style.TopColor = color.RGBA{60, 60, 70, 255}
			style.BottomColor = color.RGBA{40, 40, 50, 255}
			style.BorderColor = color.RGBA{80, 80, 90, 255}
		} else if voltage < 0.5 {
			// Read zone - blue/cyan (safe)
			style.TopColor = color.RGBA{60, 140, 200, 255}
			style.BottomColor = color.RGBA{40, 100, 160, 255}
			style.BorderColor = color.RGBA{100, 180, 255, 255}
		} else if voltage < 0.8 {
			// Caution zone - yellow/amber
			style.TopColor = color.RGBA{200, 180, 60, 255}
			style.BottomColor = color.RGBA{160, 140, 40, 255}
			style.BorderColor = color.RGBA{255, 220, 100, 255}
		} else {
			// Write zone - orange/red (programming)
			style.TopColor = color.RGBA{220, 100, 60, 255}
			style.BottomColor = color.RGBA{180, 70, 40, 255}
			style.BorderColor = color.RGBA{255, 140, 100, 255}
		}
	}

	vText := fmt.Sprintf("%.2fV", voltage)
	drawPeripheralBox(img, x, y, w, h, style, vText)

	// Draw label above
	if label != "" {
		labelColor := color.RGBA{140, 170, 255, 255}
		if dimmed {
			labelColor = color.RGBA{100, 120, 180, 200}
		}
		utils.DrawSimpleText(img, label, x+w/2-len(label)*3, y-12, labelColor)
	}
}

// drawTIAADCRow draws TIA+ADC boxes to the right of a row with current/level display
func drawTIAADCRow(img *image.RGBA, x, y, tiaW, adcW, h int, current float64, level int, label string, highlighted, dimmed bool, tia interface{ Convert(float64) float64 }, _ interface{ Convert(float64) int }) {
	// TIA box - show output voltage (with explicit units).
	tiaStyle := TIAStyle(highlighted, dimmed)
	vout := 0.0
	if tia != nil {
		vout = tia.Convert(current * 1e-6) // current is in uA; TIA expects A
	}
	var tiaText string
	absV := math.Abs(vout)
	switch {
	case absV < 1e-3:
		tiaText = fmt.Sprintf("%.0fmV", vout*1e3)
	case absV < 10:
		tiaText = fmt.Sprintf("%.2fV", vout)
	default:
		tiaText = fmt.Sprintf("%.1fV", vout)
	}
	drawPeripheralBox(img, x, y, tiaW, h, tiaStyle, tiaText)

	// ADC box (to the right of TIA) - show output with explicit code/LSB units.
	adcStyle := ADCStyle(highlighted, dimmed)
	adcText := fmt.Sprintf("%dLSB", level)
	drawPeripheralBox(img, x+tiaW+2, y, adcW, h, adcStyle, adcText)

	// Draw label to the left (row number)
	if label != "" {
		labelColor := color.RGBA{255, 190, 140, 255}
		if dimmed {
			labelColor = color.RGBA{180, 140, 110, 200}
		}
		// Draw row label to the right of ADC box
		utils.DrawSimpleText(img, label, x+tiaW+adcW+6, y+h/2-3, labelColor)
	}
}

// drawPeripheralLabels draws the "DAC" and "TIA+ADC" labels at fixed positions
func drawPeripheralLabels(img *image.RGBA, dacLabelX, dacLabelY, adcLabelX, adcLabelY int, showDAC, showADC bool) {
	if showDAC {
		utils.DrawSimpleText(img, "DAC", dacLabelX, dacLabelY, color.RGBA{170, 140, 220, 255})
	}
	if showADC {
		utils.DrawSimpleText(img, "TIA", adcLabelX, adcLabelY, color.RGBA{220, 180, 100, 255})
		utils.DrawSimpleText(img, "ADC", adcLabelX+30, adcLabelY, color.RGBA{130, 210, 170, 255})
	}
}
