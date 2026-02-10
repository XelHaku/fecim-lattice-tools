// Package widgets provides custom GUI widgets for the hysteresis visualization.
package widgets

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	eqassets "fecim-lattice-tools/shared/assets/equations"
)

// Equation asset identifiers (used as cache keys).
const (
	lkEquationID       = "lk"
	preisachEquationID = "preisach"
)

var (
	equationSVGCacheMu  sync.Mutex
	equationSVGCache    = map[string]*equationSVGCacheEntry{}
	equationPerfEnabled = os.Getenv("FECIM_EQUATION_PERF") == "1"

	lkHotspotsOnce sync.Once
	cachedLkSpots  []hotspotDef
	cachedLkSize   fyne.Size
)

type equationSVGCacheEntry struct {
	once sync.Once
	res  fyne.Resource
	ok   bool
}

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

func logEquationPerf(format string, args ...interface{}) {
	if !equationPerfEnabled {
		return
	}
	log.Printf(format, args...)
}

func equationThemeKey() string {
	c := theme.ForegroundColor()
	r, g, b, a := c.RGBA()
	// RGBA() returns 16-bit components.
	return fmt.Sprintf("%02x%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
}

func newEquationLoadingSlot(message string) (*fyne.Container, *widget.ProgressBarInfinite) {
	label := widget.NewLabel(message)
	label.TextStyle = fyne.TextStyle{Italic: true}
	bar := widget.NewProgressBarInfinite()
	bar.Start()
	return container.NewVBox(label, bar), bar
}

func swapEquationSlotContent(slot *fyne.Container, bar *widget.ProgressBarInfinite, content fyne.CanvasObject) {
	if bar != nil {
		bar.Stop()
	}
	slot.Objects = []fyne.CanvasObject{content}
	slot.Refresh()
}

// NewPhysicsEquationsWidget builds the equation display with tooltips.
// The first (visible) tab is built synchronously; the other two are
// constructed lazily on first selection so the dialog opens fast.
func NewPhysicsEquationsWidget(parent fyne.Window) fyne.CanvasObject {
	type lazyTab struct {
		slot    *fyne.Container
		bar     *widget.ProgressBarInfinite
		loaded  bool
		builder func() fyne.CanvasObject
	}

	// First tab: build synchronously (user sees it immediately).
	lkContent := buildLkEquationTab(parent)
	lk := &lazyTab{
		slot:   container.NewStack(lkContent),
		loaded: true,
	}

	// Other tabs: deferred until selected.
	makeDeferred := func(builder func() fyne.CanvasObject) *lazyTab {
		slot, bar := newEquationLoadingSlot("Loading…")
		return &lazyTab{slot: slot, bar: bar, builder: builder}
	}
	preisach := makeDeferred(func() fyne.CanvasObject { return buildPreisachEquationTab(parent) })
	ispp := makeDeferred(func() fyne.CanvasObject { return buildIsppControllerTab(parent) })

	lazyTabs := []*lazyTab{lk, preisach, ispp}

	tabs := container.NewAppTabs(
		container.NewTabItem("L-K (dynamic)", lk.slot),
		container.NewTabItem("Preisach (quasi-static)", preisach.slot),
		container.NewTabItem("ISPP / WRD", ispp.slot),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	tabs.OnSelected = func(item *container.TabItem) {
		idx := -1
		for i, t := range tabs.Items {
			if t == item {
				idx = i
				break
			}
		}
		if idx < 0 || idx >= len(lazyTabs) {
			return
		}
		lt := lazyTabs[idx]
		if lt.loaded {
			return
		}
		lt.loaded = true
		go func() {
			content := lt.builder()
			fyne.Do(func() {
				swapEquationSlotContent(lt.slot, lt.bar, content)
			})
		}()
	}

	return tabs
}

func buildLkEquationTab(parent fyne.Window) fyne.CanvasObject {
	detailPanel, detailCard := newTermDetailPanel()

	infoTabs, termTab := buildLkInfoTabsWithDetail(detailCard)

	selectTerm := func(termID, fallback string) {
		detailPanel.SetDetail(termID, fallback)
		infoTabs.Select(termTab) // auto-switch to Selected Term
	}

	eqPanel := buildLkEquationPanel(parent, selectTerm)

	eqScroll := container.NewScroll(eqPanel)
	eqScroll.SetMinSize(fyne.NewSize(240, 120))

	title := widget.NewLabelWithStyle(
		"Landau-Khalatnikov Equation (Module 1)",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	top := container.NewBorder(title, nil, nil, nil, eqScroll)
	split := container.NewVSplit(top, infoTabs)
	split.Offset = 0.45
	return split
}

func buildPreisachEquationTab(parent fyne.Window) fyne.CanvasObject {
	detailPanel, detailCard := newTermDetailPanel()

	infoTabs, termTab := buildPreisachInfoTabsWithDetail(detailCard)

	selectTerm := func(termID, fallback string) {
		detailPanel.SetDetail(termID, fallback)
		infoTabs.Select(termTab)
	}

	eqPanel := buildPreisachEquationPanel(parent, selectTerm)

	eqScroll := container.NewScroll(eqPanel)
	eqScroll.SetMinSize(fyne.NewSize(240, 120))

	title := widget.NewLabelWithStyle(
		"Preisach Model (Module 1)",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	top := container.NewBorder(title, nil, nil, nil, eqScroll)
	split := container.NewVSplit(top, infoTabs)
	split.Offset = 0.45
	return split
}

func buildIsppControllerTab(parent fyne.Window) fyne.CanvasObject {
	detailPanel, detailCard := newTermDetailPanel()

	infoTabs, termTab := buildIsppInfoTabsWithDetail(detailCard)

	selectTerm := func(termID, fallback string) {
		detailPanel.SetDetail(termID, fallback)
		infoTabs.Select(termTab)
	}

	// Simple, code-faithful controller summary. This is intentionally not a single
	// closed-form equation; it is a state machine with bounds/bisection.
	line1 := container.NewHBox(
		mathLabel("Goal: write discrete level L* such that "),
		NewTermChip("ispp_verify", "L_read == L_target", "VERIFY compares the read-back discrete level against the active controller target.", selectTerm),
	)
	line2 := container.NewHBox(
		NewTermChip("ispp_bounds", "Bounds", "Maintain Vmin<Vmax as a bracket around the target; bisection uses the bracket when available.", selectTerm),
		mathLabel(": Vmin < Vpulse < Vmax"),
	)
	line3 := container.NewHBox(
		NewTermChip("ispp_step", "Step", "Adaptive step-size: increases when stuck/no-improve; shrinks near target to avoid overshoot.", selectTerm),
		mathLabel(": V_{k+1} = clamp(V_k ± ΔV, [Vmin,Vmax])"),
	)
	line4 := container.NewHBox(
		NewTermChip("ispp_overshoot", "Overshoot", "If we cross past the target level, reset bounds/state and apply a short reverse correction.", selectTerm),
	)
	caption := widget.NewLabel("See code for exact states/transitions: module1-hysteresis/pkg/controller + headless runner in cmd/fecim-lattice-tools/mode.go")
	caption.Wrapping = fyne.TextWrapWord
	caption.TextStyle = fyne.TextStyle{Italic: true}

	eqPanel := container.NewVBox(
		widget.NewLabelWithStyle("ISPP / Write-Read-Demo Controller", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		line1,
		line2,
		line3,
		line4,
		caption,
	)

	eqScroll := container.NewScroll(eqPanel)
	eqScroll.SetMinSize(fyne.NewSize(240, 120))

	split := container.NewVSplit(eqScroll, infoTabs)
	split.Offset = 0.35
	return split
}

func buildLkEquationPanel(parent fyne.Window, selectTerm func(string, string)) fyne.CanvasObject {
	textPanel := buildLkEquationTextPanel(selectTerm, false)
	textContainer, _ := textPanel.(*fyne.Container)
	caption := widget.NewLabel("Tap a coefficient or the LK nonlinearity row to see its purpose in Module 1.")
	caption.TextStyle = fyne.TextStyle{Italic: true}

	imageSlot, bar := newEquationLoadingSlot("Loading L-K equation diagram...")

	res, ok := loadEquationSVGResource(lkEquationID)
	if !ok {
		swapEquationSlotContent(imageSlot, bar, widget.NewLabel("Equation SVG unavailable; showing text-only equation."))
		if textContainer != nil {
			textContainer.Objects = append(textContainer.Objects, caption)
			textContainer.Refresh()
		}
		return container.NewVBox(imageSlot, textPanel)
	}

	hotspots, minSize := loadLkHotspots()
	panel := buildLkEquationImagePanel(parent, selectTerm, res, hotspots, minSize)
	if panel == nil {
		swapEquationSlotContent(imageSlot, bar, widget.NewLabel("Equation SVG unavailable; showing text-only equation."))
		if textContainer != nil {
			textContainer.Objects = append(textContainer.Objects, caption)
			textContainer.Refresh()
		}
		return container.NewVBox(imageSlot, textPanel)
	}
	swapEquationSlotContent(imageSlot, bar, panel)

	// SVG loaded — show image only; text panel is redundant.
	return imageSlot
}

func buildLkEquationTextPanel(selectTerm func(string, string), withCaption bool) fyne.CanvasObject {
	line1 := container.NewHBox(
		NewTermChip("rho_eff_main", "ρ_eff", "Effective viscosity: intrinsic damping plus series-resistance RC delay.", selectTerm),
		mathLabel(" dP/dt = "),
		NewTermChip("e_applied", "E_applied", "Applied electric field drive term (external voltage across the film).", selectTerm),
		mathLabel(" - "),
		NewTermChip("k_dep", "k_dep", "Depolarization factor: models interfacial layer; slants the loop for analog states.", selectTerm),
		mathLabel(" P - ("),
	)

	line2 := container.NewHBox(
		NewTermChip("alpha", "2α", "Dynamic stiffness: temperature + stress dependent curvature of energy wells.", selectTerm),
		mathLabel(" P + "),
		NewTermChip("beta", "4β", "First-order nonlinearity: negative for HZO to create the switching barrier.", selectTerm),
		mathLabel(" P³ + "),
		NewTermChip("gamma", "6γ", "Sixth-order stabilizer: keeps energy bounded at large polarization.", selectTerm),
		mathLabel(" P⁵)"),
	)

	lkRow := container.NewHBox(
		NewTermChip("lk_terms", "LK nonlinearity", "Landau-Khalatnikov nonlinear energy term: 2αP + 4βP³ + 6γP⁵.", selectTerm),
	)

	line3 := container.NewHBox(
		mathLabel("+ "),
		NewTermChip("noise", "ξ(t)", "Stochastic noise term (optional): captures thermal variability.", selectTerm),
	)

	line4 := container.NewHBox(
		NewTermChip("rho_eff_def", "ρ_eff", "Effective viscosity definition used in the headless hysteresis path.", selectTerm),
		mathLabel(" = "),
		NewTermChip("rho", "ρ", "Intrinsic viscosity / damping coefficient.", selectTerm),
		mathLabel(" + ("),
		NewTermChip("r_series", "R_series", "Series resistance: absorbs RC delay into viscosity.", selectTerm),
		mathLabel(" A) / d"),
	)

	objects := []fyne.CanvasObject{
		line1,
		line2,
		lkRow,
		line3,
		line4,
	}

	if withCaption {
		caption := widget.NewLabel("Tap a coefficient or the LK nonlinearity row to see its purpose in Module 1.")
		caption.TextStyle = fyne.TextStyle{Italic: true}
		objects = append(objects, caption)
	}

	return container.NewVBox(objects...)
}

func buildEquationDiagramWithZoom(parent fyne.Window, stack fyne.CanvasObject, baseSize fyne.Size, aspectFallback float32) fyne.CanvasObject {
	// baseSize is the "natural" SVG size (used for aspect ratio). When unavailable,
	// fall back to aspectFallback = height/width.
	zoom := float32(1.0)

	// update sizes based on window width; on narrow screens use more width.
	update := func() {
		canvasW := float32(0)
		if parent != nil {
			canvasW = parent.Canvas().Size().Width
		}
		if canvasW <= 0 {
			canvasW = 900
		}
		frac := float32(0.62)
		if canvasW < 620 {
			frac = 0.92
		}
		targetW := canvasW * frac * zoom
		if targetW < 260 {
			targetW = 260
		}
		var targetH float32
		if baseSize.Width > 0 && baseSize.Height > 0 {
			s := targetW / baseSize.Width
			targetH = baseSize.Height * s
		} else if aspectFallback > 0 {
			targetH = targetW * aspectFallback
		} else {
			targetH = 180
		}

		// Best-effort: if the stack contains a canvas.Image as the first child,
		// set its min size. (LK uses this for correct contain sizing.)
		if c, ok := stack.(*fyne.Container); ok {
			if len(c.Objects) > 0 {
				if img, ok := c.Objects[0].(*canvas.Image); ok {
					img.SetMinSize(fyne.NewSize(targetW, targetH))
				}
			}
		}
	}

	slider := widget.NewSlider(0.7, 2.2)
	slider.Step = 0.05
	slider.Value = float64(zoom)
	slider.OnChanged = func(v float64) {
		zoom = float32(v)
		update()
	}

	minusBtn := widget.NewButton("-", func() {
		z := zoom - 0.1
		if z < 0.7 {
			z = 0.7
		}
		slider.SetValue(float64(z))
	})
	plusBtn := widget.NewButton("+", func() {
		z := zoom + 0.1
		if z > 2.2 {
			z = 2.2
		}
		slider.SetValue(float64(z))
	})

	zoomLabel := widget.NewLabel("Zoom")
	zoomLabel.TextStyle = fyne.TextStyle{Bold: true}
	controls := container.NewHBox(zoomLabel, minusBtn, slider, plusBtn)

	bg := canvas.NewRectangle(theme.InputBackgroundColor())
	bg.StrokeColor = theme.ShadowColor()
	bg.StrokeWidth = 1
	framed := container.NewPadded(container.NewStack(bg, stack))

	update()
	return container.NewVBox(controls, framed)
}

func buildLkEquationImagePanel(parent fyne.Window, selectTerm func(string, string), res fyne.Resource, hotspots []hotspotDef, minSize fyne.Size) fyne.CanvasObject {
	debug := os.Getenv("FECIM_EQUATION_DEBUG") == "1"

	var hotspotWidgets []fyne.CanvasObject
	for _, spot := range hotspots {
		hotspotWidgets = append(hotspotWidgets, NewHotspot(spot.ID, spot.Tooltip, debug, selectTerm))
	}

	if res == nil {
		return nil
	}

	image := canvas.NewImageFromResource(res)
	image.FillMode = canvas.ImageFillContain

	overlay := container.New(&normalizedHotspotLayout{hotspots: hotspots}, hotspotWidgets...)
	stack := container.NewStack(image, overlay)

	// LK: we know the SVG base size from the hotspot file, so aspect stays correct.
	return buildEquationDiagramWithZoom(parent, stack, minSize, 0.35)
}

func buildPreisachEquationPanel(parent fyne.Window, selectTerm func(string, string)) fyne.CanvasObject {
	imageSlot, bar := newEquationLoadingSlot("Loading Preisach equation diagram...")
	res, ok := loadEquationSVGResource(preisachEquationID)
	if !ok {
		swapEquationSlotContent(imageSlot, bar, widget.NewLabel("Equation SVG unavailable; showing text-only equation."))
		return container.NewVBox(
			imageSlot,
			buildPreisachEquationTextPanel(selectTerm),
		)
	}
	panel := buildPreisachEquationImagePanel(parent, res)
	swapEquationSlotContent(imageSlot, bar, panel)

	// SVG loaded — show image only; text panel is redundant.
	return imageSlot
}

func buildPreisachEquationImagePanel(parent fyne.Window, res fyne.Resource) fyne.CanvasObject {
	if res == nil {
		return widget.NewLabel("Equation SVG unavailable; showing text-only equation.")
	}

	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillContain

	stack := container.NewStack(img)
	diagram := buildEquationDiagramWithZoom(parent, stack, fyne.NewSize(0, 0), 0.40)

	caption := widget.NewLabel("Quasi-static hysteron superposition model (no explicit dP/dt term).")
	caption.TextStyle = fyne.TextStyle{Italic: true}
	return container.NewVBox(diagram, caption)
}

func buildPreisachEquationTextPanel(selectTerm func(string, string)) fyne.CanvasObject {
	term := func(id, label, tooltip string) *TermChip {
		return NewTermChip(id, label, tooltip, selectTerm)
	}

	line1 := container.NewHBox(
		mathLabel("P(E) = ∬ "),
		term("preisach_mu", "μ(α,β)", "Weight/density of hysterons with thresholds (α, β)."),
		mathLabel(" · "),
		term("preisach_gamma", "γ_{α,β}(E)", "Bistable hysteron state (+1/-1) with memory."),
		mathLabel(" dα dβ"),
	)

	line2 := container.NewHBox(
		mathLabel("γ_{α,β}(E) = +1 if E >= "),
		term("preisach_alpha", "α", "Upper switching threshold for a hysteron."),
		mathLabel("; -1 if E <= "),
		term("preisach_beta", "β", "Lower switching threshold for a hysteron."),
	)

	line3 := container.NewHBox(
		mathLabel("hold if "),
		term("preisach_beta", "β", "Lower switching threshold for a hysteron."),
		mathLabel(" < E < "),
		term("preisach_alpha", "α", "Upper switching threshold for a hysteron."),
		mathLabel("; state follows history "),
		term("preisach_history", "history", "Turning-point memory that makes the model hysteretic."),
	)

	caption := widget.NewLabel("Tap a term to see its meaning and code mapping.")
	caption.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		line1,
		line2,
		line3,
		caption,
	)
}

func buildPreisachNotesSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Model Notes"),
		bodyLabel("Preisach treats hysteresis as a weighted sum of bistable hysterons:"),
		bodyLabel(bullets([]string{
			"Each hysteron flips at thresholds (alpha, beta) and retains memory between them.",
			"Quasi-static means rate-independent: no explicit dP/dt term or inertial delay.",
			"Output depends only on the input history ordering, not the sweep speed.",
			"Use Preisach for static loop shape; use L-K for switching dynamics.",
		})),
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

// embeddedSVGData returns the embedded SVG bytes for a given equation ID.
func embeddedSVGData(eqID string) []byte {
	switch eqID {
	case lkEquationID:
		return eqassets.LkEquationSVG
	case preisachEquationID:
		return eqassets.PreisachEquationSVG
	default:
		return nil
	}
}

func loadEquationSVGResource(eqID string) (fyne.Resource, bool) {
	equationSVGCacheMu.Lock()
	cacheKey := eqID + "|" + equationThemeKey()
	entry := equationSVGCache[cacheKey]
	if entry == nil {
		entry = &equationSVGCacheEntry{}
		equationSVGCache[cacheKey] = entry
	}
	equationSVGCacheMu.Unlock()

	entry.once.Do(func() {
		start := time.Now()
		data := embeddedSVGData(eqID)
		if data == nil {
			logEquationPerf("equation svg embed missing id=%s", eqID)
			return
		}

		fg := theme.ForegroundColor()
		recolorStart := time.Now()
		recolored, err := canvas.RecolorSVG(data, fg)
		recolorDur := time.Since(recolorStart)
		if err != nil {
			logEquationPerf("equation svg recolor failed id=%s err=%v", eqID, err)
			recolored = data
		}
		entry.res = fyne.NewStaticResource(eqID+".svg", recolored)
		entry.ok = true

		logEquationPerf(
			"equation svg loaded id=%s bytes=%d recolor=%s total=%s",
			eqID,
			len(recolored),
			recolorDur,
			time.Since(start),
		)
	})

	return entry.res, entry.ok
}

func loadEquationSVG(eqID string) *canvas.Image {
	if res, ok := loadEquationSVGResource(eqID); ok {
		return canvas.NewImageFromResource(res)
	}
	return nil
}

// PrefetchEquationAssets pre-warms the SVG recolor cache for the current theme.
// Call from a background goroutine at app startup so the equations dialog opens instantly.
func PrefetchEquationAssets() {
	loadEquationSVGResource(lkEquationID)
	loadEquationSVGResource(preisachEquationID)
	loadLkHotspots()
}

func loadLkHotspots() ([]hotspotDef, fyne.Size) {
	lkHotspotsOnce.Do(func() {
		defaultHotspots, defaultSize := defaultLkHotspots()
		start := time.Now()
		data := eqassets.LkHotspotsJSON

		var cfg hotspotConfig
		parseStart := time.Now()
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Printf("failed to parse embedded hotspots: %v", err)
			logEquationPerf("equation hotspots parse failed err=%v", err)
			cachedLkSpots = defaultHotspots
			cachedLkSize = defaultSize
			return
		}
		parseDur := time.Since(parseStart)
		_ = start // used below

		hotspots := defaultHotspots
		if len(cfg.Hotspots) > 0 {
			hotspots = cfg.Hotspots
		}

		size := defaultSize
		if cfg.BaseWidth > 0 && cfg.BaseHeight > 0 {
			size = fyne.NewSize(cfg.BaseWidth, cfg.BaseHeight)
		}

		cachedLkSpots = hotspots
		cachedLkSize = size

		logEquationPerf(
			"equation hotspots loaded (embedded) spots=%d parse=%s total=%s",
			len(hotspots),
			parseDur,
			time.Since(start),
		)
	})

	return cachedLkSpots, cachedLkSize
}

func defaultLkHotspots() ([]hotspotDef, fyne.Size) {
	return []hotspotDef{
		{
			ID:      "rho_eff_main",
			Tooltip: "Effective viscosity: intrinsic damping plus series-resistance RC delay.",
			X:       0.0018, Y: 0.0663, W: 0.0693, H: 0.1868,
		},
		{
			ID:      "e_applied",
			Tooltip: "Applied electric field drive term (external voltage across the film).",
			X:       0.1708, Y: 0.0663, W: 0.1462, H: 0.1868,
		},
		{
			ID:      "k_dep",
			Tooltip: "Depolarization factor: models interfacial layer; slants the loop for analog states.",
			X:       0.3481, Y: 0.0663, W: 0.0785, H: 0.1868,
		},
		{
			ID:      "alpha",
			Tooltip: "Dynamic stiffness: temperature + stress dependent curvature of energy wells.",
			X:       0.5149, Y: 0.0663, W: 0.0917, H: 0.1762,
		},
		{
			ID:      "beta",
			Tooltip: "First-order nonlinearity: negative for HZO to create the switching barrier.",
			X:       0.5987, Y: 0.0419, W: 0.1472, H: 0.2006,
		},
		{
			ID:      "gamma",
			Tooltip: "Sixth-order stabilizer: keeps energy bounded at large polarization.",
			X:       0.7381, Y: 0.0419, W: 0.1368, H: 0.2006,
		},
		{
			ID:      "lk_terms",
			Tooltip: "Landau-Khalatnikov nonlinear energy term: 2αP + 4βP^3 + 6γP^5.",
			X:       0.51, Y: 0.27, W: 0.37, H: 0.08,
		},
		{
			ID:      "noise",
			Tooltip: "Stochastic noise term (optional): captures thermal variability.",
			X:       0.9327, Y: 0.0663, W: 0.0673, H: 0.1762,
		},
		{
			ID:      "rho_eff_def",
			Tooltip: "Effective viscosity definition used in the headless hysteresis path.",
			X:       0.062, Y: 0.4108, W: 0.0754, H: 0.1868,
		},
		{
			ID:      "rho",
			Tooltip: "Intrinsic viscosity / damping coefficient.",
			X:       0.1708, Y: 0.4108, W: 0.0367, H: 0.1762,
		},
		{
			ID:      "r_series",
			Tooltip: "Series resistance: absorbs RC delay into viscosity.",
			X:       0.2434, Y: 0.3155, W: 0.0649, H: 0.1868,
		},
		{
			ID:      "alpha_def",
			Tooltip: "Alpha definition: temperature + stress dependent stiffness coefficient.",
			X:       0.0173, Y: 0.646, W: 0.4309, H: 0.354,
		},
	}, fyne.NewSize(1200, 332.6155)
}
