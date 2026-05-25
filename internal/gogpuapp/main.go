//go:build !cgo

package gogpuapp

import (
	"fmt"
	"image"
	"log"
	"sort"

	"fecim-lattice-tools/internal/gogpuapp/design"
	fecimrender "fecim-lattice-tools/internal/gogpuapp/render"
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

type Options struct {
	ActiveModuleID viewmodel.ModuleID
}

func Run(options Options) error {
	model := NewAppModel(options.ActiveModuleID)
	spec := model.Spec
	activePort := model.ActivePort()

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
	var rebuildRoot func()
	var selectModule func(viewmodel.ModuleID)
	var dispatchAction func(viewmodel.Action)
	rebuildRoot = func() {
		app.SetRoot(buildRootWithSelectAndActions(model, materialTheme, selectModule, dispatchAction))
	}
	selectModule = func(id viewmodel.ModuleID) {
		if !model.SelectModule(id) {
			return
		}
		activePort = model.ActivePort()
		rebuildRoot()
	}
	dispatchAction = func(action viewmodel.Action) {
		port := model.ActivePort()
		if port == nil {
			return
		}
		if err := port.ApplyAction(action); err != nil {
			log.Printf("module action %s: %v", action.ID, err)
			return
		}
		activePort = model.ActivePort()
		rebuildRoot()
	}
	rebuildRoot()

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
			drawAppFrame(cc, app, activePort, cw, ch)
		}); err != nil {
			log.Printf("draw: %v", err)
			return
		}
		if err := canvas.Render(dc.RenderTarget()); err != nil {
			log.Printf("render: %v", err)
		}
	})
	gpuApp.OnClose(func() { gg.CloseAccelerator() })

	return gpuApp.Run()
}

func CaptureFrameImage(active viewmodel.ModuleID, w, h int) (image.Image, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("capture frame dimensions must be positive, got %dx%d", w, h)
	}

	model := NewAppModel(active)
	seed := widget.Hex(0x2F5D50)
	appTheme := uitheme.DefaultLight()
	appTheme.Colors.Primary = seed
	appTheme.Colors.PrimaryDark = widget.Hex(0x1F463C)
	appTheme.Colors.PrimaryLight = widget.Hex(0x6F9C8D)
	materialTheme := material3.New(seed)

	app := uiapp.New(uiapp.WithTheme(appTheme))
	app.Window().HandleResize(w, h)
	app.SetRoot(buildRoot(model, materialTheme))
	app.Frame()

	cc := newOffscreenContext(w, h)
	defer cc.Close()
	drawAppFrame(cc, app, model.ActivePort(), w, h)
	if err := cc.FlushGPU(); err != nil {
		return nil, fmt.Errorf("capture frame flush: %w", err)
	}
	return cc.Image(), nil
}

func drawAppFrame(cc *gg.Context, app *uiapp.App, activePort viewmodel.ModulePort, w, h int) {
	cc.SetRGBA(0.96, 0.97, 0.96, 1)
	cc.DrawRectangle(0, 0, float64(w), float64(h))
	cc.Fill()
	if app != nil {
		cc.Push()
		app.Window().DrawTo(uirender.NewCanvas(cc, w, h))
		cc.Pop()
	}
	if activePort != nil {
		drawModuleOverlays(cc, activePort.Snapshot(), w, h)
	}
}

func drawModuleOverlays(cc *gg.Context, snapshot viewmodel.ModuleSnapshot, w, h int) {
	switch snapshot.Descriptor.ID {
	case viewmodel.ModuleHysteresis:
		drawHysteresisOverlayFromSnapshot(cc, snapshot, w, h)
	case viewmodel.ModuleCrossbar:
		drawCrossbarOverlay(cc, snapshot, 8, 8, w, h)
	case viewmodel.ModuleCircuits:
		drawCircuitsOverlay(cc, snapshot, w, h)
	}
}

func drawHysteresisOverlayFromSnapshot(cc *gg.Context, snapshot viewmodel.ModuleSnapshot, w, h int) {
	if plot := snapshotPlotByID(snapshot, "pe_loop"); len(plot.Series) > 0 {
		pts := make([]design.PlotPoint, len(plot.Series[0].Points))
		for i, p := range plot.Series[0].Points {
			pts[i] = design.PlotPoint{X: p.X, Y: p.Y}
		}
		drawHysteresisOverlay(cc, pts, w, h)
	}
	if plot := snapshotPlotByID(snapshot, "pund_current_waveforms"); len(plot.Series) > 0 {
		drawPUNDWaveformOverlay(cc, plot, w, h)
	}
	if plot := snapshotPlotByID(snapshot, "forc_density_heatmap"); len(plot.Series) > 0 {
		drawFORCDensityOverlay(cc, plot, w, h)
	}
}

func drawHysteresisOverlay(cc *gg.Context, points []design.PlotPoint, w, h int) {
	if len(points) == 0 {
		return
	}
	x, y, plotW, plotH := hysteresisOverlayPlotRegion(w, h)
	if plotW < 100 || plotH < 100 {
		return
	}
	data := design.NewPlotData("P-E Hysteresis Loop", "Field (kV/cm)", "P (µC/cm²)")
	data.AddSeries("P-E", points)
	fecimrender.DrawPlot(cc, fecimrender.PlotConfig{
		Data: data, X: x, Y: y, Width: plotW, Height: plotH,
	})
}

func hysteresisOverlayPlotRegion(w, h int) (x, y, width, height float64) {
	fw, fh := float64(w), float64(h)
	x = 300
	if fw < 900 {
		x = 240
	}
	y = 420
	if fh < 800 {
		y = 410
	}
	if fh < 700 {
		y = 400
	}
	if fh < 520 {
		y = 260
	}
	width = fw - x - 40
	if width > 940 {
		width = 940
	}
	height = fh - y - 60
	if height > 440 {
		height = 440
	}
	return x, y, width, height
}

func compactWorkspaceSubtitle() string {
	return "Simulation-first FeCIM workspace — educational, not a device claim."
}

func compactSimulationNotice() string {
	return "EDUCATIONAL SIMULATION — Published-physics estimates; not silicon measurements."
}

func drawPUNDWaveformOverlay(cc *gg.Context, plot viewmodel.PlotData, w, h int) {
	data := designPlotFromSnapshotPlot(plot)
	if data == nil || len(data.Series) == 0 {
		return
	}
	width := minFloat(520, float64(w)-320)
	height := 165.0
	if width < 220 || float64(h) < 360 {
		return
	}
	fecimrender.DrawPlot(cc, fecimrender.PlotConfig{
		Data:       data,
		X:          280,
		Y:          float64(h) - height - 44,
		Width:      width,
		Height:     height,
		Background: "#14141E",
		GridColor:  "#32323C",
	})
}

func drawFORCDensityOverlay(cc *gg.Context, plot viewmodel.PlotData, w, h int) {
	matrix := densityMatrixFromPlot(plot)
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return
	}
	cell := clampFloat((float64(h)-240)/float64(len(matrix)), 6, 16)
	width := float64(len(matrix[0])) * cell
	x := float64(w) - width - 52
	if x < 300 {
		x = 300
	}
	fecimrender.DrawHeatmap(cc, fecimrender.HeatmapConfig{
		Data:     matrix,
		X:        x,
		Y:        130,
		CellSize: cell,
		Title:    plot.Title,
	})
}

func drawCrossbarOverlay(cc *gg.Context, snapshot viewmodel.ModuleSnapshot, rows, cols, w, h int) {
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
			data[i][j] = float64(30 - ((i*cols + j) % 30) + i)
		}
	}
	fecimrender.DrawHeatmap(cc, fecimrender.HeatmapConfig{
		Data: data, X: 260, Y: 100, CellSize: 24,
		Title: "Crossbar Conductance Matrix",
	})
}

func snapshotPlotByID(snapshot viewmodel.ModuleSnapshot, id string) viewmodel.PlotData {
	for _, plot := range snapshot.Plots {
		if plot.ID == id {
			return plot
		}
	}
	return viewmodel.PlotData{}
}

func designPlotFromSnapshotPlot(plot viewmodel.PlotData) *design.PlotData {
	data := design.NewPlotData(plot.Title, plot.XLabel, plot.YLabel)
	for _, series := range plot.Series {
		points := make([]design.PlotPoint, len(series.Points))
		for i, point := range series.Points {
			points[i] = design.PlotPoint{X: point.X, Y: point.Y}
		}
		data.AddSeries(series.Name, points)
	}
	if len(data.Series) == 0 {
		return nil
	}
	return data
}

func densityMatrixFromPlot(plot viewmodel.PlotData) [][]float64 {
	if len(plot.Series) == 0 || len(plot.Series[0].Points) == 0 {
		return nil
	}
	xs := sortedUniquePlotValues(plot.Series[0].Points, func(point viewmodel.PlotPoint) float64 { return point.X })
	ys := sortedUniquePlotValues(plot.Series[0].Points, func(point viewmodel.PlotPoint) float64 { return point.Y })
	if len(xs) == 0 || len(ys) == 0 {
		return nil
	}
	xIndex := make(map[float64]int, len(xs))
	for i, x := range xs {
		xIndex[x] = i
	}
	yIndex := make(map[float64]int, len(ys))
	for i, y := range ys {
		yIndex[y] = i
	}
	matrix := make([][]float64, len(ys))
	for i := range matrix {
		matrix[i] = make([]float64, len(xs))
	}
	for _, point := range plot.Series[0].Points {
		row, rowOK := yIndex[point.Y]
		col, colOK := xIndex[point.X]
		if rowOK && colOK {
			matrix[row][col] = point.V
		}
	}
	return matrix
}

func sortedUniquePlotValues(points []viewmodel.PlotPoint, value func(viewmodel.PlotPoint) float64) []float64 {
	seen := map[float64]struct{}{}
	for _, point := range points {
		seen[value(point)] = struct{}{}
	}
	values := make([]float64, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sort.Float64s(values)
	return values
}

func buildRoot(model AppModel, theme *material3.Theme) widget.Widget {
	return buildRootWithSelect(model, theme, nil)
}

func buildRootWithSelect(model AppModel, theme *material3.Theme, onSelect func(viewmodel.ModuleID)) widget.Widget {
	return buildRootWithSelectAndActions(model, theme, onSelect, nil)
}

func buildRootWithSelectAndActions(model AppModel, theme *material3.Theme, onSelect func(viewmodel.ModuleID), onAction func(viewmodel.Action)) widget.Widget {
	descriptors := make([]viewmodel.ModuleDescriptor, len(model.Ports))
	for i, p := range model.Ports {
		descriptors[i] = p.Descriptor()
	}
	sidebar := buildSidebarMaterialWithSelect(descriptors, model.ActiveIndex, theme, onSelect)

	children := []widget.Widget{
		primitives.Text(model.Spec.Title).FontSize(22).Bold(),
		primitives.Text(compactWorkspaceSubtitle()).FontSize(12).Color(widget.Hex(0x5C3B00)),
		primitives.Box(
			primitives.Text(compactSimulationNotice()).FontSize(11).Color(widget.Hex(0x5C3B00)),
		).Padding(10).Background(widget.Hex(0xFFF4D8)).Rounded(6).BorderStyle(1, widget.Hex(0xE7C66A)),
	}

	if port := model.ActivePort(); port != nil {
		snapshot := port.Snapshot()
		switch snapshot.Descriptor.ID {
		case viewmodel.ModuleComparison:
			children = append(children, buildComparisonView(snapshot, theme))
		case viewmodel.ModuleHysteresis:
			children = append(children, buildHysteresisViewWithActions(snapshot, theme, onAction))
		case viewmodel.ModuleCrossbar:
			children = append(children, buildCrossbarView(snapshot, theme))
		case viewmodel.ModuleCircuits:
			children = append(children, buildCircuitsViewWithActions(snapshot, theme, onAction))
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
	return primitives.HBox(sidebar, content).Gap(0)
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
