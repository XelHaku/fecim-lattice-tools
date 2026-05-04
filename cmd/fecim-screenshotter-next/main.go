//go:build !cgo

package main

import (
	"log"
	"math"

	"fecim-lattice-tools/cmd/fecim-lattice-tools-next/design"
	"fecim-lattice-tools/cmd/fecim-lattice-tools-next/render"

	"github.com/gogpu/gg"
)

func main() {
	genLoop()
	genHeatmap8()
	genHeatmap16()
	genComparison()
	genMVM()
	log.Print("done — screens generated in screenshots/")
}

func genLoop() {
	log.Print("[1/5] hysteresis-p-e-loop.png")
	render.CaptureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := design.NewPlotData("P-E Hysteresis Loop", "Field (kV/cm)", "Polarization (µC/cm²)")
		data.AddSeries("HZO Default", loop(3000, 20, 0.3))
		render.DrawPlot(dc, render.PlotConfig{
			Data: data, X: 80, Y: 60, Width: 1240, Height: 780, Background: "#FFFFFF",
		})
	}, "hysteresis-p-e-loop.png")
}

func genHeatmap8() {
	log.Print("[2/5] crossbar-heatmap-8x8.png")
	render.CaptureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := make([][]float64, 8)
		for i := range 8 {
			data[i] = make([]float64, 8)
			for j := range 8 {
				data[i][j] = float64(30 - ((i*8+j)%30) + i)
			}
		}
		render.DrawHeatmap(dc, render.HeatmapConfig{
			Data: data, X: 200, Y: 120, CellSize: 60,
			Title: "Crossbar Conductance Matrix (8×8, 30 levels)",
		})
	}, "crossbar-heatmap-8x8.png")
}

func genHeatmap16() {
	log.Print("[3/5] crossbar-heatmap-16x16.png")
	render.CaptureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := make([][]float64, 16)
		for i := range 16 {
			data[i] = make([]float64, 16)
			for j := range 16 {
				data[i][j] = float64(30 - ((i*16+j)%30) + i/2)
			}
		}
		render.DrawHeatmap(dc, render.HeatmapConfig{
			Data: data, X: 120, Y: 100, CellSize: 36,
			Title: "Crossbar Conductance Matrix (16×16, 30 levels)",
		})
	}, "crossbar-heatmap-16x16.png")
}

func genComparison() {
	log.Print("[4/5] hysteresis-material-comparison.png")
	render.CaptureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := design.NewPlotData("P-E Hysteresis — Material Comparison", "Field (kV/cm)", "Polarization (µC/cm²)")
		data.AddSeries("HZO Default", loop(3000, 20, 0.3))
		data.AddSeries("BTO (high Pr)", loop(2500, 30, 0.2))
		data.AddSeries("PZT (low Ec)", loop(1500, 25, 0.15))
		render.DrawPlot(dc, render.PlotConfig{
			Data: data, X: 80, Y: 60, Width: 1240, Height: 780, Background: "#FFFFFF",
		})
	}, "hysteresis-material-comparison.png")
}

func genMVM() {
	log.Print("[5/5] mvm-diagram.png")
	render.CaptureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		// G matrix grid
		for r := 0; r < 8; r++ {
			for c := 0; c < 8; c++ {
				x := 120.0 + float64(c)*40
				y := 180.0 + float64(r)*40
				v := float64(30 - ((r*8+c)%30) + r)
				t := clamp01(v / 30.0)
				rr, gg, bb := heatmap(t)
				dc.SetRGBA(float64(rr)/255, float64(gg)/255, float64(bb)/255, 1)
				dc.DrawRectangle(x, y, 32, 32)
				dc.Fill()
			}
		}
		// V vector
		for r := 0; r < 8; r++ {
			dc.SetRGBA(0.18, 0.36, 0.31, 1)
			dc.DrawRectangle(550, 180+float64(r)*40, 50, 32)
			dc.Fill()
		}
		// Arrow
		dc.SetHexColor("#1A1C1A")
		dc.SetLineWidth(3)
		dc.MoveTo(470, 340)
		dc.LineTo(540, 340)
		dc.Stroke()
		dc.MoveTo(610, 340)
		dc.LineTo(680, 340)
		dc.Stroke()
		// I output
		for r := 0; r < 8; r++ {
			dc.SetRGBA(0.35, 0.53, 0.49, 1)
			dc.DrawRectangle(700, 180+float64(r)*40, 80, 32)
			dc.Fill()
		}
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("G", 392, 120, 0.5, 0)
		dc.DrawStringAnchored("V", 575, 120, 0.5, 0)
		dc.DrawStringAnchored("I = G × V", 740, 120, 0.5, 0)
	}, "mvm-diagram.png")
}

func loop(maxF, maxP, w float64) []design.PlotPoint {
	pts := make([]design.PlotPoint, 200)
	for i := range 200 {
		t := float64(i) * 2 * math.Pi / 199
		pts[i] = design.PlotPoint{X: maxF * math.Sin(t), Y: maxP * math.Sin(t-math.Pi/6) * (1 - w*math.Abs(math.Sin(t)))}
	}
	return pts
}

func clamp01(v float64) float64 {
	if v < 0 { return 0 }
	if v > 1 { return 1 }
	return v
}

func heatmap(t float64) (uint8, uint8, uint8) {
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
