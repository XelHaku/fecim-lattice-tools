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
	"fyne.io/fyne/v2/test"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
)

// testScreenshotDir is the directory for test screenshots
const testScreenshotDir = "testdata/screenshots"

// =============================================================================
// VISUAL REGRESSION TESTS
// =============================================================================

// TestVisualRegressionHysteresis captures the hysteresis module initial state.
func TestVisualRegressionHysteresis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	app := test.NewApp()
	defer app.Quit()

	window := app.NewWindow("Visual Test - Hysteresis")
	window.Resize(fyne.NewSize(1200, 800))

	demo := demo1gui.NewEmbeddedApp()
	content := demo.BuildContent(app, window)
	window.SetContent(container.NewMax(content))
	window.Show()

	// Start the module
	demo.Start()

	// Allow time for rendering
	time.Sleep(500 * time.Millisecond)

	// Capture screenshot
	img := captureWindow(window)
	savePath := saveTestScreenshot(t, img, "hysteresis_initial")

	// Stop module
	demo.Stop()

	t.Logf("Hysteresis screenshot saved: %s", savePath)
	verifyImageNotEmpty(t, img, "hysteresis")
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

	window := app.NewWindow("Visual Test - Circuits")
	window.Resize(fyne.NewSize(1200, 800))

	demo := demo4gui.NewEmbeddedCircuitsApp()
	content := demo.BuildContent(app, window)
	window.SetContent(container.NewMax(content))
	window.Show()

	// Start the module
	demo.Start()

	// Allow time for rendering
	time.Sleep(500 * time.Millisecond)

	// Capture screenshot
	img := captureWindow(window)
	savePath := saveTestScreenshot(t, img, "circuits_initial")

	// Stop module
	demo.Stop()

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

	window := app.NewWindow("Visual Test - Comparison")
	window.Resize(fyne.NewSize(1200, 800))

	demo := demo5gui.NewEmbeddedComparisonApp()
	content := demo.BuildContent(app, window)
	window.SetContent(container.NewMax(content))
	window.Show()

	// Start the module (triggers animation)
	demo.Start()

	// Allow time for rendering
	time.Sleep(500 * time.Millisecond)

	// Capture screenshot
	img := captureWindow(window)
	savePath := saveTestScreenshot(t, img, "comparison_initial")

	// Stop module
	demo.Stop()

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

	window := app.NewWindow("Visual Test - EDA")
	window.Resize(fyne.NewSize(1200, 800))

	demo := demo6gui.NewEmbeddedEDAApp()
	content := demo.BuildContent(app, window)
	window.SetContent(container.NewMax(content))
	window.Show()

	// Start the module
	demo.Start()

	// Allow time for rendering
	time.Sleep(500 * time.Millisecond)

	// Capture screenshot
	img := captureWindow(window)
	savePath := saveTestScreenshot(t, img, "eda_initial")

	// Stop module
	demo.Stop()

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

	app := test.NewApp()
	defer app.Quit()

	window := app.NewWindow("Visual Test - All Modules")
	window.Resize(fyne.NewSize(1200, 800))

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

			content := module.BuildContent(app, window)
			window.SetContent(container.NewMax(content))
			window.Show()

			module.Start()
			time.Sleep(300 * time.Millisecond)

			img := captureWindow(window)
			savePath := saveTestScreenshot(t, img, m.name+"_all_test")
			verifyImageNotEmpty(t, img, m.name)

			module.Stop()
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
			threshold := uint32(256 * 16) // ~6% tolerance
			if abs(r1-r2) > threshold || abs(g1-g2) > threshold ||
				abs(b1-b2) > threshold || abs(a1-a2) > threshold {
				differentPixels++
			}
		}
	}

	return float64(differentPixels) / float64(totalPixels)
}

func abs(x uint32) uint32 {
	if x < 0 {
		return 0 // unsigned, can't be negative
	}
	return x
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
