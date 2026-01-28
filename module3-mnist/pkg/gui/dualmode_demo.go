// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode_demo.go provides quick demo functionality.
package gui

import (
	"time"

	"fyne.io/fyne/v2"
)

// StartQuickDemo runs an automated 30-second demonstration of FeCIM's key insight.
// Shows: Ideal (30 levels) → Success → Break (2 levels) → Failure → Restore
func (app *DualModeApp) StartQuickDemo() {
	if app.quickDemoRunning {
		return
	}

	app.quickDemoRunning = true
	app.quickDemoStopChan = make(chan struct{})
	app.animationEnabled = true

	go func() {
		defer func() {
			app.quickDemoRunning = false
			app.animationEnabled = false
		}()

		// Step 1: Introduction (2s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 1/5: Welcome to FeCIM - Watch the magic of 30 analog levels!")
		})
		if app.waitOrStop(2 * time.Second) {
			return
		}

		// Step 2: Load sample and show ideal prediction (3s)
		fyne.Do(func() {
			app.applyPreset(30, 0.01, 8, 8)
			app.statusLabel.SetText("QUICK DEMO | Step 2/5: Loading test digit with 30 levels (ideal)...")
		})
		if app.waitOrStop(500 * time.Millisecond) {
			return
		}
		fyne.Do(func() {
			app.loadRandomSample()
		})
		if app.waitOrStop(2500 * time.Millisecond) {
			return
		}

		// Step 3: Show success with 30 levels (3s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 3/5: 30 LEVELS = HIGH ACCURACY! FP and CIM predictions match.")
		})
		if app.waitOrStop(3 * time.Second) {
			return
		}

		// Step 4: Break it with 2 levels (4s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 4/5: Now watch what happens with only 2 levels (binary)...")
		})
		if app.waitOrStop(1 * time.Second) {
			return
		}
		fyne.Do(func() {
			app.applyPreset(2, 0.01, 8, 8)
		})
		if app.waitOrStop(500 * time.Millisecond) {
			return
		}
		// Re-run inference with same digit
		if len(app.lastPixels) > 0 {
			fyne.Do(func() {
				app.runInference(app.lastPixels)
			})
		}
		if app.waitOrStop(2500 * time.Millisecond) {
			return
		}

		// Step 5: Show failure explanation (3s)
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 5/5: 2 LEVELS = FAILURE! Binary weights cannot represent the network.")
		})
		if app.waitOrStop(3 * time.Second) {
			return
		}

		// Restore ideal settings (2s)
		fyne.Do(func() {
			app.applyPreset(30, 0.01, 8, 8)
			app.statusLabel.SetText("DEMO COMPLETE | Key insight: 30 levels enable high accuracy with 25-100× energy efficiency (Samsung Nature 2025)")
		})
		if app.waitOrStop(500 * time.Millisecond) {
			return
		}
		// Re-run inference with restored settings
		if len(app.lastPixels) > 0 {
			fyne.Do(func() {
				app.runInference(app.lastPixels)
			})
		}
	}()
}

// StopQuickDemo stops the running quick demo.
func (app *DualModeApp) StopQuickDemo() {
	if app.quickDemoRunning && app.quickDemoStopChan != nil {
		close(app.quickDemoStopChan)
	}
}

// waitOrStop waits for duration or returns true if demo was stopped.
func (app *DualModeApp) waitOrStop(d time.Duration) bool {
	select {
	case <-app.quickDemoStopChan:
		return true
	case <-time.After(d):
		return false
	}
}
