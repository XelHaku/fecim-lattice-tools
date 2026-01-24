// Package gui provides Fyne-based GUI components for crossbar visualization.
package gui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/shared/logging"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// Package-level logger for crossbar GUI
var log *logging.Logger

func init() {
	log = logging.NewLogger("crossbar")
}

// ControlPanel provides interactive controls for the crossbar demo.
type ControlPanel struct {
	widget.BaseWidget

	// Controls
	ArraySizeSlider *widget.Slider
	NoiseSlider     *widget.Slider
	ADCBitsSlider   *widget.Slider
	ColormapSelect  *widget.Select
	DemoModeSelect  *widget.Select

	// Buttons
	RunMVMButton       *widget.Button
	AnalyzeIRButton    *widget.Button
	AnalyzeSneakButton *widget.Button
	ResetButton        *widget.Button

	// Labels for current values
	arraySizeLabel *widget.Label
	noiseLabel     *widget.Label
	adcBitsLabel   *widget.Label

	// Fixed size
	minSize fyne.Size

	// Callbacks
	OnArraySizeChanged func(size int)
	OnNoiseChanged     func(noise float64)
	OnADCBitsChanged   func(bits int)
	OnColormapChanged  func(colormap string)
	OnDemoModeChanged  func(mode string)
	OnRunMVM           func()
	OnAnalyzeIR        func()
	OnAnalyzeSneak     func()
	OnReset            func()
}

// NewControlPanel creates a new control panel.
func NewControlPanel() *ControlPanel {
	cp := &ControlPanel{
		minSize: fyne.NewSize(200, 350), // Fixed size
	}

	// Array size slider (8 to 128)
	cp.arraySizeLabel = widget.NewLabel("Array Size: 64x64")
	cp.ArraySizeSlider = widget.NewSlider(8, 128)
	cp.ArraySizeSlider.Step = 8
	cp.ArraySizeSlider.Value = 64
	cp.ArraySizeSlider.OnChanged = func(v float64) {
		log.SliderChange("ArraySize", v)
		size := int(v)
		cp.arraySizeLabel.SetText(fmt.Sprintf("Array Size: %dx%d", size, size))
		if cp.OnArraySizeChanged != nil {
			cp.OnArraySizeChanged(size)
		}
	}

	// Noise level slider (0 to 20%)
	cp.noiseLabel = widget.NewLabel("Noise: 2.0%")
	cp.NoiseSlider = widget.NewSlider(0, 20)
	cp.NoiseSlider.Step = 0.5
	cp.NoiseSlider.Value = 2
	cp.NoiseSlider.OnChanged = func(v float64) {
		log.SliderChange("Noise", v)
		cp.noiseLabel.SetText(fmt.Sprintf("Noise: %.1f%%", v))
		if cp.OnNoiseChanged != nil {
			cp.OnNoiseChanged(v / 100.0)
		}
	}

	// ADC bits slider (4 to 10)
	cp.adcBitsLabel = widget.NewLabel("ADC Bits: 6")
	cp.ADCBitsSlider = widget.NewSlider(4, 10)
	cp.ADCBitsSlider.Step = 1
	cp.ADCBitsSlider.Value = 6
	cp.ADCBitsSlider.OnChanged = func(v float64) {
		log.SliderChange("ADCBits", v)
		bits := int(v)
		cp.adcBitsLabel.SetText(fmt.Sprintf("ADC Bits: %d", bits))
		if cp.OnADCBitsChanged != nil {
			cp.OnADCBitsChanged(bits)
		}
	}

	// Colormap selector
	cp.ColormapSelect = widget.NewSelect(
		[]string{"fecim", "viridis", "plasma", "coolwarm"},
		func(s string) {
			sharedwidgets.DebugInteraction(fmt.Sprintf("Crossbar ColormapSelect changed to '%s'", s))
			log.Selection("Colormap", s)
			if cp.OnColormapChanged != nil {
				cp.OnColormapChanged(s)
			}
		},
	)
	cp.ColormapSelect.SetSelected("fecim")

	// Demo mode selector
	cp.DemoModeSelect = widget.NewSelect(
		[]string{"Manual", "Auto Demo", "Step-by-Step"},
		func(s string) {
			sharedwidgets.DebugInteraction(fmt.Sprintf("Crossbar DemoModeSelect changed to '%s'", s))
			log.Selection("DemoMode", s)
			if cp.OnDemoModeChanged != nil {
				cp.OnDemoModeChanged(s)
			}
		},
	)
	cp.DemoModeSelect.SetSelected("Manual")

	// Action buttons
	cp.RunMVMButton = widget.NewButton("Run MVM", func() {
		log.Button("RunMVM")
		if cp.OnRunMVM != nil {
			cp.OnRunMVM()
		}
	})
	cp.RunMVMButton.Importance = widget.HighImportance

	cp.AnalyzeIRButton = widget.NewButton("Analyze IR Drop", func() {
		log.Button("AnalyzeIRDrop")
		if cp.OnAnalyzeIR != nil {
			cp.OnAnalyzeIR()
		}
	})

	cp.AnalyzeSneakButton = widget.NewButton("Analyze Sneak Paths", func() {
		log.Button("AnalyzeSneakPaths")
		if cp.OnAnalyzeSneak != nil {
			cp.OnAnalyzeSneak()
		}
	})

	cp.ResetButton = widget.NewButton("Reset Array", func() {
		log.Button("ResetArray")
		if cp.OnReset != nil {
			cp.OnReset()
		}
	})

	cp.ExtendBaseWidget(cp)
	return cp
}

// MinSize returns minimum size - small to allow flexible layout.
func (cp *ControlPanel) MinSize() fyne.Size {
	return fyne.NewSize(180, 200)
}

// CreateRenderer implements fyne.Widget.
func (cp *ControlPanel) CreateRenderer() fyne.WidgetRenderer {
	// Simplified content to fit in fixed height
	content := container.NewVBox(
		cp.RunMVMButton,
		cp.AnalyzeIRButton,
		cp.AnalyzeSneakButton,
		cp.ResetButton,
		widget.NewSeparator(),
		cp.arraySizeLabel,
		cp.ArraySizeSlider,
		widget.NewLabel("Colormap:"),
		cp.ColormapSelect,
	)

	return widget.NewSimpleRenderer(content)
}

// StatsPanel displays statistics and analysis results.
type StatsPanel struct {
	widget.BaseWidget

	// Labels
	titleLabel  *widget.Label
	statsText   *widget.Label
	progressBar *widget.ProgressBar

	// Data
	title    string
	stats    string
	progress float64
	minSize  fyne.Size
}

// NewStatsPanel creates a new statistics panel.
func NewStatsPanel(title string) *StatsPanel {
	sp := &StatsPanel{
		title:   title,
		minSize: fyne.NewSize(200, 250), // Fixed size to prevent resize
	}

	sp.titleLabel = widget.NewLabel(title)
	sp.titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	sp.statsText = widget.NewLabel("No data")
	sp.statsText.Wrapping = fyne.TextWrapOff // Prevent resize on content change
	sp.progressBar = widget.NewProgressBar()
	sp.progressBar.Hide()

	sp.ExtendBaseWidget(sp)
	return sp
}

// MinSize returns minimum size - small to allow flexible layout.
func (sp *StatsPanel) MinSize() fyne.Size {
	return fyne.NewSize(180, 150)
}

// SetStats updates the statistics display.
func (sp *StatsPanel) SetStats(stats string) {
	sp.stats = stats
	sp.statsText.SetText(stats)
}

// SetProgress shows/hides and updates the progress bar.
func (sp *StatsPanel) SetProgress(progress float64, show bool) {
	sp.progress = progress
	sp.progressBar.SetValue(progress)
	if show {
		sp.progressBar.Show()
	} else {
		sp.progressBar.Hide()
	}
}

// CreateRenderer implements fyne.Widget.
func (sp *StatsPanel) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewVBox(
		sp.titleLabel,
		widget.NewSeparator(),
		sp.statsText,
		sp.progressBar,
	)

	return widget.NewSimpleRenderer(content)
}

// InputVectorPanel allows editing of the input vector for MVM.
type InputVectorPanel struct {
	widget.BaseWidget

	entries []*widget.Entry
	size    int
	values  []float64

	OnValueChanged func(index int, value float64)
}

// NewInputVectorPanel creates a panel for editing input vectors.
func NewInputVectorPanel(size int) *InputVectorPanel {
	ivp := &InputVectorPanel{
		size:   size,
		values: make([]float64, size),
	}

	ivp.entries = make([]*widget.Entry, size)
	for i := 0; i < size; i++ {
		idx := i // Capture for closure
		ivp.entries[i] = widget.NewEntry()
		ivp.entries[i].SetPlaceHolder("0.0")
		ivp.entries[i].OnChanged = func(s string) {
			if v, err := strconv.ParseFloat(s, 64); err == nil {
				ivp.values[idx] = v
				if ivp.OnValueChanged != nil {
					ivp.OnValueChanged(idx, v)
				}
			}
		}
	}

	ivp.ExtendBaseWidget(ivp)
	return ivp
}

// SetValues sets all input values.
func (ivp *InputVectorPanel) SetValues(values []float64) {
	for i := 0; i < len(values) && i < ivp.size; i++ {
		ivp.values[i] = values[i]
		ivp.entries[i].SetText(fmt.Sprintf("%.3f", values[i]))
	}
}

// GetValues returns the current input values.
func (ivp *InputVectorPanel) GetValues() []float64 {
	return ivp.values
}

// CreateRenderer implements fyne.Widget.
func (ivp *InputVectorPanel) CreateRenderer() fyne.WidgetRenderer {
	// Show only first 8 entries for compact display
	maxShow := 8
	if ivp.size < maxShow {
		maxShow = ivp.size
	}

	objects := make([]fyne.CanvasObject, 0)
	objects = append(objects, widget.NewLabel("Input Vector:"))

	for i := 0; i < maxShow; i++ {
		row := container.NewHBox(
			widget.NewLabel(fmt.Sprintf("[%d]", i)),
			ivp.entries[i],
		)
		objects = append(objects, row)
	}

	if ivp.size > maxShow {
		objects = append(objects, widget.NewLabel(fmt.Sprintf("... +%d more", ivp.size-maxShow)))
	}

	content := container.NewVBox(objects...)
	return widget.NewSimpleRenderer(content)
}

// LevelIndicator shows the 30 discrete FeCIM levels.
type LevelIndicator struct {
	widget.BaseWidget

	currentLevel int
	levels       int
}

// NewLevelIndicator creates a new level indicator.
func NewLevelIndicator() *LevelIndicator {
	li := &LevelIndicator{
		levels:       30,
		currentLevel: 0,
	}
	li.ExtendBaseWidget(li)
	return li
}

// SetLevel sets the current level (0-29).
func (li *LevelIndicator) SetLevel(level int) {
	if level < 0 {
		level = 0
	}
	if level >= li.levels {
		level = li.levels - 1
	}
	li.currentLevel = level
	fyne.Do(func() {
		li.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (li *LevelIndicator) CreateRenderer() fyne.WidgetRenderer {
	labels := make([]fyne.CanvasObject, 0)
	labels = append(labels, widget.NewLabel("30 Discrete Levels:"))

	// Create level indicators
	levelStr := ""
	for i := 0; i < li.levels; i++ {
		if i == li.currentLevel {
			levelStr += "█"
		} else if i < li.currentLevel {
			levelStr += "▓"
		} else {
			levelStr += "░"
		}
	}

	levelLabel := widget.NewLabel(levelStr)
	levelLabel.TextStyle = fyne.TextStyle{Monospace: true}
	labels = append(labels, levelLabel)

	currentLabel := widget.NewLabel(fmt.Sprintf("Current: Level %d / %d", li.currentLevel, li.levels-1))
	labels = append(labels, currentLabel)

	content := container.NewVBox(labels...)
	return widget.NewSimpleRenderer(content)
}
