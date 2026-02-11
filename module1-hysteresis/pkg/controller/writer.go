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
	MinStep         float64 // Hard lower bound on step size (prevents tiny increments)
	MaxRetries      int     // Max ISPP pulses per attempt before forcing a reset (<=0 disables)
	ForceResetLimit int     // Max resets before giving up (<=0 disables)
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

	// Guard pulse tracking: counts consecutive guard corrections at the target level.
	// After MaxGuardPulses, accept the level as converged to prevent direction flips.
	GuardPulseCount int
}

func NewWriteController(numLevels int, ec, emax float64, calib *algo.CalibrationManager) *WriteController {
	minStep := 0.04 * ec
	if emax > 0 && minStep > emax {
		minStep = emax
	}
	return &WriteController{
		NumLevels:       numLevels,
		EcField:         ec,
		MaxField:        emax,
		MinStep:         minStep, // Avoid overly tiny steps near target
		MaxRetries:      0,       // Unlimited pulses per attempt (no forced reset)
		ForceResetLimit: 0,       // Unlimited resets by default (never give up)
		PulseDuration:   0.15,    // Default safe value
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
	wc.GuardPulseCount = 0

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

// raiseStepFloor increases the minimum step size to break out of stalls.
// It only raises the floor (never decreases) and logs once per change.
func (wc *WriteController) raiseStepFloor(minStep float64, reason string) {
	if minStep <= 0 {
		return
	}
	if wc.MaxField > 0 && minStep > wc.MaxField {
		minStep = wc.MaxField
	}
	if minStep <= wc.stepFloor {
		return
	}
	old := wc.stepFloor
	wc.stepFloor = minStep
	scale := wc.EcField
	if scale <= 0 {
		scale = 1
	}
	if old == 0 {
		log.Printf("ISPP STEP FLOOR SET: %.3f×Ec (%s)", wc.stepFloor/scale, reason)
	} else {
		log.Printf("ISPP STEP FLOOR RAISED: %.3f×Ec→%.3f×Ec (%s)", old/scale, wc.stepFloor/scale, reason)
	}
}

// Update advances the controller state logic.
func (wc *WriteController) Update(dt float64, currentField float64, currentLevel int, guardSign int) (targetField float64, done bool) {
	wc.PhaseTimer += dt
	if guardSign > 0 {
		guardSign = 1
	} else if guardSign < 0 {
		guardSign = -1
	}

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
		// Guard against triggering mid-pulse: PulseCount stays 0 until the first VERIFY,
		// so only consider this "early exit" at the very start of the operation.
		if wc.PulseCount == 0 && wc.PhaseTimer <= dt && currentLevel == wc.TargetLevel {
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
		// Overshoot recovery: apply a reset pulse on the opposite branch,
		// then restart the search with a lower field.
		resetDir := wc.resetDirection
		if resetDir == 0 {
			resetDir = -pulseDirection(wc.CurrentField) // Opposite of last pulse
			if resetDir == 0 {
				resetDir = wc.directionToTarget(currentLevel)
			}
		}

		resetMag := wc.EcField
		if resetMag <= 0 {
			resetMag = math.Abs(wc.CurrentField)
		}
		if wc.VMaxSet && wc.VMax > 0 && wc.VMax < resetMag {
			resetMag = wc.VMax
		}
		minReset := 0.8 * wc.EcField
		if minReset > 0 && resetMag < minReset {
			resetMag = minReset
		}
		if wc.MaxField > 0 && resetMag > wc.MaxField {
			resetMag = wc.MaxField
		}

		if resetDir < 0 {
			targetField = -resetMag
		} else {
			targetField = resetMag
		}

		if wc.PhaseTimer <= dt {
			log.Printf("ISPP REVERSE: overshoot recovery, applying E=%.3f×Ec (opposite direction)", targetField/wc.EcField)
		}

		// Wait for correction pulse to complete
		if wc.PhaseTimer > pulseDur*0.6 && math.Abs(currentField-targetField) < 0.05*wc.MaxField {
			// Count the overshoot recovery pulse before choosing the next programming
			// pulse. This ensures calculateNextField uses the post-first-pulse logic
			// (bisection/bracketing), instead of restarting with the aggressive
			// first-pulse initializer that tends to re-overshoot near saturation.
			wc.PulseCount++

			wc.State = StateApply
			wc.PhaseTimer = 0

			// Set up next pulse toward target using the updated bounds.
			nextDir := wc.directionToTarget(currentLevel)
			if nextDir == 0 {
				// Hit target during reset - go to verify
				wc.CurrentField = 0
				wc.State = StateVerify
			} else {
				wc.CurrentField = targetField
				wc.calculateNextField(currentLevel)
			}

			// Don't reset InitialLevel - keep tracking original direction
			wc.resetDirection = 0
			log.Printf("ISPP RESET DONE: level=%d, next E=%.3f×Ec toward target=%d",
				currentLevel, wc.CurrentField/wc.EcField, wc.TargetLevel)
		}
		return targetField, false

	case StateVerify:
		// Target is 0V for verification
		targetField = 0.0
		guardActive := false

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
			if wc.LastError == 0 && guardSign != 0 {
				// Guard-band correction: the cell is at the target level but
				// polarization is near a bin edge. Allow up to 2 guard pulses
				// to nudge it toward center; after that, accept convergence
				// to avoid catastrophic direction flips.
				const maxGuardPulses = 2
				wc.GuardPulseCount++
				if wc.GuardPulseCount <= maxGuardPulses {
					wc.LastError = guardSign
					absErr = int(math.Abs(float64(wc.LastError)))
					guardActive = true
					log.Printf("ISPP GUARD: level=%d target=%d sign=%d guardPulse=%d/%d",
						currentLevel, wc.TargetLevel, guardSign, wc.GuardPulseCount, maxGuardPulses)
				} else {
					log.Printf("ISPP GUARD ACCEPT: level=%d target=%d (guard limit reached, accepting convergence)",
						currentLevel, wc.TargetLevel)
				}
			} else if wc.LastError != 0 {
				// Reset guard pulse counter when not at target level.
				wc.GuardPulseCount = 0
			}
			wc.LastAbsError = absErr

			if guardActive {
				// Guard-band correction: avoid treating "no level change" as stuck.
				wc.StuckCount = 0
				wc.NoImproveCount = 0
			}

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

			// If bounds collapse or invert, widen the bracket minimally
			// instead of discarding all convergence progress.
			if wc.VMinSet && wc.VMaxSet && wc.VMin >= wc.VMax {
				minSep := 0.04 * wc.EcField // Minimum bracket width
				if wc.stepFloor > minSep {
					minSep = wc.stepFloor
				}
				if needMore {
					// We need higher voltage: keep VMin, push VMax up
					wc.VMax = wc.VMin + minSep
					if wc.VMax > wc.MaxField {
						wc.VMax = wc.MaxField
					}
				} else if needLess {
					// We need lower voltage: keep VMax, push VMin down
					wc.VMin = wc.VMax - minSep
					if wc.VMin < 0 {
						wc.VMin = 0
					}
				} else {
					// Unknown direction: full reset as last resort
					wc.VMin = 0
					wc.VMax = wc.MaxField
					wc.VMinSet = false
					wc.VMaxSet = false
				}
				log.Printf("ISPP BOUNDS FIX: bracket collapsed, widened to [%.3f, %.3f]×Ec",
					wc.VMin/wc.EcField, wc.VMax/wc.EcField)
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
				// Keep a bracket so the next pulse restarts with a lower field.
				absField := math.Abs(wc.CurrentField)
				if absField > 0 {
					if !wc.VMaxSet || absField < wc.VMax {
						wc.VMax = absField
					}
					wc.VMaxSet = true
				}
				if !wc.VMinSet {
					// Avoid resetting the lower bracket bound to exactly 0.
					// This shows up in logs/CSVs as "VMin reset to minimum" and is
					// rarely useful in practice (fields below MinStep are typically
					// ineffective for state changes).
					min := wc.MinStep
					if min <= 0 {
						min = 0.02 * wc.EcField
					}
					if min < 0 {
						min = 0
					}
					wc.VMin = min
					wc.VMinSet = true
				}
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
			if wc.MaxRetries > 0 && wc.PulseCount >= wc.MaxRetries {
				// Per-attempt pulse budget exceeded: force a full reset and retry.
				wc.RetryCount++
				if wc.ForceResetLimit > 0 && wc.RetryCount > wc.ForceResetLimit {
					wc.FailureCount++
					wc.State = StateFailed
					return 0, true
				}
				log.Printf("ISPP RETRY LIMIT: forcing full reset (retry=%d target=%d)", wc.RetryCount, wc.TargetLevel)
				wc.State = StateForceReset
				return 0, true
			}

			// Continue ISPP (Next Pulse)
			wc.PulseCount++
			calcLevel := currentLevel
			if guardActive {
				// Guard correction: nudge calcLevel toward the side that needs
				// more polarization, but NEVER flip the overall write direction.
				// Flipping direction at the target level causes catastrophic overshoot.
				candidateLevel := currentLevel + guardSign
				prevDir := wc.directionToTarget(currentLevel)
				candidateDir := wc.directionToTarget(candidateLevel)
				if prevDir != 0 && candidateDir != 0 && candidateDir != prevDir {
					// Would flip direction — keep calcLevel at currentLevel
					// and let the small voltage increment handle it.
					log.Printf("ISPP GUARD CLAMP: avoiding direction flip (current=%d candidate=%d target=%d)",
						currentLevel, candidateLevel, wc.TargetLevel)
				} else {
					calcLevel = candidateLevel
				}
			}
			wc.calculateNextField(calcLevel)
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

			// Start at 0.7×Ec, increment by 0.05×Ec, cap at 1.5×Ec
			// Note: L-K dynamics need higher fields than Preisach's sharp Ec threshold
			minReverseField := 0.7 * wc.EcField
			reverseStep := 0.05 * wc.EcField
			maxReverseField := 1.5 * wc.EcField // Allow higher fields for L-K gradual switching

			// After multiple overshoots, use even finer steps
			if wc.OvershootCount > 2 {
				reverseStep = 0.03 * wc.EcField
			}
			if wc.MinStep > reverseStep {
				reverseStep = wc.MinStep
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
		if wc.MinStep > minStep {
			minStep = wc.MinStep
		}
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

	// Stuck or no-improvement escalation: boost step and relax bounds.
	if wc.StuckCount >= 3 || wc.NoImproveCount >= 3 {
		wc.raiseStepFloor(0.12*wc.EcField, "stuck/no-improve>=3")
		// Don't hard-reset bounds to 0. That shows up in logs/CSVs as "VMin reset",
		// and it also discards useful convergence progress. Instead:
		// - Keep (or set) a non-zero lower bound at the last attempted magnitude.
		// - Drop the upper bound (set VMax to MaxField, mark VMaxSet=false) so the
		//   next steps can grow beyond any stale bracket.
		if absCur := math.Abs(wc.CurrentField); absCur > 0 {
			wc.VMin = absCur
			wc.VMinSet = true
		}
		wc.VMax = wc.MaxField
		wc.VMaxSet = false
		wc.StuckCount = 0
		wc.NoImproveCount = 0
		log.Printf("ISPP STUCK: boosting step (floor=%.3f×Ec) and relaxing bounds to [%.3f, %.3f]×Ec",
			wc.stepFloor/wc.EcField, wc.VMin/wc.EcField, wc.VMax/wc.EcField)
	}

	// Enforce hard minimum step (and dynamic step floor) to avoid tiny increments.
	minStep := wc.MinStep
	if wc.stepFloor > minStep {
		minStep = wc.stepFloor
	}
	if minStep > 0 && stepSize < minStep {
		stepSize = minStep
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
