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
	StateResetting  // Internal reset due to overshoot
)

// WriteController manages the ISPP (Incremental Step Pulse Programming) loop
type WriteController struct {
	// Configuration
	NumLevels       int
	EcField         float64
	MaxField        float64
	MaxRetries      int // Max ISPP pulses before giving up
	ForceResetLimit int // Max retries before forcing a full reset
	Attempts        int
	SuccessCount    int
	FailureCount    int
	PulseDuration   float64 // Configuration field for simulation sync

	// Added for search acceleration (Slope estimation)
	previousLevel int
	previousField float64

	// Dependencies
	CalibManager *algo.CalibrationManager

	// Current Target
	TargetLevel    int
	FromSaturation bool

	// Dynamic State
	State          WriteState
	PhaseTimer     float64
	CurrentField   float64 // The target field for the current pulse
	PulseCount     int
	TotalPulses    int // Accumulated pulses across retries
	RetryCount     int // Number of times we've restarted the ISPP loop
	
	InitialLevel    int
	InitialLevelSet bool

	// Binary Search State (for ISPP)
	VMin float64 // Lower bound of safe voltage (won't overshoot)
	VMax float64 // Upper bound of voltage search space

	// Outputs
	LastVerifyLevel int
	LastError       int
}

func NewWriteController(numLevels int, ec, emax float64, calib *algo.CalibrationManager) *WriteController {
	return &WriteController{
		NumLevels:       numLevels,
		EcField:         ec,
		MaxField:        emax,
		MaxRetries:      50,   // Stubborn: Try hard to converge
		ForceResetLimit: 100,  // Effectively disabled, only for total failure
		PulseDuration:   0.15, // Default safe value
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
	
	// Initialize Binary Search Bounds
	wc.VMin = 0
	wc.VMax = wc.MaxField
	
	wc.InitialLevel = 0 // Will be captured in Update
	wc.InitialLevelSet = false
	// CurrentField will be set when calculateNextField is called during first Update
}

// Reset clears the controller state for a completely new operation
func (wc *WriteController) ResetState() {
	wc.State = StateIdle
	wc.TotalPulses = 0
	wc.RetryCount = 0
	wc.Attempts = 0
	wc.SuccessCount = 0
	wc.FailureCount = 0
}

// Update advances the controller state logic.
func (wc *WriteController) Update(dt float64, currentField float64, currentLevel int) (targetField float64, done bool) {
	wc.PhaseTimer += dt
	
	// Capture initial level on first update
	if !wc.InitialLevelSet {
		wc.InitialLevel = currentLevel
		wc.InitialLevelSet = true
	}

	// Pulse duration constant (could be configurable)
	pulseDur := wc.PulseDuration

	switch wc.State {
	case StateApply:
		// Calculate field for first pulse if not done yet
		if wc.PulseCount == 0 && wc.CurrentField == 0 {
			wc.calculateNextField(currentLevel)
		}
		
		// Target is the pulse field
		targetField = wc.CurrentField

		// If we reached the target field (approx), switch to WAIT
		if wc.PhaseTimer > pulseDur*0.4 && math.Abs(currentField-wc.CurrentField) < 0.01*wc.MaxField {
			wc.State = StateWait
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateWait:
		targetField = wc.CurrentField
		if wc.PhaseTimer > pulseDur*0.3 {
			wc.State = StateVerify // Go to Verify
			wc.PhaseTimer = 0
		}
		return targetField, false
	
	case StateResetting:
		// Determine reset polarity based on direction
		// If we were going UP and overshot, we are stuck High. Reset Low (-Max).
		// If we were going DOWN and overshot, we are stuck Low. Reset High (+Max).
		goingUp := wc.TargetLevel > wc.InitialLevel
		
		if goingUp {
			targetField = -wc.MaxField * 1.5 // Deep Negative
		} else {
			targetField = wc.MaxField * 1.5 // Deep Positive
		}
		
		// Wait for reset pulse to actually REACH the target (critical for ramp speed)
		if wc.PhaseTimer > pulseDur*0.8 && math.Abs(currentField-targetField) < 0.01*wc.MaxField {
			wc.State = StateApply
			wc.PhaseTimer = 0
			
			// BINARY SEARCH BOUNDS UPDATE
			// The pulse that caused overshoot was too strong
			// Set VMax to that failing voltage, VMin stays at 0 (or previous VMin if it exists)
			// Next try: VPulse = (VMin + VMax) / 2
			failedVoltage := math.Abs(wc.CurrentField)
			wc.VMax = failedVoltage
			wc.VMin = 0  // Reset to safe baseline
			wc.CurrentField = (wc.VMin + wc.VMax) / 2.0
			
			// Apply sign based on direction
			if !goingUp {
				wc.CurrentField = -wc.CurrentField
			}
			
			// CRITICAL: Reset InitialLevel to current level after reset
			// This ensures direction logic is consistent between calculateNextField and overshoot detection
			wc.InitialLevel = currentLevel
			
			wc.PulseCount++
			log.Printf("ISPP RESET DONE. Binary search: VMax=%.3f×Ec (failed), VMin=%.3f×Ec, trying E=%.3f×Ec",
				wc.VMax/wc.EcField, wc.VMin/wc.EcField, wc.CurrentField/wc.EcField)
		}
		return targetField, false

	case StateVerify:
		// Target is 0V for verification
		targetField = 0.0

		// Wait for field to settle to 0
		if wc.PhaseTimer > pulseDur*0.3 && math.Abs(currentField) < 0.01*wc.MaxField {
			// VERIFY LOGIC
			wc.LastVerifyLevel = currentLevel
			wc.LastError = currentLevel - wc.TargetLevel

			// STRICT CONVERGENCE: Only accept exact match
			if wc.LastError == 0 {
				wc.State = StateSuccess
				wc.SuccessCount++
				// Update calibration (simple average learning)
				// Note: In real FeCIM, we'd be more careful about updating calib from "stubborn" writes
				// as they might represent edges of the distribution.
				return 0, true
			}
			
			// Check for OVERSHOOT
			// If we passed the target, we are on the wrong hysteresis branch.
			// Standard servoing (reducing voltage) won't work due to remanence.
			// Must RESET.
			goingUp := wc.TargetLevel > wc.InitialLevel
			overshoot := (goingUp && currentLevel > wc.TargetLevel) || (!goingUp && currentLevel < wc.TargetLevel)
			
			// DEBUG: Always log the overshoot check
			log.Printf("ISPP VERIFY: currentLevel=%d, targetLevel=%d, initialLevel=%d, goingUp=%v, overshoot=%v",
				currentLevel, wc.TargetLevel, wc.InitialLevel, goingUp, overshoot)
			
			if overshoot {
				log.Printf("ISPP OVERSHOOT detected! Resetting state...")
				wc.State = StateResetting
				wc.PhaseTimer = 0
				return 0, false
			}

			// Not converged. Check retries.
			if wc.PulseCount >= wc.MaxRetries {
				// We tried hard. If we still failed, we might need a BIG reset.
				// But we are "Stubborn", so we only give up if we hit the limit.
				wc.RetryCount++
				wc.FailureCount++

				// Don't resets immediately. Just count it as a "failed cycle" but maybe keep trying?
				// For now, adhere to the "Give Up" logic if MaxRetries hit, but MaxRetries is high (50).
				wc.State = StateFailed
				return 0, true
			}

			// Continue ISPP (Next Pulse)
			wc.PulseCount++
			wc.calculateNextField(currentLevel)
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

// calculateNextField determines the field for the next pulse using binary search
func (wc *WriteController) calculateNextField(currentLevel int) {
	targetLevel := wc.TargetLevel
	
	// Determine which saturation state we're writing from
	atNegativeSat := currentLevel <= 3                    // Near Level 1 (-Pr)
	atPositiveSat := currentLevel >= (wc.NumLevels - 2)   // Near Level 30 (+Pr)
	
	// Initial Pulse: Use calibration as starting point
	if wc.PulseCount == 0 {
		wrdTargetIdx := targetLevel - 1
		var initialGuess float64
		
		// Use calibration if available and we're at saturation
		if wc.FromSaturation && wrdTargetIdx >= 0 && wrdTargetIdx < len(wc.CalibManager.CalibrationUp) {
			if atNegativeSat {
				// Starting from -Pr, use CalibrationUp
				initialGuess = wc.CalibManager.CalibrationUp[wrdTargetIdx]
				log.Printf("ISPP INIT: target=%d, calib=%.3f×Ec (from -Pr), bounds=[%.3f, %.3f]×Ec",
					targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
			} else if atPositiveSat {
				// Starting from +Pr, use CalibrationDown (already negative)
				initialGuess = wc.CalibManager.CalibrationDown[wrdTargetIdx]
				log.Printf("ISPP INIT: target=%d, calib=%.3f×Ec (from +Pr), bounds=[%.3f, %.3f]×Ec",
					targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
			} else {
				// Not at saturation - use binary search midpoint
				initialGuess = (wc.VMin + wc.VMax) / 2.0
				// Apply sign based on target direction
				if targetLevel < currentLevel {
					initialGuess = -initialGuess
				}
				log.Printf("ISPP INIT: target=%d, midpoint=%.3f×Ec (not at sat), bounds=[%.3f, %.3f]×Ec",
					targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
			}
		} else {
			// No calibration or not from saturation - use midpoint
			initialGuess = (wc.VMin + wc.VMax) / 2.0
			if targetLevel < currentLevel {
				initialGuess = -initialGuess
			}
			log.Printf("ISPP INIT: target=%d, midpoint=%.3f×Ec (no calib), bounds=[%.3f, %.3f]×Ec",
				targetLevel, initialGuess/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
		}
		
		wc.CurrentField = initialGuess
		return
	}
	
	// SUBSEQUENT PULSES: Binary Search Logic
	// error = currentLevel - targetLevel
	// error > 0: We're ABOVE target (overshoot if we were going up, OR undershoot if going down)
	// error < 0: We're BELOW target (undershoot if we were going up, OR overshoot if going down)
	
	error := currentLevel - targetLevel
	
	if error == 0 {
		// Perfect hit! (shouldn't reach here as StateVerify handles this)
		log.Printf("ISPP: Perfect hit at level %d", currentLevel)
		return
	}
	
	// Determine if we undershot or made progress
	// Case 1: Undershoot - field too weak, need stronger field
	//   - Going UP (target > initial): currentLevel < targetLevel → error < 0
	//   - Going DOWN (target < initial): currentLevel > targetLevel → error > 0
	// Case 2: We're making progress in the right direction
	
	// Update binary search bounds
	absCurrentField := math.Abs(wc.CurrentField)
	
	if error < 0 {
		// We're BELOW target
		// If we were trying to go UP → Undershoot, need more voltage
		// Update VMin = currentVoltage (this voltage is safe, didn't overshoot)
		wc.VMin = absCurrentField
		log.Printf("ISPP BINARY: Undershoot (level=%d < target=%d), VMin=%.3f×Ec → VMin=%.3f×Ec",
			currentLevel, targetLevel, wc.VMin/wc.EcField, absCurrentField/wc.EcField)
	} else if error > 0 {
		// We're ABOVE target  
		// If we were trying to go DOWN → Undershoot in reverse direction
		// But wait - this means we haven't moved enough in the negative direction
		// Actually, for going DOWN, we need to treat this differently
		
		// Check the intended direction
		goingDown := targetLevel < wc.InitialLevel
		if goingDown {
			// We're trying to go DOWN but still above target
			// Current field wasn't strong enough (in negative direction)
			// Need stronger negative field
			wc.VMin = absCurrentField
			log.Printf("ISPP BINARY: Going DOWN, still above (level=%d > target=%d), V Min=%.3f×Ec",
				currentLevel, targetLevel, absCurrentField/wc.EcField)
		} else {
			// Going UP but we're above target - this shouldn't happen in VERIFY
			// because it should have been caught as overshoot
			// Treat as needing less voltage
			wc.VMax = absCurrentField
			log.Printf("ISPP BINARY: Unexpected state (level=%d > target=%d while going UP), VMax=%.3f×Ec",
				currentLevel, targetLevel, absCurrentField/wc.EcField)
		}
	}
	
	// Calculate next voltage as midpoint of bounds
	nextVoltage := (wc.VMin + wc.VMax) / 2.0
	
	// Apply sign based on direction
	goingDown := targetLevel < wc.InitialLevel
	if goingDown {
		nextVoltage = -nextVoltage
	}
	
	wc.CurrentField = nextVoltage
	log.Printf("ISPP BINARY: Next pulse E=%.3f×Ec, bounds=[%.3f, %.3f]×E c",
		wc.CurrentField/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
}
