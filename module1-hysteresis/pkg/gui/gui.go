// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// Uses fyne.io/fyne/v2 for cross-platform native GUI with proper graphics.
package gui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path/filepath"
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
	"multilayer-ferroelectric-cim-visualizer/shared/logging"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// Package-level logger for hysteresis GUI
var log *logging.Logger

func init() {
	log = logging.NewLogger("hysteresis")
}

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

// abs returns absolute value of int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

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

	// Write/Read Demo state (improved physics)
	wrdPhase       int     // 0=saturate, 1=settle, 2=hold, 3=read, 4=display, 5=retention, 6=verify
	wrdTargetLevel int     // Target level to write (1-30)
	wrdReadLevel   int     // Level read back
	wrdPhaseTimer  float64 // Time in current phase
	wrdWriteE      float64 // E-field during write
	wrdSaturateE   float64 // Saturation E-field (±Emax)
	wrdSettleE     float64 // Settle E-field (determines final level)
	wrdStartLevel  int     // Level at start of write cycle
	wrdDebugLog    *WriteReadDebugLog
	wrdPrevP       float64 // Previous polarization for energy calculation (E·dP integration)
	wrdCycleEnergy float64 // Energy for current write/read cycle

	// Dr. Tour Demo Metrics (impressive stats!)
	wrdTotalWrites   int     // Total write operations
	wrdSuccessWrites int     // Successful writes (read matches target)
	wrdTotalEnergyfJ float64 // Total energy consumed in femtojoules
	wrdDemoMode      int     // 0=multilevel, 1=retention, 2=endurance
	wrdRetentionTime float64 // Time at E=0 during retention demo
	wrdBitsStored    float64 // log2(30) = 4.91 bits per cell

	// Interactive Click Mode - user clicks level indicator to go to that level
	interactiveMode      bool    // True when in click-to-select mode
	interactiveIsWrite   bool    // True = Write operation, False = Read operation
	interactiveTarget    int     // Target level user clicked
	interactivePhase     int     // 0=idle, 1=saturate, 2=settle, 3=hold, 4=sense
	interactivePhaseTime float64 // Time in current phase

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

	// Wake-up/Fatigue display labels (Dr. Tour recommendation)
	cyclesLabel  *widget.Label
	wakeupLabel  *widget.Label
	fatigueLabel *widget.Label

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
	WaveformInteractive // Click level indicator to go there
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

// WriteReadDebugLog stores debug data for write/read operations
type WriteReadDebugLog struct {
	Timestamp   string             `json:"timestamp"`
	Material    string             `json:"material"`
	Ec          float64            `json:"ec_v_per_m"`
	EcMVcm      float64            `json:"ec_mv_per_cm"`
	Ps          float64            `json:"ps_c_per_m2"`
	Cycles      []WriteReadCycle   `json:"cycles"`
}

// WriteReadCycle stores one complete write/read cycle
type WriteReadCycle struct {
	CycleNum     int                `json:"cycle_num"`
	TargetLevel  int                `json:"target_level"`
	StartLevel   int                `json:"start_level"`
	ReadLevel    int                `json:"read_level"`
	Success      bool               `json:"success"`
	Phases       []WriteReadPhase   `json:"phases"`
}

// WriteReadPhase stores data for one phase of the write/read cycle
type WriteReadPhase struct {
	Phase       string    `json:"phase"`
	Duration    float64   `json:"duration_s"`
	EFieldStart float64   `json:"e_field_start_mv_cm"`
	EFieldEnd   float64   `json:"e_field_end_mv_cm"`
	EFieldPeak  float64   `json:"e_field_peak_mv_cm"`
	PStart      float64   `json:"p_start_uc_cm2"`
	PEnd        float64   `json:"p_end_uc_cm2"`
	LevelStart  int       `json:"level_start"`
	LevelEnd    int       `json:"level_end"`
	Samples     []PhaseDataPoint `json:"samples,omitempty"`
}

// PhaseDataPoint stores a single data point during a phase
type PhaseDataPoint struct {
	Time   float64 `json:"t"`
	E      float64 `json:"e"`
	P      float64 `json:"p"`
	Level  int     `json:"level"`
}

// saveDebugLog saves the debug log to a JSON file
func (a *App) saveDebugLog() {
	if a.wrdDebugLog == nil || len(a.wrdDebugLog.Cycles) == 0 {
		return
	}

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Error creating logs dir: %v", err)
		return
	}

	// Generate filename with timestamp
	filename := filepath.Join(logsDir, fmt.Sprintf("hysteresis-%s.json",
		time.Now().Format("2006-01-02T15-04-05")))

	// Marshal to JSON
	data, err := json.MarshalIndent(a.wrdDebugLog, "", "  ")
	if err != nil {
		log.Printf("Error marshaling debug log: %v", err)
		return
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("Error writing debug log: %v", err)
		return
	}

	log.Info("Debug log saved: %s", filename)
}

// initDebugLog initializes the debug log
func (a *App) initDebugLog() {
	// NOTE: Caller must hold a.mu.Lock() before calling this function
	// Defensive: ensure material exists
	materialName := "Unknown"
	materialEc := 0.0
	materialPs := 0.0
	if a.material != nil {
		materialName = a.material.Name
		materialEc = a.material.Ec
		materialPs = a.material.Ps
	}

	a.wrdDebugLog = &WriteReadDebugLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Material:  materialName,
		Ec:        materialEc,
		EcMVcm:    materialEc / 1e8,
		Ps:        materialPs,
		Cycles:    make([]WriteReadCycle, 0),
	}
	// Clear previous log entries and add impressive startup banner
	// Defensive: initialize logEntries if nil
	if a.logEntries == nil {
		a.logEntries = make([]string, 0, 12)
	} else {
		a.logEntries = a.logEntries[:0]
	}
	a.logEntries = append(a.logEntries, "══════════════════════")
	a.logEntries = append(a.logEntries, "  FeCIM WRITE/READ    ")
	a.logEntries = append(a.logEntries, "══════════════════════")
	a.logEntries = append(a.logEntries, fmt.Sprintf("Material: %s", materialName))
	a.logEntries = append(a.logEntries, fmt.Sprintf("Ec: %.1f MV/cm", materialEc/1e8))
	a.logEntries = append(a.logEntries, "30 LEVELS = 4.91 bits/cell")
	a.logEntries = append(a.logEntries, "──────────────────────")
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
	// Create cell visualizer (THE memory cell!) - prominent size
	a.cellViz = NewCellVisualizer()
	a.cellViz.SetMinSize(fyne.NewSize(200, 240))

	// Create P-E plot - will expand to fill space
	a.plot = NewPEPlot(a.material.Ec*2.5, a.material.Ps*1.2)
	a.plot.SetMinSize(fyne.NewSize(400, 350))

	// Create level indicator (wider for better labels, clickable for interactive mode)
	a.levelIndicator = NewLevelIndicator()
	a.levelIndicator.SetMinSize(fyne.NewSize(80, 350))
	// Wire up click callback for interactive mode
	a.levelIndicator.OnLevelClicked = func(targetLevel int) {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.waveform == WaveformInteractive && a.interactivePhase == 0 {
			// Start animation to target level
			a.interactiveTarget = targetLevel
			a.interactivePhase = 1 // Start saturate phase
			a.interactivePhaseTime = 0
			// Log the action
			action := "WRITE"
			if !a.interactiveIsWrite {
				action = "READ"
				a.interactivePhase = 4 // Skip to sense phase for read
			}
			a.addLogEntry(fmt.Sprintf("%s → Level %d", action, targetLevel))
		}
	}

	// Create controls panel
	controls := a.createControlsPanel()

	// Create info panel
	info := a.createInfoPanel()

	// Create educational slide panel
	slidePanel := a.createSlidePanel()

	// Create log panel
	logPanel := a.createLogPanel()

	// Cell title with underline effect
	cellTitle := canvas.NewText("MEMORY CELL", color.RGBA{0, 212, 255, 255})
	cellTitle.TextSize = 18
	cellTitle.TextStyle = fyne.TextStyle{Bold: true}
	cellTitle.Alignment = fyne.TextAlignCenter

	cellUnderline := canvas.NewRectangle(color.RGBA{0, 212, 255, 150})
	cellUnderline.SetMinSize(fyne.NewSize(120, 2))

	cellHeader := container.NewVBox(
		container.NewCenter(cellTitle),
		container.NewCenter(cellUnderline),
	)

	// Cell container - vertically centered with padding
	cellContainer := container.NewBorder(
		cellHeader,
		nil, nil, nil,
		container.NewPadded(container.NewCenter(a.cellViz)),
	)

	// Right panel - controls stay outside scroll to avoid Select popup issues
	scrollableContent := container.NewVBox(
		container.NewPadded(info),
		a.createSectionDivider(),
		container.NewPadded(slidePanel),
		a.createSectionDivider(),
		container.NewPadded(logPanel),
	)
	rightScroll := container.NewScroll(scrollableContent)

	// Controls at top (with Select widgets), scroll below
	// Use a min size container to prevent resize when dropdown opens
	controlsWithMinWidth := container.New(&fixedMinWidthLayout{minWidth: 280},
		container.NewVBox(
			container.NewPadded(controls),
			a.createSectionDivider(),
		),
	)
	rightArea := container.NewBorder(
		controlsWithMinWidth,
		nil, nil, nil,
		rightScroll,
	)

	// Plot + level in same row using Border layout
	// Border allows level indicator to be tappable (HSplit may intercept events)
	// Level indicator gets fixed width on right, plot fills the center
	plotAndLevel := container.NewBorder(
		nil, nil, nil, // top, bottom, left
		a.levelIndicator, // right side - the tappable level indicator
		a.plot,           // center - the plot takes remaining space
	)

	// Center area (plot takes full space, no separate title - it's inside plot now)
	centerArea := plotAndLevel

	// Status bar at bottom - more informative
	a.statusLabel = widget.NewLabel("Running...")
	statusBar := container.NewHBox(
		layout.NewSpacer(),
		a.statusLabel,
		layout.NewSpacer(),
	)

	// Adaptive layout: supports mobile (tabs) and desktop (splits)
	// Zones: [0]=Cell, [1]=Plot+Level, [2]=Controls
	zones := []fyne.CanvasObject{
		cellContainer,
		centerArea,
		rightArea,
	}
	tabLabels := []string{"Memory Cell", "P-E Plot", "Controls"}

	adaptive := sharedwidgets.NewAdaptiveLayout(zones, tabLabels)
	adaptive.SetDesktopLayout(func(zones []fyne.CanvasObject) fyne.CanvasObject {
		// Desktop: Cell | Plot | Controls using HSplit
		innerSplit := container.NewHSplit(zones[1], zones[2])
		innerSplit.SetOffset(0.63) // Plot gets 63% of remaining space

		outerSplit := container.NewHSplit(zones[0], innerSplit)
		outerSplit.SetOffset(0.14) // Cell gets 14% of total width

		return outerSplit
	})

	return container.NewBorder(
		nil, // No header - cleaner look
		statusBar,
		nil, nil,
		adaptive.Content(),
	)
}

// createSectionDivider creates a subtle divider between sections
func (a *App) createSectionDivider() fyne.CanvasObject {
	line := canvas.NewRectangle(color.RGBA{0, 100, 180, 100})
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.NewPadded(line)
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
	// E-field slider (no logging - too frequent)
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

	// Waveform selector (added Interactive mode for click-to-select)
	waveforms := []string{"Manual", "Sine Wave", "Triangle Wave", "Square Wave", "Random Walk", "Write/Read Demo", "Interactive"}
	a.waveformSelect = widget.NewSelect(waveforms, func(s string) {
		log.Selection("Waveform", s)
		a.mu.Lock()
		defer a.mu.Unlock()
		switch s {
		case "Manual":
			a.waveform = WaveformManual
			a.autoMode = false
			a.interactiveMode = false
			if a.eFieldSlider != nil {
				a.eFieldSlider.Enable()
			}
		case "Sine Wave":
			a.waveform = WaveformSine
			a.autoMode = true
			a.interactiveMode = false
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
		case "Triangle Wave":
			a.waveform = WaveformTriangle
			a.autoMode = true
			a.interactiveMode = false
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
		case "Square Wave":
			a.waveform = WaveformSquare
			a.autoMode = true
			a.interactiveMode = false
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
		case "Random Walk":
			a.waveform = WaveformRandomWalk
			a.autoMode = true
			a.interactiveMode = false
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			// Reset random walk state
			a.rwStepTimer = 0
			a.rwCurrentE = a.electricField
			a.rwTargetLevel = rand.Intn(30) + 1
		case "Write/Read Demo":
			a.waveform = WaveformWriteReadDemo
			a.autoMode = true
			a.interactiveMode = false
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			// Reset write/read demo state with improved physics
			a.wrdPhase = 0
			a.wrdPhaseTimer = 0
			a.wrdTargetLevel = rand.Intn(30) + 1
			a.wrdStartLevel = a.discreteLevel + 1
			// Reset Dr. Tour demo metrics
			a.wrdTotalWrites = 0
			a.wrdSuccessWrites = 0
			a.wrdTotalEnergyfJ = 0
			a.wrdCycleEnergy = 0
			a.wrdBitsStored = 4.91 // log2(30)
			// Initialize debug log (requires material to be set)
			if a.material != nil {
				a.initDebugLog()
			}
		case "Interactive":
			a.waveform = WaveformInteractive
			a.autoMode = true
			a.interactiveMode = true
			a.interactivePhase = 0 // Idle - waiting for click
			a.interactiveIsWrite = true // Default to Write mode
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			// Clear log and show instructions
			if a.logEntries == nil {
				a.logEntries = make([]string, 0, 12)
			} else {
				a.logEntries = a.logEntries[:0]
			}
			a.logEntries = append(a.logEntries, "══════════════════════")
			a.logEntries = append(a.logEntries, "  INTERACTIVE MODE    ")
			a.logEntries = append(a.logEntries, "══════════════════════")
			a.logEntries = append(a.logEntries, "Click level bar to")
			a.logEntries = append(a.logEntries, "go to that level!")
			a.logEntries = append(a.logEntries, "")
			a.logEntries = append(a.logEntries, "Toggle Read/Write below")
		}
	})
	a.waveformSelect.SetSelected("Sine Wave")

	// Material selector
	matNames := []string{"Default HZO", "Optimized Superlattice", "FeCIM HZO"}
	a.materialSelect = widget.NewSelect(matNames, func(s string) {
		log.Selection("Material", s)
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
			log.Button("Pause (now paused)")
			a.pauseBtn.SetText("Resume")
		} else {
			log.Button("Resume (now running)")
			a.pauseBtn.SetText("Pause")
		}
	})

	// Reset button
	resetBtn := widget.NewButton("Reset", func() {
		log.Button("Reset")
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

	// ELI5 (Explain Like I'm 5) button
	eli5Btn := widget.NewButton("ELI5", func() {
		log.Button("ELI5")
		a.showELI5Dialog()
	})
	eli5Btn.Importance = widget.LowImportance

	// Frequency slider
	freqSlider := widget.NewSlider(0.01, 1.0)
	freqSlider.Step = 0.01
	freqSlider.Value = 0.5
	freqLabel := widget.NewLabel("Freq: 0.50 Hz")
	freqSlider.OnChanged = func(v float64) {
		log.SliderChange("Frequency", v)
		a.mu.Lock()
		a.frequency = v
		// Reset trail when frequency changes
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.mu.Unlock()
		freqLabel.SetText(fmt.Sprintf("Freq: %.2f Hz", v))
	}

	// Temperature slider (Dr. Tour recommendation: show HZO's high Tc advantage)
	tempSlider := widget.NewSlider(200, 700)
	tempSlider.Step = 25
	tempSlider.Value = 300 // Room temperature (300K = 27°C)
	tempLabel := widget.NewLabel("T: 300 K (27°C)")
	tempSlider.OnChanged = func(v float64) {
		log.SliderChange("Temperature", v)
		a.mu.Lock()
		// Update Preisach model temperature
		a.preisach.SetTemperature(v)
		a.mu.Unlock()
		celsius := v - 273
		tcRatio := v / a.material.CurieTemp * 100
		tempLabel.SetText(fmt.Sprintf("T: %.0f K (%.0f°C) [%.0f%% Tc]", v, celsius, tcRatio))
	}

	// Trail length slider
	trailSlider := widget.NewSlider(50, 2000)
	trailSlider.Step = 50
	trailSlider.Value = float64(a.maxHistory)
	trailLabel := widget.NewLabel(fmt.Sprintf("Trail: %d", a.maxHistory))
	trailSlider.OnChanged = func(v float64) {
		log.SliderChange("TrailLength", v)
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

	// Grouped sections with styled headers
	modeGroup := container.NewVBox(
		a.createSectionTitle("MODE"),
		a.materialSelect,
		a.waveformSelect,
	)

	eFieldGroup := container.NewVBox(
		a.createSectionTitle("E-FIELD (×Ec)"),
		a.eFieldSlider,
		a.eFieldLabel,
	)

	timingGroup := container.NewVBox(
		a.createSectionTitle("TIMING"),
		container.NewGridWithColumns(2, freqLabel, trailLabel),
		freqSlider,
		trailSlider,
	)

	// Temperature group (Dr. Tour: HZO remains ferroelectric to ~450°C)
	tempGroup := container.NewVBox(
		a.createSectionTitle("TEMPERATURE"),
		tempLabel,
		tempSlider,
	)

	// Action buttons - styled
	actionGroup := container.NewHBox(
		layout.NewSpacer(),
		a.pauseBtn,
		resetBtn,
		eli5Btn,
		layout.NewSpacer(),
	)

	// Interactive mode controls - Read/Write toggle
	// Create a radio group for Write vs Read selection
	rwToggle := widget.NewRadioGroup([]string{"WRITE", "READ"}, func(selected string) {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.interactiveIsWrite = (selected == "WRITE")
		// Add log entry showing mode change
		if a.waveform == WaveformInteractive {
			if a.interactiveIsWrite {
				a.addLogEntry("Mode: WRITE (full cycle)")
			} else {
				a.addLogEntry("Mode: READ (sense only)")
			}
		}
	})
	rwToggle.Horizontal = true
	rwToggle.SetSelected("WRITE")

	interactiveHint := widget.NewLabel("Click level bar →")
	interactiveHint.TextStyle = fyne.TextStyle{Italic: true}

	interactiveGroup := container.NewVBox(
		a.createSectionTitle("INTERACTIVE"),
		rwToggle,
		container.NewCenter(interactiveHint),
	)

	// Add spacer at end to absorb any extra vertical space
	return container.NewVBox(
		modeGroup,
		eFieldGroup,
		timingGroup,
		tempGroup,
		actionGroup,
		interactiveGroup,
		layout.NewSpacer(),
	)
}

// createSectionTitle creates a styled section title with underline
func (a *App) createSectionTitle(text string) fyne.CanvasObject {
	title := canvas.NewText(text, color.RGBA{0, 212, 255, 255})
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	underline := canvas.NewRectangle(color.RGBA{0, 212, 255, 80})
	underline.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(
		container.NewCenter(title),
		underline,
	)
}

func (a *App) createInfoPanel() fyne.CanvasObject {
	a.pLabel = widget.NewLabel("P: 0.00 µC/cm²")
	a.levelLabel = widget.NewLabel("Level: 15/30")
	a.modeIndicator = NewModeIndicator()
	a.modeIndicator.SetMinSize(fyne.NewSize(220, 80))

	// State display - 2-column grid with larger labels
	stateGrid := container.NewGridWithColumns(2,
		widget.NewLabel("P:"), a.pLabel,
		widget.NewLabel("Level:"), a.levelLabel,
	)

	// Material params - compact
	matParams := widget.NewLabel(fmt.Sprintf(
		"Pr=%.0f Ps=%.0f µC/cm²  Ec=%.2f MV/cm\nEndurance: %.0e cycles",
		a.material.Pr*100, a.material.Ps*100,
		a.material.Ec/1e8,
		a.material.EnduranceCycles,
	))
	matParams.Wrapping = fyne.TextWrapOff

	// Wake-up/Fatigue display (Dr. Tour recommendation)
	a.cyclesLabel = widget.NewLabel("Cycles: 0")
	a.wakeupLabel = widget.NewLabel("Wake-up: 80%")
	a.fatigueLabel = widget.NewLabel("Fatigue: 0.0%")

	fatigueGrid := container.NewGridWithColumns(2,
		widget.NewLabel("Cycles:"), a.cyclesLabel,
		widget.NewLabel("Wake-up:"), a.wakeupLabel,
		widget.NewLabel("Fatigue:"), a.fatigueLabel,
	)

	return container.NewVBox(
		a.createSectionTitle("STATE"),
		stateGrid,
		container.NewCenter(a.modeIndicator),
		a.createSectionTitle("MATERIAL"),
		matParams,
		a.createSectionTitle("CYCLING (Dr. Tour)"),
		fatigueGrid,
	)
}

func (a *App) createSlidePanel() fyne.CanvasObject {
	a.slideTitle = widget.NewLabel("") // Keep for compatibility
	a.slideText = widget.NewLabel(a.getSlideText())
	a.slideText.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		a.createSectionTitle("WHAT YOU'RE SEEING"),
		a.slideText,
	)
}

func (a *App) createLogPanel() fyne.CanvasObject {
	a.logText = widget.NewLabel("Memory operations will appear here...")
	a.logText.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		a.createSectionTitle("MEMORY LOG"),
		a.logText,
	)
}

func (a *App) getSlideText() string {
	a.mu.RLock()
	level := a.discreteLevel + 1 // 1-indexed for display
	wrdPhase := a.wrdPhase
	wrdTarget := a.wrdTargetLevel
	isWrite := math.Abs(a.electricField) > a.material.Ec
	waveform := a.waveform
	wrdTotalWrites := a.wrdTotalWrites
	wrdSuccessWrites := a.wrdSuccessWrites
	wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
	a.mu.RUnlock()

	switch waveform {
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
		// Calculate stats (using local copies from RLock above)
		successRate := 0.0
		if wrdTotalWrites > 0 {
			successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
		}
		energyPerOp := 10.0 // ~10 fJ per operation (FeFET switching energy)

		var phaseExplanation string
		switch wrdPhase {
		case 0: // SATURATE
			phaseExplanation = fmt.Sprintf("▓▓ SATURATE → %d ▓▓\n\n"+
				"|E| = 2×Ec (maximum field)\n"+
				"ALL ferroelectric domains\n"+
				"switching simultaneously.\n\n"+
				"Energy: ~%.0f fJ\n"+
				"(10M× less than NAND!)\n\n"+
				"\"The same device does memory\n"+
				"AND computation.\" - Dr. Tour", wrdTarget, energyPerOp)
		case 1: // SETTLE
			phaseExplanation = fmt.Sprintf("██ SETTLE → %d ██\n\n"+
				"MINOR LOOP formation:\n"+
				"Partial domain reversal\n"+
				"sets the analog level.\n\n"+
				"This is how we store\n"+
				"4.91 BITS in ONE cell!\n"+
				"(Binary = only 1 bit)\n\n"+
				"30 levels = 5× density", wrdTarget)
		case 2: // HOLD
			phaseExplanation = fmt.Sprintf("░░ HOLD LEVEL %d ░░\n\n"+
				"E = 0, P persists!\n\n"+
				"ZERO POWER NEEDED.\n"+
				"Data retention: 10+ years\n"+
				"(demonstrated: 10⁷ sec)\n\n"+
				"This is TRUE non-volatile:\n"+
				"No refresh like DRAM.\n"+
				"No charge leakage.\n\n"+
				"Just ferroelectric\n"+
				"polarization.", level)
		case 3: // READ
			phaseExplanation = fmt.Sprintf("▒▒ READING LEVEL %d ▒▒\n\n"+
				"Sense pulse: |E| < Ec\n"+
				"State UNCHANGED!\n\n"+
				"Non-destructive read:\n"+
				"Unlike NAND, data stays.\n"+
				"No rewrite needed.\n\n"+
				"Read energy: ~%.0f fJ\n"+
				"(1000× less than DRAM)", level, energyPerOp*0.1)
		case 4: // DISPLAY
			status := "✓ SUCCESS"
			accuracy := ""
			if level != wrdTarget {
				status = fmt.Sprintf("△ Level %d (target %d)", level, wrdTarget)
				accuracy = "\n(Within ±1 is normal)"
			} else {
				accuracy = "\nPerfect analog storage!"
			}
			phaseExplanation = fmt.Sprintf("%s%s\n\n"+
				"─── SESSION STATS ───\n"+
				"Writes: %d\n"+
				"Success: %.0f%%\n"+
				"Energy: %.1f pJ total\n\n"+
				"Each write: ~10 fJ\n"+
				"Each read: ~1 fJ\n"+
				"─────────────────\n"+
				"Next level coming...", status, accuracy,
				wrdTotalWrites, successRate, wrdTotalEnergyfJ/1000)
		}
		// Add Dr. Tour footer
		return phaseExplanation + "\n\n═══════════════════\n" +
			"FeCIM ADVANTAGE:\n" +
			"• 30 levels = 4.9 bits/cell\n" +
			"• 10M× better than NAND\n" +
			"• 1000× better than DRAM\n" +
			"• CMOS compatible"

	case WaveformInteractive:
		// Interactive mode - explain what user can do
		a.mu.RLock()
		interPhase := a.interactivePhase
		interTarget := a.interactiveTarget
		interIsWrite := a.interactiveIsWrite
		a.mu.RUnlock()

		modeStr := "WRITE"
		modeExplain := "Full write cycle:\nSATURATE → SETTLE → HOLD"
		if !interIsWrite {
			modeStr = "READ"
			modeExplain = "Non-destructive sense:\nSmall pulse |E| < Ec"
		}

		if interPhase == 0 {
			return fmt.Sprintf("🎯 INTERACTIVE MODE\n\n"+
				"Current Level: %d/30\n\n"+
				"Mode: %s\n%s\n\n"+
				"─────────────────\n"+
				"HOW TO USE:\n"+
				"1. Select WRITE or READ\n"+
				"2. Click the level bar →\n"+
				"3. Watch the physics!\n\n"+
				"WRITE: Changes the level\n"+
				"READ: Just senses (no change)", level, modeStr, modeExplain)
		}

		// Show animation progress
		phaseNames := []string{"IDLE", "SATURATE", "SETTLE", "HOLD", "SENSE"}
		phaseName := "UNKNOWN"
		if interPhase >= 0 && interPhase < len(phaseNames) {
			phaseName = phaseNames[interPhase]
		}
		return fmt.Sprintf("🎯 %s → Level %d\n\n"+
			"Phase: %s\n"+
			"Current: Level %d\n\n"+
			"Watch the P-E plot\n"+
			"trace the path!\n\n"+
			"─────────────────\n"+
			"WRITE physics:\n"+
			"• SATURATE: |E|=2×Ec\n"+
			"• SETTLE: Minor loop\n"+
			"• HOLD: E=0, P persists", modeStr, interTarget, phaseName, level)

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
				// IMPROVED Write/Read Demo with correct ferroelectric physics
				//
				// Key insight: To write ANY level, you must first SATURATE then SETTLE
				// 1. SATURATE: Apply large |E| >> Ec to fully switch all hysterons
				// 2. SETTLE: Apply opposite field to create minor loop and set level
				// 3. HOLD: E = 0, polarization persists (non-volatile!)
				// 4. READ: Small sense pulse |E| < Ec, doesn't disturb state
				//
				// Phase mapping:
				// 0 = SATURATE (ramp to ±Emax)
				// 1 = SETTLE (ramp to settling field to set target level)
				// 2 = HOLD (return to zero)
				// 3 = READ (small sense pulse)
				// 4 = DISPLAY (show result, pick next target)

				a.wrdPhaseTimer += dt
				phaseDuration := 1.0 / a.frequency
				rampRate := 2.5 * Emax * a.frequency
				Ec := a.material.Ec

				// Calculate target normalized polarization (-1 to +1)
				// Level 1 = -1 (negative saturation)
				// Level 15-16 = ~0 (intermediate)
				// Level 30 = +1 (positive saturation)
				targetNormP := (float64(a.wrdTargetLevel) - 15.5) / 14.5

				// Determine saturation direction and settle field
				// For positive targets: saturate positive, settle negative
				// For negative targets: saturate negative, settle positive
				if targetNormP >= 0 {
					a.wrdSaturateE = Emax // Positive saturation
					// Settle field: partial negative field to create minor loop
					// More negative field → lower final level
					// The relationship is approximately: final P ≈ Ps * tanh((E_settle + Ec) / delta)
					// For level 30: settle at ~0 (stay at positive saturation)
					// For level 16: settle at ~ -Ec (intermediate)
					settleRatio := 1.0 - (float64(a.wrdTargetLevel-16) / 14.0) // 0 to 1 for levels 16-30
					a.wrdSettleE = -Ec * (0.5 + settleRatio*1.0) // Range: -0.5*Ec to -1.5*Ec
				} else {
					a.wrdSaturateE = -Emax // Negative saturation
					// For negative targets, settle with positive field
					// For level 1: settle at ~0 (stay at negative saturation)
					// For level 15: settle at ~ +Ec (intermediate)
					settleRatio := float64(a.wrdTargetLevel-1) / 14.0 // 0 to 1 for levels 1-15
					a.wrdSettleE = Ec * (0.5 + settleRatio*1.0) // Range: 0.5*Ec to 1.5*Ec
				}

				switch a.wrdPhase {
				case 0: // SATURATE phase - ramp to saturation voltage (±Emax)
					diff := a.wrdSaturateE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = a.wrdSaturateE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Transition when we've reached saturation and held briefly
					if a.wrdPhaseTimer > phaseDuration*0.6 && math.Abs(a.electricField-a.wrdSaturateE) < 0.01*Emax {
						a.wrdPhase = 1
						a.wrdPhaseTimer = 0
					}

				case 1: // SETTLE phase - ramp to settling field (creates minor loop)
					diff := a.wrdSettleE - a.electricField
					step := rampRate * 0.7 * dt // Slower for precision
					if math.Abs(diff) < step {
						a.electricField = a.wrdSettleE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Transition when we've reached settle field
					if a.wrdPhaseTimer > phaseDuration*0.5 && math.Abs(a.electricField-a.wrdSettleE) < 0.01*Emax {
						a.wrdPhase = 2
						a.wrdPhaseTimer = 0
					}

				case 2: // HOLD phase - return to zero (polarization persists!)
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					// Transition when at zero
					if a.wrdPhaseTimer > phaseDuration*0.5 && math.Abs(a.electricField) < 0.01*Emax {
						a.wrdPhase = 3
						a.wrdPhaseTimer = 0
					}

				case 3: // READ phase - small sense pulse below Ec
					readE := Ec * 0.3 // Well below Ec - won't switch
					if a.wrdSaturateE < 0 {
						readE = -readE // Match polarity of written state
					}
					step := rampRate * 0.4 * dt
					diff := readE - a.electricField
					if math.Abs(diff) < step {
						a.electricField = readE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Capture read level and transition
					if a.wrdPhaseTimer > phaseDuration*0.4 {
						a.wrdReadLevel = a.discreteLevel + 1
						a.wrdPhase = 4
						a.wrdPhaseTimer = 0

						// Track Dr. Tour demo metrics
						a.wrdTotalWrites++
						// Success if within ±1 level (analog tolerance)
						if abs(a.wrdReadLevel-a.wrdTargetLevel) <= 1 {
							a.wrdSuccessWrites++
						}
						// Add accumulated energy for this cycle (calculated from E·dP integration)
						a.wrdTotalEnergyfJ += a.wrdCycleEnergy
						a.wrdCycleEnergy = 0 // Reset for next cycle

						// Log this cycle for debugging
						if a.wrdDebugLog != nil {
							cycle := WriteReadCycle{
								CycleNum:    len(a.wrdDebugLog.Cycles) + 1,
								TargetLevel: a.wrdTargetLevel,
								StartLevel:  a.wrdStartLevel,
								ReadLevel:   a.wrdReadLevel,
								Success:     abs(a.wrdReadLevel-a.wrdTargetLevel) <= 1,
								Phases: []WriteReadPhase{
									{Phase: "SATURATE", EFieldPeak: a.wrdSaturateE / 1e8},
									{Phase: "SETTLE", EFieldEnd: a.wrdSettleE / 1e8},
									{Phase: "READ", EFieldPeak: readE / 1e8, LevelEnd: a.wrdReadLevel},
								},
							}
							a.wrdDebugLog.Cycles = append(a.wrdDebugLog.Cycles, cycle)

							// Save after every 5 cycles
							if len(a.wrdDebugLog.Cycles)%5 == 0 {
								go a.saveDebugLog()
							}
						}
					}

				case 4: // DISPLAY phase - return to zero, show result
					step := rampRate * 0.4 * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					// Transition to next cycle
					if a.wrdPhaseTimer > phaseDuration*0.8 {
						// Record start level for next cycle
						a.wrdStartLevel = a.discreteLevel + 1

						// Add comparison callout every 5 cycles
						if a.wrdTotalWrites > 0 && a.wrdTotalWrites%5 == 0 {
							fecimEnergy := a.wrdTotalEnergyfJ / 1000 // pJ
							nandEquiv := fecimEnergy * 10000000     // 10M× worse
							dramEquiv := fecimEnergy * 1000         // 1000× worse
							bitsStored := float64(a.wrdTotalWrites) * 4.91
							a.addLogEntry("━━ ENERGY COMPARISON ━━")
							a.addLogEntry(fmt.Sprintf("FeCIM: %.0f pJ total", fecimEnergy))
							a.addLogEntry(fmt.Sprintf("NAND:  %.0f pJ (10M×!)", nandEquiv))
							a.addLogEntry(fmt.Sprintf("DRAM:  %.0f pJ (1000×)", dramEquiv))
							a.addLogEntry(fmt.Sprintf("Bits stored: %.0f (%.1f×binary)", bitsStored, 4.91))
							a.addLogEntry("━━━━━━━━━━━━━━━━━━━━━━")
						}

						// Milestone celebrations
						switch a.wrdTotalWrites {
						case 10:
							a.addLogEntry("★★ 10 ops! ~49 bits stored ★★")
						case 25:
							a.addLogEntry("★★★ 25 ops! ~123 bits stored ★★★")
						case 50:
							a.addLogEntry("★★★★ 50 ops! ~245 bits stored ★★★★")
							a.addLogEntry("Binary would need 245 cells!")
							a.addLogEntry("FeCIM: only 50 cells! (5× denser)")
						case 100:
							a.addLogEntry("★★★★★ 100 OPERATIONS! ★★★★★")
							a.addLogEntry("~491 bits in 100 FeCIM cells")
							a.addLogEntry("Binary: 491 cells needed!")
							successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100
							a.addLogEntry(fmt.Sprintf("Accuracy: %.0f%%", successRate))
						}

						// Pick new target - alternate between high and low
						if a.wrdTargetLevel > 15 {
							a.wrdTargetLevel = rand.Intn(12) + 2 // Low: 2-13 (avoid extremes)
						} else {
							a.wrdTargetLevel = rand.Intn(12) + 18 // High: 18-29 (avoid extremes)
						}
						a.wrdPhase = 0
						a.wrdPhaseTimer = 0
						a.wrdCycleEnergy = 0 // Reset energy accumulator for next cycle
					}
				}

			case WaveformInteractive:
				// Interactive mode: user clicks level indicator to animate to that level
				// Phase: 0=idle, 1=saturate, 2=settle, 3=hold, 4=sense (read-only)
				Ec := a.material.Ec
				phaseDuration := 0.8 / a.frequency // Faster for interactive feel
				rampRate := 3.0 * Emax * a.frequency

				if a.interactivePhase > 0 {
					a.interactivePhaseTime += dt

					// Calculate target fields same as Write/Read demo
					targetNormP := (float64(a.interactiveTarget) - 15.5) / 14.5
					var saturateE, settleE float64
					if targetNormP >= 0 {
						saturateE = Emax
						settleRatio := 1.0 - (float64(a.interactiveTarget-16) / 14.0)
						settleE = -Ec * (0.5 + settleRatio*1.0)
					} else {
						saturateE = -Emax
						settleRatio := float64(a.interactiveTarget-1) / 14.0
						settleE = Ec * (0.5 + settleRatio*1.0)
					}

					switch a.interactivePhase {
					case 1: // SATURATE - ramp to saturation
						diff := saturateE - a.electricField
						step := rampRate * dt
						if math.Abs(diff) < step {
							a.electricField = saturateE
						} else if diff > 0 {
							a.electricField += step
						} else {
							a.electricField -= step
						}
						if a.interactivePhaseTime > phaseDuration*0.5 && math.Abs(a.electricField-saturateE) < 0.01*Emax {
							a.interactivePhase = 2
							a.interactivePhaseTime = 0
							a.addLogEntry(fmt.Sprintf("SATURATE done → %.1f MV/cm", saturateE/1e8))
						}

					case 2: // SETTLE - create minor loop
						diff := settleE - a.electricField
						step := rampRate * 0.7 * dt
						if math.Abs(diff) < step {
							a.electricField = settleE
						} else if diff > 0 {
							a.electricField += step
						} else {
							a.electricField -= step
						}
						if a.interactivePhaseTime > phaseDuration*0.4 && math.Abs(a.electricField-settleE) < 0.01*Emax {
							a.interactivePhase = 3
							a.interactivePhaseTime = 0
							a.addLogEntry(fmt.Sprintf("SETTLE → Level %d", a.discreteLevel+1))
						}

					case 3: // HOLD - return to zero
						step := rampRate * dt
						if math.Abs(a.electricField) < step {
							a.electricField = 0
						} else if a.electricField > 0 {
							a.electricField -= step
						} else {
							a.electricField += step
						}
						if a.interactivePhaseTime > phaseDuration*0.3 && math.Abs(a.electricField) < 0.01*Emax {
							a.interactivePhase = 0 // Done - back to idle
							a.addLogEntry(fmt.Sprintf("HOLD Level %d ✓", a.discreteLevel+1))
							a.addLogEntry("Click another level!")
						}

					case 4: // SENSE (read-only mode) - small pulse
						readE := Ec * 0.3
						if a.discreteLevel < 15 {
							readE = -readE
						}
						diff := readE - a.electricField
						step := rampRate * 0.5 * dt
						if math.Abs(diff) < step {
							a.electricField = readE
						} else if diff > 0 {
							a.electricField += step
						} else {
							a.electricField -= step
						}
						if a.interactivePhaseTime > phaseDuration*0.3 {
							// Return to zero
							a.electricField = 0
							a.interactivePhase = 0
							a.addLogEntry(fmt.Sprintf("READ Level %d (|E|<Ec)", a.discreteLevel+1))
							a.addLogEntry("State unchanged ✓")
						}
					}
				}
			}
		}

		// Update physics
		prevP := a.polarization
		a.polarization = a.preisach.Update(a.electricField)
		a.normalizedP = a.preisach.NormalizedPolarization()
		a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * 29))
		if a.discreteLevel < 0 {
			a.discreteLevel = 0
		}
		if a.discreteLevel > 29 {
			a.discreteLevel = 29
		}

		// Calculate energy: integral of E·dP ≈ |E| * |ΔP|
		// During write/read cycles, accumulate energy for the cycle
		if a.waveform == WaveformWriteReadDemo && a.wrdPhase >= 0 && a.wrdPhase <= 3 {
			deltaP := a.polarization - prevP
			// Energy per unit volume: E·dP in J/m³
			// Convert to fJ for a typical cell (assume 100nm x 100nm x 20nm = 2e-22 m³)
			cellVolume := 2e-22 // m³
			energyJ := math.Abs(a.electricField * deltaP) * cellVolume
			energyfJ := energyJ * 1e15 // Convert J to fJ
			a.wrdCycleEnergy += energyfJ
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

		// Update wake-up/fatigue labels (Dr. Tour recommendation)
		cycles, degradation, wakeup := a.preisach.GetFatigueState()
		if a.cyclesLabel != nil {
			if cycles >= 1000000 {
				a.cyclesLabel.SetText(fmt.Sprintf("%.1fM", float64(cycles)/1e6))
			} else if cycles >= 1000 {
				a.cyclesLabel.SetText(fmt.Sprintf("%.1fK", float64(cycles)/1e3))
			} else {
				a.cyclesLabel.SetText(fmt.Sprintf("%d", cycles))
			}
		}
		if a.wakeupLabel != nil {
			a.wakeupLabel.SetText(fmt.Sprintf("%.1f%%", wakeup*100))
		}
		if a.fatigueLabel != nil {
			a.fatigueLabel.SetText(fmt.Sprintf("%.4f%%", degradation*100))
		}

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
			wrdTotalWrites := a.wrdTotalWrites
			wrdSuccessWrites := a.wrdSuccessWrites
			wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
			a.mu.RUnlock()

			switch waveform {
			case WaveformRandomWalk:
				a.statusLabel.SetText(fmt.Sprintf("● Random Walk | Target: %d | Current: %d", rwTarget, level+1))
			case WaveformWriteReadDemo:
				var phaseStr string
				// Log phase transitions (now 5 phases: SATURATE, SETTLE, HOLD, READ, DISPLAY)
				if wrdPhase != lastPhase {
					a.mu.Lock()
					a.lastLogPhase = wrdPhase
					switch wrdPhase {
					case 0:
						// SATURATE: Show domain saturation with energy
						a.addLogEntry(fmt.Sprintf("▓▓ WRITE L%d | E=2Ec | ~10fJ", wrdTarget))
					case 1:
						// SETTLE: Show minor loop formation
						polarization := float64(wrdTarget-15) / 15.0 * 100 // % of Pr
						a.addLogEntry(fmt.Sprintf("██ SETTLE  | P→%.0f%% Pr", polarization))
					case 2:
						// HOLD: Emphasize zero-power retention
						a.addLogEntry(fmt.Sprintf("░░ HOLD L%d | E=0 | 0 fJ!", level+1))
					case 3:
						// READ: Show non-destructive sense
						a.addLogEntry("▒▒ READ    | E<Ec | ~1fJ")
					case 4:
						// RESULT: Show success/tolerance and running stats
						status := "✓ MATCH"
						if wrdRead != wrdTarget {
							diff := abs(wrdRead - wrdTarget)
							if diff == 1 {
								status = fmt.Sprintf("△ ±1 (got %d)", wrdRead)
							} else {
								status = fmt.Sprintf("✗ miss (got %d)", wrdRead)
							}
						}
						successRate := 0.0
						if wrdTotalWrites > 0 {
							successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
						}
						a.addLogEntry(fmt.Sprintf("●● L%d %s [%.0f%% rate]", wrdTarget, status, successRate))
					}
					a.mu.Unlock()
				}

				// Enhanced status with energy metrics (using local copies from RLock above)
				energyTotal := wrdTotalEnergyfJ
				writeCount := wrdTotalWrites

				switch wrdPhase {
				case 0:
					phaseStr = fmt.Sprintf("▓ WRITE L%d | E=2×Ec | ~10fJ switching", wrdTarget)
				case 1:
					polarization := float64(wrdTarget-15) / 15.0 * 100
					phaseStr = fmt.Sprintf("█ SETTLE L%d | P→%.0f%% of Pr", wrdTarget, polarization)
				case 2:
					phaseStr = fmt.Sprintf("░ HOLD L%d | E=0 | ZERO POWER", level+1)
				case 3:
					phaseStr = fmt.Sprintf("▒ READ | Sense L%d | ~1fJ", level+1)
				case 4:
					successRate := 0.0
					if writeCount > 0 {
						successRate = float64(wrdSuccessWrites) / float64(writeCount) * 100
					}
					if wrdRead == wrdTarget {
						phaseStr = fmt.Sprintf("● L%d ✓ | Ops:%d | %.0f%% | %.0fpJ", wrdRead, writeCount, successRate, energyTotal/1000)
					} else {
						phaseStr = fmt.Sprintf("● L%d (want %d) | Ops:%d | %.0f%%", wrdRead, wrdTarget, writeCount, successRate)
					}
				}
				a.statusLabel.SetText(fmt.Sprintf("⚡ FeCIM Write/Read | %s", phaseStr))
			case WaveformInteractive:
				// Interactive mode status
				a.mu.RLock()
				interPhase := a.interactivePhase
				interTarget := a.interactiveTarget
				interIsWrite := a.interactiveIsWrite
				a.mu.RUnlock()

				var phaseStr string
				modeStr := "WRITE"
				if !interIsWrite {
					modeStr = "READ"
				}
				switch interPhase {
				case 0:
					phaseStr = fmt.Sprintf("Click level bar to %s | Current: L%d", modeStr, level+1)
				case 1:
					phaseStr = fmt.Sprintf("SATURATING → L%d...", interTarget)
				case 2:
					phaseStr = fmt.Sprintf("SETTLING → L%d...", interTarget)
				case 3:
					phaseStr = fmt.Sprintf("HOLDING L%d...", level+1)
				case 4:
					phaseStr = fmt.Sprintf("READING L%d...", level+1)
				}
				a.statusLabel.SetText(fmt.Sprintf("🎯 Interactive [%s] | %s", modeStr, phaseStr))
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

// showELI5Dialog displays a compact explanation of hysteresis concepts
func (a *App) showELI5Dialog() {
	// Create content with key concepts from the ELI5 guide
	content := widget.NewLabel(
		"HYSTERESIS EXPLAINED LIKE YOU'RE 5\n\n" +
			"🔁 What is Hysteresis?\n" +
			"Like a rubber band that \"remembers\" being stretched.\n" +
			"The path going UP is different from the path coming DOWN.\n\n" +
			"💾 Why It Matters for Memory?\n" +
			"• Regular memory (DRAM): Like a whiteboard - erase & gone\n" +
			"• FeCIM memory: Like carving in clay - stays after power off!\n\n" +
			"📊 The P-E Loop:\n" +
			"• E = Electric Field (the \"push\" you apply)\n" +
			"• P = Polarization (material's response)\n" +
			"• When E = 0, P stays at ±Pr → MEMORY!\n\n" +
			"🎚️ Why 30 Levels?\n" +
			"• Binary: Like a light switch (ON/OFF) = 1 bit\n" +
			"• FeCIM: Like a dimmer with 30 positions = ~5 bits\n" +
			"• Same chip, 5× more storage!\n\n" +
			"📝 Write vs Read:\n" +
			"• WRITE: |E| > Ec → Data changes\n" +
			"• READ: |E| < Ec → Data unchanged, just sense\n" +
			"• HOLD: E = 0 → Data persists (no power!)\n\n" +
			"🎯 The Key Insight:\n" +
			"Hysteresis isn't a bug - it's the FEATURE that\n" +
			"enables memory! The loop REMEMBERS which\n" +
			"way you pushed it.\n\n" +
			"📚 Full Documentation:\n" +
			"See docs/hysteresis/hysteresis.ELI5.md for\n" +
			"detailed explanations with diagrams.")
	content.Wrapping = fyne.TextWrapWord

	// Create scrollable container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 500))

	// Create dialog (declare as var first so button callback can reference it)
	var dialog *widget.PopUp
	closeBtn := widget.NewButton("Got it!", func() {
		if dialog != nil {
			dialog.Hide()
		}
	})

	dialog = widget.NewModalPopUp(
		container.NewVBox(
			container.NewPadded(scroll),
			closeBtn,
		),
		a.mainWindow.Canvas(),
	)

	dialog.Show()
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
	cache   sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *peplotRenderer) MinSize() fyne.Size {
	return r.plot.minSize
}

func (r *peplotRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("peplotRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *peplotRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("peplotRenderer", r.plot.Size())
	// Always re-layout on Refresh for this dynamic widget (data changes)
	// layoutWithSize handles size validation and cache marking internally
	r.layoutWithSize(r.plot.Size())
}

func (r *peplotRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.plot.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.plot.mu.RLock()
	defer r.plot.mu.RUnlock()

	r.objects = r.objects[:0]

	// Background with subtle border
	bg := canvas.NewRectangle(color.RGBA{0, 40, 90, 255})
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Inner background
	innerBg := canvas.NewRectangle(colorBackground)
	innerBg.Resize(fyne.NewSize(size.Width-4, size.Height-4))
	innerBg.Move(fyne.NewPos(2, 2))
	r.objects = append(r.objects, innerBg)

	// Margins - adjusted for better proportions
	marginLeft := float32(60)
	marginRight := float32(20)
	marginTop := float32(50)
	marginBottom := float32(35)
	plotW := size.Width - marginLeft - marginRight
	plotH := size.Height - marginTop - marginBottom

	// Title inside plot area
	plotTitle := canvas.NewText("P-E Hysteresis Loop", color.RGBA{255, 255, 255, 255})
	plotTitle.TextSize = 16
	plotTitle.TextStyle = fyne.TextStyle{Bold: true}
	plotTitle.Move(fyne.NewPos(size.Width/2-80, 8))
	r.objects = append(r.objects, plotTitle)

	// Legend inside plot area (top-left corner)
	legendX := marginLeft + 10
	legendY := marginTop + 10
	legendItems := []struct {
		color color.RGBA
		text  string
	}{
		{color.RGBA{255, 150, 0, 255}, "Ec"},
		{color.RGBA{0, 200, 150, 255}, "Pr"},
		{colorWarning, "Current"},
	}
	for _, item := range legendItems {
		// Color box - increased from 14 to 16 for better visibility
		box := canvas.NewRectangle(item.color)
		box.Resize(fyne.NewSize(16, 16))
		box.Move(fyne.NewPos(legendX, legendY))
		r.objects = append(r.objects, box)
		// Label - increased from 12pt to 13pt
		label := canvas.NewText(item.text, color.RGBA{200, 200, 200, 255})
		label.TextSize = 13
		label.Move(fyne.NewPos(legendX+20, legendY))
		r.objects = append(r.objects, label)
		legendX += 65
	}

	// Grid lines (fewer divisions to avoid overlap)
	numDivisions := 2 // divisions on each side of center (shows -1, -0.5, 0, 0.5, 1)
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

		// X-axis tick labels (E-field in MV/cm) - only at edges and center
		if i != 0 && (i == -numDivisions || i == numDivisions) {
			eVal := float64(t) * r.plot.eMax / 1e8 // Convert to MV/cm
			eTickLabel := canvas.NewText(fmt.Sprintf("%.1f", eVal), color.RGBA{200, 200, 200, 255})
			eTickLabel.TextSize = 12
			eTickLabel.Move(fyne.NewPos(x-15, marginTop+plotH+5))
			r.objects = append(r.objects, eTickLabel)
		}

		// Y-axis tick labels (P in µC/cm²) - only at edges and center
		if i != 0 && (i == -numDivisions || i == numDivisions) {
			pVal := float64(t) * r.plot.pMax * 100 // Convert to µC/cm²
			pTickLabel := canvas.NewText(fmt.Sprintf("%.0f", pVal), color.RGBA{200, 200, 200, 255})
			pTickLabel.TextSize = 12
			pTickLabel.Move(fyne.NewPos(marginLeft-35, y-7))
			r.objects = append(r.objects, pTickLabel)
		}
	}

	// Additional subtle grid lines for easier value reading
	// These are finer subdivisions with very low alpha to not distract from main plot
	subtleGridColor := color.RGBA{0, 80, 140, 25} // Very low alpha for subtle appearance
	numSubdivisions := 4                          // More divisions for finer grid
	for i := -numSubdivisions; i <= numSubdivisions; i++ {
		// Skip the main grid lines (already drawn above)
		if i == 0 || i == -numSubdivisions || i == numSubdivisions || i == -numSubdivisions/2 || i == numSubdivisions/2 {
			continue
		}
		t := float32(i) / float32(numSubdivisions)

		// Subtle vertical grid line
		x := marginLeft + plotW/2 + t*plotW/2
		vLine := canvas.NewLine(subtleGridColor)
		vLine.Position1 = fyne.NewPos(x, marginTop)
		vLine.Position2 = fyne.NewPos(x, marginTop+plotH)
		vLine.StrokeWidth = 1
		r.objects = append(r.objects, vLine)

		// Subtle horizontal grid line
		y := marginTop + plotH/2 - t*plotH/2
		hLine := canvas.NewLine(subtleGridColor)
		hLine.Position1 = fyne.NewPos(marginLeft, y)
		hLine.Position2 = fyne.NewPos(marginLeft+plotW, y)
		hLine.StrokeWidth = 1
		r.objects = append(r.objects, hLine)
	}

	// Center coordinates
	centerX := marginLeft + plotW/2
	centerY := marginTop + plotH/2

	// Zero label (single, at corner to avoid overlap)
	zeroLabel := canvas.NewText("0", color.RGBA{200, 200, 200, 255})
	zeroLabel.TextSize = 12
	zeroLabel.Move(fyne.NewPos(centerX-10, centerY+4))
	r.objects = append(r.objects, zeroLabel)

	// Axis title labels (positioned at ends of axes) - larger
	eLabel := canvas.NewText("E (MV/cm)", color.RGBA{0, 212, 255, 255})
	eLabel.TextSize = 13
	eLabel.TextStyle = fyne.TextStyle{Bold: true}
	eLabel.Move(fyne.NewPos(marginLeft+plotW-70, marginTop+plotH+15))
	r.objects = append(r.objects, eLabel)

	pLabelText := canvas.NewText("P (µC/cm²)", color.RGBA{0, 212, 255, 255})
	pLabelText.TextSize = 13
	pLabelText.TextStyle = fyne.TextStyle{Bold: true}
	pLabelText.Move(fyne.NewPos(marginLeft-55, marginTop-18))
	r.objects = append(r.objects, pLabelText)

	// Ec markers (vertical dashed lines at ±Ec)
	// Assuming Ec is roughly at 40% of eMax (since eMax = 2.5*Ec)
	ecRatio := float32(1.0 / 2.5) // Ec / eMax
	ecPosX := centerX + ecRatio*plotW/2
	ecNegX := centerX - ecRatio*plotW/2

	// +Ec marker (dashed effect with shorter line)
	ecPosLine := canvas.NewLine(color.RGBA{255, 150, 0, 120})
	ecPosLine.Position1 = fyne.NewPos(ecPosX, marginTop)
	ecPosLine.Position2 = fyne.NewPos(ecPosX, marginTop+plotH)
	ecPosLine.StrokeWidth = 2
	r.objects = append(r.objects, ecPosLine)
	// Label inside plot area
	ecPosLabel := canvas.NewText("+Ec", color.RGBA{255, 150, 0, 255})
	ecPosLabel.TextSize = 11
	ecPosLabel.Move(fyne.NewPos(ecPosX+4, marginTop+5))
	r.objects = append(r.objects, ecPosLabel)

	// -Ec marker
	ecNegLine := canvas.NewLine(color.RGBA{255, 150, 0, 120})
	ecNegLine.Position1 = fyne.NewPos(ecNegX, marginTop)
	ecNegLine.Position2 = fyne.NewPos(ecNegX, marginTop+plotH)
	ecNegLine.StrokeWidth = 2
	r.objects = append(r.objects, ecNegLine)
	// Label inside plot area
	ecNegLabel := canvas.NewText("-Ec", color.RGBA{255, 150, 0, 255})
	ecNegLabel.TextSize = 11
	ecNegLabel.Move(fyne.NewPos(ecNegX-28, marginTop+5))
	r.objects = append(r.objects, ecNegLabel)

	// Pr markers (horizontal dashed lines at ±Pr)
	prRatio := float32(0.8 / 1.2) // Pr / pMax
	prPosY := centerY - prRatio*plotH/2
	prNegY := centerY + prRatio*plotH/2

	// +Pr marker
	prPosLine := canvas.NewLine(color.RGBA{0, 200, 150, 120})
	prPosLine.Position1 = fyne.NewPos(marginLeft, prPosY)
	prPosLine.Position2 = fyne.NewPos(marginLeft+plotW, prPosY)
	prPosLine.StrokeWidth = 2
	r.objects = append(r.objects, prPosLine)
	// Label inside plot area
	prPosLabel := canvas.NewText("+Pr", color.RGBA{0, 200, 150, 255})
	prPosLabel.TextSize = 11
	prPosLabel.Move(fyne.NewPos(marginLeft+plotW-30, prPosY-14))
	r.objects = append(r.objects, prPosLabel)

	// -Pr marker
	prNegLine := canvas.NewLine(color.RGBA{0, 200, 150, 120})
	prNegLine.Position1 = fyne.NewPos(marginLeft, prNegY)
	prNegLine.Position2 = fyne.NewPos(marginLeft+plotW, prNegY)
	prNegLine.StrokeWidth = 2
	r.objects = append(r.objects, prNegLine)
	// Label inside plot area
	prNegLabel := canvas.NewText("-Pr", color.RGBA{0, 200, 150, 255})
	prNegLabel.TextSize = 11
	prNegLabel.Move(fyne.NewPos(marginLeft+plotW-28, prNegY+4))
	r.objects = append(r.objects, prNegLabel)

	// Plot the hysteresis data
	if len(r.plot.eData) > 1 {
		for i := 1; i < len(r.plot.eData); i++ {
			// Map data to screen coordinates
			x1 := marginLeft + plotW/2 + float32(r.plot.eData[i-1]/r.plot.eMax)*plotW/2
			y1 := centerY - float32(r.plot.pData[i-1]/r.plot.pMax)*plotH/2
			x2 := marginLeft + plotW/2 + float32(r.plot.eData[i]/r.plot.eMax)*plotW/2
			y2 := centerY - float32(r.plot.pData[i]/r.plot.pMax)*plotH/2

			// Color based on age (fade effect) - increased fade from 0.7 to 0.85 for better trace distinction
			age := float64(len(r.plot.eData)-i) / float64(len(r.plot.eData))
			alpha := uint8(255 * (1 - age*0.85))

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

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
}

func (r *peplotRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *peplotRenderer) Destroy() {}

// ============================================================
// Custom Level Indicator Widget
// ============================================================

// LevelIndicator shows the 30 discrete states (clickable for interactive mode)
type LevelIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	level   int
	minSize fyne.Size

	// Interactive mode callback - called when user clicks a level
	OnLevelClicked func(targetLevel int)
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

// Tapped implements fyne.Tappable - allows clicking to select a level
func (l *LevelIndicator) Tapped(e *fyne.PointEvent) {
	if l.OnLevelClicked == nil {
		return
	}
	// Calculate which level was clicked based on Y position
	size := l.Size()
	if size.Height <= 0 {
		return
	}
	// Levels are drawn from top (30) to bottom (1)
	// Account for margins (title area at top)
	marginTop := float32(25)
	marginBottom := float32(5)
	usableHeight := size.Height - marginTop - marginBottom

	// Calculate relative position in the level area
	relY := e.Position.Y - marginTop
	if relY < 0 {
		relY = 0
	}
	if relY > usableHeight {
		relY = usableHeight
	}

	// Convert to level (30 at top, 1 at bottom)
	levelFloat := 30.0 - (float64(relY) / float64(usableHeight) * 29.0)
	targetLevel := int(math.Round(levelFloat))
	if targetLevel < 1 {
		targetLevel = 1
	}
	if targetLevel > 30 {
		targetLevel = 30
	}

	l.OnLevelClicked(targetLevel)
}

func (l *LevelIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &levelRenderer{indicator: l}
}

// Ensure LevelIndicator implements Tappable
var _ fyne.Tappable = (*LevelIndicator)(nil)

type levelRenderer struct {
	indicator *LevelIndicator
	objects   []fyne.CanvasObject
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *levelRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *levelRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("levelRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *levelRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("levelRenderer", r.indicator.Size())
	size := r.indicator.Size()
	// Always re-layout on Refresh for this dynamic widget (level changes)
	r.layoutWithSize(size)
}

func (r *levelRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.indicator.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.indicator.mu.RLock()
	level := r.indicator.level
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]

	// Allow level indicator to expand to match plot height
	// Only constrain width to keep it compact
	minSize := r.indicator.minSize
	if size.Width > minSize.Width*1.5 {
		size.Width = minSize.Width * 1.5
	}
	// Don't constrain height - let it match the plot

	// Background with subtle border
	border := canvas.NewRectangle(color.RGBA{0, 100, 180, 255})
	border.Resize(size)
	r.objects = append(r.objects, border)

	bg := canvas.NewRectangle(color.RGBA{0, 40, 80, 255})
	bg.Resize(fyne.NewSize(size.Width-4, size.Height-4))
	bg.Move(fyne.NewPos(2, 2))
	r.objects = append(r.objects, bg)

	// Title at top - increased from 12pt to 14pt for better visibility
	title := canvas.NewText("30 LEVELS", color.RGBA{0, 212, 255, 255})
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(6, 8))
	r.objects = append(r.objects, title)

	// Bits label - increased from 10pt to 12pt for better readability
	bitsLabel := canvas.NewText("4.9 bits", color.RGBA{180, 180, 180, 255})
	bitsLabel.TextSize = 12
	bitsLabel.Move(fyne.NewPos(10, 26))
	r.objects = append(r.objects, bitsLabel)

	// Draw 30 level segments
	// Match P-E plot margins for proper Y-axis alignment
	marginH := float32(50)      // Top margin - matches plot marginTop
	marginBottom := float32(35) // Bottom margin - matches plot marginBottom
	marginW := float32(6)
	labelW := float32(28)
	barW := size.Width - 2*marginW - labelW
	totalH := size.Height - marginH - marginBottom

	// Calculate center Y (same as plot's centerY)
	centerY := marginH + totalH/2

	// The plot shows P from -Ps*1.2 to +Ps*1.2
	// Levels 1-30 should map to -Ps to +Ps (the inner 1/1.2 = 83% of plot range)
	pMaxScale := float32(1.2)
	// The actual Y range for the 30 levels (±Ps portion of the plot)
	levelRangeH := totalH / pMaxScale // ~83% of totalH
	segH := levelRangeH / 30
	gap := float32(1)

	// Y positions for level range (screen coords: y increases downward)
	// levelTop = where level 30 (+Ps) should be = centerY - levelRangeH/2
	// levelBottom = where level 1 (-Ps) should be = centerY + levelRangeH/2
	levelTop := centerY - levelRangeH/2

	for i := 0; i < 30; i++ {
		// Level i=0 is level 1 (bottom, -Ps), i=29 is level 30 (top, +Ps)
		// Invert: level 30 at top, level 1 at bottom
		// y = levelTop + (29-i) * segH
		y := levelTop + float32(29-i)*segH

		// Color gradient
		t := float64(i) / 29.0
		var segColor color.RGBA
		if i == level {
			// Current level - bright yellow highlight with glow
			segColor = colorWarning
		} else if t < 0.5 {
			t2 := t * 2
			segColor = color.RGBA{
				uint8(80 + t2*175),
				uint8(120 + t2*135),
				255,
				180,
			}
		} else {
			t2 := (t - 0.5) * 2
			segColor = color.RGBA{
				255,
				uint8(255 - t2*175),
				uint8(255 - t2*175),
				180,
			}
		}

		// Current level gets extra emphasis
		if i == level {
			// Glow effect
			glow := canvas.NewRectangle(color.RGBA{colorWarning.R, colorWarning.G, colorWarning.B, 100})
			glow.Resize(fyne.NewSize(barW+6, segH+4))
			glow.Move(fyne.NewPos(marginW-3, y-2))
			r.objects = append(r.objects, glow)
		}

		seg := canvas.NewRectangle(segColor)
		seg.Resize(fyne.NewSize(barW, segH-gap))
		seg.Move(fyne.NewPos(marginW, y))
		r.objects = append(r.objects, seg)

		// Level number for every 5th level and current level
		if i%5 == 0 || i == 29 || i == level {
			labelColor := colorAxis
			fontSize := float32(11)
			if i == level {
				labelColor = colorWarning
				fontSize = 12
			}
			label := canvas.NewText(fmt.Sprintf("%d", i+1), labelColor)
			label.TextSize = fontSize
			if i == level {
				label.TextStyle = fyne.TextStyle{Bold: true}
			}
			label.Move(fyne.NewPos(marginW+barW+4, y+(segH-gap)/2-6))
			r.objects = append(r.objects, label)
		}
	}

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
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
		minSize: fyne.NewSize(180, 220),
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
	cache   sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *cellRenderer) MinSize() fyne.Size {
	return r.cell.minSize
}

func (r *cellRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("cellRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *cellRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("cellRenderer", r.cell.Size())
	size := r.cell.Size()
	// Always re-layout on Refresh for this dynamic widget (level changes)
	r.layoutWithSize(size)
}

func (r *cellRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.cell.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.cell.mu.RLock()
	level := r.cell.level
	r.cell.mu.RUnlock()

	r.objects = r.objects[:0]

	// Constrain to minimum size to prevent growing
	minSize := r.cell.minSize
	if size.Width > minSize.Width {
		size.Width = minSize.Width
	}
	if size.Height > minSize.Height {
		size.Height = minSize.Height
	}

	// Background with subtle gradient effect
	bg := canvas.NewRectangle(color.RGBA{0, 35, 70, 255})
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Calculate cell size and position - larger cell
	margin := float32(15)
	cellSize := size.Width - 2*margin
	if size.Height-70 < cellSize {
		cellSize = size.Height - 70
	}
	cellX := (size.Width - cellSize) / 2
	cellY := margin + 5

	// Outer glow (based on cell color)
	t := float64(level) / 29.0
	var glowColor color.RGBA
	if t < 0.5 {
		glowColor = color.RGBA{100, 150, 255, 60} // Blue glow
	} else {
		glowColor = color.RGBA{255, 150, 100, 60} // Red glow
	}
	outerGlow := canvas.NewRectangle(glowColor)
	outerGlow.Resize(fyne.NewSize(cellSize+20, cellSize+20))
	outerGlow.Move(fyne.NewPos(cellX-10, cellY-10))
	r.objects = append(r.objects, outerGlow)

	// Cell border (electrode representation) - cyan accent
	borderWidth := float32(4)
	border := canvas.NewRectangle(color.RGBA{0, 180, 220, 255})
	border.Resize(fyne.NewSize(cellSize+borderWidth*2, cellSize+borderWidth*2))
	border.Move(fyne.NewPos(cellX-borderWidth, cellY-borderWidth))
	r.objects = append(r.objects, border)

	// Cell color based on level - intuitive gradient:
	// Level 1 (negative P) = Blue, Level 15 = White/neutral, Level 30 (positive P) = Red
	// This matches physics: negative polarization -> positive polarization
	// t is already calculated above (0.0 to 1.0)
	var cellColor color.RGBA
	if t < 0.5 {
		// Blue to white transition (levels 1-15, negative to neutral)
		t2 := t * 2 // 0 to 1
		cellColor = color.RGBA{
			uint8(80 + t2*175),  // R: 80 -> 255
			uint8(120 + t2*135), // G: 120 -> 255
			255,                 // B stays high
			255,
		}
	} else {
		// White to red transition (levels 16-30, neutral to positive)
		t2 := (t - 0.5) * 2 // 0 to 1
		cellColor = color.RGBA{
			255,                 // R stays high
			uint8(255 - t2*175), // G: 255 -> 80
			uint8(255 - t2*175), // B: 255 -> 80
			255,
		}
	}

	// The memory cell square
	cell := canvas.NewRectangle(cellColor)
	cell.Resize(fyne.NewSize(cellSize, cellSize))
	cell.Move(fyne.NewPos(cellX, cellY))
	r.objects = append(r.objects, cell)

	// Inner glow effect
	glowSize := cellSize * 0.5
	glowX := cellX + (cellSize-glowSize)/2
	glowY := cellY + (cellSize-glowSize)/2
	glow := canvas.NewRectangle(color.RGBA{
		uint8(min(int(cellColor.R)+20, 255)),
		uint8(min(int(cellColor.G)+20, 255)),
		uint8(min(int(cellColor.B)+20, 255)),
		80,
	})
	glow.Resize(fyne.NewSize(glowSize, glowSize))
	glow.Move(fyne.NewPos(glowX, glowY))
	r.objects = append(r.objects, glow)

	// Level text inside cell - larger and centered
	levelStr := fmt.Sprintf("%d", level+1)
	// Choose text color based on background brightness
	var textColor color.RGBA
	brightness := (int(cellColor.R) + int(cellColor.G) + int(cellColor.B)) / 3
	if brightness > 180 {
		textColor = color.RGBA{0, 0, 0, 230} // Dark text on light background
	} else {
		textColor = color.RGBA{255, 255, 255, 230} // Light text on dark background
	}
	// Scale text size with cell size
	textSize := cellSize * 0.35
	if textSize < 24 {
		textSize = 24
	}
	if textSize > 48 {
		textSize = 48
	}
	levelText := canvas.NewText(levelStr, textColor)
	levelText.TextSize = textSize
	levelText.TextStyle = fyne.TextStyle{Bold: true}
	// Calculate width based on text size
	textW := float32(len(levelStr)) * textSize * 0.6
	levelText.Move(fyne.NewPos(cellX+(cellSize-textW)/2, cellY+cellSize/2-textSize/2))
	r.objects = append(r.objects, levelText)

	// Label below cell (centered) - larger
	labelY := cellY + cellSize + 8
	levelLabelStr := fmt.Sprintf("Level %d/30", level+1)
	levelLabel := canvas.NewText(levelLabelStr, color.RGBA{220, 220, 220, 255})
	levelLabel.TextSize = 14
	levelLabel.TextStyle = fyne.TextStyle{Bold: true}
	labelW := float32(len(levelLabelStr)) * 8
	levelLabel.Move(fyne.NewPos(cellX+(cellSize-labelW)/2, labelY))
	r.objects = append(r.objects, levelLabel)

	// State description (centered) - larger
	var stateText string
	if level < 10 {
		stateText = "Negative P"
	} else if level > 19 {
		stateText = "Positive P"
	} else {
		stateText = "Intermediate"
	}
	stateLabel := canvas.NewText(stateText, color.RGBA{180, 180, 180, 255})
	stateLabel.TextSize = 12
	stateLabelW := float32(len(stateText)) * 7
	stateLabel.Move(fyne.NewPos(cellX+(cellSize-stateLabelW)/2, labelY+18))
	r.objects = append(r.objects, stateLabel)

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
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
		minSize: fyne.NewSize(200, 70),
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
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *modeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *modeRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("modeRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *modeRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("modeRenderer", r.indicator.Size())
	size := r.indicator.Size()
	// Always re-layout on Refresh for this dynamic widget (mode changes)
	r.layoutWithSize(size)
}

func (r *modeRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.indicator.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.indicator.mu.RLock()
	isWrite := r.indicator.isWrite
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

	// Box colors
	var bgColor, borderColor color.RGBA
	var modeText, conditionText string

	if isWrite {
		bgColor = color.RGBA{180, 50, 50, 255} // Red background
		borderColor = color.RGBA{255, 100, 100, 255}
		modeText = "WRITE"
		conditionText = "|E| > Ec"
	} else {
		bgColor = color.RGBA{50, 150, 80, 255} // Green background
		borderColor = color.RGBA{100, 220, 130, 255}
		modeText = "READ"
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

	// Mode text (centered) - larger
	modeLabel := canvas.NewText(modeText, color.White)
	modeLabel.TextSize = 22
	modeLabel.TextStyle = fyne.TextStyle{Bold: true}
	modeTextW := float32(len(modeText)) * 13
	modeLabel.Move(fyne.NewPos((size.Width-modeTextW)/2, 12))
	r.objects = append(r.objects, modeLabel)

	// Condition text (centered, bottom) - larger
	condLabel := canvas.NewText(conditionText, color.RGBA{230, 230, 230, 255})
	condLabel.TextSize = 14
	condLabelW := float32(len(conditionText)) * 8
	condLabel.Move(fyne.NewPos((size.Width-condLabelW)/2, 40))
	r.objects = append(r.objects, condLabel)

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
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

// fixedMinWidthLayout enforces a minimum width but allows expansion
type fixedMinWidthLayout struct {
	minWidth float32
}

func (l *fixedMinWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if o.Visible() {
			minH = fyne.Max(minH, o.MinSize().Height)
		}
	}
	return fyne.NewSize(l.minWidth, minH)
}

func (l *fixedMinWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		// Only use the given width, but respect child's MinSize for height
		// This prevents vertical stretching
		childMinSize := o.MinSize()
		o.Resize(fyne.NewSize(size.Width, childMinSize.Height))
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
	// Return fixed minimum to prevent layout recalculation issues
	return fyne.NewSize(100, minH)
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
