// Package gui provides Fyne-based GUI components for MNIST visualization.
// This file implements the "Live Slide" pattern components for Demo 3.
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

// MNISTMode represents the current demo mode.
type MNISTMode int

const (
	MNISTModeIdle MNISTMode = iota
	MNISTModeDrawing
	MNISTModeInference
	MNISTModeEvaluating
	MNISTModeLoading
)

func (m MNISTMode) String() string {
	switch m {
	case MNISTModeIdle:
		return "IDLE"
	case MNISTModeDrawing:
		return "DRAWING"
	case MNISTModeInference:
		return "INFERENCE"
	case MNISTModeEvaluating:
		return "EVALUATING"
	case MNISTModeLoading:
		return "LOADING"
	default:
		return "UNKNOWN"
	}
}

// MNISTModeIndicator shows the current mode with a colored background.
type MNISTModeIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	mode    MNISTMode
	minSize fyne.Size
}

// NewMNISTModeIndicator creates a new mode indicator.
func NewMNISTModeIndicator() *MNISTModeIndicator {
	m := &MNISTModeIndicator{
		mode:    MNISTModeIdle,
		minSize: fyne.NewSize(100, 40),
	}
	m.ExtendBaseWidget(m)
	return m
}

// SetMode updates the current mode.
func (m *MNISTModeIndicator) SetMode(mode MNISTMode) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	fyne.Do(func() {
		m.Refresh()
	})
}

// GetMode returns the current mode.
func (m *MNISTModeIndicator) GetMode() MNISTMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// MinSize returns the minimum size.
func (m *MNISTModeIndicator) MinSize() fyne.Size {
	return m.minSize
}

// CreateRenderer implements fyne.Widget.
func (m *MNISTModeIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &mnistModeRenderer{indicator: m}
}

type mnistModeRenderer struct {
	indicator *MNISTModeIndicator
	objects   []fyne.CanvasObject
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *mnistModeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *mnistModeRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("mnistModeRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *mnistModeRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("mnistModeRenderer", r.indicator.Size())
	size := r.indicator.Size()
	// Always re-layout on Refresh for this dynamic widget (mode changes)
	r.layoutWithSize(size)
}

func (r *mnistModeRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.indicator.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
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
	case MNISTModeIdle:
		bgColor = color.RGBA{60, 60, 80, 255}
		borderColor = color.RGBA{100, 100, 130, 255}
		modeText = "IDLE"
	case MNISTModeDrawing:
		bgColor = color.RGBA{80, 50, 150, 255}
		borderColor = color.RGBA{140, 100, 220, 255}
		modeText = "DRAWING"
	case MNISTModeInference:
		bgColor = color.RGBA{50, 120, 180, 255}
		borderColor = color.RGBA{100, 180, 255, 255}
		modeText = "INFERENCE"
	case MNISTModeEvaluating:
		bgColor = color.RGBA{180, 120, 50, 255}
		borderColor = color.RGBA{255, 180, 100, 255}
		modeText = "EVALUATING"
	case MNISTModeLoading:
		bgColor = color.RGBA{50, 150, 80, 255}
		borderColor = color.RGBA{100, 220, 130, 255}
		modeText = "LOADING"
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

	// Mode text - scale with widget size
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

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
}

func (r *mnistModeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *mnistModeRenderer) Destroy() {}

// MNISTEducationalPanel shows context-sensitive explanations.
type MNISTEducationalPanel struct {
	widget.BaseWidget

	mu      sync.RWMutex
	title   string
	content string
	minSize fyne.Size
}

// NewMNISTEducationalPanel creates a new educational panel.
func NewMNISTEducationalPanel() *MNISTEducationalPanel {
	e := &MNISTEducationalPanel{
		title:   "What You're Seeing",
		content: "Draw a digit to see\nneural network inference.",
		minSize: fyne.NewSize(150, 150),
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetContent updates the educational content.
func (e *MNISTEducationalPanel) SetContent(title, content string) {
	e.mu.Lock()
	e.title = title
	e.content = content
	e.mu.Unlock()
	fyne.Do(func() {
		e.Refresh()
	})
}

// SetInferenceExplanation sets content for inference phases.
func (e *MNISTEducationalPanel) SetInferenceExplanation(phase int) {
	var title, content string
	switch phase {
	case 1:
		title = "Phase 1: Input"
		content = "Your drawing → 784 pixels\n\n" +
			"Each pixel becomes a voltage.\n" +
			"All 784 applied at once.\n\n" +
			"Traditional CPU: sequential\n" +
			"FeCIM: ALL AT ONCE"
	case 2:
		title = "Phase 2: Hidden Layer"
		content = "MVM: I = G × V\n\n" +
			"100,352 multiplications\n" +
			"in ONE clock cycle!\n\n" +
			"Physics does the math.\n" +
			"No fetching from memory."
	case 3:
		title = "Phase 3: Output"
		content = "10 outputs = 10 digits\n\n" +
			"Highest score wins.\n\n" +
			"Total: 101,632 MACs\n" +
			"Just 2 clock cycles.\n\n" +
			"That's compute-in-memory."
	default:
		title = "Inference Complete"
		content = "784 → 128 → 10\n\n" +
			"101,632 operations\n" +
			"2 clock cycles\n" +
			"Near-zero energy\n\n" +
			"Draw another digit!"
	}
	e.SetContent(title, content)
}

// SetDrawingExplanation sets content for drawing mode.
func (e *MNISTEducationalPanel) SetDrawingExplanation() {
	content := "Click and drag to draw.\n" +
		"Right-click to clear.\n\n" +
		"As you draw, the network\n" +
		"runs inference instantly.\n\n" +
		"28×28 = 784 inputs\n" +
		"Normalized 0.0 to 1.0\n\n" +
		"Watch the layers light up!"
	e.SetContent("Draw a Digit", content)
}

// SetEvaluationExplanation sets content for evaluation.
func (e *MNISTEducationalPanel) SetEvaluationExplanation() {
	content := "Testing on 1000 digits...\n\n" +
		"Each digit runs through\n" +
		"the full network.\n\n" +
		"Target: 87% accuracy\n" +
		"(88% theoretical max)"
	e.SetContent("Evaluating Network", content)
}

// SetIdleExplanation sets content for idle state.
func (e *MNISTEducationalPanel) SetIdleExplanation() {
	content := "Draw a digit (0-9) or\n" +
		"click Random Test.\n\n" +
		"The FeCIM chip recognizes\n" +
		"handwritten numbers with\n" +
		"87% accuracy.\n\n" +
		"784 → 128 → 10 neurons\n" +
		"30 analog levels per cell\n" +
		"2 clock cycles total"
	e.SetContent("MNIST Recognition", content)
}

// MinSize returns the minimum size.
func (e *MNISTEducationalPanel) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *MNISTEducationalPanel) CreateRenderer() fyne.WidgetRenderer {
	e.mu.RLock()
	title := e.title
	content := e.content
	e.mu.RUnlock()

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	contentLabel := widget.NewLabel(content)
	contentLabel.Wrapping = fyne.TextWrapWord

	// Wrap contentLabel in scroll container to prevent resize loops from text wrapping
	contentScroll := container.NewScroll(contentLabel)
	contentScroll.SetMinSize(fyne.NewSize(190, 140))

	box := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		contentScroll,
	)

	return widget.NewSimpleRenderer(box)
}

// MNISTOperationLog shows timestamped operation history.
type MNISTOperationLog struct {
	widget.BaseWidget

	mu         sync.RWMutex
	entries    []string
	maxEntries int
	startTime  time.Time
	minSize    fyne.Size

	titleLabel   *widget.Label
	contentLabel *widget.Label
}

// NewMNISTOperationLog creates a new operation log.
func NewMNISTOperationLog() *MNISTOperationLog {
	o := &MNISTOperationLog{
		maxEntries: 10,
		startTime:  time.Now(),
		minSize:    fyne.NewSize(150, 120),
		entries:    make([]string, 0, 10),
	}
	o.titleLabel = widget.NewLabelWithStyle("Operation Log", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	o.contentLabel = widget.NewLabel("Waiting for operations...")
	o.contentLabel.Wrapping = fyne.TextWrapWord
	o.ExtendBaseWidget(o)
	return o
}

// Add adds a new log entry with timestamp.
func (o *MNISTOperationLog) Add(entry string) {
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

// AddPrediction adds a prediction result entry.
func (o *MNISTOperationLog) AddPrediction(predicted int, confidence float64) {
	o.Add(fmt.Sprintf("Predict → %d (%.1f%%)", predicted, confidence*100))
}

// Clear clears all log entries.
func (o *MNISTOperationLog) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.entries = o.entries[:0]
	o.startTime = time.Now()
	o.updateContent()
}

func (o *MNISTOperationLog) updateContent() {
	text := "Waiting for operations..."
	if len(o.entries) > 0 {
		text = strings.Join(o.entries, "\n")
	}
	fyne.Do(func() {
		o.contentLabel.SetText(text)
	})
}

// MinSize returns the minimum size.
func (o *MNISTOperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *MNISTOperationLog) CreateRenderer() fyne.WidgetRenderer {
	// Wrap contentLabel in scroll container to prevent resize loops from text wrapping
	contentScroll := container.NewScroll(o.contentLabel)
	contentScroll.SetMinSize(fyne.NewSize(140, 90))

	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		contentScroll,
	)
	return widget.NewSimpleRenderer(box)
}

// PredictionDisplay shows the big prediction number prominently.
type PredictionDisplay struct {
	widget.BaseWidget

	mu         sync.RWMutex
	prediction int
	confidence float64
	minSize    fyne.Size
}

// NewPredictionDisplay creates a new prediction display.
func NewPredictionDisplay() *PredictionDisplay {
	p := &PredictionDisplay{
		prediction: -1,
		confidence: 0,
		minSize:    fyne.NewSize(100, 80),
	}
	p.ExtendBaseWidget(p)
	return p
}

// SetPrediction updates the displayed prediction.
func (p *PredictionDisplay) SetPrediction(pred int, conf float64) {
	p.mu.Lock()
	p.prediction = pred
	p.confidence = conf
	p.mu.Unlock()
	fyne.Do(func() {
		p.Refresh()
	})
}

// MinSize returns the minimum size.
func (p *PredictionDisplay) MinSize() fyne.Size {
	return p.minSize
}

// CreateRenderer implements fyne.Widget.
func (p *PredictionDisplay) CreateRenderer() fyne.WidgetRenderer {
	return &predictionRenderer{display: p}
}

type predictionRenderer struct {
	display *PredictionDisplay
	objects []fyne.CanvasObject
	cache   sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *predictionRenderer) MinSize() fyne.Size {
	return r.display.minSize
}

func (r *predictionRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("predictionRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *predictionRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("predictionRenderer", r.display.Size())
	size := r.display.Size()
	// Always re-layout on Refresh for this dynamic widget (prediction changes)
	r.layoutWithSize(size)
}

func (r *predictionRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.display.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.display.mu.RLock()
	pred := r.display.prediction
	conf := r.display.confidence
	r.display.mu.RUnlock()

	r.objects = r.objects[:0]

	// Constrain to minimum size to prevent growing
	minSize := r.display.minSize
	if size.Width > minSize.Width {
		size.Width = minSize.Width
	}
	if size.Height > minSize.Height {
		size.Height = minSize.Height
	}

	// Background
	bgColor := color.RGBA{30, 30, 45, 255}
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Border with confidence-based color
	var borderColor color.RGBA
	if pred < 0 {
		borderColor = color.RGBA{80, 80, 100, 255}
	} else if conf > 0.9 {
		borderColor = color.RGBA{80, 220, 150, 255} // High confidence: green
	} else if conf > 0.7 {
		borderColor = color.RGBA{220, 200, 80, 255} // Medium: yellow
	} else {
		borderColor = color.RGBA{220, 100, 80, 255} // Low: red
	}

	border := canvas.NewRectangle(borderColor)
	border.StrokeWidth = 3
	border.FillColor = color.Transparent
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Title - positioned relative to size
	title := canvas.NewText("PREDICTION", color.RGBA{180, 180, 200, 255})
	title.TextSize = 10
	titleW := float32(70)
	title.Move(fyne.NewPos((size.Width-titleW)/2, 5))
	r.objects = append(r.objects, title)

	// Big number - scale font with widget size
	var predText string
	if pred < 0 {
		predText = "?"
	} else {
		predText = fmt.Sprintf("%d", pred)
	}
	predLabel := canvas.NewText(predText, color.White)
	fontSize := size.Height * 0.35
	if fontSize > 42 {
		fontSize = 42
	}
	if fontSize < 18 {
		fontSize = 18
	}
	predLabel.TextSize = fontSize
	predLabel.TextStyle = fyne.TextStyle{Bold: true}
	textW := fontSize * 0.6
	predLabel.Move(fyne.NewPos((size.Width-textW)/2, size.Height*0.18))
	r.objects = append(r.objects, predLabel)

	// Confidence meter bar - visual indicator
	meterPadding := float32(10)
	meterHeight := float32(8)
	meterY := size.Height*0.55 + 5
	meterWidth := size.Width - 2*meterPadding

	// Meter background (track)
	meterBg := canvas.NewRectangle(color.RGBA{50, 50, 70, 255})
	meterBg.Resize(fyne.NewSize(meterWidth, meterHeight))
	meterBg.Move(fyne.NewPos(meterPadding, meterY))
	r.objects = append(r.objects, meterBg)

	// Meter fill (confidence level)
	if pred >= 0 && conf > 0 {
		fillWidth := meterWidth * float32(conf)
		meterFill := canvas.NewRectangle(borderColor)
		meterFill.Resize(fyne.NewSize(fillWidth, meterHeight))
		meterFill.Move(fyne.NewPos(meterPadding, meterY))
		r.objects = append(r.objects, meterFill)
	}

	// Confidence text below meter
	confText := fmt.Sprintf("%.1f%%", conf*100)
	if pred < 0 {
		confText = "-"
	}
	confLabel := canvas.NewText(confText, borderColor)
	confLabel.TextSize = 11
	confLabelW := float32(40)
	confLabel.Move(fyne.NewPos((size.Width-confLabelW)/2, meterY+meterHeight+4))
	r.objects = append(r.objects, confLabel)

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
}

func (r *predictionRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *predictionRenderer) Destroy() {}

// MNISTKeyStat displays a key statistic prominently.
type MNISTKeyStat struct {
	widget.BaseWidget

	mu      sync.RWMutex
	label   string
	value   string
	minSize fyne.Size
}

// NewMNISTKeyStat creates a new key stat box.
func NewMNISTKeyStat(label, value string) *MNISTKeyStat {
	k := &MNISTKeyStat{
		label:   label,
		value:   value,
		minSize: fyne.NewSize(100, 50),
	}
	k.ExtendBaseWidget(k)
	return k
}

// SetValue updates the statistic value.
func (k *MNISTKeyStat) SetValue(value string) {
	k.mu.Lock()
	k.value = value
	k.mu.Unlock()
	fyne.Do(func() {
		k.Refresh()
	})
}

// MinSize returns the minimum size.
func (k *MNISTKeyStat) MinSize() fyne.Size {
	return k.minSize
}

// CreateRenderer implements fyne.Widget.
func (k *MNISTKeyStat) CreateRenderer() fyne.WidgetRenderer {
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
