//go:build !cgo

package main

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildCrossbarView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
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
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer).Rounded(6))
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))

	rows, cols := parseDimensions(snapshot)
	children = append(children, primitives.Box(
		primitives.Text(fmt.Sprintf("Crossbar Array (%d×%d)", rows, cols)).FontSize(16).Bold(),
		primitives.Text("Conductance matrix visualized as heatmap grid overlay.").FontSize(12).Color(theme.Colors.OnSurfaceVariant),
	).Padding(20).Gap(8).Background(widget.Hex(0xF6FAF7)).Rounded(8).BorderStyle(1, widget.Hex(0xD4DED8)))

	eSections, rSections, dSections := partitionSections(snapshot.Sections)
	children = appendSectionGroup(children, "Education", eSections, widget.Hex(0xE8EEF0), theme)
	children = appendSectionGroup(children, "Research", rSections, widget.Hex(0xEBF5F0), theme)
	children = appendSectionGroup(children, "Design", dSections, widget.Hex(0xF5EEE8), theme)

	actionBoxes := []widget.Widget{}
	for _, action := range snapshot.Actions {
		actionBoxes = append(actionBoxes, primitives.Box(
			primitives.Text(action.Label).FontSize(13).Color(theme.Colors.OnPrimary),
		).Padding(10).Background(theme.Colors.Primary).Rounded(6))
	}
	if len(actionBoxes) > 0 {
		children = append(children, primitives.Box(actionBoxes...).Gap(8))
	}

	return primitives.Box(children...).Padding(24).Gap(14).Background(theme.Colors.Surface)
}

func parseDimensions(snapshot viewmodel.ModuleSnapshot) (int, int) {
	rows, cols := 4, 4
	for _, m := range snapshot.Metrics {
		switch m.ID {
		case "rows":
			fmt.Sscanf(m.Value, "%d", &rows)
		case "cols":
			fmt.Sscanf(m.Value, "%d", &cols)
		}
	}
	return rows, cols
}
