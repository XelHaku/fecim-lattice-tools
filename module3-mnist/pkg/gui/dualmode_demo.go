// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode_demo.go provides quick demo functionality.
package gui

import (
	"time"

	"fyne.io/fyne/v2"
)

// StartQuickDemo runs an automated ~30-second demonstration of FeCIM's key insight.
// Shows: Ideal (30 levels) → Success → Break (2 levels) → Failure → Restore
func (app *DualModeApp) StartQuickDemo() {
	if app.quickDemoRunning {
		return
	}

	const (
		stepIntroHold      = 3 * time.Second
		stepPresetHold     = 1 * time.Second
		stepSampleHold     = 4 * time.Second
		stepSuccessHold    = 4 * time.Second
		stepBreakIntroHold = 2 * time.Second
		stepBreakHold      = 1 * time.Second
		stepFailureHold    = 4 * time.Second
		stepExplainHold    = 4 * time.Second
		stepRestoreHold    = 2 * time.Second
		stepWrapHold       = 3 * time.Second
		stepAfterInferHold = 3 * time.Second
	)

	app.quickDemoRunning = true
	app.quickDemoStopChan = make(chan struct{})
	app.animationEnabled = true

	go func() {
		defer func() {
			app.quickDemoRunning = false
			app.animationEnabled = false
		}()

		// Step 1: Introduction
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 1/5: Welcome. We'll compare 30 levels vs 2 levels.")
		})
		if app.waitOrStop(stepIntroHold) {
			return
		}

		// Step 2: Load sample and show ideal prediction
		fyne.Do(func() {
			app.applyPreset(30, 0.01, 8, 8)
			app.statusLabel.SetText("QUICK DEMO | Step 2/5: Loading a test digit with 30 levels (ideal).")
		})
		if app.waitOrStop(stepPresetHold) {
			return
		}
		fyne.Do(func() {
			app.loadRandomSample()
		})
		if app.waitOrStop(stepSampleHold) {
			return
		}

		// Step 3: Show success with 30 levels
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 3/5: 30 levels → FP and CIM match. Compare the results.")
		})
		if app.waitOrStop(stepSuccessHold) {
			return
		}

		// Step 4: Break it with 2 levels
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 4/5: Switching to 2 levels (binary). Watch what changes.")
		})
		if app.waitOrStop(stepBreakIntroHold) {
			return
		}
		fyne.Do(func() {
			app.applyPreset(2, 0.01, 8, 8)
		})
		if app.waitOrStop(stepBreakHold) {
			return
		}
		// Re-run inference with same digit
		if len(app.lastPixels) > 0 {
			fyne.Do(func() {
				app.runInference(app.lastPixels)
			})
		}
		if app.waitOrStop(stepAfterInferHold) {
			return
		}

		// Step 5: Show failure explanation
		fyne.Do(func() {
			app.statusLabel.SetText("QUICK DEMO | Step 5/5: 2 levels → quantization bottleneck. Accuracy drops.")
		})
		if app.waitOrStop(stepFailureHold) {
			return
		}

		// Restore ideal settings
		fyne.Do(func() {
			app.applyPreset(30, 0.01, 8, 8)
			app.statusLabel.SetText("RESTORE | Returning to 30 levels. Notice the predictions recover.")
		})
		if app.waitOrStop(stepRestoreHold) {
			return
		}
		// Re-run inference with restored settings
		if len(app.lastPixels) > 0 {
			fyne.Do(func() {
				app.runInference(app.lastPixels)
			})
		}
		if app.waitOrStop(stepExplainHold) {
			return
		}
		fyne.Do(func() {
			app.statusLabel.SetText("DEMO COMPLETE | Key insight: 30 levels enable high accuracy and energy efficiency.")
		})
		if app.waitOrStop(stepWrapHold) {
			return
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
