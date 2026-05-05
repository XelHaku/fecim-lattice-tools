//go:build !cgo

package main

import (
	"fecim-lattice-tools/cmd/fecim-lattice-tools-next/design"
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildHysteresisView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(13).Color(theme.Colors.OnSurfaceVariant),
	}

	if snapshot.Descriptor.BoundaryNotice != "" {
		children = append(children, boundaryNotice(snapshot.Descriptor.BoundaryNotice))
	}

	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		unitStr := ""
		if m.Unit != "" {
			unitStr = " " + m.Unit
		}
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value + unitStr).FontSize(16).Bold(),
		).
			Padding(12).
			Gap(4).
			Background(theme.Colors.SurfaceContainer).
			Rounded(6),
		)
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))

	for _, plot := range snapshot.Plots {
		plotData := design.NewPlotData(plot.Title, plot.XLabel, plot.YLabel)
		for _, s := range plot.Series {
			pts := make([]design.PlotPoint, len(s.Points))
			for i, p := range s.Points {
				pts[i] = design.PlotPoint{X: p.X, Y: p.Y}
			}
			plotData.AddSeries(s.Name, pts)
		}
		children = append(children, plotCard(plotData, theme))
	}

	eSections, rSections, dSections := partitionSections(snapshot.Sections)
	children = appendSectionGroup(children, "Education", eSections, widget.Hex(0xE8EEF0), theme)
	children = appendSectionGroup(children, "Research", rSections, widget.Hex(0xEBF5F0), theme)
	children = appendSectionGroup(children, "Design", dSections, widget.Hex(0xF5EEE8), theme)

	actionBoxes := []widget.Widget{}
	for _, action := range snapshot.Actions {
		actionBoxes = append(actionBoxes, primitives.Box(
			primitives.Text(action.Label).FontSize(13).Color(theme.Colors.OnPrimary),
		).
			Padding(10).
			Background(theme.Colors.Primary).
			Rounded(6),
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

func plotCard(data *design.PlotData, theme *material3.Theme) widget.Widget {
	summary := data.Title
	if len(data.Series) > 0 {
		summary += " | " + data.Series[0].Name
	}
	return primitives.Box(
		primitives.Text(data.Title).FontSize(16).Bold(),
		primitives.Text(data.XLabel+" vs "+data.YLabel).FontSize(12).Color(theme.Colors.OnSurfaceVariant),
	).
		Padding(20).
		Gap(8).
		Background(widget.Hex(0xF6FAF7)).
		Rounded(8).
		BorderStyle(1, widget.Hex(0xD4DED8))
}

func partitionSections(sections []viewmodel.Section) (edu, res, des []viewmodel.Section) {
	for _, s := range sections {
		switch s.Category {
		case "education":
			edu = append(edu, s)
		case "research":
			res = append(res, s)
		case "design":
			des = append(des, s)
		default:
			res = append(res, s)
		}
	}
	return
}

func appendSectionGroup(children []widget.Widget, label string, sections []viewmodel.Section, bg widget.Color, theme *material3.Theme) []widget.Widget {
	if len(sections) == 0 {
		return children
	}
	children = append(children, primitives.Text(label).FontSize(15).Bold().Color(widget.Hex(0x24483E)))
	for _, s := range sections {
		children = append(children, primitives.Box(
			primitives.Text(s.Title).FontSize(14).Bold().Color(widget.Hex(0x183D34)),
			primitives.Text(s.Body).FontSize(12).Color(widget.Hex(0x44504B)),
		).
			Padding(12).
			Gap(6).
			Background(bg).
			Rounded(8).
			BorderStyle(1, widget.Hex(0xD4DED8)),
		)
	}
	return children
}

func boundaryNotice(text string) widget.Widget {
	return primitives.Box(
		primitives.Text(text).FontSize(12).Color(widget.Hex(0x5C3B00)),
	).Padding(12).Background(widget.Hex(0xFFF4D8)).Rounded(8).BorderStyle(1, widget.Hex(0xE7C66A))
}
