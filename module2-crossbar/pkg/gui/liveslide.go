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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
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

// newModeIndicator creates a shared ModeIndicator configured for crossbar demo modes.
func newModeIndicator() *sharedwidgets.ModeIndicator {
	return sharedwidgets.NewModeIndicator(sharedwidgets.ModeIndicatorConfig{
		MinSize: fyne.NewSize(100, 30),
		Styles: map[int]sharedwidgets.ModeStyle{
			int(DemoModeIdle): {
				Text:            "○ IDLE",
				BackgroundColor: color.RGBA{60, 60, 80, 255},
				BorderColor:     color.RGBA{100, 100, 130, 255},
			},
			int(DemoModeCompute): {
				Text:            "▶ COMPUTE",
				BackgroundColor: color.RGBA{50, 120, 180, 255},
				BorderColor:     color.RGBA{100, 180, 255, 255},
			},
			int(DemoModeWrite): {
				Text:            "↓ WRITE",
				BackgroundColor: color.RGBA{180, 50, 50, 255},
				BorderColor:     color.RGBA{255, 100, 100, 255},
			},
			int(DemoModeRead): {
				Text:            "↑ READ",
				BackgroundColor: color.RGBA{50, 150, 80, 255},
				BorderColor:     color.RGBA{100, 220, 130, 255},
			},
			int(DemoModeIRDrop): {
				Text:            "~ IR DROP",
				BackgroundColor: color.RGBA{180, 120, 50, 255},
				BorderColor:     color.RGBA{255, 180, 100, 255},
			},
			int(DemoModeSneakPath): {
				Text:            "⌇ SNEAK",
				BackgroundColor: color.RGBA{150, 50, 150, 255},
				BorderColor:     color.RGBA{220, 100, 220, 255},
			},
		},
	})
}

// newEducationalPanel creates a shared EducationalPanel for crossbar demo explanations.
func newEducationalPanel() *sharedwidgets.EducationalPanel {
	return sharedwidgets.NewEducationalPanel(sharedwidgets.EducationalPanelConfig{
		Title:   "What You're Seeing",
		Content: "Select an operation to see\nwhat's happening.",
		MinSize: fyne.NewSize(220, 280),
	})
}

// setMVMExplanation sets MVM operation content on the given panel.
func setMVMExplanation(ep *sharedwidgets.EducationalPanel, phase int) {
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
			"Analog step: ~1ns settling\n" +
			"Energy: ~10pJ/MVM (modeled)"
	}
	ep.SetContent("Compute-in-Memory", content)
}

// setIRDropExplanation sets IR drop content on the given panel.
func setIRDropExplanation(ep *sharedwidgets.EducationalPanel) {
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
	ep.SetContent("Non-Ideality: IR Drop", content)
}

// setSneakPathExplanation sets sneak path content on the given panel.
func setSneakPathExplanation(ep *sharedwidgets.EducationalPanel) {
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
	ep.SetContent("Non-Ideality: Sneak Paths", content)
}

// setIdleExplanation sets idle content on the given panel.
func setIdleExplanation(ep *sharedwidgets.EducationalPanel) {
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
	ep.SetContent("What You're Seeing", content)
}

// newOperationLog creates a shared OperationLog for crossbar demo operations.
func newOperationLog() *sharedwidgets.OperationLog {
	return sharedwidgets.NewOperationLog(sharedwidgets.OperationLogConfig{
		Title:        "Operation Log",
		MaxEntries:   6,
		MinSize:      fyne.NewSize(150, 60),
		UseMonospace: false,
	})
}

// addOperationWithResult adds an operation entry with a success/failure indicator.
func addOperationWithResult(log *sharedwidgets.OperationLog, action, result string, success bool) {
	indicator := "✓"
	if !success {
		indicator = "✗"
	}
	log.Add(fmt.Sprintf("%s → %s %s", action, result, indicator))
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

// newKeyStatBox creates a shared KeyStat with crossbar demo styling.
func newKeyStatBox(label, value string) *sharedwidgets.KeyStat {
	return sharedwidgets.NewKeyStat(sharedwidgets.KeyStatConfig{
		Label:   label,
		Value:   value,
		MinSize: fyne.NewSize(220, 60),
	})
}
