// Package gui provides Fyne-based GUI components for crossbar visualization.
// This file implements the "Live Slide" pattern components for Demo 2.
package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

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
		minSize: fyne.NewSize(120, 50),
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetMode updates the current mode.
func (m *ModeIndicatorBox) SetMode(mode DemoMode) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	m.Refresh()
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
}

func (r *modeIndicatorBoxRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *modeIndicatorBoxRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *modeIndicatorBoxRenderer) Refresh() {
	r.indicator.mu.RLock()
	mode := r.indicator.mode
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.indicator.Size()

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

	mu       sync.RWMutex
	title    string
	content  string
	phase    int
	minSize  fyne.Size
}

// NewEducationalPanel creates a new educational panel.
func NewEducationalPanel() *EducationalPanel {
	e := &EducationalPanel{
		title:   "What You're Seeing",
		content: "Select an operation to see\nwhat's happening.",
		minSize: fyne.NewSize(200, 200),
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetContent updates the educational content.
func (e *EducationalPanel) SetContent(title, content string) {
	e.mu.Lock()
	e.title = title
	e.content = content
	e.mu.Unlock()
	e.Refresh()
}

// SetPhase updates the current phase for phase-aware content.
func (e *EducationalPanel) SetPhase(phase int) {
	e.mu.Lock()
	e.phase = phase
	e.mu.Unlock()
	e.Refresh()
}

// SetMVMExplanation sets content for MVM operation.
func (e *EducationalPanel) SetMVMExplanation(phase int) {
	var content string
	switch phase {
	case 1:
		content = "MVM OPERATION\n\n" +
			"1. Input voltages V applied\n" +
			"   to column lines\n\n" +
			"Each voltage drives current\n" +
			"through ALL cells in column."
	case 2:
		content = "MVM OPERATION\n\n" +
			"2. Current flows through\n" +
			"   ALL cells simultaneously\n\n" +
			"I = G × V (Ohm's Law)\n" +
			"Each cell multiplies!"
	case 3:
		content = "MVM OPERATION\n\n" +
			"3. Row currents collected\n" +
			"   = dot product result\n\n" +
			"N² multiplications in\n" +
			"ONE clock cycle!"
	default:
		content = "MVM OPERATION\n\n" +
			"Matrix-Vector Multiplication:\n" +
			"I = G × V\n\n" +
			"The crossbar computes the\n" +
			"entire matrix operation\n" +
			"in a single step."
	}
	e.SetContent("Compute-in-Memory", content)
}

// SetIRDropExplanation sets content for IR drop analysis.
func (e *EducationalPanel) SetIRDropExplanation() {
	content := "IR DROP ANALYSIS\n\n" +
		"Wire resistance causes\n" +
		"voltage drop along lines.\n\n" +
		"Cells far from drivers\n" +
		"see reduced voltage.\n\n" +
		"This affects accuracy:\n" +
		"• Worst at corners\n" +
		"• Mitigate with drivers"
	e.SetContent("Non-Ideality: IR Drop", content)
}

// SetSneakPathExplanation sets content for sneak path analysis.
func (e *EducationalPanel) SetSneakPathExplanation() {
	content := "SNEAK PATH ANALYSIS\n\n" +
		"Current can flow through\n" +
		"unintended paths in passive\n" +
		"crossbar arrays.\n\n" +
		"Mitigation strategies:\n" +
		"• Selector devices\n" +
		"• 1T1R architecture\n" +
		"• Threshold switching"
	e.SetContent("Non-Ideality: Sneak Paths", content)
}

// SetIdleExplanation sets content for idle state.
func (e *EducationalPanel) SetIdleExplanation() {
	content := "CROSSBAR MVM\n\n" +
		"\"Compute in memory where\n" +
		"the same device does memory\n" +
		"and computation.\"\n\n" +
		"— Dr. external research group\n\n" +
		"Click a button to start\n" +
		"a demonstration."
	e.SetContent("What You're Seeing", content)
}

// MinSize returns the minimum size.
func (e *EducationalPanel) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *EducationalPanel) CreateRenderer() fyne.WidgetRenderer {
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
		maxEntries: 10,
		startTime:  time.Now(),
		minSize:    fyne.NewSize(200, 180),
		entries:    make([]string, 0, 10),
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
	if len(o.entries) == 0 {
		o.contentLabel.SetText("Waiting for operations...")
		return
	}
	o.contentLabel.SetText(strings.Join(o.entries, "\n"))
}

// MinSize returns the minimum size.
func (o *OperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *OperationLog) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		o.contentLabel,
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
		maxDisplay: 8,
		minSize:    fyne.NewSize(200, 160),
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
	d.mu.Lock()
	d.inputValues = values
	d.mu.Unlock()
	d.updateDisplay()
}

// SetOutput updates the output vector display.
func (d *InputOutputDisplay) SetOutput(values []float64) {
	d.mu.Lock()
	d.outputValues = values
	d.mu.Unlock()
	d.updateDisplay()
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

	d.inputContent.SetText(d.formatVector(input, "V"))
	d.outputContent.SetText(d.formatVector(output, "I"))
	d.Refresh()
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

// QuoteBox displays a Dr. Tour quote.
type QuoteBox struct {
	widget.BaseWidget

	quote   string
	minSize fyne.Size
}

// NewQuoteBox creates a new quote box.
func NewQuoteBox(quote string) *QuoteBox {
	q := &QuoteBox{
		quote:   quote,
		minSize: fyne.NewSize(300, 40),
	}
	q.ExtendBaseWidget(q)
	return q
}

// SetQuote updates the quote.
func (q *QuoteBox) SetQuote(quote string) {
	q.quote = quote
	q.Refresh()
}

// MinSize returns the minimum size.
func (q *QuoteBox) MinSize() fyne.Size {
	return q.minSize
}

// CreateRenderer implements fyne.Widget.
func (q *QuoteBox) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabelWithStyle(
		fmt.Sprintf("\"%s\" — Dr. external research group", q.quote),
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	label.Wrapping = fyne.TextWrapWord
	return widget.NewSimpleRenderer(label)
}

// KeyStatBox displays a key statistic prominently.
type KeyStatBox struct {
	widget.BaseWidget

	mu      sync.RWMutex
	label   string
	value   string
	minSize fyne.Size
}

// NewKeyStatBox creates a new key stat box.
func NewKeyStatBox(label, value string) *KeyStatBox {
	k := &KeyStatBox{
		label:   label,
		value:   value,
		minSize: fyne.NewSize(150, 60),
	}
	k.ExtendBaseWidget(k)
	return k
}

// SetValue updates the statistic value.
func (k *KeyStatBox) SetValue(value string) {
	k.mu.Lock()
	k.value = value
	k.mu.Unlock()
	k.Refresh()
}

// MinSize returns the minimum size.
func (k *KeyStatBox) MinSize() fyne.Size {
	return k.minSize
}

// CreateRenderer implements fyne.Widget.
func (k *KeyStatBox) CreateRenderer() fyne.WidgetRenderer {
	k.mu.RLock()
	label := k.label
	value := k.value
	k.mu.RUnlock()

	labelWidget := widget.NewLabel(label)
	labelWidget.Alignment = fyne.TextAlignCenter

	valueWidget := widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	box := container.NewVBox(labelWidget, valueWidget)
	return widget.NewSimpleRenderer(box)
}
