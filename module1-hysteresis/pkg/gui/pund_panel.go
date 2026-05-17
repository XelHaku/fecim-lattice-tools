//go:build legacy_fyne

package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

// pundColors defines the waveform colors: P=blue, U=cyan, N=red, D=magenta.
var pundColors = [4]color.RGBA{
	{0, 80, 255, 255},  // P — blue
	{0, 200, 220, 255}, // U — cyan
	{220, 30, 30, 255}, // N — red
	{200, 0, 200, 255}, // D — magenta
}

// PUNDPanel provides PUND measurement mode visualization with current waveform raster.
type PUNDPanel struct {
	resultsLabel    *widget.Label
	rasterContainer *fyne.Container
	traces          [4][]sharedphysics.PulseSample
	content         fyne.CanvasObject
}

// NewPUNDPanel creates a new PUND measurement panel and auto-runs the simulation.
func NewPUNDPanel() *PUNDPanel {
	p := &PUNDPanel{resultsLabel: widget.NewLabel("Running PUND simulation...")}
	placeholder := newPUNDRaster(p.traces)
	placeholder.SetMinSize(fyne.NewSize(400, 250))
	p.rasterContainer = container.NewStack(placeholder)
	runBtn := widget.NewButton("Run PUND", func() { go p.runPUND() })
	legend := container.NewHBox(
		newColorSwatch(pundColors[0]), widget.NewLabel("P"),
		newColorSwatch(pundColors[1]), widget.NewLabel("U"),
		newColorSwatch(pundColors[2]), widget.NewLabel("N"),
		newColorSwatch(pundColors[3]), widget.NewLabel("D"),
	)
	p.content = container.NewVBox(
		widget.NewLabelWithStyle("PUND Measurement", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("6-pulse protocol: Preset -> P -> U -> N -> D"),
		runBtn, widget.NewSeparator(),
		widget.NewLabel("Current waveforms (I vs time):"),
		legend, p.rasterContainer,
		widget.NewSeparator(), p.resultsLabel,
	)
	go p.runPUND() // auto-run with default HZO parameters
	return p
}

func (p *PUNDPanel) runPUND() {
	// Default HZO parameters
	ec := 3e7  // V/m
	ps := 0.25 // C/m²
	stack := sharedphysics.NewPreisachStack(ec, &sharedphysics.TanhEverett{Ec: ec, Ps: ps, Delta: ec * 0.3})

	result, traces, err := sharedphysics.RunPUNDSimulation(stack, 5*ec, 100e-9, 5e-9, 1e-12)

	if err != nil {
		fyne.Do(func() {
			p.resultsLabel.SetText(fmt.Sprintf("PUND error: %v", err))
		})
		return
	}

	p.traces = traces
	ratio := 0.0
	if result.SwitchingNegative_C != 0 {
		ratio = math.Abs(result.SwitchingPositive_C / result.SwitchingNegative_C)
	}

	text := fmt.Sprintf(
		"QP: %.3e C   QU: %.3e C   QN: %.3e C   QD: %.3e C\n"+
			"Qsw+ (QP-QU): %.3e C\n"+
			"Qsw- (QN-QD): %.3e C\n"+
			"Switching ratio |Qsw+/Qsw-|: %.3f",
		result.QP_C, result.QU_C, result.QN_C, result.QD_C,
		result.SwitchingPositive_C,
		result.SwitchingNegative_C,
		ratio,
	)

	fyne.Do(func() {
		p.resultsLabel.SetText(text)
		raster := newPUNDRaster(p.traces)
		raster.SetMinSize(fyne.NewSize(400, 250))
		p.rasterContainer.Objects = []fyne.CanvasObject{raster}
		p.rasterContainer.Refresh()
	})
}

// newPUNDRaster creates a canvas.Raster that plots the four PUND current waveforms.
// Each pulse is plotted sequentially along the x-axis with its assigned color.
func newPUNDRaster(traces [4][]sharedphysics.PulseSample) *canvas.Raster {
	totalSamples := 0
	for i := range traces {
		totalSamples += len(traces[i])
	}
	if totalSamples == 0 {
		return canvas.NewRaster(func(w, h int) image.Image {
			return pundFillBg(w, h)
		})
	}

	// Flatten traces into global timeline with pulse index
	type sample struct {
		t     float64
		cur   float64
		pulse int
	}
	all := make([]sample, 0, totalSamples)
	tOff := 0.0
	for pi := 0; pi < 4; pi++ {
		for _, s := range traces[pi] {
			all = append(all, sample{tOff + s.TimeS, s.CurrentA, pi})
		}
		if n := len(traces[pi]); n > 0 {
			tOff += traces[pi][n-1].TimeS
		}
	}

	// Axis bounds
	minT, maxT := all[0].t, all[0].t
	minI, maxI := all[0].cur, all[0].cur
	for _, s := range all {
		minT, maxT = math.Min(minT, s.t), math.Max(maxT, s.t)
		minI, maxI = math.Min(minI, s.cur), math.Max(maxI, s.cur)
	}
	rngI := maxI - minI
	if rngI == 0 {
		rngI = 1
	}
	minI -= rngI * 0.1
	maxI += rngI * 0.1
	rngT := maxT - minT
	if rngT == 0 {
		rngT = 1
	}

	// Pulse boundary x-fractions for vertical dividers
	var bounds [3]float64
	bT := 0.0
	for pi := 0; pi < 3; pi++ {
		if n := len(traces[pi]); n > 0 {
			bT += traces[pi][n-1].TimeS
		}
		bounds[pi] = (bT - minT) / rngT
	}

	return canvas.NewRaster(func(w, h int) image.Image {
		img := pundFillBg(w, h)
		grid := color.RGBA{50, 50, 60, 255}
		zero := color.RGBA{80, 80, 90, 255}

		// Zero-current line
		zy := h - 1 - int((0-minI)/(maxI-minI)*float64(h-1))
		if zy >= 0 && zy < h {
			for x := 0; x < w; x++ {
				img.Set(x, zy, zero)
			}
		}
		// Pulse boundary dividers
		for _, f := range bounds {
			bx := int(f * float64(w-1))
			if bx > 0 && bx < w {
				for y := 0; y < h; y++ {
					img.Set(bx, y, grid)
				}
			}
		}

		// Draw waveform line segments
		prevX, prevY := [4]int{}, [4]int{}
		started := [4]bool{}
		clamp := func(v, lo, hi int) int {
			if v < lo {
				return lo
			}
			if v > hi {
				return hi
			}
			return v
		}
		for _, s := range all {
			px := clamp(int((s.t-minT)/rngT*float64(w-1)), 0, w-1)
			py := clamp(h-1-int((s.cur-minI)/(maxI-minI)*float64(h-1)), 0, h-1)
			c := pundColors[s.pulse]
			if started[s.pulse] {
				pundDrawLine(img, prevX[s.pulse], prevY[s.pulse], px, py, c)
			}
			img.Set(px, py, c)
			prevX[s.pulse], prevY[s.pulse] = px, py
			started[s.pulse] = true
		}
		return img
	})
}

// pundFillBg creates a dark-background RGBA image.
func pundFillBg(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{20, 20, 30, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bg)
		}
	}
	return img
}

// pundDrawLine draws a line between two points using Bresenham's algorithm.
func pundDrawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx, dy := x1-x0, y1-y0
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// newColorSwatch creates a small colored rectangle for the legend.
func newColorSwatch(c color.RGBA) fyne.CanvasObject {
	rect := canvas.NewRectangle(c)
	rect.SetMinSize(fyne.NewSize(16, 16))
	return rect
}

// Content returns the panel's Fyne canvas object.
func (p *PUNDPanel) Content() fyne.CanvasObject {
	return p.content
}

// createPUNDPanel creates the PUND measurement card for the controls panel.
func (a *App) createPUNDPanel() fyne.CanvasObject {
	panel := NewPUNDPanel()
	return widget.NewCard("PUND Measurement", "Gold-standard ferroelectric characterization", panel.Content())
}
