//go:build !cgo

package main

import (
	"log"

	"fecim-lattice-tools/cmd/fecim-lattice-tools-next/design"
	fecimrender "fecim-lattice-tools/cmd/fecim-lattice-tools-next/render"
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/primitives"
	uirender "github.com/gogpu/ui/render"
	uitheme "github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

var globalPorts []viewmodel.ModulePort

func main() {
	spec := DefaultAppSpec()
	globalPorts = BuildPlaceholderPorts()

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
	app.SetRoot(buildRoot(spec, globalPorts, materialTheme))

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
			app.Window().DrawTo(uirender.NewCanvas(cc, cw, ch))
			drawModuleOverlays(cc, cw, ch)
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

func drawModuleOverlays(cc *gg.Context, w, h int) {
	for _, port := range globalPorts {
		snapshot := port.Snapshot()
		switch snapshot.Descriptor.ID {
		case viewmodel.ModuleHysteresis:
			for _, plot := range snapshot.Plots {
				if plot.ID == "pe_loop" && len(plot.Series) > 0 {
					pts := make([]design.PlotPoint, len(plot.Series[0].Points))
					for i, p := range plot.Series[0].Points {
						pts[i] = design.PlotPoint{X: p.X, Y: p.Y}
					}
					drawHysteresisOverlay(cc, pts, w, h)
				}
			}
		case viewmodel.ModuleCrossbar:
			drawCrossbarOverlay(cc, 8, 8, w, h)
		}
	}
}

func drawHysteresisOverlay(cc *gg.Context, points []design.PlotPoint, w, h int) {
	if len(points) == 0 {
		return
	}
	plotW := float64(w) - 300
	plotH := float64(h) - 180
	if plotW < 100 || plotH < 100 {
		return
	}
	data := design.NewPlotData("P-E Hysteresis Loop", "Field (kV/cm)", "P (µC/cm²)")
	data.AddSeries("P-E", points)
	fecimrender.DrawPlot(cc, fecimrender.PlotConfig{
		Data: data, X: 260, Y: 100, Width: plotW, Height: plotH,
	})
}

func drawCrossbarOverlay(cc *gg.Context, rows, cols, w, h int) {
	for _, port := range globalPorts {
		if port.Descriptor().ID == viewmodel.ModuleCrossbar {
			snapshot := port.Snapshot()
			for _, plot := range snapshot.Plots {
				if plot.ID == "conductance_matrix" && len(plot.Series) > 0 {
					data := make([][]float64, rows)
					for i := range rows {
						data[i] = make([]float64, cols)
					}
					for _, p := range plot.Series[0].Points {
						r, c := int(p.Y), int(p.X)
						if r >= 0 && r < rows && c >= 0 && c < cols {
							data[r][c] = p.V
						}
					}
					fecimrender.DrawHeatmap(cc, fecimrender.HeatmapConfig{
						Data: data, X: 260, Y: 100, CellSize: 24,
						Title: "Crossbar Conductance Matrix",
					})
					return
				}
			}
			data := make([][]float64, rows)
			for i := range rows {
				data[i] = make([]float64, cols)
				for j := range cols {
					data[i][j] = float64(30 - ((i*cols+j)%30) + i)
				}
			}
			fecimrender.DrawHeatmap(cc, fecimrender.HeatmapConfig{
				Data: data, X: 260, Y: 100, CellSize: 24,
				Title: "Crossbar Conductance Matrix",
			})
			return
		}
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
		primitives.Text("Simulation-first FeCIM design workspace — educational tool, not a validated device measurement platform").FontSize(12).Color(widget.Hex(0x5C3B00)),
		primitives.Box(
			primitives.Text("EDUCATIONAL SIMULATION — Results are model estimates based on published physics (Materlik 2015, Park 2015). Not validated against silicon measurements. See per-module boundary notices for source citations and limitations.").FontSize(11).Color(widget.Hex(0x5C3B00)),
		).Padding(10).Background(widget.Hex(0xFFF4D8)).Rounded(6).BorderStyle(1, widget.Hex(0xE7C66A)),
	}

	for _, port := range ports {
		snapshot := port.Snapshot()
		switch snapshot.Descriptor.ID {
		case viewmodel.ModuleComparison:
			children = append(children, buildComparisonView(snapshot, theme))
		case viewmodel.ModuleHysteresis:
			children = append(children, buildHysteresisView(snapshot, theme))
		case viewmodel.ModuleCrossbar:
			children = append(children, buildCrossbarView(snapshot, theme))
		case viewmodel.ModuleCircuits:
			children = append(children, buildCircuitsView(snapshot, theme))
		case viewmodel.ModuleEDA:
			children = append(children, buildEDAView(snapshot, theme))
		case viewmodel.ModuleMNIST:
			children = append(children, buildMNISTView(snapshot, theme))
		case viewmodel.ModuleDocs:
			children = append(children, buildDocsView(snapshot, theme))
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
