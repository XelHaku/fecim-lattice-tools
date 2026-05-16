//go:build !cgo

package design

import "testing"

func TestPlotData_NewPlotData(t *testing.T) {
	pd := NewPlotData("P-E Loop", "Field (kV/cm)", "Polarization (µC/cm²)")
	if pd.Title != "P-E Loop" {
		t.Errorf("Title = %v, want P-E Loop", pd.Title)
	}
	if pd.XLabel != "Field (kV/cm)" {
		t.Errorf("XLabel = %v", pd.XLabel)
	}
	if len(pd.Series) != 0 {
		t.Errorf("Series length = %v, want 0", len(pd.Series))
	}
}

func TestPlotData_AddSeries(t *testing.T) {
	pd := NewPlotData("Test", "X", "Y")
	pd.AddSeries("line1", []PlotPoint{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0}})
	if len(pd.Series) != 1 {
		t.Errorf("Series length = %v, want 1", len(pd.Series))
	}
	if pd.Series[0].Name != "line1" {
		t.Errorf("Series[0].Name = %v", pd.Series[0].Name)
	}
	if len(pd.Series[0].Points) != 3 {
		t.Errorf("Series[0] point count = %v, want 3", len(pd.Series[0].Points))
	}
}

func TestPlotData_AddSeriesAutoBounds(t *testing.T) {
	pd := NewPlotData("Auto", "X", "Y")
	pd.AddSeries("s", []PlotPoint{{X: -10, Y: 20}, {X: 10, Y: -5}})
	if pd.XMin != -10 || pd.XMax != 10 {
		t.Errorf("X bounds = [%v, %v], want [-10, 10]", pd.XMin, pd.XMax)
	}
	if pd.YMin != -5 || pd.YMax != 20 {
		t.Errorf("Y bounds = [%v, %v], want [-5, 20]", pd.YMin, pd.YMax)
	}
}

func TestPlotData_AddSeriesEmpty(t *testing.T) {
	pd := NewPlotData("Empty", "X", "Y")
	pd.AddSeries("empty", []PlotPoint{})
	if len(pd.Series) != 0 {
		t.Errorf("Series length after empty add = %v, want 0", len(pd.Series))
	}
}

func TestPlotData_MultipleSeries(t *testing.T) {
	pd := NewPlotData("Multi", "X", "Y")
	pd.AddSeries("a", []PlotPoint{{X: 0, Y: 0}})
	pd.AddSeries("b", []PlotPoint{{X: -5, Y: 10}})
	if len(pd.Series) != 2 {
		t.Errorf("Series length = %v, want 2", len(pd.Series))
	}
	if pd.XMin != -5 || pd.XMax != 0 {
		t.Errorf("Multi-series X bounds = [%v, %v], want [-5, 0]", pd.XMin, pd.XMax)
	}
}
