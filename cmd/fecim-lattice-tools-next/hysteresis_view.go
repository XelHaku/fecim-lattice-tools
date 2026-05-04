//go:build !cgo

package main

import (
	"math"

	"fecim-lattice-tools/cmd/fecim-lattice-tools-next/design"
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildHysteresisView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(14),
	}

	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(16).Bold(),
		).
			Padding(12).
			Gap(4).
			Background(theme.Colors.SurfaceContainer),
		)
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))

	loopPoints := defaultHysteresisLoop()
	plotData := design.NewPlotData("P-E Hysteresis Loop", "Electric Field (kV/cm)", "Polarization (µC/cm²)")
	plotData.AddSeries("P-E", loopPoints)
	children = append(children, plotCard(plotData, theme))

	for _, section := range snapshot.Sections {
		children = append(children, hysteresisMaterialCard(section, theme))
	}

	actionBoxes := []widget.Widget{}
	for _, action := range snapshot.Actions {
		actionBoxes = append(actionBoxes, primitives.Box(
			primitives.Text(action.Label).FontSize(13).Color(theme.Colors.OnPrimary),
		).
			Padding(10).
			Background(theme.Colors.Primary),
		)
	}
	if len(actionBoxes) > 0 {
		children = append(children, primitives.Box(actionBoxes...).Gap(8))
	}

	return primitives.Box(children...).
		Padding(24).
		Gap(14).
		Background(theme.Colors.Surface)
}

func hysteresisMaterialCard(section viewmodel.Section, theme *material3.Theme) widget.Widget {
	return primitives.Box(
		primitives.Text(section.Title).FontSize(15).Bold(),
		primitives.Text(section.Body).FontSize(12),
	).
		Padding(12).
		Gap(4).
		Background(theme.Colors.SurfaceContainer)
}

func plotCard(data *design.PlotData, theme *material3.Theme) widget.Widget {
	summary := data.Title
	if len(data.Series) > 0 {
		summary += " | " + data.Series[0].Name
	}
	return primitives.Box(
		primitives.Text(data.Title).FontSize(16).Bold(),
		primitives.Text(summary).FontSize(12).Color(theme.Colors.OnSurfaceVariant),
	).
		Padding(20).
		Gap(8).
		Background(theme.Colors.SurfaceContainer)
}

func defaultHysteresisLoop() []design.PlotPoint {
	points := make([]design.PlotPoint, 0, 200)
	for i := 0; i < 200; i++ {
		t := float64(i) * 2.0 * math.Pi / 199.0
		field := 3000.0 * math.Sin(t)
		pol := 20.0 * math.Sin(t-math.Pi/6) * (1.0 - 0.3*math.Abs(math.Sin(t)))
		points = append(points, design.PlotPoint{X: field, Y: pol})
	}
	return points
}
