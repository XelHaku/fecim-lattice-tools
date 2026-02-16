// Command fecim-screenshotter captures automated screenshots of the FeCIM GUI.
//
// Usage: fecim-screenshotter -module 1 -output screenshots/
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	flag.Parse()

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

func savePNG(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
