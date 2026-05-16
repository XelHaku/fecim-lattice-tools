//go:build !ci
// +build !ci

package main

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
)

// TestUICrawlerInfrastructureVerification tests the crawler infrastructure without GUI
func TestUICrawlerInfrastructureVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping crawler verification in short mode")
	}

	t.Run("environment_detection", func(t *testing.T) {
		// Test environment variable detection
		if uiCrawlEnabled() {
			t.Log("UI crawling is enabled via environment variable")
		} else {
			t.Log("UI crawling is disabled (set FECIM_UI_CRAWL=1 to enable)")
		}

		// Test headless environment detection
		if isHeadlessEnvironment() {
			t.Log("Running in headless environment")
		} else {
			t.Log("Running with display available")
		}
	})

	t.Run("directory_structure", func(t *testing.T) {
		// Test that we can create the output directory structure
		testDir := filepath.Join(crawlerScreenshotBaseDir, "test-verification")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Cleanup
		defer os.RemoveAll(testDir)

		// Verify directory exists
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			t.Fatalf("Test directory was not created")
		}

		t.Logf("Successfully created directory: %s", testDir)
	})

	t.Run("module_creation", func(t *testing.T) {
		// Test that we can create the circuits module
		module := demo4gui.NewEmbeddedCircuitsApp()
		if module == nil {
			t.Fatalf("Failed to create circuits module")
		}

		t.Log("Successfully created circuits module")

		// Test that module implements the required interface
		var _ moduleLifecycle = module // Compile-time check

		t.Log("Circuits module correctly implements moduleLifecycle interface")
	})

	t.Run("crawler_configuration", func(t *testing.T) {
		// Test crawler configuration
		config := crawlerConfig{
			Module: "circuits",
			Sizes: []crawlerSize{
				{"test", 800, 600},
			},
			CaptureBase:     true,
			CaptureTabs:     true,
			CaptureDialogs:  false, // Disable for verification
			CaptureOverlays: false, // Disable for verification
			CaptureTooltips: false, // Disable for verification
			OutputDir:       filepath.Join(crawlerScreenshotBaseDir, "circuits", "verification"),
		}

		// Verify configuration
		if config.Module == "" {
			t.Fatal("Module name is empty")
		}
		if len(config.Sizes) == 0 {
			t.Fatal("No sizes configured")
		}
		if config.OutputDir == "" {
			t.Fatal("Output directory is empty")
		}

		t.Logf("Crawler configuration is valid: %+v", config)
	})

	t.Run("helper_functions", func(t *testing.T) {
		// Test safe name generation
		testCases := []struct {
			input    string
			expected string
		}{
			{"Simple", "simple"},
			{"With Spaces", "with-spaces"},
			{"With/Slashes", "with-slashes"},
			{"With (Parentheses)", "with-parentheses"},
			{"", "unnamed"},
		}

		for _, tc := range testCases {
			result := safeCrawlerName(tc.input)
			if result != tc.expected {
				t.Errorf("safeCrawlerName(%q) = %q; expected %q", tc.input, result, tc.expected)
			}
		}

		t.Log("Helper functions working correctly")
	})

	t.Run("ui_discovery_helpers", func(t *testing.T) {
		// Test that UI discovery helper functions exist and are callable
		// We can't test them without actual UI objects, but we can verify they exist

		// Test with nil input (should handle gracefully)
		elements := findInteractiveElements(nil)
		if elements == nil {
			t.Fatal("findInteractiveElements returned nil instead of empty slice")
		}
		if len(elements) != 0 {
			t.Errorf("findInteractiveElements(nil) returned %d elements; expected 0", len(elements))
		}

		triggers := findPopupTriggers(elements)
		if triggers == nil {
			t.Fatal("findPopupTriggers returned nil instead of empty slice")
		}

		closers := findCloseElements(elements)
		if closers == nil {
			t.Fatal("findCloseElements returned nil instead of empty slice")
		}

		menus := findMenus(nil)
		if menus == nil {
			t.Fatal("findMenus returned nil instead of empty slice")
		}

		t.Log("UI discovery helper functions are working correctly")
	})
}

// TestUICrawlerDryRun tests crawler logic without actual GUI rendering
func TestUICrawlerDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dry run test in short mode")
	}

	// Test the crawler state structure
	config := crawlerConfig{
		Module:          "circuits",
		Sizes:           []crawlerSize{{"test", 640, 480}},
		CaptureBase:     true,
		CaptureTabs:     false,
		CaptureDialogs:  false,
		CaptureOverlays: false,
		CaptureTooltips: false,
		OutputDir:       filepath.Join(crawlerScreenshotBaseDir, "circuits", "dry-run"),
	}

	state := &crawlerState{
		Config:      config,
		Module:      nil, // No actual module for dry run
		Window:      nil, // No actual window for dry run
		App:         nil, // No actual app for dry run
		Screenshots: make(map[string]image.Image),
		Errors:      []string{},
	}

	// Test that state is properly initialized
	if state.Config.Module != "circuits" {
		t.Errorf("Expected module 'circuits', got %q", state.Config.Module)
	}

	if state.Screenshots == nil {
		t.Fatal("Screenshots map is nil")
	}

	if state.Errors == nil {
		t.Fatal("Errors slice is nil")
	}

	// Test adding errors
	state.Errors = append(state.Errors, "test error")
	if len(state.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(state.Errors))
	}

	t.Log("Crawler dry run test completed successfully")
}
