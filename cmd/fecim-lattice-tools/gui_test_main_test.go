//go:build !ci
// +build !ci

package main

import (
	"os"
	"runtime"
	"sync"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var (
	guiTestApp     fyne.App
	guiTestAppOnce sync.Once
)

// getGUITestApp returns a singleton fyne.App backed by the real GLFW driver.
// It is started in TestMain on the main goroutine when a display is available.
func getGUITestApp() fyne.App {
	guiTestAppOnce.Do(func() {
		guiTestApp = app.New()
	})
	return guiTestApp
}

func TestMain(m *testing.M) {
	// Fyne's GLFW driver requires app.Run() on the process main goroutine.
	// go test executes tests in goroutines, so we use this pattern:
	// - Create a singleton app
	// - Run tests in a goroutine
	// - Run the app event loop on the main goroutine
	// - Quit the app when tests complete
	// NOTE: Running a real GLFW-backed app event loop inside `go test` is fragile in
	// headless/Xvfb environments and has caused intermittent panics on teardown.
	// Default to the safer mode (just run tests) unless explicitly opted-in.
	//
	// To force the real event loop (local dev only): set FECIM_GLFW_TESTMAIN=1.
	if isHeadlessEnvironment() || os.Getenv("FECIM_GLFW_TESTMAIN") != "1" {
		os.Exit(m.Run())
	}

	runtime.LockOSThread()
	app := getGUITestApp()

	codeCh := make(chan int, 1)
	go func() {
		codeCh <- m.Run()
		// Stop the event loop after tests finish.
		fyne.DoAndWait(func() {
			app.Quit()
		})
	}()

	app.Run()
	os.Exit(<-codeCh)
}
