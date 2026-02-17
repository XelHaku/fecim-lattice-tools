package main

import (
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

	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/recording"
	"fecim-lattice-tools/shared/canvas"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
)

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

var (
	screenshotOutputDir = "screenshots"
	recordingOutputDir  = "recordings"
)

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

func isVerbosityToken(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "0", "off", "none", "1", "info", "2", "debug", "3", "trace", "all":
		return true
	default:
		return false
	}
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
func takeScreenshot(window fyne.Window, sectionName string) string {
	screenshotDir := screenshotOutputDir
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		logging.Printf("Error creating screenshots directory: %v", err)
		return ""
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(screenshotDir, fmt.Sprintf("fecim_%s_%s.png", sectionName, timestamp))

	img := window.Canvas().Capture()

	meta := utils.DefaultScreenshotMetadata(sectionName)
	meta.CustomData["Window-Size"] = fmt.Sprintf("%dx%d", int(window.Canvas().Size().Width), int(window.Canvas().Size().Height))

	if err := utils.SavePNGWithMetadata(img, filename, meta); err != nil {
		logging.Printf("Error saving screenshot with metadata: %v", err)
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

	recordingDir := recordingOutputDir
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return fmt.Errorf("error creating recordings directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	rs.outputFile = filepath.Join(recordingDir, fmt.Sprintf("fecim_recording_%s.mp4", timestamp))
	rs.window = window
	rs.stopChan = make(chan struct{})
	rs.stopped = false

	img := window.Canvas().Capture()
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width%2 != 0 {
		width--
	}
	if height%2 != 0 {
		height--
	}

	audioDevice := ""
	isMicrophone := false
	if recording.IsAudioAvailable() {
		if micDevice, err := recording.GetMicrophoneDevice(); err == nil {
			audioDevice = micDevice.Name
			isMicrophone = true
		} else {
			if desktopDevice, err := recording.GetDesktopAudioDevice(); err == nil {
				audioDevice = desktopDevice.Name
			}
		}
	}

	var ffmpegArgs []string
	if audioDevice != "" {
		ffmpegArgs = []string{
			"-y", "-f", "rawvideo", "-pixel_format", "rgb24",
			"-video_size", fmt.Sprintf("%dx%d", width, height),
			"-framerate", "20", "-i", "-", "-f", "pulse", "-i", audioDevice,
			"-c:v", "libx264", "-preset", "ultrafast", "-crf", "23", "-pix_fmt", "yuv420p",
		}

		if isMicrophone {
			ffmpegArgs = append(ffmpegArgs, "-af",
				"highpass=f=80,lowpass=f=7500,afftdn=nr=80:nf=-25:tn=1,agate=threshold=-35dB:ratio=3:attack=10:release=150,rubberband=pitch=0.92,equalizer=f=200:t=h:w=200:g=4,equalizer=f=100:t=h:w=100:g=3,acompressor=threshold=-20dB:ratio=3:attack=10:release=150:makeup=5dB,alimiter=limit=0.95")
		}

		ffmpegArgs = append(ffmpegArgs,
			"-c:a", "aac", "-b:a", "192k", "-ac", "2", "-shortest",
			rs.outputFile,
		)
	} else {
		ffmpegArgs = []string{
			"-y", "-f", "rawvideo", "-pixel_format", "rgb24",
			"-video_size", fmt.Sprintf("%dx%d", width, height),
			"-framerate", "20", "-i", "-",
			"-c:v", "libx264", "-preset", "ultrafast", "-crf", "23", "-pix_fmt", "yuv420p",
			rs.outputFile,
		}
	}
	rs.cmd = exec.Command("ffmpeg", ffmpegArgs...)

	var err error
	rs.stdin, err = rs.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error getting stdin pipe: %w", err)
	}

	if err := rs.cmd.Start(); err != nil {
		return fmt.Errorf("error starting FFmpeg: %w", err)
	}

	rs.isRecording = true

	utils.SafeGo("captureLoop", func() {
		rs.captureLoop(width, height)
	})

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

	if shouldCloseStop {
		close(stopChan)
	}

	if stdin != nil {
		_ = stdin.Close()
	}

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

	return outputFile, nil
}

// IsRecording returns current recording state
func (rs *RecordingState) IsRecording() bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.isRecording
}
