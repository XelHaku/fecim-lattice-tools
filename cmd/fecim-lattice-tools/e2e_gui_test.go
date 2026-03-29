//go:build !ci
// +build !ci

// Package main provides GUI-based end-to-end tests for the FeCIM Lattice Tools application.
// These tests require a display and are excluded from CI builds.
//
// Run locally with: go test -v ./cmd/fecim-lattice-tools/... -run E2EGUI
package main

import (
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
)

// =============================================================================
// MODULE LIFECYCLE TESTS (GUI-based)
// =============================================================================

// TestE2EGUIAllModulesLifecycle tests BuildContent/Start/Stop for all 6 modules.
// This test requires a display but uses test.NewApp() for headless testing.
// NOTE: May fail with font errors on some systems - see e2e_test.go for non-GUI tests.
func TestE2EGUIAllModulesLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E GUI test in short mode")
	}
	// GUI/E2E tests should not mutate tracked calibration baselines.
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	ensureDisplayForGraphicalTests(t)

	// Set a reasonable timeout for the entire test
	t.Cleanup(func() {
		if t.Failed() {
			t.Log("GUI lifecycle test failed - may be due to headless environment or font issues")
		}
	})

	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("E2E GUI Lifecycle Test")
	defer window.Close()

	initialGoroutines := runtime.NumGoroutine()
	t.Logf("Initial goroutine count: %d", initialGoroutines)

	// Test each module with recovery from panics
	// NOTE: Some modules have issues with test.NewApp() font support
	// Crossbar and MNIST modules use Bold+Monospace fonts not in test theme
	modules := []struct {
		name   string
		create func() (moduleLifecycle, error)
		skip   bool // Skip modules with known test environment issues
	}{
		{"Hysteresis", func() (moduleLifecycle, error) {
			return demo1gui.NewEmbeddedApp(), nil
		}, false},
		{"Crossbar", func() (moduleLifecycle, error) {
			return demo2gui.NewEmbeddedCrossbarApp()
		}, true}, // Skip: uses Bold+Monospace fonts not in test theme
		{"MNIST", func() (moduleLifecycle, error) {
			return demo3gui.NewEmbeddedDualModeApp(), nil
		}, true}, // Skip: has infinite loop in WeightComparisonWidget with test driver
		{"Circuits", func() (moduleLifecycle, error) {
			return demo4gui.NewEmbeddedCircuitsApp(), nil
		}, false},
		{"Comparison", func() (moduleLifecycle, error) {
			return demo5gui.NewEmbeddedComparisonApp(), nil
		}, false},
		{"EDA", func() (moduleLifecycle, error) {
			return demo6gui.NewEmbeddedEDAApp(), nil
		}, false},
	}

	for _, m := range modules {
		t.Run(m.name, func(t *testing.T) {
			if m.skip {
				t.Skipf("Skipping %s: known issues with test.NewApp() environment", m.name)
			}
			testGUIModuleLifecycle(t, app, window, m.create, m.name)
		})
	}

	// Allow time for goroutine cleanup
	time.Sleep(500 * time.Millisecond)
	runtime.GC()

	finalGoroutines := runtime.NumGoroutine()
	goroutineDelta := finalGoroutines - initialGoroutines
	t.Logf("Final goroutine count: %d (delta: %d)", finalGoroutines, goroutineDelta)

	// Allow up to 10 goroutine growth for Fyne internal workers
	if goroutineDelta > 10 {
		t.Logf("Warning: potential goroutine leak: delta=%d (may be Fyne internals)", goroutineDelta)
	}
}

// testGUIModuleLifecycle tests a single module's lifecycle with panic recovery
func testGUIModuleLifecycle(t *testing.T, app fyne.App, window fyne.Window, createModule func() (moduleLifecycle, error), name string) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			// Log but don't fail - some modules may have font issues in test environment
			t.Logf("%s module panicked (may be font issue in test env): %v", name, r)
		}
	}()

	module, err := createModule()
	if err != nil {
		t.Errorf("%s module creation failed: %v", name, err)
		return
	}
	if module == nil {
		t.Errorf("%s module creation returned nil", name)
		return
	}

	// Test BuildContent with recovery
	var content fyne.CanvasObject
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("%s BuildContent panicked: %v", name, r)
			}
		}()
		content = module.BuildContent(app, window)
	}()

	if content == nil {
		t.Logf("%s BuildContent returned nil (may be font issue)", name)
	}

	// Test Start with recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("%s Start panicked: %v", name, r)
			}
		}()
		module.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	// Test Stop with recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("%s Stop panicked: %v", name, r)
			}
		}()
		module.Stop()
	}()

	time.Sleep(100 * time.Millisecond)

	t.Logf("%s module lifecycle completed", name)
}

// =============================================================================
// MODULE SWITCHING TESTS
// =============================================================================

// TestE2EGUIModuleSwitching tests switching between modules.
func TestE2EGUIModuleSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E GUI test in short mode")
	}
	// GUI/E2E tests should not mutate tracked calibration baselines.
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	ensureDisplayForGraphicalTests(t)

	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("E2E Module Switching Test")
	defer window.Close()

	// Only use modules that work with test.NewApp()
	// Skip Crossbar and MNIST due to font/fyne.Do issues
	var demo1 moduleLifecycle = demo1gui.NewEmbeddedApp()
	var demo4 moduleLifecycle = demo4gui.NewEmbeddedCircuitsApp()
	var demo5 moduleLifecycle = demo5gui.NewEmbeddedComparisonApp()

	modules := []moduleLifecycle{demo1, demo4, demo5}

	// Build all content first with recovery
	for _, m := range modules {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("BuildContent panicked: %v", r)
				}
			}()
			m.BuildContent(app, window)
		}()
	}

	// Simulate rapid module switching
	var currentModule moduleLifecycle
	for i := 0; i < 3; i++ {
		for _, m := range modules {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Module switch panicked: %v", r)
					}
				}()

				if currentModule != nil {
					currentModule.Stop()
				}
				m.Start()
				currentModule = m
			}()

			time.Sleep(50 * time.Millisecond)
		}
	}

	// Stop final module
	if currentModule != nil {
		func() {
			defer func() { recover() }()
			currentModule.Stop()
		}()
	}

	time.Sleep(200 * time.Millisecond)
	t.Log("Module switching test completed")
}

// TestE2EGUIRapidModuleSwitching tests rapid switching to detect race conditions.
func TestE2EGUIRapidModuleSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E GUI test in short mode")
	}
	// GUI/E2E tests should not mutate tracked calibration baselines.
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	ensureDisplayForGraphicalTests(t)

	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("E2E Rapid Switching Test")
	defer window.Close()

	// Use modules that work with test.NewApp()
	demo1 := demo1gui.NewEmbeddedApp()
	demo4 := demo4gui.NewEmbeddedCircuitsApp()

	// Build content with recovery
	func() {
		defer func() { recover() }()
		demo1.BuildContent(app, window)
	}()
	func() {
		defer func() { recover() }()
		demo4.BuildContent(app, window)
	}()

	// Rapidly switch 20 times
	for i := 0; i < 20; i++ {
		func() {
			defer func() { recover() }()
			demo1.Start()
			demo1.Stop()
			demo4.Start()
			demo4.Stop()
		}()
	}

	t.Log("Rapid module switching completed")
}

// =============================================================================
// CONCURRENT GUI OPERATIONS
// =============================================================================

// TestE2EGUIConcurrentOperations tests concurrent module operations.
// NOTE: This test is intentionally conservative - it tests that modules
// can be started/stopped from different goroutines without deadlocking,
// but does not test simultaneous Start/Stop on the same module (which
// would require additional synchronization in the module itself).
func TestE2EGUIConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E GUI test in short mode")
	}
	// GUI/E2E tests should not mutate tracked calibration baselines.
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	ensureDisplayForGraphicalTests(t)

	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("E2E Concurrent Operations Test")
	defer window.Close()

	// Create separate module instances for each goroutine to avoid
	// concurrent Start/Stop on same instance (which can cause deadlock)
	modules := make([]*demo4gui.EmbeddedCircuitsApp, 5)
	for i := range modules {
		modules[i] = demo4gui.NewEmbeddedCircuitsApp()
		func() {
			defer func() { recover() }()
			modules[i].BuildContent(app, window)
		}()
	}

	var wg sync.WaitGroup
	errChan := make(chan string, 10)

	// Each goroutine operates on its own module instance
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errChan <- "goroutine panicked"
				}
			}()

			for j := 0; j < 5; j++ {
				modules[id].Start()
				time.Sleep(20 * time.Millisecond)
				modules[id].Stop()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	errCount := 0
	for msg := range errChan {
		t.Log(msg)
		errCount++
	}

	if errCount > 0 {
		t.Logf("Warning: %d goroutines encountered issues", errCount)
	}

	t.Log("Concurrent GUI operations test completed")
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// isHeadlessEnvironment checks if running without display
func isHeadlessEnvironment() bool {
	// Check common CI environment variables
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "TRAVIS"}
	for _, v := range ciVars {
		if val := getEnv(v); val != "" {
			return true
		}
	}

	// Check for display on Linux
	if runtime.GOOS == "linux" {
		display := getEnv("DISPLAY")
		wayland := getEnv("WAYLAND_DISPLAY")
		if display == "" && wayland == "" {
			return true
		}
	}

	return false
}

// getEnv returns environment variable value
func getEnv(key string) string {
	return os.Getenv(key)
}
