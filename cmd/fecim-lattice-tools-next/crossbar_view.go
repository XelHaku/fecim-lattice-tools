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

	rows, cols := parseDimensions(snapshot)
	children = append(children, primitives.Box(
		primitives.Text(fmt.Sprintf("Crossbar Array (%d×%d)", rows, cols)).FontSize(16).Bold(),
		primitives.Text("Heatmap visualization rendering as grid preview.").FontSize(12),
	).
		Padding(20).
		Gap(8).
		Background(theme.Colors.SurfaceContainer),
	)

	for _, section := range snapshot.Sections {
		children = append(children, primitives.Box(
			primitives.Text(section.Title).FontSize(15).Bold(),
			primitives.Text(section.Body).FontSize(12),
		).
			Padding(12).
			Gap(4).
			Background(theme.Colors.SurfaceContainer),
		)
	}

	return primitives.Box(children...).
		Padding(24).
		Gap(14).
		Background(theme.Colors.Surface)
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
