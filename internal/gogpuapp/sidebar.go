//go:build !cgo

package gogpuapp

import (
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// buildSidebar returns a sidebar widget listing all module descriptors.
func buildSidebar(descriptors []viewmodel.ModuleDescriptor, activeIndex int) widget.Widget {
	items := []widget.Widget{
		primitives.Text("FeCIM Modules").FontSize(14).Bold(),
	}

	for i, d := range descriptors {
		highlight := ""
		if i == activeIndex {
			highlight = " →"
		}
		items = append(items, primitives.Box(
			primitives.Text(highlight+" "+d.Title).FontSize(13),
			primitives.Text(string(d.ID)+" · "+d.Status).FontSize(10),
		).
			Padding(8).
			Gap(2),
		)
	}

	return primitives.Box(items...).
		Padding(16).
		Gap(8)
}

// buildSidebarMaterial returns a themed sidebar widget with Material 3 colors.
func buildSidebarMaterial(descriptors []viewmodel.ModuleDescriptor, activeIndex int, theme *material3.Theme) widget.Widget {
	items := []widget.Widget{
		primitives.Text("FeCIM Modules").FontSize(14).Bold(),
	}

	for i, d := range descriptors {
		bgColor := theme.Colors.Surface
		textColor := theme.Colors.OnSurface
		if i == activeIndex {
			bgColor = theme.Colors.Primary
			textColor = theme.Colors.OnPrimary
		}
		statusColor := theme.Colors.OnSurfaceVariant
		if d.Status == viewmodel.StatusFunctional {
			statusColor = theme.Colors.Primary
		}

		items = append(items, primitives.Box(
			primitives.Text(d.Title).FontSize(13).Color(textColor),
			primitives.Text(string(d.ID)+" · "+d.Status).FontSize(10).Color(statusColor),
		).
			Padding(10).
			Gap(2).
			Background(bgColor),
		)
	}

	return primitives.Box(items...).
		Padding(16).
		Gap(8).
		Background(theme.Colors.SurfaceContainer)
}
