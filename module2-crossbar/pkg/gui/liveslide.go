// Package gui provides Fyne-based GUI components for crossbar visualization.
// This file implements the "Live Slide" pattern components for Demo 2.
package gui

import (
	"fmt"
	"image/color"
	stdlog "log"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

var lsDebug = stdlog.New(os.Stdout, "[WIDGET] ", stdlog.Ltime|stdlog.Lmicroseconds)

// DemoMode represents the current demo mode.
type DemoMode int

const (
	DemoModeIdle DemoMode = iota
	DemoModeCompute
	DemoModeWrite
	DemoModeRead
	DemoModeIRDrop
	DemoModeSneakPath
)

func (m DemoMode) String() string {
	switch m {
	case DemoModeIdle:
		return "IDLE"
	case DemoModeCompute:
		return "COMPUTE"
	case DemoModeWrite:
		return "WRITE"
	case DemoModeRead:
		return "READ"
	case DemoModeIRDrop:
		return "IR DROP"
	case DemoModeSneakPath:
		return "SNEAK"
	default:
		return "UNKNOWN"
	}
}

// ModeIndicatorBox shows the current mode with a colored background.
type ModeIndicatorBox struct {
	widget.BaseWidget

	mu      sync.RWMutex
	mode    DemoMode
	minSize fyne.Size
}

// NewModeIndicatorBox creates a new mode indicator.
func NewModeIndicatorBox() *ModeIndicatorBox {
	m := &ModeIndicatorBox{
		mode:    DemoModeIdle,
		minSize: fyne.NewSize(100, 30),
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetMode updates the current mode.
func (m *ModeIndicatorBox) SetMode(mode DemoMode) {
	lsDebug.Printf("ModeIndicator: SetMode(%s)", mode.String())
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	lsDebug.Println("ModeIndicator: Calling Refresh")
	fyne.Do(func() {
		m.Refresh()
	})
	lsDebug.Println("ModeIndicator: Refresh done")
}

// GetMode returns the current mode.
func (m *ModeIndicatorBox) GetMode() DemoMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// MinSize returns the minimum size.
func (m *ModeIndicatorBox) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *ModeIndicatorBox) CreateRenderer() fyne.WidgetRenderer {
	return &modeIndicatorBoxRenderer{indicator: m}
}

type modeIndicatorBoxRenderer struct {
	indicator *ModeIndicatorBox
	objects   []fyne.CanvasObject
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *modeIndicatorBoxRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *modeIndicatorBoxRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("modeIndicatorBoxRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
	r.cache.MarkLayout(size)
}

func (r *modeIndicatorBoxRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("modeIndicatorBoxRenderer", r.indicator.Size())
	size := r.indicator.Size()
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
	r.cache.MarkLayout(size)
}

func (r *modeIndicatorBoxRenderer) layoutWithSize(size fyne.Size) {
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

	// Colors based on mode
	var bgColor, borderColor color.RGBA
	var modeText string

	switch mode {
	case DemoModeIdle:
		bgColor = color.RGBA{60, 60, 80, 255}
		borderColor = color.RGBA{100, 100, 130, 255}
		modeText = "░░ IDLE ░░"
	case DemoModeCompute:
		bgColor = color.RGBA{50, 120, 180, 255}
		borderColor = color.RGBA{100, 180, 255, 255}
		modeText = "▶▶ COMPUTE ▶▶"
	case DemoModeWrite:
		bgColor = color.RGBA{180, 50, 50, 255}
		borderColor = color.RGBA{255, 100, 100, 255}
		modeText = "██ WRITE ██"
	case DemoModeRead:
		bgColor = color.RGBA{50, 150, 80, 255}
		borderColor = color.RGBA{100, 220, 130, 255}
		modeText = "░░ READ ░░"
	case DemoModeIRDrop:
		bgColor = color.RGBA{180, 120, 50, 255}
		borderColor = color.RGBA{255, 180, 100, 255}
		modeText = "⚡ IR DROP ⚡"
	case DemoModeSneakPath:
		bgColor = color.RGBA{150, 50, 150, 255}
		borderColor = color.RGBA{220, 100, 220, 255}
		modeText = "↯ SNEAK ↯"
	}

	// Border
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background
	padding := float32(3)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Mode text
	text := canvas.NewText(modeText, color.White)
	text.TextSize = 14
	text.TextStyle = fyne.TextStyle{Bold: true}
	textWidth := float32(len(modeText) * 8)
	text.Move(fyne.NewPos((size.Width-textWidth)/2, (size.Height-20)/2))
	r.objects = append(r.objects, text)
}

func (r *modeIndicatorBoxRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *modeIndicatorBoxRenderer) Destroy() {}

// EducationalPanel shows context-sensitive explanations of what's happening.
type EducationalPanel struct {
	widget.BaseWidget

	mu           sync.RWMutex
	title        string
	content      string
	phase        int
	minSize      fyne.Size
	titleLabel   *widget.Label
	contentLabel *widget.Label
}

// NewEducationalPanel creates a new educational panel.
func NewEducationalPanel() *EducationalPanel {
	e := &EducationalPanel{
		title:   "What You're Seeing",
		content: "Select an operation to see\nwhat's happening.",
		minSize: fyne.NewSize(220, 280), // Larger size to fit content
	}
	e.titleLabel = widget.NewLabelWithStyle(e.title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	e.contentLabel = widget.NewLabel(e.content)
	e.contentLabel.Wrapping = fyne.TextWrapOff
	e.ExtendBaseWidget(e)
	return e
}

// SetContent updates the educational content.
func (e *EducationalPanel) SetContent(title, content string) {
	lsDebug.Printf("EducationalPanel: SetContent(%s)", title)
	e.mu.Lock()
	e.title = title
	e.content = content
	e.mu.Unlock()
	fyne.Do(func() {
		e.titleLabel.SetText(title)
		e.contentLabel.SetText(content)
	})
	lsDebug.Println("EducationalPanel: SetContent done")
}

// SetPhase updates the current phase for phase-aware content.
func (e *EducationalPanel) SetPhase(phase int) {
	e.mu.Lock()
	e.phase = phase
	e.mu.Unlock()
	fyne.Do(func() {
		e.Refresh()
	})
}

// SetMVMExplanation sets content for MVM operation.
func (e *EducationalPanel) SetMVMExplanation(phase int) {
	var content string
	switch phase {
	case 1:
		content = "MVM PHASE 1: Input\n\n" +
			"DAC converts digital input\n" +
			"to analog voltages V[0...N-1]\n\n" +
			"Each voltage applied to\n" +
			"one column (bitline).\n\n" +
			"This drives current through\n" +
			"ALL cells in that column\n" +
			"simultaneously."
	case 2:
		content = "MVM PHASE 2: Multiply\n\n" +
			"Current flows through every\n" +
			"cell in parallel.\n\n" +
			"Physics does the math:\n" +
			"I_ij = G_ij × V_j\n\n" +
			"Each cell performs one\n" +
			"multiplication using\n" +
			"Ohm's Law - no transistors!\n\n" +
			"This is TRUE analog compute."
	case 3:
		content = "MVM PHASE 3: Accumulate\n\n" +
			"Row currents (wordlines)\n" +
			"sum automatically via\n" +
			"Kirchhoff's Current Law.\n\n" +
			"I_row[i] = Σ(G_ij × V_j)\n\n" +
			"ADC converts analog currents\n" +
			"to digital output.\n\n" +
			"Result: N² MACs in ~1ns!\n" +
			"(vs. N² cycles in CPU)"
	default:
		content = "MVM OPERATION\n\n" +
			"Matrix-Vector Multiply:\n" +
			"Output = Weights × Input\n" +
			"I = G × V\n\n" +
			"The crossbar computes\n" +
			"the entire matrix-vector\n" +
			"product in ONE analog step.\n\n" +
			"Speed: ~1ns (vs. µs in CPU)\n" +
			"Energy: ~10pJ (vs. nJ in GPU)"
	}
	e.SetContent("Compute-in-Memory", content)
}

// SetIRDropExplanation sets content for IR drop analysis.
func (e *EducationalPanel) SetIRDropExplanation() {
	content := "IR DROP ANALYSIS\n\n" +
		"Problem: Metal wires have\n" +
		"finite resistance (~1-10Ω/cell).\n\n" +
		"Effect: Voltage drops as\n" +
		"current flows along wires:\n" +
		"V_drop = I × R_wire\n\n" +
		"Impact: Cells far from drivers\n" +
		"compute with lower voltage,\n" +
		"reducing accuracy.\n\n" +
		"Worst at: Array corners\n" +
		"(longest wire paths)\n\n" +
		"Mitigation:\n" +
		"• Multiple voltage drivers\n" +
		"• Lower wire resistance (Cu)\n" +
		"• Smaller tile sizes"
	e.SetContent("Non-Ideality: IR Drop", content)
}

// SetSneakPathExplanation sets content for sneak path analysis.
func (e *EducationalPanel) SetSneakPathExplanation() {
	content := "SNEAK PATH ANALYSIS\n\n" +
		"Problem: In passive (0T1R)\n" +
		"crossbars, current can flow\n" +
		"through unselected cells.\n\n" +
		"Effect: Parasitic currents\n" +
		"reduce Signal-to-Noise Ratio.\n\n" +
		"Example sneak path:\n" +
		"WL[i] → Cell[i,j] → BL[j] ✓\n" +
		"WL[i] → Cell[i,k] → BL[k]\n" +
		"        → Cell[m,k] → BL[j] ✗\n\n" +
		"Impact: 2-15% error in large\n" +
		"arrays (worse as N increases)\n\n" +
		"Solutions:\n" +
		"• 1T1R (transistor switch)\n" +
		"• Selector device (diode)\n" +
		"• Self-rectifying FeFET"
	e.SetContent("Non-Ideality: Sneak Paths", content)
}

// SetIdleExplanation sets content for idle state.
func (e *EducationalPanel) SetIdleExplanation() {
	content := "FeFET CROSSBAR ARRAY\n\n" +
		"Compute-in-Memory (CIM)\n" +
		"where the same physical\n" +
		"device stores weights AND\n" +
		"performs computation.\n\n" +
		"Key advantage: No data\n" +
		"movement between memory\n" +
		"and processor!\n\n" +
		"Traditional: DRAM → CPU\n" +
		"FeCIM: Compute WHERE stored\n\n" +
		"Click 'Run MVM' to see\n" +
		"analog computation in action!"
	e.SetContent("What You're Seeing", content)
}

// MinSize returns the minimum size.
func (e *EducationalPanel) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *EducationalPanel) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(
		e.titleLabel,
		widget.NewSeparator(),
		e.contentLabel,
	)

	// Use padded container with border
	padded := container.NewPadded(box)
	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, nil, nil, padded))
}

// OperationLog shows timestamped operation history.
type OperationLog struct {
	widget.BaseWidget

	mu         sync.RWMutex
	entries    []string
	maxEntries int
	startTime  time.Time
	minSize    fyne.Size

	// UI components
	titleLabel   *widget.Label
	contentLabel *widget.Label
}

// NewOperationLog creates a new operation log.
func NewOperationLog() *OperationLog {
	o := &OperationLog{
		maxEntries: 6,
		startTime:  time.Now(),
		minSize:    fyne.NewSize(150, 60),
		entries:    make([]string, 0, 6),
	}
	o.titleLabel = widget.NewLabelWithStyle("Operation Log", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	o.contentLabel = widget.NewLabel("Waiting for operations...")
	o.contentLabel.Wrapping = fyne.TextWrapWord
	o.ExtendBaseWidget(o)
	return o
}

// Add adds a new log entry with timestamp.
func (o *OperationLog) Add(entry string) {
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

// AddWithResult adds an entry with a result indicator.
func (o *OperationLog) AddWithResult(action string, result string, success bool) {
	indicator := "✓"
	if !success {
		indicator = "✗"
	}
	o.Add(fmt.Sprintf("%s → %s %s", action, result, indicator))
}

// Clear clears all log entries.
func (o *OperationLog) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.entries = o.entries[:0]
	o.startTime = time.Now()
	o.updateContent()
}

func (o *OperationLog) updateContent() {
	text := "Waiting for operations..."
	if len(o.entries) > 0 {
		text = strings.Join(o.entries, "\n")
	}
	fyne.Do(func() {
		o.contentLabel.SetText(text)
	})
}

// MinSize returns the minimum size.
func (o *OperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *OperationLog) CreateRenderer() fyne.WidgetRenderer {
	// Wrap contentLabel in scroll container to prevent resize loops from text wrapping
	contentScroll := container.NewScroll(o.contentLabel)
	contentScroll.SetMinSize(fyne.NewSize(140, 40))

	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		contentScroll,
	)
	return widget.NewSimpleRenderer(box)
}

// InputOutputDisplay shows the input and output vectors.
type InputOutputDisplay struct {
	widget.BaseWidget

	mu           sync.RWMutex
	inputValues  []float64
	outputValues []float64
	inputLabels  []string
	outputLabels []string
	maxDisplay   int
	minSize      fyne.Size

	// UI components
	inputLabel    *widget.Label
	outputLabel   *widget.Label
	inputContent  *widget.Label
	outputContent *widget.Label
}

// NewInputOutputDisplay creates a new I/O display.
func NewInputOutputDisplay() *InputOutputDisplay {
	d := &InputOutputDisplay{
		maxDisplay: 6,
		minSize:    fyne.NewSize(150, 60),
	}
	d.inputLabel = widget.NewLabelWithStyle("Input Vector V", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	d.outputLabel = widget.NewLabelWithStyle("Output Vector I", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	d.inputContent = widget.NewLabel("V = [...]")
	d.outputContent = widget.NewLabel("I = [...]")
	d.ExtendBaseWidget(d)
	return d
}

// SetInput updates the input vector display.
func (d *InputOutputDisplay) SetInput(values []float64) {
	lsDebug.Printf("IODisplay: SetInput (len=%d)", len(values))
	d.mu.Lock()
	d.inputValues = values
	d.mu.Unlock()
	d.updateDisplay()
	lsDebug.Println("IODisplay: SetInput done")
}

// SetOutput updates the output vector display.
func (d *InputOutputDisplay) SetOutput(values []float64) {
	lsDebug.Printf("IODisplay: SetOutput (len=%d)", len(values))
	d.mu.Lock()
	d.outputValues = values
	d.mu.Unlock()
	d.updateDisplay()
	lsDebug.Println("IODisplay: SetOutput done")
}

func (d *InputOutputDisplay) formatVector(values []float64, prefix string) string {
	if len(values) == 0 {
		return fmt.Sprintf("%s = [...]", prefix)
	}

	// Show first few values
	maxShow := d.maxDisplay
	if len(values) < maxShow {
		maxShow = len(values)
	}

	parts := make([]string, maxShow)
	for i := 0; i < maxShow; i++ {
		parts[i] = fmt.Sprintf("%.2f", values[i])
	}

	result := fmt.Sprintf("%s = [%s", prefix, strings.Join(parts, ", "))
	if len(values) > maxShow {
		result += fmt.Sprintf(", ...+%d]", len(values)-maxShow)
	} else {
		result += "]"
	}
	return result
}

func (d *InputOutputDisplay) updateDisplay() {
	d.mu.RLock()
	input := d.inputValues
	output := d.outputValues
	d.mu.RUnlock()

	fyne.Do(func() {
		d.inputContent.SetText(d.formatVector(input, "V"))
		d.outputContent.SetText(d.formatVector(output, "I"))
		d.Refresh()
	})
}

// MinSize returns the minimum size.
func (d *InputOutputDisplay) MinSize() fyne.Size {
	return d.minSize
}

// CreateRenderer implements fyne.Widget.
func (d *InputOutputDisplay) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(
		d.inputLabel,
		d.inputContent,
		widget.NewSeparator(),
		d.outputLabel,
		d.outputContent,
	)
	return widget.NewSimpleRenderer(box)
}

// QuoteBox displays a centered italic text.
type QuoteBox struct {
	widget.BaseWidget

	text    string
	minSize fyne.Size
}

// NewQuoteBox creates a new quote box.
func NewQuoteBox(text string) *QuoteBox {
	q := &QuoteBox{
		text:    text,
		minSize: fyne.NewSize(200, 30),
	}
	q.ExtendBaseWidget(q)
	return q
}

// SetQuote updates the text.
func (q *QuoteBox) SetQuote(text string) {
	q.text = text
	fyne.Do(func() {
		q.Refresh()
	})
}

// MinSize returns the minimum size.
func (q *QuoteBox) MinSize() fyne.Size {
	return q.minSize
}

// CreateRenderer implements fyne.Widget.
func (q *QuoteBox) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabelWithStyle(
		q.text,
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	label.Wrapping = fyne.TextWrapWord
	return widget.NewSimpleRenderer(label)
}

// KeyStatBox displays a key statistic prominently.
type KeyStatBox struct {
	widget.BaseWidget

	mu          sync.RWMutex
	label       string
	value       string
	minSize     fyne.Size
	labelWidget *widget.Label
	valueWidget *widget.Label
}

// NewKeyStatBox creates a new key stat box.
func NewKeyStatBox(label, value string) *KeyStatBox {
	k := &KeyStatBox{
		label:   label,
		value:   value,
		minSize: fyne.NewSize(220, 60),
	}
	k.labelWidget = widget.NewLabel(label)
	k.labelWidget.Alignment = fyne.TextAlignCenter
	k.valueWidget = widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	k.ExtendBaseWidget(k)
	return k
}

// SetValue updates the statistic value.
func (k *KeyStatBox) SetValue(value string) {
	k.mu.Lock()
	k.value = value
	k.mu.Unlock()
	fyne.Do(func() {
		k.valueWidget.SetText(value)
	})
}

// MinSize returns the minimum size.
func (k *KeyStatBox) MinSize() fyne.Size {
	return k.minSize
}

// CreateRenderer implements fyne.Widget.
func (k *KeyStatBox) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(k.labelWidget, k.valueWidget)
	padded := container.NewPadded(box)
	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, nil, nil, padded))
}
