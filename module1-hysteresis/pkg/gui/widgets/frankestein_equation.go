// Package widgets provides custom GUI widgets for the hysteresis visualization.
package widgets

import (
	"encoding/json"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	frankesteinEquationSVGPath     = "shared/assets/equations/frankestein.svg"
	frankesteinEquationHotspotPath = "shared/assets/equations/frankestein.hotspots.json"
)

// TermChip is a small hoverable label that shows a tooltip for a coefficient.
type TermChip struct {
	widget.BaseWidget
	termID   string
	tooltip  string
	label    *widget.Label
	onSelect func(string, string)
}

// NewTermChip creates a new term chip with hover tooltip text.
func NewTermChip(termID, text, tooltip string, onSelect func(string, string)) *TermChip {
	t := &TermChip{
		termID:   termID,
		tooltip:  tooltip,
		onSelect: onSelect,
	}
	t.label = widget.NewLabel(text)
	t.label.TextStyle = fyne.TextStyle{Monospace: true}
	t.ExtendBaseWidget(t)
	return t
}

// CreateRenderer implements fyne.Widget.
func (t *TermChip) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.label)
}

// MouseIn shows the tooltip on hover.
func (t *TermChip) MouseIn(e *desktop.MouseEvent) {
	_ = e
}

// MouseMoved keeps the tooltip near the cursor.
func (t *TermChip) MouseMoved(e *desktop.MouseEvent) {
	_ = e
}

// MouseOut hides the tooltip.
func (t *TermChip) MouseOut() {
}

// Tapped notifies the selected term without showing hover tooltips.
func (t *TermChip) Tapped(_ *fyne.PointEvent) {
	if t.onSelect != nil {
		t.onSelect(t.termID, t.tooltip)
	}
}

// TappedSecondary mirrors tap behavior.
func (t *TermChip) TappedSecondary(_ *fyne.PointEvent) {
	t.Tapped(nil)
}

func mathLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Monospace: true}
	return label
}

// NewFrankesteinEquationWidget builds the equation display with tooltips.
func NewFrankesteinEquationWidget(parent fyne.Window) fyne.CanvasObject {
	if _, err := os.Stat(frankesteinEquationSVGPath); err == nil {
		if widget := newFrankesteinEquationImageWidget(parent, frankesteinEquationSVGPath); widget != nil {
			return widget
		}
	}
	return newFrankesteinEquationTextWidget(parent)
}

func newFrankesteinEquationTextWidget(parent fyne.Window) fyne.CanvasObject {
	detailPanel, detailCard := newTermDetailPanel()
	selectTerm := func(termID, fallback string) {
		detailPanel.SetDetail(termID, fallback)
	}
	infoTabs := buildEquationInfoTabs()

	title := widget.NewLabelWithStyle(
		"Frankestein Equation (Module 1)",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	line1 := container.NewHBox(
		NewTermChip("rho_eff_main", "\\rho_{eff}", "Effective viscosity: intrinsic damping plus series-resistance RC delay.", selectTerm),
		mathLabel(" dP/dt = "),
		NewTermChip("e_applied", "E_{applied}", "Applied electric field drive term (external voltage across the film).", selectTerm),
		mathLabel(" - "),
		NewTermChip("k_dep", "k_{dep}", "Depolarization factor: models interfacial layer; slants the loop for analog states.", selectTerm),
		mathLabel(" P - ("),
	)

	line2 := container.NewHBox(
		NewTermChip("alpha", "2\\alpha", "Dynamic stiffness: temperature + stress dependent curvature of energy wells.", selectTerm),
		mathLabel(" P + "),
		NewTermChip("beta", "4\\beta", "First-order nonlinearity: negative for HZO to create the switching barrier.", selectTerm),
		mathLabel(" P^3 + "),
		NewTermChip("gamma", "6\\gamma", "Sixth-order stabilizer: keeps energy bounded at large polarization.", selectTerm),
		mathLabel(" P^5)"),
	)

	lkRow := container.NewHBox(
		NewTermChip("lk_terms", "LK nonlinearity", "Landau-Khalatnikov nonlinear energy term: 2αP + 4βP^3 + 6γP^5.", selectTerm),
	)

	line3 := container.NewHBox(
		mathLabel("+ "),
		NewTermChip("noise", "\\xi(t)", "Stochastic noise term (optional): captures thermal variability.", selectTerm),
	)

	line4 := container.NewHBox(
		NewTermChip("rho_eff_def", "\\rho_{eff}", "Effective viscosity definition used in the headless hysteresis path.", selectTerm),
		mathLabel(" = "),
		NewTermChip("rho", "\\rho", "Intrinsic viscosity / damping coefficient.", selectTerm),
		mathLabel(" + ("),
		NewTermChip("r_series", "R_{series}", "Series resistance: absorbs RC delay into viscosity.", selectTerm),
		mathLabel(" A) / d"),
	)

	caption := widget.NewLabel("Tap a coefficient or the LK nonlinearity row to see its purpose in Module 1.")
	caption.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		title,
		line1,
		line2,
		lkRow,
		line3,
		line4,
		caption,
		detailCard,
		infoTabs,
	)
}

type hotspotDef struct {
	ID      string  `json:"id"`
	Tooltip string  `json:"tooltip"`
	X       float32 `json:"x"`
	Y       float32 `json:"y"`
	W       float32 `json:"w"`
	H       float32 `json:"h"`
}

type hotspotConfig struct {
	BaseWidth  float32      `json:"base_width"`
	BaseHeight float32      `json:"base_height"`
	Hotspots   []hotspotDef `json:"hotspots"`
}

type normalizedHotspotLayout struct {
	hotspots []hotspotDef
}

func (l *normalizedHotspotLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for i, obj := range objects {
		if i >= len(l.hotspots) {
			break
		}
		spot := l.hotspots[i]
		obj.Move(fyne.NewPos(size.Width*spot.X, size.Height*spot.Y))
		obj.Resize(fyne.NewSize(size.Width*spot.W, size.Height*spot.H))
	}
}

func (l *normalizedHotspotLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}

type Hotspot struct {
	widget.BaseWidget
	termID   string
	tooltip  string
	onSelect func(string, string)
	debug    bool
}

func NewHotspot(termID, tooltip string, debug bool, onSelect func(string, string)) *Hotspot {
	h := &Hotspot{
		termID:   termID,
		tooltip:  tooltip,
		onSelect: onSelect,
		debug:    debug,
	}
	h.ExtendBaseWidget(h)
	return h
}

func (h *Hotspot) CreateRenderer() fyne.WidgetRenderer {
	fill := color.NRGBA{A: 0}
	stroke := color.NRGBA{A: 0}
	if h.debug {
		fill = color.NRGBA{R: 255, G: 0, B: 0, A: 48}
		stroke = color.NRGBA{R: 255, G: 0, B: 0, A: 120}
	}
	rect := canvas.NewRectangle(fill)
	rect.StrokeColor = stroke
	rect.StrokeWidth = 1
	return widget.NewSimpleRenderer(rect)
}

func (h *Hotspot) MouseIn(e *desktop.MouseEvent) {
	_ = e
}

func (h *Hotspot) MouseMoved(e *desktop.MouseEvent) {
	_ = e
}

func (h *Hotspot) MouseOut() {
}

func (h *Hotspot) Tapped(_ *fyne.PointEvent) {
	if h.onSelect != nil {
		h.onSelect(h.termID, h.tooltip)
	}
}

func (h *Hotspot) TappedSecondary(_ *fyne.PointEvent) {
	h.Tapped(nil)
}

func newFrankesteinEquationImageWidget(parent fyne.Window, svgPath string) fyne.CanvasObject {
	detailPanel, detailCard := newTermDetailPanel()
	selectTerm := func(termID, fallback string) {
		detailPanel.SetDetail(termID, fallback)
	}
	infoTabs := buildEquationInfoTabs()

	title := widget.NewLabelWithStyle(
		"Frankestein Equation (Module 1)",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	hotspots, minSize := loadFrankesteinHotspots()
	debug := os.Getenv("FECIM_EQUATION_DEBUG") == "1"

	var hotspotWidgets []fyne.CanvasObject
	for _, spot := range hotspots {
		hotspotWidgets = append(hotspotWidgets, NewHotspot(spot.ID, spot.Tooltip, debug, selectTerm))
	}

	image := loadFrankesteinEquationSVG(svgPath)
	image.FillMode = canvas.ImageFillContain
	if minSize.Width > 0 && minSize.Height > 0 {
		canvasSize := parent.Canvas().Size()
		targetWidth := minSize.Width
		if canvasSize.Width > 0 {
			targetWidth = canvasSize.Width * 0.85
		}
		scale := targetWidth / minSize.Width
		image.SetMinSize(fyne.NewSize(targetWidth, minSize.Height*scale))
	}

	overlay := container.New(&normalizedHotspotLayout{hotspots: hotspots}, hotspotWidgets...)
	stack := container.NewStack(image, overlay)

	caption := widget.NewLabel("Tap a coefficient or the LK nonlinearity row to see its purpose in Module 1.")
	caption.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		title,
		stack,
		caption,
		detailCard,
		infoTabs,
	)
}

func loadFrankesteinEquationSVG(svgPath string) *canvas.Image {
	data, err := os.ReadFile(svgPath)
	if err != nil {
		return canvas.NewImageFromFile(svgPath)
	}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	recolored, err := canvas.RecolorSVG(data, white)
	if err != nil {
		recolored = data
	}
	resource := fyne.NewStaticResource(filepath.Base(svgPath), recolored)
	return canvas.NewImageFromResource(resource)
}

func loadFrankesteinHotspots() ([]hotspotDef, fyne.Size) {
	defaultHotspots, defaultSize := defaultFrankesteinHotspots()
	data, err := os.ReadFile(frankesteinEquationHotspotPath)
	if err != nil {
		return defaultHotspots, defaultSize
	}

	var cfg hotspotConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("failed to parse hotspots file: %v", err)
		return defaultHotspots, defaultSize
	}

	hotspots := defaultHotspots
	if len(cfg.Hotspots) > 0 {
		hotspots = cfg.Hotspots
	}

	size := defaultSize
	if cfg.BaseWidth > 0 && cfg.BaseHeight > 0 {
		size = fyne.NewSize(cfg.BaseWidth, cfg.BaseHeight)
	}

	return hotspots, size
}

func defaultFrankesteinHotspots() ([]hotspotDef, fyne.Size) {
	return []hotspotDef{
		{
			ID:      "rho_eff_main",
			Tooltip: "Effective viscosity: intrinsic damping plus series-resistance RC delay.",
			X:       0.0, Y: 0.1045, W: 0.0657, H: 0.2944,
		},
		{
			ID:      "e_applied",
			Tooltip: "Applied electric field drive term (external voltage across the film).",
			X:       0.166, Y: 0.1045, W: 0.147, H: 0.2944,
		},
		{
			ID:      "k_dep",
			Tooltip: "Depolarization factor: models interfacial layer; slants the loop for analog states.",
			X:       0.3444, Y: 0.1045, W: 0.0789, H: 0.2944,
		},
		{
			ID:      "alpha",
			Tooltip: "Dynamic stiffness: temperature + stress dependent curvature of energy wells.",
			X:       0.5121, Y: 0.1045, W: 0.0922, H: 0.2778,
		},
		{
			ID:      "beta",
			Tooltip: "First-order nonlinearity: negative for HZO to create the switching barrier.",
			X:       0.5964, Y: 0.0661, W: 0.1481, H: 0.3162,
		},
		{
			ID:      "gamma",
			Tooltip: "Sixth-order stabilizer: keeps energy bounded at large polarization.",
			X:       0.7366, Y: 0.0661, W: 0.1376, H: 0.3162,
		},
		{
			ID:      "lk_terms",
			Tooltip: "Landau-Khalatnikov nonlinear energy term: 2αP + 4βP^3 + 6γP^5.",
			X:       0.51, Y: 0.4, W: 0.37, H: 0.12,
		},
		{
			ID:      "noise",
			Tooltip: "Stochastic noise term (optional): captures thermal variability.",
			X:       0.9323, Y: 0.1045, W: 0.0677, H: 0.2778,
		},
		{
			ID:      "rho_eff_def",
			Tooltip: "Effective viscosity definition used in the headless hysteresis path.",
			X:       0.0566, Y: 0.6477, W: 0.0758, H: 0.2944,
		},
		{
			ID:      "rho",
			Tooltip: "Intrinsic viscosity / damping coefficient.",
			X:       0.166, Y: 0.6477, W: 0.0369, H: 0.2778,
		},
		{
			ID:      "r_series",
			Tooltip: "Series resistance: absorbs RC delay into viscosity.",
			X:       0.239, Y: 0.4974, W: 0.0653, H: 0.2944,
		},
	}, fyne.NewSize(1200, 212.1974)
}
