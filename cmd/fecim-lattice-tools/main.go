// Command fecim-lattice-tools provides a unified GUI application with all FeCIM demos.
//
// This is the main entry point for FeCIM Lattice Tools.
// It combines all 6 demos into a single application with tab navigation.
//
// The 6-Demo Story:
//
//	Demo 1: The Memory Cell (Hysteresis) - How the cell works
//	Demo 2: The Crossbar Computer (MVM + Non-Idealities) - How we compute
//	Demo 3: The AI Brain (MNIST) - What we can build
//	Demo 4: The Chip System (Circuits) - How it fits in a chip
//	Demo 5: Why FeCIM Wins (Comparison) - The business case
//	Demo 6: EDA Design Suite - Bridge to open-source EDA tools
package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
	"fecim-lattice-tools/shared/accessibility"
	"fecim-lattice-tools/shared/help"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/recentfiles"
	"fecim-lattice-tools/shared/recording"
	"fecim-lattice-tools/shared/themes"
	"fecim-lattice-tools/shared/undo"
	"fecim-lattice-tools/shared/utils"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Global undo manager for parameter changes across all modules
var undoManager *undo.Manager

// Global logger for the main application
var log *logging.Logger

var (
	screenshotOutputDir = "screenshots"
	recordingOutputDir  = "recordings"
)

func isVerbosityToken(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "0", "off", "none", "1", "info", "2", "debug", "3", "trace", "all":
		return true
	default:
		return false
	}
}

// ForceMinSizeLayout ignores the child's MinSize and returns a fixed small size.
// This prevents Fyne from requesting window resizes based on content size,
// which causes oscillation loops on tiling window managers (Wayland/Sway).
type ForceMinSizeLayout struct {
	Min fyne.Size
}

func (l *ForceMinSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return l.Min
}

func (l *ForceMinSizeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

// RecordingState manages FFmpeg video recording using canvas capture
type RecordingState struct {
	mu          sync.Mutex
	isRecording bool
	stopped     bool // Prevents double-close of stopChan
	stopChan    chan struct{}
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	outputFile  string
	window      fyne.Window
}

func newRecordingState() *RecordingState {
	return &RecordingState{}
}

// sectionNameFromTab converts view name to a filename-friendly section name
func sectionNameFromTab(viewName string) string {
	switch viewName {
	case "Home":
		return "home"
	case "FeCIM Hysteresis Simulation":
		return "module01-hysteresis"
	case "FeCIM Crossbar Array Visualization":
		return "module02-crossbar"
	case "FeCIM MNIST Neural Network":
		return "module03-mnist"
	case "FeCIM Peripheral Circuits Visualizer":
		return "module04-circuits"
	case "FeCIM: The Energy Revolution":
		return "module05-comparison"
	case "FeCIM EDA Design Suite (Work In Progress)":
		return "module06-eda"
	default:
		return "unknown"
	}
}

// takeScreenshot captures the current window and saves it as PNG with metadata
// L02: Embeds PNG metadata (title, timestamp, module name, simulation parameters)
func takeScreenshot(window fyne.Window, sectionName string) string {
	// Create screenshots directory if it doesn't exist
	screenshotDir := screenshotOutputDir
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		fmt.Println("Error creating screenshots directory:", err)
		return ""
	}

	// Generate filename with timestamp and section name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(screenshotDir, fmt.Sprintf("fecim_%s_%s.png", sectionName, timestamp))

	// Capture the canvas
	img := window.Canvas().Capture()

	// L02: Create metadata for the screenshot
	meta := utils.DefaultScreenshotMetadata(sectionName)
	meta.CustomData["Window-Size"] = fmt.Sprintf("%dx%d", int(window.Canvas().Size().Width), int(window.Canvas().Size().Height))

	// L02: Save with embedded metadata
	if err := utils.SavePNGWithMetadata(img, filename, meta); err != nil {
		fmt.Println("Error saving screenshot with metadata:", err)
		return ""
	}

	fmt.Println("Screenshot saved:", filename)
	return filename
}

// startRecording begins FFmpeg video recording using Fyne canvas capture
func (rs *RecordingState) startRecording(window fyne.Window) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.isRecording {
		return fmt.Errorf("already recording")
	}

	// Create recordings directory if it doesn't exist
	recordingDir := recordingOutputDir
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return fmt.Errorf("error creating recordings directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	rs.outputFile = filepath.Join(recordingDir, fmt.Sprintf("fecim_recording_%s.mp4", timestamp))
	rs.window = window
	rs.stopChan = make(chan struct{})
	rs.stopped = false // Reset stopped flag for new recording

	// Get initial frame to determine size
	img := window.Canvas().Capture()
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Ensure dimensions are even (required by H.264)
	if width%2 != 0 {
		width--
	}
	if height%2 != 0 {
		height--
	}

	// Get audio device for recording - try mic first, fallback to desktop audio
	audioDevice := ""
	isMicrophone := false
	if recording.IsAudioAvailable() {
		// Try to find a microphone (Bluetooth or wired)
		if micDevice, err := recording.GetMicrophoneDevice(); err == nil {
			audioDevice = micDevice.Name
			isMicrophone = true
			log.Info("Recording with microphone: %s", audioDevice)
		} else {
			// Fallback to desktop audio (monitor source)
			if desktopDevice, err := recording.GetDesktopAudioDevice(); err == nil {
				audioDevice = desktopDevice.Name
				log.Info("Recording desktop audio: %s", audioDevice)
			}
		}
	}

	// FFmpeg command to receive raw RGB frames from stdin + audio from PulseAudio
	var ffmpegArgs []string
	if audioDevice != "" {
		// Base args for video + audio input
		ffmpegArgs = []string{
			"-y",             // Overwrite output file
			"-f", "rawvideo", // Raw video input
			"-pixel_format", "rgb24", // RGB24 format
			"-video_size", fmt.Sprintf("%dx%d", width, height), // Frame size
			"-framerate", "20", // 20 FPS
			"-i", "-", // Video from stdin
			"-f", "pulse", // Audio from PulseAudio
			"-i", audioDevice, // Audio device
			"-c:v", "libx264", // H.264 codec
			"-preset", "ultrafast", // Fast encoding
			"-crf", "23", // Quality
			"-pix_fmt", "yuv420p", // Output pixel format
		}

		// Add audio filter based on source type
		if isMicrophone {
			// Voice enhancement with deeper tone:
			// 1. highpass=f=80 - Allow more bass through
			// 2. lowpass=f=7500 - Cut high hiss
			// 3. afftdn - Noise reduction
			// 4. agate - Noise gate
			// 5. rubberband - Pitch down for deeper voice
			// 6. equalizer - Boost bass frequencies for fuller sound
			// 7. acompressor - Even out levels
			// 8. alimiter - Prevent clipping
			ffmpegArgs = append(ffmpegArgs, "-af",
				"highpass=f=80,lowpass=f=7500,afftdn=nr=80:nf=-25:tn=1,agate=threshold=-35dB:ratio=3:attack=10:release=150,rubberband=pitch=0.92,equalizer=f=200:t=h:w=200:g=4,equalizer=f=100:t=h:w=100:g=3,acompressor=threshold=-20dB:ratio=3:attack=10:release=150:makeup=5dB,alimiter=limit=0.95")
		}

		// Common audio encoding args
		ffmpegArgs = append(ffmpegArgs,
			"-c:a", "aac", // AAC audio codec
			"-b:a", "192k", // Higher bitrate for better quality
			"-ac", "2", // Stereo
			"-shortest", // Stop when shortest input ends
			rs.outputFile,
		)
	} else {
		// Without audio: video only
		log.Info("Recording without audio (no audio device available)")
		ffmpegArgs = []string{
			"-y",             // Overwrite output file
			"-f", "rawvideo", // Raw video input
			"-pixel_format", "rgb24", // RGB24 format
			"-video_size", fmt.Sprintf("%dx%d", width, height), // Frame size
			"-framerate", "20", // 20 FPS
			"-i", "-", // Read from stdin
			"-c:v", "libx264", // H.264 codec
			"-preset", "ultrafast", // Fast encoding
			"-crf", "23", // Quality
			"-pix_fmt", "yuv420p", // Output pixel format
			rs.outputFile,
		}
	}
	rs.cmd = exec.Command("ffmpeg", ffmpegArgs...)

	// Get stdin pipe
	var err error
	rs.stdin, err = rs.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error getting stdin pipe: %w", err)
	}

	// Start FFmpeg
	if err := rs.cmd.Start(); err != nil {
		return fmt.Errorf("error starting FFmpeg: %w", err)
	}

	rs.isRecording = true

	// Start capture goroutine with panic recovery
	utils.SafeGo("captureLoop", func() {
		rs.captureLoop(width, height)
	})

	fmt.Println("Recording started:", rs.outputFile)
	return nil
}

// captureLoop continuously captures frames and sends to FFmpeg
func (rs *RecordingState) captureLoop(width, height int) {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 FPS
	defer ticker.Stop()

	frameBuffer := make([]byte, width*height*3) // RGB24

	for {
		select {
		case <-rs.stopChan:
			return
		case <-ticker.C:
			// Capture frame on main thread
			var img image.Image
			done := make(chan struct{})
			fyne.Do(func() {
				img = rs.window.Canvas().Capture()
				close(done)
			})
			<-done

			if img == nil {
				continue
			}

			// Convert to RGB24 and write to FFmpeg
			bounds := img.Bounds()
			idx := 0
			for y := bounds.Min.Y; y < bounds.Min.Y+height && y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Min.X+width && x < bounds.Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					frameBuffer[idx] = byte(r >> 8)
					frameBuffer[idx+1] = byte(g >> 8)
					frameBuffer[idx+2] = byte(b >> 8)
					idx += 3
				}
			}

			// Write frame to FFmpeg without holding the lock during IO.
			rs.mu.Lock()
			stdin := rs.stdin
			recording := rs.isRecording
			rs.mu.Unlock()

			if stdin != nil && recording {
				_, _ = stdin.Write(frameBuffer)
			}
		}
	}
}

// stopRecording stops the FFmpeg recording
func (rs *RecordingState) stopRecording() (string, error) {
	rs.mu.Lock()
	if !rs.isRecording {
		rs.mu.Unlock()
		return "", fmt.Errorf("not recording")
	}

	stopChan := rs.stopChan
	shouldCloseStop := !rs.stopped && stopChan != nil
	rs.stopped = true
	rs.stopChan = nil

	stdin := rs.stdin
	rs.stdin = nil

	cmd := rs.cmd
	rs.cmd = nil

	outputFile := rs.outputFile
	rs.isRecording = false
	rs.mu.Unlock()

	// Signal capture loop to stop.
	if shouldCloseStop {
		close(stopChan)
	}

	// Close stdin to signal EOF to FFmpeg.
	if stdin != nil {
		_ = stdin.Close()
	}

	// Wait for FFmpeg to finish, but don't block the UI thread indefinitely.
	if cmd != nil {
		waitDone := make(chan error, 1)
		go func() {
			waitDone <- cmd.Wait()
		}()

		select {
		case <-waitDone:
		case <-time.After(2 * time.Second):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			select {
			case <-waitDone:
			case <-time.After(1 * time.Second):
			}
		}
	}

	fmt.Println("Recording stopped:", outputFile)
	return outputFile, nil
}

// IsRecording returns current recording state
func (rs *RecordingState) IsRecording() bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.isRecording
}

// DemoApp holds the demo instances
type DemoApp struct {
	demo1 *demo1gui.EmbeddedApp           // Hysteresis
	demo2 *demo2gui.EmbeddedCrossbarApp   // Crossbar (original single-view)
	demo3 *demo3gui.EmbeddedDualModeApp   // MNIST FP vs CIM (full-featured)
	demo4 *demo4gui.EmbeddedCircuitsApp   // Circuits
	demo5 *demo5gui.EmbeddedComparisonApp // Comparison (technical briefing)
	demo6 *demo6gui.EmbeddedEDAApp        // EDA Design Suite
	demo7 *demo7gui.EmbeddedDocsApp       // Documentation
}

// Preference keys for window state persistence
const (
	prefKeyWindowWidth  = "window_width"
	prefKeyWindowHeight = "window_height"
	prefKeyLastTab      = "last_tab"

	// Default window dimensions
	defaultWindowWidth  = 1400
	defaultWindowHeight = 900

	// Minimum supported GUI size (G13)
	minWindowWidth  = 1024
	minWindowHeight = 768
)

// loadWindowSize loads saved window dimensions from preferences
func loadWindowSize(prefs fyne.Preferences) fyne.Size {
	w := prefs.FloatWithFallback(prefKeyWindowWidth, defaultWindowWidth)
	h := prefs.FloatWithFallback(prefKeyWindowHeight, defaultWindowHeight)

	// Clamp to minimum supported GUI size (G13)
	if w < minWindowWidth {
		w = minWindowWidth
	}
	if h < minWindowHeight {
		h = minWindowHeight
	}

	// Round to even integers and add 2-pixel buffer to prevent Fyne/Wayland
	// resize oscillation caused by MinSize floating-point rounding
	w = float64(int(w+1) / 2 * 2)
	h = float64(int(h+1) / 2 * 2)

	return fyne.NewSize(float32(w), float32(h))
}

// saveWindowSize saves current window dimensions to preferences
func saveWindowSize(prefs fyne.Preferences, size fyne.Size) {
	// Round to even integers to prevent floating-point drift
	w := float64(int(size.Width+1) / 2 * 2)
	h := float64(int(size.Height+1) / 2 * 2)
	prefs.SetFloat(prefKeyWindowWidth, w)
	prefs.SetFloat(prefKeyWindowHeight, h)
}

// saveLastTab saves the currently selected tab index
func saveLastTab(prefs fyne.Preferences, tabIndex int) {
	prefs.SetInt(prefKeyLastTab, tabIndex)
}

// loadLastTab loads the last selected tab index
func loadLastTab(prefs fyne.Preferences) int {
	return prefs.IntWithFallback(prefKeyLastTab, 0)
}

func main() {
	if maybeDispatchSubcommand(os.Args[1:]) {
		return
	}

	// Parse command-line flags
	loggerFlag := flag.Bool("logger", false, "Enable file logging (logs to logs/ directory). Optional shorthand: --logger debug|info|trace|off")
	verbosityFlag := flag.String("verbosity", "info", "Logging verbosity: 0|off, 1|info, 2|debug, 3|trace (only used with --logger)")
	calibrateFlag := flag.Bool("calibrate", false, "Run hysteresis calibration and exit (no GUI)")
	materialFlag := flag.String("material", "all", "Material to calibrate (use 'all' for all materials, or specify name)")
	forceFlag := flag.Bool("force", false, "Force recalibration even if calibration file exists")
	verifyFlag := flag.Bool("verify", false, "Verify calibration accuracy after calibrating")
	listMaterialsFlag := flag.Bool("list-materials", false, "List available materials and exit")
	modeFlag := flag.String("mode", "", "Run a headless mode (e.g., hysteresis) and exit")
	engineFlag := flag.String("engine", "", "Headless hysteresis engine for --mode hysteresis: preisach|lk (default: preisach)")
	screenshotDirFlag := flag.String("screenshot-dir", screenshotOutputDir, "Output directory for GUI screenshots")
	recordingDirFlag := flag.String("recording-dir", recordingOutputDir, "Output directory for GUI recordings")
	var moduleFlag = flag.String("module", "home", "Start module: home, hysteresis, crossbar, mnist, circuits, comparison, eda, docs")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "FeCIM Lattice Tools")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools [flags]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Flags:")
		flag.PrintDefaults()
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Headless hysteresis environment variables (--mode hysteresis):")
		fmt.Fprintln(out, "  FECIM_MATERIAL              Material preset (fecim_hzo, literature_superlattice, default_hzo, etc.)")
		fmt.Fprintln(out, "  FECIM_RANGE_FRAC            Reachable polarization fraction (0,1], clamped for LK reachability")
		fmt.Fprintln(out, "  FECIM_ISPP_STEPS_PER_PULSE  Target simulation steps per pulse (integer)")
		fmt.Fprintln(out, "  FECIM_HEADLESS_FAST         Fast preset for CI (1 enables shorter LK budgets)")
		fmt.Fprintln(out, "  FECIM_ISPP_TARGETS          Number of randomized ISPP targets")
		fmt.Fprintln(out, "  FECIM_ISPP_TARGET_SEED      Deterministic seed for randomized ISPP target fill")
		fmt.Fprintln(out, "  FECIM_ISPP_TARGET_LEVELS    Explicit ISPP target levels (e.g., lo,mid,hi or numeric list)")
		fmt.Fprintln(out, "  FECIM_ISPP_MAX_PULSES       Per-target pulse budget override")
		fmt.Fprintln(out, "  FECIM_HEADLESS_ALLOW_TIMEOUT Continue after per-target timeout when set to 1")
		fmt.Fprintln(out, "  FECIM_PPROF                 Enable headless pprof endpoint when set to 1")
		fmt.Fprintln(out, "  FECIM_PPROF_ADDR            Optional pprof listen address (default 127.0.0.1:6060)")
	}
	flag.Parse()

	verbosityProvided := false
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if f.Name == "verbosity" {
			verbosityProvided = true
		}
	})
	if dir := strings.TrimSpace(*screenshotDirFlag); dir != "" {
		screenshotOutputDir = filepath.Clean(dir)
	}
	if dir := strings.TrimSpace(*recordingDirFlag); dir != "" {
		recordingOutputDir = filepath.Clean(dir)
	}

	if *loggerFlag && !verbosityProvided {
		if args := flag.Args(); len(args) > 0 && isVerbosityToken(args[0]) {
			*verbosityFlag = args[0]
		}
	}

	// Handle --list-materials
	if *listMaterialsFlag {
		fmt.Println("Available materials:")
		for _, name := range demo1gui.ListMaterials() {
			fmt.Printf("  - %s\n", name)
		}
		return
	}

	// Handle --calibrate (CLI mode, no GUI)
	if *calibrateFlag {
		opts := demo1gui.CLICalibrationOptions{
			MaterialName: *materialFlag,
			NumLevels:    0, // 0 = use material's native level count
			Temperature:  300,
			Force:        *forceFlag,
			Verbose:      true,
			Verify:       *verifyFlag,
		}
		if err := demo1gui.RunCLICalibration(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Calibration error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Calibration complete.")
		return
	}

	// Initialize logger only if --logger flag is set
	if *loggerFlag {
		logging.EnableFileLogging() // Must be called before NewLogger to enable file output
		verbosity := logging.ParseVerbosityFlag(*verbosityFlag)
		logging.SetVerbosity(verbosity)
		log = logging.NewLogger("fecim-lattice-tools")
		defer log.Close()
		log.Info("FeCIM Visualizer starting with verbosity=%s", logging.VerbosityString(verbosity))
	} else {
		// File logging is disabled by default, no action needed
		log = logging.NewNoOpLogger()
	}

	// Module 4 compute logging (JSON) is opt-in; tie to --logger for convenience.
	if *loggerFlag {
		demo4gui.EnableComputeLog(true)
	}

	// Handle --mode (headless diagnostics) after logging is initialized
	if *modeFlag != "" {
		if err := runMode(*modeFlag, *engineFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Mode error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create Fyne app
	fmt.Println("[STARTUP] Creating Fyne app...")
	fyneApp := app.NewWithID("com.fecim.visualizer")

	// Initialize theme manager and load saved preference
	fmt.Println("[STARTUP] Initializing theme manager...")
	themeManager := themes.NewManager(fyneApp)
	themeManager.LoadPreference()
	themeManager.ApplyAccessibilityPreferences()
	fmt.Printf("[STARTUP] Theme loaded: %s\n", themeManager.CurrentTheme())

	// Load saved window size from preferences
	prefs := fyneApp.Preferences()
	savedSize := loadWindowSize(prefs)
	log.Debug("Loaded window size from preferences: %.0fx%.0f", savedSize.Width, savedSize.Height)
	fmt.Printf("[STARTUP] Loaded window size: %.0fx%.0f\n", savedSize.Width, savedSize.Height)

	// Create main window with saved size
	fmt.Println("[STARTUP] Creating window...")
	window := fyneApp.NewWindow("FeCIM Lattice Tools")
	window.Resize(savedSize)
	fmt.Println("[STARTUP] Window created")

	// Create notification manager for toast notifications
	fmt.Println("[STARTUP] Creating notification manager...")
	notificationManager := sharedwidgets.NewNotificationManager(window, 100)
	toastContainer := sharedwidgets.NewToastContainer(5)

	// Create undo manager for parameter changes
	fmt.Println("[STARTUP] Creating undo manager...")
	undoManager = undo.NewManager(100) // Keep 100 actions in history
	undo.SetGlobalManager(undoManager) // Register globally so modules can access it

	// Initialize recent files manager
	fmt.Println("[STARTUP] Creating recent files manager...")
	recentFilesManager := recentfiles.InitGlobal(prefs)
	// Clean up any files that no longer exist
	removed := recentFilesManager.CleanupMissing()
	if removed > 0 {
		log.Debug("Cleaned up %d missing recent files", removed)
	}
	fmt.Println("[STARTUP] Recent files manager created")

	// Wire up toast display callbacks
	notificationManager.SetToastCallbacks(
		func(toast *sharedwidgets.Toast) {
			toastContainer.Add(toast)
		},
		func(toast *sharedwidgets.Toast) {
			toastContainer.Remove(toast)
		},
	)

	// Set global notification manager for demos to use
	sharedwidgets.SetGlobalNotificationManager(notificationManager)
	fmt.Println("[STARTUP] Notification manager created")

	// Create demo instances with error handling
	demo2, err := demo2gui.NewEmbeddedCrossbarApp()
	if err != nil {
		log.Printf("[ERROR] Failed to create crossbar demo: %v", err)
		// Defer notification until after UI is ready
		utils.SafeGo("crossbar-error-notification", func() {
			time.Sleep(1 * time.Second)
			notificationManager.Error("Crossbar Demo Error", "Failed to initialize crossbar demo. Check logs for details.")
		})
		// Create a placeholder - the demo tab will show an error
		demo2 = nil
	}

	fmt.Println("[STARTUP] Creating demo1 (hysteresis)...")
	d1 := demo1gui.NewEmbeddedApp()
	fmt.Println("[STARTUP] Creating demo3 (MNIST)...")
	d3 := demo3gui.NewEmbeddedDualModeApp()
	fmt.Println("[STARTUP] Creating demo4 (circuits)...")
	d4 := demo4gui.NewEmbeddedCircuitsApp()
	fmt.Println("[STARTUP] Creating demo5 (comparison)...")
	d5 := demo5gui.NewEmbeddedComparisonApp()
	fmt.Println("[STARTUP] Creating demo6 (EDA)...")
	d6 := demo6gui.NewEmbeddedEDAApp()
	fmt.Println("[STARTUP] Creating demo7 (Docs)...")
	d7 := demo7gui.NewEmbeddedDocsApp()
	fmt.Println("[STARTUP] All demos created")

	demos := &DemoApp{
		demo1: d1,
		demo2: demo2,
		demo3: d3,
		demo4: d4,
		demo5: d5,
		demo6: d6,
		demo7: d7,
	}

	// View names for navigation (index matches view index)
	viewNames := []string{
		"Home",
		"FeCIM Hysteresis Simulation",
		"FeCIM Crossbar Array Visualization",
		"FeCIM MNIST Neural Network",
		"FeCIM Peripheral Circuits Visualizer",
		"FeCIM: The Energy Revolution",
		"FeCIM EDA Design Suite (Work In Progress)",
		"Documentation",
	}

	// Track current view index and content stack
	currentViewIndex := 0
	var contentStack *fyne.Container
	var views []fyne.CanvasObject
	var onViewChange func(index int) // Callback for view changes

	// selectView switches to the specified view index
	selectView := func(index int) {
		if index < 0 || index >= len(views) {
			return
		}
		// Hide all views
		for _, v := range views {
			v.Hide()
		}
		// Show selected view
		views[index].Show()
		currentViewIndex = index
		// Trigger view change callback
		if onViewChange != nil {
			onViewChange(index)
		}
		if contentStack != nil {
			contentStack.Refresh()
		}
	}

	// Create launcher content with callback to switch views
	launcherContent := CreateLauncherContent(func(demoNum int) {
		// Map demo number to view index (demoNum 1-6 maps to index 1-6)
		if demoNum >= 1 && demoNum <= 6 {
			selectView(demoNum)
		}
	})

	// Build content for each demo
	fmt.Println("[STARTUP] Building demo1 content...")
	demo1Content := demos.demo1.BuildContent(fyneApp, window)
	fmt.Println("[STARTUP] demo1 content built")

	// Handle potentially nil demo2 (initialization error)
	var demo2Content fyne.CanvasObject
	if demos.demo2 != nil {
		fmt.Println("[STARTUP] Building demo2 content...")
		demo2Content = demos.demo2.BuildContent(fyneApp, window)
		fmt.Println("[STARTUP] demo2 content built")
	} else {
		demo2Content = container.NewCenter(widget.NewLabel("Crossbar demo failed to initialize. Check logs for details."))
	}

	fmt.Println("[STARTUP] Building demo3 content...")
	demo3Content := demos.demo3.BuildContent(fyneApp, window)
	fmt.Println("[STARTUP] demo3 content built")
	fmt.Println("[STARTUP] Building demo4 content...")
	demo4Content := demos.demo4.BuildContent(fyneApp, window)
	fmt.Println("[STARTUP] demo4 content built")
	fmt.Println("[STARTUP] Building demo5 content...")
	demo5Content := demos.demo5.BuildContent(fyneApp, window)
	fmt.Println("[STARTUP] demo5 content built")
	fmt.Println("[STARTUP] Building demo6 content...")
	demo6Content := demos.demo6.BuildContent(fyneApp, window)
	fmt.Println("[STARTUP] demo6 content built")
	fmt.Println("[STARTUP] Building demo7 content...")
	demo7Content := demos.demo7.BuildContent(fyneApp, window)
	fmt.Println("[STARTUP] demo7 content built")

	// Create views - 7 demos total (plus home)
	fmt.Println("[STARTUP] Creating views...")
	views = []fyne.CanvasObject{
		launcherContent,
		container.NewMax(demo1Content),
		container.NewMax(demo2Content),
		container.NewMax(demo3Content),
		container.NewMax(demo4Content),
		container.NewMax(demo5Content),
		container.NewMax(demo6Content),
		container.NewMax(demo7Content),
	}
	// Hide all views except Home initially
	for i, v := range views {
		if i != 0 {
			v.Hide()
		}
	}
	// Create stack container with all views
	contentStack = container.NewStack(views...)
	fmt.Println("[STARTUP] Views created")
	fmt.Println("[STARTUP] Creating recording state...")

	// Create recording state
	recordingState := newRecordingState()

	// Create mic level widget (shows audio levels even when not recording)
	fmt.Println("[STARTUP] Creating mic level widget...")
	micLevelWidget := sharedwidgets.NewMicLevel()
	var audioMonitor *recording.AudioMonitor
	var audioMonitorMu sync.Mutex
	var micController *sharedwidgets.MicLevelController

	// Defer audio initialization to background - IsAudioAvailable() can take 7+ seconds
	fmt.Println("[STARTUP] Deferring audio initialization to background...")

	fmt.Println("[STARTUP] Creating buttons...")

	// Create screenshot button
	screenshotBtn := widget.NewButtonWithIcon("Screenshot", theme.MediaPhotoIcon(), func() {
		log.Button("Screenshot")
		// Get current section name from selected view
		sectionName := sectionNameFromTab(viewNames[currentViewIndex])
		filename := takeScreenshot(window, sectionName)
		if filename != "" {
			log.Debug("Screenshot saved: %s", filename)
			// Save silently to avoid intrusive UI interruptions during demos.
		}
	})

	// Create record button with toggle functionality
	// Use a separate label for recording time to prevent button resize
	recordBtn := widget.NewButtonWithIcon("Record", theme.MediaRecordIcon(), nil)
	recordTimeLabel := widget.NewLabel("")
	recordTimeLabel.TextStyle = fyne.TextStyle{Monospace: true}
	recordTimeLabel.Hide() // Hidden until recording starts
	var recordingTimerStop chan struct{}
	recordBtn.OnTapped = func() {
		log.Button("Record")
		if recordingState.IsRecording() {
			log.Debug("Stopping recording...")
			// Stop the timer goroutine
			if recordingTimerStop != nil {
				close(recordingTimerStop)
				recordingTimerStop = nil
			}
			// Stop recording
			outputFile, err := recordingState.stopRecording()
			if err != nil {
				log.Debug("Error stopping recording: %v", err)
				fmt.Println("Error stopping recording:", err)
				return
			}
			log.Debug("Recording saved: %s", outputFile)
			recordBtn.SetText("Record")
			recordBtn.SetIcon(theme.MediaRecordIcon())
			recordTimeLabel.Hide()
			// Update mic indicator
			micLevelWidget.SetRecording(false)
			// Show toast notification
			notificationManager.Success("Recording Saved", outputFile)
		} else {
			log.Debug("Starting recording...")

			// Lazy-start audio monitoring on first record (saves startup time)
			audioMonitorMu.Lock()
			if micController != nil && audioMonitor != nil && !audioMonitor.IsRunning() {
				log.Debug("Starting audio monitoring (lazy init on first record)")
				if err := micController.Start(); err != nil {
					log.Debug("Failed to start audio monitoring: %v", err)
				} else {
					log.Info("Audio level monitoring started")
				}
			}
			audioMonitorMu.Unlock()

			// Start recording
			if err := recordingState.startRecording(window); err != nil {
				fmt.Println("Error starting recording:", err)
				// Show error toast notification
				notificationManager.Error("Recording Error", err.Error())
				return
			}
			recordBtn.SetText("Stop")
			recordBtn.SetIcon(theme.MediaStopIcon())
			recordTimeLabel.Show()
			// Update mic indicator to show recording state (red dot)
			micLevelWidget.SetRecording(true)
			// Show info notification that recording started
			notificationManager.Info("Recording Started", "Recording video with audio...")
			// Start timer to show real-time datetime with milliseconds and take screenshots
			recordingTimerStop = make(chan struct{})
			utils.SafeGo("recording-timer", func() {
				displayTicker := time.NewTicker(50 * time.Millisecond) // Update display ~20 times per second
				screenshotTicker := time.NewTicker(5 * time.Second)    // Screenshot every 5 seconds
				defer displayTicker.Stop()
				defer screenshotTicker.Stop()
				for {
					select {
					case <-recordingTimerStop:
						return
					case <-displayTicker.C:
						fyne.Do(func() {
							now := time.Now()
							recordTimeLabel.SetText(now.Format("15:04:05"))
						})
					case <-screenshotTicker.C:
						// Take screenshot on main thread
						fyne.Do(func() {
							sectionName := sectionNameFromTab(viewNames[currentViewIndex])
							takeScreenshot(window, sectionName)
						})
					}
				}
			})
		}
	}

	// Create close button (triggers close intercept for proper state saving)
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		log.Button("Close")
		// Trigger the close intercept which handles state saving
		window.Close()
	})

	// Create home button to navigate to Home view
	homeBtn := widget.NewButtonWithIcon("", theme.HomeIcon(), func() {
		log.Button("Home")
		selectView(0)
	})

	// Create current module label (left-justified in toolbar)
	currentModuleLabel := widget.NewLabel("Home")
	currentModuleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Track current demo for start/stop
	currentDemo := 0

	// Map view index to help context key
	viewContextKeys := []string{
		"home",       // 0 - Home
		"hysteresis", // 1 - Hysteresis
		"crossbar",   // 2 - Crossbar
		"mnist",      // 3 - MNIST
		"circuits",   // 4 - Circuits
		"comparison", // 5 - Comparison
		"eda",        // 6 - EDA
		"docs",       // 7 - Documentation
	}

	// Handle view changes - start/stop simulations as needed
	onViewChange = func(index int) {
		viewName := viewNames[index]
		log.TabChange(viewName)
		sharedwidgets.DebugInteraction(fmt.Sprintf("View changed to '%s'", viewName))

		// Update current module label
		currentModuleLabel.SetText(viewName)

		// Update help system context for F1 contextual help
		if index < len(viewContextKeys) {
			help.GetHelpSystem().SetContext(viewContextKeys[index])
		}

		// Stop previous demo
		switch currentDemo {
		case 1:
			log.Debug("Stopping demo1 (Hysteresis)")
			demos.demo1.Stop()
		case 2:
			log.Debug("Stopping demo2 (Crossbar+)")
			if demos.demo2 != nil {
				demos.demo2.Stop()
			}
		case 3:
			log.Debug("Stopping demo3 (MNIST)")
			demos.demo3.Stop()
		case 4:
			log.Debug("Stopping demo4 (Circuits)")
			demos.demo4.Stop()
		case 5:
			log.Debug("Stopping demo5 (Comparison)")
			demos.demo5.Stop()
		case 6:
			log.Debug("Stopping demo6 (EDA)")
			demos.demo6.Stop()
		case 7:
			log.Debug("Stopping demo7 (Docs)")
			demos.demo7.Stop()
		}

		// Start new demo (index 0=Home, 1-7=demos)
		currentDemo = index
		switch index {
		case 1:
			log.Debug("Starting demo1 (Hysteresis)")
			demos.demo1.Start()
		case 2:
			log.Debug("Starting demo2 (Crossbar+)")
			if demos.demo2 != nil {
				demos.demo2.Start()
			}
		case 3:
			log.Debug("Starting demo3 (MNIST)")
			demos.demo3.Start()
		case 4:
			log.Debug("Starting demo4 (Circuits)")
			demos.demo4.Start()
		case 5:
			log.Debug("Starting demo5 (Comparison)")
			demos.demo5.Start()
		case 6:
			log.Debug("Starting demo6 (EDA)")
			demos.demo6.Start()
		case 7:
			log.Debug("Starting demo7 (Docs)")
			demos.demo7.Start()
		}
	}

	// Create docs button to navigate to Documentation view
	docsBtn := widget.NewButtonWithIcon("", theme.HelpIcon(), func() {
		log.Button("Docs")
		selectView(7) // Documentation is view index 7
	})

	// Create theme toggle button (cycles through dark/light/high-contrast)
	themeToggleBtn := themes.CreateQuickToggle(themeManager)

	// Create notification history panel (initially hidden)
	fmt.Println("[STARTUP] Creating notification history panel...")
	notificationHistoryPanel := sharedwidgets.NewNotificationHistoryPanel(notificationManager)
	notificationHistoryPanel.Hide()

	// Create notification button with unread badge
	notificationBtn := sharedwidgets.NewNotificationButton(notificationManager, func() {
		log.Button("Notifications")
		notificationHistoryPanel.Toggle()
		if notificationHistoryPanel.IsVisible() {
			notificationManager.MarkAllRead()
		}
	})

	// Create undo/redo toolbar widget
	fmt.Println("[STARTUP] Creating undo toolbar...")
	undoToolbar := sharedwidgets.NewUndoToolbar(undoManager)

	// Add keyboard shortcuts for undo/redo (Ctrl+Z / Ctrl+Y / Ctrl+Shift+Z)
	canvas := window.Canvas()
	canvas.AddShortcut(&fyne.ShortcutUndo{}, func(_ fyne.Shortcut) {
		if undoManager.Undo() {
			log.Debug("Undo via keyboard shortcut")
			sharedwidgets.DebugInteraction("Undo (Ctrl+Z)")
		}
	})
	canvas.AddShortcut(&fyne.ShortcutRedo{}, func(_ fyne.Shortcut) {
		if undoManager.Redo() {
			log.Debug("Redo via keyboard shortcut")
			sharedwidgets.DebugInteraction("Redo (Ctrl+Y)")
		}
	})

	// Initialize help system
	fmt.Println("[STARTUP] Initializing help system...")
	helpSystem := help.GetHelpSystem()
	helpSystem.Init(window)

	// Create help browser
	helpBrowser := help.NewHelpBrowser(helpSystem, window)

	// Set up help callbacks
	helpSystem.SetCallbacks(
		func(topic *help.HelpTopic) {
			// Show contextual help popup
			help.NewContextualHelpDialog(topic, window).Show()
		},
		func() {
			// Show full help browser
			helpBrowser.Show()
		},
	)

	// Set up F1 keyboard shortcut for contextual help
	help.SetupF1Shortcut(window, helpSystem)

	// Create tip manager for startup tips
	tipManager := help.NewTipManager(fyneApp.Preferences(), window)
	fmt.Println("[STARTUP] Help system initialized")

	// Create help browser button (Shift+F1)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		log.Button("Help")
		helpBrowser.Show()
	})

	// Create toolbar with module label left, buttons aligned right
	fmt.Println("[STARTUP] Creating toolbar...")
	toolbar := container.NewBorder(
		nil, nil,
		currentModuleLabel, // Left side: current module name
		container.NewHBox(homeBtn, docsBtn, helpBtn, undoToolbar, themeToggleBtn, notificationBtn, micLevelWidget, screenshotBtn, recordBtn, recordTimeLabel, closeBtn), // Right side: buttons (mic shows levels always)
	)

	// Stack content with toolbar on top
	fmt.Println("[STARTUP] Creating main content...")
	mainContent := container.NewBorder(
		toolbar, nil, nil, nil,
		contentStack,
	)

	// Create toast overlay positioned at top-right
	toastOverlay := container.NewBorder(
		container.NewHBox(layout.NewSpacer(), container.NewPadded(toastContainer)), // Top-right
		nil, nil, nil,
	)

	// Create notification panel overlay (positioned at top-right, below toolbar)
	notificationPanelOverlay := container.NewBorder(
		nil, nil, nil,
		container.NewVBox(
			container.NewPadded(notificationHistoryPanel),
		),
	)

	// Stack main content with overlays
	contentWithOverlays := container.NewStack(
		mainContent,
		notificationPanelOverlay,
		toastOverlay,
	)

	// Wrap in ForceMinSize container to prevent Wayland resize loops
	rootContainer := container.New(&ForceMinSizeLayout{Min: fyne.NewSize(100, 100)}, contentWithOverlays)

	// Set content in a deferred way to avoid fyne.Do() deadlock
	// First set a placeholder, then set real content after event loop starts
	fmt.Println("[STARTUP] Setting placeholder content...")
	loadingLabel := widget.NewLabel("Loading FeCIM Visualizer...")
	window.SetContent(container.NewCenter(loadingLabel))

	// Defer setting actual content until after main loop starts
	utils.SafeGo("startup-content-loader", func() {
		time.Sleep(200 * time.Millisecond)
		fyne.Do(func() {
			fmt.Println("[STARTUP] Setting actual window content...")
			window.SetContent(rootContainer)
			fmt.Println("[STARTUP] Window content set")

			// Start at requested module
			startIdx := 0
			switch *moduleFlag {
			case "hysteresis":
				startIdx = 1
			case "crossbar":
				startIdx = 2
			case "mnist":
				startIdx = 3
			case "circuits":
				startIdx = 4
			case "comparison":
				startIdx = 5
			case "eda":
				startIdx = 6
			case "docs":
				startIdx = 7
			}

			fmt.Printf("[STARTUP] Starting at module: %s (index %d)\n", *moduleFlag, startIdx)
			selectView(startIdx)
			log.Debug("Started at module %s (index %d)", *moduleFlag, startIdx)
			fmt.Printf("[STARTUP] %s tab selected\n", *moduleFlag)

			// Show startup tip after a short delay
			utils.SafeGo("startup-tip", func() {
				time.Sleep(800 * time.Millisecond)
				fyne.Do(func() {
					tipManager.ShowStartupTip()
				})
			})
		})
	})

	// Prepare audio monitoring (but don't start FFmpeg yet - that's expensive)
	// Audio monitoring will be started lazily when Record button is clicked
	utils.SafeGo("audio-init", func() {
		time.Sleep(500 * time.Millisecond) // Let UI render first
		fmt.Println("[STARTUP] Checking audio availability (background)...")

		if recording.IsAudioAvailable() {
			log.Info("Audio input detected, mic level indicator available")
			monitor := recording.NewAudioMonitor()

			// Try to set default device
			if defaultDevice, err := recording.GetDefaultAudioDevice(); err == nil {
				monitor.SetDevice(defaultDevice)
				log.Debug("Using audio device: %s", defaultDevice.Name)
			}

			// Create controller to connect monitor to widget
			controller := sharedwidgets.NewMicLevelController(micLevelWidget, monitor)

			// Store references for lazy start (don't start FFmpeg now - saves 5-11s at startup)
			audioMonitorMu.Lock()
			audioMonitor = monitor
			micController = controller
			audioMonitorMu.Unlock()

			log.Debug("Audio monitoring prepared (will start on first record)")
		} else {
			log.Info("No audio input detected, mic level indicator disabled")
		}
		fmt.Println("[STARTUP] Audio initialization complete")
	})
	fmt.Println("[STARTUP] Placeholder content set")

	// Start window resize tracking for debugging (if FYNE_DEBUG_RESIZE is set)
	if sharedwidgets.DebugResize {
		utils.SafeGo("resize-debugger", func() {
			var lastSize fyne.Size
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				currentSize := window.Canvas().Size()
				if currentSize.Width != lastSize.Width || currentSize.Height != lastSize.Height {
					sharedwidgets.DebugWindowResize(currentSize)
					lastSize = currentSize
				}
			}
		})
		log.Info("Resize debugging enabled - set FYNE_DEBUG_RESIZE=1")
	}

	// Set close intercept to save final state
	window.SetCloseIntercept(func() {
		log.Info("Application closing, saving state...")

		// Save final window size
		finalSize := window.Canvas().Size()
		saveWindowSize(prefs, finalSize)
		log.Debug("Final window size saved: %.0fx%.0f", finalSize.Width, finalSize.Height)

		// Save current view index
		saveLastTab(prefs, currentViewIndex)
		log.Debug("Last view saved: %d (%s)", currentViewIndex, viewNames[currentViewIndex])

		// Stop recording if active
		if recordingState.IsRecording() {
			recordingState.stopRecording()
		}

		// Stop audio monitoring (with mutex since it may still be initializing)
		audioMonitorMu.Lock()
		if micController != nil {
			micController.Stop()
		}
		if audioMonitor != nil {
			audioMonitor.Stop()
		}
		audioMonitorMu.Unlock()

		// Stop all demos to clean up resources (e.g., auto demo contexts)
		log.Debug("Stopping all demos before window close...")
		demos.demo1.Stop()
		if demos.demo2 != nil {
			demos.demo2.Stop()
		}
		demos.demo3.Stop()
		demos.demo4.Stop()
		demos.demo5.Stop()
		demos.demo6.Stop()

		// Close the window
		window.Close()
	})

	// Create File menu with recent files
	fmt.Println("[STARTUP] Creating File menu...")
	recentFilesMenu := recentfiles.NewRefreshableRecentMenu(recentFilesManager, "Recent Files", recentfiles.FileTypeAny)
	recentFilesMenu.SetOnOpen(func(file *recentfiles.RecentFile) {
		if file == nil {
			return
		}
		log.Debug("Opening recent file: %s", file.Path)
		if _, err := os.Stat(file.Path); err != nil {
			notificationManager.Error("Recent File Missing", fmt.Sprintf("%s no longer exists", file.Name))
			return
		}

		cmd := exec.Command("xdg-open", file.Path)
		if err := cmd.Start(); err != nil {
			notificationManager.Error("Open Failed", err.Error())
			return
		}

		recentFilesManager.Add(file.Path, file.Type, viewContextKeys[currentViewIndex])
		notificationManager.Success("Opening Recent File", file.Name)
	})

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Config...", func() {
			log.Button("File > Open Config")
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					notificationManager.Error("Error", err.Error())
					return
				}
				if reader == nil {
					return // User cancelled
				}
				defer reader.Close()
				path := reader.URI().Path()
				recentFilesManager.AddConfig(path, viewContextKeys[currentViewIndex])
				notificationManager.Success("Config Loaded", filepath.Base(path))
			}, window)
		}),
		recentFilesMenu.MenuItem(),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Export...", func() {
			log.Button("File > Export")
			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					notificationManager.Error("Error", err.Error())
					return
				}
				if writer == nil {
					return // User cancelled
				}
				defer writer.Close()
				path := writer.URI().Path()
				recentFilesManager.AddExport(path, viewContextKeys[currentViewIndex])
				notificationManager.Info("Export", "Export destination: "+filepath.Base(path))
			}, window)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			log.Button("File > Quit")
			window.Close()
		}),
	)

	// Create Edit menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Undo", func() {
			if undoManager.Undo() {
				log.Debug("Undo via menu")
			}
		}),
		fyne.NewMenuItem("Redo", func() {
			if undoManager.Redo() {
				log.Debug("Redo via menu")
			}
		}),
	)

	// Create View menu for module navigation
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Home", func() { selectView(0) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Hysteresis", func() { selectView(1) }),
		fyne.NewMenuItem("Crossbar", func() { selectView(2) }),
		fyne.NewMenuItem("MNIST", func() { selectView(3) }),
		fyne.NewMenuItem("Circuits", func() { selectView(4) }),
		fyne.NewMenuItem("Comparison", func() { selectView(5) }),
		fyne.NewMenuItem("EDA", func() { selectView(6) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Documentation", func() { selectView(7) }),
	)

	// Create Settings menu (includes accessibility options)
	largeTextItem := fyne.NewMenuItem("Toggle Large Text Mode", func() {
		enabled := !accessibility.IsLargeTextModeEnabled(fyneApp)
		accessibility.SetLargeTextMode(fyneApp, enabled)
		themeManager.ApplyAccessibilityPreferences()
		notificationManager.Info("Accessibility", fmt.Sprintf("Large text mode %s", map[bool]string{true: "enabled", false: "disabled"}[enabled]))
	})
	reducedMotionItem := fyne.NewMenuItem("Toggle Reduced Motion", func() {
		enabled := !accessibility.IsReducedMotionEnabled(fyneApp)
		accessibility.SetReducedMotion(fyneApp, enabled)
		notificationManager.Info("Accessibility", fmt.Sprintf("Reduced motion %s", map[bool]string{true: "enabled", false: "disabled"}[enabled]))
	})
	settingsMenu := fyne.NewMenu("Settings",
		fyne.NewMenuItem("Theme: Dark (FeCIM)", func() {
			themeManager.SetTheme(themes.ThemeDark)
		}),
		fyne.NewMenuItem("Theme: Light", func() {
			themeManager.SetTheme(themes.ThemeLight)
		}),
		fyne.NewMenuItem("Theme: High Contrast", func() {
			themeManager.SetTheme(themes.ThemeHighContrast)
		}),
		fyne.NewMenuItemSeparator(),
		largeTextItem,
		reducedMotionItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Keyboard Shortcuts", func() {
			sharedwidgets.ShowKeyboardHelp(window)
		}),
	)

	// Create Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Help Topics", func() {
			helpBrowser.Show()
		}),
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About FeCIM Lattice Tools",
				"FeCIM Lattice Tools v1.0\n\n"+
					"A unified visualization and simulation suite for\n"+
					"Ferroelectric Compute-in-Memory (FeCIM) technology.\n\n"+
					"Includes 6 interactive demos covering:\n"+
					"• Hysteresis simulation\n"+
					"• Crossbar array visualization\n"+
					"• MNIST neural network inference\n"+
					"• Peripheral circuits\n"+
					"• Technology comparison\n"+
					"• EDA design suite",
				window)
		}),
	)

	// Set main menu
	mainMenu := fyne.NewMainMenu(fileMenu, editMenu, viewMenu, settingsMenu, helpMenu)
	window.SetMainMenu(mainMenu)
	fmt.Println("[STARTUP] File menu created")

	// Run the application
	fmt.Println("[STARTUP] About to call ShowAndRun()...")
	window.ShowAndRun()

	// Cleanup all demos on exit
	demos.demo1.Stop()
	if demos.demo2 != nil {
		demos.demo2.Stop()
	}
	demos.demo3.Stop()
	demos.demo4.Stop()
	demos.demo5.Stop()
	demos.demo6.Stop()

	// Stop recording if still running
	if recordingState.IsRecording() {
		recordingState.stopRecording()
	}
}
