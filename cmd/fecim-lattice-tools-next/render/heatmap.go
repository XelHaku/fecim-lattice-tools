//go:build !cgo

package render

import (
	"fmt"
	"math"

	"github.com/gogpu/gg"
)

// HeatmapConfig describes a 2D heatmap to render.
type HeatmapConfig struct {
	Data      [][]float64
	X         float64
	Y         float64
	CellSize  float64
	MinVal    float64
	MaxVal    float64
	Title     string
}

// DrawHeatmap renders a 2D heatmap into the gg.Context.
func DrawHeatmap(dc *gg.Context, cfg HeatmapConfig) {
	if len(cfg.Data) == 0 || len(cfg.Data[0]) == 0 {
		return
	}
	rows := len(cfg.Data)
	cols := len(cfg.Data[0])
	if cfg.CellSize <= 0 {
		cfg.CellSize = 20
	}
	if cfg.MinVal == 0 && cfg.MaxVal == 0 {
		cfg.MinVal = math.MaxFloat64
		cfg.MaxVal = -math.MaxFloat64
		for r := range rows {
			for c := range cols {
				v := cfg.Data[r][c]
				if v < cfg.MinVal {
					cfg.MinVal = v
				}
				if v > cfg.MaxVal {
					cfg.MaxVal = v
				}
			}
		}
	}

	totalW := float64(cols) * cfg.CellSize
	totalH := float64(rows) * cfg.CellSize

	dc.Push()
	dc.Translate(cfg.X, cfg.Y)

	// Background
	dc.SetHexColor("#1A1C1A")
	dc.DrawRectangle(0, 0, totalW, totalH)
	dc.Fill()

	valRange := cfg.MaxVal - cfg.MinVal
	if valRange == 0 {
		valRange = 1
	}

	for r := range rows {
		for c := range cols {
			v := cfg.Data[r][c]
			t := (v - cfg.MinVal) / valRange
			t = clamp(t, 0, 1)

			// Color gradient: blue (low) → green (mid) → yellow (high)
			rCol, gCol, bCol := heatmapColor(t)

			dc.SetRGBA(float64(rCol)/255, float64(gCol)/255, float64(bCol)/255, 1)
			dc.DrawRectangle(float64(c)*cfg.CellSize, float64(r)*cfg.CellSize,
				cfg.CellSize-1, cfg.CellSize-1)
			dc.Fill()
		}
	}

	// Labels
	dc.SetRGBA(0.7, 0.7, 0.7, 1)
	for i := range rows {
		label := fmt.Sprintf("R%d", i)
		y := float64(i)*cfg.CellSize + cfg.CellSize/2
		dc.DrawStringAnchored(label, -5, y, 1, 0.5)
	}
	for i := range cols {
		label := fmt.Sprintf("C%d", i)
		x := float64(i)*cfg.CellSize + cfg.CellSize/2
		dc.DrawStringAnchored(label, x, totalH+12, 0.5, 0)
	}

	dc.SetRGBA(0.9, 0.9, 0.9, 1)
	dc.DrawStringAnchored(cfg.Title, totalW/2, -18, 0.5, 0)

	dc.Pop()
}

func heatmapColor(t float64) (r, g, b uint8) {
	t = clamp(t, 0, 1)
	if t < 0.33 {
		s := t / 0.33
		return uint8(41 + s*49), uint8(93 + s*93), uint8(80 + s*76)
	}
	if t < 0.66 {
		s := (t - 0.33) / 0.33
		return uint8(90 + s*165), uint8(186 - s*34), uint8(156 - s*116)
	}
	s := (t - 0.66) / 0.34
	return uint8(255), uint8(152 + s*103), uint8(40 - s*40)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
