//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
)

// These tests are designed specifically for xvfb headless environments
// Run with: xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run XvfbCrawler

const (
	xvfbCrawlEnvVar = "FECIM_UI_CRAWL"
	xvfbSettleTime  = 800 * time.Millisecond
	xvfbTabDelay    = 200 * time.Millisecond
)

// TestXvfbCrawlerCircuits tests the UI crawler specifically under xvfb with real app driver
func TestXvfbCrawlerCircuits(t *testing.T) {
	if !xvfbCrawlEnabled() {
		t.Skipf("Skipping xvfb crawler: set %s=1 to enable", xvfbCrawlEnvVar)
	}
	if testing.Short() {
		t.Skip("Skipping xvfb crawler in short mode")
	}
	ensureDisplayForGraphicalTests(t)

	module := demo4gui.NewEmbeddedCircuitsApp()
	if module == nil {
		t.Fatalf("Failed to create circuits module")
	}

	sizes := []struct {
		name   string
		width  float32
		height float32
	}{
		{"desktop", 1200, 800},
		{"tablet", 768, 1024},
		{"mobile", 390, 844},
	}

	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			img := captureWithXvfbRealApp(t, module, fyne.NewSize(size.width, size.height), xvfbSettleTime)

			outputDir := filepath.Join(crawlerScreenshotBaseDir, "circuits", "xvfb")
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				t.Fatalf("Failed to create output directory: %v", err)
			}

			filename := filepath.Join(outputDir, fmt.Sprintf("base_%s_%dx%d.png", size.name, int(size.width), int(size.height)))
			if err := saveCrawlerScreenshot(img, filename); err != nil {
				t.Fatalf("Failed to save screenshot: %v", err)
			}

			t.Logf("Circuits %s screenshot saved: %s", size.name, filename)
			verifyImageNotEmpty(t, img, fmt.Sprintf("circuits_%s", size.name))
		})
	}
}

// TestXvfbCrawlerCircuitsDetailed performs detailed UI crawling with xvfb
func TestXvfbCrawlerCircuitsDetailed(t *testing.T) {
	if !xvfbCrawlEnabled() {
		t.Skipf("Skipping detailed xvfb crawler: set %s=1 to enable", xvfbCrawlEnvVar)
	}
	if testing.Short() {
		t.Skip("Skipping detailed xvfb crawler in short mode")
	}
	ensureDisplayForGraphicalTests(t)

	module := demo4gui.NewEmbeddedCircuitsApp()
	if module == nil {
		t.Fatalf("Failed to create circuits module")
	}

	// Run detailed crawling with real app for better compatibility
	runXvfbDetailedCrawl(t, module, "circuits", fyne.NewSize(1200, 800))
}

// runXvfbDetailedCrawl performs comprehensive crawling with real app under xvfb
func runXvfbDetailedCrawl(t *testing.T, module moduleLifecycle, moduleName string, size fyne.Size) {
	t.Helper()

	outputDir := filepath.Join(crawlerScreenshotBaseDir, moduleName, "xvfb-detailed")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	screenshots := make(map[string]image.Image)
	errors := []string{}

	// Create real app for xvfb compatibility
	a := app.New()
	w := a.NewWindow(fmt.Sprintf("XvfbCrawler - %s", moduleName))
	w.Resize(size)

	content := module.BuildContent(a, w)
	w.SetContent(container.NewMax(content))
	w.Show()

	// Let UI stabilize (don't start module to avoid simulation loops)
	time.Sleep(xvfbSettleTime)

	// Capture base state
	baseImg := w.Canvas().Capture()
	if baseImg != nil {
		baseName := fmt.Sprintf("base_%dx%d", int(size.Width), int(size.Height))
		filename := filepath.Join(outputDir, baseName+".png")
		if err := saveCrawlerScreenshot(baseImg, filename); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to save base image: %v", err))
		} else {
			screenshots[baseName] = baseImg
		}
	}

	// Crawl tabs systematically
	tabSets := findAllAppTabs(content)
	for tabSetIdx, tabs := range tabSets {
		originalSelection := tabs.SelectedIndex()

		for tabIdx := 0; tabIdx < len(tabs.Items); tabIdx++ {
			tabs.SelectIndex(tabIdx)
			time.Sleep(xvfbTabDelay)

			tabText := ""
			if tabIdx < len(tabs.Items) {
				tabText = tabs.Items[tabIdx].Text
			}

			img := w.Canvas().Capture()
			if img != nil {
				name := fmt.Sprintf("tab_set%d_idx%d_%s_%dx%d",
					tabSetIdx, tabIdx, safeCrawlerName(tabText), int(size.Width), int(size.Height))
				filename := filepath.Join(outputDir, name+".png")
				if err := saveCrawlerScreenshot(img, filename); err != nil {
					errors = append(errors, fmt.Sprintf("Failed to save tab image %s: %v", name, err))
				} else {
					screenshots[name] = img
				}
			}
		}

		// Restore original selection
		if originalSelection >= 0 && originalSelection < len(tabs.Items) {
			tabs.SelectIndex(originalSelection)
		}
	}

	// Crawl interactive elements
	elements := findInteractiveElements(content)
	triggers := findPopupTriggers(elements)

	for _, trigger := range triggers {
		if triggerInteractiveElement(trigger) {
			time.Sleep(xvfbTabDelay)

			img := w.Canvas().Capture()
			if img != nil {
				name := fmt.Sprintf("trigger_%s_%s_%dx%d",
					trigger.Type, safeCrawlerName(trigger.Text), int(size.Width), int(size.Height))
				filename := filepath.Join(outputDir, name+".png")
				if err := saveCrawlerScreenshot(img, filename); err != nil {
					errors = append(errors, fmt.Sprintf("Failed to save trigger image %s: %v", name, err))
				} else {
					screenshots[name] = img
				}
			}

			// Try to close any opened dialogs/overlays
			closeElements := findCloseElements(findInteractiveElements(content))
			for _, closer := range closeElements {
				if triggerInteractiveElement(closer) {
					break
				}
			}
			time.Sleep(xvfbTabDelay)
		}
	}

	// Cleanup
	w.Close()
	a.Quit()

	// Report results
	if len(errors) > 0 {
		t.Logf("Xvfb crawler collected %d warnings/errors:", len(errors))
		for i, err := range errors {
			t.Logf("  %d: %s", i+1, err)
		}
	}

	t.Logf("Successfully captured %d screenshots for module %s under xvfb", len(screenshots), moduleName)
}

// captureWithXvfbRealApp captures a screenshot using real app under xvfb
func captureWithXvfbRealApp(t *testing.T, module moduleLifecycle, size fyne.Size, settle time.Duration) image.Image {
	t.Helper()

	// Use the existing captureWithRealApp function but ensure it works with xvfb
	var img image.Image
	var err error

	// Ensure we're on the main goroutine for GLFW compatibility
	runOnMainGoroutine(func() {
		img, err = captureWithXvfbOnMain(module, size, settle)
	})

	if err != nil {
		t.Fatalf("Failed to capture screenshot with xvfb real app: %v", err)
	}

	return img
}

// captureWithXvfbOnMain performs capture on main goroutine (required for GLFW)
func captureWithXvfbOnMain(module moduleLifecycle, size fyne.Size, settle time.Duration) (image.Image, error) {
	a := app.New()
	defer a.Quit()

	w := a.NewWindow("XvfbCapture")
	w.Resize(size)

	content := module.BuildContent(a, w)
	w.SetContent(container.NewMax(content))
	w.Show()

	// Don't start module to avoid simulation loops
	time.Sleep(settle)

	return w.Canvas().Capture(), nil
}

// Helper functions

func xvfbCrawlEnabled() bool {
	return uiCrawlEnabled() // Reuse the same environment variable
}

// Reuse existing verifyImageNotEmpty from e2e_visual_test.go
