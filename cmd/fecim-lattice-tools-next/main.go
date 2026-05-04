//go:build !cgo

package main

import (
	"log"

	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/render"
	uitheme "github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func main() {
	spec := DefaultAppSpec()
	ports := BuildPlaceholderPorts()

	gpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle(spec.Title).
		WithSize(spec.Width, spec.Height).
		WithContinuousRender(false))

	seed := widget.Hex(0x2F5D50)
	appTheme := uitheme.DefaultLight()
	appTheme.Colors.Primary = seed
	appTheme.Colors.PrimaryDark = widget.Hex(0x1F463C)
	appTheme.Colors.PrimaryLight = widget.Hex(0x6F9C8D)
	materialTheme := material3.New(seed)

	app := uiapp.New(
		uiapp.WithWindowProvider(gpuApp),
		uiapp.WithPlatformProvider(gpuApp),
		uiapp.WithEventSource(gpuApp.EventSource()),
		uiapp.WithTheme(appTheme),
	)
	app.SetRoot(buildRoot(spec, ports, materialTheme))

	var canvas *ggcanvas.Canvas
	gpuApp.OnDraw(func(dc *gogpu.Context) {
		w, h := dc.Width(), dc.Height()
		if w <= 0 || h <= 0 {
			return
		}
		if canvas == nil {
			provider := gpuApp.GPUContextProvider()
			if provider == nil {
				return
			}
			var err error
			canvas, err = ggcanvas.New(provider, w, h)
			if err != nil {
				log.Printf("ggcanvas: %v", err)
				return
			}
		}

		app.Frame()
		cw, ch := canvas.Size()
		if cw != w || ch != h {
			if err := canvas.Resize(w, h); err != nil {
				log.Printf("resize: %v", err)
				return
			}
			cw, ch = w, h
		}

		if err := canvas.Draw(func(cc *gg.Context) {
			cc.SetRGBA(0.96, 0.97, 0.96, 1)
			cc.DrawRectangle(0, 0, float64(cw), float64(ch))
			cc.Fill()
			app.Window().DrawTo(render.NewCanvas(cc, cw, ch))
		}); err != nil {
			log.Printf("draw: %v", err)
			return
		}
		if err := canvas.Render(dc.RenderTarget()); err != nil {
			log.Printf("render: %v", err)
		}
	})
	gpuApp.OnClose(func() { gg.CloseAccelerator() })

	if err := gpuApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func buildRoot(spec AppSpec, ports []viewmodel.ModulePort, theme *material3.Theme) widget.Widget {
	descriptors := make([]viewmodel.ModuleDescriptor, len(ports))
	for i, p := range ports {
		descriptors[i] = p.Descriptor()
	}
	sidebar := buildSidebarMaterial(descriptors, 0, theme)

	children := []widget.Widget{
		primitives.Text(spec.Title).FontSize(22).Bold(),
		primitives.Text("Simulation-first FeCIM design workspace").FontSize(14),
	}

	for _, port := range ports {
		snapshot := port.Snapshot()
		switch snapshot.Descriptor.ID {
		case viewmodel.ModuleComparison:
			children = append(children, buildComparisonView(snapshot, theme))
		case viewmodel.ModuleHysteresis:
			children = append(children, buildHysteresisView(snapshot, theme))
		default:
			children = append(children, moduleCardEnhanced(snapshot, theme))
		}
	}

	content := primitives.Box(children...).Padding(24).Gap(14)
	return primitives.Box(sidebar, content).Gap(0)
}

func moduleCardEnhanced(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	descriptor := snapshot.Descriptor
	body := descriptor.Description
	if len(snapshot.Sections) > 0 && snapshot.Sections[0].Body != "" {
		body = body + "\n" + snapshot.Sections[0].Body
	}
	statusBadge := "PLACEHOLDER"
	badgeColor := theme.Colors.OnSurfaceVariant
	if descriptor.Status == viewmodel.StatusFunctional {
		statusBadge = "FUNCTIONAL"
		badgeColor = theme.Colors.Primary
	}
	return primitives.Box(
		primitives.Box(
			primitives.Text(descriptor.Title).FontSize(18).Bold(),
			primitives.Text(statusBadge).FontSize(11).Color(badgeColor),
		).Gap(8),
		primitives.Text(body).FontSize(14),
	).
		Padding(16).
		Gap(8).
		Background(theme.Colors.SurfaceContainer)
}
