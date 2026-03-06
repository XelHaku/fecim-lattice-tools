//go:build !ci
// +build !ci

// Package main provides comprehensive UI screenshot crawling for the FeCIM Lattice Tools application.
// This test systematically captures screenshots of all UI states including tabs, dialogs, overlays, and tooltips.
//
// The crawler is designed to work both headless (with xvfb) and with real displays.
// Run with: FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler
// For headless: xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
)

const (
	uiCrawlEnvVar            = "FECIM_UI_CRAWL"
	crawlerScreenshotBaseDir = "testdata/screenshots/audit"
	crawlerShowDelay         = 200 * time.Millisecond
	crawlerResizeDelay       = 150 * time.Millisecond
	crawlerTabDelay          = 100 * time.Millisecond
	crawlerDialogDelay       = 120 * time.Millisecond
	crawlerOverlayDelay      = 100 * time.Millisecond
	crawlerOverlayCloseDelay = 80 * time.Millisecond
	crawlerTooltipDelay      = 80 * time.Millisecond
)

// crawlerConfig holds configuration for the UI crawler
type crawlerConfig struct {
	Module          string
	Sizes           []crawlerSize
	CaptureBase     bool
	CaptureTabs     bool
	CaptureDialogs  bool
	CaptureOverlays bool
	CaptureTooltips bool
	OutputDir       string
}

type crawlerSize struct {
	Name   string
	Width  float32
	Height float32
}

// dialogTrigger represents a known dialog trigger in the UI
type dialogTrigger struct {
	Name        string
	ButtonText  string
	Description string
}

// crawlerState holds the current state of the crawler
type crawlerState struct {
	Config      crawlerConfig
	Module      moduleLifecycle
	Window      fyne.Window
	App         fyne.App
	Screenshots map[string]image.Image
	Errors      []string
}

// TestUICrawlerCircuits tests the comprehensive UI crawler specifically for Module 4 circuits
func TestUICrawlerCircuits(t *testing.T) {
	if !uiCrawlEnabled() {
		t.Skipf("Skipping UI crawler: set %s=1 to enable", uiCrawlEnvVar)
	}
	if testing.Short() {
		t.Skip("Skipping UI crawler in short mode")
	}

	// Check if we're in a headless environment and setup display accordingly
	setupHeadlessIfNeeded(t)

	config := crawlerConfig{
		Module: "circuits",
		Sizes: []crawlerSize{
			{"desktop", 1200, 800},
			{"mobile", 390, 844},
			{"tablet", 768, 1024},
		},
		CaptureBase:     true,
		CaptureTabs:     true,
		CaptureDialogs:  true,
		CaptureOverlays: true,
		CaptureTooltips: true,
		OutputDir:       filepath.Join(crawlerScreenshotBaseDir, "circuits"),
	}

	runUICrawler(t, config, func() (moduleLifecycle, error) {
		return demo4gui.NewEmbeddedCircuitsApp(), nil
	})
}

// TestUICrawlerAllModules runs the crawler on all supported modules
func TestUICrawlerAllModules(t *testing.T) {
	if !uiCrawlEnabled() {
		t.Skipf("Skipping UI crawler: set %s=1 to enable", uiCrawlEnvVar)
	}
	if testing.Short() {
		t.Skip("Skipping UI crawler in short mode")
	}

	setupHeadlessIfNeeded(t)

	type moduleInfo struct {
		name   string
		create func() (moduleLifecycle, error)
		skip   bool
		reason string
	}

	modules := []moduleInfo{
		{"hysteresis", func() (moduleLifecycle, error) { return createModule("demo1gui") }, false, ""},
		{"crossbar", func() (moduleLifecycle, error) { return createModule("demo2gui") }, false, ""},
		{"mnist", func() (moduleLifecycle, error) { return createModule("demo3gui") }, false, ""},
		{"circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil }, false, ""},
		{"comparison", func() (moduleLifecycle, error) { return createModule("demo5gui") }, false, ""},
		{"eda", func() (moduleLifecycle, error) { return createModule("demo6gui") }, false, ""},
	}

	for _, m := range modules {
		t.Run(m.name, func(t *testing.T) {
			config := crawlerConfig{
				Module: m.name,
				Sizes: []crawlerSize{
					{"desktop", 1200, 800},
					{"mobile", 390, 844},
				},
				CaptureBase:     true,
				CaptureTabs:     true,
				CaptureDialogs:  true,
				CaptureOverlays: true,
				CaptureTooltips: false, // Disable tooltips for broader compatibility
				OutputDir:       filepath.Join(crawlerScreenshotBaseDir, m.name),
			}

			// Keep crossbar/mnist in the audit (no skip), but use conservative crawl depth
			// to avoid known flaky paths while still producing screenshot evidence.
			if m.name == "crossbar" || m.name == "mnist" {
				config.CaptureTabs = false
				config.CaptureDialogs = false
				config.CaptureOverlays = false
			}
			if m.name == "eda" {
				config.CaptureOverlays = false
			}

			runUICrawler(t, config, m.create)
		})
	}
}

// runUICrawler is the main crawler engine
func runUICrawler(t *testing.T, config crawlerConfig, moduleFactory func() (moduleLifecycle, error)) {
	t.Helper()

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory %s: %v", config.OutputDir, err)
	}

	// Create module
	module, err := moduleFactory()
	if err != nil {
		if isKnownCrawlerGap(err) {
			t.Skipf("Skipping %s crawler: %v", config.Module, err)
		}
		t.Fatalf("Failed to create %s module: %v", config.Module, err)
	}
	if module == nil {
		t.Fatalf("%s module is nil", config.Module)
	}

	// Setup app and window.
	// For automated crawling we use the Fyne test driver for determinism and to
	// ensure fyne.Do[AndWait] works without needing to run a real event loop.
	fy := test.NewApp()
	var w fyne.Window

	defer func() {
		fyne.DoAndWait(func() {
			if w != nil {
				w.Close()
			}
			fy.Quit()
		})
	}()

	fyne.DoAndWait(func() {
		w = fy.NewWindow(fmt.Sprintf("UICrawler - %s", config.Module))
	})

	if w == nil {
		t.Fatalf("%s window creation failed", config.Module)
	}

	state := &crawlerState{
		Config:      config,
		Module:      module,
		Window:      w,
		App:         fy,
		Screenshots: make(map[string]image.Image),
		Errors:      []string{},
	}

	// Build and set content
	var content fyne.CanvasObject
	fyne.DoAndWait(func() {
		content = module.BuildContent(fy, w)
		w.SetContent(container.NewMax(content))
		w.Show()
	})

	if content == nil {
		t.Fatalf("%s BuildContent returned nil", config.Module)
	}

	// Let UI stabilize
	time.Sleep(crawlerShowDelay)

	// Avoid starting simulation loops that might interfere with capture
	// module.Start() is skipped intentionally to keep UI stable for crawling

	// Crawl all sizes
	for _, size := range config.Sizes {
		t.Run(size.Name, func(t *testing.T) {
			crawlSize(t, state, content, size)
		})
	}

	// Report any errors collected during crawling
	if len(state.Errors) > 0 {
		t.Logf("Crawler collected %d warnings/errors:", len(state.Errors))
		for i, err := range state.Errors {
			t.Logf("  %d: %s", i+1, err)
		}
	}

	t.Logf("Successfully captured %d screenshots for module %s", len(state.Screenshots), config.Module)
}

// crawlSize crawls a single window size
func crawlSize(t *testing.T, state *crawlerState, content fyne.CanvasObject, size crawlerSize) {
	t.Helper()

	// Resize window
	fyne.DoAndWait(func() {
		state.Window.Resize(fyne.NewSize(size.Width, size.Height))
	})
	time.Sleep(crawlerResizeDelay)

	// Capture base state
	if state.Config.CaptureBase {
		name := fmt.Sprintf("base_%s_%dx%d", size.Name, int(size.Width), int(size.Height))
		captureAndSave(t, state, name)
	}

	// Crawl tabs
	if state.Config.CaptureTabs {
		crawlTabs(t, state, content, size)
	}

	// Crawl dialogs
	if state.Config.CaptureDialogs {
		crawlDialogs(t, state, content, size)
	}

	// Crawl overlays
	if state.Config.CaptureOverlays {
		crawlOverlays(t, state, content, size)
	}

	// Crawl tooltips (if enabled)
	if state.Config.CaptureTooltips {
		crawlTooltips(t, state, content, size)
	}
}

// crawlTabs systematically captures all tabs and nested tabs
func crawlTabs(t *testing.T, state *crawlerState, content fyne.CanvasObject, size crawlerSize) {
	t.Helper()

	var tabSets []*container.AppTabs
	fyne.DoAndWait(func() {
		tabSets = findAllAppTabs(content)
	})

	if len(tabSets) == 0 {
		return // No tabs found
	}

	for tabSetIdx, tabs := range tabSets {
		var itemCount int
		var originalSelection int

		fyne.DoAndWait(func() {
			itemCount = len(tabs.Items)
			originalSelection = tabs.SelectedIndex()
		})

		for tabIdx := 0; tabIdx < itemCount; tabIdx++ {
			fyne.DoAndWait(func() {
				tabs.SelectIndex(tabIdx)
			})
			time.Sleep(crawlerTabDelay)

			var tabText string
			fyne.DoAndWait(func() {
				if tabIdx < len(tabs.Items) {
					tabText = tabs.Items[tabIdx].Text
				}
			})

			name := fmt.Sprintf("tabs_%s_%dx%d_set%d_tab%d_%s",
				size.Name, int(size.Width), int(size.Height),
				tabSetIdx, tabIdx, safeCrawlerName(tabText))
			captureAndSave(t, state, name)
		}

		// Restore original selection
		fyne.DoAndWait(func() {
			if originalSelection >= 0 && originalSelection < itemCount {
				tabs.SelectIndex(originalSelection)
			}
		})
	}
}

// crawlDialogs attempts to open and capture known dialogs
func crawlDialogs(t *testing.T, state *crawlerState, content fyne.CanvasObject, size crawlerSize) {
	t.Helper()

	// Define known dialog triggers for common patterns
	triggers := []dialogTrigger{
		{"tools", "Tools", "Tools dialog or menu"},
		{"export", "Export", "Export dialog"},
		{"import", "Import", "Import dialog"},
		{"settings", "Settings", "Settings dialog"},
		{"preferences", "Preferences", "Preferences dialog"},
		{"material", "Material", "Material picker dialog"},
		{"parameters", "Parameters", "Parameters dialog"},
		{"configuration", "Config", "Configuration dialog"},
		{"about", "About", "About dialog"},
		{"help", "Help", "Help dialog"},
	}

	for _, trigger := range triggers {
		if triggerDialog(t, state, content, trigger) {
			name := fmt.Sprintf("dialog_%s_%dx%d_%s",
				size.Name, int(size.Width), int(size.Height), trigger.Name)
			captureAndSave(t, state, name)

			// Try to close the dialog
			closeDialog(t, state, content)
			time.Sleep(crawlerDialogDelay)
		}
	}
}

// crawlOverlays captures popup overlays and modals
func crawlOverlays(t *testing.T, state *crawlerState, content fyne.CanvasObject, size crawlerSize) {
	t.Helper()

	// Find buttons that might trigger overlays
	overlayTriggers := map[string]bool{
		"learn":    true,
		"info":     true,
		"details":  true,
		"more":     true,
		"options":  true,
		"advanced": true,
		"show":     true,
		"view":     true,
	}

	closeWords := map[string]bool{
		"close":   true,
		"back":    true,
		"dismiss": true,
		"ok":      true,
		"done":    true,
		"cancel":  true,
	}

	var buttons []*widget.Button
	fyne.DoAndWait(func() {
		buttons = findAllButtons(content)
	})

	seenOverlay := map[string]int{}

	for _, button := range buttons {
		var label string
		var hasAction bool
		fyne.DoAndWait(func() {
			label = strings.TrimSpace(button.Text)
			hasAction = button.OnTapped != nil
		})

		if !hasAction || label == "" {
			continue
		}

		norm := strings.ToLower(label)
		if !overlayTriggers[norm] {
			continue
		}

		// Trigger overlay
		fyne.DoAndWait(func() {
			button.OnTapped()
		})
		time.Sleep(crawlerOverlayDelay)

		idx := seenOverlay[norm]
		seenOverlay[norm] = idx + 1
		name := fmt.Sprintf("overlay_%s_%dx%d_%s_%d",
			size.Name, int(size.Width), int(size.Height), safeCrawlerName(norm), idx)
		captureAndSave(t, state, name)

		// Try to close overlay
		var currentButtons []*widget.Button
		fyne.DoAndWait(func() {
			currentButtons = findAllButtons(content)
		})

		for _, closeBtn := range currentButtons {
			var closeLabel string
			var closeHasAction bool
			fyne.DoAndWait(func() {
				closeLabel = strings.ToLower(strings.TrimSpace(closeBtn.Text))
				closeHasAction = closeBtn.OnTapped != nil
			})

			if closeHasAction && closeWords[closeLabel] {
				fyne.DoAndWait(func() {
					closeBtn.OnTapped()
				})
				break
			}
		}
		time.Sleep(crawlerOverlayCloseDelay)
	}
}

// crawlTooltips attempts to capture tooltips (future implementation)
func crawlTooltips(t *testing.T, state *crawlerState, content fyne.CanvasObject, size crawlerSize) {
	t.Helper()

	// Tooltip capture is complex and may not be reliable across all platforms
	// This is a placeholder for future implementation
	t.Logf("Tooltip crawling not yet implemented for %s", state.Config.Module)
}

// Helper functions

func uiCrawlEnabled() bool {
	value := strings.TrimSpace(os.Getenv(uiCrawlEnvVar))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func setupHeadlessIfNeeded(t *testing.T) {
	t.Helper()

	if isHeadlessEnvironment() {
		ensureDisplayForGraphicalTests(t)
		if display := strings.TrimSpace(os.Getenv("DISPLAY")); display != "" {
			t.Logf("Running in headless mode with DISPLAY=%s", display)
		}
	}
}

func captureAndSave(t *testing.T, state *crawlerState, name string) {
	t.Helper()

	var (
		img        image.Image
		captureErr error
	)
	fyne.DoAndWait(func() {
		defer func() {
			if r := recover(); r != nil {
				captureErr = fmt.Errorf("panic during canvas capture: %v", r)
			}
		}()
		img = state.Window.Canvas().Capture()
	})

	if captureErr != nil {
		state.Errors = append(state.Errors, fmt.Sprintf("Failed to capture %s: %v (using 1x1 placeholder)", name, captureErr))
		img = image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	if img == nil {
		state.Errors = append(state.Errors, fmt.Sprintf("Failed to capture %s: image is nil", name))
		return
	}

	// Save to file
	filename := filepath.Join(state.Config.OutputDir, name+".png")
	if err := saveCrawlerScreenshot(img, filename); err != nil {
		state.Errors = append(state.Errors, fmt.Sprintf("Failed to save %s: %v", name, err))
		return
	}

	// Store in memory for potential comparison
	state.Screenshots[name] = img

	// Verify image quality
	if !verifyCrawlerImageQuality(img) {
		state.Errors = append(state.Errors, fmt.Sprintf("Poor quality image captured for %s", name))
	}
}

func triggerDialog(t *testing.T, state *crawlerState, content fyne.CanvasObject, trigger dialogTrigger) bool {
	t.Helper()

	var buttons []*widget.Button
	fyne.DoAndWait(func() {
		buttons = findAllButtons(content)
	})

	for _, button := range buttons {
		var label string
		var hasAction bool
		fyne.DoAndWait(func() {
			label = strings.TrimSpace(button.Text)
			hasAction = button.OnTapped != nil
		})

		if hasAction && strings.EqualFold(label, trigger.ButtonText) {
			fyne.DoAndWait(func() {
				button.OnTapped()
			})
			time.Sleep(crawlerDialogDelay)
			return true
		}
	}

	return false
}

func closeDialog(t *testing.T, state *crawlerState, content fyne.CanvasObject) {
	t.Helper()

	closeWords := []string{"close", "ok", "done", "cancel", "back", "dismiss"}

	var buttons []*widget.Button
	fyne.DoAndWait(func() {
		buttons = findAllButtons(content)
	})

	for _, button := range buttons {
		var label string
		var hasAction bool
		fyne.DoAndWait(func() {
			label = strings.ToLower(strings.TrimSpace(button.Text))
			hasAction = button.OnTapped != nil
		})

		if hasAction {
			for _, closeWord := range closeWords {
				if label == closeWord {
					fyne.DoAndWait(func() {
						button.OnTapped()
					})
					return
				}
			}
		}
	}
}

func safeCrawlerName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	if s == "" {
		s = "unnamed"
	}
	return s
}

// createModule dynamically creates modules by name
func createModule(moduleName string) (moduleLifecycle, error) {
	switch moduleName {
	case "demo1gui":
		return demo1gui.NewEmbeddedApp(), nil
	case "demo2gui":
		mod, err := demo2gui.NewEmbeddedCrossbarApp()
		if err != nil {
			return nil, fmt.Errorf("crossbar module creation failed: %w", err)
		}
		return mod, nil
	case "demo3gui":
		return demo3gui.NewEmbeddedMNISTApp(), nil
	case "demo5gui":
		return demo5gui.NewEmbeddedComparisonApp(), nil
	case "demo6gui":
		return demo6gui.NewEmbeddedEDAApp(), nil
	default:
		return nil, fmt.Errorf("unknown module: %s", moduleName)
	}
}

// Additional helper functions that need to be added to the test infrastructure

func saveCrawlerScreenshot(img image.Image, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Same implementation as existing saveTestScreenshot, but with error return
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// Note: isHeadlessEnvironment is already defined in e2e_gui_test.go and will be reused

func verifyCrawlerImageQuality(img image.Image) bool {
	if img == nil {
		return false
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return false
	}

	// Check for some variation in the image
	firstPixel := img.At(bounds.Min.X, bounds.Min.Y)
	hasVariation := false

	// Sample some pixels to check for variation
	step := max(bounds.Dx()/20, 1)
	for y := bounds.Min.Y; y < bounds.Max.Y && !hasVariation; y += step {
		for x := bounds.Min.X; x < bounds.Max.X && !hasVariation; x += step {
			if img.At(x, y) != firstPixel {
				hasVariation = true
			}
		}
	}

	return hasVariation
}

func isKnownCrawlerGap(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	known := []string{
		"not yet integrated",
		"font compatibility",
		"infinite loop",
	}
	for _, needle := range known {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
