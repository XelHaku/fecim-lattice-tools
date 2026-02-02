package main

import (
	"fmt"
	"math"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	hysgui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
)

func runMode(mode string) error {
	switch mode {
	case "hysteresis":
		return runHysteresisMode()
	default:
		return fmt.Errorf("unknown mode %q (expected: hysteresis)", mode)
	}
}

func runHysteresisMode() error {
	log := logging.NewLogger("hysteresis-mode")

	mat := physics.FeCIMMaterial()
	numLevels := mat.GetNumLevels()
	if numLevels <= 0 {
		numLevels = 30
	}
	midLevel := numLevels / 2

	gmin := mat.Gmin
	gmax := mat.Gmax
	if gmin == 0 && gmax == 0 {
		gmin = 1e-6
		gmax = 100e-6
	}

	var dataLogger *hysgui.HysteresisDataLogger
	if logger, err := hysgui.NewHysteresisDataLogger(mat.Name); err != nil {
		log.Info("Headless CSV logging disabled: %v", err)
	} else {
		dataLogger = logger
		log.Info("Headless CSV logging enabled: %s", logger.Path())
		defer func() {
			if err := dataLogger.Close(); err != nil {
				log.Info("Headless CSV logging close failed: %v", err)
			}
		}()
	}

	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false
	solver.UpdateParams()

	log.Info("LK config: Beta=%.3e Gamma=%.3e Rho=%.3e K_dep=%.3e Q12=%.3e Stress=%.2f GPa SeriesR=%.1f Ohm Thickness=%.2e m Area=%.2e m^2 CurieTemp=%.1fK CurieConst=%.2e UseEffVisc=%v UseNLS=%v Noise=%v",
		solver.Beta, solver.Gamma, solver.Rho, solver.K_dep, solver.Q12, solver.Stress/1e9, solver.SeriesResistance, solver.Thickness, solver.Area,
		solver.CurieTemp, solver.CurieConst, solver.UseEffectiveViscosity, solver.UseNLS, solver.EnableNoise)
	log.Info("LK alpha(T,σ)=%.3e at T=%.1fK", solver.Alpha, solver.Temperature)

	Emax := mat.Ec * 1.2
	dt := 1e-12

	log.Info("Landau-Khalatnikov diagnostic sweep starting")
	for _, E := range []float64{-Emax, -0.5 * Emax, 0, 0.5 * Emax, Emax} {
		solver.Step(E, dt)
		recordHeadlessSnapshot(dataLogger, solver, mat, gmin, gmax, numLevels, nil, dt, E, "LK_SWEEP", "SWEEP", nil)
	}

	log.Info("ISPP write-verify sequence starting")
	calibManager := algo.NewCalibrationManager(numLevels)
	writeController := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, calibManager)

	pulseDuration := mat.Tau
	if pulseDuration <= 0 {
		pulseDuration = 10e-9
	}
	writeController.PulseDuration = pulseDuration
	phaseDuration := pulseDuration / 0.4
	if phaseDuration <= 0 {
		phaseDuration = pulseDuration
	}
	frequency := 1.0 / phaseDuration

	dtNominal := 1e-4
	dtMin := 1e-6
	dtMax := 0.025
	stableNominal := pulseDuration / 10000.0
	if stableNominal > 0 && stableNominal < dtNominal {
		dtNominal = stableNominal
	}
	if dtNominal <= 0 {
		dtNominal = 1e-12
	}
	if dtMin > dtNominal {
		dtMin = dtNominal
	}
	if pulseDuration > 0 && pulseDuration < dtMax {
		dtMax = pulseDuration
	}

	gWindow := gmax - gmin
	// Test targets matching baseline: levels 28, 5, 27, 3, 20
	// Level-to-G: G = gmin + (level-1)/(numLevels-1) * gWindow
	levelToG := func(level int) float64 {
		return gmin + float64(level-1)/float64(numLevels-1)*gWindow
	}
	steps := []struct {
		label   string
		targetG float64
	}{
		{"L28", levelToG(28)},
		{"L5", levelToG(5)},
		{"L27", levelToG(27)},
		{"L3", levelToG(3)},
		{"L20", levelToG(20)},
	}

	wrdTotalWrites := 0
	wrdSuccessWrites := 0

	for i, step := range steps {
		_, targetLevel := headlessLevelFromConductance(step.targetG, gmin, gmax, numLevels)
		currentField := 0.0
		currentP := solver.GetState()
		currentG := physics.PolarizationToConductance(currentP, mat.Ps, gmin, gmax)
		_, currentLevel := headlessLevelFromConductance(currentG, gmin, gmax, numLevels)
		stepStart := solver.Time
		maxSimTime := phaseDuration * float64(writeController.MaxRetries+100)
		wrd := &headlessWRDState{
			phase:         0,
			phaseTimer:    0,
			targetLevel:   targetLevel,
			startLevel:    currentLevel,
			readLevel:     currentLevel,
			retryCount:    0,
			cycleEnergy:   0,
			totalWrites:   wrdTotalWrites,
			successWrites: wrdSuccessWrites,
		}

		recordHeadlessSnapshot(dataLogger, solver, mat, gmin, gmax, numLevels, writeController, 0, currentField, "ISPP", "", wrd)

		targetDone := false
		// Track "written P" - the polarization captured at end of pulse (before E→0)
		// This is critical because high K_dep causes P to relax toward 0 when E=0,
		// which would give wrong level readings during VERIFY phase.
		var writtenP float64
		var lastControllerState = controller.StateIdle
		// Track the state at success for accurate reporting (before DISPLAY relaxation)
		var successP float64
		var successLevel int
		for solver.Time-stepStart < maxSimTime && !targetDone {
			// Compute current level from solver state (previous step)
			currentP := solver.GetState()

			// Capture "written P" at transition from WAIT to VERIFY (before E→0)
			// This gives the true written state before depolarization relaxation
			if writeController.State == controller.StateVerify && lastControllerState == controller.StateWait {
				writtenP = currentP
			}
			lastControllerState = writeController.State

			// Use writtenP for level calculation during VERIFY/SUCCESS/FAILED states
			// to avoid depolarization-induced errors at E=0
			effectiveP := currentP
			if writeController.State == controller.StateVerify ||
				writeController.State == controller.StateSuccess ||
				writeController.State == controller.StateFailed {
				if writtenP != 0 {
					effectiveP = writtenP
				}
			}

			currentG := physics.PolarizationToConductance(effectiveP, mat.Ps, gmin, gmax)
			_, currentLevel = headlessLevelFromConductance(currentG, gmin, gmax, numLevels)
			wrd.readLevel = currentLevel

			currentStep := dtNominal
			if mat.Ec > 0 {
				distPlus := math.Abs(currentField - mat.Ec)
				distMinus := math.Abs(currentField + mat.Ec)
				minDist := math.Min(distPlus, distMinus)
				if minDist < 1e7 {
					currentStep = dtMin
				}
			}
			if currentStep > dtMax {
				currentStep = dtMax
			}

			wrd.phaseTimer += currentStep

			switch wrd.phase {
			case 0: // PREP - drive to saturation in opposite direction of target
				// For upper targets (>15): saturate NEGATIVE first (level ~1), then write UP
				// For lower targets (<=15): saturate POSITIVE first (level ~30), then write DOWN
				prepE := 0.0
				saturated := false
				// Use 0.75*Ps threshold (not 0.9) because K_dep depolarization feedback
				// limits achievable polarization. With K_dep=2.5e8, steady-state P < 0.9*Ps.
				// Also use 2.0×Ec drive field for faster saturation.
				if targetLevel > numLevels/2 {
					// Upper target: drive to negative saturation
					prepE = -mat.Ec * 2.0
					saturated = currentP <= -0.75*mat.Ps
				} else {
					// Lower target: drive to positive saturation
					prepE = mat.Ec * 2.0
					saturated = currentP >= 0.75*mat.Ps
				}
				wrd.prepE = prepE

				diff := prepE - currentField
				stepSize := 3.0 * mat.Ec * 2 * frequency * currentStep
				if math.Abs(diff) < stepSize {
					currentField = prepE
				} else if diff > 0 {
					currentField += stepSize
				} else {
					currentField -= stepSize
				}

				// Transition only after field reached AND level actually saturated
				if wrd.phaseTimer > phaseDuration*0.25 && math.Abs(currentField-prepE) < 0.01*mat.Ec*2 && saturated {
					if targetLevel > numLevels/2 {
						wrd.saturatedDirection = -1
					} else {
						wrd.saturatedDirection = 1
					}
					// Skip HOLD_RESET (phase 1) to prevent polarization relaxation
					// Go directly to WRITE phase with fromSaturation=true
					fromSaturation := true
					writeController.PulseDuration = phaseDuration * 0.4
					writeController.Start(targetLevel, fromSaturation)
					wrd.phase = 2
					wrd.phaseTimer = 0
				}

			// case 1: HOLD_RESET - DELETED to prevent polarization relaxation
			// Phase 1 is now skipped - transition goes directly from PREP (0) to WRITE (2)

			case 2: // WRITE - delegated to WriteController
				targetField, done := writeController.Update(currentStep, currentField, currentLevel)

				diff := targetField - currentField
				stepSize := 3.0 * mat.Ec * 2 * frequency * 1.5 * currentStep
				if math.Abs(diff) < stepSize {
					currentField = targetField
				} else if diff > 0 {
					currentField += stepSize
				} else {
					currentField -= stepSize
				}

				wrd.writeE = targetField

				if done {
					switch writeController.State {
					case controller.StateSuccess:
						// Capture the written state BEFORE DISPLAY phase relaxation
						successP = effectiveP
						successLevel = currentLevel
						wrdTotalWrites++
						wrdSuccessWrites++
						wrd.totalWrites = wrdTotalWrites
						wrd.successWrites = wrdSuccessWrites

						// Learn calibration from successful write (match GUI behavior)
						learnedE := writeController.CurrentField
						targetIdx := targetLevel - 1
						if calibManager != nil && targetIdx >= 0 && targetIdx < len(calibManager.CalibrationUp) {
							if targetLevel > midLevel {
								// Ascending (written from reset negative)
								calibManager.UpdateCalibrationUp(targetIdx, 0, mat.Ec)
								calibManager.CalibrationUp[targetIdx] = learnedE
							} else {
								// Descending (written from reset positive)
								calibManager.UpdateCalibrationDown(targetIdx, 0, mat.Ec)
								calibManager.CalibrationDown[targetIdx] = learnedE
							}
							log.Info("CALIB LEARN: target=%d learnedE=%.3f×Ec (updated via CM)", targetLevel, learnedE/mat.Ec)
						}

						wrd.phase = 5
						wrd.phaseTimer = 0
					case controller.StateForceReset:
						wrd.phase = 0
						wrd.phaseTimer = 0
					case controller.StateFailed:
						wrdTotalWrites++
						wrd.totalWrites = wrdTotalWrites
						wrd.phase = 5
						wrd.phaseTimer = 0
					}
				}

			case 5: // DISPLAY - return to zero, then finish target
				stepSize := 3.0 * mat.Ec * 2 * frequency * 0.4 * currentStep
				if math.Abs(currentField) < stepSize {
					currentField = 0
				} else if currentField > 0 {
					currentField -= stepSize
				} else {
					currentField += stepSize
				}

				if wrd.phaseTimer > phaseDuration*0.6 && math.Abs(currentField) < 0.01*mat.Ec*2 {
					targetDone = true
				}
			}

			wrd.retryCount = writeController.RetryCount

			solver.Step(currentField, currentStep)
			recordHeadlessSnapshot(dataLogger, solver, mat, gmin, gmax, numLevels, writeController, currentStep, currentField, "ISPP", "", wrd)
		}

		if !targetDone {
			log.Info("ISPP step %d (%s): timed out after %.3fs", i+1, step.label, solver.Time-stepStart)
		}

		// Use the captured success state if available, otherwise use relaxed state
		finalP := solver.GetState()
		finalG := physics.PolarizationToConductance(finalP, mat.Ps, gmin, gmax)
		_, finalLevel := headlessLevelFromConductance(finalG, gmin, gmax, numLevels)
		success := writeController.State == controller.StateSuccess
		attempts := writeController.PulseCount
		overshoots := writeController.OvershootCount

		// Report the written state (at success) not the relaxed state (after DISPLAY)
		reportP := finalP
		reportLevel := finalLevel
		if success && successP != 0 {
			reportP = successP
			reportLevel = successLevel
		}
		reportG := physics.PolarizationToConductance(reportP, mat.Ps, gmin, gmax)

		log.Info("ISPP step %d (%s): targetG=%.3e targetLevel=%d attempts=%d success=%v overshoots=%d writtenLevel=%d writtenP=%.3e writtenG=%.3e (relaxedP=%.3e)",
			i+1, step.label, step.targetG, targetLevel, attempts, success, overshoots, reportLevel, reportP, reportG, finalP)
	}

	return nil
}

const (
	headlessMVPerCM  = 1e8
	headlessUCPerCM2 = 1e2
)

type headlessWRDState struct {
	phase              int
	phaseTimer         float64
	targetLevel        int
	readLevel          int
	retryCount         int
	cycleEnergy        float64
	totalWrites        int
	successWrites      int
	writeE             float64
	prepE              float64
	settleE            float64
	startLevel         int
	saturatedDirection int // -1 for negative sat, +1 for positive sat, 0 for not saturated
}

func headlessLevelFromConductance(g, gmin, gmax float64, numLevels int) (int, int) {
	if numLevels <= 1 || gmax <= gmin {
		return 0, 1
	}
	ratio := (g - gmin) / (gmax - gmin)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	idx := int(math.Round(ratio * float64(numLevels-1)))
	if idx < 0 {
		idx = 0
	} else if idx > numLevels-1 {
		idx = numLevels - 1
	}
	return idx, idx + 1
}

func headlessStateBand(levelIndex int, numLevels int) string {
	if numLevels <= 0 {
		return "UNKNOWN"
	}
	lowThird := numLevels / 3
	highThird := numLevels * 2 / 3
	if levelIndex < lowThird {
		return "NEGATIVE"
	}
	if levelIndex >= highThird {
		return "POSITIVE"
	}
	return "INTERMEDIATE"
}

func headlessWRDPhaseName(phase int) string {
	switch phase {
	case 0:
		return "RESET"
	case 1:
		return "HOLD_RESET"
	case 2:
		return "WRITE"
	case 3:
		return "HOLD_WRITE"
	case 4:
		return "READ"
	case 5:
		return "DISPLAY"
	case 6:
		return "BOOST"
	default:
		return "UNKNOWN"
	}
}

func headlessPhaseForState(state, waveform string) (int, string) {
	if waveform == "LK_SWEEP" {
		return -1, "LK_SWEEP"
	}
	switch state {
	case "RESETTING":
		return 0, "RESET"
	case "VERIFY":
		return 4, "READ"
	case "SUCCESS":
		return 5, "DISPLAY"
	case "FAILED", "FORCE_RESET":
		return 5, "FAILED"
	case "IDLE":
		return -1, "IDLE"
	default:
		return 2, "WRITE"
	}
}

func recordHeadlessSnapshot(
	logger *hysgui.HysteresisDataLogger,
	solver *physics.LKSolver,
	mat *physics.HZOMaterial,
	gmin float64,
	gmax float64,
	numLevels int,
	ctrl *controller.WriteController,
	dt float64,
	eField float64,
	waveform string,
	stateOverride string,
	wrd *headlessWRDState,
) {
	if logger == nil || solver == nil || mat == nil {
		return
	}

	currentP := solver.GetState()
	currentG := physics.PolarizationToConductance(currentP, mat.Ps, gmin, gmax)
	levelIdx, level := headlessLevelFromConductance(currentG, gmin, gmax, numLevels)
	normalizedP := 0.0
	if mat.Ps != 0 {
		normalizedP = currentP / mat.Ps
		if normalizedP < -1 {
			normalizedP = -1
		} else if normalizedP > 1 {
			normalizedP = 1
		}
	}

	state := stateOverride
	if ctrl != nil {
		state = ctrl.State.String()
	}

	wrdPhase := -1
	wrdPhaseName := ""
	wrdPhaseTimer := 0.0
	wrdTargetLevel := 0
	wrdReadLevel := level
	wrdRetryCount := 0
	wrdCycleEnergy := 0.0
	wrdTotalWrites := 0
	wrdSuccessWrites := 0
	wrdWriteE := 0.0
	wrdPrepE := 0.0
	wrdSettleE := 0.0
	wrdStartLevel := 0
	if wrd != nil {
		wrdPhase = wrd.phase
		wrdPhaseName = headlessWRDPhaseName(wrd.phase)
		wrdPhaseTimer = wrd.phaseTimer
		wrdTargetLevel = wrd.targetLevel
		wrdReadLevel = wrd.readLevel
		wrdRetryCount = wrd.retryCount
		wrdCycleEnergy = wrd.cycleEnergy
		wrdTotalWrites = wrd.totalWrites
		wrdSuccessWrites = wrd.successWrites
		wrdWriteE = wrd.writeE
		wrdPrepE = wrd.prepE
		wrdSettleE = wrd.settleE
		wrdStartLevel = wrd.startLevel
	} else {
		wrdPhase, wrdPhaseName = headlessPhaseForState(state, waveform)
	}

	snapshot := hysgui.HysteresisSnapshot{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		SimTime:       solver.Time,
		Dt:            dt,
		Waveform:      waveform,
		AutoMode:      true,
		Material:      mat.Name,
		TemperatureK:  solver.Temperature,
		EcMVcm:        mat.Ec / headlessMVPerCM,
		PsUcCm2:       mat.Ps * headlessUCPerCM2,
		PrUcCm2:       mat.Pr * headlessUCPerCM2,
		NumLevels:     numLevels,
		LevelIndex:    levelIdx,
		Level:         level,
		StateBand:     headlessStateBand(levelIdx, numLevels),
		EField:        eField,
		EFieldMVcm:    eField / headlessMVPerCM,
		Polarization:  currentP,
		PolarizationU: currentP * headlessUCPerCM2,
		NormalizedP:   normalizedP,

		WrdPhase:       wrdPhase,
		WrdPhaseName:   wrdPhaseName,
		WrdPhaseTimer:  wrdPhaseTimer,
		WrdTargetLevel: wrdTargetLevel,
		WrdReadLevel:   wrdReadLevel,
		WrdRetryCount:  wrdRetryCount,
		WrdCycleEnergy: wrdCycleEnergy,
		WrdTotalWrites: wrdTotalWrites,
		WrdSuccess:     wrdSuccessWrites,
		WrdWriteE:      wrdWriteE,
		WrdPrepE:       wrdPrepE,
		WrdSettleE:     wrdSettleE,
		WrdStartLevel:  wrdStartLevel,

		ControllerState:          state,
		ControllerPhaseTimer:     0,
		ControllerTargetLevel:    0,
		ControllerCurrentField:   0,
		ControllerCurrentFieldMV: 0,
		ControllerPulseCount:     0,
		ControllerTotalPulses:    0,
		ControllerRetryCount:     0,
		ControllerOvershootCount: 0,
		ControllerOvershootTotal: 0,
		ControllerLastVerify:     0,
		ControllerLastError:      0,
		ControllerVMin:           0,
		ControllerVMax:           0,
		ControllerInitialLevel:   0,
		ControllerFromSaturation: false,
		ControllerResetDirection: 0,
	}

	if ctrl != nil {
		snapshot.ControllerPhaseTimer = ctrl.PhaseTimer
		snapshot.ControllerTargetLevel = ctrl.TargetLevel
		snapshot.ControllerCurrentField = ctrl.CurrentField
		snapshot.ControllerCurrentFieldMV = ctrl.CurrentField / headlessMVPerCM
		snapshot.ControllerPulseCount = ctrl.PulseCount
		snapshot.ControllerTotalPulses = ctrl.TotalPulses
		snapshot.ControllerRetryCount = ctrl.RetryCount
		snapshot.ControllerOvershootCount = ctrl.OvershootCount
		snapshot.ControllerOvershootTotal = ctrl.OvershootTotal
		snapshot.ControllerLastVerify = ctrl.LastVerifyLevel
		snapshot.ControllerLastError = ctrl.LastError
		snapshot.ControllerVMin = ctrl.VMin
		snapshot.ControllerVMax = ctrl.VMax
		snapshot.ControllerInitialLevel = ctrl.InitialLevel
		snapshot.ControllerFromSaturation = ctrl.FromSaturation
		snapshot.ControllerResetDirection = ctrl.ResetDirection()

		if ctrl.EcField != 0 {
			snapshot.ControllerVMinEc = ctrl.VMin / ctrl.EcField
			snapshot.ControllerVMaxEc = ctrl.VMax / ctrl.EcField
		}
	}

	logger.Record(snapshot)
}
