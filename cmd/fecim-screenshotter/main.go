// Command fecim-screenshotter captures automated screenshots of the FeCIM GUI.
//
// Usage: fecim-screenshotter -module 1 -output screenshots/
//
// # Xwayland note
//
// If the session is running under Xwayland (XDG_SESSION_TYPE=wayland, X server
// is "Xwayland :0 -rootless"), external X11 screenshot tools such as maim,
// scrot, and xwd will always produce all-black images. This is an architectural
// limitation: Xwayland composites X11 windows inside the Wayland compositor
// buffer; the X11 root-window framebuffer is unmapped and contains no pixels.
//
// This tool is NOT affected by that limitation. It uses fyne.Canvas.Capture(),
// which calls glReadPixels(GL_FRONT) on the application's own OpenGL context,
// bypassing X11 screen capture entirely.
//
// For external capture on a Wayland session use grim(1):
//
//	grim /tmp/full-desktop.png
//	grim -g "$(slurp)" /tmp/window.png
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"flag"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	sharedtheme "fecim-lattice-tools/shared/theme"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
)

type moduleLifecycle interface {
	BuildContent(fyne.App, fyne.Window) fyne.CanvasObject
	Start()
	Stop()
}

func main() {
	outDir := flag.String("out", "cmd/fecim-lattice-tools/testdata/screenshots", "output directory for screenshots")
	sizeW := flag.Int("w", 1200, "window width")
	sizeH := flag.Int("h", 800, "window height")
	only := flag.String("only", "", "capture only a single module (hysteresis|crossbar|mnist|circuits|comparison|eda|docs)")
	tag := flag.String("tag", "initial", "filename tag suffix (e.g. initial|after_badges)")
	hysEngine := flag.String("hys-engine", "preisach", "hysteresis physics engine: preisach|lk")
	autoXvfb := flag.Bool("auto-xvfb", true, "auto-start Xvfb when DISPLAY/WAYLAND_DISPLAY is missing")
	xvfbScreen := flag.String("xvfb-screen", "1920x1080x24", "Xvfb virtual screen (WIDTHxHEIGHTxDEPTH)")
	xvfbDisplayStart := flag.Int("xvfb-display-start", 99, "first Xvfb display number to try")
	xvfbDisplayEnd := flag.Int("xvfb-display-end", 110, "last Xvfb display number to try")
	flag.Parse()

	cleanupDisplay, err := ensureDisplayForCapture(*autoXvfb, *xvfbScreen, *xvfbDisplayStart, *xvfbDisplayEnd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize graphical display: %v\n", err)
		os.Exit(2)
	}
	if cleanupDisplay != nil {
		defer cleanupDisplay()
	}

	checkDisplayBrightness()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create out dir: %v\n", err)
		os.Exit(2)
	}

	mods := []struct {
		name   string
		settle time.Duration
		start  bool
		create func() (moduleLifecycle, error)
	}{
		{"hysteresis", 800 * time.Millisecond, true, func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil }},
		{"crossbar", 1500 * time.Millisecond, true, func() (moduleLifecycle, error) { return demo2gui.NewEmbeddedCrossbarApp() }},
		// MNIST Start() can spawn long-running loops; for initial-state screenshots we skip Start.
		{"mnist", 1200 * time.Millisecond, false, func() (moduleLifecycle, error) { return demo3gui.NewEmbeddedDualModeApp(), nil }},
		{"circuits", 1200 * time.Millisecond, true, func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil }},
		{"comparison", 1200 * time.Millisecond, true, func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil }},
		{"eda", 1200 * time.Millisecond, true, func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil }},
		{"docs", 900 * time.Millisecond, false, func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil }},
	}

	for _, m := range mods {
		if *only != "" && m.name != *only {
			continue
		}
		fmt.Printf("[screenshotter] capturing %s...\n", m.name)
		settle := m.settle
		if m.name == "hysteresis" {
			// LK needs a bit more time to populate a meaningful loop.
			if strings.ToLower(strings.TrimSpace(*hysEngine)) == "lk" {
				settle = 2500 * time.Millisecond
			}
		}
		fileBase := m.name + "_" + strings.TrimSpace(*tag)
		if err := captureOne(*outDir, fileBase, fyne.NewSize(float32(*sizeW), float32(*sizeH)), settle, m.start, *hysEngine, m.create); err != nil {
			fmt.Fprintf(os.Stderr, "[screenshotter] %s failed: %v\n", m.name, err)
		}
	}
}

func captureOne(outDir, fileBase string, size fyne.Size, settle time.Duration, start bool, hysEngine string, create func() (moduleLifecycle, error)) error {
	module, err := create()
	if err != nil {
		return err
	}

	a := app.NewWithID("com.fecim.screenshotter")
	a.Settings().SetTheme(&sharedtheme.FeCIMTheme{})
	w := a.NewWindow("FeCIM Screenshotter - " + fileBase)
	w.Resize(size)

	// Quit after we capture, so app.Run returns and we can proceed to next module.
	captured := make(chan error, 1)

	a.Lifecycle().SetOnStarted(func() {
		fmt.Printf("[screenshotter] onStarted %s\n", fileBase)
		fmt.Printf("[screenshotter] %s: BuildContent...\n", fileBase)
		content := module.BuildContent(a, w)
		fmt.Printf("[screenshotter] %s: SetContent/Show...\n", fileBase)
		bg := canvas.NewRectangle(theme.BackgroundColor())
		w.SetContent(container.NewMax(bg, content))
		w.Show()

		if strings.HasPrefix(fileBase, "hysteresis_") {
			if h, ok := module.(*demo1gui.EmbeddedApp); ok {
				// OnStarted runs on the UI thread; calling fyne.DoAndWait here trips Fyne's guard.
				h.SetPhysicsEngineName(hysEngine)
				w.Canvas().Refresh(w.Content())
			}
		}

		if start {
			module.Start()
		}

		time.AfterFunc(settle, func() {
			fmt.Printf("[screenshotter] capturing now %s\n", fileBase)
			if !waitForCanvasSize(w, size.Width, size.Height, 3*time.Second) {
				err := fmt.Errorf("canvas size not ready for capture")
				fyne.DoAndWait(func() {
					module.Stop()
					w.Close()
				})
				captured <- err
				fyne.DoAndWait(func() {
					a.Quit()
				})
				return
			}
			fyne.DoAndWait(func() {
				filename := filepath.Join(outDir, fileBase+".png")
				// Some widgets paint one frame late under headless/software rendering.
				// Capture a few times with small delays; keep the last capture.
				var img image.Image
				for i := 0; i < 4; i++ {
					w.Canvas().Refresh(w.Content())
					img = w.Canvas().Capture()
					time.Sleep(150 * time.Millisecond)
				}

				// Retry once on all-black result. Under Xwayland + Mesa exchange-swap the
				// GL front buffer may be undefined immediately after the first SwapBuffers.
				// A second Refresh + short sleep gives Mesa time to populate it.
				if isImageAllBlack(img) {
					fmt.Printf("[screenshotter] WARNING: Canvas().Capture() returned all-black for %s — retrying after extra refresh\n", fileBase)
					if content := w.Content(); content != nil {
						w.Canvas().Refresh(content)
						time.Sleep(200 * time.Millisecond)
						fallbackImg := w.Canvas().Capture()
						if !isImageAllBlack(fallbackImg) {
							fmt.Printf("[screenshotter] retry capture succeeded for %s\n", fileBase)
							img = fallbackImg
						} else {
							diag := captureBlackDiagnostic()
							fmt.Fprintf(os.Stderr, "[screenshotter] WARNING: capture still all-black for %s.\n%s\n", fileBase, diag)
						}
					}
				}

				module.Stop()
				w.Close()
				captured <- savePNG(filename, img)
				a.Quit()
			})
		})
	})

	fmt.Printf("[screenshotter] %s: Run...\n", fileBase)
	a.Run()
	fmt.Printf("[screenshotter] %s: Run returned\n", fileBase)

	select {
	case err := <-captured:
		if err == nil {
			fmt.Printf("[screenshotter] saved %s/%s.png\n", outDir, fileBase)
		}
		return err
	case <-time.After(settle + 10*time.Second):
		return fmt.Errorf("timed out waiting for capture")
	}
}

func waitForCanvasSize(w fyne.Window, wantW, wantH float32, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var got fyne.Size
		fyne.DoAndWait(func() {
			got = w.Canvas().Size()
		})
		fmt.Printf("[screenshotter] canvas size %0.1fx%0.1f (want %0.1fx%0.1f)\n", got.Width, got.Height, wantW, wantH)
		if got.Width >= wantW-2 && got.Height >= wantH-2 {
			return true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false
}

type xvfbSession struct {
	display string
	cmd     *exec.Cmd
	waitCh  chan error
	stderr  bytes.Buffer
}

func (s *xvfbSession) stop() {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return
	}
	_ = s.cmd.Process.Kill()
	if s.waitCh == nil {
		return
	}
	select {
	case <-s.waitCh:
	case <-time.After(2 * time.Second):
	}
}

func ensureDisplayForCapture(auto bool, screen string, displayStart, displayEnd int) (func(), error) {
	if hasGraphicalSession() {
		return nil, nil
	}
	if !auto {
		return nil, fmt.Errorf("DISPLAY and WAYLAND_DISPLAY are both unset (enable -auto-xvfb or run inside an X11/Wayland session)")
	}

	sess, err := startAutoXvfb(screen, displayStart, displayEnd)
	if err != nil {
		return nil, err
	}
	if err := os.Setenv("DISPLAY", sess.display); err != nil {
		sess.stop()
		return nil, fmt.Errorf("failed to set DISPLAY=%s: %w", sess.display, err)
	}
	_ = os.Unsetenv("WAYLAND_DISPLAY")

	fmt.Printf("[screenshotter] started Xvfb on DISPLAY=%s\n", sess.display)
	return func() {
		sess.stop()
		fmt.Printf("[screenshotter] stopped Xvfb on DISPLAY=%s\n", sess.display)
	}, nil
}

func hasGraphicalSession() bool {
	return strings.TrimSpace(os.Getenv("DISPLAY")) != "" || strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) != ""
}

func startAutoXvfb(screen string, displayStart, displayEnd int) (*xvfbSession, error) {
	xvfbBin, err := exec.LookPath("Xvfb")
	if err != nil {
		return nil, fmt.Errorf("Xvfb binary not found in PATH: %w", err)
	}
	if displayStart > displayEnd {
		displayStart, displayEnd = displayEnd, displayStart
	}
	if displayStart <= 0 {
		displayStart = 99
	}
	if displayEnd < displayStart {
		displayEnd = displayStart
	}

	var launchErrors []string
	for n := displayStart; n <= displayEnd; n++ {
		display := fmt.Sprintf(":%d", n)
		sess, err := launchXvfb(xvfbBin, display, screen)
		if err != nil {
			launchErrors = append(launchErrors, fmt.Sprintf("%s: %v", display, err))
			continue
		}
		if waitForXDisplay(display, 4*time.Second) {
			return sess, nil
		}
		stderr := strings.TrimSpace(sess.stderr.String())
		if stderr == "" {
			stderr = "display did not become ready"
		}
		launchErrors = append(launchErrors, fmt.Sprintf("%s: %s", display, stderr))
		sess.stop()
	}

	return nil, fmt.Errorf("unable to start Xvfb on displays %d-%d (%s)", displayStart, displayEnd, strings.Join(launchErrors, "; "))
}

func launchXvfb(xvfbBin, display, screen string) (*xvfbSession, error) {
	sess := &xvfbSession{display: display, waitCh: make(chan error, 1)}
	sess.cmd = exec.Command(xvfbBin, display, "-screen", "0", screen, "-nolisten", "tcp", "+extension", "GLX", "+render", "-ac")
	sess.cmd.Stdout = io.Discard
	sess.cmd.Stderr = &sess.stderr
	if err := sess.cmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		sess.waitCh <- sess.cmd.Wait()
	}()
	return sess, nil
}

func waitForXDisplay(display string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if xDisplayReady(display) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return xDisplayReady(display)
}

func xDisplayReady(display string) bool {
	if xdpyinfoBin, err := exec.LookPath("xdpyinfo"); err == nil {
		cmd := exec.Command(xdpyinfoBin, "-display", display)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return cmd.Run() == nil
	}
	displayNum, ok := parseDisplayNumber(display)
	if !ok {
		return false
	}
	socket := filepath.Join("/tmp/.X11-unix", "X"+strconv.Itoa(displayNum))
	if _, err := os.Stat(socket); err != nil {
		return false
	}
	return true
}

func parseDisplayNumber(display string) (int, bool) {
	trimmed := strings.TrimSpace(display)
	if !strings.HasPrefix(trimmed, ":") {
		return 0, false
	}
	num, err := strconv.Atoi(strings.TrimPrefix(trimmed, ":"))
	if err != nil {
		return 0, false
	}
	return num, true
}

func savePNG(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// isImageAllBlack returns true when every pixel in img has R=G=B=A=0.
func isImageAllBlack(img image.Image) bool {
	if img == nil {
		return true
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r != 0 || g != 0 || b != 0 || a != 0 {
				return false
			}
		}
	}
	return true
}

// isXwaylandDisplay returns true when the process is running under Xwayland
// (i.e. WAYLAND_DISPLAY is set and an Xwayland process is present).
func isXwaylandDisplay() bool {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return false
	}
	out, err := exec.Command("pgrep", "-x", "Xwayland").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// checkDisplayBrightness warns when an xrandr output has brightness 0.0.
// Under Xwayland this is a NOTE only — Canvas.Capture() is not affected by
// xrandr brightness because it reads from the GL framebuffer, not from X11.
// Under native X11 a brightness-0.0 output causes external capture tools
// (maim, scrot, xwd) to return black images.
func checkDisplayBrightness() {
	onXwayland := isXwaylandDisplay()
	outputs := parseXrandrBrightness()
	for _, o := range outputs {
		if o.connected && o.brightness == 0.0 {
			if onXwayland {
				fmt.Fprintf(os.Stderr,
					"NOTE: Output %s has xrandr brightness 0.0 (Xwayland session)."+
						" This does NOT affect Canvas.Capture() — it only affects external X11"+
						" capture tools (maim/scrot/xwd), which always produce black images on"+
						" Xwayland regardless of brightness. Use grim(1) for external window"+
						" capture on Wayland: grim -g \"$(slurp)\" /tmp/capture.png\n", o.name)
			} else {
				if trySetOutputBrightness(o.name, 1.0) {
					fmt.Fprintf(os.Stderr,
						"NOTE: Output %s brightness was 0.0; auto-set to 1.0 for capture stability.\n",
						o.name)
				} else {
					fmt.Fprintf(os.Stderr,
						"WARNING: Output %s has brightness 0.0 — captures via external X11 tools"+
							" may be black. Run: xrandr --output %s --brightness 1.0\n",
						o.name, o.name)
				}
			}
		}
	}
}

// captureBlackDiagnostic returns an environment-aware diagnostic string
// explaining why Canvas().Capture() might have returned an all-black image.
func captureBlackDiagnostic() string {
	var sb strings.Builder
	display := os.Getenv("DISPLAY")
	wayland := os.Getenv("WAYLAND_DISPLAY")
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	sb.WriteString("  Diagnostic:\n")
	sb.WriteString(fmt.Sprintf("    DISPLAY=%s  WAYLAND_DISPLAY=%s  XDG_SESSION_TYPE=%s\n",
		display, wayland, sessionType))
	if isXwaylandDisplay() {
		sb.WriteString("    Session type: Xwayland (X11-over-Wayland)\n")
		sb.WriteString("    Canvas.Capture() uses glReadPixels(GL_FRONT) on the application's own\n")
		sb.WriteString("    OpenGL context and is NOT affected by Xwayland root-window limitations.\n")
		sb.WriteString("    A black result here likely means one of:\n")
		sb.WriteString("      1. The window was not fully rendered before capture (timing issue).\n")
		sb.WriteString("         Try increasing the -settle duration (e.g. -settle 3s).\n")
		sb.WriteString("      2. Mesa GLX on Xwayland uses exchange-swap, leaving the GL front\n")
		sb.WriteString("         buffer undefined immediately after the first SwapBuffers call.\n")
		sb.WriteString("         The extra Refresh+sleep retry above should mitigate this.\n")
		sb.WriteString("    Note: External tools (maim, scrot, xwd) ALWAYS produce black images\n")
		sb.WriteString("    on Xwayland — this is a known architectural limitation of Xwayland.\n")
		sb.WriteString("    Use grim(1) for external capture: grim -g \"$(slurp)\" /tmp/cap.png\n")
	} else {
		sb.WriteString("    Session type: native X11\n")
		outputs := parseXrandrBrightness()
		for _, o := range outputs {
			if o.connected && o.brightness == 0.0 {
				sb.WriteString(fmt.Sprintf("    Output %s has brightness 0.0 — run with xrandr --output %s --brightness 1.0\n",
					o.name, o.name))
			}
		}
		sb.WriteString("    If display brightness is fine, try increasing the -settle duration.\n")
	}
	return sb.String()
}

func trySetOutputBrightness(output string, brightness float64) bool {
	if strings.TrimSpace(output) == "" {
		return false
	}
	if _, err := exec.LookPath("xrandr"); err != nil {
		return false
	}
	cmd := exec.Command("xrandr", "--output", output, "--brightness", fmt.Sprintf("%.1f", brightness))
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// xrandrOutput holds parsed state for one xrandr output.
type xrandrOutput struct {
	name       string
	connected  bool
	brightness float64
}

// parseXrandrBrightness runs xrandr --verbose and parses brightness values.
// Returns an empty slice if xrandr is not available or parsing fails.
func parseXrandrBrightness() []xrandrOutput {
	out, err := exec.Command("xrandr", "--verbose").Output()
	if err != nil {
		return nil
	}
	var outputs []xrandrOutput
	var current *xrandrOutput
	for _, line := range strings.Split(string(out), "\n") {
		// Detect "OUTPUT connected" lines.
		if fields := strings.Fields(line); len(fields) >= 2 {
			if fields[1] == "connected" {
				outputs = append(outputs, xrandrOutput{name: fields[0], connected: true, brightness: 1.0})
				current = &outputs[len(outputs)-1]
				continue
			}
			if fields[1] == "disconnected" {
				outputs = append(outputs, xrandrOutput{name: fields[0], connected: false, brightness: 1.0})
				current = &outputs[len(outputs)-1]
				continue
			}
		}
		// Detect "\tBrightness: 0.00" lines.
		trimmed := strings.TrimSpace(line)
		if current != nil && strings.HasPrefix(trimmed, "Brightness:") {
			parts := strings.Fields(trimmed)
			if len(parts) == 2 {
				var v float64
				if _, err := fmt.Sscanf(parts[1], "%f", &v); err == nil {
					current.brightness = v
				}
			}
		}
	}
	return outputs
}
