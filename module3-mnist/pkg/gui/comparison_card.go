// Package gui provides Fyne-based GUI components for MNIST visualization.
// comparison_card.go implements P1.2: Enhanced FP vs CIM Comparison Card
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Note: math is used for glow circle calculations

// ComparisonResult holds the results of FP vs CIM comparison.
type ComparisonResult struct {
	FPPrediction    int
	FPConfidence    float64
	FPProbabilities []float64

	CIMPrediction    int
	CIMConfidence    float64
	CIMProbabilities []float64

	Match           bool
	ConfidenceDelta float64
	EnergyFeCIM     float64 // nanojoules
	EnergyGPU       float64 // nanojoules
	EnergyRatio     float64 // GPU/FeCIM
}

// ComparisonCard provides enhanced FP vs CIM comparison visualization.
// This is the hero widget showing why FeCIM's accuracy-energy tradeoff matters.
type ComparisonCard struct {
	widget.BaseWidget

	mu     sync.RWMutex
	result *ComparisonResult

	// Visual components
	titleLabel  *widget.Label
	statusLabel *widget.Label
	raster      *canvas.Raster
}

// NewComparisonCard creates a new comparison card widget.
func NewComparisonCard() *ComparisonCard {
	cc := &ComparisonCard{}
	cc.titleLabel = widget.NewLabelWithStyle("FP vs CIM Comparison", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	cc.statusLabel = widget.NewLabel("Draw a digit to compare predictions")
	cc.ExtendBaseWidget(cc)
	return cc
}

// SetResult updates the comparison with new inference results.
func (cc *ComparisonCard) SetResult(result *ComparisonResult) {
	cc.mu.Lock()
	cc.result = result
	cc.mu.Unlock()

	// Update status
	if result != nil {
		if result.Match {
			cc.statusLabel.SetText(fmt.Sprintf("CIM predicts: %d | Validated by FP | %.0fx energy improvement",
				result.CIMPrediction, result.EnergyRatio))
		} else {
			cc.statusLabel.SetText(fmt.Sprintf("CIM predicts: %d | FP disagrees: %d | Quantization effects visible",
				result.CIMPrediction, result.FPPrediction))
		}
	}

	fyne.Do(func() {
		cc.Refresh()
	})
}

// Clear resets the card to idle state.
func (cc *ComparisonCard) Clear() {
	cc.mu.Lock()
	cc.result = nil
	cc.mu.Unlock()
	cc.statusLabel.SetText("Draw a digit to compare predictions")
	fyne.Do(func() {
		cc.Refresh()
	})
}

// MinSize returns the minimum size for the widget.
func (cc *ComparisonCard) MinSize() fyne.Size {
	return fyne.NewSize(550, 480) // Redesigned with hero layout
}

// CreateRenderer implements fyne.Widget.
func (cc *ComparisonCard) CreateRenderer() fyne.WidgetRenderer {
	cc.raster = canvas.NewRaster(cc.generateImage)

	content := container.NewBorder(
		container.NewVBox(
			cc.titleLabel,
			widget.NewSeparator(),
		),
		cc.statusLabel,
		nil, nil,
		container.NewMax(cc.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage creates the comparison visualization.
func (cc *ComparisonCard) generateImage(w, h int) image.Image {
	if w < 10 {
		w = 550
	}
	if h < 10 {
		h = 420
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	cc.mu.RLock()
	result := cc.result
	cc.mu.RUnlock()

	// Background - Deep dark blue with subtle gradient
	for y := 0; y < h; y++ {
		gradFactor := float64(y) / float64(h)
		r := uint8(12 + gradFactor*8)
		g := uint8(15 + gradFactor*10)
		b := uint8(25 + gradFactor*15)
		bgColor := color.RGBA{r, g, b, 255}
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if result == nil {
		// Idle state - centered message
		drawScaledText(img, "Ready - Draw a digit", w/2-105, h/2-20, 2, color.RGBA{120, 130, 150, 255})
		drawSimpleText(img, "or click Random to load from MNIST", w/2-120, h/2+10, color.RGBA{80, 90, 110, 255})
		return img
	}

	// Color palette - Bold and distinct
	matchColor := color.RGBA{50, 255, 150, 255}     // Vibrant green
	mismatchColor := color.RGBA{255, 140, 50, 255}  // Bright orange
	fpColor := color.RGBA{80, 150, 255, 255}        // Azure blue
	cimColor := color.RGBA{0, 220, 200, 255}        // Cyan
	whiteColor := color.RGBA{245, 250, 255, 255}
	dimColor := color.RGBA{120, 130, 150, 255}
	accentColor := color.RGBA{255, 200, 100, 255}   // Gold

	// Layout sections (height percentages)
	topSectionH := int(float64(h) * 0.18)      // 18% - Status badge and predictions
	heroSectionH := int(float64(h) * 0.38)     // 38% - Giant digit
	chartSectionH := int(float64(h) * 0.38)    // 38% - Probability chart
	footerH := h - topSectionH - heroSectionH - chartSectionH // 6% - Energy info

	padding := 20
	currentY := padding

	// ========== TOP SECTION: Match Badge & Predictions ==========
	topY := currentY

	// Centered badge
	badgeW := 140
	badgeH := 32
	badgeX := (w - badgeW) / 2
	badgeY := topY

	var badgeColor color.RGBA
	var badgeText string
	if result.Match {
		badgeColor = matchColor
		badgeText = "MATCH"
	} else {
		badgeColor = mismatchColor
		badgeText = "DIFFERS"
	}

	// Badge background with glow
	for bx := badgeX; bx < badgeX+badgeW; bx++ {
		for by := badgeY; by < badgeY+badgeH; by++ {
			img.Set(bx, by, color.RGBA{30, 35, 50, 255})
		}
	}
	// Badge border - 2px thick
	for bx := badgeX; bx < badgeX+badgeW; bx++ {
		img.Set(bx, badgeY, badgeColor)
		img.Set(bx, badgeY+1, badgeColor)
		img.Set(bx, badgeY+badgeH-1, badgeColor)
		img.Set(bx, badgeY+badgeH-2, badgeColor)
	}
	for by := badgeY; by < badgeY+badgeH; by++ {
		img.Set(badgeX, by, badgeColor)
		img.Set(badgeX+1, by, badgeColor)
		img.Set(badgeX+badgeW-1, by, badgeColor)
		img.Set(badgeX+badgeW-2, by, badgeColor)
	}
	// Badge text
	textW := len(badgeText) * 7 * 2
	badgeTextX := badgeX + (badgeW-textW)/2
	drawScaledText(img, badgeText, badgeTextX, badgeY+10, 2, badgeColor)

	// FP and CIM predictions side-by-side below badge
	predY := badgeY + badgeH + 15
	fpText := fmt.Sprintf("FP: %d (%.0f%%)", result.FPPrediction, result.FPConfidence*100)
	cimText := fmt.Sprintf("CIM: %d (%.0f%%)", result.CIMPrediction, result.CIMConfidence*100)

	// FP on left
	fpX := w/4 - len(fpText)*7/2
	drawSimpleText(img, fpText, fpX, predY, fpColor)

	// CIM on right
	cimX := 3*w/4 - len(cimText)*7/2
	drawSimpleText(img, cimText, cimX, predY, cimColor)

	currentY = topY + topSectionH

	// Subtle divider
	for x := padding; x < w-padding; x++ {
		img.Set(x, currentY, color.RGBA{40, 50, 70, 255})
	}

	// ========== HERO SECTION: Giant Prediction Digit ==========
	heroY := currentY + 10

	// HUGE digit scale - 12x for maximum impact
	digitScale := 12
	digitHeight := 7 * digitScale
	digitWidth := 5 * digitScale
	digitX := (w - digitWidth) / 2
	digitY := heroY + (heroSectionH-digitHeight-30)/2

	digitText := fmt.Sprintf("%d", result.CIMPrediction)
	if result.CIMPrediction < 0 {
		digitText = "?"
	}

	// Digit color based on match
	var digitColor color.RGBA
	if result.Match {
		digitColor = matchColor
	} else {
		digitColor = mismatchColor
	}

	// Draw hero digit
	cc.drawScaledDigit(img, digitX, digitY, digitText, digitColor, digitScale)

	// Confidence percentage below digit - large and bold
	confText := fmt.Sprintf("%.1f%%", result.CIMConfidence*100)
	confScale := 3
	confW := len(confText) * 7 * confScale
	confX := (w - confW) / 2
	confY := digitY + digitHeight + 10
	drawScaledText(img, confText, confX, confY, confScale, whiteColor)

	currentY = heroY + heroSectionH

	// Subtle divider
	for x := padding; x < w-padding; x++ {
		img.Set(x, currentY, color.RGBA{40, 50, 70, 255})
	}

	// ========== CHART SECTION: Full-Width Probability Comparison ==========
	chartY := currentY + 15

	// Title
	titleText := "Probability Distribution"
	titleX := (w - len(titleText)*7) / 2
	drawSimpleText(img, titleText, titleX, chartY, dimColor)
	chartY += 18

	// Bar chart dimensions - much larger bars
	chartWidth := w - 2*padding
	barGroupWidth := chartWidth / 10
	singleBarWidth := (barGroupWidth - 6) / 2  // Gap between FP and CIM bars
	maxBarHeight := chartSectionH - 60  // Taller bars

	// Draw bars for each digit 0-9
	for i := 0; i < 10; i++ {
		groupX := padding + i*barGroupWidth

		// FP bar (blue)
		fpProb := 0.0
		if len(result.FPProbabilities) > i {
			fpProb = result.FPProbabilities[i]
		}
		fpBarH := int(float64(maxBarHeight) * fpProb)
		if fpBarH < 2 && fpProb > 0.005 {
			fpBarH = 2
		}
		fpBarY := chartY + maxBarHeight - fpBarH

		fpBarColor := fpColor
		if i == result.FPPrediction {
			fpBarColor = color.RGBA{120, 180, 255, 255} // Brighter
		}

		// Draw FP bar with top highlight
		for bx := groupX + 2; bx < groupX+2+singleBarWidth; bx++ {
			for by := fpBarY; by < chartY+maxBarHeight; by++ {
				if by == fpBarY {
					// Top highlight
					lighter := color.RGBA{
						uint8(min(255, int(fpBarColor.R)+50)),
						uint8(min(255, int(fpBarColor.G)+50)),
						uint8(min(255, int(fpBarColor.B)+50)),
						255,
					}
					img.Set(bx, by, lighter)
				} else {
					img.Set(bx, by, fpBarColor)
				}
			}
		}

		// CIM bar (cyan/green)
		cimProb := 0.0
		if len(result.CIMProbabilities) > i {
			cimProb = result.CIMProbabilities[i]
		}
		cimBarH := int(float64(maxBarHeight) * cimProb)
		if cimBarH < 2 && cimProb > 0.005 {
			cimBarH = 2
		}
		cimBarY := chartY + maxBarHeight - cimBarH

		cimBarColor := cimColor
		if i == result.CIMPrediction {
			cimBarColor = matchColor // Vibrant green for prediction
		}

		// Draw CIM bar with top highlight
		for bx := groupX + 4 + singleBarWidth; bx < groupX+4+singleBarWidth*2; bx++ {
			for by := cimBarY; by < chartY+maxBarHeight; by++ {
				if by == cimBarY {
					// Top highlight
					lighter := color.RGBA{
						uint8(min(255, int(cimBarColor.R)+50)),
						uint8(min(255, int(cimBarColor.G)+50)),
						uint8(min(255, int(cimBarColor.B)+50)),
						255,
					}
					img.Set(bx, by, lighter)
				} else {
					img.Set(bx, by, cimBarColor)
				}
			}
		}

		// Digit label below bars
		labelX := groupX + barGroupWidth/2 - 3
		labelY := chartY + maxBarHeight + 6
		digitLabel := fmt.Sprintf("%d", i)
		drawSimpleText(img, digitLabel, labelX, labelY, whiteColor)
	}

	// Legend below chart
	legendY := chartY + maxBarHeight + 22
	legendX := padding + 10

	// FP legend
	for bx := legendX; bx < legendX+15; bx++ {
		for by := legendY; by < legendY+10; by++ {
			img.Set(bx, by, fpColor)
		}
	}
	drawSimpleText(img, "FP32", legendX+20, legendY+1, dimColor)

	// CIM legend
	cimLegendX := legendX + 100
	for bx := cimLegendX; bx < cimLegendX+15; bx++ {
		for by := legendY; by < legendY+10; by++ {
			img.Set(bx, by, cimColor)
		}
	}
	drawSimpleText(img, "FeCIM", cimLegendX+20, legendY+1, dimColor)

	currentY = chartY + maxBarHeight + 40

	// ========== FOOTER: Energy Efficiency ==========
	footerY := h - footerH - 8
	effText := fmt.Sprintf("%.0f-%.0fx Energy Efficiency", 25.0, 100.0)
	effX := (w - len(effText)*7) / 2
	drawSimpleText(img, effText, effX, footerY, accentColor)

	return img
}

// drawPredictionCard draws a single prediction card (legacy).
func (cc *ComparisonCard) drawPredictionCard(img *image.RGBA, x, y, w, h int,
	title, subtitle string, prediction int, confidence float64, accentColor color.RGBA, probs []float64) {
	cc.drawPredictionCardEnhanced(img, x, y, w, h, title, subtitle, prediction, confidence, accentColor, probs, 3)
}

// drawPredictionCardEnhanced draws a prediction card with configurable digit scale.
func (cc *ComparisonCard) drawPredictionCardEnhanced(img *image.RGBA, x, y, w, h int,
	title, subtitle string, prediction int, confidence float64, accentColor color.RGBA, probs []float64, digitScale int) {

	// Card background with subtle gradient
	for cx := x; cx < x+w; cx++ {
		for cy := y; cy < y+h; cy++ {
			// Subtle vertical gradient
			gradientFactor := float64(cy-y) / float64(h)
			r := uint8(30 + gradientFactor*10)
			g := uint8(35 + gradientFactor*10)
			b := uint8(50 + gradientFactor*10)
			img.Set(cx, cy, color.RGBA{r, g, b, 255})
		}
	}

	// Thicker border (accent color) - 2px
	for cx := x; cx < x+w; cx++ {
		img.Set(cx, y, accentColor)
		img.Set(cx, y+1, accentColor)
		img.Set(cx, y+h-1, accentColor)
		img.Set(cx, y+h-2, accentColor)
	}
	for cy := y; cy < y+h; cy++ {
		img.Set(x, cy, accentColor)
		img.Set(x+1, cy, accentColor)
		img.Set(x+w-1, cy, accentColor)
		img.Set(x+w-2, cy, accentColor)
	}

	// Title
	titleY := y + 10
	drawSimpleText(img, title, x+10, titleY, accentColor)

	// Subtitle
	subtitleY := titleY + 14
	drawSimpleText(img, subtitle, x+10, subtitleY, color.RGBA{120, 120, 140, 255})

	// LARGE prediction digit (centered)
	digitHeight := 7 * digitScale
	digitWidth := 5 * digitScale
	digitY := subtitleY + 18
	digitX := x + (w-digitWidth)/2

	digitText := fmt.Sprintf("%d", prediction)
	if prediction < 0 {
		digitText = "?"
	}
	cc.drawScaledDigit(img, digitX, digitY, digitText, accentColor, digitScale)

	// Confidence bar - wider and more prominent (UI-021 fix: increased from 16 to 24px)
	barY := digitY + digitHeight + 12
	barX := x + 15
	barWidth := w - 30
	barHeight := 24

	// Background
	for bx := barX; bx < barX+barWidth; bx++ {
		for by := barY; by < barY+barHeight; by++ {
			img.Set(bx, by, color.RGBA{40, 45, 60, 255})
		}
	}

	// Fill with gradient
	fillWidth := int(float64(barWidth) * confidence)
	for bx := barX; bx < barX+fillWidth; bx++ {
		for by := barY; by < barY+barHeight; by++ {
			// Vertical gradient on fill
			t := float64(by-barY) / float64(barHeight)
			r := uint8(float64(accentColor.R) * (1 - t*0.3))
			g := uint8(float64(accentColor.G) * (1 - t*0.3))
			b := uint8(float64(accentColor.B) * (1 - t*0.3))
			img.Set(bx, by, color.RGBA{r, g, b, 255})
		}
	}

	// Confidence percentage text (UI-021 fix: moved to right side of bar)
	confY := barY + (barHeight / 2) - 5
	confText := fmt.Sprintf("%.1f%%", confidence*100)
	confX := barX + barWidth + 10
	drawSimpleText(img, confText, confX, confY, color.RGBA{220, 220, 240, 255})

	// Mini probability distribution (bottom of card)
	if len(probs) == 10 {
		probY := confY + 22
		probBarWidth := (w - 30) / 10
		probBarMaxH := h - (probY - y) - 12

		for i, p := range probs {
			probBarX := x + 15 + i*probBarWidth
			probBarH := int(float64(probBarMaxH) * p)
			if probBarH < 1 {
				probBarH = 1
			}

			probBarY := y + h - 8 - probBarH

			barColor := color.RGBA{60, 65, 85, 255}
			if i == prediction {
				barColor = accentColor
			}

			for bx := probBarX; bx < probBarX+probBarWidth-2; bx++ {
				for by := probBarY; by < y+h-8; by++ {
					img.Set(bx, by, barColor)
				}
			}
		}
	}
}

// drawGlowCircle draws a circle with a glow effect.
func (cc *ComparisonCard) drawGlowCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	// Outer glow (larger, faded)
	for dy := -r - 5; dy <= r+5; dy++ {
		for dx := -r - 5; dx <= r+5; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > float64(r) && dist <= float64(r+5) {
				// Fade based on distance
				alpha := uint8(80 * (1 - (dist-float64(r))/5))
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					// Blend with background
					bg := img.RGBAAt(px, py)
					blended := color.RGBA{
						R: uint8((int(bg.R)*(255-int(alpha)) + int(c.R)*int(alpha)) / 255),
						G: uint8((int(bg.G)*(255-int(alpha)) + int(c.G)*int(alpha)) / 255),
						B: uint8((int(bg.B)*(255-int(alpha)) + int(c.B)*int(alpha)) / 255),
						A: 255,
					}
					img.Set(px, py, blended)
				}
			}
		}
	}

	// Main circle
	cc.drawCircle(img, cx, cy, r, c)
}

// drawSmallCheckmark draws a small checkmark symbol.
func (cc *ComparisonCard) drawSmallCheckmark(img *image.RGBA, cx, cy int, c color.RGBA) {
	// Draw a checkmark using lines
	// Short leg: from bottom-left going up-right
	for i := 0; i < 4; i++ {
		x := cx - 3 + i
		y := cy + 1 - i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}
	}
	// Long leg: from middle going up-right
	for i := 0; i < 6; i++ {
		x := cx + i
		y := cy - 2 + i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}
	}
}

// drawSmallWarning draws a small warning triangle.
func (cc *ComparisonCard) drawSmallWarning(img *image.RGBA, cx, cy int, c color.RGBA) {
	// Draw triangle outline
	// Triangle vertices: top (cx, cy-4), bottom-left (cx-4, cy+4), bottom-right (cx+4, cy+4)
	// Top to bottom-left
	for i := 0; i < 5; i++ {
		x := cx - i
		y := cy - 4 + 2*i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}
	}
	// Top to bottom-right
	for i := 0; i < 5; i++ {
		x := cx + i
		y := cy - 4 + 2*i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}
	}
	// Bottom edge
	for i := -4; i <= 4; i++ {
		x := cx + i
		y := cy + 4
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}
	}
	// Exclamation mark inside
	// Vertical line
	for i := -2; i <= 1; i++ {
		x := cx
		y := cy + i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}
	}
	// Dot
	img.Set(cx, cy+3, c)
}

// drawCheckmark draws a clean checkmark symbol for the new layout.
func (cc *ComparisonCard) drawCheckmark(img *image.RGBA, cx, cy int, c color.RGBA) {
	// Draw a simple checkmark
	// Short leg
	for i := 0; i < 4; i++ {
		x := cx - 2 + i
		y := cy + 2 - i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
			img.Set(x, y+1, c)
		}
	}
	// Long leg
	for i := 0; i < 7; i++ {
		x := cx + 1 + i
		y := cy - 1 + i
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
			img.Set(x, y+1, c)
		}
	}
}

// drawWarningIcon draws a warning triangle/exclamation for the new layout.
func (cc *ComparisonCard) drawWarningIcon(img *image.RGBA, cx, cy int, c color.RGBA) {
	// Simple filled circle with exclamation
	r := 5
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				px, py := cx+dx, cy+dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
	// Exclamation mark (dark)
	dark := color.RGBA{20, 28, 38, 255}
	for i := -2; i <= 1; i++ {
		if cx >= 0 && cx < img.Bounds().Dx() && cy+i >= 0 && cy+i < img.Bounds().Dy() {
			img.Set(cx, cy+i, dark)
		}
	}
	if cx >= 0 && cx < img.Bounds().Dx() && cy+3 >= 0 && cy+3 < img.Bounds().Dy() {
		img.Set(cx, cy+3, dark)
	}
}

// drawLargeCheckmark draws a checkmark symbol.
func (cc *ComparisonCard) drawLargeCheckmark(img *image.RGBA, cx, cy int, c color.RGBA) {
	white := color.RGBA{255, 255, 255, 255}
	// Draw a checkmark using lines
	// Short leg: from bottom-left going up-right
	for i := 0; i < 8; i++ {
		x := cx - 8 + i
		y := cy + 2 - i
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				px, py := x+dx, y+dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, white)
				}
			}
		}
	}
	// Long leg: from middle going up-right
	for i := 0; i < 12; i++ {
		x := cx + i
		y := cy - 6 + i
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				px, py := x+dx, y+dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, white)
				}
			}
		}
	}
}

// drawLargeX draws an X symbol for mismatch.
func (cc *ComparisonCard) drawLargeX(img *image.RGBA, cx, cy int, c color.RGBA) {
	white := color.RGBA{255, 255, 255, 255}
	// Draw X using two diagonal lines
	for i := -10; i <= 10; i++ {
		// Line 1: top-left to bottom-right
		x1, y1 := cx+i, cy+i
		// Line 2: top-right to bottom-left
		x2, y2 := cx+i, cy-i

		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				px1, py1 := x1+dx, y1+dy
				px2, py2 := x2+dx, y2+dy
				if px1 >= 0 && px1 < img.Bounds().Dx() && py1 >= 0 && py1 < img.Bounds().Dy() {
					img.Set(px1, py1, white)
				}
				if px2 >= 0 && px2 < img.Bounds().Dx() && py2 >= 0 && py2 < img.Bounds().Dy() {
					img.Set(px2, py2, white)
				}
			}
		}
	}
}

// drawScaledDigit draws a single digit with configurable scale.
func (cc *ComparisonCard) drawScaledDigit(img *image.RGBA, x, y int, digit string, c color.RGBA, scale int) {
	patterns := map[rune][]string{
		'0': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
		'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
		'2': {"01110", "10001", "00001", "00110", "01000", "10000", "11111"},
		'3': {"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
		'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
		'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
		'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
		'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
		'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
		'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
		'?': {"01110", "10001", "00001", "00110", "00100", "00000", "00100"},
	}

	for _, ch := range digit {
		pattern, ok := patterns[ch]
		if !ok {
			continue
		}

		for dy, row := range pattern {
			for dx, pixel := range row {
				if pixel == '1' {
					// Draw scaled pixel
					for sy := 0; sy < scale; sy++ {
						for sx := 0; sx < scale; sx++ {
							px := x + dx*scale + sx
							py := y + dy*scale + sy
							if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
								img.Set(px, py, c)
							}
						}
					}
				}
			}
		}
	}
}

// drawLargeDigit draws a scaled-up digit.
func (cc *ComparisonCard) drawLargeDigit(img *image.RGBA, x, y int, digit string, c color.RGBA) {
	// 3x scale for the digit
	scale := 3

	patterns := map[rune][]string{
		'0': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
		'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
		'2': {"01110", "10001", "00001", "00110", "01000", "10000", "11111"},
		'3': {"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
		'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
		'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
		'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
		'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
		'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
		'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
		'?': {"01110", "10001", "00001", "00110", "00100", "00000", "00100"},
	}

	for _, ch := range digit {
		pattern, ok := patterns[ch]
		if !ok {
			continue
		}

		for dy, row := range pattern {
			for dx, pixel := range row {
				if pixel == '1' {
					// Draw scaled pixel
					for sy := 0; sy < scale; sy++ {
						for sx := 0; sx < scale; sx++ {
							px := x + dx*scale + sx
							py := y + dy*scale + sy
							if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
								img.Set(px, py, c)
							}
						}
					}
				}
			}
		}
	}
}

// drawCircle draws a filled circle.
func (cc *ComparisonCard) drawCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// getSecondBest returns the second-highest prediction and its confidence.
func (cc *ComparisonCard) getSecondBest(probs []float64) (int, float64) {
	if len(probs) < 2 {
		return -1, 0
	}

	bestIdx, secondIdx := 0, 1
	bestVal, secondVal := probs[0], probs[1]

	if secondVal > bestVal {
		bestIdx, secondIdx = secondIdx, bestIdx
		bestVal, secondVal = secondVal, bestVal
	}

	for i := 2; i < len(probs); i++ {
		if probs[i] > bestVal {
			secondIdx, secondVal = bestIdx, bestVal
			bestIdx, bestVal = i, probs[i]
		} else if probs[i] > secondVal {
			secondIdx, secondVal = i, probs[i]
		}
	}

	return secondIdx, secondVal
}

// DualProbabilityChart shows FP vs CIM probability comparison with divergence highlighting.
type DualProbabilityChart struct {
	widget.BaseWidget

	mu          sync.RWMutex
	fpProbs     []float64
	cimProbs    []float64
	divergences []float64
	fpPred      int
	cimPred     int

	raster *canvas.Raster
}

// NewDualProbabilityChart creates a new dual probability chart.
func NewDualProbabilityChart() *DualProbabilityChart {
	dpc := &DualProbabilityChart{
		fpProbs:     make([]float64, 10),
		cimProbs:    make([]float64, 10),
		divergences: make([]float64, 10),
		fpPred:      -1,
		cimPred:     -1,
	}
	dpc.ExtendBaseWidget(dpc)
	return dpc
}

// SetProbabilities updates both FP and CIM probabilities.
func (dpc *DualProbabilityChart) SetProbabilities(fpProbs, cimProbs []float64, fpPred, cimPred int) {
	dpc.mu.Lock()
	defer dpc.mu.Unlock()

	dpc.fpProbs = fpProbs
	dpc.cimProbs = cimProbs
	dpc.fpPred = fpPred
	dpc.cimPred = cimPred

	// Calculate divergences
	dpc.divergences = make([]float64, len(fpProbs))
	for i := range fpProbs {
		if i < len(cimProbs) {
			dpc.divergences[i] = math.Abs(fpProbs[i] - cimProbs[i])
		}
	}

	fyne.Do(func() {
		dpc.Refresh()
	})
}

// Clear resets the chart.
func (dpc *DualProbabilityChart) Clear() {
	dpc.mu.Lock()
	dpc.fpProbs = make([]float64, 10)
	dpc.cimProbs = make([]float64, 10)
	dpc.divergences = make([]float64, 10)
	dpc.fpPred = -1
	dpc.cimPred = -1
	dpc.mu.Unlock()
	fyne.Do(func() {
		dpc.Refresh()
	})
}

// MinSize returns the minimum size.
func (dpc *DualProbabilityChart) MinSize() fyne.Size {
	return fyne.NewSize(400, 150)
}

// CreateRenderer implements fyne.Widget.
func (dpc *DualProbabilityChart) CreateRenderer() fyne.WidgetRenderer {
	dpc.raster = canvas.NewRaster(dpc.generateImage)

	title := widget.NewLabelWithStyle("Probability Distribution (FP vs CIM)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewBorder(
		title,
		nil, nil, nil,
		container.NewMax(dpc.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage creates the dual probability bar chart.
func (dpc *DualProbabilityChart) generateImage(w, h int) image.Image {
	if w < 10 {
		w = 400
	}
	if h < 10 {
		h = 130
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 30, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	dpc.mu.RLock()
	fpProbs := dpc.fpProbs
	cimProbs := dpc.cimProbs
	divergences := dpc.divergences
	fpPred := dpc.fpPred
	cimPred := dpc.cimPred
	dpc.mu.RUnlock()

	if len(fpProbs) < 10 {
		return img
	}

	padding := 20
	labelHeight := 15
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding - labelHeight

	groupWidth := chartWidth / 10
	barWidth := (groupWidth - 4) / 2

	fpColor := color.RGBA{100, 150, 255, 255}   // Blue
	cimColor := color.RGBA{100, 255, 180, 255}  // Green
	warnColor := color.RGBA{255, 200, 100, 255} // Yellow for divergence

	for i := 0; i < 10; i++ {
		groupX := padding + i*groupWidth

		// FP bar
		fpHeight := int(float64(chartHeight) * fpProbs[i])
		if fpHeight < 1 && fpProbs[i] > 0 {
			fpHeight = 1
		}
		fpBarX := groupX + 1
		fpBarY := padding + chartHeight - fpHeight

		barCol := fpColor
		if i == fpPred {
			barCol = color.RGBA{150, 200, 255, 255} // Brighter for prediction
		}

		for bx := fpBarX; bx < fpBarX+barWidth; bx++ {
			for by := fpBarY; by < padding+chartHeight; by++ {
				img.Set(bx, by, barCol)
			}
		}

		// CIM bar
		cimHeight := int(float64(chartHeight) * cimProbs[i])
		if cimHeight < 1 && cimProbs[i] > 0 {
			cimHeight = 1
		}
		cimBarX := groupX + barWidth + 2
		cimBarY := padding + chartHeight - cimHeight

		barCol = cimColor
		if i == cimPred {
			barCol = color.RGBA{150, 255, 200, 255} // Brighter for prediction
		}

		for bx := cimBarX; bx < cimBarX+barWidth; bx++ {
			for by := cimBarY; by < padding+chartHeight; by++ {
				img.Set(bx, by, barCol)
			}
		}

		// Divergence warning marker (if > 2%)
		if divergences[i] > 0.02 {
			warnY := padding + chartHeight + 2
			warnX := groupX + groupWidth/2 - 2
			for wx := warnX; wx < warnX+4; wx++ {
				for wy := warnY; wy < warnY+4; wy++ {
					img.Set(wx, wy, warnColor)
				}
			}
		}

		// Digit label
		labelX := groupX + groupWidth/2 - 6 // Adjusted for scale 2
		labelY := h - 15
		drawScaledChar(img, rune('0'+i), labelX, labelY, 2, color.RGBA{150, 150, 170, 255})
	}

	// Legend
	legendY := 5
	drawScaledText(img, "FP", padding, legendY, 2, fpColor)
	drawScaledText(img, "CIM", padding+60, legendY, 2, cimColor)

	return img
}
