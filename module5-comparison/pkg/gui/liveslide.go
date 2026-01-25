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
	fyne.Do(func() {
		m.Refresh()
	})
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
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *comparisonModeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *comparisonModeRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("comparisonModeRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
	r.cache.MarkLayout(size)
}

func (r *comparisonModeRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("comparisonModeRenderer", r.indicator.Size())
	size := r.indicator.Size()
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
	r.cache.MarkLayout(size)
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
	fyne.Do(func() {
		e.Refresh()
	})
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
			"THE OPPORTUNITY\n\n"+
				"$721B addressable market by 2030\n"+
				"1000× claimed energy reduction\n"+
				"CMOS-compatible fabrication\n"+
				"Proven research pedigree\n\n"+
				"COMMERCIALIZATION STRATEGY\n"+
				"Phase 1: NAND replacement\n"+
				"  → Drop-in compatible design\n"+
				"  → Minimal integration risk\n\n"+
				"Clear TRL 4 → TRL 9 pathway")

	case PresentationModeEngineer:
		e.SetContent("Technical Deep-Dive",
			"FERROELECTRIC PHYSICS\n\n"+
				"HfO₂-ZrO₂ superlattice structure\n"+
				"Remanent polarization: 15-34 µC/cm²\n"+
				"Coercive field: 1.0-1.5 MV/cm\n"+
				"30 discrete analog levels\n\n"+
				"CROSSBAR ARCHITECTURE\n"+
				"Matrix-vector multiply: O(1) time\n"+
				"Physical parallelism via Kirchhoff's law\n"+
				"Current summation: I = Σ(Gᵢⱼ × Vⱼ)\n\n"+
				"ENGINEERING CHALLENGES\n"+
				"IR voltage drop mitigation\n"+
				"Sneak path current management\n"+
				"Long-term conductance stability")

	default:
		e.SetContent("Why Compute-in-Memory Wins",
			"THE MEMORY WALL PROBLEM\n\n"+
				"Von Neumann Architecture:\n"+
				"  • Data shuttles between\n"+
				"    separate memory and CPU\n"+
				"  • Energy dominated by data movement\n"+
				"  • Bandwidth bottleneck limits performance\n\n"+
				"Compute-in-Memory Solution:\n"+
				"  • Computation occurs at storage location\n"+
				"  • Eliminates data movement overhead\n"+
				"  • Massive parallel operations via physics")
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
		content = "$721B BY 2030\n\n" +
			"NAND Flash: $98B\n" +
			"DRAM: $220B\n" +
			"AI Semiconductor: $403B\n\n" +
			"FeCIM addresses all three segments"

	case AutoDemoPhaseCompetitive:
		title = "Competitive Position"
		content = "COMPETITIVE LANDSCAPE\n\n" +
			"Google TPU v5: Von Neumann arch\n" +
			"Intel Loihi 2: Exotic fabrication\n" +
			"IBM Analog AI: Research phase\n\n" +
			"FeCIM ADVANTAGES:\n" +
			"  ✓ True compute-in-memory\n" +
			"  ✓ Standard CMOS process\n" +
			"  ✓ Production scalability"

	case AutoDemoPhaseStrategy:
		title = "Phased Strategy"
		content = "COMMERCIALIZATION ROADMAP\n\n" +
			"Phase 1: NAND Replacement\n" +
			"  → Drop-in compatible interface\n" +
			"  → Leverage existing infrastructure\n\n" +
			"Phase 2: DRAM Displacement\n" +
			"  → Non-volatile, zero refresh power\n" +
			"  → Higher density potential\n\n" +
			"Phase 3: Full Compute-in-Memory\n" +
			"  → 80-90% claimed energy savings\n" +
			"  → Transform datacenter economics"

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

// SetComparison sets comparison explanation with calculated ratios.
func (e *ComparisonEducationalPanel) SetComparison(cpuRatio, gpuRatio float64) {
	content := "THE MEMORY WALL PROBLEM\n\n" +
		"Von Neumann Architecture:\n" +
		"  • Data shuttles between\n" +
		"    separate memory and CPU\n" +
		"  • Energy dominated by movement\n" +
		"  • Bandwidth bottleneck\n\n" +
		"Compute-in-Memory Solution:\n" +
		"  • Computation at storage location\n" +
		"  • Eliminates data movement\n" +
		"  • Physics-based parallelism\n\n" +
		"CLAIMED ADVANTAGES:\n" +
		fmt.Sprintf("  • %.0f× less power vs CPU*\n", cpuRatio) +
		fmt.Sprintf("  • %.0f× less power vs GPU*\n", gpuRatio) +
		"\n* Based on Dr. Tour's claims (TRL 4)\n" +
		"  Independent verification pending"
	e.SetContent("Why Compute-in-Memory Wins", content)
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

	// Wrap contentLabel in scroll container to prevent resize loops from text wrapping
	contentScroll := container.NewScroll(contentLabel)
	contentScroll.SetMinSize(fyne.NewSize(240, 160))

	box := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		contentScroll,
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
	text := "Ready for calculations..."
	if len(o.entries) > 0 {
		text = strings.Join(o.entries, "\n")
	}
	fyne.Do(func() {
		o.contentLabel.SetText(text)
	})
}

// MinSize returns the minimum size.
func (o *ComparisonOperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *ComparisonOperationLog) CreateRenderer() fyne.WidgetRenderer {
	// Wrap contentLabel in scroll container to prevent resize loops from text wrapping
	contentScroll := container.NewScroll(o.contentLabel)
	contentScroll.SetMinSize(fyne.NewSize(190, 120))

	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		contentScroll,
	)
	return widget.NewSimpleRenderer(box)
}
