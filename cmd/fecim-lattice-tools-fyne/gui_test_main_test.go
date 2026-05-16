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

// mainGoroutineJob allows tests to dispatch work onto the main OS thread,
// which is required by Fyne's GLFW driver for app.Run/ShowAndRun.
type mainGoroutineJob struct {
	fn   func()
	done chan struct{}
}

var mainGoroutineJobs = make(chan mainGoroutineJob)

func runOnMainGoroutine(fn func()) {
	job := mainGoroutineJob{fn: fn, done: make(chan struct{})}
	mainGoroutineJobs <- job
	<-job.done
}

func TestMain(m *testing.M) {
	_ = os.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	_ = os.Setenv("FECIM_DISABLE_STARTUP_CALIBRATION", "1")

	logDir, err := os.MkdirTemp("", "fecim-cmd-logs-*")
	if err == nil {
		_ = os.Setenv("FECIM_LOGS_DIR", logDir)
		defer os.RemoveAll(logDir)
	}

	// Fyne's GLFW driver requires app.Run() on the process main goroutine.
	// Default to the safer mode (just run tests) unless explicitly opted-in.
	// To force the real event loop (local dev only): set FECIM_GLFW_TESTMAIN=1.
	if isHeadlessEnvironment() || os.Getenv("FECIM_GLFW_TESTMAIN") != "1" {
		// Still service mainGoroutineJobs for xvfb tests.
		runtime.LockOSThread()

		exitCh := make(chan int, 1)
		go func() {
			exitCh <- m.Run()
		}()

		for {
			select {
			case job := <-mainGoroutineJobs:
				job.fn()
				close(job.done)
			case code := <-exitCh:
				shutdownAutoXvfbForTests()
				os.Exit(code)
			}
		}
	}

	runtime.LockOSThread()
	a := getGUITestApp()

	codeCh := make(chan int, 1)
	go func() {
		codeCh <- m.Run()
		fyne.DoAndWait(func() {
			a.Quit()
		})
	}()

	a.Run()
	shutdownAutoXvfbForTests()
	os.Exit(<-codeCh)
}
