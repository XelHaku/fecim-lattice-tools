// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
// Uses fyne.io/fyne/v2 for cross-platform native GUI with proper graphics.
package gui

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/presets"
	sharedtheme "fecim-lattice-tools/shared/theme"
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

const defaultWriteRangeFrac = 0.98

func clampRangeFrac(frac float64) float64 {
	if frac <= 0 || frac > 1 {
		return defaultWriteRangeFrac
	}
	return frac
}

func rangeFracForMaterial(mat *ferroelectric.HZOMaterial) float64 {
	if mat == nil {
		return defaultWriteRangeFrac
	}
	if mat.TargetRangeFrac > 0 && mat.TargetRangeFrac <= 1 {
		return mat.TargetRangeFrac
	}
	return defaultWriteRangeFrac
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
	preisach  *ferroelectric.PreisachModel
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
	eHistory          []float64
	pHistory          []float64
	maxHistory        int
	lastHistorySample float64
	historyHead       int
	historySize       int

	// Full data logging (CSV)
	dataLogger *HysteresisDataLogger

	// UI state
	running          atomic.Bool
	paused           atomic.Bool
	autoMode         bool
	waveform         WaveformType
	physicsEngine    PhysicsEngine
	physicsSelect    *widget.Select
	isppMethodSelect *widget.Select
	tempContainer    *fyne.Container // Temperature controls (L-K only)
	stressContainer  *fyne.Container // Stress controls (L-K only)

	// Plot view (presentation-only)
	plotViewMode   PlotViewMode
	plotViewSelect *widget.Select

	frequency float64
	timeScale float64 // Simulation time scaling (1.0 = real-time)
	simTime   float64

	// Write-Read-Demo (wrd) fields
	// These fields drive the ISPP write-verify-read demo cycle (WaveformWriteReadDemo).
	// "wrd" is short for Write-Read-Demo — a project-internal abbreviation.
	// All wrd* state fields are protected by a.mu. UI widget fields (wrdRangeLabel,
	// wrdRangeSlider) must only be accessed from the Fyne main thread.
	wrdPhase           int     // 0=prep, 1=unused (hold-prep), 2=write, 3/4=unused, 5=display
	wrdTargetLevel     int     // Target level to write (1-30)
	wrdNextTargetLevel int     // Queued next target (applied at next PREP start)
	wrdReadLevel       int     // Level read back
	wrdPhaseTimer      float64 // Time in current phase
	wrdWriteE          float64 // E-field during write
	wrdPrepE           float64 // Pre-bias E-field (±Ec) toward target
	wrdSettleE         float64 // Settle E-field (determines final level)
	wrdStartLevel      int     // Level at start of write cycle
	wrdDebugLog        *WriteReadDebugLog
	wrdPrevP           float64 // Previous polarization for energy calculation (E·dP integration)
	wrdCycleEnergy     float64 // Energy for current write/read cycle

	// Write-Verify-Retry loop tracking (retries indefinitely until target hit)
	wrdRetryCount int // Current retry count for this target

	// Phase transition tracking for logging
	wrdResetStartP         float64 // Polarization at start of RESET phase
	wrdResetEndP           float64 // Polarization at end of RESET phase
	wrdResetEndLvl         int     // Level at end of RESET phase
	wrdWriteStartP         float64 // Polarization at start of WRITE phase
	wrdWriteEndP           float64 // Polarization at end of WRITE phase
	wrdWriteEndLvl         int     // Level at end of WRITE phase
	wrdReadStartP          float64 // Polarization at start of READ phase
	wrdLastControllerState controller.WriteState
	wrdLastControllerPulse int
	wrdLastGuardLogPulse   int // rate-limit guard log to once per pulse
	wrdLastProgressLog     float64
	wrdLastLogState        controller.WriteState // last controller state written to CSV (force-record on transition)
	wrdLastBranch          int                   // -1 lower branch, +1 upper branch, 0 unknown
	wrdForceReset          bool                  // Force PREP on next cycle (overshoot/direction change)
	wrdSkipPrep            bool                  // Skip PREP/RESET and write directly from current state
	wrdRangeFrac           float64               // Fraction of Ps used for outer level targets (0..1)
	wrdGuardFrac           float64               // Guard band fraction for MLC verify (0..0.5)
	wrdRangeTimer          *time.Timer

	// UI update throttling
	lastUIUpdate             time.Time
	lastTargetMismatchLog    time.Time
	lastWrdBoundaryLog       time.Time
	lastWrdBoundaryLogTarget int

	// Dr. Tour Demo Metrics (impressive stats!)
	wrdTotalWrites   int     // Total write operations
	wrdSuccessWrites int     // Successful writes (read matches target)
	wrdTotalEnergyfJ float64 // Total energy consumed in femtojoules
	wrdDemoMode      int     // 0=multilevel, 1=retention, 2=endurance
	wrdRetentionTime float64 // Time at E=0 during retention demo
	wrdBitsStored    float64 // log2(numLevels) bits per cell

	// Autonomous calibration (runtime)
	autoRecalibrate         bool
	recalibratePending      bool
	recalibrateReason       string
	recalibrateOvershootMax int
	recalibratePulseMax     int

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

	// ISPP (Incremental Step Pulse Programming) convergence state
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
	plot             *widgets.PEPlot
	levelIndicator   *widgets.LevelIndicator
	cellViz          *widgets.CellVisualizer
	phaseIndicator   *widgets.PhaseIndicator // State machine phase indicator (PROGRAM|VERIFY|RESULT)
	eFieldSlider     *widget.Slider
	eFieldLabel      *widget.Label
	eFieldModeLabel  *widget.Label // Shows "MANUAL" or "AUTO" for slider control mode
	eFieldRangeLabel *widget.Label // Shows slider range in MV/cm (e.g. "±2.12 MV/cm")
	tempSlider       *widget.Slider
	stressSlider     *widget.Slider
	envAccordion     *widget.AccordionItem // Environment section (disabled for Preisach)
	pLabel           *widget.Label
	levelLabel       *widget.Label
	stateLabel       *widget.Label  // State description (Negative P, Intermediate, Positive P)
	materialBtn      *widget.Button // Opens material picker, shows current material name
	waveformSelect   *widget.Select
	waveformHelp     *widget.Label
	statusLabel      *widget.Label
	pauseBtn         *widget.Button

	// Wake-up/Fatigue display labels (Dr. Tour recommendation)
	cyclesLabel     *widget.Label
	wakeupLabel     *widget.Label
	fatigueLabel    *widget.Label
	cyclePhaseLabel *widget.Label // L01: Shows current phase (WAKE-UP, STABLE, FATIGUE)

	// UI update loop (async UI updates from simulation goroutine)
	uiUpdateOnce sync.Once
	uiUpdates    chan uiSnapshot

	// Temperature-dependent metrics labels
	effEcLabel      *widget.Label
	effPrLabel      *widget.Label
	squarenessLabel *widget.Label
	switchedLabel   *widget.Label

	// Info panel material fields (updated on material change via onMaterialPickerSelected)
	materialNameLabel  *widget.Label   // prominent material name display
	materialPropsLabel *widget.Label   // Pr/Ec summary line
	materialBadgeBox   *fyne.Container // holds confidence badge (swappable on material change)

	// Levels selector
	levelsEntry    *widget.Entry
	levelsLabel    *widget.Label
	wrdRangeLabel  *widget.Label
	wrdRangeSlider *widget.Slider

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

	// Literature overlay (M1-WC-10)
	literatureDataset      *LiteratureDataset
	literatureSummaryLabel *widget.Label
	literatureMetricsLabel *widget.Label
	literatureTableLabel   *widget.Label

	// ISPP visualization widget (H14)
	isppWidget *widgets.ISPPVisualization

	// Shared ISPP calculator
	adaptiveISPP *physics.AdaptiveISPP // Phase 2 Adaptive ISPP
	lkSolver     *physics.LKSolver     // Shared L-K Solver for time-domain physics

	// LK polydomain ensemble for ISPP — disabled in GUI, used in tests only
	lkISPPPolydomainEnabled bool
	lkISPPPolydomainDomains int
	lkISPPPolydomainSeed    uint64
	lkEnsembleActive        bool

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
		return "ISPP (Write/Read)"
	case WaveformTimeResolved:
		return "Time-Resolved Switching"
	default:
		return "Unknown"
	}
}

// PhysicsEngine selects the polarization dynamics model.
type PhysicsEngine int

const (
	PhysicsLandau PhysicsEngine = iota
	PhysicsPreisach
)

func (p PhysicsEngine) String() string {
	switch p {
	case PhysicsLandau:
		return "L-K (Dynamic)"
	case PhysicsPreisach:
		return "Preisach (Quasi-Static)"
	default:
		return "Unknown"
	}
}

// newAppFromMaterial initializes an App with the provided material and shared defaults.
// Callers are responsible for setting any constructor-specific fields afterwards.
func newAppFromMaterial(mat *ferroelectric.HZOMaterial, matIndex int, materials []*ferroelectric.HZOMaterial) *App {
	numLevels := 30 // Default: FeCIM's 30 discrete analog states
	preisach := ferroelectric.NewPreisachModel(mat)

	calibManager := algo.NewCalibrationManager(numLevels)
	writeController := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, calibManager)
	writeController.MaxRetries = 10
	writeController.ForceResetLimit = 3

	a := &App{
		calibManager:            calibManager,
		writeController:         writeController,
		material:                mat,
		preisach:                preisach,
		materials:               materials,
		matIndex:                matIndex,
		numLevels:               numLevels,
		calibrationUp:           make([]float64, numLevels),
		calibrationDown:         make([]float64, numLevels),
		calibrationTemp:         300, // Default room temperature
		tempCalibrations:        make(map[int]*TempCalibration),
		maxHistory:              50000,
		eHistory:                make([]float64, 0, 2000),
		pHistory:                make([]float64, 0, 2000),
		autoMode:                true,
		physicsEngine:           PhysicsPreisach,
		frequency:               0.5,
		timeScale:               1.0,
		autoRecalibrate:         true,
		recalibrateOvershootMax: 2,
		recalibratePulseMax:     12,
		wrdRangeFrac:            rangeFracForMaterial(mat),
		wrdGuardFrac:            0.15,
		wrdBitsStored:           math.Log2(float64(numLevels)),
		lkISPPPolydomainEnabled: false,
		lkISPPPolydomainDomains: 96,
	}

	// Initialize L-K Solver and Adaptive ISPP
	a.lkSolver = physics.NewLKSolver()
	a.lkSolver.ConfigureFromMaterial(mat) // Load material-specific params (K_dep, etc.)
	// Module 1 expects deterministic LK dynamics for interactive demos and ISPP control.
	a.lkSolver.UseNLS = false
	a.lkSolver.EnableNoise = false
	a.adaptiveISPP = physics.NewAdaptiveISPP(a.lkSolver, mat)

	// Register with global preset manager
	presets.Global().RegisterProvider(NewHysteresisPresetProvider(a))

	return a
}

// NewApp creates a new GUI application with the default material selection.
func NewApp() *App {
	materials := ferroelectric.AllMaterials()
	mat, matIndex := defaultMaterialSelection(materials)

	a := newAppFromMaterial(mat, matIndex, materials)
	a.lastHistorySample = -1
	a.wrdTargetLevel = 28 // Start high for dramatic first write
	a.wrdNextTargetLevel = 0
	a.wrdSkipPrep = true
	a.waveform = WaveformSine
	a.plotViewMode = PlotViewFullHistory
	a.maxLogLines = 12
	a.logEntries = make([]string, 0, 12)
	a.lastLogPhase = -1
	return a
}

// Run starts the GUI application with the default material.
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

	var mat *ferroelectric.HZOMaterial
	var matIndex int
	for i, m := range materials {
		if m.Name == materialName {
			mat = m
			matIndex = i
			break
		}
	}
	if mat == nil {
		mat, matIndex = defaultMaterialSelection(materials)
	}

	a := newAppFromMaterial(mat, matIndex, materials)
	a.wrdPhase = 0
	a.wrdTargetLevel = 15
	a.wrdNextTargetLevel = 0
	a.wrdStartLevel = 15
	a.manualTargetLevel = 15
	return a
}

func (a *App) run() error {
	// Initialize logger here (after EnableFileLogging() has been called in main)
	if log == nil {
		log = logging.NewLogger("hysteresis")
	}
	a.startDataLogger()
	defer a.stopDataLogger()

	a.fyneApp = app.New()
	a.fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

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
	a.running.Store(true)
	stopCh := make(chan struct{})
	var simWG sync.WaitGroup

	// Calibration goroutine: runs once at startup; exits after first calibration completes.
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

	// Simulation goroutine: runs until a.running.Store(false) on window close.
	simWG.Add(1)
	go func() {
		defer simWG.Done()
		a.simulationLoop()
	}()

	a.mainWindow.ShowAndRun()
	a.running.Store(false)
	close(stopCh)
	simWG.Wait()
	return nil
}
