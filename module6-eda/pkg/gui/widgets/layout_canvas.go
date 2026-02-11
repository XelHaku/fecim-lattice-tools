// pkg/gui/widgets/layout_canvas.go
// Interactive layout visualization widget for FeCIM crossbar arrays
// Renders the array layout using Fyne canvas primitives

package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// LayoutCanvas is a custom widget that visualizes FeCIM crossbar arrays
type LayoutCanvas struct {
	widget.BaseWidget
	config     *config.ArrayConfig
	cellWidth  float32
	cellHeight float32
	margin     float32
	container  *fyne.Container
}

// Color definitions
var (
	colorBackground = color.RGBA{10, 21, 32, 255}
	colorCell       = color.RGBA{26, 58, 74, 255}
	colorCell1T1R   = color.RGBA{42, 74, 58, 255}
	colorTransistor = color.RGBA{58, 90, 74, 255}
	colorWL         = color.RGBA{255, 102, 0, 255}
	colorBL         = color.RGBA{0, 204, 102, 255}
	colorSL         = color.RGBA{204, 102, 255, 255}
	colorPin        = color.RGBA{255, 204, 0, 255}
	colorGrid       = color.RGBA{51, 68, 85, 128}
	colorText       = color.RGBA{0, 204, 255, 255}
	colorTextDim    = color.RGBA{102, 170, 204, 255}
)

// NewLayoutCanvas creates a new layout visualization widget
func NewLayoutCanvas(cfg *config.ArrayConfig) *LayoutCanvas {
	lc := &LayoutCanvas{
		config:     cfg,
		cellWidth:  50, // Larger cells for better visibility
		cellHeight: 60,
		margin:     70,
	}
	lc.ExtendBaseWidget(lc)
	lc.container = container.NewWithoutLayout()
	lc.Refresh()
	return lc
}

// SetConfig updates the array configuration and refreshes the display
func (lc *LayoutCanvas) SetConfig(cfg *config.ArrayConfig) {
	lc.config = cfg
	lc.Refresh()
}

// CreateRenderer implements fyne.Widget
func (lc *LayoutCanvas) CreateRenderer() fyne.WidgetRenderer {
	return &layoutCanvasRenderer{
		canvas: lc,
	}
}

// MinSize returns the minimum size for the widget
func (lc *LayoutCanvas) MinSize() fyne.Size {
	if lc.config == nil {
		return fyne.NewSize(500, 400)
	}
	width := lc.margin*2 + float32(lc.config.Cols)*lc.cellWidth
	// Extra space for labels and legend (more for 1T1R with SL labels)
	legendSpace := float32(70)
	if lc.config.Architecture == "1t1r" {
		legendSpace = 90
	}
	height := lc.margin*2 + float32(lc.config.Rows)*lc.cellHeight + legendSpace
	// Ensure minimum usable size
	if width < 500 {
		width = 500
	}
	if height < 400 {
		height = 400
	}
	return fyne.NewSize(width, height)
}

type layoutCanvasRenderer struct {
	canvas  *LayoutCanvas
	objects []fyne.CanvasObject
}

func (r *layoutCanvasRenderer) Layout(size fyne.Size) {
	// Objects are positioned absolutely
}

func (r *layoutCanvasRenderer) MinSize() fyne.Size {
	return r.canvas.MinSize()
}

func (r *layoutCanvasRenderer) Refresh() {
	r.objects = r.buildLayout()
	canvas.Refresh(r.canvas)
}

func (r *layoutCanvasRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *layoutCanvasRenderer) Destroy() {}

func (r *layoutCanvasRenderer) buildLayout() []fyne.CanvasObject {
	lc := r.canvas
	cfg := lc.config
	if cfg == nil || cfg.Rows == 0 || cfg.Cols == 0 {
		// Show placeholder
		text := canvas.NewText("Configure array and click 'Generate All'", colorText)
		text.TextSize = 14
		text.Move(fyne.NewPos(20, 20))
		return []fyne.CanvasObject{text}
	}

	var objects []fyne.CanvasObject
	is1T1R := cfg.Architecture == "1t1r"

	// Calculate dimensions
	arrayWidth := float32(cfg.Cols) * lc.cellWidth
	arrayHeight := float32(cfg.Rows) * lc.cellHeight

	// Background - include space for legend
	legendSpace := float32(60)
	if is1T1R {
		legendSpace = 80 // Extra space for SL legend
	}
	bg := canvas.NewRectangle(colorBackground)
	bg.Resize(fyne.NewSize(arrayWidth+lc.margin*2, arrayHeight+lc.margin*2+legendSpace))
	objects = append(objects, bg)

	// Title
	archLabel := "Passive"
	if is1T1R {
		archLabel = "1T1R"
	}
	title := canvas.NewText(fmt.Sprintf("FeCIM %dx%d Crossbar (%s)", cfg.Rows, cfg.Cols, archLabel), colorText)
	title.TextSize = 14
	title.TextStyle.Bold = true
	title.Move(fyne.NewPos(lc.margin, 5))
	objects = append(objects, title)

	// Grid lines
	for col := 0; col <= cfg.Cols; col++ {
		x := lc.margin + float32(col)*lc.cellWidth
		line := canvas.NewLine(colorGrid)
		line.Position1 = fyne.NewPos(x, lc.margin)
		line.Position2 = fyne.NewPos(x, lc.margin+arrayHeight)
		line.StrokeWidth = 0.5
		objects = append(objects, line)
	}
	for row := 0; row <= cfg.Rows; row++ {
		y := lc.margin + float32(row)*lc.cellHeight
		line := canvas.NewLine(colorGrid)
		line.Position1 = fyne.NewPos(lc.margin, y)
		line.Position2 = fyne.NewPos(lc.margin+arrayWidth, y)
		line.StrokeWidth = 0.5
		objects = append(objects, line)
	}

	// Word Lines (horizontal)
	for row := 0; row < cfg.Rows; row++ {
		y := lc.margin + float32(row)*lc.cellHeight + lc.cellHeight/2

		// WL wire
		wlLine := canvas.NewLine(colorWL)
		wlLine.Position1 = fyne.NewPos(lc.margin-20, y)
		wlLine.Position2 = fyne.NewPos(lc.margin+arrayWidth, y)
		wlLine.StrokeWidth = 2
		objects = append(objects, wlLine)

		// WL pin
		pin := canvas.NewCircle(colorPin)
		pin.Resize(fyne.NewSize(8, 8))
		pin.Move(fyne.NewPos(lc.margin-24, y-4))
		objects = append(objects, pin)

		// WL label
		label := canvas.NewText(fmt.Sprintf("WL[%d]", row), colorText)
		label.TextSize = 14
		label.Move(fyne.NewPos(5, y-6))
		objects = append(objects, label)
	}

	// Bit Lines (vertical)
	for col := 0; col < cfg.Cols; col++ {
		x := lc.margin + float32(col)*lc.cellWidth + lc.cellWidth/2

		// BL wire
		blLine := canvas.NewLine(colorBL)
		blLine.Position1 = fyne.NewPos(x, lc.margin)
		blLine.Position2 = fyne.NewPos(x, lc.margin+arrayHeight+15)
		blLine.StrokeWidth = 2
		objects = append(objects, blLine)

		// BL pin
		pin := canvas.NewCircle(colorPin)
		pin.Resize(fyne.NewSize(8, 8))
		pin.Move(fyne.NewPos(x-4, lc.margin+arrayHeight+15))
		objects = append(objects, pin)

		// BL label
		label := canvas.NewText(fmt.Sprintf("BL[%d]", col), colorText)
		label.TextSize = 14
		label.Move(fyne.NewPos(x-12, lc.margin+arrayHeight+28))
		objects = append(objects, label)
	}

	// Source Lines for 1T1R (vertical, offset from BL)
	if is1T1R {
		for col := 0; col < cfg.Cols; col++ {
			x := lc.margin + float32(col)*lc.cellWidth + lc.cellWidth/2 + 8

			// SL wire (dashed effect via multiple short lines)
			for i := 0; i < int(arrayHeight)/8; i++ {
				y1 := lc.margin + float32(i*8)
				y2 := y1 + 4
				if y2 > lc.margin+arrayHeight {
					y2 = lc.margin + arrayHeight
				}
				slLine := canvas.NewLine(colorSL)
				slLine.Position1 = fyne.NewPos(x, y1)
				slLine.Position2 = fyne.NewPos(x, y2)
				slLine.StrokeWidth = 2
				objects = append(objects, slLine)
			}

			// SL pin
			pin := canvas.NewCircle(colorSL)
			pin.Resize(fyne.NewSize(8, 8))
			pin.Move(fyne.NewPos(x-4, lc.margin+arrayHeight+15))
			objects = append(objects, pin)

			// SL label (below BL label)
			label := canvas.NewText(fmt.Sprintf("SL[%d]", col), colorSL)
			label.TextSize = 14
			label.Move(fyne.NewPos(x-10, lc.margin+arrayHeight+42))
			objects = append(objects, label)
		}
	}

	// Cells
	for row := 0; row < cfg.Rows; row++ {
		for col := 0; col < cfg.Cols; col++ {
			x := lc.margin + float32(col)*lc.cellWidth
			y := lc.margin + float32(row)*lc.cellHeight

			if is1T1R {
				// 1T1R: transistor (top) + FeFET (bottom)
				// Transistor
				trans := canvas.NewRectangle(colorTransistor)
				trans.Resize(fyne.NewSize(lc.cellWidth-8, lc.cellHeight/2-4))
				trans.Move(fyne.NewPos(x+4, y+2))
				trans.StrokeColor = colorBL
				trans.StrokeWidth = 1
				objects = append(objects, trans)

				// FeFET
				fefet := canvas.NewRectangle(colorCell1T1R)
				fefet.Resize(fyne.NewSize(lc.cellWidth-8, lc.cellHeight/2-4))
				fefet.Move(fyne.NewPos(x+4, y+lc.cellHeight/2+2))
				fefet.StrokeColor = colorBL
				fefet.StrokeWidth = 1
				objects = append(objects, fefet)

				// Connection line
				connLine := canvas.NewLine(colorText)
				connLine.Position1 = fyne.NewPos(x+lc.cellWidth/2, y+lc.cellHeight/2-2)
				connLine.Position2 = fyne.NewPos(x+lc.cellWidth/2, y+lc.cellHeight/2+2)
				connLine.StrokeWidth = 2
				objects = append(objects, connLine)
			} else {
				// Passive: single FeFET cell
				cell := canvas.NewRectangle(colorCell)
				cell.Resize(fyne.NewSize(lc.cellWidth-8, lc.cellHeight-8))
				cell.Move(fyne.NewPos(x+4, y+4))
				cell.StrokeColor = colorBL
				cell.StrokeWidth = 1
				objects = append(objects, cell)
			}
		}
	}

	// Legend
	legendY := lc.margin + arrayHeight + 55
	if is1T1R {
		legendY += 15
	}

	// WL legend
	wlLegend := canvas.NewLine(colorWL)
	wlLegend.Position1 = fyne.NewPos(lc.margin, legendY)
	wlLegend.Position2 = fyne.NewPos(lc.margin+20, legendY)
	wlLegend.StrokeWidth = 2
	objects = append(objects, wlLegend)
	wlText := canvas.NewText("WL", colorTextDim)
	wlText.TextSize = 14
	wlText.Move(fyne.NewPos(lc.margin+25, legendY-5))
	objects = append(objects, wlText)

	// BL legend
	blLegend := canvas.NewLine(colorBL)
	blLegend.Position1 = fyne.NewPos(lc.margin+60, legendY)
	blLegend.Position2 = fyne.NewPos(lc.margin+80, legendY)
	blLegend.StrokeWidth = 2
	objects = append(objects, blLegend)
	blText := canvas.NewText("BL", colorTextDim)
	blText.TextSize = 14
	blText.Move(fyne.NewPos(lc.margin+85, legendY-5))
	objects = append(objects, blText)

	if is1T1R {
		// SL legend
		slLegend := canvas.NewLine(colorSL)
		slLegend.Position1 = fyne.NewPos(lc.margin+120, legendY)
		slLegend.Position2 = fyne.NewPos(lc.margin+140, legendY)
		slLegend.StrokeWidth = 2
		objects = append(objects, slLegend)
		slText := canvas.NewText("SL", colorTextDim)
		slText.TextSize = 14
		slText.Move(fyne.NewPos(lc.margin+145, legendY-5))
		objects = append(objects, slText)
	}

	return objects
}
