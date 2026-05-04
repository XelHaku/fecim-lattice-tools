//go:build !cgo

package design

type PlotPoint struct {
	X float64
	Y float64
}

type PlotSeries struct {
	Name   string
	Points []PlotPoint
	Color  string
}

type PlotData struct {
	Title  string
	XLabel string
	YLabel string
	Series []PlotSeries
	XMin   float64
	XMax   float64
	YMin   float64
	YMax   float64
}

func NewPlotData(title, xlabel, ylabel string) *PlotData {
	return &PlotData{Title: title, XLabel: xlabel, YLabel: ylabel}
}

func (pd *PlotData) AddSeries(name string, points []PlotPoint) {
	if len(points) == 0 {
		return
	}
	series := PlotSeries{Name: name, Points: points}
	if len(pd.Series) == 0 {
		pd.XMin, pd.XMax = points[0].X, points[0].X
		pd.YMin, pd.YMax = points[0].Y, points[0].Y
	}
	for _, p := range points {
		if p.X < pd.XMin {
			pd.XMin = p.X
		}
		if p.X > pd.XMax {
			pd.XMax = p.X
		}
		if p.Y < pd.YMin {
			pd.YMin = p.Y
		}
		if p.Y > pd.YMax {
			pd.YMax = p.Y
		}
	}
	pd.Series = append(pd.Series, series)
}
