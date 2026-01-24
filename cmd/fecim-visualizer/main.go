// Command fecim-visualizer provides a unified GUI application with all FeCIM demos as tabs.
//
// This is the main entry point for the FeCIM Visualization Suite.
// It combines all 6 demos into a single application with tab navigation.
//
// The 6-Demo Story:
//   Demo 1: The Memory Cell (Hysteresis) - How the cell works
//   Demo 2: The Crossbar Computer (MVM + Non-Idealities) - How we compute
//   Demo 3: The AI Brain (MNIST) - What we can build
//   Demo 4: The Chip System (Circuits) - How it fits in a chip
//   Demo 5: Why FeCIM Wins (Comparison) - The business case
//   Demo 6: EDA Design Suite - Bridge to open-source EDA tools
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	demo1gui "multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/gui"
	demo2gui "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/gui"
	demo3gui "multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/gui"
	demo4gui "multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/gui"
	demo5gui "multilayer-ferroelectric-cim-visualizer/module5-comparison/pkg/gui"
	demo6gui "multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui"
	"multilayer-ferroelectric-cim-visualizer/shared/logging"
	sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
	"multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// Global logger for the main application
var log *logging.Logger

// RecordingState manages FFmpeg video recording using canvas capture
type RecordingState struct {
	mu          sync.Mutex
	isRecording bool
	stopChan    chan struct{}
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	outputFile  string
	window      fyne.Window
}

func newRecordingState() *RecordingState {
	return &RecordingState{}
}

// sectionNameFromTab converts tab text to a filename-friendly section name
func sectionNameFromTab(tabText string) string {
	switch tabText {
	case "Home":
		return "home"
	case "1. Hysteresis":
		return "module01-hysteresis"
	case "2. Crossbar+":
		return "module02-crossbar"
	case "3. MNIST":
		return "module03-mnist"
	case "4. Circuits":
		return "module04-circuits"
	case "5. Comparison":
		return "module05-comparison"
	case "6. EDA (Work In Progress)":
		return "module06-eda"
	default:
		return "unknown"
	}
}

// takeScreenshot captures the current window and saves it as PNG
func takeScreenshot(window fyne.Window, sectionName string) string {
	// Create screenshots directory if it doesn't exist
	screenshotDir := "screenshots"
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		fmt.Println("Error creating screenshots directory:", err)
		return ""
	}

	// Generate filename with timestamp and section name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(screenshotDir, fmt.Sprintf("fecim_%s_%s.png", sectionName, timestamp))

	// Capture the canvas
	img := window.Canvas().Capture()

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating screenshot file:", err)
		return ""
	}
	defer file.Close()

	// Encode as PNG
	if err := png.Encode(file, img); err != nil {
		fmt.Println("Error encoding screenshot:", err)
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
	recordingDir := "recordings"
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return fmt.Errorf("error creating recordings directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	rs.outputFile = filepath.Join(recordingDir, fmt.Sprintf("fecim_recording_%s.mp4", timestamp))
	rs.window = window
	rs.stopChan = make(chan struct{})

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

	// FFmpeg command to receive raw RGB frames from stdin
	rs.cmd = exec.Command("ffmpeg",
		"-y",                                      // Overwrite output file
		"-f", "rawvideo",                          // Raw video input
		"-pixel_format", "rgb24",                  // RGB24 format
		"-video_size", fmt.Sprintf("%dx%d", width, height), // Frame size
		"-framerate", "20",                        // 20 FPS (balance between smoothness and CPU)
		"-i", "-",                                 // Read from stdin
		"-c:v", "libx264",                         // H.264 codec
		"-preset", "ultrafast",                    // Fast encoding
		"-crf", "23",                              // Quality
		"-pix_fmt", "yuv420p",                     // Output pixel format
		rs.outputFile,
	)

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

	// Start capture goroutine
	go rs.captureLoop(width, height)

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

			// Write frame to FFmpeg
			rs.mu.Lock()
			if rs.stdin != nil && rs.isRecording {
				rs.stdin.Write(frameBuffer)
			}
			rs.mu.Unlock()
		}
	}
}

// stopRecording stops the FFmpeg recording
func (rs *RecordingState) stopRecording() (string, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if !rs.isRecording {
		return "", fmt.Errorf("not recording")
	}

	// Signal capture loop to stop
	close(rs.stopChan)

	// Close stdin to signal EOF to FFmpeg
	if rs.stdin != nil {
		rs.stdin.Close()
	}

	// Wait for FFmpeg to finish
	if rs.cmd != nil {
		rs.cmd.Wait()
	}

	rs.isRecording = false
	outputFile := rs.outputFile
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
	demo1 *demo1gui.EmbeddedApp             // Hysteresis
	demo2 *demo2gui.EmbeddedCrossbarApp     // Crossbar (original single-view)
	demo3 *demo3gui.EmbeddedDualModeApp     // MNIST FP vs CIM (full-featured)
	demo4 *demo4gui.EmbeddedCircuitsApp     // Circuits
	demo5 *demo5gui.EmbeddedComparisonApp   // Comparison (technical briefing)
	demo6 *demo6gui.EmbeddedEDAApp          // EDA Design Suite
}

// Preference keys for window state persistence
const (
	prefKeyWindowWidth  = "window_width"
	prefKeyWindowHeight = "window_height"
	prefKeyLastTab      = "last_tab"

	// Default window dimensions
	defaultWindowWidth  = 1400
	defaultWindowHeight = 900
)

// loadWindowSize loads saved window dimensions from preferences
func loadWindowSize(prefs fyne.Preferences) fyne.Size {
	w := prefs.FloatWithFallback(prefKeyWindowWidth, defaultWindowWidth)
	h := prefs.FloatWithFallback(prefKeyWindowHeight, defaultWindowHeight)

	// Clamp to reasonable minimum size
	if w < 800 {
		w = 800
	}
	if h < 600 {
		h = 600
	}

	return fyne.NewSize(float32(w), float32(h))
}

// saveWindowSize saves current window dimensions to preferences
func saveWindowSize(prefs fyne.Preferences, size fyne.Size) {
	prefs.SetFloat(prefKeyWindowWidth, float64(size.Width))
	prefs.SetFloat(prefKeyWindowHeight, float64(size.Height))
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
	// Parse command-line flags
	verbosityFlag := flag.String("verbosity", "off", "Logging verbosity: 0|off, 1|info, 2|debug, 3|trace")
	flag.Parse()

	// Set global verbosity level
	verbosity := logging.ParseVerbosityFlag(*verbosityFlag)
	logging.SetVerbosity(verbosity)

	// Initialize global logger
	log = logging.NewLogger("fecim-visualizer")
	defer log.Close()

	log.Info("FeCIM Visualizer starting with verbosity=%s", logging.VerbosityString(verbosity))

	// Create Fyne app
	fyneApp := app.NewWithID("com.fecim.visualizer")
	fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

	// Load saved window size from preferences
	prefs := fyneApp.Preferences()
	savedSize := loadWindowSize(prefs)
	log.Debug("Loaded window size from preferences: %.0fx%.0f", savedSize.Width, savedSize.Height)

	// Create main window with saved size
	window := fyneApp.NewWindow("FeCIM Visualization Suite - 6 World-Class Demos")
	window.Resize(savedSize)

	// Create demo instances with error handling
	demo2, err := demo2gui.NewEmbeddedCrossbarApp()
	if err != nil {
		log.Printf("[ERROR] Failed to create crossbar demo: %v", err)
		dialog.ShowError(fmt.Errorf("crossbar demo initialization failed: %w", err), window)
		// Create a placeholder - the demo tab will show an error
		demo2 = nil
	}

	demos := &DemoApp{
		demo1: demo1gui.NewEmbeddedApp(),
		demo2: demo2,                             // Original single-view crossbar (may be nil on error)
		demo3: demo3gui.NewEmbeddedDualModeApp(), // Full-featured MNIST with FP vs CIM
		demo4: demo4gui.NewEmbeddedCircuitsApp(),
		demo5: demo5gui.NewEmbeddedComparisonApp(),
		demo6: demo6gui.NewEmbeddedEDAApp(),
	}

	// Create tabs container (will be populated below)
	var tabs *container.AppTabs

	// Create launcher content with callback to switch tabs
	launcherContent := CreateLauncherContent(func(demoNum int) {
		if tabs != nil {
			// Map demo number to tab index
			// Home=0, Demo1=1, Demo2=2, Demo3=3, Demo4=4, Demo5=5, Demo6=6
			tabIndex := 0
			switch demoNum {
			case 1:
				tabIndex = 1
			case 2:
				tabIndex = 2
			case 3:
				tabIndex = 3
			case 4:
				tabIndex = 4
			case 5:
				tabIndex = 5
			case 6:
				tabIndex = 6
			}
			tabs.SelectIndex(tabIndex)
		}
	})

	// Build content for each demo
	demo1Content := demos.demo1.BuildContent(fyneApp, window)

	// Handle potentially nil demo2 (initialization error)
	var demo2Content fyne.CanvasObject
	if demos.demo2 != nil {
		demo2Content = demos.demo2.BuildContent(fyneApp, window)
	} else {
		demo2Content = container.NewCenter(widget.NewLabel("Crossbar demo failed to initialize. Check logs for details."))
	}

	demo3Content := demos.demo3.BuildContent(fyneApp, window)
	demo4Content := demos.demo4.BuildContent(fyneApp, window)
	demo5Content := demos.demo5.BuildContent(fyneApp, window)
	demo6Content := demos.demo6.BuildContent(fyneApp, window)

	// Create tabs - 6 demos total (plus home)
	tabs = container.NewAppTabs(
		container.NewTabItem("Home", launcherContent),
		container.NewTabItem("1. Hysteresis", container.NewMax(demo1Content)),
		container.NewTabItem("2. Crossbar+", container.NewMax(demo2Content)),
		container.NewTabItem("3. MNIST", container.NewMax(demo3Content)),
		container.NewTabItem("4. Circuits", container.NewMax(demo4Content)),
		container.NewTabItem("5. Comparison", container.NewMax(demo5Content)),
		container.NewTabItem("6. EDA (Work In Progress)", container.NewMax(demo6Content)),
	)

	// Create recording state
	recordingState := newRecordingState()

	// Track current breakpoint for responsive layout (initialize based on saved window size)
	currentBreakpoint := widgets.GetBreakpoint(savedSize.Width)
	log.Debug("Initial breakpoint: %s (width: %.0f)", widgets.BreakpointName(currentBreakpoint), savedSize.Width)

	// Create screenshot button
	screenshotBtn := widget.NewButtonWithIcon("Screenshot", theme.MediaPhotoIcon(), func() {
		log.Button("Screenshot")
		// Get current section name from selected tab
		sectionName := "home"
		if tabs.Selected() != nil {
			sectionName = sectionNameFromTab(tabs.Selected().Text)
		}
		filename := takeScreenshot(window, sectionName)
		if filename != "" {
			log.Debug("Screenshot saved: %s", filename)
			// Show brief notification in window title
			originalTitle := window.Title()
			window.SetTitle("Screenshot saved: " + filename)
			go func() {
				time.Sleep(2 * time.Second)
				fyne.Do(func() {
					window.SetTitle(originalTitle)
				})
			}()
		}
	})

	// Create record button with toggle functionality
	recordBtn := widget.NewButtonWithIcon("Record", theme.MediaRecordIcon(), nil)
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
			// Show brief notification
			originalTitle := window.Title()
			window.SetTitle("Recording saved: " + outputFile)
			go func() {
				time.Sleep(2 * time.Second)
				fyne.Do(func() {
					window.SetTitle(originalTitle)
				})
			}()
		} else {
			log.Debug("Starting recording...")
			// Start recording
			if err := recordingState.startRecording(window); err != nil {
				fmt.Println("Error starting recording:", err)
				// Show error in title briefly
				originalTitle := window.Title()
				window.SetTitle("Recording error: " + err.Error())
				go func() {
					time.Sleep(2 * time.Second)
					fyne.Do(func() {
						window.SetTitle(originalTitle)
					})
				}()
				return
			}
			recordBtn.SetIcon(theme.MediaStopIcon())
			// Start timer to show real-time datetime with milliseconds and take screenshots
			recordingTimerStop = make(chan struct{})
			go func() {
				displayTicker := time.NewTicker(50 * time.Millisecond)  // Update display ~20 times per second
				screenshotTicker := time.NewTicker(5 * time.Second)     // Screenshot every 5 seconds
				defer displayTicker.Stop()
				defer screenshotTicker.Stop()
				for {
					select {
					case <-recordingTimerStop:
						return
					case <-displayTicker.C:
						fyne.Do(func() {
							now := time.Now()
							recordBtn.SetText(now.Format("02 Jan 06 15:04:05.000"))
						})
					case <-screenshotTicker.C:
						// Take screenshot on main thread
						fyne.Do(func() {
							sectionName := "recording"
							if tabs.Selected() != nil {
								sectionName = sectionNameFromTab(tabs.Selected().Text)
							}
							takeScreenshot(window, sectionName)
						})
					}
				}
			}()
		}
	}

	// Create close button (triggers close intercept for proper state saving)
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		log.Button("Close")
		// Trigger the close intercept which handles state saving
		window.Close()
	})

	// Track current demo for start/stop
	currentDemo := 0

	// Handle tab changes - start/stop simulations as needed
	tabs.OnSelected = func(tab *container.TabItem) {
		log.TabChange(tab.Text)

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
		}

		// Start new demo
		switch tab.Text {
		case "1. Hysteresis":
			currentDemo = 1
			log.Debug("Starting demo1 (Hysteresis)")
			demos.demo1.Start()
		case "2. Crossbar+":
			currentDemo = 2
			log.Debug("Starting demo2 (Crossbar+)")
			if demos.demo2 != nil {
				demos.demo2.Start()
			}
		case "3. MNIST":
			currentDemo = 3
			log.Debug("Starting demo3 (MNIST)")
			demos.demo3.Start()
		case "4. Circuits":
			currentDemo = 4
			log.Debug("Starting demo4 (Circuits)")
			demos.demo4.Start()
		case "5. Comparison":
			currentDemo = 5
			log.Debug("Starting demo5 (Comparison)")
			demos.demo5.Start()
		case "6. EDA (Work In Progress)":
			currentDemo = 6
			log.Debug("Starting demo6 (EDA)")
			demos.demo6.Start()
		default:
			currentDemo = 0
		}
	}

	// Put toolbar buttons on the right side of the tab bar (same row)
	// We'll position them absolutely in the top-right corner
	tabs.SetTabLocation(container.TabLocationTop)

	// Create overlay container with buttons in top-right
	buttonOverlay := container.NewWithoutLayout(
		screenshotBtn,
		recordBtn,
		closeBtn,
	)

	// Get button width for current breakpoint (no side effects)
	getButtonWidth := func() float32 {
		switch currentBreakpoint {
		case widgets.BreakpointSM:
			return float32(32) // Square icon buttons
		case widgets.BreakpointMD:
			return float32(80) // Shortened labels
		default:
			return float32(120) // Full labels
		}
	}

	// Update button labels based on breakpoint (only call when breakpoint changes)
	updateButtonLabels := func() {
		switch currentBreakpoint {
		case widgets.BreakpointSM:
			screenshotBtn.SetText("")
			if !recordingState.IsRecording() {
				recordBtn.SetText("")
			}
		case widgets.BreakpointMD:
			screenshotBtn.SetText("Shot")
			if !recordingState.IsRecording() {
				recordBtn.SetText("Rec")
			}
		default:
			screenshotBtn.SetText("Screenshot")
			if !recordingState.IsRecording() {
				recordBtn.SetText("Record")
			}
		}
	}

	// Position buttons in top-right corner (will be updated on resize)
	// Only handles positioning, not text changes
	updateButtonPositions := func() {
		size := window.Canvas().Size()
		btnHeight := float32(32)
		spacing := float32(5)
		btnWidth := getButtonWidth()

		closeBtn.Resize(fyne.NewSize(btnHeight, btnHeight))
		closeBtn.Move(fyne.NewPos(size.Width-btnHeight-10, 5))

		recordBtn.Resize(fyne.NewSize(btnWidth, btnHeight))
		recordBtn.Move(fyne.NewPos(size.Width-btnHeight-btnWidth-spacing-10, 5))

		screenshotBtn.Resize(fyne.NewSize(btnWidth, btnHeight))
		screenshotBtn.Move(fyne.NewPos(size.Width-btnHeight-2*btnWidth-2*spacing-10, 5))
	}

	// Initial labels and position
	updateButtonLabels()
	updateButtonPositions()

	// Stack tabs with button overlay
	mainContent := container.NewStack(
		tabs,
		buttonOverlay,
	)

	// Set window content and add resize callback
	window.SetContent(mainContent)
	window.Canvas().SetOnTypedRune(nil) // Dummy to ensure canvas is initialized

	// Create resize detector to handle window resize events (replaces polling)
	var lastSaveTime time.Time
	resizeDetector := widgets.NewResizeDetector(func(size fyne.Size) {
		// Update button positions (Move/Resize handle their own refresh)
		updateButtonPositions()

		// Debounce: only save size every 500ms to avoid excessive writes
		if time.Since(lastSaveTime) > 500*time.Millisecond {
			saveWindowSize(prefs, size)
			lastSaveTime = time.Now()
			log.Debug("Window resized and saved: %.0fx%.0f", size.Width, size.Height)
		}
	})

	// Create responsive detector for breakpoint-based layout changes
	responsiveDetector := widgets.NewResponsiveDetector(func(newBreakpoint widgets.Breakpoint, size fyne.Size) {
		currentBreakpoint = newBreakpoint
		log.Debug("Breakpoint changed to %s (width: %.0f)", widgets.BreakpointName(newBreakpoint), size.Width)

		// Update button labels and layout for new breakpoint
		updateButtonLabels()
		updateButtonPositions()
	})

	// Add resize and responsive detectors to the main content stack
	mainContent = container.NewStack(mainContent, resizeDetector, responsiveDetector)

	// Set close intercept to save final state
	window.SetCloseIntercept(func() {
		log.Info("Application closing, saving state...")

		// Save final window size
		finalSize := window.Canvas().Size()
		saveWindowSize(prefs, finalSize)
		log.Debug("Final window size saved: %.0fx%.0f", finalSize.Width, finalSize.Height)

		// Save current tab index
		if tabs.Selected() != nil {
			for i, tab := range tabs.Items {
				if tab == tabs.Selected() {
					saveLastTab(prefs, i)
					log.Debug("Last tab saved: %d (%s)", i, tab.Text)
					break
				}
			}
		}

		// Stop recording if active
		if recordingState.IsRecording() {
			recordingState.stopRecording()
		}

		// Close the window
		window.Close()
	})

	// Restore last selected tab
	lastTabIndex := loadLastTab(prefs)
	if lastTabIndex > 0 && lastTabIndex < len(tabs.Items) {
		tabs.SelectIndex(lastTabIndex)
		log.Debug("Restored last tab: %d", lastTabIndex)
	}

	// Run the application
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
