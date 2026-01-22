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
	"fmt"
	"image"
	"image/color"
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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	demo1gui "multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/gui"
	demo2gui "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/gui"
	demo3gui "multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/gui"
	demo4gui "multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/gui"
	demo5gui "multilayer-ferroelectric-cim-visualizer/module5-comparison/pkg/gui"
	demo6gui "multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui"
)

// FeCIM theme colors
var (
	colorBackground = color.RGBA{0, 50, 100, 255}  // FeCIM blue #003264
	colorPrimary    = color.RGBA{0, 212, 255, 255} // Cyan
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
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
	case "6. EDA":
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

func main() {
	// Create Fyne app
	fyneApp := app.NewWithID("com.fecim.visualizer")
	fyneApp.Settings().SetTheme(&feCIMTheme{})

	// Create main window
	window := fyneApp.NewWindow("FeCIM Visualization Suite - 6 World-Class Demos")
	window.Resize(fyne.NewSize(1400, 900))

	// Create demo instances
	demos := &DemoApp{
		demo1: demo1gui.NewEmbeddedApp(),
		demo2: demo2gui.NewEmbeddedCrossbarApp(), // Original single-view crossbar
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
	demo2Content := demos.demo2.BuildContent(fyneApp, window)
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
		container.NewTabItem("6. EDA", container.NewMax(demo6Content)),
	)

	// Create recording state
	recordingState := newRecordingState()

	// Create screenshot button
	screenshotBtn := widget.NewButtonWithIcon("Screenshot", theme.MediaPhotoIcon(), func() {
		// Get current section name from selected tab
		sectionName := "home"
		if tabs.Selected() != nil {
			sectionName = sectionNameFromTab(tabs.Selected().Text)
		}
		filename := takeScreenshot(window, sectionName)
		if filename != "" {
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
		if recordingState.IsRecording() {
			// Stop the timer goroutine
			if recordingTimerStop != nil {
				close(recordingTimerStop)
				recordingTimerStop = nil
			}
			// Stop recording
			outputFile, err := recordingState.stopRecording()
			if err != nil {
				fmt.Println("Error stopping recording:", err)
				return
			}
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

	// Create close button
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		// Stop recording if active
		if recordingState.IsRecording() {
			recordingState.stopRecording()
		}
		// Close the application
		fyneApp.Quit()
	})

	// Create toolbar with screenshot, record, and close buttons on the right
	toolbarButtons := container.NewHBox(
		layout.NewSpacer(),
		screenshotBtn,
		recordBtn,
		closeBtn,
	)

	// Track current demo for start/stop
	currentDemo := 0

	// Handle tab changes - start/stop simulations as needed
	tabs.OnSelected = func(tab *container.TabItem) {
		// Stop previous demo
		switch currentDemo {
		case 1:
			demos.demo1.Stop()
		case 2:
			demos.demo2.Stop()
		case 3:
			demos.demo3.Stop()
		case 4:
			demos.demo4.Stop()
		case 5:
			demos.demo5.Stop()
		case 6:
			demos.demo6.Stop()
		}

		// Start new demo
		switch tab.Text {
		case "1. Hysteresis":
			currentDemo = 1
			demos.demo1.Start()
		case "2. Crossbar+":
			currentDemo = 2
			demos.demo2.Start()
		case "3. MNIST":
			currentDemo = 3
			demos.demo3.Start()
		case "4. Circuits":
			currentDemo = 4
			demos.demo4.Start()
		case "5. Comparison":
			currentDemo = 5
			demos.demo5.Start()
		case "6. EDA":
			currentDemo = 6
			demos.demo6.Start()
		default:
			currentDemo = 0
		}
	}

	// Create main content with toolbar at top right
	// Use Border layout to put toolbar above tabs
	mainContent := container.NewBorder(
		toolbarButtons, // top
		nil,            // bottom
		nil,            // left
		nil,            // right
		tabs,           // center
	)

	// Set window content
	window.SetContent(mainContent)

	// Run the application
	window.ShowAndRun()

	// Cleanup all demos on exit
	demos.demo1.Stop()
	demos.demo2.Stop()
	demos.demo3.Stop()
	demos.demo4.Stop()
	demos.demo5.Stop()
	demos.demo6.Stop()

	// Stop recording if still running
	if recordingState.IsRecording() {
		recordingState.stopRecording()
	}
}
