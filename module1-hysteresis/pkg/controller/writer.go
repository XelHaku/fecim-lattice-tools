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

func (s WriteState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateApply:
		return "APPLY"
	case StateWait:
		return "WAIT"
	case StateVerify:
		return "VERIFY"
	case StateHold:
		return "HOLD"
	case StateSuccess:
		return "SUCCESS"
	case StateFailed:
		return "FAILED"
	case StateForceReset:
		return "FORCE_RESET"
	case StateResetting:
		return "RESETTING"
	default:
		return "UNKNOWN"
	}
}

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
	State        WriteState
	PhaseTimer   float64
	CurrentField float64 // The target field for the current pulse
	PulseCount   int
	TotalPulses  int // Accumulated pulses across retries
	RetryCount   int // Number of times we've restarted the ISPP loop

	InitialLevel    int
	InitialLevelSet bool

	// Overshoot tracking (for autonomous recalibration)
	OvershootCount int // Overshoots in current target cycle
	OvershootTotal int // Overshoots across runtime
	resetDirection int // Sticky reset direction after overshoot (-1 or +1)

	// Binary Search State (for ISPP)
	VMin float64 // Lower bound of safe voltage (won't overshoot)
	VMax float64 // Upper bound of voltage search space

	// Bound tracking for bisection (set after verified undershoot/overshoot)
	VMinSet bool
	VMaxSet bool

	// Outputs
	LastVerifyLevel int
	LastError       int

	// Stuck detection (no level change across verifies)
	StuckCount int

	// Progress detection (error not improving across verifies)
	LastAbsError   int
	NoImproveCount int

	// Step floor to break out of quantization stalls (in E-field units).
	stepFloor float64
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
	wc.OvershootCount = 0
	wc.CurrentField = 0
	wc.LastVerifyLevel = 0
	wc.LastError = 0
	wc.LastAbsError = -1
	wc.NoImproveCount = 0
	wc.previousLevel = 0
	wc.previousField = 0
	wc.resetDirection = 0
	wc.StuckCount = 0
	wc.stepFloor = 0

	// Initialize Binary Search Bounds
	wc.VMin = 0
	wc.VMax = wc.MaxField
	wc.VMinSet = false
	wc.VMaxSet = false

	wc.InitialLevel = 0 // Will be captured in Update
	wc.InitialLevelSet = false
	// CurrentField will be set when calculateNextField is called during first Update

	log.Printf("ISPP START: target=%d fromSaturation=%v Ec=%.3g MaxField=%.3g bounds=[%.3f, %.3f]×Ec",
		wc.TargetLevel, wc.FromSaturation, wc.EcField, wc.MaxField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
}

// Reset clears the controller state for a completely new operation
func (wc *WriteController) ResetState() {
	wc.State = StateIdle
	wc.TotalPulses = 0
	wc.RetryCount = 0
	wc.Attempts = 0
	wc.SuccessCount = 0
	wc.FailureCount = 0
	wc.OvershootCount = 0
	wc.OvershootTotal = 0
	wc.LastAbsError = -1
	wc.NoImproveCount = 0
	wc.stepFloor = 0
}

// ResetDirection exposes the current sticky reset direction for diagnostics/logging.
func (wc *WriteController) ResetDirection() int {
	return wc.resetDirection
}

// Update advances the controller state logic.
func (wc *WriteController) Update(dt float64, currentField float64, currentLevel int) (targetField float64, done bool) {
	wc.PhaseTimer += dt

	// Capture initial level on first update
	if !wc.InitialLevelSet {
		wc.InitialLevel = currentLevel
		wc.InitialLevelSet = true
		wc.LastVerifyLevel = currentLevel
		wc.previousLevel = currentLevel
	}

	// Pulse duration constant (could be configurable)
	pulseDur := wc.PulseDuration

	switch wc.State {
	case StateApply:
		// If we're already at target, skip pulses entirely.
		if wc.PulseCount == 0 && currentLevel == wc.TargetLevel {
			wc.LastVerifyLevel = currentLevel
			wc.LastError = 0
			wc.State = StateSuccess
			wc.SuccessCount++
			log.Printf("ISPP EARLY: already at target level %d, no pulse needed", currentLevel)
			return 0, true
		}

		// Calculate field for first pulse if not done yet
		if wc.PulseCount == 0 && wc.CurrentField == 0 {
			wc.calculateNextField(currentLevel)
		}

		// Target is the pulse field
		targetField = wc.CurrentField
		if wc.PhaseTimer <= dt {
			log.Printf("ISPP APPLY: pulse=%d currentLevel=%d targetLevel=%d pulseDir=%d E=%.3f×Ec bounds=[%.3f, %.3f]×Ec",
				wc.PulseCount+1, currentLevel, wc.TargetLevel, pulseDirection(wc.CurrentField),
				wc.CurrentField/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
		}

		// If we reached the target field (approx), switch to WAIT
		if wc.PhaseTimer > pulseDur*0.4 && math.Abs(currentField-wc.CurrentField) < 0.01*wc.MaxField {
			wc.State = StateWait
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateWait:
		targetField = wc.CurrentField
		if wc.PhaseTimer <= dt {
			log.Printf("ISPP WAIT: holding E=%.3f×Ec for verify (pulse=%d)", wc.CurrentField/wc.EcField, wc.PulseCount+1)
		}
		if wc.PhaseTimer > pulseDur*0.3 {
			wc.State = StateVerify // Go to Verify
			wc.PhaseTimer = 0
		}
		return targetField, false

	case StateResetting:
		// REAL ISPP OVERSHOOT RECOVERY:
		// Instead of full saturation reset, just apply voltage in OPPOSITE direction
		// with a smaller step to nudge back toward target.
		resetDir := wc.resetDirection
		if resetDir == 0 {
			resetDir = -pulseDirection(wc.CurrentField) // Opposite of last pulse
			if resetDir == 0 {
				resetDir = wc.directionToTarget(currentLevel)
			}
		}

		// Use a moderate correction pulse (not full saturation)
		// Step size decreases with more overshoots for finer control
		correctionStep := wc.EcField * 0.8 // Start with 0.8×Ec correction
		if wc.OvershootCount > 1 {
			correctionStep = wc.EcField * 0.4 // Smaller after multiple overshoots
		}
		if wc.OvershootCount > 3 {
			correctionStep = wc.EcField * 0.2 // Even smaller for persistent issues
		}

		if resetDir < 0 {
			targetField = -correctionStep
		} else {
			targetField = correctionStep
		}

		if wc.PhaseTimer <= dt {
			log.Printf("ISPP REVERSE: overshoot recovery, applying E=%.3f×Ec (opposite direction)", targetField/wc.EcField)
		}

		// Wait for correction pulse to complete
		if wc.PhaseTimer > pulseDur*0.6 && math.Abs(currentField-targetField) < 0.05*wc.MaxField {
			wc.State = StateApply
			wc.PhaseTimer = 0

			// Set up next pulse toward target
			nextDir := wc.directionToTarget(currentLevel)
			if nextDir == 0 {
				// Hit target during reset - go to verify
				wc.CurrentField = 0
				wc.State = StateVerify
			} else {
				// After overshoot correction, we need STRONG reverse pulses
				// Start at 0.6×Ec - the Preisach model needs fields near Ec to switch
				reverseField := wc.EcField * 0.6
				if nextDir < 0 {
					wc.CurrentField = -reverseField
				} else {
					wc.CurrentField = reverseField
				}
			}

			// Don't reset InitialLevel - keep tracking original direction
			wc.PulseCount++
			wc.resetDirection = 0
			log.Printf("ISPP CORRECTION DONE: level=%d, next E=%.3f×Ec toward target=%d",
				currentLevel, wc.CurrentField/wc.EcField, wc.TargetLevel)
		}
		return targetField, false

	case StateVerify:
		// Target is 0V for verification
		targetField = 0.0

		// Wait for field to settle to 0
		if wc.PhaseTimer > pulseDur*0.3 && math.Abs(currentField) < 0.01*wc.MaxField {
			// VERIFY LOGIC
			prevLevel := wc.LastVerifyLevel
			if prevLevel == 0 {
				prevLevel = currentLevel
			}
			log.Printf("ISPP READ: currentLevel=%d targetLevel=%d prevLevel=%d currentField=%.3f×Ec",
				currentLevel, wc.TargetLevel, prevLevel, currentField/wc.EcField)
			prevError := wc.LastError
			wc.LastError = currentLevel - wc.TargetLevel
			absErr := int(math.Abs(float64(wc.LastError)))

			// Stuck detection: no level change since last verify
			if currentLevel == prevLevel {
				wc.StuckCount++
			} else {
				wc.StuckCount = 0
			}

			// Progress detection: error not decreasing across verifies
			if wc.LastAbsError >= 0 {
				if absErr >= wc.LastAbsError {
					wc.NoImproveCount++
				} else {
					wc.NoImproveCount = 0
				}
			} else {
				wc.NoImproveCount = 0
			}
			wc.LastAbsError = absErr

			// Escalate minimum step size if we are not making observable progress.
			if currentLevel != prevLevel {
				if wc.stepFloor != 0 {
					log.Printf("ISPP STEP FLOOR CLEARED: level change %d→%d", prevLevel, currentLevel)
				}
				wc.stepFloor = 0
			} else {
				switch {
				case wc.StuckCount >= 6:
					wc.raiseStepFloor(0.18*wc.EcField, "stuck>=6")
				case wc.StuckCount >= 4:
					wc.raiseStepFloor(0.12*wc.EcField, "stuck>=4")
				case wc.StuckCount >= 2:
					wc.raiseStepFloor(0.08*wc.EcField, "stuck>=2")
				}
			}
			if wc.NoImproveCount >= 2 {
				wc.raiseStepFloor(0.10*wc.EcField, "no-improve>=2")
			}

			// STRICT CONVERGENCE: Only accept exact match
			if wc.LastError == 0 {
				log.Printf("ISPP VERIFY RESULT: hit target level %d", currentLevel)
				wc.LastVerifyLevel = currentLevel
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
			pulseDir := pulseDirection(wc.CurrentField)
			overshoot := false
			if pulseDir > 0 {
				overshoot = prevLevel <= wc.TargetLevel && currentLevel > wc.TargetLevel
			} else if pulseDir < 0 {
				overshoot = prevLevel >= wc.TargetLevel && currentLevel < wc.TargetLevel
			}

			// Update search bounds based on whether we need more or less magnitude.
			absField := math.Abs(wc.CurrentField)
			needMore := false
			needLess := false
			if pulseDir >= 0 {
				needMore = wc.LastError < 0
				needLess = wc.LastError > 0
			} else {
				needMore = wc.LastError > 0
				needLess = wc.LastError < 0
			}
			if needMore {
				if !wc.VMinSet || absField > wc.VMin {
					wc.VMin = absField
					wc.VMinSet = true
				}
			} else if needLess {
				if !wc.VMaxSet || absField < wc.VMax {
					wc.VMax = absField
					wc.VMaxSet = true
				}
			}

			// If bounds collapse or invert, reset them to avoid deadlocks.
			if wc.VMinSet && wc.VMaxSet && wc.VMin >= wc.VMax {
				wc.VMin = 0
				wc.VMax = wc.MaxField
				wc.VMinSet = false
				wc.VMaxSet = false
				log.Printf("ISPP BOUNDS RESET: invalid bracket (vmin>=vmax), clearing bounds")
			}

			// DEBUG: Always log the overshoot check
			log.Printf("ISPP VERIFY: currentLevel=%d, targetLevel=%d, prevLevel=%d, pulseDir=%d, overshoot=%v",
				currentLevel, wc.TargetLevel, prevLevel, pulseDir, overshoot)
			wc.LastVerifyLevel = currentLevel
			log.Printf("ISPP VERIFY RESULT: error=%d bounds=[%.3f, %.3f]×Ec",
				wc.LastError, wc.VMin/wc.EcField, wc.VMax/wc.EcField)

			if overshoot {
				wc.OvershootCount++
				wc.OvershootTotal++
				if pulseDir != 0 {
					wc.resetDirection = -pulseDir
				} else {
					wc.resetDirection = -wc.directionToTarget(currentLevel)
				}
				log.Printf("ISPP OVERSHOOT detected! Resetting state... (count=%d)", wc.OvershootCount)
				// Clear bounds after overshoot; hysteresis branch changed.
				wc.VMin = 0
				wc.VMax = wc.MaxField
				wc.VMinSet = false
				wc.VMaxSet = false
				wc.StuckCount = 0
				wc.NoImproveCount = 0
				wc.stepFloor = 0
				wc.State = StateResetting
				wc.PhaseTimer = 0
				return 0, false
			}

			// If error sign flipped, tighten bracket for bisection.
			if prevError != 0 && wc.LastError != 0 && (prevError > 0) != (wc.LastError > 0) {
				wc.VMinSet = wc.VMinSet || wc.VMin > 0
				wc.VMaxSet = wc.VMaxSet || wc.VMax < wc.MaxField
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

// calculateNextField determines the field for the next pulse using REAL ISPP
// Real ISPP: Incremental voltage stepping with verify after each pulse
// - Start with small voltage in target direction
// - If undershoot: increase voltage by dynamic step
// - If overshoot: handled by StateResetting (reverse direction)
func (wc *WriteController) calculateNextField(currentLevel int) {
	targetLevel := wc.TargetLevel
	direction := wc.directionToTarget(currentLevel)
	levelError := int(math.Abs(float64(currentLevel - targetLevel)))

	// Dynamic step size based on distance to target
	// Larger steps when far, smaller steps when close
	var stepSize float64
	switch {
	case levelError > 10:
		stepSize = 0.15 * wc.EcField // Large step when very far
	case levelError > 5:
		stepSize = 0.08 * wc.EcField // Medium-large step
	case levelError > 2:
		stepSize = 0.04 * wc.EcField // Medium step
	case levelError > 0:
		stepSize = 0.02 * wc.EcField // Small step when close
	default:
		stepSize = 0.01 * wc.EcField // Tiny step (shouldn't reach here)
	}

	// Near saturation levels (1-3 or 28-30), use larger steps due to K_dep feedback
	if targetLevel <= 3 || targetLevel >= (wc.NumLevels-2) {
		if stepSize < 0.06*wc.EcField {
			stepSize = 0.06 * wc.EcField // Larger step for saturation targets
		}
	}

	// After overshoot, use smaller steps (but not too small)
	if wc.OvershootCount > 0 {
		stepSize = 0.02 * wc.EcField // Fine step after overshoot
	}

	// FIRST PULSE: Start with calibration hint or small step
	if wc.PulseCount == 0 {
		if direction == 0 {
			wc.CurrentField = 0
			log.Printf("ISPP INIT: already at target level %d, no pulse needed", currentLevel)
			return
		}

		// Use calibration as initial hint if available
		wrdTargetIdx := targetLevel - 1
		atNegativeSat := currentLevel <= 3
		atPositiveSat := currentLevel >= (wc.NumLevels - 2)

		var initialField float64
		if wc.FromSaturation && wc.CalibManager != nil && wrdTargetIdx >= 0 && wrdTargetIdx < len(wc.CalibManager.CalibrationUp) {
			if atNegativeSat {
				initialField = math.Abs(wc.CalibManager.CalibrationUp[wrdTargetIdx])
				log.Printf("ISPP INIT: target=%d, calib=%.3f×Ec (from -Pr), step=%.3f×Ec",
					targetLevel, initialField/wc.EcField, stepSize/wc.EcField)
			} else if atPositiveSat {
				initialField = math.Abs(wc.CalibManager.CalibrationDown[wrdTargetIdx])
				log.Printf("ISPP INIT: target=%d, calib=%.3f×Ec (from +Pr), step=%.3f×Ec",
					targetLevel, initialField/wc.EcField, stepSize/wc.EcField)
			} else {
				// Not at saturation - start with coercive field
				initialField = wc.EcField
				log.Printf("ISPP INIT: target=%d, start=%.3f×Ec (not at sat), step=%.3f×Ec",
					targetLevel, initialField/wc.EcField, stepSize/wc.EcField)
			}
		} else {
			// No calibration - start with coercive field
			initialField = wc.EcField
			log.Printf("ISPP INIT: target=%d, start=%.3f×Ec (no calib), step=%.3f×Ec",
				targetLevel, initialField/wc.EcField, stepSize/wc.EcField)
		}

		// Fallback: if calibration returned 0, use coercive field as starting point
		if initialField < 0.1*wc.EcField {
			initialField = wc.EcField
			log.Printf("ISPP INIT FALLBACK: calib was 0, using Ec=%.3f×Ec", initialField/wc.EcField)
		}

		// For targets near saturation (levels 1-3 or 28-30), use stronger starting field
		if targetLevel <= 3 || targetLevel >= (wc.NumLevels-2) {
			if initialField < 1.5*wc.EcField {
				initialField = 1.5 * wc.EcField
				log.Printf("ISPP INIT: near-saturation target %d, boosting to %.3f×Ec", targetLevel, initialField/wc.EcField)
			}
		}

		// Apply direction
		if direction < 0 {
			initialField = -initialField
		}

		wc.CurrentField = initialField
		return
	}

	// SUBSEQUENT PULSES: Increment voltage if undershoot
	if direction == 0 {
		wc.CurrentField = 0
		log.Printf("ISPP STEP: hit target level %d", currentLevel)
		return
	}

	// After overshoot, we're going in REVERSE direction
	// This requires careful stepping to avoid ping-pong oscillation
	if wc.OvershootCount > 0 {
		// Check if we're going in the same direction as initial or reversed
		initialDir := 1
		if wc.InitialLevel > wc.TargetLevel {
			initialDir = -1
		}
		isReversed := direction != initialDir

		if isReversed {
			// REVERSE DIRECTION after overshoot
			// Use calibration hints if available, otherwise conservative stepping
			// Cap at 0.95×Ec to avoid sharp switching that causes ping-pong
			prevVoltage := math.Abs(wc.CurrentField)

			// Start at 0.7×Ec, increment by 0.05×Ec, cap at 0.95×Ec
			minReverseField := 0.7 * wc.EcField
			reverseStep := 0.05 * wc.EcField
			maxReverseField := 0.95 * wc.EcField // Stay below sharp switching threshold

			// After multiple overshoots, use even finer steps
			if wc.OvershootCount > 2 {
				reverseStep = 0.03 * wc.EcField
			}

			var nextVoltage float64
			if prevVoltage < minReverseField {
				nextVoltage = minReverseField
			} else {
				nextVoltage = prevVoltage + reverseStep
			}

			// Cap at max reverse field (below Ec to avoid ping-pong)
			if nextVoltage > maxReverseField {
				nextVoltage = maxReverseField
			}

			// Apply direction
			if direction < 0 {
				nextVoltage = -nextVoltage
			}

			wc.CurrentField = nextVoltage
			wc.VMin = math.Abs(nextVoltage)

			log.Printf("ISPP REVERSE STEP: level=%d→%d, E=%.3f×Ec→%.3f×Ec (+%.3f×Ec reverse step, cap=%.2f×Ec)",
				currentLevel, targetLevel, prevVoltage/wc.EcField, math.Abs(nextVoltage)/wc.EcField, reverseStep/wc.EcField, maxReverseField/wc.EcField)
			return
		}
		// Forward direction after overshoot - use fine steps
		stepSize = 0.03 * wc.EcField
	}

	// If we have bounds from verify, use bisection for faster convergence.
	if wc.VMinSet && wc.VMaxSet && wc.VMax > wc.VMin && wc.StuckCount < 3 && wc.NoImproveCount < 3 && wc.stepFloor == 0 {
		prevVoltage := math.Abs(wc.CurrentField)
		nextVoltage := 0.5 * (wc.VMin + wc.VMax)
		// Nudge if midpoint is too close to current voltage (avoid stalling).
		minStep := 0.02 * wc.EcField
		if math.Abs(nextVoltage-prevVoltage) < minStep {
			if wc.LastError < 0 {
				nextVoltage = math.Min(wc.VMax, prevVoltage+minStep)
			} else if wc.LastError > 0 {
				nextVoltage = math.Max(wc.VMin, prevVoltage-minStep)
			}
		}
		if direction < 0 {
			nextVoltage = -nextVoltage
		}
		wc.CurrentField = nextVoltage
		log.Printf("ISPP BISECT: level=%d→%d, E=%.3f×Ec→%.3f×Ec bounds=[%.3f, %.3f]×Ec",
			currentLevel, targetLevel, prevVoltage/wc.EcField, math.Abs(nextVoltage)/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
		return
	}

	// Stuck or no-improvement escalation: boost step and clear bounds.
	if wc.StuckCount >= 3 || wc.NoImproveCount >= 3 {
		wc.raiseStepFloor(0.12*wc.EcField, "stuck/no-improve>=3")
		// Reset bounds to avoid being trapped by stale bracket.
		wc.VMin = 0
		wc.VMax = wc.MaxField
		wc.VMinSet = false
		wc.VMaxSet = false
		wc.StuckCount = 0
		wc.NoImproveCount = 0
		log.Printf("ISPP STUCK: boosting step (floor=%.3f×Ec) and clearing bounds", wc.stepFloor/wc.EcField)
	}

	// Enforce step floor if set (prevents tiny steps that never cross quantization boundaries).
	if wc.stepFloor > 0 && stepSize < wc.stepFloor {
		stepSize = wc.stepFloor
	}

	// Undershoot: increase voltage by step
	prevVoltage := math.Abs(wc.CurrentField)
	nextVoltage := prevVoltage + stepSize

	// Cap at max field
	if nextVoltage > wc.MaxField {
		nextVoltage = wc.MaxField
	}

	// Apply direction
	if direction < 0 {
		nextVoltage = -nextVoltage
	}

	wc.CurrentField = nextVoltage
	wc.VMin = math.Abs(nextVoltage) // Track for next increment

	log.Printf("ISPP STEP: level=%d→%d, E=%.3f×Ec→%.3f×Ec (+%.3f×Ec step)",
		currentLevel, targetLevel, prevVoltage/wc.EcField, math.Abs(nextVoltage)/wc.EcField, stepSize/wc.EcField)
}

func (wc *WriteController) directionToTarget(currentLevel int) int {
	if currentLevel < wc.TargetLevel {
		return 1
	}
	if currentLevel > wc.TargetLevel {
		return -1
	}
	return 0
}

func pulseDirection(field float64) int {
	if field > 0 {
		return 1
	}
	if field < 0 {
		return -1
	}
	return 0
}
