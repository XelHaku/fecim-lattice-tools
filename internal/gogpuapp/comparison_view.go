//go:build !cgo

package gogpuapp

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildComparisonView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	descriptor := snapshot.Descriptor

	children := []widget.Widget{
		primitives.Text(descriptor.Title).FontSize(22).Bold(),
		primitives.Text(descriptor.Description).FontSize(13).Color(theme.Colors.OnSurfaceVariant),
		primitives.Text(string(descriptor.ID) + " | " + descriptor.Status).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
	}
	if descriptor.BoundaryNotice != "" {
		children = append(children, boundaryNotice(descriptor.BoundaryNotice))
	}

	for _, m := range snapshot.Metrics {
		children = append(children, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(14).Bold(),
		).Padding(10).Gap(4).Background(theme.Colors.SurfaceContainer).Rounded(6))
	}

	eSections, rSections, dSections := partitionSections(snapshot.Sections)
	children = appendSectionGroup(children, "Technology Comparison", rSections, widget.Hex(0xEBF5F0), theme)
	children = appendSectionGroup(children, "Education", eSections, widget.Hex(0xE8EEF0), theme)
	children = appendSectionGroup(children, "Design", dSections, widget.Hex(0xF5EEE8), theme)

	return primitives.Box(children...).
		Padding(20).
		Gap(12).
		Background(theme.Colors.Surface)
}
