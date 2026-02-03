// Package widgets provides custom GUI widgets for the hysteresis module.
package widgets

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// PEPlot is a custom widget that displays a P-E hysteresis loop.
type PEPlot struct {
	widget.BaseWidget

	mu           sync.RWMutex
	eData        []float64
	pData        []float64
	currentE     float64
	currentP     float64
	eMax         float64
	pMax         float64
	ec           float64 // Actual coercive field for markers
	pr           float64 // Actual remanent polarization for markers
	minSize      fyne.Size
	filterSpikes bool

	// Color scheme (passed in during construction or via setters)
	ColorBackground color.RGBA
	ColorGrid       color.RGBA
	ColorAxis       color.RGBA
	ColorPositive   color.RGBA
	ColorNegative   color.RGBA
	ColorWarning    color.RGBA
}

// NewPEPlot creates a new P-E plot widget
func NewPEPlot(eMax, pMax float64, colorBg, colorGrid, colorAxis, colorPos, colorNeg, colorWarn color.RGBA) *PEPlot {
	p := &PEPlot{
		eMax:            eMax,
		pMax:            pMax,
		minSize:         fyne.NewSize(400, 300),
		filterSpikes:    true,
		ColorBackground: colorBg,
		ColorGrid:       colorGrid,
		ColorAxis:       colorAxis,
		ColorPositive:   colorPos,
		ColorNegative:   colorNeg,
		ColorWarning:    colorWarn,
	}
	p.ExtendBaseWidget(p)
	return p
}

func (p *PEPlot) SetMinSize(size fyne.Size) {
	p.minSize = size
}

func (p *PEPlot) MinSize() fyne.Size {
	return p.minSize
}

func (p *PEPlot) SetBounds(eMax, pMax float64) {
	p.mu.Lock()
	p.eMax = eMax
	p.pMax = pMax
	p.mu.Unlock()
}

// SetMaterialParams sets the actual Ec and Pr values for accurate marker placement
func (p *PEPlot) SetMaterialParams(ec, pr float64) {
	p.mu.Lock()
	p.ec = ec
	p.pr = pr
	p.mu.Unlock()
	p.Refresh() // Redraw with new marker positions
}

// SetSpikeFiltering toggles spike/discontinuity filtering for line segments.
// Disable for L-K mode to avoid dropping legitimate rapid transitions.
func (p *PEPlot) SetSpikeFiltering(enabled bool) {
	p.mu.Lock()
	p.filterSpikes = enabled
	p.mu.Unlock()
}

func (p *PEPlot) SetData(eData, pData []float64, currentE, currentP float64) {
	p.mu.Lock()
	p.eData = eData
	p.pData = pData
	p.currentE = currentE
	p.currentP = currentP
	p.mu.Unlock()
}

func (p *PEPlot) CreateRenderer() fyne.WidgetRenderer {
	return &peplotRenderer{plot: p}
}

type peplotRenderer struct {
	plot    *PEPlot
	objects []fyne.CanvasObject
	cache   sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *peplotRenderer) MinSize() fyne.Size {
	return r.plot.minSize
}

func (r *peplotRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("peplotRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *peplotRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("peplotRenderer", r.plot.Size())
	// Always re-layout on Refresh for this dynamic widget (data changes)
	// layoutWithSize handles size validation and cache marking internally
	r.layoutWithSize(r.plot.Size())
}

func (r *peplotRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.plot.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.plot.mu.RLock()
	defer r.plot.mu.RUnlock()

	r.objects = r.objects[:0]

	// Background with subtle border
	bg := canvas.NewRectangle(color.RGBA{0, 40, 90, 255})
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Inner background
	innerBg := canvas.NewRectangle(r.plot.ColorBackground)
	innerBg.Resize(fyne.NewSize(size.Width-4, size.Height-4))
	innerBg.Move(fyne.NewPos(2, 2))
	r.objects = append(r.objects, innerBg)

	// Margins - adjusted for better proportions
	marginLeft := float32(60)
	marginRight := float32(20)
	marginTop := float32(50)
	marginBottom := float32(35)
	plotW := size.Width - marginLeft - marginRight
	plotH := size.Height - marginTop - marginBottom

	// Title inside plot area
	plotTitle := canvas.NewText("P-E Hysteresis Loop", color.RGBA{255, 255, 255, 255})
	plotTitle.TextSize = 16
	plotTitle.TextStyle = fyne.TextStyle{Bold: true}
	plotTitle.Move(fyne.NewPos(size.Width/2-80, 8))
	r.objects = append(r.objects, plotTitle)

	// Legend inside plot area (top-left corner)
	legendX := marginLeft + 10
	legendY := marginTop + 10
	legendItems := []struct {
		color color.RGBA
		text  string
	}{
		{color.RGBA{255, 150, 0, 255}, "Ec"},
		{color.RGBA{0, 200, 150, 255}, "Pr"},
		{r.plot.ColorWarning, "Pos"},
	}
	for _, item := range legendItems {
		// Color box - increased from 14 to 16 for better visibility
		box := canvas.NewRectangle(item.color)
		box.Resize(fyne.NewSize(16, 16))
		box.Move(fyne.NewPos(legendX, legendY))
		r.objects = append(r.objects, box)
		// Label - increased from 12pt to 13pt
		label := canvas.NewText(item.text, color.RGBA{200, 200, 200, 255})
		label.TextSize = 13
		label.Move(fyne.NewPos(legendX+20, legendY))
		r.objects = append(r.objects, label)
		legendX += 65
	}

	// Grid lines (fewer divisions to avoid overlap)
	numDivisions := 2 // divisions on each side of center (shows -1, -0.5, 0, 0.5, 1)
	for i := -numDivisions; i <= numDivisions; i++ {
		t := float32(i) / float32(numDivisions)

		// Vertical grid line
		x := marginLeft + plotW/2 + t*plotW/2
		vLine := canvas.NewLine(r.plot.ColorGrid)
		vLine.Position1 = fyne.NewPos(x, marginTop)
		vLine.Position2 = fyne.NewPos(x, marginTop+plotH)
		if i == 0 {
			vLine.StrokeWidth = 2
			vLine.StrokeColor = r.plot.ColorAxis
		} else {
			vLine.StrokeWidth = 1
		}
		r.objects = append(r.objects, vLine)

		// Horizontal grid line
		y := marginTop + plotH/2 - t*plotH/2
		hLine := canvas.NewLine(r.plot.ColorGrid)
		hLine.Position1 = fyne.NewPos(marginLeft, y)
		hLine.Position2 = fyne.NewPos(marginLeft+plotW, y)
		if i == 0 {
			hLine.StrokeWidth = 2
			hLine.StrokeColor = r.plot.ColorAxis
		} else {
			hLine.StrokeWidth = 1
		}
		r.objects = append(r.objects, hLine)

		// X-axis tick labels (E-field in MV/cm) - only at edges and center
		if i != 0 && (i == -numDivisions || i == numDivisions) {
			eVal := float64(t) * r.plot.eMax / 1e8 // Convert to MV/cm
			eTickLabel := canvas.NewText(fmt.Sprintf("%.1f", eVal), color.RGBA{200, 200, 200, 255})
			eTickLabel.TextSize = 12
			eTickLabel.Move(fyne.NewPos(x-15, marginTop+plotH+5))
			r.objects = append(r.objects, eTickLabel)
		}

		// Y-axis tick labels (P in µC/cm²) - only at edges and center
		if i != 0 && (i == -numDivisions || i == numDivisions) {
			pVal := float64(t) * r.plot.pMax * 100 // Convert to µC/cm²
			pTickLabel := canvas.NewText(fmt.Sprintf("%.0f", pVal), color.RGBA{200, 200, 200, 255})
			pTickLabel.TextSize = 12
			pTickLabel.Move(fyne.NewPos(marginLeft-35, y-7))
			r.objects = append(r.objects, pTickLabel)
		}
	}

	// Additional subtle grid lines for easier value reading
	// These are finer subdivisions with very low alpha to not distract from main plot
	subtleGridColor := color.RGBA{0, 80, 140, 25} // Very low alpha for subtle appearance
	numSubdivisions := 4                          // More divisions for finer grid
	for i := -numSubdivisions; i <= numSubdivisions; i++ {
		// Skip the main grid lines (already drawn above)
		if i == 0 || i == -numSubdivisions || i == numSubdivisions || i == -numSubdivisions/2 || i == numSubdivisions/2 {
			continue
		}
		t := float32(i) / float32(numSubdivisions)

		// Subtle vertical grid line
		x := marginLeft + plotW/2 + t*plotW/2
		vLine := canvas.NewLine(subtleGridColor)
		vLine.Position1 = fyne.NewPos(x, marginTop)
		vLine.Position2 = fyne.NewPos(x, marginTop+plotH)
		vLine.StrokeWidth = 1
		r.objects = append(r.objects, vLine)

		// Subtle horizontal grid line
		y := marginTop + plotH/2 - t*plotH/2
		hLine := canvas.NewLine(subtleGridColor)
		hLine.Position1 = fyne.NewPos(marginLeft, y)
		hLine.Position2 = fyne.NewPos(marginLeft+plotW, y)
		hLine.StrokeWidth = 1
		r.objects = append(r.objects, hLine)
	}

	// Center coordinates
	centerX := marginLeft + plotW/2
	centerY := marginTop + plotH/2

	// Zero label (single, at corner to avoid overlap)
	zeroLabel := canvas.NewText("0", color.RGBA{200, 200, 200, 255})
	zeroLabel.TextSize = 12
	zeroLabel.Move(fyne.NewPos(centerX-10, centerY+4))
	r.objects = append(r.objects, zeroLabel)

	// Axis title labels (positioned at ends of axes) - larger
	eLabel := canvas.NewText("E (MV/cm)", color.RGBA{0, 212, 255, 255})
	eLabel.TextSize = 13
	eLabel.TextStyle = fyne.TextStyle{Bold: true}
	eLabel.Move(fyne.NewPos(marginLeft+plotW-70, marginTop+plotH+15))
	r.objects = append(r.objects, eLabel)

	pLabelText := canvas.NewText("P (µC/cm²)", color.RGBA{0, 212, 255, 255})
	pLabelText.TextSize = 13
	pLabelText.TextStyle = fyne.TextStyle{Bold: true}
	pLabelText.Move(fyne.NewPos(marginLeft-55, marginTop-18))
	r.objects = append(r.objects, pLabelText)

	// Ec markers (vertical dashed lines at ±Ec) - M01: Ec threshold visualization
	// Shows coercive field threshold where switching occurs
	// Use actual Ec value if set, otherwise fall back to ratio
	var ecRatio float32
	if r.plot.ec > 0 && r.plot.eMax > 0 {
		ecRatio = float32(r.plot.ec / r.plot.eMax)
	} else {
		ecRatio = float32(1.0 / 1.5) // Fallback: assume eMax = 1.5*Ec
	}
	ecPosX := centerX + ecRatio*plotW/2
	ecNegX := centerX - ecRatio*plotW/2

	// +Ec marker (dashed effect with shorter line)
	ecPosLine := canvas.NewLine(color.RGBA{255, 150, 0, 120})
	ecPosLine.Position1 = fyne.NewPos(ecPosX, marginTop)
	ecPosLine.Position2 = fyne.NewPos(ecPosX, marginTop+plotH)
	ecPosLine.StrokeWidth = 2
	r.objects = append(r.objects, ecPosLine)
	// Label inside plot area with "Coercive Field" tooltip context
	ecPosLabel := canvas.NewText("+Ec (Coercive)", color.RGBA{255, 150, 0, 255})
	ecPosLabel.TextSize = 11
	ecPosLabel.Move(fyne.NewPos(ecPosX+4, marginTop+5))
	r.objects = append(r.objects, ecPosLabel)

	// -Ec marker
	ecNegLine := canvas.NewLine(color.RGBA{255, 150, 0, 120})
	ecNegLine.Position1 = fyne.NewPos(ecNegX, marginTop)
	ecNegLine.Position2 = fyne.NewPos(ecNegX, marginTop+plotH)
	ecNegLine.StrokeWidth = 2
	r.objects = append(r.objects, ecNegLine)
	// Label inside plot area with "Coercive Field" tooltip context
	ecNegLabel := canvas.NewText("-Ec (Coercive)", color.RGBA{255, 150, 0, 255})
	ecNegLabel.TextSize = 11
	ecNegLabel.Move(fyne.NewPos(ecNegX-75, marginTop+5))
	r.objects = append(r.objects, ecNegLabel)

	// No-switching zone shading (optional enhancement)
	// Shows region below |Ec| where polarization remains stable
	noSwitchRect := canvas.NewRectangle(color.RGBA{255, 100, 100, 15}) // Very subtle red
	noSwitchRect.Resize(fyne.NewSize(ecPosX-ecNegX, plotH))
	noSwitchRect.Move(fyne.NewPos(ecNegX, marginTop))
	r.objects = append(r.objects, noSwitchRect)
	// Label for no-switching zone
	noSwitchLabel := canvas.NewText("No switching below |Ec|", color.RGBA{255, 150, 150, 180})
	noSwitchLabel.TextSize = 9
	noSwitchLabel.Move(fyne.NewPos(centerX-60, marginTop+plotH-20))
	r.objects = append(r.objects, noSwitchLabel)

	// Pr markers (horizontal dashed lines at ±Pr)
	// Use actual Pr value if set, otherwise fall back to ratio
	var prRatio float32
	if r.plot.pr > 0 && r.plot.pMax > 0 {
		prRatio = float32(r.plot.pr / r.plot.pMax)
	} else {
		prRatio = float32(0.8 / 1.2) // Fallback: assume Pr/Ps ≈ 0.8, pMax = 1.2*Ps
	}
	prPosY := centerY - prRatio*plotH/2
	prNegY := centerY + prRatio*plotH/2

	// +Pr marker
	prPosLine := canvas.NewLine(color.RGBA{0, 200, 150, 120})
	prPosLine.Position1 = fyne.NewPos(marginLeft, prPosY)
	prPosLine.Position2 = fyne.NewPos(marginLeft+plotW, prPosY)
	prPosLine.StrokeWidth = 2
	r.objects = append(r.objects, prPosLine)
	// Label inside plot area
	prPosLabel := canvas.NewText("+Pr", color.RGBA{0, 200, 150, 255})
	prPosLabel.TextSize = 11
	prPosLabel.Move(fyne.NewPos(marginLeft+plotW-30, prPosY-14))
	r.objects = append(r.objects, prPosLabel)

	// -Pr marker
	prNegLine := canvas.NewLine(color.RGBA{0, 200, 150, 120})
	prNegLine.Position1 = fyne.NewPos(marginLeft, prNegY)
	prNegLine.Position2 = fyne.NewPos(marginLeft+plotW, prNegY)
	prNegLine.StrokeWidth = 2
	r.objects = append(r.objects, prNegLine)
	// Label inside plot area
	prNegLabel := canvas.NewText("-Pr", color.RGBA{0, 200, 150, 255})
	prNegLabel.TextSize = 11
	prNegLabel.Move(fyne.NewPos(marginLeft+plotW-28, prNegY+4))
	r.objects = append(r.objects, prNegLabel)

	// Plot the hysteresis data as smooth curve segments
	// Strategy: Only connect points that are close together on the E-P plane
	// Discontinuous jumps (phase transitions) are shown as separate curve segments
	if len(r.plot.eData) > 1 {
		for i := 1; i < len(r.plot.eData); i++ {
			// Calculate distance between consecutive points
			eDiff := r.plot.eData[i] - r.plot.eData[i-1]
			pDiff := r.plot.pData[i] - r.plot.pData[i-1]
			if eDiff < 0 {
				eDiff = -eDiff
			}
			if pDiff < 0 {
				pDiff = -pDiff
			}

			// Normalize to [0,1] range
			normE := eDiff / r.plot.eMax
			normP := pDiff / r.plot.pMax

			// Map data to screen coordinates
			x1 := marginLeft + plotW/2 + float32(r.plot.eData[i-1]/r.plot.eMax)*plotW/2
			y1 := centerY - float32(r.plot.pData[i-1]/r.plot.pMax)*plotH/2
			x2 := marginLeft + plotW/2 + float32(r.plot.eData[i]/r.plot.eMax)*plotW/2
			y2 := centerY - float32(r.plot.pData[i]/r.plot.pMax)*plotH/2

			// Color based on age (fade effect)
			age := float64(len(r.plot.eData)-i) / float64(len(r.plot.eData))
			alpha := uint8(255 - age*205)

			var lineColor color.RGBA
			if r.plot.pData[i] >= 0 {
				lineColor = color.RGBA{r.plot.ColorPositive.R, r.plot.ColorPositive.G, r.plot.ColorPositive.B, alpha}
			} else {
				lineColor = color.RGBA{r.plot.ColorNegative.R, r.plot.ColorNegative.G, r.plot.ColorNegative.B, alpha}
			}

			// Only draw line if points are close (smooth curve segment)
			// These thresholds are tuned for WRD mode where we need to filter
			// discontinuities between cycles while preserving the steep switching region
			// Typical WRD discontinuity: E≈0, ΔP > 100% (cycle boundary)
			// Typical steep switch: ΔE≈2%, ΔP≈10% (legitimate curve)
			// Also treat large diagonal jumps as discontinuities (missing PREP points).
			jumpMag := math.Hypot(normE, normP)
			isSpike := (normE < 0.05 && normP > 0.30) || // Vertical spike (P jumps while E stays)
				(normE > 0.30 && normP < 0.05) || // Horizontal spike (E jumps while P stays)
				normP > 0.50 || // Any P jump > 50% is definitely a discontinuity
				jumpMag > 0.35 || // Large diagonal jump (skip connecting line)
				(normE > 0.20 && normP > 0.20) // Moderate diagonal jump

			if !r.plot.filterSpikes || !isSpike {
				line := canvas.NewLine(lineColor)
				line.Position1 = fyne.NewPos(x1, y1)
				line.Position2 = fyne.NewPos(x2, y2)
				line.StrokeWidth = 2
				r.objects = append(r.objects, line)
			}
			// Note: We don't draw disconnected points as dots to keep the plot clean
			// The visible curve segments show the hysteresis loop shape clearly
		}
	}

	// Current position marker - large with glow effect
	markerX := marginLeft + plotW/2 + float32(r.plot.currentE/r.plot.eMax)*plotW/2
	markerY := centerY - float32(r.plot.currentP/r.plot.pMax)*plotH/2

	// Outer glow (larger, semi-transparent)
	markerGlow := canvas.NewCircle(color.RGBA{r.plot.ColorWarning.R, r.plot.ColorWarning.G, r.plot.ColorWarning.B, 80})
	markerGlow.Resize(fyne.NewSize(28, 28))
	markerGlow.Move(fyne.NewPos(markerX-14, markerY-14))
	r.objects = append(r.objects, markerGlow)

	// Middle glow
	markerMidGlow := canvas.NewCircle(color.RGBA{r.plot.ColorWarning.R, r.plot.ColorWarning.G, r.plot.ColorWarning.B, 150})
	markerMidGlow.Resize(fyne.NewSize(20, 20))
	markerMidGlow.Move(fyne.NewPos(markerX-10, markerY-10))
	r.objects = append(r.objects, markerMidGlow)

	// Main marker (solid)
	marker := canvas.NewCircle(r.plot.ColorWarning)
	marker.Resize(fyne.NewSize(14, 14))
	marker.Move(fyne.NewPos(markerX-7, markerY-7))
	r.objects = append(r.objects, marker)

	// Inner highlight (white center)
	markerInner := canvas.NewCircle(color.RGBA{255, 255, 255, 200})
	markerInner.Resize(fyne.NewSize(6, 6))
	markerInner.Move(fyne.NewPos(markerX-3, markerY-3))
	r.objects = append(r.objects, markerInner)

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
}

func (r *peplotRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *peplotRenderer) Destroy() {}
