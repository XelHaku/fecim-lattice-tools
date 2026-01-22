// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// Uses fyne.io/fyne/v2 for cross-platform native GUI with proper graphics.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/ferroelectric"
)

// Colors - FeCIM theme
var (
	colorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan
	colorSecondary  = color.RGBA{255, 107, 107, 255} // Coral red
	colorAccent     = color.RGBA{78, 205, 196, 255}  // Teal
	colorWarning    = color.RGBA{255, 230, 109, 255} // Yellow
	colorBackground = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
	colorGrid       = color.RGBA{0, 70, 130, 128}    // Grid lines (lighter blue)
	colorAxis       = color.RGBA{150, 180, 200, 255} // Axis lines
	colorPositive   = color.RGBA{255, 100, 100, 255} // Positive polarization
	colorNegative   = color.RGBA{100, 150, 255, 255} // Negative polarization
)

// App holds the main application state
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window

	// Physics
	material  *ferroelectric.HZOMaterial
	preisach  *ferroelectric.MayergoyzPreisach
	materials []*ferroelectric.HZOMaterial
	matIndex  int

	// Simulation state
	mu            sync.RWMutex
	electricField float64
	polarization  float64
	normalizedP   float64
	discreteLevel int

	// History for plotting
	eHistory   []float64
	pHistory   []float64
	maxHistory int

	// UI state
	running   bool
	paused    bool
	autoMode  bool
	waveform  WaveformType
	frequency float64
	simTime   float64

	// Random Walk state
	rwTargetLevel int     // Target level (1-30)
	rwCurrentE    float64 // Current E-field
	rwStepTimer   float64 // Time until next random level
	rwStepDelay   float64 // How often to pick new level

	// Write/Read Demo state
	wrdPhase       int     // 0=write, 1=hold, 2=read, 3=display
	wrdTargetLevel int     // Target level to write
	wrdReadLevel   int     // Level read back
	wrdPhaseTimer  float64 // Time in current phase
	wrdWriteE      float64 // E-field during write

	// UI components
	plot           *PEPlot
	levelIndicator *LevelIndicator
	cellViz        *CellVisualizer
	modeIndicator  *ModeIndicator // WRITE/READ mode indicator with colored box
	eFieldSlider   *widget.Slider
	eFieldLabel    *widget.Label
	pLabel         *widget.Label
	levelLabel     *widget.Label
	materialSelect *widget.Select
	waveformSelect *widget.Select
	statusLabel    *widget.Label
	pauseBtn       *widget.Button

	// Educational slide and log
	slideTitle   *widget.Label
	slideText    *widget.Label
	logText      *widget.Label
	logEntries   []string
	maxLogLines  int
	lastLogPhase int // Track phase changes to avoid duplicate logs
	lastRWTarget int // Track Random Walk target changes
}

// WaveformType represents the input waveform
type WaveformType int

const (
	WaveformManual WaveformType = iota
	WaveformSine
	WaveformTriangle
	WaveformSquare
	WaveformRandomWalk
	WaveformWriteReadDemo
)

func (w WaveformType) String() string {
	switch w {
	case WaveformManual:
		return "Manual"
	case WaveformSine:
		return "Sine Wave"
	case WaveformTriangle:
		return "Triangle Wave"
	case WaveformSquare:
		return "Square Wave"
	case WaveformRandomWalk:
		return "Random Walk"
	case WaveformWriteReadDemo:
		return "Write/Read Demo"
	default:
		return "Unknown"
	}
}

// NewApp creates a new GUI application
func NewApp() *App {
	materials := []*ferroelectric.HZOMaterial{
		ferroelectric.DefaultHZO(),
		ferroelectric.OptimizedHZO(),
		ferroelectric.FeCIMMaterial(),
	}

	mat := materials[0]
	preisach := ferroelectric.NewMayergoyzPreisach(mat, 30)

	return &App{
		material:       mat,
		preisach:       preisach,
		materials:      materials,
		matIndex:       0,
		maxHistory:     500,
		eHistory:       make([]float64, 0, 500),
		pHistory:       make([]float64, 0, 500),
		autoMode:       true,
		waveform:       WaveformSine,
		frequency:      0.5, // 0.5 Hz default
		rwTargetLevel:  25,
		rwStepDelay:    2.5, // 2.5 seconds between random level changes
		wrdTargetLevel: 28,  // Start high for dramatic first write
		maxLogLines:    12,
		logEntries:     make([]string, 0, 12),
		lastLogPhase:   -1,
	}
}

// Run starts the GUI application
func Run() error {
	a := NewApp()
	return a.run()
}

func (a *App) run() error {
	a.fyneApp = app.New()
	a.fyneApp.Settings().SetTheme(&feCIMTheme{})

	a.mainWindow = a.fyneApp.NewWindow("FeCIM Hysteresis Visualizer - Demo 1")
	a.mainWindow.Resize(fyne.NewSize(1280, 900))

	// Create UI components
	content := a.createUI()
	a.mainWindow.SetContent(content)

	// Start simulation loop
	a.running = true
	go a.simulationLoop()

	a.mainWindow.ShowAndRun()
	a.running = false
	return nil
}

func (a *App) createUI() fyne.CanvasObject {
	// Create cell visualizer (THE memory cell!)
	a.cellViz = NewCellVisualizer()
	a.cellViz.SetMinSize(fyne.NewSize(140, 180))

	// Create P-E plot
	a.plot = NewPEPlot(a.material.Ec*2.5, a.material.Ps*1.2)
	a.plot.SetMinSize(fyne.NewSize(400, 350))

	// Create level indicator
	a.levelIndicator = NewLevelIndicator()
	a.levelIndicator.SetMinSize(fyne.NewSize(70, 350))

	// Create controls panel
	controls := a.createControlsPanel()

	// Create info panel
	info := a.createInfoPanel()

	// Create educational slide panel
	slidePanel := a.createSlidePanel()

	// Create log panel
	logPanel := a.createLogPanel()

	// Cell container with label
	cellContainer := container.NewVBox(
		widget.NewLabelWithStyle("Memory Cell", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		a.cellViz,
		widget.NewLabelWithStyle("This is the cell", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
	)

	// Controls panel (22% of window width)
	controlsCol := container.NewScroll(container.NewVBox(
		controls,
		widget.NewSeparator(),
		info,
	))

	// Slide + Log panel (12% of window width)
	slideLogCol := container.NewScroll(container.NewVBox(
		slidePanel,
		widget.NewSeparator(),
		logPanel,
	))

	// Plot and level indicator side by side with shared title row
	plotTitle := widget.NewLabelWithStyle("P-E Hysteresis Loop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	levelTitle := widget.NewLabelWithStyle("30 Levels", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	levelBits := widget.NewLabel(fmt.Sprintf("%.1f bits", math.Log2(30)))

	// Plot + level in same row (level has fixed width)
	plotAndLevel := container.NewBorder(
		nil, nil,
		nil,
		a.levelIndicator,
		a.plot,
	)

	// Titles row
	titlesRow := container.NewBorder(
		nil, nil,
		nil,
		container.NewVBox(levelTitle, levelBits),
		plotTitle,
	)

	// Center area with titles on top
	centerArea := container.NewBorder(
		titlesRow,
		nil, nil, nil,
		plotAndLevel,
	)

	// Status bar at bottom
	a.statusLabel = widget.NewLabel("Running...")
	statusBar := container.NewHBox(
		layout.NewSpacer(),
		a.statusLabel,
		layout.NewSpacer(),
	)

	// Main layout using percentage-based widths
	// Cell: 12%, Plot: 54%, Controls: 22%, Slide/Log: 12%
	mainLayout := container.New(
		&percentHBoxLayout{weights: []float32{0.12, 0.54, 0.22, 0.12}},
		cellContainer,
		centerArea,
		controlsCol,
		slideLogCol,
	)

	return container.NewBorder(
		a.createHeader(),
		statusBar,
		nil, nil,
		mainLayout,
	)
}

func (a *App) createHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"FeCIM Ferroelectric Hysteresis Visualization",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	return container.NewVBox(title, widget.NewSeparator())
}

func (a *App) createControlsPanel() fyne.CanvasObject {
	// E-field slider
	a.eFieldSlider = widget.NewSlider(-2, 2)
	a.eFieldSlider.Step = 0.01
	a.eFieldSlider.Value = 0
	a.eFieldSlider.OnChanged = func(v float64) {
		if a.waveform == WaveformManual {
			a.mu.Lock()
			a.electricField = v * a.material.Ec
			a.mu.Unlock()
		}
	}
	a.eFieldLabel = widget.NewLabel("E: 0.00 MV/cm")

	// Waveform selector
	waveforms := []string{"Manual", "Sine Wave", "Triangle Wave", "Square Wave", "Random Walk", "Write/Read Demo"}
	a.waveformSelect = widget.NewSelect(waveforms, func(s string) {
		a.mu.Lock()
		defer a.mu.Unlock()
		switch s {
		case "Manual":
			a.waveform = WaveformManual
			a.autoMode = false
			a.eFieldSlider.Enable()
		case "Sine Wave":
			a.waveform = WaveformSine
			a.autoMode = true
			a.eFieldSlider.Disable()
		case "Triangle Wave":
			a.waveform = WaveformTriangle
			a.autoMode = true
			a.eFieldSlider.Disable()
		case "Square Wave":
			a.waveform = WaveformSquare
			a.autoMode = true
			a.eFieldSlider.Disable()
		case "Random Walk":
			a.waveform = WaveformRandomWalk
			a.autoMode = true
			a.eFieldSlider.Disable()
			// Reset random walk state
			a.rwStepTimer = 0
			a.rwCurrentE = a.electricField
			a.rwTargetLevel = rand.Intn(30) + 1
		case "Write/Read Demo":
			a.waveform = WaveformWriteReadDemo
			a.autoMode = true
			a.eFieldSlider.Disable()
			// Reset write/read demo state
			a.wrdPhase = 0
			a.wrdPhaseTimer = 0
			a.wrdTargetLevel = rand.Intn(30) + 1
		}
	})
	a.waveformSelect.SetSelected("Sine Wave")

	// Material selector
	matNames := []string{"Default HZO", "Optimized Superlattice", "FeCIM HZO"}
	a.materialSelect = widget.NewSelect(matNames, func(s string) {
		var idx int
		switch s {
		case "Default HZO":
			idx = 0
		case "Optimized Superlattice":
			idx = 1
		case "FeCIM HZO":
			idx = 2
		}
		a.mu.Lock()
		a.matIndex = idx
		a.material = a.materials[idx]
		a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 30)
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.plot.SetBounds(a.material.Ec*2.5, a.material.Ps*1.2)
		a.mu.Unlock()
	})
	a.materialSelect.SetSelected("Default HZO")

	// Pause/Resume button
	a.pauseBtn = widget.NewButton("Pause", func() {
		a.paused = !a.paused
		if a.paused {
			a.pauseBtn.SetText("Resume")
		} else {
			a.pauseBtn.SetText("Pause")
		}
	})

	// Reset button
	resetBtn := widget.NewButton("Reset", func() {
		a.mu.Lock()
		a.preisach.Reset()
		a.electricField = 0
		a.polarization = 0
		a.normalizedP = 0
		a.discreteLevel = 15
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.eFieldSlider.SetValue(0)
		a.mu.Unlock()
	})

	// Frequency slider
	freqSlider := widget.NewSlider(0.01, 1.0)
	freqSlider.Step = 0.01
	freqSlider.Value = 0.5
	freqLabel := widget.NewLabel("Freq: 0.50 Hz")
	freqSlider.OnChanged = func(v float64) {
		a.mu.Lock()
		a.frequency = v
		// Reset trail when frequency changes
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.mu.Unlock()
		freqLabel.SetText(fmt.Sprintf("Freq: %.2f Hz", v))
	}

	// Trail length slider
	trailSlider := widget.NewSlider(50, 2000)
	trailSlider.Step = 50
	trailSlider.Value = float64(a.maxHistory)
	trailLabel := widget.NewLabel(fmt.Sprintf("Trail: %d", a.maxHistory))
	trailSlider.OnChanged = func(v float64) {
		a.mu.Lock()
		a.maxHistory = int(v)
		// Immediately trim if history exceeds new max
		if len(a.eHistory) > a.maxHistory {
			a.eHistory = a.eHistory[len(a.eHistory)-a.maxHistory:]
			a.pHistory = a.pHistory[len(a.pHistory)-a.maxHistory:]
		}
		a.mu.Unlock()
		trailLabel.SetText(fmt.Sprintf("Trail: %d", int(v)))
	}

	// Grouped sections for cleaner layout
	// Simulation Mode group
	modeGroup := container.NewVBox(
		widget.NewLabelWithStyle("Mode", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		a.materialSelect,
		a.waveformSelect,
	)

	// E-field control group
	eFieldGroup := container.NewVBox(
		widget.NewLabelWithStyle("E-field (×Ec)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		a.eFieldSlider,
		a.eFieldLabel,
	)

	// Timing group - frequency and trail on same rows
	timingGroup := container.NewVBox(
		widget.NewLabelWithStyle("Timing", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2, freqLabel, trailLabel),
		freqSlider,
		trailSlider,
	)

	// Action buttons
	actionGroup := container.NewHBox(a.pauseBtn, resetBtn)

	return container.NewVBox(
		modeGroup,
		widget.NewSeparator(),
		eFieldGroup,
		widget.NewSeparator(),
		timingGroup,
		widget.NewSeparator(),
		actionGroup,
	)
}

func (a *App) createInfoPanel() fyne.CanvasObject {
	a.pLabel = widget.NewLabel("P: 0.00 µC/cm²")
	a.levelLabel = widget.NewLabel("Level: 15/30")
	a.modeIndicator = NewModeIndicator()
	a.modeIndicator.SetMinSize(fyne.NewSize(160, 50))

	// State display - compact 2-column grid
	stateGrid := container.NewGridWithColumns(2,
		widget.NewLabel("P:"), a.pLabel,
		widget.NewLabel("Level:"), a.levelLabel,
	)

	// Material params - compact display
	matParams := widget.NewLabel(fmt.Sprintf(
		"Pr=%.0f Ps=%.0f µC/cm²\nEc=%.2f MV/cm τ=%.0fns\nEndurance: %.0e",
		a.material.Pr*100, a.material.Ps*100,
		a.material.Ec/1e8, a.material.Tau*1e9,
		a.material.EnduranceCycles,
	))
	matParams.Wrapping = fyne.TextWrapOff

	return container.NewVBox(
		widget.NewLabelWithStyle("State", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		stateGrid,
		widget.NewSeparator(),
		a.modeIndicator,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Material", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		matParams,
	)
}

func (a *App) createSlidePanel() fyne.CanvasObject {
	a.slideTitle = widget.NewLabelWithStyle("What You're Seeing", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	a.slideText = widget.NewLabel(a.getSlideText())
	a.slideText.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		a.slideTitle,
		widget.NewSeparator(),
		a.slideText,
	)
}

func (a *App) createLogPanel() fyne.CanvasObject {
	a.logText = widget.NewLabel("Memory operations will\nappear here...")
	a.logText.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		widget.NewLabelWithStyle("Memory Log", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		a.logText,
	)
}

func (a *App) getSlideText() string {
	a.mu.RLock()
	level := a.discreteLevel + 1 // 1-indexed for display
	wrdPhase := a.wrdPhase
	wrdTarget := a.wrdTargetLevel
	isWrite := math.Abs(a.electricField) > a.material.Ec
	a.mu.RUnlock()

	switch a.waveform {
	case WaveformManual:
		if isWrite {
			return fmt.Sprintf("██ WRITING LEVEL %d ██\n\n"+
				"Electric field E > Ec.\n"+
				"Domains are switching.\n"+
				"Polarization is changing.\n\n"+
				"This is how we STORE data.", level)
		}
		return fmt.Sprintf("░░ HOLDING LEVEL %d ░░\n\n"+
			"E-field is low or zero.\n"+
			"Polarization PERSISTS.\n"+
			"No power needed.\n\n"+
			"This is NON-VOLATILE.", level)

	case WaveformSine, WaveformTriangle, WaveformSquare:
		phaseText := "░░ READING ░░"
		if isWrite {
			phaseText = "██ WRITING ██"
		}
		return fmt.Sprintf("%s\n\n"+
			"Level: %d/30\n\n"+
			"The P-E loop shows hysteresis:\n"+
			"• Upper branch: E increasing\n"+
			"• Lower branch: E decreasing\n"+
			"• Area inside = energy loss\n\n"+
			"The SQUARE shape means:\n"+
			"sharp switching at ±Ec.", phaseText, level)

	case WaveformRandomWalk:
		return fmt.Sprintf("RANDOM WALK\n\n"+
			"Current: Level %d\n\n"+
			"Picking random targets\n"+
			"and ramping to reach them.\n\n"+
			"This shows multi-level\n"+
			"storage: not just 0/1,\n"+
			"but 30 distinct states!\n\n"+
			"─────────────────\n"+
			"WHY 30 LEVELS?\n"+
			"Binary: 1 bit = 2 states\n"+
			"This: 4.9 bits = 30 states\n"+
			"Same chip = 5× storage", level)

	case WaveformWriteReadDemo:
		var phaseExplanation string
		switch wrdPhase {
		case 0: // WRITE
			phaseExplanation = fmt.Sprintf("██ WRITING LEVEL %d ██\n\n"+
				"Electric field E > Ec.\n"+
				"Domains are switching.\n"+
				"Polarization is changing.\n\n"+
				"This is how we STORE data.", wrdTarget)
		case 1: // HOLD
			phaseExplanation = fmt.Sprintf("░░ HOLDING LEVEL %d ░░\n\n"+
				"E-field is zero.\n"+
				"Polarization PERSISTS.\n"+
				"No power needed to retain.\n\n"+
				"This is NON-VOLATILE memory.", level)
		case 2: // READ
			phaseExplanation = fmt.Sprintf("▒▒ READING LEVEL %d ▒▒\n\n"+
				"Small E-field (E < Ec).\n"+
				"Polarization unchanged.\n"+
				"We sense, not alter.\n\n"+
				"Infinite read endurance.", level)
		case 3: // DISPLAY
			phaseExplanation = fmt.Sprintf("✓ READ COMPLETE: %d\n\n"+
				"Data retrieved successfully!\n"+
				"State unchanged after read.\n\n"+
				"Next: Writing new level...", level)
		}
		return phaseExplanation + "\n\n─────────────────\n" +
			"WHY THIS MATTERS\n\n" +
			"vs DRAM: 1000× less energy\n" +
			"vs Flash: 10⁷× faster\n" +
			"30 levels = 4.9 bits/cell\n" +
			"Non-volatile: no refresh"

	default:
		return "Select a waveform mode\nto see explanation."
	}
}

func (a *App) addLogEntry(entry string) {
	// Add timestamp prefix
	timestamp := fmt.Sprintf("t=%.1fs", a.simTime)
	fullEntry := fmt.Sprintf("%s %s", timestamp, entry)
	a.logEntries = append(a.logEntries, fullEntry)
	if len(a.logEntries) > a.maxLogLines {
		a.logEntries = a.logEntries[1:]
	}
}

func (a *App) getLogText() string {
	if len(a.logEntries) == 0 {
		return "Waiting for operations..."
	}
	result := ""
	for _, e := range a.logEntries {
		result += e + "\n"
	}
	return result
}

func (a *App) simulationLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	lastTime := time.Now()

	for a.running {
		<-ticker.C

		if a.paused {
			continue
		}

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		a.simTime += dt
		// Wrap simTime to prevent floating-point issues after long runs
		if a.simTime > 1000 {
			a.simTime = math.Mod(a.simTime, 1000)
		}

		a.mu.Lock()

		// Generate E-field based on waveform
		if a.autoMode && a.waveform != WaveformManual {
			Emax := a.material.Ec * 2
			// Wrap phase to prevent floating-point precision loss over long times
			phase := math.Mod(2*math.Pi*a.frequency*a.simTime, 2*math.Pi)

			switch a.waveform {
			case WaveformSine:
				a.electricField = Emax * math.Sin(phase)
			case WaveformTriangle:
				p := phase / (2 * math.Pi)
				if p < 0.25 {
					a.electricField = Emax * (4 * p)
				} else if p < 0.75 {
					a.electricField = Emax * (2 - 4*p)
				} else {
					a.electricField = Emax * (4*p - 4)
				}
			case WaveformSquare:
				if math.Sin(phase) >= 0 {
					a.electricField = Emax
				} else {
					a.electricField = -Emax
				}
			case WaveformRandomWalk:
				// Random Walk: pick random target levels and ramp to them
				// Frequency controls how often we pick new targets
				stepDelay := 1.5 / a.frequency // Higher freq = faster changes
				a.rwStepTimer += dt
				if a.rwStepTimer >= stepDelay {
					a.rwStepTimer = 0
					oldTarget := a.rwTargetLevel
					a.rwTargetLevel = rand.Intn(30) + 1 // Random level 1-30
					// Log target change
					a.addLogEntry(fmt.Sprintf("TARGET %d→%d", oldTarget, a.rwTargetLevel))
				}
				// Calculate E needed to reach that level
				// Level 1-30 maps to E-field: must exceed Ec to cause switching
				// Level 1 = -Emax (negative saturation)
				// Level 15-16 = near 0
				// Level 30 = +Emax (positive saturation)
				targetNormP := (float64(a.rwTargetLevel) - 15.5) / 14.5 // -1 to +1
				targetE := targetNormP * Emax
				// Ramp toward target - speed scales with frequency
				rampSpeed := 1.5 * Emax * a.frequency // Faster freq = faster ramp
				diff := targetE - a.rwCurrentE
				step := rampSpeed * dt
				if math.Abs(diff) < step {
					a.rwCurrentE = targetE
				} else if diff > 0 {
					a.rwCurrentE += step
				} else {
					a.rwCurrentE -= step
				}
				a.electricField = a.rwCurrentE
			case WaveformWriteReadDemo:
				// Write/Read Demo: demonstrates write then read cycle
				// Key insight: To write level N, we need E > Ec to cause switching
				// Level 1-15 = negative P (need negative E past -Ec)
				// Level 16-30 = positive P (need positive E past +Ec)
				a.wrdPhaseTimer += dt
				phaseDuration := 1.0 / a.frequency // Scale with frequency

				// Calculate write E-field: must exceed Ec, scaled by how far from center
				// Map level 1-30 to E-field: level 1 = -Emax, level 30 = +Emax
				targetNormP := (float64(a.wrdTargetLevel) - 15.5) / 14.5 // -1 to +1
				// E-field must exceed Ec to cause switching, so scale appropriately
				a.wrdWriteE = targetNormP * Emax

				// Ramp rate scales with frequency
				rampRate := 2.0 * Emax * a.frequency

				switch a.wrdPhase {
				case 0: // WRITE phase - ramp to write voltage (must exceed Ec)
					diff := a.wrdWriteE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = a.wrdWriteE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					if a.wrdPhaseTimer > phaseDuration {
						a.wrdPhase = 1
						a.wrdPhaseTimer = 0
					}
				case 1: // HOLD phase - return to zero (polarization persists!)
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					if a.wrdPhaseTimer > phaseDuration*0.7 && math.Abs(a.electricField) < 0.01*Emax {
						a.wrdPhase = 2
						a.wrdPhaseTimer = 0
					}
				case 2: // READ phase - small pulse below Ec (doesn't change state)
					readE := a.material.Ec * 0.4 // Below Ec - won't switch
					if a.wrdWriteE < 0 {
						readE = -readE
					}
					step := rampRate * 0.5 * dt
					diff := readE - a.electricField
					if math.Abs(diff) < step {
						a.electricField = readE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					if a.wrdPhaseTimer > phaseDuration*0.5 {
						a.wrdReadLevel = a.discreteLevel + 1
						a.wrdPhase = 3
						a.wrdPhaseTimer = 0
					}
				case 3: // DISPLAY phase - return to zero, show result
					step := rampRate * 0.5 * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					if a.wrdPhaseTimer > phaseDuration*1.2 {
						// Pick new target - alternate between high and low for drama
						if a.wrdTargetLevel > 15 {
							a.wrdTargetLevel = rand.Intn(10) + 1 // Low: 1-10
						} else {
							a.wrdTargetLevel = rand.Intn(10) + 21 // High: 21-30
						}
						a.wrdPhase = 0
						a.wrdPhaseTimer = 0
					}
				}
			}
		}

		// Update physics
		a.polarization = a.preisach.Update(a.electricField)
		a.normalizedP = a.preisach.NormalizedPolarization()
		a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * 29))
		if a.discreteLevel < 0 {
			a.discreteLevel = 0
		}
		if a.discreteLevel > 29 {
			a.discreteLevel = 29
		}

		// Record history
		a.eHistory = append(a.eHistory, a.electricField)
		a.pHistory = append(a.pHistory, a.polarization)
		if len(a.eHistory) > a.maxHistory {
			a.eHistory = a.eHistory[1:]
			a.pHistory = a.pHistory[1:]
		}

		// Copy data for UI update
		eField := a.electricField
		pol := a.polarization
		level := a.discreteLevel
		eHist := make([]float64, len(a.eHistory))
		pHist := make([]float64, len(a.pHistory))
		copy(eHist, a.eHistory)
		copy(pHist, a.pHistory)

		a.mu.Unlock()

		// Update UI (must be on main thread)
		a.updateUI(eField, pol, level, eHist, pHist)
	}
}

func (a *App) updateUI(eField, pol float64, level int, eHist, pHist []float64) {
	fyne.Do(func() {
		// Update labels
		a.eFieldLabel.SetText(fmt.Sprintf("E-field: %.3f MV/cm", eField/1e8))
		a.pLabel.SetText(fmt.Sprintf("P: %.2f µC/cm²", pol*100))
		a.levelLabel.SetText(fmt.Sprintf("Level: %d/30", level+1))

		// Update WRITE/READ mode indicator based on E vs Ec
		isWrite := math.Abs(eField) > a.material.Ec
		a.modeIndicator.SetWrite(isWrite)
		a.modeIndicator.Refresh()

		// Update slider position for auto modes
		if a.autoMode {
			a.eFieldSlider.SetValue(eField / a.material.Ec)
		}

		// Update status and logging
		if a.paused {
			a.statusLabel.SetText("⏸ Paused")
		} else {
			a.mu.RLock()
			waveform := a.waveform
			wrdPhase := a.wrdPhase
			wrdTarget := a.wrdTargetLevel
			wrdRead := a.wrdReadLevel
			rwTarget := a.rwTargetLevel
			lastPhase := a.lastLogPhase
			a.mu.RUnlock()

			switch waveform {
			case WaveformRandomWalk:
				a.statusLabel.SetText(fmt.Sprintf("● Random Walk | Target: %d | Current: %d", rwTarget, level+1))
			case WaveformWriteReadDemo:
				var phaseStr string
				// Log phase transitions
				if wrdPhase != lastPhase {
					a.mu.Lock()
					a.lastLogPhase = wrdPhase
					switch wrdPhase {
					case 0:
						a.addLogEntry(fmt.Sprintf("WRITE → %d", wrdTarget))
					case 1:
						a.addLogEntry(fmt.Sprintf("HOLD  → %d", level+1))
					case 2:
						a.addLogEntry("READ  ...")
					case 3:
						status := "✓"
						if wrdRead != wrdTarget {
							status = "~"
						}
						a.addLogEntry(fmt.Sprintf("READ  → %d %s", wrdRead, status))
					}
					a.mu.Unlock()
				}

				switch wrdPhase {
				case 0:
					phaseStr = fmt.Sprintf("WRITING %d...", wrdTarget)
				case 1:
					phaseStr = "HOLDING..."
				case 2:
					phaseStr = "READING..."
				case 3:
					match := ""
					if wrdRead == wrdTarget {
						match = " OK"
					}
					phaseStr = fmt.Sprintf("Wrote: %d, Read: %d%s", wrdTarget, wrdRead, match)
				}
				a.statusLabel.SetText(fmt.Sprintf("● Write/Read Demo | %s", phaseStr))
			default:
				frac := a.preisach.GetSwitchedFraction() * 100
				a.statusLabel.SetText(fmt.Sprintf("● Running | t=%.2fs | Switched: %.1f%%", a.simTime, frac))
			}
		}

		// Update slide text based on current waveform
		a.slideText.SetText(a.getSlideText())

		// Update log text
		a.mu.RLock()
		logText := a.getLogText()
		a.mu.RUnlock()
		a.logText.SetText(logText)

		// Update plot
		a.plot.SetData(eHist, pHist, eField, pol)
		a.plot.Refresh()

		// Update level indicator
		a.levelIndicator.SetLevel(level)
		a.levelIndicator.Refresh()

		// Update cell visualizer
		a.cellViz.SetLevel(level)
		a.cellViz.Refresh()
	})
}

// ============================================================
// Custom P-E Plot Widget
// ============================================================

// PEPlot is a custom widget for drawing P-E hysteresis curves
type PEPlot struct {
	widget.BaseWidget

	mu       sync.RWMutex
	eData    []float64
	pData    []float64
	currentE float64
	currentP float64
	eMax     float64
	pMax     float64
	minSize  fyne.Size
}

// NewPEPlot creates a new P-E plot widget
func NewPEPlot(eMax, pMax float64) *PEPlot {
	p := &PEPlot{
		eMax:    eMax,
		pMax:    pMax,
		minSize: fyne.NewSize(400, 300),
	}
	p.ExtendBaseWidget(p)
	return p
}

func (p *PEPlot) SetMinSize(size fyne.Size) {
	p.minSize = size
}

func (p *PEPlot) MinSize() fyne.Size {
	return p.minSize
}

func (p *PEPlot) SetBounds(eMax, pMax float64) {
	p.mu.Lock()
	p.eMax = eMax
	p.pMax = pMax
	p.mu.Unlock()
}

func (p *PEPlot) SetData(eData, pData []float64, currentE, currentP float64) {
	p.mu.Lock()
	p.eData = eData
	p.pData = pData
	p.currentE = currentE
	p.currentP = currentP
	p.mu.Unlock()
}

func (p *PEPlot) CreateRenderer() fyne.WidgetRenderer {
	return &peplotRenderer{plot: p}
}

type peplotRenderer struct {
	plot    *PEPlot
	objects []fyne.CanvasObject
}

func (r *peplotRenderer) MinSize() fyne.Size {
	return r.plot.minSize
}

func (r *peplotRenderer) Layout(size fyne.Size) {
	// Layout is handled in Refresh, so we must trigger it when resized
	r.Refresh()
}

func (r *peplotRenderer) Refresh() {
	r.plot.mu.RLock()
	defer r.plot.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.plot.Size()

	// Background
	bg := canvas.NewRectangle(colorBackground)
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Margins - larger to accommodate axis labels
	marginLeft := float32(50)
	marginRight := float32(20)
	marginTop := float32(25)
	marginBottom := float32(35)
	plotW := size.Width - marginLeft - marginRight
	plotH := size.Height - marginTop - marginBottom

	// Grid lines (5 divisions each direction from center)
	numDivisions := 4 // divisions on each side of center
	for i := -numDivisions; i <= numDivisions; i++ {
		t := float32(i) / float32(numDivisions)

		// Vertical grid line
		x := marginLeft + plotW/2 + t*plotW/2
		vLine := canvas.NewLine(colorGrid)
		vLine.Position1 = fyne.NewPos(x, marginTop)
		vLine.Position2 = fyne.NewPos(x, marginTop+plotH)
		if i == 0 {
			vLine.StrokeWidth = 2
			vLine.StrokeColor = colorAxis
		} else {
			vLine.StrokeWidth = 1
		}
		r.objects = append(r.objects, vLine)

		// Horizontal grid line
		y := marginTop + plotH/2 - t*plotH/2
		hLine := canvas.NewLine(colorGrid)
		hLine.Position1 = fyne.NewPos(marginLeft, y)
		hLine.Position2 = fyne.NewPos(marginLeft+plotW, y)
		if i == 0 {
			hLine.StrokeWidth = 2
			hLine.StrokeColor = colorAxis
		} else {
			hLine.StrokeWidth = 1
		}
		r.objects = append(r.objects, hLine)

		// X-axis tick labels (E-field in MV/cm)
		if i != 0 {
			eVal := float64(t) * r.plot.eMax / 1e8 // Convert to MV/cm
			eTickLabel := canvas.NewText(fmt.Sprintf("%.1f", eVal), color.RGBA{180, 180, 180, 255})
			eTickLabel.TextSize = 10
			eTickLabel.Move(fyne.NewPos(x-12, marginTop+plotH+3))
			r.objects = append(r.objects, eTickLabel)
		}

		// Y-axis tick labels (P in µC/cm²)
		if i != 0 {
			pVal := float64(t) * r.plot.pMax * 100 // Convert to µC/cm²
			pTickLabel := canvas.NewText(fmt.Sprintf("%.0f", pVal), color.RGBA{180, 180, 180, 255})
			pTickLabel.TextSize = 10
			pTickLabel.Move(fyne.NewPos(marginLeft-28, y-6))
			r.objects = append(r.objects, pTickLabel)
		}
	}

	// Center coordinates
	centerX := marginLeft + plotW/2
	centerY := marginTop + plotH/2

	// Zero labels
	zeroLabelX := canvas.NewText("0", color.RGBA{180, 180, 180, 255})
	zeroLabelX.TextSize = 10
	zeroLabelX.Move(fyne.NewPos(centerX-4, marginTop+plotH+3))
	r.objects = append(r.objects, zeroLabelX)

	zeroLabelY := canvas.NewText("0", color.RGBA{180, 180, 180, 255})
	zeroLabelY.TextSize = 10
	zeroLabelY.Move(fyne.NewPos(marginLeft-15, centerY-6))
	r.objects = append(r.objects, zeroLabelY)

	// Axis title labels
	eLabel := canvas.NewText("E (MV/cm)", color.RGBA{200, 200, 200, 255})
	eLabel.TextSize = 12
	eLabel.TextStyle = fyne.TextStyle{Bold: true}
	eLabel.Move(fyne.NewPos(marginLeft+plotW-65, centerY-18))
	r.objects = append(r.objects, eLabel)

	pLabelText := canvas.NewText("P (µC/cm²)", color.RGBA{200, 200, 200, 255})
	pLabelText.TextSize = 12
	pLabelText.TextStyle = fyne.TextStyle{Bold: true}
	pLabelText.Move(fyne.NewPos(centerX+8, marginTop-2))
	r.objects = append(r.objects, pLabelText)

	// Ec markers (vertical dashed lines at ±Ec)
	// Assuming Ec is roughly at 40% of eMax (since eMax = 2.5*Ec)
	ecRatio := float32(1.0 / 2.5) // Ec / eMax
	ecPosX := centerX + ecRatio*plotW/2
	ecNegX := centerX - ecRatio*plotW/2

	// +Ec marker
	ecPosLine := canvas.NewLine(color.RGBA{255, 150, 0, 180})
	ecPosLine.Position1 = fyne.NewPos(ecPosX, marginTop)
	ecPosLine.Position2 = fyne.NewPos(ecPosX, marginTop+plotH)
	ecPosLine.StrokeWidth = 2
	r.objects = append(r.objects, ecPosLine)
	ecPosLabel := canvas.NewText("+Ec", color.RGBA{255, 150, 0, 255})
	ecPosLabel.TextSize = 10
	ecPosLabel.Move(fyne.NewPos(ecPosX-10, marginTop+plotH+18))
	r.objects = append(r.objects, ecPosLabel)

	// -Ec marker
	ecNegLine := canvas.NewLine(color.RGBA{255, 150, 0, 180})
	ecNegLine.Position1 = fyne.NewPos(ecNegX, marginTop)
	ecNegLine.Position2 = fyne.NewPos(ecNegX, marginTop+plotH)
	ecNegLine.StrokeWidth = 2
	r.objects = append(r.objects, ecNegLine)
	ecNegLabel := canvas.NewText("-Ec", color.RGBA{255, 150, 0, 255})
	ecNegLabel.TextSize = 10
	ecNegLabel.Move(fyne.NewPos(ecNegX-10, marginTop+plotH+18))
	r.objects = append(r.objects, ecNegLabel)

	// Pr markers (horizontal dashed lines at ±Pr)
	// Pr is roughly 80% of Ps, and Ps is at pMax/1.2
	prRatio := float32(0.8 / 1.2) // Pr / pMax
	prPosY := centerY - prRatio*plotH/2
	prNegY := centerY + prRatio*plotH/2

	// +Pr marker
	prPosLine := canvas.NewLine(color.RGBA{0, 200, 150, 180})
	prPosLine.Position1 = fyne.NewPos(marginLeft, prPosY)
	prPosLine.Position2 = fyne.NewPos(marginLeft+plotW, prPosY)
	prPosLine.StrokeWidth = 2
	r.objects = append(r.objects, prPosLine)
	prPosLabel := canvas.NewText("+Pr", color.RGBA{0, 200, 150, 255})
	prPosLabel.TextSize = 10
	prPosLabel.Move(fyne.NewPos(marginLeft-35, prPosY-6))
	r.objects = append(r.objects, prPosLabel)

	// -Pr marker
	prNegLine := canvas.NewLine(color.RGBA{0, 200, 150, 180})
	prNegLine.Position1 = fyne.NewPos(marginLeft, prNegY)
	prNegLine.Position2 = fyne.NewPos(marginLeft+plotW, prNegY)
	prNegLine.StrokeWidth = 2
	r.objects = append(r.objects, prNegLine)
	prNegLabel := canvas.NewText("-Pr", color.RGBA{0, 200, 150, 255})
	prNegLabel.TextSize = 10
	prNegLabel.Move(fyne.NewPos(marginLeft-32, prNegY-6))
	r.objects = append(r.objects, prNegLabel)

	// Plot the hysteresis data
	if len(r.plot.eData) > 1 {
		for i := 1; i < len(r.plot.eData); i++ {
			// Map data to screen coordinates
			x1 := marginLeft + plotW/2 + float32(r.plot.eData[i-1]/r.plot.eMax)*plotW/2
			y1 := centerY - float32(r.plot.pData[i-1]/r.plot.pMax)*plotH/2
			x2 := marginLeft + plotW/2 + float32(r.plot.eData[i]/r.plot.eMax)*plotW/2
			y2 := centerY - float32(r.plot.pData[i]/r.plot.pMax)*plotH/2

			// Color based on age (fade effect)
			age := float64(len(r.plot.eData)-i) / float64(len(r.plot.eData))
			alpha := uint8(255 * (1 - age*0.7))

			var lineColor color.RGBA
			if r.plot.pData[i] >= 0 {
				lineColor = color.RGBA{colorPositive.R, colorPositive.G, colorPositive.B, alpha}
			} else {
				lineColor = color.RGBA{colorNegative.R, colorNegative.G, colorNegative.B, alpha}
			}

			line := canvas.NewLine(lineColor)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			line.StrokeWidth = 2
			r.objects = append(r.objects, line)
		}
	}

	// Current position marker - large with glow effect
	markerX := marginLeft + plotW/2 + float32(r.plot.currentE/r.plot.eMax)*plotW/2
	markerY := centerY - float32(r.plot.currentP/r.plot.pMax)*plotH/2

	// Outer glow (larger, semi-transparent)
	markerGlow := canvas.NewCircle(color.RGBA{colorWarning.R, colorWarning.G, colorWarning.B, 80})
	markerGlow.Resize(fyne.NewSize(28, 28))
	markerGlow.Move(fyne.NewPos(markerX-14, markerY-14))
	r.objects = append(r.objects, markerGlow)

	// Middle glow
	markerMidGlow := canvas.NewCircle(color.RGBA{colorWarning.R, colorWarning.G, colorWarning.B, 150})
	markerMidGlow.Resize(fyne.NewSize(20, 20))
	markerMidGlow.Move(fyne.NewPos(markerX-10, markerY-10))
	r.objects = append(r.objects, markerMidGlow)

	// Main marker (solid)
	marker := canvas.NewCircle(colorWarning)
	marker.Resize(fyne.NewSize(14, 14))
	marker.Move(fyne.NewPos(markerX-7, markerY-7))
	r.objects = append(r.objects, marker)

	// Inner highlight (white center)
	markerInner := canvas.NewCircle(color.RGBA{255, 255, 255, 200})
	markerInner.Resize(fyne.NewSize(6, 6))
	markerInner.Move(fyne.NewPos(markerX-3, markerY-3))
	r.objects = append(r.objects, markerInner)
}

func (r *peplotRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *peplotRenderer) Destroy() {}

// ============================================================
// Custom Level Indicator Widget
// ============================================================

// LevelIndicator shows the 30 discrete states
type LevelIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	level   int
	minSize fyne.Size
}

// NewLevelIndicator creates a new level indicator
func NewLevelIndicator() *LevelIndicator {
	l := &LevelIndicator{
		level:   15,
		minSize: fyne.NewSize(60, 400),
	}
	l.ExtendBaseWidget(l)
	return l
}

func (l *LevelIndicator) SetMinSize(size fyne.Size) {
	l.minSize = size
}

func (l *LevelIndicator) MinSize() fyne.Size {
	return l.minSize
}

func (l *LevelIndicator) SetLevel(level int) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

func (l *LevelIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &levelRenderer{indicator: l}
}

type levelRenderer struct {
	indicator *LevelIndicator
	objects   []fyne.CanvasObject
}

func (r *levelRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *levelRenderer) Layout(size fyne.Size) {}

func (r *levelRenderer) Refresh() {
	r.indicator.mu.RLock()
	level := r.indicator.level
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.indicator.Size()

	// Background
	bg := canvas.NewRectangle(color.RGBA{0, 40, 80, 255}) // Darker blue for level indicator
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Draw 30 level segments - use same margin as plot (40) to align vertically
	marginH := float32(40) // Match plot's vertical margin
	marginW := float32(5)
	barW := size.Width - 2*marginW
	totalH := size.Height - 2*marginH
	segH := totalH / 30
	gap := float32(2)

	for i := 0; i < 30; i++ {
		y := marginH + float32(29-i)*segH

		var segColor color.RGBA
		if i == level {
			// Current level - bright white/yellow
			segColor = colorWarning
		} else if i < level {
			// Below current - gradient blue to cyan
			t := float64(i) / 29.0
			segColor = color.RGBA{
				uint8(50 + t*150),
				uint8(50 + t*200),
				255,
				200,
			}
		} else {
			// Above current - gradient pink to red
			t := float64(i) / 29.0
			segColor = color.RGBA{
				255,
				uint8(200 - t*150),
				uint8(200 - t*150),
				200,
			}
		}

		seg := canvas.NewRectangle(segColor)
		seg.Resize(fyne.NewSize(barW, segH-gap))
		seg.Move(fyne.NewPos(marginW, y))
		r.objects = append(r.objects, seg)

		// Level number for every 5th level
		if i%5 == 0 || i == 29 {
			label := canvas.NewText(fmt.Sprintf("%d", i+1), colorAxis)
			label.TextSize = 10
			label.Move(fyne.NewPos(marginW+barW+2, y))
			r.objects = append(r.objects, label)
		}
	}
}

func (r *levelRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *levelRenderer) Destroy() {}

// ============================================================
// Cell Visualizer Widget - THE MEMORY CELL
// ============================================================

// CellVisualizer shows a colored square representing the memory cell state
type CellVisualizer struct {
	widget.BaseWidget

	mu      sync.RWMutex
	level   int
	minSize fyne.Size
}

// NewCellVisualizer creates a new cell visualizer
func NewCellVisualizer() *CellVisualizer {
	c := &CellVisualizer{
		level:   15,
		minSize: fyne.NewSize(120, 160),
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *CellVisualizer) SetMinSize(size fyne.Size) {
	c.minSize = size
}

func (c *CellVisualizer) MinSize() fyne.Size {
	return c.minSize
}

func (c *CellVisualizer) SetLevel(level int) {
	c.mu.Lock()
	c.level = level
	c.mu.Unlock()
}

func (c *CellVisualizer) CreateRenderer() fyne.WidgetRenderer {
	return &cellRenderer{cell: c}
}

type cellRenderer struct {
	cell    *CellVisualizer
	objects []fyne.CanvasObject
}

func (r *cellRenderer) MinSize() fyne.Size {
	return r.cell.minSize
}

func (r *cellRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *cellRenderer) Refresh() {
	r.cell.mu.RLock()
	level := r.cell.level
	r.cell.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.cell.Size()

	// Background
	bg := canvas.NewRectangle(color.RGBA{0, 40, 80, 255}) // Darker blue for cell viz
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Calculate cell size and position
	margin := float32(10)
	cellSize := size.Width - 2*margin
	if size.Height-60 < cellSize {
		cellSize = size.Height - 60
	}
	cellX := (size.Width - cellSize) / 2
	cellY := margin

	// Cell border (electrode representation)
	borderWidth := float32(4)
	border := canvas.NewRectangle(color.RGBA{100, 100, 120, 255})
	border.Resize(fyne.NewSize(cellSize+borderWidth*2, cellSize+borderWidth*2))
	border.Move(fyne.NewPos(cellX-borderWidth, cellY-borderWidth))
	r.objects = append(r.objects, border)

	// Cell color based on level - matches the level indicator gradient
	// Level 1 (i=0) = Yellow, Level 15 = Pink/salmon, Level 30 (i=29) = Red
	t := float64(level) / 29.0
	var cellColor color.RGBA
	if t < 0.5 {
		// Yellow to pink transition (levels 1-15)
		t2 := t * 2
		cellColor = color.RGBA{
			255,                // R stays high
			uint8(230 - t2*80), // G decreases: 230 -> 150
			uint8(100 + t2*80), // B increases: 100 -> 180
			255,
		}
	} else {
		// Pink to red transition (levels 16-30)
		t2 := (t - 0.5) * 2
		cellColor = color.RGBA{
			255,                 // R stays high
			uint8(150 - t2*100), // G decreases: 150 -> 50
			uint8(180 - t2*130), // B decreases: 180 -> 50
			255,
		}
	}

	// The memory cell square
	cell := canvas.NewRectangle(cellColor)
	cell.Resize(fyne.NewSize(cellSize, cellSize))
	cell.Move(fyne.NewPos(cellX, cellY))
	r.objects = append(r.objects, cell)

	// Inner glow effect
	glowSize := cellSize * 0.6
	glowX := cellX + (cellSize-glowSize)/2
	glowY := cellY + (cellSize-glowSize)/2
	glow := canvas.NewRectangle(color.RGBA{
		uint8(min(int(cellColor.R)+30, 255)),
		uint8(min(int(cellColor.G)+30, 255)),
		uint8(min(int(cellColor.B)+30, 255)),
		100,
	})
	glow.Resize(fyne.NewSize(glowSize, glowSize))
	glow.Move(fyne.NewPos(glowX, glowY))
	r.objects = append(r.objects, glow)

	// Level text inside cell
	levelText := canvas.NewText(fmt.Sprintf("%d", level+1), color.RGBA{0, 0, 0, 200})
	levelText.TextSize = 28
	levelText.TextStyle = fyne.TextStyle{Bold: true}
	textW := float32(20)
	if level+1 >= 10 {
		textW = 35
	}
	levelText.Move(fyne.NewPos(cellX+(cellSize-textW)/2, cellY+cellSize/2-16))
	r.objects = append(r.objects, levelText)

	// Label below cell
	labelY := cellY + cellSize + 8
	levelLabel := canvas.NewText(fmt.Sprintf("Level %d/30", level+1), color.RGBA{200, 200, 200, 255})
	levelLabel.TextSize = 14
	levelLabel.TextStyle = fyne.TextStyle{Bold: true}
	levelLabel.Move(fyne.NewPos(cellX+5, labelY))
	r.objects = append(r.objects, levelLabel)

	// State description
	var stateText string
	if level < 10 {
		stateText = "Negative P"
	} else if level > 19 {
		stateText = "Positive P"
	} else {
		stateText = "Intermediate"
	}
	stateLabel := canvas.NewText(stateText, color.RGBA{150, 150, 150, 255})
	stateLabel.TextSize = 11
	stateLabel.Move(fyne.NewPos(cellX+10, labelY+18))
	r.objects = append(r.objects, stateLabel)
}

func (r *cellRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *cellRenderer) Destroy() {}

// ============================================================
// Mode Indicator Widget - WRITE/READ with colored background
// ============================================================

// ModeIndicator shows the current mode (WRITE/READ) with colored background
type ModeIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	isWrite bool
	minSize fyne.Size
}

// NewModeIndicator creates a new mode indicator
func NewModeIndicator() *ModeIndicator {
	m := &ModeIndicator{
		isWrite: false,
		minSize: fyne.NewSize(180, 60),
	}
	m.ExtendBaseWidget(m)
	return m
}

func (m *ModeIndicator) SetMinSize(size fyne.Size) {
	m.minSize = size
}

func (m *ModeIndicator) MinSize() fyne.Size {
	return m.minSize
}

func (m *ModeIndicator) SetWrite(isWrite bool) {
	m.mu.Lock()
	m.isWrite = isWrite
	m.mu.Unlock()
}

func (m *ModeIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &modeRenderer{indicator: m}
}

type modeRenderer struct {
	indicator *ModeIndicator
	objects   []fyne.CanvasObject
}

func (r *modeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *modeRenderer) Layout(size fyne.Size) {
	r.Refresh()
}

func (r *modeRenderer) Refresh() {
	r.indicator.mu.RLock()
	isWrite := r.indicator.isWrite
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]
	size := r.indicator.Size()

	// Box colors
	var bgColor, borderColor color.RGBA
	var modeText, conditionText string

	if isWrite {
		bgColor = color.RGBA{180, 50, 50, 255} // Red background
		borderColor = color.RGBA{255, 100, 100, 255}
		modeText = "██ WRITE ██"
		conditionText = "|E| > Ec"
	} else {
		bgColor = color.RGBA{50, 150, 80, 255} // Green background
		borderColor = color.RGBA{100, 220, 130, 255}
		modeText = "░░ READ ░░"
		conditionText = "|E| < Ec"
	}

	// Border
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background with padding
	padding := float32(3)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Mode text (centered, top)
	modeLabel := canvas.NewText(modeText, color.White)
	modeLabel.TextSize = 16
	modeLabel.TextStyle = fyne.TextStyle{Bold: true}
	textW := float32(100)
	modeLabel.Move(fyne.NewPos((size.Width-textW)/2, 8))
	r.objects = append(r.objects, modeLabel)

	// Condition text (centered, bottom)
	condLabel := canvas.NewText(conditionText, color.RGBA{220, 220, 220, 255})
	condLabel.TextSize = 14
	condLabelW := float32(70)
	condLabel.Move(fyne.NewPos((size.Width-condLabelW)/2, 32))
	r.objects = append(r.objects, condLabel)
}

func (r *modeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *modeRenderer) Destroy() {}

// ============================================================
// Custom Theme
// ============================================================

type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground // FeCIM blue #003264
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255} // Slightly lighter blue
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255} // Darker blue for inputs
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255} // Separator lines
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// ============================================================
// Fixed Width Layout
// ============================================================

// fixedWidthLayout is a custom layout that enforces a fixed width
type fixedWidthLayout struct {
	width float32
}

func (l *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if o.Visible() {
			minH = fyne.Max(minH, o.MinSize().Height)
		}
	}
	return fyne.NewSize(l.width, minH)
}

func (l *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(l.width, size.Height))
		o.Move(fyne.NewPos(0, 0))
	}
}

// ============================================================
// Percentage-Based HBox Layout
// ============================================================

// percentHBoxLayout distributes width based on percentage weights
// Children are laid out left-to-right with widths proportional to their weights
type percentHBoxLayout struct {
	weights []float32 // percentage weights (0.0 to 1.0 each)
}

func (l *percentHBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if o.Visible() {
			minH = fyne.Max(minH, o.MinSize().Height)
		}
	}
	return fyne.NewSize(100, minH) // Minimal width, will expand
}

func (l *percentHBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for i, o := range objects {
		if !o.Visible() {
			continue
		}
		var w float32
		if i < len(l.weights) {
			w = size.Width * l.weights[i]
		}
		o.Resize(fyne.NewSize(w, size.Height))
		o.Move(fyne.NewPos(x, 0))
		x += w
	}
}
