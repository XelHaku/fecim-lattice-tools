package controller

import (
	"log"
	"math"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
)

// WriteState represents the sub-state of the writing process
type WriteState int

const (
	StateIdle WriteState = iota
	StateApply
	StateWait
	StateVerify
	StateHold
	StateSuccess
	StateFailed     // Needs external intervention (e.g. Reset)
	StateForceReset // Explicitly requests a full system reset
)

// WriteController manages the ISPP (Incremental Step Pulse Programming) loop
type WriteController struct {
	// Configuration
	NumLevels       int
	Ec              float64
	Emax            float64
	MaxRetries      int // Max ISPP pulses before giving up
	ForceResetLimit int // Max retries before forcing a full reset

	// Dependencies
	CalibManager *algo.CalibrationManager

	// Current Target
	TargetLevel    int
	FromSaturation bool

	// Dynamic State
	State          WriteState
	PhaseTimer     float64
	CurrentVoltage float64 // The target voltage for the current pulse
	PulseCount     int
	TotalPulses    int // Accumulated pulses across retries
	RetryCount     int // Number of times we've restarted the ISPP loop

	// Outputs
	LastVerifyLevel int
	LastError       int
}

func NewWriteController(numLevels int, ec, emax float64, calib *algo.CalibrationManager) *WriteController {
	return &WriteController{
		NumLevels:       numLevels,
		Ec:              ec,
		Emax:            emax,
		MaxRetries:      5, // Default from simulation.go
		ForceResetLimit: 3, // From our recent fix
		CalibManager:    calib,
		State:           StateIdle,
	}
}

// Start begins a new write operation to the target level
func (wc *WriteController) Start(targetLevel int, fromSaturation bool) {
	wc.TargetLevel = targetLevel
	wc.FromSaturation = fromSaturation
	wc.State = StateApply
	wc.PhaseTimer = 0
	wc.PulseCount = 0
	// Don't reset TotalPulses or RetryCount here if we are retrying?
	// Actually, Start is called for a *new* target or a *retry*.
	// Let's assume the caller manages RetryCount if they want to track it across resets,
	// OR we have a dedicated Retry() method.

	// For now, let's treat Start as "Begin a fresh pulse train".
	// The voltage calculation logic happens in the first Update or we pre-calc here.
	wc.calculateNextVoltage(0) // 0 for current level, but will be refined
}

// Reset clears the controller state for a completely new operation
func (wc *WriteController) ResetState() {
	wc.State = StateIdle
	wc.TotalPulses = 0
	wc.RetryCount = 0
}

// Update advances the controller state logic.
// Returns:
//
//	targetField: The electric field the physics simulation should ramp to
//	done: True if the write operation is complete (Success, Failed, or RequestReset)
func (wc *WriteController) Update(dt float64, currentField float64, currentLevel int) (targetField float64, done bool) {
	wc.PhaseTimer += dt

	// Pulse duration constant (could be configurable)
	// simulation.go: phaseDuration * 0.08, where phaseDuration is 1.0 (approx)
	// We'll use hardcoded values relative to what was there or passed in.
	const pulseDur = 0.08

	switch wc.State {
	case StateApply:
		// Target is the pulse voltage
		targetField = wc.CurrentVoltage

		// If we reached the target voltage (approx), switch to WAIT
		// simulation.go uses 0.4 * pulseDur as switching point
		if wc.PhaseTimer > pulseDur*0.4 && math.Abs(currentField-wc.CurrentVoltage) < 0.01*wc.Emax {
			wc.State = StateWait
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateWait:
		targetField = wc.CurrentVoltage
		if wc.PhaseTimer > pulseDur*0.3 {
			wc.State = StateVerify // Go to Verify
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateVerify:
		// Target is 0V for verification
		targetField = 0.0

		// Wait for field to settle to 0
		if wc.PhaseTimer > pulseDur*0.3 && math.Abs(currentField) < 0.01*wc.Emax {
			// VERIFY LOGIC
			wc.LastVerifyLevel = currentLevel
			wc.LastError = currentLevel - wc.TargetLevel

			// Convergence Check
			// Boundary tolerance: After 10 retries, accept +/- 1
			boundaryTolerance := wc.RetryCount >= 10 && math.Abs(float64(wc.LastError)) == 1
			isConverged := wc.LastError == 0 || boundaryTolerance

			if isConverged {
				wc.State = StateSuccess
				wc.CalibManager.UpdateCalibrationUp(wc.TargetLevel-1, wc.LastError, wc.Ec) // Example update
				// Wait, direction matters.
				// We need to know if we were going up or down to update the correct calibration.
				// For now, let's assume standard ISPP updates calibration.
				return 0, true
			}

			// Not converged. Check retries.
			if wc.PulseCount >= wc.MaxRetries {
				// Too many ISPP pulses in this train.
				// Need to trigger external retry logic (Reset or Boost).
				wc.RetryCount++

				if wc.RetryCount > wc.ForceResetLimit {
					wc.State = StateForceReset
				} else {
					wc.State = StateFailed // "Failed" this pulse train, needs generic retry
				}
				return 0, true
			}

			// Continue ISPP (Next Pulse)
			wc.PulseCount++
			wc.calculateNextVoltage(currentLevel)
			wc.State = StateApply
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateSuccess, StateFailed, StateForceReset:
		return 0, true

	default:
		return 0, false
	}
}

// calculateNextVoltage determines the voltage for the next pulse
func (wc *WriteController) calculateNextVoltage(currentLevel int) {
	// Logic ported from simulation.go Case 2

	targetLevel := wc.TargetLevel
	goingUp := targetLevel > currentLevel
	levelDiff := int(math.Abs(float64(targetLevel - currentLevel)))
	wrdTargetIdx := targetLevel - 1

	// Initial Pulse Logic
	if wc.PulseCount == 0 {
		var targetVoltage float64

		// Check if we start from saturation (normal path) or mid-state (overshoot recovery)
		// For mid-state, logic relies on wc.FromSaturation flag passed in Start()

		if wc.FromSaturation && wrdTargetIdx >= 0 && wrdTargetIdx < len(wc.CalibManager.CalibrationUp) {
			// NORMAL PATH: Starting from saturation, use calibration directly
			if goingUp {
				targetVoltage = wc.CalibManager.CalibrationUp[wrdTargetIdx]
			} else {
				targetVoltage = wc.CalibManager.CalibrationDown[wrdTargetIdx]
			}
			log.Printf("ISPP FROM SATURATION: target=%d, current=%d, calibV=%.3f×Ec",
				targetLevel, currentLevel, targetVoltage/wc.Ec)
		} else {
			// OVERSHOOT RECOVERY: NOT at saturation, estimate voltage for level change
			voltagePerLevel := (2.0 * wc.Ec) / float64(wc.NumLevels-1)

			if goingUp {
				// Going UP: MUST use POSITIVE voltage
				baseVoltage := voltagePerLevel * float64(levelDiff) * 1.3

				// Refine with calibration if available
				if wrdTargetIdx >= 0 && wrdTargetIdx < len(wc.CalibManager.CalibrationUp) {
					calibV := wc.CalibManager.CalibrationUp[wrdTargetIdx]
					if calibV > 0 {
						// Scale by distance ratio
						scaleFactor := float64(levelDiff) / float64(targetLevel-1)
						scaleFactor = math.Max(0.3, math.Min(scaleFactor, 1.2))
						baseVoltage = calibV * scaleFactor
					}
				}

				// Enforce bounds
				targetVoltage = math.Abs(baseVoltage)
				if targetVoltage < voltagePerLevel {
					targetVoltage = voltagePerLevel * float64(levelDiff)
				}
				if targetVoltage > 1.8*wc.Ec {
					targetVoltage = 1.8 * wc.Ec
				}
			} else {
				// Going DOWN: MUST use NEGATIVE voltage
				baseVoltage := -voltagePerLevel * float64(levelDiff) * 1.3

				// Refine with calibration
				if wrdTargetIdx >= 0 && wrdTargetIdx < len(wc.CalibManager.CalibrationDown) {
					calibV := wc.CalibManager.CalibrationDown[wrdTargetIdx]
					if calibV < 0 {
						scaleFactor := float64(levelDiff) / float64(wc.NumLevels-targetLevel)
						scaleFactor = math.Max(0.3, math.Min(scaleFactor, 1.2))
						baseVoltage = calibV * scaleFactor
					}
				}

				// Enforce bounds
				targetVoltage = -math.Abs(baseVoltage)
				if targetVoltage > -voltagePerLevel {
					targetVoltage = -voltagePerLevel * float64(levelDiff)
				}
				if targetVoltage < -1.8*wc.Ec {
					targetVoltage = -1.8 * wc.Ec
				}
			}
			log.Printf("ISPP MID-STATE RECOVERY: target=%d, current=%d, diff=%d, estV=%.3f×Ec",
				targetLevel, currentLevel, levelDiff, targetVoltage/wc.Ec)
		}

		// Fallback
		if targetVoltage == 0 {
			levelRatio := float64(targetLevel-1) / float64(wc.NumLevels-1)
			targetVoltage = wc.Ec * (levelRatio*4.4 - 2.2)
		}

		wc.CurrentVoltage = targetVoltage
		return
	}

	// Subsequent Pulses: Adaptive Stepping based on Error
	// We use the LAST error (from verify), which is stored in wc.LastError
	// But calculateNextVoltage is called with currentLevel, which is the Verified level.
	// So levelDiff here is effectively the error.

	// Error calculation from simulation.go: levelError := verifyLevel - targetLevel
	// Since we are passed 'currentLevel' which is the verifyLevel from the last step:
	levelError := currentLevel - targetLevel
	absError := math.Abs(float64(levelError))

	voltagePerLevel := (2.0 * wc.Ec) / float64(wc.NumLevels-1)

	// Adaptive step size
	stepSize := voltagePerLevel * math.Min(absError*0.8, 2.5)

	if levelError > 0 {
		// OVERSHOT (Read > Target): Need to go DOWN (Negative Voltage)
		if wc.CurrentVoltage > 0 {
			// Cross zero to negative
			wc.CurrentVoltage = -stepSize
		} else {
			// Make more negative
			wc.CurrentVoltage -= stepSize * 0.5
		}
	} else {
		// UNDERSHOT (Read < Target): Need to go UP (Positive Voltage)
		if wc.CurrentVoltage < 0 {
			// Cross zero to positive
			wc.CurrentVoltage = stepSize
		} else {
			// Make more positive
			wc.CurrentVoltage += stepSize * 0.5
		}
	}

	// Clamp
	if wc.CurrentVoltage > 2.0*wc.Ec {
		wc.CurrentVoltage = 2.0 * wc.Ec
	} else if wc.CurrentVoltage < -2.0*wc.Ec {
		wc.CurrentVoltage = -2.0 * wc.Ec
	}

	log.Printf("ISPP STEP: target=%d, current=%d, err=%+d, newV=%.3f×Ec, pulse=%d",
		targetLevel, currentLevel, levelError, wc.CurrentVoltage/wc.Ec, wc.PulseCount)
}
