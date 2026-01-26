// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// Uses fyne.io/fyne/v2 for cross-platform native GUI with proper graphics.
package gui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/ferroelectric"
	"multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/gui/widgets"
	"multilayer-ferroelectric-cim-visualizer/shared/logging"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// Package-level logger for hysteresis GUI
var log *logging.Logger

func init() {
	log = logging.NewLogger("hysteresis")
}

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

	// Manual mode click-to-level animation (merged from Interactive mode)
	manualAnimating   bool    // True when animating to clicked level
	manualTargetLevel int     // Target level from click
	manualStartLevel  int     // Level at animation start (prevents feedback loop)
	manualPhase       int     // 0=idle, 1=saturate, 2=settle, 3=hold
	manualPhaseTime   float64 // Time in current phase

	// UI components
	plot           *widgets.PEPlot
	levelIndicator *widgets.LevelIndicator
	cellViz        *widgets.CellVisualizer
	modeIndicator  *widgets.ModeIndicator // WRITE/READ mode indicator with colored box
	eFieldSlider   *widget.Slider
	eFieldLabel    *widget.Label
	pLabel         *widget.Label
	levelLabel     *widget.Label
	stateLabel     *widget.Label // State description (Negative P, Intermediate, Positive P)
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
}

// WaveformType represents the input waveform
type WaveformType int

const (
	WaveformManual WaveformType = iota
	WaveformSine
	WaveformTriangle
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
	case WaveformWriteReadDemo:
		return "Write/Read Demo"
	default:
		return "Unknown"
	}
}

// WriteReadDebugLog stores debug data for write/read operations
type WriteReadDebugLog struct {
	Timestamp string           `json:"timestamp"`
	Material  string           `json:"material"`
	Ec        float64          `json:"ec_v_per_m"`
	EcMVcm    float64          `json:"ec_mv_per_cm"`
	Ps        float64          `json:"ps_c_per_m2"`
	Cycles    []WriteReadCycle `json:"cycles"`
}

// WriteReadCycle stores one complete write/read cycle
type WriteReadCycle struct {
	CycleNum    int              `json:"cycle_num"`
	TargetLevel int              `json:"target_level"`
	StartLevel  int              `json:"start_level"`
	ReadLevel   int              `json:"read_level"`
	Success     bool             `json:"success"`
	Phases      []WriteReadPhase `json:"phases"`
}

// WriteReadPhase stores data for one phase of the write/read cycle
type WriteReadPhase struct {
	Phase       string           `json:"phase"`
	Duration    float64          `json:"duration_s"`
	EFieldStart float64          `json:"e_field_start_mv_cm"`
	EFieldEnd   float64          `json:"e_field_end_mv_cm"`
	EFieldPeak  float64          `json:"e_field_peak_mv_cm"`
	PStart      float64          `json:"p_start_uc_cm2"`
	PEnd        float64          `json:"p_end_uc_cm2"`
	LevelStart  int              `json:"level_start"`
	LevelEnd    int              `json:"level_end"`
	Samples     []PhaseDataPoint `json:"samples,omitempty"`
}

// PhaseDataPoint stores a single data point during a phase
type PhaseDataPoint struct {
	Time  float64 `json:"t"`
	E     float64 `json:"e"`
	P     float64 `json:"p"`
	Level int     `json:"level"`
}

// saveDebugLog saves the debug log to a JSON file
// THREAD-SAFE: No UI updates - only file I/O operations. Safe to call from goroutines without fyne.Do().
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

	// Register keyboard shortcuts
	a.mainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		a.handleKeyPress(ke)
	})

	// Start simulation loop
	a.running = true
	go a.simulationLoop()

	a.mainWindow.ShowAndRun()
	a.running = false
	return nil
}

func (a *App) createUI() fyne.CanvasObject {
	// SAFETY: No mutex needed - createUI() called from run() before simulation goroutine starts (line 319).
	// No concurrent access to a.material exists during UI initialization.

	// Create cell visualizer (THE memory cell!) - prominent size
	a.cellViz = widgets.NewCellVisualizer()
	a.cellViz.SetMinSize(fyne.NewSize(180, 200))

	// Create P-E plot - will expand to fill space
	a.plot = widgets.NewPEPlot(a.material.Ec*2.5, a.material.Ps*1.2, ColorBackground, ColorGrid, ColorAxis, ColorPositive, ColorNegative, ColorWarning)
	a.plot.SetMinSize(fyne.NewSize(400, 350))

	// Create level indicator (wider for better labels, clickable in Manual mode)
	a.levelIndicator = widgets.NewLevelIndicator()
	a.levelIndicator.SetMinSize(fyne.NewSize(80, 350))
	// Wire up click callback for Manual mode
	a.levelIndicator.OnLevelClicked = func(targetLevel int) {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.waveform == WaveformManual && !a.manualAnimating {
			// Start animation to target level
			a.manualTargetLevel = targetLevel
			a.manualStartLevel = a.discreteLevel + 1 // Capture starting level (1-indexed)
			a.manualAnimating = true
			a.manualPhase = 1 // Start saturate phase
			a.manualPhaseTime = 0
			a.addLogEntry(fmt.Sprintf("WRITE → Level %d", targetLevel))
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
	cellTitle.TextSize = 16
	cellTitle.TextStyle = fyne.TextStyle{Bold: true}
	cellTitle.Alignment = fyne.TextAlignCenter

	cellUnderline := canvas.NewRectangle(color.RGBA{0, 212, 255, 150})
	cellUnderline.SetMinSize(fyne.NewSize(100, 2))

	cellHeader := container.NewVBox(
		container.NewCenter(cellTitle),
		container.NewCenter(cellUnderline),
	)

	// Fixed header: Cell visualization and basic level display
	cellDisplay := container.NewVBox(
		cellHeader,
		container.NewCenter(a.cellViz),
	)

	// Scrollable content: all information panels with spacing
	leftScrollableContent := container.NewVBox(
		info,
		widget.NewSeparator(),
		slidePanel,
		widget.NewSeparator(),
		logPanel,
	)
	leftScroll := container.NewScroll(leftScrollableContent)
	leftScroll.SetMinSize(fyne.NewSize(200, 0))

	// Left column: Fixed cell at top, scrollable info below
	leftColumn := container.NewBorder(
		cellDisplay,
		nil, nil, nil,
		container.NewPadded(leftScroll),
	)

	// Right column: Controls only (compact)
	controlsWithMinWidth := container.New(&fixedMinWidthLayout{minWidth: 240},
		container.NewPadded(controls),
	)
	rightColumn := container.NewVBox(
		controlsWithMinWidth,
		layout.NewSpacer(),
	)

	// Plot + level in same row using Border layout
	// Border allows level indicator to be tappable (HSplit may intercept events)
	plotAndLevel := container.NewBorder(
		nil, nil, nil,
		a.levelIndicator,
		a.plot,
	)

	// Status bar at bottom
	a.statusLabel = widget.NewLabel("Running...")
	statusBar := container.NewHBox(
		layout.NewSpacer(),
		a.statusLabel,
		layout.NewSpacer(),
	)

	// Adaptive layout: Left | Center | Right
	// Zones: [0]=Left (Cell+Info), [1]=Plot+Level, [2]=Controls
	zones := []fyne.CanvasObject{
		leftColumn,
		plotAndLevel,
		rightColumn,
	}
	tabLabels := []string{"Info", "P-E Plot", "Controls"}

	adaptive := sharedwidgets.NewAdaptiveLayout(zones, tabLabels)
	adaptive.SetDesktopLayout(func(zones []fyne.CanvasObject) fyne.CanvasObject {
		// Desktop: Left (Cell+Info) | Plot | Controls
		innerSplit := container.NewHSplit(zones[1], zones[2])
		innerSplit.SetOffset(0.75) // Plot gets 75%, controls get 25%

		outerSplit := container.NewHSplit(zones[0], innerSplit)
		outerSplit.SetOffset(0.22) // Left column gets 22%

		return outerSplit
	})

	return container.NewBorder(
		nil,
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
