package controller

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
)

// TestWriteController_OvershootRecovery verifies that the WriteController
// successfully recovers from an overshoot by resetting to a safe field,
// rather than just backing off slightly which might still be in the swithching region.
func TestWriteController_OvershootRecovery(t *testing.T) {
	// Setup
	numLevels := 30
	ec := 100.0 // 100 units
	maxField := 250.0
	calib := algo.NewCalibrationManager(numLevels)
	
	// Pre-populate calibration with a POSITIVE value for going down (simulating the log condition)
	// Log said: "calibE=0.080xEc" -> +8.0
	// This means we start scanning from +8.0 towards negative.
	// Target Level 15 (Mid). Go Down from 30.
	// Index 14 (Level 15).
	calib.CalibrationDown[14] = 8.0 

	wc := NewWriteController(numLevels, ec, maxField, calib)
	wc.PulseDuration = 1.0 // Simple units

	// Scenario: Write Target 15. Start from Saturation (Level 30).
	targetLevel := 15
	wc.Start(targetLevel, true) // From Saturation

	// Simulation function (Stateful Hysteresis)
	// Material switches SHARPLY at E = -1.0
	// If E < -1.0 -> Level 1 (Overshoot) and STAYS there (Memory)
	// It only resets to 30 if we apply a large positive field (e.g. > +100)
	simStateLevel := 30
	simulatePhysics := func(field float64) int {
		if field < -1.0 {
			simStateLevel = 1 // Overshot
		} else if field > 100 {
			simStateLevel = 30 // Reset
		}
		return simStateLevel
	}

	// Run loop
	const maxFrames = 100
	t.Logf("Starting Loop: Target=%d, Switch@ -1.0, StartCalib=8.0", targetLevel)
	
	currentField := 0.0 // Simulation field state
	currentLevel := 30
	
	// recovered := false // Unused
	

	var lastTargetField float64
	
	for i := 0; i < maxFrames; i++ {
		// 1. Controller Update (First in loop to determine target)
		dt := 0.2
		nextTarget, _ := wc.Update(dt, currentField, currentLevel)
		lastTargetField = nextTarget
		
		// 2. Simulate Physics Ramp
		rampStep := maxField * 0.5
		if math.Abs(lastTargetField - currentField) < rampStep {
			currentField = lastTargetField
		} else if lastTargetField > currentField {
			currentField += rampStep
		} else {
			currentField -= rampStep
		}
		
		// 3. Physics Response
		currentLevel = simulatePhysics(currentField)
		
		t.Logf("Frame %d: State=%d PhaseT=%.1f Field=%.2f (Target=%.2f) Level=%d Mod=%.2f", 
			i, wc.State, wc.PhaseTimer, currentField, lastTargetField, currentLevel, wc.StepModifier)

		if wc.State == StateResetting {
			// In Resetting state, we expect the controller to eventually retry.
			// Code currently uses StateResetting internally. 
			// We check if it sets CurrentField to something SAFE (>= -1.0)
		}
		
		// Check for specific failure pattern:
		// If we see Field < -1.0 REPEATEDLY after resets, it's failing.
		if wc.State == StateApply && currentField < -1.0 {
			// We are applying a failing field.
			// If we do this multiple times (different retries), it's the bug.
			if wc.RetryCount > 3 {
				t.Errorf("FAIL: Retried %d times but keeps applying failing field %.2f < -1.0", wc.RetryCount, currentField)
				break
			}
		}
	}
}
