//go:build !cgo

package render

import (
	"math"

	"fecim-lattice-tools/internal/gogpuapp/design"

	"github.com/gogpu/gg"
)

// PlotConfig describes how to render a 2D plot into a gg.Context region.
type PlotConfig struct {
	Data       *design.PlotData
	X          float64
	Y          float64
	Width      float64
	Height     float64
	Background string
	GridColor  string
}

const defaultPlotBack = "#1A1C1A"
const defaultPlotGrid = "#333533"
const defaultPlotMargin = 60.0

// DrawPlot renders a 2D line plot into the gg.Context at the specified region.
func DrawPlot(dc *gg.Context, cfg PlotConfig) {
	if cfg.Data == nil || len(cfg.Data.Series) == 0 {
		return
	}
	if cfg.Background == "" {
		cfg.Background = defaultPlotBack
	}
	if cfg.GridColor == "" {
		cfg.GridColor = defaultPlotGrid
	}

	margin := plotMargin(cfg.Width, cfg.Height)
	plotW := cfg.Width - margin*2
	plotH := cfg.Height - margin*2
	if plotW <= 0 || plotH <= 0 {
		return
	}

	dc.Push()
	dc.Translate(cfg.X, cfg.Y)

	// Background
	dc.SetHexColor(cfg.Background)
	dc.DrawRectangle(0, 0, cfg.Width, cfg.Height)
	dc.Fill()

	dc.Push()
	dc.Translate(margin, margin)

	// Y = 0 axis
	dc.SetHexColor(cfg.GridColor)
	dc.SetLineWidth(1)
	y0 := plotH
	if cfg.Data.YMin < 0 && cfg.Data.YMax > 0 {
		y0 = plotH * (cfg.Data.YMax / (cfg.Data.YMax - cfg.Data.YMin))
	}
	dc.MoveTo(0, y0)
	dc.LineTo(plotW, y0)
	dc.Stroke()

	// X = 0 axis
	if cfg.Data.XMin < 0 && cfg.Data.XMax > 0 {
		x0 := plotW * (-cfg.Data.XMin / (cfg.Data.XMax - cfg.Data.XMin))
		dc.MoveTo(x0, 0)
		dc.LineTo(x0, plotH)
		dc.Stroke()
	}

	// Grid lines
	dc.SetLineWidth(0.5)
	for i := 0.0; i <= 4; i++ {
		y := plotH * i / 4
		dc.MoveTo(0, y)
		dc.LineTo(plotW, y)
		dc.Stroke()
		x := plotW * i / 4
		dc.MoveTo(x, 0)
		dc.LineTo(x, plotH)
		dc.Stroke()
	}

	// Data series
	colors := []string{"#2F5D50", "#58685E", "#6F9C8D", "#BA1A1A", "#1F463C"}
	for si, series := range cfg.Data.Series {
		if len(series.Points) < 2 {
			continue
		}
		seriesColor := colors[si%len(colors)]
		if series.Color != "" {
			seriesColor = series.Color
		}
		dc.SetHexColor(seriesColor)
		dc.SetLineWidth(2)

		xRange := cfg.Data.XMax - cfg.Data.XMin
		yRange := cfg.Data.YMax - cfg.Data.YMin
		if xRange == 0 {
			xRange = 1
		}
		if yRange == 0 {
			yRange = 1
		}

		for i, pt := range series.Points {
			px := plotW * (pt.X - cfg.Data.XMin) / xRange
			py := plotH * (cfg.Data.YMax - pt.Y) / yRange
			if i == 0 {
				dc.MoveTo(px, py)
			} else {
				dc.LineTo(px, py)
			}
		}
		dc.Stroke()

		// Data points as small circles
		dc.SetHexColor(seriesColor)
		for _, pt := range series.Points {
			px := plotW * (pt.X - cfg.Data.XMin) / xRange
			py := plotH * (cfg.Data.YMax - pt.Y) / yRange
			dc.DrawCircle(px, py, 2)
			dc.Fill()
		}
	}

	dc.Pop()

	// Labels
	dc.SetRGBA(0.9, 0.9, 0.9, 1)
	dc.SetLineWidth(1)
	dc.DrawStringAnchored(cfg.Data.XLabel, margin+plotW/2, cfg.Height-10, 0.5, 1)
	yLabelX, yLabelAnchor := yAxisLabelAnchor(margin)
	dc.DrawStringAnchored(cfg.Data.YLabel, yLabelX, margin+plotH/2, yLabelAnchor, 0.5)
	dc.DrawStringAnchored(cfg.Data.Title, margin+plotW/2, 18, 0.5, 0)

	dc.Pop()
}

func yAxisLabelAnchor(margin float64) (x, anchor float64) {
	return margin + 8, 0
}

func plotMargin(width, height float64) float64 {
	margin := defaultPlotMargin
	if height < 260 {
		margin = math.Max(30, height*0.22)
	}
	if width < 420 {
		margin = math.Min(margin, math.Max(28, width*0.12))
	}
	return margin
}

func sineWave(start, stop float64, n int) []design.PlotPoint {
	points := make([]design.PlotPoint, n)
	step := (stop - start) / float64(n-1)
	for i := range n {
		t := start + float64(i)*step
		points[i] = design.PlotPoint{X: t, Y: math.Sin(t)}
	}
	return points
}
