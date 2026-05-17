//go:build legacy_fyne

package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

type vcLegendSpec struct {
	Title    string
	Min      float64
	Max      float64
	Ticks    []float64
	TickText []string
	SignText string
}

func (ca *CircuitsApp) currentVCLegendSpec() vcLegendSpec {
	maxAbs := 1.0
	readMax := 0.2
	mode := OpModeRead
	if ca != nil && ca.deviceState != nil {
		wr := ca.deviceState.GetWriteRange()
		rr := ca.deviceState.GetReadRange()
		mode = ca.deviceState.GetOperationMode()
		if wr.Max > maxAbs {
			maxAbs = wr.Max
		}
		if rr.Max > 0 {
			readMax = rr.Max
		}
		if mode == OpModeRead || mode == OpModeCompute {
			maxAbs = rr.Max
			if maxAbs <= 0 {
				maxAbs = 1.0
			}
		}
	}
	if maxAbs <= 0 {
		maxAbs = 1.0
	}
	if readMax > maxAbs {
		readMax = maxAbs
	}
	if readMax < 0 {
		readMax = 0
	}

	_ = readMax
	_ = mode

	return vcLegendSpec{
		Title:    "Cell Voltage (V)",
		Min:      -maxAbs,
		Max:      maxAbs,
		Ticks:    []float64{-maxAbs, 0, maxAbs},
		TickText: []string{"-Vmax", "0", "+Vmax"},
		SignText: "+ = BL>WL", // negative implies WL>BL
	}
}

func vcOverlayColor(voltage, maxAbs float64) color.RGBA {
	if maxAbs <= 0 {
		maxAbs = 1.0
	}
	n := voltage / maxAbs
	if n > 1 {
		n = 1
	}
	if n < -1 {
		n = -1
	}

	cool := color.RGBA{70, 130, 255, 255}     // blue: negative voltage
	neutral := color.RGBA{255, 255, 255, 255} // white: zero voltage
	warm := color.RGBA{255, 80, 80, 255}      // red: positive voltage

	if n < 0 {
		return lerpRGBA(neutral, cool, -n)
	}
	return lerpRGBA(neutral, warm, n)
}

func lerpRGBA(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	mix := func(x, y uint8) uint8 {
		return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t))
	}
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}

func blendRGBA(base, overlay color.RGBA, alpha float64) color.RGBA {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	mix := func(a, b uint8) uint8 {
		return uint8(math.Round((1-alpha)*float64(a) + alpha*float64(b)))
	}
	return color.RGBA{mix(base.R, overlay.R), mix(base.G, overlay.G), mix(base.B, overlay.B), 255}
}

func drawVCLegend(img *image.RGBA, x, y, w, h int, spec vcLegendSpec) {
	if w < 40 || h < 6 {
		return
	}

	drawSimpleText(img, spec.Title, x, y-12, color.RGBA{220, 225, 240, 230})

	for i := 0; i < w; i++ {
		v := spec.Min + (spec.Max-spec.Min)*(float64(i)/float64(w-1))
		c := vcOverlayColor(v, math.Max(math.Abs(spec.Min), math.Abs(spec.Max)))
		drawRect(img, x+i, y, 1, h, c)
	}
	drawRectBorder(img, x, y, w, h, color.RGBA{210, 210, 220, 230})

	for i := range spec.Ticks {
		tv := spec.Ticks[i]
		if spec.Max == spec.Min {
			continue
		}
		n := (tv - spec.Min) / (spec.Max - spec.Min)
		tx := x + int(n*float64(w-1))
		drawRect(img, tx, y-2, 1, h+4, color.RGBA{255, 255, 255, 200})
		if i < len(spec.TickText) {
			label := spec.TickText[i]
			drawSimpleText(img, label, tx-len(label)*3, y+h+2, color.RGBA{200, 210, 230, 210})
		}
	}

	drawSimpleText(img, fmt.Sprintf("%+.2fV", spec.Min), x, y+h+12, color.RGBA{140, 180, 255, 220})
	right := fmt.Sprintf("%+.2fV", spec.Max)
	drawSimpleText(img, right, x+w-len(right)*6, y+h+12, color.RGBA{255, 160, 140, 220})
	drawSimpleText(img, spec.SignText, x, y+h+22, color.RGBA{170, 190, 220, 210})
}
