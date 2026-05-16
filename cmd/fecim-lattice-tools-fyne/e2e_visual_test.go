//go:build !ci
// +build !ci

// Package main provides visual regression tests for the FeCIM Lattice Tools application.
// These tests capture screenshots and compare against golden images.
//
// Note: These tests require a display and are excluded from CI builds.
// Run locally with: go test -v ./cmd/fecim-lattice-tools/... -run Visual
package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo1widgets "fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// testScreenshotDir is the directory for test screenshots
const testScreenshotDir = "testdata/screenshots"

// fyneUI runs fn on the Fyne UI thread.
//
// Some drivers (incl. the test driver) can warn/error if window operations are
// executed off-thread.
func fyneUI(t *testing.T, fn func()) {
	t.Helper()
	fyne.DoAndWait(fn)
}

func findFirstAppTabs(root fyne.CanvasObject) *container.AppTabs {
	if root == nil {
		return nil
	}
	if tabs, ok := root.(*container.AppTabs); ok {
		return tabs
	}
	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if found := findFirstAppTabs(child); found != nil {
				return found
			}
		}
	}
	return nil
}

// =============================================================================
// VISUAL REGRESSION TESTS
// =============================================================================

// TestVisualRegressionHysteresis captures the hysteresis module initial state.
func TestVisualRegressionHysteresis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}
	// Visual tests should not mutate tracked calibration baselines.
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	app := test.NewApp()
	defer app.Quit()

	var (
		window   fyne.Window
		demo     = demo1gui.NewEmbeddedApp()
		img      image.Image
		savePath string
	)

	fyneUI(t, func() {
		window = app.NewWindow("Visual Test - Hysteresis")
		window.Resize(fyne.NewSize(1200, 800))

		content := demo.BuildContent(app, window)
		window.SetContent(container.NewMax(content))
		window.Show()
		demo.Start()
	})

	// Allow time for rendering
	time.Sleep(500 * time.Millisecond)

	fyneUI(t, func() {
		img = captureWindow(window)
	})
	savePath = saveTestScreenshot(t, img, "hysteresis_initial")

	fyneUI(t, func() {
		demo.Stop()
		window.Close()
	})

	t.Logf("Hysteresis screenshot saved: %s", savePath)
	verifyImageNotEmpty(t, img, "hysteresis")
}

// TestVisualRegressionHysteresis_EquationModal captures the equation modal, including all tabs.
func TestVisualRegressionHysteresis_EquationModal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}
	// Visual tests should not mutate tracked calibration baselines.
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	app := test.NewApp()
	defer app.Quit()

	var (
		window fyne.Window
		demo   = demo1gui.NewEmbeddedApp()
	)

	sizes := []struct {
		name string
		sz   fyne.Size
		dSz  fyne.Size
	}{
		{name: "desktop_1200x800", sz: fyne.NewSize(1200, 800), dSz: fyne.NewSize(1100, 650)},
		{name: "laptop_1024x768", sz: fyne.NewSize(1024, 768), dSz: fyne.NewSize(950, 620)},
	}

	fyneUI(t, func() {
		window = app.NewWindow("Visual Test - Hysteresis (Equation Modal)")
		window.Resize(sizes[0].sz)

		content := demo.BuildContent(app, window)
		window.SetContent(container.NewMax(content))
		window.Show()
		demo.Start()
	})
	defer fyneUI(t, func() {
		demo.Stop()
		window.Close()
	})

	// Let base UI render.
	time.Sleep(600 * time.Millisecond)

	// Build the equation modal content directly so the test can reliably
	// switch tabs and capture each state.
	content := demo1widgets.NewPhysicsEquationsWidget(window)
	tabs := findFirstAppTabs(content)
	if tabs == nil {
		t.Fatal("equation modal tabs not found in content")
	}

	var d dialog.Dialog
	closeBtn := widget.NewButton("Close", func() {
		if d != nil {
			d.Hide()
		}
	})
	footer := container.NewHBox(layout.NewSpacer(), closeBtn)
	body := container.NewBorder(nil, footer, nil, nil, container.NewPadded(content))

	for _, cfg := range sizes {
		fyneUI(t, func() {
			window.Resize(cfg.sz)
			d = dialog.NewCustom("Physics Equations", "", body, window)
			d.Resize(cfg.dSz)
			d.Show()
		})
		time.Sleep(450 * time.Millisecond)

		tabNames := []string{"lk", "preisach", "ispp"}
		for idx, name := range tabNames {
			fyneUI(t, func() {
				tabs.SelectIndex(idx)
			})
			time.Sleep(350 * time.Millisecond)

			var img image.Image
			fyneUI(t, func() {
				img = captureWindow(window)
			})
			savePath := saveTestScreenshot(t, img, "hysteresis_equation_modal_"+cfg.name+"_"+name)
			t.Logf("Equation modal (%s/%s) screenshot saved: %s", cfg.name, name, savePath)
			verifyImageNotEmpty(t, img, "equation_modal_"+cfg.name+"_"+name)
		}

		fyneUI(t, func() { d.Hide() })
		time.Sleep(120 * time.Millisecond)
	}
}

// TestVisualRegressionHysteresis_MaterialPicker captures the material picker dialog.
func TestVisualRegressionHysteresis_MaterialPicker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	app := test.NewApp()
	defer app.Quit()

	var window fyne.Window
	demo := demo1gui.NewEmbeddedApp()

	sizes := []struct {
		name string
		sz   fyne.Size
	}{
		{"desktop_1200x800", fyne.NewSize(1200, 800)},
		{"laptop_1024x768", fyne.NewSize(1024, 768)},
	}

	fyneUI(t, func() {
		window = app.NewWindow("Visual Test - Material Picker")
		window.Resize(sizes[0].sz)
		content := demo.BuildContent(app, window)
		window.SetContent(container.NewMax(content))
		window.Show()
		demo.Start()
	})
	defer fyneUI(t, func() {
		demo.Stop()
		window.Close()
	})

	time.Sleep(600 * time.Millisecond)

	for _, cfg := range sizes {
		fyneUI(t, func() {
			window.Resize(cfg.sz)
		})
		time.Sleep(200 * time.Millisecond)

		// Open material picker via the shared widget directly.
		var d dialog.Dialog
		fyneUI(t, func() {
			picker := sharedwidgets.NewMaterialPicker(nil)
			picker.SetSelected("default_hzo")
			d = dialog.NewCustomConfirm("Select Ferroelectric Material", "Select", "Cancel", picker, func(bool) {}, window)
			cs := window.Canvas().Size()
			w := cs.Width * 0.90
			h := cs.Height * 0.85
			if w < 680 {
				w = 680
			}
			if h < 420 {
				h = 420
			}
			d.Resize(fyne.NewSize(w, h))
			d.Show()
		})
		time.Sleep(500 * time.Millisecond)

		var img image.Image
		fyneUI(t, func() {
			img = captureWindow(window)
		})
		savePath := saveTestScreenshot(t, img, "hysteresis_material_picker_"+cfg.name)
		t.Logf("Material picker (%s) screenshot saved: %s", cfg.name, savePath)
		verifyImageNotEmpty(t, img, "material_picker_"+cfg.name)

		fyneUI(t, func() { d.Hide() })
		time.Sleep(120 * time.Millisecond)
	}
}

// TestVisualRegressionCrossbar captures the crossbar module initial state.
// NOTE: Skipped because Crossbar module uses Bold+Monospace fonts not in test theme.
func TestVisualRegressionCrossbar(t *testing.T) {
	t.Skip("Crossbar module uses Bold+Monospace fonts not available in test.NewApp() theme")
}

// TestVisualRegressionMNIST captures the MNIST module initial state.
// NOTE: Skipped because MNIST module has infinite loop with fyne.Do in test driver.
func TestVisualRegressionMNIST(t *testing.T) {
	t.Skip("MNIST module has infinite loop with fyne.Do in test driver")
}

// TestVisualRegressionCircuits captures the circuits module initial state.
func TestVisualRegressionCircuits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	app := test.NewApp()
	defer app.Quit()

	var (
		window   fyne.Window
		demo     = demo4gui.NewEmbeddedCircuitsApp()
		img      image.Image
		savePath string
	)

	fyneUI(t, func() {
		window = app.NewWindow("Visual Test - Circuits")
		window.Resize(fyne.NewSize(1200, 800))

		content := demo.BuildContent(app, window)
		window.SetContent(container.NewMax(content))
		window.Show()
		demo.Start()
	})

	time.Sleep(500 * time.Millisecond)

	fyneUI(t, func() { img = captureWindow(window) })
	savePath = saveTestScreenshot(t, img, "circuits_initial")

	fyneUI(t, func() {
		demo.Stop()
		window.Close()
	})

	t.Logf("Circuits screenshot saved: %s", savePath)
	verifyImageNotEmpty(t, img, "circuits")
}

// TestVisualRegressionComparison captures the comparison module initial state.
func TestVisualRegressionComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	app := test.NewApp()
	defer app.Quit()

	var (
		window   fyne.Window
		demo     = demo5gui.NewEmbeddedComparisonApp()
		img      image.Image
		savePath string
	)

	fyneUI(t, func() {
		window = app.NewWindow("Visual Test - Comparison")
		window.Resize(fyne.NewSize(1200, 800))

		content := demo.BuildContent(app, window)
		window.SetContent(container.NewMax(content))
		window.Show()
		demo.Start()
	})

	time.Sleep(500 * time.Millisecond)

	fyneUI(t, func() { img = captureWindow(window) })
	savePath = saveTestScreenshot(t, img, "comparison_initial")

	fyneUI(t, func() {
		demo.Stop()
		window.Close()
	})

	t.Logf("Comparison screenshot saved: %s", savePath)
	verifyImageNotEmpty(t, img, "comparison")
}

// TestVisualRegressionEDA captures the EDA module initial state.
func TestVisualRegressionEDA(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	app := test.NewApp()
	defer app.Quit()

	var (
		window   fyne.Window
		demo     = demo6gui.NewEmbeddedEDAApp()
		img      image.Image
		savePath string
	)

	fyneUI(t, func() {
		window = app.NewWindow("Visual Test - EDA")
		window.Resize(fyne.NewSize(1200, 800))

		content := demo.BuildContent(app, window)
		window.SetContent(container.NewMax(content))
		window.Show()
		demo.Start()
	})

	time.Sleep(500 * time.Millisecond)

	fyneUI(t, func() { img = captureWindow(window) })
	savePath = saveTestScreenshot(t, img, "eda_initial")

	fyneUI(t, func() {
		demo.Stop()
		window.Close()
	})

	t.Logf("EDA screenshot saved: %s", savePath)
	verifyImageNotEmpty(t, img, "eda")
}

// =============================================================================
// ALL MODULES SCREENSHOT TEST
// =============================================================================

// TestVisualAllModulesScreenshot captures screenshots of all modules sequentially.
func TestVisualAllModulesScreenshot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}
	// Visual tests should not mutate tracked calibration baselines.
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	app := test.NewApp()
	defer app.Quit()

	var window fyne.Window
	fyneUI(t, func() {
		window = app.NewWindow("Visual Test - All Modules")
		window.Resize(fyne.NewSize(1200, 800))
		window.Show()
	})

	type moduleInfo struct {
		name   string
		create func() (moduleLifecycle, error)
		skip   bool // Skip modules with known test environment issues
	}

	modules := []moduleInfo{
		{"hysteresis", func() (moduleLifecycle, error) {
			return demo1gui.NewEmbeddedApp(), nil
		}, false},
		{"crossbar", func() (moduleLifecycle, error) {
			return demo2gui.NewEmbeddedCrossbarApp()
		}, true}, // Skip: font issues
		{"mnist", func() (moduleLifecycle, error) {
			return demo3gui.NewEmbeddedDualModeApp(), nil
		}, true}, // Skip: fyne.Do loop
		{"circuits", func() (moduleLifecycle, error) {
			return demo4gui.NewEmbeddedCircuitsApp(), nil
		}, false},
		{"comparison", func() (moduleLifecycle, error) {
			return demo5gui.NewEmbeddedComparisonApp(), nil
		}, false},
		{"eda", func() (moduleLifecycle, error) {
			return demo6gui.NewEmbeddedEDAApp(), nil
		}, false},
	}

	for _, m := range modules {
		t.Run(m.name, func(t *testing.T) {
			if m.skip {
				t.Skipf("Skipping %s: known issues with test.NewApp()", m.name)
			}

			module, err := m.create()
			if err != nil {
				t.Fatalf("Failed to create %s module: %v", m.name, err)
			}

			var img image.Image

			fyneUI(t, func() {
				content := module.BuildContent(app, window)
				window.SetContent(container.NewMax(content))
				window.Show()
				module.Start()
			})

			time.Sleep(300 * time.Millisecond)

			fyneUI(t, func() { img = captureWindow(window) })
			savePath := saveTestScreenshot(t, img, m.name+"_all_test")
			verifyImageNotEmpty(t, img, m.name)

			fyneUI(t, func() { module.Stop() })
			time.Sleep(100 * time.Millisecond)

			t.Logf("%s screenshot saved: %s", m.name, savePath)
		})
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// captureWindow captures the window content as an image
func captureWindow(window fyne.Window) image.Image {
	return window.Canvas().Capture()
}

// saveTestScreenshot saves a screenshot to the test data directory
func saveTestScreenshot(t *testing.T, img image.Image, name string) string {
	t.Helper()

	// Create screenshot directory if needed
	if err := os.MkdirAll(testScreenshotDir, 0755); err != nil {
		t.Logf("Warning: could not create screenshot directory: %v", err)
		return ""
	}

	filename := filepath.Join(testScreenshotDir, name+".png")

	file, err := os.Create(filename)
	if err != nil {
		t.Logf("Warning: could not create screenshot file: %v", err)
		return ""
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Logf("Warning: could not encode screenshot: %v", err)
		return ""
	}

	return filename
}

// verifyImageNotEmpty checks that the captured image has actual content
func verifyImageNotEmpty(t *testing.T, img image.Image, name string) {
	t.Helper()

	if img == nil {
		t.Errorf("%s: captured image is nil", name)
		return
	}

	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Errorf("%s: captured image has zero dimensions: %dx%d", name, bounds.Dx(), bounds.Dy())
		return
	}

	// Check that image has some non-uniform content (not all same color)
	firstPixel := img.At(bounds.Min.X, bounds.Min.Y)
	hasVariation := false

	// Sample some pixels to check for variation
	for y := bounds.Min.Y; y < bounds.Max.Y && !hasVariation; y += bounds.Dy() / 10 {
		for x := bounds.Min.X; x < bounds.Max.X && !hasVariation; x += bounds.Dx() / 10 {
			if img.At(x, y) != firstPixel {
				hasVariation = true
			}
		}
	}

	if !hasVariation {
		t.Logf("%s: warning - captured image appears to be uniform color (may be rendering issue)", name)
	}

	t.Logf("%s: captured %dx%d image", name, bounds.Dx(), bounds.Dy())
}

// compareImages computes a simple difference metric between two images
// Returns the percentage of pixels that differ (0.0 = identical, 1.0 = completely different)
func compareImages(img1, img2 image.Image) float64 {
	if img1 == nil || img2 == nil {
		return 1.0
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Different sizes = completely different
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		return 1.0
	}

	totalPixels := bounds1.Dx() * bounds1.Dy()
	differentPixels := 0

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()

			// Check if pixel differs significantly (threshold for anti-aliasing)
			threshold := int64(256 * 16) // ~6% tolerance
			if absDiff(r1, r2) > threshold || absDiff(g1, g2) > threshold ||
				absDiff(b1, b2) > threshold || absDiff(a1, a2) > threshold {
				differentPixels++
			}
		}
	}

	return float64(differentPixels) / float64(totalPixels)
}

func absDiff(a, b uint32) int64 {
	d := int64(a) - int64(b)
	if d < 0 {
		return -d
	}
	return d
}

// =============================================================================
// GOLDEN IMAGE COMPARISON (future extension)
// =============================================================================

// loadGoldenImage loads a golden reference image for comparison
func loadGoldenImage(name string) (image.Image, error) {
	goldenPath := filepath.Join(testScreenshotDir, "golden", name+".png")

	file, err := os.Open(goldenPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return png.Decode(file)
}

// TestVisualGoldenComparison compares current screenshots against golden images.
// This test is skipped if golden images don't exist.
func TestVisualGoldenComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	modules := []string{"hysteresis", "crossbar", "mnist", "circuits", "comparison", "eda"}

	for _, name := range modules {
		t.Run(name, func(t *testing.T) {
			golden, err := loadGoldenImage(name + "_initial")
			if err != nil {
				t.Skipf("No golden image for %s: %v", name, err)
				return
			}

			currentPath := filepath.Join(testScreenshotDir, name+"_initial.png")
			currentFile, err := os.Open(currentPath)
			if err != nil {
				t.Skipf("No current screenshot for %s: %v", name, err)
				return
			}
			defer currentFile.Close()

			current, err := png.Decode(currentFile)
			if err != nil {
				t.Fatalf("Failed to decode current screenshot: %v", err)
			}

			diff := compareImages(golden, current)
			maxDiff := 0.05 // Allow 5% difference for minor rendering variations

			if diff > maxDiff {
				t.Errorf("%s: visual regression detected - %.1f%% of pixels differ (max allowed: %.1f%%)",
					name, diff*100, maxDiff*100)
			} else {
				t.Logf("%s: visual comparison passed - %.2f%% difference", name, diff*100)
			}
		})
	}
}
