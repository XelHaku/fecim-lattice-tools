//go:build legacy_fyne

package gui

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/logging"
	"fyne.io/fyne/v2"
)

// simulationLoop runs the main simulation loop at ~60 FPS
// simulationLoop runs the main simulation loop at ~60 FPS with adaptive physics stepping
func (a *App) simulationLoop() {
	// Recover from closed-channel panics during shutdown.
	// Stop() closes the UI update channel, but the sim goroutine may still be
	// mid-frame between the running.Load() check and the channel send.
	defer func() {
		if r := recover(); r != nil {
			// Expected during shutdown; silently absorb.
		}
	}()

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS targeting
	defer ticker.Stop()

	lastTime := time.Now()
	perfEnabled := logging.IsVerbose(logging.VerbosityDebug)
	perfWindowStart := time.Time{}
	perfFrames := 0
	perfSubSteps := 0
	perfDtCount := 0
	perfDtMin := 0.0
	perfDtMax := 0.0
	perfDtSum := 0.0
	perfPhysics := time.Duration(0)
	perfSolver := time.Duration(0)
	perfRender := time.Duration(0)
	perfEngine := "L-K"
	perfDtMinHits := 0

	// Adaptive Time-Stepping Constants
	const (
		dtMax         = 0.025 // 25ms cap to prevent explosion after pause
		dtNominal     = 1e-4  // 0.1ms nominal physics step (good for standard loop)
		dtMin         = 1e-6  // 1µs minimum step near critical field (Ec)
		lkMaxSubSteps = 400

		// ecProximityFrac is the fraction of Ec used as the switching-region threshold.
		// When |E - Ec| < ecProximityFrac*Ec, the solver uses dtMin to capture sharp
		// switching dynamics. 0.1 means ±10% of Ec (≈0.1 MV/cm for typical HZO).
		// Material-relative scaling ensures correct behavior for high-Ec materials
		// like AlScN (Ec ≈ 5 MV/cm) where a fixed 1e7 V/m threshold is too narrow.
		ecProximityFrac = 0.1
	)

	resetPerfWindow := func(now time.Time) {
		perfWindowStart = now
		perfFrames = 0
		perfSubSteps = 0
		perfDtCount = 0
		perfDtMin = math.MaxFloat64
		perfDtMax = 0
		perfDtSum = 0
		perfPhysics = 0
		perfSolver = 0
		perfRender = 0
		perfDtMinHits = 0
	}

	recordPhysicsStep := func(step float64) {
		if perfEnabled {
			if step < perfDtMin {
				perfDtMin = step
			}
			if step > perfDtMax {
				perfDtMax = step
			}
			perfDtSum += step
			perfDtCount++
			if step <= dtMin*1.001 {
				perfDtMinHits++
			}
			stepStart := time.Now()
			solverDur := a.updatePhysics(step, true)
			perfPhysics += time.Since(stepStart)
			perfSolver += solverDur
			return
		}
		a.updatePhysics(step, false)
	}

	logPerfWindow := func(now time.Time) {
		if !perfEnabled || perfWindowStart.IsZero() {
			return
		}
		if now.Sub(perfWindowStart) < time.Second {
			return
		}
		if perfFrames == 0 || perfDtCount == 0 {
			resetPerfWindow(now)
			return
		}
		avgSubSteps := float64(perfSubSteps) / float64(perfFrames)
		dtMean := perfDtSum / float64(perfDtCount)
		physicsMs := perfPhysics.Seconds() * 1000.0
		solverMs := perfSolver.Seconds() * 1000.0
		renderMs := perfRender.Seconds() * 1000.0
		totalMs := physicsMs + renderMs
		physicsShare := 0.0
		if totalMs > 0 {
			physicsShare = (physicsMs / totalMs) * 100.0
		}
		solverShare := 0.0
		if physicsMs > 0 {
			solverShare = (solverMs / physicsMs) * 100.0
		}
		minShare := 0.0
		if perfDtCount > 0 {
			minShare = float64(perfDtMinHits) / float64(perfDtCount) * 100.0
		}
		log.Debug("LK perf (%s): frames=%d avgSubSteps=%.1f dtMin=%.3e dtMean=%.3e dtMax=%.3e dtMinHits=%d (%.1f%%) solverMs=%.2f physicsMs=%.2f renderMs=%.2f solverShare=%.1f%% physicsShare=%.1f%%",
			perfEngine, perfFrames, avgSubSteps, perfDtMin, dtMean, perfDtMax, perfDtMinHits, minShare, solverMs, physicsMs, renderMs, solverShare, physicsShare)
		resetPerfWindow(now)
	}

	if perfEnabled {
		resetPerfWindow(time.Now())
	}

	for a.running.Load() {
		<-ticker.C

		// Stop can be requested while we're sleeping on the ticker. Exit without
		// doing another full frame of physics.
		if !a.running.Load() {
			return
		}

		now := time.Now()
		frameDtRaw := now.Sub(lastTime).Seconds()
		lastTime = now

		if a.paused.Load() {
			continue
		}

		// Lock ONCE for the entire frame's physics burst
		a.mu.Lock()
		timeScale := a.timeScale
		if timeScale <= 0 {
			timeScale = 1.0
		}
		frameDt := frameDtRaw * timeScale
		if frameDt > dtMax {
			frameDt = dtMax
		}

		// --- Adaptive Sub-Stepping Loop ---
		remainingDt := frameDt

		// Safety break to prevent infinite loops if calculation is too slow
		const maxSubSteps = lkMaxSubSteps
		subSteps := 0
		fastPreisach := a.physicsEngine == PhysicsPreisach
		if fastPreisach {
			perfEngine = "Preisach"
		} else {
			perfEngine = "L-K"
		}

		matEc := 0.0
		if a.material != nil {
			matEc = a.material.Ec
		}

		if fastPreisach {
			// Quasi-static visualization still needs enough points to avoid straight segments.
			const preisachSubSteps = 24
			steps := preisachSubSteps
			if steps < 1 {
				steps = 1
			}
			for i := 0; i < steps && remainingDt > 0; i++ {
				// Stop requests should abort calibration/simulation immediately on tab/window changes.
				if !a.running.Load() {
					steps = i
					break
				}
				// Allow modal-driven pauses to interrupt a long frame immediately.
				if a.paused.Load() {
					steps = i
					break
				}
				step := remainingDt / float64(steps-i)
				recordPhysicsStep(step)
				remainingDt -= step
			}
			subSteps = steps
		} else {
			for remainingDt > 0 && subSteps < maxSubSteps {
				// Stop requests should abort calibration/simulation immediately on tab/window changes.
				if !a.running.Load() {
					break
				}
				// Allow modal-driven pauses to interrupt a long LK substep burst immediately.
				if a.paused.Load() {
					break
				}
				// Determine step size based on physics state
				// If E-field is near Ec, use smaller steps to capture switching dynamics.

				currentStep := dtNominal
				// Check proximity to Ec (switching region)
				// Switching happens at +Ec (increasing) and -Ec (decreasing)
				// But effective Ec varies. Use material Ec as baseline proxy.
				if matEc > 0 {
					distPlus := math.Abs(a.electricField - matEc)
					distMinus := math.Abs(a.electricField + matEc)
					minDist := math.Min(distPlus, distMinus)

					// Use material-relative threshold: ecProximityFrac × Ec.
					// This scales correctly for any ferroelectric (HZO, AlScN, PZT, etc.).
					threshold := ecProximityFrac * matEc
					if minDist < threshold {
						currentStep = dtMin
					}
				}

				// Ensure we can consume the remaining time within our substep budget.
				remainingBudget := maxSubSteps - subSteps
				if remainingBudget > 0 {
					minStep := remainingDt / float64(remainingBudget)
					if currentStep < minStep {
						currentStep = minStep
					}
				}

				// Don't step past the frame time
				if currentStep > remainingDt {
					currentStep = remainingDt
				}

				recordPhysicsStep(currentStep)
				remainingDt -= currentStep
				subSteps++
			}
		}

		a.mu.Unlock()

		if !a.running.Load() {
			return
		}

		// Update UI once per frame with the final state (without holding a.mu).
		renderStart := time.Time{}
		if perfEnabled {
			renderStart = time.Now()
		}
		a.updateUI()
		if perfEnabled {
			perfRender += time.Since(renderStart)
			perfFrames++
			perfSubSteps += subSteps
			logPerfWindow(time.Now())
		}
	}
}

// updatePhysics handles the physics and state transitions.
// MUST be called with a.mu held.
// Returns time spent inside the active physics engine when perfEnabled is true.
func (a *App) updatePhysics(dt float64, perfEnabled bool) time.Duration {
	// a.mu.Lock() -> Removed, caller holds lock

	mat := a.material
	if mat == nil {
		return 0
	}

	Ec := a.effectiveEc()
	Emax := Ec * 2.5 // Scale for full loop traversal
	a.simTime += dt
	phaseDuration := 1.0 / a.frequency
	rampRate := 4.0 * Emax * a.frequency

	// Generate E-field based on waveform
	if a.waveform == WaveformManual {
		if a.manualAnimating {
			a.manualPhaseTime += dt
			targetLevel := a.manualTargetLevel // 1-indexed (1-N)
			midLevel := a.numLevels / 2

			switch a.manualPhase {
			case 0: // PREP - saturate to known state opposite to target
				// For Preisach engine skip PREP (same logic as WRD wrdSkipPrep)
				if !a.useLKSolver() {
					if a.writeController != nil {
						a.writeController.PulseDuration = phaseDuration * 0.4
						a.writeController.Start(targetLevel, false)
					}
					a.manualPhase = 1
					a.manualPhaseTime = 0
					break
				}

				prepMag := Emax
				prepE := prepMag
				if targetLevel > midLevel {
					prepE = -prepMag // Upper targets: bias NEGATIVE first
				}

				diff := prepE - a.electricField
				step := rampRate * dt
				if math.Abs(diff) < step {
					a.electricField = prepE
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				biasRef := mat.Pr
				if biasRef == 0 {
					biasRef = mat.Ps
				}
				biased := true
				if biasRef > 0 {
					threshold := 0.2 * biasRef
					if prepE < 0 {
						biased = a.polarization <= -threshold
					} else {
						biased = a.polarization >= threshold
					}
				}

				if a.manualPhaseTime > phaseDuration*0.3 && math.Abs(a.electricField-prepE) < 0.01*Emax && biased {
					if a.writeController != nil {
						a.writeController.PulseDuration = phaseDuration * 0.4
						a.writeController.Start(targetLevel, true)
					}
					a.manualPhase = 1
					a.manualPhaseTime = 0
				}

			case 1: // WRITE - delegate to WriteController (same as WRD case 2)
				currentLevel := a.discreteLevel + 1
				guardSign := 0
				if a.material != nil && a.wrdGuardFrac > 0 {
					ps := a.material.Ps
					if ps == 0 {
						ps = a.material.Pr
					}
					bins := ferroelectric.NewLevelBins(ps, a.numLevels, a.wrdRangeFrac, a.wrdGuardFrac)
					level, inError, delta := bins.LevelForP(a.polarization)
					currentLevel = level
					if inError && level == targetLevel {
						if delta > 0 {
							guardSign = 1
						} else if delta < 0 {
							guardSign = -1
						}
					}
				}

				var targetField float64
				done := false
				if a.writeController != nil {
					targetField, done = a.writeController.Update(dt, a.electricField, currentLevel, guardSign)
				}

				// Hard pulse timeout guard
				if !done && a.writeController != nil {
					pulseLimit := isppPulseLimit(a.isppMaxPulses)
					totalPulses := a.writeController.TotalPulses + a.writeController.PulseCount
					if totalPulses >= pulseLimit {
						done = true
					}
				}

				// Apply voltage ramp
				step := rampRate * 1.5 * dt
				diff := targetField - a.electricField
				if math.Abs(diff) < step {
					a.electricField = targetField
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				if done {
					// Learn from successful ISPP write
					if a.writeController != nil && a.writeController.State == controller.StateSuccess {
						learnedE := a.writeController.CurrentField
						targetIdx := targetLevel - 1
						Ec := mat.Ec
						if a.calibManager != nil {
							if targetLevel > midLevel {
								a.calibManager.UpdateCalibrationUp(targetIdx, 0, Ec)
								a.calibManager.CalibrationUp[targetIdx] = learnedE
								a.calibrationUp[targetIdx] = learnedE
							} else {
								a.calibManager.UpdateCalibrationDown(targetIdx, 0, Ec)
								a.calibManager.CalibrationDown[targetIdx] = learnedE
								a.calibrationDown[targetIdx] = learnedE
							}
						}
					}
					a.manualPhase = 2
					a.manualPhaseTime = 0
				}

			case 2: // DISPLAY - return to zero, show result
				step := rampRate * 0.4 * dt
				if math.Abs(a.electricField) < step {
					a.electricField = 0
				} else if a.electricField > 0 {
					a.electricField -= step
				} else {
					a.electricField += step
				}

				if a.manualPhaseTime > phaseDuration*0.4 && math.Abs(a.electricField) < 0.01*Emax {
					finalLevel := a.discreteLevel + 1
					log.Printf("MANUAL ANIMATION COMPLETE: target=%d, final=%d", targetLevel, finalLevel)
					a.addLogEntry(fmt.Sprintf("RESULT L%d → L%d", targetLevel, finalLevel))
					a.manualAnimating = false
					a.manualPhase = 0
					a.manualPhaseTime = 0
				}
			}
		}
	} else if a.autoMode {
		Emax := mat.Ec * 2 // Default auto drive
		driveMax := Emax
		if a.physicsEngine == PhysicsLandau {
			// L-K needs a stronger drive to traverse the full loop.
			driveMax = mat.Ec * 2.5
		}
		// Wrap phase to prevent floating-point precision loss over long times
		phase := math.Mod(2*math.Pi*a.frequency*a.simTime, 2*math.Pi)

		switch a.waveform {
		case WaveformSine:
			a.electricField = driveMax * math.Sin(phase)
		case WaveformTriangle:
			p := phase / (2 * math.Pi)
			if p < 0.25 {
				a.electricField = driveMax * (4 * p)
			} else if p < 0.75 {
				a.electricField = driveMax * (2 - 4*p)
			} else {
				a.electricField = driveMax * (4*p - 4)
			}
		case WaveformWriteReadDemo:
			// Write/read demo with deterministic pre-bias + ISPP:
			// - Always pre-saturate to the opposite polarity of the target (matches headless).
			// - Skip HOLD_RESET to avoid depolarization relaxation.
			// - WriteController handles APPLY/WAIT/VERIFY/RESETTING internally.
			//
			// Phase mapping:
			// 0 = PREP (saturate opposite polarity)
			// 1 = HOLD_PREP (unused; kept for compatibility)
			// 2 = WRITE (delegated to WriteController)
			// 5 = DISPLAY (return to zero, pick next target)

			a.wrdPhaseTimer += dt
			phaseDuration := 1.0 / a.frequency
			rampRate := 3.0 * Emax * a.frequency
			Ec := mat.Ec // Use local copy for thread safety
			midLevel := a.numLevels / 2

			// Apply queued target at the start of PREP or WRITE so UI and controller stay aligned.
			if (a.wrdPhase == 0 || a.wrdPhase == 2) && a.wrdPhaseTimer <= dt && a.wrdNextTargetLevel > 0 {
				nextTarget := a.wrdNextTargetLevel
				a.wrdTargetLevel = nextTarget
				a.wrdNextTargetLevel = 0
				// Reset cumulative controller state for the new target so TotalPulses
				// from prior targets don't trip the hard timeout guard immediately.
				if a.writeController != nil {
					a.writeController.ResetState()
				}
				ctrlTarget := 0
				if a.writeController != nil {
					ctrlTarget = a.writeController.TargetLevel
				}
				if logging.IsVerbose(logging.VerbosityDebug) {
					log.Printf("WRD TARGET APPLY: active=%d queued=%d phase=%d level=%d ctrlTarget=%d E=%.3f MV/cm P=%.2f µC/cm²",
						a.wrdTargetLevel, nextTarget, a.wrdPhase, a.discreteLevel+1, ctrlTarget, a.electricField/1e8, a.polarization*100)
				}
				forcePrep := nextTarget <= 3 || nextTarget >= (a.numLevels-2)
				if a.wrdPhase == 2 && forcePrep {
					// Near saturation targets are more stable with a deterministic PREP.
					a.wrdPhase = 0
					a.wrdPhaseTimer = 0
				} else if a.wrdPhase == 2 && a.writeController != nil {
					// Skip PREP: start controller directly from current state.
					a.writeController.PulseDuration = phaseDuration * 0.4
					a.writeController.Start(a.wrdTargetLevel, false)
					a.wrdStartLevel = a.discreteLevel + 1
					a.wrdWriteStartP = a.polarization * 100
					if logging.IsVerbose(logging.VerbosityDebug) {
						log.Printf("WRD SKIP PREP: start=%d target=%d P=%.2f µC/cm²",
							a.wrdStartLevel, a.wrdTargetLevel, a.wrdWriteStartP)
					}
				}
			}
			targetLevel := a.wrdTargetLevel // 1-indexed
			// Note: startLevel (a.wrdStartLevel) no longer used for direction - we use absolute position

			switch a.wrdPhase {
			case 0: // PREP (wake-up) - prepare starting state before writing
				forcePrep := targetLevel <= 3 || targetLevel >= (a.numLevels-2)
				if a.wrdSkipPrep && !forcePrep {
					// Skip PREP entirely and jump to WRITE.
					if a.writeController != nil {
						a.writeController.PulseDuration = phaseDuration * 0.4
						a.writeController.Start(targetLevel, false)
					}
					a.wrdPhase = 2
					a.wrdPhaseTimer = 0
					break
				}
				// Pre-bias to full saturation field to match calibration procedure.
				// Calibration tests each level from Reset → ±Emax → 0 → testE → 0,
				// so PREP must reach ±Emax to produce matching Preisach states.
				prepMag := Emax
				prepE := prepMag
				if targetLevel > midLevel {
					prepE = -prepMag // Upper targets: bias NEGATIVE first
				}
				a.wrdPrepE = prepE

				// Capture start-of-PREP state for logging
				if a.wrdPhaseTimer <= dt {
					// Reset Preisach stack to clear residual turning points from
					// previous ISPP cycle. Without this, the stack retains history
					// at up to ±2.5*Ec that PREP cannot wipe out, causing subsequent
					// cycles to converge to wrong levels. This matches the calibration
					// procedure which starts each measurement from a Reset() state.
					if a.preisach != nil {
						a.preisach.Reset()
					}

					a.wrdResetStartP = a.polarization * 100
					a.wrdStartLevel = a.discreteLevel + 1
					a.wrdWriteStartP = a.polarization * 100
					if logging.IsVerbose(logging.VerbosityDebug) {
						log.Printf("WRD CYCLE START: cycle=%d | startLevel=%d | target=%d | P=%.2f µC/cm²",
							a.wrdTotalWrites+1, a.wrdStartLevel, a.wrdTargetLevel, a.wrdWriteStartP)
					}
				}

				diff := prepE - a.electricField
				step := rampRate * dt
				if math.Abs(diff) < step {
					a.electricField = prepE
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				biasRef := mat.Pr
				if biasRef == 0 {
					biasRef = mat.Ps
				}
				biasThreshold := 0.0
				if biasRef > 0 {
					biasThreshold = 0.2 * biasRef
				}

				satRef := mat.Ps
				if satRef == 0 {
					satRef = mat.Pr
				}
				satThreshold := 0.0
				if satRef > 0 {
					satThreshold = 0.5 * satRef
				}

				biased := biasThreshold == 0
				saturated := satThreshold == 0
				if prepE < 0 {
					if biasThreshold > 0 {
						biased = a.polarization <= -biasThreshold
					}
					if satThreshold > 0 {
						saturated = a.polarization <= -satThreshold
					}
				} else {
					if biasThreshold > 0 {
						biased = a.polarization >= biasThreshold
					}
					if satThreshold > 0 {
						saturated = a.polarization >= satThreshold
					}
				}
				ready := biased
				if forcePrep {
					ready = saturated
				}

				// Transition when field reached AND bias criteria satisfied
				if a.wrdPhaseTimer > phaseDuration*0.25 && math.Abs(a.electricField-prepE) < 0.01*Emax && ready {
					// Capture end-of-PREP state for logging
					a.wrdResetEndP = a.polarization * 100 // Convert to µC/cm²
					a.wrdResetEndLvl = a.discreteLevel + 1
					if logging.IsVerbose(logging.VerbosityDebug) {
						log.Printf("WRD PHASE 0→2: PREP done | E=%.3f MV/cm | P=%.2f→%.2f µC/cm² | L=%d→%d | target=%d",
							a.electricField/1e8, a.wrdResetStartP, a.wrdResetEndP, a.wrdStartLevel, a.wrdResetEndLvl, targetLevel)
					}

					// Initialize WriteController directly (skip HOLD_RESET)
					a.writeController.PulseDuration = phaseDuration * 0.4
					a.writeController.Start(targetLevel, saturated)

					a.wrdPhase = 2
					a.wrdPhaseTimer = 0
				}

			case 1: // SETTLE - return to zero (now at known saturated remanent state)
				// HOLD_RESET removed to prevent polarization relaxation; jump to WRITE.
				a.wrdPhase = 2
				a.wrdPhaseTimer = 0

			case 2: // WRITE - Delegated to WriteController
				// Refactored to use pkg/controller/WriteController
				// This consolidates Write, Wait, Verify, and Retry logic

				if a.wrdSkipPrep && a.wrdPhaseTimer <= dt && a.writeController != nil {
					if a.writeController.State == controller.StateIdle || a.writeController.TargetLevel != targetLevel {
						a.writeController.PulseDuration = phaseDuration * 0.4
						a.writeController.Start(targetLevel, false)
						a.wrdStartLevel = a.discreteLevel + 1
						a.wrdWriteStartP = a.polarization * 100
						if logging.IsVerbose(logging.VerbosityDebug) {
							log.Printf("WRD SKIP PREP: start=%d target=%d P=%.2f µC/cm²",
								a.wrdStartLevel, a.wrdTargetLevel, a.wrdWriteStartP)
						}
					}
				}

				currentLevel := a.discreteLevel + 1
				guardSign := 0
				if a.material != nil && a.wrdGuardFrac > 0 {
					ps := a.material.Ps
					if ps == 0 {
						ps = a.material.Pr
					}
					bins := ferroelectric.NewLevelBins(ps, a.numLevels, a.wrdRangeFrac, a.wrdGuardFrac)
					level, inError, delta := bins.LevelForP(a.polarization)
					currentLevel = level
					if inError && level == targetLevel {
						if delta > 0 {
							guardSign = 1
						} else if delta < 0 {
							guardSign = -1
						}
						if guardSign != 0 && logging.IsVerbose(logging.VerbosityDebug) && a.writeController != nil && a.writeController.PulseCount != a.wrdLastGuardLogPulse {
							a.wrdLastGuardLogPulse = a.writeController.PulseCount
							log.Printf("WRD GUARD: level=%d target=%d delta=%.4g (guard=%.2f)",
								level, targetLevel, delta, a.wrdGuardFrac)
						}
					}
				}
				var targetField float64
				var done bool
				if a.writeController != nil {
					targetField, done = a.writeController.Update(dt, a.electricField, currentLevel, guardSign)
				}

				if a.writeController != nil && (a.wrdLastControllerState != a.writeController.State || a.wrdLastControllerPulse != a.writeController.PulseCount) {
					if logging.IsVerbose(logging.VerbosityDebug) {
						log.Printf("WRD ISPP STATE: state=%s pulse=%d target=%d level=%d P=%.2f µC/cm² E=%.3f MV/cm targetE=%.3f×Ec pulseE=%.3f×Ec overshoots=%d",
							a.writeController.State,
							a.writeController.PulseCount,
							targetLevel,
							currentLevel,
							a.polarization*100,
							a.electricField/1e8,
							targetField/mat.Ec,
							a.writeController.CurrentField/mat.Ec,
							a.writeController.OvershootCount)
					}
					a.wrdLastControllerState = a.writeController.State
					a.wrdLastControllerPulse = a.writeController.PulseCount
				}
				if a.simTime-a.wrdLastProgressLog > 0.5 && a.writeController != nil {
					if logging.IsVerbose(logging.VerbosityDebug) {
						log.Printf("WRD ISPP PROGRESS: target=%d level=%d P=%.2f µC/cm² E=%.3f MV/cm state=%s pulse=%d targetE=%.3f×Ec pulseE=%.3f×Ec bounds=[%.3f, %.3f]×Ec",
							targetLevel,
							currentLevel,
							a.polarization*100,
							a.electricField/1e8,
							a.writeController.State,
							a.writeController.PulseCount,
							targetField/mat.Ec,
							a.writeController.CurrentField/mat.Ec,
							a.writeController.VMin/mat.Ec,
							a.writeController.VMax/mat.Ec)
					}
					a.wrdLastProgressLog = a.simTime
				}

				// Hard ISPP timeout guard: force completion after global pulse budget.
				pulseLimit := isppPulseLimit(a.isppMaxPulses)
				totalPulses := a.writeController.TotalPulses + a.writeController.PulseCount
				if !done && totalPulses >= pulseLimit {
					bestLevel := a.writeController.BestVerifyLevel
					if bestLevel <= 0 {
						bestLevel = currentLevel
					}
					a.wrdReadLevel = bestLevel
					a.wrdPhase = 5
					a.wrdPhaseTimer = 0
					a.wrdTotalWrites++
					log.Printf("WRD ISPP TIMEOUT: target=%d pulse=%d/%d state=%s bestLevel=%d (forced complete)",
						targetLevel, totalPulses, pulseLimit, a.writeController.State, bestLevel)
					done = true
				}

				// Autonomous runtime recalibration trigger (overshoots or too many pulses)
				if a.autoRecalibrate && !a.recalibratePending {
					if a.writeController.OvershootCount >= a.recalibrateOvershootMax {
						a.recalibratePending = true
						a.recalibrateReason = fmt.Sprintf("overshoots=%d target=%d", a.writeController.OvershootCount, targetLevel)
						log.Printf("AUTO RECAL scheduled: %s", a.recalibrateReason)
					} else if a.writeController.PulseCount >= a.recalibratePulseMax {
						a.recalibratePending = true
						a.recalibrateReason = fmt.Sprintf("pulses=%d target=%d", a.writeController.PulseCount, targetLevel)
						log.Printf("AUTO RECAL scheduled: %s", a.recalibrateReason)
					}
				}

				// Apply Voltage Ramp (Controller outputs target, Simulation handles physics ramp)
				step := rampRate * 1.5 * dt // Faster ramp for write pulses
				diff := targetField - a.electricField
				if math.Abs(diff) < step {
					a.electricField = targetField
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				if done {
					switch a.writeController.State {
					case controller.StateSuccess:
						// SUCCESS: Proceed to DISPLAY phase
						a.wrdPhase = 5
						a.wrdPhaseTimer = 0

						// Logging and Metrics
						a.wrdTotalWrites++
						a.wrdSuccessWrites++
						successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100

						log.Printf("WRD SUCCESS via Controller: target=%d, tries=%d", targetLevel, a.writeController.RetryCount)
						if shouldForceResetAfterISPP(a.writeController, false) {
							a.wrdForceReset = true
						}
						if targetLevel > a.wrdStartLevel {
							a.wrdLastBranch = 1
						} else if targetLevel < a.wrdStartLevel {
							a.wrdLastBranch = -1
						}

						// Log result (replaces Case 4 success log)
						log.Printf("WRD PHASE 4→5: TARGET HIT | L_read=%d L_target=%d | rate=%.1f%% (%d/%d)",
							a.writeController.LastVerifyLevel, targetLevel,
							successRate, a.wrdSuccessWrites, a.wrdTotalWrites)

						// CRITICAL FIX: Learn from the Servo using CalibrationManager
						// Instead of naïve 100% overwrite, we use binary search + monotonicity
						learnedE := a.writeController.CurrentField
						Ec := mat.Ec
						targetIdx := targetLevel - 1 // 0-indexed for CM

						if a.calibManager != nil {
							if targetLevel > midLevel {
								// Ascending (written from reset negative)
								a.calibManager.UpdateCalibrationUp(targetIdx, 0, Ec)
								// Override the midpoint with the actual successful servo voltage to anchor the search
								a.calibManager.CalibrationUp[targetIdx] = learnedE
								a.calibrationUp[targetIdx] = learnedE
							} else {
								// Descending (written from reset positive)
								a.calibManager.UpdateCalibrationDown(targetIdx, 0, Ec)
								a.calibManager.CalibrationDown[targetIdx] = learnedE
								a.calibrationDown[targetIdx] = learnedE
							}
							log.Printf("CALIB LEARN: target=%d learnedE=%.3f×Ec (updated via CM)", targetLevel, learnedE/Ec)
						}

					case controller.StateForceReset:
						// Retry limit reached. Only allow full RESET on severe overshoot.
						if shouldForceResetAfterISPP(a.writeController, false) {
							log.Printf("WRD RETRY LIMIT: Full RESET to saturation (retry %d, overshootDelta=%d)", a.writeController.RetryCount, a.writeController.MaxOvershootDelta)
							a.wrdPhase = 0 // RESET phase
							a.wrdPhaseTimer = 0
						} else {
							bestLevel := a.writeController.BestVerifyLevel
							if bestLevel <= 0 {
								bestLevel = a.writeController.LastVerifyLevel
							}
							a.wrdReadLevel = bestLevel
							a.wrdPhase = 5
							a.wrdPhaseTimer = 0
							a.wrdTotalWrites++
							log.Printf("WRD RETRY LIMIT: suppressing reset (overshootDelta=%d<=3), forced complete at level=%d target=%d", a.writeController.MaxOvershootDelta, bestLevel, targetLevel)
						}

					case controller.StateFailed:
						// Generic failure (shouldn't happen with infinite retries enabled, but just in case)
						log.Printf("WRD FAILED: Controller gave up on target %d", targetLevel)
						a.wrdPhase = 5 // Display anyway
						a.wrdPhaseTimer = 0
						a.wrdTotalWrites++ // Count as attempt
					}
				}

				// Cases 3 (HOLD) and 4 (VERIFY) are now internal to the Controller
				// We stay in Case 2 until 'done' is true.

			case 3, 4:
				// Legacy states - should not be reached with new controller
				// Jump to 2 just in case
				a.wrdPhase = 2
			case 5: // DISPLAY phase - return to zero, show result
				step := rampRate * 0.4 * dt
				if math.Abs(a.electricField) < step {
					a.electricField = 0
				} else if a.electricField > 0 {
					a.electricField -= step
				} else {
					a.electricField += step
				}
				// Final read at E=0 to verify the level (stable state).
				if math.Abs(a.electricField) < 0.01*Ec {
					a.wrdReadLevel = a.discreteLevel + 1
				}
				// Transition to next cycle
				if a.wrdPhaseTimer > phaseDuration*0.6 {
					// Runtime auto recalibration (run between targets to avoid state disruption)
					if a.autoRecalibrate && a.recalibratePending {
						log.Printf("AUTO RECAL start: %s", a.recalibrateReason)
						a.calibrateLevelsAtTemperature(a.currentTemperature())
						if err := a.saveCalibration(); err != nil {
							log.Printf("Warning: failed to save calibration: %v", err)
						}
						a.needsCalibration = false
						a.recalibratePending = false
						a.recalibrateReason = ""
					}

					// Add comparison callout every 5 cycles
					if a.wrdTotalWrites > 0 && a.wrdTotalWrites%5 == 0 {
						fecimEnergy := a.wrdTotalEnergyfJ / 1000 // pJ
						// NOTE: Ignore the 10M× conference claim here; peer-reviewed literature reports 25-100× (Samsung Nature 2025).
						nandEquiv := fecimEnergy * 50   // 25-100× better (conservative: use 50)
						dramEquiv := fecimEnergy * 1000 // 1000× worse
						bitsStored := float64(a.wrdTotalWrites) * a.wrdBitsStored
						a.addLogEntry("━━ ENERGY COMPARISON ━━")
						a.addLogEntry(fmt.Sprintf("FeCIM: %.0f pJ total", fecimEnergy))
						a.addLogEntry(fmt.Sprintf("NAND:  %.0f pJ (~50× est.)", nandEquiv))
						a.addLogEntry(fmt.Sprintf("DRAM:  %.0f pJ (~1000× est.)", dramEquiv))
						a.addLogEntry(fmt.Sprintf("Bits stored: %.0f (%.2f×binary)", bitsStored, a.wrdBitsStored))
						a.addLogEntry("━━━━━━━━━━━━━━━━━━━━━━")
					}

					// Milestone celebrations
					switch a.wrdTotalWrites {
					case 10:
						a.addLogEntry(fmt.Sprintf("★★ 10 ops! ~%.0f bits stored ★★", 10*a.wrdBitsStored))
					case 25:
						a.addLogEntry(fmt.Sprintf("★★★ 25 ops! ~%.0f bits stored ★★★", 25*a.wrdBitsStored))
					case 50:
						a.addLogEntry(fmt.Sprintf("★★★★ 50 ops! ~%.0f bits stored ★★★★", 50*a.wrdBitsStored))
						a.addLogEntry(fmt.Sprintf("Binary would need %.0f cells!", 50*a.wrdBitsStored))
						a.addLogEntry("FeCIM: only 50 cells! (5× denser)")
					case 100:
						a.addLogEntry("★★★★★ 100 OPERATIONS! ★★★★★")
						a.addLogEntry(fmt.Sprintf("~%.0f bits in 100 FeCIM cells", 100*a.wrdBitsStored))
						a.addLogEntry(fmt.Sprintf("Binary: %.0f cells needed!", 100*a.wrdBitsStored))
						successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100
						a.addLogEntry(fmt.Sprintf("Accuracy: %.0f%%", successRate))
					}

					// Pick new target - prefer same-branch moves to avoid unnecessary PREP.
					nextTarget := a.wrdNextTargetLevel
					if nextTarget == 0 {
						minLevel := 2
						maxLevel := a.numLevels - 1
						currentLevel := a.discreteLevel + 1
						preferDir := a.wrdLastBranch
						if preferDir == 0 {
							if currentLevel <= a.numLevels/2 {
								preferDir = 1
							} else {
								preferDir = -1
							}
						}
						nextTarget = a.wrdTargetLevel
						if preferDir > 0 {
							low := currentLevel + 1
							high := maxLevel
							if low <= high {
								nextTarget = rand.Intn(high-low+1) + low
							} else {
								preferDir = -1
							}
						}
						if preferDir < 0 {
							low := minLevel
							high := currentLevel - 1
							if low <= high {
								nextTarget = rand.Intn(high-low+1) + low
							} else {
								// At bottom edge — flip to ascending
								low = currentLevel + 1
								high = maxLevel
								if low <= high {
									nextTarget = rand.Intn(high-low+1) + low
								}
							}
						}
						a.wrdNextTargetLevel = nextTarget
					}
					currentLevel := a.discreteLevel + 1
					nextTarget = a.wrdNextTargetLevel
					nextBranch := 0
					if nextTarget > currentLevel {
						nextBranch = 1
					} else if nextTarget < currentLevel {
						nextBranch = -1
					}
					forcePrep := nextTarget <= 3 || nextTarget >= (a.numLevels-2)
					usePrep := forcePrep || a.wrdForceReset || (!a.wrdSkipPrep && (a.wrdLastBranch != 0 && a.wrdLastBranch != nextBranch))
					a.wrdForceReset = false
					// NOTE: Don't clear history - let the trail accumulate to show full hysteresis loop
					// Spike detection in plot widget handles any discontinuities
					// Use PREP only on branch change or overshoot.
					if usePrep {
						a.wrdPhase = 0
					} else {
						a.wrdPhase = 2
					}
					a.wrdPhaseTimer = 0
					a.wrdCycleEnergy = 0  // Reset energy accumulator for next cycle
					a.isppTotalPulses = 0 // Reset ISPP pulse counter for next target

					// Reset ISPP widget to idle state for new target
					if a.isppWidget != nil {
						// UI update must be on main thread
						fyne.Do(func() {
							a.isppWidget.SetAnimationState(0, 0, a.discreteLevel+1, 0, false)
						})
					}
				}

			case 6: // Legacy BOOST phase - redirect to Controller
				// The generic WriteController handles retries/boost internally or via re-entry to Case 2
				a.wrdPhase = 2
			}
		case WaveformTimeResolved:
			if !a.timeResAnimating {
				a.addLogEntry("NOTE: Time-Resolved Switching")
				a.addLogEntry("requires legacy physics engine.")
				a.timeResAnimating = true // To prevent repeated logging
			}
		}
	}

	// Update physics
	prevP := a.polarization
	var solverDur time.Duration
	if a.useLKSolver() {
		if a.lkSolver != nil {
			if perfEnabled {
				stepStart := time.Now()
				a.polarization = a.lkSolver.Step(a.electricField, dt)
				solverDur = time.Since(stepStart)
			} else {
				a.polarization = a.lkSolver.Step(a.electricField, dt)
			}
			if math.IsNaN(a.polarization) || math.IsInf(a.polarization, 0) {
				fallback := prevP
				if math.IsNaN(fallback) || math.IsInf(fallback, 0) {
					fallback = a.lkDefaultPolarization()
				}
				a.polarization = fallback
				a.lkSolver.SetState(fallback)
				if log != nil {
					log.Printf("LK solver returned invalid polarization; reset to %.3e", fallback)
				}
			}
		} else {
			a.polarization = 0
		}
		if mat.Ps != 0 {
			a.normalizedP = a.polarization / mat.Ps
		} else {
			a.normalizedP = 0
		}
		if a.normalizedP > 1 {
			a.normalizedP = 1
		} else if a.normalizedP < -1 {
			a.normalizedP = -1
		}
	} else {
		if a.preisach != nil {
			if perfEnabled {
				stepStart := time.Now()
				a.polarization = a.preisach.Update(a.electricField)
				solverDur = time.Since(stepStart)
			} else {
				a.polarization = a.preisach.Update(a.electricField)
			}
			a.normalizedP = a.preisach.NormalizedPolarization()
		} else {
			a.polarization = 0
			a.normalizedP = 0
		}
	}
	a.syncDiscreteLevelLocked()

	// Calculate energy: integral of E·dP ≈ |E| * |ΔP|
	// During write/read cycles, accumulate energy for the cycle (phases 0-4)
	if a.waveform == WaveformWriteReadDemo && a.wrdPhase >= 0 && a.wrdPhase <= 4 {
		deltaP := a.polarization - prevP
		// Energy per unit volume: E·dP in J/m³
		// Use actual cell dimensions from material (use local copy for thread safety)
		cellVolume := mat.Area * mat.Thickness
		// Fallback if material doesn't have dimensions
		if cellVolume <= 0 {
			cellVolume = 2e-22 // Default: 100nm x 100nm x 20nm
		}
		energyJ := math.Abs(a.electricField*deltaP) * cellVolume
		energyfJ := energyJ * 1e15 // Convert J to fJ
		a.wrdCycleEnergy += energyfJ
	}

	// Record history unconditionally — trail should always be visible.
	// The renderer's spike filter handles visual discontinuities.
	{
		shouldAppend := true
		if a.physicsEngine == PhysicsLandau {
			if a.lastHistorySample < 0 || a.historyLengthLocked() == 0 {
				a.lastHistorySample = a.simTime
			} else if (a.simTime - a.lastHistorySample) < lkHistorySampleInterval {
				shouldAppend = false
			} else {
				a.lastHistorySample = a.simTime
			}
		}
		if shouldAppend {
			a.appendHistoryLocked(a.electricField, a.polarization)
		}
	}

	// Always record full-resolution data snapshot for offline analysis.
	a.recordDataSnapshot(dt)

	return solverDur
}
