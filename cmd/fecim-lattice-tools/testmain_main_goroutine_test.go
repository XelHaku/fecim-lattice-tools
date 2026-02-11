package main

import (
	"os"
	"runtime"
	"testing"
)

// Some Fyne/GLFW operations (notably app.Run/ShowAndRun) must execute on the
// process main goroutine. The testing framework runs individual tests in their
// own goroutines, so we provide a small "run on main goroutine" dispatcher that
// tests can use when they need the real GLFW driver (e.g. Xvfb screenshot
// capture).
//
// This keeps default tests working, while allowing opt-in visual tests to
// safely call a.Run() without panicking.

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

	// Fyne's GLFW driver is also sensitive to the main OS thread.
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
			os.Exit(code)
		}
	}
}
