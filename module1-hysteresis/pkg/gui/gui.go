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

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Package-level logger for hysteresis GUI
// Initialized in run() after EnableFileLogging() is called in main
var log *logging.Logger

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

	// Core Logic Components (Refactoring)
	calibManager    *algo.CalibrationManager
	writeController *controller.WriteController

	// Physics
	material  *ferroelectric.HZOMaterial
	preisach  *ferroelectric.MayergoyzPreisach
	materials []*ferroelectric.HZOMaterial
	matIndex  int
	numLevels int // Number of discrete analog levels (default 30)

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

	// Write-Verify-Retry loop tracking (retries indefinitely until target hit)
	wrdRetryCount int // Current retry count for this target

	// Phase transition tracking for logging
	wrdResetStartP float64 // Polarization at start of RESET phase
	wrdResetEndP   float64 // Polarization at end of RESET phase
	wrdResetEndLvl int     // Level at end of RESET phase
	wrdWriteStartP float64 // Polarization at start of WRITE phase
	wrdWriteEndP   float64 // Polarization at end of WRITE phase
	wrdWriteEndLvl int     // Level at end of WRITE phase
	wrdReadStartP  float64 // Polarization at start of READ phase

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

	// Time-resolved switching animation (KAI dynamics visualization)
	timeResAnimating  bool      // True when animating time-resolved switching
	timeResDataTimes  []float64 // Time points from SimulateDomainSwitching
	timeResDataPols   []float64 // Polarization values from SimulateDomainSwitching
	timeResDataSwitch []int     // Switched hysteron counts
	timeResIndex      int       // Current index into time-resolved data arrays

	// Level calibration data (populated at startup/material change)
	// Maps field values needed to reach each level from known starting states
	calibrationUp    []float64 // Field needed to reach level N from level 1 (ascending)
	calibrationDown  []float64 // Field needed to reach level N from level N (descending)
	calibrated       bool      // Whether calibration has been performed
	needsCalibration bool      // True if calibration should run on first manual/WRD mode use

	// Adaptive calibration bounds (binary search approach)
	// Tracks proven lower/upper bounds for each level to converge faster
	calibUpLow    []float64 // Proven lower bound (field too weak, undershot)
	calibUpHigh   []float64 // Proven upper bound (field too strong, overshot)
	calibDownLow  []float64 // Proven lower bound (more negative, overshot down)
	calibDownHigh []float64 // Proven upper bound (less negative, undershot down)
	lastErrorUp   []int     // Last error for each level (for oscillation detection)
	lastErrorDown []int     // Last error for each level (for oscillation detection)
	relaxCompUp   []float64 // Relaxation compensation factors (ascending, 0-indexed)
	relaxCompDown []float64 // Relaxation compensation factors (descending, 0-indexed)

	// ISPP (Incremental Step Pulse Programming) state
	// True ISPP: climb hysteresis S-curve incrementally instead of single calibrated pulse
	isppEnabled        bool    // Whether to use ISPP algorithm (default: true)
	isppPulseCount     int     // Current pulse number in ISPP sequence (1-based)
	isppCurrentVoltage float64 // Current ISPP pulse voltage (E-field)
	isppStartVoltage   float64 // ISPP starting voltage for this write
	isppVoltageStep    float64 // Voltage increment per ISPP step (default: 0.05*Ec)
	isppMaxPulses      int     // Maximum ISPP pulses before giving up (default: 10)
	isppPhase          int     // ISPP sub-phase: 0=APPLY, 1=WAIT, 2=VERIFY, 3=ADJUST
	isppPhaseTimer     float64 // Timer for ISPP sub-phases
	isppLastVerifyLvl  int     // Level at last ISPP verification (for detecting stuck states)
	isppTotalPulses    int     // Total ISPP pulses across all retries for current write target

	// ISPP oscillation/bisection tracking (convergence improvement)
	isppLastError        int     // Last level error for oscillation detection
	isppOscillationCount int     // Number of direction sign changes (oscillations)
	isppUseBisection     bool    // True when using bisection mode after oscillations
	isppReversedOnce     bool    // True after first direction reversal
	isppLowVoltage       float64 // Lower bound voltage for bisection
	isppHighVoltage      float64 // Upper bound voltage for bisection

	// Temperature-aware calibration (v2)
	calibrationTemp  float64                  // Temperature of current active calibration (K)
	tempCalibrations map[int]*TempCalibration // Cache of calibrations at key temperatures (key: temp in K)

	// UI components
	plot            *widgets.PEPlot
	levelIndicator  *widgets.LevelIndicator
	cellViz         *widgets.CellVisualizer
	phaseIndicator  *widgets.PhaseIndicator // State machine phase indicator (RESET|SETTLE|WRITE|READ|VERIFY)
	eFieldSlider    *widget.Slider
	eFieldLabel     *widget.Label
	eFieldModeLabel *widget.Label // Shows "MANUAL" or "AUTO" for slider control mode
	pLabel          *widget.Label
	levelLabel      *widget.Label
	stateLabel      *widget.Label  // State description (Negative P, Intermediate, Positive P)
	materialBtn     *widget.Button // Opens material picker, shows current material name
	waveformSelect  *widget.Select
	statusLabel     *widget.Label
	pauseBtn        *widget.Button

	// Wake-up/Fatigue display labels (Dr. Tour recommendation)
	cyclesLabel     *widget.Label
	wakeupLabel     *widget.Label
	fatigueLabel    *widget.Label
	cyclePhaseLabel *widget.Label // L01: Shows current phase (WAKE-UP, STABLE, FATIGUE)

	// Temperature-dependent metrics labels
	effEcLabel      *widget.Label
	effPrLabel      *widget.Label
	squarenessLabel *widget.Label
	switchedLabel   *widget.Label

	// Levels selector
	levelsEntry *widget.Entry
	levelsLabel *widget.Label

	// Educational slide and log
	slideTitle   *widget.Label
	slideText    *widget.Label
	logText      *widget.Label
	logEntries   []string
	maxLogLines  int
	lastLogPhase int            // Track phase changes to avoid duplicate logs
	logVerbose   bool           // M07: Toggle between verbose and compact log views
	logToggleBtn *widget.Button // M07: Button to toggle log verbosity

	// Simulation vs Experiment comparison widget (H16)
	simVsExpWidget *widgets.SimVsExpComparison

	// ISPP visualization widget (H14)
	isppWidget *widgets.ISPPVisualization

	// Shared ISPP calculator
	isppCalc *physics.ISPPCalculator

	// State stability indicator (M12)
	stabilityIndicator *widgets.StabilityIndicator
}

// WaveformType represents the input waveform
type WaveformType int

const (
	WaveformManual WaveformType = iota
	WaveformSine
	WaveformTriangle
	WaveformWriteReadDemo
	WaveformTimeResolved
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
	case WaveformTimeResolved:
		return "Time-Resolved Switching"
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
	materials := ferroelectric.AllMaterials()

	mat := materials[0]
	numLevels := 30         // Default: FeCIM's 30 discrete analog states
	preisachGridSize := 200 // High-resolution physics simulation (independent of quantization)
	preisach := ferroelectric.NewMayergoyzPreisach(mat, preisachGridSize)

	// Refactoring: Initialize managers
	calibManager := algo.NewCalibrationManager(numLevels)
	writeController := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, calibManager)

	return &App{
		calibManager:     calibManager,
		writeController:  writeController,
		material:         mat,
		preisach:         preisach,
		materials:        materials,
		matIndex:         0,
		numLevels:        numLevels,
		calibrationUp:    make([]float64, numLevels),
		calibrationDown:  make([]float64, numLevels),
		calibrationTemp:  300, // Default room temperature
		tempCalibrations: make(map[int]*TempCalibration),
		maxHistory:       2000,
		eHistory:         make([]float64, 0, 2000),
		pHistory:         make([]float64, 0, 2000),
		autoMode:         true,
		waveform:         WaveformSine,
		frequency:        0.5, // 0.5 Hz default
		wrdTargetLevel:   28,  // Start high for dramatic first write
		maxLogLines:      12,
		logEntries:       make([]string, 0, 12),
		lastLogPhase:     -1,
		isppCalc:         physics.NewISPPCalculator(preisach.GetEffectiveEc(), numLevels),
	}
}

// Run starts the GUI application
// Run starts the GUI with the default material (first in AllMaterials).
func Run() error {
	a := NewApp()
	return a.run()
}

// RunWithMaterial starts the GUI with a specific initial material.
// materialName should match one of the material names from AllMaterials().
// If not found, falls back to the default material.
func RunWithMaterial(materialName string) error {
	a := NewAppWithMaterial(materialName)
	return a.run()
}

// NewAppWithMaterial creates a new GUI application with a specific initial material.
func NewAppWithMaterial(materialName string) *App {
	materials := ferroelectric.AllMaterials()

	// Find the material by name
	var mat *ferroelectric.HZOMaterial
	var matIndex int
	for i, m := range materials {
		if m.Name == materialName {
			mat = m
			matIndex = i
			break
		}
	}
	// Fallback to first material if not found
	if mat == nil {
		mat = materials[0]
		matIndex = 0
	}

	numLevels := 30         // Default: FeCIM's 30 discrete analog states
	preisachGridSize := 200 // High-resolution physics simulation (independent of quantization)
	preisach := ferroelectric.NewMayergoyzPreisach(mat, preisachGridSize)

	// Refactoring: Initialize managers
	calibManager := algo.NewCalibrationManager(numLevels)
	writeController := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, calibManager)

	return &App{
		calibManager:     calibManager,
		writeController:  writeController,
		material:         mat,
		preisach:         preisach,
		materials:        materials,
		matIndex:         matIndex,
		numLevels:        numLevels,
		calibrationUp:    make([]float64, numLevels),
		calibrationDown:  make([]float64, numLevels),
		calibrationTemp:  300, // Default room temperature
		tempCalibrations: make(map[int]*TempCalibration),
		maxHistory:       2000,
		eHistory:         make([]float64, 0, 2000),
		pHistory:         make([]float64, 0, 2000),
		autoMode:         true,
		frequency:        0.5,
		paused:           false,
		// Write/Read Demo fields initialized to defaults
		wrdPhase:          0,
		wrdTargetLevel:    15,
		wrdStartLevel:     15,
		wrdBitsStored:     4.91,
		manualTargetLevel: 15,
		isppCalc:          physics.NewISPPCalculator(preisach.GetEffectiveEc(), numLevels),
	}
}

func (a *App) run() error {
	// Initialize logger here (after EnableFileLogging() has been called in main)
	if log == nil {
		log = logging.NewLogger("hysteresis")
	}

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

	// Setup custom shortcuts (Ctrl+E for export, etc.)
	a.setupShortcuts()

	// Start simulation loop
	a.running = true

	// Try to load saved calibration, or perform fresh calibration at room temp
	go func() {
		time.Sleep(100 * time.Millisecond) // Let UI settle
		a.mu.Lock()
		if !a.loadCalibration() {
			// Calibrate at default room temperature (300K)
			a.calibrateLevelsAtTemperature(300)
			if err := a.saveCalibration(); err != nil {
				log.Printf("Warning: failed to save calibration: %v", err)
			}
		}
		a.mu.Unlock()
	}()

	go a.simulationLoop()

	a.mainWindow.ShowAndRun()
	a.running = false
	return nil
}

func (a *App) createUI() fyne.CanvasObject {
	// SAFETY: No mutex needed - createUI() called from run() before simulation goroutine starts (line 319).
	// No concurrent access to a.material exists during UI initialization.

	// Create cell visualizer (THE memory cell!) - larger for better visibility
	a.cellViz = widgets.NewCellVisualizer()
	a.cellViz.SetMinSize(fyne.NewSize(180, 200)) // Increased 30% for prominence

	// Create P-E plot - will expand to fill space
	// Use temperature-corrected values for initial plot setup
	effEc := a.preisach.GetEffectiveEc()
	effPr := a.preisach.GetEffectivePr()
	a.plot = widgets.NewPEPlot(effEc*2.5, effPr*1.2, ColorBackground, ColorGrid, ColorAxis, ColorPositive, ColorNegative, ColorWarning)
	a.plot.SetMinSize(fyne.NewSize(400, 350))
	a.plot.SetMaterialParams(effEc, effPr)

	// Create level indicator (wider for better labels, clickable in Manual mode)
	a.levelIndicator = widgets.NewLevelIndicator()
	a.levelIndicator.SetMinSize(fyne.NewSize(80, 350))
	// Wire up click callback for Manual mode
	a.levelIndicator.OnLevelClicked = func(targetLevel int) {
		a.mu.Lock()
		defer a.mu.Unlock()

		// Log click with detailed state (always log for debugging)
		log.Printf("LEVEL CLICK: target=%d, currentDiscrete=%d, normalizedP=%.4f, waveform=%v, animating=%v",
			targetLevel, a.discreteLevel, a.normalizedP, a.waveform, a.manualAnimating)

		if a.waveform == WaveformManual && !a.manualAnimating {
			// Start animation to target level
			a.manualTargetLevel = targetLevel
			a.manualStartLevel = a.discreteLevel + 1 // Capture starting level (1-indexed)
			a.manualAnimating = true
			a.manualPhase = 0 // Start at RESET phase (saturate opposite direction first)
			a.manualPhaseTime = 0

			log.Printf("ANIMATION START: target=%d, start=%d, numLevels=%d",
				targetLevel, a.manualStartLevel, a.numLevels)
			a.addLogEntry(fmt.Sprintf("WRITE → Level %d", targetLevel))
		}
	}

	// Create controls panel
	controls := a.createControlsPanel()

	// Create info panel
	info := a.createInfoPanel()

	// Create log panel
	logPanel := a.createLogPanel()

	// Create simulation vs experiment comparison widget (H16)
	a.simVsExpWidget = widgets.NewSimVsExpComparison()
	// Set initial values from Preisach model
	simPr := a.preisach.GetEffectivePr()
	simEc := a.preisach.GetEffectiveEc()
	// Calculate squareness from model
	a.preisach.Reset()
	a.preisach.Update(simEc * 2) // Saturate
	psat := a.preisach.Polarization()
	a.preisach.Update(0) // Return to zero
	pr := a.preisach.Polarization()
	squareness := 0.0
	if psat > 0 {
		squareness = pr / psat
	}
	a.simVsExpWidget.SetSimulatedValues(simPr, simEc, squareness)
	a.preisach.Reset() // Reset for normal operation

	// Create ISPP visualization widget (H14)
	a.isppWidget = widgets.NewISPPVisualization()

	// Cell title with underline effect - more prominent
	cellTitle := canvas.NewText("MEMORY CELL", color.RGBA{0, 212, 255, 255})
	cellTitle.TextSize = 18 // Increased from 16
	cellTitle.TextStyle = fyne.TextStyle{Bold: true}
	cellTitle.Alignment = fyne.TextAlignCenter

	cellUnderline := canvas.NewRectangle(color.RGBA{0, 212, 255, 200}) // More opaque
	cellUnderline.SetMinSize(fyne.NewSize(140, 3))                     // Wider and thicker

	cellHeader := container.NewVBox(
		container.NewCenter(cellTitle),
		container.NewCenter(cellUnderline),
	)

	// Fixed header: Cell visualization and basic level display
	cellDisplay := container.NewVBox(
		cellHeader,
		container.NewCenter(a.cellViz),
	)

	// Scrollable content: all information panels with improved spacing
	leftScrollableContent := container.NewVBox(
		info,
		container.NewPadded(widget.NewSeparator()),
		logPanel,
		container.NewPadded(widget.NewSeparator()),
		layout.NewSpacer(),
		a.isppWidget,
		container.NewPadded(widget.NewSeparator()),
		layout.NewSpacer(),
		a.simVsExpWidget,
	)
	leftScroll := container.NewScroll(leftScrollableContent)
	leftScroll.SetMinSize(fyne.NewSize(200, 0)) // Wider for better breathing room

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
		outerSplit.SetOffset(0.20) // Left column gets 20%

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
