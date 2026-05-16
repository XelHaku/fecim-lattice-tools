//go:build !cgo

package gogpuapp

import (
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildDocsView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(13).Color(theme.Colors.OnSurfaceVariant),
	}
	if snapshot.Descriptor.BoundaryNotice != "" {
		children = append(children, boundaryNotice(snapshot.Descriptor.BoundaryNotice))
	}

	for _, m := range snapshot.Metrics {
		children = append(children, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(14).Bold(),
		).Padding(10).Gap(4).Background(theme.Colors.SurfaceContainer).Rounded(6))
	}

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
