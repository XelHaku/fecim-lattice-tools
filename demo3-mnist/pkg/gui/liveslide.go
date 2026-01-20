// Package gui provides Fyne-based GUI components for MNIST visualization.
// This file implements the "Live Slide" pattern components for Demo 3.
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
	m.Refresh()
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
}

func (r *mnistModeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *mnistModeRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *mnistModeRenderer) Refresh() {
	r.indicator.mu.RLock()
	mode := r.indicator.mode
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.indicator.Size()

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
	e.Refresh()
}

// SetInferenceExplanation sets content for inference phases.
func (e *MNISTEducationalPanel) SetInferenceExplanation(phase int) {
	var content string
	switch phase {
	case 1:
		content = "NEURAL NETWORK INFERENCE\n\n" +
			"1. Input pixels (784)\n" +
			"   fed to crossbar\n\n" +
			"28×28 = 784 grayscale\n" +
			"values normalized 0-1"
	case 2:
		content = "NEURAL NETWORK INFERENCE\n\n" +
			"2. Hidden layer (128)\n" +
			"   MVM: I = G × V\n\n" +
			"100,352 multiplications\n" +
			"in ONE clock cycle!"
	case 3:
		content = "NEURAL NETWORK INFERENCE\n\n" +
			"3. Output layer (10)\n" +
			"   MVM: I = G × V\n\n" +
			"1,280 more MACs\n" +
			"Softmax → probabilities"
	default:
		content = "NEURAL NETWORK INFERENCE\n\n" +
			"Layer Flow:\n" +
			"Input (784) →\n" +
			"Hidden (128) →\n" +
			"Output (10)\n\n" +
			"Total: 101,632 MACs\n" +
			"in 2 clock cycles"
	}
	e.SetContent("Compute-in-Memory", content)
}

// SetDrawingExplanation sets content for drawing mode.
func (e *MNISTEducationalPanel) SetDrawingExplanation() {
	content := "DRAW A DIGIT\n\n" +
		"• Click and drag to draw\n" +
		"• Right-click to clear\n\n" +
		"The network runs inference\n" +
		"in real-time as you draw.\n\n" +
		"28×28 pixels → 784 inputs\n" +
		"Each pixel is normalized\n" +
		"to range 0.0 - 1.0"
	e.SetContent("Drawing Input", content)
}

// SetEvaluationExplanation sets content for evaluation.
func (e *MNISTEducationalPanel) SetEvaluationExplanation() {
	content := "FULL EVALUATION\n\n" +
		"Testing on MNIST dataset:\n" +
		"• 10,000 test images\n" +
		"• Ground truth labels\n\n" +
		"FeCIM Target:\n" +
		"87% accuracy\n" +
		"(88% theoretical max)\n\n" +
		"\"We're at 87% validation\n" +
		"here\" — Dr. external research group"
	e.SetContent("Network Accuracy", content)
}

// SetIdleExplanation sets content for idle state.
func (e *MNISTEducationalPanel) SetIdleExplanation() {
	content := "MNIST RECOGNITION\n\n" +
		"\"We're at 87% validation\n" +
		"here\"\n\n" +
		"— Dr. external research group\n\n" +
		"Architecture: 784→128→10\n" +
		"30 discrete analog levels\n" +
		"Compute-in-memory"
	e.SetContent("What You're Seeing", content)
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

	box := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		contentLabel,
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
	if len(o.entries) == 0 {
		o.contentLabel.SetText("Waiting for operations...")
		return
	}
	o.contentLabel.SetText(strings.Join(o.entries, "\n"))
}

// MinSize returns the minimum size.
func (o *MNISTOperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *MNISTOperationLog) CreateRenderer() fyne.WidgetRenderer {
	box := container.NewVBox(
		o.titleLabel,
		widget.NewSeparator(),
		o.contentLabel,
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
	p.Refresh()
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
}

func (r *predictionRenderer) MinSize() fyne.Size {
	return r.display.minSize
}

func (r *predictionRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *predictionRenderer) Refresh() {
	r.display.mu.RLock()
	pred := r.display.prediction
	conf := r.display.confidence
	r.display.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.display.Size()

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
	fontSize := size.Height * 0.45
	if fontSize > 48 {
		fontSize = 48
	}
	if fontSize < 20 {
		fontSize = 20
	}
	predLabel.TextSize = fontSize
	predLabel.TextStyle = fyne.TextStyle{Bold: true}
	textW := fontSize * 0.6
	predLabel.Move(fyne.NewPos((size.Width-textW)/2, size.Height*0.25))
	r.objects = append(r.objects, predLabel)

	// Confidence
	confText := fmt.Sprintf("%.1f%%", conf*100)
	if pred < 0 {
		confText = "-"
	}
	confLabel := canvas.NewText(confText, borderColor)
	confLabel.TextSize = 12
	confLabelW := float32(40)
	confLabel.Move(fyne.NewPos((size.Width-confLabelW)/2, size.Height-18))
	r.objects = append(r.objects, confLabel)
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
	k.Refresh()
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
