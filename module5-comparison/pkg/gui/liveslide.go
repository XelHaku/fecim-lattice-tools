// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// ComparisonMode represents the current demo mode.
type ComparisonMode int

const (
	ComparisonModeIdle ComparisonMode = iota
	ComparisonModeCalculating
	ComparisonModeComparing
)

func (m ComparisonMode) String() string {
	switch m {
	case ComparisonModeIdle:
		return "IDLE"
	case ComparisonModeCalculating:
		return "CALCULATING"
	case ComparisonModeComparing:
		return "COMPARING"
	default:
		return "UNKNOWN"
	}
}

// PresentationMode represents the presentation/demo mode.
type PresentationMode int

const (
	PresentationModeManual   PresentationMode = iota // User controls navigation
	PresentationModeAuto                             // Self-running 30s per section
	PresentationModeInvestor                         // Large numbers, minimal jargon
	PresentationModeEngineer                         // Technical deep-dive
)

func (p PresentationMode) String() string {
	switch p {
	case PresentationModeManual:
		return "Manual"
	case PresentationModeAuto:
		return "Auto Demo"
	case PresentationModeInvestor:
		return "Investor"
	case PresentationModeEngineer:
		return "Engineer"
	default:
		return "Unknown"
	}
}

// PresentationModeFromString converts string to PresentationMode.
func PresentationModeFromString(s string) PresentationMode {
	switch s {
	case "Manual":
		return PresentationModeManual
	case "Auto Demo":
		return PresentationModeAuto
	case "Investor":
		return PresentationModeInvestor
	case "Engineer":
		return PresentationModeEngineer
	default:
		return PresentationModeManual
	}
}

// AutoDemoPhase represents phases in the auto demo sequence.
type AutoDemoPhase int

const (
	AutoDemoPhaseEnergyRace AutoDemoPhase = iota
	AutoDemoPhaseMarket
	AutoDemoPhaseCompetitive
	AutoDemoPhaseStrategy
	AutoDemoPhaseCalculator
	AutoDemoPhaseCount // Total number of phases
)

func (p AutoDemoPhase) String() string {
	switch p {
	case AutoDemoPhaseEnergyRace:
		return "Energy Comparison"
	case AutoDemoPhaseMarket:
		return "Market Opportunity"
	case AutoDemoPhaseCompetitive:
		return "Competitive Matrix"
	case AutoDemoPhaseStrategy:
		return "Phased Strategy"
	case AutoDemoPhaseCalculator:
		return "Calculator Demo"
	default:
		return "Unknown"
	}
}

// PhaseDuration returns the duration for each auto-demo phase.
func (p AutoDemoPhase) PhaseDuration() time.Duration {
	switch p {
	case AutoDemoPhaseEnergyRace:
		return 10 * time.Second
	case AutoDemoPhaseMarket:
		return 10 * time.Second
	case AutoDemoPhaseCompetitive:
		return 10 * time.Second
	case AutoDemoPhaseStrategy:
		return 10 * time.Second
	case AutoDemoPhaseCalculator:
		return 15 * time.Second
	default:
		return 10 * time.Second
	}
}

// ComparisonModeIndicator shows the current mode.
type ComparisonModeIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	mode    ComparisonMode
	minSize fyne.Size
}

// NewComparisonModeIndicator creates a new mode indicator.
func NewComparisonModeIndicator() *ComparisonModeIndicator {
	m := &ComparisonModeIndicator{
		mode:    ComparisonModeIdle,
		minSize: fyne.NewSize(120, 40),
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetMode updates the current mode.
func (m *ComparisonModeIndicator) SetMode(mode ComparisonMode) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	m.Refresh()
}

// MinSize returns the minimum size.
func (m *ComparisonModeIndicator) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *ComparisonModeIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &comparisonModeRenderer{indicator: m}
}

type comparisonModeRenderer struct {
	indicator *ComparisonModeIndicator
	objects   []fyne.CanvasObject
}

func (r *comparisonModeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *comparisonModeRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("comparisonModeRenderer", size)
	r.layoutWithSize(size)
}

func (r *comparisonModeRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("comparisonModeRenderer", r.indicator.Size())
	r.layoutWithSize(r.indicator.Size())
}

func (r *comparisonModeRenderer) layoutWithSize(size fyne.Size) {
	// Skip layout with invalid sizes
	if size.Width <= 0 || size.Height <= 0 {
		return
	}

	r.indicator.mu.RLock()
	mode := r.indicator.mode
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]

	// Constrain to minimum size to prevent growing
	minSize := r.indicator.minSize
	if size.Width > minSize.Width {
		size.Width = minSize.Width
	}
	if size.Height > minSize.Height {
		size.Height = minSize.Height
	}

	var bgColor, borderColor color.RGBA
	var modeText string

	switch mode {
	case ComparisonModeIdle:
		bgColor = color.RGBA{60, 60, 80, 255}
		borderColor = color.RGBA{100, 100, 130, 255}
		modeText = "IDLE"
	case ComparisonModeCalculating:
		bgColor = color.RGBA{80, 120, 50, 255}
		borderColor = color.RGBA{140, 200, 100, 255}
		modeText = "CALCULATING"
	case ComparisonModeComparing:
		bgColor = color.RGBA{50, 80, 150, 255}
		borderColor = color.RGBA{100, 150, 255, 255}
		modeText = "COMPARING"
	}

	// Border
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background
	padding := float32(2)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Mode text
	text := canvas.NewText(modeText, color.White)
	fontSize := size.Height * 0.35
	if fontSize > 14 {
		fontSize = 14
	}
	if fontSize < 10 {
		fontSize = 10
	}
	text.TextSize = fontSize
	text.TextStyle = fyne.TextStyle{Bold: true}
	textWidth := float32(len(modeText)) * fontSize * 0.6
	text.Move(fyne.NewPos((size.Width-textWidth)/2, (size.Height-fontSize)/2))
	r.objects = append(r.objects, text)
}

func (r *comparisonModeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *comparisonModeRenderer) Destroy() {}

// ComparisonEducationalPanel shows explanations.
type ComparisonEducationalPanel struct {
	widget.BaseWidget

	mu               sync.RWMutex
	title            string
	content          string
	minSize          fyne.Size
	presentationMode PresentationMode
	currentPhase     AutoDemoPhase
}

// NewComparisonEducationalPanel creates a new educational panel.
func NewComparisonEducationalPanel() *ComparisonEducationalPanel {
	e := &ComparisonEducationalPanel{
		title:            "Why CIM Wins",
		content:          "Compute-in-memory eliminates\nthe memory bottleneck.",
		minSize:          fyne.NewSize(200, 200),
		presentationMode: PresentationModeManual,
		currentPhase:     AutoDemoPhaseEnergyRace,
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetContent updates the content.
func (e *ComparisonEducationalPanel) SetContent(title, content string) {
	e.mu.Lock()
	e.title = title
	e.content = content
	e.mu.Unlock()
	e.Refresh()
}

// SetPresentationMode sets the current presentation mode.
func (e *ComparisonEducationalPanel) SetPresentationMode(mode PresentationMode) {
	e.mu.Lock()
	e.presentationMode = mode
	e.mu.Unlock()
	e.updateForMode()
}

// SetPhase sets the current auto-demo phase.
func (e *ComparisonEducationalPanel) SetPhase(phase AutoDemoPhase) {
	e.mu.Lock()
	e.currentPhase = phase
	e.mu.Unlock()
	e.updateForPhase()
}

// updateForMode updates content based on presentation mode.
func (e *ComparisonEducationalPanel) updateForMode() {
	e.mu.RLock()
	mode := e.presentationMode
	e.mu.RUnlock()

	switch mode {
	case PresentationModeInvestor:
		e.SetContent("Scenario Summary",
			"THE PITCH\n\n"+
				"$711B market by 2030\n"+
				"100× energy reduction\n"+
				"CMOS compatible fab\n"+
				"Proven research team\n\n"+
				"PHASE 1: NAND Replacement\n"+
				"Drop-in compatible\n"+
				"Low adoption risk\n\n"+
				"TRL 4 → TRL 9 path clear")

	case PresentationModeEngineer:
		e.SetContent("Technical Deep-Dive",
			"PHYSICS\n\n"+
				"HfO2-ZrO2 superlattice\n"+
				"Pr ≈ 25 µC/cm²\n"+
				"Ec ≈ 1 MV/cm\n"+
				"30 analog levels\n\n"+
				"CROSSBAR ARRAY\n"+
				"MVM in O(1) time\n"+
				"Kirchhoff's law\n"+
				"I = G × V summation\n\n"+
				"NON-IDEALITIES\n"+
				"IR drop, sneak paths\n"+
				"Conductance drift")

	default:
		e.SetContent("Why CIM Wins",
			"THE MEMORY WALL\n\n"+
				"Traditional CPUs/GPUs:\n"+
				"Data moves between\n"+
				"memory and processor.\n"+
				"This wastes energy.\n\n"+
				"Compute-in-Memory:\n"+
				"Computation happens\n"+
				"WHERE data lives.\n"+
				"No movement = no waste.")
	}
}

// updateForPhase updates content based on auto-demo phase.
func (e *ComparisonEducationalPanel) updateForPhase() {
	e.mu.RLock()
	phase := e.currentPhase
	mode := e.presentationMode
	e.mu.RUnlock()

	var title, content string

	switch phase {
	case AutoDemoPhaseEnergyRace:
		title = "Energy Comparison"
		if mode == PresentationModeInvestor {
			content = "THE HEADLINE\n\n" +
				"1000× less energy\n" +
				"than current CPUs\n" +
				"100× less than GPUs\n\n" +
				"= 90% cost reduction\n" +
				"= 10× more inference\n" +
				"= same power budget"
		} else {
			content = "ENERGY PER MAC\n\n" +
				"CPU + DRAM: 1000 pJ\n" +
				"GPU + HBM: 100 pJ\n" +
				"FeCIM: ~1 pJ*\n\n" +
				"* TRL 4 claims\n" +
				"(1 pJ = 1000 fJ)"
		}

	case AutoDemoPhaseMarket:
		title = "Market Opportunity"
		content = "$711B BY 2030\n\n" +
			"NAND Flash: $98B\n" +
			"DRAM: $220B\n" +
			"AI Semiconductor: $403B\n\n" +
			"FeCIM addresses ALL THREE"

	case AutoDemoPhaseCompetitive:
		title = "Competitive Position"
		content = "VS COMPETITION\n\n" +
			"Google TPU: Not in-memory\n" +
			"Intel Loihi: Non-CMOS\n" +
			"Mythic AI: Not scalable\n\n" +
			"FeCIM: ✓ In-memory\n" +
			"       ✓ CMOS fab\n" +
			"       ✓ Scalable"

	case AutoDemoPhaseStrategy:
		title = "Phased Strategy"
		content = "COMMERCIALIZATION\n\n" +
			"Phase 1: NAND replacement\n" +
			"  → Drop-in compatible\n\n" +
			"Phase 2: DRAM displacement\n" +
			"  → No refresh needed\n\n" +
			"Phase 3: Full CIM\n" +
			"  → 80-90% energy savings"

	case AutoDemoPhaseCalculator:
		title = "Real Impact"
		content = "DATA CENTER SAVINGS\n\n" +
			"At 10,000 inferences/sec:\n\n" +
			"GPU: $X,XXX/month\n" +
			"FeCIM: $XXX/month\n\n" +
			"Try the calculator\n" +
			"with your workload!"
	}

	e.SetContent(title, content)
}

// SetComparison sets comparison explanation.
func (e *ComparisonEducationalPanel) SetComparison(cpuRatio, gpuRatio float64) {
	content := "THE MEMORY WALL\n\n" +
		"Traditional CPUs/GPUs:\n" +
		"  Data moves between\n" +
		"  memory and processor.\n" +
		"  This wastes energy.\n\n" +
		"Compute-in-Memory:\n" +
		"  Computation happens\n" +
		"  WHERE data lives.\n" +
		"  No movement = no waste.\n\n" +
		fmt.Sprintf("FeCIM vs CPU: %.0f× less power*\n", cpuRatio) +
		fmt.Sprintf("FeCIM vs GPU: %.0f× less power*\n", gpuRatio) +
		"\n* If claims hold (TRL 4)"
	e.SetContent("Why CIM Wins", content)
}

// MinSize returns the minimum size.
func (e *ComparisonEducationalPanel) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *ComparisonEducationalPanel) CreateRenderer() fyne.WidgetRenderer {
	e.mu.RLock()
	title := e.title
	content := e.content
	e.mu.RUnlock()

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	contentLabel := widget.NewLabel(content)
	contentLabel.Wrapping = fyne.TextWrapWord

	box := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		contentLabel,
	)

	return widget.NewSimpleRenderer(box)
}

// ComparisonOperationLog shows timestamped operations.
type ComparisonOperationLog struct {
	widget.BaseWidget

	mu         sync.RWMutex
	entries    []string
	maxEntries int
	startTime  time.Time
	minSize    fyne.Size

	titleLabel   *widget.Label
	contentLabel *widget.Label
}

// NewComparisonOperationLog creates a new operation log.
func NewComparisonOperationLog() *ComparisonOperationLog {
	o := &ComparisonOperationLog{
		maxEntries: 8,
		startTime:  time.Now(),
		minSize:    fyne.NewSize(200, 150),
		entries:    make([]string, 0, 8),
	}
	o.titleLabel = widget.NewLabelWithStyle("Calculation Log", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	o.contentLabel = widget.NewLabel("Ready for calculations...")
	o.contentLabel.Wrapping = fyne.TextWrapWord
	o.ExtendBaseWidget(o)
	return o
}

// Add adds a new log entry.
func (o *ComparisonOperationLog) Add(entry string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	elapsed := time.Since(o.startTime).Seconds()
	timestamped := fmt.Sprintf("t=%.1fs >> %s", elapsed, entry)
	o.entries = append(o.entries, timestamped)

	if len(o.entries) > o.maxEntries {
		o.entries = o.entries[1:]
	}

	o.updateContent()
}

func (o *ComparisonOperationLog) updateContent() {
	if len(o.entries) == 0 {
		o.contentLabel.SetText("Ready for calculations...")
		return
	}
	o.contentLabel.SetText(strings.Join(o.entries, "\n"))
}

// MinSize returns the minimum size.
func (o *ComparisonOperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *ComparisonOperationLog) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		o.contentLabel,
	)
	return widget.NewSimpleRenderer(box)
}
