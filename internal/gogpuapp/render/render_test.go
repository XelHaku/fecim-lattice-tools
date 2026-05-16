//go:build !cgo

package render

import (
	"testing"

	"fecim-lattice-tools/internal/gogpuapp/design"

	"github.com/gogpu/gg"
)

func TestDrawPlot_NonNil(t *testing.T) {
	data := design.NewPlotData("Test", "X", "Y")
	data.AddSeries("line", []design.PlotPoint{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0}})

	dc := gg.NewContext(400, 300)
	defer dc.Close()

	DrawPlot(dc, PlotConfig{
		Data: data, X: 0, Y: 0, Width: 400, Height: 300,
	})
}

func TestDrawPlot_EmptySeries(t *testing.T) {
	data := design.NewPlotData("Empty", "X", "Y")
	dc := gg.NewContext(100, 100)
	defer dc.Close()

	DrawPlot(dc, PlotConfig{Data: data, X: 0, Y: 0, Width: 100, Height: 100})
}

func TestDrawPlot_ZeroSize(t *testing.T) {
	data := design.NewPlotData("Tiny", "X", "Y")
	data.AddSeries("l", []design.PlotPoint{{X: 0, Y: 0}})
	dc := gg.NewContext(50, 50)
	defer dc.Close()

	DrawPlot(dc, PlotConfig{Data: data, X: 0, Y: 0, Width: 10, Height: 10})
}

func TestDrawPlot_NegativeAxis(t *testing.T) {
	data := design.NewPlotData("Neg", "X", "Y")
	data.AddSeries("l", []design.PlotPoint{{X: -10, Y: -5}, {X: 10, Y: 5}})
	dc := gg.NewContext(400, 300)
	defer dc.Close()

	DrawPlot(dc, PlotConfig{Data: data, X: 0, Y: 0, Width: 400, Height: 300})
}

func TestDrawHeatmap_NonNil(t *testing.T) {
	data := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	dc := gg.NewContext(200, 150)
	defer dc.Close()

	DrawHeatmap(dc, HeatmapConfig{
		Data: data, X: 0, Y: 0, CellSize: 30, Title: "Test",
	})
}

func TestDrawHeatmap_Empty(t *testing.T) {
	dc := gg.NewContext(100, 100)
	defer dc.Close()

	DrawHeatmap(dc, HeatmapConfig{Data: [][]float64{}})
}

func TestDrawHeatmap_SingleCell(t *testing.T) {
	data := [][]float64{{42}}
	dc := gg.NewContext(100, 100)
	defer dc.Close()

	DrawHeatmap(dc, HeatmapConfig{
		Data: data, X: 0, Y: 0, CellSize: 40, Title: "One",
	})
}

func TestHeatmapColor_Range(t *testing.T) {
	for _, tc := range []struct{ t, rMin, rMax, gMin, gMax, bMin, bMax float64 }{
		{0.0, 0, 100, 0, 200, 0, 160},
		{0.5, 50, 200, 100, 200, 20, 120},
		{1.0, 200, 255, 200, 255, 0, 40},
	} {
		r, g, b := heatmapColor(tc.t)
		if float64(r) < tc.rMin || float64(r) > tc.rMax {
			t.Errorf("t=%.1f r=%d", tc.t, r)
		}
		if float64(g) < tc.gMin || float64(g) > tc.gMax {
			t.Errorf("t=%.1f g=%d", tc.t, g)
		}
		if float64(b) < tc.bMin || float64(b) > tc.bMax {
			t.Errorf("t=%.1f b=%d", tc.t, b)
		}
	}
}
